// Copyright 2020 Google Inc. All Rights Reserved.
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

// Package vrf implements various Verifiable Random Functions (VRF)
//
// A Verifiable Random Function (VRF) is the public-key version of a
// keyed cryptographic hash.  Only the holder of the private key can
// compute the hash, but anyone with public key can verify the
// correctness of the hash.  VRFs are useful for preventing enumeration
// of hash-based data structures.
//
// These algorithms are secure in the cryptographic random oracle model.
//
// Reference: https://tools.ietf.org/html/draft-irtf-cfrg-vrf-06
package vrf

type VRF interface {
	Params() *ECVRFParams

	// Prove returns proof pi that beta is the correct hash output.
	// beta is deterministic in the sense that it always
	// produces the same output beta given a pair of inputs (sk, alpha).
	Prove(sk *PrivateKey, alpha []byte) (pi []byte)

	// ProofToHash allows anyone to deterministically obtain the VRF hash
	// output beta directly from the proof value pi.
	//
	// ProofToHash should be run only on pi that is known to have been produced by Prove
	// Clients attempting to verify untrusted inputs should use Verify.
	ProofToHash(pi []byte) (beta []byte, err error)

	// Verify that beta is the correct VRF hash of alpha using PublicKey pub.
	Verify(pub *PublicKey, pi, alpha []byte) (beta []byte, err error)
}