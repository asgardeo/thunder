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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/executor"
)

type FlowInferenceServiceTestSuite struct {
	suite.Suite
	service flowInferenceServiceInterface
}

func TestFlowInferenceServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FlowInferenceServiceTestSuite))
}

func (s *FlowInferenceServiceTestSuite) SetupTest() {
	s.service = newFlowInferenceService()
}

// Test InferRegistrationFlow

func (s *FlowInferenceServiceTestSuite) TestInferRegistrationFlow_Success() {
	authFlow := &FlowDefinition{
		Name:     "Basic Authentication",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "prompt"},
			{ID: "prompt", Type: "PROMPT", OnSuccess: "auth"},
			{
				ID:   "auth",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Name: executor.ExecutorNameBasicAuth,
				},
				OnSuccess: "end",
			},
			{ID: "end", Type: "END"},
		},
	}

	regFlow, err := s.service.InferRegistrationFlow(authFlow)

	s.NoError(err)
	s.NotNil(regFlow)
	s.Equal("Basic Registration", regFlow.Name)
	s.Equal(common.FlowTypeRegistration, regFlow.FlowType)

	// Verify provisioning node was inserted
	s.True(s.hasNode(regFlow.Nodes, provisioningNodeID))
	provNode := s.getNode(regFlow.Nodes, provisioningNodeID)
	s.Equal(executor.ExecutorNameProvisioning, provNode.Executor.Name)
	s.Equal("end", provNode.OnSuccess)

	// Verify user type resolver was inserted
	s.True(s.hasNode(regFlow.Nodes, userTypeResolverNodeID))
	resolverNode := s.getNode(regFlow.Nodes, userTypeResolverNodeID)
	s.Equal(executor.ExecutorNameUserTypeResolver, resolverNode.Executor.Name)

	// Verify start node points to user type resolver
	startNode := s.getNode(regFlow.Nodes, "start")
	s.Equal(userTypeResolverNodeID, startNode.OnSuccess)
}

func (s *FlowInferenceServiceTestSuite) TestInferRegistrationFlow_WithExistingProvisioning() {
	authFlow := &FlowDefinition{
		Name:     "Auth Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task"},
			{
				ID:   "task",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Name: executor.ExecutorNameProvisioning,
				},
				OnSuccess: "end",
			},
			{ID: "end", Type: "END"},
		},
	}

	regFlow, err := s.service.InferRegistrationFlow(authFlow)

	s.NoError(err)
	s.NotNil(regFlow)
	// Should only have 4 nodes (start, user-type-resolver, task, end) - no duplicate provisioning
	s.Len(regFlow.Nodes, 4)
}

func (s *FlowInferenceServiceTestSuite) TestInferRegistrationFlow_WithExistingUserTypeResolver() {
	authFlow := &FlowDefinition{
		Name:     "Auth Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "resolver"},
			{
				ID:   "resolver",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Name: executor.ExecutorNameUserTypeResolver,
				},
				OnSuccess: "end",
			},
			{ID: "end", Type: "END"},
		},
	}

	regFlow, err := s.service.InferRegistrationFlow(authFlow)

	s.NoError(err)
	s.NotNil(regFlow)
	// Should have 4 nodes (start, resolver, provisioning, end) - no duplicate user type resolver
	s.Len(regFlow.Nodes, 4)
}

func (s *FlowInferenceServiceTestSuite) TestInferRegistrationFlow_CleansAuthProperties() {
	authFlow := &FlowDefinition{
		Name:     "Auth Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task"},
			{
				ID:   "task",
				Type: "TASK_EXECUTION",
				Properties: map[string]interface{}{
					common.NodePropertyAllowAuthenticationWithoutLocalUser: true,
					"otherProp": "value",
				},
				OnSuccess: "end",
			},
			{ID: "end", Type: "END"},
		},
	}

	regFlow, err := s.service.InferRegistrationFlow(authFlow)

	s.NoError(err)
	taskNode := s.getNode(regFlow.Nodes, "task")
	s.NotNil(taskNode.Properties)
	_, hasAuthProp := taskNode.Properties[common.NodePropertyAllowAuthenticationWithoutLocalUser]
	s.False(hasAuthProp)
	s.Equal("value", taskNode.Properties["otherProp"])
}

func (s *FlowInferenceServiceTestSuite) TestInferRegistrationFlow_NoStartNode() {
	authFlow := &FlowDefinition{
		Name:     "Invalid Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "task", Type: "TASK_EXECUTION", OnSuccess: "end"},
			{ID: "end", Type: "END"},
		},
	}

	regFlow, err := s.service.InferRegistrationFlow(authFlow)

	s.Error(err)
	s.Nil(regFlow)
	s.Contains(err.Error(), "no START node found")
}

func (s *FlowInferenceServiceTestSuite) TestInferRegistrationFlow_NoEndNode() {
	authFlow := &FlowDefinition{
		Name:     "Invalid Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task"},
			{ID: "task", Type: "TASK_EXECUTION"},
		},
	}

	regFlow, err := s.service.InferRegistrationFlow(authFlow)

	s.Error(err)
	s.Nil(regFlow)
	s.Contains(err.Error(), "no END node found")
}

func (s *FlowInferenceServiceTestSuite) TestInferRegistrationFlow_WithActions() {
	authFlow := &FlowDefinition{
		Name:     "Auth Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "prompt"},
			{
				ID:   "prompt",
				Type: "PROMPT",
				Actions: []ActionDefinition{
					{Ref: "login", NextNode: "end"},
					{Ref: "signup", NextNode: "end"},
				},
			},
			{ID: "end", Type: "END"},
		},
	}

	regFlow, err := s.service.InferRegistrationFlow(authFlow)

	s.NoError(err)
	promptNode := s.getNode(regFlow.Nodes, "prompt")
	s.Len(promptNode.Actions, 2)
	// Actions should now point to provisioning node instead of end
	s.Equal(provisioningNodeID, promptNode.Actions[0].NextNode)
	s.Equal(provisioningNodeID, promptNode.Actions[1].NextNode)
}

func (s *FlowInferenceServiceTestSuite) TestInferRegistrationFlow_WithOnFailure() {
	authFlow := &FlowDefinition{
		Name:     "Auth Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START", OnSuccess: "task"},
			{
				ID:        "task",
				Type:      "TASK_EXECUTION",
				OnSuccess: "end",
				OnFailure: "end",
			},
			{ID: "end", Type: "END"},
		},
	}

	regFlow, err := s.service.InferRegistrationFlow(authFlow)

	s.NoError(err)
	taskNode := s.getNode(regFlow.Nodes, "task")
	// OnSuccess should point to provisioning
	s.Equal(provisioningNodeID, taskNode.OnSuccess)
	// OnFailure should also point to provisioning (was pointing to end)
	s.Equal(provisioningNodeID, taskNode.OnFailure)
}

// Test generateRegistrationFlowName

func (s *FlowInferenceServiceTestSuite) TestGenerateRegistrationFlowName_Authentication() {
	service := s.service.(*flowInferenceService)

	result := service.generateRegistrationFlowName("Basic Authentication")
	s.Equal("Basic Registration", result)
}

func (s *FlowInferenceServiceTestSuite) TestGenerateRegistrationFlowName_Authenticate() {
	service := s.service.(*flowInferenceService)

	result := service.generateRegistrationFlowName("Basic Authenticate")
	s.Equal("Basic Registration", result)
}

func (s *FlowInferenceServiceTestSuite) TestGenerateRegistrationFlowName_Login() {
	service := s.service.(*flowInferenceService)

	result := service.generateRegistrationFlowName("Login Flow")
	s.Equal("Registration Flow", result)
}

func (s *FlowInferenceServiceTestSuite) TestGenerateRegistrationFlowName_SignIn() {
	service := s.service.(*flowInferenceService)

	result := service.generateRegistrationFlowName("Sign-in Flow")
	s.Equal("Registration Flow", result)
}

func (s *FlowInferenceServiceTestSuite) TestGenerateRegistrationFlowName_Signin() {
	service := s.service.(*flowInferenceService)

	result := service.generateRegistrationFlowName("Signin Flow")
	s.Equal("Registration Flow", result)
}

func (s *FlowInferenceServiceTestSuite) TestGenerateRegistrationFlowName_SignInWithSpace() {
	service := s.service.(*flowInferenceService)

	result := service.generateRegistrationFlowName("Sign in Flow")
	s.Equal("Registration Flow", result)
}

func (s *FlowInferenceServiceTestSuite) TestGenerateRegistrationFlowName_Auth() {
	service := s.service.(*flowInferenceService)

	result := service.generateRegistrationFlowName("Basic Auth")
	s.Equal("Basic Registration", result)
}

func (s *FlowInferenceServiceTestSuite) TestGenerateRegistrationFlowName_NoAuthTerm() {
	service := s.service.(*flowInferenceService)

	result := service.generateRegistrationFlowName("Custom Flow")
	s.Equal("Custom Flow - Registration", result)
}

func (s *FlowInferenceServiceTestSuite) TestGenerateRegistrationFlowName_CaseInsensitive() {
	service := s.service.(*flowInferenceService)

	result := service.generateRegistrationFlowName("basic authentication")
	s.Equal("basic Registration", result)
}

// Test cloneNodes

func (s *FlowInferenceServiceTestSuite) TestCloneNodes_Success() {
	service := s.service.(*flowInferenceService)
	original := []NodeDefinition{
		{ID: "node1", Type: "START", OnSuccess: "node2"},
		{
			ID:   "node2",
			Type: "TASK_EXECUTION",
			Properties: map[string]interface{}{
				"key": "value",
			},
		},
	}

	cloned, err := service.cloneNodes(original)

	s.NoError(err)
	s.Len(cloned, 2)
	s.Equal(original[0].ID, cloned[0].ID)

	// Verify it's a deep copy by modifying clone
	cloned[0].ID = "modified"
	s.NotEqual(original[0].ID, cloned[0].ID)
}

func (s *FlowInferenceServiceTestSuite) TestCloneNodes_EmptyArray() {
	service := s.service.(*flowInferenceService)

	cloned, err := service.cloneNodes([]NodeDefinition{})

	s.NoError(err)
	s.Empty(cloned)
}

// Test cleanAuthenticationProperties

func (s *FlowInferenceServiceTestSuite) TestCleanAuthenticationProperties_RemovesAuthProp() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{
			ID:   "node1",
			Type: "TASK_EXECUTION",
			Properties: map[string]interface{}{
				common.NodePropertyAllowAuthenticationWithoutLocalUser: true,
				"otherProp": "value",
			},
		},
	}

	service.cleanAuthenticationProperties(nodes)

	_, hasAuthProp := nodes[0].Properties[common.NodePropertyAllowAuthenticationWithoutLocalUser]
	s.False(hasAuthProp)
	s.Equal("value", nodes[0].Properties["otherProp"])
}

func (s *FlowInferenceServiceTestSuite) TestCleanAuthenticationProperties_NilProperties() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "node1", Type: "TASK_EXECUTION"},
	}

	// Should not panic
	s.NotPanics(func() {
		service.cleanAuthenticationProperties(nodes)
	})
}

// Test findStartNode

func (s *FlowInferenceServiceTestSuite) TestFindStartNode_Success() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "task", Type: "TASK_EXECUTION"},
		{ID: "start", Type: "START"},
	}

	id, err := service.findStartNode(nodes)

	s.NoError(err)
	s.Equal("start", id)
}

func (s *FlowInferenceServiceTestSuite) TestFindStartNode_NotFound() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "task", Type: "TASK_EXECUTION"},
	}

	id, err := service.findStartNode(nodes)

	s.Error(err)
	s.Empty(id)
	s.Contains(err.Error(), "no START node found")
}

// Test findEndNode

func (s *FlowInferenceServiceTestSuite) TestFindEndNode_Success() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "task", Type: "TASK_EXECUTION"},
		{ID: "end", Type: "END"},
	}

	id, err := service.findEndNode(nodes)

	s.NoError(err)
	s.Equal("end", id)
}

func (s *FlowInferenceServiceTestSuite) TestFindEndNode_NotFound() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "task", Type: "TASK_EXECUTION"},
	}

	id, err := service.findEndNode(nodes)

	s.Error(err)
	s.Empty(id)
	s.Contains(err.Error(), "no END node found")
}

// Test hasProvisioningNode

func (s *FlowInferenceServiceTestSuite) TestHasProvisioningNode_Exists() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{
			ID:   "prov",
			Type: "TASK_EXECUTION",
			Executor: &ExecutorDefinition{
				Name: executor.ExecutorNameProvisioning,
			},
		},
	}

	result := service.hasProvisioningNode(nodes)

	s.True(result)
}

func (s *FlowInferenceServiceTestSuite) TestHasProvisioningNode_NotExists() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "task", Type: "TASK_EXECUTION"},
	}

	result := service.hasProvisioningNode(nodes)

	s.False(result)
}

// Test createProvisioningNode

func (s *FlowInferenceServiceTestSuite) TestCreateProvisioningNode() {
	service := s.service.(*flowInferenceService)

	node := service.createProvisioningNode("end-node")

	s.Equal(provisioningNodeID, node.ID)
	s.Equal(string(common.NodeTypeTaskExecution), node.Type)
	s.NotNil(node.Executor)
	s.Equal(executor.ExecutorNameProvisioning, node.Executor.Name)
	s.Equal("end-node", node.OnSuccess)
}

// Test hasUserTypeResolverNode

func (s *FlowInferenceServiceTestSuite) TestHasUserTypeResolverNode_Exists() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{
			ID:   "resolver",
			Type: "TASK_EXECUTION",
			Executor: &ExecutorDefinition{
				Name: executor.ExecutorNameUserTypeResolver,
			},
		},
	}

	result := service.hasUserTypeResolverNode(nodes)

	s.True(result)
}

func (s *FlowInferenceServiceTestSuite) TestHasUserTypeResolverNode_NotExists() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "task", Type: "TASK_EXECUTION"},
	}

	result := service.hasUserTypeResolverNode(nodes)

	s.False(result)
}

// Test createUserTypeResolverNode

func (s *FlowInferenceServiceTestSuite) TestCreateUserTypeResolverNode() {
	service := s.service.(*flowInferenceService)

	node := service.createUserTypeResolverNode()

	s.Equal(userTypeResolverNodeID, node.ID)
	s.Equal(string(common.NodeTypeTaskExecution), node.Type)
	s.NotNil(node.Executor)
	s.Equal(executor.ExecutorNameUserTypeResolver, node.Executor.Name)
}

// Test insertNodeBeforeEnd

func (s *FlowInferenceServiceTestSuite) TestInsertNodeBeforeEnd_WithOnSuccess() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "start", Type: "START", OnSuccess: "task"},
		{ID: "task", Type: "TASK_EXECUTION", OnSuccess: "end"},
		{ID: "end", Type: "END"},
	}
	newNode := NodeDefinition{ID: "new", Type: "TASK_EXECUTION", OnSuccess: "end"}

	err := service.insertNodeBeforeEnd(&nodes, newNode, "end")

	s.NoError(err)
	s.Len(nodes, 4)
	s.Equal("new", nodes[1].OnSuccess) // task now points to new
	s.Equal("end", nodes[3].OnSuccess) // new node points to end
}

func (s *FlowInferenceServiceTestSuite) TestInsertNodeBeforeEnd_WithOnFailure() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "start", Type: "START", OnSuccess: "task"},
		{ID: "task", Type: "TASK_EXECUTION", OnSuccess: "success", OnFailure: "end"},
		{ID: "success", Type: "END"},
		{ID: "end", Type: "END"},
	}
	newNode := NodeDefinition{ID: "new", Type: "TASK_EXECUTION", OnSuccess: "end"}

	err := service.insertNodeBeforeEnd(&nodes, newNode, "end")

	s.NoError(err)
	s.Equal("new", nodes[1].OnFailure)
}

func (s *FlowInferenceServiceTestSuite) TestInsertNodeBeforeEnd_WithActions() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "start", Type: "START", OnSuccess: "prompt"},
		{
			ID:   "prompt",
			Type: "PROMPT",
			Actions: []ActionDefinition{
				{Ref: "action1", NextNode: "end"},
				{Ref: "action2", NextNode: "task"},
			},
		},
		{ID: "task", Type: "TASK_EXECUTION"},
		{ID: "end", Type: "END"},
	}
	newNode := NodeDefinition{ID: "new", Type: "TASK_EXECUTION", OnSuccess: "end"}

	err := service.insertNodeBeforeEnd(&nodes, newNode, "end")

	s.NoError(err)
	s.Equal("new", nodes[1].Actions[0].NextNode)
	s.Equal("task", nodes[1].Actions[1].NextNode) // unchanged
}

func (s *FlowInferenceServiceTestSuite) TestInsertNodeBeforeEnd_NoNodesPointingToEnd() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "start", Type: "START", OnSuccess: "task"},
		{ID: "task", Type: "TASK_EXECUTION"},
		{ID: "end", Type: "END"},
	}
	newNode := NodeDefinition{ID: "new", Type: "TASK_EXECUTION", OnSuccess: "end"}

	err := service.insertNodeBeforeEnd(&nodes, newNode, "end")

	s.Error(err)
	s.Contains(err.Error(), "no nodes pointing to END node found")
}

// Test insertNodeAfterStart

func (s *FlowInferenceServiceTestSuite) TestInsertNodeAfterStart_Success() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "start", Type: "START", OnSuccess: "task"},
		{ID: "task", Type: "TASK_EXECUTION", OnSuccess: "end"},
		{ID: "end", Type: "END"},
	}
	newNode := NodeDefinition{ID: "new", Type: "TASK_EXECUTION"}

	err := service.insertNodeAfterStart(&nodes, newNode, "start")

	s.NoError(err)
	s.Len(nodes, 4)
	s.Equal("new", nodes[0].OnSuccess)  // start now points to new
	s.Equal("task", nodes[3].OnSuccess) // new node points to task
}

func (s *FlowInferenceServiceTestSuite) TestInsertNodeAfterStart_NoOnSuccess() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "start", Type: "START"},
		{ID: "end", Type: "END"},
	}
	newNode := NodeDefinition{ID: "new", Type: "TASK_EXECUTION"}

	err := service.insertNodeAfterStart(&nodes, newNode, "start")

	s.Error(err)
	s.Contains(err.Error(), "START node has no onSuccess defined")
}

func (s *FlowInferenceServiceTestSuite) TestInsertNodeAfterStart_StartNodeNotFound() {
	service := s.service.(*flowInferenceService)
	nodes := []NodeDefinition{
		{ID: "task", Type: "TASK_EXECUTION"},
	}
	newNode := NodeDefinition{ID: "new", Type: "TASK_EXECUTION"}

	err := service.insertNodeAfterStart(&nodes, newNode, "start")

	s.Error(err)
	s.Contains(err.Error(), "START node not found")
}

// Helper methods

func (s *FlowInferenceServiceTestSuite) hasNode(nodes []NodeDefinition, nodeID string) bool {
	for _, node := range nodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}

func (s *FlowInferenceServiceTestSuite) getNode(nodes []NodeDefinition, nodeID string) *NodeDefinition {
	for i := range nodes {
		if nodes[i].ID == nodeID {
			return &nodes[i]
		}
	}
	return nil
}
