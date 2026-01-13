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
)

// ThunderRuntime holds the runtime configuration for the Thunder server.
type ThunderRuntime struct {
	ThunderHome string `yaml:"thunder_home"`
	Config      Config `yaml:"config"`
}

var (
	runtimeConfig *ThunderRuntime
	once          sync.Once
)

// InitializeThunderRuntime initializes the ThunderRuntime configuration.
func InitializeThunderRuntime(thunderHome string, config *Config) error {
	once.Do(func() {
		runtimeConfig = &ThunderRuntime{
			ThunderHome: thunderHome,
			Config:      *config,
		}
	})

	return nil
}

// GetThunderRuntime returns the ThunderRuntime configuration.
func GetThunderRuntime() *ThunderRuntime {
	if runtimeConfig == nil {
		panic("ThunderRuntime is not initialized")
	}
	return runtimeConfig
}

// ResetThunderRuntime resets the ThunderRuntime.
// This should only be used in tests to reset the singleton state.
func ResetThunderRuntime() {
	runtimeConfig = nil
	once = sync.Once{}
}
