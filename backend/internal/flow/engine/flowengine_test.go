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

package engine

import (
	"errors"
	"sync"
	"testing"

	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlowEngineTestSuite struct {
	suite.Suite
	engine FlowEngineInterface
}

func TestFlowEngineTestSuite(t *testing.T) {
	suite.Run(t, new(FlowEngineTestSuite))
}

func (suite *FlowEngineTestSuite) SetupTest() {
	// Reset singleton
	instance = nil
	once = sync.Once{}
	suite.engine = GetFlowEngine()
}

func (suite *FlowEngineTestSuite) TestGetFlowEngine() {
	engine1 := GetFlowEngine()
	engine2 := GetFlowEngine()

	assert.NotNil(suite.T(), engine1)
	assert.NotNil(suite.T(), engine2)
	assert.Same(suite.T(), engine1, engine2, "Should return the same singleton instance")
}

func (suite *FlowEngineTestSuite) TestExecute_GraphNotInitialized() {
	ctx := &model.EngineContext{
		FlowID:   "test-flow",
		FlowType: constants.FlowTypeAuthentication,
		AppID:    "test-app",
		Graph:    nil, // Graph not initialized
	}

	step, err := suite.engine.Execute(ctx)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorFlowGraphNotInitialized.Code, err.Code)
	assert.Equal(suite.T(), "test-flow", step.FlowID)
}

func (suite *FlowEngineTestSuite) TestExecute_StartNodeNotFound() {
	mockGraph := &MockGraph{
		startNodeErr: errors.New("start node not found"),
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorStartNodeNotFoundInGraph.Code, err.Code)
	assert.Equal(suite.T(), "test-flow", step.FlowID)
}

func (suite *FlowEngineTestSuite) TestExecute_TaskExecutionNode_Success() {
	mockNode := &MockNode{
		id:             "task-node",
		nodeType:       constants.NodeTypeTaskExecution,
		nextNodes:      []string{},
		executor:       &MockExecutor{}, // Set executor so it doesn't need construction
		executorConfig: &model.ExecutorConfig{Name: "TestExecutor"},
		response:       &model.NodeResponse{Status: constants.NodeStatusComplete},
		executeErr:     nil,
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"task-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:        "test-flow",
		FlowType:      constants.FlowTypeAuthentication,
		AppID:         "test-app",
		Graph:         mockGraph,
		CurrentNode:   nil,
		UserInputData: map[string]string{"username": "test"},
		RuntimeData:   map[string]string{"session": "12345"},
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test-flow", step.FlowID)
	assert.Equal(suite.T(), constants.FlowStatusComplete, step.Status)
}

func (suite *FlowEngineTestSuite) TestExecute_PromptNode_Incomplete() {
	mockNode := &MockNode{
		id:       "prompt-node",
		nodeType: constants.NodeTypePromptOnly,
		response: &model.NodeResponse{
			Status: constants.NodeStatusIncomplete,
			Type:   constants.NodeResponseTypeView,
			RequiredData: []model.InputData{
				{Name: "username", Type: "text", Required: true},
			},
		},
		executeErr: nil,
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"prompt-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test-flow", step.FlowID)
	assert.Equal(suite.T(), constants.FlowStatusIncomplete, step.Status)
	assert.Equal(suite.T(), constants.StepTypeView, step.Type)
	assert.Len(suite.T(), step.Data.Inputs, 1)
	assert.Equal(suite.T(), "username", step.Data.Inputs[0].Name)
}

func (suite *FlowEngineTestSuite) TestExecute_RedirectionNode_Incomplete() {
	mockNode := &MockNode{
		id:             "redirect-node",
		nodeType:       constants.NodeTypeTaskExecution,
		executor:       &MockExecutor{}, // Set executor so it doesn't need construction
		executorConfig: &model.ExecutorConfig{Name: "TestExecutor"},
		response: &model.NodeResponse{
			Status:      constants.NodeStatusIncomplete,
			Type:        constants.NodeResponseTypeRedirection,
			RedirectURL: "https://external-idp.com/auth",
			AdditionalData: map[string]string{
				"state": "abc123",
			},
		},
		executeErr: nil,
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"redirect-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test-flow", step.FlowID)
	assert.Equal(suite.T(), constants.FlowStatusIncomplete, step.Status)
	assert.Equal(suite.T(), constants.StepTypeRedirection, step.Type)
	assert.Equal(suite.T(), "https://external-idp.com/auth", step.Data.RedirectURL)
	assert.Equal(suite.T(), "abc123", step.Data.AdditionalData["state"])
}

func (suite *FlowEngineTestSuite) TestExecute_NodeExecutionError() {
	mockNode := &MockNode{
		id:             "error-node",
		nodeType:       constants.NodeTypeTaskExecution,
		executor:       &MockExecutor{}, // Set executor so it doesn't need construction
		executorConfig: &model.ExecutorConfig{Name: "TestExecutor"},
		executeErr:     &serviceerror.ServiceError{Code: "TEST_ERROR", Error: "Test execution error"},
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"error-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "TEST_ERROR", err.Code)
	assert.Equal(suite.T(), "test-flow", step.FlowID)
}

func (suite *FlowEngineTestSuite) TestExecute_NodeFailureResponse() {
	mockNode := &MockNode{
		id:             "failure-node",
		nodeType:       constants.NodeTypeTaskExecution,
		executor:       &MockExecutor{}, // Set executor so it doesn't need construction
		executorConfig: &model.ExecutorConfig{Name: "TestExecutor"},
		response: &model.NodeResponse{
			Status:        constants.NodeStatusFailure,
			FailureReason: "Authentication failed",
		},
		executeErr: nil,
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"failure-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test-flow", step.FlowID)
	assert.Equal(suite.T(), constants.FlowStatusError, step.Status)
	assert.Equal(suite.T(), "Authentication failed", step.FailureReason)
}

func (suite *FlowEngineTestSuite) TestExecute_DecisionNode_WithNextNode() {
	nextNode := &MockNode{
		id:             "next-node",
		nodeType:       constants.NodeTypeTaskExecution,
		executor:       &MockExecutor{}, // Set executor so it doesn't need construction
		executorConfig: &model.ExecutorConfig{Name: "TestExecutor"},
		response:       &model.NodeResponse{Status: constants.NodeStatusComplete},
		executeErr:     nil,
	}

	decisionNode := &MockNode{
		id:       "decision-node",
		nodeType: constants.NodeTypeDecision,
		response: &model.NodeResponse{
			Status:     constants.NodeStatusComplete,
			NextNodeID: "next-node",
		},
		executeErr: nil,
	}

	mockGraph := &MockGraph{
		startNode: decisionNode,
		nodes: map[string]model.NodeInterface{
			"decision-node": decisionNode,
			"next-node":     nextNode,
		},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test-flow", step.FlowID)
	assert.Equal(suite.T(), constants.FlowStatusComplete, step.Status)
}

func (suite *FlowEngineTestSuite) TestExecute_UnsupportedNodeResponseStatus() {
	mockNode := &MockNode{
		id:             "unsupported-node",
		nodeType:       constants.NodeTypeTaskExecution,
		executor:       &MockExecutor{}, // Set executor so it doesn't need construction
		executorConfig: &model.ExecutorConfig{Name: "TestExecutor"},
		response: &model.NodeResponse{
			Status: "UNSUPPORTED_STATUS",
		},
		executeErr: nil,
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"unsupported-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	_, err := suite.engine.Execute(ctx)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorUnsupportedNodeResponseStatus.Code, err.Code)
	assert.Contains(suite.T(), err.ErrorDescription, "UNSUPPORTED_STATUS")
}

func (suite *FlowEngineTestSuite) TestExecute_UnsupportedNodeResponseType() {
	mockNode := &MockNode{
		id:             "unsupported-type-node",
		nodeType:       constants.NodeTypeTaskExecution,
		executor:       &MockExecutor{}, // Set executor so it doesn't need construction
		executorConfig: &model.ExecutorConfig{Name: "TestExecutor"},
		response: &model.NodeResponse{
			Status: constants.NodeStatusIncomplete,
			Type:   "UNSUPPORTED_TYPE",
		},
		executeErr: nil,
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"unsupported-type-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	_, err := suite.engine.Execute(ctx)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorUnsupportedNodeResponseType.Code, err.Code)
	assert.Contains(suite.T(), err.ErrorDescription, "UNSUPPORTED_TYPE")
}

// Mock implementations for testing

type MockGraph struct {
	startNode    model.NodeInterface
	startNodeErr error
	nodes        map[string]model.NodeInterface
}

func (m *MockGraph) GetID() string                                 { return "mock-graph" }
func (m *MockGraph) GetType() constants.FlowType                   { return constants.FlowTypeAuthentication }
func (m *MockGraph) GetNodes() map[string]model.NodeInterface      { return m.nodes }
func (m *MockGraph) SetNodes(nodes map[string]model.NodeInterface) { m.nodes = nodes }
func (m *MockGraph) GetEdges() map[string][]string                 { return make(map[string][]string) }
func (m *MockGraph) SetEdges(edges map[string][]string)            {}
func (m *MockGraph) GetStartNodeID() string                        { return "start" }
func (m *MockGraph) SetStartNode(nodeID string) error              { return nil }
func (m *MockGraph) AddNode(node model.NodeInterface) error        { return nil }
func (m *MockGraph) RemoveNode(nodeID string) error                { return nil }
func (m *MockGraph) AddEdge(fromNodeID, toNodeID string) error     { return nil }
func (m *MockGraph) RemoveEdge(fromNodeID, toNodeID string) error  { return nil }
func (m *MockGraph) ToJSON() (string, error)                       { return "{}", nil }

func (m *MockGraph) GetStartNode() (model.NodeInterface, error) {
	if m.startNodeErr != nil {
		return nil, m.startNodeErr
	}
	return m.startNode, nil
}

func (m *MockGraph) GetNode(nodeID string) (model.NodeInterface, bool) {
	if m.nodes == nil {
		return nil, false
	}
	node, exists := m.nodes[nodeID]
	return node, exists
}

type MockNode struct {
	id             string
	nodeType       constants.NodeType
	nextNodes      []string
	previousNodes  []string
	inputData      []model.InputData
	executor       model.ExecutorInterface
	executorConfig *model.ExecutorConfig
	response       *model.NodeResponse
	executeErr     *serviceerror.ServiceError
	isStartNode    bool
	isFinalNode    bool
}

func (m *MockNode) GetID() string                      { return m.id }
func (m *MockNode) GetType() constants.NodeType        { return m.nodeType }
func (m *MockNode) GetNextNodeList() []string          { return m.nextNodes }
func (m *MockNode) SetNextNodeList(nextNodes []string) { m.nextNodes = nextNodes }
func (m *MockNode) AddNextNodeID(nodeID string)        { m.nextNodes = append(m.nextNodes, nodeID) }
func (m *MockNode) RemoveNextNodeID(nodeID string) {
	for i, n := range m.nextNodes {
		if n == nodeID {
			m.nextNodes = append(m.nextNodes[:i], m.nextNodes[i+1:]...)
			break
		}
	}
}
func (m *MockNode) GetPreviousNodeList() []string              { return m.previousNodes }
func (m *MockNode) SetPreviousNodeList(previousNodes []string) { m.previousNodes = previousNodes }
func (m *MockNode) AddPreviousNodeID(nodeID string) {
	m.previousNodes = append(m.previousNodes, nodeID)
}
func (m *MockNode) RemovePreviousNodeID(nodeID string) {
	for i, n := range m.previousNodes {
		if n == nodeID {
			m.previousNodes = append(m.previousNodes[:i], m.previousNodes[i+1:]...)
			break
		}
	}
}
func (m *MockNode) GetInputData() []model.InputData                { return m.inputData }
func (m *MockNode) SetInputData(inputData []model.InputData)       { m.inputData = inputData }
func (m *MockNode) GetExecutor() model.ExecutorInterface           { return m.executor }
func (m *MockNode) SetExecutor(executor model.ExecutorInterface)   { m.executor = executor }
func (m *MockNode) GetExecutorConfig() *model.ExecutorConfig       { return m.executorConfig }
func (m *MockNode) SetExecutorConfig(config *model.ExecutorConfig) { m.executorConfig = config }
func (m *MockNode) IsStartNode() bool                              { return m.isStartNode }
func (m *MockNode) SetAsStartNode()                                { m.isStartNode = true }
func (m *MockNode) IsFinalNode() bool                              { return m.isFinalNode }
func (m *MockNode) SetAsFinalNode()                                { m.isFinalNode = true }

func (m *MockNode) Execute(ctx *model.NodeContext) (*model.NodeResponse, *serviceerror.ServiceError) {
	if m.executeErr != nil {
		return nil, m.executeErr
	}
	return m.response, nil
}

type MockExecutor struct {
	err error
}

func (m *MockExecutor) Execute(ctx *model.NodeContext) (*model.ExecutorResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &model.ExecutorResponse{}, nil
}

func (m *MockExecutor) GetID() string                               { return "mock-executor" }
func (m *MockExecutor) GetName() string                             { return "MockExecutor" }
func (m *MockExecutor) GetProperties() model.ExecutorProperties     { return model.ExecutorProperties{} }
func (m *MockExecutor) GetDefaultExecutorInputs() []model.InputData { return []model.InputData{} }
func (m *MockExecutor) GetPrerequisites() []model.InputData         { return []model.InputData{} }
func (m *MockExecutor) CheckInputData(ctx *model.NodeContext, execResp *model.ExecutorResponse) bool {
	return false
}
func (m *MockExecutor) ValidatePrerequisites(ctx *model.NodeContext, execResp *model.ExecutorResponse) bool {
	return true
}
func (m *MockExecutor) GetUserIDFromContext(ctx *model.NodeContext) (string, error) { return "", nil }
func (m *MockExecutor) GetRequiredData(ctx *model.NodeContext) []model.InputData {
	return []model.InputData{}
}

// Additional comprehensive engine tests to reach 90% coverage

func (suite *FlowEngineTestSuite) TestExecute_ViewResponse() {
	mockNode := &MockNode{
		id:       "view-node",
		nodeType: constants.NodeTypePromptOnly,
		response: &model.NodeResponse{
			Type:   constants.NodeResponseTypeView,
			Status: constants.NodeStatusIncomplete,
			RequiredData: []model.InputData{
				{Name: "username", Type: "string", Required: true},
			},
			Actions: []model.Action{
				{Type: constants.ActionTypeUserInput, ID: "login"},
			},
		},
		nextNodes: []string{"next-node"},
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes: map[string]model.NodeInterface{
			"view-node": mockNode,
			"next-node": &MockNode{id: "next-node", nodeType: constants.NodeTypeTaskExecution},
		},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), constants.StepTypeView, step.Type)
	assert.Equal(suite.T(), constants.FlowStatusIncomplete, step.Status)
	assert.NotEmpty(suite.T(), step.Data.Inputs)
	assert.Equal(suite.T(), "username", step.Data.Inputs[0].Name)
}

func (suite *FlowEngineTestSuite) TestExecute_RedirectionResponse() {
	mockNode := &MockNode{
		id:       "redirect-node",
		nodeType: constants.NodeTypeTaskExecution,
		executor: &MockExecutor{}, // Provide executor to avoid setNodeExecutor
		response: &model.NodeResponse{
			Type:        constants.NodeResponseTypeRedirection,
			Status:      constants.NodeStatusIncomplete,
			RedirectURL: "https://example.com/oauth",
		},
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"redirect-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), constants.StepTypeRedirection, step.Type)
	assert.Equal(suite.T(), constants.FlowStatusIncomplete, step.Status)
	assert.Equal(suite.T(), "https://example.com/oauth", step.Data.RedirectURL)
}

func (suite *FlowEngineTestSuite) TestExecute_CompleteResponse() {
	mockNode := &MockNode{
		id:       "complete-node",
		nodeType: constants.NodeTypeTaskExecution,
		executor: &MockExecutor{}, // Provide executor
		response: &model.NodeResponse{
			Type:   constants.NodeResponseTypeView,
			Status: constants.NodeStatusComplete,
		},
		isFinalNode: true,
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"complete-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), constants.FlowStatusComplete, step.Status)
	assert.Equal(suite.T(), constants.FlowStepType(""), step.Type)
}

func (suite *FlowEngineTestSuite) TestExecute_WithCurrentNodeSet() {
	// Test when CurrentNode is already set in context
	mockNode := &MockNode{
		id:       "current-node",
		nodeType: constants.NodeTypePromptOnly,
		response: &model.NodeResponse{
			Type:   constants.NodeResponseTypeView,
			Status: constants.NodeStatusIncomplete, // Add proper status
			RequiredData: []model.InputData{ // Add required data for VIEW response
				{
					Name:     "username",
					Type:     "text",
					Required: true,
				},
			},
		},
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"current-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: mockNode, // Current node already set
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), constants.FlowStatusIncomplete, step.Status)
	assert.Equal(suite.T(), constants.StepTypeView, step.Type)
}

func (suite *FlowEngineTestSuite) TestExecute_MultipleNodeTransition() {
	// Test scenario with multiple nodes and transitions
	nextNode := &MockNode{
		id:       "next-node",
		nodeType: constants.NodeTypeTaskExecution,
		executor: &MockExecutor{}, // Provide executor
		response: &model.NodeResponse{
			Type:   constants.NodeResponseTypeView,
			Status: constants.NodeStatusComplete,
		},
		isFinalNode: true,
	}

	startNode := &MockNode{
		id:       "start-node",
		nodeType: constants.NodeTypeDecision,
		response: &model.NodeResponse{
			Type:       constants.NodeResponseTypeView,
			Status:     constants.NodeStatusComplete,
			NextNodeID: "next-node",
		},
		nextNodes: []string{"next-node"},
	}

	mockGraph := &MockGraph{
		startNode: startNode,
		nodes: map[string]model.NodeInterface{
			"start-node": startNode,
			"next-node":  nextNode,
		},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err)
	// Should move to next node and complete
	assert.Equal(suite.T(), constants.FlowStatusComplete, step.Status)
}

func (suite *FlowEngineTestSuite) TestExecute_FailureResponse() {
	mockNode := &MockNode{
		id:       "failure-node",
		nodeType: constants.NodeTypeTaskExecution,
		executor: &MockExecutor{}, // Provide executor
		response: &model.NodeResponse{
			Type:          constants.NodeResponseTypeView,
			Status:        constants.NodeStatusFailure,
			FailureReason: "Authentication failed",
		},
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"failure-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	assert.Nil(suite.T(), err) // No error returned for FAILURE status
	assert.Equal(suite.T(), constants.FlowStatusError, step.Status)
	assert.Equal(suite.T(), "Authentication failed", step.FailureReason)
}

func (suite *FlowEngineTestSuite) TestExecute_RetryResponse() {
	mockNode := &MockNode{
		id:       "retry-node",
		nodeType: constants.NodeTypeTaskExecution,
		executor: &MockExecutor{}, // Provide executor
		response: &model.NodeResponse{
			Type:          constants.NodeResponseTypeRetry,
			Status:        constants.NodeStatusIncomplete, // Add status
			FailureReason: "Authentication failed - please retry",
		},
	}

	mockGraph := &MockGraph{
		startNode: mockNode,
		nodes:     map[string]model.NodeInterface{"retry-node": mockNode},
	}

	ctx := &model.EngineContext{
		FlowID:      "test-flow",
		FlowType:    constants.FlowTypeAuthentication,
		AppID:       "test-app",
		Graph:       mockGraph,
		CurrentNode: nil,
	}

	step, err := suite.engine.Execute(ctx)

	// RETRY response type is not yet implemented, so expect an error
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.ErrorDescription, "unsupported node response type: RETRY")
	assert.Equal(suite.T(), constants.FlowStepType(""), step.Type)
	assert.Equal(suite.T(), constants.FlowStatus(""), step.Status)
}
