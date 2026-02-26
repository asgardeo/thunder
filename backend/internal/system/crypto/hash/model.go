/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

package hash

// CredAlgorithm represents the supported credential hashing algorithms
type CredAlgorithm string

const (
	// SHA256 represents SHA-256 hashing algorithm
	SHA256 CredAlgorithm = "SHA256"
	// PBKDF2 represents PBKDF2 key derivation function
	PBKDF2 CredAlgorithm = "PBKDF2"
	// ARGON2ID represents Argon2id key derivation function
	ARGON2ID CredAlgorithm = "ARGON2ID"
)

// CredParameters holds the parameters for credential hashing algorithms
type CredParameters struct {
	Iterations  int
	KeySize     int
	Salt        string
	Memory      int
	Parallelism int
}

// Credential represents the credentials of a hashed value.
type Credential struct {
	Algorithm  CredAlgorithm
	Hash       string
	Parameters CredParameters
}

// HashAlgorithm represents the supported generic hash algorithms
type HashAlgorithm string

const (
	// GenericSHA256 represents SHA-256 hash algorithm
	GenericSHA256 HashAlgorithm = "SHA-256"
	// GenericSHA384 represents SHA-384 hash algorithm
	GenericSHA384 HashAlgorithm = "SHA-384"
	// GenericSHA512 represents SHA-512 hash algorithm
	GenericSHA512 HashAlgorithm = "SHA-512"
)
