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

// Package flowmgt provides the flow management service implementation.
package flowmgt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// FlowMgtServiceInterface defines the interface for the flow management service
type FlowMgtServiceInterface interface {
	RegisterGraph(graphID string, g core.GraphInterface)
	GetGraph(graphID string) (core.GraphInterface, bool)
	IsValidGraphID(graphID string) bool
}

// flowMgtService is the implementation of FlowMgtServiceInterface
type flowMgtService struct {
	graphs           map[string]core.GraphInterface
	mu               sync.Mutex
	flowFactory      core.FlowFactoryInterface
	executorRegistry executor.ExecutorRegistryInterface
	logger           *log.Logger
}

// newFlowMgtService creates a new instance of FlowMgtServiceInterface
func newFlowMgtService(flowFactory core.FlowFactoryInterface,
	executorRegistry executor.ExecutorRegistryInterface) (FlowMgtServiceInterface, error) {
	flowMgtInstance := &flowMgtService{
		graphs:           make(map[string]core.GraphInterface),
		mu:               sync.Mutex{},
		flowFactory:      flowFactory,
		executorRegistry: executorRegistry,
		logger:           log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowMgtService")),
	}
	err := flowMgtInstance.init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize flow management service: %w", err)
	}

	err = flowMgtInstance.validateDefaultFlowConfigs()
	if err != nil {
		return nil, errors.New("failed to validate default flow configurations: " + err.Error())
	}
	return flowMgtInstance, nil
}

// Init initializes the FlowMgtService by loading graph configurations into runtime
func (s *flowMgtService) init() error {
	logger := s.logger
	logger.Debug("Initializing the flow management service")

	configDir := config.GetThunderRuntime().Config.Flow.GraphDirectory
	if configDir == "" {
		logger.Info("Graph directory is not set. No graphs will be loaded.")
		return nil
	}

	configDir = filepath.Join(config.GetThunderRuntime().ThunderHome, configDir)
	configDir = filepath.Clean(configDir)

	logger.Debug("Loading graphs from config directory", log.String("configDir", configDir))

	files, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Config directory does not exist. No graphs will be loaded.",
				log.String("configDir", configDir))
			return nil
		}
		return fmt.Errorf("failed to read config directory %s: %w", configDir, err)
	}

	if len(files) == 0 {
		logger.Info("No graph configuration files found in the configured directory. No graphs will be loaded.")
		return nil
	}
	logger.Debug("Found graph definition files in the graph directory", log.Int("fileCount", len(files)))

	// Process each JSON file in the directory
	flowGraphs := make(map[string]core.GraphInterface)
	for _, file := range files {
		// Skip directories and non-JSON files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			logger.Debug("Skipping non-JSON file or directory",
				log.String("fileName", file.Name()), log.Bool("isDir", file.IsDir()))
			continue
		}
		filePath := filepath.Join(configDir, file.Name())
		filePath = filepath.Clean(filePath)

		// Read the file content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("Failed to read graph file", log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Parse the JSON into the flow model
		var jsonGraph graphDefinition
		if err := json.Unmarshal(fileContent, &jsonGraph); err != nil {
			logger.Warn("Failed to parse JSON in file", log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Convert the JSON graph definition to the graph model
		graphModel, err := s.buildGraphFromDefinition(&jsonGraph)
		if err != nil {
			logger.Warn("Failed to convert graph definition to graph model",
				log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Log the graph model as JSON for debugging
		if logger.IsDebugEnabled() {
			jsonString, err := graphModel.ToJSON()
			if err != nil {
				logger.Warn("Failed to convert graph model to JSON", log.String("filePath", filePath), log.Error(err))
			} else {
				logger.Debug("Graph model loaded successfully", log.String("graphID", graphModel.GetID()),
					log.String("json", jsonString))
			}
		}

		// Append graph to the flowGraphs map
		flowGraphs[graphModel.GetID()] = graphModel
	}

	// Register all loaded graphs
	inferredGraphCount := 0
	for graphID, graph := range flowGraphs {
		// Create and register the equivalent registration graph if not found already.
		registrationGraphID := s.getRegistrationGraphID(graphID)
		_, exists := s.graphs[registrationGraphID]
		if !exists && graph.GetType() == common.FlowTypeAuthentication {
			if err := s.createAndRegisterRegistrationGraph(registrationGraphID, graph); err != nil {
				logger.Error("Failed creating registration graph", log.String("graphID", graphID), log.Error(err))
				continue
			}
			inferredGraphCount++
		}

		logger.Debug("Registering graph", log.String("graphType", string(graph.GetType())),
			log.String("graphID", graphID))
		s.RegisterGraph(graphID, graph)
	}

	logger.Debug("Flow management service initialized successfully", log.Int("configuredGraphCount", len(flowGraphs)),
		log.Int("inferredGraphCount", inferredGraphCount))

	return nil
}

// RegisterGraph registers a graph with the FlowMgtService by its ID
func (s *flowMgtService) RegisterGraph(graphID string, g core.GraphInterface) {
	s.graphs[graphID] = g
}

// GetGraph retrieves a graph by its ID
func (s *flowMgtService) GetGraph(graphID string) (core.GraphInterface, bool) {
	g, ok := s.graphs[graphID]
	return g, ok
}

// IsValidGraphID checks if the provided graph ID is valid and exists in the service
func (s *flowMgtService) IsValidGraphID(graphID string) bool {
	if graphID == "" {
		return false
	}
	_, exists := s.graphs[graphID]
	return exists
}

// buildGraphFromDefinition builds a graph from a graph definition json
func (s *flowMgtService) buildGraphFromDefinition(definition *graphDefinition) (core.GraphInterface, error) {
	if definition == nil || len(definition.Nodes) == 0 {
		return nil, fmt.Errorf("graph definition is nil or has no nodes")
	}

	// Create a graph
	_type, err := getGraphType(definition.Type)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving graph type: %w", err)
	}
	g := s.flowFactory.CreateGraph(definition.ID, _type)

	// Process all nodes and build the graph structure
	edges := make(map[string][]string)
	for _, nodeDef := range definition.Nodes {
		if err := s.processNodeDefinition(&nodeDef, definition.Nodes, g, edges); err != nil {
			return nil, err
		}
	}

	if err := s.addGraphEdges(g, edges); err != nil {
		return nil, err
	}

	if err := s.determineAndSetStartNode(g); err != nil {
		return nil, err
	}

	return g, nil
}

// processNodeDefinition processes a single node definition and adds it to the graph
func (s *flowMgtService) processNodeDefinition(nodeDef *nodeDefinition, allNodes []nodeDefinition,
	g core.GraphInterface, edges map[string][]string) error {
	isFinalNode := nodeDef.OnSuccess == "" && nodeDef.OnFailure == "" && len(nodeDef.Actions) == 0

	// Construct a new node. Here we set isStartNode to false by default
	node, err := s.flowFactory.CreateNode(nodeDef.ID, nodeDef.Type, nodeDef.Properties, false, isFinalNode)
	if err != nil {
		return fmt.Errorf("failed to create node %s: %w", nodeDef.ID, err)
	}

	if err := s.configureNodeNavigation(nodeDef, allNodes, node, edges); err != nil {
		return err
	}

	s.configureNodeInputs(nodeDef, node)

	if err := s.configureNodeActions(nodeDef, node, edges); err != nil {
		return err
	}

	s.configureNodeCondition(nodeDef, node)

	if err := s.configureNodeExecutor(nodeDef, node, g.GetType()); err != nil {
		return err
	}

	// Add node to the graph
	if err := g.AddNode(node); err != nil {
		return fmt.Errorf("failed to add node %s to the graph: %w", nodeDef.ID, err)
	}

	return nil
}

// configureNodeNavigation configures the onSuccess and onFailure properties for a node
func (s *flowMgtService) configureNodeNavigation(nodeDef *nodeDefinition, allNodes []nodeDefinition,
	node core.NodeInterface, edges map[string][]string) error {
	// Set onSuccess if defined
	if nodeDef.OnSuccess != "" {
		if taskNode, ok := node.(core.ExecutorBackedNodeInterface); ok {
			taskNode.SetOnSuccess(nodeDef.OnSuccess)
		}

		// Add edge for graph structure
		if _, exists := edges[nodeDef.ID]; !exists {
			edges[nodeDef.ID] = []string{}
		}
		edges[nodeDef.ID] = append(edges[nodeDef.ID], nodeDef.OnSuccess)
	}

	// Set onFailure if defined
	if nodeDef.OnFailure != "" {
		if err := validateOnFailureTarget(allNodes, nodeDef.OnFailure); err != nil {
			return fmt.Errorf("invalid onFailure configuration for node %s: %w", nodeDef.ID, err)
		}
		if taskNode, ok := node.(core.ExecutorBackedNodeInterface); ok {
			taskNode.SetOnFailure(nodeDef.OnFailure)
		}

		// Add edge for graph structure
		if _, exists := edges[nodeDef.ID]; !exists {
			edges[nodeDef.ID] = []string{}
		}
		edges[nodeDef.ID] = append(edges[nodeDef.ID], nodeDef.OnFailure)
	}

	return nil
}

// validateOnFailureTarget validates that the onFailure target node is a PROMPT node
func validateOnFailureTarget(nodes []nodeDefinition, targetNodeID string) error {
	for _, node := range nodes {
		if node.ID == targetNodeID {
			if node.Type != "PROMPT" {
				return errors.New("onFailure must point to a PROMPT node")
			}
			return nil
		}
	}
	return errors.New("onFailure target node not found")
}

// configureNodeInputs configures the inputs for a node
func (s *flowMgtService) configureNodeInputs(nodeDef *nodeDefinition, node core.NodeInterface) {
	inputs := make([]common.Input, len(nodeDef.Inputs))
	for i, input := range nodeDef.Inputs {
		inputs[i] = common.Input{
			Ref:        input.Ref,
			Identifier: input.Identifier,
			Type:       input.Type,
			Required:   input.Required,
		}
	}
	node.SetInputs(inputs)
}

// configureNodeActions configures the actions for a prompt node
func (s *flowMgtService) configureNodeActions(nodeDef *nodeDefinition, node core.NodeInterface,
	edges map[string][]string) error {
	if len(nodeDef.Actions) == 0 {
		return nil
	}

	actions := make([]common.Action, len(nodeDef.Actions))
	for i, action := range nodeDef.Actions {
		actions[i] = common.Action{
			Ref:      action.Ref,
			NextNode: action.NextNode,
		}
		if _, exists := edges[nodeDef.ID]; !exists {
			edges[nodeDef.ID] = []string{}
		}
		edges[nodeDef.ID] = append(edges[nodeDef.ID], action.NextNode)
	}

	// Set actions only if the node is a prompt node
	if promptNode, ok := node.(core.PromptNodeInterface); ok {
		promptNode.SetActions(actions)
	}

	return nil
}

// configureNodeCondition configures the condition for a node
func (s *flowMgtService) configureNodeCondition(nodeDef *nodeDefinition, node core.NodeInterface) {
	if nodeDef.Condition != nil && (nodeDef.Condition.Key != "" || nodeDef.Condition.Value != "") {
		node.SetCondition(&core.NodeCondition{
			Key:    nodeDef.Condition.Key,
			Value:  nodeDef.Condition.Value,
			OnSkip: nodeDef.Condition.OnSkip,
		})
	}
}

// configureNodeExecutor configures the executor for a node
func (s *flowMgtService) configureNodeExecutor(nodeDef *nodeDefinition, node core.NodeInterface,
	flowType common.FlowType) error {
	executorName := nodeDef.Executor.Name

	// Determine executor name for special node types if not explicitly defined
	if executorName == "" {
		if nodeDef.Type == string(common.NodeTypeAuthSuccess) {
			executorName = executor.ExecutorNameAuthAssert
		} else if nodeDef.Type == string(common.NodeTypeStart) {
			if flowType == common.FlowTypeRegistration || flowType == common.FlowTypeAuthentication {
				executorName = executor.ExecutorNameUserTypeResolver
			}
			// Related init executors for other flow types can be added here
		}
	}

	// Set executor name if defined
	if executorName != "" {
		if err := s.validateExecutorName(executorName); err != nil {
			return fmt.Errorf("error while validating executor %s: %w", executorName, err)
		}
		executableNode, ok := node.(core.ExecutorBackedNodeInterface)
		if !ok {
			return fmt.Errorf("node %s of type %s does not support executors", nodeDef.ID, nodeDef.Type)
		}
		executableNode.SetExecutorName(executorName)
	}

	return nil
}

// addGraphEdges adds all collected edges to the graph
func (s *flowMgtService) addGraphEdges(g core.GraphInterface, edges map[string][]string) error {
	for sourceID, targetIDs := range edges {
		for _, targetID := range targetIDs {
			if err := g.AddEdge(sourceID, targetID); err != nil {
				return fmt.Errorf("failed to add edge from %s to %s: %w", sourceID, targetID, err)
			}
		}
	}
	return nil
}

// determineAndSetStartNode determines the start node and sets it in the graph
func (s *flowMgtService) determineAndSetStartNode(g core.GraphInterface) error {
	for _, node := range g.GetNodes() {
		if len(node.GetPreviousNodeList()) == 0 {
			return g.SetStartNode(node.GetID())
		}
	}
	return fmt.Errorf("no start node found in the graph definition")
}

// validateExecutorName validates that an executor with the given name is registered
func (s *flowMgtService) validateExecutorName(executorName string) error {
	if executorName == "" {
		return fmt.Errorf("executor name cannot be empty")
	}
	if !s.executorRegistry.IsRegistered(executorName) {
		return fmt.Errorf("executor with name %s not registered", executorName)
	}

	return nil
}

// getRegistrationGraphID constructs the registration graph ID from the auth graph ID
func (s *flowMgtService) getRegistrationGraphID(authGraphID string) string {
	return common.RegistrationFlowGraphPrefix + strings.TrimPrefix(authGraphID, common.AuthFlowGraphPrefix)
}

// createAndRegisterRegistrationGraph creates a registration graph from an authentication graph and registers it
func (s *flowMgtService) createAndRegisterRegistrationGraph(
	registrationGraphID string, authGraph core.GraphInterface) error {
	logger := s.logger
	registrationGraph, err := s.createRegistrationGraph(registrationGraphID, authGraph)
	if err != nil {
		return fmt.Errorf("failed to infer registration graph: %w", err)
	}

	if logger.IsDebugEnabled() {
		registrationGraphJSON, err := registrationGraph.ToJSON()
		if err != nil {
			logger.Warn("Failed to convert graph model to JSON", log.String("graphID", registrationGraphID),
				log.Error(err))
		} else {
			logger.Debug("Graph model loaded successfully", log.String("graphID", registrationGraph.GetID()),
				log.String("json", registrationGraphJSON))
		}
	}

	logger.Debug("Registering inferred registration graph", log.String("graphID", registrationGraph.GetID()))
	s.RegisterGraph(registrationGraph.GetID(), registrationGraph)
	return nil
}

// createRegistrationGraph creates a registration graph from an authentication graph
func (s *flowMgtService) createRegistrationGraph(registrationGraphID string,
	authGraph core.GraphInterface) (core.GraphInterface, error) {
	logger := s.logger
	logger.Debug("Creating registration graph from authentication graph",
		log.String("registrationGraphID", registrationGraphID),
		log.String("authGraphID", authGraph.GetID()))

	// Create a new graph from the authentication graph
	registrationGraph := s.flowFactory.CreateGraph(registrationGraphID, common.FlowTypeRegistration)
	nodesCopy, err := s.flowFactory.CloneNodes(authGraph.GetNodes())
	if err != nil {
		return nil, fmt.Errorf("failed to clone nodes from auth graph: %w", err)
	}
	registrationGraph.SetNodes(nodesCopy)
	registrationGraph.SetEdges(sysutils.DeepCopyMapOfStringSlices(authGraph.GetEdges()))

	// Find authentication success node to insert provisioning node before it
	authSuccessNodeID := ""
	nodes := registrationGraph.GetNodes()
	for nodeID, node := range nodes {
		if node.IsFinalNode() {
			authSuccessNodeID = nodeID
			break
		}
	}
	if authSuccessNodeID == "" {
		return nil, fmt.Errorf("no authentication success node found in the authentication graph")
	}

	// Create and add provisioning node
	provisioningNode, err := s.createProvisioningNode()
	if err != nil {
		return nil, fmt.Errorf("failed to create provisioning node: %w", err)
	}
	err = registrationGraph.AddNode(provisioningNode)
	if err != nil {
		return nil, fmt.Errorf("failed to add provisioning node to registration graph: %w", err)
	}

	// Modify the edges that lead to the auth success node to point to the provisioning node
	for fromNodeID, toNodeIDs := range registrationGraph.GetEdges() {
		for _, toNodeID := range toNodeIDs {
			if toNodeID == authSuccessNodeID {
				err := registrationGraph.RemoveEdge(fromNodeID, toNodeID)
				if err != nil {
					return nil, fmt.Errorf("failed to remove edge from %s to %s: %w", fromNodeID, toNodeID, err)
				}

				err = registrationGraph.AddEdge(fromNodeID, provisioningNode.GetID())
				if err != nil {
					return nil, fmt.Errorf("failed to add edge from %s to provisioning node: %w", fromNodeID, err)
				}
			}
		}
	}

	// Add an edge from the provisioning node to the auth success node
	err = registrationGraph.AddEdge(provisioningNode.GetID(), authSuccessNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to add edge from provisioning node to auth success node: %w", err)
	}

	// Insert the registration start node
	regStartNode, err := s.createRegistrationStartNode()
	if err != nil {
		return nil, fmt.Errorf("failed to create registration start node: %w", err)
	}
	err = registrationGraph.AddNode(regStartNode)
	if err != nil {
		return nil, fmt.Errorf("failed to add registration start node: %w", err)
	}

	// Add edge from registration start to the previous start node
	if err := registrationGraph.AddEdge(regStartNode.GetID(), authGraph.GetStartNodeID()); err != nil {
		return nil, fmt.Errorf("failed to connect registration start to previous start: %w", err)
	}
	if err := registrationGraph.SetStartNode(regStartNode.GetID()); err != nil {
		return nil, fmt.Errorf("failed to set registration start node as start: %w", err)
	}

	// Log the graph model as JSON for debugging
	if logger.IsDebugEnabled() {
		jsonString, err := registrationGraph.ToJSON()
		if err != nil {
			logger.Warn("Failed to convert graph model to JSON",
				log.String("graphID", registrationGraph.GetID()), log.Error(err))
		} else {
			logger.Debug("Registration graph model loaded successfully",
				log.String("graphID", registrationGraph.GetID()), log.String("json", jsonString))
		}
	}

	return registrationGraph, nil
}

// createProvisioningNode creates a provisioning node that leads to the specified auth success node
func (s *flowMgtService) createProvisioningNode() (core.NodeInterface, error) {
	provisioningNode, err := s.flowFactory.CreateNode(
		"provisioning",
		string(common.NodeTypeTaskExecution),
		map[string]interface{}{},
		false,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create provisioning node: %w", err)
	}

	// Set executor name for the provisioning node
	if executableNode, ok := provisioningNode.(core.ExecutorBackedNodeInterface); ok {
		executableNode.SetExecutorName(executor.ExecutorNameProvisioning)
	} else {
		return nil, fmt.Errorf("provisioning node does not implement ExecutorBackedNodeInterface")
	}

	return provisioningNode, nil
}

// createRegistrationStartNode creates the registration start node
func (s *flowMgtService) createRegistrationStartNode() (core.NodeInterface, error) {
	regStartNode, err := s.flowFactory.CreateNode(
		"start",
		string(common.NodeTypeStart),
		map[string]interface{}{},
		false,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create registration start node: %w", err)
	}

	// Set executor name for the registration start node
	if executableNode, ok := regStartNode.(core.ExecutorBackedNodeInterface); ok {
		executableNode.SetExecutorName(executor.ExecutorNameUserTypeResolver)
	} else {
		return nil, fmt.Errorf("registration start node does not implement ExecutorBackedNodeInterface")
	}

	return regStartNode, nil
}

// validateDefaultFlowConfigs validates the default flow configurations
func (s *flowMgtService) validateDefaultFlowConfigs() error {
	flowConfig := config.GetThunderRuntime().Config.Flow

	// Validate auth flow.
	if flowConfig.Authn.DefaultFlow == "" {
		return errors.New("default authentication flow is not configured")
	}
	if !s.IsValidGraphID(flowConfig.Authn.DefaultFlow) {
		return errors.New("default authentication flow graph ID is invalid")
	}

	return nil
}
