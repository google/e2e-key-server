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

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"
)

func TestBasicValidateCreds(t *testing.T) {
	auth := NewFake()
	for _, tc := range []struct {
		ctx            context.Context
		requiredUserID string
		want           error
	}{
		{context.Background(), "foo", ErrMissingAuth},
		{auth.NewContext("foo"), "bar", ErrWrongUser},
		{auth.NewContext("foo"), "foo", nil},
	} {
		// Convert outgoing to incoming context.
		outMD, _ := metadata.FromOutgoingContext(tc.ctx)
		inCtx := metadata.NewIncomingContext(tc.ctx, outMD)

		if got, want := auth.ValidateCreds(inCtx, tc.requiredUserID), tc.want; got != want {
			t.Errorf("ValidateCreds(%v, %v): %v, want %v", tc.ctx, tc.requiredUserID, got, want)
		}
	}
}
