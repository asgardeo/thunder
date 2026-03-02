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

// Package clientauth provides shared client authentication logic for OAuth2 endpoints.
package clientauth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/application"
	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/discovery"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/jose/jws"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
)

// authenticate authenticates the OAuth2 client from the request.
// It extracts credentials, validates them, and returns OAuthClientInfo on success.
// Returns an authError on failure.
func authenticate(
	r *http.Request,
	appService application.ApplicationServiceInterface,
	jwtService jwt.JWTServiceInterface,
	discoveryService discovery.DiscoveryServiceInterface,
) (*OAuthClientInfo, *authError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ClientAuthMiddleware"))

	// Extract all possible auth fields
	hasAuthHeader := r.Header.Get(serverconst.AuthorizationHeaderName) != ""
	clientIDFromBody := r.FormValue(constants.RequestParamClientID)
	clientSecretFromBody := r.FormValue(constants.RequestParamClientSecret)
	clientAssertionType := r.FormValue(constants.RequestParamClientAssertionType)
	clientAssertion := r.FormValue(constants.RequestParamClientAssertion)

	var detectedMethod constants.TokenEndpointAuthMethod

	// Method 1: Basic Auth (header)
	if hasAuthHeader {
		detectedMethod = constants.TokenEndpointAuthMethodClientSecretBasic
	}

	// Method 2: Client credentials in body
	if clientSecretFromBody != "" {
		if detectedMethod != "" {
			return nil, errMultipleAuthMethods
		}
		detectedMethod = constants.TokenEndpointAuthMethodClientSecretPost
	}

	// Method 3: Client assertion (private_key_jwt)
	if clientAssertionType != "" || clientAssertion != "" {
		if detectedMethod != "" {
			return nil, errMultipleAuthMethods
		}
		detectedMethod = constants.TokenEndpointAuthMethodPrivateKeyJWT
	}

	// If no auth method but client_id exists -> public client
	if detectedMethod == "" && clientIDFromBody != "" {
		detectedMethod = constants.TokenEndpointAuthMethodNone
	}

	// Now process based on detected method
	var clientID string
	var clientSecret string

	switch detectedMethod {
	case constants.TokenEndpointAuthMethodClientSecretBasic:
		var err *authError
		clientID, clientSecret, err = extractBasicAuthCredentials(r)
		if err != nil {
			return nil, err
		}

	case constants.TokenEndpointAuthMethodClientSecretPost:
		if clientIDFromBody == "" {
			return nil, errMissingClientID
		}
		clientID = clientIDFromBody
		clientSecret = clientSecretFromBody

	case constants.TokenEndpointAuthMethodPrivateKeyJWT:
		if clientAssertionType != constants.SupportedClientAssertionType {
			logger.Error("Invalid client assertion: unsupported client assertion type")
			return nil, errInvalidClientAssertion
		}
		extracted, err := extractClientIDFromAssertion(clientAssertion)
		if err != nil {
			return nil, err
		}
		clientID = extracted

	case constants.TokenEndpointAuthMethodNone:
		clientID = clientIDFromBody

	default:
		return nil, errMissingClientID
	}

	if clientIDFromBody != "" && clientID != clientIDFromBody {
		return nil, errClientIDMismatch
	}

	oauthApp, err := appService.GetOAuthApplication(clientID)
	if err != nil || oauthApp == nil {
		return nil, errInvalidClientCredentials
	}

	if !oauthApp.IsAllowedTokenEndpointAuthMethod(detectedMethod) {
		return nil, errUnauthorizedAuthMethod
	}

	// Validate credentials based on method
	switch detectedMethod {
	case constants.TokenEndpointAuthMethodPrivateKeyJWT:
		if err := validateClientAssertion(oauthApp, jwtService, discoveryService, clientID,
			clientAssertion); err != nil {
			logger.Error("Invalid client assertion: " + err.Error())
			return nil, errInvalidClientAssertion
		}
	case constants.TokenEndpointAuthMethodClientSecretBasic,
		constants.TokenEndpointAuthMethodClientSecretPost:
		if !oauthApp.ValidateCredentials(clientID, clientSecret) {
			return nil, errInvalidClientCredentials
		}
	}

	return &OAuthClientInfo{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		OAuthApp:     oauthApp,
	}, nil
}

// extractBasicAuthCredentials extracts the basic authentication credentials from the request header.
func extractBasicAuthCredentials(r *http.Request) (string, string, *authError) {
	authHeader := r.Header.Get(serverconst.AuthorizationHeaderName)
	if !strings.HasPrefix(authHeader, serverconst.AuthSchemeBasic) {
		return "", "", errInvalidAuthorizationHeader
	}

	encodedCredentials := strings.TrimPrefix(authHeader, serverconst.AuthSchemeBasic)
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return "", "", errInvalidAuthorizationHeader
	}

	credentials := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(credentials) != 2 || credentials[0] == "" || credentials[1] == "" {
		return "", "", errInvalidAuthorizationHeader
	}

	return credentials[0], credentials[1], nil
}

// extractClientIDFromAssertion extracts the client_id from the JWT assertion's 'sub' claim.
// This parses the JWT WITHOUT signature verification to extract the subject.
func extractClientIDFromAssertion(assertion string) (string, *authError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ClientAuthMiddleware"))

	payload, err := jwt.DecodeJWTPayload(assertion)
	if err != nil {
		logger.Error("Invalid client assertion: failed to decode jwt")
		return "", errInvalidClientAssertion
	}

	subject, ok := payload["sub"].(string)

	if !ok || subject == "" {
		logger.Error("Invalid client assertion: missing 'sub' claim or 'sub' claim is not a string")
		return "", errInvalidClientAssertion
	}

	return subject, nil
}

// validateClientAssertion validates the provided client assertion JWT using the configured certificate and JWT service.
func validateClientAssertion(
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
	jwtService jwt.JWTServiceInterface,
	discoveryService discovery.DiscoveryServiceInterface,
	clientID, clientAssertion string) error {
	if oauthApp.Certificate == nil || oauthApp.Certificate.Type == cert.CertificateTypeNone {
		return fmt.Errorf("No certificate configured for client assertion validation")
	}

	// Get token endpoint from discovery service for aud validation
	tokenEndpoint := discoveryService.GetOAuth2AuthorizationServerMetadata().TokenEndpoint

	if oauthApp.Certificate.Type == cert.CertificateTypeJWKSURI {
		if err := jwtService.VerifyJWTWithJWKS(clientAssertion, oauthApp.Certificate.Value, tokenEndpoint, clientID); err != nil {
			return fmt.Errorf("Client assertion verification with JWKS URI failed: %v", err.Error)
		}
		return nil
	}

	var jwks struct {
		Keys []map[string]any `json:"keys"`
	}
	if err := json.Unmarshal([]byte(oauthApp.Certificate.Value), &jwks); err != nil {
		return fmt.Errorf("Invalid jwks certificate format: %v", err.Error())
	}

	var kid string
	if header, err := jwt.DecodeJWTHeader(clientAssertion); err != nil {
		return fmt.Errorf("Failed to decode header: %v", err.Error())
	} else if k, ok := header["kid"].(string); !ok || k == "" {
		return fmt.Errorf("JWT header missing 'kid' claim or 'kid' is not a string")
	} else {
		kid = k
	}

	var jwk map[string]any
	for _, key := range jwks.Keys {
		if keyID, ok := key["kid"].(string); ok && keyID == kid {
			jwk = key
			break
		}
	}
	if jwk == nil {
		return fmt.Errorf("No matching key found in JWKS for kid: %v", kid)
	}

	pubKey, err := jws.JWKToPublicKey(jwk)
	if err != nil {
		return fmt.Errorf("Failed to convert JWK to public key: %v", err.Error())
	}

	if err := jwtService.VerifyJWTWithPublicKey(clientAssertion, pubKey, tokenEndpoint, clientID); err != nil {
		return fmt.Errorf("Client assertion verification failed: %v", err.Error)
	}

	return nil
}
