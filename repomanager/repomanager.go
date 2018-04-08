// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package repomanager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
	"gopkg.in/src-d/go-git.v4"
	git_obj "gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/swinslow/lfscanning/config"
	"github.com/swinslow/lfscanning/database"
)

type RepoManager struct {
	ReposPath string
	db        *database.DB
}

func InitRepoManager() *RepoManager {
	return &RepoManager{}
}

func (rm *RepoManager) PrepareRepoManager(cfg *config.Config, db *database.DB) error {
	if rm == nil {
		return errors.New("must pass non-nil RepoManager")
	}
	if cfg == nil || cfg.ReposLocation == "" {
		return errors.New("must pass config string")
	}
	if db == nil {
		return errors.New("must prepare and pass database first")
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
		return errors.New(path + " is not a directory")
	}

	// check whether path is writable
	if unix.Access(path, unix.W_OK) != nil {
		return errors.New(path + " is not writable")
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
	_, err := git.PlainClone(dstPath, false, &git.CloneOptions{
		URL:               srcURL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	})

	if err != nil {
		return err
	}

	rm.db.UpdateRepoLastRetrieval(repo, time.Now())

	return nil
}

func (rm *RepoManager) GetRepoLatestCommit(repo *database.Repo) (*git_obj.Commit, error) {
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

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	rm.db.UpdateRepoLastRetrieval(repo, time.Now())

	return nil
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
	tree.Files().ForEach(func(f *git_obj.File) error {
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

	// walk through files; 50 is arbitrary
	filePaths := make([]string, 50)
	tree.Files().ForEach(func(f *git_obj.File) error {
		filePaths = append(filePaths, f.Name)
		return nil
	})

	return filePaths, nil
}
