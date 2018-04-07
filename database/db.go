// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"

	"github.com/swinslow/lfscanning/config"
)

type DB struct {
}

func PrepareDB(cfg *config.Config) (*sql.DB, error) {
	if cfg == nil || cfg.DBConnectString == "" {
		return nil, errors.New("must pass config string")
	}
	db, err := sql.Open("postgres", cfg.DBConnectString)
	if err != nil {
		return nil, err
	}

	// check that we can in fact connect
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateDBTablesIfNotExists(db *sql.DB) error {
	if db == nil {
		return errors.New("got nil *sql.DB in CreateDBTablesIfNotExists")
	}

	err := CreateDBRepoTableIfNotExists(db)
	if err != nil {
		return err
	}

	return nil
}
