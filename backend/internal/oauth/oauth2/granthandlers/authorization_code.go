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

package granthandlers

import (
	"net/url"
	"slices"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/pkce"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/user"
)

// authorizationCodeGrantHandler handles the authorization code grant type.
type authorizationCodeGrantHandler struct {
	authzService authz.AuthorizeServiceInterface
	userService  user.UserServiceInterface
	tokenBuilder tokenservice.TokenBuilderInterface
}

// newAuthorizationCodeGrantHandler creates a new instance of AuthorizationCodeGrantHandler.
func newAuthorizationCodeGrantHandler(
	userService user.UserServiceInterface,
	authzService authz.AuthorizeServiceInterface,
	tokenBuilder tokenservice.TokenBuilderInterface,
) GrantHandlerInterface {
	return &authorizationCodeGrantHandler{
		authzService: authzService,
		userService:  userService,
		tokenBuilder: tokenBuilder,
	}
}

// ValidateGrant validates the authorization code grant request.
func (h *authorizationCodeGrantHandler) ValidateGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) *model.ErrorResponse {
	if tokenRequest.GrantType == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Missing grant type",
		}
	}
	if constants.GrantType(tokenRequest.GrantType) != constants.GrantTypeAuthorizationCode {
		return &model.ErrorResponse{
			Error:            constants.ErrorUnsupportedGrantType,
			ErrorDescription: "Unsupported grant type",
		}
	}
	if tokenRequest.Code == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Authorization code is required",
		}
	}
	if tokenRequest.ClientID == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidClient,
			ErrorDescription: "Client Id is required",
		}
	}

	// TODO: Redirect uri is not mandatory when excluded in the authorize request and is valid scenario.
	//  This should be removed when supporting other means of authorization.
	if tokenRequest.RedirectURI == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Redirect URI is required",
		}
	}

	// The resource parameter must be an absolute URI without a fragment component
	if tokenRequest.Resource != "" {
		if !utils.IsValidURI(tokenRequest.Resource) {
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: "Invalid resource parameter: must be an absolute URI",
			}
		}
		parsedURI, err := url.Parse(tokenRequest.Resource)
		if err != nil || parsedURI.Fragment != "" {
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: "Invalid resource parameter: must not contain a fragment component",
			}
		}
	}

	return nil
}

// HandleGrant processes the authorization code grant request and generates a token response.
func (h *authorizationCodeGrantHandler) HandleGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) (
	*model.TokenResponseDTO, *model.ErrorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthorizationCodeGrantHandler"))

	// Retrieve and validate authorization code
	authCode, errResponse := h.retrieveAndValidateAuthCode(tokenRequest, oauthApp, logger)
	if errResponse != nil {
		return nil, errResponse
	}

	// Parse authorized scopes
	authorizedScopes := tokenservice.ParseScopes(authCode.Scopes)

	// Use user attributes from the authorization code
	attrs := authCode.UserAttributes
	if attrs == nil {
		attrs = make(map[string]interface{})
	}

	audience := tokenservice.DetermineAudience("", authCode.Resource, "", authCode.ClientID)

	// Generate access token using tokenBuilder (attributes will be filtered in BuildAccessToken)
	accessToken, err := h.tokenBuilder.BuildAccessToken(&tokenservice.AccessTokenBuildContext{
		Subject:        authCode.AuthorizedUserID,
		Audience:       audience,
		ClientID:       tokenRequest.ClientID,
		Scopes:         authorizedScopes,
		UserAttributes: attrs,
		GrantType:      string(constants.GrantTypeAuthorizationCode),
		OAuthApp:       oauthApp,
	})
	if err != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate token",
		}
	}

	// Build token response
	tokenResponse := &model.TokenResponseDTO{
		AccessToken: *accessToken,
	}

	// Generate ID token if 'openid' scope is present
	if slices.Contains(authorizedScopes, "openid") {
		idToken, err := h.tokenBuilder.BuildIDToken(&tokenservice.IDTokenBuildContext{
			Subject:        authCode.AuthorizedUserID,
			Audience:       tokenRequest.ClientID,
			Scopes:         authorizedScopes,
			UserAttributes: attrs,
			AuthTime:       authCode.TimeCreated.Unix(),
			OAuthApp:       oauthApp,
		})
		if err != nil {
			logger.Error("Failed to generate ID token", log.Error(err))
			return nil, &model.ErrorResponse{
				Error:            constants.ErrorServerError,
				ErrorDescription: "Failed to generate ID token",
			}
		}
		tokenResponse.IDToken = *idToken
	}

	return tokenResponse, nil
}

// retrieveAndValidateAuthCode retrieves and validates the authorization code, including PKCE validation.
func (h *authorizationCodeGrantHandler) retrieveAndValidateAuthCode(
	tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
	logger *log.Logger,
) (*authz.AuthorizationCode, *model.ErrorResponse) {
	authCode, codeErr := h.authzService.GetAuthorizationCodeDetails(tokenRequest.ClientID, tokenRequest.Code)
	if codeErr != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Invalid authorization code",
		}
	}

	// Validate the retrieved authorization code
	errResponse := validateAuthorizationCode(tokenRequest, *authCode)
	if errResponse != nil && errResponse.Error != "" {
		return nil, errResponse
	}

	// Validate PKCE if required or if code challenge was provided during authorization
	if oauthApp.RequiresPKCE() || authCode.CodeChallenge != "" {
		if tokenRequest.CodeVerifier == "" {
			return nil, &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: "code_verifier is required",
			}
		}

		// Validate PKCE
		if err := pkce.ValidatePKCE(authCode.CodeChallenge, authCode.CodeChallengeMethod,
			tokenRequest.CodeVerifier); err != nil {
			logger.Debug("PKCE validation failed", log.Error(err))
			return nil, &model.ErrorResponse{
				Error:            constants.ErrorInvalidGrant,
				ErrorDescription: "Invalid code verifier",
			}
		}
	}
	return authCode, nil
}

// validateAuthorizationCode validates the authorization code against the token request.
func validateAuthorizationCode(tokenRequest *model.TokenRequest,
	code authz.AuthorizationCode) *model.ErrorResponse {
	if tokenRequest.ClientID != code.ClientID {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidClient,
			ErrorDescription: "Invalid client Id",
		}
	}

	// redirect_uri is not mandatory in certain scenarios. Should match if provided with the authorization.
	if code.RedirectURI != "" && tokenRequest.RedirectURI != code.RedirectURI {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Invalid redirect URI",
		}
	}

	// Validate resource parameter consistency with authorization code
	if code.Resource != "" && code.Resource != tokenRequest.Resource {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidTarget,
			ErrorDescription: "Resource parameter mismatch",
		}
	}

	if code.State == authz.AuthCodeStateInactive {
		// TODO: Revoke all the tokens issued for this authorization code.

		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Inactive authorization code",
		}
	} else if code.State != authz.AuthCodeStateActive {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Inactive authorization code",
		}
	}

	if code.ExpiryTime.Before(time.Now()) {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Expired authorization code",
		}
	}

	return nil
}
