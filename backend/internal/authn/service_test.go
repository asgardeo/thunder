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

package authn

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/authn/assert"
	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/idp"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/authn/assertmock"
	"github.com/asgardeo/thunder/tests/mocks/authn/credentialsmock"
	"github.com/asgardeo/thunder/tests/mocks/authn/githubmock"
	"github.com/asgardeo/thunder/tests/mocks/authn/googlemock"
	"github.com/asgardeo/thunder/tests/mocks/authn/oauthmock"
	"github.com/asgardeo/thunder/tests/mocks/authn/oidcmock"
	"github.com/asgardeo/thunder/tests/mocks/authn/otpmock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

const (
	testUserID       = "user123"
	testIDPID        = "idp_123"
	testOrgUnit      = "org_unit_123"
	testAuthCode     = "auth_code_123"
	testToken        = "token_123"
	testSessionTkn   = "session_token_123"
	testJWTToken     = "jwt_token_123" // #nosec G101
	testRedirectURL  = "https://oauth.provider.com/authorize"
	invalidAssertion = "invalid.jwt.token"
)

type AuthenticationServiceTestSuite struct {
	suite.Suite
	mockIDPService         *idpmock.IDPServiceInterfaceMock
	mockJWTService         *jwtmock.JWTServiceInterfaceMock
	mockCredentialsService *credentialsmock.CredentialsAuthnServiceInterfaceMock
	mockOTPService         *otpmock.OTPAuthnServiceInterfaceMock
	mockOAuthService       *oauthmock.OAuthAuthnServiceInterfaceMock
	mockOIDCService        *oidcmock.OIDCAuthnServiceInterfaceMock
	mockGoogleService      *googlemock.GoogleOIDCAuthnServiceInterfaceMock
	mockGithubService      *githubmock.GithubOAuthAuthnServiceInterfaceMock
	service                *authenticationService
}

func TestAuthenticationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationServiceTestSuite))
}

func (suite *AuthenticationServiceTestSuite) SetupSuite() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "test-issuer",
			ValidityPeriod: 3600,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	if err != nil {
		suite.T().Fatalf("Failed to initialize ThunderRuntime: %v", err)
	}

	// Register authenticators for IDP types
	common.RegisterAuthenticator(common.AuthenticatorMeta{
		Name:          "OAuthAuthenticator",
		AssociatedIDP: idp.IDPTypeOAuth,
	})
	common.RegisterAuthenticator(common.AuthenticatorMeta{
		Name:          "OIDCAuthenticator",
		AssociatedIDP: idp.IDPTypeOIDC,
	})
	common.RegisterAuthenticator(common.AuthenticatorMeta{
		Name:          "GoogleAuthenticator",
		AssociatedIDP: idp.IDPTypeGoogle,
	})
	common.RegisterAuthenticator(common.AuthenticatorMeta{
		Name:          "GitHubAuthenticator",
		AssociatedIDP: idp.IDPTypeGitHub,
	})
}

func (suite *AuthenticationServiceTestSuite) SetupTest() {
	suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockCredentialsService = credentialsmock.NewCredentialsAuthnServiceInterfaceMock(suite.T())
	suite.mockOTPService = otpmock.NewOTPAuthnServiceInterfaceMock(suite.T())
	suite.mockOAuthService = oauthmock.NewOAuthAuthnServiceInterfaceMock(suite.T())
	suite.mockOIDCService = oidcmock.NewOIDCAuthnServiceInterfaceMock(suite.T())
	suite.mockGoogleService = googlemock.NewGoogleOIDCAuthnServiceInterfaceMock(suite.T())
	suite.mockGithubService = githubmock.NewGithubOAuthAuthnServiceInterfaceMock(suite.T())

	suite.service = &authenticationService{
		idpService:             suite.mockIDPService,
		jwtService:             suite.mockJWTService,
		authAssertionGenerator: assert.NewAuthAssertGenerator(),
		credentialsService:     suite.mockCredentialsService,
		otpService:             suite.mockOTPService,
		oauthService:           suite.mockOAuthService,
		oidcService:            suite.mockOIDCService,
		googleService:          suite.mockGoogleService,
		githubService:          suite.mockGithubService,
	}
}

func (suite *AuthenticationServiceTestSuite) TestAuthenticateWithCredentials() {
	attributes := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}

	testCases := []struct {
		name              string
		skipAssertion     bool
		existingAssertion string
		expectAssertion   bool
		validateClaims    bool
		setupMocks        func()
		validateAssertion func(result *common.AuthenticationResponse)
	}{
		{
			name:            "Success without assertion",
			skipAssertion:   true,
			expectAssertion: false,
			setupMocks: func() {
				suite.mockCredentialsService.On("Authenticate", attributes).Return(testUser, nil).Once()
			},
			validateAssertion: func(result *common.AuthenticationResponse) {
				suite.Empty(result.Assertion)
			},
		},
		{
			name:            "Success with assertion generation",
			skipAssertion:   false,
			expectAssertion: true,
			validateClaims:  true,
			setupMocks: func() {
				suite.mockCredentialsService.On("Authenticate", attributes).Return(testUser, nil).Once()
				suite.mockJWTService.On("GenerateJWT", testUserID, "application", mock.Anything, mock.Anything,
					mock.MatchedBy(func(claims map[string]interface{}) bool {
						// Verify that assurance claims are present
						_, hasAssurance := claims["assurance"]
						return hasAssurance
					})).Return(testJWTToken, int64(3600), nil).Once()
			},
			validateAssertion: func(result *common.AuthenticationResponse) {
				suite.Equal(testJWTToken, result.Assertion)
			},
		},
		{
			name:              "Success with existing assertion",
			skipAssertion:     false,
			existingAssertion: "", // Will be set in setupMocks
			expectAssertion:   true,
			validateClaims:    true,
			setupMocks: func() {
				suite.mockCredentialsService.On("Authenticate", attributes).Return(testUser, nil).Once()
				suite.mockJWTService.On("VerifyJWT", mock.Anything, "", mock.Anything).Return(nil).Once()
				suite.mockJWTService.On("GenerateJWT", testUserID, "application", mock.Anything, mock.Anything,
					mock.Anything).Return(testJWTToken, int64(3600), nil).Once()
			},
			validateAssertion: func(result *common.AuthenticationResponse) {
				suite.Equal(testJWTToken, result.Assertion)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setupMocks()

			// Create existing assertion if needed
			existingAssertion := tc.existingAssertion
			if tc.name == "Success with existing assertion" {
				existingAssertion = suite.createTestAssertion(testUserID)
			}

			result, err := suite.service.AuthenticateWithCredentials(attributes, tc.skipAssertion, existingAssertion)

			suite.Nil(err)
			suite.NotNil(result)
			suite.Equal(testUserID, result.ID)
			suite.Equal(testOrgUnit, result.OrganizationUnit)
			tc.validateAssertion(result)
		})
	}
}

func (suite *AuthenticationServiceTestSuite) TestAuthenticateWithCredentialsServiceError() {
	attributes := map[string]interface{}{
		"username": "testuser",
		"password": "wrongpass",
	}
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "INVALID_CREDENTIALS",
		Error:            "Invalid credentials",
		ErrorDescription: "The provided credentials are invalid",
	}

	suite.mockCredentialsService.On("Authenticate", attributes).Return(nil, svcErr)

	result, err := suite.service.AuthenticateWithCredentials(attributes, false, "")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(svcErr.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestAuthenticateWithCredentialsJWTGenerationError() {
	attributes := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}

	suite.mockCredentialsService.On("Authenticate", attributes).Return(testUser, nil)
	suite.mockJWTService.On("GenerateJWT", testUserID, "application", mock.Anything, mock.Anything, mock.Anything).
		Return("", int64(0), errors.New("JWT generation failed"))

	result, err := suite.service.AuthenticateWithCredentials(attributes, false, "")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInternalServerError.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestAuthenticateWithCredentialsSubjectMismatch() {
	attributes := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}

	// Create assertion with different subject
	existingAssertion := suite.createTestAssertion("different_user_id")

	suite.mockCredentialsService.On("Authenticate", attributes).Return(testUser, nil)
	suite.mockJWTService.On("VerifyJWT", existingAssertion, "", mock.Anything).Return(nil)

	result, err := suite.service.AuthenticateWithCredentials(attributes, false, existingAssertion)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorAssertionSubjectMismatch.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestAuthenticateWithCredentialsInvalidExistingAssertion() {
	attributes := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}

	suite.mockCredentialsService.On("Authenticate", attributes).Return(testUser, nil)
	suite.mockJWTService.On("VerifyJWT", invalidAssertion, "", mock.Anything).Return(errors.New("invalid JWT"))

	result, err := suite.service.AuthenticateWithCredentials(attributes, false, invalidAssertion)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidAssertion.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestAuthenticateWithCredentialsExistingAssertionWithoutAssurance() {
	attributes := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}

	// Create assertion without assurance claim
	existingAssertion := suite.createTestAssertionWithoutAssurance(testUserID)

	suite.mockCredentialsService.On("Authenticate", attributes).Return(testUser, nil)
	suite.mockJWTService.On("VerifyJWT", existingAssertion, "", mock.Anything).Return(nil)

	result, err := suite.service.AuthenticateWithCredentials(attributes, false, existingAssertion)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidAssertion.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestSendOTPSuccess() {
	senderID := "sender_123"
	recipient := "+1234567890"
	sessionToken := testSessionTkn

	suite.mockOTPService.On("SendOTP", senderID, notifcommon.ChannelTypeSMS, recipient).
		Return(sessionToken, nil)

	result, err := suite.service.SendOTP(senderID, notifcommon.ChannelTypeSMS, recipient)

	suite.Nil(err)
	suite.Equal(sessionToken, result)
}

func (suite *AuthenticationServiceTestSuite) TestSendOTPServiceError() {
	senderID := "sender_123"
	recipient := "+1234567890"
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OTP_ERROR",
		Error:            "OTP error",
		ErrorDescription: "Failed to send OTP",
	}

	suite.mockOTPService.On("SendOTP", senderID, notifcommon.ChannelTypeSMS, recipient).
		Return("", svcErr)

	result, err := suite.service.SendOTP(senderID, notifcommon.ChannelTypeSMS, recipient)

	suite.Empty(result)
	suite.NotNil(err)
	suite.Equal(svcErr.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestVerifyOTP() {
	sessionToken := testSessionTkn
	otpCode := "123456"
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}

	testCases := []struct {
		name              string
		skipAssertion     bool
		existingAssertion string
		expectAssertion   bool
		setupMocks        func()
		validateAssertion func(result *common.AuthenticationResponse)
	}{
		{
			name:              "Success without assertion",
			skipAssertion:     true,
			existingAssertion: "",
			expectAssertion:   false,
			setupMocks: func() {
				suite.mockOTPService.On("VerifyOTP", sessionToken, otpCode).Return(testUser, nil).Once()
			},
			validateAssertion: func(result *common.AuthenticationResponse) {
				suite.Empty(result.Assertion)
			},
		},
		{
			name:              "Success with assertion generation",
			skipAssertion:     false,
			existingAssertion: "",
			expectAssertion:   true,
			setupMocks: func() {
				suite.mockOTPService.On("VerifyOTP", sessionToken, otpCode).Return(testUser, nil).Once()
				suite.mockJWTService.On("GenerateJWT", testUserID, "application", mock.Anything, mock.Anything,
					mock.MatchedBy(func(claims map[string]interface{}) bool {
						// Verify that assurance claims are present
						_, hasAssurance := claims["assurance"]
						return hasAssurance
					})).Return(testJWTToken, int64(3600), nil).Once()
			},
			validateAssertion: func(result *common.AuthenticationResponse) {
				suite.Equal(testJWTToken, result.Assertion)
			},
		},
		{
			name:              "Success with existing assertion (MFA)",
			skipAssertion:     false,
			existingAssertion: suite.createTestAssertion(testUserID),
			expectAssertion:   true,
			setupMocks: func() {
				existingAssertion := suite.createTestAssertion(testUserID)
				suite.mockOTPService.On("VerifyOTP", sessionToken, otpCode).Return(testUser, nil).Once()
				suite.mockJWTService.On("VerifyJWT", existingAssertion, "", mock.Anything).Return(nil).Once()
				suite.mockJWTService.On("GenerateJWT", testUserID, "application", mock.Anything, mock.Anything,
					mock.MatchedBy(func(claims map[string]interface{}) bool {
						// Verify that assurance claims are present for MFA
						_, hasAssurance := claims["assurance"]
						return hasAssurance
					})).Return("new_jwt_token_with_mfa", int64(3600), nil).Once()
			},
			validateAssertion: func(result *common.AuthenticationResponse) {
				suite.NotEmpty(result.Assertion)
				suite.Equal("new_jwt_token_with_mfa", result.Assertion)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setupMocks()

			result, err := suite.service.VerifyOTP(sessionToken, tc.skipAssertion, tc.existingAssertion, otpCode)

			suite.Nil(err)
			suite.NotNil(result)
			suite.Equal(testUserID, result.ID)
			tc.validateAssertion(result)
		})
	}
}

func (suite *AuthenticationServiceTestSuite) TestVerifyOTPServiceError() {
	sessionToken := testSessionTkn
	otpCode := "wrong_otp"
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "INCORRECT_OTP",
		Error:            "Incorrect OTP",
		ErrorDescription: "The provided OTP is incorrect",
	}

	suite.mockOTPService.On("VerifyOTP", sessionToken, otpCode).Return(nil, svcErr)

	result, err := suite.service.VerifyOTP(sessionToken, false, "", otpCode)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(svcErr.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationOAuthSuccess() {
	idpID := testIDPID
	redirectURL := testRedirectURL
	identityProvider := &idp.IDPDTO{
		ID:   idpID,
		Type: idp.IDPTypeOAuth,
	}

	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(identityProvider, nil)
	suite.mockOAuthService.On("BuildAuthorizeURL", idpID).Return(redirectURL, nil)
	suite.mockJWTService.On("GenerateJWT", "auth-svc", "auth-svc", mock.Anything, mock.Anything, mock.Anything).
		Return(testSessionTkn, int64(600), nil)

	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeOAuth, idpID)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(redirectURL, result.RedirectURL)
	suite.Equal(testSessionTkn, result.SessionToken)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationOIDCSuccess() {
	idpID := testIDPID
	redirectURL := "https://oidc.provider.com/authorize"
	identityProvider := &idp.IDPDTO{
		ID:   idpID,
		Type: idp.IDPTypeOIDC,
	}

	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(identityProvider, nil)
	suite.mockOIDCService.On("BuildAuthorizeURL", idpID).Return(redirectURL, nil)
	suite.mockJWTService.On("GenerateJWT", "auth-svc", "auth-svc", mock.Anything, mock.Anything, mock.Anything).
		Return(testSessionTkn, int64(600), nil)

	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeOIDC, idpID)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(redirectURL, result.RedirectURL)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationGoogleSuccess() {
	idpID := testIDPID
	redirectURL := "https://accounts.google.com/o/oauth2/v2/auth"
	identityProvider := &idp.IDPDTO{
		ID:   idpID,
		Type: idp.IDPTypeGoogle,
	}

	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(identityProvider, nil)
	suite.mockGoogleService.On("BuildAuthorizeURL", idpID).Return(redirectURL, nil)
	suite.mockJWTService.On("GenerateJWT", "auth-svc", "auth-svc", mock.Anything, mock.Anything, mock.Anything).
		Return(testSessionTkn, int64(600), nil)

	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeGoogle, idpID)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(redirectURL, result.RedirectURL)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationGitHubSuccess() {
	idpID := testIDPID
	redirectURL := "https://github.com/login/oauth/authorize"
	identityProvider := &idp.IDPDTO{
		ID:   idpID,
		Type: idp.IDPTypeGitHub,
	}

	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(identityProvider, nil)
	suite.mockGithubService.On("BuildAuthorizeURL", idpID).Return(redirectURL, nil)
	suite.mockJWTService.On("GenerateJWT", "auth-svc", "auth-svc", mock.Anything, mock.Anything, mock.Anything).
		Return(testSessionTkn, int64(600), nil)

	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeGitHub, idpID)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(redirectURL, result.RedirectURL)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationEmptyIDPID() {
	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeOAuth, "")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidIDPID.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationIDPNotFound() {
	idpID := "nonexistent_idp"
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP_NOT_FOUND",
		Error:            "IDP not found",
		ErrorDescription: "The identity provider was not found",
	}

	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(nil, svcErr)

	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeOAuth, idpID)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Contains(err.ErrorDescription, idpID)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationInvalidIDPType() {
	idpID := testIDPID
	identityProvider := &idp.IDPDTO{
		ID:   idpID,
		Type: idp.IDPTypeGoogle,
	}

	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(identityProvider, nil)

	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeGitHub, idpID)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidIDPType.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationCrossTypeAllowed() {
	idpID := testIDPID
	redirectURL := testRedirectURL
	identityProvider := &idp.IDPDTO{
		ID:   idpID,
		Type: idp.IDPTypeOAuth,
	}

	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(identityProvider, nil)
	suite.mockOAuthService.On("BuildAuthorizeURL", idpID).Return(redirectURL, nil)
	suite.mockJWTService.On("GenerateJWT", "auth-svc", "auth-svc", mock.Anything, mock.Anything, mock.Anything).
		Return(testSessionTkn, int64(600), nil)

	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeOIDC, idpID)

	suite.Nil(err)
	suite.NotNil(result)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationJWTGenerationError() {
	idpID := testIDPID
	redirectURL := testRedirectURL
	identityProvider := &idp.IDPDTO{
		ID:   idpID,
		Type: idp.IDPTypeOAuth,
	}

	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(identityProvider, nil)
	suite.mockOAuthService.On("BuildAuthorizeURL", idpID).Return(redirectURL, nil)
	suite.mockJWTService.On("GenerateJWT", "auth-svc", "auth-svc", mock.Anything, mock.Anything, mock.Anything).
		Return("", int64(0), errors.New("JWT generation failed"))

	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeOAuth, idpID)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInternalServerError.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationOAuthSuccess() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	tokenResp := &oauth.TokenResponse{
		AccessToken: testToken,
		TokenType:   "Bearer",
	}
	userInfo := map[string]interface{}{
		"sub": testUserID,
	}

	sessionToken := suite.createSessionToken(idp.IDPTypeOAuth)
	suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil)
	suite.mockOAuthService.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", testIDPID, testToken).Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", testUserID).Return(testUser, nil)

	result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeOAuth, sessionToken, true, "", testAuthCode)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(testUserID, result.ID)
	suite.Empty(result.Assertion)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationOIDCSuccess() {
	suite.testFinishOIDCBasedAuth(idp.IDPTypeOIDC, suite.mockOIDCService)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationGoogleSuccess() {
	suite.testFinishOIDCBasedAuth(idp.IDPTypeGoogle, suite.mockGoogleService)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationGitHubSuccess() {
	userInfo := map[string]interface{}{
		"sub": testUserID,
	}
	suite.testFinishOAuthBasedAuth(idp.IDPTypeGitHub, suite.mockGithubService, userInfo)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationWithAssertion() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	tokenResp := &oauth.TokenResponse{
		AccessToken: testToken,
		TokenType:   "Bearer",
	}
	userInfo := map[string]interface{}{
		"sub": testUserID,
	}

	testCases := []struct {
		name              string
		skipAssertion     bool
		existingAssertion string
		setupMocks        func()
		validateAssertion func(result *common.AuthenticationResponse)
	}{
		{
			name:              "Success with assertion generation",
			skipAssertion:     false,
			existingAssertion: "",
			setupMocks: func() {
				sessionToken := suite.createSessionToken(idp.IDPTypeOAuth)
				suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil).Once()
				suite.mockOAuthService.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil).Once()
				suite.mockOAuthService.On("FetchUserInfo", testIDPID, testToken).Return(userInfo, nil).Once()
				suite.mockOAuthService.On("GetInternalUser", testUserID).Return(testUser, nil).Once()
				suite.mockJWTService.On("GenerateJWT", testUserID, "application", mock.Anything, mock.Anything, mock.Anything).
					Return(testJWTToken, int64(3600), nil).Once()
			},
			validateAssertion: func(result *common.AuthenticationResponse) {
				suite.Equal(testJWTToken, result.Assertion)
			},
		},
		{
			name:              "Success with existing assertion (MFA)",
			skipAssertion:     false,
			existingAssertion: suite.createTestAssertion(testUserID),
			setupMocks: func() {
				sessionToken := suite.createSessionToken(idp.IDPTypeOAuth)
				existingAssertion := suite.createTestAssertion(testUserID)
				suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil).Once()
				suite.mockJWTService.On("VerifyJWT", existingAssertion, "", mock.Anything).Return(nil).Once()
				suite.mockOAuthService.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil).Once()
				suite.mockOAuthService.On("FetchUserInfo", testIDPID, testToken).Return(userInfo, nil).Once()
				suite.mockOAuthService.On("GetInternalUser", testUserID).Return(testUser, nil).Once()
				suite.mockJWTService.On("GenerateJWT", testUserID, "application", mock.Anything, mock.Anything,
					mock.MatchedBy(func(claims map[string]interface{}) bool {
						// Verify that assurance claims are present for MFA
						_, hasAssurance := claims["assurance"]
						return hasAssurance
					})).Return("new_jwt_token_with_mfa", int64(3600), nil).Once()
			},
			validateAssertion: func(result *common.AuthenticationResponse) {
				suite.NotEmpty(result.Assertion)
				suite.Equal("new_jwt_token_with_mfa", result.Assertion)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setupMocks()

			sessionToken := suite.createSessionToken(idp.IDPTypeOAuth)
			result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeOAuth, sessionToken,
				tc.skipAssertion, tc.existingAssertion, testAuthCode)

			suite.Nil(err)
			suite.NotNil(result)
			suite.Equal(testUserID, result.ID)
			tc.validateAssertion(result)
		})
	}
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationEmptySessionToken() {
	result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeOAuth, "", false, "", testAuthCode)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorEmptySessionToken.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationEmptyAuthCode() {
	sessionToken := suite.createSessionToken(idp.IDPTypeOAuth)

	result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeOAuth, sessionToken, false, "", "")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorEmptyAuthCode.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationInvalidSessionToken() {
	suite.mockJWTService.On("VerifyJWT", "invalid_token", "auth-svc", mock.Anything).
		Return(errors.New("invalid token"))

	result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeOAuth, "invalid_token", false, "", testAuthCode)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidSessionToken.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationTypeMismatch() {
	sessionToken := suite.createSessionToken(idp.IDPTypeGoogle)
	suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil)

	result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeGitHub, sessionToken, false, "", testAuthCode)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidIDPType.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationSubClaimNotFound() {
	tokenResp := &oauth.TokenResponse{
		AccessToken: testToken,
		TokenType:   "Bearer",
	}
	userInfo := map[string]interface{}{
		"name": "Test User",
	}

	sessionToken := suite.createSessionToken(idp.IDPTypeOAuth)
	suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil)
	suite.mockOAuthService.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", testIDPID, testToken).Return(userInfo, nil)

	result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeOAuth, sessionToken, false, "", testAuthCode)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorSubClaimNotFound.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationSubClaimFallbackToID() {
	userInfo := map[string]interface{}{
		"id": testUserID,
	}
	suite.testFinishOAuthBasedAuth(idp.IDPTypeOAuth, suite.mockOAuthService, userInfo)
}

func (suite *AuthenticationServiceTestSuite) TestValidateIDPTypeExactMatch() {
	err := suite.service.validateIDPType(idp.IDPTypeOAuth, idp.IDPTypeOAuth, nil)
	suite.Nil(err)
}

func (suite *AuthenticationServiceTestSuite) TestValidateIDPTypeEmptyRequested() {
	err := suite.service.validateIDPType("", idp.IDPTypeOAuth, nil)
	suite.Nil(err)
}

func (suite *AuthenticationServiceTestSuite) TestValidateIDPTypeCrossAllowed() {
	err := suite.service.validateIDPType(idp.IDPTypeOAuth, idp.IDPTypeOIDC, nil)
	suite.Nil(err)

	err = suite.service.validateIDPType(idp.IDPTypeOIDC, idp.IDPTypeOAuth, nil)
	suite.Nil(err)
}

func (suite *AuthenticationServiceTestSuite) TestValidateIDPTypeMismatch() {
	logger := log.GetLogger()
	err := suite.service.validateIDPType(idp.IDPTypeGoogle, idp.IDPTypeGitHub, logger)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidIDPType.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestGetSubClaimFromSub() {
	userClaims := map[string]interface{}{
		"sub": testUserID,
	}

	result, err := suite.service.getSubClaim(userClaims, nil)

	suite.Nil(err)
	suite.Equal(testUserID, result)
}

func (suite *AuthenticationServiceTestSuite) TestGetSubClaimFromID() {
	userClaims := map[string]interface{}{
		"id": testUserID,
	}

	result, err := suite.service.getSubClaim(userClaims, nil)

	suite.Nil(err)
	suite.Equal(testUserID, result)
}

func (suite *AuthenticationServiceTestSuite) TestGetSubClaimFromIDNumeric() {
	userClaims := map[string]interface{}{
		"id": 12345,
	}

	result, err := suite.service.getSubClaim(userClaims, nil)

	suite.Nil(err)
	suite.Equal("12345", result)
}

func (suite *AuthenticationServiceTestSuite) TestGetSubClaimNotFound() {
	logger := log.GetLogger()
	userClaims := map[string]interface{}{
		"name": "Test User",
	}

	result, err := suite.service.getSubClaim(userClaims, logger)

	suite.Empty(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorSubClaimNotFound.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestHandleIDPServiceErrorServerError() {
	idpID := "test_idp"
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "INTERNAL_ERROR",
		Error:            "Internal error",
		ErrorDescription: "Database connection failed",
	}
	logger := log.GetLogger()

	result := suite.service.handleIDPServiceError(idpID, svcErr, logger)

	suite.NotNil(result)
	suite.Equal(common.ErrorInternalServerError.Code, result.Code)
}

func (suite *AuthenticationServiceTestSuite) TestVerifyAndDecodeSessionTokenMalformedPayload() {
	logger := log.GetLogger()
	badToken := "header.invalid-base64.signature"

	suite.mockJWTService.On("VerifyJWT", badToken, "auth-svc", mock.Anything).Return(nil)

	result, err := suite.service.verifyAndDecodeSessionToken(badToken, logger)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidSessionToken.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestVerifyAndDecodeSessionTokenMissingAuthData() {
	logger := log.GetLogger()
	payload := map[string]interface{}{
		"sub": "test",
	}
	payloadBytes, _ := json.Marshal(payload)
	encoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	tokenWithoutAuthData := "header." + encoded + ".signature"

	suite.mockJWTService.On("VerifyJWT", tokenWithoutAuthData, "auth-svc", mock.Anything).Return(nil)

	result, err := suite.service.verifyAndDecodeSessionToken(tokenWithoutAuthData, logger)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidSessionToken.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestStartIDPAuthenticationBuildURLError() {
	idpID := testIDPID
	identityProvider := &idp.IDPDTO{
		ID:   idpID,
		Type: idp.IDPTypeOAuth,
	}
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "INVALID_CONFIG",
		Error:            "Invalid configuration",
		ErrorDescription: "Missing redirect URI",
	}

	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(identityProvider, nil)
	suite.mockOAuthService.On("BuildAuthorizeURL", idpID).Return("", svcErr)

	result, err := suite.service.StartIDPAuthentication(idp.IDPTypeOAuth, idpID)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(svcErr.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestFinishOIDCAuthenticationFetchUserInfoError() {
	tokenResp := &oauth.TokenResponse{
		AccessToken: testToken,
		IDToken:     "id_token_123",
		TokenType:   "Bearer",
	}
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FETCH_ERROR",
		Error:            "Fetch error",
		ErrorDescription: "Failed to fetch ID token claims",
	}

	sessionToken := suite.createSessionToken(idp.IDPTypeOIDC)
	suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil)
	suite.mockOIDCService.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_123").Return(nil, svcErr)

	result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeOIDC, sessionToken, true, "", testAuthCode)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(svcErr.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestFinishGoogleAuthenticationGetInternalUserError() {
	tokenResp := &oauth.TokenResponse{
		AccessToken: testToken,
		IDToken:     "id_token_123",
		TokenType:   "Bearer",
	}
	claims := map[string]interface{}{
		"sub": testUserID,
	}
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USER_NOT_FOUND",
		Error:            "User not found",
		ErrorDescription: "Internal user not found",
	}

	sessionToken := suite.createSessionToken(idp.IDPTypeGoogle)
	suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil)
	suite.mockGoogleService.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil)
	suite.mockGoogleService.On("GetIDTokenClaims", "id_token_123").Return(claims, nil)
	suite.mockGoogleService.On("GetInternalUser", testUserID).Return(nil, svcErr)

	result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeGoogle, sessionToken, true, "", testAuthCode)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(svcErr.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) testFinishOIDCBasedAuth(
	idpType idp.IDPType,
	mockService interface{},
) {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	tokenResp := &oauth.TokenResponse{
		AccessToken: testToken,
		IDToken:     "id_token_123",
		TokenType:   "Bearer",
	}
	claims := map[string]interface{}{
		"sub": testUserID,
	}

	sessionToken := suite.createSessionToken(idpType)
	suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil)

	switch service := mockService.(type) {
	case *oidcmock.OIDCAuthnServiceInterfaceMock:
		service.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil)
		service.On("GetIDTokenClaims", "id_token_123").Return(claims, nil)
		service.On("GetInternalUser", testUserID).Return(testUser, nil)
	case *googlemock.GoogleOIDCAuthnServiceInterfaceMock:
		service.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil)
		service.On("GetIDTokenClaims", "id_token_123").Return(claims, nil)
		service.On("GetInternalUser", testUserID).Return(testUser, nil)
	}

	result, err := suite.service.FinishIDPAuthentication(idpType, sessionToken, true, "", testAuthCode)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(testUserID, result.ID)
}

func (suite *AuthenticationServiceTestSuite) testFinishOAuthBasedAuth(
	idpType idp.IDPType,
	mockService interface{},
	userInfo map[string]interface{},
) {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	tokenResp := &oauth.TokenResponse{
		AccessToken: testToken,
		TokenType:   "Bearer",
	}

	sessionToken := suite.createSessionToken(idpType)
	suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil)

	switch service := mockService.(type) {
	case *githubmock.GithubOAuthAuthnServiceInterfaceMock:
		service.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil)
		service.On("FetchUserInfo", testIDPID, testToken).Return(userInfo, nil)
		service.On("GetInternalUser", testUserID).Return(testUser, nil)
	case *oauthmock.OAuthAuthnServiceInterfaceMock:
		service.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).Return(tokenResp, nil)
		service.On("FetchUserInfo", testIDPID, testToken).Return(userInfo, nil)
		service.On("GetInternalUser", testUserID).Return(testUser, nil)
	}

	result, err := suite.service.FinishIDPAuthentication(idpType, sessionToken, true, "", testAuthCode)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(testUserID, result.ID)
}

func (suite *AuthenticationServiceTestSuite) createSessionToken(idpType idp.IDPType) string {
	sessionData := AuthSessionData{
		IDPID:   testIDPID,
		IDPType: idpType,
	}
	payload := map[string]interface{}{
		"auth_data": sessionData,
	}
	payloadBytes, _ := json.Marshal(payload)
	encoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return "header." + encoded + ".signature"
}

func (suite *AuthenticationServiceTestSuite) TestValidateAndAppendAuthAssertionExtractClaimsError() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	authResponse := &common.AuthenticationResponse{
		ID:               testUserID,
		Type:             testUser.Type,
		OrganizationUnit: testUser.OrganizationUnit,
	}
	logger := log.GetLogger()

	// Create assertion without sub claim
	payload := map[string]interface{}{
		"assurance": map[string]interface{}{
			"aal": "aal1",
			"ial": "ial1",
			"authenticators": []map[string]interface{}{
				{
					"authenticator": common.AuthenticatorCredentials,
					"step":          1,
					"timestamp":     int64(1735689600),
				},
			},
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	invalidAssertion := "header." + encodedPayload + ".signature"

	suite.mockJWTService.On("VerifyJWT", invalidAssertion, "", mock.Anything).Return(nil).Once()

	svcErr := suite.service.validateAndAppendAuthAssertion(
		authResponse, testUser, common.AuthenticatorSMSOTP, invalidAssertion, logger)

	suite.NotNil(svcErr)
	suite.Equal(common.ErrorInvalidAssertion.Code, svcErr.Code)
}

func (suite *AuthenticationServiceTestSuite) TestFinishIDPAuthenticationAssertionGenerationError() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	tokenResp := &oauth.TokenResponse{
		AccessToken: testToken,
		TokenType:   "Bearer",
	}
	userInfo := map[string]interface{}{
		"sub": testUserID,
	}

	sessionToken := suite.createSessionToken(idp.IDPTypeOAuth)
	suite.mockJWTService.On("VerifyJWT", sessionToken, "auth-svc", mock.Anything).Return(nil).Once()

	suite.mockOAuthService.On("ExchangeCodeForToken", testIDPID, testAuthCode, true).
		Return(tokenResp, nil).Once()
	suite.mockOAuthService.On("FetchUserInfo", testIDPID, testToken).Return(userInfo, nil).Once()
	suite.mockOAuthService.On("GetInternalUser", testUserID).Return(testUser, nil).Once()

	// Create invalid existing assertion that will fail JWT verification
	suite.mockJWTService.On("VerifyJWT", invalidAssertion, "", mock.Anything).
		Return(errors.New("invalid signature")).Once()

	result, err := suite.service.FinishIDPAuthentication(idp.IDPTypeOAuth, sessionToken, false,
		invalidAssertion, testAuthCode)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidAssertion.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestValidateAndAppendAuthAssertionStepOne() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	authResponse := &common.AuthenticationResponse{
		ID:               testUserID,
		Type:             testUser.Type,
		OrganizationUnit: testUser.OrganizationUnit,
	}
	logger := log.GetLogger()

	suite.mockJWTService.On("GenerateJWT", testUserID, "application", mock.Anything, mock.Anything, mock.Anything).
		Return(testJWTToken, int64(3600), nil).Once()

	// Test with empty existingAssertion
	svcErr := suite.service.validateAndAppendAuthAssertion(
		authResponse, testUser, common.AuthenticatorCredentials, "", logger)
	suite.Nil(svcErr)
	suite.Equal(testJWTToken, authResponse.Assertion)
}

func (suite *AuthenticationServiceTestSuite) TestValidateAndAppendAuthAssertionSubjectMismatch() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	authResponse := &common.AuthenticationResponse{
		ID:               testUserID,
		Type:             testUser.Type,
		OrganizationUnit: testUser.OrganizationUnit,
	}

	// Create assertion with different subject
	existingAssertion := suite.createTestAssertion("different_user_id")

	suite.mockJWTService.On("VerifyJWT", existingAssertion, "", mock.Anything).Return(nil)

	svcErr := suite.service.validateAndAppendAuthAssertion(
		authResponse, testUser, common.AuthenticatorSMSOTP, existingAssertion, log.GetLogger())

	suite.NotNil(svcErr)
	suite.Equal(common.ErrorAssertionSubjectMismatch.Code, svcErr.Code)
}

func (suite *AuthenticationServiceTestSuite) TestExtractClaimsFromAssertionMissingAssurance() {
	// Create assertion without assurance claim
	assertionWithoutAssurance := suite.createTestAssertionWithoutAssurance(testUserID)

	suite.mockJWTService.On("VerifyJWT", assertionWithoutAssurance, "", mock.Anything).Return(nil)

	_, _, svcErr := suite.service.extractClaimsFromAssertion(
		assertionWithoutAssurance, log.GetLogger())

	suite.NotNil(svcErr)
	suite.Equal(common.ErrorInvalidAssertion.Code, svcErr.Code)
}

func (suite *AuthenticationServiceTestSuite) TestExtractClaimsFromAssertionErrorCases() {
	logger := log.GetLogger()

	testCases := []struct {
		name      string
		payload   map[string]interface{}
		setupMock func(assertion string)
	}{
		{
			name: "MissingSubClaim",
			payload: map[string]interface{}{
				"assurance": map[string]interface{}{
					"aal": "aal1",
					"ial": "ial1",
					"authenticators": []map[string]interface{}{
						{
							"authenticator": common.AuthenticatorCredentials,
							"step":          1,
							"timestamp":     int64(1735689600),
						},
					},
				},
			},
			setupMock: func(assertion string) {
				suite.mockJWTService.On("VerifyJWT", assertion, "", mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "InvalidSubClaimType",
			payload: map[string]interface{}{
				"sub": 12345, // Invalid: should be string
				"assurance": map[string]interface{}{
					"aal": "aal1",
					"ial": "ial1",
					"authenticators": []map[string]interface{}{
						{
							"authenticator": common.AuthenticatorCredentials,
							"step":          1,
							"timestamp":     int64(1735689600),
						},
					},
				},
			},
			setupMock: func(assertion string) {
				suite.mockJWTService.On("VerifyJWT", assertion, "", mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "EmptySubClaim",
			payload: map[string]interface{}{
				"sub": "", // Empty string
				"assurance": map[string]interface{}{
					"aal": "aal1",
					"ial": "ial1",
					"authenticators": []map[string]interface{}{
						{
							"authenticator": common.AuthenticatorCredentials,
							"step":          1,
							"timestamp":     int64(1735689600),
						},
					},
				},
			},
			setupMock: func(assertion string) {
				suite.mockJWTService.On("VerifyJWT", assertion, "", mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "MissingAssuranceClaim",
			payload: map[string]interface{}{
				"sub": testUserID,
			},
			setupMock: func(assertion string) {
				suite.mockJWTService.On("VerifyJWT", assertion, "", mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			payloadBytes, _ := json.Marshal(tc.payload)
			encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
			testAssertion := "header." + encodedPayload + ".signature"

			tc.setupMock(testAssertion)

			assuranceCtx, sub, err := suite.service.extractClaimsFromAssertion(testAssertion, logger)

			suite.Nil(assuranceCtx)
			suite.Empty(sub, "sub should be empty for test case: %s", tc.name)
			suite.NotNil(err, "error should not be nil for test case: %s", tc.name)
			suite.Equal(common.ErrorInvalidAssertion.Code, err.Code)
		})
	}
}

func (suite *AuthenticationServiceTestSuite) TestExtractClaimsFromAssertionDecodeError() {
	logger := log.GetLogger()

	// Create a malformed JWT that will fail payload decoding
	malformedAssertion := "header.not-valid-base64!!.signature"
	suite.mockJWTService.On("VerifyJWT", malformedAssertion, "", mock.Anything).Return(nil).Once()

	assuranceCtx, sub, err := suite.service.extractClaimsFromAssertion(malformedAssertion, logger)
	suite.Nil(assuranceCtx)
	suite.Empty(sub)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidAssertion.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestExtractClaimsFromAssertionUnmarshalError() {
	logger := log.GetLogger()

	// Create assertion with assurance as a value that will fail to unmarshal into AssuranceContext
	validPayload := map[string]interface{}{
		"sub":       testUserID,
		"assurance": []int{1, 2, 3},
	}
	payloadBytes, _ := json.Marshal(validPayload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	testAssertion := "header." + encodedPayload + ".signature"
	suite.mockJWTService.On("VerifyJWT", testAssertion, "", mock.Anything).Return(nil).Once()

	assuranceCtx, sub, err := suite.service.extractClaimsFromAssertion(testAssertion, logger)
	suite.Nil(assuranceCtx)
	suite.Empty(sub)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidAssertion.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestVerifyOTPJWTGenerationError() {
	sessionToken := testSessionTkn
	otpCode := "123456"
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}

	suite.mockOTPService.On("VerifyOTP", sessionToken, otpCode).Return(testUser, nil)
	suite.mockJWTService.On("GenerateJWT", testUserID, "application", mock.Anything, mock.Anything, mock.Anything).
		Return("", int64(0), errors.New("JWT generation failed"))

	result, err := suite.service.VerifyOTP(sessionToken, false, "", otpCode)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorInternalServerError.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestExtractClaimsFromAssertionInvalidJWTSignature() {
	logger := log.GetLogger()

	suite.mockJWTService.On("VerifyJWT", invalidAssertion, "", mock.Anything).
		Return(errors.New("invalid signature"))

	assuranceCtx, sub, err := suite.service.extractClaimsFromAssertion(invalidAssertion, logger)

	suite.Nil(assuranceCtx)
	suite.Empty(sub)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidAssertion.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestExtractClaimsFromAssertionMalformedAssurance() {
	logger := log.GetLogger()

	// Create assertion with invalid assurance structure
	payload := map[string]interface{}{
		"sub":       testUserID,
		"assurance": "invalid_string_instead_of_object",
	}
	payloadBytes, _ := json.Marshal(payload)
	encoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	malformedAssertion := "header." + encoded + ".signature"

	suite.mockJWTService.On("VerifyJWT", malformedAssertion, "", mock.Anything).Return(nil)

	assuranceCtx, sub, err := suite.service.extractClaimsFromAssertion(malformedAssertion, logger)

	suite.Nil(assuranceCtx)
	suite.Empty(sub)
	suite.NotNil(err)
	suite.Equal(common.ErrorInvalidAssertion.Code, err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestValidateAndAppendAuthAssertionGenerationError() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	authResponse := &common.AuthenticationResponse{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	logger := log.GetLogger()

	// Create a service with a mock assertion generator that returns an error
	mockAssertGenerator := assertmock.NewAuthAssertGeneratorInterfaceMock(suite.T())
	mockAssertGenerator.On("GenerateAssertion", mock.Anything).
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "ASSERTION_ERROR",
			Error:            "Assertion generation failed",
			ErrorDescription: "Failed to generate assertion",
		})

	service := &authenticationService{
		authAssertionGenerator: mockAssertGenerator,
		jwtService:             suite.mockJWTService,
	}

	err := service.validateAndAppendAuthAssertion(authResponse, testUser, common.AuthenticatorCredentials, "", logger)

	suite.NotNil(err)
	suite.Equal("ASSERTION_ERROR", err.Code)
}

func (suite *AuthenticationServiceTestSuite) TestValidateAndAppendAuthAssertionUpdateError() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	authResponse := &common.AuthenticationResponse{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: testOrgUnit,
	}
	logger := log.GetLogger()
	existingAssertion := suite.createTestAssertion(testUserID)

	suite.mockJWTService.On("VerifyJWT", existingAssertion, "", mock.Anything).Return(nil)

	// Create a service with a mock assertion generator that returns an error on update
	mockAssertGenerator := assertmock.NewAuthAssertGeneratorInterfaceMock(suite.T())
	mockAssertGenerator.On("UpdateAssertion", mock.Anything, mock.Anything).
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "UPDATE_ERROR",
			Error:            "Assertion update failed",
			ErrorDescription: "Failed to update assertion",
		})

	service := &authenticationService{
		authAssertionGenerator: mockAssertGenerator,
		jwtService:             suite.mockJWTService,
	}

	err := service.validateAndAppendAuthAssertion(authResponse, testUser, common.AuthenticatorSMSOTP,
		existingAssertion, logger)

	suite.NotNil(err)
	suite.Equal("UPDATE_ERROR", err.Code)
}

func (suite *AuthenticationServiceTestSuite) createTestAssertion(subject string) string {
	assuranceCtx := map[string]interface{}{
		"aal": "aal1",
		"ial": "ial1",
		"authenticators": []map[string]interface{}{
			{
				"authenticator": common.AuthenticatorCredentials,
				"step":          1,
				"timestamp":     int64(1735689600), // 2025-01-01T00:00:00Z in Unix epoch
			},
		},
	}

	payload := map[string]interface{}{
		"sub":       subject,
		"assurance": assuranceCtx,
	}

	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return fmt.Sprintf("header.%s.signature", encodedPayload)
}

func (suite *AuthenticationServiceTestSuite) createTestAssertionWithoutAssurance(subject string) string {
	payload := map[string]interface{}{
		"sub": subject,
	}

	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return fmt.Sprintf("header.%s.signature", encodedPayload)
}
