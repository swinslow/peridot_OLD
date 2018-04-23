// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"path/filepath"
	"sort"
)

func (db *DB) createDBRepoDirsTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS repodirs (
			id SERIAL NOT NULL PRIMARY KEY,
			reporetrieval_id INTEGER NOT NULL,
			dir_parent_id INTEGER,
			path TEXT NOT NULL,
			UNIQUE (reporetrieval_id, path),
			FOREIGN KEY (reporetrieval_id) REFERENCES reporetrievals (id),
			FOREIGN KEY (dir_parent_id) REFERENCES repodirs (id)
		)
	`)
	return err
}

// RepoDir represents a directory within a single retrieval fo a source code
// repository.
type RepoDir struct {
	ID              int
	RepoRetrievalID int
	DirParentID     int
	Path            string
}

// GetRepoDirByID looks up and returns a RepoDir in the database by its ID.
// It returns nil if no RepoDir with the requested ID is found.
func (db *DB) GetRepoDirByID(id int) (*RepoDir, error) {
	stmt, err := db.getStatement(stmtRepoDirGet)
	if err != nil {
		return nil, err
	}

	var repodir RepoDir
	err = stmt.QueryRow(id).Scan(&repodir.ID, &repodir.RepoRetrievalID,
		&repodir.DirParentID, &repodir.Path)
	if err != nil {
		return nil, err
	}

	return &repodir, nil
}

// GetRepoDirsForRepoRetrieval takes the ID of a RepoRetrieval and returns a
// map of IDs to RepoDirs, for all RepoDirs from that RepoRetrieval.
func (db *DB) GetRepoDirsForRepoRetrieval(repoRetrievalID int) (map[int]*RepoDir, error) {
	stmt, err := db.getStatement(stmtRepoDirGetForRepoRetrieval)
	if err != nil {
		return nil, err
	}

	// 20 is arbitrary
	repoDirs := make(map[int]*RepoDir, 20)
	rows, err := stmt.Query(repoRetrievalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		repoDir := &RepoDir{}
		err := rows.Scan(&repoDir.ID, &repoDir.RepoRetrievalID, &repoDir.DirParentID,
			&repoDir.Path)
		if err != nil {
			return nil, err
		}
		repoDirs[repoDir.ID] = repoDir
	}

	// check at end for error
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return repoDirs, nil
}

// GetRepoDirsForRepoRetrievalByPath takes the ID of a RepoRetrieval and returns a
// map of RepoDir paths to RepoDirs, for all RepoDirs from that RepoRetrieval.
func (db *DB) GetRepoDirsForRepoRetrievalByPath(repoRetrievalID int) (map[string]*RepoDir, error) {
	stmt, err := db.getStatement(stmtRepoDirGetForRepoRetrieval)
	if err != nil {
		return nil, err
	}

	// 20 is arbitrary
	repoDirs := make(map[string]*RepoDir, 20)
	rows, err := stmt.Query(repoRetrievalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		repoDir := &RepoDir{}
		err := rows.Scan(&repoDir.ID, &repoDir.RepoRetrievalID, &repoDir.DirParentID,
			&repoDir.Path)
		if err != nil {
			return nil, err
		}
		repoDirs[repoDir.Path] = repoDir
	}

	// check at end for error
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return repoDirs, nil
}

var exists = struct{}{}

// ExtractDirsFromPaths takes a slice of paths, and returns a slice of
// directory paths that recursively includes all parent folders.
func ExtractDirsFromPaths(paths []string) []string {
	// mapping to struct{} since it's a zero-byte data structure
	dirs := make(map[string]struct{})

	// for each dir, walk through to get each path, and add it to the "set"
	for _, path := range paths {
		for i := path; i != "." && i != "/"; {
			i = filepath.Dir(i)
			dirs[i] = exists
		}
	}

	// now, convert to a list of just the keys
	// FIXME consider switching to var dirPaths []string and append()'ing
	dirPaths := make([]string, len(dirs))
	i := 0
	for dirPath := range dirs {
		dirPaths[i] = dirPath
		i++
	}

	sort.Strings(dirPaths)
	return dirPaths
}

// BulkInsertRepoDirs inserts a collection of directory paths into the
// database, wrapped in a single transaction. It takes the ID of the
// corresponding RepoRetrieval and a slice of string paths to insert.
func (db *DB) BulkInsertRepoDirs(repoRetrievalID int, dirs []string) error {
	// first, get a transaction and prepare a stmt on it
	// (we can't use stmts prepared on the main DB from within a Tx)
	tx, err := db.sqldb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertStmt, err := tx.Prepare(`
		INSERT INTO repodirs (reporetrieval_id, path)
		VALUES ($1, $2)
		RETURNING id
	`)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	// create temp dir holder, mapping paths to dirs
	repoDirs := make(map[string]*RepoDir)

	var repoDir *RepoDir

	for _, dir := range dirs {
		var id int
		err = insertStmt.QueryRow(repoRetrievalID, dir).Scan(&id)
		if err != nil {
			return err
		}
		repoDir = &RepoDir{ID: id, RepoRetrievalID: repoRetrievalID, Path: dir}
		repoDirs[dir] = repoDir
	}

	// go through and fix the dir_parent_id links, first in memory
	for _, repoDir = range repoDirs {
		parentPath := filepath.Dir(repoDir.Path)
		repoDir.DirParentID = repoDirs[parentPath].ID
	}

	// and then update in the database
	updateStmt, err := tx.Prepare(`
		UPDATE repodirs
		SET dir_parent_id = $1
		WHERE id = $2
	`)
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	for _, repoDir = range repoDirs {
		_, err = updateStmt.Exec(repoDir.DirParentID, repoDir.ID)
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
