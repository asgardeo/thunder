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

type number struct {
	required bool
	unique   bool
	enum     map[float64]struct{}
}

func (p *number) isRequired() bool {
	return p.required
}

func (p *number) isIndexed() bool {
	// Booleans are never indexed
	return false
}

func (p *number) validateValue(value interface{}, path string, logger *log.Logger) (bool, error) {
	numberValue, ok := convertToFloat64(value)
	if !ok {
		logger.Debug("Expected number but got different type",
			log.String("property", path), log.String("value", fmt.Sprintf("%v", value)))
		return false, nil
	}

	if p.enum != nil {
		if _, exists := p.enum[numberValue]; !exists {
			logger.Debug("Value not in enum", log.String("property", path), log.String("value", fmt.Sprintf("%v", value)))
			return false, nil
		}
	}

	return true, nil
}

func (p *number) validateUniqueness(
	value interface{},
	path string,
	identifyUser func(map[string]interface{}) (*string, error),
	logger *log.Logger,
) (bool, error) {
	if !p.unique {
		return true, nil
	}

	filter := map[string]interface{}{path: value}
	existingUserID, err := identifyUser(filter)
	if err != nil {
		return false, err
	}

	return existingUserID == nil, nil
}

func compileNumberProperty(propMap map[string]json.RawMessage) (property, error) {
	allowedFields := map[string]struct{}{
		"type":     {},
		"required": {},
		"unique":   {},
		"enum":     {},
	}

	for field := range propMap {
		if _, ok := allowedFields[field]; !ok {
			return nil, fmt.Errorf("invalid field '%s' for leaf property", field)
		}
	}

	prop := &number{}

	if raw, exists := propMap["required"]; exists {
		if err := json.Unmarshal(raw, &prop.required); err != nil {
			return nil, fmt.Errorf("'required' field must be a boolean")
		}
	}

	if raw, exists := propMap["unique"]; exists {
		if err := json.Unmarshal(raw, &prop.unique); err != nil {
			return nil, fmt.Errorf("'unique' field must be a boolean")
		}
	}

	if raw, exists := propMap["enum"]; exists {
		var enumRaw []json.RawMessage
		if err := json.Unmarshal(raw, &enumRaw); err != nil {
			return nil, fmt.Errorf("'enum' field must be an array")
		}
		if len(enumRaw) == 0 {
			return nil, fmt.Errorf("'enum' array cannot be empty")
		}

		prop.enum = make(map[float64]struct{}, len(enumRaw))
		for i, itemRaw := range enumRaw {
			var value float64
			if err := json.Unmarshal(itemRaw, &value); err != nil {
				return nil, fmt.Errorf("'enum' array item at index %d must be a number to match property type", i)
			}
			prop.enum[value] = struct{}{}
		}
	}

	return prop, nil
}
