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

package scope

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ScopeValidatorTestSuite struct {
	suite.Suite
	validator ScopeValidatorInterface
}

func TestScopeValidatorSuite(t *testing.T) {
	suite.Run(t, new(ScopeValidatorTestSuite))
}

func (suite *ScopeValidatorTestSuite) SetupTest() {
	suite.validator = newAPIScopeValidator()
}

func (suite *ScopeValidatorTestSuite) TestNewAPIScopeValidator() {
	validator := newAPIScopeValidator()
	assert.NotNil(suite.T(), validator)
	assert.IsType(suite.T(), &apiScopeValidator{}, validator)
}

func (suite *ScopeValidatorTestSuite) TestValidateScopes() {
	testCases := []struct {
		name            string
		requestedScopes string
		clientID        string
		expectedScopes  string
		expectedError   *ScopeError
	}{
		{
			name:            "EmptyScopes",
			requestedScopes: "",
			clientID:        "test-client",
			expectedScopes:  "",
			expectedError:   nil,
		},
		{
			name:            "SingleScope",
			requestedScopes: "read",
			clientID:        "test-client",
			expectedScopes:  "read",
			expectedError:   nil,
		},
		{
			name:            "MultipleScopes",
			requestedScopes: "read write delete",
			clientID:        "test-client",
			expectedScopes:  "read write delete",
			expectedError:   nil,
		},
		{
			name:            "ScopesWithSpecialCharacters",
			requestedScopes: "api:read profile:write",
			clientID:        "test-client",
			expectedScopes:  "api:read profile:write",
			expectedError:   nil,
		},
		{
			name:            "EmptyClientID",
			requestedScopes: "read",
			clientID:        "",
			expectedScopes:  "read",
			expectedError:   nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			scopes, err := suite.validator.ValidateScopes(context.Background(), tc.requestedScopes, tc.clientID)

			assert.Equal(t, tc.expectedScopes, scopes)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func (suite *ScopeValidatorTestSuite) TestValidateScopesInterface() {
	var _ ScopeValidatorInterface = &apiScopeValidator{}

	validator := newAPIScopeValidator()
	scopes, err := validator.ValidateScopes(context.Background(), "test", "client")
	assert.Equal(suite.T(), "test", scopes)
	assert.Nil(suite.T(), err)
}
