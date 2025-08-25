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

// Package hash provides generic hashing utilities for sensitive data.
package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HashProviderTestSuite struct {
	suite.Suite
}

func TestProviderHashSuite(t *testing.T) {
	suite.Run(t, new(HashProviderTestSuite))
}

func (suite *HashProviderTestSuite) TestSha256HashWithSalt() {
	testCases := []struct {
		name     string
		input    string
		salt     string
		expected string
	}{
		{
			name:     "EmptyStringAndSalt",
			input:    "",
			salt:     "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "NormalStringWithSalt",
			input:    "password",
			salt:     "somesalt",
			expected: "6bceb6d53d51a11c3bde77e8cafe1f152782c5e52a13e514da12a9e35b0c2bcb",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			hash := newSHA256HashProvider().Hash([]byte(tc.input), []byte(tc.salt))

			assert.Equal(t, tc.expected, hash)
		})
	}
}

func (suite *HashProviderTestSuite) TestSha256HashWithSaltDeterministic() {
	input := "test-input"
	salt := "test-salt"

	hash1 := newSHA256HashProvider().Hash([]byte(input), []byte(salt))
	hash2 := newSHA256HashProvider().Hash([]byte(input), []byte(salt))

	assert.Equal(suite.T(), hash1, hash2, "Hash should be deterministic for the same input and salt")
}

func (suite *HashProviderTestSuite) TestSha256HashWithDifferentInputs() {
	salt := "common-salt"
	input1 := "input-one"
	input2 := "input-two"

	hash1 := newSHA256HashProvider().Hash([]byte(input1), []byte(salt))
	hash2 := newSHA256HashProvider().Hash([]byte(input2), []byte(salt))

	assert.NotEqual(suite.T(), hash1, hash2, "Different inputs should produce different hashes")
}

func (suite *HashProviderTestSuite) TestSha256HashWithDifferentSalts() {
	input := "common-input"
	salt1 := "salt-one"
	salt2 := "salt-two"

	hash1 := newSHA256HashProvider().Hash([]byte(input), []byte(salt1))
	hash2 := newSHA256HashProvider().Hash([]byte(input), []byte(salt2))

	assert.NotEqual(suite.T(), hash1, hash2, "Different salts should produce different hashes")
}
