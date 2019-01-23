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

package client

import (
	"context"

	"fmt"

	"github.com/google/tink/go/tink"
	"google.golang.org/grpc"

	"github.com/google/keytransparency/core/mutator/entry"

	tpb "github.com/google/keytransparency/core/api/type/type_go_proto"
	pb "github.com/google/keytransparency/core/api/v1/keytransparency_go_proto"
)

// BatchCreateUser inserts mutations for new users that do not currently have entries.
// Calling BatchCreate for a user that already exists will produce no change.
func (c *Client) BatchCreateUser(ctx context.Context, users []*tpb.User,
	signers []tink.Signer, opts ...grpc.CallOption) error {
	// 1. Fetch user indexes
	userIDs := make([]string, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, u.UserId)
	}
	indexByUser, err := c.BatchVerifyGetUserIndex(ctx, userIDs)
	if err != nil {
		return err
	}

	mutations := make([]*entry.Mutation, 0, len(users))
	for _, u := range users {
		mutation := entry.NewMutation(indexByUser[u.UserId], c.DirectoryID, u.UserId)

		if err := mutation.SetCommitment(u.PublicKeyData); err != nil {
			return err
		}
		if len(u.AuthorizedKeys.Key) != 0 {
			if err := mutation.ReplaceAuthorizedKeys(u.AuthorizedKeys); err != nil {
				return err
			}
		}
		mutations = append(mutations, mutation)
	}
	return c.BatchQueueUserUpdate(ctx, mutations, signers, opts...)
}

// BatchQueueUserUpdate signs the mutations and sends them to the server.
func (c *Client) BatchQueueUserUpdate(ctx context.Context, mutations []*entry.Mutation,
	signers []tink.Signer, opts ...grpc.CallOption) error {
	updates := make([]*pb.EntryUpdate, 0, len(mutations))
	for _, m := range mutations {
		update, err := m.SerializeAndSign(signers)
		if err != nil {
			return err
		}
		updates = append(updates, update)
	}

	req := &pb.BatchQueueUserUpdateRequest{DirectoryId: c.DirectoryID, Updates: updates}
	_, err := c.cli.BatchQueueUserUpdate(ctx, req, opts...)
	return err
}

// BatchCreateMutation fetches the current index and value for a list of users and prepares mutations.
func (c *Client) BatchCreateMutation(ctx context.Context, users []*tpb.User) ([]*entry.Mutation, error) {
	userIDs := make([]string, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, u.UserId)
	}

	leavesByUserID, err := c.BatchVerifiedGetUser(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	mutations := make([]*entry.Mutation, 0, len(users))

	for _, u := range users {
		leaf, ok := leavesByUserID[u.UserId]
		if !ok {
			return nil, fmt.Errorf("no leaf found for %v", u.UserId)
		}
		index, err := c.Index(leaf.GetVrfProof(), c.DirectoryID, u.UserId)
		if err != nil {
			return nil, err
		}
		mutation := entry.NewMutation(index, c.DirectoryID, u.UserId)

		leafValue := leaf.MapInclusion.GetLeaf().GetLeafValue()
		if err := mutation.SetPrevious(leafValue, true); err != nil {
			return nil, err
		}

		if err := mutation.SetCommitment(u.PublicKeyData); err != nil {
			return nil, err
		}

		if len(u.AuthorizedKeys.Key) != 0 {
			if err := mutation.ReplaceAuthorizedKeys(u.AuthorizedKeys); err != nil {
				return nil, err
			}
		}
		mutations = append(mutations, mutation)
	}
	return mutations, nil
}
