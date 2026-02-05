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
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	oauth2utils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const serviceLoggerComponentName = "UserInfoService"

type userInfoServiceInterface interface {
	GetUserInfo(accessToken string) (map[string]interface{}, *serviceerror.ServiceError)
}

type userInfoService struct {
	jwtService         jwt.JWTServiceInterface
	applicationService application.ApplicationServiceInterface
	userService        user.UserServiceInterface
	logger             *log.Logger
}

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

func (s *userInfoService) GetUserInfo(accessToken string) (map[string]interface{}, *serviceerror.ServiceError) {
	if accessToken == "" {
		return nil, &errorInvalidAccessToken
	}
	tokenClaims, svcErr := s.validateAndDecodeToken(accessToken)
	if svcErr != nil {
		return nil, svcErr
	}
	sub, svcErr := s.extractSubClaim(tokenClaims)
	if svcErr != nil {
		return nil, svcErr
	}
	if svcErr := s.validateGrantType(tokenClaims); svcErr != nil {
		return nil, svcErr
	}
	scopes := s.extractScopes(tokenClaims)
	if len(scopes) == 0 {
		return map[string]interface{}{"sub": sub}, nil
	}
	oauthApp := s.getOAuthApp(tokenClaims)
	includeGroups := oauthApp != nil && oauthApp.Token != nil && oauthApp.Token.IDToken != nil &&
		slices.Contains(oauthApp.Token.IDToken.UserAttributes, constants.UserAttributeGroups)
	userAttributes, userGroups, err := tokenservice.FetchUserAttributesAndGroups(s.userService, sub, includeGroups)
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
	response, svcErr := s.buildUserInfoResponse(sub, scopes, userAttributes, oauthApp, tokenClaims)
	if svcErr != nil {
		return nil, svcErr
	}
	return response, nil
}

func (s *userInfoService) validateAndDecodeToken(accessToken string) (map[string]interface{}, *serviceerror.ServiceError) {
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

func (s *userInfoService) extractSubClaim(claims map[string]interface{}) (string, *serviceerror.ServiceError) {
	sub, ok := claims[constants.ClaimSub].(string)
	if !ok || sub == "" {
		return "", &errorMissingSubClaim
	}
	return sub, nil
}

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
		s.logger.Debug("UserInfo endpoint called with client_credentials grant token", log.String("grant_type", grantTypeString))
		return &errorClientCredentialsNotSupported
	}
	return nil
}

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

func (s *userInfoService) buildUserInfoResponse(sub string, scopes []string, userAttributes map[string]interface{},
	oauthApp *appmodel.OAuthAppConfigProcessedDTO, tokenClaims map[string]interface{}) (map[string]interface{}, *serviceerror.ServiceError) {
	response := map[string]interface{}{"sub": sub}
	claimsRequest, svcErr := s.extractClaimsRequest(tokenClaims)
	if svcErr != nil {
		return nil, svcErr
	}
	var userInfoClaims map[string]*model.IndividualClaimRequest
	if claimsRequest != nil {
		userInfoClaims = claimsRequest.UserInfo
	}
	claimData := tokenservice.BuildClaims(scopes, userInfoClaims, userAttributes, oauthApp)
	for key, value := range claimData {
		response[key] = value
	}
	return response, nil
}

func (s *userInfoService) extractClaimsRequest(tokenClaims map[string]interface{}) (*model.ClaimsRequest, *serviceerror.ServiceError) {
	claimsRequestStr, ok := tokenClaims[constants.ClaimClaimsRequest].(string)
	if !ok || claimsRequestStr == "" {
		return nil, nil
	}
	claimsRequest, err := oauth2utils.ParseClaimsRequest(claimsRequestStr)
	if err != nil {
		s.logger.Error("Failed to parse claims request from access token", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	return claimsRequest, nil
}
