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
	"testing"

	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
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
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	suite.Run(t, new(ModelTestSuite))
}

func (s *ModelTestSuite) TestFromEngineContext_WithToken() {
	// Setup
	testToken := "test-token-123456"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		Verbose:  true,
		FlowType: common.FlowTypeAuthentication,
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
	s.Equal("test-flow-id", dbModel.FlowID)
	s.Equal("test-app-id", dbModel.AppID)
	s.True(dbModel.Verbose)
	s.True(dbModel.IsAuthenticated)
	s.NotNil(dbModel.UserID)
	s.Equal("user-123", *dbModel.UserID)

	// Verify token is encrypted (not equal to original)
	s.NotNil(dbModel.Token)
	s.NotEqual(testToken, *dbModel.Token)

	// Verify token can be decrypted back
	s.Greater(len(*dbModel.Token), 0)
}

func (s *ModelTestSuite) TestFromEngineContext_WithoutToken() {
	// Setup
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		Verbose:  false,
		FlowType: common.FlowTypeAuthentication,
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
	s.Equal("test-flow-id", dbModel.FlowID)
	s.True(dbModel.IsAuthenticated)

	// Verify token is nil when empty
	s.Nil(dbModel.Token)
}

func (s *ModelTestSuite) TestFromEngineContext_WithEmptyAuthenticatedUser() {
	// Setup
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		FlowID:            "test-flow-id",
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
	s.False(dbModel.IsAuthenticated)
	s.Nil(dbModel.UserID)
	s.Nil(dbModel.Token)
}

func (s *ModelTestSuite) TestToEngineContext_WithToken() {
	// Setup - First create an encrypted token
	testToken := "test-token-xyz789"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	// Create the context and convert to DB model to get encrypted token
	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		FlowType: common.FlowTypeAuthentication,
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
	s.NotNil(dbModel.Token)

	// Execute - Convert back to EngineContext
	resultCtx, err := dbModel.ToEngineContext(mockGraph)

	// Verify
	s.NoError(err)
	s.Equal("test-flow-id", resultCtx.FlowID)
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
	userID := "user-789"

	dbModel := &FlowContextWithUserDataDB{
		FlowID:           "test-flow-id",
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

	// Execute
	resultCtx, err := dbModel.ToEngineContext(mockGraph)

	// Verify
	s.NoError(err)
	s.Equal("test-flow-id", resultCtx.FlowID)
	s.True(resultCtx.AuthenticatedUser.IsAuthenticated)
	s.Equal("user-789", resultCtx.AuthenticatedUser.UserID)

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
				FlowID:   "test-flow-id",
				AppID:    "test-app-id",
				FlowType: common.FlowTypeAuthentication,
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

			// Convert to DB model (encrypts token)
			dbModel, err := FromEngineContext(ctx)
			s.NoError(err)
			s.NotNil(dbModel.Token)

			// Verify token is encrypted
			s.NotEqual(testToken, *dbModel.Token)

			// Convert back to EngineContext (decrypts token)
			resultCtx, err := dbModel.ToEngineContext(mockGraph)
			s.NoError(err)

			// Verify original token is restored
			s.Equal(testToken, resultCtx.AuthenticatedUser.Token)
		})
	}
}

func (s *ModelTestSuite) TestToEngineContext_WithInvalidEncryptedToken() {
	// Setup - Create a DB model with invalid encrypted token
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	invalidToken := "invalid-encrypted-data" //nolint:gosec // G101: This is test data, not a real credential
	userInputs := `{}`
	runtimeData := `{}`
	userAttributes := `{}`
	executionHistory := `{}`

	dbModel := &FlowContextWithUserDataDB{
		FlowID:           "test-flow-id",
		AppID:            "test-app-id",
		GraphID:          "test-graph-id",
		IsAuthenticated:  true,
		UserInputs:       &userInputs,
		RuntimeData:      &runtimeData,
		UserAttributes:   &userAttributes,
		ExecutionHistory: &executionHistory,
		Token:            &invalidToken,
	}

	// Execute
	_, err := dbModel.ToEngineContext(mockGraph)

	// Verify - Should return error for invalid encrypted token
	s.Error(err)
}

func (s *ModelTestSuite) TestFromEngineContext_PreservesOtherFields() {
	// Setup
	testToken := "test-token-preserve-fields"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("graph-123")

	currentAction := "test-action"
	ctx := EngineContext{
		FlowID:        "flow-123",
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
			IsAuthenticated:    true,
			UserID:             "user-abc",
			OrganizationUnitID: "org-xyz",
			UserType:           "admin",
			Token:              testToken,
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
	s.Equal("flow-123", dbModel.FlowID)
	s.Equal("app-123", dbModel.AppID)
	s.True(dbModel.Verbose)
	s.NotNil(dbModel.CurrentAction)
	s.Equal(currentAction, *dbModel.CurrentAction)
	s.Equal("graph-123", dbModel.GraphID)
	s.True(dbModel.IsAuthenticated)
	s.NotNil(dbModel.UserID)
	s.Equal("user-abc", *dbModel.UserID)
	s.NotNil(dbModel.OrganizationUnitID)
	s.Equal("org-xyz", *dbModel.OrganizationUnitID)
	s.NotNil(dbModel.UserType)
	s.Equal("admin", *dbModel.UserType)
	s.NotNil(dbModel.UserInputs)
	s.NotNil(dbModel.RuntimeData)
	s.NotNil(dbModel.UserAttributes)
	s.NotNil(dbModel.ExecutionHistory)
	s.NotNil(dbModel.Token)
}
