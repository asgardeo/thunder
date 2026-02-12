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

// Package declarativeresource provides a generic framework for managing declarative resources.
package declarativeresource

// ResourceConfig defines the configuration for a declarative resource type.
type ResourceConfig struct {
	// ResourceType is the name of the resource type (e.g., "IdentityProvider", "Application")
	ResourceType string

	// DirectoryName is the directory name under declarative_resources/ (e.g., "identity_providers")
	DirectoryName string

	// Parser converts YAML bytes to a DTO
	// Returns the parsed DTO or an error
	Parser func([]byte) (interface{}, error)

	// Validator validates the parsed DTO (optional)
	// Returns an error if validation fails
	Validator func(interface{}) error

	// DependencyValidator validates dependencies (optional)
	// This is called after basic validation to check cross-resource dependencies
	DependencyValidator func(interface{}) error

	// IDExtractor extracts the ID from the DTO
	IDExtractor func(interface{}) string
}
