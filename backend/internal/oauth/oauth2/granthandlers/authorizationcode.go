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
	"encoding/json"
	"strings"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	authzconstants "github.com/asgardeo/thunder/internal/oauth/oauth2/authz/constants"
	authzmodel "github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/store"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/pkce"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

// authorizationCodeGrantHandler handles the authorization code grant type.
type authorizationCodeGrantHandler struct {
	JWTService  jwt.JWTServiceInterface
	AuthZStore  store.AuthorizationCodeStoreInterface
	UserService user.UserServiceInterface
}

// newAuthorizationCodeGrantHandler creates a new instance of AuthorizationCodeGrantHandler.
func newAuthorizationCodeGrantHandler() GrantHandlerInterface {
	return &authorizationCodeGrantHandler{
		JWTService:  jwt.GetJWTService(),
		AuthZStore:  store.NewAuthorizationCodeStore(),
		UserService: user.GetUserService(),
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

	// Validate the authorization code.
	if tokenRequest.Code == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Authorization code is required",
		}
	}

	return nil
}

// HandleGrant processes the authorization code grant request and generates a token response.
func (h *authorizationCodeGrantHandler) HandleGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO, ctx *model.TokenContext) (
	*model.TokenResponseDTO, *model.ErrorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthorizationCodeGrantHandler"))

	authCode, err := h.AuthZStore.GetAuthorizationCode(tokenRequest.ClientID, tokenRequest.Code)
	if err != nil || authCode.Code == "" {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Invalid authorization code",
		}
	}

	// Validate the retrieved authorization code.
	errResponse := validateAuthorizationCode(tokenRequest, authCode)
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

	// Invalidate the authorization code after use.
	err = h.AuthZStore.DeactivateAuthorizationCode(authCode)
	if err != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to invalidate authorization code",
		}
	}

	// Get authorized scopes from the authorization code
	authorizedScopesStr := strings.TrimSpace(authCode.Scopes)
	authorizedScopes := []string{}
	if authorizedScopesStr != "" {
		authorizedScopes = strings.Split(authorizedScopesStr, " ")
	}

	// Generate a JWT token for the client
	jwtClaims := make(map[string]interface{})
	if authorizedScopesStr != "" {
		jwtClaims["scope"] = authorizedScopesStr
	}

	// Get token configuration from OAuth app
	iss := ""
	validityPeriod := int64(0)
	if oauthApp.Token != nil && oauthApp.Token.AccessToken != nil {
		iss = oauthApp.Token.AccessToken.Issuer
		validityPeriod = oauthApp.Token.AccessToken.ValidityPeriod
	}
	if iss == "" {
		iss = config.GetThunderRuntime().Config.OAuth.JWT.Issuer
	}
	if validityPeriod == 0 {
		validityPeriod = config.GetThunderRuntime().Config.OAuth.JWT.ValidityPeriod
	}

	userAttributes := map[string]interface{}{}
	if len(oauthApp.Token.AccessToken.UserAttributes) > 0 {
		user, svcErr := h.UserService.GetUser(authCode.AuthorizedUserID)
		if svcErr != nil {
			logger.Error("Failed to fetch user attributes",
				log.String("userID", authCode.AuthorizedUserID), log.Any("error", svcErr))
			return nil, &model.ErrorResponse{
				Error:            constants.ErrorServerError,
				ErrorDescription: "Something went wrong",
			}
		}

		var attrs map[string]interface{}
		if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
			logger.Error("Failed to unmarshal user attributes", log.String("userID", authCode.AuthorizedUserID),
				log.Error(err))
			return nil, &model.ErrorResponse{
				Error:            constants.ErrorServerError,
				ErrorDescription: "Something went wrong",
			}
		}

		for _, attr := range oauthApp.Token.AccessToken.UserAttributes {
			if val, ok := attrs[attr]; ok {
				jwtClaims[attr] = val
				userAttributes[attr] = val
			}
		}
	}

	token, _, err := h.JWTService.GenerateJWT(authCode.AuthorizedUserID, authCode.ClientID,
		iss, validityPeriod, jwtClaims)
	if err != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate token",
		}
	}

	// Add context attributes.
	if ctx.TokenAttributes == nil {
		ctx.TokenAttributes = make(map[string]interface{})
	}
	ctx.TokenAttributes["sub"] = authCode.AuthorizedUserID
	ctx.TokenAttributes["aud"] = authCode.ClientID

	// Prepare the token response.
	accessToken := &model.TokenDTO{
		Token:          token,
		TokenType:      constants.TokenTypeBearer,
		IssuedAt:       time.Now().Unix(),
		ExpiresIn:      3600,
		Scopes:         authorizedScopes,
		ClientID:       tokenRequest.ClientID,
		UserAttributes: userAttributes,
	}

	return &model.TokenResponseDTO{
		AccessToken: *accessToken,
	}, nil
}

// validateAuthorizationCode validates the authorization code against the token request.
func validateAuthorizationCode(tokenRequest *model.TokenRequest,
	code authzmodel.AuthorizationCode) *model.ErrorResponse {
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

	if code.State == authzconstants.AuthCodeStateInactive {
		// TODO: Revoke all the tokens issued for this authorization code.

		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Inactive authorization code",
		}
	} else if code.State != authzconstants.AuthCodeStateActive {
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
