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

package userinfo

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type UserInfoServiceTestSuite struct {
	suite.Suite
	mockJWTService  *jwtmock.JWTServiceInterfaceMock
	mockAppService  *applicationmock.ApplicationServiceInterfaceMock
	mockUserService *usermock.UserServiceInterfaceMock
	mockOUService   *oumock.OrganizationUnitServiceInterfaceMock
	userInfoService userInfoServiceInterface
	privateKey      *rsa.PrivateKey
}

func TestUserInfoServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserInfoServiceTestSuite))
}

func (s *UserInfoServiceTestSuite) SetupTest() {
	s.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(s.T())
	s.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(s.T())
	s.mockUserService = usermock.NewUserServiceInterfaceMock(s.T())
	s.mockOUService = oumock.NewOrganizationUnitServiceInterfaceMock(s.T())
	s.userInfoService = newUserInfoService(s.mockJWTService, s.mockAppService, s.mockUserService, s.mockOUService)

	// Initialize Thunder runtime for tests
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime(
		"test-home",
		&config.Config{
			JWT: config.JWTConfig{
				Issuer:         "test-issuer",
				ValidityPeriod: 600,
			},
		},
	)

	// Create a private key for signing JWT tokens
	var err error
	s.privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		s.T().Fatal("Error generating RSA key:", err)
	}
}

// TestGetUserInfo_EmptyToken tests that empty token returns an error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_EmptyToken() {
	response, svcErr := s.userInfoService.GetUserInfo("")
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), errorInvalidAccessToken.Code, svcErr.Code)
	assert.Nil(s.T(), response)
}

// TestGetUserInfo_InvalidTokenSignature tests that invalid token signature returns an error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_InvalidTokenSignature() {
	token := "invalid.token.signature"
	s.mockJWTService.On("VerifyJWT", token, "", "").Return(&serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "INVALID_SIGNATURE",
		Error:            "Invalid signature",
		ErrorDescription: "invalid signature",
	})

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), errorInvalidAccessToken.Code, svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// createToken creates a JWT token with the given claims
func (s *UserInfoServiceTestSuite) createToken(claims map[string]interface{}) string {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}

	headerBytes, _ := json.Marshal(header)
	claimsBytes, _ := json.Marshal(claims)

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)

	signingInput := headerEncoded + "." + claimsEncoded
	hashed := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		s.T().Fatal("Error signing token:", err)
	}
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureEncoded
}

// TestGetUserInfo_InvalidTokenFormat tests that invalid token format returns an error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_InvalidTokenFormat() {
	// nolint:gosec // This is a test token, not a real credential
	invalidToken := "not.a.valid.jwt"
	s.mockJWTService.On("VerifyJWT", invalidToken, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(invalidToken)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), errorInvalidAccessToken.Code, svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// TestGetUserInfo_MissingSubClaim tests that missing sub claim returns an error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_MissingSubClaim() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"scope": "openid profile",
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), errorMissingSubClaim.Code, svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// TestGetUserInfo_EmptySubClaim tests that empty sub claim returns an error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_EmptySubClaim() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "",
		"scope": "openid profile",
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), errorMissingSubClaim.Code, svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// TestGetUserInfo_NoScopes tests that token with no scopes returns insufficient_scope error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_NoScopes() {
	claims := map[string]interface{}{
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"nbf": float64(time.Now().Add(-time.Minute).Unix()),
		"sub": "user123",
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), "insufficient_scope", svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// TestGetUserInfo_NoScopesEmptyScopeString tests that empty scope string returns insufficient_scope error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_NoScopesEmptyScopeString() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "user123",
		"scope": "",
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), "insufficient_scope", svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// TestGetUserInfo_ErrorFetchingUserAttributes tests error when fetching user attributes fails
func (s *UserInfoServiceTestSuite) TestGetUserInfo_ErrorFetchingUserAttributes() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "user123",
		"scope": "openid profile",
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(nil, &serviceerror.ServiceError{
		Code:  "USER_NOT_FOUND",
		Error: "User not found",
	})

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
}

// TestGetUserInfo_ErrorFetchingGroups tests error when fetching groups fails
func (s *UserInfoServiceTestSuite) TestGetUserInfo_ErrorFetchingGroups() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	// This test verifies error handling when groups are needed but fetching fails
	// So we need an OAuth app with groups in UserAttributes
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"name", constants.UserAttributeGroups},
		},
	}
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)
	s.mockUserService.On("GetUserGroups", mock.Anything, "user123",
		constants.DefaultGroupListLimit, 0).Return(nil, &serviceerror.ServiceError{
		Code:  "INTERNAL_ERROR",
		Error: "Failed to fetch groups",
	})

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
}

// TestGetUserInfo_Success_StandardScopes tests successful response with standard OIDC scopes
func (s *UserInfoServiceTestSuite) TestGetUserInfo_Success_StandardScopes() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile email",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				UserAttributes: []string{"name", "email"},
			},
		},
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"name", "email"},
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	assert.Equal(s.T(), "John Doe", response.JSONBody["name"])
	assert.Equal(s.T(), "john@example.com", response.JSONBody["email"])
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_Success_WithGroups tests successful response with groups
func (s *UserInfoServiceTestSuite) TestGetUserInfo_Success_WithGroups() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name": "John Doe",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				UserAttributes: []string{"name", constants.UserAttributeGroups},
			},
		},
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"name", constants.UserAttributeGroups},
		},
		ScopeClaims: map[string][]string{
			"profile": {"name", constants.UserAttributeGroups}, // Add groups to profile scope
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockUserService.On("GetUserGroups", mock.Anything, "user123",
		constants.DefaultGroupListLimit, 0).Return(&user.UserGroupListResponse{
		Groups: []user.UserGroup{
			{Name: "admin"},
			{Name: "users"},
		},
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	assert.Equal(s.T(), "John Doe", response.JSONBody["name"])
	groupsValue := response.JSONBody[constants.UserAttributeGroups]
	assert.NotNil(s.T(), groupsValue, "groups should be present")
	// Groups can be []string or []interface{} depending on JSON unmarshaling
	var groups []string
	switch v := groupsValue.(type) {
	case []string:
		groups = v
	case []interface{}:
		groups = make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				groups = append(groups, str)
			}
		}
	default:
		s.T().Fatalf("groups should be []string or []interface{}, got %T", v)
	}
	assert.Len(s.T(), groups, 2)
	assert.Contains(s.T(), groups, "admin")
	assert.Contains(s.T(), groups, "users")
	assert.Equal(s.T(), []string{"admin", "users"}, groups)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_Success_WithScopeClaimsMapping tests successful response with app-specific scope-to-claims mapping
func (s *UserInfoServiceTestSuite) TestGetUserInfo_Success_WithScopeClaimsMapping() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid custom_scope",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"phone": "1234567890",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				UserAttributes: []string{"name", "email", "phone"},
			},
		},
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"name", "email", "phone"},
		},
		ScopeClaims: map[string][]string{
			"custom_scope": {"name", "phone"},
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	assert.Equal(s.T(), "John Doe", response.JSONBody["name"])
	assert.Equal(s.T(), "1234567890", response.JSONBody["phone"])
	assert.NotContains(s.T(), response.JSONBody, "email") // email not in custom_scope mapping
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_Success_NoAppConfig tests successful response without app config
func (s *UserInfoServiceTestSuite) TestGetUserInfo_Success_NoAppConfig() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "user123",
		"scope": "openid profile",
		// No client_id
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)

	// When no app config, BuildClaims returns empty (no allowedUserAttributes)
	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	// No other claims because allowedUserAttributes is empty
	assert.Len(s.T(), response.JSONBody, 1)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
}

// TestGetUserInfo_Success_AppNotFound tests successful response when app is not found
func (s *UserInfoServiceTestSuite) TestGetUserInfo_Success_AppNotFound() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name": "John Doe",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(nil, &serviceerror.ServiceError{
		Code:  "APP_NOT_FOUND",
		Error: "App not found",
	})

	// When app not found, continue without app config
	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	// No other claims because allowedUserAttributes is empty
	assert.Len(s.T(), response.JSONBody, 1)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_Success_GroupsNotInAllowedAttributes tests that groups are not included if not in allowed attributes
func (s *UserInfoServiceTestSuite) TestGetUserInfo_Success_GroupsNotInAllowedAttributes() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name": "John Doe",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				UserAttributes: []string{"name"}, // groups not in allowed attributes
			},
		},
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"name"},
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	assert.Equal(s.T(), "John Doe", response.JSONBody["name"])
	assert.NotContains(s.T(), response.JSONBody, constants.UserAttributeGroups) // groups not included
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_Success_EmptyUserAttributes tests successful response with empty user attributes
func (s *UserInfoServiceTestSuite) TestGetUserInfo_Success_EmptyUserAttributes() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				UserAttributes: []string{"name", "email"},
			},
		},
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"name", "email"},
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: nil, // No attributes
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	// No other claims because user has no attributes
	assert.Len(s.T(), response.JSONBody, 1)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_Success_InvalidSubClaimType tests that non-string sub claim returns an error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_Success_InvalidSubClaimType() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   123, // Invalid type
		"scope": "openid profile",
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), errorMissingSubClaim.Code, svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// TestGetUserInfo_Success_ScopeAsNonString tests that non-string scope returns insufficient_scope error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_Success_ScopeAsNonString() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "user123",
		"scope": 123, // Invalid type
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), "insufficient_scope", svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// TestGetUserInfo_ScopeExistsButNotString tests when scope exists but is not a string returns insufficient_scope error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_ScopeExistsButNotString() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "user123",
		"scope": []string{"openid"}, // Scope as array instead of string
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), "insufficient_scope", svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// testGetUserInfoInvalidClientID is a helper function for testing invalid client_id scenarios
func (s *UserInfoServiceTestSuite) testGetUserInfoInvalidClientID(clientIDValue interface{}, description string) {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": clientIDValue,
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name": "John Doe",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)

	// When client_id is invalid, app lookup is skipped
	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr, description)
	assert.NotNil(s.T(), response, description)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type, description)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"], description)
	// No other claims because allowedUserAttributes is empty
	assert.Len(s.T(), response.JSONBody, 1, description)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
}

// TestGetUserInfo_ClientIDNotString tests when client_id exists but is not a string
func (s *UserInfoServiceTestSuite) TestGetUserInfo_ClientIDNotString() {
	s.testGetUserInfoInvalidClientID(123, "When client_id is not a string, app lookup is skipped")
}

// TestGetUserInfo_ClientIDEmptyString tests when client_id is empty string
func (s *UserInfoServiceTestSuite) TestGetUserInfo_ClientIDEmptyString() {
	s.testGetUserInfoInvalidClientID("", "When client_id is empty, app lookup is skipped")
}

// TestGetUserInfo_GroupsWithNilOAuthApp tests groups when oauthApp is nil
func (s *UserInfoServiceTestSuite) TestGetUserInfo_GroupsWithNilOAuthApp() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "user123",
		"scope": "openid profile",
		// No client_id
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name": "John Doe",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	// Groups not included because oauthApp is nil
	assert.NotContains(s.T(), response.JSONBody, constants.UserAttributeGroups)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
}

// TestGetUserInfo_GroupsWithNilToken tests groups when Token is nil
func (s *UserInfoServiceTestSuite) TestGetUserInfo_GroupsWithNilToken() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name": "John Doe",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: nil, // Token is nil
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	// When Token is nil, groups are not added
	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	// Groups not included because Token is nil
	assert.NotContains(s.T(), response.JSONBody, constants.UserAttributeGroups)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_GroupsWithNilIDToken tests groups when IDToken is nil
func (s *UserInfoServiceTestSuite) TestGetUserInfo_GroupsWithNilIDToken() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name": "John Doe",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{
			IDToken: nil, // IDToken is nil
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	// When IDToken is nil, groups are not added
	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	// Groups not included because IDToken is nil
	assert.NotContains(s.T(), response.JSONBody, constants.UserAttributeGroups)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_GroupsWithEmptyGroups tests when groups list is empty
func (s *UserInfoServiceTestSuite) TestGetUserInfo_GroupsWithEmptyGroups() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name": "John Doe",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				UserAttributes: []string{"name", constants.UserAttributeGroups},
			},
		},
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"name", constants.UserAttributeGroups},
		},
		ScopeClaims: map[string][]string{
			"profile": {"name", constants.UserAttributeGroups},
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockUserService.On("GetUserGroups", mock.Anything, "user123",
		constants.DefaultGroupListLimit, 0).Return(&user.UserGroupListResponse{
		Groups: []user.UserGroup{}, // Empty groups
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	// When groups is empty, groups are not added to userAttributes
	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	assert.Equal(s.T(), "John Doe", response.JSONBody["name"])
	// Groups not included because len(userGroups) == 0
	assert.NotContains(s.T(), response.JSONBody, constants.UserAttributeGroups)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_ClientCredentialsGrant_Rejected tests that client_credentials grant is rejected
func (s *UserInfoServiceTestSuite) TestGetUserInfo_ClientCredentialsGrant_Rejected() {
	claims := map[string]interface{}{
		"exp":        float64(time.Now().Add(time.Hour).Unix()),
		"nbf":        float64(time.Now().Add(-time.Minute).Unix()),
		"sub":        "client123",
		"scope":      "read write",
		"grant_type": "client_credentials",
		"client_id":  "client123",
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), errorClientCredentialsNotSupported.Code, svcErr.Code)
	assert.Equal(s.T(), errorClientCredentialsNotSupported.ErrorDescription, svcErr.ErrorDescription)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
	// Verify that user service is not called
	s.mockUserService.AssertNotCalled(s.T(), "GetUser", mock.Anything)
}

// testGetUserInfoAllowedGrantType is a helper function for testing allowed grant types
func (s *UserInfoServiceTestSuite) testGetUserInfoAllowedGrantType(grantTypeValue interface{}, description string) {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid profile",
		"client_id": "client123",
	}
	if grantTypeValue != nil {
		claims["grant_type"] = grantTypeValue
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name": "John Doe",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				UserAttributes: []string{"name"},
			},
		},
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"name"},
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr, description)
	assert.NotNil(s.T(), response, description)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type, description)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"], description)
	assert.Equal(s.T(), "John Doe", response.JSONBody["name"], description)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_AuthorizationCodeGrant_Allowed tests that authorization_code grant is allowed
func (s *UserInfoServiceTestSuite) TestGetUserInfo_AuthorizationCodeGrant_Allowed() {
	s.testGetUserInfoAllowedGrantType("authorization_code", "authorization_code grant should be allowed")
}

// TestGetUserInfo_RefreshTokenGrant_Allowed tests that refresh_token grant is allowed
func (s *UserInfoServiceTestSuite) TestGetUserInfo_RefreshTokenGrant_Allowed() {
	s.testGetUserInfoAllowedGrantType("refresh_token", "refresh_token grant should be allowed")
}

// TestGetUserInfo_TokenExchangeGrant_Allowed tests that token_exchange grant is allowed
func (s *UserInfoServiceTestSuite) TestGetUserInfo_TokenExchangeGrant_Allowed() {
	s.testGetUserInfoAllowedGrantType(
		"urn:ietf:params:oauth:grant-type:token-exchange",
		"token_exchange grant should be allowed")
}

// TestGetUserInfo_NoGrantType_Allowed tests that tokens without grant_type claim are allowed (backward compatibility)
func (s *UserInfoServiceTestSuite) TestGetUserInfo_NoGrantType_Allowed() {
	s.testGetUserInfoAllowedGrantType(nil, "tokens without grant_type should be allowed")
}

// TestGetUserInfo_GrantTypeNotString_Allowed tests that non-string grant_type is ignored and allowed
func (s *UserInfoServiceTestSuite) TestGetUserInfo_GrantTypeNotString_Allowed() {
	s.testGetUserInfoAllowedGrantType(123, "non-string grant_type should be ignored and allowed")
}

// TestGetUserInfo_MissingOpenIDScope_WithOtherScopes tests that missing openid scope returns insufficient_scope error
func (s *UserInfoServiceTestSuite) TestGetUserInfo_MissingOpenIDScope_WithOtherScopes() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "user123",
		"scope": "profile email", // Missing 'openid' scope
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), "insufficient_scope", svcErr.Code)
	assert.Contains(s.T(), svcErr.ErrorDescription, "openid")
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
	// Verify that user service is NOT called (fail fast)
	s.mockUserService.AssertNotCalled(s.T(), "GetUser")
}

// TestGetUserInfo_OpenIDScope_CaseSensitive tests that scope matching is case-sensitive
func (s *UserInfoServiceTestSuite) TestGetUserInfo_OpenIDScope_CaseSensitive() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "user123",
		"scope": "OpenID profile", // Wrong case - should fail
	}
	token := s.createToken(claims)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), "insufficient_scope", svcErr.Code)
	assert.Nil(s.T(), response)
	s.mockJWTService.AssertExpectations(s.T())
}

// TestGetUserInfo_OnlyOpenIDScope_Success tests that only openid scope returns sub claim
func (s *UserInfoServiceTestSuite) TestGetUserInfo_OnlyOpenIDScope_Success() {
	claims := map[string]interface{}{
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"nbf":   float64(time.Now().Add(-time.Minute).Unix()),
		"sub":   "user123",
		"scope": "openid", // Only openid scope
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	// Only sub claim should be present
	assert.Len(s.T(), response.JSONBody, 1)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
}

// TestGetUserInfo_OpenIDScope_InMiddleOfScopeString tests openid scope in middle position
func (s *UserInfoServiceTestSuite) TestGetUserInfo_OpenIDScope_InMiddleOfScopeString() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "profile openid email", // openid in middle
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"name", "email"},
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	assert.Equal(s.T(), "John Doe", response.JSONBody["name"])
	assert.Equal(s.T(), "john@example.com", response.JSONBody["email"])
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_OpenIDScope_AtEnd tests openid scope at end of scope string
func (s *UserInfoServiceTestSuite) TestGetUserInfo_OpenIDScope_AtEnd() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "profile email openid", // openid at end
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"email": "john@example.com",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		UserInfo: &appmodel.UserInfoConfig{
			UserAttributes: []string{"email"},
		},
	}

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJSON, response.Type)
	assert.NotNil(s.T(), response.JSONBody)
	assert.Equal(s.T(), "user123", response.JSONBody["sub"])
	assert.Equal(s.T(), "john@example.com", response.JSONBody["email"])
	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_JWS_ResponseType tests that when the OAuth application
// is configured with UserInfo response_type as JWS, the service generates
// and returns a signed JWT response instead of JSON.
func (s *UserInfoServiceTestSuite) TestGetUserInfo_JWS_ResponseType() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid email",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"email": "john@example.com",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{},
		UserInfo: &appmodel.UserInfoConfig{
			ResponseType:   appmodel.UserInfoResponseTypeJWS,
			UserAttributes: []string{"email"},
		},
	}

	issuer := "test-issuer"

	// JWT verification
	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	// User fetch
	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)

	// App fetch
	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	// JWT generation
	s.mockJWTService.On(
		"GenerateJWT",
		"user123",
		"client123",
		issuer,
		config.GetThunderRuntime().Config.JWT.ValidityPeriod,
		mock.Anything,
	).Return("signed.jwt.token", int64(0), nil)

	response, svcErr := s.userInfoService.GetUserInfo(token)

	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), response)
	assert.Equal(s.T(), appmodel.UserInfoResponseTypeJWS, response.Type)
	assert.Equal(s.T(), "signed.jwt.token", response.JWTBody)
	assert.Nil(s.T(), response.JSONBody)

	s.mockJWTService.AssertExpectations(s.T())
	s.mockUserService.AssertExpectations(s.T())
	s.mockAppService.AssertExpectations(s.T())
}

// TestGetUserInfo_JWS_GenerateJWTFailure tests that
// an internal server error is returned when JWT generation fails.
func (s *UserInfoServiceTestSuite) TestGetUserInfo_JWS_GenerateJWTFailure() {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"sub":       "user123",
		"scope":     "openid email",
		"client_id": "client123",
	}
	token := s.createToken(claims)

	userAttrs := map[string]interface{}{
		"email": "john@example.com",
	}
	userAttrsJSON, _ := json.Marshal(userAttrs)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		Token: &appmodel.OAuthTokenConfig{},
		UserInfo: &appmodel.UserInfoConfig{
			ResponseType:   appmodel.UserInfoResponseTypeJWS,
			UserAttributes: []string{"email"},
		},
	}
	issuer := "test-issuer"

	s.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	s.mockUserService.On("GetUser", mock.Anything, "user123").Return(&user.User{
		ID:         "user123",
		Attributes: userAttrsJSON,
	}, nil)

	s.mockAppService.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	// Simulate signing failure
	s.mockJWTService.On(
		"GenerateJWT",
		"user123",
		"client123",
		issuer,
		config.GetThunderRuntime().Config.JWT.ValidityPeriod,
		mock.Anything,
	).Return("", int64(0),
		&serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "JWT_SIGNING_FAILED",
			Error:            "JWT signing failed",
			ErrorDescription: "JWT signing failed",
		})

	response, svcErr := s.userInfoService.GetUserInfo(token)

	assert.Nil(s.T(), response)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)

	s.mockJWTService.AssertExpectations(s.T())
}
