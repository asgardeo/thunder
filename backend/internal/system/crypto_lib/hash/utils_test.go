/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package hash

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HashUtilsTestSuite struct {
	suite.Suite
}

func TestHashUtilsSuite(t *testing.T) {
	suite.Run(t, new(HashUtilsTestSuite))
}

func (suite *HashUtilsTestSuite) TestThumbprint() {
	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "EmptyInput",
			input:    []byte(""),
			expected: "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
		},
		{
			name:     "NormalInput",
			input:    []byte("hello world"),
			expected: "uU0nuZNNPgilLlLX2n2r+sSE7+N6U4DukIj3rOLvzek=",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			hash := GenerateThumbprint(tc.input)
			suite.Equal(tc.expected, hash, "Hash should match expected value")
		})
	}
}

func (suite *HashUtilsTestSuite) TestThumbprintString() {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "EmptyString",
			input:    "",
			expected: "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
		},
		{
			name:     "NormalString",
			input:    "hello world",
			expected: "uU0nuZNNPgilLlLX2n2r+sSE7+N6U4DukIj3rOLvzek=",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			hash := GenerateThumbprintFromString(tc.input)
			suite.Equal(tc.expected, hash, "Hash should match expected value")
		})
	}
}

func (suite *HashUtilsTestSuite) TestHash() {
	data := []byte("hello world")

	testCases := []struct {
		name string
		alg  HashAlgorithm
	}{
		{"SHA256", GenericSHA256},
		{"SHA384", GenericSHA384},
		{"SHA512", GenericSHA512},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			hashed, err := Hash(data, tc.alg)
			suite.NoError(err)
			suite.NotEmpty(hashed)
		})
	}

	_, err := Hash(data, "INVALID")
	suite.Error(err)
}

func (suite *HashUtilsTestSuite) TestGetHash() {
	testCases := []struct {
		name string
		alg  HashAlgorithm
	}{
		{"SHA256", GenericSHA256},
		{"SHA384", GenericSHA384},
		{"SHA512", GenericSHA512},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			h, err := GetHash(tc.alg)
			suite.NoError(err)
			suite.NotNil(h)
		})
	}

	_, err := GetHash("INVALID")
	suite.Error(err)
}
