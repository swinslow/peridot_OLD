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

type RepoManager struct {
	ReposPath string
	db        *database.DB
}

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

func (rm *RepoManager) GetPathToRepo(repo *database.Repo) string {
	return filepath.Join(rm.ReposPath, repo.OrgName, repo.RepoName)
}

func (rm *RepoManager) GetURLToRepo(repo *database.Repo) string {
	// FIXME check for e.g. no slashes or problematic chars in OrgName or RepoName!
	// FIXME perhaps use path.Join, except not clear yet how to use it with two slashes
	// FIXME for protocol name
	return "https://github.com/" + repo.OrgName + "/" + repo.RepoName + ".git"
}

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

// returns map of paths to array of file hashes: SHA1, SHA256, MD5
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
		h_sha1 := sha1.New()
		h_sha256 := sha256.New()
		h_md5 := md5.New()
		hMulti := io.MultiWriter(h_sha1, h_sha256, h_md5)

		if _, err := io.Copy(hMulti, f); err != nil {
			f.Close()
			return nil, err
		}
		hashes[0] = fmt.Sprintf("%x", h_sha1.Sum(nil))
		hashes[1] = fmt.Sprintf("%x", h_sha256.Sum(nil))
		hashes[2] = fmt.Sprintf("%x", h_md5.Sum(nil))

		pathsToHashes[path] = hashes

		f.Close()
	}

	return pathsToHashes, nil
}
