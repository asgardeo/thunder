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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SessionUtilsTestSuite struct {
	suite.Suite
}

func TestSessionUtilsSuite(t *testing.T) {
	suite.Run(t, new(SessionUtilsTestSuite))
}

func (suite *SessionUtilsTestSuite) TestGenerateNewSessionDataKey() {
	// Generate a session data key
	sessionKey := GenerateNewSessionDataKey()
	
	// Should not be empty
	assert.NotEmpty(suite.T(), sessionKey)
	
	// Should be a valid UUID format (36 characters with hyphens)
	assert.Len(suite.T(), sessionKey, 36)
	assert.Contains(suite.T(), sessionKey, "-")
}

func (suite *SessionUtilsTestSuite) TestGenerateNewSessionDataKeyUniqueness() {
	// Generate multiple session keys
	keys := make(map[string]bool)
	numKeys := 100
	
	for i := 0; i < numKeys; i++ {
		key := GenerateNewSessionDataKey()
		
		// Check that each key is unique
		assert.False(suite.T(), keys[key], "Duplicate session key generated: %s", key)
		keys[key] = true
		
		// Verify it's not empty
		assert.NotEmpty(suite.T(), key)
	}
	
	// Should have generated all unique keys
	assert.Len(suite.T(), keys, numKeys)
}

func (suite *SessionUtilsTestSuite) TestGenerateNewSessionDataKeyFormat() {
	sessionKey := GenerateNewSessionDataKey()
	
	// Verify UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	// Total length should be 36 characters
	assert.Len(suite.T(), sessionKey, 36)
	
	// Check hyphen positions (8, 13, 18, 23)
	assert.Equal(suite.T(), "-", string(sessionKey[8]))
	assert.Equal(suite.T(), "-", string(sessionKey[13]))
	assert.Equal(suite.T(), "-", string(sessionKey[18]))
	assert.Equal(suite.T(), "-", string(sessionKey[23]))
}

func (suite *SessionUtilsTestSuite) TestGenerateNewSessionDataKeyConsistentLength() {
	// Test that all generated keys have consistent length
	for i := 0; i < 10; i++ {
		key := GenerateNewSessionDataKey()
		assert.Len(suite.T(), key, 36, "Session key should always be 36 characters")
	}
}
