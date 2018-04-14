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

type RepoRetrieval struct {
	Id            int
	RepoId        int
	LastRetrieval time.Time
	CommitHash    string
}

func (db *DB) GetRepoRetrievalById(id int) (*RepoRetrieval, error) {
	stmt, err := db.getStatement(stmtRepoRetrievalGet)
	if err != nil {
		return nil, err
	}

	var repoRetrieval RepoRetrieval
	err = stmt.QueryRow(id).Scan(&repoRetrieval.Id, &repoRetrieval.RepoId,
		&repoRetrieval.LastRetrieval, &repoRetrieval.CommitHash)
	if err != nil {
		return nil, err
	}

	return &repoRetrieval, nil
}

func (db *DB) GetRepoRetrievalLatest(repoId int) (*RepoRetrieval, error) {
	stmt, err := db.getStatement(stmtRepoRetrievalGetLatest)
	if err != nil {
		return nil, err
	}

	var repoRetrieval RepoRetrieval
	err = stmt.QueryRow(repoId).Scan(&repoRetrieval.Id, &repoRetrieval.RepoId,
		&repoRetrieval.LastRetrieval, &repoRetrieval.CommitHash)
	if err != nil {
		return nil, err
	}

	return &repoRetrieval, nil
}

func (db *DB) InsertRepoRetrieval(repoId int, lr time.Time, ch string) (*RepoRetrieval, error) {
	stmt, err := db.getStatement(stmtRepoRetrievalInsert)
	if err != nil {
		return nil, err
	}

	var id int
	err = stmt.QueryRow(repoId, lr, ch).Scan(&id)
	if err != nil {
		return nil, err
	}

	repoRet := &RepoRetrieval{Id: id, RepoId: repoId, LastRetrieval: lr, CommitHash: ch}
	return repoRet, nil
}

func (db *DB) UpdateRepoRetrieval(repoRetrieval *RepoRetrieval, lr time.Time, ch string) error {
	stmt, err := db.getStatement(stmtRepoRetrievalUpdate)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(lr, ch, repoRetrieval.Id)
	if err != nil {
		return err
	}

	rowCount, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowCount != 1 {
		return fmt.Errorf("UpdateRepoRetrieval for ID %d modified %d rows, should be 1",
			strconv.FormatInt(rowCount, 10), strconv.Itoa(repoRetrieval.Id))
	}

	// update in-memory copy of repo
	repoRetrieval.LastRetrieval = lr
	repoRetrieval.CommitHash = ch

	return nil
}
