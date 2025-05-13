/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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
	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/log"
)

// AuthorizationValidatorInterface defines the interface for validating OAuth2 authorization requests.
type AuthorizationValidatorInterface interface {
	validateInitialAuthorizationRequest(msg *model.OAuthMessage,
		app *appmodel.OAuthApplication) (bool, string, string)
}

// AuthorizationValidator implements the AuthorizationValidatorInterface for validating OAuth2 authorization requests.
type AuthorizationValidator struct{}

// NewAuthorizationValidator creates a new instance of AuthorizationValidator.
func NewAuthorizationValidator() AuthorizationValidatorInterface {
	return &AuthorizationValidator{}
}

// validateInitialAuthorizationRequest validates the initial authorization request parameters.
func (av *AuthorizationValidator) validateInitialAuthorizationRequest(msg *model.OAuthMessage,
	app *appmodel.OAuthApplication) (bool, string, string) {
	logger := log.GetLogger()
	// Extract required parameters.
	responseType := msg.RequestQueryParams[constants.RESPONSE_TYPE]
	clientId := msg.RequestQueryParams[constants.CLIENT_ID]
	redirectUri := msg.RequestQueryParams[constants.REDIRECT_URI]

	if clientId == "" {
		return false, constants.ERROR_INVALID_REQUEST, "Missing client_id parameter"
	}

	// Validate if the authorization code grant type is allowed for the app.
	if !app.IsAllowedGrantType(constants.GRANT_TYPE_AUTHORIZATION_CODE) {
		return false, constants.ERROR_UNSUPPORTED_GRANT_TYPE,
			"Authorization code grant type is not allowed for the client"
	}

	// Validate the redirect URI against the registered application.
	if err := app.ValidateRedirectURI(redirectUri); err != nil {
		logger.Error("Validation failed for redirect URI", log.Error(err))
		return false, constants.ERROR_INVALID_REQUEST, "Invalid redirect URI"
	}

	// Validate the authorization request.
	if responseType == "" {
		return true, constants.ERROR_INVALID_REQUEST, "Missing response_type parameter"
	}
	if responseType != constants.RESPONSE_TYPE_CODE {
		return true, constants.ERROR_UNSUPPORTED_RESPONSE_TYPE, "Unsupported response type"
	}

	return false, "", ""
}
