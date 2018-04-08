// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"strconv"
	"time"
)

func (db *DB) createDBRepoTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS repos (
			id SERIAL NOT NULL PRIMARY KEY,
			org_name TEXT NOT NULL,
			repo_name TEXT NOT NULL,
			last_retrieval TIMESTAMP NOT NULL
		)
	`)
	return err
}

type Repo struct {
	Id            int
	OrgName       string
	RepoName      string
	LastRetrieval time.Time
}

func (db *DB) GetRepoById(id int) (*Repo, error) {
	stmt, err := db.getStatement(stmtRepoGet)
	if err != nil {
		return nil, err
	}

	var repo Repo
	err = stmt.QueryRow(id).Scan(&repo.Id, &repo.OrgName, &repo.RepoName, &repo.LastRetrieval)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func (db *DB) InsertRepo(orgName string, repoName string) (*Repo, error) {
	stmt, err := db.getStatement(stmtRepoInsert)
	if err != nil {
		return nil, err
	}

	var id int
	zeroTime := time.Time{}
	err = stmt.QueryRow(orgName, repoName, zeroTime).Scan(&id)
	if err != nil {
		return nil, err
	}

	repo := &Repo{Id: id, OrgName: orgName, RepoName: repoName, LastRetrieval: zeroTime}
	return repo, nil
}

func (db *DB) UpdateRepoLastRetrieval(repo *Repo, lr time.Time) error {
	stmt, err := db.getStatement(stmtRepoUpdateLastRetrieval)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(lr, repo.Id)
	if err != nil {
		return err
	}

	rowCount, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowCount != 1 {
		return errors.New("UpdateRepoLastRetrieval for ID " + strconv.Itoa(repo.Id) +
			" modified " + strconv.FormatInt(rowCount, 10) + " rows, should be 1")
	}

	// update in-memory copy of repo
	repo.LastRetrieval = lr

	return nil
}
