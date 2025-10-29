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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SliceUtilTestSuite struct {
	suite.Suite
}

func TestSliceUtilTestSuite(t *testing.T) {
	suite.Run(t, new(SliceUtilTestSuite))
}

func (suite *SliceUtilTestSuite) TestUniqueStrings() {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "No duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "With duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "All duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			name:     "Single element",
			input:    []string{"a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := UniqueStrings(tt.input)

			if tt.expected == nil {
				assert.Nil(suite.T(), result)
				return
			}

			// Check length matches
			assert.Equal(suite.T(), len(tt.expected), len(result))

			// Convert result to map for order-independent comparison
			resultMap := make(map[string]bool)
			for _, v := range result {
				resultMap[v] = true
			}

			// Verify all expected values are present
			for _, v := range tt.expected {
				assert.True(suite.T(), resultMap[v], "Expected value %s not found in result", v)
			}
		})
	}
}

func (suite *SliceUtilTestSuite) TestDeepCopyMapOfStrings() {
	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{
			name:     "Nil map",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Empty map",
			input:    map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "Map with values",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := DeepCopyMapOfStrings(tt.input)

			if tt.expected == nil {
				assert.Nil(suite.T(), result)
				return
			}

			assert.Equal(suite.T(), tt.expected, result)

			// Verify it's a deep copy (modifying original doesn't affect copy)
			if len(tt.input) > 0 {
				for k := range tt.input {
					tt.input[k] = "modified"
					assert.NotEqual(suite.T(), "modified", result[k])
					break
				}
			}
		})
	}
}

func (suite *SliceUtilTestSuite) TestDeepCopyMapOfStringSlices() {
	tests := []struct {
		name     string
		input    map[string][]string
		expected map[string][]string
	}{
		{
			name:     "Nil map",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Empty map",
			input:    map[string][]string{},
			expected: map[string][]string{},
		},
		{
			name: "Map with values",
			input: map[string][]string{
				"key1": {"value1", "value2"},
				"key2": {"value3"},
			},
			expected: map[string][]string{
				"key1": {"value1", "value2"},
				"key2": {"value3"},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := DeepCopyMapOfStringSlices(tt.input)

			if tt.expected == nil {
				assert.Nil(suite.T(), result)
				return
			}

			assert.Equal(suite.T(), tt.expected, result)

			// Verify it's a deep copy (modifying original doesn't affect copy)
			if len(tt.input) > 0 {
				for k := range tt.input {
					if len(tt.input[k]) > 0 {
						tt.input[k][0] = "modified"
						assert.NotEqual(suite.T(), "modified", result[k][0])
					}
					break
				}
			}
		})
	}
}

func (suite *SliceUtilTestSuite) TestMergeInterfaceMaps() {
	tests := []struct {
		name     string
		dst      map[string]interface{}
		src      map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "Both nil",
			dst:      nil,
			src:      nil,
			expected: map[string]interface{}{},
		},
		{
			name:     "Dst nil, src with values",
			dst:      nil,
			src:      map[string]interface{}{"key1": "value1"},
			expected: map[string]interface{}{"key1": "value1"},
		},
		{
			name:     "Dst with values, src nil",
			dst:      map[string]interface{}{"key1": "value1"},
			src:      nil,
			expected: map[string]interface{}{"key1": "value1"},
		},
		{
			name:     "Both with non-overlapping keys",
			dst:      map[string]interface{}{"key1": "value1"},
			src:      map[string]interface{}{"key2": "value2"},
			expected: map[string]interface{}{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "Both with overlapping keys - src overrides dst",
			dst:      map[string]interface{}{"key1": "value1", "key2": "value2"},
			src:      map[string]interface{}{"key2": "newValue2", "key3": "value3"},
			expected: map[string]interface{}{"key1": "value1", "key2": "newValue2", "key3": "value3"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := MergeInterfaceMaps(tt.dst, tt.src)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}
