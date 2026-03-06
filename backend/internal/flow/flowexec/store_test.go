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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/database/modelmock"
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

func (s *StoreTestSuite) TestStoreFlowContext_WithToken() {
	// Setup
	testToken := "test-auth-token-12345" //nolint:gosec // G101: This is test data, not a real credential
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())
	mockTx := modelmock.NewTxInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("BeginTx").Return(mockTx, nil)

	// Expect two Exec calls: one for FLOW_CONTEXT, one for FLOW_USER_DATA
	// Use mock.Anything for pointer parameters since they're created inside FromEngineContext
	mockTx.On("Exec", QueryCreateFlowContext, "test-flow-id", "test-app-id", false,
		mock.Anything, mock.Anything, "test-graph-id", mock.Anything, mock.Anything, "test-deployment").Return(nil, nil)

	// Token encryption/decryption is tested in model_test.go, so we just use mock.Anything here
	mockTx.On("Exec", QueryCreateFlowUserData, "test-flow-id", true, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "test-deployment").
		Return(nil, nil)

	mockTx.On("Commit").Return(nil)

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
	err := store.StoreFlowContext(ctx)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
	mockTx.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestStoreFlowContext_WithoutToken() {
	// Setup
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())
	mockTx := modelmock.NewTxInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("BeginTx").Return(mockTx, nil)

	mockTx.On("Exec", QueryCreateFlowContext, "test-flow-id", "test-app-id", false,
		mock.Anything, mock.Anything, "test-graph-id", mock.Anything, mock.Anything, "test-deployment").Return(nil, nil)

	// Token should be nil when not provided
	mockTx.On("Exec", QueryCreateFlowUserData, "test-flow-id", false, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "test-deployment").
		Return(nil, nil)

	mockTx.On("Commit").Return(nil)

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
	err := store.StoreFlowContext(ctx)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
	mockTx.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestUpdateFlowContext_WithToken() {
	// Setup
	testToken := "updated-token-xyz"
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())
	mockTx := modelmock.NewTxInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("BeginTx").Return(mockTx, nil)

	mockTx.On("Exec", QueryUpdateFlowContext, "test-flow-id", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, "test-deployment").Return(nil, nil)

	// Token encryption/decryption is tested in model_test.go, so we just use mock.Anything here
	mockTx.On("Exec", QueryUpdateFlowUserData, "test-flow-id", true, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "test-deployment").
		Return(nil, nil)

	mockTx.On("Commit").Return(nil)

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
	err := store.UpdateFlowContext(ctx)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
	mockTx.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestGetFlowContext_WithToken() {
	// Setup - First encrypt a token to use as test data
	testToken := "retrieved-token-abc"
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")
	mockGraph.On("GetType").Return(common.FlowTypeAuthentication)

	// Create encrypted token
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
	s.NotNil(dbModel.Token)

	// Setup mocks
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())

	userID := "user-789"
	results := []map[string]interface{}{
		{
			"flow_id":           "test-flow-id",
			"app_id":            "test-app-id",
			"verbose":           false,
			"current_node_id":   nil,
			"current_action":    nil,
			"graph_id":          "test-graph-id",
			"runtime_data":      "{}",
			"execution_history": "{}",
			"is_authenticated":  true,
			"user_id":           userID,
			"ou_id":             nil,
			"user_type":         nil,
			"user_inputs":       "{}",
			"user_attributes":   "{}",
			"token":             *dbModel.Token, // Use the encrypted token
		},
	}

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("Query", QueryGetFlowContextWithUserData, "test-flow-id", "test-deployment").Return(results, nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	result, err := store.GetFlowContext("test-flow-id")

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.FlowID)
	s.True(result.IsAuthenticated)
	s.NotNil(result.Token)
	s.Equal(*dbModel.Token, *result.Token) // Encrypted token should match

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

	results := []map[string]interface{}{
		{
			"flow_id":           "test-flow-id",
			"app_id":            "test-app-id",
			"verbose":           false,
			"current_node_id":   nil,
			"current_action":    nil,
			"graph_id":          "test-graph-id",
			"runtime_data":      "{}",
			"execution_history": "{}",
			"is_authenticated":  false,
			"user_id":           nil,
			"ou_id":             nil,
			"user_type":         nil,
			"user_inputs":       "{}",
			"user_attributes":   "{}",
			"token":             nil, // No token
		},
	}

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("Query", QueryGetFlowContextWithUserData, "test-flow-id", "test-deployment").Return(results, nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	result, err := store.GetFlowContext("test-flow-id")

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.FlowID)
	s.False(result.IsAuthenticated)
	s.Nil(result.Token)

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
		FlowID:   "integration-flow-id",
		AppID:    "integration-app-id",
		Verbose:  true,
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:    true,
			UserID:             "integration-user-123",
			OrganizationUnitID: "integration-org-456",
			UserType:           "premium",
			Token:              originalToken,
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
	s.NotNil(dbModel.Token)
	s.NotEqual(originalToken, *dbModel.Token, "Token should be encrypted")

	// Step 2: Simulate storing and retrieving from DB
	// In a real scenario, this would be inserted into DB and read back
	// For this test, we'll directly use the dbModel

	// Step 3: Convert back to EngineContext (decrypts token)
	retrievedCtx, err := dbModel.ToEngineContext(mockGraph)
	s.NoError(err)

	// Step 4: Verify all data is preserved correctly
	s.Equal(originalCtx.FlowID, retrievedCtx.FlowID)
	s.Equal(originalCtx.AppID, retrievedCtx.AppID)
	s.Equal(originalCtx.Verbose, retrievedCtx.Verbose)
	s.Equal(originalCtx.AuthenticatedUser.IsAuthenticated, retrievedCtx.AuthenticatedUser.IsAuthenticated)
	s.Equal(originalCtx.AuthenticatedUser.UserID, retrievedCtx.AuthenticatedUser.UserID)
	s.Equal(originalCtx.AuthenticatedUser.OrganizationUnitID, retrievedCtx.AuthenticatedUser.OrganizationUnitID)
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

	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		FlowType: common.FlowTypeAuthentication,
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
	row := map[string]interface{}{
		"flow_id":           "test-flow-id",
		"app_id":            "test-app-id",
		"verbose":           false,
		"current_node_id":   nil,
		"current_action":    nil,
		"graph_id":          "test-graph-id",
		"runtime_data":      "{}",
		"execution_history": "{}",
		"is_authenticated":  true,
		"user_id":           userID,
		"ou_id":             nil,
		"user_type":         nil,
		"user_inputs":       "{}",
		"user_attributes":   "{}",
		"token":             *dbModel.Token,
	}

	// Execute
	result, err := store.buildFlowContextFromResultRow(row)

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Token)
	s.Equal(*dbModel.Token, *result.Token)
}

func (s *StoreTestSuite) TestBuildFlowContextFromResultRow_WithByteToken() {
	// Test handling when database returns token as []byte (common with PostgreSQL)
	// Setup
	testToken := "byte-token-test" //nolint:gosec // G101: This is test data, not a real credential
	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockGraph.On("GetID").Return("test-graph-id")

	ctx := EngineContext{
		FlowID:   "test-flow-id",
		AppID:    "test-app-id",
		FlowType: common.FlowTypeAuthentication,
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

	// Simulate PostgreSQL returning the token as []byte
	tokenBytes := []byte(*dbModel.Token)

	row := map[string]interface{}{
		"flow_id":           "test-flow-id",
		"app_id":            "test-app-id",
		"verbose":           false,
		"graph_id":          "test-graph-id",
		"runtime_data":      "{}",
		"execution_history": "{}",
		"is_authenticated":  false,
		"user_inputs":       "{}",
		"user_attributes":   "{}",
		"token":             tokenBytes, // Token as []byte
	}

	// Execute
	result, err := store.buildFlowContextFromResultRow(row)

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Token)
	s.Equal(*dbModel.Token, *result.Token)
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
	mockTx := modelmock.NewTxInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("BeginTx").Return(mockTx, nil)

	// Expect two Exec calls: one for FLOW_CONTEXT, one for FLOW_USER_DATA
	mockTx.On("Exec", QueryCreateFlowContext, "test-flow-id", "test-app-id", false,
		mock.Anything, mock.Anything, "test-graph-id", mock.Anything, mock.Anything, "test-deployment").Return(nil, nil)

	// Available attributes serialization is tested in model_test.go, so we just use mock.Anything here
	mockTx.On("Exec", QueryCreateFlowUserData, "test-flow-id", true, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "test-deployment").
		Return(nil, nil)

	mockTx.On("Commit").Return(nil)

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
			IsAuthenticated:     true,
			UserID:              "user-123",
			AvailableAttributes: testAvailableAttributes,
			Attributes:          map[string]interface{}{},
		},
		UserInputs:       map[string]string{},
		RuntimeData:      map[string]string{},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Graph:            mockGraph,
	}

	// Execute
	err := store.StoreFlowContext(ctx)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
	mockTx.AssertExpectations(s.T())
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
	mockTx := modelmock.NewTxInterfaceMock(s.T())
	mockGraph := coremock.NewGraphInterfaceMock(s.T())

	mockGraph.On("GetID").Return("test-graph-id")

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("BeginTx").Return(mockTx, nil)

	mockTx.On("Exec", QueryUpdateFlowContext, "test-flow-id", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, "test-deployment").Return(nil, nil)

	// Available attributes serialization is tested in model_test.go, so we just use mock.Anything here
	mockTx.On("Exec", QueryUpdateFlowUserData, "test-flow-id", true, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "test-deployment").
		Return(nil, nil)

	mockTx.On("Commit").Return(nil)

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
	err := store.UpdateFlowContext(ctx)

	// Verify
	s.NoError(err)
	mockDBProvider.AssertExpectations(s.T())
	mockDBClient.AssertExpectations(s.T())
	mockTx.AssertExpectations(s.T())
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

	// Create serialized available attributes
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
	s.NotNil(dbModel.AvailableAttributes)

	// Setup mocks
	mockDBProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockDBClient := providermock.NewDBClientInterfaceMock(s.T())

	userID := "user-789"
	results := []map[string]interface{}{
		{
			"flow_id":              "test-flow-id",
			"app_id":               "test-app-id",
			"verbose":              false,
			"current_node_id":      nil,
			"current_action":       nil,
			"graph_id":             "test-graph-id",
			"runtime_data":         "{}",
			"execution_history":    "{}",
			"is_authenticated":     true,
			"user_id":              userID,
			"ou_id":                nil,
			"user_type":            nil,
			"user_inputs":          "{}",
			"user_attributes":      "{}",
			"available_attributes": *dbModel.AvailableAttributes,
		},
	}

	mockDBProvider.On("GetRuntimeDBClient").Return(mockDBClient, nil)
	mockDBClient.On("Query", QueryGetFlowContextWithUserData, "test-flow-id", "test-deployment").Return(results, nil)

	store := &flowStore{
		dbProvider:   mockDBProvider,
		deploymentID: "test-deployment",
	}

	// Execute
	result, err := store.GetFlowContext("test-flow-id")

	// Verify
	s.NoError(err)
	s.NotNil(result)
	s.Equal("test-flow-id", result.FlowID)
	s.True(result.IsAuthenticated)
	s.NotNil(result.AvailableAttributes)
	s.Equal(*dbModel.AvailableAttributes, *result.AvailableAttributes) // Serialized attributes should match

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
