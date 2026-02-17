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
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const serviceLoggerComponentName = "UserInfoService"

// userInfoServiceInterface defines the interface for OIDC UserInfo endpoint.
type userInfoServiceInterface interface {
	GetUserInfo(accessToken string) (*UserInfoResponse, *serviceerror.ServiceError)
}

// userInfoService implements the userInfoServiceInterface.
type userInfoService struct {
	jwtService         jwt.JWTServiceInterface
	applicationService application.ApplicationServiceInterface
	userService        user.UserServiceInterface
	ouService          ou.OrganizationUnitServiceInterface
	logger             *log.Logger
}

// newUserInfoService creates a new userInfoService instance.
func newUserInfoService(
	jwtService jwt.JWTServiceInterface,
	applicationService application.ApplicationServiceInterface,
	userService user.UserServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
) userInfoServiceInterface {
	return &userInfoService{
		jwtService:         jwtService,
		applicationService: applicationService,
		userService:        userService,
		ouService:          ouService,
		logger:             log.GetLogger().With(log.String(log.LoggerKeyComponentName, serviceLoggerComponentName)),
	}
}

// GetUserInfo validates the access token and returns user information based on authorized scopes.
func (s *userInfoService) GetUserInfo(accessToken string) (*UserInfoResponse, *serviceerror.ServiceError) {
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

	// Validate that the 'openid' scope is present
	if svcErr := s.validateOpenIDScope(scopes); svcErr != nil {
		return nil, svcErr
	}

	oauthApp := s.getOAuthApp(tokenClaims)

	// Extract allowed user attributes
	var allowedUserAttributes []string
	if oauthApp != nil && oauthApp.UserInfo != nil {
		allowedUserAttributes = oauthApp.UserInfo.UserAttributes
	}

	// Fetch user attributes with groups and default claims
	userAttributes, err := tokenservice.FetchUserAttributes(s.userService, s.ouService,
		sub, allowedUserAttributes)
	if err != nil {
		s.logger.Error("Failed to fetch user attributes", log.String("userID", sub), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	response, svcErr := s.buildUserInfoResponse(sub, scopes, userAttributes, oauthApp, tokenClaims)
	if svcErr != nil {
		return nil, svcErr
	}

	// Decide response type
	responseType := appmodel.UserInfoResponseTypeJSON
	if oauthApp != nil && oauthApp.UserInfo != nil {
		responseType = oauthApp.UserInfo.ResponseType
	}

	if responseType == appmodel.UserInfoResponseTypeJWS {
		return s.generateJWSUserInfo(sub, tokenClaims, response)
	}

	return &UserInfoResponse{
		Type:     appmodel.UserInfoResponseTypeJSON,
		JSONBody: response,
	}, nil
}

// generateJWSUserInfo creates a signed JWT UserInfo response
// based on the application configuration.
func (s *userInfoService) generateJWSUserInfo(
	sub string,
	tokenClaims map[string]interface{},
	response map[string]interface{},
) (*UserInfoResponse, *serviceerror.ServiceError) {
	clientID := ""
	if cid, ok := tokenClaims["client_id"].(string); ok {
		clientID = cid
	}

	runtime := config.GetThunderRuntime()

	issuer := runtime.Config.JWT.Issuer
	validity := runtime.Config.JWT.ValidityPeriod

	signedJWT, _, err := s.jwtService.GenerateJWT(
		sub,
		clientID,
		issuer,
		validity,
		response,
	)
	if err != nil {
		s.logger.Error("Failed to generate signed UserInfo JWT")
		return nil, &serviceerror.InternalServerError
	}

	return &UserInfoResponse{
		Type:    appmodel.UserInfoResponseTypeJWS,
		JWTBody: signedJWT,
	}, nil
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

// validateOpenIDScope validates that the access token contains the required 'openid' scope.
func (s *userInfoService) validateOpenIDScope(scopes []string) *serviceerror.ServiceError {
	if !slices.Contains(scopes, "openid") {
		s.logger.Debug("UserInfo request missing required 'openid' scope",
			log.String("scopes", tokenservice.JoinScopes(scopes)))
		return &errorInsufficientScope
	}
	return nil
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
// It also processes any explicit claims request embedded in the access token.
func (s *userInfoService) buildUserInfoResponse(
	sub string,
	scopes []string,
	userAttributes map[string]interface{},
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
	tokenClaims map[string]interface{},
) (map[string]interface{}, *serviceerror.ServiceError) {
	response := map[string]interface{}{
		"sub": sub,
	}

	// Build claims from scopes and explicit claims request
	// Extract only the UserInfo claims map from the access token
	claimsRequest, svcErr := s.extractClaimsRequest(tokenClaims)
	if svcErr != nil {
		return nil, svcErr
	}
	var userInfoClaims map[string]*model.IndividualClaimRequest
	if claimsRequest != nil {
		userInfoClaims = claimsRequest.UserInfo
	}

	// Get scope claims mapping and allowed user attributes from app config
	var scopeClaimsMapping map[string][]string
	var allowedUserAttributes []string
	if oauthApp != nil {
		scopeClaimsMapping = oauthApp.ScopeClaims
		if oauthApp.UserInfo != nil && len(oauthApp.UserInfo.UserAttributes) > 0 {
			allowedUserAttributes = oauthApp.UserInfo.UserAttributes
		}
	}

	claimData := tokenservice.BuildClaims(
		scopes,
		userInfoClaims,
		userAttributes,
		scopeClaimsMapping,
		allowedUserAttributes,
	)
	for key, value := range claimData {
		response[key] = value
	}

	return response, nil
}

// extractClaimsRequest extracts the claims request from the access token if present.
func (s *userInfoService) extractClaimsRequest(
	tokenClaims map[string]interface{},
) (*model.ClaimsRequest, *serviceerror.ServiceError) {
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
