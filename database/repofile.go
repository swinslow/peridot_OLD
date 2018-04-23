// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func (db *DB) createDBRepoFilesTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS repofiles (
			id SERIAL NOT NULL PRIMARY KEY,
			reporetrieval_id INTEGER NOT NULL,
			dir_parent_id INTEGER NOT NULL,
			nextfile_id INTEGER,
			prevfile_id INTEGER,
			path TEXT NOT NULL,
			hash_sha1 TEXT NOT NULL,
			hash_sha256 TEXT NOT NULL,
			hash_md5 TEXT NOT NULL,
			FOREIGN KEY (reporetrieval_id) REFERENCES reporetrievals (id),
			FOREIGN KEY (dir_parent_id) REFERENCES repodirs (id),
			FOREIGN KEY (nextfile_id) REFERENCES repofiles (id),
			FOREIGN KEY (prevfile_id) REFERENCES repofiles (id)
		)
	`)
	return err
}

// RepoFile represents a file within a single retrieval fo a source code
// repository.
type RepoFile struct {
	ID              int
	RepoRetrievalID int
	DirParentID     int
	NextFileID      int
	PrevFileID      int
	Path            string
	HashSHA1        string
	HashSHA256      string
	HashMD5         string
}

// GetRepoFileByID looks up and returns a RepoFile in the database by its ID.
// It returns nil if no RepoFile with the requested ID is found.
func (db *DB) GetRepoFileByID(id int) (*RepoFile, error) {
	stmt, err := db.getStatement(stmtRepoFileGet)
	if err != nil {
		return nil, err
	}

	var repofile RepoFile
	err = stmt.QueryRow(id).Scan(&repofile.ID, &repofile.RepoRetrievalID,
		&repofile.DirParentID, &repofile.NextFileID, &repofile.PrevFileID,
		&repofile.Path,
		&repofile.HashSHA1, &repofile.HashSHA256, &repofile.HashMD5)
	if err != nil {
		return nil, err
	}

	return &repofile, nil
}

// GetRepoFilesForRepoRetrieval takes the ID of a RepoRetrieval and returns a
// map of IDs to RepoFiles, for all RepoFiles from that RepoRetrieval.
func (db *DB) GetRepoFilesForRepoRetrieval(repoRetrievalID int) (map[int]*RepoFile, error) {
	stmt, err := db.getStatement(stmtRepoFileGetForRepoRetrieval)
	if err != nil {
		return nil, err
	}

	// 20 is arbitrary
	repoFiles := make(map[int]*RepoFile, 20)
	rows, err := stmt.Query(repoRetrievalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		repoFile := &RepoFile{}
		err := rows.Scan(&repoFile.ID, &repoFile.RepoRetrievalID,
			&repoFile.DirParentID, &repoFile.NextFileID, &repoFile.PrevFileID,
			&repoFile.Path,
			&repoFile.HashSHA1, &repoFile.HashSHA256, &repoFile.HashMD5)
		if err != nil {
			return nil, err
		}
		repoFiles[repoFile.ID] = repoFile
	}

	// check at end for error
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return repoFiles, nil
}

// BulkInsertRepoFiles inserts a collection of files into the database,
// wrapped in a single transaction. It takes a map from a path to a 3-element
// string array, with SHA1, SHA256 and MD5 hashes in that order.
func (db *DB) BulkInsertRepoFiles(repoRetrievalID int, pathsToHashes map[string][3]string) error {
	// first, get the corresponding repo directories from the database
	repoDirs, err := db.GetRepoDirsForRepoRetrievalByPath(repoRetrievalID)
	if err != nil {
		return err
	}

	// now, get a transaction and prepare a stmt on it
	// (we can't use stmts prepared on the main DB from within a Tx)
	tx, err := db.sqldb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertStmt, err := tx.Prepare(`
		INSERT INTO repofiles (reporetrieval_id, dir_parent_id, path, hash_sha1, hash_sha256, hash_md5)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	// create temp files holder, mapping paths to RepoFiles
	repoFiles := make(map[string]*RepoFile)
	// and create temp list of all file paths
	var repoFilePaths []string
	var repoFile *RepoFile

	for path, hashes := range pathsToHashes {
		hashSHA1 := hashes[0]
		hashSHA256 := hashes[1]
		hashMD5 := hashes[2]

		dirParentPath := filepath.Dir(path)
		dirParent, ok := repoDirs[dirParentPath]
		if !ok {
			return fmt.Errorf("Couldn't find parent directory object for file %s", path)
		}

		var id int
		err = insertStmt.QueryRow(repoRetrievalID, dirParent.ID, path,
			hashSHA1, hashSHA256, hashMD5).Scan(&id)
		if err != nil {
			return err
		}
		repoFile = &RepoFile{ID: id, RepoRetrievalID: repoRetrievalID,
			DirParentID: dirParent.ID, Path: path,
			HashSHA1: hashSHA1, HashSHA256: hashSHA256, HashMD5: hashMD5}
		repoFiles[path] = repoFile
		repoFilePaths = append(repoFilePaths, path)
	}

	err = fillInNextAndPrevRepoFiles(repoFiles, repoFilePaths)
	if err != nil {
		return err
	}

	// finally, we can update to save the prev and next file IDs
	updateStmt, err := tx.Prepare(`
		UPDATE repofiles
		SET nextfile_id = $1, prevfile_id = $2
		WHERE id = $3
	`)
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	for _, repoFile = range repoFiles {
		_, err = updateStmt.Exec(repoFile.NextFileID, repoFile.PrevFileID, repoFile.ID)
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

func fillInNextAndPrevRepoFiles(repoFiles map[string]*RepoFile, repoFilePaths []string) error {
	// sort the paths so we can do next / prev linking
	// note, we want to sort case-insensitive, which sort.Strings() doesn't do for us
	sort.Slice(repoFilePaths, func(i, j int) bool {
		return strings.ToLower(repoFilePaths[i]) < strings.ToLower(repoFilePaths[j])
	})

	// go through paths in sorted order, looking up IDs and filling them in
	var prevRepoFile, curRepoFile, nextRepoFile *RepoFile
	var prevPath, nextPath string
	var prevRepoFileID, nextRepoFileID int
	var ok bool
	lenPaths := len(repoFilePaths)
	for i, curPath := range repoFilePaths {
		// determine path strings
		if i == 0 {
			prevPath = ""
			prevRepoFile = nil
			prevRepoFileID = 0
		} else {
			prevPath = repoFilePaths[i-1]
			prevRepoFile, ok = repoFiles[prevPath]
			if !ok {
				return fmt.Errorf("No file found when setting previous ID for %s (prevPath = %s)",
					curPath, prevPath)
			}
			prevRepoFileID = prevRepoFile.ID
		}

		if i == (lenPaths - 1) {
			nextPath = ""
			nextRepoFile = nil
			nextRepoFileID = 0
		} else {
			nextPath = repoFilePaths[i+1]
			nextRepoFile, ok = repoFiles[nextPath]
			if !ok {
				return fmt.Errorf("No file found when setting next ID for %s (nextPath = %s)",
					curPath, nextPath)
			}
			nextRepoFileID = nextRepoFile.ID
		}

		curRepoFile, ok = repoFiles[curPath]
		if !ok {
			return fmt.Errorf("No file found when setting IDs for %s", curPath)
		}

		// and, finally, fill in IDs
		// files at beginning / end will point to themselves
		if prevRepoFileID == 0 {
			curRepoFile.PrevFileID = curRepoFile.ID
		} else {
			curRepoFile.PrevFileID = prevRepoFileID
		}
		if nextRepoFileID == 0 {
			curRepoFile.NextFileID = curRepoFile.ID
		} else {
			curRepoFile.NextFileID = nextRepoFileID
		}
	}

	return nil
}
