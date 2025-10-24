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

// Package assert provides functionality to generate and verify authentication assertions with support
// for authentication assurance levels (AAL, IAL).
package assert

import (
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

// TODO: Refactor this to be a centralized auth assertion generator with appropriate token generation logics.

const loggerComponentName = "AuthAssertGenerator"

// AuthAssertGeneratorInterface defines the interface for generating auth assertion claims.
type AuthAssertGeneratorInterface interface {
	GenerateAssertion(authenticators []authncm.AuthenticatorReference) (*AssertionResult, *serviceerror.ServiceError)
	UpdateAssertion(context *AssuranceContext, authenticator authncm.AuthenticatorReference) (
		*AssertionResult, *serviceerror.ServiceError)
	VerifyAssurance(context *AssuranceContext, requiredAAL AssuranceLevel, requiredIAL AssuranceLevel) bool
}

// authAssertGenerator implements the AuthAssertGeneratorInterface.
type authAssertGenerator struct{}

// NewAuthAssertGenerator creates a new instance of AuthAssertGeneratorInterface.
func NewAuthAssertGenerator() AuthAssertGeneratorInterface {
	return &authAssertGenerator{}
}

// GenerateAssertion generates authenticator assertion based on the provided authenticators.
func (ag *authAssertGenerator) GenerateAssertion(
	authenticators []authncm.AuthenticatorReference) (*AssertionResult, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Generating authentication assertion")

	if len(authenticators) == 0 {
		logger.Debug("No authenticators provided for assertion generation")
		return nil, &ErrorNoAuthenticators
	}

	uniqueAuthenticatorsMap := ag.extractUniqueAuthenticators(authenticators)
	aal := ag.calculateAAL(uniqueAuthenticatorsMap)
	ial := ag.calculateIAL()

	return &AssertionResult{
		Context: &AssuranceContext{
			AAL:            aal,
			IAL:            ial,
			Authenticators: authenticators,
		},
	}, nil
}

// UpdateAssertion updates existing assurance context with the provided authenticator.
func (ag *authAssertGenerator) UpdateAssertion(context *AssuranceContext,
	authenticator authncm.AuthenticatorReference) (*AssertionResult, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating authentication assertion with new authenticator")

	if context == nil {
		logger.Debug("No existing assurance context found, generating new assertion")
		return ag.GenerateAssertion([]authncm.AuthenticatorReference{authenticator})
	}

	// Validate authenticator name is present
	if authenticator.Authenticator == "" {
		logger.Debug("Invalid authenticator: missing authenticator name")
		return nil, &ErrorInvalidAuthenticator
	}

	// Merge authenticators
	allAuthenticators := make([]authncm.AuthenticatorReference, 0, len(context.Authenticators)+1)
	allAuthenticators = append(allAuthenticators, context.Authenticators...)
	allAuthenticators = append(allAuthenticators, authenticator)

	// Regenerate claims with all authenticators
	return ag.GenerateAssertion(allAuthenticators)
}

// VerifyAssurance verifies if actual assurance meets the required assurance level.
func (ag *authAssertGenerator) VerifyAssurance(context *AssuranceContext, requiredAAL AssuranceLevel,
	requiredIAL AssuranceLevel) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Verifying assurance levels")

	if context == nil {
		return false
	}
	if requiredAAL == "" && requiredIAL == "" {
		logger.Debug("No required assurance levels specified, considering verification as successful")
		return true
	}

	// Check AAL level
	if !ag.meetsAssuranceLevel(context.AAL, requiredAAL) {
		logger.Debug("Actual AAL does not meet required AAL", log.String("actualAAL", string(context.AAL)),
			log.String("requiredAAL", string(requiredAAL)))
		return false
	}

	// Check IAL level
	if !ag.meetsAssuranceLevel(context.IAL, requiredIAL) {
		logger.Debug("Actual IAL does not meet required IAL", log.String("actualIAL", string(context.IAL)),
			log.String("requiredIAL", string(requiredIAL)))
		return false
	}

	return true
}

// extractUniqueAuthenticators extracts unique authenticator names from authenticator references.
func (ag *authAssertGenerator) extractUniqueAuthenticators(
	authenticators []authncm.AuthenticatorReference) map[string]bool {
	authenticatorsMap := make(map[string]bool)
	for _, auth := range authenticators {
		authenticatorsMap[auth.Authenticator] = true
	}

	return authenticatorsMap
}

// calculateAAL calculates the AAL based on the unique authenticators used.
// For single authenticators, uses individual AAL weight from authenticatorRegistry.
// For multiple authenticators, checks against defined valid combinations.
func (ag *authAssertGenerator) calculateAAL(uniqueAuthenticatorsMap map[string]bool) AssuranceLevel {
	authenticatorCount := len(uniqueAuthenticatorsMap)

	// Single authenticator - use individual AAL weight
	if authenticatorCount == 1 {
		authenticator := ""
		for auth := range uniqueAuthenticatorsMap {
			authenticator = auth
			break
		}

		weight := authncm.GetAuthenticatorWeight(authenticator)
		switch weight {
		case 3:
			return AALLevel3
		case 2:
			return AALLevel2
		default:
			return AALLevel1
		}
	}

	// Multiple authenticators - check against valid combinations
	for _, combination := range validAALCombinations {
		if ag.matchesCombination(uniqueAuthenticatorsMap, combination.Authenticators) {
			return combination.AAL
		}
	}

	// No valid combination found - default to AAL1
	return AALLevel1
}

// matchesCombination checks if the provided authenticators exactly match a defined combination.
// Order doesn't matter, but all authenticators in the combination must be present.
func (ag *authAssertGenerator) matchesCombination(actualAuthenticatorsMap map[string]bool,
	requiredAuthenticators []string) bool {
	if len(actualAuthenticatorsMap) != len(requiredAuthenticators) {
		return false
	}

	for _, required := range requiredAuthenticators {
		if !actualAuthenticatorsMap[required] {
			return false
		}
	}

	return true
}

// calculateIAL calculates the IAL based on authenticators.
// For now, returns default IAL1. Can be enhanced based on user verification status.
func (ag *authAssertGenerator) calculateIAL() AssuranceLevel {
	// Default implementation - can be enhanced to check user verification status
	// For example, check if email/phone is verified, document verification, etc.
	return IALLevel1
}

// meetsAssuranceLevel checks if actual assurance level meets or exceeds the required level.
func (ag *authAssertGenerator) meetsAssuranceLevel(actual, required AssuranceLevel) bool {
	actualLevel := assuranceLevelOrder[actual]
	requiredLevel := assuranceLevelOrder[required]

	return actualLevel >= requiredLevel
}
