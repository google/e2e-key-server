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

// Package monitor implements the monitor service. A monitor repeatedly polls a
// key-transparency server's Mutations API and signs Map Roots if it could
// reconstruct
// clients can query.
package monitor

import (
	"bytes"
	"errors"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/google/trillian"

	tcrypto "github.com/google/trillian/crypto"

	cmon "github.com/google/keytransparency/core/monitor"
	ktpb "github.com/google/keytransparency/core/proto/keytransparency_v1_types"
	mopb "github.com/google/keytransparency/core/proto/monitor_v1_types"

	mupb "github.com/google/keytransparency/impl/proto/mutation_v1_service"
)

// Each page contains pageSize profiles. Each profile contains multiple
// keys. Assuming 2 keys per profile (each of size 2048-bit), a page of
// size 16 will contain about 8KB of data.
const pageSize = 16

var (
	// ErrNothingProcessed occurs when the monitor did not process any mutations /
	// smrs yet.
	ErrNothingProcessed = errors.New("did not process any mutations yet")
)

// Server holds internal state for the monitor server.
type Server struct {
	client     mupb.MutationServiceClient
	pollPeriod time.Duration

	monitor        *cmon.Monitor
	signer         *tcrypto.Signer
	proccessedSMRs []*mopb.GetMonitoringResponse
}

// New creates a new instance of the monitor server.
func New(cli mupb.MutationServiceClient,
	signer *tcrypto.Signer,
	logTree, mapTree *trillian.Tree,
	poll time.Duration) *Server {
	return &Server{
		client:     cli,
		pollPeriod: poll,
		// TODO(ismail) use domain info to properly init. the monitor:
		monitor:        &cmon.Monitor{},
		signer:         signer,
		proccessedSMRs: make([]*mopb.GetMonitoringResponse, 256),
	}
}

// StartPolling initiates polling and processing mutations every pollPeriod.
func (s *Server) StartPolling() error {
	t := time.NewTicker(s.pollPeriod)
	for now := range t.C {
		glog.Infof("Polling: %v", now)
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		if _, err := s.pollMutations(ctx); err != nil {
			glog.Errorf("pollMutations(_): %v", err)
		}
	}
	return nil
}

// GetSignedMapRoot returns the latest valid signed map root the monitor
// observed. Additionally, the response contains additional data necessary to
// reproduce errors on failure.
//
// Returns the signed map root for the latest epoch the monitor observed. If
// the monitor could not reconstruct the map root given the set of mutations
// from the previous to the current epoch it won't sign the map root and
// additional data will be provided to reproduce the failure.
func (s *Server) GetSignedMapRoot(ctx context.Context, in *mopb.GetMonitoringRequest) (*mopb.GetMonitoringResponse, error) {
	if len(s.proccessedSMRs) > 0 {
		return s.proccessedSMRs[len(s.proccessedSMRs)-1], nil
	}
	return nil, ErrNothingProcessed
}

// GetSignedMapRootByRevision works similar to GetSignedMapRoot but returns
// the monitor's result for a specific map revision.
//
// Returns the signed map root for the specified epoch the monitor observed.
// If the monitor could not reconstruct the map root given the set of
// mutations from the previous to the current epoch it won't sign the map root
// and additional data will be provided to reproduce the failure.
func (s *Server) GetSignedMapRootByRevision(ctx context.Context, in *mopb.GetMonitoringRequest) (*mopb.GetMonitoringResponse, error) {
	// TODO(ismail): implement by revision API
	return nil, grpc.Errorf(codes.Unimplemented, "GetSignedMapRoot is unimplemented")
}

func (s *Server) pollMutations(ctx context.Context, opts ...grpc.CallOption) ([]*ktpb.Mutation, error) {
	// TODO(ismail): move everything that does not rely on impl packages (e.g.,
	// the client here) into core:
	resp, err := s.client.GetMutations(ctx, &ktpb.GetMutationsRequest{
		PageSize: pageSize,
		Epoch:    s.nextRevToQuery(),
	}, opts...)
	if err != nil {
		return nil, err
	}

	if got, want := resp.GetSmr(), s.lastSeenSMR(); bytes.Equal(got.GetRootHash(), want.GetRootHash()) &&
		got.GetMapRevision() == want.GetMapRevision() {
		// We already processed this SMR. Do not update seen SMRs. Do not scroll
		// pages for further mutations. Return empty mutations list.
		glog.Infof("Already processed this SMR with revision %v. Continuing.", got.GetMapRevision())
		return nil, nil
	}

	mutations, err := s.pageMutations(ctx, resp, opts...)
	if err != nil {
		glog.Errorf("s.pageMutations(_): %v", err)
		return nil, err
	}

	// TODO(Ismail): let the verification method in core directly return the response
	monitorResp := s.monitor.VerifyResponse(resp, mutations)
	// Update seen/processed signed map roots:
	s.proccessedSMRs = append(s.proccessedSMRs, monitorResp)

	return mutations, nil
}

// pageMutations iterates/pages through all mutations in the case there were
// more then maximum pageSize mutations in between epochs.
// It will modify the passed GetMutationsResponse resp.
func (s *Server) pageMutations(ctx context.Context, resp *ktpb.GetMutationsResponse,
	opts ...grpc.CallOption) ([]*ktpb.Mutation, error) {
	ms := make([]*ktpb.Mutation, pageSize*2)
	ms = append(ms, resp.GetMutations()...)

	// Query all mutations in the current epoch
	for resp.GetNextPageToken() != "" {
		req := &ktpb.GetMutationsRequest{PageSize: pageSize}
		resp, err := s.client.GetMutations(ctx, req, opts...)
		if err != nil {
			return nil, err
		}
		ms = append(ms, resp.GetMutations()...)
	}
	return ms, nil
}

func (s *Server) lastSeenSMR() *trillian.SignedMapRoot {
	if len(s.proccessedSMRs) > 0 {
		return s.proccessedSMRs[len(s.proccessedSMRs)-1].GetSmr()
	}
	return nil
}

func (s *Server) nextRevToQuery() int64 {
	smr := s.lastSeenSMR()
	if smr == nil {
		return 1
	}
	return smr.GetMapRevision() + 1
}
