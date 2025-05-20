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

// Package engine provides the flow engine for orchestrating flow executions.
package engine

import (
	"encoding/json"
	"fmt"
	syslog "log"
	"os"
	"path/filepath"
	"strings"

	"github.com/asgardeo/thunder/internal/flow/graph"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/flow/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

// FlowEngineInterface defines the interface for the flow engine.
type FlowEngineInterface interface {
	init() error
	RegisterGraph(graphID string, g *graph.Graph)
	GetGraph(graphID string) (*graph.Graph, bool)
	Execute(ctx *model.FlowContext) (*model.FlowStep, error)
}

// FlowEngine is the main engine implementation for orchestrating flow executions.
type FlowEngine struct {
	graphs map[string]*graph.Graph
}

// NewFlowEngine creates a new FlowEngine.
func NewFlowEngine() FlowEngineInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowEngine"))

	engine := &FlowEngine{
		graphs: make(map[string]*graph.Graph),
	}
	if err := engine.init(); err != nil {
		logger.Fatal("Failed to initialize FlowEngine: %v", log.Error(err))
	}

	return engine
}

// init initializes the FlowEngine by loading graph configurations into runtime.
func (e *FlowEngine) init() error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowEngine"))
	logger.Info("Initializing FlowEngine")

	configDir := config.GetThunderRuntime().Config.Authenticator.GraphDirectory
	if configDir == "" {
		logger.Info("Graph directory is not set. No graphs will be loaded.")
		return nil
	}

	logger.Debug("Loading graphs from %s", log.String("configDir", configDir))

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
	logger.Debug("Found %d files in the graph directory", log.Int("fileCount", len(files)))

	// Process each JSON file in the directory
	for _, file := range files {
		// Skip directories and non-JSON files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			logger.Debug("Skipping non-JSON file: %s", log.String("fileName", file.Name()))
			continue
		}
		filePath := filepath.Join(configDir, file.Name())
		filePath = filepath.Clean(filePath)

		// Read the file content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("Failed to read graph file %s: %v", log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Parse the JSON into the flow model
		var flowModel model.Flow
		if err := json.Unmarshal(fileContent, &flowModel); err != nil {
			logger.Warn("Failed to parse JSON in file %s: %v", log.String("filePath", filePath), log.Error(err))
			continue
		}

		// TODO: Temporarily print the flow model for debugging.
		syslog.Println("====================================================")
		syslog.Println("----- Flow Model -----")
		syslog.Printf("%+v", flowModel)
		syslog.Println("====================================================")

		// Convert the flow model to a graph
		graphModel, err := utils.ConvertFlowModelToGraph(&flowModel)
		if err != nil {
			logger.Warn("Failed to convert flow model to graph for file %s: %v",
				log.String("filePath", filePath), log.Error(err))
			continue
		}

		// TODO: Temporarily print the graph model for debugging.
		syslog.Println("====================================================")
		syslog.Println("----- Graph Model -----")
		syslog.Printf("%+v", graphModel)
		syslog.Println("====================================================")

		// Register the graph with the flow engine
		logger.Debug("Registering graph with ID %s", log.String("graphID", graphModel.ID))
		e.RegisterGraph(graphModel.ID, graphModel)
	}

	return nil
}

// RegisterGraph registers a graph with the flow engine
func (e *FlowEngine) RegisterGraph(graphID string, g *graph.Graph) {
	e.graphs[graphID] = g
}

// GetGraph retrieves a graph by its ID
func (e *FlowEngine) GetGraph(graphID string) (*graph.Graph, bool) {
	g, ok := e.graphs[graphID]
	return g, ok
}

// Execute executes a step in the flow
func (e *FlowEngine) Execute(ctx *model.FlowContext) (*model.FlowStep, error) {
	// TODO: Implement the execution logic for the flow engine
	return nil, nil
}
