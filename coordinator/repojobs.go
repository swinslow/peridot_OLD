// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package coordinator

import (
	"fmt"

	"github.com/swinslow/peridot/database"
)

// DoCloneRepo is the function for JobCloneRepo, and performs the first
// retrieval of files for a new repo.
func (co *Coordinator) DoCloneRepo(repoID int) error {
	repo, err := co.db.GetRepoById(repoID)
	if err != nil {
		return fmt.Errorf("couldn't get repo data from DB: %v", err)
	}

	err = co.rm.CloneRepo(repo)
	if err != nil {
		return fmt.Errorf("couldn't clone repo: %v", err)
	}

	return nil
}

// DoUpdateRepo is the function for JobUpdateRepo, and returns true if
// an update occurred, or false if there are no changes to the repo.
func (co *Coordinator) DoUpdateRepo(repoID int) (bool, error) {
	repo, err := co.db.GetRepoById(repoID)
	if err != nil {
		return false, fmt.Errorf("couldn't get repo from DB: %v", err)
	}

	repoRetBefore, err := co.db.GetRepoRetrievalLatest(repoID)
	if err != nil {
		return false, fmt.Errorf("couldn't get repo retrieval from DB before update: %v", err)
	}

	err = co.rm.UpdateRepo(repo)
	if err != nil {
		return false, fmt.Errorf("couldn't checking remote for updates: %v", err)
	}

	repoRetAfter, err := co.db.GetRepoRetrievalLatest(repoID)
	if err != nil {
		return false, fmt.Errorf("couldn't get repo retrieval from DB after update: %v", err)
	}

	if repoRetBefore.Id == repoRetAfter.Id {
		// no update occurred
		return false, nil
	}

	// a new RepoRetrieval was created, so there was an update
	return true, nil
}

// DoPrepareFiles is the function for JobPrepareFiles, and is called after a
// JobCloneRepo or JobUpdateRepo to set up the files in the repo and hash
// managers.
func (co *Coordinator) DoPrepareFiles(repoID int) error {
	repo, err := co.db.GetRepoById(repoID)
	if err != nil {
		return fmt.Errorf("couldn't get repo from DB: %v", err)
	}

	repoRetrieval, err := co.db.GetRepoRetrievalLatest(repoID)
	if err != nil {
		return fmt.Errorf("couldn't get repo retrieval from DB: %v", err)
	}

	allPaths, err := co.rm.GetAllFilepaths(repo)
	if err != nil {
		return fmt.Errorf("couldn't get filepaths for repo: %v", err)
	}

	dirPaths := database.ExtractDirsFromPaths(allPaths)

	// split and add directories to DB
	err = co.db.BulkInsertRepoDirs(repoRetrieval.Id, dirPaths)
	if err != nil {
		return fmt.Errorf("couldn't insert repo directories into DB: %v", err)
	}

	pathsToHashes, err := co.rm.GetFileHashes(repo)
	if err != nil {
		return fmt.Errorf("couldn't get file hashes: %v", err)
	}

	// add files to DB for this repo
	err = co.db.BulkInsertRepoFiles(repoRetrieval.Id, pathsToHashes)
	if err != nil {
		return fmt.Errorf("couldn't insert repo files into DB: %v", err)
	}

	// also add files to hashmanager
	pathRoot := co.rm.GetPathToRepo(repo)
	copiedPathsToHashes, err := co.hm.CopyAllFilesToHash(pathRoot, pathsToHashes)
	if err != nil {
		return fmt.Errorf("couldn't copy files to hashes: %v", err)
	}

	// and finally add just the newly-copied files as hashfiles to DB
	err = co.db.BulkInsertHashFiles(copiedPathsToHashes)
	if err != nil {
		return fmt.Errorf("couldn't insert hash files into DB: %v", err)
	}

	return nil
}
