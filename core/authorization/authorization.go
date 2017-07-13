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

// Package authorization defines the authorization interface of Key Transparency.
package authorization

import (
	"github.com/google/keytransparency/core/authentication"

	authzpb "github.com/google/keytransparency/core/proto/authorization"
)

// Authorization authorizes access to RPCs.
type Authorization interface {
	// IsAuthorized verifies that the identity issuing the call
	// (from ctx) is authorized to carry the given permission.
	IsAuthorized(ctx *authentication.SecurityContext, mapID, appID int64,
		userID string, permission authzpb.Permission) error
}
