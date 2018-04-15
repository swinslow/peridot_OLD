// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
)

func (db *DB) createDBReposTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS repos (
			id SERIAL NOT NULL PRIMARY KEY,
			org_name TEXT NOT NULL,
			repo_name TEXT NOT NULL
		)
	`)
	return err
}

type Repo struct {
	Id       int
	OrgName  string
	RepoName string
}

func (db *DB) GetRepoById(id int) (*Repo, error) {
	stmt, err := db.getStatement(stmtRepoGet)
	if err != nil {
		return nil, err
	}

	var repo Repo
	err = stmt.QueryRow(id).Scan(&repo.Id, &repo.OrgName, &repo.RepoName)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

// returns 0, nil if repo not found for these coords
func (db *DB) GetRepoIdFromCoords(orgName string, repoName string) (int, error) {
	stmt, err := db.getStatement(stmtRepoGetByCoords)
	if err != nil {
		return -1, err
	}

	var id int
	err = stmt.QueryRow(orgName, repoName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		} else {
			return -1, err
		}
	}

	return id, nil
}

func (db *DB) InsertRepo(orgName string, repoName string) (*Repo, error) {
	stmt, err := db.getStatement(stmtRepoInsert)
	if err != nil {
		return nil, err
	}

	var id int
	err = stmt.QueryRow(orgName, repoName).Scan(&id)
	if err != nil {
		return nil, err
	}

	repo := &Repo{Id: id, OrgName: orgName, RepoName: repoName}
	return repo, nil
}
