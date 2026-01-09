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

package config

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RuntimeConfigTestSuite struct {
	suite.Suite
}

func TestRuntimeConfigSuite(t *testing.T) {
	suite.Run(t, new(RuntimeConfigTestSuite))
}

func (suite *RuntimeConfigTestSuite) BeforeTest(suiteName, testName string) {
	runtimeConfig = nil
	once = sync.Once{}
}

func (suite *RuntimeConfigTestSuite) TestInitializeThunderRuntime() {
	config := &Config{
		Server: ServerConfig{
			Hostname: "testhost",
			Port:     9000,
		},
		TLS: TLSConfig{
			CertFile: "test-cert.pem",
			KeyFile:  "test-key.pem",
		},
	}

	err := InitializeThunderRuntime("/test/thunder/home", config)

	assert.NoError(suite.T(), err)

	runtime := runtimeConfig
	assert.NotNil(suite.T(), runtime)
	assert.Equal(suite.T(), "/test/thunder/home", runtime.ThunderHome)
	assert.Equal(suite.T(), config.Server.Hostname, runtime.Config.Server.Hostname)
	assert.Equal(suite.T(), config.Server.Port, runtime.Config.Server.Port)
	assert.Equal(suite.T(), config.TLS.CertFile, runtime.Config.TLS.CertFile)
}

func (suite *RuntimeConfigTestSuite) TestInitializeThunderRuntimeOnlyOnce() {
	// First initialization
	firstConfig := &Config{
		Server: ServerConfig{
			Hostname: "firsthost",
			Port:     8000,
		},
	}

	err := InitializeThunderRuntime("/first/path", firstConfig)
	assert.NoError(suite.T(), err)

	// Try second initialization
	secondConfig := &Config{
		Server: ServerConfig{
			Hostname: "secondhost",
			Port:     9000,
		},
	}

	err = InitializeThunderRuntime("/second/path", secondConfig)
	assert.NoError(suite.T(), err) // Should not return error

	// Verify that the first initialization remains
	runtime := GetThunderRuntime()
	assert.Equal(suite.T(), "/first/path", runtime.ThunderHome)
	assert.Equal(suite.T(), "firsthost", runtime.Config.Server.Hostname)
	assert.Equal(suite.T(), 8000, runtime.Config.Server.Port)
}

func (suite *RuntimeConfigTestSuite) TestGetThunderRuntime() {
	config := &Config{
		Server: ServerConfig{
			Hostname: "gettest",
			Port:     8888,
		},
	}

	err := InitializeThunderRuntime("/get/test/path", config)
	assert.NoError(suite.T(), err)

	runtime := GetThunderRuntime()

	assert.NotNil(suite.T(), runtime)
	assert.Equal(suite.T(), "/get/test/path", runtime.ThunderHome)
	assert.Equal(suite.T(), "gettest", runtime.Config.Server.Hostname)
}

func (suite *RuntimeConfigTestSuite) TestGetThunderRuntimePanic() {
	runtimeConfig = nil

	assert.Panics(suite.T(), func() {
		GetThunderRuntime()
	})
}
