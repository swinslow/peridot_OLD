// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"fmt"

	"github.com/swinslow/peridot/licenses"
)

func (db *DB) createDBLicenseLeafsTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS licenseleafs (
			id SERIAL NOT NULL PRIMARY KEY,
			identifier TEXT NOT NULL,
			name TEXT NOT NULL,
			is_spdx INTEGER NOT NULL,
			type INTEGER NOT NULL
		)
	`)
	return err
}

// LicenseLeaf represents a single, simple SPDX-formatted license name.
// It can be used in more complex / compound license expressions by being
// in a LicenseNode.
// Under current SPDX naming rules, it should either be taken from the
// SPDX License List (IsSPDX=True) or else its name should begin with
// "LicenseRef-" (IsSPDX=False).
type LicenseLeaf struct {
	ID         int
	Identifier string
	Name       string
	IsSPDX     bool
	Type       int
}

// GetLicenseLeafByID looks up and returns a LicenseLeaf in the database by
// its numerical ID (NOT its Identifier / Name). It returns nil if no
// LicenseLeaf with the requested ID is found.
func (db *DB) GetLicenseLeafByID(id int) (*LicenseLeaf, error) {
	stmt, err := db.getStatement(stmtLicenseLeafGetByID)
	if err != nil {
		return nil, err
	}

	var licenseLeaf LicenseLeaf
	err = stmt.QueryRow(id).Scan(&licenseLeaf.ID,
		&licenseLeaf.Identifier, &licenseLeaf.Name,
		&licenseLeaf.IsSPDX, &licenseLeaf.Type)
	if err != nil {
		return nil, err
	}

	return &licenseLeaf, nil
}

// GetLicenseLeafByIdentifier looks up and returns a LicenseLeaf in the
// database by its SPDX identifier (NOT its numerical database ID). It returns
// nil if no LicenseLeaf with the requested identifier is found.
func (db *DB) GetLicenseLeafByIdentifier(identifier string) (*LicenseLeaf, error) {
	stmt, err := db.getStatement(stmtLicenseLeafGetByIdentifier)
	if err != nil {
		return nil, err
	}

	var licenseLeaf LicenseLeaf
	err = stmt.QueryRow(identifier).Scan(&licenseLeaf.ID,
		&licenseLeaf.Identifier, &licenseLeaf.Name,
		&licenseLeaf.IsSPDX, &licenseLeaf.Type)
	if err != nil {
		return nil, err
	}

	return &licenseLeaf, nil
}

// InsertLicenseLeaf takes data for a license leaf, creates a new LicenseLeaf
// struct, adds it to the database, and returns the new struct with its ID
// from the DB.
func (db *DB) InsertLicenseLeaf(identifier string, name string,
	isSPDX bool, licType int) (*LicenseLeaf, error) {
	stmt, err := db.getStatement(stmtLicenseLeafInsert)
	if err != nil {
		return nil, err
	}

	var id int
	var isSPDXInt int
	if isSPDX {
		isSPDXInt = 1
	} else {
		isSPDXInt = 0
	}
	err = stmt.QueryRow(identifier, name, isSPDXInt, licType).Scan(&id)
	if err != nil {
		return nil, err
	}

	ll := &LicenseLeaf{ID: id, Identifier: identifier, Name: name,
		IsSPDX: isSPDX, Type: licType}
	return ll, nil
}

// InsertFromLicenseList takes a path to the SPDX license list data JSON
// directory. It reads the license list and, for each one it finds, checks
// to see whether there is already an entry for its identifier in the
// database. If there isn't, it creates a LicenseLeaf for it in the database.
func (db *DB) InsertFromLicenseList(spdxLLJSONLocation string) error {
	ll, err := licenses.LoadFromJSON(spdxLLJSONLocation)
	if err != nil {
		return fmt.Errorf("couldn't insert new licenses from license list: %v", err)
	}

	lics := ll.Licenses
	fmt.Printf("Got lics: %d\n", len(lics))
	for _, lic := range lics {
		// check if already present
		_, err := db.GetLicenseLeafByIdentifier(lic.Identifier)
		if err == nil {
			// found it; don't insert this one
			continue
		}
		// it wasn't found, so try adding a new one
		// FIXME should likely check whether err is the "row not found" type,
		// FIXME and not some other error.
		_, err = db.InsertLicenseLeaf(lic.Identifier, lic.Name, true, 1)
		if err != nil {
			return fmt.Errorf("couldn't insert new license leaf for %s: %v",
				lic.Identifier, err)
		}
	}

	return nil
}

// func (ll *LicenseLeaf) IsValidLeaf() (bool, error) {

// }
