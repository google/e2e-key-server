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

package client

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/keytransparency/core/testutil"
	"github.com/google/trillian"
	"github.com/google/trillian/types"
	"github.com/kylelemons/godebug/pretty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/google/keytransparency/core/api/v1/keytransparency_go_proto"
)

func TestMapRevisionFor(t *testing.T) {
	for _, tc := range []struct {
		treeSize     uint64
		wantRevision uint64
		wantErr      error
	}{
		{treeSize: 1, wantRevision: 0},
		{treeSize: 0, wantRevision: 0, wantErr: ErrLogEmpty},
		{treeSize: ^uint64(0), wantRevision: ^uint64(0) - 1},
	} {
		revision, err := mapRevisionFor(&types.LogRootV1{TreeSize: tc.treeSize})
		if got, want := revision, tc.wantRevision; got != want {
			t.Errorf("mapRevisionFor(%v).Revision: %v, want: %v", tc.treeSize, got, want)
		}
		if got, want := err, tc.wantErr; got != want {
			t.Errorf("mapRevisionFor(%v).err: %v, want: %v", tc.treeSize, got, want)
		}
	}
}

func TestCompressHistory(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		roots   map[uint64][]byte
		want    map[uint64][]byte
		wantErr error
	}{
		{
			desc: "Single",
			roots: map[uint64][]byte{
				1: []byte("a"),
			},
			want: map[uint64][]byte{
				1: []byte("a"),
			},
		},
		{
			desc: "Compress",
			roots: map[uint64][]byte{
				0: []byte("a"),
				1: []byte("a"),
				2: []byte("a"),
			},
			want: map[uint64][]byte{
				0: []byte("a"),
			},
		},
		{
			desc: "Not Contiguous",
			roots: map[uint64][]byte{
				0: []byte("a"),
				2: []byte("a"),
			},
			wantErr: ErrNonContiguous,
		},
		{
			desc: "Complex",
			roots: map[uint64][]byte{
				1: []byte("a"),
				2: []byte("a"),
				3: []byte("b"),
				4: []byte("b"),
				5: []byte("c"),
			},
			want: map[uint64][]byte{
				1: []byte("a"),
				3: []byte("b"),
				5: []byte("c"),
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := CompressHistory(tc.roots)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("compressHistory(): %#v, want %#v", got, tc.want)
			}
			if err != tc.wantErr {
				t.Errorf("compressHistory(): %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestPaginateHistory(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	appID := "fakeapp"
	userID := "fakeuser"

	srv := &fakeKeyServer{
		revisions: map[int64]*pb.GetEntryResponse{
			0:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{0}}},
			1:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{1}}},
			2:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{2}}},
			3:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{3}}},
			4:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{4}}},
			5:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{5}}},
			6:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{6}}},
			7:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{7}}},
			8:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{8}}},
			9:  {Smr: &trillian.SignedMapRoot{MapRoot: []byte{9}}},
			10: {Smr: &trillian.SignedMapRoot{MapRoot: []byte{10}}},
		},
	}
	s, stop, err := testutil.NewFakeKT(srv)
	if err != nil {
		t.Fatalf("NewFakeKT(): %v", err)
	}
	defer stop()

	for _, tc := range []struct {
		desc       string
		start, end int64
		wantErr    error
		wantValues map[uint64][]byte
	}{
		{
			desc:    "incomplete",
			start:   9,
			end:     15,
			wantErr: ErrIncomplete,
		},
		{
			desc: "1Item",
			end:  0,
			wantValues: map[uint64][]byte{
				0: nil,
			},
		},
		{
			desc: "2Items",
			end:  1,
			wantValues: map[uint64][]byte{
				0: nil,
				1: nil,
			},
		},
		{
			desc:  "3Times",
			start: 0,
			end:   10,
			wantValues: map[uint64][]byte{
				0:  nil,
				1:  nil,
				2:  nil,
				3:  nil,
				4:  nil,
				5:  nil,
				6:  nil,
				7:  nil,
				8:  nil,
				9:  nil,
				10: nil,
			},
		},
		{
			desc:  "pageSize",
			start: 0,
			end:   5,
			wantValues: map[uint64][]byte{
				0: nil,
				1: nil,
				2: nil,
				3: nil,
				4: nil,
				5: nil,
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			c := Client{
				Verifier: &fakeVerifier{},
				cli:      s.Client,
			}

			_, values, err := c.PaginateHistory(ctx, appID, userID, tc.start, tc.end)
			if err != tc.wantErr {
				t.Errorf("PaginateHistory(): %v, want %v", err, tc.wantErr)
			}
			if got, want := values, tc.wantValues; !reflect.DeepEqual(got, want) {
				t.Errorf("PaginateHistory().values: \n%#v, want \n%#v, diff: \n%v",
					got, want, pretty.Compare(got, want))
			}

		})
	}
}

type fakeKeyServer struct {
	revisions map[int64]*pb.GetEntryResponse
}

func (f *fakeKeyServer) ListEntryHistory(ctx context.Context, in *pb.ListEntryHistoryRequest) (*pb.ListEntryHistoryResponse, error) {
	currentEpoch := int64(len(f.revisions)) - 1 // len(1) contains map revision 0.
	if in.PageSize > 5 || in.PageSize == 0 {
		in.PageSize = 5 // Test maximum page size limits.
	}
	if in.Start+int64(in.PageSize) > currentEpoch {
		in.PageSize = int32(currentEpoch - in.Start + 1)
	}

	values := make([]*pb.GetEntryResponse, in.PageSize)
	for i := range values {
		values[i] = f.revisions[in.Start+int64(i)]
	}
	next := in.Start + int64(len(values))
	if next > currentEpoch {
		next = 0 // no more!
	}

	return &pb.ListEntryHistoryResponse{
		Values:    values,
		NextStart: next,
	}, nil
}

func (f *fakeKeyServer) GetDomain(context.Context, *pb.GetDomainRequest) (*pb.Domain, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (f *fakeKeyServer) GetEpoch(context.Context, *pb.GetEpochRequest) (*pb.Epoch, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (f *fakeKeyServer) GetLatestEpoch(context.Context, *pb.GetLatestEpochRequest) (*pb.Epoch, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (f *fakeKeyServer) GetEpochStream(*pb.GetEpochRequest, pb.KeyTransparency_GetEpochStreamServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

func (f *fakeKeyServer) ListMutations(context.Context, *pb.ListMutationsRequest) (*pb.ListMutationsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (f *fakeKeyServer) ListMutationsStream(*pb.ListMutationsRequest, pb.KeyTransparency_ListMutationsStreamServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

func (f *fakeKeyServer) GetEntry(context.Context, *pb.GetEntryRequest) (*pb.GetEntryResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (f *fakeKeyServer) UpdateEntry(context.Context, *pb.UpdateEntryRequest) (*pb.UpdateEntryResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

type fakeVerifier struct{}

func (f *fakeVerifier) Index(vrfProof []byte, domainID string, appID string, userID string) ([]byte, error) {
	return make([]byte, 32), nil
}

func (f *fakeVerifier) VerifyGetEntryResponse(ctx context.Context, domainID string, appID string, userID string, trusted types.LogRootV1, in *pb.GetEntryResponse) (*types.MapRootV1, *types.LogRootV1, error) {
	smr, err := f.VerifySignedMapRoot(in.Smr)
	return smr, &types.LogRootV1{}, err
}

func (f *fakeVerifier) VerifyEpoch(in *pb.Epoch, trusted types.LogRootV1) (*types.LogRootV1, *types.MapRootV1, error) {
	smr, err := f.VerifySignedMapRoot(in.Smr)
	return &types.LogRootV1{}, smr, err
}

func (f *fakeVerifier) VerifySignedMapRoot(smr *trillian.SignedMapRoot) (*types.MapRootV1, error) {
	return &types.MapRootV1{Revision: uint64(smr.MapRoot[0])}, nil
}
