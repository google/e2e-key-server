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

// Package mapper contains a transformation pipeline from log messages to map revisions.
package mapper

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	"github.com/google/keytransparency/core/mutator"
	"github.com/google/keytransparency/core/mutator/entry"

	pb "github.com/google/keytransparency/core/api/v1/keytransparency_go_proto"
	tpb "github.com/google/trillian"
)

// IndexedUpdate is a KV<Index, Update> type.
type IndexedUpdate struct {
	Index  []byte
	Update *pb.EntryUpdate
}

// MapUpdateFn converts an update into an IndexedUpdate.
func MapUpdateFn(msg *pb.EntryUpdate) (*IndexedUpdate, error) {
	var e pb.Entry
	if err := proto.Unmarshal(msg.Mutation.Entry, &e); err != nil {
		return nil, err
	}
	return &IndexedUpdate{
		Index:  e.Index,
		Update: msg,
	}, nil
}

// ReduceFn decides which of multiple updates can be applied in this revision.
// TODO(gbelvin): Move to mutator interface.
func ReduceFn(mutatorFn mutator.ReduceMutationFn,
	index []byte, leaves []*tpb.MapLeaf, msgs []*pb.EntryUpdate, emit func(*tpb.MapLeaf)) error {
	if got := len(leaves); got > 1 {
		return fmt.Errorf("expected 0 or 1 map leaf for index %x, got %v", index, got)
	}
	var oldValue *pb.SignedEntry // If no map leaf was found, oldValue will be nil.
	var err error
	if len(leaves) > 0 {
		oldValue, err = entry.FromLeafValue(leaves[0].GetLeafValue())
		if err != nil {
			return fmt.Errorf("entry.FromLeafValue(): %v", err)
		}
	}

	if len(msgs) == 0 {
		return fmt.Errorf("no msgs for index %x", index)
	}

	// TODO(gbelvin): Choose the mutation deterministically, regardless of the messages order.
	// (optional): Select the mutation based on it's correctness.
	msg := msgs[0]
	newValue, err := mutatorFn(oldValue, msg.Mutation)
	if err != nil {
		glog.Warningf("Mutate(): %v", err)
		return nil // A bad mutation should not make the whole batch to fail.
	}
	leafValue, err := entry.ToLeafValue(newValue)
	if err != nil {
		glog.Warningf("ToLeafValue(): %v", err)
		return nil // A bad mutation should not cause the entire pipeline to fail.
	}
	extraData, err := proto.Marshal(msg.Committed)
	if err != nil {
		glog.Warningf("proto.Marshal(): %v", err)
		return nil
	}
	emit(&tpb.MapLeaf{Index: index, LeafValue: leafValue, ExtraData: extraData})
	return nil
}
