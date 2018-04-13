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

type RepoDir struct {
	Id              int
	RepoRetrievalId int
	DirParentId     int
	Path            string
}

func (db *DB) GetRepoDirById(id int) (*RepoDir, error) {
	stmt, err := db.getStatement(stmtRepoDirGet)
	if err != nil {
		return nil, err
	}

	var repodir RepoDir
	err = stmt.QueryRow(id).Scan(&repodir.Id, &repodir.RepoRetrievalId,
		&repodir.DirParentId, &repodir.Path)
	if err != nil {
		return nil, err
	}

	return &repodir, nil
}

var exists = struct{}{}

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
	dirPaths := make([]string, len(dirs))
	i := 0
	for dirPath, _ := range dirs {
		dirPaths[i] = dirPath
		i++
	}

	sort.Strings(dirPaths)
	return dirPaths
}

func (db *DB) BulkInsertRepoDirs(repoRetrievalId int, dirs []string) error {
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
		err = insertStmt.QueryRow(repoRetrievalId, dir).Scan(&id)
		if err != nil {
			return err
		}
		repoDir = &RepoDir{Id: id, RepoRetrievalId: repoRetrievalId, Path: dir}
		repoDirs[dir] = repoDir
	}

	// go through and fix the dir_parent_id links, first in memory
	for _, repoDir = range repoDirs {
		parentPath := filepath.Dir(repoDir.Path)
		repoDir.DirParentId = repoDirs[parentPath].Id
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
		_, err = updateStmt.Exec(repoDir.DirParentId, repoDir.Id)
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
