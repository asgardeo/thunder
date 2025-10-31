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

package authz

import (
	"fmt"
	"net/url"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/pkce"
	"github.com/asgardeo/thunder/internal/system/log"
)

// AuthorizationValidatorInterface defines the interface for validating OAuth2 authorization requests.
type AuthorizationValidatorInterface interface {
	validateInitialAuthorizationRequest(msg *OAuthMessage, oauthApp *appmodel.OAuthAppConfigProcessedDTO) (
		bool, string, string)
}

// authorizationValidator implements the AuthorizationValidatorInterface for validating OAuth2 authorization requests.
type authorizationValidator struct{}

// newAuthorizationValidator creates a new instance of authorizationValidator.
func newAuthorizationValidator() AuthorizationValidatorInterface {
	return &authorizationValidator{}
}

// validateInitialAuthorizationRequest validates the initial authorization request parameters.
func (av *authorizationValidator) validateInitialAuthorizationRequest(msg *OAuthMessage,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) (bool, string, string) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthorizationValidator"))

	// Extract required parameters.
	responseType := msg.RequestQueryParams[constants.RequestParamResponseType]
	clientID := msg.RequestQueryParams[constants.RequestParamClientID]
	redirectURI := msg.RequestQueryParams[constants.RequestParamRedirectURI]

	if clientID == "" {
		return false, constants.ErrorInvalidRequest, "Missing client_id parameter"
	}

	// Validate the redirect URI against the registered application.
	if err := oauthApp.ValidateRedirectURI(redirectURI); err != nil {
		logger.Error("Validation failed for redirect URI", log.Error(err))
		return false, constants.ErrorInvalidRequest, "Invalid redirect URI"
	}

	// Validate if the authorization code grant type is allowed for the app.
	if !oauthApp.IsAllowedGrantType(constants.GrantTypeAuthorizationCode) {
		return true, constants.ErrorUnsupportedGrantType,
			"Authorization code grant type is not allowed for the client"
	}

	// Validate the authorization request.
	if responseType == "" {
		return true, constants.ErrorInvalidRequest, "Missing response_type parameter"
	}
	if !oauthApp.IsAllowedResponseType(responseType) {
		return true, constants.ErrorUnsupportedResponseType, "Unsupported response type"
	}

	// Validate PKCE parameters if required
	if oauthApp.RequiresPKCE() && responseType == string(constants.ResponseTypeCode) {
		codeChallenge := msg.RequestQueryParams[constants.RequestParamCodeChallenge]
		codeChallengeMethod := msg.RequestQueryParams[constants.RequestParamCodeChallengeMethod]

		if codeChallenge == "" {
			return true, constants.ErrorInvalidRequest, "code_challenge is required for this application"
		}

		// Validate code challenge format and method
		if err := pkce.ValidateCodeChallenge(codeChallenge, codeChallengeMethod); err != nil {
			return true, constants.ErrorInvalidRequest, "Invalid PKCE parameters"
		}
	}

	// Validate resource parameter if present
	resource := msg.RequestQueryParams[constants.RequestParamResource]
	if resource != "" {
		if err := validateResourceParameter(resource); err != nil {
			return true, constants.ErrorInvalidTarget, err.Error()
		}
	}

	return false, "", ""
}

// validateResourceParameter validates the resource parameter.
// TODO: Need to add other validations after introducing resources.
func validateResourceParameter(resource string) error {
	parsedURL, err := url.Parse(resource)
	if err != nil {
		return fmt.Errorf("resource parameter must be a valid absolute URI: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("resource parameter must be an absolute URI with a scheme")
	}

	if parsedURL.Fragment != "" {
		return fmt.Errorf("resource parameter must not include a fragment component")
	}

	return nil
}
