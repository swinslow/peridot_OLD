// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"fmt"
	"strconv"
	"time"
)

func (db *DB) createDBRepoRetrievalsTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS reporetrievals (
			id SERIAL NOT NULL PRIMARY KEY,
			repo_id INTEGER NOT NULL,
			last_retrieval TIMESTAMP NOT NULL,
			commit_hash TEXT NOT NULL,
			FOREIGN KEY (repo_id) REFERENCES repos (id)
		)
	`)
	return err
}

// RepoRetrieval stores the data for a single point-in-time retrieval of a
// source code repository that is being tracked in peridot.
type RepoRetrieval struct {
	ID            int
	RepoID        int
	LastRetrieval time.Time
	CommitHash    string
}

// GetRepoRetrievalByID looks up and returns a RepoRetrieval in the database
// by its ID. It returns nil if no Repo with the requested ID is found.
func (db *DB) GetRepoRetrievalByID(id int) (*RepoRetrieval, error) {
	stmt, err := db.getStatement(stmtRepoRetrievalGet)
	if err != nil {
		return nil, err
	}

	var repoRetrieval RepoRetrieval
	err = stmt.QueryRow(id).Scan(&repoRetrieval.ID, &repoRetrieval.RepoID,
		&repoRetrieval.LastRetrieval, &repoRetrieval.CommitHash)
	if err != nil {
		return nil, err
	}

	return &repoRetrieval, nil
}

// GetRepoRetrievalLatest looks up and returns the most recent RepoRetrieval
// in the database for a given Repo's ID. It returns nil if no Repo for the
// requested Repo ID is found.
func (db *DB) GetRepoRetrievalLatest(repoID int) (*RepoRetrieval, error) {
	stmt, err := db.getStatement(stmtRepoRetrievalGetLatest)
	if err != nil {
		return nil, err
	}

	var repoRetrieval RepoRetrieval
	err = stmt.QueryRow(repoID).Scan(&repoRetrieval.ID, &repoRetrieval.RepoID,
		&repoRetrieval.LastRetrieval, &repoRetrieval.CommitHash)
	if err != nil {
		return nil, err
	}

	return &repoRetrieval, nil
}

// InsertRepoRetrieval takes a new repo retrieval's data, creates a new
// RepoRetrieval struct, adds it to the database, and returns the new struct
// with its ID from the DB.
func (db *DB) InsertRepoRetrieval(repoID int, lr time.Time, ch string) (*RepoRetrieval, error) {
	stmt, err := db.getStatement(stmtRepoRetrievalInsert)
	if err != nil {
		return nil, err
	}

	var id int
	err = stmt.QueryRow(repoID, lr, ch).Scan(&id)
	if err != nil {
		return nil, err
	}

	repoRet := &RepoRetrieval{ID: id, RepoID: repoID, LastRetrieval: lr, CommitHash: ch}
	return repoRet, nil
}

// UpdateRepoRetrieval updates a given RepoRetrieval's data in both the
// database and its in-memory struct.
func (db *DB) UpdateRepoRetrieval(repoRetrieval *RepoRetrieval, lr time.Time, ch string) error {
	stmt, err := db.getStatement(stmtRepoRetrievalUpdate)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(lr, ch, repoRetrieval.ID)
	if err != nil {
		return err
	}

	rowCount, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowCount != 1 {
		return fmt.Errorf("UpdateRepoRetrieval for ID %s modified %s rows, should be 1",
			strconv.FormatInt(rowCount, 10), strconv.Itoa(repoRetrieval.ID))
	}

	// update in-memory copy of repo
	repoRetrieval.LastRetrieval = lr
	repoRetrieval.CommitHash = ch

	return nil
}
