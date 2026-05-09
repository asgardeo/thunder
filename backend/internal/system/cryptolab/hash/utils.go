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

package hash

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
)

// GenerateThumbprint generates a SHA-256 thumbprint for the given data.
func GenerateThumbprint(data []byte) string {
	h := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(h[:])
}

// GenerateThumbprintFromString generates a SHA-256 thumbprint for the given string data.
func GenerateThumbprintFromString(data string) string {
	return GenerateThumbprint([]byte(data))
}

// Hash returns the hash of the given data using the specified algorithm.
func Hash(data []byte, alg HashAlgorithm) ([]byte, error) {
	switch alg {
	case GenericSHA256:
		h := sha256.Sum256(data)
		return h[:], nil
	case GenericSHA384:
		h := sha512.Sum384(data)
		return h[:], nil
	case GenericSHA512:
		h := sha512.Sum512(data)
		return h[:], nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", alg)
	}
}

// GetHash returns a hash.Hash for the given algorithm.
func GetHash(alg HashAlgorithm) (hash.Hash, error) {
	switch alg {
	case GenericSHA256:
		return sha256.New(), nil
	case GenericSHA384:
		return sha512.New384(), nil
	case GenericSHA512:
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", alg)
	}
}
