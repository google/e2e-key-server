// Copyright 2018 Google Inc. All Rights Reserved.
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

package mutationstorage

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/google/keytransparency/core/mutator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/google/keytransparency/core/api/v1/keytransparency_go_proto"
)

// AddLogs creates and adds new logs for writing to a domain.
func (m *Mutations) AddLogs(ctx context.Context, domainID string, logIDs ...int64) error {
	glog.Infof("mutationstorage: AddLog(%v, %v)", domainID, logIDs)
	for _, logID := range logIDs {
		// TODO(gdbelvin): Use INSERT IGNORE to allow this function to be retried.
		// TODO(gdbelvin): Migrate to a MySQL Docker image for unit tests.
		// MySQL and SQLite do not have the same syntax for INSERT IGNORE.
		if _, err := m.db.ExecContext(ctx,
			`INSERT INTO Logs (DomainID, LogID, Enabled)  Values(?, ?, ?);`,
			domainID, logID, true); err != nil {
			return err
		}
	}
	return nil
}

// Send writes mutations to the leading edge (by sequence number) of the mutations table.
func (m *Mutations) Send(ctx context.Context, domainID string, update *pb.EntryUpdate) error {
	glog.Infof("mutationstorage: Send(%v, <mutation>)", domainID)
	logID, err := m.randLog(ctx, domainID)
	if err != nil {
		return err
	}
	mData, err := proto.Marshal(update)
	if err != nil {
		return err
	}
	// TODO(gbelvin): Implement retry with backoff for retryable errors if
	// we get timestamp contention.
	return m.send(ctx, domainID, logID, mData, time.Now())
}

// ListLogs returns a list of all logs for domainID, optionally filtered for writable logs.
func (m *Mutations) ListLogs(ctx context.Context, domainID string, writable bool) ([]int64, error) {
	var query string
	if writable {
		query = `SELECT LogID from Logs WHERE DomainID = ? AND Enabled = True;`
	} else {
		query = `SELECT LogID from Logs WHERE DomainID = ?;`
	}
	var logIDs []int64
	rows, err := m.db.QueryContext(ctx, query, domainID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var logID int64
		if err := rows.Scan(&logID); err != nil {
			return nil, err
		}
		logIDs = append(logIDs, logID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(logIDs) == 0 {
		return nil, status.Errorf(codes.NotFound, "no log found for domain %v", domainID)
	}
	return logIDs, nil
}

// randLog returns a random, enabled log for domainID.
func (m *Mutations) randLog(ctx context.Context, domainID string) (int64, error) {
	// TODO(gbelvin): Cache these results.
	writable := true
	logIDs, err := m.ListLogs(ctx, domainID, writable)
	if err != nil {
		return 0, err
	}

	// Return a random log.
	return logIDs[rand.Intn(len(logIDs))], nil
}

// ts must be greater than all other timestamps currently recorded for domainID.
func (m *Mutations) send(ctx context.Context, domainID string, logID int64, mData []byte, ts time.Time) (ret error) {
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() {
		if ret != nil {
			if err := tx.Rollback(); err != nil {
				ret = status.Errorf(codes.Internal, "%v, and could not rollback: %v", ret, err)
			}
		}
	}()

	var maxTime int64
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(Time), 0) FROM Queue WHERE DomainID = ? AND LogID = ?;`,
		domainID, logID).Scan(&maxTime); err != nil {
		return status.Errorf(codes.Internal, "could not find max timestamp: %v", err)
	}
	tsTime := ts.UnixNano()
	if tsTime <= maxTime {
		return status.Errorf(codes.Aborted,
			"current timestamp: %v, want > max-timestamp of queued mutations: %v",
			tsTime, maxTime)
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO Queue (DomainID, LogID, Time, Mutation) VALUES (?, ?, ?, ?);`,
		domainID, logID, tsTime, mData); err != nil {
		return status.Errorf(codes.Internal, "failed inserting into queue: %v", err)
	}
	return tx.Commit()
}

// HighWatermark returns the highest watermark in logID that is less than or
// equal to batchSize items greater than start.
func (m *Mutations) HighWatermark(ctx context.Context, domainID string, logID,
	start int64, batchSize int32) (int32, int64, error) {
	var count int32
	var high int64
	if err := m.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(MAX(T1.Time), ?) FROM 
		(
			SELECT Q.Time FROM Queue as Q
			WHERE Q.DomainID = ? AND Q.LogID = ? AND Q.Time > ?
			ORDER BY Q.Time ASC
			LIMIT ?
		) AS T1`,
		start, domainID, logID, start, batchSize).
		Scan(&count, &high); err != nil {
		return 0, 0, err
	}
	return count, high, nil
}

// ReadLog reads all mutations in logID between (low, high].
func (m *Mutations) ReadLog(ctx context.Context, domainID string,
	logID, low, high int64, batchSize, offset int32) ([]*mutator.LogMessage, error) {
	rows, err := m.db.QueryContext(ctx,
		`SELECT Time, Mutation FROM Queue
		WHERE DomainID = ? AND LogID = ? AND Time > ? AND Time <= ?
		ORDER BY Time ASC
		LIMIT ? OFFSET ?;`,
		domainID, logID, low, high, batchSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return readQueueMessages(rows)
}

func readQueueMessages(rows *sql.Rows) ([]*mutator.LogMessage, error) {
	results := make([]*mutator.LogMessage, 0)
	for rows.Next() {
		var timestamp int64
		var mData []byte
		if err := rows.Scan(&timestamp, &mData); err != nil {
			return nil, err
		}
		entryUpdate := new(pb.EntryUpdate)
		if err := proto.Unmarshal(mData, entryUpdate); err != nil {
			return nil, err
		}
		results = append(results, &mutator.LogMessage{
			ID:        timestamp,
			Mutation:  entryUpdate.Mutation,
			ExtraData: entryUpdate.Committed,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
