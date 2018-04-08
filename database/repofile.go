// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

func (db *DB) createDBRepoFilesTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS repofiles (
			id SERIAL NOT NULL PRIMARY KEY,
			repo_id INTEGER NOT NULL,
			dir_parent_id INTEGER NOT NULL,
			nextfile_id INTEGER,
			prevfile_id INTEGER,
			path TEXT NOT NULL,
			hash_sha1 TEXT NOT NULL,
			FOREIGN KEY (repo_id) REFERENCES repos (id),
			FOREIGN KEY (dir_parent_id) REFERENCES repodirs (id),
			FOREIGN KEY (nextfile_id) REFERENCES repofiles (id),
			FOREIGN KEY (prevfile_id) REFERENCES repofiles (id)
		)
	`)
	return err
}

type RepoFile struct {
	Id          int
	RepoId      int
	DirParentId int
	NextFileId  int
	PrevFileId  int
	Path        string
	Hash_SHA1   string
}

func (db *DB) GetRepoFileById(id int) (*RepoFile, error) {
	stmt, err := db.getStatement(stmtRepoFileGet)
	if err != nil {
		return nil, err
	}

	var repofile RepoFile
	err = stmt.QueryRow(id).Scan(&repofile.Id, &repofile.RepoId,
		&repofile.DirParentId, &repofile.NextFileId, &repofile.PrevFileId,
		&repofile.Path, &repofile.Hash_SHA1)
	if err != nil {
		return nil, err
	}

	return &repofile, nil
}
