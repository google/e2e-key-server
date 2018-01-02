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
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	pb "github.com/google/keytransparency/core/api/v1/keytransparency_proto"
	"github.com/google/keytransparency/core/mutator"
)

// Send writes mutations to the leading edge (by sequence number) of the mutations table.
func (m *Mutations) Send(ctx context.Context, mapID int64, update *pb.EntryUpdate) error {
	index := update.GetMutation().GetIndex()
	mData, err := proto.Marshal(update)
	if err != nil {
		return err
	}
	writeStmt, err := m.db.Prepare(insertExpr)
	if err != nil {
		return err
	}
	defer writeStmt.Close()
	_, err = writeStmt.ExecContext(ctx, mapID, index, mData)
	return err
}

// NewReciever starts recieving messages sent to the queue. As batches become ready, recieveFunc will be called.
func (m *Mutations) NewReciever(ctx context.Context, last time.Time, mapID, start int64, recieveFunc func([]*pb.EntryUpdate) error, ropts mutator.RecieverOptions) *Reciever {
	r := &Reciever{
		m:           m,
		mapID:       mapID,
		opts:        ropts,
		ticker:      time.NewTicker(ropts.Period),
		maxTicker:   time.NewTicker(ropts.MaxPeriod),
		done:        make(chan bool),
		recieveFunc: recieveFunc,
	}

	go r.run(ctx, last)

	return r
}

// Reciever recieves messages from a queue.
type Reciever struct {
	m           mutator.MutationStorage
	start       uint64
	mapID       int64
	recieveFunc func([]*pb.EntryUpdate) error
	opts        mutator.RecieverOptions
	ticker      *time.Ticker
	maxTicker   *time.Ticker
	done        chan interface{}
	finished    sync.WaitGroup
}

// Close stops the reciever and returns only when all callbacks are complete.
func (r *Reciever) Close() {
	close(r.done)
	r.finished.Wait()
}

func (r *Reciever) run(ctx context.Context, last time.Time) {
	r.finished.Add(1)
	defer r.finished.Done()

	more := make(chan bool, 1)
	if time.Since(last) > (r.opts.MaxPeriod - r.opts.Period) {
		more <- true // We will be over due for an epoch soon.
	}

	for {
		select {
		case <-more:
		case <-r.ticker.C:
		case <-r.maxTicker.C:
			if count := r.sendBatch(ctx); count > r.opts.MaxBatchSize {
				// Continue sending until we drop below batch size.
				more <- true
			}
		case <-ctx.Done():
		case <-r.done:
			break
		}
	}
}

// sendBatch sends up to batchSize items to the reciever. Returns the number of sent items.
func (r *Reciever) sendBatch(ctx context.Context) int {
	max, ms, err := r.m.ReadAll(ctx, r.mapID, r.start)
	if err != nil {
		glog.Infof("ReadAll(%v): %v", r.start, err)
		return 0
	}

	// TODO: find a way to send mutation ids so the reciever can determine
	// the high water mark.
	if err := r.recieveFunc(ms); err != nil {
		glog.Infof("queue.SendBatch failed: %v", err)
		return 0
	}
	// TODO(gbelvin): Do we need finer grained errors?
	// We could put an ack'ed field in a QueueMessage object.
	// But I don't think we need that level of granularity -- yet?

	// Update our own high water mark.
	r.start = max
	return len(ms)
}