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

package sign

// SignAlgorithm represents the supported digital signature algorithms
type SignAlgorithm string

const (
	// RSASHA256 represents RSA signature with SHA-256 hash
	RSASHA256 SignAlgorithm = "RSA-SHA256"
	// RSASHA512 represents RSA signature with SHA-512 hash
	RSASHA512 SignAlgorithm = "RSA-SHA512"
	// ECDSASHA256 represents ECDSA signature with SHA-256 hash
	ECDSASHA256 SignAlgorithm = "ECDSA-SHA256"
	// ECDSASHA384 represents ECDSA signature with SHA-384 hash
	ECDSASHA384 SignAlgorithm = "ECDSA-SHA384"
	// ECDSASHA512 represents ECDSA signature with SHA-512 hash
	ECDSASHA512 SignAlgorithm = "ECDSA-SHA512"
	// ED25519 represents ED25519 signature algorithm
	ED25519 SignAlgorithm = "ED25519"
)
