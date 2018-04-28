// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"fmt"

	"github.com/swinslow/peridot/licenses"
)

func (db *DB) createDBLicenseNodesTableIfNotExists() error {
	_, err := db.sqldb.Exec(`
		CREATE TABLE IF NOT EXISTS licensenodes (
			id SERIAL NOT NULL PRIMARY KEY,
			type INTEGER NOT NULL,
			left_id INTEGER NOT NULL,
			right_id INTEGER NOT NULL,
			leaf_id INTEGER NOT NULL,
			FOREIGN KEY (left_id) REFERENCES licensenodes (id),
			FOREIGN KEY (right_id) REFERENCES licensenodes (id),
			FOREIGN KEY (leaf_id) REFERENCES licenseleafs (id)
		)
	`)
	if err != nil {
		return err
	}

	// check for zero row and add if not present
	var tmp int
	err = db.sqldb.QueryRow(`SELECT id FROM licensenodes WHERE id=0`).Scan(&tmp)
	switch {
	case err == sql.ErrNoRows:
		_, err := db.sqldb.Exec(`INSERT INTO licensenodes (id, type, left_id, right_id, leaf_id) VALUES (0, 0, 0, 0, 0)`)
		return err
	case err != nil:
		return err
	default:
		return nil
	}
}

const (
	lnodeErr  = 0
	lnodeLeaf = 1
	lnodeAnd  = 2
	lnodeOr   = 3
	lnodeWith = 4
	lnodePlus = 5
)

// LicenseNode represents a component of a simple or complex SPDX license
// expression.
type LicenseNode struct {
	ID      int
	Type    int
	LeftID  int
	RightID int
	LeafID  int
}

// GetIntForNodeType converts a license NodeType to the corresponding
// integer used in the database. It returns an error for an invalid type.
func (db *DB) GetIntForNodeType(nodeType licenses.ParsedLicenseNodeType) (int, error) {
	switch nodeType {
	case licenses.NodeIdentifier:
		return lnodeLeaf, nil
	case licenses.NodeAnd:
		return lnodeAnd, nil
	case licenses.NodeOr:
		return lnodeOr, nil
	case licenses.NodeWith:
		return lnodeWith, nil
	case licenses.NodePlus:
		return lnodePlus, nil
	case licenses.NodeError:
		return lnodeErr, fmt.Errorf("error node type: %v", nodeType)
	default:
		return lnodeErr, fmt.Errorf("unknown node type: %v", nodeType)
	}
}

// GetNodeTypeForInt converts an integer used for LicenseNode.Type in
// the database into the corresponding license NodeType. It returns an
// error for an invalid type.
func (db *DB) GetNodeTypeForInt(nodeType int) (licenses.ParsedLicenseNodeType, error) {
	switch nodeType {
	case lnodeLeaf:
		return licenses.NodeIdentifier, nil
	case lnodeAnd:
		return licenses.NodeAnd, nil
	case lnodeOr:
		return licenses.NodeOr, nil
	case lnodeWith:
		return licenses.NodeWith, nil
	case lnodePlus:
		return licenses.NodePlus, nil
	default:
		return licenses.NodeError, fmt.Errorf("unknown node type: %d", nodeType)
	}
}

// GetLicenseNodeAll looks up and returns a map of IDs to LicenseNodes for
// all LicenseNodes in the database.
func (db *DB) GetLicenseNodeAll() (map[int]*LicenseNode, error) {
	stmt, err := db.getStatement(stmtLicenseNodeGetAll)
	if err != nil {
		return nil, err
	}

	lns := make(map[int]*LicenseNode)
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		ln := &LicenseNode{}
		err = rows.Scan(&ln.ID, &ln.Type, &ln.LeftID, &ln.RightID, &ln.LeafID)
		if err != nil {
			return nil, err
		}
		lns[ln.ID] = ln
	}

	// check at end for error
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return lns, nil
}

// GetLicenseNodeByID looks up and returns a LicenseNode in the database by
// its numerical ID. It returns nil if no LicenseNode with the requested ID
// is found.
func (db *DB) GetLicenseNodeByID(id int) (*LicenseNode, error) {
	stmt, err := db.getStatement(stmtLicenseNodeGetByID)
	if err != nil {
		return nil, err
	}

	var ln LicenseNode
	err = stmt.QueryRow(id).Scan(&ln.ID, &ln.Type, &ln.LeftID, &ln.RightID, &ln.LeafID)
	if err != nil {
		return nil, err
	}

	return &ln, nil
}

// GetLicenseNodeByContents looks up and returns a complete LicenseNode from
// the database with the given node type and children, if one exists in the
// database. It returns nil (with nil error) if it is not found.
func (db *DB) GetLicenseNodeByContents(nodeType int, leftID int, rightID int, leafID int) (*LicenseNode, error) {
	switch nodeType {
	case lnodeLeaf:
		return db.getLicenseNodeByContentsLeaf(leafID)
	case lnodeAnd:
		return db.getLicenseNodeByContentsAnd(leftID, rightID)
	case lnodeOr:
		return db.getLicenseNodeByContentsOr(leftID, rightID)
	case lnodeWith:
		return db.getLicenseNodeByContentsWith(leftID, rightID)
	case lnodePlus:
		return db.getLicenseNodeByContentsPlus(leftID)
	default:
		return nil, fmt.Errorf("unknown license node type in GetLicenseNodeByContents: %d", nodeType)
	}
}

func (db *DB) getLicenseNodeByContentsLeaf(leafID int) (*LicenseNode, error) {
	stmt, err := db.getStatement(stmtLicenseNodeGetLeafByContents)
	if err != nil {
		return nil, err
	}

	var ln LicenseNode
	err = stmt.QueryRow(leafID).Scan(&ln.ID, &ln.Type, &ln.LeftID, &ln.RightID, &ln.LeafID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &ln, nil
	}
}

func (db *DB) getLicenseNodeByContentsAnd(leftID int, rightID int) (*LicenseNode, error) {
	stmt, err := db.getStatement(stmtLicenseNodeGetAndByContents)
	if err != nil {
		return nil, err
	}

	var ln LicenseNode
	err = stmt.QueryRow(leftID, rightID).Scan(&ln.ID, &ln.Type, &ln.LeftID, &ln.RightID, &ln.LeafID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &ln, nil
	}
}

func (db *DB) getLicenseNodeByContentsOr(leftID int, rightID int) (*LicenseNode, error) {
	stmt, err := db.getStatement(stmtLicenseNodeGetOrByContents)
	if err != nil {
		return nil, err
	}

	var ln LicenseNode
	err = stmt.QueryRow(leftID, rightID).Scan(&ln.ID, &ln.Type, &ln.LeftID, &ln.RightID, &ln.LeafID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &ln, nil
	}
}

func (db *DB) getLicenseNodeByContentsWith(leftID int, rightID int) (*LicenseNode, error) {
	stmt, err := db.getStatement(stmtLicenseNodeGetWithByContents)
	if err != nil {
		return nil, err
	}

	var ln LicenseNode
	err = stmt.QueryRow(leftID, rightID).Scan(&ln.ID, &ln.Type, &ln.LeftID, &ln.RightID, &ln.LeafID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &ln, nil
	}
}

func (db *DB) getLicenseNodeByContentsPlus(leftID int) (*LicenseNode, error) {
	stmt, err := db.getStatement(stmtLicenseNodeGetPlusByContents)
	if err != nil {
		return nil, err
	}

	var ln LicenseNode
	err = stmt.QueryRow(leftID).Scan(&ln.ID, &ln.Type, &ln.LeftID, &ln.RightID, &ln.LeafID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &ln, nil
	}
}

// InsertLicenseNode takes data for a license node, creates a new LicenseNode
// struct, adds it to the database, and returns the new struct with its ID
// from the DB.
func (db *DB) InsertLicenseNode(nodeType int, leftID int, rightID int, leafID int) (*LicenseNode, error) {
	stmt, err := db.getStatement(stmtLicenseNodeInsert)
	if err != nil {
		return nil, err
	}

	var id int
	err = stmt.QueryRow(nodeType, leftID, rightID, leafID).Scan(&id)
	if err != nil {
		return nil, err
	}

	ln := &LicenseNode{ID: id, Type: nodeType, LeftID: leftID, RightID: rightID, LeafID: leafID}
	return ln, nil
}

// CheckAndInsertLicenseNodes takes a pointer to a ParsedLicenseNode, and
// for it and each of its children, it recursively (1) checks whether a
// LicenseNode for that (sub)expression already exists, and (2) inserts one
// if there isn't one already. It returns the applicable LicenseNode pointer.
// Note that it will not insert LicenseLeafs, and assumes that all are already
// present; if any aren't, it will return an error.
func (db *DB) CheckAndInsertLicenseNodes(pln *licenses.ParsedLicenseNode) (*LicenseNode, error) {
	if pln == nil {
		return nil, nil
	}
	switch pln.NodeType {
	case licenses.NodeIdentifier:
		return db.checkAndInsertNodeIdentifier(pln)
	case licenses.NodeAnd:
		return db.checkAndInsertNodeAnd(pln)
	case licenses.NodeOr:
		return db.checkAndInsertNodeOr(pln)
	case licenses.NodeWith:
		return db.checkAndInsertNodeWith(pln)
	case licenses.NodePlus:
		return db.checkAndInsertNodePlus(pln)
	default:
		return nil, fmt.Errorf("unknown node type when checking and inserting license nodes: %v", pln.NodeType)
	}
}

func (db *DB) checkAndInsertNodeIdentifier(pln *licenses.ParsedLicenseNode) (*LicenseNode, error) {
	if pln == nil {
		return nil, nil
	}

	// get the leaf so we can point to it
	leaf, err := db.GetLicenseLeafByIdentifier(pln.Expression)
	if err != nil {
		return nil, fmt.Errorf("license leaf not found when adding nodes: %s", pln.Expression)
	}

	// see whether we've already got a node for this expression
	ln, err := db.getLicenseNodeByContentsLeaf(leaf.ID)
	if err != nil {
		return nil, fmt.Errorf("error when checking for existing leaf node: %v", err)
	}

	if ln != nil {
		// a node already exists, so return it
		return ln, nil
	}

	// a node doesn't exist, so create and return it
	return db.InsertLicenseNode(lnodeLeaf, 0, 0, leaf.ID)
}

func (db *DB) checkAndInsertNodeAnd(pln *licenses.ParsedLicenseNode) (*LicenseNode, error) {
	if pln == nil {
		return nil, nil
	}

	// get the left and right nodes first, so we can point to them
	leftChild, err := db.CheckAndInsertLicenseNodes(pln.LeftChild)
	if err != nil {
		return nil, fmt.Errorf("left child not found when adding nodes: %v", pln.LeftChild)
	}
	rightChild, err := db.CheckAndInsertLicenseNodes(pln.RightChild)
	if err != nil {
		return nil, fmt.Errorf("right child not found when adding nodes: %v", pln.RightChild)
	}

	// see whether we've already got a node for this expression
	ln, err := db.getLicenseNodeByContentsAnd(leftChild.ID, rightChild.ID)
	if err != nil {
		return nil, fmt.Errorf("error when checking for existing AND node: %v", err)
	}

	if ln != nil {
		// a node already exists, so return it
		return ln, nil
	}

	// a node doesn't exist, so create and return it
	return db.InsertLicenseNode(lnodeAnd, leftChild.ID, rightChild.ID, 0)
}

func (db *DB) checkAndInsertNodeOr(pln *licenses.ParsedLicenseNode) (*LicenseNode, error) {
	if pln == nil {
		return nil, nil
	}

	// get the left and right nodes first, so we can point to them
	leftChild, err := db.CheckAndInsertLicenseNodes(pln.LeftChild)
	if err != nil {
		return nil, fmt.Errorf("left child not found when adding nodes: %v", pln.LeftChild)
	}
	rightChild, err := db.CheckAndInsertLicenseNodes(pln.RightChild)
	if err != nil {
		return nil, fmt.Errorf("right child not found when adding nodes: %v", pln.RightChild)
	}

	// see whether we've already got a node for this expression
	ln, err := db.getLicenseNodeByContentsOr(leftChild.ID, rightChild.ID)
	if err != nil {
		return nil, fmt.Errorf("error when checking for existing OR node: %v", err)
	}

	if ln != nil {
		// a node already exists, so return it
		return ln, nil
	}

	// a node doesn't exist, so create and return it
	return db.InsertLicenseNode(lnodeOr, leftChild.ID, rightChild.ID, 0)
}

func (db *DB) checkAndInsertNodeWith(pln *licenses.ParsedLicenseNode) (*LicenseNode, error) {
	if pln == nil {
		return nil, nil
	}

	// get the left and right nodes first, so we can point to them
	leftChild, err := db.CheckAndInsertLicenseNodes(pln.LeftChild)
	if err != nil {
		return nil, fmt.Errorf("left child not found when adding nodes: %v", pln.LeftChild)
	}
	rightChild, err := db.CheckAndInsertLicenseNodes(pln.RightChild)
	if err != nil {
		return nil, fmt.Errorf("right child not found when adding nodes: %v", pln.RightChild)
	}

	// see whether we've already got a node for this expression
	ln, err := db.getLicenseNodeByContentsWith(leftChild.ID, rightChild.ID)
	if err != nil {
		return nil, fmt.Errorf("error when checking for existing WITH node: %v", err)
	}

	if ln != nil {
		// a node already exists, so return it
		return ln, nil
	}

	// a node doesn't exist, so create and return it
	return db.InsertLicenseNode(lnodeWith, leftChild.ID, rightChild.ID, 0)
}

func (db *DB) checkAndInsertNodePlus(pln *licenses.ParsedLicenseNode) (*LicenseNode, error) {
	if pln == nil {
		return nil, nil
	}

	// get the left node first, so we can point to it
	leftChild, err := db.CheckAndInsertLicenseNodes(pln.LeftChild)
	if err != nil {
		return nil, fmt.Errorf("left child not found when adding nodes: %v", pln.LeftChild)
	}

	// see whether we've already got a node for this expression
	ln, err := db.getLicenseNodeByContentsPlus(leftChild.ID)
	if err != nil {
		return nil, fmt.Errorf("error when checking for existing PLUS node: %v", err)
	}

	if ln != nil {
		// a node already exists, so return it
		return ln, nil
	}

	// a node doesn't exist, so create and return it
	return db.InsertLicenseNode(lnodePlus, leftChild.ID, 0, 0)
}
