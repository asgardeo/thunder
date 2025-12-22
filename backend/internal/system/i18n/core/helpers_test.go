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

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefault(t *testing.T) {
	// Setup: Mock DefaultMessages
	origDefaults := defaultMessages
	defer func() { defaultMessages = origDefaults }()

	testKey := "test.key"
	testValue := "Test Value"
	defaultMessages = map[string]string{
		testKey: testValue,
	}

	t.Run("Existing Key", func(t *testing.T) {
		val, ok := GetDefault(testKey)
		assert.True(t, ok)
		assert.Equal(t, testValue, val)
	})

	t.Run("Non-Existing Key", func(t *testing.T) {
		val, ok := GetDefault("non.existent.key")
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}

func TestGetAllDefaults(t *testing.T) {
	// Setup: Mock DefaultMessages
	origDefaults := defaultMessages
	defer func() { defaultMessages = origDefaults }()

	defaultMessages = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	defaults := GetAllDefaults()

	assert.Equal(t, len(defaultMessages), len(defaults))

	// Verify content match
	for k, v := range defaultMessages {
		assert.Equal(t, v, defaults[k])
	}

	// Verify it is a copy
	defaults["new_temp_key"] = "temp_value"
	_, ok := defaultMessages["new_temp_key"]
	assert.False(t, ok, "GetAllDefaults should return a copy, not reference")
}

func TestGetAllKeys(t *testing.T) {
	// Setup: Mock DefaultMessages
	origDefaults := defaultMessages
	defer func() { defaultMessages = origDefaults }()

	defaultMessages = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	keys := GetAllKeys()

	assert.Equal(t, len(defaultMessages), len(keys))

	// Verify all keys are present
	for _, k := range keys {
		_, ok := defaultMessages[k]
		assert.True(t, ok, "Key returned by GetAllKeys should exist in DefaultMessages")
	}
}
