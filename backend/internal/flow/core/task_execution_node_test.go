/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
)

type TaskExecutionNodeTestSuite struct {
	suite.Suite
	mockExecutor *ExecutorInterfaceMock
}

func TestTaskExecutionNodeTestSuite(t *testing.T) {
	suite.Run(t, new(TaskExecutionNodeTestSuite))
}

func (s *TaskExecutionNodeTestSuite) SetupTest() {
	s.mockExecutor = NewExecutorInterfaceMock(s.T())
}

func (s *TaskExecutionNodeTestSuite) TestNewTaskExecutionNode() {
	node := newTaskExecutionNode("task-1", map[string]interface{}{"key": "value"}, true, false)

	s.NotNil(node)
	s.Equal("task-1", node.GetID())
	s.Equal(common.NodeTypeTaskExecution, node.GetType())
	s.True(node.IsStartNode())
	s.False(node.IsFinalNode())
}

func (s *TaskExecutionNodeTestSuite) TestExecutorMethods() {
	node := newTaskExecutionNode("task-1", map[string]interface{}{}, false, false)
	execNode, ok := node.(ExecutorBackedNodeInterface)
	s.True(ok)

	s.Empty(execNode.GetExecutorName())
	s.Nil(execNode.GetExecutor())

	execNode.SetExecutorName("test-executor")
	s.Equal("test-executor", execNode.GetExecutorName())

	s.mockExecutor.On("GetName").Return("mock-executor")
	execNode.SetExecutor(s.mockExecutor)
	s.NotNil(execNode.GetExecutor())
	s.Equal("mock-executor", execNode.GetExecutorName())
}

func (s *TaskExecutionNodeTestSuite) TestExecuteNoExecutor() {
	node := newTaskExecutionNode("task-1", map[string]interface{}{}, false, false)
	ctx := &NodeContext{FlowID: "test-flow"}

	resp, err := node.Execute(ctx)

	s.NotNil(err)
	s.Nil(resp)
}

func (s *TaskExecutionNodeTestSuite) TestExecuteSuccess() {
	tests := []struct {
		name           string
		setupMock      func(*ExecutorInterfaceMock)
		expectedStatus common.NodeStatus
		expectedType   common.NodeResponseType
	}{
		{
			name: "Complete execution",
			setupMock: func(m *ExecutorInterfaceMock) {
				m.On("GetName").Return("test-executor").Once()
				m.On("Execute", mock.Anything).Return(
					&common.ExecutorResponse{
						Status:         common.ExecComplete,
						AdditionalData: map[string]string{"key": "value"},
						RuntimeData:    map[string]string{"runtime": "data"},
						AuthenticatedUser: authncm.AuthenticatedUser{
							UserID: "user-123",
						},
					}, nil,
				).Once()
			},
			expectedStatus: common.NodeStatusComplete,
			expectedType:   "",
		},
		{
			name: "User input required",
			setupMock: func(m *ExecutorInterfaceMock) {
				m.On("GetName").Return("test-executor").Once()
				m.On("Execute", mock.Anything).Return(
					&common.ExecutorResponse{
						Status: common.ExecUserInputRequired,
						Inputs: []common.Input{{Identifier: "username", Required: true}},
					}, nil,
				).Once()
			},
			expectedStatus: common.NodeStatusIncomplete,
			expectedType:   common.NodeResponseTypeView,
		},
		{
			name: "External redirection",
			setupMock: func(m *ExecutorInterfaceMock) {
				m.On("GetName").Return("test-executor").Once()
				m.On("Execute", mock.Anything).Return(
					&common.ExecutorResponse{
						Status:      common.ExecExternalRedirection,
						RedirectURL: "https://example.com/auth",
					}, nil,
				).Once()
			},
			expectedStatus: common.NodeStatusIncomplete,
			expectedType:   common.NodeResponseTypeRedirection,
		},
		{
			name: "Retry execution",
			setupMock: func(m *ExecutorInterfaceMock) {
				m.On("GetName").Return("test-executor").Once()
				m.On("Execute", mock.Anything).Return(
					&common.ExecutorResponse{Status: common.ExecRetry},
					nil,
				).Once()
			},
			expectedStatus: common.NodeStatusIncomplete,
			expectedType:   common.NodeResponseTypeRetry,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockExec := NewExecutorInterfaceMock(s.T())
			node := newTaskExecutionNode("task-1", map[string]interface{}{}, false, false)
			execNode, _ := node.(ExecutorBackedNodeInterface)
			tt.setupMock(mockExec)
			execNode.SetExecutor(mockExec)

			ctx := &NodeContext{FlowID: "test-flow"}
			resp, err := node.Execute(ctx)

			s.Nil(err)
			s.NotNil(resp)
			s.Equal(tt.expectedStatus, resp.Status)
			s.Equal(tt.expectedType, resp.Type)
		})
	}
}

func (s *TaskExecutionNodeTestSuite) TestExecuteFailure() {
	s.mockExecutor.On("GetName").Return("test-executor").Once()
	s.mockExecutor.On("Execute", mock.Anything).Return(
		&common.ExecutorResponse{Status: common.ExecFailure, FailureReason: "AUTH_FAILED"},
		nil,
	).Once()

	node := newTaskExecutionNode("task-1", map[string]interface{}{}, false, false)
	execNode, _ := node.(ExecutorBackedNodeInterface)
	execNode.SetExecutor(s.mockExecutor)

	ctx := &NodeContext{FlowID: "test-flow"}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusFailure, resp.Status)
	s.Equal("AUTH_FAILED", resp.FailureReason)
}

func (s *TaskExecutionNodeTestSuite) TestExecuteFailureWithOnFailureHandler() {
	s.mockExecutor.On("GetName").Return("test-executor").Once()
	s.mockExecutor.On("Execute", mock.Anything).Return(
		&common.ExecutorResponse{Status: common.ExecFailure, FailureReason: "AUTH_FAILED"},
		nil,
	).Once()

	node := newTaskExecutionNode("task-1", map[string]interface{}{}, false, false)
	execNode, _ := node.(ExecutorBackedNodeInterface)
	execNode.SetOnFailure("error-prompt")
	execNode.SetExecutor(s.mockExecutor)

	ctx := &NodeContext{FlowID: "test-flow"}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusForward, resp.Status)
	s.Equal("error-prompt", resp.NextNodeID)
	s.Equal("AUTH_FAILED", resp.FailureReason)
	s.NotNil(resp.RuntimeData)
	s.Equal("AUTH_FAILED", resp.RuntimeData["failureReason"])
}

func (s *TaskExecutionNodeTestSuite) TestExecuteExecutorError() {
	s.mockExecutor.On("GetName").Return("test-executor").Once()
	s.mockExecutor.On("Execute", mock.Anything).Return(nil, assert.AnError).Once()

	node := newTaskExecutionNode("task-1", map[string]interface{}{}, false, false)
	execNode, _ := node.(ExecutorBackedNodeInterface)
	execNode.SetExecutor(s.mockExecutor)

	ctx := &NodeContext{FlowID: "test-flow"}
	resp, err := node.Execute(ctx)

	s.NotNil(err)
	s.Nil(resp)
}

func (s *TaskExecutionNodeTestSuite) TestExecuteNilExecutorResponse() {
	node := newTaskExecutionNode("task-1", map[string]interface{}{}, false, false)
	execNode, _ := node.(ExecutorBackedNodeInterface)

	s.mockExecutor.On("GetName").Return("test-executor").Once()
	s.mockExecutor.On("Execute", mock.Anything).Return(nil, nil).Once()
	execNode.SetExecutor(s.mockExecutor)

	ctx := &NodeContext{FlowID: "test-flow"}
	resp, err := node.Execute(ctx)

	s.NotNil(err)
	s.Nil(resp)
}

func (s *TaskExecutionNodeTestSuite) TestExecutePopulatedNodeProperties() {
	mockExec := NewExecutorInterfaceMock(s.T())

	props := map[string]interface{}{"k": "v"}
	node := newTaskExecutionNode("task-props", props, false, false)
	execNode, _ := node.(ExecutorBackedNodeInterface)

	mockExec.On("GetName").Return("test-executor").Once()
	mockExec.On("Execute", mock.Anything).Return(
		&common.ExecutorResponse{Status: common.ExecComplete}, nil,
	).Once()

	execNode.SetExecutor(mockExec)

	ctx := &NodeContext{FlowID: "test-flow"}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(props, ctx.NodeProperties)
}

func (s *TaskExecutionNodeTestSuite) TestBuildNodeResponse() {
	tests := []struct {
		name         string
		execStatus   common.ExecutorStatus
		nodeStatus   common.NodeStatus
		responseType common.NodeResponseType
	}{
		{"ExecComplete", common.ExecComplete, common.NodeStatusComplete, ""},
		{"ExecUserInputRequired", common.ExecUserInputRequired, common.NodeStatusIncomplete,
			common.NodeResponseTypeView},
		{"ExecExternalRedirection", common.ExecExternalRedirection, common.NodeStatusIncomplete,
			common.NodeResponseTypeRedirection},
		{"ExecRetry", common.ExecRetry, common.NodeStatusIncomplete, common.NodeResponseTypeRetry},
		{"ExecFailure", common.ExecFailure, common.NodeStatusFailure, ""},
		{"Unknown status", common.ExecutorStatus("UNKNOWN"), common.NodeStatusIncomplete, ""},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			execResp := &common.ExecutorResponse{Status: tt.execStatus}
			nodeResp := buildNodeResponse(execResp)

			s.NotNil(nodeResp)
			s.Equal(tt.nodeStatus, nodeResp.Status)
			s.Equal(tt.responseType, nodeResp.Type)
			s.NotNil(nodeResp.AdditionalData)
			s.NotNil(nodeResp.RuntimeData)
			s.NotNil(nodeResp.Inputs)
			s.NotNil(nodeResp.Actions)
		})
	}
}
