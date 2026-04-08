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

package userprovider

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type InitUserProviderTestSuite struct {
	suite.Suite
	mockUserService *usermock.UserServiceInterfaceMock
}

func (suite *InitUserProviderTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())

	// Initialize ThunderRuntime with a basic test config
	testConfig := &config.Config{
		Database: config.DatabaseConfig{
			Config: config.DataSource{
				Type:   "sqlite",
				SQLite: config.SQLiteDataSource{Path: ":memory:"},
			},
			Runtime: config.DataSource{
				Type:   "sqlite",
				SQLite: config.SQLiteDataSource{Path: ":memory:"},
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (suite *InitUserProviderTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func TestInitUserProviderTestSuite(t *testing.T) {
	suite.Run(t, new(InitUserProviderTestSuite))
}

func (suite *InitUserProviderTestSuite) TestInitializeUserProvider_WithDisabledType() {
	// Set user provider type to "disabled"
	config.GetThunderRuntime().Config.UserProvider = config.UserProviderConfig{
		Type: "disabled",
	}

	provider := InitializeUserProvider(suite.mockUserService)

	// Assert that the provider is of type disabledUserProvider
	suite.NotNil(provider)
	_, ok := provider.(*disabledUserProvider)
	suite.True(ok, "Expected provider to be of type *disabledUserProvider")
}

func (suite *InitUserProviderTestSuite) TestInitializeUserProvider_WithDefaultType() {
	// Set user provider type to "default"
	config.GetThunderRuntime().Config.UserProvider = config.UserProviderConfig{
		Type: "default",
	}

	provider := InitializeUserProvider(suite.mockUserService)

	// Assert that the provider is of type defaultUserProvider
	suite.NotNil(provider)
	_, ok := provider.(*defaultUserProvider)
	suite.True(ok, "Expected provider to be of type *defaultUserProvider")
}

func (suite *InitUserProviderTestSuite) TestInitializeUserProvider_WithEmptyType() {
	// Set user provider type to empty (should default to default provider)
	config.GetThunderRuntime().Config.UserProvider = config.UserProviderConfig{
		Type: "",
	}

	provider := InitializeUserProvider(suite.mockUserService)

	// Assert that the provider is of type defaultUserProvider (default case)
	suite.NotNil(provider)
	_, ok := provider.(*defaultUserProvider)
	suite.True(ok, "Expected provider to be of type *defaultUserProvider when type is empty")
}

func (suite *InitUserProviderTestSuite) TestInitializeUserProvider_WithUnknownType() {
	// Set user provider type to an unknown value (should default to default provider)
	config.GetThunderRuntime().Config.UserProvider = config.UserProviderConfig{
		Type: "unknown",
	}

	provider := InitializeUserProvider(suite.mockUserService)

	// Assert that the provider is of type defaultUserProvider (default case)
	suite.NotNil(provider)
	_, ok := provider.(*defaultUserProvider)
	suite.True(ok, "Expected provider to be of type *defaultUserProvider for unknown type")
}
