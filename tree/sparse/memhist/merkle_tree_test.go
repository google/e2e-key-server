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

package memhist

import (
	"bytes"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/gdbelvin/e2e-key-server/tree"
	"github.com/gdbelvin/e2e-key-server/tree/sparse"
)

const (
	testCommitmentTimestamp = 1
)

var (
	AllZeros = strings.Repeat("0", 256)
	// validTreeLeaves contains valid leaves that will be added to the test
	// valid tree. All these leaves will be added to the test tree using the
	// addValidLeaves function.
	validTreeLeaves = []Leaf{
		{0, "0000000000000000000000000000000000000000000000000000000000000000", 1},
		{0, "0000000000000000000000000000000000000000000000000000000000000001", 2},
		{0, "8000000000000000000000000000000000000000000000000000000000000001", 3},
		{1, "8000000000000000000000000000000000000000000000000000000000000001", 4},
		{1, "0000000000000000000000000000000000000000000000000000000000000001", 5},
	}
)

type Leaf struct {
	epoch        int64
	hindex       string
	commitmentTS int64
}

type Env struct {
	m *Tree
}

func NewEnv(t *testing.T) *Env {
	m := New()
	// Adding few leaves with commitment timestamps to the tree.
	addValidLeaves(t, m)

	return &Env{m}
}

func addValidLeaves(t *testing.T, m *Tree) {
	for i, leaf := range validTreeLeaves {
		index := hexToBytes(t, leaf.hindex)
		if err := m.AddLeaf([]byte{}, leaf.epoch, index, leaf.commitmentTS); err != nil {
			t.Fatalf("%v: AddLeaf(-, %v, %v)=%v", i, leaf.epoch, leaf.hindex, err)
		}
	}
}

func hexToBytes(t testing.TB, h string) []byte {
	result, err := hex.DecodeString(h)
	if err != nil {
		t.Fatalf("DecodeString(%v)= %v", h, err)
	}
	return result
}

func TestWriteRead(t *testing.T) {
	t.Parallel()
	m := New()
	ctx := context.Background()
	tests := []struct {
		hindex string
		data   []byte
	}{
		{"0000000000000000000000000000000000000000000000000000000000000000", []byte("test data")},
		{"F000000000000000000000000000000000000000000000000000000000000000", []byte("test foo")},
		{"2000000000000000000000000000000000000000000000000000000000000000", []byte("test bar")},
		{"C000000000000000000000000000000000000000000000000000000000000000", []byte("test zed")},
	}
	for _, test := range tests {
		index := hexToBytes(t, test.hindex)
		if err := m.WriteLeaf(ctx, index, test.data); err != nil {
			t.Errorf("WriteLeaf(%v, %v)=%v)", index, test.data, err)
		}
		data, err := m.ReadLeaf(ctx, index)
		if err != nil {
			t.Errorf("ReadLeaf(%v)=%v)", index, err)
		}
		if got, want, equal := data, test.data, bytes.Equal(data, test.data); !equal {
			t.Errorf("ReadLeaf(%v)=%v, want %v", index, got, want)
		}
	}
}

func TestAddRoot(t *testing.T) {
	t.Parallel()

	m := New()
	tests := []struct {
		epoch int64
		code  codes.Code
	}{
		{0, codes.OK},
		{1, codes.OK},
		{2, codes.OK},
		{3, codes.OK},
		{3, codes.OK},
		{1, codes.FailedPrecondition},
		{10, codes.FailedPrecondition},
	}
	for i, test := range tests {
		err := m.AddRoot(test.epoch)
		if got, want := grpc.Code(err), test.code; got != want {
			t.Errorf("Test[%v]: addRoot(%v)=%v, want %v", i, test.epoch, got, want)
		}
	}
}

func TestAddExistingLeaf(t *testing.T) {
	t.Parallel()

	env := NewEnv(t)
	tests := []struct {
		leaf Leaf
		code codes.Code
	}{
		{validTreeLeaves[4], codes.OK},
	}
	for i, test := range tests {
		index := hexToBytes(t, test.leaf.hindex)
		err := env.m.AddLeaf([]byte{}, test.leaf.epoch, index, test.leaf.commitmentTS)
		if got, want := grpc.Code(err), codes.OK; got != want {
			t.Errorf("Test[%v]: AddLeaf(_, %v, %v)=%v, want %v, %v",
				i, test.leaf.epoch, test.leaf.hindex, got, want, err)
		}
	}
}

var letters = []rune("01234567890abcdef")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func BenchmarkAddLeaf(b *testing.B) {
	m := New()
	var epoch int64
	for i := 0; i < b.N; i++ {
		hindex := randSeq(64)
		index := hexToBytes(b, hindex)
		err := m.AddLeaf([]byte{}, epoch, index, testCommitmentTimestamp)
		if got, want := grpc.Code(err), codes.OK; got != want {
			b.Errorf("%v: AddLeaf(_, %v, %v)=%v, want %v",
				i, epoch, hindex, got, want)
		}
	}
}

func BenchmarkAddLeafAdvanceEpoch(b *testing.B) {
	m := New()
	var epoch int64
	for i := 0; i < b.N; i++ {
		hindex := randSeq(64)
		index := hexToBytes(b, hindex)
		epoch++
		err := m.AddLeaf([]byte{}, epoch, index, testCommitmentTimestamp)
		if got, want := grpc.Code(err), codes.OK; got != want {
			b.Errorf("%v: AddLeaf(_, %v, %v)=%v, want %v",
				i, epoch, hindex, got, want)
		}
	}
}

func BenchmarkAudit(b *testing.B) {
	m := New()
	var epoch int64
	items := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		hindex := randSeq(64)
		index := hexToBytes(b, hindex)
		items[i] = hindex
		err := m.AddLeaf([]byte{}, epoch, index, testCommitmentTimestamp)
		if got, want := grpc.Code(err), codes.OK; got != want {
			b.Errorf("%v: AddLeaf(_, %v, %v)=%v, want %v",
				i, epoch, hindex, got, want)
		}
	}
	for _, v := range items {
		index := hexToBytes(b, v)
		m.AuditPath(epoch, index)
	}
}

func TestPushDown(t *testing.T) {
	t.Parallel()

	index := hexToBytes(t, AllZeros)
	n := &node{bindex: tree.BitString(index)}
	if !n.leaf() {
		t.Errorf("node without children was a leaf")
	}
	n.pushDown()
	if n.leaf() {
		t.Errorf("node was still a leaf after push")
	}
	if !n.left.leaf() {
		t.Errorf("new child was not a leaf after push")
	}
}

func TestCreateBranch(t *testing.T) {
	t.Parallel()

	index := hexToBytes(t, AllZeros)
	n := &node{bindex: tree.BitString(index)}
	n.createBranch("0")
	if n.left == nil {
		t.Errorf("nil branch after create")
	}
}

func TestCreateBranchCOW(t *testing.T) {
	t.Parallel()

	la := &node{epoch: 0, bindex: "0", depth: 1}
	lb := &node{epoch: 0, bindex: "1", depth: 1}
	r0 := &node{epoch: 0, bindex: "", left: la, right: lb}
	r1 := &node{epoch: 1, bindex: "", left: la, right: lb}

	var e0 int64
	var e1 int64 = 1

	r1.createBranch("0")
	if got, want := r1.left.epoch, e1; got != want {
		t.Errorf("r1.left.epoch = %v, want %v", got, want)
	}
	if got, want := r0.left.epoch, e0; got != want {
		t.Errorf("r0.left.epoch = %v, want %v", got, want)
	}
}

func TestAuditDepth(t *testing.T) {
	t.Parallel()

	env := NewEnv(t)

	tests := []struct {
		leaf  Leaf
		depth int
	}{
		{validTreeLeaves[0], 256},
		{validTreeLeaves[1], 256},
		{validTreeLeaves[2], 1},
		{validTreeLeaves[3], 1},
		{validTreeLeaves[4], 256},
	}

	for i, test := range tests {
		index := hexToBytes(t, test.leaf.hindex)
		audit, _, err := env.m.AuditPath(test.leaf.epoch, index)
		if got, want := grpc.Code(err), codes.OK; got != want {
			t.Errorf("Test[%v]: AuditPath(_, %v, %v)=%v, want %v",
				i, test.leaf.epoch, test.leaf.hindex, got, want)
		}
		if got, want := len(audit), test.depth; got != want {
			for j, a := range audit {
				fmt.Println(j, ": ", a)
			}
			t.Errorf("Test[%v]: len(audit(%v, %v))=%v, want %v", i, test.leaf.epoch, test.leaf.hindex, got, want)
		}
	}
}

func TestAuditNeighors(t *testing.T) {
	t.Parallel()

	m := New()
	tests := []struct {
		epoch         int64
		hindex        string
		emptyNeighors []bool
	}{
		{0, "0000000000000000000000000000000000000000000000000000000000000000", []bool{}},
		{0, "F000000000000000000000000000000000000000000000000000000000000000", []bool{false}},
		{0, "2000000000000000000000000000000000000000000000000000000000000000", []bool{false, true, false}},
		{0, "C000000000000000000000000000000000000000000000000000000000000000", []bool{false, true, false}},
	}
	for i, test := range tests {
		index := hexToBytes(t, test.hindex)
		// Insert.
		err := m.AddLeaf([]byte{}, test.epoch, index, testCommitmentTimestamp)
		if got, want := grpc.Code(err), codes.OK; got != want {
			t.Errorf("Test[%v]: AddLeaf(_, %v, %v)=%v, want %v",
				i, test.epoch, test.hindex, got, want)
		}
		// Verify audit path.
		audit, _, err := m.AuditPath(test.epoch, index)
		if got, want := grpc.Code(err), codes.OK; got != want {
			t.Errorf("Test[%v]: AuditPath(_, %v, %v)=%v, want %v",
				i, test.epoch, test.hindex, got, want)
		}
		if got, want := len(audit), len(test.emptyNeighors); got != want {
			for j, a := range audit {
				fmt.Println(j, ": ", a)
			}
			t.Errorf("Test[%v]: len(audit(%v, %v))=%v, want %v", i, test.epoch, test.hindex, got, want)
		}
		for j, v := range test.emptyNeighors {
			// Starting from the leaf's neighbor, going to the root.
			depth := len(audit) - j
			nstr := tree.NeighborString(tree.BitString(index)[:depth])
			value := sparse.EmptyLeafValue(nstr)
			if got, want := bytes.Equal(audit[j], value), v; got != want {
				t.Errorf("Test[%v]: AuditPath(%v)[%v]=%v, want %v", i, test.hindex, j, got, want)
			}
		}
	}
}

func TestLongestPrefixMatch(t *testing.T) {
	t.Parallel()

	env := NewEnv(t)

	// Get commitment timestamps.
	tests := []struct {
		leaf            Leaf
		outCommitmentTS int64
		code            codes.Code
	}{
		// Get commitment timestamps of all added leaves. Ordering doesn't matter
		{validTreeLeaves[3], validTreeLeaves[3].commitmentTS, codes.OK},
		{validTreeLeaves[0], validTreeLeaves[0].commitmentTS, codes.OK},
		{validTreeLeaves[4], validTreeLeaves[4].commitmentTS, codes.OK},
		{validTreeLeaves[1], validTreeLeaves[1].commitmentTS, codes.OK},
		{validTreeLeaves[2], validTreeLeaves[2].commitmentTS, codes.OK},
		// Add custom testing leaves.
		// Invalid index lengh.
		{Leaf{1, "8000", 0}, 0, codes.InvalidArgument},
		// Not found due to missing epoch.
		{Leaf{3, "8000000000000000000000000000000000000000000000000000000000000001", 0}, 0, codes.InvalidArgument},
		// Found another leaf.
		{Leaf{1, "8000000000000000000000000000000000000000000000000000000000000002", 0}, 4, codes.OK},
		// Found empty branch.
		{Leaf{0, "0000000000000000000000000000000000000000000000000000000000000002", 0}, 0, codes.OK},
	}
	for i, test := range tests {
		index := hexToBytes(t, test.leaf.hindex)
		_, commitmentTS, err := env.m.AuditPath(test.leaf.epoch, index)
		if gotc, wantc, gote, wante := commitmentTS, test.outCommitmentTS, grpc.Code(err), test.code; gotc != wantc || gote != wante {
			t.Errorf("Test[%v]: LongestPrefixMatch(%v, %v)=(%v, %v), want (%v, %v), err = %v",
				i, test.leaf.epoch, test.leaf.hindex, gotc, gote, wantc, wante, err)
		}
	}
}

func TestRoot(t *testing.T) {
	t.Parallel()

	env := NewEnv(t)

	tests := []struct {
		epoch int64
		code  codes.Code
	}{
		{0, codes.OK},
		{1, codes.OK},
		{5, codes.NotFound},
	}

	for i, test := range tests {
		r := env.m.Root(test.epoch)
		if got, want := len(r) == 0, test.code == codes.NotFound; got != want {
			t.Errorf("%v: Root(%v)=%v, want %v", i, test.epoch, r, test.code)
		}
	}
}

func TestFromNeighbors(t *testing.T) {
	t.Parallel()
	m := New()

	for i, test := range validTreeLeaves {
		index := hexToBytes(t, test.hindex)
		data := []byte("hi")
		if err := m.AddLeaf(data, test.epoch, index, test.commitmentTS); err != nil {
			t.Fatalf("%v: AddLeaf(-, %v, %v)=%v", i, test.epoch, test.hindex, err)
		}
		neighbors, _, err := m.AuditPath(test.epoch, index)

		tmp, err := FromNeighbors(neighbors, index, data)
		if err != nil {
			t.Fatalf("FromNeighbors()= %v", err)
		}

		got, want := tmp.Root(0), m.Root(test.epoch)
		if ok := hmac.Equal(got, want); !ok {
			t.Errorf("%v: FromNeighbors().Root=%v, want %v", i, got, want)
		}
	}
}
