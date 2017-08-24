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

package entry

import (
	"github.com/benlaurie/objecthash/go/objecthash"
	"github.com/golang/protobuf/proto"
	"github.com/google/keytransparency/core/crypto/commitments"
	"github.com/google/keytransparency/core/crypto/signatures"
	"github.com/google/keytransparency/core/mutator"
	tpb "github.com/google/keytransparency/core/proto/keytransparency_v1_types"
	"github.com/google/trillian/crypto/sigpb"
)

// Mutation provides APIs for manipulating entries.
type Mutation struct {
	userID, appID string
	index         []byte
	data, nonce   []byte

	prevEntry *tpb.Entry
	entry     *tpb.Entry
}

// NewMutation creates a mutation object from a previous value which can be modified.
// To create a new value:
// - Create a new mutation for a user starting with the previous value with NewMutation.
// - Change the value with SetCommitment and ReplaceAuthorizedKeys.
// - Finalize the changes and create the mutation with SerializeAndSign.
func NewMutation(oldValue, index []byte, userID, appID string) (*Mutation, error) {
	prevEntry, err := FromLeafValue(oldValue)
	if err != nil {
		return nil, err
	}

	// TODO(ismail): Change this to plain sha256.
	hash := objecthash.ObjectHash(prevEntry)
	return &Mutation{
		userID:    userID,
		appID:     appID,
		index:     index,
		prevEntry: prevEntry,
		entry: &tpb.Entry{
			AuthorizedKeys: prevEntry.GetAuthorizedKeys(),
			Previous:       hash[:],
			Commitment:     prevEntry.GetCommitment(),
		},
	}, nil
}

// SetCommitment updates entry to be a commitment to data.
func (m *Mutation) SetCommitment(data []byte) error {
	// Commit to profile.
	commitmentNonce, err := commitments.GenCommitmentKey()
	if err != nil {
		return err
	}
	m.data = data
	m.nonce = commitmentNonce
	m.entry.Commitment = commitments.Commit(m.userID, m.appID, data, commitmentNonce)
	return nil
}

// ReplaceAuthorizedKeys sets authorized keys to pubkeys.
// pubkeys must contain at least one key.
func (m *Mutation) ReplaceAuthorizedKeys(pubkeys []*tpb.PublicKey) error {
	if got, want := len(pubkeys), 1; got < want {
		return mutator.ErrMissingKey
	}
	m.entry.AuthorizedKeys = pubkeys
	return nil
}

// SerializeAndSign produces the mutation.
func (m *Mutation) SerializeAndSign(signers []signatures.Signer) (*tpb.UpdateEntryRequest, error) {
	signedkv, err := m.sign(signers)
	if err != nil {
		return nil, err
	}

	// Check authorization.
	if err := verifyKeys(m.prevEntry.GetAuthorizedKeys(),
		m.entry.GetAuthorizedKeys(),
		signedkv.GetKeyValue(),
		signedkv.GetSignatures()); err != nil {
		return nil, err
	}

	return &tpb.UpdateEntryRequest{
		UserId: m.userID,
		AppId:  m.appID,
		EntryUpdate: &tpb.EntryUpdate{
			Update: signedkv,
			Committed: &tpb.Committed{
				Key:  m.nonce,
				Data: m.data,
			},
		},
	}, nil
}

// Sign produces the SignedKV
func (m *Mutation) sign(signers []signatures.Signer) (*tpb.SignedKV, error) {
	entryData, err := proto.Marshal(m.entry)
	if err != nil {
		return nil, err
	}
	kv := &tpb.KeyValue{
		Key:   m.index,
		Value: entryData,
	}

	sigs := make(map[string]*sigpb.DigitallySigned)
	for _, signer := range signers {
		sig, err := signer.Sign(kv)
		if err != nil {
			return nil, err
		}
		sigs[signer.KeyID()] = sig
	}

	return &tpb.SignedKV{
		KeyValue:   kv,
		Signatures: sigs,
	}, nil
}
