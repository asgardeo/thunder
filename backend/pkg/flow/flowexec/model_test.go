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

package flowexec

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/system/config"
	authnprovidercm "github.com/asgardeo/thunder/pkg/authnprovider/common"
	"github.com/asgardeo/thunder/pkg/flow/common"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
)

const (
	testUserID789 = "user-789"
)

type ModelTestSuite struct {
	suite.Suite
}

func TestModelTestSuite(t *testing.T) {
	// Setup test config with encryption key
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Encryption: config.EncryptionConfig{
				Key: "2729a7928c79371e5f312167269294a14bb0660fd166b02a408a20fa73271580",
			},
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/test/thunderid/home", testConfig)
	if err != nil {
		t.Fatalf("failed to initialize server runtime: %v", err)
	}

	suite.Run(t, new(ModelTestSuite))
}

func (s *ModelTestSuite) getContextContent(dbModel *FlowContextDB) flowContextContent {
	err := dbModel.decrypt(context.Background())
	s.NoError(err)
	var content flowContextContent
	err = json.Unmarshal([]byte(dbModel.Context), &content)
	s.NoError(err)
	return content
}

func (s *ModelTestSuite) TestFromEngineContext_WithToken() {
	// Setup
	testToken := "test-token-123456"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		Context:     context.Background(),
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		Verbose:     true,
		FlowType:    common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "testuser",
		},
		RuntimeData: map[string]string{
			"key": "value",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
			Token:           testToken,
			Attributes: map[string]interface{}{
				"email": "test@example.com",
			},
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	// Execute
	dbModel, err := FromEngineContext(ctx)

	// Verify
	s.NoError(err)
	s.NotNil(dbModel)
	s.Equal("test-flow-id", dbModel.ExecutionID)

	content := s.getContextContent(dbModel)
	s.Equal("test-app-id", content.AppID)
	s.True(content.Verbose)
	s.True(content.IsAuthenticated)
	s.NotNil(content.UserID)
	s.Equal("user-123", *content.UserID)

	// Verify token is stored in decrypted context
	s.NotNil(content.Token)
	s.Equal(testToken, *content.Token)
}

func (s *ModelTestSuite) TestFromEngineContext_WithoutToken() {
	// Setup
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		Context:     context.Background(),
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		Verbose:     false,
		FlowType:    common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "testuser",
		},
		RuntimeData: map[string]string{},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
			Token:           "", // Empty token
			Attributes:      map[string]interface{}{},
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	// Execute
	dbModel, err := FromEngineContext(ctx)

	// Verify
	s.NoError(err)
	s.NotNil(dbModel)
	s.Equal("test-flow-id", dbModel.ExecutionID)

	content := s.getContextContent(dbModel)
	s.True(content.IsAuthenticated)

	// Verify token is nil when empty
	s.Nil(content.Token)
}

func (s *ModelTestSuite) TestFromEngineContext_WithEmptyAuthenticatedUser() {
	// Setup
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		Context:           context.Background(),
		ExecutionID:       "test-flow-id",
		AppID:             "test-app-id",
		Verbose:           false,
		FlowType:          common.FlowTypeAuthentication,
		UserInputs:        map[string]string{},
		RuntimeData:       map[string]string{},
		AuthenticatedUser: authncm.AuthenticatedUser{}, // Empty authenticated user
		ExecutionHistory:  map[string]*common.NodeExecutionRecord{},
		Graph:             mockGraph,
	}

	// Execute
	dbModel, err := FromEngineContext(ctx)

	// Verify
	s.NoError(err)
	s.NotNil(dbModel)

	content := s.getContextContent(dbModel)
	s.False(content.IsAuthenticated)
	s.Nil(content.UserID)
	s.Nil(content.Token)
}

func (s *ModelTestSuite) TestToEngineContext_WithToken() {
	// Setup - First create an encrypted token
	testToken := "test-token-xyz789"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	// Create the context and convert to DB model to get encrypted token
	ctx := EngineContext{
		Context:     context.Background(),
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-456",
			Token:           testToken,
			Attributes: map[string]interface{}{
				"role": "admin",
			},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	dbModel, err := FromEngineContext(ctx)
	s.NoError(err)
	content := s.getContextContent(dbModel)
	s.NotNil(content.Token)

	// Execute - Convert back to EngineContext
	resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)

	// Verify
	s.NoError(err)
	s.Equal("test-flow-id", resultCtx.ExecutionID)
	s.Equal("test-app-id", resultCtx.AppID)
	s.True(resultCtx.AuthenticatedUser.IsAuthenticated)
	s.Equal("user-456", resultCtx.AuthenticatedUser.UserID)

	// Verify token is decrypted correctly
	s.Equal(testToken, resultCtx.AuthenticatedUser.Token)
}

func (s *ModelTestSuite) TestToEngineContext_WithoutToken() {
	// Setup
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	userInputs := `{"username":"testuser"}`
	runtimeData := `{"key":"value"}`
	userAttributes := `{"email":"test@example.com"}`
	executionHistory := `{}`
	userID := testUserID789

	content := flowContextContent{
		AppID:            "test-app-id",
		Verbose:          true,
		GraphID:          "test-graph-id",
		IsAuthenticated:  true,
		UserID:           &userID,
		UserInputs:       &userInputs,
		RuntimeData:      &runtimeData,
		UserAttributes:   &userAttributes,
		ExecutionHistory: &executionHistory,
		Token:            nil, // No token
	}
	contextJSON, _ := json.Marshal(content)
	dbModel := &FlowContextDB{
		ExecutionID: "test-flow-id",
		Context:     string(contextJSON),
	}

	// Execute
	resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)

	// Verify
	s.NoError(err)
	s.Equal("test-flow-id", resultCtx.ExecutionID)
	s.True(resultCtx.AuthenticatedUser.IsAuthenticated)
	s.Equal(testUserID789, resultCtx.AuthenticatedUser.UserID)

	// Verify token is empty string when nil
	s.Equal("", resultCtx.AuthenticatedUser.Token)
}

func (s *ModelTestSuite) TestTokenEncryptionDecryptionRoundTrip() {
	// Setup
	testTokens := []string{
		"simple-token",
		"token-with-special-chars-!@#$%^&*()",
		"very-long-token-" + string(make([]byte, 1000)),
		"unicode-token-🔐🔑",
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id").Maybe()
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication).Maybe()

	for _, testToken := range testTokens {
		s.Run("Token: "+testToken[:min(20, len(testToken))], func() {
			// Create context with token
			ctx := EngineContext{
				Context:     context.Background(),
				ExecutionID: "test-flow-id",
				AppID:       "test-app-id",
				FlowType:    common.FlowTypeAuthentication,
				AuthenticatedUser: authncm.AuthenticatedUser{
					IsAuthenticated: true,
					UserID:          "user-123",
					Token:           testToken,
					Attributes:      map[string]interface{}{},
				},
				UserInputs:       map[string]string{},
				RuntimeData:      map[string]string{},
				ExecutionHistory: map[string]*common.NodeExecutionRecord{},
				Graph:            mockGraph,
			}

			// Convert to DB model (encrypts entire context)
			dbModel, err := FromEngineContext(ctx)
			s.NoError(err)

			// Verify context is encrypted (contains ciphertext envelope)
			s.Contains(dbModel.Context, `"ct"`, "context should be encrypted")

			// Decrypt and convert back to EngineContext
			s.NoError(dbModel.decrypt(context.Background()))
			resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)
			s.NoError(err)

			// Verify original token is restored
			s.Equal(testToken, resultCtx.AuthenticatedUser.Token)
		})
	}
}

func (s *ModelTestSuite) TestDecrypt_WithInvalidCiphertext() {
	// Context is valid JSON with the encrypted envelope structure but has corrupted ciphertext.
	testCases := []struct {
		name    string
		context string
	}{
		{
			name:    "invalid base64 in ct field",
			context: `{"alg":"AES-GCM","ct":"not-valid-base64!!!","kid":"key-1"}`,
		},
		{
			name:    "empty ct field",
			context: `{"alg":"AES-GCM","ct":"","kid":"key-1"}`,
		},
		{
			name:    "ct too short to contain nonce",
			context: `{"alg":"AES-GCM","ct":"dGVzdA==","kid":"key-1"}`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			dbModel := &FlowContextDB{ExecutionID: "test-flow-id", Context: tc.context}
			err := dbModel.decrypt(context.Background())
			s.Error(err)
		})
	}
}

func (s *ModelTestSuite) TestGetGraphID_WithDecryptionFailure() {
	// Verifies that GetGraphID propagates a decryption error when the context
	// looks encrypted but cannot be decrypted.
	dbModel := &FlowContextDB{
		ExecutionID: "test-flow-id",
		Context:     `{"alg":"AES-GCM","ct":"not-valid-base64!!!","kid":"key-1"}`,
	}

	graphID, err := dbModel.GetGraphID(context.Background())
	s.Error(err)
	s.Empty(graphID)
}

func (s *ModelTestSuite) TestToEngineContext_WithDecryptionFailure() {
	// Verifies that ToEngineContext propagates a decryption error when the context
	// looks encrypted but cannot be decrypted.
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	dbModel := &FlowContextDB{
		ExecutionID: "test-flow-id",
		Context:     `{"alg":"AES-GCM","ct":"not-valid-base64!!!","kid":"key-1"}`,
	}

	result, err := dbModel.ToEngineContext(context.Background(), mockGraph)
	s.Error(err)
	s.Equal(EngineContext{}, result)
}

func (s *ModelTestSuite) TestDecrypt_WithDecryptionFailure() {
	// Verifies that decrypt returns an error (and does not panic) when
	// the context looks encrypted but the ciphertext is invalid.
	dbModel := &FlowContextDB{
		ExecutionID: "test-flow-id",
		Context:     `{"alg":"AES-GCM","ct":"not-valid-base64!!!","kid":"key-1"}`,
	}

	s.True(dbModel.isEncrypted(), "context should be detected as encrypted before attempt")

	err := dbModel.decrypt(context.Background())
	s.Error(err)
	// Context should remain unchanged (still in its original encrypted form)
	s.True(dbModel.isEncrypted(), "context should remain encrypted after failed decryption")
}

func (s *ModelTestSuite) TestContextEncryptionRoundTrip() {
	// Tests that the entire context JSON is encrypted and all fields survive the round trip.
	testCases := []struct {
		name    string
		appID   string
		userID  string
		inputs  map[string]string
		runtime map[string]string
	}{
		{
			name:    "full context",
			appID:   "app-full-context",
			userID:  "user-full-context",
			inputs:  map[string]string{"username": "testuser", "password": "secret"},
			runtime: map[string]string{"state": "abc123", "nonce": "xyz789"},
		},
		{
			name:    "minimal context",
			appID:   "app-minimal",
			userID:  "",
			inputs:  map[string]string{},
			runtime: map[string]string{},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id").Maybe()
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication).Maybe()

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := EngineContext{
				ExecutionID: "test-flow-id",
				AppID:       tc.appID,
				FlowType:    common.FlowTypeAuthentication,
				AuthenticatedUser: authncm.AuthenticatedUser{
					IsAuthenticated: tc.userID != "",
					UserID:          tc.userID,
					Attributes:      map[string]interface{}{},
				},
				UserInputs:       tc.inputs,
				RuntimeData:      tc.runtime,
				ExecutionHistory: map[string]*common.NodeExecutionRecord{},
				Graph:            mockGraph,
			}

			// Encrypt
			dbModel, err := FromEngineContext(ctx)
			s.NoError(err)

			// Verify context is encrypted: ciphertext envelope present, plain fields hidden
			s.Contains(dbModel.Context, `"ct"`, "encrypted context should contain ciphertext field")
			s.NotContains(dbModel.Context, tc.appID, "appId should not be visible in encrypted context")

			// Decrypt and verify all fields are restored
			err = dbModel.decrypt(context.Background())
			s.NoError(err)
			s.NotContains(dbModel.Context, `"ct"`, "decrypted context should not contain ciphertext field")
			s.Contains(dbModel.Context, `"appId"`, "decrypted context should expose plain appId field")

			resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)
			s.NoError(err)
			s.Equal(tc.appID, resultCtx.AppID)
			s.Equal(tc.userID, resultCtx.AuthenticatedUser.UserID)
			s.Equal(len(tc.inputs), len(resultCtx.UserInputs))
			s.Equal(len(tc.runtime), len(resultCtx.RuntimeData))
		})
	}
}

func (s *ModelTestSuite) TestIsEncrypted() {
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			Attributes: map[string]interface{}{},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	dbModel, err := FromEngineContext(ctx)
	s.NoError(err)

	// Freshly created DB model should have encrypted context
	s.True(dbModel.isEncrypted(), "freshly encrypted context should be detected as encrypted")

	// After decryption, should no longer be detected as encrypted
	err = dbModel.decrypt(context.Background())
	s.NoError(err)
	s.False(dbModel.isEncrypted(), "decrypted context should not be detected as encrypted")

	// Plain JSON without "alg" field should not be detected as encrypted
	plainModel := &FlowContextDB{
		ExecutionID: "plain-id",
		Context:     `{"appId":"test","graphId":"graph-1","isAuthenticated":false}`,
	}
	s.False(plainModel.isEncrypted(), "plain JSON context should not be detected as encrypted")

	// Non-JSON string should not be detected as encrypted
	invalidModel := &FlowContextDB{ExecutionID: "invalid-id", Context: "not-json-at-all"}
	s.False(invalidModel.isEncrypted(), "non-JSON string should not be detected as encrypted")
}

func (s *ModelTestSuite) TestDecrypt_WithEncryptedContext() {
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "ensure-decrypt-app",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-ensure",
			Token:           "ensure-token",
			Attributes:      map[string]interface{}{},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	dbModel, err := FromEngineContext(ctx)
	s.NoError(err)
	s.True(dbModel.isEncrypted())

	err = dbModel.decrypt(context.Background())
	s.NoError(err)
	s.False(dbModel.isEncrypted(), "context should be decrypted after decrypt")

	// Verify context is usable after decrypt
	resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)
	s.NoError(err)
	s.Equal("ensure-decrypt-app", resultCtx.AppID)
	s.Equal("user-ensure", resultCtx.AuthenticatedUser.UserID)
	s.Equal("ensure-token", resultCtx.AuthenticatedUser.Token)
}

func (s *ModelTestSuite) TestDecrypt_WithPlainContext() {
	// Plain JSON context should pass through decrypt unchanged
	plainJSON := `{"appId":"test-app","graphId":"graph-1","isAuthenticated":false}`
	dbModel := &FlowContextDB{ExecutionID: "test-flow-id", Context: plainJSON}

	s.False(dbModel.isEncrypted())
	originalContext := dbModel.Context

	err := dbModel.decrypt(context.Background())
	s.NoError(err)
	s.Equal(originalContext, dbModel.Context, "plain context should be unchanged by decrypt")
}

func (s *ModelTestSuite) TestFromEngineContext_PreservesOtherFields() {
	// Setup
	testToken := "test-token-preserve-fields"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("graph-123")

	currentAction := "test-action"
	ctx := EngineContext{
		Context:       context.Background(),
		ExecutionID:   "flow-123",
		AppID:         "app-123",
		Verbose:       true,
		FlowType:      common.FlowTypeAuthentication,
		CurrentAction: currentAction,
		UserInputs: map[string]string{
			"input1": "value1",
			"input2": "value2",
		},
		RuntimeData: map[string]string{
			"runtime1": "val1",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-abc",
			OUID:            "org-xyz",
			UserType:        "admin",
			Token:           testToken,
			Attributes: map[string]interface{}{
				"attr1": "value1",
			},
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{
			"node1": {NodeID: "node1"},
		},
		Graph: mockGraph,
	}

	// Execute
	dbModel, err := FromEngineContext(ctx)

	// Verify all fields are preserved
	s.NoError(err)
	s.Equal("flow-123", dbModel.ExecutionID)

	content := s.getContextContent(dbModel)
	s.Equal("app-123", content.AppID)
	s.True(content.Verbose)
	s.NotNil(content.CurrentAction)
	s.Equal(currentAction, *content.CurrentAction)
	s.Equal("graph-123", content.GraphID)
	s.True(content.IsAuthenticated)
	s.NotNil(content.UserID)
	s.Equal("user-abc", *content.UserID)
	s.NotNil(content.OUID)
	s.Equal("org-xyz", *content.OUID)
	s.NotNil(content.UserType)
	s.Equal("admin", *content.UserType)
	s.NotNil(content.UserInputs)
	s.NotNil(content.RuntimeData)
	s.NotNil(content.UserAttributes)
	s.NotNil(content.ExecutionHistory)
	s.NotNil(content.Token)
}

func (s *ModelTestSuite) TestFromEngineContext_WithAvailableAttributes() {
	// Setup
	testAvailableAttributes := &authnprovidercm.AttributesResponse{
		Attributes: map[string]*authnprovidercm.AttributeResponse{
			"email": {
				AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
					IsVerified: true,
				},
			},
			"phoneNumber": {
				AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
					IsVerified: false,
				},
			},
		},
		Verifications: map[string]*authnprovidercm.VerificationResponse{},
	}
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		Context:     context.Background(),
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		Verbose:     true,
		FlowType:    common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "testuser",
		},
		RuntimeData: map[string]string{
			"key": "value",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:     true,
			UserID:              "user-123",
			AvailableAttributes: testAvailableAttributes,
			Attributes: map[string]interface{}{
				"email": "test@example.com",
			},
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	// Execute
	dbModel, err := FromEngineContext(ctx)

	// Verify
	s.NoError(err)
	s.NotNil(dbModel)
	s.Equal("test-flow-id", dbModel.ExecutionID)

	content := s.getContextContent(dbModel)
	s.Equal("test-app-id", content.AppID)
	s.True(content.Verbose)
	s.True(content.IsAuthenticated)
	s.NotNil(content.UserID)
	s.Equal("user-123", *content.UserID)

	// Verify available attributes are serialized (not encrypted)
	s.NotNil(content.AvailableAttributes)
	s.Greater(len(*content.AvailableAttributes), 0)

	// Verify available attributes can be deserialized back
	s.Contains(*content.AvailableAttributes, "\"email\"")
	s.Contains(*content.AvailableAttributes, "\"phoneNumber\"")
}

func (s *ModelTestSuite) TestFromEngineContext_WithoutAvailableAttributes() {
	// Setup
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		Context:     context.Background(),
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		Verbose:     false,
		FlowType:    common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "testuser",
		},
		RuntimeData: map[string]string{},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:     true,
			UserID:              "user-123",
			AvailableAttributes: nil, // No available attributes
			Attributes:          map[string]interface{}{},
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	// Execute
	dbModel, err := FromEngineContext(ctx)

	// Verify
	s.NoError(err)
	s.NotNil(dbModel)
	s.Equal("test-flow-id", dbModel.ExecutionID)

	content := s.getContextContent(dbModel)
	s.True(content.IsAuthenticated)

	// Verify available attributes is nil when empty
	s.Nil(content.AvailableAttributes)
}

func (s *ModelTestSuite) TestToEngineContext_WithAvailableAttributes() {
	// Setup
	testAvailableAttributes := &authnprovidercm.AttributesResponse{
		Attributes: map[string]*authnprovidercm.AttributeResponse{
			"email": {
				AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
					IsVerified: true,
				},
			},
			"address": {
				AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
					IsVerified: false,
				},
			},
		},
		Verifications: map[string]*authnprovidercm.VerificationResponse{},
	}
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	// Create the context and convert to DB model to get serialized available attributes
	ctx := EngineContext{
		Context:     context.Background(),
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:     true,
			UserID:              "user-456",
			AvailableAttributes: testAvailableAttributes,
			Attributes: map[string]interface{}{
				"role": "admin",
			},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	dbModel, err := FromEngineContext(ctx)
	s.NoError(err)
	content := s.getContextContent(dbModel)
	s.NotNil(content.AvailableAttributes)

	// Execute - Convert back to EngineContext
	resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)

	// Verify
	s.NoError(err)
	s.Equal("test-flow-id", resultCtx.ExecutionID)
	s.Equal("test-app-id", resultCtx.AppID)
	s.True(resultCtx.AuthenticatedUser.IsAuthenticated)
	s.Equal("user-456", resultCtx.AuthenticatedUser.UserID)

	// Verify available attributes are deserialized correctly
	s.NotNil(resultCtx.AuthenticatedUser.AvailableAttributes)
	s.Len(resultCtx.AuthenticatedUser.AvailableAttributes.Attributes, 2)
	s.Contains(resultCtx.AuthenticatedUser.AvailableAttributes.Attributes, "email")
	s.Contains(resultCtx.AuthenticatedUser.AvailableAttributes.Attributes, "address")
	s.True(resultCtx.AuthenticatedUser.AvailableAttributes.Attributes["email"].AssuranceMetadataResponse.IsVerified)
	s.False(resultCtx.AuthenticatedUser.AvailableAttributes.Attributes["address"].AssuranceMetadataResponse.IsVerified)
}

func (s *ModelTestSuite) TestToEngineContext_WithoutAvailableAttributes() {
	// Setup
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	userInputs := `{"username":"testuser"}`
	runtimeData := `{"key":"value"}`
	userAttributes := `{"email":"test@example.com"}`
	executionHistory := `{}`
	userID := "user-987"

	content := flowContextContent{
		AppID:               "test-flow-id",
		Verbose:             true,
		GraphID:             "test-graph-id",
		IsAuthenticated:     true,
		UserID:              &userID,
		UserInputs:          &userInputs,
		RuntimeData:         &runtimeData,
		UserAttributes:      &userAttributes,
		ExecutionHistory:    &executionHistory,
		AvailableAttributes: nil, // No available attributes
	}
	contextJSON, _ := json.Marshal(content)
	dbModel := &FlowContextDB{
		ExecutionID: "test-flow-id",
		Context:     string(contextJSON),
	}

	// Execute
	resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)

	// Verify
	s.NoError(err)
	s.Equal("test-flow-id", resultCtx.ExecutionID)
	s.True(resultCtx.AuthenticatedUser.IsAuthenticated)
	s.Equal("user-987", resultCtx.AuthenticatedUser.UserID)

	// Verify available attributes is nil/empty when not provided
	s.Nil(resultCtx.AuthenticatedUser.AvailableAttributes)
}

func (s *ModelTestSuite) TestAvailableAttributesSerializationRoundTrip() {
	// Setup
	testCases := []struct {
		name       string
		attributes *authnprovidercm.AttributesResponse
	}{
		{
			name: "Single attribute",
			attributes: &authnprovidercm.AttributesResponse{
				Attributes: map[string]*authnprovidercm.AttributeResponse{
					"email": {
						AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
							IsVerified: true,
						},
					},
				},
				Verifications: map[string]*authnprovidercm.VerificationResponse{},
			},
		},
		{
			name: "Multiple attributes",
			attributes: &authnprovidercm.AttributesResponse{
				Attributes: map[string]*authnprovidercm.AttributeResponse{
					"email": {
						AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
							IsVerified: true,
						},
					},
					"phone": {
						AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
							IsVerified: false,
						},
					},
					"address": {
						AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
							IsVerified: true,
						},
					},
				},
				Verifications: map[string]*authnprovidercm.VerificationResponse{},
			},
		},
		{
			name: "Special characters in names",
			attributes: &authnprovidercm.AttributesResponse{
				Attributes: map[string]*authnprovidercm.AttributeResponse{
					"custom-attr-1": {
						AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
							IsVerified: true,
						},
					},
					"attr_with_underscore": {
						AssuranceMetadataResponse: &authnprovidercm.AssuranceMetadataResponse{
							IsVerified: false,
						},
					},
				},
				Verifications: map[string]*authnprovidercm.VerificationResponse{},
			},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id").Maybe()
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication).Maybe()

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create context with available attributes
			ctx := EngineContext{
				Context:     context.Background(),
				ExecutionID: "test-flow-id",
				AppID:       "test-app-id",
				FlowType:    common.FlowTypeAuthentication,
				AuthenticatedUser: authncm.AuthenticatedUser{
					IsAuthenticated:     true,
					UserID:              "user-123",
					AvailableAttributes: tc.attributes,
					Attributes:          map[string]interface{}{},
				},
				UserInputs:       map[string]string{},
				RuntimeData:      map[string]string{},
				ExecutionHistory: map[string]*common.NodeExecutionRecord{},
				Graph:            mockGraph,
			}

			// Convert to DB model (serializes available attributes)
			dbModel, err := FromEngineContext(ctx)
			s.NoError(err)
			content := s.getContextContent(dbModel)
			s.NotNil(content.AvailableAttributes)

			// Convert back to EngineContext (deserializes available attributes)
			resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)
			s.NoError(err)

			// Verify original available attributes are restored
			s.NotNil(resultCtx.AuthenticatedUser.AvailableAttributes)
			s.Len(resultCtx.AuthenticatedUser.AvailableAttributes.Attributes, len(tc.attributes.Attributes))
			for attrName, attrMetadata := range tc.attributes.Attributes {
				s.Contains(resultCtx.AuthenticatedUser.AvailableAttributes.Attributes, attrName)
				expectedVerified := attrMetadata.AssuranceMetadataResponse.IsVerified
				actualVerified := resultCtx.AuthenticatedUser.AvailableAttributes.Attributes[attrName].
					AssuranceMetadataResponse.IsVerified
				s.Equal(expectedVerified, actualVerified)
			}
		})
	}
}

func (s *ModelTestSuite) TestFromEngineContext_WithCurrentSegmentID() {
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		Context:          context.Background(),
		ExecutionID:      "test-exec-id",
		FlowType:         common.FlowTypeAuthentication,
		CurrentSegmentID: "seg-1",
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	dbModel, err := FromEngineContext(ctx)
	s.NoError(err)

	content := s.getContextContent(dbModel)
	s.NotNil(content.CurrentSegmentID)
	s.Equal("seg-1", *content.CurrentSegmentID)
}

func (s *ModelTestSuite) TestFromEngineContext_EmptyCurrentSegmentID_OmitsField() {
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		Context:          context.Background(),
		ExecutionID:      "test-exec-id",
		FlowType:         common.FlowTypeAuthentication,
		CurrentSegmentID: "",
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	dbModel, err := FromEngineContext(ctx)
	s.NoError(err)

	content := s.getContextContent(dbModel)
	s.Nil(content.CurrentSegmentID)
}

func (s *ModelTestSuite) TestToEngineContext_WithCurrentSegmentID() {
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	segID := "seg-1"
	content := flowContextContent{
		GraphID:          "test-graph-id",
		CurrentSegmentID: &segID,
		UserInputs:       func() *string { v := `{}`; return &v }(),
		RuntimeData:      func() *string { v := `{}`; return &v }(),
		ExecutionHistory: func() *string { v := `{}`; return &v }(),
	}
	ctxJSON, _ := json.Marshal(content)
	dbModel := &FlowContextDB{
		ExecutionID: "test-exec-id",
		Context:     string(ctxJSON),
	}

	resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)

	s.NoError(err)
	s.Equal("seg-1", resultCtx.CurrentSegmentID)
}

func (s *ModelTestSuite) TestToEngineContext_MissingCurrentSegmentID_IsEmpty() {
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	content := flowContextContent{
		GraphID:          "test-graph-id",
		CurrentSegmentID: nil,
		UserInputs:       func() *string { v := `{}`; return &v }(),
		RuntimeData:      func() *string { v := `{}`; return &v }(),
		ExecutionHistory: func() *string { v := `{}`; return &v }(),
	}
	ctxJSON, _ := json.Marshal(content)
	dbModel := &FlowContextDB{
		ExecutionID: "test-exec-id",
		Context:     string(ctxJSON),
	}

	resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)

	s.NoError(err)
	s.Equal("", resultCtx.CurrentSegmentID)
}

func (s *ModelTestSuite) TestCurrentSegmentID_RoundTrip() {
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	ctx := EngineContext{
		Context:          context.Background(),
		ExecutionID:      "test-exec-id",
		FlowType:         common.FlowTypeAuthentication,
		CurrentSegmentID: "seg-2",
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	dbModel, err := FromEngineContext(ctx)
	s.NoError(err)

	resultCtx, err := dbModel.ToEngineContext(context.Background(), mockGraph)
	s.NoError(err)
	s.Equal("seg-2", resultCtx.CurrentSegmentID)
}
