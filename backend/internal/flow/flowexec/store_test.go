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

// utcTimeMatcher matches any time.Time value whose location is UTC.
var utcTimeMatcher = mock.MatchedBy(func(t time.Time) bool {
	return t.Location() == time.UTC
})

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

func (s *StoreTestSuite) TestStoreFlowContext_WithToken() {
	// Setup
	testToken := "test-auth-token-12345" //nolint:gosec // G101: This is test data, not a real credential
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)

	// Expect a single ExecuteContext call for FLOW_CONTEXT with the context JSON
	mockDBClient.EXPECT().ExecuteContext(mock.Anything, QueryCreateFlowContext,
		"test-flow-id", "test-deployment", mock.Anything, utcTimeMatcher).Return(int64(0), nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	expirySeconds := int64(1800) // 30 minutes
	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		Verbose:  false,
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

	mockDBClient.EXPECT().ExecuteContext(mock.Anything, QueryCreateFlowContext,
		"test-flow-id", "test-deployment", mock.Anything, utcTimeMatcher).Return(int64(0), nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		Verbose:  false,
		FlowType: common.FlowTypeAuthentication,
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
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		FlowType: common.FlowTypeAuthentication,
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

	// Create a FlowContextDB with the context JSON containing an encrypted token
	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		FlowType: common.FlowTypeAuthentication,
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

	// Verify token is encrypted inside the context JSON
	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
	s.NotNil(content.Token)
	s.NotEqual(testToken, *content.Token)

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
		"test-flow-id", "test-deployment", utcTimeMatcher).Return(results, nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	result, err := store.GetFlowContext(context.Background(), "test-flow-id")

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.FlowID)
	s.Equal(dbModel.Context, result.Context)

	// Verify we can decrypt the token back to original
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

	// Build a context JSON without a token
	content := flowContextContent{
		AppID:            "test-app-id",
		GraphID:          "test-graph-id",
		IsAuthenticated:  false,
		RuntimeData:      strPtr("{}"),
		ExecutionHistory: strPtr("{}"),
		UserInputs:       strPtr("{}"),
		UserAttributes:   strPtr("{}"),
	}
	contextBytes, err := json.Marshal(content)
	s.NoError(err)

	results := []map[string]interface{}{
		{
			"flow_id":     "test-flow-id",
			"context":     string(contextBytes),
			"expiry_time": expiryTime,
		},
	}

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("QueryContext", mock.Anything, QueryGetFlowContext,
		"test-flow-id", "test-deployment", utcTimeMatcher).Return(results, nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	result, err := store.GetFlowContext(context.Background(), "test-flow-id")

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.FlowID)

	// Verify no token in context
	var parsedContent flowContextContent
	s.NoError(json.Unmarshal([]byte(result.Context), &parsedContent))
	s.Nil(parsedContent.Token)
	s.False(parsedContent.IsAuthenticated)

	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestDeleteFlowContext() {
	// Setup
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)

	mockDBClient.EXPECT().ExecuteContext(mock.Anything, QueryDeleteFlowContext,
		"test-flow-id", "test-deployment").Return(int64(1), nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	err := store.DeleteFlowContext(context.Background(), "test-flow-id")

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestStoreAndRetrieve_TokenRoundTrip() {
	originalToken := "integration-test-token-secret"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("integration-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	originalCtx := EngineContext{
		FlowID:   "integration-flow-id",
		AppID:    "integration-app-id",
		Verbose:  true,
		FlowType: common.FlowTypeAuthentication,
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

	// Step 1: Convert to DB model (encrypts token inside context JSON)
	dbModel, err := FromEngineContext(originalCtx)
	s.NoError(err)
	s.NotNil(dbModel)
	s.Equal("integration-flow-id", dbModel.FlowID)

	// Verify token is encrypted inside the JSON
	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
	s.NotNil(content.Token)
	s.NotEqual(originalToken, *content.Token, "Token should be encrypted")

	// Step 2: Convert back to EngineContext (decrypts token)
	retrievedCtx, err := dbModel.ToEngineContext(mockGraph)
	s.NoError(err)

	// Step 3: Verify all data is preserved correctly
	s.Equal(originalCtx.FlowID, retrievedCtx.FlowID)
	s.Equal(originalCtx.AppID, retrievedCtx.AppID)
	s.Equal(originalCtx.Verbose, retrievedCtx.Verbose)
	s.Equal(originalCtx.AuthenticatedUser.IsAuthenticated, retrievedCtx.AuthenticatedUser.IsAuthenticated)
	s.Equal(originalCtx.AuthenticatedUser.UserID, retrievedCtx.AuthenticatedUser.UserID)
	s.Equal(originalCtx.AuthenticatedUser.OUID, retrievedCtx.AuthenticatedUser.OUID)
	s.Equal(originalCtx.AuthenticatedUser.UserType, retrievedCtx.AuthenticatedUser.UserType)
	s.Equal(originalToken, retrievedCtx.AuthenticatedUser.Token, "Token should be decrypted to original value")
	s.Equal(len(originalCtx.UserInputs), len(retrievedCtx.UserInputs))
	s.Equal(len(originalCtx.RuntimeData), len(retrievedCtx.RuntimeData))
	s.Equal(len(originalCtx.ExecutionHistory), len(retrievedCtx.ExecutionHistory))
}

func (s *StoreTestSuite) TestBuildFlowContextFromResultRow_WithContextJSON() {
	expiryTime := time.Now().Add(30 * time.Minute)

	content := flowContextContent{
		AppID:            "test-app-id",
		GraphID:          "test-graph-id",
		IsAuthenticated:  true,
		RuntimeData:      strPtr("{}"),
		ExecutionHistory: strPtr("{}"),
		UserInputs:       strPtr("{}"),
		UserAttributes:   strPtr("{}"),
	}
	contextBytes, err := json.Marshal(content)
	s.NoError(err)

	store := &flowStore{deploymentID: "test-deployment"}

	row := map[string]interface{}{
		"flow_id":     "test-flow-id",
		"context":     string(contextBytes),
		"expiry_time": expiryTime,
	}

	// Execute
	result, err := store.buildFlowContextFromResultRow(row)

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.FlowID)
	s.NotEmpty(result.Context)

	var parsedContent flowContextContent
	s.NoError(json.Unmarshal([]byte(result.Context), &parsedContent))
	s.Equal("test-app-id", parsedContent.AppID)
	s.Equal("test-graph-id", parsedContent.GraphID)
	s.True(parsedContent.IsAuthenticated)
}

func (s *StoreTestSuite) TestBuildFlowContextFromResultRow_WithByteContext() {
	// Test handling when database returns context as []byte (common with PostgreSQL)
	expiryTime := time.Now().Add(30 * time.Minute)

	content := flowContextContent{
		AppID:            "test-app-id",
		GraphID:          "test-graph-id",
		IsAuthenticated:  false,
		RuntimeData:      strPtr("{}"),
		ExecutionHistory: strPtr("{}"),
		UserInputs:       strPtr("{}"),
		UserAttributes:   strPtr("{}"),
	}
	contextBytes, err := json.Marshal(content)
	s.NoError(err)

	store := &flowStore{deploymentID: "test-deployment"}

	row := map[string]interface{}{
		"flow_id":     "test-flow-id",
		"context":     contextBytes, // []byte as returned by PostgreSQL
		"expiry_time": expiryTime,
	}

	// Execute
	result, err := store.buildFlowContextFromResultRow(row)

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.FlowID)

	var parsedContent flowContextContent
	s.NoError(json.Unmarshal([]byte(result.Context), &parsedContent))
	s.Equal("test-app-id", parsedContent.AppID)
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

	mockDBClient.EXPECT().ExecuteContext(mock.Anything, QueryCreateFlowContext,
		"test-flow-id", "test-deployment", mock.Anything, utcTimeMatcher).Return(int64(0), nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	expirySeconds := int64(1800) // 30 minutes
	ctx := EngineContext{
		FlowID:      "test-flow-id",
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
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		FlowType: common.FlowTypeAuthentication,
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

	// Create context and serialize to DB model to get the JSON with available attributes
	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		FlowType: common.FlowTypeAuthentication,
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

	// Verify available attributes are in the context JSON
	var content flowContextContent
	s.NoError(json.Unmarshal([]byte(dbModel.Context), &content))
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
		"test-flow-id", "test-deployment", utcTimeMatcher).Return(results, nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	result, err := store.GetFlowContext(context.Background(), "test-flow-id")

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.FlowID)
	s.Equal(dbModel.Context, result.Context)

	// Verify we can deserialize available attributes back to original
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

// strPtr is a helper to get a pointer to a string literal.
func strPtr(s string) *string {
	return &s
}
