/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

import "sync"

// ThunderRuntime holds the runtime configuration for the Thunder server.
type ThunderRuntime struct {
	ThunderHome string `yaml:"thunder_home"`
	Config      Config `yaml:"config"`
}

var (
	runtimeConfig *ThunderRuntime
	once          sync.Once
	mu            sync.RWMutex
)

// InitializeThunderRuntime initializes the ThunderRuntime configuration.
func InitializeThunderRuntime(thunderHome string, config *Config) error {
	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()
		runtimeConfig = &ThunderRuntime{
			ThunderHome: thunderHome,
			Config:      *config,
		}
	})

	return nil
}

// GetThunderRuntime returns the ThunderRuntime configuration.
func GetThunderRuntime() *ThunderRuntime {
	mu.RLock()
	defer mu.RUnlock()
	if runtimeConfig == nil {
		panic("ThunderRuntime is not initialized")
	}
	return runtimeConfig
}

// ResetThunderRuntimeForTest resets the runtime config for testing purposes.
// This function should only be used in tests.
func ResetThunderRuntimeForTest() {
	mu.Lock()
	defer mu.Unlock()
	runtimeConfig = nil
	once = sync.Once{}
}
