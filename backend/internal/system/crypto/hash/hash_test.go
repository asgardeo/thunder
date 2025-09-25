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

type HashUtilTestSuite struct {
	suite.Suite
}

func TestHashSuite(t *testing.T) {
	suite.Run(t, new(HashUtilTestSuite))
}

func (suite *HashUtilTestSuite) TestVerifySha256() {
	testCases := []struct {
		name     string
		input    string
		expected Credential
	}{
		{
			name:  "EmptyStringAndSalt",
			input: "",
			expected: Credential{
				Algorithm: "SHA256",
				Hash:      "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				Salt:      "",
			},
		},
		{
			name:  "NormalStringWithoutSalt",
			input: "password",
			expected: Credential{
				Algorithm: "SHA256",
				Hash:      "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
				Salt:      "",
			},
		},
		{
			name:  "NormalStringWithSalt",
			input: "password",
			expected: Credential{
				Algorithm: "SHA256",
				Hash:      "4b2dcea502b405a479a69fd2478ea891fa9f02966db9ee5cbcbee53137c8ae4d",
				Salt:      "12f4576d7432bd8020db7202b6492a37",
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			hash := verifySHA256Credential([]byte(tc.input), tc.expected)

			assert.True(t, hash)
		})
	}
}

func (suite *HashUtilTestSuite) TestSha256HashAndVerify() {
	input := "test-input"
	cred := newSHA256Credential([]byte(input))

	assert.True(suite.T(), Verify([]byte(input), cred),
		"Hash verification should succeed for the same input")
}

func (suite *HashUtilTestSuite) TestSha256HashWithDifferentInputs() {
	input1 := "input-one"
	input2 := "input-two"
	cred1 := newSHA256Credential([]byte(input1))
	cred2 := newSHA256Credential([]byte(input2))

	assert.NotEqual(suite.T(), cred1.Hash, cred2.Hash, "Different inputs should produce different hashes")
}

func (suite *HashUtilTestSuite) TestSha256HashWithDifferentSalts() {
	input := "common-input"
	cred1 := newSHA256Credential([]byte(input))
	cred2 := newSHA256Credential([]byte(input))

	assert.NotEqual(suite.T(), cred1.Hash, cred2.Hash, "Different salts should produce different hashes")
}

func (suite *HashUtilTestSuite) TestVerifyBKDF2() {
	testCases := []struct {
		name     string
		input    string
		expected Credential
	}{
		{
			name:  "EmptyStringAndSalt",
			input: "",
			expected: Credential{
				Algorithm: "PBKDF2",
				Hash:      "3106cb5743a54114a36bb7d3b2afa0242360b58243264728a9ca208548082281",
				Salt:      "",
			},
		},
		{
			name:  "NormalStringWithoutSalt",
			input: "password",
			expected: Credential{
				Algorithm: "PBKDF2",
				Hash:      "fdc25be00b18ba5c79d8bf7a452d98c248b11f2c7e9c871d24f1f880381e95cf",
				Salt:      "",
			},
		},
		{
			name:  "NormalStringWithSalt",
			input: "password",
			expected: Credential{
				Algorithm: "PBKDF2",
				Hash:      "b500f5369698b4bcdde08267c406c12ff95e8de1d431e4472bf6ea95b620da5c",
				Salt:      "36d2dde7dfbafe8e04ea49450f659b1c",
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			hash := verifyPBKDF2Credential([]byte(tc.input), tc.expected)

			assert.True(t, hash)
		})
	}
}

func (suite *HashUtilTestSuite) TestPBKDF2HashWithAndVerify() {
	input := "test-input"
	cred := newPBKDF2Credential([]byte(input))

	assert.True(suite.T(), Verify([]byte(input), cred),
		"Hash verification should succeed for the same input")
}

func (suite *HashUtilTestSuite) TestPBKDF2HashWithDifferentInputs() {
	input1 := "input-one"
	input2 := "input-two"
	hash1 := newPBKDF2Credential([]byte(input1))
	hash2 := newPBKDF2Credential([]byte(input2))

	assert.NotEqual(suite.T(), hash1, hash2, "Different inputs should produce different hashes")
}

func (suite *HashUtilTestSuite) TestPBKDF2HashWithDifferentSalts() {
	input := "common-input"
	hash1 := newPBKDF2Credential([]byte(input))
	hash2 := newPBKDF2Credential([]byte(input))
	assert.NotEqual(suite.T(), hash1, hash2, "Different salts should produce different hashes")
}

func (suite *HashUtilTestSuite) TestThumbprint() {
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

func (suite *HashUtilTestSuite) TestThumbprintString() {
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

func (suite *HashUtilTestSuite) TestGenerateSalt() {
	salt, err := generateSalt()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), salt)
}

func (suite *HashUtilTestSuite) TestGenerateSaltUniqueness() {
	salt1, err1 := generateSalt()
	salt2, err2 := generateSalt()

	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)
	assert.NotEqual(suite.T(), salt1, salt2, "Generated salts should be different")
}

func (suite *HashUtilTestSuite) TestGenerateSaltLength() {
	salt, err := generateSalt()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 16, len(salt), "Generated salt should be 16 bytes")
}
