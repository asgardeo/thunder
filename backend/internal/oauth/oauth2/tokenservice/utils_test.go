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

package tokenservice

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type UtilsTestSuite struct {
	suite.Suite
}

const (
	testTokenAud        = "https://token-aud.example.com" //nolint:gosec // Test data, not a real credential
	testDefaultAudience = "default-app"
)

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func (suite *UtilsTestSuite) SetupTest() {
	// Initialize Thunder Runtime for tests
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default.thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithNilOAuthApp() {
	// When oauthApp is nil, should return default issuer from config
	validIssuers := getValidIssuers(nil)

	assert.NotNil(suite.T(), validIssuers)
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithOnlyDefaultIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	assert.Len(suite.T(), validIssuers, 1)
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithTokenConfig() {
	// OAuthApp with a Token config should still use Thunder-level issuer from config
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token:    &appmodel.OAuthTokenConfig{},
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	assert.Len(suite.T(), validIssuers, 1)
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithAccessTokenConfig() {
	// OAuthApp with Token and AccessToken config should still use Thunder-level issuer from config
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{
				ValidityPeriod: 7200,
			},
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	assert.Len(suite.T(), validIssuers, 1)
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithIDTokenConfig() {
	// OAuthApp with Token and IDToken config should still use Thunder-level issuer from config
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				ValidityPeriod: 3600,
			},
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	assert.Len(suite.T(), validIssuers, 1)
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_AlwaysUsesThunderIssuer() {
	// Valid issuers always come from Thunder config, never empty strings
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{},
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
	assert.NotContains(suite.T(), validIssuers, "")
}

// ============================================================================
// validateIssuer Tests
// ============================================================================

func (suite *UtilsTestSuite) TestvalidateIssuer_WithValidDefaultIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	err := validateIssuer("https://thunder.io", oauthApp)

	assert.NoError(suite.T(), err)
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithThunderIssuerAndTokenConfig() {
	// Thunder-level issuer is always valid regardless of token config presence
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token:    &appmodel.OAuthTokenConfig{},
	}

	err := validateIssuer("https://thunder.io", oauthApp)

	assert.NoError(suite.T(), err)
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithThunderIssuerAndAccessTokenConfig() {
	// Thunder-level issuer is always valid regardless of access token config presence
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{
				ValidityPeriod: 3600,
			},
		},
	}

	err := validateIssuer("https://thunder.io", oauthApp)

	assert.NoError(suite.T(), err)
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithInvalidIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	err := validateIssuer("https://evil.example.com", oauthApp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
	assert.Contains(suite.T(), err.Error(), "https://evil.example.com")
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithEmptyIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	err := validateIssuer("", oauthApp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithNilOAuthApp() {
	// Should still validate against default issuer from config
	err := validateIssuer("https://thunder.io", nil)

	assert.NoError(suite.T(), err)
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithNilOAuthAppInvalidIssuer() {
	err := validateIssuer("https://invalid.com", nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *UtilsTestSuite) TestFederationScenario_OnlyThunderIssuerIsValid() {
	// Only the Thunder-level issuer from config is accepted; app-level issuers are no longer supported
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{},
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	// Only the Thunder-level issuer from config is returned
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")

	// Validate the Thunder-level issuer passes
	assert.NoError(suite.T(), validateIssuer("https://thunder.io", oauthApp))

	// Should reject unknown issuers
	assert.Error(suite.T(), validateIssuer("https://thunder-staging.company.com", oauthApp))
	assert.Error(suite.T(), validateIssuer("https://unknown.company.com", oauthApp))
}

func (suite *UtilsTestSuite) TestFederationScenario_FutureExternalIssuerSupport() {
	// This test documents the intended behavior for future external issuer support
	// TODO: When external issuer support is added, update GetValidIssuers to include
	// external federated issuers from configuration

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	validIssuers := getValidIssuers(oauthApp)

	// Currently only the Thunder-level issuer from config is returned
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")

	// In the future, external issuers should also be included
	// assert.Contains(suite.T(), validIssuers, "https://external-idp.com")
}

func (suite *UtilsTestSuite) TestJoinScopes_WithMultipleScopes() {
	scopes := []string{"read", "write", "admin"}
	result := JoinScopes(scopes)

	assert.Equal(suite.T(), "read write admin", result)
}

func (suite *UtilsTestSuite) TestJoinScopes_WithSingleScope() {
	scopes := []string{"read"}
	result := JoinScopes(scopes)

	assert.Equal(suite.T(), "read", result)
}

func (suite *UtilsTestSuite) TestJoinScopes_WithEmptySlice() {
	scopes := []string{}
	result := JoinScopes(scopes)

	assert.Equal(suite.T(), "", result)
}

func (suite *UtilsTestSuite) TestJoinScopes_WithNilSlice() {
	scopes := []string(nil)
	result := JoinScopes(scopes)

	assert.Equal(suite.T(), "", result)
}

// ============================================================================
// DetermineAudience Tests
// ============================================================================

func (suite *UtilsTestSuite) TestDetermineAudience_WithAudience() {
	audience := "https://api.example.com"
	resource := "https://other-api.com"
	tokenAud := testTokenAud
	defaultAudience := testDefaultAudience

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), audience, result)
}

func (suite *UtilsTestSuite) TestDetermineAudience_WithResource() {
	audience := ""
	resource := "https://api.example.com"
	tokenAud := testTokenAud
	defaultAudience := testDefaultAudience

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), resource, result)
}

func (suite *UtilsTestSuite) TestDetermineAudience_WithTokenAud() {
	audience := ""
	resource := ""
	tokenAud := testTokenAud
	defaultAudience := testDefaultAudience

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), tokenAud, result)
}

func (suite *UtilsTestSuite) TestDetermineAudience_WithoutResource() {
	audience := ""
	resource := ""
	tokenAud := ""
	defaultAudience := testDefaultAudience

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), defaultAudience, result)
}

func (suite *UtilsTestSuite) TestDetermineAudience_EmptyDefault() {
	audience := ""
	resource := ""
	tokenAud := ""
	defaultAudience := ""

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), "", result)
}

// ============================================================================
// getStandardJWTClaims Tests
// ============================================================================

func (suite *UtilsTestSuite) TestgetStandardJWTClaims_ContainsAllStandardClaims() {
	claims := getStandardJWTClaims()

	assert.True(suite.T(), claims["sub"])
	assert.True(suite.T(), claims["iss"])
	assert.True(suite.T(), claims["aud"])
	assert.True(suite.T(), claims["exp"])
	assert.True(suite.T(), claims["nbf"])
	assert.True(suite.T(), claims["iat"])
	assert.True(suite.T(), claims["jti"])
	assert.True(suite.T(), claims["scope"])
	assert.True(suite.T(), claims["client_id"])
	assert.True(suite.T(), claims["act"])
}

func (suite *UtilsTestSuite) TestgetStandardJWTClaims_ReturnsNewMap() {
	claims1 := getStandardJWTClaims()
	claims2 := getStandardJWTClaims()

	// Should be independent - modifying one shouldn't affect the other
	claims1["test"] = true
	assert.NotContains(suite.T(), claims2, "test")
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_WithStandardClaimsOnly() {
	claims := map[string]interface{}{
		"sub":   "user123",
		"iss":   "https://thunder.io",
		"aud":   "app123",
		"exp":   1234567890,
		"scope": "read write",
	}

	result := ExtractUserAttributes(claims)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_WithCustomClaims() {
	claims := map[string]interface{}{
		"sub":    "user123",
		"iss":    "https://thunder.io",
		"aud":    "app123",
		"exp":    1234567890,
		"scope":  "read write",
		"name":   "John Doe",
		"email":  "john@example.com",
		"groups": []string{"admin", "user"},
	}

	result := ExtractUserAttributes(claims)

	assert.Equal(suite.T(), "John Doe", result["name"])
	assert.Equal(suite.T(), "john@example.com", result["email"])
	assert.Equal(suite.T(), []string{"admin", "user"}, result["groups"])
	assert.NotContains(suite.T(), result, "sub")
	assert.NotContains(suite.T(), result, "iss")
	assert.NotContains(suite.T(), result, "aud")
	assert.NotContains(suite.T(), result, "exp")
	assert.NotContains(suite.T(), result, "scope")
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_WithRefreshTokenSpecificClaims() {
	claims := map[string]interface{}{
		"sub":                          "user123",
		"iss":                          "https://thunder.io",
		"aud":                          "app123",
		"exp":                          1234567890,
		"scope":                        "read write",
		"grant_type":                   "authorization_code",
		"access_token_sub":             "user123",
		"access_token_aud":             "app123",
		"access_token_user_attributes": map[string]interface{}{"name": "John"},
		"name":                         "John Doe",
		"email":                        "john@example.com",
	}

	result := ExtractUserAttributes(claims)

	// Should include refresh token specific claims as they're not standard JWT claims
	assert.Equal(suite.T(), "John Doe", result["name"])
	assert.Equal(suite.T(), "john@example.com", result["email"])
	assert.Equal(suite.T(), "authorization_code", result["grant_type"])
	assert.Equal(suite.T(), "user123", result["access_token_sub"])
	assert.Equal(suite.T(), "app123", result["access_token_aud"])
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_EmptyClaims() {
	claims := map[string]interface{}{}

	result := ExtractUserAttributes(claims)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_NilClaims() {
	claims := map[string]interface{}(nil)

	result := ExtractUserAttributes(claims)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestextractInt64Claim_WithIntType() {
	claims := map[string]interface{}{
		"iat": int(1234567890),
	}

	result, err := extractInt64Claim(claims, "iat")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1234567890), result)
}

func (suite *UtilsTestSuite) TestextractInt64Claim_WithInt64Type() {
	claims := map[string]interface{}{
		"iat": int64(1234567890),
	}

	result, err := extractInt64Claim(claims, "iat")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1234567890), result)
}

func (suite *UtilsTestSuite) TestextractInt64Claim_WithInvalidType() {
	claims := map[string]interface{}{
		"iat": "not-a-number",
	}

	result, err := extractInt64Claim(claims, "iat")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), int64(0), result)
	assert.Contains(suite.T(), err.Error(), "not a number")
}

func (suite *UtilsTestSuite) TestParseScopes_WithMultipleSpaces() {
	scopeString := "read  write   admin"
	result := ParseScopes(scopeString)

	assert.Equal(suite.T(), []string{"read", "write", "admin"}, result)
}

func (suite *UtilsTestSuite) TestParseScopes_WithLeadingTrailingSpaces() {
	scopeString := "  read write  "
	result := ParseScopes(scopeString)

	assert.Equal(suite.T(), []string{"read", "write"}, result)
}

func (suite *UtilsTestSuite) TestParseScopes_WithSingleScope() {
	scopeString := "read"
	result := ParseScopes(scopeString)

	assert.Equal(suite.T(), []string{"read"}, result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithValidScope() {
	claims := map[string]interface{}{
		"scope": "read write admin",
	}

	result := extractScopesFromClaims(claims, false)

	assert.Equal(suite.T(), []string{"read", "write", "admin"}, result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithEmptyScopeString() {
	claims := map[string]interface{}{
		"scope": "", // Empty string
	}

	result := extractScopesFromClaims(claims, false)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithInvalidScopeType() {
	claims := map[string]interface{}{
		"scope": 12345, // Invalid type (not string)
	}

	result := extractScopesFromClaims(claims, false)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithNoScopeButAuthorizedPermissions_IsAuthAssertion() {
	claims := map[string]interface{}{
		"authorized_permissions": "read:documents write:documents",
	}

	result := extractScopesFromClaims(claims, true)

	assert.Equal(suite.T(), []string{"read:documents", "write:documents"}, result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithNoScopeButAuthorizedPermissions_NotAuthAssertion() {
	claims := map[string]interface{}{
		"authorized_permissions": "read:documents write:documents",
	}

	result := extractScopesFromClaims(claims, false)

	assert.Empty(suite.T(), result) // Should not use authorized_permissions when not auth assertion
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithEmptyScopeButAuthorizedPermissions_IsAuthAssertion() {
	claims := map[string]interface{}{
		"scope":                  "", // Empty scope
		"authorized_permissions": "read write",
	}

	result := extractScopesFromClaims(claims, true)

	assert.Equal(suite.T(), []string{"read", "write"}, result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithEmptyAuthorizedPermissions_IsAuthAssertion() {
	claims := map[string]interface{}{
		"authorized_permissions": "", // Empty string
	}

	result := extractScopesFromClaims(claims, true)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithInvalidAuthorizedPermissionsType_IsAuthAssertion() {
	claims := map[string]interface{}{
		"authorized_permissions": 12345, // Invalid type (not string)
	}

	result := extractScopesFromClaims(claims, true)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithNoScopeAndNoAuthorizedPermissions() {
	claims := map[string]interface{}{
		// No scope or authorized_permissions
	}

	result := extractScopesFromClaims(claims, true)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_ScopeTakesPriorityOverAuthorizedPermissions() {
	claims := map[string]interface{}{
		"scope":                  "openid profile",
		"authorized_permissions": "read:documents write:documents",
	}

	result := extractScopesFromClaims(claims, true)

	// Scope should take priority
	assert.Equal(suite.T(), []string{"openid", "profile"}, result)
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_GetUserError() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return error
	serverErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "USER_NOT_FOUND",
		ErrorDescription: "user not found",
	}
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(nil, serverErr)

	_, err := FetchUserAttributes(mockUserService, nil, "test-user", nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to fetch user")

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_UnmarshalError() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return user with invalid JSON in attributes
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:         "test-user",
		Attributes: json.RawMessage(`{invalid json}`), // Invalid JSON
		Type:       "local",
	}, nil)

	_, err := FetchUserAttributes(mockUserService, nil, "test-user", nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to unmarshal user attributes")

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_NilAttributes() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return user with nil attributes
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:         "test-user",
		Attributes: nil,
		Type:       "local",
	}, nil)

	allowedClaims := []string{constants.ClaimUserType}
	attrs, err := FetchUserAttributes(mockUserService, nil, "test-user", allowedClaims)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	assert.Equal(suite.T(), "local", attrs[constants.ClaimUserType])

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_EmptyAllowedClaims() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return user with full data
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:               "test-user",
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
		Type:             "local",
		OrganizationUnit: "ou-123",
	}, nil)

	// Empty allowedClaims - no special claims should be added
	attrs, err := FetchUserAttributes(mockUserService, nil, "test-user", []string{})

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	// Only user attributes should be present, not default claims
	assert.Equal(suite.T(), "test@example.com", attrs["email"])
	assert.Nil(suite.T(), attrs[constants.ClaimUserType])
	assert.Nil(suite.T(), attrs[constants.ClaimOUID])
	assert.Nil(suite.T(), attrs[constants.UserAttributeGroups])

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_NilAllowedClaims() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return user with full data
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:               "test-user",
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
		Type:             "local",
		OrganizationUnit: "ou-123",
	}, nil)

	// Nil allowedClaims - no special claims should be added
	attrs, err := FetchUserAttributes(mockUserService, nil, "test-user", nil)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	// Only user attributes should be present
	assert.Equal(suite.T(), "test@example.com", attrs["email"])
	assert.Nil(suite.T(), attrs[constants.ClaimUserType])
	assert.Nil(suite.T(), attrs[constants.ClaimOUID])

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_UserWithNoType() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return user without type
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:               "test-user",
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
		Type:             "", // Empty type
		OrganizationUnit: "ou-123",
	}, nil)

	allowedClaims := []string{constants.ClaimUserType, constants.ClaimOUID}
	attrs, err := FetchUserAttributes(mockUserService, nil, "test-user", allowedClaims)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	// userType should not be present since it's empty
	assert.Nil(suite.T(), attrs[constants.ClaimUserType])
	// ouId should still be present
	assert.Equal(suite.T(), "ou-123", attrs[constants.ClaimOUID])

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_UserWithNoOrganizationUnit() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return user without organization unit
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:               "test-user",
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
		Type:             "local",
		OrganizationUnit: "", // Empty OU
	}, nil)

	allowedClaims := []string{constants.ClaimUserType, constants.ClaimOUID}
	attrs, err := FetchUserAttributes(mockUserService, nil, "test-user", allowedClaims)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	// userType should be present
	assert.Equal(suite.T(), "local", attrs[constants.ClaimUserType])
	// ouId should not be present since OU is empty
	assert.Nil(suite.T(), attrs[constants.ClaimOUID])

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_WithOUServiceSuccess() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

	// Mock GetUser to return valid user with OU
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:               "test-user",
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
		Type:             "local",
		OrganizationUnit: "ou-123",
	}, nil)

	// Mock GetOrganizationUnit to return OU details
	mockOUService.On("GetOrganizationUnit", "ou-123").Return(ou.OrganizationUnit{
		ID:     "ou-123",
		Handle: "test-org",
		Name:   "Test Organization",
	}, nil)

	// Request all OU-related claims
	allowedClaims := []string{constants.ClaimOUID, constants.ClaimOUHandle, constants.ClaimOUName}
	attrs, err := FetchUserAttributes(mockUserService, mockOUService, "test-user", allowedClaims)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	assert.Equal(suite.T(), "ou-123", attrs[constants.ClaimOUID])
	assert.Equal(suite.T(), "test-org", attrs[constants.ClaimOUHandle])
	assert.Equal(suite.T(), "Test Organization", attrs[constants.ClaimOUName])

	mockUserService.AssertExpectations(suite.T())
	mockOUService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_WithOUServiceError() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

	// Mock GetUser to return valid user with OU
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:               "test-user",
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
		Type:             "local",
		OrganizationUnit: "ou-123",
	}, nil)

	// Mock GetOrganizationUnit to return error
	mockOUService.On("GetOrganizationUnit", "ou-123").Return(ou.OrganizationUnit{}, &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "OU_NOT_FOUND",
		ErrorDescription: "organization unit not found",
	})

	// Request ouHandle which requires OU service
	allowedClaims := []string{constants.ClaimOUHandle}
	_, err := FetchUserAttributes(mockUserService, mockOUService, "test-user", allowedClaims)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to fetch organization unit details")

	mockUserService.AssertExpectations(suite.T())
	mockOUService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_OUDetailsNotRequestedSkipsOUService() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

	// Mock GetUser to return valid user with OU
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:               "test-user",
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
		Type:             "local",
		OrganizationUnit: "ou-123",
	}, nil)

	// No mock for GetOrganizationUnit - it should NOT be called

	// Request only ouId (not ouHandle or ouName)
	allowedClaims := []string{constants.ClaimOUID}
	attrs, err := FetchUserAttributes(mockUserService, mockOUService, "test-user", allowedClaims)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	assert.Equal(suite.T(), "ou-123", attrs[constants.ClaimOUID])
	// ouHandle and ouName should not be present
	assert.Nil(suite.T(), attrs[constants.ClaimOUHandle])
	assert.Nil(suite.T(), attrs[constants.ClaimOUName])

	// Verify GetOrganizationUnit was NOT called
	mockOUService.AssertNotCalled(suite.T(), "GetOrganizationUnit", mock.Anything)
	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_NilOUServiceSkipsOUDetails() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return valid user with OU
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:               "test-user",
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
		Type:             "local",
		OrganizationUnit: "ou-123",
	}, nil)

	// Request ouHandle but pass nil ouService
	allowedClaims := []string{constants.ClaimOUID, constants.ClaimOUHandle}
	attrs, err := FetchUserAttributes(mockUserService, nil, "test-user", allowedClaims)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	// ouId should be present
	assert.Equal(suite.T(), "ou-123", attrs[constants.ClaimOUID])
	// ouHandle should NOT be present since ouService is nil
	assert.Nil(suite.T(), attrs[constants.ClaimOUHandle])

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_WithGroups() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return valid user
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:               "test-user",
		Attributes:       json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
		Type:             "local",
		OrganizationUnit: "ou-123",
	}, nil)

	// Mock GetUserGroups to return groups
	mockGroups := &user.UserGroupListResponse{
		TotalResults: 2,
		StartIndex:   0,
		Count:        2,
		Groups: []user.UserGroup{
			{ID: "group1", Name: "Admin"},
			{ID: "group2", Name: "Users"},
		},
	}
	mockUserService.On("GetUserGroups", mock.Anything, "test-user", constants.DefaultGroupListLimit, 0).
		Return(mockGroups, nil)

	// Include default claims and groups in allowed list
	allowedClaims := []string{
		"email",
		"username",
		constants.ClaimUserType,
		constants.ClaimOUID,
		constants.UserAttributeGroups,
	}
	attrs, err := FetchUserAttributes(mockUserService, nil, "test-user", allowedClaims)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	assert.Equal(suite.T(), "test@example.com", attrs["email"])
	assert.Equal(suite.T(), "testuser", attrs["username"])
	// Verify groups are merged into attrs
	assert.Equal(suite.T(), []string{"Admin", "Users"}, attrs[constants.UserAttributeGroups])
	// Verify default claims are merged into attrs
	assert.Equal(suite.T(), "local", attrs[constants.ClaimUserType])
	assert.Equal(suite.T(), "ou-123", attrs[constants.ClaimOUID])

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_WithEmptyGroups() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return valid user
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:         "test-user",
		Attributes: json.RawMessage(`{"email":"test@example.com"}`),
		Type:       "local",
	}, nil)

	// Mock GetUserGroups to return empty groups
	mockGroups := &user.UserGroupListResponse{
		TotalResults: 0,
		StartIndex:   0,
		Count:        0,
		Groups:       []user.UserGroup{},
	}
	mockUserService.On("GetUserGroups", mock.Anything, "test-user", constants.DefaultGroupListLimit, 0).
		Return(mockGroups, nil)

	// Request groups
	allowedClaims := []string{constants.UserAttributeGroups}
	attrs, err := FetchUserAttributes(mockUserService, nil, "test-user", allowedClaims)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	// Groups should not be added when empty
	assert.Nil(suite.T(), attrs[constants.UserAttributeGroups])

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_GetUserGroupsError() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return valid user
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:         "test-user",
		Attributes: json.RawMessage(`{"email":"test@example.com"}`),
		Type:       "local",
	}, nil)

	// Mock GetUserGroups to return error
	serverErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "INTERNAL_ERROR",
		ErrorDescription: "failed to fetch groups",
	}
	mockUserService.On("GetUserGroups", mock.Anything, "test-user", constants.DefaultGroupListLimit, 0).
		Return(nil, serverErr)

	// Request groups in allowed claims to trigger the error
	allowedClaims := []string{constants.UserAttributeGroups}
	_, err := FetchUserAttributes(mockUserService, nil, "test-user", allowedClaims)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to fetch user groups")

	mockUserService.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_WithoutGroups() {
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	// Mock GetUser to return valid user
	mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
		ID:         "test-user",
		Attributes: json.RawMessage(`{"email":"test@example.com"}`),
		Type:       "local",
	}, nil)

	// Include userType but NOT groups - so GetUserGroups should not be called
	allowedClaims := []string{"email", constants.ClaimUserType}
	attrs, err := FetchUserAttributes(mockUserService, nil, "test-user", allowedClaims)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrs)
	// Verify userType is in attrs
	assert.Equal(suite.T(), "local", attrs[constants.ClaimUserType])
	// Verify groups is not present
	assert.Nil(suite.T(), attrs[constants.UserAttributeGroups])

	// Verify GetUserGroups was NOT called
	mockUserService.AssertNotCalled(
		suite.T(), "GetUserGroups", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *UtilsTestSuite) TestFetchUserAttributes_SingleOUClaim() {
	// Test cases for requesting only one OU claim at a time
	testCases := []struct {
		name          string
		allowedClaims []string
		expectedClaim string
		expectedValue string
		absentClaim   string
	}{
		{
			name:          "OnlyOUHandle",
			allowedClaims: []string{constants.ClaimOUHandle},
			expectedClaim: constants.ClaimOUHandle,
			expectedValue: "test-org",
			absentClaim:   constants.ClaimOUName,
		},
		{
			name:          "OnlyOUName",
			allowedClaims: []string{constants.ClaimOUName},
			expectedClaim: constants.ClaimOUName,
			expectedValue: "Test Organization",
			absentClaim:   constants.ClaimOUHandle,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())
			mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

			mockUserService.On("GetUser", mock.Anything, "test-user").Return(&user.User{
				ID:               "test-user",
				Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
				OrganizationUnit: "ou-123",
			}, nil)

			mockOUService.On("GetOrganizationUnit", "ou-123").Return(ou.OrganizationUnit{
				ID:     "ou-123",
				Handle: "test-org",
				Name:   "Test Organization",
			}, nil)

			attrs, err := FetchUserAttributes(mockUserService, mockOUService, "test-user", tc.allowedClaims)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), attrs)
			assert.Equal(suite.T(), tc.expectedValue, attrs[tc.expectedClaim])
			assert.Nil(suite.T(), attrs[tc.absentClaim])

			mockUserService.AssertExpectations(suite.T())
			mockOUService.AssertExpectations(suite.T())
		})
	}
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_RefreshToken_WithServerLevelConfig() {
	// Reset and initialize config with refresh token validity period
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
		OAuth: config.OAuthConfig{
			RefreshToken: config.RefreshTokenConfig{
				ValidityPeriod: 86400, // 24 hours
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	result := resolveTokenConfig(oauthApp, TokenTypeRefresh)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(86400), result.ValidityPeriod)
	assert.Equal(suite.T(), "https://thunder.io", result.Issuer)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_RefreshToken_WithoutServerLevelConfig() {
	// Reset and initialize config without refresh token validity period (zero value)
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
		OAuth: config.OAuthConfig{
			RefreshToken: config.RefreshTokenConfig{
				ValidityPeriod: 0, // Not set
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	result := resolveTokenConfig(oauthApp, TokenTypeRefresh)

	assert.NotNil(suite.T(), result)
	// Should fallback to default JWT validity period
	assert.Equal(suite.T(), int64(3600), result.ValidityPeriod)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_RefreshToken_WithNilOAuthApp() {
	// Reset and initialize config with refresh token validity period
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
		OAuth: config.OAuthConfig{
			RefreshToken: config.RefreshTokenConfig{
				ValidityPeriod: 604800, // 7 days
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	// oauthApp is nil
	result := resolveTokenConfig(nil, TokenTypeRefresh)

	assert.NotNil(suite.T(), result)
	// Should still use server-level refresh token config
	assert.Equal(suite.T(), int64(604800), result.ValidityPeriod)
	assert.Equal(suite.T(), "https://thunder.io", result.Issuer)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_RefreshToken_WithTokenConfig() {
	// Refresh token always uses Thunder-level issuer from config
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
		OAuth: config.OAuthConfig{
			RefreshToken: config.RefreshTokenConfig{
				ValidityPeriod: 86400,
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token:    &appmodel.OAuthTokenConfig{},
	}

	result := resolveTokenConfig(oauthApp, TokenTypeRefresh)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(86400), result.ValidityPeriod)
	assert.Equal(suite.T(), "https://thunder.io", result.Issuer)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_AccessToken_WithNilOAuthApp() {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	// oauthApp is nil - should use default config
	result := resolveTokenConfig(nil, TokenTypeAccess)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(3600), result.ValidityPeriod)
	assert.Equal(suite.T(), "https://thunder.io", result.Issuer)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_AccessToken_WithNilToken() {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	// oauthApp.Token is nil - should use default config
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token:    nil,
	}

	result := resolveTokenConfig(oauthApp, TokenTypeAccess)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(3600), result.ValidityPeriod)
	assert.Equal(suite.T(), "https://thunder.io", result.Issuer)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_AccessToken_WithAppLevelConfig() {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{
				ValidityPeriod: 7200,
			},
		},
	}

	result := resolveTokenConfig(oauthApp, TokenTypeAccess)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(7200), result.ValidityPeriod)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_IDToken_WithNilOAuthApp() {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	// oauthApp is nil - should use default config
	result := resolveTokenConfig(nil, TokenTypeID)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(3600), result.ValidityPeriod)
	assert.Equal(suite.T(), "https://thunder.io", result.Issuer)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_IDToken_WithNilToken() {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	// oauthApp.Token is nil - should use default config
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token:    nil,
	}

	result := resolveTokenConfig(oauthApp, TokenTypeID)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(3600), result.ValidityPeriod)
	assert.Equal(suite.T(), "https://thunder.io", result.Issuer)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_IDToken_WithAppLevelConfig() {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				ValidityPeriod: 1800,
			},
		},
	}

	result := resolveTokenConfig(oauthApp, TokenTypeID)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(1800), result.ValidityPeriod)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_WithCustomIssuer_NilOAuthApp() {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	// With nil oauthApp, should use default issuer
	result := resolveTokenConfig(nil, TokenTypeAccess)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "https://thunder.io", result.Issuer)
}

func (suite *UtilsTestSuite) TestResolveTokenConfig_WithTokenConfig_UsesThunderIssuer() {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	// OAuthApp with token config always uses Thunder-level issuer from config
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token:    &appmodel.OAuthTokenConfig{},
	}

	result := resolveTokenConfig(oauthApp, TokenTypeAccess)

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "https://thunder.io", result.Issuer)
}
