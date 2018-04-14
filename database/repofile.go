// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
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
			FOREIGN KEY (reporetrieval_id) REFERENCES reporetrievals (id),
			FOREIGN KEY (dir_parent_id) REFERENCES repodirs (id),
			FOREIGN KEY (nextfile_id) REFERENCES repofiles (id),
			FOREIGN KEY (prevfile_id) REFERENCES repofiles (id)
		)
	`)
	return err
}

type RepoFile struct {
	Id              int
	RepoRetrievalId int
	DirParentId     int
	NextFileId      int
	PrevFileId      int
	Path            string
	Hash_SHA1       string
}

func (db *DB) GetRepoFileById(id int) (*RepoFile, error) {
	stmt, err := db.getStatement(stmtRepoFileGet)
	if err != nil {
		return nil, err
	}

	var repofile RepoFile
	err = stmt.QueryRow(id).Scan(&repofile.Id, &repofile.RepoRetrievalId,
		&repofile.DirParentId, &repofile.NextFileId, &repofile.PrevFileId,
		&repofile.Path, &repofile.Hash_SHA1)
	if err != nil {
		return nil, err
	}

	return &repofile, nil
}

func (db *DB) GetRepoFilesForRepoRetrieval(repoRetrievalId int) (map[int]*RepoFile, error) {
	stmt, err := db.getStatement(stmtRepoFileGetForRepoRetrieval)
	if err != nil {
		return nil, err
	}

	// 20 is arbitrary
	repoFiles := make(map[int]*RepoFile, 20)
	rows, err := stmt.Query(repoRetrievalId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		repoFile := &RepoFile{}
		err := rows.Scan(&repoFile.Id, &repoFile.RepoRetrievalId,
			&repoFile.DirParentId, &repoFile.NextFileId, &repoFile.PrevFileId,
			&repoFile.Path, &repoFile.Hash_SHA1)
		if err != nil {
			return nil, err
		}
		repoFiles[repoFile.Id] = repoFile
	}

	// check at end for error
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return repoFiles, nil
}

func (db *DB) BulkInsertRepoFiles(repoRetrievalId int, pathsToHashes map[string]string) error {
	// first, get the corresponding repo directories from the database
	repoDirs, err := db.GetRepoDirsForRepoRetrievalByPath(repoRetrievalId)
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
		INSERT INTO repofiles (reporetrieval_id, dir_parent_id, path, hash_sha1)
		VALUES ($1, $2, $3, $4)
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

	for path, hash := range pathsToHashes {
		dirParentPath := filepath.Dir(path)
		dirParent, ok := repoDirs[dirParentPath]
		if !ok {
			return errors.New("Couldn't find parent directory object for file " + path)
		}

		var id int
		err = insertStmt.QueryRow(repoRetrievalId, dirParent.Id, path, hash).Scan(&id)
		if err != nil {
			return err
		}
		repoFile = &RepoFile{Id: id, RepoRetrievalId: repoRetrievalId,
			DirParentId: dirParent.Id, Path: path, Hash_SHA1: hash}
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
		_, err = updateStmt.Exec(repoFile.NextFileId, repoFile.PrevFileId, repoFile.Id)
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
	var prevRepoFileId, nextRepoFileId int
	var ok bool
	lenPaths := len(repoFilePaths)
	for i, curPath := range repoFilePaths {
		// determine path strings
		if i == 0 {
			prevPath = ""
			prevRepoFile = nil
			prevRepoFileId = 0
		} else {
			prevPath = repoFilePaths[i-1]
			prevRepoFile, ok = repoFiles[prevPath]
			if !ok {
				return errors.New("No file found when setting previous ID for " +
					curPath + " (prevPath = " + prevPath + ")")
			}
			prevRepoFileId = prevRepoFile.Id
		}

		if i == (lenPaths - 1) {
			nextPath = ""
			nextRepoFile = nil
			nextRepoFileId = 0
		} else {
			nextPath = repoFilePaths[i+1]
			nextRepoFile, ok = repoFiles[nextPath]
			if !ok {
				return errors.New("No file found when setting next ID for " +
					curPath + " (nextPath = " + nextPath + ")")
			}
			nextRepoFileId = nextRepoFile.Id
		}

		curRepoFile, ok = repoFiles[curPath]
		if !ok {
			return errors.New("No file found when setting IDs for " + curPath)
		}

		// and, finally, fill in IDs
		// files at beginning / end will point to themselves
		if prevRepoFileId == 0 {
			curRepoFile.PrevFileId = curRepoFile.Id
		} else {
			curRepoFile.PrevFileId = prevRepoFileId
		}
		if nextRepoFileId == 0 {
			curRepoFile.NextFileId = curRepoFile.Id
		} else {
			curRepoFile.NextFileId = nextRepoFileId
		}
	}

	return nil
}
