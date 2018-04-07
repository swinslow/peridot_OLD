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
	stmts []*sql.Stmt
}

func InitDB() *DB {
	var db DB
	db.stmts = make([]*sql.Stmt, 1)
	return &db
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

	// create tables if they don't already exist
	err = db.createDBTablesIfNotExists()
	if err != nil {
		return err
	}

	// and prepare statements (must do this after ensuring tables exist)
	err = db.prepareStatements()
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) createDBTablesIfNotExists() error {
	err := db.createDBRepoTableIfNotExists()
	if err != nil {
		return err
	}

	return nil
}
