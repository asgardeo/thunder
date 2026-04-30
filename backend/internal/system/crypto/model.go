/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package crypto

import (
	gocrypto "crypto"
	"crypto/tls"
)

// KeyRef identifies a cryptographic key by its ID.
type KeyRef struct {
	KeyID string
}

// Algorithm represents a cryptographic algorithm identifier.
type Algorithm string

const (
	// AlgorithmRS256 represents RSA PKCS1v15 signature with SHA-256 (JWA, RFC 7518).
	AlgorithmRS256 Algorithm = "RS256"
	// AlgorithmRS512 represents RSA PKCS1v15 signature with SHA-512 (JWA, RFC 7518).
	AlgorithmRS512 Algorithm = "RS512"
	// AlgorithmPS256 represents RSA-PSS signature with SHA-256 (JWA, RFC 7518).
	AlgorithmPS256 Algorithm = "PS256"
	// AlgorithmES256 represents ECDSA signature with SHA-256 (JWA, RFC 7518).
	AlgorithmES256 Algorithm = "ES256"
	// AlgorithmES384 represents ECDSA signature with SHA-384 (JWA, RFC 7518).
	AlgorithmES384 Algorithm = "ES384"
	// AlgorithmES512 represents ECDSA signature with SHA-512 (JWA, RFC 7518).
	AlgorithmES512 Algorithm = "ES512"
	// AlgorithmEdDSA represents EdDSA signature algorithm (JWA, RFC 7518).
	AlgorithmEdDSA Algorithm = "EdDSA"
	// AlgorithmRSAOAEP256 represents RSA-OAEP key encryption with SHA-256 (JWA, RFC 7518).
	AlgorithmRSAOAEP256 Algorithm = "RSA-OAEP-256"
	// AlgorithmAESGCM represents AES-GCM symmetric authenticated encryption.
	AlgorithmAESGCM Algorithm = "AES-GCM"
)

// PublicKeyFilter specifies criteria for filtering public keys in GetPublicKeys.
type PublicKeyFilter struct {
	KeyID     string
	Algorithm Algorithm
}

// PublicKeyInfo describes a public key returned by GetPublicKeys.
type PublicKeyInfo struct {
	KeyID      string
	Algorithm  Algorithm
	PublicKey  gocrypto.PublicKey
	Thumbprint string
}

// TLSMaterial holds the TLS certificate material for a key reference.
type TLSMaterial struct {
	Certificate tls.Certificate
}
