// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"fmt"
)

// Note that the order is important, so that DBReset() can drop tables
// in the correct order (e.g., dependent tables dropped before those
// they depend upon).
var tables = []string{
	"hashfiles",
	"repofiles",
	"repodirs",
	"reporetrievals",
	"licensenodes",
	"licenseleafs",
	"repos",
}

// getStatement is not exported, because we don't want anyone outside
// the database package touching the database directly, even to
// retrieve data.
func (db *DB) getStatement(sv dbStatementVal) (*sql.Stmt, error) {
	if int(sv) > len(db.stmts) {
		return nil, fmt.Errorf("invalid statement number %d > len(db.stmts) (%d) in getStatement, statement not prepared",
			sv, len(db.stmts))
	}

	return db.stmts[sv], nil
}

func (db *DB) addStatement(sv dbStatementVal, s string) error {
	if sv < 0 {
		return fmt.Errorf("negative statement number %d in addStatement", sv)
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

	err = db.createDBLicenseLeafsTableIfNotExists()
	if err != nil {
		return err
	}

	err = db.createDBLicenseNodesTableIfNotExists()
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

	err = db.createDBHashFilesTableIfNotExists()
	if err != nil {
		return err
	}

	return nil
}

// statement type and enum
type dbStatementVal int

const (
	stmtRepoGet dbStatementVal = iota
	stmtRepoGetByCoords
	stmtRepoInsert
	stmtLicenseLeafGetAll
	stmtLicenseLeafGetByID
	stmtLicenseLeafGetByIdentifier
	stmtLicenseLeafSearchByName
	stmtLicenseLeafInsert
	stmtLicenseNodeGetAll
	stmtLicenseNodeGetByID
	stmtLicenseNodeGetAndByContents
	stmtLicenseNodeGetOrByContents
	stmtLicenseNodeGetWithByContents
	stmtLicenseNodeGetPlusByContents
	stmtLicenseNodeGetLeafByContents
	stmtLicenseNodeInsert
	stmtRepoRetrievalGet
	stmtRepoRetrievalGetLatest
	stmtRepoRetrievalInsert
	stmtRepoRetrievalUpdate
	stmtRepoFileGet
	stmtRepoFileGetForRepoRetrieval
	stmtRepoFileInsert
	stmtRepoDirGet
	stmtRepoDirGetForRepoRetrieval
	stmtRepoDirInsert
	stmtHashFileGet
	stmtHashFileGetByHashes
	stmtHashFileInsert
)

// master prepare function
func (db *DB) prepareStatements() error {
	var err error

	err = db.prepareStatementsRepos()
	if err != nil {
		return err
	}
	err = db.prepareStatementsLicenseLeafs()
	if err != nil {
		return err
	}
	err = db.prepareStatementsLicenseNodes()
	if err != nil {
		return err
	}
	err = db.prepareStatementsRepoRetrievals()
	if err != nil {
		return err
	}
	err = db.prepareStatementsRepoDirs()
	if err != nil {
		return err
	}
	err = db.prepareStatementsRepoFiles()
	if err != nil {
		return err
	}
	err = db.prepareStatementsHashFiles()
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

	err = db.addStatement(stmtRepoGetByCoords, `
		SELECT id
		FROM repos
		WHERE org_name = $1 AND repo_name = $2
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

// table licenseleafs
func (db *DB) prepareStatementsLicenseLeafs() error {
	var err error

	err = db.addStatement(stmtLicenseLeafGetAll, `
		SELECT id, identifier, name, is_spdx, type
		FROM licenseleafs
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseLeafGetByID, `
		SELECT id, identifier, name, is_spdx, type
		FROM licenseleafs
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseLeafGetByIdentifier, `
		SELECT id, identifier, name, is_spdx, type
		FROM licenseleafs
		WHERE identifier = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseLeafSearchByName, `
		SELECT id, identifier, name, is_spdx, type
		FROM licenseleafs
		WHERE name LIKE '%$1%'
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseLeafInsert, `
		INSERT INTO licenseleafs (identifier, name, is_spdx, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`)
	if err != nil {
		return err
	}

	return nil
}

// table licensenodes
func (db *DB) prepareStatementsLicenseNodes() error {
	var err error

	err = db.addStatement(stmtLicenseNodeGetAll, `
		SELECT id, type, left_id, right_id, leaf_id
		FROM licensenodes
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseNodeGetByID, `
		SELECT id, type, left_id, right_id, leaf_id
		FROM licensenodes
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseNodeGetAndByContents, `
		SELECT id, type, left_id, right_id, leaf_id
		FROM licensenodes
		WHERE type = 2 AND left_id = $1 AND right_id = $2
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseNodeGetOrByContents, `
		SELECT id, type, left_id, right_id, leaf_id
		FROM licensenodes
		WHERE type = 3 AND left_id = $1 AND right_id = $2
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseNodeGetWithByContents, `
		SELECT id, type, left_id, right_id, leaf_id
		FROM licensenodes
		WHERE type = 4 AND left_id = $1 AND right_id = $2
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseNodeGetPlusByContents, `
		SELECT id, type, left_id, right_id, leaf_id
		FROM licensenodes
		WHERE type = 5 AND left_id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseNodeGetLeafByContents, `
		SELECT id, type, left_id, right_id, leaf_id
		FROM licensenodes
		WHERE type = 1 AND leaf_id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtLicenseNodeInsert, `
		INSERT INTO licensenodes (type, left_id, right_id, leaf_id)
		VALUES ($1, $2, $3, $4)
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

// table repodirs
func (db *DB) prepareStatementsRepoDirs() error {
	var err error

	err = db.addStatement(stmtRepoDirGet, `
		SELECT id, reporetrieval_id, dir_parent_id, path
		FROM repodirs
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoDirGetForRepoRetrieval, `
		SELECT id, reporetrieval_id, dir_parent_id, path
		FROM repodirs
		WHERE reporetrieval_id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoDirInsert, `
		INSERT INTO repodirs (reporetrieval_id, dir_parent_id, path)
		VALUES ($1, $2, $3)
		RETURNING id
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
		SELECT id, reporetrieval_id, dir_parent_id, nextfile_id, prevfile_id,
		       path, hash_sha1, hash_sha256, hash_md5
		FROM repofiles
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoFileGetForRepoRetrieval, `
		SELECT id, reporetrieval_id, dir_parent_id, nextfile_id, prevfile_id,
		       path, hash_sha1, hash_sha256, hash_md5
		FROM repofiles
		WHERE reporetrieval_id = $1
		ORDER BY path
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtRepoFileInsert, `
		INSERT INTO repofiles (reporetrieval_id, dir_parent_id,
			nextfile_id, prevfile_id, path, hash_sha1, hash_sha256, hash_md5)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`)
	if err != nil {
		return err
	}

	return nil
}

// table hashfiles
func (db *DB) prepareStatementsHashFiles() error {
	var err error

	err = db.addStatement(stmtHashFileGet, `
		SELECT id, hash_sha1, hash_sha256, hash_md5
		FROM hashfiles
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtHashFileGetByHashes, `
		SELECT id, hash_sha1, hash_sha256, hash_md5
		FROM hashfiles
		WHERE hash_sha1 = $1 AND hash_sha256 = $2
	`)
	if err != nil {
		return err
	}

	err = db.addStatement(stmtHashFileInsert, `
		INSERT INTO hashfiles (hash_sha1, hash_sha256, hash_md5)
		VALUES ($1, $2, $3)
		RETURNING id
	`)
	if err != nil {
		return err
	}

	return nil
}
