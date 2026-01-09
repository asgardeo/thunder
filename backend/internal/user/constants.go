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

package user

import "slices"

// CredentialType represents the type of credential.
type CredentialType string

// Credential type constants define the supported credential types.
const (
	CredentialTypePassword CredentialType = "password"
	CredentialTypePin      CredentialType = "pin"
	CredentialTypeSecret   CredentialType = "secret"
	CredentialTypePasskey  CredentialType = "passkey"
)

// SupportedCredentialTypes defines the set of credential types that are supported.
var SupportedCredentialTypes = []CredentialType{
	CredentialTypePassword,
	CredentialTypePin,
	CredentialTypeSecret,
	CredentialTypePasskey,
}

// HashedCredentialTypes defines credential types that require hashing.
var HashedCredentialTypes = []CredentialType{
	CredentialTypePassword,
	CredentialTypePin,
	CredentialTypeSecret,
}

// String returns the string representation of the credential type.
func (ct CredentialType) String() string {
	return string(ct)
}

// RequiresHashing checks if this credential type requires hashing.
func (ct CredentialType) RequiresHashing() bool {
	return slices.Contains(HashedCredentialTypes, ct)
}

// IsValid checks if the credential type is supported.
func (ct CredentialType) IsValid() bool {
	return slices.Contains(SupportedCredentialTypes, ct)
}
