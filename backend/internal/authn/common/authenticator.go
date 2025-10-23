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

package common

import (
	"time"

	"github.com/asgardeo/thunder/internal/idp"
)

// Authenticator name constants.
const (
	AuthenticatorCredentials = "CredentialsAuthenticator"
	AuthenticatorSMSOTP      = "SMSOTPAuthenticator"
	AuthenticatorGoogle      = "GoogleOIDCAuthenticator"
	AuthenticatorGithub      = "GithubOAuthAuthenticator"
	AuthenticatorOAuth       = "OAuthAuthenticator"
	AuthenticatorOIDC        = "OIDCAuthenticator"
)

// AuthenticatorMeta represents an authenticator's metadata including its AAL weight.
type AuthenticatorMeta struct {
	// Name is the unique identifier for the authenticator (used in individual authentication APIs)
	Name string
	// AALWeight represents the strength of the authenticator for AAL calculation
	AALWeight int
}

// AuthenticatorReference represents an engaged authenticator in the authentication flow.
type AuthenticatorReference struct {
	// Authenticator is the name of the authenticator
	Authenticator string `json:"authenticator"`
	// Step is the step number in the flow where this authenticator was engaged
	Step int `json:"step"`
	// Timestamp is the time when the authenticator was engaged
	Timestamp time.Time `json:"timestamp"`
}

// authenticatorRegistry holds the metadata for all authenticators in the system.
var authenticatorRegistry = map[string]AuthenticatorMeta{
	AuthenticatorCredentials: {
		Name:      AuthenticatorCredentials,
		AALWeight: 1,
	},
	AuthenticatorSMSOTP: {
		Name:      AuthenticatorSMSOTP,
		AALWeight: 1,
	},
	AuthenticatorGoogle: {
		Name:      AuthenticatorGoogle,
		AALWeight: 1,
	},
	AuthenticatorGithub: {
		Name:      AuthenticatorGithub,
		AALWeight: 1,
	},
	AuthenticatorOAuth: {
		Name:      AuthenticatorOAuth,
		AALWeight: 1,
	},
	AuthenticatorOIDC: {
		Name:      AuthenticatorOIDC,
		AALWeight: 1,
	},
}

// getAuthenticatorMetaData returns the authenticator metadata for the given authenticator.
func getAuthenticatorMetaData(name string) *AuthenticatorMeta {
	if auth, ok := authenticatorRegistry[name]; ok {
		return &auth
	}
	return nil
}

// GetAuthenticatorWeight returns the AAL weight for the given authenticator or executor name.
func GetAuthenticatorWeight(name string) int {
	if auth := getAuthenticatorMetaData(name); auth != nil {
		return auth.AALWeight
	}
	return 0
}

// GetAuthenticatorNameForIDPType returns the authenticator name for a given IDP type.
func GetAuthenticatorNameForIDPType(idpType idp.IDPType) string {
	switch idpType {
	case idp.IDPTypeGoogle:
		return AuthenticatorGoogle
	case idp.IDPTypeGitHub:
		return AuthenticatorGithub
	case idp.IDPTypeOAuth:
		return AuthenticatorOAuth
	case idp.IDPTypeOIDC:
		return AuthenticatorOIDC
	default:
		return AuthenticatorOAuth
	}
}
