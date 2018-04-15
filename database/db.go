// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"

	"github.com/swinslow/peridot/config"
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
		return fmt.Errorf("must pass non-nil DB to PrepareDB")
	}
	if cfg == nil || cfg.DBConnectString == "" {
		return fmt.Errorf("must pass config string to PrepareDB")
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

func (db *DB) ResetDB() error {
	// we control the contents of tables, it isn't dependent on user input,
	// so building the statements this way shouldn't be a problem
	for _, tablename := range tables {
		_, err := db.sqldb.Exec("DROP TABLE " + tablename)
		if err != nil {
			return err
		}
	}

	return nil
}