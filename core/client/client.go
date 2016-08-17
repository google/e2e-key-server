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

// Client for communicating with the Key Server.
// Implements verification and convenience functions.

package client

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/key-transparency/core/commitments"
	"github.com/google/key-transparency/core/signatures"
	"github.com/google/key-transparency/core/tree/sparse"
	tv "github.com/google/key-transparency/core/tree/sparse/verifier"
	"github.com/google/key-transparency/core/vrf"

	"github.com/benlaurie/objecthash/go/objecthash"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/google/key-transparency/core/proto/keytransparency_v1"
)

const (
	// Each page contains pageSize profiles. Each profile contains multiple
	// keys. Assuming 2 keys per profile (each of size 2048-bit), a page of
	// size 16 will contain about 8KB of data.
	defaultPageSize = int32(16)
	// The delay before the client retry updating the profile if ErrRetry is
	// received.
	retryDelay = 3 * time.Second
	// TODO: Public key of signer.
	// TODO: Public keys of trusted monitors.
	// TODO: Domain to API URL resolution.
)

var (
	// ErrRetry occurs when an update request has been submitted, but the
	// results of the udpate are not visible on the server yet. The client
	// must retry until the request is visible.
	ErrRetry = errors.New("update not present on server yet")
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
	cli          pb.KeyTransparencyServiceClient
	vrf          vrf.PublicKey
	RetryCount   int
	treeVerifier *tv.Verifier
	verifier     *signatures.Verifier
	log          LogVerifier
}

// New creates a new client.
func New(client pb.KeyTransparencyServiceClient, vrf vrf.PublicKey, verifier *signatures.Verifier, log LogVerifier) *Client {
	return &Client{
		cli:          client,
		vrf:          vrf,
		RetryCount:   1,
		treeVerifier: tv.New(sparse.CONIKSHasher),
		verifier:     verifier,
		log:          log,
	}
}

// GetEntry returns an entry if it exists, and nil if it does not.
func (c *Client) GetEntry(ctx context.Context, userID string, opts ...grpc.CallOption) (*pb.Profile, error) {
	// Error, ctx is not being passed
	e, err := c.cli.GetEntry(ctx, &pb.GetEntryRequest{
		UserId: userID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	if err := c.verifyGetEntryResponse(userID, e); err != nil {
		return nil, err
	}

	// Empty case.
	if e.GetCommitted() == nil {
		return nil, nil
	}

	profile := new(pb.Profile)
	if err := proto.Unmarshal(e.GetCommitted().Data, profile); err != nil {
		log.Printf("Error unmarshaling profile: %v", err)
		return nil, err
	}
	return profile, nil
}

// ListHistory returns a list of profiles starting and ending at given epochs.
// It also filters out all identical consecutive profiles.
func (c *Client) ListHistory(ctx context.Context, userID string, startEpoch, endEpoch int64, opts ...grpc.CallOption) ([]*pb.Profile, error) {
	currentProfile := []byte(nil)
	profiles := make([]*pb.Profile, 0, endEpoch-startEpoch+1)

	// Setup loop-related variables.
	pageSize := defaultPageSize
	epoch := startEpoch
	for epoch <= endEpoch {
		resp, err := c.cli.ListEntryHistory(ctx, &pb.ListEntryHistoryRequest{
			UserId:   userID,
			Start:    epoch,
			PageSize: pageSize,
		}, opts...)
		if err != nil {
			return nil, err
		}

		// Iterate of the responses in reverse order.
		for i := 0; i < len(resp.GetValues()); i++ {
			// If empty profile, skip.
			if resp.GetValues()[i].GetCommitted() == nil {
				continue
			}

			pData := resp.GetValues()[i].GetCommitted().Data
			profile := new(pb.Profile)
			if err := proto.Unmarshal(pData, profile); err != nil {
				log.Printf("Error unmarshaling profile: %v", err)
				return nil, err
			}
			// Ignore the extracted profile if it is nil or similar
			// to the current one. Nil profiles can occur in epochs
			// before the user updates his profile for the first
			// time.
			if bytes.Equal(currentProfile, pData) {
				continue
			}

			// Verify the response.
			err = c.verifyGetEntryResponse(userID, resp.GetValues()[i])
			if err != nil {
				return nil, err
			}

			// Append the slice and update currentProfile.
			profiles = append(profiles, profile)
			currentProfile = pData
		}

		epoch = resp.NextStart
	}

	return profiles, nil
}

// Update creates an UpdateEntryRequest for a user, attempt to submit it multiple
// times depending on RetryCount.
func (c *Client) Update(ctx context.Context, userID string, profile *pb.Profile, opts ...grpc.CallOption) (*pb.UpdateEntryRequest, error) {
	getResp, err := c.cli.GetEntry(ctx, &pb.GetEntryRequest{UserId: userID}, opts...)
	if err != nil {
		return nil, err
	}

	// Extract index from a prior GetEntry call.
	index := c.vrf.Index(getResp.Vrf)
	prevEntry := new(pb.Entry)
	if err := proto.Unmarshal(getResp.GetLeafProof().LeafData, prevEntry); err != nil {
		return nil, fmt.Errorf("Error unmarshaling Entry from leaf proof: %v", err)
	}

	// Commit to profile.
	profileData, err := proto.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("Unexpected profile marshaling error: %v", err)
	}
	commitment, committed, err := commitments.Commit(userID, profileData)
	if err != nil {
		return nil, err
	}

	// Create new Entry.
	entry := &pb.Entry{
		Commitment:     commitment,
		AuthorizedKeys: prevEntry.AuthorizedKeys,
	}

	// Sign Entry.
	entryData, err := proto.Marshal(entry)
	if err != nil {
		return nil, err
	}
	kv := &pb.KeyValue{
		Key:   index[:],
		Value: entryData,
	}
	kvData, err := proto.Marshal(kv)
	if err != nil {
		return nil, err
	}
	previous := objecthash.ObjectHash(getResp.GetLeafProof().LeafData)
	signedkv := &pb.SignedKV{
		KeyValue:   kvData,
		Signatures: nil, // TODO: Apply Signatures.
		Previous:   previous[:],
	}

	// Send request.
	req := &pb.UpdateEntryRequest{
		UserId: userID,
		EntryUpdate: &pb.EntryUpdate{
			Update:    signedkv,
			Committed: committed,
		},
	}

	err = c.Retry(ctx, req)
	// Retry submitting until an incluion proof is returned.
	for i := 0; err == ErrRetry && i < c.RetryCount; i++ {
		time.Sleep(retryDelay)
		err = c.Retry(ctx, req)
	}
	return req, err
}

// Retry will take a pre-fabricated request and send it again.
func (c *Client) Retry(ctx context.Context, req *pb.UpdateEntryRequest) error {
	updateResp, err := c.cli.UpdateEntry(ctx, req)
	if err != nil {
		return err
	}

	// Check if the response is a replay.
	kv := new(pb.KeyValue)
	if err := proto.Unmarshal(req.GetEntryUpdate().GetUpdate().KeyValue, kv); err != nil {
		return fmt.Errorf("Error unmarshaling KeyValue: %v", err)
	}
	got := updateResp.GetProof().GetLeafProof().LeafData
	if bytes.Equal(got, kv.Value) {
		log.Printf("Retry(%v) Matched", req.UserId)
		return nil
	}

	return ErrRetry
	// TODO: Update previous entry pointer
}
