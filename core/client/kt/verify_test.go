// Copyright 2017 Google Inc. All Rights Reserved.
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

package kt

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/google/keytransparency/core/crypto/vrf/p256"
	"github.com/google/keytransparency/core/fake"
	"github.com/google/keytransparency/core/proto/keytransparency_v1_types"
	"github.com/google/trillian"
	"github.com/google/trillian/crypto/keys"
	"github.com/google/trillian/crypto/sigpb"
	"github.com/google/trillian/merkle/maphasher"
)

var (
	VRFPub = []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE5AV2WCmStBt4N2Dx+7BrycJFbxhW
f5JqSoyp0uiL8LeNYyj5vgklK8pLcyDbRqch9Az8jXVAmcBAkvaSrLW8wQ==
-----END PUBLIC KEY-----`)
	MapPub = []byte{0x30, 0x59, 0x30, 0x13, 0x6, 0x7, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x2, 0x1, 0x6, 0x8, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x3, 0x1, 0x7, 0x3, 0x42, 0x0, 0x4, 0xb0, 0x5, 0x60, 0xdd, 0x80, 0x74, 0xb4, 0xe1, 0x5f, 0xdc, 0x37, 0x42, 0xd9, 0x81, 0xcf, 0x2f, 0x65, 0xa2, 0xb8, 0x23, 0x51, 0xd6, 0x2c, 0xb0, 0xa8, 0x68, 0xe3, 0xb6, 0xed, 0x9d, 0x1, 0xd5, 0xa4, 0xb5, 0x6a, 0xa0, 0x44, 0xee, 0xd, 0x4e, 0xa3, 0xc9, 0x5d, 0x48, 0x20, 0x49, 0x2, 0xf7, 0x9e, 0xf3, 0xae, 0xa4, 0x70, 0x78, 0x59, 0x91, 0x59, 0xe2, 0x3d, 0xd0, 0x86, 0xd9, 0x96, 0x46}
)

// Test vectors were obtained by observing the integration tests.
func TestVerifyGetEntyrResponse(t *testing.T) {
	ctx := context.Background()
	vrfPub, err := p256.NewVRFVerifierFromPEM(VRFPub)
	if err != nil {
		t.Fatal(err)
	}
	mapPub, err := keys.NewFromPublicDER(MapPub)
	if err != nil {
		t.Fatal(err)
	}
	v := New(vrfPub, maphasher.Default, mapPub, fake.NewFakeTrillianLogVerifier())
	for _, tc := range []struct {
		desc          string
		wantErr       bool
		userID, appID string
		trusted       *trillian.SignedLogRoot
		in            *keytransparency_v1_types.GetEntryResponse
	}{
		{
			desc:    "empty",
			userID:  "noalice",
			appID:   "app",
			trusted: &trillian.SignedLogRoot{},
			in: &keytransparency_v1_types.GetEntryResponse{
				VrfProof:  []byte{0x8, 0xca, 0x1, 0xaa, 0x1, 0x74, 0x70, 0xf4, 0xbf, 0x96, 0x69, 0xff, 0x73, 0x5d, 0xc7, 0x9d, 0x2b, 0x51, 0xb0, 0xa5, 0x42, 0xb8, 0x8c, 0x42, 0x34, 0xfc, 0xa1, 0xa2, 0xc0, 0x80, 0xdf, 0x76, 0x33, 0xfc, 0x7d, 0xee, 0x4a, 0xf3, 0x18, 0xea, 0x30, 0xc4, 0xa4, 0x6f, 0x31, 0xc9, 0x20, 0x73, 0x42, 0x84, 0xd9, 0x71, 0x39, 0x52, 0xb2, 0x8f, 0x58, 0x52, 0x4, 0x53, 0x87, 0x8, 0x3e, 0x81, 0x4, 0xb, 0x13, 0x89, 0xd7, 0xc6, 0x63, 0x22, 0x39, 0x18, 0x73, 0x72, 0xfa, 0x32, 0xf6, 0xeb, 0x3, 0x8, 0x5d, 0x7, 0x4e, 0x2, 0x3a, 0xc6, 0x7f, 0x89, 0xe8, 0x44, 0x27, 0xcb, 0x73, 0xdc, 0xf2, 0x2f, 0xcc, 0xcd, 0x90, 0x6e, 0x97, 0xcb, 0x22, 0xff, 0x6e, 0xdb, 0x74, 0x22, 0xbf, 0x28, 0x27, 0x9b, 0x9e, 0x26, 0x1a, 0xe4, 0xc6, 0x16, 0x59, 0x4f, 0x7d, 0xcc, 0xb9, 0x8e, 0x7d, 0x41, 0xf7},
				Committed: nil,
				LeafProof: &trillian.MapLeafInclusion{
					Leaf:      &trillian.MapLeaf{},
					Inclusion: make([][]byte, 256),
				},
				Smr: &trillian.SignedMapRoot{
					TimestampNanos: 1502231274209403635,
					RootHash:       []byte{0xc6, 0x68, 0x9f, 0x10, 0x81, 0x2a, 0x9, 0x80, 0x97, 0x6d, 0x95, 0x33, 0xd8, 0x38, 0x75, 0x28, 0x21, 0x66, 0x15, 0x95, 0x67, 0xec, 0x35, 0x15, 0x57, 0x16, 0xc1, 0x41, 0x3a, 0xf5, 0x3d, 0x6a},
					Metadata:       &trillian.MapperMetadata{},
					Signature: &sigpb.DigitallySigned{
						HashAlgorithm:        4,
						SignatureAlgorithm:   3,
						SignatureCipherSuite: 0,
						Signature:            []byte{0x30, 0x45, 0x2, 0x21, 0x0, 0xbf, 0x13, 0x6a, 0xe4, 0xc3, 0x58, 0x23, 0xf3, 0x99, 0xb5, 0xe, 0x84, 0x2, 0x88, 0x40, 0x5c, 0xeb, 0x1a, 0x9a, 0xd3, 0x65, 0xb2, 0x21, 0x43, 0xbb, 0xce, 0xaf, 0xa7, 0x8c, 0x6b, 0xe1, 0xf, 0x2, 0x20, 0x71, 0x60, 0x94, 0xf8, 0x70, 0x2a, 0x64, 0x49, 0xa9, 0xdc, 0xa6, 0xde, 0x1, 0x9a, 0x8, 0xb6, 0xad, 0x76, 0x86, 0x16, 0x24, 0xa3, 0xab, 0xa7, 0x4b, 0x6c, 0x27, 0x8c, 0x6b, 0x79, 0x2a, 0xea},
					},
					MapId:       8245331544573830053,
					MapRevision: 1,
				},
			},
			wantErr: false,
		},
		{
			desc:    "Tree size 2",
			userID:  "nocarol",
			appID:   "app",
			trusted: &trillian.SignedLogRoot{},
			in: &keytransparency_v1_types.GetEntryResponse{
				VrfProof:  []byte{0xc5, 0x6c, 0x7a, 0xc0, 0xdf, 0x50, 0x8c, 0x6b, 0xc6, 0x72, 0x21, 0x9d, 0xf2, 0xdc, 0xf6, 0x36, 0xf1, 0xff, 0x4d, 0xe0, 0xa1, 0x1c, 0xc9, 0x95, 0x4b, 0x56, 0x85, 0x5c, 0xd2, 0x0, 0xb5, 0x8e, 0x52, 0x47, 0x9d, 0xfb, 0x42, 0x48, 0xb7, 0x87, 0x87, 0x59, 0xed, 0x2, 0xd8, 0x8f, 0x10, 0x84, 0xac, 0x45, 0x94, 0xa4, 0x29, 0xb1, 0x34, 0xeb, 0x1a, 0xe9, 0xfe, 0x47, 0xeb, 0x8a, 0xef, 0xba, 0x4, 0x7b, 0xcd, 0x7, 0x3, 0xb9, 0x69, 0xe3, 0x72, 0x35, 0xdb, 0xfc, 0xb2, 0xa3, 0x4c, 0x22, 0x6b, 0xaa, 0xce, 0x92, 0x6b, 0xcf, 0x2, 0x11, 0x78, 0x7b, 0x1f, 0x5c, 0x2f, 0xff, 0xb9, 0x34, 0x32, 0xa1, 0xd9, 0xba, 0xec, 0xa5, 0x9d, 0x5e, 0xa6, 0xbb, 0xb6, 0x77, 0x92, 0x4c, 0xc, 0x2d, 0x76, 0xdf, 0xbe, 0x9e, 0xa0, 0x93, 0xde, 0xf5, 0xa1, 0xc1, 0x4e, 0x9e, 0x19, 0x39, 0x16, 0xfe, 0x60},
				Committed: nil,
				LeafProof: &trillian.MapLeafInclusion{
					Leaf: &trillian.MapLeaf{},
					Inclusion: append(make([][]byte, 255),
						[]byte{0xf2, 0x61, 0x90, 0xf8, 0x52, 0x2c, 0xfa, 0x8, 0xfa, 0xc0, 0xb2, 0x54, 0x4c, 0x78, 0x54, 0xad, 0x7, 0x5, 0xe6, 0xec, 0x43, 0xd3, 0xba, 0xed, 0x5f, 0x25, 0x9e, 0x7, 0x1c, 0x63, 0xa6, 0x97}),
				},
				Smr: &trillian.SignedMapRoot{
					TimestampNanos: 1502231274356137738,
					RootHash:       []byte{0x6d, 0xfd, 0x5f, 0xda, 0xbe, 0x1c, 0x77, 0xf4, 0x90, 0xa7, 0x10, 0x65, 0xc3, 0x4, 0xfe, 0xdb, 0x53, 0x98, 0x17, 0x5e, 0xbd, 0x24, 0x74, 0x6f, 0xc4, 0xa5, 0x47, 0xbe, 0xa4, 0x24, 0x42, 0xcf},
					Metadata: &trillian.MapperMetadata{
						HighestFullyCompletedSeq: 1,
					},
					Signature: &sigpb.DigitallySigned{
						HashAlgorithm:      4,
						SignatureAlgorithm: 3,
						Signature:          []byte{0x30, 0x46, 0x2, 0x21, 0x0, 0xaf, 0xff, 0x55, 0xb0, 0x9f, 0xdc, 0x3c, 0x72, 0x17, 0x39, 0x84, 0xda, 0x67, 0x5b, 0x59, 0x71, 0x6, 0xea, 0x7b, 0xfa, 0x34, 0x12, 0xb1, 0xcd, 0x90, 0xf, 0x21, 0x15, 0xee, 0x61, 0x7d, 0xed, 0x2, 0x21, 0x0, 0x8c, 0xf5, 0x51, 0xd, 0x9e, 0xf8, 0x2a, 0x22, 0xb8, 0xfd, 0xc1, 0xee, 0x5d, 0x14, 0xbe, 0x87, 0x7c, 0x6e, 0x2a, 0x8a, 0x8f, 0x8f, 0xae, 0x98, 0x8b, 0x10, 0xc2, 0x3e, 0xb4, 0xbc, 0x3c, 0x12},
					},
					MapId:       8245331544573830053,
					MapRevision: 2,
				},
			},
		},
	} {
		err := v.VerifyGetEntryResponse(ctx, tc.userID, tc.appID, tc.trusted, tc.in)
		if got, want := err != nil, tc.wantErr; got != want {
			t.Errorf("VerifyGetEntryResponse(%v, %v, %v, %v): %v, wantErr %v",
				tc.userID, tc.appID, tc.trusted, tc.in, got, want)
		}
	}
}
