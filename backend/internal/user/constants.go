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

// Credential type constants for system-managed credential types.
// System-managed credentials are not defined in user schemas.
const (
	CredentialTypePasskey CredentialType = "passkey"
)

// systemManagedCredentialTypes defines credential types that are managed by the system,
// not through user schemas. These may support multiple values per user.
var systemManagedCredentialTypes = []CredentialType{
	CredentialTypePasskey,
}

// String returns the string representation of the credential type.
func (ct CredentialType) String() string {
	return string(ct)
}

// IsSystemManaged checks if the credential type is a system-managed credential type.
func (ct CredentialType) IsSystemManaged() bool {
	return slices.Contains(systemManagedCredentialTypes, ct)
}
