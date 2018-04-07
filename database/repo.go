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
	var repo Repo
	err := db.sqldb.QueryRow("SELECT id, org_name, repo_name, last_retrieval FROM repos WHERE id = $1", id).Scan(&repo.Id, &repo.OrgName, &repo.RepoName, &repo.LastRetrieval)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func (db *DB) InsertRepo(orgName string, repoName string) (*Repo, error) {
	// FIXME Change to prepared statements!
	var id int
	zeroTime := time.Time{}
	err := db.sqldb.QueryRow("INSERT INTO repos (org_name, repo_name, last_retrieval) VALUES ($1, $2, $3) RETURNING id", orgName, repoName, zeroTime).Scan(&id)
	if err != nil {
		return nil, err
	}

	repo := &Repo{Id: id, OrgName: orgName, RepoName: repoName, LastRetrieval: zeroTime}
	return repo, nil
}
