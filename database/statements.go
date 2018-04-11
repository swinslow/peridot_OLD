// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"errors"
)

// getStatement is not exported, because we don't want anyone outside
// the database package touching the database directly, even to
// retrieve data
func (db *DB) getStatement(sv dbStatementVal) (*sql.Stmt, error) {
	if int(sv) > len(db.stmts) {
		return nil, errors.New("invalid statement number in getStatement, statement not prepared")
	}

	return db.stmts[sv], nil
}

func (db *DB) addStatement(sv dbStatementVal, s string) error {
	if sv < 0 {
		return errors.New("negative statement number in addStatement")
	}

	stmt, err := db.sqldb.Prepare(s)
	if err != nil {
		return err
	}

	// increase stmts length until we are large enough
	// we'll come back and fill in any holes later, if statements are
	// added out of order
	for i := len(db.stmts); i <= int(sv); i++ {
		db.stmts = append(db.stmts, nil)
	}

	// now that there's enough space, insert the statement
	db.stmts[sv] = stmt
	return nil
}

func (db *DB) createDBTablesIfNotExists() error {
	var err error
	err = db.createDBReposTableIfNotExists()
	if err != nil {
		return err
	}

	err = db.createDBRepoRetrievalsTableIfNotExists()
	if err != nil {
		return err
	}

	err = db.createDBRepoDirsTableIfNotExists()
	if err != nil {
		return err
	}

	err = db.createDBRepoFilesTableIfNotExists()
	if err != nil {
		return err
	}

	return nil
}

// statement type and enum
type dbStatementVal int

const (
	stmtRepoGet dbStatementVal = iota
	stmtRepoInsert
	stmtRepoRetrievalGet
	stmtRepoRetrievalGetLatest
	stmtRepoRetrievalInsert
	stmtRepoRetrievalUpdate
	stmtRepoFileGet
	stmtRepoFileInsert
	stmtRepoDirGet
	stmtRepoDirInsert
)

// master prepare function
func (db *DB) prepareStatements() error {
	var err error

	err = db.prepareStatementsRepos()
	if err != nil {
		return err
	}
	err = db.prepareStatementsRepoRetrievals()
	if err != nil {
		return err
	}
	err = db.prepareStatementsRepoFiles()
	if err != nil {
		return err
	}
	err = db.prepareStatementsRepoDirs()
	if err != nil {
		return err
	}

	return nil
}

// table-specific prepare functions

// table repos
func (db *DB) prepareStatementsRepos() error {
	var err error

	err = db.addStatement(stmtRepoGet, `
		SELECT id, org_name, repo_name
		FROM repos
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoInsert, `
		INSERT INTO repos (org_name, repo_name)
		VALUES ($1, $2)
		RETURNING id
	`)
	if err != nil {
		return err
	}

	return nil
}

// table reporetrievals
func (db *DB) prepareStatementsRepoRetrievals() error {
	var err error

	err = db.addStatement(stmtRepoRetrievalGet, `
		SELECT id, repo_id, last_retrieval, commit_hash
		FROM reporetrievals
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoRetrievalGetLatest, `
		SELECT id, repo_id, last_retrieval, commit_hash
		FROM reporetrievals
		WHERE repo_id = $1
		ORDER BY last_retrieval DESC
		LIMIT 1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoRetrievalInsert, `
		INSERT INTO reporetrievals (repo_id, last_retrieval, commit_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoRetrievalUpdate, `
		UPDATE reporetrievals
		SET last_retrieval = $1, commit_hash = $2
		WHERE id = $3
	`)
	if err != nil {
		return err
	}

	return nil
}

// table repofiles
func (db *DB) prepareStatementsRepoFiles() error {
	var err error

	err = db.addStatement(stmtRepoFileGet, `
		SELECT id, repo_id, dir_parent_id, nextfile_id, prevfile_id,
		       path, hash_sha1
		FROM repofiles
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoFileInsert, `
		INSERT INTO repofiles (repo_id, dir_parent_id,
			nextfile_id, prevfile_id, path, hash_sha1)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`)
	if err != nil {
		return err
	}

	return nil
}

// table repofiles
func (db *DB) prepareStatementsRepoDirs() error {
	var err error

	err = db.addStatement(stmtRepoDirGet, `
		SELECT id, repo_id, dir_parent_id, path
		FROM repodirs
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoDirInsert, `
		INSERT INTO repodirs (repo_id, dir_parent_id, path)
		VALUES ($1, $2, $3)
		RETURNING id
	`)
	if err != nil {
		return err
	}

	return nil
}
