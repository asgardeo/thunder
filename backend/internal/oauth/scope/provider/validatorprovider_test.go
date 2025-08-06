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

package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/oauth/scope/validator"
)

type ScopeValidatorProviderTestSuite struct {
	suite.Suite
	provider ScopeValidatorProviderInterface
}

func TestScopeValidatorProviderSuite(t *testing.T) {
	suite.Run(t, new(ScopeValidatorProviderTestSuite))
}

func (suite *ScopeValidatorProviderTestSuite) SetupTest() {
	suite.provider = NewScopeValidatorProvider()
}

func (suite *ScopeValidatorProviderTestSuite) TestNewScopeValidatorProvider() {
	provider := NewScopeValidatorProvider()
	assert.NotNil(suite.T(), provider)
	assert.IsType(suite.T(), &ScopeValidatorProvider{}, provider)
}

func (suite *ScopeValidatorProviderTestSuite) TestGetScopeValidator() {
	scopeValidator := suite.provider.GetScopeValidator()
	
	assert.NotNil(suite.T(), scopeValidator)
	assert.Implements(suite.T(), (*validator.ScopeValidatorInterface)(nil), scopeValidator)
}

func (suite *ScopeValidatorProviderTestSuite) TestGetScopeValidatorReturnsConsistentInstance() {
	// Test that multiple calls return the same type of validator
	validator1 := suite.provider.GetScopeValidator()
	validator2 := suite.provider.GetScopeValidator()
	
	assert.NotNil(suite.T(), validator1)
	assert.NotNil(suite.T(), validator2)
	assert.IsType(suite.T(), validator1, validator2)
}

func (suite *ScopeValidatorProviderTestSuite) TestScopeValidatorProviderInterface() {
	// Test that ScopeValidatorProvider implements ScopeValidatorProviderInterface
	var _ ScopeValidatorProviderInterface = &ScopeValidatorProvider{}
	
	// Verify the provider can be used through the interface
	var provider ScopeValidatorProviderInterface = NewScopeValidatorProvider()
	scopeValidator := provider.GetScopeValidator()
	
	assert.NotNil(suite.T(), scopeValidator)
	assert.Implements(suite.T(), (*validator.ScopeValidatorInterface)(nil), scopeValidator)
}

func (suite *ScopeValidatorProviderTestSuite) TestGetScopeValidatorFunctionality() {
	// Test that the returned validator actually works
	scopeValidator := suite.provider.GetScopeValidator()
	
	result, err := scopeValidator.ValidateScopes("read write", "test-client")
	
	assert.Equal(suite.T(), "read write", result)
	assert.Nil(suite.T(), err)
}
