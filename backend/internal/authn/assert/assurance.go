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

package assert

import (
	authncm "github.com/asgardeo/thunder/internal/authn/common"
)

// assuranceLevelOrder maps assurance levels to their hierarchical order for comparison.
var assuranceLevelOrder = map[AssuranceLevel]int{
	AALLevel1: 1,
	AALLevel2: 2,
	AALLevel3: 3,
	IALLevel1: 1,
	IALLevel2: 2,
	IALLevel3: 3,
}

// authenticatorCombination defines a valid combination of authenticators and their resulting AAL.
type authenticatorCombination struct {
	// Authenticators is a list of authenticator names that must be present in the combination.
	// Order doesn't matter - combinations are matched based on presence of all required authenticators.
	Authenticators []string
	// AAL is the resulting authenticator assurance level for this combination.
	AAL AssuranceLevel
	// Description provides human-readable context for this combination.
	Description string
}

// validAALCombinations defines all valid authenticator combinations and their resulting AAL.
// Single authenticator combinations inherit their individual AAL weight from authenticatorRegistry.
// Multi-factor combinations are explicitly defined here.
var validAALCombinations = []authenticatorCombination{
	// AAL2 - Multi-factor combinations
	{
		Authenticators: []string{authncm.AuthenticatorCredentials, authncm.AuthenticatorSMSOTP},
		AAL:            AALLevel2,
		Description:    "Password + SMS OTP (MFA)",
	},
	{
		Authenticators: []string{authncm.AuthenticatorGoogle, authncm.AuthenticatorSMSOTP},
		AAL:            AALLevel2,
		Description:    "Google + SMS OTP (MFA)",
	},
	{
		Authenticators: []string{authncm.AuthenticatorGithub, authncm.AuthenticatorSMSOTP},
		AAL:            AALLevel2,
		Description:    "GitHub + SMS OTP (MFA)",
	},
	{
		Authenticators: []string{authncm.AuthenticatorOAuth, authncm.AuthenticatorSMSOTP},
		AAL:            AALLevel2,
		Description:    "OAuth + SMS OTP (MFA)",
	},
	{
		Authenticators: []string{authncm.AuthenticatorOIDC, authncm.AuthenticatorSMSOTP},
		AAL:            AALLevel2,
		Description:    "OIDC + SMS OTP (MFA)",
	},
}
