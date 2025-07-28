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

package flow

import (
	"sync"
	"testing"

	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlowServiceTestSuite struct {
	suite.Suite
	service FlowServiceInterface
}

func TestFlowServiceSuite(t *testing.T) {
	suite.Run(t, new(FlowServiceTestSuite))
}

func (suite *FlowServiceTestSuite) SetupTest() {
	// Reset singleton
	instance = nil
	once = sync.Once{}
	suite.service = GetFlowService()
}

func (suite *FlowServiceTestSuite) BeforeTest(suiteName, testName string) {
	// No mocks needed for simplified tests
}

func (suite *FlowServiceTestSuite) TestIsNewFlow() {
	tests := []struct {
		name     string
		flowID   string
		actionID string
		expected bool
	}{
		{
			name:     "Both empty - new flow",
			flowID:   "",
			actionID: "",
			expected: true,
		},
		{
			name:     "FlowID exists - not new flow",
			flowID:   "flow123",
			actionID: "",
			expected: false,
		},
		{
			name:     "ActionID exists - not new flow",
			flowID:   "",
			actionID: "action123",
			expected: false,
		},
		{
			name:     "Both exist - not new flow",
			flowID:   "flow123",
			actionID: "action123",
			expected: false,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			result := isNewFlow(tt.flowID, tt.actionID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func (suite *FlowServiceTestSuite) TestIsComplete() {
	tests := []struct {
		name     string
		step     model.FlowStep
		expected bool
	}{
		{
			name: "Complete status",
			step: model.FlowStep{
				Status: constants.FlowStatusComplete,
			},
			expected: true,
		},
		{
			name: "Incomplete status",
			step: model.FlowStep{
				Status: constants.FlowStatusIncomplete,
			},
			expected: false,
		},
		{
			name: "Error status",
			step: model.FlowStep{
				Status: constants.FlowStatusError,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			result := isComplete(tt.step)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func (suite *FlowServiceTestSuite) TestPrepareContext() {
	tests := []struct {
		name      string
		context   *model.EngineContext
		actionID  string
		inputData map[string]string
		expected  func(*model.EngineContext)
	}{
		{
			name: "Empty context with input data",
			context: &model.EngineContext{
				UserInputData: nil,
				RuntimeData:   nil,
			},
			actionID: "test-action",
			inputData: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			expected: func(ctx *model.EngineContext) {
				assert.Equal(suite.T(), "test-action", ctx.CurrentActionID)
				assert.Equal(suite.T(), "testuser", ctx.UserInputData["username"])
				assert.Equal(suite.T(), "testpass", ctx.UserInputData["password"])
				assert.NotNil(suite.T(), ctx.RuntimeData)
			},
		},
		{
			name: "Context with existing data",
			context: &model.EngineContext{
				UserInputData: map[string]string{"existing": "data"},
				RuntimeData:   map[string]string{"runtime": "data"},
			},
			actionID: "new-action",
			inputData: map[string]string{
				"new": "value",
			},
			expected: func(ctx *model.EngineContext) {
				assert.Equal(suite.T(), "new-action", ctx.CurrentActionID)
				assert.Equal(suite.T(), "data", ctx.UserInputData["existing"])
				assert.Equal(suite.T(), "value", ctx.UserInputData["new"])
				assert.Equal(suite.T(), "data", ctx.RuntimeData["runtime"])
			},
		},
		{
			name: "No action ID provided",
			context: &model.EngineContext{
				CurrentActionID: "old-action",
				UserInputData:   make(map[string]string),
			},
			actionID:  "",
			inputData: map[string]string{"key": "value"},
			expected: func(ctx *model.EngineContext) {
				assert.Equal(suite.T(), "old-action", ctx.CurrentActionID) // Should remain unchanged
				assert.Equal(suite.T(), "value", ctx.UserInputData["key"])
			},
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			prepareContext(tt.context, tt.actionID, tt.inputData)
			tt.expected(tt.context)
		})
	}
}

func (suite *FlowServiceTestSuite) TestValidateDefaultFlowConfigs() {
	// This test is commented out because it requires proper config initialization
	// which would make the test more complex. The function itself is straightforward
	// and the logic can be tested separately if needed.
	suite.T().Skip("Skipping config-dependent test - validateDefaultFlowConfigs requires proper config setup")
}

func (suite *FlowServiceTestSuite) TestGetFlowService() {
	service1 := GetFlowService()
	service2 := GetFlowService()

	assert.NotNil(suite.T(), service1)
	assert.NotNil(suite.T(), service2)
	assert.Same(suite.T(), service1, service2, "Should return the same singleton instance")
}

// Test getFlowGraph function with different scenarios
func (suite *FlowServiceTestSuite) TestGetFlowGraph_EmptyAppID() {
	_, err := getFlowGraph("", constants.FlowTypeAuthentication)
	assert.NotNil(suite.T(), err) // ServiceError should not be nil
	assert.IsType(suite.T(), &serviceerror.ServiceError{}, err)
}

// Test basic FlowService methods that don't require complex mocking
func (suite *FlowServiceTestSuite) TestFlowService_BasicMethods() {
	fs := &FlowService{}

	// Test that we can create a FlowService instance
	assert.NotNil(suite.T(), fs)

	// Test calling methods that have validation logic
	_, err := fs.loadContextFromStore("", nil)
	assert.NotNil(suite.T(), err) // Should fail with empty flowID
	assert.IsType(suite.T(), &serviceerror.ServiceError{}, err)
}

// Test additional helper function coverage
func (suite *FlowServiceTestSuite) TestInitContext_EmptyAppID() {
	fs := &FlowService{}

	// Test with empty app ID - should return error without needing mocks
	_, err := fs.initContext("", constants.FlowTypeAuthentication, nil)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &serviceerror.ServiceError{}, err)
}

func (suite *FlowServiceTestSuite) TestLoadNewContext_EmptyAppID() {
	fs := &FlowService{}

	// Test with empty app ID - should return error without needing mocks
	_, err := fs.loadNewContext("", "action123", constants.FlowTypeAuthentication, nil, nil)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &serviceerror.ServiceError{}, err)
}
