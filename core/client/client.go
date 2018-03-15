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

// Package client is a client for communicating with the Key Server.
// It wraps the gRPC APIs and verifies all responses.
package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/google/keytransparency/core/crypto/signatures"
	"github.com/google/keytransparency/core/mutator"
	"github.com/google/keytransparency/core/mutator/entry"

	"github.com/google/trillian"
	"github.com/google/trillian/client/backoff"
	"github.com/google/trillian/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tpb "github.com/google/keytransparency/core/api/type/type_proto"
	pb "github.com/google/keytransparency/core/api/v1/keytransparency_proto"
	tcrypto "github.com/google/trillian/crypto"
)

const (
	// Each page contains pageSize profiles. Each profile contains multiple
	// keys. Assuming 2 keys per profile (each of size 2048-bit), a page of
	// size 16 will contain about 8KB of data.
	pageSize = 16
	// TODO: Public keys of trusted monitors.
)

var (
	// ErrRetry occurs when an update has been queued, but the
	// results of the update differ from the one requested.
	// This indicates that a separate update was in-flight while
	// this update was being submitted. To continue, the client
	// should make a fresh update and try again.
	ErrRetry = errors.New("client: update race condition - try again")
	// ErrWait occurs when an update has been queued, but no change has been
	// observed in the user's account yet.
	ErrWait = errors.New("client: update not present yet - wait some more")
	// ErrIncomplete occurs when the server indicates that requested epochs
	// are not available.
	ErrIncomplete = errors.New("incomplete account history")
	// Vlog is the verbose logger. By default it outputs to /dev/null.
	Vlog = log.New(ioutil.Discard, "", 0)
)

// Client is a helper library for issuing updates to the key server.
// Client Responsibilities
// - Trust Model:
// - - Trusted Monitors
// - - Verify last X days
// - Gossip - What is the current value of the root?
// -  - Gossip advancement: advance state between current and server.
// - Sender queries - Do queries match up against the gossip root?
// - - List trusted monitors.
// - Key Owner
// - - Periodically query own keys. Do they match the private keys I have?
// - - Sign key update requests.
type Client struct {
	*Verifier
	cli        pb.KeyTransparencyClient
	domainID   string
	mutator    mutator.Func
	RetryDelay time.Duration
	trusted    types.LogRootV1
}

// NewFromConfig creates a new client from a config
func NewFromConfig(ktClient pb.KeyTransparencyClient, config *pb.Domain) (*Client, error) {
	ktVerifier, err := NewVerifierFromDomain(config)
	if err != nil {
		return nil, err
	}
	return New(ktClient, config.DomainId, ktVerifier), nil
}

// New creates a new client.
// TODO(gbelvin): set retry delay.
func New(ktClient pb.KeyTransparencyClient,
	domainID string,
	ktVerifier *Verifier) *Client {
	return &Client{
		Verifier:   ktVerifier,
		cli:        ktClient,
		domainID:   domainID,
		mutator:    entry.New(),
		RetryDelay: 3 * time.Second,
	}
}

// updateTrusted sets the local reference for the latest SignedLogRoot if
// newTrusted is correctly signed and newer than the current stored root.
func (c *Client) updateTrusted(newTrusted *trillian.SignedLogRoot) error {
	r, err := tcrypto.VerifySignedLogRoot(c.logVerifier.PubKey, newTrusted)
	if err != nil {
		return err
	}

	if r.TimestampNanos <= c.trusted.TimestampNanos ||
		r.TreeSize < c.trusted.TreeSize {
		// Valid root, but it's older than the one we currently have.
		return nil
	}
	c.trusted = *r
	return nil
}

// GetEntry returns an entry if it exists, and nil if it does not.
func (c *Client) GetEntry(ctx context.Context, userID, appID string, opts ...grpc.CallOption) ([]byte, *trillian.SignedMapRoot, error) {
	e, err := c.VerifiedGetEntry(ctx, appID, userID)
	if err != nil {
		return nil, nil, err
	}
	// Empty case.
	if e.GetCommitted() == nil {
		return nil, e.GetSmr(), nil
	}

	return e.GetCommitted().GetData(), e.GetSmr(), nil
}

func min(x, y int32) int32 {
	if x < y {
		return x
	}
	return y
}

// ListHistory returns a list of profiles starting and ending at given epochs.
// It also filters out all identical consecutive profiles.
func (c *Client) ListHistory(ctx context.Context, userID, appID string, start, end int64, opts ...grpc.CallOption) (map[*trillian.SignedMapRoot][]byte, error) {
	if start < 0 {
		return nil, fmt.Errorf("start=%v, want >= 0", start)
	}
	var currentProfile []byte
	profiles := make(map[*trillian.SignedMapRoot][]byte)
	epochsReceived := int64(0)
	epochsWant := end - start + 1
	for epochsReceived < epochsWant {
		trustedSnapshot := c.trusted
		resp, err := c.cli.ListEntryHistory(ctx, &pb.ListEntryHistoryRequest{
			DomainId:      c.domainID,
			UserId:        userID,
			AppId:         appID,
			FirstTreeSize: int64(trustedSnapshot.TreeSize),
			Start:         start,
			PageSize:      min(int32((end-start)+1), pageSize),
		}, opts...)
		if err != nil {
			return nil, err
		}
		epochsReceived += int64(len(resp.GetValues()))

		for i, v := range resp.GetValues() {
			Vlog.Printf("Processing entry for %v, epoch %v", userID, start+int64(i))
			err = c.VerifyGetEntryResponse(ctx, c.domainID, appID, userID, trustedSnapshot, v)
			if err != nil {
				return nil, err
			}
			c.updateTrusted(v.GetLogRoot())

			// Compress profiles that are equal through time.  All
			// nil profiles before the first profile are ignored.
			profile := v.GetCommitted().GetData()
			if bytes.Equal(currentProfile, profile) {
				continue
			}

			// Append the slice and update currentProfile.
			profiles[v.GetSmr()] = profile
			currentProfile = profile
		}
		if resp.NextStart == 0 {
			break // No more data.
		}
		start = resp.NextStart // Fetch the next block of results.
	}

	if epochsReceived < epochsWant {
		return nil, ErrIncomplete
	}

	return profiles, nil
}

// Update creates an UpdateEntryRequest for a user,
// attempt to submit it multiple times depending until ctx times out.
// Returns context.DeadlineExceeded if ctx times out.
func (c *Client) Update(ctx context.Context, u *tpb.User, signers []signatures.Signer) (*entry.Mutation, error) {
	if got, want := u.DomainId, c.domainID; got != want {
		return nil, fmt.Errorf("u.DomainID: %v, want %v", got, want)
	}
	// 1. pb.User + ExistingEntry -> Mutation.
	m, err := c.newMutation(ctx, u)
	if err != nil {
		return nil, err
	}

	// 2. Queue Mutation.
	if err := c.QueueMutation(ctx, m, signers); err != nil {
		return nil, err
	}

	// 3. Wait for update.
	m, err = c.waitOnceForUserUpdate(ctx, m)
	for {
		switch {
		case err == ErrWait:
			// Try again.
		case err == ErrRetry:
			if err := c.QueueMutation(ctx, m, signers); err != nil {
				return nil, err
			}
		case status.Code(err) == codes.DeadlineExceeded:
			// Sometimes the timeout occurs during an rpc.
			// Convert to a standard context.DeadlineExceeded for consistent error handling.
			return m, context.DeadlineExceeded
		default:
			return m, err
		}
		m, err = c.waitOnceForUserUpdate(ctx, m)
	}
}

// QueueMutation signs an entry.Mutation and sends it to the server.
func (c *Client) QueueMutation(ctx context.Context, m *entry.Mutation, signers []signatures.Signer) error {
	req, err := m.SerializeAndSign(signers, int64(c.trusted.TreeSize))
	if err != nil {
		return fmt.Errorf("SerializeAndSign(): %v", err)
	}

	Vlog.Printf("Sending Update request...")
	// TODO(gdbelvin): Change name from UpdateEntry to QueueUpdate.
	_, err = c.cli.UpdateEntry(ctx, req)
	return err
}

// newMutation fetches the current index and value for a user and prepares a mutation.
func (c *Client) newMutation(ctx context.Context, u *tpb.User) (*entry.Mutation, error) {
	e, err := c.VerifiedGetEntry(ctx, u.AppId, u.UserId)
	if err != nil {
		return nil, err
	}
	oldLeaf := e.GetLeafProof().GetLeaf().GetLeafValue()
	Vlog.Printf("Got current entry...")

	index, err := c.Index(e.GetVrfProof(), u.DomainId, u.AppId, u.UserId)
	if err != nil {
		return nil, err
	}

	mutation := entry.NewMutation(index, u.DomainId, u.AppId, u.UserId)

	if err := mutation.SetPrevious(oldLeaf, true); err != nil {
		return nil, err
	}

	if err := mutation.SetCommitment(u.PublicKeyData); err != nil {
		return nil, err
	}

	if len(u.AuthorizedKeys) != 0 {
		if err := mutation.ReplaceAuthorizedKeys(u.AuthorizedKeys); err != nil {
			return nil, err
		}
	}

	return mutation, nil
}

// WaitForUserUpdate waits for the mutation to be applied or the context to timeout or cancel.
func (c *Client) WaitForUserUpdate(ctx context.Context, m *entry.Mutation) (*entry.Mutation, error) {
	for {
		m, err := c.waitOnceForUserUpdate(ctx, m)
		switch {
		case err == ErrWait:
			// Try again.
		case status.Code(err) == codes.DeadlineExceeded:
			// Sometimes the timeout occurs during an rpc.
			// Convert to a standard context.DeadlineExceeded for consistent error handling.
			return m, context.DeadlineExceeded
		default:
			return m, err
		}
	}
}

// waitOnceForUserUpdate waits for the STH to be updated, indicating the next epoch has been created,
// it then queries the current value for the user and checks it against the requested mutation.
// If the current value has not changed, WaitForUpdate returns ErrWait.
// If the current value has changed, but does not match the requested mutation,
// WaitForUpdate returns a new mutation, built with the current value and ErrRetry.
// If the current value matches the request, no mutation and no error are returned.
func (c *Client) waitOnceForUserUpdate(ctx context.Context, m *entry.Mutation) (*entry.Mutation, error) {
	if m == nil {
		return nil, fmt.Errorf("nil mutation")
	}
	// Wait for STH to change.
	if err := c.WaitForSTHUpdate(ctx, int64(c.trusted.TreeSize)+1); err != nil {
		return m, err
	}

	// GetEntry.
	e, err := c.VerifiedGetEntry(ctx, m.AppID, m.UserID)
	if err != nil {
		return m, err
	}
	Vlog.Printf("Got current entry...")

	// Verify.
	cntLeaf := e.GetLeafProof().GetLeaf().GetLeafValue()
	cntValue, err := entry.FromLeafValue(cntLeaf)
	if err != nil {
		return m, err
	}
	switch {
	case m.EqualsRequested(cntValue):
		return nil, nil
	case m.EqualsPrevious(cntValue):
		return m, ErrWait
	default:
		// Race condition: some change got in first.
		// Value has changed, but it's not what we asked for.
		// Retry based on new cntValue.

		// To break the tie between two devices that are fighting
		// each other, this error should be propagated back to the user.
		copyPreviousLeafData := false
		if err := m.SetPrevious(cntLeaf, copyPreviousLeafData); err != nil {
			return nil, fmt.Errorf("waitforupdate: SetPrevious(): %v", err)
		}
		return m, errors.New("client: update race condition - try again")
	}
}

// sthForRevision returns the minimum STH.TreeSize that will contain the map revision.
// Map revision N is stored at Log index N, the minimum TreeSize will be N+1.
func sthForRevision(revision int64) int64 {
	return revision + 1
}

// WaitForRevision waits until a given map revision is available.
func (c *Client) WaitForRevision(ctx context.Context, revision int64) error {
	return c.WaitForSTHUpdate(ctx, sthForRevision(revision))
}

// WaitForSTHUpdate blocks until the log root reported by the server has moved
// to at least treeSize or times out.
func (c *Client) WaitForSTHUpdate(ctx context.Context, treeSize int64) error {
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    10 * time.Second,
		Factor: 1.2,
		Jitter: true,
	}

	for {
		select {
		case <-time.After(b.Duration()):
			resp, err := c.cli.GetLatestEpoch(ctx, &pb.GetLatestEpochRequest{
				DomainId: c.domainID,
			})
			if err != nil {
				return err
			}
			if resp.GetLogRoot().TreeSize >= treeSize {
				return nil // We're done!
			}
			// The LogRoot is not updated yet.
			// Wait some more.
			continue

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
