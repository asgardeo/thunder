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

// Package github implements an authentication service for authentication via GitHub OAuth.
package github

import (
	"context"
	"slices"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	syshttp "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	loggerComponentName = "GithubAuthnService"
)

// GithubOAuthAuthnServiceInterface defines the contract for GitHub OAuth based authenticator services.
type GithubOAuthAuthnServiceInterface interface {
	authnoauth.OAuthAuthnCoreServiceInterface
}

// githubOAuthAuthnService is the default implementation of GithubOAuthAuthnServiceInterface.
type githubOAuthAuthnService struct {
	internal   authnoauth.OAuthAuthnServiceInterface
	httpClient syshttp.HTTPClientInterface
	logger     *log.Logger
}

// newGithubOAuthAuthnService creates a new instance of GitHub OAuth authenticator service.
func newGithubOAuthAuthnService(idpSvc idp.IDPServiceInterface,
	userSvc user.UserServiceInterface) GithubOAuthAuthnServiceInterface {
	httpClient := syshttp.NewHTTPClient()
	internal := authnoauth.NewOAuthAuthnService(httpClient, idpSvc, userSvc)

	service := &githubOAuthAuthnService{
		internal:   internal,
		httpClient: httpClient,
		logger:     log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
	authncm.RegisterAuthenticator(service.getMetadata())

	return service
}

// BuildAuthorizeURL constructs the authorization request URL for GitHub OAuth authentication.
func (g *githubOAuthAuthnService) BuildAuthorizeURL(ctx context.Context, idpID string) (
	string, *serviceerror.ServiceError) {
	return g.internal.BuildAuthorizeURL(ctx, idpID)
}

// ExchangeCodeForToken exchanges the authorization code for a token with GitHub.
func (g *githubOAuthAuthnService) ExchangeCodeForToken(ctx context.Context, idpID, code string,
	validateResponse bool) (*authnoauth.TokenResponse, *serviceerror.ServiceError) {
	return g.internal.ExchangeCodeForToken(ctx, idpID, code, validateResponse)
}

// FetchUserInfo retrieves user information from the Github API, ensuring email resolution if necessary.
func (g *githubOAuthAuthnService) FetchUserInfo(ctx context.Context, idpID, accessToken string) (
	map[string]interface{}, *serviceerror.ServiceError) {
	logger := g.logger
	oAuthClientConfig, svcErr := g.internal.GetOAuthClientConfig(ctx, idpID)
	if svcErr != nil {
		return nil, svcErr
	}

	userInfo, svcErr := g.internal.FetchUserInfoWithClientConfig(ctx, oAuthClientConfig, accessToken)
	if svcErr != nil {
		return userInfo, svcErr
	}

	// If email is already present in the user info or email scope is not requested, return it directly.
	email := authnoauth.GetStringUserClaimValue(userInfo, "email")
	if email != "" || !g.shouldFetchEmail(oAuthClientConfig.Scopes) {
		logger.Debug("Email is already present in the user info or email scope not requested")
		authnoauth.ProcessSubClaim(userInfo)
		return userInfo, nil
	}

	// Fetch primary email from the GitHub user emails endpoint.
	primaryEmail, svcErr := g.fetchPrimaryEmail(oAuthClientConfig, accessToken)
	if svcErr != nil {
		return nil, svcErr
	}
	if primaryEmail != "" {
		userInfo["email"] = primaryEmail
	}

	authnoauth.ProcessSubClaim(userInfo)
	return userInfo, nil
}

// shouldFetchEmail check whether user email should be fetched from the emails endpoint based on the scopes.
func (g *githubOAuthAuthnService) shouldFetchEmail(scopes []string) bool {
	return slices.Contains(scopes, UserScope) || slices.Contains(scopes, UserEmailScope)
}

// fetchPrimaryEmail fetches the primary email of the user from the GitHub user emails endpoint.
func (g *githubOAuthAuthnService) fetchPrimaryEmail(
	oAuthClientConfig *authnoauth.OAuthClientConfig, accessToken string) (
	string, *serviceerror.ServiceError) {
	logger := g.logger
	logger.Debug("Fetching primary email from GitHub user emails endpoint")

	if oAuthClientConfig.OAuthEndpoints.UserEmailEndpoint == "" {
		logger.Error("User email endpoint is not configured in OAuth client config")
		return "", &serviceerror.InternalServerError
	}

	req, svcErr := buildUserEmailRequest(oAuthClientConfig.OAuthEndpoints.UserEmailEndpoint, accessToken, logger)
	if svcErr != nil {
		return "", svcErr
	}

	emails, svcErr := sendUserEmailRequest(req, g.httpClient, logger)
	if svcErr != nil {
		return "", svcErr
	}

	for _, emailEntry := range emails {
		if isPrimary, ok := emailEntry["primary"].(bool); ok && isPrimary {
			if primaryEmail, ok := emailEntry["email"].(string); ok {
				return primaryEmail, nil
			}
		}
	}

	return "", nil
}

// GetInternalUser retrieves the internal user based on the external subject identifier.
func (g *githubOAuthAuthnService) GetInternalUser(ctx context.Context, sub string) (
	*user.User, *serviceerror.ServiceError) {
	return g.internal.GetInternalUser(ctx, sub)
}

// GetOAuthClientConfig retrieves and validates the OAuth client configuration for the given identity provider ID.
func (g *githubOAuthAuthnService) GetOAuthClientConfig(ctx context.Context, idpID string) (
	*authnoauth.OAuthClientConfig, *serviceerror.ServiceError) {
	return g.internal.GetOAuthClientConfig(ctx, idpID)
}

// getMetadata returns the authenticator metadata for GitHub OAuth authenticator.
func (g *githubOAuthAuthnService) getMetadata() authncm.AuthenticatorMeta {
	return authncm.AuthenticatorMeta{
		Name:          authncm.AuthenticatorGithub,
		Factors:       []authncm.AuthenticationFactor{authncm.FactorKnowledge},
		AssociatedIDP: idp.IDPTypeGitHub,
	}
}
