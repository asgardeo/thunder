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

// Package utils provides utility functions for flow processing.
package utils

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/flow/model"
)

// ConvertFlowModelToGraph converts a flow model to a graph representation.
func ConvertFlowModelToGraph(flowModel *model.Flow) (*model.Graph, error) {
	if flowModel == nil || len(flowModel.Steps) == 0 {
		return nil, fmt.Errorf("flow model is nil or has no steps")
	}

	// Create a graph
	g := &model.Graph{
		ID:    flowModel.ID,
		Nodes: make(map[string]model.Node),
		Edges: make(map[string][]string),
	}

	// Map to track which steps are pointed to by others (have incoming references)
	hasIncomingRef := make(map[string]bool)

	// First pass: identify steps that are pointed to by ACTION components
	for _, step := range flowModel.Steps {
		for _, component := range step.Data.Components {
			// Recursively check all components and their nested components
			findNextReferences(component, hasIncomingRef)
		}
	}

	// Find the start node (the one without incoming references)
	startNodeID := ""
	for _, step := range flowModel.Steps {
		if !hasIncomingRef[step.ID] {
			startNodeID = step.ID
			break
		}
	}

	// If no start node found, fallback to the first step
	if startNodeID == "" {
		startNodeID = flowModel.Steps[0].ID
	}

	g.StartNodeID = startNodeID

	// Second pass: create nodes and build connections
	for _, step := range flowModel.Steps {
		// Create a new node
		isStartNode := (step.ID == startNodeID)

		// Mark AUTHENTICATION_SUCCESS type as final node
		isFinalNode := (step.Type == "AUTHENTICATION_SUCCESS")

		node := model.Node{
			ID:          step.ID,
			Type:        step.Type,
			IsStartNode: isStartNode,
			IsFinalNode: isFinalNode,
			Page:        step,
		}

		// Add the node to the graph
		g.AddNode(node)

		// Create edges based on ACTION component references
		for _, component := range step.Data.Components {
			addEdgesFromComponent(component, step.ID, g)
		}
	}

	// Third pass: set PreviousNodeID and NextNodeID based on edges
	for fromNodeID, toNodeIDs := range g.Edges {
		if len(toNodeIDs) > 0 {
			// Set the NextNodeID for the source node (using the first target as next)
			if sourceNode, exists := g.Nodes[fromNodeID]; exists {
				sourceNode.NextNodeID = toNodeIDs[0] // Use the first edge as the default next
				g.Nodes[fromNodeID] = sourceNode     // Update the source node
			}

			// Set the PreviousNodeID for each target node
			for _, toNodeID := range toNodeIDs {
				if targetNode, exists := g.Nodes[toNodeID]; exists {
					targetNode.PreviousNodeID = fromNodeID
					g.Nodes[toNodeID] = targetNode // Update the target node
				}
			}
		}
	}

	return g, nil
}

// findNextReferences recursively checks components for action.next references
// and marks the referenced steps as having incoming references
func findNextReferences(component model.Component, hasIncomingRef map[string]bool) {
	if component.Category == "ACTION" {
		// First check if action is directly defined in the component
		if component.Action != nil && component.Action.Next != "" {
			hasIncomingRef[component.Action.Next] = true
		} else {
			// Otherwise check if action is in config
			if nextID, exists := getNextStepID(component.Config); exists && nextID != "" {
				hasIncomingRef[nextID] = true
			}
		}
	}

	// Recursively check nested components
	for _, nestedComp := range component.Components {
		findNextReferences(nestedComp, hasIncomingRef)
	}
}

// addEdgesFromComponent adds graph edges based on component action.next references
func addEdgesFromComponent(component model.Component, sourceStepID string, g *model.Graph) {
	if component.Category == "ACTION" {
		// First check if action is directly defined in the component
		if component.Action != nil && component.Action.Next != "" {
			nextID := component.Action.Next
			g.AddEdge(sourceStepID, nextID)
		} else {
			// Otherwise check if action is in config
			if nextID, exists := getNextStepID(component.Config); exists && nextID != "" {
				g.AddEdge(sourceStepID, nextID)
			}
		}
	}

	// Recursively process nested components
	for _, nestedComp := range component.Components {
		addEdgesFromComponent(nestedComp, sourceStepID, g)
	}
}

// getNextStepID extracts the next step ID from component configuration
func getNextStepID(config map[string]interface{}) (string, bool) {
	// Check for action configuration
	actionValue, actionExists := config["action"]
	if !actionExists {
		return "", false
	}

	// Check if action is a map
	actionMap, isMap := actionValue.(map[string]interface{})
	if !isMap {
		return "", false
	}

	// Get the next property
	nextValue, nextExists := actionMap["next"]
	if !nextExists {
		return "", false
	}

	// Convert to string
	nextID, isString := nextValue.(string)
	if !isString {
		return "", false
	}

	return nextID, true
}
