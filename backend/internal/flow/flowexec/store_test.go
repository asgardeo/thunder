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
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/config"

	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
)

type StoreTestSuite struct {
	suite.Suite
}

func TestStoreTestSuite(t *testing.T) {
	// Setup test config with encryption key
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Encryption: config.EncryptionConfig{
				Key: "2729a7928c79371e5f312167269294a14bb0660fd166b02a408a20fa73271580",
			},
		},
		Server: config.ServerConfig{
			Identifier: "test-deployment",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/test/thunder/home", testConfig)
	if err != nil {
		t.Fatalf("Failed to initialize Thunder runtime: %v", err)
	}

	suite.Run(t, new(StoreTestSuite))
}

func (s *StoreTestSuite) getContextContent(dbModel *FlowContextDB) flowContextContent {
	var content flowContextContent
	err := json.Unmarshal([]byte(dbModel.Context), &content)
	s.NoError(err)
	return content
}

func (s *StoreTestSuite) TestStoreFlowContext_WithToken() {
	// Setup
	testToken := "test-auth-token-12345" //nolint:gosec // G101: This is test data, not a real credential
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)

	// Expect one ExecuteContext call for FLOW_CONTEXT
	// Use mock.Anything for pointer parameters since they're created inside FromEngineContext
	mockDBClient.EXPECT().ExecuteContext(mock.Anything, QueryCreateFlowContext, "test-flow-id", "test-deployment",
		mock.Anything, mock.Anything).Return(int64(0), nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	expirySeconds := int64(1800) // 30 minutes
	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		Verbose:     false,
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

	// Execute
	err := store.StoreFlowContext(context.Background(), ctx, expirySeconds)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestStoreFlowContext_WithoutToken() {
	// Setup
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)

	expirySeconds := int64(1800) // 30 minutes

	mockDBClient.EXPECT().ExecuteContext(mock.Anything, QueryCreateFlowContext, "test-flow-id", "test-deployment",
		mock.Anything, mock.Anything).Return(int64(0), nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		Verbose:     false,
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: false,
			Token:           "", // No token
			Attributes:      map[string]interface{}{},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	// Execute
	err := store.StoreFlowContext(context.Background(), ctx, expirySeconds)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestUpdateFlowContext_WithToken() {
	// Setup
	testToken := "updated-token-xyz"
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)

	mockDBClient.EXPECT().ExecuteContext(mock.Anything, QueryUpdateFlowContext,
		"test-flow-id", mock.Anything, "test-deployment").Return(int64(0), nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-456",
			Token:           testToken,
			Attributes:      map[string]interface{}{},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	// Execute
	err := store.UpdateFlowContext(context.Background(), ctx)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestGetFlowContext_WithToken() {
	// Setup - First encrypt a token to use as test data
	testToken := "retrieved-token-abc"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	expiryTime := time.Now().Add(30 * time.Minute)

	// Create encrypted token
	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-789",
			Token:           testToken,
			Attributes:      map[string]interface{}{},
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

	// Setup mocks
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())

	results := []map[string]interface{}{
		{
			"flow_id":     "test-flow-id",
			"context":     dbModel.Context,
			"expiry_time": expiryTime,
		},
	}

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("QueryContext", mock.Anything, QueryGetFlowContext,
		"test-flow-id", "test-deployment", mock.Anything).Return(results, nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	result, err := store.GetFlowContext(context.Background(), "test-flow-id")

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.ExecutionID)

	content = s.getContextContent(result)
	s.True(content.IsAuthenticated)
	s.NotNil(content.Token)
	s.Equal(dbModel.Context, result.Context) // JSON should match

	// Verify we can decrypt it back to original
	restoredCtx, err := result.ToEngineContext(mockGraph)
	s.NoError(err)
	s.Equal(testToken, restoredCtx.AuthenticatedUser.Token)

	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestGetFlowContext_WithoutToken() {
	// Setup
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())

	expiryTime := time.Now().Add(30 * time.Minute)

	// Create serialized context
	contextJSON, err := json.Marshal(flowContextContent{
		AppID:           "test-app-id",
		IsAuthenticated: false,
		GraphID:         "test-graph-id",
	})
	s.NoError(err)

	results := []map[string]interface{}{
		{
			"flow_id":     "test-flow-id",
			"context":     string(contextJSON),
			"expiry_time": expiryTime,
		},
	}

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("QueryContext", mock.Anything, QueryGetFlowContext,
		"test-flow-id", "test-deployment", mock.Anything).Return(results, nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	result, err := store.GetFlowContext(context.Background(), "test-flow-id")

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.ExecutionID)

	content := s.getContextContent(result)
	s.False(content.IsAuthenticated)
	s.Nil(content.Token)

	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestStoreAndRetrieve_TokenRoundTrip() {
	// This is an integration-style test that simulates the full round trip
	// of storing and retrieving a flow context with an encrypted token

	// Setup
	originalToken := "integration-test-token-secret"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("integration-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	originalCtx := EngineContext{
		ExecutionID: "integration-flow-id",
		AppID:       "integration-app-id",
		Verbose:     true,
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "integration-user-123",
			OUID:            "integration-org-456",
			UserType:        "premium",
			Token:           originalToken,
			Attributes: map[string]interface{}{
				"email": "integration@test.com",
				"role":  "admin",
			},
		},
		UserInputs: map[string]string{
			"username": "testuser",
			"password": "secret",
		},
		RuntimeData: map[string]string{
			"state": "abc123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{
			"node-1": {NodeID: "node-1"},
		},
		Graph: mockGraph,
	}

	// Step 1: Convert to DB model (encrypts token)
	dbModel, err := FromEngineContext(originalCtx)
	s.NoError(err)
	s.NotNil(dbModel)

	content := s.getContextContent(dbModel)
	s.NotNil(content.Token)
	s.NotEqual(originalToken, *content.Token, "Token should be encrypted")

	// Step 2: Simulate storing and retrieving from DB
	// In a real scenario, this would be inserted into DB and read back
	// For this test, we'll directly use the dbModel

	// Step 3: Convert back to EngineContext (decrypts token)
	retrievedCtx, err := dbModel.ToEngineContext(mockGraph)
	s.NoError(err)

	// Step 4: Verify all data is preserved correctly
	s.Equal(originalCtx.ExecutionID, retrievedCtx.ExecutionID)
	s.Equal(originalCtx.AppID, retrievedCtx.AppID)
	s.Equal(originalCtx.Verbose, retrievedCtx.Verbose)
	s.Equal(originalCtx.AuthenticatedUser.IsAuthenticated, retrievedCtx.AuthenticatedUser.IsAuthenticated)
	s.Equal(originalCtx.AuthenticatedUser.UserID, retrievedCtx.AuthenticatedUser.UserID)
	s.Equal(originalCtx.AuthenticatedUser.OUID, retrievedCtx.AuthenticatedUser.OUID)
	s.Equal(originalCtx.AuthenticatedUser.UserType, retrievedCtx.AuthenticatedUser.UserType)

	// Most importantly, verify the token was decrypted correctly
	s.Equal(originalToken, retrievedCtx.AuthenticatedUser.Token, "Token should be decrypted to original value")

	// Verify other fields
	s.Equal(len(originalCtx.UserInputs), len(retrievedCtx.UserInputs))
	s.Equal(len(originalCtx.RuntimeData), len(retrievedCtx.RuntimeData))
	s.Equal(len(originalCtx.ExecutionHistory), len(retrievedCtx.ExecutionHistory))
}

func (s *StoreTestSuite) TestBuildFlowContextFromResultRow_WithToken() {
	// Setup - First create an encrypted token
	testToken := "parse-test-token"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	expiryTime := time.Now().Add(30 * time.Minute)

	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			Token:      testToken,
			Attributes: map[string]interface{}{},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	dbModel, err := FromEngineContext(ctx)
	s.NoError(err)

	store := &flowStore{deploymentID: "test-deployment"}

	userID := "user-123"
	_ = userID
	row := map[string]interface{}{
		"flow_id":     "test-flow-id",
		"context":     dbModel.Context,
		"expiry_time": expiryTime,
	}

	// Execute
	result, err := store.buildFlowContextFromResultRow(row)

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal(dbModel.Context, result.Context)
}

func (s *StoreTestSuite) TestBuildFlowContextFromResultRow_WithByteToken() {
	// Test handling when database returns token as []byte (common with PostgreSQL)
	// Setup
	testToken := "byte-token-test" //nolint:gosec // G101: This is test data, not a real credential
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	expiryTime := time.Now().Add(30 * time.Minute)

	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			Token:      testToken,
			Attributes: map[string]interface{}{},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	dbModel, err := FromEngineContext(ctx)
	s.NoError(err)

	store := &flowStore{deploymentID: "test-deployment"}

	row := map[string]interface{}{
		"flow_id":     "test-flow-id",
		"context":     dbModel.Context,
		"expiry_time": expiryTime,
	}

	// Execute
	result, err := store.buildFlowContextFromResultRow(row)

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal(dbModel.Context, result.Context)
}

func (s *StoreTestSuite) TestStoreFlowContext_WithAvailableAttributes() {
	// Setup
	testAvailableAttributes := &authnprovider.AvailableAttributes{
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
		},
		Verifications: map[string]*authnprovider.VerificationResponse{},
	}
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)

	// Expect one ExecuteContext call for FLOW_CONTEXT
	mockDBClient.EXPECT().ExecuteContext(mock.Anything, QueryCreateFlowContext, "test-flow-id", "test-deployment",
		mock.Anything, mock.Anything).Return(int64(0), nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	expirySeconds := int64(1800) // 30 minutes
	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		Verbose:     false,
		FlowType:    common.FlowTypeAuthentication,
		RuntimeData: map[string]string{"key": "value"},
		UserInputs:  map[string]string{"input1": "val1"},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:     true,
			UserID:              "test-user",
			AvailableAttributes: testAvailableAttributes,
		},
		Graph: mockGraph,
	}

	// Execute
	err := store.StoreFlowContext(context.Background(), ctx, expirySeconds)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestUpdateFlowContext_WithAvailableAttributes() {
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
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)

	mockDBClient.EXPECT().ExecuteContext(mock.Anything, QueryUpdateFlowContext,
		"test-flow-id", mock.Anything, "test-deployment").Return(int64(0), nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:     true,
			UserID:              "user-456",
			AvailableAttributes: testAvailableAttributes,
			Attributes:          map[string]interface{}{},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	// Execute
	err := store.UpdateFlowContext(context.Background(), ctx)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestGetFlowContext_WithAvailableAttributes() {
	// Setup
	testAvailableAttributes := &authnprovider.AvailableAttributes{
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
		},
		Verifications: map[string]*authnprovider.VerificationResponse{},
	}
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	expiryTime := time.Now().Add(30 * time.Minute)

	// Create serialized available attributes
	ctx := EngineContext{
		ExecutionID: "test-flow-id",
		AppID:       "test-app-id",
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:     true,
			UserID:              "user-789",
			AvailableAttributes: testAvailableAttributes,
			Attributes:          map[string]interface{}{},
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

	// Setup mocks
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())

	results := []map[string]interface{}{
		{
			"flow_id":     "test-flow-id",
			"context":     dbModel.Context,
			"expiry_time": expiryTime,
		},
	}

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("QueryContext", mock.Anything, QueryGetFlowContext,
		"test-flow-id", "test-deployment", mock.Anything).Return(results, nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	result, err := store.GetFlowContext(context.Background(), "test-flow-id")

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.ExecutionID)

	content = s.getContextContent(result)
	s.True(content.IsAuthenticated)
	s.NotNil(content.AvailableAttributes)
	s.Equal(dbModel.Context, result.Context) // Serialized context should match

	// Verify we can deserialize it back to original
	restoredCtx, err := result.ToEngineContext(mockGraph)
	s.NoError(err)
	s.NotNil(restoredCtx.AuthenticatedUser.AvailableAttributes)
	s.Len(restoredCtx.AuthenticatedUser.AvailableAttributes.Attributes, 2)
	s.Contains(restoredCtx.AuthenticatedUser.AvailableAttributes.Attributes, "email")
	s.Contains(restoredCtx.AuthenticatedUser.AvailableAttributes.Attributes, "phone")
	s.True(restoredCtx.AuthenticatedUser.AvailableAttributes.Attributes["email"].AssuranceMetadataResponse.IsVerified)
	s.False(restoredCtx.AuthenticatedUser.AvailableAttributes.Attributes["phone"].AssuranceMetadataResponse.IsVerified)

	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}
