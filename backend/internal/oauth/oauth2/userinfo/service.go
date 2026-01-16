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

// Package userinfo provides functionality for the OIDC UserInfo endpoint.
package userinfo

import (
	"slices"

	"github.com/asgardeo/thunder/internal/application"
	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const serviceLoggerComponentName = "UserInfoService"

// userInfoServiceInterface defines the interface for OIDC UserInfo endpoint.
type userInfoServiceInterface interface {
	GetUserInfo(accessToken string) (map[string]interface{}, *serviceerror.ServiceError)
}

// userInfoService implements the userInfoServiceInterface.
type userInfoService struct {
	jwtService         jwt.JWTServiceInterface
	applicationService application.ApplicationServiceInterface
	userService        user.UserServiceInterface
	logger             *log.Logger
}

// newUserInfoService creates a new userInfoService instance.
func newUserInfoService(
	jwtService jwt.JWTServiceInterface,
	applicationService application.ApplicationServiceInterface,
	userService user.UserServiceInterface,
) userInfoServiceInterface {
	return &userInfoService{
		jwtService:         jwtService,
		applicationService: applicationService,
		userService:        userService,
		logger:             log.GetLogger().With(log.String(log.LoggerKeyComponentName, serviceLoggerComponentName)),
	}
}

// GetUserInfo validates the access token and returns user information based on authorized scopes.
func (s *userInfoService) GetUserInfo(accessToken string) (map[string]interface{}, *serviceerror.ServiceError) {
	if accessToken == "" {
		return nil, &errorInvalidAccessToken
	}

	claims, svcErr := s.validateAndDecodeToken(accessToken)
	if svcErr != nil {
		return nil, svcErr
	}

	sub, svcErr := s.extractSubClaim(claims)
	if svcErr != nil {
		return nil, svcErr
	}

	if svcErr := s.validateGrantType(claims); svcErr != nil {
		return nil, svcErr
	}

	scopes := s.extractScopes(claims)
	if len(scopes) == 0 {
		return map[string]interface{}{"sub": sub}, nil
	}

	oauthApp := s.getOAuthApp(claims)

	includeGroups := oauthApp != nil &&
		oauthApp.Token != nil &&
		oauthApp.Token.IDToken != nil &&
		slices.Contains(oauthApp.Token.IDToken.UserAttributes, constants.UserAttributeGroups)

	userAttributes, userGroups, err := tokenservice.FetchUserAttributesAndGroups(s.userService,
		sub, includeGroups)
	if err != nil {
		s.logger.Error("Failed to fetch user attributes", log.String("userID", sub), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if len(userGroups) > 0 && includeGroups {
		if userAttributes == nil {
			userAttributes = make(map[string]interface{})
		}
		userAttributes[constants.UserAttributeGroups] = userGroups
	}

	return s.buildUserInfoResponse(sub, scopes, userAttributes, oauthApp), nil
}

// validateAndDecodeToken validates the JWT signature and decodes the payload.
func (s *userInfoService) validateAndDecodeToken(accessToken string) (
	map[string]interface{}, *serviceerror.ServiceError) {
	if err := s.jwtService.VerifyJWT(accessToken, "", ""); err != nil {
		s.logger.Debug("Failed to verify access token", log.String("error", err.Error))
		return nil, &errorInvalidAccessToken
	}

	claims, err := jwt.DecodeJWTPayload(accessToken)
	if err != nil {
		s.logger.Debug("Failed to decode access token", log.Error(err))
		return nil, &errorInvalidAccessToken
	}

	return claims, nil
}

// extractSubClaim extracts and validates the sub claim from the token claims.
func (s *userInfoService) extractSubClaim(claims map[string]interface{}) (string, *serviceerror.ServiceError) {
	sub, ok := claims[constants.ClaimSub].(string)
	if !ok || sub == "" {
		return "", &errorMissingSubClaim
	}
	return sub, nil
}

// validateGrantType validates that the token was not issued using client_credentials grant.
func (s *userInfoService) validateGrantType(claims map[string]interface{}) *serviceerror.ServiceError {
	grantTypeValue, ok := claims["grant_type"]
	if !ok {
		return nil
	}

	grantTypeString, ok := grantTypeValue.(string)
	if !ok {
		return nil
	}

	if constants.GrantType(grantTypeString) == constants.GrantTypeClientCredentials {
		s.logger.Debug("UserInfo endpoint called with client_credentials grant token",
			log.String("grant_type", grantTypeString))
		return &errorClientCredentialsNotSupported
	}

	return nil
}

// extractScopes extracts scopes from the token claims.
func (s *userInfoService) extractScopes(claims map[string]interface{}) []string {
	scopeValue, ok := claims["scope"]
	if !ok {
		return nil
	}

	scopeString, ok := scopeValue.(string)
	if !ok {
		return nil
	}

	return tokenservice.ParseScopes(scopeString)
}

// getOAuthApp retrieves the OAuth application configuration if client_id is present in claims.
func (s *userInfoService) getOAuthApp(claims map[string]interface{}) *appmodel.OAuthAppConfigProcessedDTO {
	clientID, ok := claims["client_id"].(string)
	if !ok || clientID == "" {
		return nil
	}

	app, err := s.applicationService.GetOAuthApplication(clientID)
	if err != nil || app == nil {
		return nil
	}

	return app
}

// buildUserInfoResponse builds the final UserInfo response from sub, scopes, and user attributes.
func (s *userInfoService) buildUserInfoResponse(
	sub string,
	scopes []string,
	userAttributes map[string]interface{},
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
) map[string]interface{} {
	scopeClaims := tokenservice.BuildOIDCClaimsFromScopes(
		scopes,
		userAttributes,
		oauthApp,
	)

	response := map[string]interface{}{
		"sub": sub,
	}

	for key, value := range scopeClaims {
		response[key] = value
	}

	return response
}
