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

package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ScopeValidatorTestSuite struct {
	suite.Suite
	validator *APIScopeValidator
}

func TestScopeValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ScopeValidatorTestSuite))
}

func (suite *ScopeValidatorTestSuite) SetupTest() {
	suite.validator = NewAPIScopeValidator()
}

func (suite *ScopeValidatorTestSuite) TestNewAPIScopeValidator() {
	validator := NewAPIScopeValidator()
	assert.NotNil(suite.T(), validator)
	assert.IsType(suite.T(), &APIScopeValidator{}, validator)
}

func (suite *ScopeValidatorTestSuite) TestValidateScopes_EmptyRequestedScopes() {
	validScopes, err := suite.validator.ValidateScopes("", "test_client")

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", validScopes)
}

func (suite *ScopeValidatorTestSuite) TestScopeError_ErrorCreation() {
	// Test ScopeError struct creation
	scopeError := &ScopeError{
		Error:            "invalid_scope",
		ErrorDescription: "The requested scope is invalid",
	}

	assert.Equal(suite.T(), "invalid_scope", scopeError.Error)
	assert.Equal(suite.T(), "The requested scope is invalid", scopeError.ErrorDescription)
}

// Note: Tests involving database interactions would require database mocking
// which is more complex and would need the database provider interfaces
// to be mockable. For now, we're focusing on the basic functionality
// that can be tested without database dependencies.
