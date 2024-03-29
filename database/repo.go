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

// Repo stores the coordinates for a source code repository that is being
// tracked in peridot.
type Repo struct {
	ID       int
	OrgName  string
	RepoName string
}

// GetRepoByID looks up and returns a Repo in the database by its ID.
// It returns nil if no Repo with the requested ID is found.
func (db *DB) GetRepoByID(id int) (*Repo, error) {
	stmt, err := db.getStatement(stmtRepoGet)
	if err != nil {
		return nil, err
	}

	var repo Repo
	err = stmt.QueryRow(id).Scan(&repo.ID, &repo.OrgName, &repo.RepoName)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetRepoIDFromCoords takes a Github repo's coordinates and returns their ID
// from the database, or returns 0, nil if repo not found for these coords.
func (db *DB) GetRepoIDFromCoords(orgName string, repoName string) (int, error) {
	stmt, err := db.getStatement(stmtRepoGetByCoords)
	if err != nil {
		return -1, err
	}

	var id int
	err = stmt.QueryRow(orgName, repoName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return -1, err
	}

	return id, nil
}

// InsertRepo takes a Github repo's coordinates, creates a new Repo struct,
// adds it to the database, and returns the new struct with its ID from the DB.
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

	repo := &Repo{ID: id, OrgName: orgName, RepoName: repoName}
	return repo, nil
}
