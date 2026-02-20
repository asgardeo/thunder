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

// Package oauth implements an authentication service for authenticating via an OAuth 2.0 based identity provider.
package oauth

import (
	"strings"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/idp"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	syshttp "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userprovider"
)

const (
	loggerComponentName = "OAuthAuthnService"
)

// OAuthAuthnCoreServiceInterface defines the core contract for OAuth based authenticator services.
type OAuthAuthnCoreServiceInterface interface {
	BuildAuthorizeURL(idpID string) (string, *serviceerror.ServiceError)
	ExchangeCodeForToken(idpID, code string, validateResponse bool) (*TokenResponse, *serviceerror.ServiceError)
	FetchUserInfo(idpID, accessToken string) (map[string]interface{}, *serviceerror.ServiceError)
	GetInternalUser(sub string) (*userprovider.User, *serviceerror.ServiceError)
	GetOAuthClientConfig(idpID string) (*OAuthClientConfig, *serviceerror.ServiceError)
}

// OAuthAuthnServiceInterface defines the contract for OAuth based authenticator services.
type OAuthAuthnServiceInterface interface {
	OAuthAuthnCoreServiceInterface
	ValidateTokenResponse(idpID string, tokenResp *TokenResponse) *serviceerror.ServiceError
	FetchUserInfoWithClientConfig(oAuthClientConfig *OAuthClientConfig, accessToken string) (
		map[string]interface{}, *serviceerror.ServiceError)
}

// oAuthAuthnService is the default implementation of OAuthAuthnServiceInterface.
type oAuthAuthnService struct {
	httpClient   syshttp.HTTPClientInterface
	idpService   idp.IDPServiceInterface
	userProvider userprovider.UserProviderInterface
	logger       *log.Logger
}

// newOAuthAuthnService creates a new instance of OAuth authenticator service.
func newOAuthAuthnService(httpClient syshttp.HTTPClientInterface,
	idpSvc idp.IDPServiceInterface, userProvider userprovider.UserProviderInterface,
) OAuthAuthnServiceInterface {
	service := &oAuthAuthnService{
		httpClient:   httpClient,
		idpService:   idpSvc,
		userProvider: userProvider,
		logger:       log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
	common.RegisterAuthenticator(service.getMetadata())

	return service
}

// NewOAuthAuthnService creates a new instance of OAuth authenticator service.
// [Deprecated: use dependency injection to get the instance instead].
// TODO: Should be removed when executors are migrated to di pattern.
func NewOAuthAuthnService(httpClient syshttp.HTTPClientInterface,
	idpSvc idp.IDPServiceInterface, userProvider userprovider.UserProviderInterface) OAuthAuthnServiceInterface {
	return newOAuthAuthnService(httpClient, idpSvc, userProvider)
}

// GetOAuthClientConfig retrieves the OAuth client configuration for the given identity provider ID.
func (s *oAuthAuthnService) GetOAuthClientConfig(idpID string) (
	*OAuthClientConfig, *serviceerror.ServiceError) {
	logger := s.logger.With(log.String("idpId", idpID))
	if strings.TrimSpace(idpID) == "" {
		return nil, &ErrorEmptyIdpID
	}

	idp, svcErr := s.idpService.GetIdentityProvider(idpID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			return nil, serviceerror.CustomServiceError(ErrorClientErrorWhileRetrievingIDP,
				"Error while retrieving identity provider: "+svcErr.ErrorDescription)
		}
		logger.Error("Error while retrieving identity provider", log.String("errorCode", svcErr.Code),
			log.String("description", svcErr.ErrorDescription))
		return nil, &serviceerror.InternalServerError
	}
	if idp == nil {
		return nil, &ErrorInvalidIDP
	}

	oAuthClientConfig, err := parseIDPConfig(idp)
	if err != nil {
		logger.Error("Failed to parse identity provider configurations", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return oAuthClientConfig, nil
}

// BuildAuthorizeURL constructs the authorization request URL for the external identity provider.
func (s *oAuthAuthnService) BuildAuthorizeURL(idpID string) (string, *serviceerror.ServiceError) {
	logger := s.logger.With(log.String("idpId", idpID))
	logger.Debug("Building authorize URL")

	oAuthClientConfig, svcErr := s.GetOAuthClientConfig(idpID)
	if svcErr != nil {
		return "", svcErr
	}
	if oAuthClientConfig.OAuthEndpoints.AuthorizationEndpoint == "" {
		logger.Error("Authorization endpoint is not configured for the identity provider")
		return "", &serviceerror.InternalServerError
	}

	queryParams := map[string]string{
		oauth2const.RequestParamClientID:     oAuthClientConfig.ClientID,
		oauth2const.RequestParamRedirectURI:  oAuthClientConfig.RedirectURI,
		oauth2const.RequestParamResponseType: oauth2const.RequestParamCode,
	}
	if len(oAuthClientConfig.Scopes) > 0 {
		queryParams[oauth2const.RequestParamScope] = sysutils.StringifyStringArray(oAuthClientConfig.Scopes, " ")
	}

	for key, value := range oAuthClientConfig.AdditionalParams {
		if key == "" || value == "" {
			continue
		}
		queryParams[key] = value
	}

	authZURL, err := sysutils.GetURIWithQueryParams(oAuthClientConfig.OAuthEndpoints.AuthorizationEndpoint,
		queryParams)
	if err != nil {
		logger.Error("Failed to build authorize URL", log.Error(err))
		return "", &serviceerror.InternalServerError
	}

	return authZURL, nil
}

// ExchangeCodeForToken exchanges the authorization code for a token with the external identity provider
// and validates the token response if validateResponse is true.
func (s *oAuthAuthnService) ExchangeCodeForToken(idpID, code string, validateResponse bool) (
	*TokenResponse, *serviceerror.ServiceError) {
	logger := s.logger.With(log.String("idpId", idpID))
	logger.Debug("Exchanging authorization code for token")

	if strings.TrimSpace(code) == "" {
		return nil, &ErrorEmptyAuthorizationCode
	}

	oAuthClientConfig, svcErr := s.GetOAuthClientConfig(idpID)
	if svcErr != nil {
		return nil, svcErr
	}
	if oAuthClientConfig.OAuthEndpoints.TokenEndpoint == "" {
		logger.Error("Token endpoint is not configured for the identity provider")
		return nil, &serviceerror.InternalServerError
	}

	httpReq, svcErr := buildTokenRequest(oAuthClientConfig, code, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	tokenResp, svcErr := sendTokenRequest(httpReq, s.httpClient, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	if validateResponse {
		svcErr = s.ValidateTokenResponse(idpID, tokenResp)
		if svcErr != nil {
			return nil, svcErr
		}
	}

	return tokenResp, nil
}

// ValidateTokenResponse validates the token response returned by the identity provider.
// ExchangeCodeForToken method calls this method to validate the token response if validateResponse is set
// to true. Hence generally you may not need to call this method explicitly.
func (s *oAuthAuthnService) ValidateTokenResponse(idpID string, tokenResp *TokenResponse) *serviceerror.ServiceError {
	logger := s.logger.With(log.String("idpId", idpID))
	logger.Debug("Validating token response")

	if tokenResp == nil {
		logger.Debug("Empty token response received from identity provider")
		return &ErrorInvalidTokenResponse
	}
	if tokenResp.AccessToken == "" {
		logger.Debug("Access token is empty in the token response")
		return &ErrorInvalidTokenResponse
	}

	return nil
}

// FetchUserInfo retrieves user information from the external identity provider.
func (s *oAuthAuthnService) FetchUserInfo(idpID, accessToken string) (
	map[string]interface{}, *serviceerror.ServiceError) {
	oAuthClientConfig, svcErr := s.GetOAuthClientConfig(idpID)
	if svcErr != nil {
		return nil, svcErr
	}

	return s.FetchUserInfoWithClientConfig(oAuthClientConfig, accessToken)
}

// FetchUserInfoWithClientConfig retrieves user information using the provided OAuth client configuration.
func (s *oAuthAuthnService) FetchUserInfoWithClientConfig(oAuthClientConfig *OAuthClientConfig,
	accessToken string) (map[string]interface{}, *serviceerror.ServiceError) {
	logger := s.logger
	logger.Debug("Fetching user info")

	if strings.TrimSpace(accessToken) == "" {
		return nil, &ErrorEmptyAccessToken
	}

	if oAuthClientConfig.OAuthEndpoints.UserInfoEndpoint == "" {
		logger.Error("User info endpoint is not configured for the identity provider")
		return nil, &serviceerror.InternalServerError
	}

	httpReq, svcErr := buildUserInfoRequest(oAuthClientConfig.OAuthEndpoints.UserInfoEndpoint,
		accessToken, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	userInfo, svcErr := sendUserInfoRequest(httpReq, s.httpClient, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	ProcessSubClaim(userInfo)
	return userInfo, nil
}

// GetInternalUser retrieves the internal user based on the external subject identifier.
func (s *oAuthAuthnService) GetInternalUser(sub string) (*userprovider.User, *serviceerror.ServiceError) {
	logger := s.logger.With(log.String("sub", log.MaskString(sub)))
	logger.Debug("Retrieving internal user for the given sub claim")

	if strings.TrimSpace(sub) == "" {
		return nil, &ErrorEmptySubClaim
	}

	filters := map[string]interface{}{
		"sub": sub,
	}
	userID, upErr := s.userProvider.IdentifyUser(filters)
	if upErr != nil {
		if upErr.Code == userprovider.ErrorCodeUserNotFound {
			logger.Debug("No user found for the provided sub claim")
			return nil, &common.ErrorUserNotFound
		}
		logger.Error("Error while identifying user", log.String("errorCode", string(upErr.Code)),
			log.String("description", upErr.Description))
		return nil, &serviceerror.InternalServerError
	}

	if userID == nil {
		logger.Debug("User id is nil, no user found for the provided sub claim")
		return nil, &common.ErrorUserNotFound
	}

	user, upErr := s.userProvider.GetUser(*userID)
	if upErr != nil {
		if upErr.Code == userprovider.ErrorCodeUserNotFound {
			return nil, &common.ErrorUserNotFound
		}
		logger.Error("Error while retrieving user", log.String("errorCode", string(upErr.Code)),
			log.String("description", upErr.Description))
		return nil, &serviceerror.InternalServerError
	}

	return user, nil
}

// getMetadata returns the authenticator metadata for OAuth authenticator.
func (s *oAuthAuthnService) getMetadata() common.AuthenticatorMeta {
	return common.AuthenticatorMeta{
		Name:          common.AuthenticatorOAuth,
		Factors:       []common.AuthenticationFactor{common.FactorKnowledge},
		AssociatedIDP: idp.IDPTypeOAuth,
	}
}
