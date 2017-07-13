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

package authentication

import (
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func TestBasicValidateCreds(t *testing.T) {
	auth := NewFake()
	for _, tc := range []struct {
		description    string
		cred           credentials.PerRPCCredentials
		requiredUserID string
		want           error
	}{
		{"missing authentication", nil, "foo", ErrMissingAuth},
		{"working case", GetFakeCredential("foo"), "foo", nil},
	} {
		// Build context by adding the credential information.
		var inCtx context.Context
		if tc.cred == nil {
			inCtx = metadata.NewIncomingContext(context.Background(), nil)
		} else {
			md, _ := tc.cred.GetRequestMetadata(context.Background())
			inCtx = metadata.NewIncomingContext(context.Background(), metadata.New(md))
		}

		sctx, err := auth.ValidateCreds(inCtx)
		if got, want := err, tc.want; got != want {
			t.Errorf("%v: ValidateCreds()=(_, %v), want (_, %v)", tc.description, got, want)
		}
		if err != nil {
			continue
		}
		if got, want := sctx.Identity(), tc.requiredUserID; got != want {
			t.Errorf("%v: sctx.Identity()=%v, want %v", tc.description, got, want)
		}
	}
}
