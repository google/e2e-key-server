// Copyright 2015 Google Inc. All Rights Reserved.
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

// Package proxy converts v1 API requests into v2 API calls.

package keyserver

import "github.com/yahoo/coname/vrf"

// Vuf is a mock verifiable unpredictable function.
func (s *Server) Vuf(userID string) (string, string, error) {
	sk := new([vrf.SecretKeySize]byte) // TODO: keep a persistent secret key
	idx := vrf.Compute([]byte(userID), sk)
	pf := vrf.Prove([]byte(userID), sk)
	return string(pf), string(idx), nil
}
