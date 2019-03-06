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

// Package runner executes the mapper pipeline.
package runner

import (
	"fmt"

	"github.com/google/keytransparency/core/mutator"
	"github.com/google/keytransparency/core/mutator/entry"

	pb "github.com/google/keytransparency/core/api/v1/keytransparency_go_proto"
	tpb "github.com/google/trillian"
)

// Joined is the result of a CoGroupByKey on []*MapLeaf and []*IndexedValue.
type Joined struct {
	Index   []byte
	Values1 []*pb.EntryUpdate
	Values2 []*pb.EntryUpdate
}

// Join pairs up MapLeaves and IndexedValue by index.
func Join(leaves []*entry.IndexedValue, msgs []*entry.IndexedValue) []*Joined {
	joinMap := make(map[string]*Joined)
	for _, l := range leaves {
		row, ok := joinMap[string(l.Index)]
		if !ok {
			row = &Joined{Index: l.Index}
		}
		row.Values1 = append(row.Values1, l.Value)
		joinMap[string(l.Index)] = row
	}
	for _, m := range msgs {
		row, ok := joinMap[string(m.Index)]
		if !ok {
			row = &Joined{Index: m.Index}
		}
		row.Values2 = append(row.Values2, m.Value)
		joinMap[string(m.Index)] = row
	}
	ret := make([]*Joined, 0, len(joinMap))
	for _, r := range joinMap {
		ret = append(ret, r)
	}
	return ret
}

// DoMapLogItemsFn runs the MapLogItemsFn on each element of msgs.
func DoMapLogItemsFn(fn mutator.MapLogItemFn, msgs []*mutator.LogMessage, emitErr func(error)) []*entry.IndexedValue {
	outs := make([]*entry.IndexedValue, 0, len(msgs))
	for _, m := range msgs {
		fn(m,
			func(index []byte, value *pb.EntryUpdate) {
				outs = append(outs, &entry.IndexedValue{Index: index, Value: value})
			},
			func(err error) { emitErr(fmt.Errorf("mapLogItemFn: %v", err)) },
		)
	}
	return outs
}

// MapMapLeafFn converts an update into an IndexedValue.
type MapMapLeafFn func(*tpb.MapLeaf) (*entry.IndexedValue, error)

func DoMapMapLeafFn(fn MapMapLeafFn, leaves []*tpb.MapLeaf) ([]*entry.IndexedValue, error) {
	outs := make([]*entry.IndexedValue, 0, len(leaves))
	for _, m := range leaves {
		out, err := fn(m)
		if err != nil {
			return nil, err
		}
		outs = append(outs, out)
	}
	return outs, nil
}

// ReduceMutationFn takes all the mutations for an index and an auxiliary input
// of existing mapleaf(s) and emits a new value for the index.
// ReduceMutationFn must be  idempotent, commutative, and associative.  i.e.
// must produce the same output  regardless of input order or grouping,
// and it must be safe to run multiple times.
type ReduceMutationFn func(msgs []*pb.EntryUpdate, leaves []*pb.EntryUpdate,
	emit func(*pb.EntryUpdate), emitErr func(error))

// DoReduceFn takes the set of mutations and applies them to given leaves.
// Returns a list of key value pairs that should be written to the map.
func DoReduceFn(reduceFn ReduceMutationFn, joined []*Joined, emitErr func(error)) []*entry.IndexedValue {
	ret := make([]*entry.IndexedValue, 0, len(joined))
	for _, j := range joined {
		reduceFn(j.Values1, j.Values2,
			func(e *pb.EntryUpdate) {
				ret = append(ret, &entry.IndexedValue{Index: j.Index, Value: e})
			},
			func(err error) { emitErr(fmt.Errorf("reduceFn on index %x: %v", j.Index, err)) },
		)
	}
	return ret
}

// DoMarshalIndexedValues executes Marshal on each IndexedValue
// If marshal fails, it will emit an error and continue with a subset of ivs.
func DoMarshalIndexedValues(ivs []*entry.IndexedValue, emitErr func(error)) []*tpb.MapLeaf {
	ret := make([]*tpb.MapLeaf, 0, len(ivs))
	for _, iv := range ivs {
		mapLeaf, err := iv.Marshal()
		if err != nil {
			emitErr(err)
			continue
		}
		ret = append(ret, mapLeaf)
	}
	return ret
}
