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

package keyserver

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/key-transparency/core/appender"
	"github.com/google/key-transparency/core/authentication"
	"github.com/google/key-transparency/core/commitments"
	"github.com/google/key-transparency/core/queue"
	"github.com/google/key-transparency/core/tree"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/google/key-transparency/proto/keytransparency_v1"
)

func TestListEntryHistory(t *testing.T) {
	profileCount := 24
	ctx := context.Background()
	for i, tc := range []struct {
		start       int64
		page        int32
		wantNext    int64
		wantHistory []int
		err         codes.Code
	}{
		{1, 1, 2, []int{0}, codes.OK},                                                           // one entry per page.
		{1, 10, 11, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, codes.OK},                              // 10 entries per page.
		{4, 10, 14, []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, codes.OK},                           // start epoch is not 1.
		{1, 0, 17, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, codes.OK},       // zero page size.
		{20, 10, 25, []int{19, 20, 21, 22, 23}, codes.OK},                                       // adjusted page size.
		{24, 10, 25, []int{23}, codes.OK},                                                       // requesting the very last entry.
		{1, 1000000, 17, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, codes.OK}, // DOS prevention.
		{40, 10, 0, []int{}, codes.InvalidArgument},                                             // start epoch is beyond current epoch.
		{0, 10, 0, []int{}, codes.InvalidArgument},                                              // start epoch is less than 1.
	} {
		// Test case setup.
		c := &FakeCommitter{make(map[string]*pb.Committed)}
		st := &FakeSparseHist{make(map[int64][]byte)}
		a := &FakeAppender{0, 0}
		srv := New(c, FakeQueue{}, st, a, FakePrivateKey{}, FakeMutator{}, authentication.NewFake())

		if err := addProfiles(profileCount, c, st, a); err != nil {
			t.Fatalf("addProfile(%v, _, _, _)=%v", profileCount, err)
		}

		// Run test case.
		req := &pb.ListEntryHistoryRequest{
			UserId:   "",
			Start:    tc.start,
			PageSize: tc.page,
		}
		resp, err := srv.ListEntryHistory(ctx, req)
		if got, want := grpc.Code(err), tc.err; got != want {
			t.Errorf("%v: ListEntryHistory(_, %v)=(_, %v), want (err==nil) %v", i, req, err, tc.err)
		}
		// Skip the rest of the test if there is an error.
		if err != nil {
			continue
		}

		// Check next epoch.
		if got, want := resp.NextStart, tc.wantNext; got != want {
			t.Errorf("%v: NextEpoch=%v, want %v", i, got, want)
		}

		// Ensure that history has the correct number of entries.
		if got, want := len(resp.Values), len(tc.wantHistory); got != want {
			t.Errorf("%v: len(resp.Values)=%v, want %v", i, got, want)
			// Skip the rest of the test if the returned history is
			// not of the expected length.
			continue
		}

		if got := checkProfiles(tc.wantHistory, resp.Values); got != nil {
			t.Errorf("%v: checkProfiles failed: %v, want nil", i, got)
		}

		// Verify mocks.
		// Ensure that latest is called only once.
		if got, want := a.LatestCount, 1; got != want {
			t.Errorf("%v: Incorrect number of Latest() call(s), got %v, want %v", i, got, want)
		}
	}
}

func addProfiles(profileCount int, c commitments.Committer, st tree.SparseHist, a appender.Appender) error {
	for i := 0; i < profileCount; i++ {
		commitment := []byte{uint8(i)}

		// Fill the committer map.
		p := createProfile(i)
		pData, err := proto.Marshal(p)
		if err != nil {
			return fmt.Errorf("%v: Failed to Marshal: %v", i, err)
		}
		committed := &pb.Committed{Data: pData}
		c.(*FakeCommitter).M[string(commitment)] = committed

		// Increase epoch
		a.(*FakeAppender).CurrentEpoch++

		// Fill the tree map.
		st.(*FakeSparseHist).M[a.(*FakeAppender).CurrentEpoch] = commitment
	}
	return nil
}

// checkProfiles Ensure that the history has the correct entries in the correct
// order.
func checkProfiles(wantHistory []int, values []*pb.GetEntryResponse) error {
	for i, tag := range wantHistory {
		p := new(pb.Profile)
		if err := proto.Unmarshal(values[i].Committed.Data, p); err != nil {
			return fmt.Errorf("%v: Failed to Unmarshal: %v", i, err)
		}
		if got, want := p, createProfile(tag); !reflect.DeepEqual(got, want) {
			return fmt.Errorf("%v: Invalid profile: %v, want %v", i, got, want)
		}
	}
	return nil
}

// createProfile creates a dummy profile using the passed tag.
func createProfile(tag int) *pb.Profile {
	return &pb.Profile{
		Keys: map[string][]byte{
			fmt.Sprintf("foo%v", tag): []byte(fmt.Sprintf("bar%v", tag)),
		},
	}
}

///////////
// Fakes //
///////////

// commitments.Committer fake.
type FakeCommitter struct {
	M map[string]*pb.Committed
}

func (*FakeCommitter) Write(ctx context.Context, commitment []byte, committed *pb.Committed) error {
	return nil
}

func (f *FakeCommitter) Read(ctx context.Context, commitment []byte) (*pb.Committed, error) {
	committed, ok := f.M[string(commitment)]
	if !ok {
		return nil, nil
	}
	return committed, nil
}

// queue.Queuer fake.
type FakeQueue struct {
}

func (FakeQueue) Enqueue(key, value []byte) error {
	return nil
}

func (FakeQueue) AdvanceEpoch() error {
	return nil
}

func (FakeQueue) Dequeue(processFunc queue.ProcessKeyValueFunc, advanceFunc queue.AdvanceEpochFunc) error {
	return nil
}

// tree.SparseHist fake.
type FakeSparseHist struct {
	M map[int64][]byte
}

func (*FakeSparseHist) QueueLeaf(ctx context.Context, index, leaf []byte) error {
	return nil
}

func (*FakeSparseHist) Commit() (epoch int64, err error) {
	return 0, nil
}

func (*FakeSparseHist) ReadRootAt(ctx context.Context, epoch int64) ([]byte, error) {
	return nil, nil
}

func (f *FakeSparseHist) ReadLeafAt(ctx context.Context, index []byte, epoch int64) ([]byte, error) {
	commitment, ok := f.M[epoch]
	if !ok {
		return nil, errors.New("not found")
	}
	entry := &pb.Entry{Commitment: commitment}
	entryData, err := proto.Marshal(entry)
	if err != nil {
		return nil, errors.New("marshaling error")
	}
	return entryData, nil
}

func (*FakeSparseHist) NeighborsAt(ctx context.Context, index []byte, epoch int64) ([][]byte, error) {
	return nil, nil
}

func (*FakeSparseHist) Epoch() int64 {
	return 0
}

// appender.Appender fake.
type FakeAppender struct {
	CurrentEpoch int64
	LatestCount  int
}

func (*FakeAppender) Append(ctx context.Context, epoch int64, obj interface{}) error {
	return nil
}

func (*FakeAppender) Epoch(ctx context.Context, epoch int64, obj interface{}) ([]byte, error) {
	return nil, nil
}

func (f *FakeAppender) Latest(ctx context.Context, obj interface{}) (int64, []byte, error) {
	f.LatestCount++
	return f.CurrentEpoch, nil, nil
}

// vrf.PrivateKey fake.
type FakePrivateKey struct {
}

func (FakePrivateKey) Evaluate(m []byte) (vrf []byte, proof []byte) {
	return nil, nil
}

func (FakePrivateKey) Index(vrf []byte) [32]byte {
	return [32]byte{}
}

// mutator.Mutator fake.
type FakeMutator struct {
}

func (FakeMutator) CheckMutation(value, mutation []byte) error {
	return nil
}

func (FakeMutator) Mutate(value, mutation []byte) ([]byte, error) {
	return nil, nil
}
