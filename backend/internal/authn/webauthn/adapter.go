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

package webauthn

import (
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/asgardeo/thunder/internal/user"
)

// webAuthnUser adapts our user model to implement the webauthn.User interface.
type webAuthnUser struct {
	id          []byte
	name        string
	displayName string
	credentials []webauthn.Credential
}

// WebAuthnID returns the user's ID as required by webauthn.User interface.
func (u *webAuthnUser) WebAuthnID() []byte {
	return u.id
}

// WebAuthnName returns the user's name as required by webauthn.User interface.
func (u *webAuthnUser) WebAuthnName() string {
	return u.name
}

// WebAuthnDisplayName returns the user's display name as required by webauthn.User interface.
func (u *webAuthnUser) WebAuthnDisplayName() string {
	return u.displayName
}

// WebAuthnCredentials returns the user's credentials as required by webauthn.User interface.
func (u *webAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// newWebAuthnUser creates a new WebAuthn user adapter.
func newWebAuthnUser(userID string, userName, displayName string, credentials []webauthn.Credential) *webAuthnUser {
	return &webAuthnUser{
		id:          []byte(userID),
		name:        userName,
		displayName: displayName,
		credentials: credentials,
	}
}

// createWebAuthnUserFromUserInfo creates a WebAuthn user adapter from user info.
func createWebAuthnUserFromUserInfo(userInfo *user.User, credentials []webauthn.Credential) *webAuthnUser {
	displayName, userName := extractUserInfo(userInfo)
	return newWebAuthnUser(userInfo.ID, userName, displayName, credentials)
}
