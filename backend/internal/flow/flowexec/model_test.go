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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/config"
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
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/test/thunder/home", testConfig)
	if err != nil {
		t.Fatalf("failed to initialize Thunder runtime: %v", err)
	}

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
	s.NotEmpty(dbModel.Context)

	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
	s.Equal("test-app-id", content.AppID)
	s.True(content.Verbose)
	s.True(content.IsAuthenticated)
	s.NotNil(content.UserID)
	s.Equal("user-123", *content.UserID)

	// Verify token is encrypted (not equal to original)
	s.NotNil(content.Token)
	s.NotEqual(testToken, *content.Token)
	s.Greater(len(*content.Token), 0)
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

	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
	s.True(content.IsAuthenticated)

	// Verify token is nil when empty
	s.Nil(content.Token)
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

	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
	s.False(content.IsAuthenticated)
	s.Nil(content.UserID)
	s.Nil(content.Token)
}

func (s *ModelTestSuite) TestToEngineContext_WithToken() {
	// Setup - First create an encrypted token via FromEngineContext
	testToken := "test-token-xyz789"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

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

	// Verify token is encrypted in the context JSON
	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
	s.NotNil(content.Token)

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
	// Setup - Build a FlowContextDB directly with no token in the context JSON
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	userID := testUserID789
	content := flowContextContent{
		AppID:            "test-app-id",
		Verbose:          true,
		GraphID:          "test-graph-id",
		IsAuthenticated:  true,
		UserID:           &userID,
		UserInputs:       strPtr(`{"username":"testuser"}`),
		RuntimeData:      strPtr(`{"key":"value"}`),
		UserAttributes:   strPtr(`{"email":"test@example.com"}`),
		ExecutionHistory: strPtr(`{}`),
		Token:            nil, // No token
	}
	contextBytes, err := json.Marshal(content)
	s.NoError(err)

	dbModel := &FlowContextDB{
		FlowID:  "test-flow-id",
		Context: string(contextBytes),
	}

	// Execute
	resultCtx, err := dbModel.ToEngineContext(mockGraph)

	// Verify
	s.NoError(err)
	s.Equal("test-flow-id", resultCtx.FlowID)
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

			// Convert to DB model (encrypts token inside context JSON)
			dbModel, err := FromEngineContext(ctx)
			s.NoError(err)

			var content flowContextContent
			s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
			s.NotNil(content.Token)
			s.NotEqual(testToken, *content.Token)

			// Convert back to EngineContext (decrypts token)
			resultCtx, err := dbModel.ToEngineContext(mockGraph)
			s.NoError(err)

			// Verify original token is restored
			s.Equal(testToken, resultCtx.AuthenticatedUser.Token)
		})
	}
}

func (s *ModelTestSuite) TestToEngineContext_WithInvalidEncryptedToken() {
	// Setup - Build a FlowContextDB with an invalid encrypted token in the context JSON
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	invalidToken := "invalid-encrypted-data" //nolint:gosec // G101: This is test data, not a real credential
	content := flowContextContent{
		AppID:            "test-app-id",
		GraphID:          "test-graph-id",
		IsAuthenticated:  true,
		UserInputs:       strPtr(`{}`),
		RuntimeData:      strPtr(`{}`),
		UserAttributes:   strPtr(`{}`),
		ExecutionHistory: strPtr(`{}`),
		Token:            &invalidToken,
	}
	contextBytes, err := json.Marshal(content)
	s.NoError(err)

	dbModel := &FlowContextDB{
		FlowID:  "test-flow-id",
		Context: string(contextBytes),
	}

	// Execute
	_, err = dbModel.ToEngineContext(mockGraph)

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

	// Verify
	s.NoError(err)
	s.Equal("flow-123", dbModel.FlowID)

	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
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
	testAvailableAttributes := &authnprovider.AvailableAttributes{
		Attributes: map[string]*authnprovider.AttributeMetadataResponse{
			"email": {
				AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
					IsVerified: true,
				},
			},
			"phoneNumber": {
				AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
					IsVerified: false,
				},
			},
		},
		Verifications: map[string]*authnprovider.VerificationResponse{},
	}
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
	s.Equal("test-flow-id", dbModel.FlowID)

	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
	s.Equal("test-app-id", content.AppID)
	s.True(content.Verbose)
	s.True(content.IsAuthenticated)
	s.NotNil(content.UserID)
	s.Equal("user-123", *content.UserID)

	// Verify available attributes are serialized (not encrypted)
	s.NotNil(content.AvailableAttributes)
	s.Greater(len(*content.AvailableAttributes), 0)
	s.Contains(*content.AvailableAttributes, "\"email\"")
	s.Contains(*content.AvailableAttributes, "\"phoneNumber\"")
}

func (s *ModelTestSuite) TestFromEngineContext_WithoutAvailableAttributes() {
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
	s.Equal("test-flow-id", dbModel.FlowID)

	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
	s.True(content.IsAuthenticated)

	// Verify available attributes is nil when empty
	s.Nil(content.AvailableAttributes)
}

func (s *ModelTestSuite) TestToEngineContext_WithAvailableAttributes() {
	// Setup
	testAvailableAttributes := &authnprovider.AvailableAttributes{
		Attributes: map[string]*authnprovider.AttributeMetadataResponse{
			"email": {
				AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
					IsVerified: true,
				},
			},
			"address": {
				AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
					IsVerified: false,
				},
			},
		},
		Verifications: map[string]*authnprovider.VerificationResponse{},
	}
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		FlowType: common.FlowTypeAuthentication,
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

	// Execute - Convert back to EngineContext
	resultCtx, err := dbModel.ToEngineContext(mockGraph)

	// Verify
	s.NoError(err)
	s.Equal("test-flow-id", resultCtx.FlowID)
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

	userID := "user-987"
	content := flowContextContent{
		AppID:               "test-app-id",
		Verbose:             true,
		GraphID:             "test-graph-id",
		IsAuthenticated:     true,
		UserID:              &userID,
		UserInputs:          strPtr(`{"username":"testuser"}`),
		RuntimeData:         strPtr(`{"key":"value"}`),
		UserAttributes:      strPtr(`{"email":"test@example.com"}`),
		ExecutionHistory:    strPtr(`{}`),
		AvailableAttributes: nil, // No available attributes
	}
	contextBytes, err := json.Marshal(content)
	s.NoError(err)

	dbModel := &FlowContextDB{
		FlowID:  "test-flow-id",
		Context: string(contextBytes),
	}

	// Execute
	resultCtx, err := dbModel.ToEngineContext(mockGraph)

	// Verify
	s.NoError(err)
	s.Equal("test-flow-id", resultCtx.FlowID)
	s.True(resultCtx.AuthenticatedUser.IsAuthenticated)
	s.Equal("user-987", resultCtx.AuthenticatedUser.UserID)

	// Verify available attributes is nil/empty when not provided
	s.Nil(resultCtx.AuthenticatedUser.AvailableAttributes)
}

func (s *ModelTestSuite) TestAvailableAttributesSerializationRoundTrip() {
	// Setup
	testCases := []struct {
		name       string
		attributes *authnprovider.AvailableAttributes
	}{
		{
			name: "Single attribute",
			attributes: &authnprovider.AvailableAttributes{
				Attributes: map[string]*authnprovider.AttributeMetadataResponse{
					"email": {
						AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
							IsVerified: true,
						},
					},
				},
				Verifications: map[string]*authnprovider.VerificationResponse{},
			},
		},
		{
			name: "Multiple attributes",
			attributes: &authnprovider.AvailableAttributes{
				Attributes: map[string]*authnprovider.AttributeMetadataResponse{
					"email": {
						AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
							IsVerified: true,
						},
					},
					"phone": {
						AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
							IsVerified: false,
						},
					},
					"address": {
						AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
							IsVerified: true,
						},
					},
				},
				Verifications: map[string]*authnprovider.VerificationResponse{},
			},
		},
		{
			name: "Special characters in names",
			attributes: &authnprovider.AvailableAttributes{
				Attributes: map[string]*authnprovider.AttributeMetadataResponse{
					"custom-attr-1": {
						AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
							IsVerified: true,
						},
					},
					"attr_with_underscore": {
						AssuranceMetadataResponse: &authnprovider.AssuranceMetadataResponse{
							IsVerified: false,
						},
					},
				},
				Verifications: map[string]*authnprovider.VerificationResponse{},
			},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id").Maybe()
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication).Maybe()

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := EngineContext{
				FlowID:   "test-flow-id",
				AppID:    "test-app-id",
				FlowType: common.FlowTypeAuthentication,
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

			// Convert to DB model (serializes available attributes into context JSON)
			dbModel, err := FromEngineContext(ctx)
			s.NoError(err)

			var content flowContextContent
			s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
			s.NotNil(content.AvailableAttributes)

			// Convert back to EngineContext (deserializes available attributes)
			resultCtx, err := dbModel.ToEngineContext(mockGraph)
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
