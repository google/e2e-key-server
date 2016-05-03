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

package sparse

import (
	"testing"
)

func TestComputeNodeValues(t *testing.T) {
	tests := []struct {
		bindex    string
		leafHash  []byte
		neighbors [][]byte
		expected  []string
	}{
		{"0100", []byte(""), make([][]byte, 4), []string{"0100", "010", "01", "0", ""}},
	}
	for _, tc := range tests {
		actual := NodeValues(Coniks, tc.bindex, tc.leafHash, tc.neighbors)
		if got, want := len(actual), len(tc.expected); got != want {
			t.Errorf("len(%v)=%v, want %v", actual, got, want)
		}
	}
}
