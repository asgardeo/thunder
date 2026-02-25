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

package model

import (
	"encoding/json"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/log"
)

// JSON Schema type constants.
const (
	// TypeString represents the string type in JSON Schema.
	TypeString = "string"
	// TypeNumber represents the number type in JSON Schema.
	TypeNumber = "number"
	// TypeBoolean represents the boolean type in JSON Schema.
	TypeBoolean = "boolean"
	// TypeObject represents the object type in JSON Schema.
	TypeObject = "object"
	// TypeArray represents the array type in JSON Schema.
	TypeArray = "array"
)

type property interface {
	isRequired() bool
	isCredential() bool
	validateValue(value interface{}, path string, logger *log.Logger) (bool, error)
	validateUniqueness(value interface{}, path string,
		identifyUser func(map[string]interface{}) (*string, error), logger *log.Logger) (bool, error)
}

// Schema represents a user schema with a set of properties.
type Schema struct {
	properties map[string]property
}

// GetCredentialAttributes returns the names of top-level properties marked as credentials.
func (cs *Schema) GetCredentialAttributes() []string {
	var fields []string
	for name, prop := range cs.properties {
		if prop.isCredential() {
			fields = append(fields, name)
		}
	}

	return fields
}

// Validate validates the user attributes against the schema.
func (cs *Schema) Validate(attributes json.RawMessage, logger *log.Logger) (bool, error) {
	if len(attributes) == 0 {
		logger.Debug("User has no attributes to validate")
		return true, nil
	}

	var userAttrs map[string]interface{}
	if err := json.Unmarshal(attributes, &userAttrs); err != nil {
		return false, fmt.Errorf("failed to unmarshal user attributes: %w", err)
	}

	if len(cs.properties) == 0 {
		return true, nil
	}

	for propName, prop := range cs.properties {
		value, exists := userAttrs[propName]
		if !exists {
			if prop.isRequired() {
				return false, nil
			}
			continue
		}

		isValid, err := prop.validateValue(value, propName, logger)
		if err != nil {
			return false, err
		}
		if !isValid {
			return false, nil
		}
	}

	return true, nil
}

// ValidateUniqueness checks uniqueness constraints for the schema properties.
func (cs *Schema) ValidateUniqueness(
	attrs map[string]interface{},
	identifyUser func(map[string]interface{}) (*string, error),
	logger *log.Logger,
) (bool, error) {
	if len(cs.properties) == 0 {
		return true, nil
	}

	for propName, prop := range cs.properties {
		value, exists := attrs[propName]
		if !exists {
			continue
		}

		isValid, err := prop.validateUniqueness(value, propName, identifyUser, logger)
		if err != nil {
			return false, err
		}
		if !isValid {
			return false, nil
		}
	}

	return true, nil
}

func convertToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// CompileUserSchema compiles a user schema from the provided JSON.
func CompileUserSchema(schema json.RawMessage) (*Schema, error) {
	var schemaMap map[string]json.RawMessage
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if len(schemaMap) == 0 {
		return nil, fmt.Errorf("schema cannot be empty - must contain at least one property definition")
	}

	compiled := &Schema{
		properties: make(map[string]property, len(schemaMap)),
	}

	for propName, propRaw := range schemaMap {
		compiledProp, err := compileProperty(propName, propRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid property '%s': %w", propName, err)
		}
		compiled.properties[propName] = compiledProp
	}

	return compiled, nil
}

func compileProperty(propName string, propRaw json.RawMessage) (property, error) {
	var propMap map[string]json.RawMessage
	if err := json.Unmarshal(propRaw, &propMap); err != nil {
		return nil, fmt.Errorf("property definition must be an object")
	}

	typeRaw, exists := propMap["type"]
	if !exists {
		return nil, fmt.Errorf("missing required 'type' field")
	}

	var typeStr string
	if err := json.Unmarshal(typeRaw, &typeStr); err != nil {
		return nil, fmt.Errorf("'type' field must be a string")
	}

	switch typeStr {
	case TypeString:
		return compileStringProperty(propMap)
	case TypeNumber:
		return compileNumberProperty(propMap)
	case TypeBoolean:
		return compileBooleanProperty(propMap)
	case TypeObject:
		return compileObjectProperty(propMap)
	case TypeArray:
		return compileArrayProperty(propName, propMap)
	default:
		return nil, fmt.Errorf("invalid type '%s', must be one of: string, number, boolean, object, array", typeStr)
	}
}
