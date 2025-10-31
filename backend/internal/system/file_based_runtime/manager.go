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

// Package filebasedruntime provides functionality to read file-based runtime configurations.
package filebasedruntime

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

// envVarPatterns defines the environment variable patterns in order of precedence (most specific first)
var envVarPatterns = []struct {
	regex *regexp.Regexp
	name  string
}{
	{
		regex: regexp.MustCompile(`\$\{\{([A-Za-z_][A-Za-z0-9_]*)\}\}`),
		name:  "double-brace",
	},
	{
		regex: regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`),
		name:  "single-brace",
	},
	{
		regex: regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`),
		name:  "direct",
	},
}

// substituteEnvironmentVariables replaces environment variable placeholders in the given content.
//
// Supported patterns (in order of precedence):
//  1. ${{VAR}}   - Double-brace syntax (highest precedence)
//  2. ${VAR}     - Single-brace syntax
//  3. $VAR       - Direct variable syntax (lowest precedence)
//
// If an environment variable is not set, an error is returned.
func substituteEnvironmentVariables(content []byte) ([]byte, error) {
	contentStr := string(content)

	// Process each pattern
	for _, pattern := range envVarPatterns {
		matches := pattern.regex.FindAllStringSubmatch(contentStr, -1)
		for _, match := range matches {
			if len(match) != 2 {
				continue
			}

			fullMatch := match[0]
			varName := match[1]

			envValue, exists := os.LookupEnv(varName)
			if !exists {
				return nil, fmt.Errorf("environment variable '%s' is not set", varName)
			}

			contentStr = strings.ReplaceAll(contentStr, fullMatch, envValue)
		}
	}

	return []byte(contentStr), nil
}

// GetConfigs reads all configuration files from the specified directory within the immutable_resources directory.
func GetConfigs(configDirectoryPath string) ([][]byte, error) {
	logger := log.GetLogger().With(log.String("component", "FileBasedRuntime"))
	thunderHome := config.GetThunderRuntime().ThunderHome
	immutableConfigFilePath := path.Join(thunderHome, "repository/conf/immutable_resources/")
	absoluteDirectoryPath := filepath.Join(immutableConfigFilePath, configDirectoryPath)
	files, err := os.ReadDir(absoluteDirectoryPath)
	if err != nil {
		logger.Error("Failed to read configuration directory",
			log.String("path", absoluteDirectoryPath), log.Error(err))
		return nil, err
	}

	// Count non-directory files
	var fileCount int
	for _, file := range files {
		if !file.IsDir() {
			fileCount++
		}
	}

	configs := make([][]byte, 0, fileCount)
	if fileCount == 0 {
		return configs, nil
	}

	// Use channels to collect results from goroutines
	type configResult struct {
		content []byte
		err     error
	}
	configChan := make(chan configResult)
	var wg sync.WaitGroup

	for _, file := range files {
		if !file.IsDir() {
			wg.Add(1)
			go func(fileName string) {
				defer wg.Done()
				filePath := filepath.Join(absoluteDirectoryPath, fileName)
				filePath = filepath.Clean(filePath)
				// #nosec G304 -- File path is controlled and within a trusted directory
				fileContent, err := os.ReadFile(filePath)
				if err != nil {
					logger.Warn("Failed to read configuration file", log.String("filePath", fileName), log.Error(err))
					configChan <- configResult{content: nil, err: err}
					return
				}
				// Substitute environment variables
				processedContent, err := substituteEnvironmentVariables(fileContent)
				if err != nil {
					logger.Warn("Failed to substitute environment variables in configuration file",
						log.String("filePath", fileName), log.Error(err))
					configChan <- configResult{content: nil, err: err}
					return
				}

				configChan <- configResult{content: processedContent, err: nil}
			}(file.Name())
		}
	}

	// Wait for all goroutines to complete and close the channel
	go func() {
		wg.Wait()
		close(configChan)
	}()

	// Collect results from the channel
	var errors []error
	for result := range configChan {
		if result.err != nil {
			errors = append(errors, result.err)
			continue
		}
		configs = append(configs, result.content)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("errors occurred while reading configuration files: %v", errors)
	}

	return configs, nil
}
