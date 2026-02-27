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

package security

import (
	"net/http"
)

// AuthenticatorInterface defines the interface for pluggable authentication mechanisms.
// Implementations handle different authentication methods (JWT, API keys, mTLS, etc.).
type AuthenticatorInterface interface {
	// CanHandle determines if this authenticator can process the given request.
	// Returns true if the authenticator recognizes the authentication mechanism in the request.
	CanHandle(r *http.Request) bool

	// Authenticate validates credentials and builds a SecurityContext on success.
	// On success, returns a SecurityContext with authenticated subject information and permissions.
	// On failure, returns an authentication error (401).
	Authenticate(r *http.Request) (*SecurityContext, error)
}
