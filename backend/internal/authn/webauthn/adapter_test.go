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
	"encoding/json"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"

	"github.com/asgardeo/thunder/internal/user"
)

func TestNewWebAuthnUser(t *testing.T) {
	userID := testUserID
	userName := "johndoe"
	displayName := "John Doe"
	credentials := []webauthn.Credential{}

	webAuthnUser := newWebAuthnUser(userID, userName, displayName, credentials)

	assert.NotNil(t, webAuthnUser)
	assert.Equal(t, []byte(userID), webAuthnUser.WebAuthnID())
	assert.Equal(t, userName, webAuthnUser.WebAuthnName())
	assert.Equal(t, displayName, webAuthnUser.WebAuthnDisplayName())
	assert.Equal(t, credentials, webAuthnUser.WebAuthnCredentials())
}

func TestWebAuthnUser_Methods(t *testing.T) {
	userID := testUserID
	userName := "johndoe"
	displayName := "John Doe"
	credentials := []webauthn.Credential{
		{
			ID: []byte("credential1"),
		},
	}

	webAuthnUser := &webAuthnUser{
		id:          []byte(userID),
		name:        userName,
		displayName: displayName,
		credentials: credentials,
	}

	assert.Equal(t, []byte(userID), webAuthnUser.WebAuthnID())
	assert.Equal(t, userName, webAuthnUser.WebAuthnName())
	assert.Equal(t, displayName, webAuthnUser.WebAuthnDisplayName())
	assert.Len(t, webAuthnUser.WebAuthnCredentials(), 1)
	assert.Equal(t, []byte("credential1"), webAuthnUser.WebAuthnCredentials()[0].ID)
}

func TestCreateWebAuthnUserFromUserInfo_WithFullAttributes(t *testing.T) {
	attrs := json.RawMessage(`{"firstName":"John","lastName":"Doe","username":"johndoe"}`)
	userInfo := &user.User{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "org123",
		Attributes:       attrs,
	}
	credentials := []webauthn.Credential{}

	webAuthnUser := createWebAuthnUserFromUserInfo(userInfo, credentials)

	assert.NotNil(t, webAuthnUser)
	assert.Equal(t, []byte("user123"), webAuthnUser.WebAuthnID())
	assert.Equal(t, "johndoe", webAuthnUser.WebAuthnName())
	assert.Equal(t, "John Doe", webAuthnUser.WebAuthnDisplayName())
	assert.Equal(t, credentials, webAuthnUser.WebAuthnCredentials())
}

func TestCreateWebAuthnUserFromUserInfo_WithEmailOnly(t *testing.T) {
	attrs := json.RawMessage(`{"email":"john@example.com"}`)
	userInfo := &user.User{
		ID:         "user123",
		Attributes: attrs,
	}
	credentials := []webauthn.Credential{}

	webAuthnUser := createWebAuthnUserFromUserInfo(userInfo, credentials)

	assert.NotNil(t, webAuthnUser)
	assert.Equal(t, []byte("user123"), webAuthnUser.WebAuthnID())
	assert.Equal(t, "john@example.com", webAuthnUser.WebAuthnName())
	assert.Equal(t, "user123", webAuthnUser.WebAuthnDisplayName()) // Falls back to ID
}

func TestCreateWebAuthnUserFromUserInfo_NoAttributes(t *testing.T) {
	userInfo := &user.User{
		ID: "user123",
	}
	credentials := []webauthn.Credential{}

	webAuthnUser := createWebAuthnUserFromUserInfo(userInfo, credentials)

	assert.NotNil(t, webAuthnUser)
	assert.Equal(t, []byte("user123"), webAuthnUser.WebAuthnID())
	assert.Equal(t, "user123", webAuthnUser.WebAuthnName())
	assert.Equal(t, "user123", webAuthnUser.WebAuthnDisplayName())
}
