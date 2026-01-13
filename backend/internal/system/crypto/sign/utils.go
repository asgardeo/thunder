/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

// Package sign provides utilities for digital signature generation and verification.
package sign

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"hash"

	"github.com/asgardeo/thunder/internal/system/log"
)

// Error definitions
var (
	ErrUnsupportedAlgorithm = errors.New("unsupported signature algorithm")
	ErrInvalidPrivateKey    = errors.New("invalid private key type for algorithm")
	ErrInvalidPublicKey     = errors.New("invalid public key type for algorithm")
	ErrInvalidSignature     = errors.New("signature verification failed")
)

// Generate takes a byte array, hashes it according to the specified algorithm,
// and returns the digital signature using the provided private key.
func Generate(data []byte, alg SignAlgorithm, privateKey crypto.PrivateKey) ([]byte, error) {
	// Hash the data according to the algorithm
	hashed, hashFunc := hashData(data, alg)

	// Sign based on the algorithm
	switch alg {
	case RSASHA256, RSASHA512:
		return newRSASign(hashed, hashFunc, privateKey)
	case ECDSASHA256, ECDSASHA384, ECDSASHA512:
		return newECDSASign(hashed, privateKey)
	case ED25519:
		return newED25519Sign(data, privateKey)
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}

// Verify takes a byte array, signature, algorithm, and public key, then verifies
// that the signature is valid for the given data.
func Verify(data []byte, signature []byte, alg SignAlgorithm, publicKey crypto.PublicKey) error {
	// Hash the data according to the algorithm
	hashed, hashFunc := hashData(data, alg)

	// Verify based on the algorithm
	switch alg {
	case RSASHA256, RSASHA512:
		return verifyRSA(hashed, signature, hashFunc, publicKey)
	case ECDSASHA256, ECDSASHA384, ECDSASHA512:
		return verifyECDSA(hashed, signature, publicKey)
	case ED25519:
		return verifyED25519(data, signature, publicKey)
	default:
		return ErrUnsupportedAlgorithm
	}
}

// hashData hashes the input data using the hash function specified by the algorithm.
// For ED25519, no pre-hashing is performed and the original data is returned.
func hashData(data []byte, alg SignAlgorithm) ([]byte, crypto.Hash) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "SignatureUtil"))
	var h hash.Hash
	var hashFunc crypto.Hash

	switch alg {
	case RSASHA256, ECDSASHA256:
		h = sha256.New()
		hashFunc = crypto.SHA256
	case RSASHA512, ECDSASHA512:
		h = sha512.New()
		hashFunc = crypto.SHA512
	case ECDSASHA384:
		h = sha512.New384()
		hashFunc = crypto.SHA384
	case ED25519:
		// ED25519 performs internal hashing, no pre-hash needed
		return data, crypto.Hash(0)
	default:
		logger.Error("Unsupported signature algorithm: %s", log.String("algorithm", string(alg)))
		return nil, crypto.Hash(0)
	}

	h.Write(data)
	return h.Sum(nil), hashFunc
}

// newRSASign creates a digital signature using PKCS1v15
func newRSASign(hashed []byte, hashFunc crypto.Hash, privateKey crypto.PrivateKey) ([]byte, error) {
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidPrivateKey
	}

	// Use PKCS1v15 for compatibility
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, hashFunc, hashed)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verifyRSA verifies an RSA signature
func verifyRSA(hashed, signature []byte, hashFunc crypto.Hash, publicKey crypto.PublicKey) error {
	rsaPub, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return ErrInvalidPublicKey
	}

	err := rsa.VerifyPKCS1v15(rsaPub, hashFunc, hashed, signature)
	if err != nil {
		return ErrInvalidSignature
	}

	return nil
}

// newECDSASign creates a digital signature using ECDSA
func newECDSASign(hashed []byte, privateKey crypto.PrivateKey) ([]byte, error) {
	ecdsaKey, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidPrivateKey
	}

	// SignASN1 returns the signature in ASN.1 DER format
	signature, err := ecdsa.SignASN1(rand.Reader, ecdsaKey, hashed)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verifyECDSA verifies an ECDSA signature
func verifyECDSA(hashed, signature []byte, publicKey crypto.PublicKey) error {
	ecdsaPub, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return ErrInvalidPublicKey
	}

	valid := ecdsa.VerifyASN1(ecdsaPub, hashed, signature)
	if !valid {
		return ErrInvalidSignature
	}
	return nil
}

// newED25519Sign creates a digital signature using ED25519
func newED25519Sign(data []byte, privateKey crypto.PrivateKey) ([]byte, error) {
	ed25519Key, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, ErrInvalidPrivateKey
	}

	signature := ed25519.Sign(ed25519Key, data)
	return signature, nil
}

// verifyED25519 verifies an ED25519 signature
func verifyED25519(data, signature []byte, publicKey crypto.PublicKey) error {
	ed25519Pub, ok := publicKey.(ed25519.PublicKey)
	if !ok {
		return ErrInvalidPublicKey
	}

	valid := ed25519.Verify(ed25519Pub, data, signature)
	if !valid {
		return ErrInvalidSignature
	}
	return nil
}
