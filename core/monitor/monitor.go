// Copyright 2017 Google Inc. All Rights Reserved.
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

package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/keytransparency/core/client"
	"github.com/google/keytransparency/core/monitorstorage"

	"github.com/google/trillian"
	"github.com/google/trillian/types"

	"github.com/golang/glog"

	pb "github.com/google/keytransparency/core/api/v1/keytransparency_proto"
	tclient "github.com/google/trillian/client"
	tcrypto "github.com/google/trillian/crypto"
)

// Monitor holds the internal state for a monitor accessing the mutations API
// and for verifying its responses.
type Monitor struct {
	cli         *client.Client
	signer      *tcrypto.Signer
	trusted     types.LogRootV1
	logVerifier *tclient.LogVerifier
	mapVerifier *tclient.MapVerifier
	store       monitorstorage.Interface
}

// NewFromDomain produces a new monitor from a Domain object.
func NewFromDomain(cli pb.KeyTransparencyClient,
	config *pb.Domain,
	signer *tcrypto.Signer,
	store monitorstorage.Interface) (*Monitor, error) {
	logVerifier, err := tclient.NewLogVerifierFromTree(config.GetLog())
	if err != nil {
		return nil, fmt.Errorf("could not initialize log verifier: %v", err)
	}
	mapVerifier, err := tclient.NewMapVerifierFromTree(config.GetMap())
	if err != nil {
		return nil, fmt.Errorf("could not initialize map verifier: %v", err)
	}

	ktClient, err := client.NewFromConfig(cli, config)
	if err != nil {
		return nil, fmt.Errorf("could not create kt client: %v", err)
	}

	return New(ktClient, logVerifier, mapVerifier, signer, store)
}

// New creates a new instance of the monitor.
func New(cli *client.Client,
	logVerifier *tclient.LogVerifier,
	mapVerifier *tclient.MapVerifier,
	signer *tcrypto.Signer,
	store monitorstorage.Interface) (*Monitor, error) {
	return &Monitor{
		cli:         cli,
		logVerifier: logVerifier,
		mapVerifier: mapVerifier,
		signer:      signer,
		store:       store,
	}, nil
}

// EpochPair is two adjacent epochs.
type EpochPair struct {
	A, B *pb.Epoch
}

// EpochPairs consumes epochs (0, 1, 2) and produces pairs (0,1), (1,2).
func EpochPairs(ctx context.Context, epochs <-chan *pb.Epoch, pairs chan<- EpochPair) error {
	defer close(pairs)
	var epochA *pb.Epoch
	for epoch := range epochs {
		if epochA == nil {
			epochA = epoch
			continue
		}
		pair := EpochPair{
			A: epochA,
			B: epoch,
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case pairs <- pair:
		}
		epochA = epoch
	}
	return nil
}

// ProcessLoop continuously fetches mutations and processes them.
func (m *Monitor) ProcessLoop(ctx context.Context, domainID string, trusted types.LogRootV1) error {
	cctx, cancel := context.WithCancel(ctx)
	errc := make(chan error)
	epochs := make(chan *pb.Epoch)
	pairs := make(chan EpochPair)

	go func(ctx context.Context) {
		errc <- m.cli.StreamEpochs(ctx, domainID, int64(trusted.TreeSize), epochs)
	}(cctx)
	go func(ctx context.Context) {
		errc <- EpochPairs(ctx, epochs, pairs)
	}(cctx)
	defer cancel()

	for pair := range pairs {
		mapRoot, err := m.mapVerifier.VerifySignedMapRoot(pair.B.GetSmr())
		if err != nil {
			return err
		}
		revision := int64(mapRoot.Revision)
		mutations, err := m.cli.EpochMutations(ctx, pair.B)
		if err != nil {
			return err
		}

		var smr *trillian.SignedMapRoot
		var errList []error
		if errs := m.VerifyEpochMutations(pair.A, pair.B, trusted, mutations); len(errs) > 0 {
			glog.Infof("Epoch %v did not verify: %v", revision, errs)
			errList = errs
		} else {
			// Sign if successful.
			smr, err = m.signer.SignMapRoot(mapRoot)
			if err != nil {
				return err
			}
		}

		// Save result.
		if err := m.store.Set(revision, &monitorstorage.Result{
			Smr:    smr,
			Seen:   time.Now(),
			Errors: errList,
		}); err != nil {
			return fmt.Errorf("monitorstorage.Set(%v, _): %v", revision, err)
		}
	}
	errA := <-errc
	errB := <-errc
	if err := errA; err != nil {
		return err
	}
	return errB
}

// VerifyEpochMutations validates that epochA + mutations = epochB.
func (m *Monitor) VerifyEpochMutations(epochA, epochB *pb.Epoch, trusted types.LogRootV1, mutations []*pb.MutationProof) []error {
	mapRoot, err := m.mapVerifier.VerifySignedMapRoot(epochB.GetSmr())
	if err != nil {
		return []error{err}
	}
	revision := int64(mapRoot.Revision)
	if errs := m.VerifyEpoch(epochB, trusted); len(errs) > 0 {
		glog.Errorf("Invalid Epoch %v: %v", revision, errs)
		return errs
	}

	// Fetch Previous root.
	if errs := m.verifyMutations(mutations, epochA.GetSmr(), epochB.GetSmr()); len(errs) > 0 {
		glog.Errorf("Invalid Epoch %v Mutations: %v", revision, errs)
		return errs
	}
	return nil

}
