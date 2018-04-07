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

// statement type and enum
type dbStatementVal int

const (
	stmtRepoGet dbStatementVal = iota
	stmtRepoInsert
)

// master prepare function
func (db *DB) prepareStatements() error {
	var err error

	err = db.prepareStatementsRepos()
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
		SELECT id, org_name, repo_name, last_retrieval
		FROM repos
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoInsert, `
		INSERT INTO repos (org_name, repo_name, last_retrieval)
		VALUES ($1, $2, $3)
		RETURNING id
	`)
	if err != nil {
		return err
	}

	return nil
}
