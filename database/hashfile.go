// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

func (db *DB) createDBHashFilesTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS hashfiles (
			id SERIAL NOT NULL PRIMARY KEY,
			hash_sha1 TEXT NOT NULL,
			hash_sha256 TEXT NOT NULL,
			hash_md5 TEXT NOT NULL
		)
	`)
	return err
}

type HashFile struct {
	Id              int
	Hash_SHA1       string
	Hash_SHA256     string
	Hash_MD5        string
}

func (db *DB) GetHashFileById(id int) (*HashFile, error) {
	stmt, err := db.getStatement(stmtHashFileGet)
	if err != nil {
		return nil, err
	}

	var hashfile HashFile
	err = stmt.QueryRow(id).Scan(&hashfile.Id,
		&hashfile.Hash_SHA1, &hashfile.Hash_SHA256, &hashfile.Hash_MD5)
	if err != nil {
		return nil, err
	}

	return &hashfile, nil
}

func (db *DB) GetHashFileByHashes(h_sha1 string, h_sha256 string) (*HashFile, error) {
	stmt, err := db.getStatement(stmtHashFileGetByHashes)
	if err != nil {
		return nil, err
	}

	var hashfile HashFile
	// FIXME this assumes that there will only ever be one file in the catalog
	// FIXME with a given SHA1 X SHA256. consider whether that's okay.
	err = stmt.QueryRow(h_sha1, h_sha256).Scan(&hashfile.Id,
		&hashfile.Hash_SHA1, &hashfile.Hash_SHA256, &hashfile.Hash_MD5)
	if err != nil {
		return nil, err
	}

	return &hashfile, nil
}

// call with map of paths to array with hashes: SHA1, SHA256, MD5
func (db *DB) BulkInsertHashFiles(pathsToHashes map[string][3]string) error {
	// we're ignoring the paths, just getting the hashes

	// get a transaction and prepare a stmt on it
	// (we can't use stmts prepared on the main DB from within a Tx)
	tx, err := db.sqldb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertStmt, err := tx.Prepare(`
		INSERT INTO hashfiles (hash_sha1, hash_sha256, hash_md5)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	for _, hashes := range pathsToHashes {
		hash_sha1 := hashes[0]
		hash_sha256 := hashes[1]
		hash_md5 := hashes[2]

		_, err = insertStmt.Exec(hash_sha1, hash_sha256, hash_md5)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}