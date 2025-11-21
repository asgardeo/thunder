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

package hash

import (
	"testing"

	"github.com/asgardeo/thunder/internal/system/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HashUtilsTestSuite struct {
	suite.Suite
	input []byte
}

func TestHashUtilsSuite(t *testing.T) {
	suite.Run(t, new(HashUtilsTestSuite))
}

func (suite *HashUtilsTestSuite) SetupSuite() {
	suite.input = []byte("secretPassword123")
}

func (suite *HashUtilsTestSuite) TearDownSuite() {
	config.ResetThunderRuntime()
}

func (suite *HashUtilsTestSuite) TestGenerateSha256() {
	// Set runtime config to SHA256
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(SHA256),
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	cred := Generate(suite.input)

	assert.Equal(suite.T(), SHA256, cred.Algorithm, "Algorithm should be SHA256")
	assert.NotEmpty(suite.T(), cred.Hash, "Hash should not be empty")
}

func (suite *HashUtilsTestSuite) TestVerifySha256() {
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
			},
		},
		{
			name:  "NormalStringWithoutSalt",
			input: "password",
			expected: Credential{
				Algorithm: "SHA256",
				Hash:      "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
			},
		},
		{
			name:  "NormalStringWithSalt",
			input: "password",
			expected: Credential{
				Algorithm: "SHA256",
				Hash:      "4b2dcea502b405a479a69fd2478ea891fa9f02966db9ee5cbcbee53137c8ae4d",
				Parameters: CredParameters{
					Salt: "12f4576d7432bd8020db7202b6492a37",
				},
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			hash := Verify([]byte(tc.input), tc.expected)

			assert.True(t, hash)
		})
	}
}

func (suite *HashUtilsTestSuite) TestVerifySha256_Failure() {
	testCases := []struct {
		name     string
		input    string
		expected Credential
	}{
		{
			name:  "IncorrectHash",
			input: "password",
			expected: Credential{
				Algorithm: "SHA256",
				Hash:      "incorrecthashvalue",
			},
		},
		{
			name:  "IncorrectSalt",
			input: "password",
			expected: Credential{
				Algorithm: "SHA256",
				Hash:      "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
				Parameters: CredParameters{
					Salt: "incorrectsalt",
				},
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			hash := Verify([]byte(tc.input), tc.expected)

			assert.False(t, hash)
		})
	}
}

func (suite *HashUtilsTestSuite) TestSha256HashAndVerify() {
	// Set runtime config to SHA256
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(SHA256),
				Parameters: config.PasswordHashingParamsConfig{
					SaltSize: 16,
				},
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	input := "test-input"
	cred := Generate([]byte(input))

	assert.True(suite.T(), Verify([]byte(input), cred),
		"Hash verification should succeed for the same input")
}

func (suite *HashUtilsTestSuite) TestGeneratePBKDF2() {
	// Set runtime config to PBKDF2
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(PBKDF2),
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	cred := Generate(suite.input)

	assert.Equal(suite.T(), PBKDF2, cred.Algorithm, "Algorithm should be PBKDF2")
	assert.NotEmpty(suite.T(), cred.Hash, "Hash should not be empty")
}

func (suite *HashUtilsTestSuite) TestGeneratePBKDF2_Failure() {
	// Set runtime config to PBKDF2 with invalid parameters
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(PBKDF2),
				Parameters: config.PasswordHashingParamsConfig{
					Iterations: -1,
					KeySize:    -1,
				},
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	cred := Generate(suite.input)

	assert.Empty(suite.T(), cred.Hash, "Hash should be empty")
}

func (suite *HashUtilsTestSuite) TestVerifyBKDF2() {
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
			},
		},
		{
			name:  "NormalStringWithoutSalt",
			input: "password",
			expected: Credential{
				Algorithm: "PBKDF2",
				Hash:      "fdc25be00b18ba5c79d8bf7a452d98c248b11f2c7e9c871d24f1f880381e95cf",
			},
		},
		{
			name:  "NormalStringWithSaltAndParameters",
			input: "password",
			expected: Credential{
				Algorithm: "PBKDF2",
				Hash:      "b500f5369698b4bcdde08267c406c12ff95e8de1d431e4472bf6ea95b620da5c",
				Parameters: CredParameters{
					Salt:       "36d2dde7dfbafe8e04ea49450f659b1c",
					Iterations: defaultPBKDF2Iterations,
					KeySize:    defaultPBKDF2KeySize,
				},
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			hash := Verify([]byte(tc.input), tc.expected)

			assert.True(t, hash)
		})
	}
}

func (suite *HashUtilsTestSuite) TestVerifyPBKDF2_Failure() {
	{
		testCases := []struct {
			name     string
			input    string
			expected Credential
		}{
			{
				name:  "IncorrectHash",
				input: "password",
				expected: Credential{
					Algorithm: "PBKDF2",
					Hash:      "incorrecthashvalue",
				},
			},
			{
				name:  "IncorrectSalt",
				input: "password",
				expected: Credential{
					Algorithm: "PBKDF2",
					Hash:      "fdc25be00b18ba5c79d8bf7a452d98c248b11f2c7e9c871d24f1f880381e95cf",
					Parameters: CredParameters{
						Salt: "incorrectsalt",
					},
				},
			},
			{
				name:  "IncorrectParameters",
				input: "password",
				expected: Credential{
					Algorithm: "PBKDF2",
					Hash:      "b500f5369698b4bcdde08267c406c12ff95e8de1d431e4472bf6ea95b620da5c",
					Parameters: CredParameters{
						Salt:    "36d2dde7dfbafe8e04ea49450f659b1c",
						KeySize: -1,
					},
				},
			},
		}

		for _, tc := range testCases {
			suite.T().Run(tc.name, func(t *testing.T) {
				hash := Verify([]byte(tc.input), tc.expected)

				assert.False(t, hash)
			})
		}
	}
}

func (suite *HashUtilsTestSuite) TestPBKDF2HashWithAndVerify() {
	// Set runtime config to PBKDF2
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(PBKDF2),
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	input := "test-input"
	cred := Generate([]byte(input))

	assert.True(suite.T(), Verify([]byte(input), cred),
		"Hash verification should succeed for the same input")
}

func (suite *HashUtilsTestSuite) TestUnsupportedAlgorithm_Failure() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: "UNSUPPORTED",
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	// Expectign error log and empty credential
	cred := Generate(suite.input)
	assert.Equal(suite.T(), Credential{}, cred, "Credential should be empty for unsupported algorithm")
}

func (suite *HashUtilsTestSuite) TestUnsupportedAlgorithmVerify_Failure() {
	referenceCredential := Credential{
		Algorithm: "UNSUPPORTED",
		Hash:      "somehash",
		Parameters: CredParameters{
			Salt: "somesalt",
		},
	}
	result := Verify(suite.input, referenceCredential)

	assert.False(suite.T(), result, "Verification should fail for unsupported algorithm")
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

func (suite *HashUtilsTestSuite) TestGenerateSalt() {
	salt, err := generateSalt(defaultSaltSize)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), salt)
	assert.Equal(suite.T(), 16, len(salt), "Generated salt should be 16 bytes")
}

func (suite *HashUtilsTestSuite) TestGenerateSaltUniqueness() {
	salt1, err1 := generateSalt(defaultSaltSize)
	salt2, err2 := generateSalt(defaultSaltSize)

	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)
	assert.NotEqual(suite.T(), salt1, salt2, "Generated salts should be different")
}
