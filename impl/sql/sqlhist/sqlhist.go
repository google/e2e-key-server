// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package sqlhist implements a temporal sparse merkle tree using SQL.
// Each epoch has its own sparse tree. By default, each new epoch is equal to
// the contents of the previous epoch.
package sqlhist

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/key-transparency/core/transaction"
	"github.com/google/key-transparency/core/tree"
	"github.com/google/key-transparency/core/tree/sparse"

	"golang.org/x/net/context"
)

var (
	createStmt = []string{
		`
	CREATE TABLE IF NOT EXISTS Maps (
		MapId   VARCHAR(32) NOT NULL,
		PRIMARY KEY(MapID)
	);`,
		`
	CREATE TABLE IF NOT EXISTS Leaves (
		MapId   VARCHAR(32)   NOT NULL,
		LeafId  VARBINARY(32) NOT NULL,
		Version INTEGER       NOT NULL,
		Data    BLOB          NOT NULL,
		PRIMARY KEY(MapID, LeafId, Version),
		FOREIGN KEY(MapId) REFERENCES Maps(MapId) ON DELETE CASCADE
	);`,
		`
	CREATE TABLE IF NOT EXISTS Nodes (
		MapId   VARCHAR(32)   NOT NULL,
		NodeId  VARBINARY(32) NOT NULL,
		Version	INTEGER       NOT NULL,
		Value	BLOB(32)      NOT NULL,
		PRIMARY KEY(MapId, NodeId, Version),
		FOREIGN KEY(MapId) REFERENCES Maps(MapId) ON DELETE CASCADE
	);`,
	}
	hasher          = sparse.CONIKSHasher
	errNilLeaf      = errors.New("nil leaf")
	errIndexLen     = errors.New("index len != 32")
	errInvalidEpoch = errors.New("invalid epoch")
)

const (
	maxDepth = sparse.IndexLen
	size     = sparse.HashSize
	readExpr = `
	SELECT Value FROM Nodes
	WHERE MapId = ? AND NodeId = ? and Version <= ?
	ORDER BY Version DESC LIMIT 1;`
	leafExpr = `
	SELECT Data FROM Leaves
	WHERE MapId = ? AND LeafId = ? and Version <= ?
	ORDER BY Version DESC LIMIT 1;`
	queueExpr = `
	REPLACE INTO Leaves (MapId, LeafId, Version, Data)
	VALUES (?, ?, ?, ?);`
	pendingLeafsExpr = `
	SELECT LeafId, Version, Data FROM Leaves 
	WHERE MapId = ? AND Version >= ?;`
	setNodeExpr = `
	REPLACE INTO Nodes (MapId, NodeId, Version, Value)
	VALUES (?, ?, ?, ?);`
	mapRowExpr    = `REPLACE INTO Maps (MapId) VALUES (?);`
	readEpochExpr = `
	SELECT Version FROM Nodes
	WHERE MapId = ? AND NodeId = ?
	ORDER BY Version DESC LIMIT 1;`
)

// Map stores a temporal sparse merkle tree, backed by an SQL database.
type Map struct {
	mapID []byte
	db    *sql.DB
	epoch int64 // The currently valid epoch. Insert at epoch+1.
}

// New creates a new map.
func New(db *sql.DB, mapID string) (*Map, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("No DB connection: %v", err)
	}

	m := &Map{
		mapID: []byte(mapID),
		db:    db,
	}
	if err := m.create(); err != nil {
		return nil, err
	}
	if err := m.insertMapRow(); err != nil {
		return nil, err
	}
	index, depth := tree.InvertBitString("")
	nodeValue := hasher.HashEmpty(m.mapID, index, depth)
	if err := m.setRootAt(nil, nodeValue, -1); err != nil {
		return nil, err
	}
	epoch, err := m.readEpoch()
	if err != nil {
		return nil, err
	}
	m.epoch = epoch
	return m, nil
}

// Epoch returns the current epoch of the merkle tree.
func (m *Map) Epoch() int64 {
	return m.epoch
}

// QueueLeaf should only be called by the sequencer. If txn is nil, the operation
// will not run in a transaction.
func (m *Map) QueueLeaf(txn transaction.Txn, index, leaf []byte) error {
	if got, want := len(index), size; got != want {
		return errIndexLen
	}
	if leaf == nil {
		return errNilLeaf
	}

	// Write leaf nodes
	var stmt *sql.Stmt
	var err error
	if txn != nil {
		stmt, err = txn.Prepare(queueExpr)
	} else {
		stmt, err = m.db.Prepare(queueExpr)
	}
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.mapID, index, m.epoch+1, leaf)
	return err
}

type leafRow struct {
	index   []byte
	version int64
	data    []byte
}

// Commit takes all the Queued values since the last Commmit() and writes them.
// Commit is NOT multi-process safe. It should only be called from the sequencer.
func (m *Map) Commit(ctx context.Context) (int64, error) {
	// Get the list of pending leafs
	stmt, err := m.db.Prepare(pendingLeafsExpr)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(m.mapID, m.epoch+1)
	if err != nil {
		return -1, err
	}
	leafRows := make([]leafRow, 0, 10)
	for rows.Next() {
		var r leafRow
		err = rows.Scan(&r.index, &r.version, &r.data)
		if err != nil {
			return -1, err
		}
		leafRows = append(leafRows, r)
	}

	for _, r := range leafRows {
		if err := m.setLeafAt(nil, r.index, maxDepth, r.data, m.epoch+1); err != nil {
			// Recovery from here would mean updating nodes that
			// didn't get included so that they would be included
			// in the next epoch.
			return -1, fmt.Errorf("Failed to set node: %v", err)
		}
	}
	// Always update the root node.
	if len(leafRows) == 0 {
		root, err := m.ReadRootAt(nil, m.epoch)
		if err != nil {
			return -1, fmt.Errorf("No root for epoch %d: %v", m.epoch, err)
		}
		if err := m.setRootAt(nil, sparse.FromBytes(root), m.epoch+1); err != nil {
			return -1, fmt.Errorf("Failed to set root: %v", err)
		}
	}

	m.epoch++
	return m.epoch, nil
}

// ReadRootAt returns the value of the root node in a specific epoch. If txn is
// nil the operation will not run in a transaction.
func (m *Map) ReadRootAt(txn transaction.Txn, epoch int64) ([]byte, error) {
	var stmt *sql.Stmt
	var err error
	if txn != nil {
		stmt, err = txn.Prepare(readExpr)
	} else {
		stmt, err = m.db.Prepare(readExpr)
	}
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var value []byte
	if err := stmt.QueryRow(m.mapID, m.nodeID(""), epoch).Scan(&value); err != nil {
		return nil, err
	}
	return value, nil
}

// ReadLeafAt returns the leaf value at epoch. If txn is nil, the operation will
// not run in a transaction.
func (m *Map) ReadLeafAt(txn transaction.Txn, index []byte, epoch int64) ([]byte, error) {
	var readStmt *sql.Stmt
	var err error
	if txn != nil {
		readStmt, err = txn.Prepare(leafExpr)
	} else {
		readStmt, err = m.db.Prepare(leafExpr)
	}
	if err != nil {
		return nil, err
	}
	defer readStmt.Close()

	var value []byte
	if err = readStmt.QueryRow(m.mapID, index, epoch).Scan(&value); err == sql.ErrNoRows {
		return nil, nil // Not found is not an error.
	} else if err != nil {
		return nil, err
	}
	return value, nil
}

// NeighborsAt returns the list of neighbors from the neighbor leaf to just below the root at epoch.
func (m *Map) NeighborsAt(ctx context.Context, index []byte, epoch int64) ([][]byte, error) {
	tx, err := m.db.Begin()
	if err != nil {
		return nil, err
	}
	nbrs, err := m.neighborsAt(tx, index, maxDepth, epoch)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			err = fmt.Errorf("neighborsAt failed: %v, and Rollback failed: %v", err, rbErr)
		}
		return nil, err
	}
	cNbrs := compressNeighbors(m.mapID, nbrs, index, maxDepth)
	return cNbrs, tx.Commit()
}

func (m *Map) neighborsAt(tx *sql.Tx, index []byte, depth int, epoch int64) ([]sparse.Hash, error) {
	bindex := tree.BitString(index)[:depth]
	neighborBIndexes := tree.Neighbors(bindex)
	neighborIDs := m.nodeIDs(neighborBIndexes)

	readStmt, err := tx.Prepare(readExpr)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			err = fmt.Errorf("Prepare failed: %v, and Rollback failed: %v", err, rbErr)
		}
		return nil, err
	}
	defer readStmt.Close()

	// Get neighbors.
	nbrValues := make([]sparse.Hash, len(neighborIDs))
	for i, nodeID := range neighborIDs {
		var tmp []byte
		if err := readStmt.QueryRow(m.mapID, nodeID, epoch).Scan(&tmp); err == sql.ErrNoRows {
			nIndex, nDepth := tree.InvertBitString(neighborBIndexes[i])
			nbrValues[i] = hasher.HashEmpty(m.mapID, nIndex, nDepth)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("QueryRow failed: %v, and Rollback failed: %v", err, rbErr)
			}
			return nil, err
		} else {
			nbrValues[i] = sparse.FromBytes(tmp)
		}
	}

	return nbrValues, nil
}

func compressNeighbors(mapID []byte, neighbors []sparse.Hash, index []byte, depth int) [][]byte {
	bindex := tree.BitString(index)[:depth]
	neighborBIndexes := tree.Neighbors(bindex)
	compressed := make([][]byte, len(neighbors))
	for i, v := range neighbors {
		// TODO: convert values to arrays rather than slices for comparison.
		nIndex, nDepth := tree.InvertBitString(neighborBIndexes[i])
		if v != hasher.HashEmpty(mapID, nIndex, nDepth) {
			compressed[i] = v.Bytes()
		}
	}
	return compressed
}

// setLeafAt sets leaf node values directly at epoch.
func (m *Map) setLeafAt(ctx context.Context, index []byte, depth int, value []byte, epoch int64) (returnErr error) {
	if len(value) == 0 {
		return nil
	}
	bindex := tree.BitString(index)[:depth]
	nodeBindexes := tree.Path(bindex)
	nodeIDs := m.nodeIDs(nodeBindexes)

	// Read the neighbor nodes
	// Set the node
	// Compute new values
	// Set those values.

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if returnErr != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				returnErr = fmt.Errorf("setLeafAt failed: %v, and Rollback failed: %v", returnErr, rbErr)
			}
		}
	}()

	writeStmt, err := tx.Prepare(setNodeExpr)
	if err != nil {
		return err
	}
	defer writeStmt.Close()

	// Get neighbors.
	nbrValues, err := m.neighborsAt(tx, index, depth, epoch)
	if err != nil {
		return err
	}

	nodeValues := sparse.NodeValues(m.mapID, hasher, bindex, value, nbrValues)

	// Save new nodes.
	for i, nodeValue := range nodeValues {
		_, err = writeStmt.Exec(m.mapID, nodeIDs[i], epoch, nodeValue.Bytes())
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// setRootAt sets root node values directly at epoch.
func (m *Map) setRootAt(ctx context.Context, value sparse.Hash, epoch int64) error {
	writeStmt, err := m.db.Prepare(setNodeExpr)
	if err != nil {
		return fmt.Errorf("setRootAt(): %v", err)
	}
	defer writeStmt.Close()
	_, err = writeStmt.Exec(m.mapID, m.nodeID(""), epoch, value.Bytes())
	if err != nil {
		return fmt.Errorf("setRootAt(): %v", err)
	}
	return nil
}

// Create creates a new database.
func (m *Map) create() error {
	for _, stmt := range createStmt {
		_, err := m.db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("Failed to create map tables: %v", err)
		}
	}
	return nil
}

func (m *Map) insertMapRow() error {
	stmt, err := m.db.Prepare(mapRowExpr)
	if err != nil {
		return fmt.Errorf("insertMapRow(): %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.mapID)
	if err != nil {
		return fmt.Errorf("insertMapRow(): %v", err)
	}
	return nil
}

func (m *Map) readEpoch() (int64, error) {
	stmt, err := m.db.Prepare(readEpochExpr)
	if err != nil {
		return -1, fmt.Errorf("readEpoch(): %v", err)
	}
	defer stmt.Close()
	var epoch sql.NullInt64
	if err := stmt.QueryRow(m.mapID, m.nodeID("")).Scan(&epoch); err != nil {
		return -1, fmt.Errorf("Error reading epoch: %v", err)
	}
	if !epoch.Valid {
		return -1, errInvalidEpoch
	}
	return epoch.Int64, nil
}

// Converts a list of bit strings into their node IDs.
func (m *Map) nodeIDs(bindexes []string) [][]byte {
	nodes := make([][]byte, len(bindexes))
	for i, bindex := range bindexes {
		nodes[i] = m.nodeID(bindex)
	}
	return nodes
}

// nodeID computes the location of a node, given its bit string index.
func (m *Map) nodeID(bindex string) []byte {
	h := sha256.New()
	h.Write(m.mapID)
	h.Write([]byte{0})
	h.Write([]byte(bindex))
	return h.Sum(nil)
}

// PrefixLen returns the index of the last non-zero item in the list
func PrefixLen(nodes [][]byte) int {
	// Iterate over the nodes from leaf to root.
	for i, v := range nodes {
		if v != nil {
			// return the first non-empty node.
			return len(nodes) - i
		}
	}
	return 0
}
