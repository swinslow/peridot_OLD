// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

func (db *DB) createDBLicenseLeafTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS licenseleafs (
			id SERIAL NOT NULL PRIMARY KEY,
			identifier TEXT NOT NULL,
			name TEXT NOT NULL,
			is_spdx INTEGER NOT NULL,
			type INTEGER NOT NULL,
			plus INTEGER NOT NULL
		)
	`)
	return err
}

type LicenseLeaf struct {
	Id         int
	Identifier string
	Name       string
	IsSPDX     bool
	Type       int
	Plus       bool
}

// ===== License utility functions =====

// func (ll *LicenseLeaf) IsValidLeaf() (bool, error) {

// }
