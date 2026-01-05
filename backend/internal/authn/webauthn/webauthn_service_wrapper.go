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

// Package webauthn implements the WebAuthn/Passkey authentication service.
package webauthn

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// webAuthnLibraryService provides high-level WebAuthn operations using the library adapter.
// This service layer acts as a facade, providing convenience methods for common WebAuthn workflows.
type webAuthnLibraryService struct {
	adapter webAuthnLibraryAdapter
}

// newWebAuthnLibraryService creates a new WebAuthn library service.
func newWebAuthnLibraryService(relyingPartyID, rpDisplayName string) (*webAuthnLibraryService, error) {
	rpOrigins := getConfiguredOrigins()

	adapter, err := newDefaultWebAuthnLibraryAdapter(relyingPartyID, rpDisplayName, rpOrigins)
	if err != nil {
		return nil, err
	}

	return &webAuthnLibraryService{
		adapter: adapter,
	}, nil
}

// BeginRegistration initiates the WebAuthn registration ceremony.
// It wraps the adapter's BeginRegistration method with the provided user and options.
func (s *webAuthnLibraryService) BeginRegistration(
	user webauthnUserInterface,
	options []webauthn.RegistrationOption,
) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	return s.adapter.BeginRegistration(user, options)
}

// FinishRegistration completes the WebAuthn registration ceremony.
// It validates the attestation response and creates a new credential.
func (s *webAuthnLibraryService) FinishRegistration(
	user webauthnUserInterface,
	sessionData webauthn.SessionData,
	parsedResponse *protocol.ParsedCredentialCreationData,
) (*webauthn.Credential, error) {
	return s.adapter.CreateCredential(user, sessionData, parsedResponse)
}

// BeginAuthentication initiates the WebAuthn authentication ceremony.
// It wraps the adapter's BeginLogin method with the provided user.
func (s *webAuthnLibraryService) BeginAuthentication(
	user webauthnUserInterface,
) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	return s.adapter.BeginLogin(user)
}

// FinishAuthentication completes the WebAuthn authentication ceremony.
// It validates the assertion response and returns the verified credential.
func (s *webAuthnLibraryService) FinishAuthentication(
	user webauthnUserInterface,
	sessionData webauthn.SessionData,
	parsedResponse *protocol.ParsedCredentialAssertionData,
) (*webauthn.Credential, error) {
	return s.adapter.ValidateLogin(user, sessionData, parsedResponse)
}
