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

package user

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
)

// ConfigTestSuite tests the user store configuration.
type ConfigTestSuite struct {
	suite.Suite
}

// SetupTest initializes Thunder runtime for config tests.
func (suite *ConfigTestSuite) SetupTest() {
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("test", &config.Config{})
	suite.Require().NoError(err)
}

// TestConfigFunctions verifies user store mode resolution.
func (suite *ConfigTestSuite) TestConfigFunctions() {
	testCases := []struct {
		name                string
		userStore           string
		globalDeclarative   bool
		expectedMode        serverconst.StoreMode
		expectedDeclarative bool
	}{
		{
			name:                "GlobalMutable_Default",
			userStore:           "",
			globalDeclarative:   false,
			expectedMode:        serverconst.StoreModeMutable,
			expectedDeclarative: false,
		},
		{
			name:                "GlobalDeclarative_Default",
			userStore:           "",
			globalDeclarative:   true,
			expectedMode:        serverconst.StoreModeDeclarative,
			expectedDeclarative: true,
		},
		{
			name:                "ExplicitMutable_OverridesGlobal",
			userStore:           string(serverconst.StoreModeMutable),
			globalDeclarative:   true,
			expectedMode:        serverconst.StoreModeMutable,
			expectedDeclarative: false,
		},
		{
			name:                "ExplicitDeclarative_OverridesGlobal",
			userStore:           string(serverconst.StoreModeDeclarative),
			globalDeclarative:   false,
			expectedMode:        serverconst.StoreModeDeclarative,
			expectedDeclarative: true,
		},
		{
			name:                "ExplicitComposite",
			userStore:           string(serverconst.StoreModeComposite),
			globalDeclarative:   true,
			expectedMode:        serverconst.StoreModeComposite,
			expectedDeclarative: false,
		},
		{
			name:                "InvalidStore_FallsBackToGlobalDisabled",
			userStore:           "invalid",
			globalDeclarative:   false,
			expectedMode:        serverconst.StoreModeMutable,
			expectedDeclarative: false,
		},
		{
			name:                "InvalidStore_FallsBackToGlobalEnabled",
			userStore:           "invalid",
			globalDeclarative:   true,
			expectedMode:        serverconst.StoreModeDeclarative,
			expectedDeclarative: true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			config.ResetThunderRuntime()
			err := config.InitializeThunderRuntime("test", &config.Config{
				User:                 config.UserConfig{Store: tc.userStore},
				DeclarativeResources: config.DeclarativeResources{Enabled: tc.globalDeclarative},
			})
			require.NoError(t, err)
			suite.Equal(tc.expectedMode, getUserStoreMode())
			suite.Equal(tc.expectedDeclarative, isDeclarativeModeEnabled())
		})
	}
}

// TestConfigTestSuite runs the test suite.
func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
