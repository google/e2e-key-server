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

package keyserver

import (
	"testing"

	"github.com/golang/protobuf/proto"

	rtpb "github.com/google/keytransparency/core/keyserver/readtoken_go_proto"
	"github.com/google/keytransparency/core/mutator"
	spb "github.com/google/keytransparency/core/sequencer/sequencer_go_proto"
)

func TestEncodeToken(t *testing.T) {
	for _, tc := range []struct {
		rt   *rtpb.ReadToken
		want string
	}{
		{rt: &rtpb.ReadToken{}, want: ""},
		//{rt: nil, want: ""},
	} {
		got, err := EncodeToken(tc.rt)
		if err != nil {
			t.Fatalf("EncodeToken(%v): %v", tc.rt, err)
		}
		if got != tc.want {
			t.Fatalf("EncodeToken(%v): %v, want %v", tc.rt, got, tc.want)
		}
	}
}

func TestTokenEncodeDecode(t *testing.T) {
	rt1 := &rtpb.ReadToken{ShardId: 2, LowWatermark: 5}
	rt1Token, err := EncodeToken(rt1)
	if err != nil {
		t.Fatalf("EncodeToken(%v): %v", rt1, err)
	}
	for _, tc := range []struct {
		desc  string
		token string
		want  *rtpb.ReadToken
	}{
		{desc: "empty", token: "", want: &rtpb.ReadToken{}},
		{desc: "notempty", token: rt1Token, want: rt1},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			var got rtpb.ReadToken
			if err := DecodeToken(tc.token, &got); err != nil {
				t.Errorf("DecodeToken(%v): %v", tc.token, err)
			}
			if !proto.Equal(&got, tc.want) {
				t.Errorf("DecodeToken(%v): %v, want %v", tc.token, got, tc.want)
			}
		})
	}
}

func TestFirst(t *testing.T) {
	s := SourceMap{
		2: &spb.MapMetadata_SourceSlice{LowestWatermark: 1, HighestWatermark: 10},
		3: &spb.MapMetadata_SourceSlice{LowestWatermark: 10, HighestWatermark: 20},
	}
	want := &rtpb.ReadToken{ShardId: 2, LowWatermark: 1}
	if got := s.First(); !proto.Equal(got, want) {
		t.Errorf("First(): %v, want %v", got, want)
	}
}

func TestNext(t *testing.T) {
	s := SourceMap{
		2: &spb.MapMetadata_SourceSlice{LowestWatermark: 1, HighestWatermark: 10},
		3: &spb.MapMetadata_SourceSlice{LowestWatermark: 10, HighestWatermark: 20},
	}
	for _, tc := range []struct {
		desc    string
		rt      *rtpb.ReadToken
		lastRow *mutator.LogMessage
		want    *rtpb.ReadToken
	}{
		{
			desc:    "first page",
			rt:      &rtpb.ReadToken{ShardId: 2, LowWatermark: 1},
			lastRow: &mutator.LogMessage{ID: 6},
			want:    &rtpb.ReadToken{ShardId: 2, LowWatermark: 6},
		},
		{
			desc:    "next source",
			rt:      &rtpb.ReadToken{ShardId: 2, LowWatermark: 1},
			lastRow: nil,
			want:    &rtpb.ReadToken{ShardId: 3, LowWatermark: 10},
		},
		{
			desc:    "last page",
			rt:      &rtpb.ReadToken{ShardId: 3, LowWatermark: 1},
			lastRow: nil,
			want:    &rtpb.ReadToken{},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got := s.Next(tc.rt, tc.lastRow)
			if !proto.Equal(got, tc.want) {
				t.Errorf("Next(): %v, want %v", got, tc.want)
			}
		})
	}
}
