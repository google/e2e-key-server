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

// Package authorization contains the authorization module implementation.
package authorization

import (
	"context"
	"testing"

	"github.com/google/keytransparency/impl/authentication"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/google/keytransparency/core/api/v1/keytransparency_go_proto"
	authzpb "github.com/google/keytransparency/impl/authorization/authz_go_proto"
)

const (
	testUser = "user@example.com"
	l1       = "r1"
	l2       = "r2"
	l3       = "r3"
	l4       = "r4"
	l5       = "r5"
	admin1   = "admin1@example.com"
	admin2   = "admin2@example.com"
	admin3   = "admin3@example.com"
	admin4   = "admin4@example.com"
	res1     = "domains/1/apps/1"
	res2     = "domains/1/apps/2"
	res3     = "domains/1/apps/3"
	res4     = "domains/1/apps/4"
)

var authz = AuthzPolicy{
	Policy: &authzpb.AuthorizationPolicy{
		Roles: map[string]*authzpb.AuthorizationPolicy_Role{
			l1: {
				Principals: []string{admin1},
			},
			l2: {
				Principals: []string{admin1, admin2},
			},
			l3: {
				Principals: []string{admin3},
			},
			l4: {},
		},
		ResourceToRoleLabels: map[string]*authzpb.AuthorizationPolicy_RoleLabels{
			res1: {
				Labels: []string{l1, l2},
			},
			res2: {
				Labels: []string{l3},
			},
			res3: {
				Labels: []string{l4},
			},
			res4: {
				Labels: []string{l5},
			},
		},
	},
}

func TestIsAuthorized(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		description string
		ctx         context.Context
		domainID    string
		appID       string
		userID      string
		wantCode    codes.Code
	}{
		{
			description: "self updating own profile",
			ctx:         authentication.WithOutgoingFakeAuth(ctx, testUser),
			domainID:    "1",
			appID:       "1",
			userID:      testUser,
		},
		{
			description: "other accessing profile, authorized with one role",
			ctx:         authentication.WithOutgoingFakeAuth(ctx, admin1),
			domainID:    "1",
			appID:       "1",
			userID:      "",
		},
		{
			description: "other accessing profile, authorized with multiple roles",
			ctx:         authentication.WithOutgoingFakeAuth(ctx, admin2),
			domainID:    "1",
			appID:       "1",
			userID:      "",
		},
		{
			description: "other accessing profile, authorized second resource",
			ctx:         authentication.WithOutgoingFakeAuth(ctx, admin3),
			domainID:    "1",
			appID:       "2",
			userID:      "",
		},
		{
			description: "not authorized, no resource label",
			ctx:         authentication.WithOutgoingFakeAuth(ctx, admin1),
			domainID:    "1",
			appID:       "10",
			userID:      "",
			wantCode:    codes.PermissionDenied,
		},
		{
			description: "not authorized, no label_to_role defined",
			ctx:         authentication.WithOutgoingFakeAuth(ctx, admin1),
			domainID:    "1",
			appID:       "4",
			userID:      "",
			wantCode:    codes.PermissionDenied,
		},
		{
			description: "not authorized, empty role definition",
			ctx:         authentication.WithOutgoingFakeAuth(ctx, admin1),
			domainID:    "1",
			appID:       "3",
			userID:      "",
			wantCode:    codes.PermissionDenied,
		},
		{
			description: "not authorized principal",
			ctx:         authentication.WithOutgoingFakeAuth(ctx, admin4),
			domainID:    "1",
			appID:       "1",
			userID:      "",
			wantCode:    codes.PermissionDenied,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			// Convert outgoing context to incoming context.
			inCtx := metautils.ExtractOutgoing(tc.ctx).ToIncoming(ctx)
			sctx, err := authentication.FakeAuthFunc(inCtx)
			if err != nil {
				t.Fatalf("FakeAuthFunc(): %v", err)
			}
			req := &pb.UpdateEntryRequest{
				DomainId: tc.domainID,
				AppId:    tc.appID,
				UserId:   tc.userID,
			}
			err = authz.Authorize(sctx, req)
			if got, want := status.Code(err), tc.wantCode; got != want {
				t.Errorf("IsAuthorized(): %v, want %v", err, want)
			}
		})
	}
}

func TestResouceLabel(t *testing.T) {
	for _, tc := range []struct {
		domainID string
		appID    string
		out      string
		wantCode codes.Code
	}{
		{domainID: "1", appID: "1", out: "domains/1/apps/1"},
		{domainID: "1", appID: "2", out: "domains/1/apps/2"},
		{domainID: "1", appID: "111", out: "domains/1/apps/111"},
		{domainID: "111", appID: "1", out: "domains/111/apps/1"},
		{domainID: "111", appID: "111", out: "domains/111/apps/111"},
		{domainID: "1/apps/1", wantCode: codes.InvalidArgument},
	} {
		label, err := resourceLabel(tc.domainID, tc.appID)
		if got, want := label, tc.out; got != want {
			t.Errorf("resourceLabel(%v, %v): %v, want %v", tc.domainID, tc.appID, got, want)
		}
		if got, want := status.Code(err), tc.wantCode; got != want {
			t.Errorf("resourceLabel(%v, %v): %v, want %v", tc.domainID, tc.appID, err, want)
		}
	}
}
