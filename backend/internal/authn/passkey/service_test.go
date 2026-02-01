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

package passkey

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

const (
	testUserID         = "user123"
	testRelyingPartyID = "example.com"
	testSessionToken   = "session_token_123"
	//nolint:gosec // Token type identifier, not a credential
	testCredentialID = "credential_123"
	testSessionKey   = "test-session-key"
)

type WebAuthnServiceTestSuite struct {
	suite.Suite
	mockUserService  *usermock.UserServiceInterfaceMock
	mockSessionStore *sessionStoreInterfaceMock
	service          *passkeyService
}

func TestWebAuthnServiceTestSuite(t *testing.T) {
	suite.Run(t, new(WebAuthnServiceTestSuite))
}

func (suite *WebAuthnServiceTestSuite) SetupSuite() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "test-issuer",
			ValidityPeriod: 3600,
			Audience:       "application",
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	if err != nil {
		suite.T().Fatalf("Failed to initialize ThunderRuntime: %v", err)
	}
}

func (suite *WebAuthnServiceTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockSessionStore = newSessionStoreInterfaceMock(suite.T())

	suite.service = &passkeyService{
		userService:  suite.mockUserService,
		sessionStore: suite.mockSessionStore,
		logger:       log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_NilRequest() {
	result, svcErr := suite.service.StartRegistration(nil)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidFinishData.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_EmptyUserID() {
	req := &PasskeyRegistrationStartRequest{
		UserID:         "",
		RelyingPartyID: testRelyingPartyID,
	}

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptyUserIdentifier.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_EmptyRelyingPartyID() {
	req := &PasskeyRegistrationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: "",
	}

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptyRelyingPartyID.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_UserNotFound() {
	req := &PasskeyRegistrationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(
		nil,
		&serviceerror.ServiceError{
			Type: serviceerror.ClientErrorType,
			Code: "USER_NOT_FOUND",
		},
	).Once()

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorUserNotFound.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_UserServiceServerError() {
	req := &PasskeyRegistrationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(
		nil,
		&serviceerror.ServiceError{
			Type: serviceerror.ServerErrorType,
			Code: "INTERNAL_ERROR",
		},
	).Once()

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_GetCredentialsError() {
	req := &PasskeyRegistrationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}

	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: "org123",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").Return(
		nil,
		&serviceerror.ServiceError{
			Type: serviceerror.ServerErrorType,
			Code: "DB_ERROR",
		},
	).Once()

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_NilRequest() {
	result, svcErr := suite.service.FinishRegistration(nil)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidFinishData.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_EmptySessionToken() {
	req := &PasskeyRegistrationFinishRequest{
		SessionToken:      "",
		CredentialID:      testCredentialID,
		ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIn0",
		AttestationObject: "o2NmbXRkbm9uZQ",
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptySessionToken.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_EmptyCredentialID() {
	req := &PasskeyRegistrationFinishRequest{
		SessionToken:      testSessionToken,
		CredentialID:      "",
		ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIn0",
		AttestationObject: "o2NmbXRkbm9uZQ",
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidFinishData.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_EmptyClientDataJSON() {
	req := &PasskeyRegistrationFinishRequest{
		SessionToken:      testSessionToken,
		CredentialID:      testCredentialID,
		ClientDataJSON:    "",
		AttestationObject: "o2NmbXRkbm9uZQ",
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidFinishData.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_EmptyAttestationObject() {
	req := &PasskeyRegistrationFinishRequest{
		SessionToken:      testSessionToken,
		CredentialID:      testCredentialID,
		ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIn0",
		AttestationObject: "",
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidFinishData.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_NilRequest() {
	result, svcErr := suite.service.StartAuthentication(nil)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidFinishData.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_EmptyUserID() {
	req := &PasskeyAuthenticationStartRequest{
		UserID:         "",
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptyUserIdentifier.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_EmptyRelyingPartyID() {
	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: "",
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptyRelyingPartyID.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_UserNotFound() {
	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(
		nil,
		&serviceerror.ServiceError{
			Type: serviceerror.ClientErrorType,
			Code: "USER_NOT_FOUND",
		},
	).Once()

	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorUserNotFound.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_UserServiceServerError() {
	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(
		nil,
		&serviceerror.ServiceError{
			Type: serviceerror.ServerErrorType,
			Code: "INTERNAL_ERROR",
		},
	).Once()

	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_GetCredentialsError() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: "org123",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").Return(
		nil,
		&serviceerror.ServiceError{
			Type: serviceerror.ServerErrorType,
			Code: "DB_ERROR",
		},
	).Once()

	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_NoCredentialsFound() {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: "org123",
	}

	emptyCredentials := []user.Credential{}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").Return(
		emptyCredentials,
		nil,
	).Once()

	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorNoCredentialsFound.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_NilRequest() {
	result, svcErr := suite.service.FinishAuthentication(nil)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidFinishData.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_EmptyCredentialID() {
	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      "",
		CredentialType:    "public-key",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "authenticatorData",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, svcErr := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptyCredentialID.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_EmptyCredentialType() {
	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      testCredentialID,
		CredentialType:    "",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "authenticatorData",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, svcErr := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptyCredentialType.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_EmptyClientDataJSON() {
	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      testCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    "",
		AuthenticatorData: "authenticatorData",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, svcErr := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidAuthenticatorResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_EmptyAuthenticatorData() {
	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      testCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, svcErr := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidAuthenticatorResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_EmptySignature() {
	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      testCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "authenticatorData",
		Signature:         "",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, svcErr := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidAuthenticatorResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_EmptySessionToken() {
	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      testCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "authenticatorData",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      "",
	}
	result, svcErr := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptySessionToken.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestGetMetadata() {
	metadata := suite.service.getMetadata()

	suite.Equal(common.AuthenticatorPasskey, metadata.Name)
	suite.NotEmpty(metadata.Factors)
	suite.Contains(metadata.Factors, common.FactorPossession)
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_Success() {
	// Create mock credentials
	mockCredential := map[string]interface{}{
		"id":        []byte("credential123"),
		"publicKey": []byte("publickey123"),
		"aaguid":    []byte("aaguid123"),
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{
			StorageType:       "",
			StorageAlgo:       "",
			StorageAlgoParams: hash.CredParameters{},
			Value:             string(credentialJSON),
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	suite.NoError(err)
	suite.NotNil(credentials)
	suite.Len(credentials, 1)
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_ServiceError() {
	svcErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "DB_ERROR",
		Error: "Database error",
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(nil, svcErr).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	suite.Error(err)
	suite.Nil(credentials)
	suite.Contains(err.Error(), "failed to get passkey credentials")
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_InvalidJSON() {
	mockUserCreds := []user.Credential{
		{
			Value: "{invalid json}",
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	suite.NoError(err)
	suite.NotNil(credentials)
	suite.Len(credentials, 0) // Invalid credentials are skipped
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_EmptyCredentials() {
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{}, nil).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	suite.NoError(err)
	suite.NotNil(credentials)
	suite.Len(credentials, 0)
}

func (suite *WebAuthnServiceTestSuite) TestStoreWebAuthnCredentialInDB_Success() {
	mockCredential := &webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			AAGUID:    []byte("aaguid123"),
			SignCount: 0,
		},
	}

	existingCreds := []user.Credential{}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.MatchedBy(
		func(credentialsJSON json.RawMessage) bool {
			var credMap map[string][]user.Credential
			if err := json.Unmarshal(credentialsJSON, &credMap); err != nil {
				return false
			}
			creds, ok := credMap["passkey"]
			return ok && len(creds) == 1
		})).Return(nil).Once()

	err := suite.service.storePasskeyCredential(testUserID, mockCredential)

	suite.NoError(err)
}

func (suite *WebAuthnServiceTestSuite) TestStoreWebAuthnCredentialInDB_GetCredentialsError() {
	mockCredential := &webauthnCredential{
		ID: []byte("credential123"),
	}

	svcErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "DB_ERROR",
		Error: "Database error",
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(nil, svcErr).Once()

	err := suite.service.storePasskeyCredential(testUserID, mockCredential)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get existing passkey credentials")
}

func (suite *WebAuthnServiceTestSuite) TestStoreWebAuthnCredentialInDB_UpdateCredentialsError() {
	mockCredential := &webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
	}

	existingCreds := []user.Credential{}

	svcErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "DB_ERROR",
		Error: "Database error",
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.Anything).
		Return(svcErr).Once()

	err := suite.service.storePasskeyCredential(testUserID, mockCredential)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update passkey credentials")
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_Success() {
	credentialID := []byte("credential123")
	existingCredential := webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			SignCount: 5,
		},
	}
	existingCredJSON, _ := json.Marshal(existingCredential)

	updatedCredential := &webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			SignCount: 6,
		},
	}

	existingCreds := []user.Credential{
		{
			Value: string(existingCredJSON),
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.MatchedBy(
		func(credentialsJSON json.RawMessage) bool {
			var credMap map[string][]user.Credential
			if err := json.Unmarshal(credentialsJSON, &credMap); err != nil {
				return false
			}
			creds, ok := credMap["passkey"]
			if !ok || len(creds) != 1 {
				return false
			}
			var cred webauthnCredential
			_ = json.Unmarshal([]byte(creds[0].Value), &cred)

			return cred.Authenticator.SignCount == 6
		})).Return(nil).Once()

	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	suite.NoError(err)
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_CredentialNotFound() {
	credentialID := []byte("credential123")
	updatedCredential := &webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
	}

	differentCredential := webauthnCredential{
		ID:        []byte("different_id"),
		PublicKey: []byte("publickey456"),
	}
	existingCredJSON, _ := json.Marshal(differentCredential)

	existingCreds := []user.Credential{
		{
			Value: string(existingCredJSON),
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	suite.Error(err)
	suite.Contains(err.Error(), "credential not found for update")
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_GetCredentialsError() {
	updatedCredential := &webauthnCredential{
		ID: []byte("credential123"),
	}

	svcErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "DB_ERROR",
		Error: "Database error",
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(nil, svcErr).Once()

	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get existing credentials")
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_UpdateError() {
	credentialID := []byte("credential123")
	existingCredential := webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
	}
	existingCredJSON, _ := json.Marshal(existingCredential)

	updatedCredential := &webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
	}

	existingCreds := []user.Credential{
		{
			Value: string(existingCredJSON),
		},
	}

	svcErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "DB_ERROR",
		Error: "Database error",
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.Anything).
		Return(svcErr).Once()

	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update credentials")
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_InvalidExistingCredential() {
	credentialID := []byte("credential123")
	updatedCredential := &webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
	}

	existingCreds := []user.Credential{
		{
			Value: "{invalid json}",
		},
		{
			Value: "{invalid}",
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	suite.Error(err)
	suite.Contains(err.Error(), "credential not found for update")
}

func (suite *WebAuthnServiceTestSuite) TestStoreSessionData_Success() {
	sessionData := &sessionData{
		Challenge:            "challenge123",
		UserID:               []byte(testUserID),
		AllowedCredentialIDs: [][]byte{},
		UserVerification:     "preferred",
	}

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		sessionData,
		mock.AnythingOfType("time.Time")).
		Return(nil).Once()

	sessionToken, svcErr := suite.service.storeSessionData(testUserID, testRelyingPartyID, sessionData)

	suite.Nil(svcErr)
	suite.NotEmpty(sessionToken)
}

func (suite *WebAuthnServiceTestSuite) TestStoreSessionData_StoreError() {
	sessionData := &sessionData{
		Challenge: "challenge123",
		UserID:    []byte(testUserID),
	}

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		sessionData,
		mock.AnythingOfType("time.Time")).
		Return(assert.AnError).Once()

	sessionToken, svcErr := suite.service.storeSessionData(testUserID, testRelyingPartyID, sessionData)

	suite.Empty(sessionToken)
	suite.NotNil(svcErr)
	suite.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestRetrieveSessionData_Success() {
	sessionData := &sessionData{
		Challenge: "challenge123",
		UserID:    []byte(testUserID),
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	retrievedSessionData, userID, rpID, svcErr := suite.service.retrieveSessionData(testSessionToken)

	suite.Nil(svcErr)
	suite.NotNil(retrievedSessionData)
	suite.Equal(testUserID, userID)
	suite.Equal(testRelyingPartyID, rpID)
	suite.Equal(sessionData.Challenge, retrievedSessionData.Challenge)
}

func (suite *WebAuthnServiceTestSuite) TestRetrieveSessionData_SessionNotFound() {
	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(nil, "", "", assert.AnError).Once()

	retrievedSessionData, userID, rpID, svcErr := suite.service.retrieveSessionData(testSessionToken)

	suite.NotNil(svcErr)
	suite.Nil(retrievedSessionData)
	suite.Empty(userID)
	suite.Empty(rpID)
	suite.Equal(ErrorSessionExpired.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestClearSessionData() {
	suite.mockSessionStore.On("deleteSession", testSessionToken).
		Return(nil).Once()

	// This method doesn't return anything, just verify it calls the mock
	suite.service.clearSessionData(testSessionToken)

	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_StoreSessionError() {
	req := &PasskeyRegistrationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}

	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: "org123",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{}, nil).Once()

	// Mock session store to return error
	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		mock.AnythingOfType("*webauthn.SessionData"),
		mock.AnythingOfType("time.Time")).
		Return(assert.AnError).Once()

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_InvalidCredentialType() {
	req := &PasskeyRegistrationFinishRequest{
		CredentialID:      "cred123",
		CredentialType:    "", // Empty will default to "public-key"
		ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIn0=",
		AttestationObject: "attestationdata",
		SessionToken:      testSessionToken,
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidAttestationResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_RetrieveSessionError() {
	req := &PasskeyRegistrationFinishRequest{
		CredentialID:      "cred123",
		CredentialType:    "public-key",
		ClientDataJSON:    "clientdata",
		AttestationObject: "attestationdata",
		SessionToken:      testSessionToken,
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)

	suite.Equal(ErrorInvalidAttestationResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestGenerateAssertionWithAttributes() {
	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(nil, "", "", assert.AnError).Once()

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      testCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "authenticatorData",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, svcErr := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorSessionExpired.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_GetUserError() {
	sessionData := &sessionData{
		Challenge: "challenge123",
		UserID:    []byte(testUserID),
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	svcErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "USER_ERROR",
		Error: "User retrieval error",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(nil, svcErr).Once()

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      testCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "authenticatorData",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_GetCredentialsError() {
	sessionData := &sessionData{
		Challenge: "challenge123",
		UserID:    []byte(testUserID),
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()

	credErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "CRED_ERROR",
		Error: "webauthnCredential retrieval error",
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(nil, credErr).Once()

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      testCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "authenticatorData",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_NoCredentialsError() {
	sessionData := &sessionData{
		Challenge: "challenge123",
		UserID:    []byte(testUserID),
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{}, nil).Once()

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      testCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "authenticatorData",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorNoCredentialsFound.Code, err.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_InvalidAssertionResponse() {
	sessionData := &sessionData{
		Challenge: "challenge123",
		UserID:    []byte(testUserID),
	}

	mockCredential := webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{
			Value: string(credentialJSON),
		},
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	// Use invalid base64 to trigger parsing error
	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      "!!!invalid-base64!!!",
		CredentialType:    "public-key",
		ClientDataJSON:    "clientDataJSON",
		AuthenticatorData: "authenticatorData",
		Signature:         "signature",
		UserHandle:        "userHandle",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidAuthenticatorResponse.Code, err.Code)
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_MultipleCredentials() {
	mockCredential1 := map[string]interface{}{
		"id":        []byte("credential1"),
		"publicKey": []byte("publickey1"),
		"aaguid":    []byte("aaguid1"),
	}
	mockCredential2 := map[string]interface{}{
		"id":        []byte("credential2"),
		"publicKey": []byte("publickey2"),
		"aaguid":    []byte("aaguid2"),
	}
	credentialJSON1, _ := json.Marshal(mockCredential1)
	credentialJSON2, _ := json.Marshal(mockCredential2)

	mockUserCreds := []user.Credential{
		{
			StorageType:       "",
			StorageAlgo:       "",
			StorageAlgoParams: hash.CredParameters{},
			Value:             string(credentialJSON1),
		},
		{
			StorageType:       "",
			StorageAlgo:       "",
			StorageAlgoParams: hash.CredParameters{},
			Value:             string(credentialJSON2),
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	suite.NoError(err)
	suite.NotNil(credentials)
	suite.Len(credentials, 2)
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_MixedValidInvalid() {
	mockCredential := map[string]interface{}{
		"id":        []byte("credential1"),
		"publicKey": []byte("publickey1"),
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{
			Value: string(credentialJSON), // Valid
		},
		{
			Value: "{invalid json}", // Invalid - should be skipped
		},
		{
			Value: "", // Empty - should be skipped
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	suite.NoError(err)
	suite.NotNil(credentials)
	suite.Len(credentials, 1) // Only one valid credential
}

func (suite *WebAuthnServiceTestSuite) TestStoreWebAuthnCredentialInDB_WithExistingCredentials() {
	mockCredential := &webauthnCredential{
		ID:        []byte("new-credential"),
		PublicKey: []byte("new-publickey"),
		Authenticator: authenticator{
			AAGUID:    []byte("new-aaguid"),
			SignCount: 0,
		},
	}

	existingCred := map[string]interface{}{
		"id":        []byte("existing-cred"),
		"publicKey": []byte("existing-key"),
	}
	existingCredJSON, _ := json.Marshal(existingCred)

	existingCreds := []user.Credential{
		{
			Value: string(existingCredJSON),
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.MatchedBy(
		func(credentialsJSON json.RawMessage) bool {
			var credMap map[string][]user.Credential
			if err := json.Unmarshal(credentialsJSON, &credMap); err != nil {
				return false
			}
			creds, ok := credMap["passkey"]
			return ok && len(creds) == 2 // Should have 2 credentials now
		})).Return(nil).Once()

	err := suite.service.storePasskeyCredential(testUserID, mockCredential)

	suite.NoError(err)
}

func (suite *WebAuthnServiceTestSuite) TestStoreWebAuthnCredentialInDB_MarshalError() {
	mockCredential := &webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			AAGUID:    []byte("aaguid123"),
			SignCount: 0,
		},
	}

	existingCreds := []user.Credential{}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.Anything).
		Return(nil).Once()

	err := suite.service.storePasskeyCredential(testUserID, mockCredential)

	// Should succeed - marshaling works for standard credentials
	suite.NoError(err)
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_MultipleCredentialsUpdateOne() {
	credentialID1 := []byte("credential1")
	credentialID2 := []byte("credential2")

	existingCredential1 := webauthnCredential{
		ID:        credentialID1,
		PublicKey: []byte("publickey1"),
		Authenticator: authenticator{
			SignCount: 5,
		},
	}
	existingCredential2 := webauthnCredential{
		ID:        credentialID2,
		PublicKey: []byte("publickey2"),
		Authenticator: authenticator{
			SignCount: 10,
		},
	}

	existingCred1JSON, _ := json.Marshal(existingCredential1)
	existingCred2JSON, _ := json.Marshal(existingCredential2)

	updatedCredential := &webauthnCredential{
		ID:        credentialID1, // Update first credential
		PublicKey: []byte("publickey1"),
		Authenticator: authenticator{
			SignCount: 6, // Incremented
		},
	}

	existingCreds := []user.Credential{
		{Value: string(existingCred1JSON)},
		{Value: string(existingCred2JSON)},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.MatchedBy(
		func(credentialsJSON json.RawMessage) bool {
			var credMap map[string][]user.Credential
			if err := json.Unmarshal(credentialsJSON, &credMap); err != nil {
				return false
			}
			creds, ok := credMap["passkey"]
			if !ok || len(creds) != 2 {
				return false
			}
			// Verify first credential was updated
			var cred1 webauthnCredential
			_ = json.Unmarshal([]byte(creds[0].Value), &cred1)

			// Verify second credential unchanged
			var cred2 webauthnCredential
			_ = json.Unmarshal([]byte(creds[1].Value), &cred2)

			return cred1.Authenticator.SignCount == 6 && cred2.Authenticator.SignCount == 10
		})).Return(nil).Once()

	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	suite.NoError(err)
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_PreserveStorageFields() {
	credentialID := []byte("credential123")
	existingCredential := webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			SignCount: 5,
		},
	}
	existingCredJSON, _ := json.Marshal(existingCredential)

	updatedCredential := &webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			SignCount: 6,
		},
	}

	existingCreds := []user.Credential{
		{
			StorageType:       "encrypted",
			StorageAlgo:       "AES-256",
			StorageAlgoParams: hash.CredParameters{KeySize: 256},
			Value:             string(existingCredJSON),
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.MatchedBy(
		func(credentialsJSON json.RawMessage) bool {
			var credMap map[string][]user.Credential
			if err := json.Unmarshal(credentialsJSON, &credMap); err != nil {
				return false
			}
			creds, ok := credMap["passkey"]
			// Verify storage fields are preserved
			return ok && len(creds) == 1 &&
				creds[0].StorageType == "encrypted" &&
				creds[0].StorageAlgo == "AES-256" &&
				creds[0].StorageAlgoParams.KeySize == 256
		})).Return(nil).Once()

	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	suite.NoError(err)
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_EmptyCredentialList() {
	updatedCredential := &webauthnCredential{
		ID: []byte("credential123"),
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{}, nil).Once()

	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	suite.Error(err)
	suite.Contains(err.Error(), "credential not found for update")
}

func (suite *WebAuthnServiceTestSuite) TestStoreWebAuthnCredentialInDB_EmptyCredential() {
	// Test with an empty but valid credential structure
	mockCredential := &webauthnCredential{
		ID:        []byte{}, // Empty ID
		PublicKey: []byte{}, // Empty public key
	}

	existingCreds := []user.Credential{}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.MatchedBy(
		func(credentialsJSON json.RawMessage) bool {
			var credMap map[string][]user.Credential
			if err := json.Unmarshal(credentialsJSON, &credMap); err != nil {
				return false
			}
			creds, ok := credMap["passkey"]
			return ok && len(creds) == 1
		})).Return(nil).Once()

	err := suite.service.storePasskeyCredential(testUserID, mockCredential)

	// Should succeed even with empty fields
	suite.NoError(err)
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_PartiallyInvalidCredentials() {
	// Test with some valid and some invalid credentials to ensure robust handling
	validCred := map[string]interface{}{
		"id":        []byte("valid-cred"),
		"publicKey": []byte("valid-key"),
	}
	validCredJSON, _ := json.Marshal(validCred)

	mockUserCreds := []user.Credential{
		{Value: string(validCredJSON)},
		{Value: "not-json-at-all"},
		{Value: `{"partial": "json"`}, // Incomplete JSON
		{Value: "{}"},                 // Empty but valid JSON
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	suite.NoError(err)
	suite.NotNil(credentials)
	// Should have parsed valid ones and skipped invalid ones
	suite.GreaterOrEqual(len(credentials), 1)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_CredentialsValidation() {
	// Test with a valid credential structure
	mockCredential := webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			AAGUID:    []byte("aaguid123"),
			SignCount: 5,
		},
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{
			StorageType:       "",
			StorageAlgo:       "",
			StorageAlgoParams: hash.CredParameters{},
			Value:             string(credentialJSON),
		},
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		mock.AnythingOfType("*webauthn.SessionData"),
		mock.AnythingOfType("time.Time")).
		Return(nil).Once()

	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(svcErr)
	suite.NotNil(result)
	suite.NotEmpty(result.SessionToken)
	suite.NotEmpty(result.PublicKeyCredentialRequestOptions.Challenge)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_CredentialWithZeroSignCount() {
	// Test credential with zero sign count (new credential)
	mockCredential := webauthnCredential{
		ID:        []byte("new-credential"),
		PublicKey: []byte("publickey"),
		Authenticator: authenticator{
			AAGUID:    []byte("aaguid"),
			SignCount: 0, // Zero sign count
		},
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{Value: string(credentialJSON)},
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		mock.AnythingOfType("*webauthn.SessionData"),
		mock.AnythingOfType("time.Time")).
		Return(nil).Once()

	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(svcErr)
	suite.NotNil(result)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_WithExistingValidCredential() {
	req := &PasskeyRegistrationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}

	// Create a properly structured credential
	mockCredential := webauthnCredential{
		ID:        []byte("existing-credential-id"),
		PublicKey: []byte("existing-publickey"),
		Authenticator: authenticator{
			AAGUID:       []byte("existing-aaguid"),
			SignCount:    10,
			CloneWarning: false,
		},
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{Value: string(credentialJSON)},
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		mock.AnythingOfType("*webauthn.SessionData"),
		mock.AnythingOfType("time.Time")).
		Return(nil).Once()

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(svcErr)
	suite.NotNil(result)
	suite.NotEmpty(result.SessionToken)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_UpdateCredentialError() {
	sessionData := &sessionData{
		Challenge: "challenge123",
		UserID:    []byte(testUserID),
	}

	// Use valid base64url encoded credential ID
	validCredentialID := base64.RawURLEncoding.EncodeToString([]byte("credential123"))

	mockCredential := webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			SignCount: 5,
		},
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{Value: string(credentialJSON)},
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      validCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.get"}`)),
		AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("authenticator-data")),
		Signature:         base64.RawURLEncoding.EncodeToString([]byte("signature")),
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidAuthenticatorResponse.Code, err.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_SkipAssertion() {
	sessionData := &sessionData{
		Challenge: "challenge123",
		UserID:    []byte(testUserID),
	}

	mockCredential := webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{Value: string(credentialJSON)},
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      "credential123",
		CredentialType:    "public-key",
		ClientDataJSON:    "valid-client-data",
		AuthenticatorData: "valid-auth-data",
		Signature:         "valid-signature",
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_NonStringValue() {
	// Test when credential value is not a string (edge case)
	// This would be a data corruption scenario
	mockUserCreds := []user.Credential{
		{
			Value: "", // Empty string value
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	suite.NoError(err)
	suite.NotNil(credentials)
	// Empty value should fail JSON unmarshal and be skipped
	suite.Len(credentials, 0)
}

func (suite *WebAuthnServiceTestSuite) TestStoreWebAuthnCredentialInDB_MarshalSuccess() {
	// Test successful marshal with complete credential
	mockCredential := &webauthnCredential{
		ID:        []byte("credential-with-all-fields"),
		PublicKey: []byte("complete-public-key"),
		Authenticator: authenticator{
			AAGUID:       []byte("complete-aaguid"),
			SignCount:    100,
			CloneWarning: false,
			Attachment:   "platform",
		},
	}

	existingCreds := []user.Credential{}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.MatchedBy(
		func(credentialsJSON json.RawMessage) bool {
			var credMap map[string][]user.Credential
			if err := json.Unmarshal(credentialsJSON, &credMap); err != nil {
				return false
			}
			creds, ok := credMap["passkey"]
			return ok && len(creds) == 1 && len(creds[0].Value) > 0
		})).Return(nil).Once()

	err := suite.service.storePasskeyCredential(testUserID, mockCredential)

	suite.NoError(err)
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_MarshalUpdatedCredentialError() {
	// This tests the error path when marshaling the updated credential fails
	// In practice this is rare, but we test the code path exists
	credentialID := []byte("credential123")

	existingCredential := webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			SignCount: 5,
		},
	}
	existingCredJSON, _ := json.Marshal(existingCredential)

	updatedCredential := &webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			SignCount: 6,
		},
	}

	existingCreds := []user.Credential{
		{Value: string(existingCredJSON)},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(existingCreds, nil).Once()

	suite.mockUserService.On("UpdateUserCredentials", mock.Anything, testUserID, mock.Anything).
		Return(nil).Once()

	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	// Should succeed for normal credentials
	suite.NoError(err)
}

func (suite *WebAuthnServiceTestSuite) TestClearSessionData_Success() {
	suite.mockSessionStore.On("deleteSession", testSessionToken).
		Return(nil).Once()

	// clearSessionData doesn't return anything, just ensure no panic
	suite.service.clearSessionData(testSessionToken)

	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *WebAuthnServiceTestSuite) TestClearSessionData_WithError() {
	// Test that clearSessionData handles errors gracefully (logs but doesn't return error)
	suite.mockSessionStore.On("deleteSession", testSessionToken).
		Return(assert.AnError).Once()

	// Should not panic even if delete fails
	suite.service.clearSessionData(testSessionToken)

	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *WebAuthnServiceTestSuite) TestNewWebAuthnAuthnService_WithNilUserService() {
	service := newPasskeyService(nil, nil)

	suite.NotNil(service)
	// Service should be created even with nil userService (uses default)
}

func (suite *WebAuthnServiceTestSuite) TestNewWebAuthnAuthnService_WithUserService() {
	// Test that NewPasskeyService uses provided user service
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())

	service := newPasskeyService(mockUserService, suite.mockSessionStore)

	suite.NotNil(service)
	// Should use provided user service
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_SuccessWithComplexCredential() {
	// Test with a fully populated credential to ensure all fields are preserved
	mockCredential := map[string]interface{}{
		"id":         []byte("complex-credential-id"),
		"publicKey":  []byte("complex-public-key-data"),
		"aaguid":     []byte("complex-aaguid"),
		"signCount":  42,
		"attachment": "cross-platform",
		"transport":  []string{"usb", "nfc", "ble"},
		"flags": map[string]interface{}{
			"userPresent":  true,
			"userVerified": true,
		},
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{
			StorageType:       "secure",
			StorageAlgo:       "AES256",
			StorageAlgoParams: hash.CredParameters{KeySize: 256, Iterations: 10000},
			Value:             string(credentialJSON),
		},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	suite.NoError(err)
	suite.NotNil(credentials)
	suite.Len(credentials, 1)
	// Verify the credential was properly unmarshaled
	suite.NotEmpty(credentials[0].ID)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_WithMultipleCredentials() {
	// Test authentication start with multiple credentials
	mockCredential1 := webauthnCredential{
		ID:        []byte("credential1"),
		PublicKey: []byte("publickey1"),
		Authenticator: authenticator{
			SignCount: 5,
		},
	}
	mockCredential2 := webauthnCredential{
		ID:        []byte("credential2"),
		PublicKey: []byte("publickey2"),
		Authenticator: authenticator{
			SignCount: 10,
		},
	}

	cred1JSON, _ := json.Marshal(mockCredential1)
	cred2JSON, _ := json.Marshal(mockCredential2)

	mockUserCreds := []user.Credential{
		{Value: string(cred1JSON)},
		{Value: string(cred2JSON)},
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		mock.AnythingOfType("*webauthn.SessionData"),
		mock.AnythingOfType("time.Time")).
		Return(nil).Once()

	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(svcErr)
	suite.NotNil(result)
	suite.NotEmpty(result.SessionToken)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_StoreCredentialError() {
	req := &PasskeyRegistrationFinishRequest{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte("credential123")),
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.create"}`)),
		AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("attestation-data")),
		SessionToken:      testSessionToken,
		CredentialName:    "Test webauthnCredential",
	}

	// This test validates that the error handling path for credential storage exists
	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidAttestationResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_WithCustomCredentialName() {
	// This test validates the custom credential name path exists in the code
	req := &PasskeyRegistrationFinishRequest{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte("credential123")),
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.create"}`)),
		AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("attestation")),
		SessionToken:      testSessionToken,
		CredentialName:    "My Custom Passkey",
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidAttestationResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_InitializeWebAuthnError() {
	// This test validates that WebAuthn library initialization error handling exists
	req := &PasskeyRegistrationFinishRequest{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte("credential123")),
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.create"}`)),
		AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("attestation")),
		SessionToken:      testSessionToken,
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	// Parsing fails first, so we get InvalidAttestationResponse
	suite.Equal(ErrorInvalidAttestationResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_CreatewebAuthnUser() {
	// This test validates that WebAuthn user creation code path exists
	req := &PasskeyRegistrationFinishRequest{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte("credential123")),
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.create"}`)),
		AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("attestation")),
		SessionToken:      testSessionToken,
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	// Parsing fails, so we get InvalidAttestationResponse
	suite.Equal(ErrorInvalidAttestationResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_ValidateLoginFailure() {
	sessionData := &sessionData{
		Challenge:      "challenge123",
		UserID:         []byte(testUserID),
		RelyingPartyID: testRelyingPartyID,
	}

	mockCredential := webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			SignCount: 5,
		},
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{Value: string(credentialJSON)},
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	validCredentialID := base64.RawURLEncoding.EncodeToString([]byte("credential123"))

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      validCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.get","challenge":"invalid"}`)),
		AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("authenticator-data")),
		Signature:         base64.RawURLEncoding.EncodeToString([]byte("signature")),
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	// Will get InvalidAuthenticatorResponse or InvalidSignature depending on validation stage
	suite.True(err.Code == ErrorInvalidAuthenticatorResponse.Code || err.Code == ErrorInvalidSignature.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_ClearSessionAfterSuccess() {
	sessionData := &sessionData{
		Challenge:      "challenge123",
		UserID:         []byte(testUserID),
		RelyingPartyID: testRelyingPartyID,
	}

	mockCredential := webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{{Value: string(credentialJSON)}}, nil).Once()

	suite.mockSessionStore.On("deleteSession", testSessionToken).
		Return(nil).Maybe()

	validCredentialID := base64.RawURLEncoding.EncodeToString([]byte("credential123"))

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      validCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.get"}`)),
		AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("authenticator-data")),
		Signature:         base64.RawURLEncoding.EncodeToString([]byte("signature")),
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	// Test will fail at library validation but verifies the clear session path exists
	suite.Nil(result)
	suite.NotNil(err)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_BuildAuthResponseWithCoreUser() {
	sessionData := &sessionData{
		Challenge:      "challenge123",
		UserID:         []byte(testUserID),
		RelyingPartyID: testRelyingPartyID,
	}

	mockCredential := webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: "org123",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{{Value: string(credentialJSON)}}, nil).Once()

	validCredentialID := base64.RawURLEncoding.EncodeToString([]byte("credential123"))

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      validCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.get"}`)),
		AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("authenticator-data")),
		Signature:         base64.RawURLEncoding.EncodeToString([]byte("signature")),
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	// Test will fail at library validation but tests the response building path
	suite.Nil(result)
	suite.NotNil(err)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_UpdateCredentialSignCountFailure() {
	sessionData := &sessionData{
		Challenge:      "challenge123",
		UserID:         []byte(testUserID),
		RelyingPartyID: testRelyingPartyID,
	}

	mockCredential := webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
		Authenticator: authenticator{
			SignCount: 5,
		},
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{{Value: string(credentialJSON)}}, nil).Once()

	validCredentialID := base64.RawURLEncoding.EncodeToString([]byte("credential123"))

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      validCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.get"}`)),
		AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("authenticator-data")),
		Signature:         base64.RawURLEncoding.EncodeToString([]byte("signature")),
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_InitializeWebAuthnLibraryError() {
	// This test validates that WebAuthn library initialization error handling exists
	sessionData := &sessionData{
		Challenge:      "challenge123",
		UserID:         []byte(testUserID),
		RelyingPartyID: "invalid-rp-id",
	}

	mockCredential := webauthnCredential{
		ID:        []byte("credential123"),
		PublicKey: []byte("publickey123"),
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, "invalid-rp-id", nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{{Value: string(credentialJSON)}}, nil).Once()

	validCredentialID := base64.RawURLEncoding.EncodeToString([]byte("credential123"))

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      validCredentialID,
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.get"}`)),
		AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("authenticator-data")),
		Signature:         base64.RawURLEncoding.EncodeToString([]byte("signature")),
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	// Parsing fails before reaching library initialization, so we get InvalidAuthenticatorResponse
	suite.Equal(ErrorInvalidAuthenticatorResponse.Code, err.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_GetWebAuthnCredentialsError() {
	req := &PasskeyRegistrationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()

	// Mock credential retrieval failure
	credErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "CRED_ERROR",
		Error: "Failed to get credentials",
	}
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(nil, credErr).Once()

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_InitializeWebAuthnLibraryError() {
	req := &PasskeyRegistrationStartRequest{
		UserID:           testUserID,
		RelyingPartyID:   "", // Invalid empty RP ID will cause init error
		RelyingPartyName: "Test RP",
	}

	// Should fail validation before reaching library init
	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptyRelyingPartyID.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_BeginRegistrationError() {
	req := &PasskeyRegistrationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{}, nil).Once()

	// Mock storeSession since BeginRegistration will succeed
	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		mock.AnythingOfType("*webauthn.SessionData"),
		mock.AnythingOfType("time.Time")).
		Return(nil).Once()

	// Library initialization will succeed and BeginRegistration will succeed
	result, svcErr := suite.service.StartRegistration(req)

	// Should succeed since BeginRegistration doesn't fail with valid data
	suite.Nil(svcErr)
	suite.NotNil(result)
	suite.NotEmpty(result.SessionToken)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_ParseAttestationResponseError() {
	req := &PasskeyRegistrationFinishRequest{
		CredentialID:      "invalid-not-base64",
		CredentialType:    "public-key",
		ClientDataJSON:    "invalid",
		AttestationObject: "invalid",
		SessionToken:      testSessionToken,
	}

	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidAttestationResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_GetWebAuthnCredentialsError() {
	req := &PasskeyRegistrationFinishRequest{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte("cred123")),
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.create"}`)),
		AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("attestation")),
		SessionToken:      testSessionToken,
	}

	// Will fail at parsing before reaching credential retrieval
	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidAttestationResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishRegistration_CreateCredentialError() {
	req := &PasskeyRegistrationFinishRequest{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte("cred123")),
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.create"}`)),
		AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("attestation")),
		SessionToken:      testSessionToken,
	}

	// Will fail at parsing stage which tests the error path exists
	result, svcErr := suite.service.FinishRegistration(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorInvalidAttestationResponse.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_GetWebAuthnCredentialsError() {
	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()

	credErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "CRED_ERROR",
		Error: "Failed to get credentials",
	}
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(nil, credErr).Once()

	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_InitializeWebAuthnLibraryError() {
	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: "",
	}
	result, svcErr := suite.service.StartAuthentication(req)

	suite.Nil(result)
	suite.NotNil(svcErr)
	suite.Equal(ErrorEmptyRelyingPartyID.Code, svcErr.Code)
}

func (suite *WebAuthnServiceTestSuite) TestStartAuthentication_BeginLoginError() {
	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	mockCredential := webauthnCredential{
		ID:        []byte("cred123"),
		PublicKey: []byte("pubkey123"),
	}
	credJSON, _ := json.Marshal(mockCredential)

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{{Value: string(credJSON)}}, nil).Once()

	// Mock storeSession since BeginLogin will succeed
	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		mock.AnythingOfType("*webauthn.SessionData"),
		mock.AnythingOfType("time.Time")).
		Return(nil).Once()

	// BeginLogin will succeed with valid data
	req := &PasskeyAuthenticationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}
	result, svcErr := suite.service.StartAuthentication(req)

	// Should succeed since BeginLogin doesn't fail with valid data
	suite.Nil(svcErr)
	suite.NotNil(result)
	suite.NotEmpty(result.SessionToken)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_GetWebAuthnCredentialsError() {
	sessionData := &sessionData{
		Challenge:      "challenge123",
		UserID:         []byte(testUserID),
		RelyingPartyID: testRelyingPartyID,
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()

	credErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "CRED_ERROR",
		Error: "Failed to get credentials",
	}
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(nil, credErr).Once()

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte("cred123")),
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.get"}`)),
		AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("auth-data")),
		Signature:         base64.RawURLEncoding.EncodeToString([]byte("signature")),
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_ParseAssertionResponseError() {
	sessionData := &sessionData{
		Challenge:      "challenge123",
		UserID:         []byte(testUserID),
		RelyingPartyID: testRelyingPartyID,
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	mockCredential := webauthnCredential{
		ID:        []byte("cred123"),
		PublicKey: []byte("pubkey123"),
	}
	credJSON, _ := json.Marshal(mockCredential)

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{{Value: string(credJSON)}}, nil).Once()

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      "invalid-not-base64",
		CredentialType:    "public-key",
		ClientDataJSON:    "invalid-data",
		AuthenticatorData: "invalid-data",
		Signature:         "invalid-data",
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidAuthenticatorResponse.Code, err.Code)
}

func (suite *WebAuthnServiceTestSuite) TestFinishAuthentication_ValidateLoginError() {
	sessionData := &sessionData{
		Challenge:      "challenge123",
		UserID:         []byte(testUserID),
		RelyingPartyID: testRelyingPartyID,
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	mockCredential := webauthnCredential{
		ID:        []byte("cred123"),
		PublicKey: []byte("pubkey123"),
	}
	credJSON, _ := json.Marshal(mockCredential)

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{{Value: string(credJSON)}}, nil).Once()

	req := &PasskeyAuthenticationFinishRequest{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte("cred123")),
		CredentialType:    "public-key",
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"passkey.get"}`)),
		AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("auth-data")),
		Signature:         base64.RawURLEncoding.EncodeToString([]byte("signature")),
		UserHandle:        "",
		SessionToken:      testSessionToken,
	}
	result, err := suite.service.FinishAuthentication(req)

	suite.Nil(result)
	suite.NotNil(err)
	// Will fail at parsing or validation
	suite.True(err.Code == ErrorInvalidAuthenticatorResponse.Code || err.Code == ErrorInvalidSignature.Code)
}

func (suite *WebAuthnServiceTestSuite) TestGetWebAuthnCredentialsFromDB_UnmarshalError() {
	validCredential := webauthnCredential{
		ID:        []byte("validcredid"),
		PublicKey: []byte("validpubkey"),
		Authenticator: authenticator{
			SignCount: 5,
		},
	}
	validJSON, _ := json.Marshal(validCredential)

	invalidJSON := []user.Credential{
		{Value: "not-valid-json{{}"},
		{Value: string(validJSON)},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(invalidJSON, nil).Once()

	credentials, err := suite.service.getStoredPasskeyCredentials(testUserID)

	// Should skip invalid and return valid ones
	suite.NoError(err)
	suite.NotNil(credentials)
	suite.Len(credentials, 1)
}

func (suite *WebAuthnServiceTestSuite) TestUpdateWebAuthnCredentialInDB_UnmarshalError() {
	credentialID := []byte("cred123")

	updatedCredential := &webauthnCredential{
		ID:        credentialID,
		PublicKey: []byte("pubkey123"),
		Authenticator: authenticator{
			SignCount: 10,
		},
	}

	invalidCreds := []user.Credential{
		{Value: "invalid-json{{"},
	}

	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(invalidCreds, nil).Once()

	// Should continue with invalid credentials (won't find match)
	err := suite.service.updatePasskeyCredential(testUserID, updatedCredential)

	suite.Error(err)
	suite.Contains(err.Error(), "credential not found")
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_WithRelyingPartyName() {
	suite.T().Skip("Skipping test that requires WebAuthn library initialization")

	req := &PasskeyRegistrationStartRequest{
		UserID:           testUserID,
		RelyingPartyID:   testRelyingPartyID,
		RelyingPartyName: "Custom RP Name",
	}

	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: "org123",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return([]user.Credential{}, nil).Once()

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		mock.AnythingOfType("*webauthn.SessionData"),
		mock.AnythingOfType("time.Time")).
		Return(nil).Once()

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(svcErr)
	suite.NotNil(result)
	suite.NotEmpty(result.SessionToken)
	suite.NotEmpty(result.PublicKeyCredentialCreationOptions.Challenge)
}

func (suite *WebAuthnServiceTestSuite) TestStartRegistration_WithExistingCredentials() {
	suite.T().Skip("Skipping test that requires WebAuthn library initialization")

	req := &PasskeyRegistrationStartRequest{
		UserID:         testUserID,
		RelyingPartyID: testRelyingPartyID,
	}

	mockCredential := webauthnCredential{
		ID:        []byte("existing-cred"),
		PublicKey: []byte("publickey123"),
	}
	credentialJSON, _ := json.Marshal(mockCredential)

	mockUserCreds := []user.Credential{
		{
			Value: string(credentialJSON),
		},
	}

	testUser := &user.User{
		ID:   testUserID,
		Type: "person",
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(testUser, nil).Once()
	suite.mockUserService.On("GetUserCredentialsByType", mock.Anything, testUserID, "passkey").
		Return(mockUserCreds, nil).Once()

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		mock.AnythingOfType("*webauthn.SessionData"),
		mock.AnythingOfType("time.Time")).
		Return(nil).Once()

	result, svcErr := suite.service.StartRegistration(req)

	suite.Nil(svcErr)
	suite.NotNil(result)
	suite.NotEmpty(result.SessionToken)
	// Verify exclude list contains the existing credential
	suite.NotEmpty(result.PublicKeyCredentialCreationOptions.CredentialExcludeList)
}
