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

package jwe

// KeyEncAlgorithm represents the JWE key management algorithm (alg header parameter)
type KeyEncAlgorithm string

const (
	// RSAOAEP256 represents RSA-OAEP using SHA-256 key encryption
	RSAOAEP256 KeyEncAlgorithm = "RSA-OAEP-256"
	// ECDHES represents ECDH-ES key encryption
	ECDHES KeyEncAlgorithm = "ECDH-ES"
	// ECDHESA128KW represents ECDH-ES with AES 128 Key Wrap
	ECDHESA128KW KeyEncAlgorithm = "ECDH-ES+A128KW"
	// ECDHESA256KW represents ECDH-ES with AES 256 Key Wrap
	ECDHESA256KW KeyEncAlgorithm = "ECDH-ES+A256KW"
)

// ContentEncAlgorithm represents the JWE content encryption algorithm (enc header parameter)
type ContentEncAlgorithm string

const (
	// A128GCM represents AES GCM using 128-bit key
	A128GCM ContentEncAlgorithm = "A128GCM"
	// A192GCM represents AES GCM using 192-bit key
	A192GCM ContentEncAlgorithm = "A192GCM"
	// A256GCM represents AES GCM using 256-bit key
	A256GCM ContentEncAlgorithm = "A256GCM"
)
