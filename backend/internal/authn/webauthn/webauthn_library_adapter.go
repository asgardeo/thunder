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
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// webAuthnLibraryAdapter provides an abstraction layer over the GO-WebAuthn library.
// This adapter decouples the business logic from the library implementation.
type webAuthnLibraryAdapter interface {
	// BeginRegistration initiates the registration ceremony and returns credential creation options.
	BeginRegistration(
		user webauthnUserInterface,
		options []webauthn.RegistrationOption,
	) (*protocol.CredentialCreation, *webauthn.SessionData, error)

	// CreateCredential validates the registration response and creates a credential.
	CreateCredential(
		user webauthnUserInterface,
		session webauthn.SessionData,
		response *protocol.ParsedCredentialCreationData,
	) (*webauthn.Credential, error)

	// BeginLogin initiates the authentication ceremony and returns credential request options.
	BeginLogin(
		user webauthnUserInterface,
	) (*protocol.CredentialAssertion, *webauthn.SessionData, error)

	// ValidateLogin validates the authentication response and returns the verified credential.
	ValidateLogin(
		user webauthnUserInterface,
		session webauthn.SessionData,
		response *protocol.ParsedCredentialAssertionData,
	) (*webauthn.Credential, error)
}

// webauthnUserInterface defines the interface for WebAuthn user operations.
// This interface abstracts the webauthn.User interface to reduce direct dependencies.
type webauthnUserInterface interface {
	WebAuthnID() []byte
	WebAuthnName() string
	WebAuthnDisplayName() string
	WebAuthnCredentials() []webauthn.Credential
}

// defaultWebAuthnLibraryAdapter is the default implementation using the GO-WebAuthn library.
type defaultWebAuthnLibraryAdapter struct {
	webAuthnLib *webauthn.WebAuthn
}

// newDefaultWebAuthnLibraryAdapter creates a new adapter instance with the given configuration.
func newDefaultWebAuthnLibraryAdapter(
	relyingPartyID, rpDisplayName string,
	rpOrigins []string,
) (webAuthnLibraryAdapter, error) {
	config := &webauthn.Config{
		RPDisplayName: rpDisplayName,
		RPID:          relyingPartyID,
		RPOrigins:     rpOrigins,
	}

	webAuthnLib, err := webauthn.New(config)
	if err != nil {
		return nil, err
	}

	return &defaultWebAuthnLibraryAdapter{
		webAuthnLib: webAuthnLib,
	}, nil
}

// BeginRegistration wraps the WebAuthn library's BeginRegistration method.
func (a *defaultWebAuthnLibraryAdapter) BeginRegistration(
	user webauthnUserInterface,
	options []webauthn.RegistrationOption,
) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	return a.webAuthnLib.BeginRegistration(user, options...)
}

// CreateCredential wraps the WebAuthn library's CreateCredential method.
func (a *defaultWebAuthnLibraryAdapter) CreateCredential(
	user webauthnUserInterface,
	session webauthn.SessionData,
	response *protocol.ParsedCredentialCreationData,
) (*webauthn.Credential, error) {
	return a.webAuthnLib.CreateCredential(user, session, response)
}

// BeginLogin wraps the WebAuthn library's BeginLogin method.
func (a *defaultWebAuthnLibraryAdapter) BeginLogin(
	user webauthnUserInterface,
) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	return a.webAuthnLib.BeginLogin(user)
}

// ValidateLogin wraps the WebAuthn library's ValidateLogin method.
func (a *defaultWebAuthnLibraryAdapter) ValidateLogin(
	user webauthnUserInterface,
	session webauthn.SessionData,
	response *protocol.ParsedCredentialAssertionData,
) (*webauthn.Credential, error) {
	return a.webAuthnLib.ValidateLogin(user, session, response)
}
