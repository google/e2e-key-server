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
	"time"

	"github.com/google/keytransparency/core/mutator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	pb "github.com/google/keytransparency/core/api/v1/keytransparency_go_proto"
)

// Send writes mutations to the leading edge (by sequence number) of the mutations table.
func (m *Mutations) Send(ctx context.Context, domainID string, update *pb.EntryUpdate) error {
	glog.Infof("mutationstorage: Send(%v, <mutation>)", domainID)
	mData, err := proto.Marshal(update)
	if err != nil {
		return err
	}
	// TODO(gbelvin): Implement retry with backoff for retryable errors if
	// we get timestamp contention.
	return m.send(ctx, domainID, mData, time.Now())
}

// ts must be greater than all other timestamps currently recorded for domainID.
func (m *Mutations) send(ctx context.Context, domainID string, mData []byte, ts time.Time) error {
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	var maxTime int64
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(Time), 0) FROM Queue WHERE DomainID = ?;`,
		domainID).Scan(&maxTime); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return status.Errorf(codes.Internal,
				"query err: %v and could not roll back: %v", err, rollbackErr)
		}
		return err
	}
	tsTime := ts.UnixNano()
	if tsTime <= maxTime {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return status.Errorf(codes.Internal, "could not roll back: %v", rollbackErr)
		}
		return status.Errorf(codes.Aborted,
			"current timestamp: %v, want > max-timestamp of queued mutations: %v",
			tsTime, maxTime)
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO Queue (DomainID, Time, Mutation) VALUES (?, ?, ?);`,
		domainID, tsTime, mData); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return status.Errorf(codes.Internal,
				"insert err: %v and could not roll back: %v", err, rollbackErr)
		}
		return status.Errorf(codes.Internal, "failed inserting into queue: %v", err)
	}
	return tx.Commit()
}

// HighWatermark returns the highest timestamp in the mutations table.
func (m *Mutations) HighWatermark(ctx context.Context, domainID string) (int64, error) {
	var watermark int64
	if err := m.db.QueryRowContext(ctx,
		`SELECT Time FROM Queue WHERE DomainID = ? ORDER BY Time DESC LIMIT 1;`,
		domainID).Scan(&watermark); err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return watermark, nil
}

// ReadQueue reads all mutations that are still in the queue up to batchSize.
func (m *Mutations) ReadQueue(ctx context.Context, domainID string, low, high int64) ([]*mutator.QueueMessage, error) {
	rows, err := m.db.QueryContext(ctx,
		`SELECT Time, Mutation FROM Queue
		WHERE DomainID = ? AND
		Time > ? AND Time <= ?
		ORDER BY Time ASC;`,
		domainID, low, high)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return readQueueMessages(rows)
}

func readQueueMessages(rows *sql.Rows) ([]*mutator.QueueMessage, error) {
	results := make([]*mutator.QueueMessage, 0)
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
		results = append(results, &mutator.QueueMessage{
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

// DeleteMessages removes messages from the queue.
func (m *Mutations) DeleteMessages(ctx context.Context, domainID string, mutations []*mutator.QueueMessage) error {
	glog.V(4).Infof("mutationstorage: DeleteMessages(%v, <mutation>)", domainID)
	delStmt, err := m.db.Prepare(deleteQueueExpr)
	if err != nil {
		return err
	}
	defer delStmt.Close()
	var retErr error
	for _, mutation := range mutations {
		if _, err = delStmt.ExecContext(ctx, domainID, mutation.ID); err != nil {
			// If an error occurs, take note, but continue deleting
			// the other referenced mutations.
			retErr = err
		}
	}
	return retErr
}
