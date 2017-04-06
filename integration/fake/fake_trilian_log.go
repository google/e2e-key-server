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

package fake

import (
	"context"
	"fmt"

	"github.com/google/trillian"
)

// FakeClient implements trillian/client.LogVerifier
type FakeClient struct {
	leaves []*trillian.LogLeaf
}

// NewFakeTrillianClient returns a client that mimicks a trillian log.
func NewFakeTrillianClient() *FakeClient {
	return &FakeClient{
		leaves: make([]*trillian.LogLeaf, 0),
	}
}

// AddLeaf adds a leaf to the log.
func (f *FakeClient) AddLeaf(ctx context.Context, data []byte) error {
	f.leaves = append(f.leaves, &trillian.LogLeaf{
		LeafValue: data,
	})
	return nil
}

// GetByIndex returns the requested leaf.
func (f *FakeClient) GetByIndex(ctx context.Context, index int64) (*trillian.LogLeaf, error) {
	if got, want := index, int64(len(f.leaves)); got > want {
		return nil, fmt.Errorf("Index out of range. Got %v, want %v", got, want)
	}
	return f.leaves[index], nil
}

// ListByIndex returns the set of requested leaves.
func (f *FakeClient) ListByIndex(ctx context.Context, start int64, count int64) ([]*trillian.LogLeaf, error) {
	if got, want := start+count, int64(len(f.leaves)); got > want {
		return nil, fmt.Errorf("Index out of range. Got %v, want %v", got, want)
	}
	return f.leaves[start : start+count], nil
}

// UpdateRoot fetches the latest signed tree root.
func (f *FakeClient) UpdateRoot(ctx context.Context) error {
	return nil
}

// Root returns the latest local copy of the signed log root.
func (f *FakeClient) Root() trillian.SignedLogRoot {
	return trillian.SignedLogRoot{
		TreeSize: int64(len(f.leaves)),
	}
}
