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

package flow

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	flowCommon "github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/asgardeo/thunder/internal/mcp/tools/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/flow/flowmgtmock"
)

type FlowToolsTestSuite struct {
	suite.Suite
	mockFlowService *flowmgtmock.FlowMgtServiceInterfaceMock
	tools           *flowTools
}

func TestFlowToolsTestSuite(t *testing.T) {
	suite.Run(t, new(FlowToolsTestSuite))
}

func (suite *FlowToolsTestSuite) SetupTest() {
	suite.mockFlowService = flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	suite.tools = NewFlowTools(suite.mockFlowService)
}

func (suite *FlowToolsTestSuite) TestListFlows() {
	mockFlows := []flowmgt.BasicFlowDefinition{
		{
			ID:       "flow-1",
			Handle:   "handle-1",
			FlowType: flowCommon.FlowTypeAuthentication,
		},
	}

	mockResponse := &flowmgt.FlowListResponse{
		TotalResults: 1,
		Flows:        mockFlows,
	}
	suite.mockFlowService.EXPECT().ListFlows(mock.Anything, 100, 0, flowCommon.FlowType("")).Return(mockResponse, nil)

	input := listFlowsInput{
		PaginationInput: common.PaginationInput{
			Limit: 100,
		},
	}
	result, output, err := suite.tools.listFlows(ctx(), nil, input)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 1, output.TotalCount)
	assert.Equal(suite.T(), "flow-1", output.Flows[0].ID)
}

func (suite *FlowToolsTestSuite) TestListFlows_Error() {
	expectedErr := &serviceerror.ServiceError{
		Code:  "ERR_LIST",
		Error: "Failed to list flows",
	}
	suite.mockFlowService.EXPECT().ListFlows(mock.Anything, 10, 0, flowCommon.FlowType("")).Return(nil, expectedErr)

	input := listFlowsInput{
		PaginationInput: common.PaginationInput{
			Limit: 10,
		},
	}
	result, _, err := suite.tools.listFlows(ctx(), nil, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to list flows")
}

func (suite *FlowToolsTestSuite) TestGetFlowByHandle() {
	mockFlow := &flowmgt.CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "handle-1",
		FlowType: flowCommon.FlowTypeAuthentication,
	}
	suite.mockFlowService.EXPECT().GetFlowByHandle(mock.Anything, "handle-1", flowCommon.FlowTypeAuthentication).
		Return(mockFlow, nil)

	input := getFlowByHandleInput{
		Handle:   "handle-1",
		FlowType: string(flowCommon.FlowTypeAuthentication),
	}
	result, output, err := suite.tools.getFlowByHandle(ctx(), nil, input)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "flow-1", output.ID)
}

func (suite *FlowToolsTestSuite) TestGetFlowByID() {
	mockFlow := &flowmgt.CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "handle-1",
		FlowType: flowCommon.FlowTypeAuthentication,
	}
	suite.mockFlowService.EXPECT().GetFlow(mock.Anything, "flow-1").Return(mockFlow, nil)

	input := common.IDInput{
		ID: "flow-1",
	}
	result, output, err := suite.tools.getFlowByID(ctx(), nil, input)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "handle-1", output.Handle)
}

func (suite *FlowToolsTestSuite) TestCreateFlow() {
	input := flowmgt.FlowDefinition{
		Handle:   "new-flow",
		FlowType: flowCommon.FlowTypeRegistration,
	}
	createdFlow := &flowmgt.CompleteFlowDefinition{
		ID:       "flow-new",
		Handle:   "new-flow",
		FlowType: flowCommon.FlowTypeRegistration,
	}
	suite.mockFlowService.EXPECT().CreateFlow(mock.Anything, &input).Return(createdFlow, nil)

	result, output, err := suite.tools.createFlow(ctx(), nil, input)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "flow-new", output.ID)
}

func (suite *FlowToolsTestSuite) TestCreateFlow_Error() {
	input := flowmgt.FlowDefinition{
		Handle:   "error-flow",
		FlowType: flowCommon.FlowTypeAuthentication,
	}
	expectedErr := &serviceerror.ServiceError{
		Code:  "ERR_CREATE",
		Error: "Failed to create flow",
	}
	suite.mockFlowService.EXPECT().CreateFlow(mock.Anything, &input).Return(nil, expectedErr)

	result, output, err := suite.tools.createFlow(ctx(), nil, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), output)
	assert.Contains(suite.T(), err.Error(), "failed to create flow")
}

func (suite *FlowToolsTestSuite) TestUpdateFlow() {
	input := updateFlowInput{
		ID:   "flow-1",
		Name: "Updated Flow",
		Nodes: []flowmgt.NodeDefinition{
			{ID: "start", Type: "START"},
			{ID: "end", Type: "END"},
		},
	}
	updatedFlow := &flowmgt.CompleteFlowDefinition{
		ID:     "flow-1",
		Handle: "updated-handle",
		Nodes:  input.Nodes,
	}

	// Expect GetFlow to be called first
	currentFlow := &flowmgt.CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "updated-handle",
		FlowType: flowCommon.FlowTypeAuthentication,
	}
	suite.mockFlowService.EXPECT().GetFlow(mock.Anything, "flow-1").Return(currentFlow, nil)

	// Expect UpdateFlow with correct definition construction
	suite.mockFlowService.EXPECT().UpdateFlow(mock.Anything, "flow-1", mock.MatchedBy(
		func(def *flowmgt.FlowDefinition) bool {
			return def.Handle == "updated-handle" && def.Name == "Updated Flow"
		})).Return(updatedFlow, nil)

	result, output, err := suite.tools.updateFlow(ctx(), nil, input)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "updated-handle", output.Handle)
}

func (suite *FlowToolsTestSuite) TestFlowSchemas() {
	// listFlowsInputSchema
	assert.NotNil(suite.T(), listFlowsInputSchema)
	assert.Contains(suite.T(), listFlowsInputSchema.Properties, "flow_type")
	assert.ElementsMatch(suite.T(), listFlowsInputSchema.Properties["flow_type"].Enum,
		[]string{string(flowCommon.FlowTypeAuthentication), string(flowCommon.FlowTypeRegistration)})
	assert.Equal(suite.T(), json.RawMessage("30"), listFlowsInputSchema.Properties["limit"].Default)
	assert.Equal(suite.T(), json.RawMessage("0"), listFlowsInputSchema.Properties["offset"].Default)

	// getFlowByHandleInputSchema
	assert.NotNil(suite.T(), getFlowByHandleInputSchema)
	assert.Contains(suite.T(), getFlowByHandleInputSchema.Properties, "flow_type")
	assert.ElementsMatch(suite.T(), getFlowByHandleInputSchema.Properties["flow_type"].Enum,
		[]string{string(flowCommon.FlowTypeAuthentication), string(flowCommon.FlowTypeRegistration)})
	assert.Contains(suite.T(), getFlowByHandleInputSchema.Required, "handle")
	assert.Contains(suite.T(), getFlowByHandleInputSchema.Required, "flow_type")

	// getFlowByIDInputSchema
	assert.NotNil(suite.T(), getFlowByIDInputSchema)
	assert.Contains(suite.T(), getFlowByIDInputSchema.Required, "id")

	// createFlowInputSchema
	assert.NotNil(suite.T(), createFlowInputSchema)
	assert.Contains(suite.T(), createFlowInputSchema.Properties, "flowType")
	assert.ElementsMatch(suite.T(), createFlowInputSchema.Properties["flowType"].Enum,
		[]string{string(flowCommon.FlowTypeAuthentication), string(flowCommon.FlowTypeRegistration)})

	// updateFlowInputSchema
	assert.NotNil(suite.T(), updateFlowInputSchema)
	assert.Contains(suite.T(), updateFlowInputSchema.Required, "id")
	assert.Contains(suite.T(), updateFlowInputSchema.Required, "name")
	assert.Contains(suite.T(), updateFlowInputSchema.Required, "nodes")
}

func (suite *FlowToolsTestSuite) TestRegisterTools() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}, nil)

	// Just verifying it runs without panic and registers
	suite.tools.RegisterTools(server)
}

// Helpers
func ctx() context.Context {
	return context.Background()
}
