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
			repo_id INTEGER NOT NULL,
			dir_parent_id INTEGER NOT NULL,
			path TEXT NOT NULL,
			UNIQUE (id, path),
			FOREIGN KEY (repo_id) REFERENCES repos (id),
			FOREIGN KEY (dir_parent_id) REFERENCES repodirs (id)
		)
	`)
	return err
}

type RepoDir struct {
	Id          int
	RepoId      int
	DirParentId int
	Path        string
}

func (db *DB) GetRepoDirById(id int) (*RepoDir, error) {
	stmt, err := db.getStatement(stmtRepoDirGet)
	if err != nil {
		return nil, err
	}

	var repodir RepoDir
	err = stmt.QueryRow(id).Scan(&repodir.Id, &repodir.RepoId,
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

//func (db *DB) BulkInsertDirs(repoId int, paths []string)
