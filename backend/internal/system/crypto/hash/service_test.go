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

type HashServiceTestSuite struct {
	suite.Suite
	input []byte
}

func TestHashServiceSuite(t *testing.T) {
	suite.Run(t, new(HashServiceTestSuite))
}

func (suite *HashServiceTestSuite) SetupSuite() {
	suite.input = []byte("secretPassword123")
}

func (suite *HashServiceTestSuite) TearDownSuite() {
	config.ResetThunderRuntime()
}

func (suite *HashServiceTestSuite) TestGenerateSha256() {
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

	cred, err := newHashService().Generate(suite.input)

	assert.NoError(suite.T(), err, "Error should be nil when generating hash")
	assert.Equal(suite.T(), SHA256, cred.Algorithm, "Algorithm should be SHA256")
	assert.NotEmpty(suite.T(), cred.Hash, "Hash should not be empty")
}

func (suite *HashServiceTestSuite) TestSHA256HashWithCustomSaltSize() {
	// Set runtime config to SHA256 with custom salt size
	customSaltSize := 32

	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(SHA256),
				Parameters: config.PasswordHashingParamsConfig{
					SaltSize: customSaltSize,
				},
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	cred, err := newHashService().Generate(suite.input)
	assert.NoError(suite.T(), err, "Error should be nil when generating hash")
	assert.Equal(suite.T(), SHA256, cred.Algorithm, "Algorithm should be SHA256")
	assert.NotEmpty(suite.T(), cred.Hash, "Hash should not be empty")
	assert.NotEmpty(suite.T(), cred.Parameters.Salt, "Salt should not be empty")

	// Verify the salt size is as expected (hex encoded, so 2x the byte size)
	expectedSaltLength := customSaltSize * 2 // hex encoding doubles the length
	assert.Equal(suite.T(), expectedSaltLength, len(cred.Parameters.Salt),
		"Salt should be hex encoded with expected length")

	// Verify that the generated credential can be verified
	ok, err := newHashService().Verify(suite.input, cred)
	assert.NoError(suite.T(), err, "Error should be nil when verifying hash")
	assert.True(suite.T(), ok, "Hash verification should succeed for the same input with custom salt size")
}

func (suite *HashServiceTestSuite) TestVerifySha256() {
	testCases := []struct {
		name     string
		input    string
		expected Credential
	}{
		{
			name:  "EmptyStringAndSalt",
			input: "",
			expected: Credential{
				Algorithm: SHA256,
				Hash:      "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
		},
		{
			name:  "NormalStringWithoutSalt",
			input: "password",
			expected: Credential{
				Algorithm: SHA256,
				Hash:      "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
			},
		},
		{
			name:  "NormalStringWithSalt",
			input: "password",
			expected: Credential{
				Algorithm: SHA256,
				Hash:      "4b2dcea502b405a479a69fd2478ea891fa9f02966db9ee5cbcbee53137c8ae4d",
				Parameters: CredParameters{
					Salt: "12f4576d7432bd8020db7202b6492a37",
				},
			},
		},
	}

	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(SHA256),
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			ok, err := newHashService().Verify([]byte(tc.input), tc.expected)
			assert.NoError(t, err, "Error should be nil when verifying hash")
			assert.True(t, ok)
		})
	}
}

func (suite *HashServiceTestSuite) TestVerifySha256_Failure() {
	testCases := []struct {
		name     string
		input    string
		expected Credential
		error    bool
	}{
		{
			name:  "IncorrectHash",
			input: "password",
			expected: Credential{
				Algorithm: SHA256,
				Hash:      "incorrecthashvalue",
			},
			error: false,
		},
		{
			name:  "IncorrectSalt",
			input: "password",
			expected: Credential{
				Algorithm: SHA256,
				Hash:      "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
				Parameters: CredParameters{
					Salt: "incorrectsalt",
				},
			},
			error: true,
		},
	}

	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(SHA256),
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			ok, err := newHashService().Verify([]byte(tc.input), tc.expected)
			assert.False(t, ok)
			if !tc.error {
				assert.NoError(t, err, "Error should be nil when verifying hash")
			} else {
				assert.Error(t, err, "Error should not be nil when verifying hash with invalid parameters")
			}
		})
	}
}

func (suite *HashServiceTestSuite) TestSha256HashAndVerify() {
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

	cred, err := newHashService().Generate(suite.input)
	assert.NoError(suite.T(), err, "Error should be nil when generating hash")

	ok, err := newHashService().Verify(suite.input, cred)
	assert.NoError(suite.T(), err, "Error should be nil when verifying hash")
	assert.True(suite.T(), ok, "Hash verification should succeed for the same input")
}

func (suite *HashServiceTestSuite) TestGeneratePBKDF2() {
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

	cred, err := newHashService().Generate(suite.input)
	assert.NoError(suite.T(), err, "Error should be nil when generating hash")
	assert.Equal(suite.T(), PBKDF2, cred.Algorithm, "Algorithm should be PBKDF2")
	assert.NotEmpty(suite.T(), cred.Hash, "Hash should not be empty")
}

func (suite *HashServiceTestSuite) TestPBKDF2HashWithCustomParameters() {
	// Set runtime config to PBKDF2 with custom parameters
	customIterations := 100000
	customKeySize := 64
	customSaltSize := 32

	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(PBKDF2),
				Parameters: config.PasswordHashingParamsConfig{
					SaltSize:   customSaltSize,
					Iterations: customIterations,
					KeySize:    customKeySize,
				},
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	cred, err := newHashService().Generate(suite.input)
	assert.NoError(suite.T(), err, "Error should be nil when generating hash")
	assert.Equal(suite.T(), PBKDF2, cred.Algorithm, "Algorithm should be PBKDF2")
	assert.NotEmpty(suite.T(), cred.Hash, "Hash should not be empty")
	assert.NotEmpty(suite.T(), cred.Parameters.Salt, "Salt should not be empty")
	assert.Equal(suite.T(), customIterations, cred.Parameters.Iterations,
		"Credential should contain configured custom iterations")
	assert.Equal(suite.T(), customKeySize, cred.Parameters.KeySize,
		"Credential should contain configured custom key size")

	// Verify the salt size is as expected (hex encoded, so 2x the byte size)
	expectedSaltLength := customSaltSize * 2 // hex encoding doubles the length
	assert.Equal(suite.T(), expectedSaltLength, len(cred.Parameters.Salt),
		"Salt should be hex encoded with expected length")

	// Verify the hash length matches the key size (hex encoded, so 2x the byte size)
	expectedHashLength := customKeySize * 2 // hex encoding doubles the length
	assert.Equal(suite.T(), expectedHashLength, len(cred.Hash),
		"Hash length should match configured key size")

	// Verify that the generated credential can be verified
	ok, err := newHashService().Verify(suite.input, cred)
	assert.NoError(suite.T(), err, "Error should be nil when verifying hash")
	assert.True(suite.T(), ok, "Hash verification should succeed for the same input with custom parameters")
}

func (suite *HashServiceTestSuite) TestGeneratePBKDF2_Failure() {
	// Set runtime config to PBKDF2 with invalid parameters
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(PBKDF2),
				Parameters: config.PasswordHashingParamsConfig{
					KeySize: 137438953473,
				},
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	cred, err := newHashService().Generate(suite.input)
	assert.Error(suite.T(), err, "Error should not be nil when generating hash with invalid parameters")
	assert.Empty(suite.T(), cred.Hash, "Hash should be empty")
}

func (suite *HashServiceTestSuite) TestVerifyBKDF2() {
	testCases := []struct {
		name     string
		input    string
		expected Credential
	}{
		{
			name:  "EmptyStringAndSalt",
			input: "",
			expected: Credential{
				Algorithm: PBKDF2,
				Hash:      "3106cb5743a54114a36bb7d3b2afa0242360b58243264728a9ca208548082281",
			},
		},
		{
			name:  "NormalStringWithoutSalt",
			input: "password",
			expected: Credential{
				Algorithm: PBKDF2,
				Hash:      "fdc25be00b18ba5c79d8bf7a452d98c248b11f2c7e9c871d24f1f880381e95cf",
			},
		},
		{
			name:  "NormalStringWithSaltAndParameters",
			input: "password",
			expected: Credential{
				Algorithm: PBKDF2,
				Hash:      "b500f5369698b4bcdde08267c406c12ff95e8de1d431e4472bf6ea95b620da5c",
				Parameters: CredParameters{
					Salt:       "36d2dde7dfbafe8e04ea49450f659b1c",
					Iterations: defaultPBKDF2Iterations,
					KeySize:    defaultPBKDF2KeySize,
				},
			},
		},
	}

	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(PBKDF2),
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			ok, err := newHashService().Verify([]byte(tc.input), tc.expected)
			assert.NoError(t, err, "Error should be nil when verifying hash")
			assert.True(t, ok)
		})
	}
}

func (suite *HashServiceTestSuite) TestVerifyPBKDF2_Failure() {
	testCases := []struct {
		name     string
		input    string
		expected Credential
		error    bool
	}{
		{
			name:  "IncorrectHash",
			input: "password",
			expected: Credential{
				Algorithm: PBKDF2,
				Hash:      "incorrecthashvalue",
			},
			error: false,
		},
		{
			name:  "IncorrectSalt",
			input: "password",
			expected: Credential{
				Algorithm: PBKDF2,
				Hash:      "fdc25be00b18ba5c79d8bf7a452d98c248b11f2c7e9c871d24f1f880381e95cf",
				Parameters: CredParameters{
					Salt: "incorrectsalt",
				},
			},
			error: true,
		},
		{
			name:  "IncorrectParameters",
			input: "password",
			expected: Credential{
				Algorithm: PBKDF2,
				Hash:      "b500f5369698b4bcdde08267c406c12ff95e8de1d431e4472bf6ea95b620da5c",
				Parameters: CredParameters{
					Salt:    "36d2dde7dfbafe8e04ea49450f659b1c",
					KeySize: 137438953473,
				},
			},
			error: true,
		},
	}

	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(PBKDF2),
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			ok, err := newHashService().Verify([]byte(tc.input), tc.expected)
			if !tc.error {
				assert.NoError(t, err, "Error should be nil when verifying hash")
			} else {
				assert.Error(t, err, "Error should not be nil when verifying hash with invalid parameters")
			}
			assert.False(t, ok)
		})
	}
}

func (suite *HashServiceTestSuite) TestPBKDF2HashWithAndVerify() {
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

	cred, err := newHashService().Generate(suite.input)
	assert.NoError(suite.T(), err, "Error should be nil when generating hash")

	ok, err := newHashService().Verify(suite.input, cred)
	assert.NoError(suite.T(), err, "Error should be nil when verifying hash")
	assert.True(suite.T(), ok, "Hash verification should succeed for the same input")
}

func (suite *HashServiceTestSuite) TestUnsupportedAlgorithm_Failure() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: "UNSUPPORTED",
			},
		},
	}
	config.ResetThunderRuntime()
	_ = config.InitializeThunderRuntime("/test/thunder/home", testConfig)

	// Expecting error log and empty credential
	defer func() {
		if r := recover(); r != nil {
			suite.T().Logf("Recovered from panic: %v", r)
		}
	}()

	// Expecting error log and empty credential
	cred, err := newHashService().Generate(suite.input)
	assert.NoError(suite.T(), err, "Error should be nil when generating hash")
	assert.Equal(suite.T(), Credential{}, cred, "Credential should be empty for unsupported algorithm")
}

func (suite *HashServiceTestSuite) TestUnsupportedAlgorithmVerify_Failure() {
	referenceCredential := Credential{
		Algorithm: "UNSUPPORTED",
		Hash:      "somehash",
		Parameters: CredParameters{
			Salt: "somesalt",
		},
	}
	// Expecting error log and empty credential
	defer func() {
		if r := recover(); r != nil {
			suite.T().Logf("Recovered from panic: %v", r)
		}
	}()

	ok, err := newHashService().Verify(suite.input, referenceCredential)
	assert.Error(suite.T(), err, "Error should not be nil when verifying hash with unsupported algorithm")
	assert.False(suite.T(), ok, "Verification should fail for unsupported algorithm")
}

func (suite *HashServiceTestSuite) TestGenerateSalt() {
	salt, err := generateSalt(defaultSaltSize)
	assert.NoError(suite.T(), err, "Error should be nil when generating salt")
	assert.NotEmpty(suite.T(), salt)
	assert.Equal(suite.T(), 16, len(salt), "Generated salt should be 16 bytes")
}

func (suite *HashServiceTestSuite) TestGenerateSaltUniqueness() {
	salt1, err := generateSalt(defaultSaltSize)
	assert.NoError(suite.T(), err, "Error should be nil when generating salt")
	salt2, err := generateSalt(defaultSaltSize)
	assert.NoError(suite.T(), err, "Error should be nil when generating salt")

	assert.NotEqual(suite.T(), salt1, salt2, "Generated salts should be different")
}
