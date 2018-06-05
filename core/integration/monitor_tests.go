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

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/keytransparency/core/fake"
	"github.com/google/keytransparency/core/monitor"
	"github.com/google/keytransparency/core/testutil"
	"github.com/google/tink/go/tink"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/trillian/crypto"
	"github.com/google/trillian/crypto/keys/pem"
	"github.com/google/trillian/types"

	tpb "github.com/google/keytransparency/core/api/type/type_go_proto"
)

const (
	monitorPrivKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAV7H3qRi/cj/w04vEQBFjLdhcXRbZR4ouT5zaAy1XUHoAoGCCqGSM49
AwEHoUQDQgAEqUDbATN2maGIm6YQLpjx67bYN1hxPPdF0VrPTZe36yQhH+GCwZQV
amFdON6OhjYnBmJWe4fVnbxny0PfpkvXtg==
-----END EC PRIVATE KEY-----`
)

// TestMonitor verifies that the monitor correctly verifies transitions between epochs.
func TestMonitor(ctx context.Context, env *Env, t *testing.T) {
	// setup monitor:
	privKey, err := pem.UnmarshalPrivateKey(monitorPrivKey, "")
	if err != nil {
		t.Fatalf("Couldn't create signer: %v", err)
	}
	signer := crypto.NewSHA256Signer(privKey)
	store := fake.NewMonitorStorage()
	mon, err := monitor.NewFromDomain(env.Cli, env.Domain, signer, store)
	if err != nil {
		t.Fatalf("Couldn't create monitor: %v", err)
	}

	// Setup a bunch of epochs with data to verify.
	for _, e := range []struct {
		epoch       int64
		signers     []*tink.KeysetHandle
		userUpdates []*tpb.User
	}{
		{
			epoch: 1,
		},
		{
			epoch:   2,
			signers: testutil.SignKeysetsFromPEMs(testPrivKey1),
			userUpdates: []*tpb.User{
				{
					DomainId:       env.Domain.DomainId,
					AppId:          "app1",
					UserId:         "alice@test.com",
					PublicKeyData:  []byte("alice-key1"),
					AuthorizedKeys: testutil.VerifyKeysetFromPEMs(testPubKey1).Keyset(),
				},
			},
		},
		{
			epoch:   3,
			signers: testutil.SignKeysetsFromPEMs(testPrivKey1),
			userUpdates: []*tpb.User{
				{
					DomainId:       env.Domain.DomainId,
					AppId:          "app1",
					UserId:         "bob@test.com",
					PublicKeyData:  []byte("bob-key1"),
					AuthorizedKeys: testutil.VerifyKeysetFromPEMs(testPubKey1).Keyset(),
				},
				{
					DomainId:       env.Domain.DomainId,
					AppId:          "app1",
					UserId:         "carol@test.com",
					PublicKeyData:  []byte("carol-key1"),
					AuthorizedKeys: testutil.VerifyKeysetFromPEMs(testPubKey1).Keyset(),
				},
			},
		},
	} {
		for _, u := range e.userUpdates {
			opts := env.CallOpts(u.UserId)
			cctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			defer cancel()
			m, err := env.Client.CreateMutation(cctx, u)
			if err != nil {
				t.Fatalf("CreateMutation(%v): %v", u.UserId, err)
			}
			if err := env.Client.QueueMutation(cctx, m, e.signers, opts...); err != nil {
				t.Fatalf("QueueMutation(%v): %v", u.UserId, err)
			}
		}

		if err := env.Receiver.FlushN(ctx, len(e.userUpdates)); err != nil {
			t.Fatalf("FlushN(%v): %v", len(e.userUpdates), err)
		}
		if err := env.Client.WaitForRevision(ctx, e.epoch); err != nil {
			t.Fatalf("WaitForRevision(): %v", err)
		}
	}

	trusted := types.LogRootV1{}
	cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	if err := mon.ProcessLoop(cctx, env.Domain.DomainId, trusted); err != context.DeadlineExceeded && status.Code(err) != codes.DeadlineExceeded {
		t.Errorf("Monitor could not process mutations: %v", err)
	}
	cancel()

	for i := int64(1); i < 4; i++ {
		mresp, err := store.Get(i)
		if err != nil {
			t.Errorf("Could not read monitoring response for epoch %v: %v", i, err)
			continue
		}
		for _, err := range mresp.Errors {
			t.Errorf("Got error: %v", err)
		}
	}
}
