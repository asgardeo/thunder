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

	"github.com/asgardeo/thunder/internal/flow/graph"
	"github.com/asgardeo/thunder/internal/flow/model"
)

// ConvertFlowModelToGraph converts a flow model to a graph representation.
func ConvertFlowModelToGraph(flowModel *model.Flow) (*graph.Graph, error) {
	if flowModel == nil || len(flowModel.Steps) == 0 {
		return nil, fmt.Errorf("flow model is nil or has no steps")
	}

	// Assume the first step is the start step
	startNodeID := flowModel.Steps[0].ID
	g := &graph.Graph{
		ID:          flowModel.ID,
		Nodes:       make(map[string]graph.Node),
		Edges:       make(map[string][]string),
		StartNodeID: startNodeID,
	}

	// Convert steps to nodes and add them to the graph
	for i, step := range flowModel.Steps {
		// Create a new node
		isStartNode := (i == 0)
		node := graph.Node{
			ID:          step.ID,
			Type:        step.Type,
			IsStartNode: isStartNode,
			Page:        step,
		}

		// If this is not the first step, set previous node ID
		if i > 0 {
			node.PreviousNodeID = flowModel.Steps[i-1].ID
		}

		// If this is not the last step, set next node ID
		if i < len(flowModel.Steps)-1 {
			node.NextNodeID = flowModel.Steps[i+1].ID
		}

		// Add the node to the graph
		g.AddNode(node)

		// If this is not the first step, create an edge from the previous step to this one
		if i > 0 {
			g.AddEdge(node.PreviousNodeID, node.ID)
		}
	}

	return g, nil
}
