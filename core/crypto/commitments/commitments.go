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

// Package commitments implements a cryptographic commitment.
//
// Commitment scheme is as follows:
// T = HMAC(fixedKey, "Key Transparency Commitment" || 16 byte nonce || message)
// message is defined as: len(userID) || userID || data
package commitments

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"errors"

	"golang.org/x/net/context"
)

const (
	// commitmentKeyLen should be robust against the birthday attack.
	// One commitment is given for each leaf node throughout time.
	commitmentKeyLen = 16 // 128 bits of security, supports 2^64 nodes.
	// prefix is a string used to make the commitments from this package unique.
	prefix = "Key Transparency Commitment"
)

var (
	hashAlgo = sha512.New512_256
	// key is publicly known random fixed key for use in the HMAC function.
	// This fixed key allows the commitment scheme to be modeled as a random oracle.
	fixedKey = []byte{0x19, 0x6e, 0x7e, 0x52, 0x84, 0xa7, 0xef, 0x93, 0x0e, 0xcb, 0x9a, 0x19, 0x78, 0x74, 0x97, 0x55}
	// ErrInvalidCommitment occurs when the commitment doesn't match the profile.
	ErrInvalidCommitment = errors.New("invalid commitment")
	// Rand is the PRNG reader. It can be overwritten in tests.
	Rand = rand.Reader
)

// Committer saves cryptographic commitments.
type Committer interface {
	// Write saves a cryptographic commitment and associated data.
	Write(ctx context.Context, commitment, data, nonce []byte) error
	// Read looks up a cryptograpic commitment and returns associated data.
	Read(ctx context.Context, commitment []byte) (data, nonce []byte, err error)
}

// Commit makes a cryptographic commitment under a specific userID to data.
func Commit(userID, appID string, data []byte) (commitment, nonce []byte, err error) {
	// Generate commitment nonce.
	nonce = make([]byte, commitmentKeyLen)
	if _, err := Rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	return createCommitment(userID, appID, data, nonce), nonce, nil
}

// Verify customizes a commitment with a userID.
func Verify(userID, appID string, commitment, data, nonce []byte) error {
	if got, want := createCommitment(userID, appID, data, nonce),
		commitment; !hmac.Equal(got, want) {
		return ErrInvalidCommitment
	}
	return nil
}

func createCommitment(userID, appID string, data, nonce []byte) []byte {
	mac := hmac.New(hashAlgo, fixedKey)
	mac.Write([]byte(prefix))
	mac.Write(nonce)

	// Message
	binary.Write(mac, binary.BigEndian, uint32(len(userID)))
	mac.Write([]byte(userID))
	binary.Write(mac, binary.BigEndian, uint32(len(appID)))
	mac.Write([]byte(appID))
	mac.Write(data)

	return mac.Sum(nil)
}
