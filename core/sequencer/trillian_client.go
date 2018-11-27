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

package sequencer

import (
	"context"

	"github.com/golang/glog"
	"github.com/google/trillian/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/keytransparency/core/directory"

	tpb "github.com/google/trillian"
	tclient "github.com/google/trillian/client"
)

// Trillian contains Trillian gRPC clients and metadata about them.
type Trillian struct {
	directories directory.Storage
	logAdmin    tpb.TrillianAdminClient
	mapAdmin    tpb.TrillianAdminClient
	tmap        tpb.TrillianMapClient
	tlog        tpb.TrillianLogClient
}

// MapClient returns a verifying MapClient
func (t *Trillian) MapClient(ctx context.Context, dirID string) (*MapClient, error) {
	directory, err := t.directories.Read(ctx, dirID, false)
	if err != nil {
		glog.Errorf("directories.Read(%v): %v", dirID, err)
		return nil, status.Errorf(codes.Internal, "Cannot fetch directory info for %v", dirID)
	}
	mapTree, err := t.mapAdmin.GetTree(ctx, &tpb.GetTreeRequest{TreeId: directory.MapID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot fetch map info for %v: %v", dirID, err)
	}

	c, err := tclient.NewMapClientFromTree(t.tmap, mapTree)
	if err != nil {
		return nil, err
	}
	return &MapClient{MapClient: c}, nil
}

// LogClient returns a verifying LogClient.
func (t *Trillian) LogClient(ctx context.Context, dirID string) (*tclient.LogClient, error) {
	directory, err := t.directories.Read(ctx, dirID, false)
	if err != nil {
		glog.Errorf("directories.Read(%v): %v", dirID, err)
		return nil, status.Errorf(codes.Internal, "Cannot fetch directory info for %v", dirID)
	}
	logTree, err := t.logAdmin.GetTree(ctx, &tpb.GetTreeRequest{TreeId: directory.LogID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot fetch log info for %v: %v", dirID, err)
	}

	// Create verifying log client.
	trustedRoot := types.LogRootV1{} // TODO(gbelvin): Store and track trustedRoot.
	return tclient.NewFromTree(t.tlog, logTree, trustedRoot)
}

// MapClient interacts with the Trillian Map and verifies its responses.
type MapClient struct {
	*tclient.MapClient
}

// SetLeaves creates a new map revision and returns its verified root.
// TODO(gbelvin): Move to Trillian Map client.
func (c *MapClient) SetLeaves(ctx context.Context, leaves []*tpb.MapLeaf, metadata []byte) (*types.MapRootV1, error) {
	// Set new leaf values.
	setResp, err := c.Conn.SetLeaves(ctx, &tpb.SetMapLeavesRequest{
		MapId:    c.MapID,
		Leaves:   leaves,
		Metadata: metadata,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "tmap.SetLeaves(): %v", err)
	}
	mapRoot, err := c.VerifySignedMapRoot(setResp.GetMapRoot())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "VerifySignedMapRoot(): %v", err)
	}
	return mapRoot, nil
}

// GetAndVerifyLatestMapRoot verifies and returns the latest map root.
func (c *MapClient) GetAndVerifyLatestMapRoot(ctx context.Context) (*tpb.SignedMapRoot, *types.MapRootV1, error) {
	rootResp, err := c.Conn.GetSignedMapRoot(ctx, &tpb.GetSignedMapRootRequest{MapId: c.MapID})
	if err != nil {
		return nil, nil, status.Errorf(status.Code(err), "GetSignedMapRoot(%v): %v", c.MapID, err)
	}
	mapRoot, err := c.VerifySignedMapRoot(rootResp.GetMapRoot())
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "VerifySignedMapRoot(%v): %v", c.MapID, err)
	}
	return rootResp.GetMapRoot(), mapRoot, nil
}

// GetAndVerifyMapRootByRevision verifies and returns a specific map root.
func (c *MapClient) GetAndVerifyMapRootByRevision(ctx context.Context,
	rev int64) (*tpb.SignedMapRoot, *types.MapRootV1, error) {
	req := &tpb.GetSignedMapRootByRevisionRequest{
		MapId:    c.MapID,
		Revision: rev,
	}
	resp, err := c.Conn.GetSignedMapRootByRevision(ctx, req)
	if err != nil {
		return nil, nil, status.Errorf(status.Code(err), "GetSignedMapRootByRevision(%v, %v): %v", c.MapID, rev, err)
	}
	rawMapRoot := resp.GetMapRoot()
	mapRoot, err := c.VerifySignedMapRoot(rawMapRoot)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "VerifySignedMapRoot(): %v", err)
	}
	return rawMapRoot, mapRoot, nil
}
