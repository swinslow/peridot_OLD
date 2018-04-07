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
	sqldb *sql.DB
}

func (db *DB) PrepareDB(cfg *config.Config) error {
	if db == nil {
		return errors.New("must pass non-nil DB")
	}
	if cfg == nil || cfg.DBConnectString == "" {
		return errors.New("must pass config string")
	}
	sqldb, err := sql.Open("postgres", cfg.DBConnectString)
	if err != nil {
		return err
	}

	// check that we can in fact connect
	err = sqldb.Ping()
	if err != nil {
		return err
	}

	// we're good; set as database connect
	db.sqldb = sqldb
	return nil
}

func (db *DB) CreateDBTablesIfNotExists() error {
	if db == nil {
		return errors.New("got nil *DB in CreateDBTablesIfNotExists")
	}
	if db.sqldb == nil {
		return errors.New("got nil *sql.DB in CreateDBTablesIfNotExists")
	}

	err := db.CreateDBRepoTableIfNotExists()
	if err != nil {
		return err
	}

	return nil
}
