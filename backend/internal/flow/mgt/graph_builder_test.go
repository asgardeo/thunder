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

package flowmgt

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/flow/executormock"
)

type GraphBuilderTestSuite struct {
	suite.Suite
	mockFlowFactory      *coremock.FlowFactoryInterfaceMock
	mockExecutorRegistry *executormock.ExecutorRegistryInterfaceMock
	mockGraphCache       *coremock.GraphCacheInterfaceMock
	builder              *graphBuilder
}

func TestGraphBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(GraphBuilderTestSuite))
}

func (s *GraphBuilderTestSuite) SetupTest() {
	_ = config.InitializeThunderRuntime("test", &config.Config{
		Server: config.ServerConfig{Identifier: "test-deployment"},
	})

	s.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(s.T())
	s.mockExecutorRegistry = executormock.NewExecutorRegistryInterfaceMock(s.T())
	s.mockGraphCache = coremock.NewGraphCacheInterfaceMock(s.T())

	s.builder = &graphBuilder{
		flowFactory:      s.mockFlowFactory,
		executorRegistry: s.mockExecutorRegistry,
		graphCache:       s.mockGraphCache,
		logger:           log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowGraphBuilder")),
	}
}

// Test GetGraph method

func (s *GraphBuilderTestSuite) TestGetGraph_NilFlow() {
	graph, err := s.builder.GetGraph(nil)

	s.Nil(graph)
	s.NotNil(err)
	s.Equal(ErrorInvalidFlowData.Code, err.Code)
	s.Contains(err.ErrorDescription, "Flow definition is nil or has no nodes")
}

func (s *GraphBuilderTestSuite) TestGetGraph_EmptyNodes() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{},
	}

	graph, err := s.builder.GetGraph(flow)

	s.Nil(graph)
	s.NotNil(err)
	s.Equal(ErrorInvalidFlowData.Code, err.Code)
	s.Contains(err.ErrorDescription, "Flow definition is nil or has no nodes")
}

func (s *GraphBuilderTestSuite) TestGetGraph_CacheHit() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	s.mockGraphCache.EXPECT().Get("flow-1").Return(mockGraph, true)

	graph, err := s.builder.GetGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
	s.Equal(mockGraph, graph)
}

func (s *GraphBuilderTestSuite) TestGetGraph_BuildAndCache() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "end"},
			{ID: "end", Type: "END"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewRepresentationNodeInterfaceMock(s.T())
	mockEndNode := coremock.NewRepresentationNodeInterfaceMock(s.T())

	s.mockGraphCache.EXPECT().Get("flow-1").Return(nil, false)
	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"end", "END", map[string]interface{}(nil), false, true).Return(
		mockEndNode, nil)

	mockStartNode.EXPECT().SetOnSuccess("end")

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockEndNode).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "end").Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"start": mockStartNode, "end": mockEndNode})
	// Map iteration order is non-deterministic, so other nodes might be checked before START is found
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockEndNode.EXPECT().GetType().Return(common.NodeTypeEnd).Maybe()
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	s.mockGraphCache.EXPECT().Set("flow-1", mockGraph).Return(nil)

	graph, err := s.builder.GetGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
	s.Equal(mockGraph, graph)
}

func (s *GraphBuilderTestSuite) TestGetGraph_BuildFailure() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	s.mockGraphCache.EXPECT().Get("flow-1").Return(nil, false)
	s.mockFlowFactory.EXPECT().CreateGraph("flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, true).Return(
		nil, errors.New("node creation error"))

	graph, err := s.builder.GetGraph(flow)

	s.Nil(graph)
	s.NotNil(err)
	s.Equal(ErrorGraphBuildFailure.Code, err.Code)
	s.Contains(err.ErrorDescription, "node creation error")
}

func (s *GraphBuilderTestSuite) TestGetGraph_CacheSetError() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewNodeInterfaceMock(s.T())

	s.mockGraphCache.EXPECT().Get("flow-1").Return(nil, false)
	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, true).Return(
		mockStartNode, nil)

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().GetNodes().Return(map[string]core.NodeInterface{"start": mockStartNode})
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	s.mockGraphCache.EXPECT().Set("flow-1", mockGraph).Return(errors.New("cache error"))

	graph, err := s.builder.GetGraph(flow)

	// Should still return graph even if caching fails
	s.NotNil(graph)
	s.Nil(err)
	s.Equal(mockGraph, graph)
}

// Test InvalidateCache method

func (s *GraphBuilderTestSuite) TestInvalidateCache_EmptyFlowID() {
	// Should not panic or error
	s.builder.InvalidateCache("")
}

func (s *GraphBuilderTestSuite) TestInvalidateCache_Success() {
	s.mockGraphCache.EXPECT().Invalidate("flow-1").Return(nil)

	s.builder.InvalidateCache("flow-1")
}

func (s *GraphBuilderTestSuite) TestInvalidateCache_Error() {
	s.mockGraphCache.EXPECT().Invalidate("flow-1").Return(errors.New("cache error"))

	// Should log error but not panic
	s.builder.InvalidateCache("flow-1")
}

// Test buildGraph method

func (s *GraphBuilderTestSuite) TestBuildGraph_WithExecutor() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task"},
			{
				ID:       "task",
				Type:     "TASK_EXECUTION",
				Executor: &ExecutorDefinition{Name: "test-executor"},
			},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewNodeInterfaceMock(s.T())
	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, true).Return(
		mockTaskNode, nil)

	s.mockExecutorRegistry.EXPECT().IsRegistered("test-executor").Return(true)
	mockTaskNode.EXPECT().SetExecutorName("test-executor")
	mockTaskNode.EXPECT().SetInputs([]common.Input{})

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockTaskNode).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "task").Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"start": mockStartNode, "task": mockTaskNode})
	// Map iteration order is non-deterministic, so other nodes might be checked before START is found
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockTaskNode.EXPECT().GetType().Return(common.NodeTypeTaskExecution).Maybe()
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	graph, err := s.builder.buildGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestBuildGraph_ExecutorNotRegistered() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{
				ID:       "task",
				Type:     "TASK_EXECUTION",
				Executor: &ExecutorDefinition{Name: "unknown-executor"},
			},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, true).Return(
		mockTaskNode, nil)
	mockTaskNode.EXPECT().SetInputs([]common.Input{})

	s.mockExecutorRegistry.EXPECT().IsRegistered("unknown-executor").Return(false)

	graph, err := s.builder.buildGraph(flow)

	s.Nil(graph)
	s.NotNil(err)
	s.Contains(err.Error(), "executor with name unknown-executor not registered")
}

func (s *GraphBuilderTestSuite) TestBuildGraph_WithOnFailure() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task"},
			{
				ID:        "task",
				Type:      "TASK_EXECUTION",
				OnSuccess: "end",
				OnFailure: "error-prompt",
				Executor:  &ExecutorDefinition{Name: "test-executor"},
			},
			{ID: "error-prompt", Type: "PROMPT"},
			{ID: "end", Type: "END"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewRepresentationNodeInterfaceMock(s.T())
	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockPromptNode := coremock.NewPromptNodeInterfaceMock(s.T())
	mockEndNode := coremock.NewRepresentationNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, false).Return(
		mockTaskNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"error-prompt", "PROMPT", map[string]interface{}(nil), false, true).Return(
		mockPromptNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"end", "END", map[string]interface{}(nil), false, true).Return(
		mockEndNode, nil)

	mockStartNode.EXPECT().SetOnSuccess("task")
	mockTaskNode.EXPECT().SetOnSuccess("end")
	mockTaskNode.EXPECT().SetOnFailure("error-prompt")
	mockTaskNode.EXPECT().SetInputs([]common.Input{})

	s.mockExecutorRegistry.EXPECT().IsRegistered("test-executor").Return(true)
	mockTaskNode.EXPECT().SetExecutorName("test-executor")

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockTaskNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockPromptNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockEndNode).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "task").Return(nil)
	mockGraph.EXPECT().AddEdge("task", "end").Return(nil)
	mockGraph.EXPECT().AddEdge("task", "error-prompt").Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"start": mockStartNode, "task": mockTaskNode,
			"error-prompt": mockPromptNode, "end": mockEndNode})
	// Map iteration order is non-deterministic, so other nodes might be checked before START is found
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockTaskNode.EXPECT().GetType().Return(common.NodeTypeTaskExecution).Maybe()
	mockPromptNode.EXPECT().GetType().Return(common.NodeTypePrompt).Maybe()
	mockEndNode.EXPECT().GetType().Return(common.NodeTypeEnd).Maybe()
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	graph, err := s.builder.buildGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestBuildGraph_OnFailureNotPromptNode() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{
				ID:        "task",
				Type:      "TASK_EXECUTION",
				OnFailure: "end",
			},
			{ID: "end", Type: "END"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, false).Return(
		mockTaskNode, nil)

	// Validation fails during configureNodeNavigation, before SetInputs is called
	// END node is not created because task node processing fails first

	graph, err := s.builder.buildGraph(flow)

	s.Nil(graph)
	s.NotNil(err)
	s.Contains(err.Error(), "onFailure must point to a PROMPT node")
}

func (s *GraphBuilderTestSuite) TestBuildGraph_OnFailureTargetNotFound() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{
				ID:        "task",
				Type:      "TASK_EXECUTION",
				OnFailure: "non-existent",
			},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, false).Return(
		mockTaskNode, nil)

	// Validation fails during configureNodeNavigation, before SetInputs is called

	graph, err := s.builder.buildGraph(flow)

	s.Nil(graph)
	s.NotNil(err)
	s.Contains(err.Error(), "onFailure target node not found")
}

func (s *GraphBuilderTestSuite) TestBuildGraph_WithInputs() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task"},
			{
				ID:   "task",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Inputs: []InputDefinition{
						{Ref: "username", Type: "string", Identifier: "user", Required: true},
						{Ref: "password", Type: "string", Identifier: "pass", Required: true},
					},
				},
			},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewRepresentationNodeInterfaceMock(s.T())
	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, true).Return(
		mockTaskNode, nil)

	mockStartNode.EXPECT().SetOnSuccess("task")
	mockTaskNode.EXPECT().SetInputs([]common.Input{
		{Ref: "username", Type: "string", Identifier: "user", Required: true},
		{Ref: "password", Type: "string", Identifier: "pass", Required: true},
	})

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockTaskNode).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "task").Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"start": mockStartNode, "task": mockTaskNode})
	// Map iteration order is non-deterministic, so other nodes might be checked before START is found
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockTaskNode.EXPECT().GetType().Return(common.NodeTypeTaskExecution).Maybe()
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	graph, err := s.builder.buildGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestBuildGraph_WithActions() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "prompt"},
			{
				ID:   "prompt",
				Type: "PROMPT",
				Prompts: []PromptDefinition{
					{Action: &ActionDefinition{Ref: "login", NextNode: "task1"}},
					{Action: &ActionDefinition{Ref: "signup", NextNode: "task2"}},
				},
			},
			{ID: "task1", Type: "TASK_EXECUTION"},
			{ID: "task2", Type: "TASK_EXECUTION"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewRepresentationNodeInterfaceMock(s.T())
	mockPromptNode := coremock.NewPromptNodeInterfaceMock(s.T())
	mockTask1Node := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockTask2Node := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"prompt", "PROMPT", map[string]interface{}(nil), false, false).Return(
		mockPromptNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task1", "TASK_EXECUTION", map[string]interface{}(nil), false, true).Return(
		mockTask1Node, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task2", "TASK_EXECUTION", map[string]interface{}(nil), false, true).Return(
		mockTask2Node, nil)

	mockStartNode.EXPECT().SetOnSuccess("prompt")

	mockPromptNode.EXPECT().SetPrompts(mock.MatchedBy(func(prompts []common.Prompt) bool {
		if len(prompts) != 2 {
			return false
		}
		if prompts[0].Action.Ref != "login" || prompts[0].Action.NextNode != "task1" {
			return false
		}
		if prompts[1].Action.Ref != "signup" || prompts[1].Action.NextNode != "task2" {
			return false
		}
		return true
	}))
	mockTask1Node.EXPECT().SetInputs([]common.Input{})
	mockTask2Node.EXPECT().SetInputs([]common.Input{})

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockPromptNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockTask1Node).Return(nil)
	mockGraph.EXPECT().AddNode(mockTask2Node).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "prompt").Return(nil)
	mockGraph.EXPECT().AddEdge("prompt", "task1").Return(nil)
	mockGraph.EXPECT().AddEdge("prompt", "task2").Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"start": mockStartNode, "prompt": mockPromptNode,
			"task1": mockTask1Node, "task2": mockTask2Node})
	// Map iteration order is non-deterministic, so other nodes might be checked before START is found
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockPromptNode.EXPECT().GetType().Return(common.NodeTypePrompt).Maybe()
	mockTask1Node.EXPECT().GetType().Return(common.NodeTypeTaskExecution).Maybe()
	mockTask2Node.EXPECT().GetType().Return(common.NodeTypeTaskExecution).Maybe()
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	graph, err := s.builder.buildGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestBuildGraph_WithMeta() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "prompt"},
			{
				ID:   "prompt",
				Type: "PROMPT",
				Meta: map[string]interface{}{"title": "Login", "description": "Enter credentials"},
			},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewRepresentationNodeInterfaceMock(s.T())
	mockPromptNode := coremock.NewPromptNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"prompt", "PROMPT", map[string]interface{}(nil), false, true).Return(
		mockPromptNode, nil)

	mockStartNode.EXPECT().SetOnSuccess("prompt")

	mockPromptNode.EXPECT().SetMeta(map[string]interface{}{
		"title": "Login", "description": "Enter credentials"})

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockPromptNode).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "prompt").Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"start": mockStartNode, "prompt": mockPromptNode})
	// Map iteration order is non-deterministic, so other nodes might be checked before START is found
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockPromptNode.EXPECT().GetType().Return(common.NodeTypePrompt).Maybe()
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	graph, err := s.builder.buildGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestBuildGraph_WithCondition() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task"},
			{
				ID:   "task",
				Type: "TASK_EXECUTION",
				Condition: &ConditionDefinition{
					Key:    "userType",
					Value:  "premium",
					OnSkip: "end",
				},
			},
			{ID: "end", Type: "END"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewRepresentationNodeInterfaceMock(s.T())
	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockEndNode := coremock.NewRepresentationNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, true).Return(
		mockTaskNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"end", "END", map[string]interface{}(nil), false, true).Return(
		mockEndNode, nil)

	mockStartNode.EXPECT().SetOnSuccess("task")
	mockTaskNode.EXPECT().SetInputs([]common.Input{})
	mockTaskNode.EXPECT().SetCondition(&core.NodeCondition{
		Key:    "userType",
		Value:  "premium",
		OnSkip: "end",
	})

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockTaskNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockEndNode).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "task").Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"start": mockStartNode, "task": mockTaskNode,
			"end": mockEndNode})
	// Map iteration order is non-deterministic, so other nodes might be checked before START is found
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockTaskNode.EXPECT().GetType().Return(common.NodeTypeTaskExecution).Maybe()
	mockEndNode.EXPECT().GetType().Return(common.NodeTypeEnd).Maybe()
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	graph, err := s.builder.buildGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestBuildGraph_NoStartNode() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "task", Type: "TASK_EXECUTION"},
			{ID: "end", Type: "END"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockTaskNode := coremock.NewNodeInterfaceMock(s.T())
	mockEndNode := coremock.NewNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, true).Return(
		mockTaskNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"end", "END", map[string]interface{}(nil), false, true).Return(
		mockEndNode, nil)

	mockGraph.EXPECT().AddNode(mockTaskNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockEndNode).Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"task": mockTaskNode, "end": mockEndNode})
	mockTaskNode.EXPECT().GetType().Return(common.NodeTypeTaskExecution)
	mockEndNode.EXPECT().GetType().Return(common.NodeTypeEnd)

	graph, err := s.builder.buildGraph(flow)

	s.Nil(graph)
	s.NotNil(err)
	s.Contains(err.Error(), "no start node found")
}

func (s *GraphBuilderTestSuite) TestBuildGraph_AddNodeError() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, true).Return(
		mockStartNode, nil)

	mockGraph.EXPECT().AddNode(mockStartNode).Return(errors.New("duplicate node"))

	graph, err := s.builder.buildGraph(flow)

	s.Nil(graph)
	s.NotNil(err)
	s.Contains(err.Error(), "failed to add node start to the graph")
}

func (s *GraphBuilderTestSuite) TestBuildGraph_AddEdgeError() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "end"},
			{ID: "end", Type: "END"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewNodeInterfaceMock(s.T())
	mockEndNode := coremock.NewNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"end", "END", map[string]interface{}(nil), false, true).Return(
		mockEndNode, nil)

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockEndNode).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "end").Return(errors.New("edge creation error"))

	graph, err := s.builder.buildGraph(flow)

	s.Nil(graph)
	s.NotNil(err)
	s.Contains(err.Error(), "failed to add edge from start to end")
}

func (s *GraphBuilderTestSuite) TestBuildGraph_SetStartNodeError() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, true).Return(
		mockStartNode, nil)

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().GetNodes().Return(map[string]core.NodeInterface{"start": mockStartNode})
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(errors.New("start node already set"))

	graph, err := s.builder.buildGraph(flow)

	s.Nil(graph)
	s.NotNil(err)
	s.Contains(err.Error(), "start node already set")
}

func (s *GraphBuilderTestSuite) TestBuildGraph_WithProperties() {
	properties := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task", Properties: properties},
			{ID: "task", Type: "TASK_EXECUTION"},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewRepresentationNodeInterfaceMock(s.T())
	mockTaskNode := coremock.NewRepresentationNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", properties, false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, true).Return(
		mockTaskNode, nil)

	mockStartNode.EXPECT().SetOnSuccess("task")

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockTaskNode).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "task").Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"start": mockStartNode, "task": mockTaskNode})
	// Map iteration order is non-deterministic, so other nodes might be checked before START is found
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockTaskNode.EXPECT().GetType().Return(common.NodeTypeTaskExecution).Maybe()
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	graph, err := s.builder.buildGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestValidateExecutorName_EmptyName() {
	err := s.builder.validateExecutorName("")

	s.NotNil(err)
	s.Contains(err.Error(), "executor name cannot be empty")
}

func (s *GraphBuilderTestSuite) TestValidateExecutorName_NotRegistered() {
	s.mockExecutorRegistry.EXPECT().IsRegistered("unknown").Return(false)

	err := s.builder.validateExecutorName("unknown")

	s.NotNil(err)
	s.Contains(err.Error(), "executor with name unknown not registered")
}

func (s *GraphBuilderTestSuite) TestValidateExecutorName_Success() {
	s.mockExecutorRegistry.EXPECT().IsRegistered("test-executor").Return(true)

	err := s.builder.validateExecutorName("test-executor")

	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestBuildGraph_WithExecutorMode() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-1",
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task"},
			{
				ID:       "task",
				Type:     "TASK_EXECUTION",
				Executor: &ExecutorDefinition{Name: "SMSOTPAuthExecutor", Mode: "send"},
			},
		},
	}

	mockGraph := coremock.NewGraphInterfaceMock(s.T())
	mockStartNode := coremock.NewNodeInterfaceMock(s.T())
	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockFlowFactory.EXPECT().CreateGraph(
		"flow-1", common.FlowTypeAuthentication).Return(
		mockGraph)
	s.mockFlowFactory.EXPECT().CreateNode(
		"start", "START", map[string]interface{}(nil), false, false).Return(
		mockStartNode, nil)
	s.mockFlowFactory.EXPECT().CreateNode(
		"task", "TASK_EXECUTION", map[string]interface{}(nil), false, true).Return(
		mockTaskNode, nil)

	s.mockExecutorRegistry.EXPECT().IsRegistered("SMSOTPAuthExecutor").Return(true)
	mockTaskNode.EXPECT().SetExecutorName("SMSOTPAuthExecutor")
	mockTaskNode.EXPECT().SetMode("send") // Verify mode is set
	mockTaskNode.EXPECT().SetInputs([]common.Input{})

	mockGraph.EXPECT().AddNode(mockStartNode).Return(nil)
	mockGraph.EXPECT().AddNode(mockTaskNode).Return(nil)
	mockGraph.EXPECT().AddEdge("start", "task").Return(nil)
	mockGraph.EXPECT().GetNodes().Return(
		map[string]core.NodeInterface{"start": mockStartNode, "task": mockTaskNode})
	mockStartNode.EXPECT().GetType().Return(common.NodeTypeStart)
	mockTaskNode.EXPECT().GetType().Return(common.NodeTypeTaskExecution).Maybe()
	mockStartNode.EXPECT().GetID().Return("start")
	mockGraph.EXPECT().SetStartNode("start").Return(nil)

	graph, err := s.builder.buildGraph(flow)

	s.NotNil(graph)
	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestConfigureNodeExecutor_NilExecutor() {
	nodeDef := &NodeDefinition{
		ID:       "task",
		Type:     "TASK_EXECUTION",
		Executor: nil, // Nil executor
	}

	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	err := s.builder.configureNodeExecutor(nodeDef, mockTaskNode)

	s.Nil(err)
	// No mock expectations should be called
}

func (s *GraphBuilderTestSuite) TestConfigureNodeExecutor_EmptyExecutorName() {
	nodeDef := &NodeDefinition{
		ID:   "task",
		Type: "TASK_EXECUTION",
		Executor: &ExecutorDefinition{
			Name: "", // Empty executor name
			Mode: "send",
		},
	}

	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	err := s.builder.configureNodeExecutor(nodeDef, mockTaskNode)

	s.Nil(err)
	// No mock expectations should be called since name is empty
}

func (s *GraphBuilderTestSuite) TestConfigureNodeExecutor_NodeDoesNotSupportExecutors() {
	nodeDef := &NodeDefinition{
		ID:   "prompt",
		Type: "PROMPT",
		Executor: &ExecutorDefinition{
			Name: "test-executor",
		},
	}

	// Use a regular NodeInterface that doesn't support executors
	mockPromptNode := coremock.NewNodeInterfaceMock(s.T())

	// Should silently skip executor configuration for non-executor nodes
	err := s.builder.configureNodeExecutor(nodeDef, mockPromptNode)
	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestConfigureNodeExecutor_ExecutorNameValidationFails() {
	nodeDef := &NodeDefinition{
		ID:   "task",
		Type: "TASK_EXECUTION",
		Executor: &ExecutorDefinition{
			Name: "unregistered-executor",
		},
	}

	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockExecutorRegistry.EXPECT().IsRegistered("unregistered-executor").Return(false)

	err := s.builder.configureNodeExecutor(nodeDef, mockTaskNode)

	s.NotNil(err)
	s.Contains(err.Error(), "error while validating executor")
	s.Contains(err.Error(), "executor with name unregistered-executor not registered")
}

func (s *GraphBuilderTestSuite) TestConfigureNodeExecutor_WithModeSuccess() {
	nodeDef := &NodeDefinition{
		ID:   "task",
		Type: "TASK_EXECUTION",
		Executor: &ExecutorDefinition{
			Name: "test-executor",
			Mode: "verify",
		},
	}

	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockExecutorRegistry.EXPECT().IsRegistered("test-executor").Return(true)
	mockTaskNode.EXPECT().SetExecutorName("test-executor")
	mockTaskNode.EXPECT().SetMode("verify")

	err := s.builder.configureNodeExecutor(nodeDef, mockTaskNode)

	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestConfigureNodeExecutor_WithoutModeSuccess() {
	nodeDef := &NodeDefinition{
		ID:   "task",
		Type: "TASK_EXECUTION",
		Executor: &ExecutorDefinition{
			Name: "test-executor",
			Mode: "", // Empty mode - should not call SetMode
		},
	}

	mockTaskNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())

	s.mockExecutorRegistry.EXPECT().IsRegistered("test-executor").Return(true)
	mockTaskNode.EXPECT().SetExecutorName("test-executor")
	// SetMode should NOT be called when mode is empty

	err := s.builder.configureNodeExecutor(nodeDef, mockTaskNode)

	s.Nil(err)
}

func (s *GraphBuilderTestSuite) TestConfigureNodeExecutor_EmptyExecutorNameInValidation() {
	// This tests the validateExecutorName method with empty name
	err := s.builder.validateExecutorName("")

	s.NotNil(err)
	s.Contains(err.Error(), "executor name cannot be empty")
}
