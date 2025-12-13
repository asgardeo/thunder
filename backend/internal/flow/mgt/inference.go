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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/system/log"
)

// flowInferenceServiceInterface defines the interface for flow inference services
type flowInferenceServiceInterface interface {
	InferRegistrationFlow(authFlow *FlowDefinition) (*FlowDefinition, error)
}

// flowInferenceService implements FlowInferenceServiceInterface
type flowInferenceService struct {
	logger *log.Logger
}

// newFlowInferenceService creates a new flow inference service instance
func newFlowInferenceService() flowInferenceServiceInterface {
	return &flowInferenceService{
		logger: log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowInferenceService")),
	}
}

// InferRegistrationFlow creates a registration flow definition from an authentication flow
func (s *flowInferenceService) InferRegistrationFlow(authFlow *FlowDefinition) (*FlowDefinition, error) {
	s.logger.Debug("Inferring registration flow from authentication flow",
		log.String("authFlowName", authFlow.Name))

	regFlowName := s.generateRegistrationFlowName(authFlow.Name)

	// Deep copy nodes to avoid modifying the original flow
	regNodes, err := s.cloneNodes(authFlow.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to clone nodes: %w", err)
	}

	s.cleanAuthenticationProperties(regNodes)

	endNodeID, err := s.findEndNode(regNodes)
	if err != nil {
		return nil, err
	}

	if !s.hasProvisioningNode(regNodes) {
		provisioningNode := s.createProvisioningNode(endNodeID)
		if err := s.insertNodeBeforeEnd(&regNodes, provisioningNode, endNodeID); err != nil {
			return nil, err
		}
		s.logger.Debug("Inserted provisioning node into registration flow")
	} else {
		s.logger.Debug("Provisioning node already exists, skipping insertion")
	}

	startNodeID, err := s.findStartNode(regNodes)
	if err != nil {
		return nil, err
	}

	if !s.hasUserTypeResolverNode(regNodes) {
		userTypeResolverNode := s.createUserTypeResolverNode()
		if err := s.insertNodeAfterStart(&regNodes, userTypeResolverNode, startNodeID); err != nil {
			return nil, err
		}
		s.logger.Debug("Inserted user type resolver node into registration flow")
	} else {
		s.logger.Debug("User type resolver node already exists, skipping insertion")
	}

	return &FlowDefinition{
		Name:     regFlowName,
		FlowType: common.FlowTypeRegistration,
		Nodes:    regNodes,
	}, nil
}

// generateRegistrationFlowName generates a registration flow name from an auth flow name.
func (s *flowInferenceService) generateRegistrationFlowName(authFlowName string) string {
	// List of authentication-related terms to replace (ordered by specificity)
	authTerms := []string{"Authentication", "Authenticate", "Sign-in", "Signin", "Sign in", "Login", "Auth"}
	regTerm := "Registration"

	// Try to replace any authentication term with "Registration" (case-insensitive)
	for _, term := range authTerms {
		lowerName := strings.ToLower(authFlowName)
		lowerTerm := strings.ToLower(term)

		if index := strings.Index(lowerName, lowerTerm); index != -1 {
			// Replace preserving the structure of the original name
			return authFlowName[:index] + regTerm + authFlowName[index+len(term):]
		}
	}

	// If no auth term found, append suffix
	return authFlowName + " - Registration"
}

// cloneNodes creates a deep copy of the nodes array
func (s *flowInferenceService) cloneNodes(nodes []NodeDefinition) ([]NodeDefinition, error) {
	// Use JSON marshaling for deep copy
	data, err := json.Marshal(nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nodes: %w", err)
	}

	var clonedNodes []NodeDefinition
	if err := json.Unmarshal(data, &clonedNodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	return clonedNodes, nil
}

// cleanAuthenticationProperties removes authentication-specific properties from nodes
// and sets appropriate registration-specific defaults
func (s *flowInferenceService) cleanAuthenticationProperties(nodes []NodeDefinition) {
	for i := range nodes {
		node := &nodes[i]
		if node.Properties == nil {
			continue
		}

		// Remove authentication-specific properties that don't apply to registration
		delete(node.Properties, common.NodePropertyAllowAuthenticationWithoutLocalUser)

		// Ensure registration-specific properties have appropriate defaults if needed
	}
}

// findStartNode finds the START node in the flow
func (s *flowInferenceService) findStartNode(nodes []NodeDefinition) (string, error) {
	for _, node := range nodes {
		if node.Type == string(common.NodeTypeStart) {
			return node.ID, nil
		}
	}
	return "", fmt.Errorf("no START node found in flow")
}

// findEndNode finds the END node in the flow
func (s *flowInferenceService) findEndNode(nodes []NodeDefinition) (string, error) {
	for _, node := range nodes {
		if node.Type == string(common.NodeTypeEnd) {
			return node.ID, nil
		}
	}
	return "", fmt.Errorf("no END node found in flow")
}

// hasProvisioningNode checks if a provisioning node already exists in the flow
func (s *flowInferenceService) hasProvisioningNode(nodes []NodeDefinition) bool {
	for _, node := range nodes {
		if node.Executor != nil && node.Executor.Name == executor.ExecutorNameProvisioning {
			return true
		}
	}
	return false
}

// createProvisioningNode creates a TASK_EXECUTION node with ProvisioningExecutor
func (s *flowInferenceService) createProvisioningNode(endNodeID string) NodeDefinition {
	return NodeDefinition{
		ID:   provisioningNodeID,
		Type: string(common.NodeTypeTaskExecution),
		Executor: &ExecutorDefinition{
			Name: executor.ExecutorNameProvisioning,
		},
		OnSuccess: endNodeID,
	}
}

// hasUserTypeResolverNode checks if a user type resolver node already exists in the flow
func (s *flowInferenceService) hasUserTypeResolverNode(nodes []NodeDefinition) bool {
	for _, node := range nodes {
		if node.Executor != nil && node.Executor.Name == executor.ExecutorNameUserTypeResolver {
			return true
		}
	}
	return false
}

// createUserTypeResolverNode creates a TASK_EXECUTION node with UserTypeResolverExecutor
func (s *flowInferenceService) createUserTypeResolverNode() NodeDefinition {
	return NodeDefinition{
		ID:   userTypeResolverNodeID,
		Type: string(common.NodeTypeTaskExecution),
		Executor: &ExecutorDefinition{
			Name: executor.ExecutorNameUserTypeResolver,
		},
	}
}

// insertNodeBeforeEnd inserts a node before the END node by updating all nodes that point to END
func (s *flowInferenceService) insertNodeBeforeEnd(nodes *[]NodeDefinition,
	newNode NodeDefinition, endNodeID string) error {
	modified := false
	for i := range *nodes {
		node := &(*nodes)[i]

		// Update onSuccess if it points to END
		if node.OnSuccess != "" && node.OnSuccess == endNodeID {
			node.OnSuccess = newNode.ID
			modified = true
		}

		// Update onFailure if it points to END
		if node.OnFailure != "" && node.OnFailure == endNodeID {
			node.OnFailure = newNode.ID
			modified = true
		}

		// Update actions that point to END
		for j := range node.Actions {
			if node.Actions[j].NextNode == endNodeID {
				node.Actions[j].NextNode = newNode.ID
				modified = true
			}
		}
	}

	if !modified {
		return fmt.Errorf("no nodes pointing to END node found")
	}

	// Append the new node to the array
	*nodes = append(*nodes, newNode)
	return nil
}

// insertNodeAfterStart inserts a node after the START node
func (s *flowInferenceService) insertNodeAfterStart(nodes *[]NodeDefinition,
	newNode NodeDefinition, startNodeID string) error {
	// Append the new node to the array first
	*nodes = append(*nodes, newNode)

	// Find START node and get its original next node
	var originalNext string
	for i := range *nodes {
		if (*nodes)[i].ID == startNodeID {
			if (*nodes)[i].OnSuccess == "" {
				return fmt.Errorf("START node has no onSuccess defined")
			}
			originalNext = (*nodes)[i].OnSuccess
			(*nodes)[i].OnSuccess = newNode.ID
			break
		}
	}

	if originalNext == "" {
		return fmt.Errorf("START node not found")
	}

	// Find the newly appended node and update its onSuccess
	for i := range *nodes {
		if (*nodes)[i].ID == newNode.ID {
			(*nodes)[i].OnSuccess = originalNext
			return nil
		}
	}

	return fmt.Errorf("new node %s not found in array", newNode.ID)
}
