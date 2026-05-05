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

// Package scope provides functionality for validating scopes.
package scope

import "context"

// ScopeError represents an error during scope validation.
type ScopeError struct {
	Error            string
	ErrorDescription string
}

// ScopeValidatorInterface defines the interface for scope validation.
type ScopeValidatorInterface interface {
	ValidateScopes(ctx context.Context, requestedScopes, clientID string) (string, *ScopeError)
}

// apiScopeValidator is the implementation of API scope validation.
type apiScopeValidator struct{}

// newAPIScopeValidator creates a new instance of the apiScopeValidator.
func newAPIScopeValidator() *apiScopeValidator {
	return &apiScopeValidator{}
}

// ValidateScopes validates and filters the requested scopes against the authorized scopes for the application.
func (sv *apiScopeValidator) ValidateScopes(
	ctx context.Context, requestedScopes, clientID string,
) (string, *ScopeError) {
	if requestedScopes == "" {
		return "", nil
	}

	// Return all requested scopes for now.
	return requestedScopes, nil
}
