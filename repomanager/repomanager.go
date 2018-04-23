// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package repomanager

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
	"gopkg.in/src-d/go-git.v4"
	gitObject "gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/swinslow/peridot/config"
	"github.com/swinslow/peridot/database"
)

// RepoManager holds the data objects needed to manage copies of files
// scanned by peridot that are stored on disk in git-managed repos.
type RepoManager struct {
	ReposPath string
	db        *database.DB
}

// PrepareRM is called with existing Config and Database objects and
// initializes the repo manager's on-disk storage location.
func (rm *RepoManager) PrepareRM(cfg *config.Config, db *database.DB) error {
	if rm == nil {
		return fmt.Errorf("must pass non-nil RepoManager")
	}
	if cfg == nil || cfg.ReposLocation == "" {
		return fmt.Errorf("must pass config string")
	}
	if db == nil {
		return fmt.Errorf("must prepare and pass database")
	}

	err := rm.setReposLocation(cfg.ReposLocation)
	if err != nil {
		return err
	}

	rm.db = db
	return nil
}

func (rm *RepoManager) setReposLocation(path string) error {
	// check whether path exists in filesystem
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	// check whether path is a directory
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	// check whether path is writable
	if unix.Access(path, unix.W_OK) != nil {
		return fmt.Errorf("%s is not writable", path)
	}

	// we're good
	rm.ReposPath = path
	return nil
}

// GetPathToRepo takes a repo structure and returns the full on-disk
// pathname to locate that repo.
func (rm *RepoManager) GetPathToRepo(repo *database.Repo) string {
	return filepath.Join(rm.ReposPath, repo.OrgName, repo.RepoName)
}

// GetURLToRepo takes a repo structure and returns the full URL for where
// that repo can be found.
func (rm *RepoManager) GetURLToRepo(repo *database.Repo) string {
	// FIXME check for e.g. no slashes or problematic chars in OrgName or RepoName!
	// FIXME perhaps use path.Join, except not clear yet how to use it with two slashes
	// FIXME for protocol name
	return "https://github.com/" + repo.OrgName + "/" + repo.RepoName + ".git"
}

// CloneRepo takes a Repo that is already in the database, and makes an
// initial clone of its contents onto disk, creating and adding a first
// RepoRetrieval to the database.
func (rm *RepoManager) CloneRepo(repo *database.Repo) error {
	dstPath := rm.GetPathToRepo(repo)
	srcURL := rm.GetURLToRepo(repo)
	r, err := git.PlainClone(dstPath, false, &git.CloneOptions{
		URL:               srcURL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	})
	if err != nil {
		return err
	}

	// get HEAD reference
	ref, err := r.Head()
	if err != nil {
		return err
	}

	// get commit history
	ci, err := r.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return err
	}
	defer ci.Close()

	// get first commit from iter
	commit, err := ci.Next()
	if err != nil {
		return err
	}

	// and insert time and hash for commit
	// FIXME determine whether we need to do all this or can just use ref.Hash()
	_, err = rm.db.InsertRepoRetrieval(repo.Id, time.Now(), commit.Hash.String())
	if err != nil {
		return err
	}

	return nil
}

// GetRepoLatestCommit takes a Repo and returns a reference to the most recent
// git commit on disk for that repo.
func (rm *RepoManager) GetRepoLatestCommit(repo *database.Repo) (*gitObject.Commit, error) {
	repoPath := rm.GetPathToRepo(repo)
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	// get HEAD reference
	ref, err := r.Head()
	if err != nil {
		return nil, err
	}

	// get commit history
	ci, err := r.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}
	defer ci.Close()

	// get first commit from iter
	commit, err := ci.Next()
	if err != nil {
		return nil, err
	}

	return commit, nil
}

// UpdateRepo takes a Repo that has already been cloned to disk previously
// via CloneRepo, and checks with the remote origin to see if there are any
// newer updates. If there are, it retrieves them and creates a new
// RepoRetrieval in the database. If there aren't, it updates the most
// recent RepoRetrieval in the database to flag that it is still current as
// of the present time.
func (rm *RepoManager) UpdateRepo(repo *database.Repo) error {
	repoPath := rm.GetPathToRepo(repo)
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// pull an update
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	// get HEAD reference
	ref, err := r.Head()
	if err != nil {
		return err
	}

	// get commit history
	ci, err := r.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return err
	}
	defer ci.Close()

	// get first commit from iter
	commit, err := ci.Next()
	if err != nil {
		return err
	}

	// FIXME determine whether we need to do all this or can just use ref.Hash()

	// get the most current RepoRetrieval so we can decide whether to
	// update it (if commit is the same) or to insert a new one (otherwise)
	repoRet, err := rm.db.GetRepoRetrievalLatest(repo.Id)
	commitHash := commit.Hash.String()
	if err != nil || commitHash != repoRet.CommitHash {
		_, err = rm.db.InsertRepoRetrieval(repo.Id, time.Now(), commitHash)
	} else {
		err = rm.db.UpdateRepoRetrieval(repoRet, time.Now(), commitHash)
	}

	// return back whatever err (or nil) we got from the insert / update call
	return err
}

// WalkAndPrintFiles is a testing / convenience function that walks through a
// repo and prints each corresponding file object in that repo. If a file's
// path matches the path parameter, it will be highlighted with an arrow
// prefix in the output.
func (rm *RepoManager) WalkAndPrintFiles(repo *database.Repo, path string) error {
	repoPath := rm.GetPathToRepo(repo)
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	// get HEAD reference
	ref, err := r.Head()
	if err != nil {
		return err
	}

	// get latest commit object
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	// get tree from commit
	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	// walk through files
	tree.Files().ForEach(func(f *gitObject.File) error {
		if f.Name == path {
			fmt.Printf("===> ")
		}
		fmt.Printf("%s: %s\n", f.Name, f.Hash)
		return nil
	})

	return nil
}

// GetAllFilepaths takes a Repo and returns a string slice containing the
// paths for all files in that repo as currently found on disk.
func (rm *RepoManager) GetAllFilepaths(repo *database.Repo) ([]string, error) {
	// FIXME would this be more efficient if it went through the filesystem
	// FIXME directly, via path/filepath.Walk()?
	repoPath := rm.GetPathToRepo(repo)
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	// get HEAD reference
	ref, err := r.Head()
	if err != nil {
		return nil, err
	}

	// get latest commit object
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	// get tree from commit
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	// walk through files
	var filePaths []string
	tree.Files().ForEach(func(f *gitObject.File) error {
		filePaths = append(filePaths, f.Name)
		return nil
	})

	return filePaths, nil
}

// GetFileHashes takes a Repo and returns a map of strings from path (for
// all files in that repo as currently found on disk) to a 3-element string
// array, with that file's hashes in order: SHA1, SHA256, MD5.
func (rm *RepoManager) GetFileHashes(repo *database.Repo) (map[string][3]string, error) {
	allPaths, err := rm.GetAllFilepaths(repo)
	if err != nil {
		return nil, err
	}

	pathsToHashes := make(map[string][3]string)
	pathRoot := rm.GetPathToRepo(repo)

	for _, path := range allPaths {
		fullPath := filepath.Join(pathRoot, path)
		f, err := os.Open(fullPath)
		if err != nil {
			return nil, err
		}
		// don't defer f.Close() here, b/c we're in a loop

		var hashes [3]string
		hSHA1 := sha1.New()
		hSHA256 := sha256.New()
		hMD5 := md5.New()
		hMulti := io.MultiWriter(hSHA1, hSHA256, hMD5)

		if _, err := io.Copy(hMulti, f); err != nil {
			f.Close()
			return nil, err
		}
		hashes[0] = fmt.Sprintf("%x", hSHA1.Sum(nil))
		hashes[1] = fmt.Sprintf("%x", hSHA256.Sum(nil))
		hashes[2] = fmt.Sprintf("%x", hMD5.Sum(nil))

		pathsToHashes[path] = hashes

		f.Close()
	}

	return pathsToHashes, nil
}
