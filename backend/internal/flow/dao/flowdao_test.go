/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package dao

import (
	"sync"
	"testing"

	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlowDAOTestSuite struct {
	suite.Suite
	dao FlowDAOInterface
}

func TestFlowDAOSuite(t *testing.T) {
	suite.Run(t, new(FlowDAOTestSuite))
}

func (suite *FlowDAOTestSuite) SetupTest() {
	// Reset singleton
	instance = nil
	once = sync.Once{}
	suite.dao = GetFlowDAO()
}

func (suite *FlowDAOTestSuite) TestGetFlowDAO() {
	dao1 := GetFlowDAO()
	dao2 := GetFlowDAO()

	assert.NotNil(suite.T(), dao1)
	assert.NotNil(suite.T(), dao2)
	assert.Same(suite.T(), dao1, dao2, "Should return the same singleton instance")
}

func (suite *FlowDAOTestSuite) TestRegisterAndGetGraph() {
	// Create a mock graph
	graphID := "test-graph-123"
	mockGraph := model.NewGraph(graphID, constants.FlowTypeAuthentication)

	// Register the graph
	suite.dao.RegisterGraph(graphID, mockGraph)

	// Retrieve the graph
	retrievedGraph, exists := suite.dao.GetGraph(graphID)

	assert.True(suite.T(), exists, "Graph should exist after registration")
	assert.NotNil(suite.T(), retrievedGraph, "Retrieved graph should not be nil")
	assert.Equal(suite.T(), graphID, retrievedGraph.GetID(), "Graph ID should match")
}

func (suite *FlowDAOTestSuite) TestGetGraphNotFound() {
	nonExistentID := "non-existent-graph"

	retrievedGraph, exists := suite.dao.GetGraph(nonExistentID)

	assert.False(suite.T(), exists, "Graph should not exist")
	assert.Nil(suite.T(), retrievedGraph, "Retrieved graph should be nil")
}

func (suite *FlowDAOTestSuite) TestIsValidGraphID() {
	// Test with non-existent graph
	assert.False(suite.T(), suite.dao.IsValidGraphID("non-existent"), "Non-existent graph should be invalid")

	// Test with empty string
	assert.False(suite.T(), suite.dao.IsValidGraphID(""), "Empty graph ID should be invalid")

	// Register a graph and test
	graphID := "valid-graph-123"
	mockGraph := model.NewGraph(graphID, constants.FlowTypeAuthentication)
	suite.dao.RegisterGraph(graphID, mockGraph)

	assert.True(suite.T(), suite.dao.IsValidGraphID(graphID), "Registered graph should be valid")
}

func (suite *FlowDAOTestSuite) TestStoreAndGetContextFromStore() {
	flowID := "test-flow-123"
	testContext := model.EngineContext{
		FlowID:   flowID,
		FlowType: constants.FlowTypeAuthentication,
		AppID:    "test-app",
		UserInputData: map[string]string{
			"username": "testuser",
		},
		RuntimeData: map[string]string{
			"step": "1",
		},
	}

	// Store the context
	err := suite.dao.StoreContextInStore(flowID, testContext)
	assert.NoError(suite.T(), err, "Storing context should not return error")

	// Retrieve the context
	retrievedContext, exists := suite.dao.GetContextFromStore(flowID)

	assert.True(suite.T(), exists, "Context should exist after storing")
	assert.Equal(suite.T(), flowID, retrievedContext.FlowID, "Flow ID should match")
	assert.Equal(suite.T(), constants.FlowTypeAuthentication, retrievedContext.FlowType, "Flow type should match")
	assert.Equal(suite.T(), "test-app", retrievedContext.AppID, "App ID should match")
	assert.Equal(suite.T(), "testuser", retrievedContext.UserInputData["username"], "User input data should match")
	assert.Equal(suite.T(), "1", retrievedContext.RuntimeData["step"], "Runtime data should match")
}

func (suite *FlowDAOTestSuite) TestStoreContextWithEmptyFlowID() {
	testContext := model.EngineContext{
		FlowID: "",
		AppID:  "test-app",
	}

	err := suite.dao.StoreContextInStore("", testContext)
	assert.Error(suite.T(), err, "Storing context with empty flow ID should return error")
	assert.Contains(suite.T(), err.Error(), "flow ID cannot be empty", "Error message should indicate empty flow ID")
}

func (suite *FlowDAOTestSuite) TestGetContextFromStoreNotFound() {
	nonExistentFlowID := "non-existent-flow"

	retrievedContext, exists := suite.dao.GetContextFromStore(nonExistentFlowID)

	assert.False(suite.T(), exists, "Context should not exist")
	assert.Equal(suite.T(), model.EngineContext{}, retrievedContext, "Retrieved context should be empty")
}

func (suite *FlowDAOTestSuite) TestRemoveContextFromStore() {
	flowID := "test-flow-remove"
	testContext := model.EngineContext{
		FlowID: flowID,
		AppID:  "test-app",
	}

	// Store the context first
	err := suite.dao.StoreContextInStore(flowID, testContext)
	assert.NoError(suite.T(), err, "Storing context should not return error")

	// Verify it exists
	_, exists := suite.dao.GetContextFromStore(flowID)
	assert.True(suite.T(), exists, "Context should exist before removal")

	// Remove the context
	err = suite.dao.RemoveContextFromStore(flowID)
	assert.NoError(suite.T(), err, "Removing context should not return error")

	// Verify it's removed
	_, exists = suite.dao.GetContextFromStore(flowID)
	assert.False(suite.T(), exists, "Context should not exist after removal")
}

func (suite *FlowDAOTestSuite) TestRemoveContextWithEmptyFlowID() {
	err := suite.dao.RemoveContextFromStore("")
	assert.Error(suite.T(), err, "Removing context with empty flow ID should return error")
	assert.Contains(suite.T(), err.Error(), "flow ID cannot be empty", "Error message should indicate empty flow ID")
}

func (suite *FlowDAOTestSuite) TestRemoveNonExistentContext() {
	// Removing a non-existent context should not return an error (idempotent operation)
	err := suite.dao.RemoveContextFromStore("non-existent-flow")
	assert.NoError(suite.T(), err, "Removing non-existent context should not return error")
}

func (suite *FlowDAOTestSuite) TestConcurrentContextOperations() {
	// Test concurrent access to context store
	flowID := "concurrent-test-flow"
	testContext := model.EngineContext{
		FlowID: flowID,
		AppID:  "test-app",
	}

	// Store context
	err := suite.dao.StoreContextInStore(flowID, testContext)
	assert.NoError(suite.T(), err)

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			ctx, exists := suite.dao.GetContextFromStore(flowID)
			assert.True(suite.T(), exists)
			assert.Equal(suite.T(), flowID, ctx.FlowID)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
