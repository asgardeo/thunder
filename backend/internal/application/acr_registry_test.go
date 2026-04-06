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

package application

import (
	"testing"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initAcrRegistry initialises the Thunder runtime with the given ACR-AMR mapping for testing.
// It resets the runtime after the test completes.
func initAcrRegistry(t *testing.T, mapping config.ACRAMRMappingConfig) {
	t.Helper()
	config.ResetThunderRuntime()
	require.NoError(t, config.InitializeThunderRuntime("", &config.Config{
		ACRAMRMapping: mapping,
	}))
	t.Cleanup(config.ResetThunderRuntime)
}

var validAcrMapping = config.ACRAMRMappingConfig{
	AMR: map[string]config.AMRFactor{
		"Password": {Type: "PWD"},
		"OTP":      {Type: "OTP"},
	},
	AcrAMR: map[string][]string{
		"mosip:idp:acr:password":       {"Password"},
		"mosip:idp:acr:generated-code": {"OTP"},
	},
}

func TestIsValidACR_KnownACR(t *testing.T) {
	initAcrRegistry(t, validAcrMapping)

	ok := isValidACR("mosip:idp:acr:password")

	assert.True(t, ok)
}

func TestIsValidACR_UnknownACR(t *testing.T) {
	initAcrRegistry(t, validAcrMapping)

	ok := isValidACR("mosip:idp:acr:unknown")

	assert.False(t, ok)
}

func TestIsValidACR_EmptyString(t *testing.T) {
	initAcrRegistry(t, validAcrMapping)

	ok := isValidACR("")

	assert.False(t, ok)
}

func TestIsValidACR_AllMappedACRs(t *testing.T) {
	initAcrRegistry(t, validAcrMapping)

	knownACRs := []string{
		"mosip:idp:acr:password",
		"mosip:idp:acr:generated-code",
	}
	for _, acr := range knownACRs {
		ok := isValidACR(acr)
		assert.True(t, ok, "expected ACR %q to be valid", acr)
	}
}

func TestIsValidACR_EmptyMapping(t *testing.T) {
	initAcrRegistry(t, config.ACRAMRMappingConfig{})

	ok := isValidACR("mosip:idp:acr:password")

	assert.False(t, ok)
}
