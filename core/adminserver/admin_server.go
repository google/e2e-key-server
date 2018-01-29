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

// Package adminserver contains the KeyTransparencyAdmin implementation
package adminserver

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/keytransparency/core/crypto/vrf/p256"
	"github.com/google/keytransparency/core/domain"
	"github.com/google/trillian/crypto/keys"
	"github.com/google/trillian/crypto/keys/der"
	"github.com/google/trillian/crypto/keyspb"
	"github.com/google/trillian/crypto/sigpb"

	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/client_golang/prometheus"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/google/keytransparency/core/api/v1/keytransparency_proto"
	tpb "github.com/google/trillian"
)

var (
	vrfKeySpec = &keyspb.Specification{
		Params: &keyspb.Specification_EcdsaParams{
			EcdsaParams: &keyspb.Specification_ECDSA{
				Curve: keyspb.Specification_ECDSA_P256,
			},
		},
	}
	logArgs = &tpb.CreateTreeRequest{
		Tree: &tpb.Tree{
			DisplayName:        "KT SMH Log",
			TreeState:          tpb.TreeState_ACTIVE,
			TreeType:           tpb.TreeType_LOG,
			HashStrategy:       tpb.HashStrategy_OBJECT_RFC6962_SHA256,
			SignatureAlgorithm: sigpb.DigitallySigned_ECDSA,
			HashAlgorithm:      sigpb.DigitallySigned_SHA256,
			MaxRootDuration:    ptypes.DurationProto(0 * time.Millisecond),
		},
		KeySpec: &keyspb.Specification{
			Params: &keyspb.Specification_EcdsaParams{
				EcdsaParams: &keyspb.Specification_ECDSA{
					Curve: keyspb.Specification_ECDSA_P256,
				},
			},
		},
	}
	mapArgs = &tpb.CreateTreeRequest{
		Tree: &tpb.Tree{
			DisplayName:        "KT Map",
			TreeState:          tpb.TreeState_ACTIVE,
			TreeType:           tpb.TreeType_MAP,
			HashStrategy:       tpb.HashStrategy_CONIKS_SHA512_256,
			SignatureAlgorithm: sigpb.DigitallySigned_ECDSA,
			HashAlgorithm:      sigpb.DigitallySigned_SHA256,
			MaxRootDuration:    ptypes.DurationProto(0 * time.Millisecond),
		},
		KeySpec: &keyspb.Specification{
			Params: &keyspb.Specification_EcdsaParams{
				EcdsaParams: &keyspb.Specification_ECDSA{
					Curve: keyspb.Specification_ECDSA_P256,
				},
			},
		},
	}
)

// Server implements pb.KeyTransparencyAdminServer
type Server struct {
	domains  domain.Storage
	logAdmin tpb.TrillianAdminClient
	mapAdmin tpb.TrillianAdminClient
	keygen   keys.ProtoGenerator
}

// New returns a KeyTransparencyAdmin implementation.
func New(domains domain.Storage, logAdmin, mapAdmin tpb.TrillianAdminClient, keygen keys.ProtoGenerator) *Server {
	s := &Server{
		domains:  domains,
		logAdmin: logAdmin,
		mapAdmin: mapAdmin,
		keygen:   keygen,
	}
	if err := prometheus.Register(
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "domain_count",
			Help: "Number of active (not deleted) domains.",
		},
			func() float64 {
				ctx := context.Background()
				showDeleted := false
				domains, _ := s.domains.List(ctx, showDeleted)
				return float64(len(domains))
			},
		),
	); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			log.Fatalf("Could not register domain_count gauge: %v", err)
		}
	}
	return s
}

// ListDomains produces a list of the configured domains
func (s *Server) ListDomains(ctx context.Context, in *pb.ListDomainsRequest) (*pb.ListDomainsResponse, error) {
	domains, err := s.domains.List(ctx, in.GetShowDeleted())
	if err != nil {
		return nil, err
	}

	resp := make([]*pb.Domain, 0, len(domains))
	for _, d := range domains {
		info, err := s.fetchDomain(ctx, d)
		if err != nil {
			return nil, err
		}
		resp = append(resp, info)

	}
	return &pb.ListDomainsResponse{
		Domains: resp,
	}, nil
}

// fetchDomain converts an adminstorage.Domain object into a pb.Domain object
// by fetching the relevant info from Trillian.
func (s *Server) fetchDomain(ctx context.Context, d *domain.Domain) (*pb.Domain, error) {
	logTree, err := s.logAdmin.GetTree(ctx, &tpb.GetTreeRequest{TreeId: d.LogID})
	if err != nil {
		return nil, err
	}
	mapTree, err := s.mapAdmin.GetTree(ctx, &tpb.GetTreeRequest{TreeId: d.MapID})
	if err != nil {
		return nil, err
	}
	return &pb.Domain{
		DomainId:    d.DomainID,
		Log:         logTree,
		Map:         mapTree,
		Vrf:         d.VRF,
		MinInterval: ptypes.DurationProto(d.MinInterval),
		MaxInterval: ptypes.DurationProto(d.MaxInterval),
		Deleted:     d.Deleted,
	}, nil
}

// GetDomain retrieves the domain info for a given domain.
func (s *Server) GetDomain(ctx context.Context, in *pb.GetDomainRequest) (*pb.Domain, error) {
	domain, err := s.domains.Read(ctx, in.GetDomainId(), in.GetShowDeleted())
	if err != nil {
		return nil, err
	}
	info, err := s.fetchDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// CreateDomain reachs out to Trillian to produce new trees.
func (s *Server) CreateDomain(ctx context.Context, in *pb.CreateDomainRequest) (*pb.Domain, error) {
	// TODO(gbelvin): Test whether the domain exists before creating trees.

	// Generate VRF key.
	wrapped, err := s.keygen(ctx, vrfKeySpec)
	if err != nil {
		return nil, fmt.Errorf("keygen: %v", err)
	}
	vrfPriv, err := p256.NewFromWrappedKey(ctx, wrapped)
	if err != nil {
		return nil, fmt.Errorf("NewFromWrappedKey(): %v", err)
	}
	vrfPublicPB, err := der.ToPublicProto(vrfPriv.Public())
	if err != nil {
		return nil, err
	}

	// Create Trillian keys.
	logTreeArgs := *logArgs
	logTreeArgs.Tree.Description = fmt.Sprintf("KT domain %s's SMH Log", in.GetDomainId())
	logTree, err := s.logAdmin.CreateTree(ctx, &logTreeArgs)
	if err != nil {
		return nil, fmt.Errorf("CreateTree(log): %v", err)
	}
	mapTreeArgs := *mapArgs
	mapTreeArgs.Tree.Description = fmt.Sprintf("KT domain %s's Map", in.GetDomainId())
	mapTree, err := s.mapAdmin.CreateTree(ctx, &mapTreeArgs)
	if err != nil {
		return nil, fmt.Errorf("CreateTree(map): %v", err)
	}
	minInterval, err := ptypes.Duration(in.MinInterval)
	if err != nil {
		return nil, fmt.Errorf("Duration(%v): %v", in.MinInterval, err)
	}
	maxInterval, err := ptypes.Duration(in.MaxInterval)
	if err != nil {
		return nil, fmt.Errorf("Duration(%v): %v", in.MaxInterval, err)
	}

	if err := s.domains.Write(ctx, &domain.Domain{
		DomainID:    in.GetDomainId(),
		MapID:       mapTree.TreeId,
		LogID:       logTree.TreeId,
		VRF:         vrfPublicPB,
		VRFPriv:     wrapped,
		MinInterval: minInterval,
		MaxInterval: maxInterval,
	}); err != nil {
		return nil, fmt.Errorf("adminstorage.Write(): %v", err)
	}
	return &pb.Domain{
		DomainId: in.GetDomainId(),
		Log:      logTree,
		Map:      mapTree,
		Vrf:      vrfPublicPB,
	}, nil
}

// DeleteDomain marks a domain as deleted, but does not immediately delete it.
func (s *Server) DeleteDomain(ctx context.Context, in *pb.DeleteDomainRequest) (*google_protobuf.Empty, error) {
	if err := s.domains.SetDelete(ctx, in.GetDomainId(), true); err != nil {
		return nil, err
	}
	return &google_protobuf.Empty{}, nil
}

// UndeleteDomain reactivates a deleted domain - provided that UndeleteDomain is called sufficiently soon after DeleteDomain.
func (s *Server) UndeleteDomain(ctx context.Context, in *pb.UndeleteDomainRequest) (*google_protobuf.Empty, error) {
	if err := s.domains.SetDelete(ctx, in.GetDomainId(), false); err != nil {
		return nil, err
	}
	return &google_protobuf.Empty{}, nil
}
