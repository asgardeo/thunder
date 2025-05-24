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

// Package composer provides the flow composer for managing flow graphs.
package composer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	syslog "log"

	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/flow/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

var (
	instance *FlowComposer
	once     sync.Once
)

// FlowComposerInterface defines the flow composer that manages the flow graphs.
type FlowComposerInterface interface {
	Init() error
	RegisterGraph(graphID string, g *model.Graph)
	GetGraph(graphID string) (*model.Graph, bool)
}

// FlowComposer is the implementation of FlowComposerInterface.
type FlowComposer struct {
	graphs map[string]*model.Graph
}

// GetFlowComposer returns a singleton instance of FlowComposer.
func GetFlowComposer() FlowComposerInterface {
	once.Do(func() {
		instance = &FlowComposer{
			graphs: make(map[string]*model.Graph),
		}
	})
	return instance
}

// Init initializes the FlowComposer by loading graph configurations into runtime.
func (c *FlowComposer) Init() error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowComposer"))
	logger.Info("Initializing the flow composer")

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
		syslog.Println("----- Graph Model JSON -----")
		jsonGraph, err := graphModel.ToJSON()
		if err != nil {
			logger.Warn("Failed to convert graph model to JSON for file %s: %v",
				log.String("filePath", filePath), log.Error(err))
		} else {
			syslog.Printf("%s", jsonGraph)
		}
		syslog.Println("====================================================")

		// Register the graph with the flow composer
		logger.Debug("Registering graph with ID %s", log.String("graphID", graphModel.ID))
		c.RegisterGraph(graphModel.ID, graphModel)
	}

	return nil
}

// RegisterGraph registers a graph with the flow composer
func (c *FlowComposer) RegisterGraph(graphID string, g *model.Graph) {
	c.graphs[graphID] = g
}

// GetGraph retrieves a graph by its ID
func (c *FlowComposer) GetGraph(graphID string) (*model.Graph, bool) {
	g, ok := c.graphs[graphID]
	return g, ok
}
