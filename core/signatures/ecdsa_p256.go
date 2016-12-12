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

package signatures

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/asn1"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"math/big"

	"github.com/benlaurie/objecthash/go/objecthash"
	"github.com/google/key-transparency/core/proto/ctmap"

	tpb "github.com/google/key-transparency/core/proto/keytransparency_v1_types"
)

// p256Signer generates signatures with a single key using ECDSA P256.
type p256Signer struct {
	privKey *ecdsa.PrivateKey
	keyID   string
	rand    io.Reader
}

func generateP256KeyPair(rand io.Reader) ([]byte, []byte, error) {
	p256Curve := elliptic.P256()
	sk, err := ecdsa.GenerateKey(p256Curve, rand)
	if err != nil {
		return nil, nil, err
	}
	skBytes, err := x509.MarshalECPrivateKey(sk)
	if err != nil {
		return nil, nil, err
	}
	pkBytes, err := x509.MarshalPKIXPublicKey(sk.Public())
	if err != nil {
		return nil, nil, err
	}
	return skBytes, pkBytes, nil
}

func generatePEMP256KeyPair(rand io.Reader) ([]byte, []byte, error) {
	skBytes, pkBytes, err := generateP256KeyPair(rand)
	if err != nil {
		return nil, nil, err
	}
	skPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: skBytes,
		},
	)
	pkPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pkBytes,
		},
	)
	return skPEM, pkPEM, nil
}

// newP256Signer creates a signer object from a private key.
func newP256Signer(rand io.Reader, pk crypto.Signer) (Signer, error) {
	var privKey *ecdsa.PrivateKey
	switch pkType := pk.(type) {
	case *ecdsa.PrivateKey:
		params := *(pkType.Params())
		if params != *elliptic.P256().Params() {
			return nil, ErrPointNotOnCurve
		}
		if !elliptic.P256().IsOnCurve(pkType.X, pkType.Y) {
			return nil, ErrPointNotOnCurve
		}
		privKey = pkType
	default:
		return nil, ErrWrongKeyType
	}

	id, err := keyID(&privKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return &p256Signer{
		privKey: privKey,
		keyID:   id,
		rand:    rand,
	}, nil
}

// Sign generates a digital signature object.
func (s *p256Signer) Sign(data interface{}) (*ctmap.DigitallySigned, error) {
	j, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	hash := objecthash.CommonJSONHash(string(j))

	var ecSig struct {
		R, S *big.Int
	}
	ecSig.R, ecSig.S, err = ecdsa.Sign(s.rand, s.privKey, hash[:])
	if err != nil {
		log.Print("signature generation failed")
		return nil, ErrSign
	}
	sig, err := asn1.Marshal(ecSig)
	if err != nil {
		log.Print("failed to marshal ECDSA signature")
		return nil, ErrSign
	}
	return &ctmap.DigitallySigned{
		HashAlgorithm: ctmap.DigitallySigned_SHA256,
		SigAlgorithm:  ctmap.DigitallySigned_ECDSA,
		Signature:     sig,
	}, nil
}

// PublicKey returns the signer public key as tpb.PublicKey proto
// message.
func (s *p256Signer) PublicKey() (*tpb.PublicKey, error) {
	return publicKey(&s.privKey.PublicKey)
}

// KeyID returns the ID of the associated public key.
func (s *p256Signer) KeyID() string {
	return s.keyID
}

// p256Verifier verifies signatures using ECDSA P256.
type p256Verifier struct {
	pubKey *ecdsa.PublicKey
	keyID  string
}

// newP256Verifier creates a verifier from a ECDSA public key.
func newP256Verifier(pk *ecdsa.PublicKey) (Verifier, error) {
	params := *(pk.Params())
	if params != *elliptic.P256().Params() {
		return nil, ErrPointNotOnCurve
	}
	if !elliptic.P256().IsOnCurve(pk.X, pk.Y) {
		return nil, ErrPointNotOnCurve
	}
	id, err := keyID(pk)
	if err != nil {
		return nil, err
	}

	return &p256Verifier{
		pubKey: pk,
		keyID:  id,
	}, nil
}

// Verify checks the digital signature associated applied to data.
func (s *p256Verifier) Verify(data interface{}, sig *ctmap.DigitallySigned) error {
	if sig == nil {
		return ErrMissingSig
	}
	if sig.HashAlgorithm != ctmap.DigitallySigned_SHA256 {
		log.Print("not SHA256 hash algorithm")
		return ErrVerify
	}
	if sig.SigAlgorithm != ctmap.DigitallySigned_ECDSA {
		log.Print("not ECDSA signature algorithm")
		return ErrVerify
	}

	j, err := json.Marshal(data)
	if err != nil {
		log.Print("json.Marshal failed")
		return ErrVerify
	}
	hash := objecthash.CommonJSONHash(string(j))

	var ecdsaSig struct {
		R, S *big.Int
	}
	rest, err := asn1.Unmarshal(sig.Signature, &ecdsaSig)
	if err != nil {
		log.Print("failed to unmarshal ECDSA signature")
		return ErrVerify
	}
	if len(rest) != 0 {
		log.Print("extra data found after signature")
		return ErrVerify
	}

	if !ecdsa.Verify(s.pubKey, hash[:], ecdsaSig.R, ecdsaSig.S) {
		log.Print("failed to verify ECDSA signature")
		return ErrVerify
	}
	return nil
}

// PublicKey returns the verifier public key as tpb.PublicKey proto
// message.
func (s *p256Verifier) PublicKey() (*tpb.PublicKey, error) {
	return publicKey(s.pubKey)
}

// KeyID returns the ID of the associated public key.
func (s *p256Verifier) KeyID() string {
	return s.keyID
}

func publicKey(k *ecdsa.PublicKey) (*tpb.PublicKey, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(k)
	if err != nil {
		return nil, err
	}
	return &tpb.PublicKey{
		KeyType: &tpb.PublicKey_EcdsaVerifyingP256{
			EcdsaVerifyingP256: pubBytes,
		},
	}, nil
}
