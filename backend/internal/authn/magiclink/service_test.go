/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package magiclink

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/email"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/template"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/tests/mocks/emailmock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/templatemock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

const (
	testUserOUID  = "test-ou"
	testRecipient = "test@example.com"
	testFlowID    = "flow-123"
	testToken     = "jwt-token-123" // nolint:gosec // G101: test data, not a real secret
	testIssuedAt  = int64(1609459200)
)

// testValidJWT is a valid JWT with recipient and user_id in the standard sub claim.
// nolint:gosec // G101: test data, not a real secret
var testValidJWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
	"eyJyZWNpcGllbnQiOiJ0ZXN0QGV4YW1wbGUuY29tIiwic3ViIjoidXNlci0xMjMifQ." +
	"test-signature"

var testMissingSubJWT = "eyJhbGciOiAiSFMyNTYiLCAidHlwIjogIkpXVCJ9." +
	"eyJyZWNpcGllbnQiOiAidGVzdEBleGFtcGxlLmNvbSJ9." +
	"test-signature"

var testMismatchedUserIDJWT = "eyJhbGciOiAiSFMyNTYiLCAidHlwIjogIkpXVCJ9." +
	"eyJyZWNpcGllbnQiOiAidGVzdEBleGFtcGxlLmNvbSIsICJzdWIiOiAidXNlci00NTYifQ." +
	"test-signature"

var (
	testUserID   = "user-123"
	runtimeMutex sync.Mutex
)

func initializeTestRuntime(root string) error {
	testConfig := &config.Config{
		Server: config.ServerConfig{
			Hostname: "localhost",
			Port:     8090,
		},
		JWT: config.JWTConfig{
			Issuer: "magiclink-svc",
		},
		GateClient: config.GateClientConfig{
			Hostname:  "localhost",
			Port:      8090,
			Scheme:    "https",
			LoginPath: "/gate/signin",
		},
	}
	return config.InitializeThunderRuntime(root, testConfig)
}

type MagicLinkServiceTestSuite struct {
	suite.Suite
	mockJWTService      *jwtmock.JWTServiceInterfaceMock
	mockEmailClient     *emailmock.EmailClientInterfaceMock
	mockUserService     *userprovidermock.UserProviderInterfaceMock
	mockTemplateService *templatemock.TemplateServiceInterfaceMock
	service             MagicLinkAuthnServiceInterface
}

func TestMagicLinkServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MagicLinkServiceTestSuite))
}

func (suite *MagicLinkServiceTestSuite) SetupSuite() {
	runtimeMutex.Lock()
	config.ResetThunderRuntime()
	suite.Require().NoError(initializeTestRuntime(suite.T().TempDir()))
}

func (suite *MagicLinkServiceTestSuite) TearDownSuite() {
	config.ResetThunderRuntime()
	runtimeMutex.Unlock()
}

func (suite *MagicLinkServiceTestSuite) SetupTest() {
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockEmailClient = emailmock.NewEmailClientInterfaceMock(suite.T())
	suite.mockUserService = userprovidermock.NewUserProviderInterfaceMock(suite.T())
	suite.mockTemplateService = templatemock.NewTemplateServiceInterfaceMock(suite.T())
	suite.service = newMagicLinkAuthnService(
		suite.mockJWTService,
		suite.mockEmailClient,
		suite.mockUserService,
		suite.mockTemplateService,
	)
}

func (suite *MagicLinkServiceTestSuite) TestSendMagicLinkSuccess() {
	suite.mockUserService.On("IdentifyUser", mock.MatchedBy(func(filters map[string]interface{}) bool {
		return filters["email"] == testRecipient
	})).Return(&testUserID, nil)

	suite.mockJWTService.On("GenerateJWT",
		testUserID, tokenAudience, mock.Anything, int64(DefaultExpirySeconds),
		mock.Anything,
		jwt.TokenTypeJWT,
	).Return(testToken, testIssuedAt, nil)

	suite.mockTemplateService.On("Render", mock.Anything, template.ScenarioMagicLink,
		mock.MatchedBy(func(data template.TemplateData) bool {
			magicLinkValue := data["magicLink"]
			parsedURL, err := url.Parse(magicLinkValue)
			if err != nil {
				return false
			}

			queryValues := parsedURL.Query()
			return queryValues.Get("flowId") == testFlowID &&
				queryValues.Get("code") == testToken
		})).
		Return(&template.RenderedTemplate{
			Subject: "Sign in to your account",
			Body:    "<html>Test body</html>",
			IsHTML:  true,
		}, nil)

	suite.mockEmailClient.On("Send", mock.Anything).Return(nil)

	err := suite.service.SendMagicLink(
		context.Background(), testRecipient, int64(DefaultExpirySeconds), testFlowID, nil)
	suite.Nil(err)
}

func (suite *MagicLinkServiceTestSuite) TestSendMagicLinkInvalidRecipient() {
	tests := []struct {
		name         string
		recipient    string
		expectedCode string
	}{
		{"EmptyRecipient", "", ErrorInvalidRecipient.Code},
		{"InvalidEmailFormat", "not-an-email", ErrorInvalidRecipient.Code},
		{"WhitespaceOnly", "   ", ErrorInvalidRecipient.Code},
		{"RecipientWithTrailingCRLF", "test@example.com\r\n", ErrorInvalidRecipient.Code},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			err := suite.service.SendMagicLink(
				context.Background(),
				tc.recipient,
				int64(DefaultExpirySeconds),
				testFlowID,
				nil,
			)
			suite.NotNil(err)
			suite.Equal(tc.expectedCode, err.Code)
		})
	}
}

func (suite *MagicLinkServiceTestSuite) TestSendMagicLinkUserNotFound() {
	suite.mockUserService.On("IdentifyUser", mock.Anything).Return(nil, nil)

	err := suite.service.SendMagicLink(
		context.Background(), testRecipient, int64(DefaultExpirySeconds), testFlowID, nil)
	suite.Nil(err)
}

func (suite *MagicLinkServiceTestSuite) TestSendMagicLinkUserProviderError() {
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockUserSvc.On("IdentifyUser", mock.Anything).Return(nil, &userprovider.UserProviderError{
		Code:    userprovider.ErrorCodeUserNotFound,
		Message: "User not found",
	})

	err := svc.SendMagicLink(context.Background(), testRecipient, int64(DefaultExpirySeconds), testFlowID, nil)
	suite.Nil(err)
}

func (suite *MagicLinkServiceTestSuite) TestSendMagicLinkJWTGenerationError() {
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockUserSvc.On("IdentifyUser", mock.Anything).Return(&testUserID, nil)
	mockJWT.On("GenerateJWT",
		testUserID, tokenAudience, mock.Anything, int64(DefaultExpirySeconds), mock.Anything, jwt.TokenTypeJWT,
	).Return("", int64(0), &serviceerror.InternalServerError)

	err := svc.SendMagicLink(context.Background(), testRecipient, int64(DefaultExpirySeconds), testFlowID, nil)
	suite.NotNil(err)
	suite.Equal(ErrorTokenGenerationFailed.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestSendMagicLinkEmailSendError() {
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockUserSvc.On("IdentifyUser", mock.Anything).Return(&testUserID, nil)
	mockJWT.On("GenerateJWT",
		testUserID, tokenAudience, mock.Anything, int64(DefaultExpirySeconds), mock.Anything, jwt.TokenTypeJWT,
	).Return(testToken, testIssuedAt, nil)

	mockTemplate.On("Render", mock.Anything, template.ScenarioMagicLink, mock.Anything).
		Return(&template.RenderedTemplate{
			Subject: "Sign in to your account",
			Body:    "<html>Test body</html>",
			IsHTML:  true,
		}, nil)

	// Return a regular error, not a serviceerror.I18nServiceError
	mockEmail.On("Send", mock.Anything).Return(errors.New("email send failed"))

	err := svc.SendMagicLink(context.Background(), testRecipient, int64(DefaultExpirySeconds), testFlowID, nil)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerErrorWithI18n.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestVerifyMagicLinkEmptyToken() {
	result, err := suite.service.VerifyMagicLink(context.Background(), "")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidToken.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestVerifyMagicLinkExpiredToken() {
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	expiredErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             jwt.ErrorTokenExpired.Code,
		ErrorDescription: "Token has expired",
	}
	mockJWT.On("VerifyJWT", testToken, tokenAudience, mock.Anything).Return(expiredErr)

	result, err := svc.VerifyMagicLink(context.Background(), testToken)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorExpiredToken.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestVerifyMagicLinkInvalidToken() {
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockJWT.On("VerifyJWT", testToken, tokenAudience, mock.Anything).Return(&serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT_INVALID",
		ErrorDescription: "Token is invalid",
	})

	result, err := svc.VerifyMagicLink(context.Background(), testToken)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidToken.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestVerifyMagicLinkSuccess() {
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockJWT.On("VerifyJWT", testValidJWT, tokenAudience, mock.Anything).Return(nil)

	testUser := &userprovider.User{
		UserID:   testUserID,
		OUID:     testUserOUID,
		UserType: "person",
	}
	mockUserSvc.On("GetUser", testUserID).Return(testUser, nil)

	result, err := svc.VerifyMagicLink(context.Background(), testValidJWT)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(testUserID, result.UserID)
	suite.Equal(testUserOUID, result.OUID)
}
func (suite *MagicLinkServiceTestSuite) TestVerifyMagicLinkMissingSubjectClaim() {
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockJWT.On("VerifyJWT", testMissingSubJWT, tokenAudience, mock.Anything).Return(nil)

	result, err := svc.VerifyMagicLink(context.Background(), testMissingSubJWT)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMalformedTokenClaims.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestVerifyMagicLinkUserIDMismatchClaim() {
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockJWT.On("VerifyJWT", testMismatchedUserIDJWT, tokenAudience, mock.Anything).Return(nil)
	mockUserSvc.On("GetUser", "user-456").Return(nil, &userprovider.UserProviderError{
		Code:    userprovider.ErrorCodeUserNotFound,
		Message: "User not found",
	})

	result, err := svc.VerifyMagicLink(context.Background(), testMismatchedUserIDJWT)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorUserNotFoundWithI18n.Code, err.Code)
}
func (suite *MagicLinkServiceTestSuite) TestVerifyMagicLinkUserNotFoundOnVerify() {
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockJWT.On("VerifyJWT", testValidJWT, tokenAudience, mock.Anything).Return(nil)
	mockUserSvc.On("GetUser", testUserID).Return(nil, &userprovider.UserProviderError{
		Code:    userprovider.ErrorCodeUserNotFound,
		Message: "User not found",
	})

	result, err := svc.VerifyMagicLink(context.Background(), testValidJWT)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorUserNotFoundWithI18n.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestVerifyMagicLinkGetUserError() {
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockJWT.On("VerifyJWT", testValidJWT, tokenAudience, mock.Anything).Return(nil)
	mockUserSvc.On("GetUser", testUserID).Return(nil, &userprovider.UserProviderError{
		Code:    userprovider.ErrorCodeUserNotFound,
		Message: "User not found",
	})

	result, err := svc.VerifyMagicLink(context.Background(), testValidJWT)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(common.ErrorUserNotFoundWithI18n.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestVerifyMagicLinkUserProviderSystemError() {
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockJWT.On("VerifyJWT", testValidJWT, tokenAudience, mock.Anything).Return(nil)
	mockUserSvc.On("GetUser", testUserID).Return(nil, &userprovider.UserProviderError{
		Code:        userprovider.ErrorCodeSystemError,
		Message:     "System error",
		Description: "Database connection failed",
	})

	result, err := svc.VerifyMagicLink(context.Background(), testValidJWT)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerErrorWithI18n.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestGetAuthenticatorMetadata() {
	metadata := suite.service.(*magicLinkAuthnService).getMetadata()
	suite.Equal(common.AuthenticatorMagicLink, metadata.Name)
	suite.Len(metadata.Factors, 1)
	suite.Contains(metadata.Factors, common.FactorPossession)
}

func (suite *MagicLinkServiceTestSuite) TestIsValidEmail() {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.org", true},
		{"invalid", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
	}

	for _, tc := range tests {
		suite.Run(tc.email, func() {
			result := email.IsValidEmail(tc.email)
			suite.Equal(tc.expected, result, "email.IsValidEmail(%q) should be %v", tc.email, tc.expected)
		})
	}
}

func (suite *MagicLinkServiceTestSuite) TestSendMagicLinkTemplateRenderError() {
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockUserSvc.On("IdentifyUser", mock.Anything).Return(&testUserID, nil)
	mockJWT.On("GenerateJWT",
		testUserID, tokenAudience, mock.Anything, int64(DefaultExpirySeconds), mock.Anything, jwt.TokenTypeJWT,
	).Return(testToken, testIssuedAt, nil)

	mockTemplate.On("Render", mock.Anything, template.ScenarioMagicLink, mock.Anything).
		Return(nil, &serviceerror.I18nServiceError{Code: "TEMPLATE_ERROR"})

	err := svc.SendMagicLink(context.Background(), testRecipient, int64(DefaultExpirySeconds), testFlowID, nil)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerErrorWithI18n.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestSendMagicLinkNilEmailClient() {
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, nil, mockUserSvc, mockTemplate)

	mockUserSvc.On("IdentifyUser", mock.Anything).Return(&testUserID, nil)
	mockJWT.On("GenerateJWT",
		testUserID, tokenAudience, mock.Anything, int64(DefaultExpirySeconds), mock.Anything, jwt.TokenTypeJWT,
	).Return(testToken, testIssuedAt, nil)

	mockTemplate.On("Render", mock.Anything, template.ScenarioMagicLink, mock.Anything).
		Return(&template.RenderedTemplate{
			Subject: "Sign in to your account",
			Body:    "<html>Test body</html>",
			IsHTML:  true,
		}, nil)

	err := svc.SendMagicLink(context.Background(), testRecipient, int64(DefaultExpirySeconds), testFlowID, nil)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerErrorWithI18n.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestSendMagicLinkUserProviderSystemError() {
	mockUserSvc := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockEmail := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplate := templatemock.NewTemplateServiceInterfaceMock(suite.T())
	svc := newMagicLinkAuthnService(mockJWT, mockEmail, mockUserSvc, mockTemplate)

	mockUserSvc.On("IdentifyUser", mock.Anything).Return(nil, &userprovider.UserProviderError{
		Code:    userprovider.ErrorCodeSystemError,
		Message: "System error",
	})

	err := svc.SendMagicLink(context.Background(), testRecipient, int64(DefaultExpirySeconds), testFlowID, nil)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerErrorWithI18n.Code, err.Code)
}

func (suite *MagicLinkServiceTestSuite) TestBuildMagicLinkURLUsesQueryParams() {
	service := suite.service.(*magicLinkAuthnService)

	result := service.buildMagicLinkURL(testFlowID, testToken, nil)
	parsedURL, err := url.Parse(result)

	suite.Require().NoError(err)
	suite.Equal("/gate/signin", parsedURL.Path)
	suite.Equal(testFlowID, parsedURL.Query().Get("flowId"))
	suite.Equal(testToken, parsedURL.Query().Get("code"))
}

func (suite *MagicLinkServiceTestSuite) TestBuildMagicLinkURLUsesQueryParamsForCustomURL() {
	service := suite.service.(*magicLinkAuthnService)
	result := service.buildMagicLinkURL(testFlowID, testToken, new("https://example.com/signin?tenant=alpha"))
	parsedURL, err := url.Parse(result)

	suite.Require().NoError(err)
	suite.Equal("alpha", parsedURL.Query().Get("tenant"))
	suite.Equal(testFlowID, parsedURL.Query().Get("flowId"))
	suite.Equal(testToken, parsedURL.Query().Get("code"))
}
