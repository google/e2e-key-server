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

// package vrf defines the interface to a verifiable random function.
package vrf

// A VRF is a pseudorandom function f_k from a secret key k, such that that
// knowledge of k not only enables one to evaluate f_k at any point x, but also
// to provide an NP-proof that the value f_k(x) is indeed correct without
// compromising the unpredictability of f_k at any other point for which no
// such a proof was provided.
// http://ieeexplore.ieee.org/stamp/stamp.jsp?tp=&arnumber=814584

type PrivateKey interface {
	// VRF returns the output of f_k(m).
	Vrf(m []byte) ([32]byte, error)
	// Proof returns an NP-proof that f_k(m) is correct.
	Proof(m []byte) ([]byte, error)
}

type PublicKey interface {
	// Verify verifies the NP-proof supplied by Proof.
	Verify(m, proof []byte, vrf [32]byte) error
}
