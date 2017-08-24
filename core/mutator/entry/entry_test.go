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

package entry

import (
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/keytransparency/core/crypto/dev"
	"github.com/google/keytransparency/core/crypto/signatures"
	"github.com/google/keytransparency/core/crypto/signatures/factory"

	"github.com/golang/protobuf/proto"

	tpb "github.com/google/keytransparency/core/proto/keytransparency_v1_types"
	"github.com/google/trillian/crypto/sigpb"
)

const (
	// openssl ecparam -name prime256v1 -genkey -out p256-key.pem
	testPrivKey1 = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBoLpoKGPbrFbEzF/ZktBSuGP+Llmx2wVKSkbdAdQ+3JoAoGCCqGSM49
AwEHoUQDQgAE+xVOdphkfpEtl7OF8oCyvWw31dV4hnGbXDPbdFlL1nmayhnqyEfR
dXNlpBT2U9hXcSxliKI1rHrAJFDx3ncttA==
-----END EC PRIVATE KEY-----`
	// openssl ec -in p256-key.pem -pubout -out p256-pubkey.pem
	testPubKey1 = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE+xVOdphkfpEtl7OF8oCyvWw31dV4
hnGbXDPbdFlL1nmayhnqyEfRdXNlpBT2U9hXcSxliKI1rHrAJFDx3ncttA==
-----END PUBLIC KEY-----`
	// openssl ecparam -name prime256v1 -genkey -out p256-key.pem
	testPrivKey2 = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIGugtYzUjyysX/JtjAFA6K3SzgBSmNjog/3e//VWRLQQoAoGCCqGSM49
AwEHoUQDQgAEJKDbR4uyhSMXW80x02NtYRUFlMQbLOA+tLe/MbwZ69SRdG6Rx92f
9tbC6dz7UVsyI7vIjS+961sELA6FeR91lA==
-----END EC PRIVATE KEY-----`
	// openssl ec -in p256-key.pem -pubout -out p256-pubkey.pem
	testPubKey2 = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEJKDbR4uyhSMXW80x02NtYRUFlMQb
LOA+tLe/MbwZ69SRdG6Rx92f9tbC6dz7UVsyI7vIjS+961sELA6FeR91lA==
-----END PUBLIC KEY-----`
)

func createEntry(commitment []byte, pkeys []string) (*tpb.Entry, error) {
	authKeys := make([]*tpb.PublicKey, len(pkeys))
	for i, key := range pkeys {
		p, _ := pem.Decode([]byte(key))
		if p == nil {
			return nil, errors.New("no PEM block found")
		}
		authKeys[i] = &tpb.PublicKey{
			KeyType: &tpb.PublicKey_EcdsaVerifyingP256{
				EcdsaVerifyingP256: p.Bytes,
			},
		}
	}

	return &tpb.Entry{
		Commitment:     commitment,
		AuthorizedKeys: authKeys,
		Previous:       nil,
	}, nil
}

func prepareMutation(key []byte, newEntry *tpb.Entry, previous []byte, signers []signatures.Signer) (*tpb.SignedKV, error) {
	newEntry.Previous = previous
	entryData, err := proto.Marshal(newEntry)
	if err != nil {
		return nil, fmt.Errorf("Marshal(%v)=%v", newEntry, err)
	}
	kv := &tpb.KeyValue{
		Key:   key,
		Value: entryData,
	}

	// Populate signatures map.
	sigs := make(map[string]*sigpb.DigitallySigned)
	for _, signer := range signers {
		sig, err := signer.Sign(*kv)
		if err != nil {
			return nil, fmt.Errorf("signerSign() failed: %v", err)
		}
		sigs[signer.KeyID()] = sig
	}

	return &tpb.SignedKV{
		KeyValue:   kv,
		Signatures: sigs,
	}, nil
}

func signersFromPEMs(t *testing.T, keys [][]byte) []signatures.Signer {
	signatures.Rand = dev.Zeros
	signers := make([]signatures.Signer, 0, len(keys))
	for _, key := range keys {
		signer, err := factory.NewSignerFromPEM(key)
		if err != nil {
			t.Fatalf("NewSigner(): %v", err)
		}
		signers = append(signers, signer)
	}
	return signers
}

func TestFromLeafValue(t *testing.T) {
	entry := &tpb.Entry{Commitment: []byte{1, 2}}
	entryB, _ := proto.Marshal(entry)
	for i, tc := range []struct {
		leafVal []byte
		want    *tpb.Entry
		wantErr bool
	}{
		{[]byte{}, &tpb.Entry{}, false},          // empty leaf bytes -> return 'empty' proto, no error
		{nil, nil, false},                        // non-existing leaf -> return nil, no error
		{[]byte{2, 2, 2, 2, 2, 2, 2}, nil, true}, // no valid proto Message
		{entryB, entry, false},                   // valid leaf
	} {
		if got, _ := FromLeafValue(tc.leafVal); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("FromLeafValue(%v)=%v, _ , want %v", tc.leafVal, got, tc.want)
			t.Error(i)
		}
		if _, gotErr := FromLeafValue(tc.leafVal); (gotErr != nil) != tc.wantErr {
			t.Errorf("FromLeafValue(%v)=_, %v", tc.leafVal, gotErr)
		}
	}
}
