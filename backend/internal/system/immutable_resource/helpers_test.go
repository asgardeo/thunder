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

package immutableresource

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

func TestIsImmutableModeEnabled(t *testing.T) {
	t.Run("Returns true when immutable resources are enabled", func(t *testing.T) {
		config.ResetThunderRuntime()
		defer config.ResetThunderRuntime() // Clean up after test
		testConfig := &config.Config{
			ImmutableResources: config.ImmutableResources{
				Enabled: true,
			},
		}
		err := config.InitializeThunderRuntime("", testConfig)
		assert.NoError(t, err)

		result := IsImmutableModeEnabled()
		assert.True(t, result)
	})

	t.Run("Returns false when immutable resources are disabled", func(t *testing.T) {
		config.ResetThunderRuntime()
		defer config.ResetThunderRuntime() // Clean up after test
		testConfig := &config.Config{
			ImmutableResources: config.ImmutableResources{
				Enabled: false,
			},
		}
		err := config.InitializeThunderRuntime("", testConfig)
		assert.NoError(t, err)

		result := IsImmutableModeEnabled()
		assert.False(t, result)
	})
}

func TestCheckImmutableOperations(t *testing.T) {
	testCases := []struct {
		name            string
		operation       string
		checkFunc       func() *serviceerror.ServiceError
		expectedErrText string
	}{
		{
			name:            "CheckImmutableCreate",
			operation:       "create",
			checkFunc:       CheckImmutableCreate,
			expectedErrText: "Immutable resource create operation",
		},
		{
			name:            "CheckImmutableUpdate",
			operation:       "update",
			checkFunc:       CheckImmutableUpdate,
			expectedErrText: "Immutable resource update operation",
		},
		{
			name:            "CheckImmutableDelete",
			operation:       "delete",
			checkFunc:       CheckImmutableDelete,
			expectedErrText: "Immutable resource delete operation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("Returns error when immutable mode is enabled", func(t *testing.T) {
				config.ResetThunderRuntime()
				defer config.ResetThunderRuntime() // Clean up after test
				testConfig := &config.Config{
					ImmutableResources: config.ImmutableResources{
						Enabled: true,
					},
				}
				err := config.InitializeThunderRuntime("", testConfig)
				assert.NoError(t, err)

				result := tc.checkFunc()
				assert.NotNil(t, result)
				assert.Contains(t, result.Error, tc.expectedErrText)
			})

			t.Run("Returns nil when immutable mode is disabled", func(t *testing.T) {
				config.ResetThunderRuntime()
				defer config.ResetThunderRuntime() // Clean up after test
				testConfig := &config.Config{
					ImmutableResources: config.ImmutableResources{
						Enabled: false,
					},
				}
				err := config.InitializeThunderRuntime("", testConfig)
				assert.NoError(t, err)

				result := tc.checkFunc()
				assert.Nil(t, result)
			})
		})
	}
}
