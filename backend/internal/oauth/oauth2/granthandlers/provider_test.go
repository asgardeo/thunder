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

package granthandlers

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	_ "modernc.org/sqlite"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
)

type GrantHandlerProviderTestSuite struct {
	suite.Suite
	provider GrantHandlerProviderInterface
	tempDir  string
}

func TestGrantHandlerProviderSuite(t *testing.T) {
	suite.Run(t, new(GrantHandlerProviderTestSuite))
}

func (suite *GrantHandlerProviderTestSuite) SetupTest() {
	// Create a temporary directory for the test database.
	tempDir, err := os.MkdirTemp("", "provider-test-")
	if err != nil {
		suite.T().Fatalf("Failed to create temp dir: %v", err)
	}
	suite.tempDir = tempDir

	testConfig := &config.Config{
		Database: config.DatabaseConfig{
			Identity: config.DataSource{
				Type: "sqlite",
				Path: filepath.Join(suite.tempDir, "identity.db"),
			},
			Runtime: config.DataSource{
				Type: "sqlite",
				Path: filepath.Join(suite.tempDir, "runtime.db"),
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
	suite.provider = NewGrantHandlerProvider()
}

func (suite *GrantHandlerProviderTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
	// Clean up the temporary directory.
	if err := os.RemoveAll(suite.tempDir); err != nil {
		suite.T().Logf("Failed to clean up temp dir: %v", err)
	}
}

func (suite *GrantHandlerProviderTestSuite) TestNewGrantHandlerProvider() {
	provider := NewGrantHandlerProvider()
	assert.NotNil(suite.T(), provider)
	assert.Implements(suite.T(), (*GrantHandlerProviderInterface)(nil), provider)
}

func (suite *GrantHandlerProviderTestSuite) TestGetGrantHandler_ClientCredentials() {
	handler, err := suite.provider.GetGrantHandler(constants.GrantTypeClientCredentials)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*GrantHandlerInterface)(nil), handler)
}

func (suite *GrantHandlerProviderTestSuite) TestGetGrantHandler_AuthorizationCode() {
	handler, err := suite.provider.GetGrantHandler(constants.GrantTypeAuthorizationCode)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*GrantHandlerInterface)(nil), handler)
}

func (suite *GrantHandlerProviderTestSuite) TestGetGrantHandler_RefreshToken() {
	handler, err := suite.provider.GetGrantHandler(constants.GrantTypeRefreshToken)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*GrantHandlerInterface)(nil), handler)
	assert.Implements(suite.T(), (*RefreshTokenGrantHandlerInterface)(nil), handler)
}

func (suite *GrantHandlerProviderTestSuite) TestGetGrantHandler_UnsupportedGrantType() {
	unsupportedGrantTypes := []struct {
		name      string
		grantType constants.GrantType
	}{
		{"Password", constants.GrantTypePassword},
		{"Implicit", constants.GrantTypeImplicit},
		{"InvalidType", constants.GrantType("invalid_type")},
		{"EmptyType", constants.GrantType("")},
	}

	for _, tc := range unsupportedGrantTypes {
		suite.T().Run(tc.name, func(t *testing.T) {
			handler, err := suite.provider.GetGrantHandler(tc.grantType)

			assert.Error(t, err)
			assert.Nil(t, handler)
			assert.Equal(t, constants.UnSupportedGrantTypeError, err)
		})
	}
}

func (suite *GrantHandlerProviderTestSuite) TestGetGrantHandler_AllSupportedTypes() {
	supportedTypes := []constants.GrantType{
		constants.GrantTypeClientCredentials,
		constants.GrantTypeAuthorizationCode,
		constants.GrantTypeRefreshToken,
	}

	for _, grantType := range supportedTypes {
		suite.T().Run(string(grantType), func(t *testing.T) {
			handler, err := suite.provider.GetGrantHandler(grantType)

			assert.NoError(t, err)
			assert.NotNil(t, handler)
			assert.Implements(t, (*GrantHandlerInterface)(nil), handler)
		})
	}
}
