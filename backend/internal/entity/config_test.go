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

package entity

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	config.ResetThunderRuntime()
}

func (s *ConfigTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func (s *ConfigTestSuite) initRuntime(cfg *config.Config) {
	err := config.InitializeThunderRuntime("", cfg)
	s.Require().NoError(err)
}

func (s *ConfigTestSuite) TestGetEntityStoreMode_EntityStoreExplicit() {
	tests := []struct {
		store    string
		expected serverconst.StoreMode
	}{
		{"mutable", serverconst.StoreModeMutable},
		{"declarative", serverconst.StoreModeDeclarative},
		{"composite", serverconst.StoreModeComposite},
		{"MUTABLE", serverconst.StoreModeMutable},
		{"  Composite  ", serverconst.StoreModeComposite},
	}

	for _, tt := range tests {
		config.ResetThunderRuntime()
		s.initRuntime(&config.Config{Entity: config.EntityConfig{Store: tt.store}})
		s.Equal(tt.expected, getEntityStoreMode())
	}
}

func (s *ConfigTestSuite) TestGetEntityStoreMode_EntityStoreInvalid_FallsBackToUserStore() {
	config.ResetThunderRuntime()
	s.initRuntime(&config.Config{
		Entity: config.EntityConfig{Store: "invalid-mode"},
		User:   config.UserConfig{Store: "composite"},
	})
	s.Equal(serverconst.StoreModeComposite, getEntityStoreMode())
}

func (s *ConfigTestSuite) TestGetEntityStoreMode_EntityStoreInvalid_UserStoreInvalid_FallsBackToDeclarative() {
	config.ResetThunderRuntime()
	s.initRuntime(&config.Config{
		Entity:               config.EntityConfig{Store: "invalid"},
		User:                 config.UserConfig{Store: "also-invalid"},
		DeclarativeResources: config.DeclarativeResources{Enabled: true},
	})
	s.Equal(serverconst.StoreModeDeclarative, getEntityStoreMode())
}

func (s *ConfigTestSuite) TestGetEntityStoreMode_FallbackDeclarativeEnabled() {
	config.ResetThunderRuntime()
	s.initRuntime(&config.Config{
		DeclarativeResources: config.DeclarativeResources{Enabled: true},
	})
	s.Equal(serverconst.StoreModeDeclarative, getEntityStoreMode())
}

func (s *ConfigTestSuite) TestGetEntityStoreMode_FallbackDeclarativeDisabled() {
	config.ResetThunderRuntime()
	s.initRuntime(&config.Config{
		DeclarativeResources: config.DeclarativeResources{Enabled: false},
	})
	s.Equal(serverconst.StoreModeMutable, getEntityStoreMode())
}

func (s *ConfigTestSuite) TestGetEntityStoreMode_UserStoreFallback() {
	tests := []struct {
		userStore string
		expected  serverconst.StoreMode
	}{
		{"mutable", serverconst.StoreModeMutable},
		{"declarative", serverconst.StoreModeDeclarative},
		{"composite", serverconst.StoreModeComposite},
	}

	for _, tt := range tests {
		config.ResetThunderRuntime()
		s.initRuntime(&config.Config{User: config.UserConfig{Store: tt.userStore}})
		s.Equal(tt.expected, getEntityStoreMode())
	}
}

func (s *ConfigTestSuite) TestGetEntityStoreMode_UserStoreInvalid_FallbackMutable() {
	config.ResetThunderRuntime()
	s.initRuntime(&config.Config{
		User:                 config.UserConfig{Store: "invalid"},
		DeclarativeResources: config.DeclarativeResources{Enabled: false},
	})
	s.Equal(serverconst.StoreModeMutable, getEntityStoreMode())
}

func (s *ConfigTestSuite) TestGetIndexedAttributes_EntityConfig() {
	config.ResetThunderRuntime()
	s.initRuntime(&config.Config{
		Entity: config.EntityConfig{IndexedAttributes: []string{"email", "username"}},
	})
	s.Equal([]string{"email", "username"}, getIndexedAttributes())
}

func (s *ConfigTestSuite) TestGetIndexedAttributes_FallbackUserConfig() {
	config.ResetThunderRuntime()
	s.initRuntime(&config.Config{
		User: config.UserConfig{IndexedAttributes: []string{"phone"}},
	})
	s.Equal([]string{"phone"}, getIndexedAttributes())
}

func (s *ConfigTestSuite) TestGetIndexedAttributes_Empty() {
	config.ResetThunderRuntime()
	s.initRuntime(&config.Config{})
	s.Nil(getIndexedAttributes())
}
