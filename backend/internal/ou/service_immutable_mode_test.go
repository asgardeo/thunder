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

package ou

import (
	"testing"

	"github.com/asgardeo/thunder/internal/system/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ImmutableModeServiceTestSuite tests service behavior in immutable mode.
type ImmutableModeServiceTestSuite struct {
	suite.Suite
	service OrganizationUnitServiceInterface
}

func (suite *ImmutableModeServiceTestSuite) SetupTest() {
	// Initialize runtime with immutable mode enabled
	testConfig := &config.Config{
		OrganizationUnit: config.OrganizationUnitConfig{
			Store: "immutable", // Explicit immutable mode
		},
	}
	_ = config.InitializeThunderRuntime("/tmp/test", testConfig)

	// Create service with mock store (store won't be called in immutable mode)
	mockStore := new(organizationUnitStoreInterfaceMock)
	suite.service = newOrganizationUnitService(mockStore)
}

func (suite *ImmutableModeServiceTestSuite) TestCreateOrganizationUnit_FailsInImmutableMode() {
	request := OrganizationUnitRequest{
		Name:        "Test OU",
		Handle:      "test-ou",
		Description: "Test Description",
	}

	ou, err := suite.service.CreateOrganizationUnit(request)

	// Should fail with immutable resource error
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorCannotModifyImmutableResource.Code, err.Code)
	assert.Equal(suite.T(), OrganizationUnit{}, ou)
}

func (suite *ImmutableModeServiceTestSuite) TestUpdateOrganizationUnit_FailsInImmutableMode() {
	request := OrganizationUnitRequest{
		Name:        "Updated OU",
		Handle:      "updated-ou",
		Description: "Updated Description",
	}

	ou, err := suite.service.UpdateOrganizationUnit("ou-1", request)

	// Should fail with immutable resource error
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorCannotModifyImmutableResource.Code, err.Code)
	assert.Equal(suite.T(), OrganizationUnit{}, ou)
}

func (suite *ImmutableModeServiceTestSuite) TestUpdateOrganizationUnitByPath_FailsInImmutableMode() {
	request := OrganizationUnitRequest{
		Name:        "Updated OU",
		Handle:      "updated-ou",
		Description: "Updated Description",
	}

	ou, err := suite.service.UpdateOrganizationUnitByPath("/path/to/ou", request)

	// Should fail because even getting the OU to update will check immutable mode
	// Or fail during update operation
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), OrganizationUnit{}, ou)
}

func (suite *ImmutableModeServiceTestSuite) TestDeleteOrganizationUnit_FailsInImmutableMode() {
	err := suite.service.DeleteOrganizationUnit("ou-1")

	// Should fail with immutable resource error
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorCannotModifyImmutableResource.Code, err.Code)
}

func (suite *ImmutableModeServiceTestSuite) TestDeleteOrganizationUnitByPath_FailsInImmutableMode() {
	err := suite.service.DeleteOrganizationUnitByPath("/path/to/ou")

	// Should fail because even getting the OU to delete will check immutable mode
	// Or fail during delete operation
	assert.NotNil(suite.T(), err)
}

func TestImmutableModeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ImmutableModeServiceTestSuite))
}
