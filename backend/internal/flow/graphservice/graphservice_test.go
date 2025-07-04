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

package graphservice

import (
	"sync"
	"testing"

	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/dao"
	"github.com/asgardeo/thunder/internal/flow/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GraphServiceTestSuite struct {
	suite.Suite
	service GraphServiceInterface
}

func TestGraphServiceTestSuite(t *testing.T) {
	suite.Run(t, new(GraphServiceTestSuite))
}

func (suite *GraphServiceTestSuite) SetupTest() {
	// Reset singleton
	instance = nil
	once = sync.Once{}
	suite.service = GetGraphService()
}

func (suite *GraphServiceTestSuite) TestGetGraphService() {
	service1 := GetGraphService()
	service2 := GetGraphService()

	assert.NotNil(suite.T(), service1)
	assert.NotNil(suite.T(), service2)
	assert.Same(suite.T(), service1, service2, "Should return the same singleton instance")
}

func (suite *GraphServiceTestSuite) TestIsValidGraphID_ValidGraph() {
	// First register a graph in DAO
	flowDAO := dao.GetFlowDAO()
	graphID := "test-valid-graph"
	mockGraph := model.NewGraph(graphID, constants.FlowTypeAuthentication)
	flowDAO.RegisterGraph(graphID, mockGraph)

	// Test that the service recognizes it as valid
	isValid := suite.service.IsValidGraphID(graphID)

	assert.True(suite.T(), isValid, "Registered graph should be valid")
}

func (suite *GraphServiceTestSuite) TestIsValidGraphID_InvalidGraph() {
	// Test with a non-existent graph ID
	nonExistentID := "non-existent-graph-id"
	isValid := suite.service.IsValidGraphID(nonExistentID)

	assert.False(suite.T(), isValid, "Non-existent graph should be invalid")
}

func (suite *GraphServiceTestSuite) TestIsValidGraphID_EmptyID() {
	// Test with empty graph ID
	isValid := suite.service.IsValidGraphID("")

	assert.False(suite.T(), isValid, "Empty graph ID should be invalid")
}

func (suite *GraphServiceTestSuite) TestIsValidGraphID_WhitespaceID() {
	// Test with whitespace-only graph ID
	isValid := suite.service.IsValidGraphID("   ")

	assert.False(suite.T(), isValid, "Whitespace-only graph ID should be invalid")
}

func (suite *GraphServiceTestSuite) TestIsValidGraphID_MultipleGraphs() {
	// Register multiple graphs
	flowDAO := dao.GetFlowDAO()

	graphs := map[string]constants.FlowType{
		"auth-graph": constants.FlowTypeAuthentication,
		"reg-graph":  constants.FlowTypeRegistration,
	}

	// Register all graphs
	for graphID, flowType := range graphs {
		mockGraph := model.NewGraph(graphID, flowType)
		flowDAO.RegisterGraph(graphID, mockGraph)
	}

	// Test that all registered graphs are valid
	for graphID := range graphs {
		isValid := suite.service.IsValidGraphID(graphID)
		assert.True(suite.T(), isValid, "Registered graph %s should be valid", graphID)
	}

	// Test that a non-registered graph is invalid
	isValid := suite.service.IsValidGraphID("unregistered-graph")
	assert.False(suite.T(), isValid, "Unregistered graph should be invalid")
}
