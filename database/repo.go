// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"time"
)

func (db *DB) CreateDBRepoTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS repos (
			id SERIAL NOT NULL PRIMARY KEY,
			org_name TEXT NOT NULL,
			repo_name TEXT NOT NULL,
			last_retrieval TIMESTAMP NOT NULL
		);
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
	// FIXME Change to prepared statements!
	row := db.sqldb.QueryRow("SELECT id, org_name, repo_name, last_retrieval FROM repos WHERE id = $1", id)

	repo := &Repo{}
	err := row.Scan(&repo.Id, &repo.OrgName, &repo.RepoName, &repo.LastRetrieval)
	if err != nil {
		return nil, err
	}

	return repo, nil
}
