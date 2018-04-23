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

// HashFile stores the hash values of a given file that peridot has
// seen before.
type HashFile struct {
	ID         int
	HashSHA1   string
	HashSHA256 string
	HashMD5    string
}

// GetHashFileByID looks up and returns a HashFile in the database by its ID.
// It returns nil if no HashFile with the requested ID is found.
func (db *DB) GetHashFileByID(id int) (*HashFile, error) {
	stmt, err := db.getStatement(stmtHashFileGet)
	if err != nil {
		return nil, err
	}

	var hashfile HashFile
	err = stmt.QueryRow(id).Scan(&hashfile.ID,
		&hashfile.HashSHA1, &hashfile.HashSHA256, &hashfile.HashMD5)
	if err != nil {
		return nil, err
	}

	return &hashfile, nil
}

// GetHashFileByHashes looks up and returns a HashFile in the database by both
// its SHA1 and SHA256 hashes. It returns nil if no HashFile with both the
// requested hashes is found.
func (db *DB) GetHashFileByHashes(hSHA1 string, hSHA256 string) (*HashFile, error) {
	stmt, err := db.getStatement(stmtHashFileGetByHashes)
	if err != nil {
		return nil, err
	}

	var hashfile HashFile
	// FIXME this assumes that there will only ever be one file in the catalog
	// FIXME with a given SHA1 X SHA256. consider whether that's okay.
	err = stmt.QueryRow(hSHA1, hSHA256).Scan(&hashfile.ID,
		&hashfile.HashSHA1, &hashfile.HashSHA256, &hashfile.HashMD5)
	if err != nil {
		return nil, err
	}

	return &hashfile, nil
}

// BulkInsertHashFiles inserts a collection of hash files into the database,
// wrapped in a single transaction. It takes a map from a path to a 3-element
// string array, with SHA1, SHA256 and MD5 hashes in that order.
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
		hashSHA1 := hashes[0]
		hashSHA256 := hashes[1]
		hashMD5 := hashes[2]

		_, err = insertStmt.Exec(hashSHA1, hashSHA256, hashMD5)
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
