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

package passkey

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Wrapper types to abstract the underlying WebAuthn library. .

// RegistrationOption wraps library-specific registration options.
type RegistrationOption = webauthn.RegistrationOption

// SessionData wraps library-specific session data.
type SessionData = webauthn.SessionData

// CredentialCreation wraps library-specific credential creation options.
type CredentialCreation = protocol.CredentialCreation

// CredentialAssertion wraps library-specific credential assertion options.
type CredentialAssertion = protocol.CredentialAssertion

// ParsedCredentialCreationData wraps library-specific parsed credential creation data.
type ParsedCredentialCreationData = protocol.ParsedCredentialCreationData

// ParsedCredentialAssertionData wraps library-specific parsed credential assertion data.
type ParsedCredentialAssertionData = protocol.ParsedCredentialAssertionData

// WebauthnCredential wraps library-specific credential data.
type WebauthnCredential = webauthn.Credential

// webAuthnService provides an abstraction layer over the WebAuthn library.
type webAuthnService interface {
	// BeginRegistration initiates the registration ceremony and returns credential creation options.
	BeginRegistration(
		user webauthnUserInterface,
		options []RegistrationOption,
	) (*CredentialCreation, *SessionData, error)

	// CreateCredential validates the registration response and creates a credential.
	CreateCredential(
		user webauthnUserInterface,
		session SessionData,
		response *ParsedCredentialCreationData,
	) (*WebauthnCredential, error)

	// BeginLogin initiates the authentication ceremony and returns credential request options.
	BeginLogin(
		user webauthnUserInterface,
	) (*CredentialAssertion, *SessionData, error)

	// ValidateLogin validates the authentication response and returns the verified credential.
	ValidateLogin(
		user webauthnUserInterface,
		session SessionData,
		response *ParsedCredentialAssertionData,
	) (*WebauthnCredential, error)
}

// defaultWebAuthnService is the default implementation using the GO-WebAuthn library.
type defaultWebAuthnService struct {
	webAuthnLib *webauthn.WebAuthn
}

// newDefaultWebAuthnService creates a new service instance with the given configuration.
func newDefaultWebAuthnService(
	relyingPartyID, rpDisplayName string,
	rpOrigins []string,
) (webAuthnService, error) {
	config := &webauthn.Config{
		RPDisplayName: rpDisplayName,
		RPID:          relyingPartyID,
		RPOrigins:     rpOrigins,
	}

	webAuthnLib, err := webauthn.New(config)
	if err != nil {
		return nil, err
	}

	return &defaultWebAuthnService{
		webAuthnLib: webAuthnLib,
	}, nil
}

// BeginRegistration wraps the WebAuthn library's BeginRegistration method.
func (a *defaultWebAuthnService) BeginRegistration(
	user webauthnUserInterface,
	options []RegistrationOption,
) (*CredentialCreation, *SessionData, error) {
	return a.webAuthnLib.BeginRegistration(user, options...)
}

// CreateCredential wraps the WebAuthn library's CreateCredential method.
func (a *defaultWebAuthnService) CreateCredential(
	user webauthnUserInterface,
	session SessionData,
	response *ParsedCredentialCreationData,
) (*WebauthnCredential, error) {
	return a.webAuthnLib.CreateCredential(user, session, response)
}

// BeginLogin wraps the WebAuthn library's BeginLogin method.
func (a *defaultWebAuthnService) BeginLogin(
	user webauthnUserInterface,
) (*CredentialAssertion, *SessionData, error) {
	return a.webAuthnLib.BeginLogin(user)
}

// ValidateLogin wraps the WebAuthn library's ValidateLogin method.
func (a *defaultWebAuthnService) ValidateLogin(
	user webauthnUserInterface,
	session SessionData,
	response *ParsedCredentialAssertionData,
) (*WebauthnCredential, error) {
	return a.webAuthnLib.ValidateLogin(user, session, response)
}
