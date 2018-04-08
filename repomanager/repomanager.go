// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package repomanager

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
	"gopkg.in/src-d/go-git.v4"

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
	return "https://github.com/" + repo.OrgName + "/" + repo.RepoName
}

func (rm *RepoManager) CloneRepo(repo *database.Repo) error {
	dstPath := rm.GetPathToRepo(repo)
	srcURL := rm.GetURLToRepo(repo)
	_, err := git.PlainClone(dstPath, false, &git.CloneOptions{
		URL:      srcURL,
		Progress: os.Stdout,
	})

	if err != nil {
		return err
	}

	rm.db.UpdateRepoLastRetrieval(repo, time.Now())

	return nil
}
