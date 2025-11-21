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

package resource

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

const (
	testParentResourceID = "parent-123"
	testOriginalName     = "original-name"
	testOriginalHandle   = "original-handle"
	testUpdatedName      = "updated-name"
	testNewDescription   = "new description"
)

// Test Suite
type ResourceServiceTestSuite struct {
	suite.Suite
	mockStore *resourceStoreInterfaceMock
	mockOU    *oumock.OrganizationUnitServiceInterfaceMock
	service   ResourceServiceInterface
}

func TestResourceServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceServiceTestSuite))
}

func (suite *ResourceServiceTestSuite) SetupTest() {
	suite.mockStore = newResourceStoreInterfaceMock(suite.T())
	suite.mockOU = new(oumock.OrganizationUnitServiceInterfaceMock)
	suite.service = newResourceService(suite.mockStore, suite.mockOU)
}

// Resource Server Tests

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_Success() {
	rs := ResourceServer{
		Name:               "test-rs",
		Description:        "Test resource server",
		Identifier:         "test-identifier",
		OrganizationUnitID: "ou-123",
	}

	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExists", "test-rs").
		Return(false, nil)
	suite.mockStore.On("CheckResourceServerIdentifierExists", "test-identifier").
		Return(false, nil)
	suite.mockStore.On("CreateResourceServer", mock.AnythingOfType("string"), rs).
		Return(nil)

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(err)
	suite.NotNil(result)
	suite.NotEmpty(result.ID)
	suite.Equal("test-rs", result.Name)
	suite.Equal("Test resource server", result.Description)
	suite.mockStore.AssertExpectations(suite.T())
	suite.mockOU.AssertExpectations(suite.T())
}

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_InvalidInput_EmptyName() {
	rs := ResourceServer{
		Name:               "",
		OrganizationUnitID: "ou-123",
	}

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidRequestFormat.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_InvalidInput_EmptyOU() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "",
	}

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidRequestFormat.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_OUNotFound() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound)

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorOrganizationUnitNotFound.Code, err.Code)
	suite.mockOU.AssertExpectations(suite.T())
}

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_OUServiceError() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{}, &serviceerror.InternalServerError)

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_NameConflict() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExists", "test-rs").
		Return(true, nil)

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorNameConflict.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_StoreError() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
		Identifier:         "", // Empty identifier - no need to check
	}

	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExists", "test-rs").
		Return(false, nil)
	// No identifier check needed since identifier is empty
	suite.mockStore.On("CreateResourceServer", mock.AnythingOfType("string"), rs).
		Return(errors.New("database error"))

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_IdentifierConflict() {
	rs := ResourceServer{
		Name:               "test-rs",
		Identifier:         "test-identifier",
		OrganizationUnitID: "ou-123",
	}

	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExists", "test-rs").
		Return(false, nil)
	suite.mockStore.On("CheckResourceServerIdentifierExists", "test-identifier").
		Return(true, nil)

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorIdentifierConflict.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_CheckNameError() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExists", "test-rs").
		Return(false, errors.New("database error"))

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResourceServer_CheckIdentifierError() {
	rs := ResourceServer{
		Name:               "test-rs",
		Identifier:         "test-identifier",
		OrganizationUnitID: "ou-123",
	}

	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExists", "test-rs").
		Return(false, nil)
	suite.mockStore.On("CheckResourceServerIdentifierExists", "test-identifier").
		Return(false, errors.New("database error"))

	result, err := suite.service.CreateResourceServer(rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceServer_Success() {
	expectedRS := ResourceServer{
		ID:                 "rs-123",
		Name:               "test-rs",
		Description:        "Test",
		OrganizationUnitID: "ou-123",
	}

	suite.mockStore.On("GetResourceServer", "rs-123").
		Return(expectedRS, nil)

	result, err := suite.service.GetResourceServer("rs-123")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("rs-123", result.ID)
	suite.Equal("test-rs", result.Name)
}

func (suite *ResourceServiceTestSuite) TestGetResourceServer_MissingID() {
	result, err := suite.service.GetResourceServer("")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceServer_NotFound() {
	suite.mockStore.On("GetResourceServer", "rs-123").
		Return(ResourceServer{}, ErrResourceServerNotFound)

	result, err := suite.service.GetResourceServer("rs-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceServer_StoreError() {
	suite.mockStore.On("GetResourceServer", "rs-123").
		Return(ResourceServer{}, errors.New("database error"))

	result, err := suite.service.GetResourceServer("rs-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceServerList_Success() {
	resourceServers := []ResourceServer{
		{ID: "rs-1", Name: "RS 1"},
		{ID: "rs-2", Name: "RS 2"},
	}

	suite.mockStore.On("GetResourceServerListCount").Return(2, nil)
	suite.mockStore.On("GetResourceServerList", 30, 0).Return(resourceServers, nil)

	result, err := suite.service.GetResourceServerList(30, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, result.Count)
	suite.Equal(2, len(result.ResourceServers))
}

// TestGetResourceServerList_InvalidPagination removed - redundant with TestValidatePaginationParams

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_Success() {
	rs := ResourceServer{
		Name:               "updated-rs",
		Description:        "Updated",
		OrganizationUnitID: "ou-123",
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExistsExcludingID", "updated-rs", "rs-123").
		Return(false, nil)
	suite.mockStore.On("UpdateResourceServer", "rs-123", rs).Return(nil)

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("rs-123", result.ID)
	suite.Equal("updated-rs", result.Name)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_NotFound() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_MissingID() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	result, err := suite.service.UpdateResourceServer("", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_InvalidInput_EmptyName() {
	rs := ResourceServer{
		Name:               "",
		OrganizationUnitID: "ou-123",
	}

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidRequestFormat.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_InvalidInput_EmptyOU() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "",
	}

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidRequestFormat.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_OUNotFound() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound)

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorOrganizationUnitNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_OUServiceError() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{}, &serviceerror.InternalServerError)

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_NameConflict() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExistsExcludingID", "test-rs", "rs-123").
		Return(true, nil)

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorNameConflict.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_IdentifierConflict() {
	rs := ResourceServer{
		Name:               "test-rs",
		Identifier:         "test-identifier",
		OrganizationUnitID: "ou-123",
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExistsExcludingID", "test-rs", "rs-123").
		Return(false, nil)
	suite.mockStore.On("CheckResourceServerIdentifierExistsExcludingID", "test-identifier", "rs-123").
		Return(true, nil)

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorIdentifierConflict.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_CheckIdentifierError() {
	rs := ResourceServer{
		Name:               "test-rs",
		Identifier:         "test-identifier",
		OrganizationUnitID: "ou-123",
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExistsExcludingID", "test-rs", "rs-123").
		Return(false, nil)
	suite.mockStore.On("CheckResourceServerIdentifierExistsExcludingID", "test-identifier", "rs-123").
		Return(false, errors.New("database error"))

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResourceServer_StoreError() {
	rs := ResourceServer{
		Name:               "test-rs",
		OrganizationUnitID: "ou-123",
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockOU.On("GetOrganizationUnit", "ou-123").
		Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil)
	suite.mockStore.On("CheckResourceServerNameExistsExcludingID", "test-rs", "rs-123").
		Return(false, nil)
	suite.mockStore.On("UpdateResourceServer", "rs-123", rs).
		Return(errors.New("database error"))

	result, err := suite.service.UpdateResourceServer("rs-123", rs)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteResourceServer_Success() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("CheckResourceServerHasDependencies", "rs-123").Return(false, nil)
	suite.mockStore.On("DeleteResourceServer", "rs-123").Return(nil)

	err := suite.service.DeleteResourceServer("rs-123")

	suite.Nil(err)
}

func (suite *ResourceServiceTestSuite) TestDeleteResourceServer_IdempotentWhenNotExists() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	err := suite.service.DeleteResourceServer("rs-123")

	suite.Nil(err)
}

func (suite *ResourceServiceTestSuite) TestDeleteResourceServer_MissingID() {
	err := suite.service.DeleteResourceServer("")

	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteResourceServer_CheckExistenceError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, errors.New("database error"))

	err := suite.service.DeleteResourceServer("rs-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteResourceServer_CheckDependenciesError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("CheckResourceServerHasDependencies", "rs-123").Return(false, errors.New("database error"))

	err := suite.service.DeleteResourceServer("rs-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteResourceServer_DeleteError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("CheckResourceServerHasDependencies", "rs-123").Return(false, nil)
	suite.mockStore.On("DeleteResourceServer", "rs-123").Return(errors.New("database error"))

	err := suite.service.DeleteResourceServer("rs-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteResourceServer_HasDependencies() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("CheckResourceServerHasDependencies", "rs-123").Return(true, nil)

	err := suite.service.DeleteResourceServer("rs-123")

	suite.NotNil(err)
	suite.Equal(ErrorCannotDelete.Code, err.Code)
}

// Resource Tests

func (suite *ResourceServiceTestSuite) TestCreateResource_Success() {
	res := Resource{
		Name:   "test-resource",
		Handle: "test-handle",
		Parent: nil,
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceHandleExistsUnderParent", "rs-123", "test-handle", (*string)(nil)).
		Return(false, nil)
	suite.mockStore.On("CreateResource", mock.AnythingOfType("string"), 123, (*int)(nil), res).
		Return(nil)

	result, err := suite.service.CreateResource("rs-123", res)

	suite.Nil(err)
	suite.NotNil(result)
	suite.NotEmpty(result.ID)
	suite.Equal("test-resource", result.Name)
	suite.Equal("test-handle", result.Handle)
}

func (suite *ResourceServiceTestSuite) TestCreateResource_EmptyName() {
	res := Resource{
		Name:   "",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)

	result, err := suite.service.CreateResource("rs-123", res)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidRequestFormat.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResource_WithParent_Success() {
	parentID := testParentResourceID
	res := Resource{
		Name:   "test-resource",
		Handle: "test-handle",
		Parent: &parentID,
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", testParentResourceID, "rs-123").Return(456, nil)
	suite.mockStore.On("CheckResourceHandleExistsUnderParent", "rs-123", "test-handle", &parentID).
		Return(false, nil)
	parentInternalID := 456
	suite.mockStore.On("CreateResource", mock.AnythingOfType("string"), 123, &parentInternalID, res).
		Return(nil)

	result, err := suite.service.CreateResource("rs-123", res)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("test-resource", result.Name)
	suite.Equal("test-handle", result.Handle)
	suite.Equal(&parentID, result.Parent)
}

func (suite *ResourceServiceTestSuite) TestCreateResource_ParentNotFound() {
	parentID := testParentResourceID
	res := Resource{
		Name:   "test-resource",
		Handle: "test-handle",
		Parent: &parentID,
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", testParentResourceID, "rs-123").Return(0, ErrResourceNotFound)

	result, err := suite.service.CreateResource("rs-123", res)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorParentResourceNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResource_ResourceServerNotFound() {
	res := Resource{
		Name:   "test-resource",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(0, ErrResourceServerNotFound)

	result, err := suite.service.CreateResource("rs-123", res)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResource_HandleConflict() {
	res := Resource{
		Name:   "test-resource",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceHandleExistsUnderParent", "rs-123", "test-handle", (*string)(nil)).
		Return(true, nil)

	result, err := suite.service.CreateResource("rs-123", res)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorHandleConflict.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResource_StoreError() {
	res := Resource{
		Name:   "test-resource",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceHandleExistsUnderParent", "rs-123", "test-handle", (*string)(nil)).
		Return(false, nil)
	suite.mockStore.On("CreateResource", mock.AnythingOfType("string"), 123, (*int)(nil), res).
		Return(errors.New("database error"))

	result, err := suite.service.CreateResource("rs-123", res)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResource_CheckHandleError() {
	res := Resource{
		Name:   "test-resource",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceHandleExistsUnderParent", "rs-123", "test-handle", (*string)(nil)).
		Return(false, errors.New("database error"))

	result, err := suite.service.CreateResource("rs-123", res)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResource_ParentCheckError() {
	parentID := testParentResourceID
	res := Resource{
		Name:   "test-resource",
		Handle: "test-handle",
		Parent: &parentID,
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", parentID, "rs-123").
		Return(0, errors.New("database error"))

	result, err := suite.service.CreateResource("rs-123", res)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateResource_CircularDependency_SelfReference() {
	// Test creating a resource with itself as parent
	res := Resource{
		Name:   "test-resource",
		Handle: "test-handle",
		Parent: nil, // Will be set to its own ID after creation attempt
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceHandleExistsUnderParent", "rs-123", "test-handle", (*string)(nil)).
		Return(false, nil)
	suite.mockStore.On("CreateResource", mock.AnythingOfType("string"), 123, (*int)(nil), res).
		Return(nil)

	result, err := suite.service.CreateResource("rs-123", res)

	// Should succeed initially - circular check would need to be in update
	suite.Nil(err)
	suite.NotNil(result)
}

func (suite *ResourceServiceTestSuite) TestUpdateResource_Success() {
	currentResource := Resource{
		ID:          "res-123",
		Name:        testOriginalName,
		Handle:      testOriginalHandle,
		Description: "old description",
	}

	updateReq := Resource{
		Name:        testUpdatedName,
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResource", "res-123", "rs-123").Return(currentResource, nil)
	suite.mockStore.On("UpdateResource", "res-123", "rs-123", mock.MatchedBy(func(r Resource) bool {
		return r.Name == testUpdatedName && r.Handle == testOriginalHandle && r.Description == testNewDescription
	})).Return(nil)

	result, err := suite.service.UpdateResource("rs-123", "res-123", updateReq)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("res-123", result.ID)
	suite.Equal(testUpdatedName, result.Name)
	suite.Equal(testOriginalHandle, result.Handle) // Handle is immutable
	suite.Equal(testNewDescription, result.Description)
}

func (suite *ResourceServiceTestSuite) TestUpdateResource_ParentIsImmutable() {
	// Parent is immutable and preserved from current resource, so no circular dependency check needed
	currentResource := Resource{
		ID:          "res-123",
		Name:        testOriginalName,
		Handle:      testOriginalHandle,
		Description: "old description",
		Parent:      nil,
	}

	newParentID := testParentResourceID
	updateReq := Resource{
		Name:        testUpdatedName,
		Description: testNewDescription,
		Parent:      &newParentID, // This will be ignored, parent is immutable
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResource", "res-123", "rs-123").Return(currentResource, nil)
	suite.mockStore.On("UpdateResource", "res-123", "rs-123", mock.MatchedBy(func(r Resource) bool {
		// Parent should be preserved from current resource (nil), not from updateReq
		return r.Name == testUpdatedName && r.Handle == testOriginalHandle && r.Parent == nil
	})).Return(nil)

	result, err := suite.service.UpdateResource("rs-123", "res-123", updateReq)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("res-123", result.ID)
	suite.Nil(result.Parent) // Parent is preserved from current resource
}

func (suite *ResourceServiceTestSuite) TestUpdateResource_MissingID() {
	updateReq := Resource{
		Description: testNewDescription,
	}

	result, err := suite.service.UpdateResource("", "res-123", updateReq)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	result, err = suite.service.UpdateResource("rs-123", "", updateReq)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResource_ResourceNotFound() {
	updateReq := Resource{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(false, nil)

	result, err := suite.service.UpdateResource("rs-123", "res-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResource_HandleIsImmutable() {
	// Handle is immutable and preserved from current resource
	currentResource := Resource{
		ID:          "res-123",
		Name:        testOriginalName,
		Handle:      testOriginalHandle,
		Description: "old description",
	}

	updateReq := Resource{
		Name:        testUpdatedName,
		Handle:      "new-handle", // This will be ignored, handle is immutable
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResource", "res-123", "rs-123").Return(currentResource, nil)
	suite.mockStore.On("UpdateResource", "res-123", "rs-123", mock.MatchedBy(func(r Resource) bool {
		// Handle should be preserved from current resource, not from updateReq
		return r.Handle == testOriginalHandle && r.Name == testUpdatedName
	})).Return(nil)

	result, err := suite.service.UpdateResource("rs-123", "res-123", updateReq)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(testOriginalHandle, result.Handle) // Handle is preserved
}

func (suite *ResourceServiceTestSuite) TestUpdateResource_StoreError() {
	currentResource := Resource{
		ID:          "res-123",
		Name:        testOriginalName,
		Handle:      testOriginalHandle,
		Description: "old description",
	}

	updateReq := Resource{
		Name:        testUpdatedName,
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResource", "res-123", "rs-123").Return(currentResource, nil)
	suite.mockStore.On("UpdateResource", "res-123", "rs-123", mock.Anything).Return(errors.New("database error"))

	result, err := suite.service.UpdateResource("rs-123", "res-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResource_GetResourceError() {
	updateReq := Resource{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResource", "res-123", "rs-123").Return(Resource{}, errors.New("database error"))

	result, err := suite.service.UpdateResource("rs-123", "res-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResource_ResourceServerNotFound() {
	updateReq := Resource{
		Name:        testUpdatedName,
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.UpdateResource("rs-123", "res-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateResource_CheckServerError() {
	updateReq := Resource{
		Name:        testUpdatedName,
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.UpdateResource("rs-123", "res-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteResource_Success() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(true, nil)
	suite.mockStore.On("CheckResourceHasDependencies", "res-123").Return(false, nil)
	suite.mockStore.On("DeleteResource", "res-123", "rs-123").Return(nil)

	err := suite.service.DeleteResource("rs-123", "res-123")

	suite.Nil(err)
}

func (suite *ResourceServiceTestSuite) TestDeleteResource_HasDependencies() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(true, nil)
	suite.mockStore.On("CheckResourceHasDependencies", "res-123").Return(true, nil)

	err := suite.service.DeleteResource("rs-123", "res-123")

	suite.NotNil(err)
	suite.Equal(ErrorCannotDelete.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteResource_MissingID() {
	err := suite.service.DeleteResource("", "res-123")
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	err = suite.service.DeleteResource("rs-123", "")
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteResource_Idempotent() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(false, nil)

	err := suite.service.DeleteResource("rs-123", "res-123")

	suite.Nil(err)
}

func (suite *ResourceServiceTestSuite) TestDeleteResource_DeleteError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(true, nil)
	suite.mockStore.On("CheckResourceHasDependencies", "res-123").Return(false, nil)
	suite.mockStore.On("DeleteResource", "res-123", "rs-123").Return(errors.New("database error"))

	err := suite.service.DeleteResource("rs-123", "res-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteResource_CheckExistenceError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", "res-123", "rs-123").Return(false, errors.New("database error"))

	err := suite.service.DeleteResource("rs-123", "res-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// GetResource Tests

func (suite *ResourceServiceTestSuite) TestGetResource_Success() {
	expectedRes := Resource{
		ID:          "res-123",
		Name:        "test-resource",
		Description: "Test",
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResource", "res-123", "rs-123").Return(expectedRes, nil)

	result, err := suite.service.GetResource("rs-123", "res-123")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("res-123", result.ID)
	suite.Equal("test-resource", result.Name)
}

func (suite *ResourceServiceTestSuite) TestGetResource_MissingID() {
	result, err := suite.service.GetResource("", "res-123")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	result, err = suite.service.GetResource("rs-123", "")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResource_ResourceServerNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.GetResource("rs-123", "res-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResource_NotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResource", "res-123", "rs-123").Return(Resource{}, ErrResourceNotFound)

	result, err := suite.service.GetResource("rs-123", "res-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResource_StoreError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResource", "res-123", "rs-123").Return(Resource{}, errors.New("database error"))

	result, err := suite.service.GetResource("rs-123", "res-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResource_CheckServerError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.GetResource("rs-123", "res-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// GetResourceList Tests

func (suite *ResourceServiceTestSuite) TestGetResourceList_NoFilter_Success() {
	resources := []Resource{
		{ID: "res-1", Name: "Resource 1"},
		{ID: "res-2", Name: "Resource 2"},
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResourceListCountByParent", "rs-123", (*string)(nil)).Return(2, nil)
	suite.mockStore.On("GetResourceListByParent", "rs-123", (*string)(nil), 30, 0).Return(resources, nil)

	result, err := suite.service.GetResourceList("rs-123", nil, 30, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, result.Count)
	suite.Equal(2, len(result.Resources))
}

func (suite *ResourceServiceTestSuite) TestGetResourceList_ByParent_Success() {
	parentID := testParentResourceID
	resources := []Resource{
		{ID: "res-1", Name: "Resource 1", Parent: &parentID},
		{ID: "res-2", Name: "Resource 2", Parent: &parentID},
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", parentID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetResourceListCountByParent", "rs-123", &parentID).Return(2, nil)
	suite.mockStore.On("GetResourceListByParent", "rs-123", &parentID, 30, 0).Return(resources, nil)

	result, err := suite.service.GetResourceList("rs-123", &parentID, 30, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, len(result.Resources))
}

func (suite *ResourceServiceTestSuite) TestGetResourceList_ByEmptyParent_Success() {
	emptyParent := ""
	resources := []Resource{
		{ID: "res-1", Name: "Top Level 1"},
		{ID: "res-2", Name: "Top Level 2"},
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResourceListCountByParent", "rs-123", &emptyParent).Return(2, nil)
	suite.mockStore.On("GetResourceListByParent", "rs-123", &emptyParent, 30, 0).Return(resources, nil)

	result, err := suite.service.GetResourceList("rs-123", &emptyParent, 30, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, len(result.Resources))
}

func (suite *ResourceServiceTestSuite) TestGetResourceList_ResourceServerNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.GetResourceList("rs-123", nil, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceList_StoreError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResourceListCountByParent", "rs-123", (*string)(nil)).Return(0, errors.New("database error"))

	result, err := suite.service.GetResourceList("rs-123", nil, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceList_CountByParentError() {
	parentID := testParentResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", parentID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetResourceListCountByParent", "rs-123", &parentID).
		Return(0, errors.New("database error"))

	result, err := suite.service.GetResourceList("rs-123", &parentID, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceList_ListByParentError() {
	parentID := testParentResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", parentID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetResourceListCountByParent", "rs-123", &parentID).Return(10, nil)
	suite.mockStore.On("GetResourceListByParent", "rs-123", &parentID, 30, 0).
		Return(nil, errors.New("database error"))

	result, err := suite.service.GetResourceList("rs-123", &parentID, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceList_ParentResourceNotFound() {
	parentID := testParentResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", parentID, "rs-123").Return(false, nil)

	result, err := suite.service.GetResourceList("rs-123", &parentID, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceList_CheckParentExistError() {
	parentID := testParentResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", parentID, "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.GetResourceList("rs-123", &parentID, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceList_ListError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetResourceListCountByParent", "rs-123", (*string)(nil)).Return(10, nil)
	suite.mockStore.On("GetResourceListByParent", "rs-123", (*string)(nil), 30, 0).
		Return([]Resource{}, errors.New("database error"))

	result, err := suite.service.GetResourceList("rs-123", nil, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// Action Tests

func (suite *ResourceServiceTestSuite) TestCreateActionAtResourceServer_Success() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckActionHandleExists", "rs-123", (*string)(nil), "test-handle").
		Return(false, nil)
	suite.mockStore.On("CreateAction", mock.AnythingOfType("string"), 123, (*int)(nil), action).
		Return(nil)

	result, err := suite.service.CreateActionAtResourceServer("rs-123", action)

	suite.Nil(err)
	suite.NotNil(result)
	suite.NotEmpty(result.ID)
	suite.Equal("test-action", result.Name)
	suite.Equal("test-handle", result.Handle)
	suite.Nil(result.ResourceID)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResourceServer_EmptyName() {
	action := Action{
		Name: "",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)

	result, err := suite.service.CreateActionAtResourceServer("rs-123", action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidRequestFormat.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResourceServer_ResourceServerNotFound() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(0, ErrResourceServerNotFound)

	result, err := suite.service.CreateActionAtResourceServer("rs-123", action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResourceServer_HandleConflict() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckActionHandleExists", "rs-123", (*string)(nil), "test-handle").
		Return(true, nil)

	result, err := suite.service.CreateActionAtResourceServer("rs-123", action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorHandleConflict.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResourceServer_StoreError() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckActionHandleExists", "rs-123", (*string)(nil), "test-handle").
		Return(false, nil)
	suite.mockStore.On("CreateAction", mock.AnythingOfType("string"), 123, (*int)(nil), action).
		Return(errors.New("database error"))

	result, err := suite.service.CreateActionAtResourceServer("rs-123", action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResourceServer_CheckHandleError() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckActionHandleExists", "rs-123", (*string)(nil), "test-handle").
		Return(false, errors.New("database error"))

	result, err := suite.service.CreateActionAtResourceServer("rs-123", action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResource_EmptyName() {
	action := Action{
		Name:   "",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", "res-123", "rs-123").Return(789, nil)

	result, err := suite.service.CreateActionAtResource("rs-123", "res-123", action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidRequestFormat.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResource_Success() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", "res-123", "rs-123").Return(789, nil)
	suite.mockStore.On("CheckActionHandleExists", "rs-123", mock.MatchedBy(func(id *string) bool {
		return id != nil && *id == "res-123"
	}), "test-handle").Return(false, nil)
	resourceInternalID := 789
	suite.mockStore.On("CreateAction", mock.AnythingOfType("string"), 123, &resourceInternalID, action).Return(nil)

	result, err := suite.service.CreateActionAtResource("rs-123", "res-123", action)

	suite.Nil(err)
	suite.NotNil(result)
	suite.NotNil(result.ResourceID)
	suite.Equal("res-123", *result.ResourceID)
	suite.Equal("test-action", result.Name)
	suite.Equal("test-handle", result.Handle)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResource_ResourceServerNotFound() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(0, ErrResourceServerNotFound)

	result, err := suite.service.CreateActionAtResource("rs-123", testResourceID, action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResource_ResourceNotFound() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", testResourceID, "rs-123").Return(0, ErrResourceNotFound)

	result, err := suite.service.CreateActionAtResource("rs-123", testResourceID, action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResource_HandleConflict() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", testResourceID, "rs-123").Return(789, nil)
	suite.mockStore.On("CheckActionHandleExists", "rs-123", mock.MatchedBy(func(id *string) bool {
		return id != nil && *id == testResourceID
	}), "test-handle").Return(true, nil)

	result, err := suite.service.CreateActionAtResource("rs-123", testResourceID, action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorHandleConflict.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResource_StoreError() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", testResourceID, "rs-123").Return(789, nil)
	suite.mockStore.On("CheckActionHandleExists", "rs-123", mock.MatchedBy(func(id *string) bool {
		return id != nil && *id == testResourceID
	}), "test-handle").Return(false, nil)
	resourceInternalID := 789
	suite.mockStore.On("CreateAction", mock.AnythingOfType("string"), 123, &resourceInternalID, action).
		Return(errors.New("database error"))

	result, err := suite.service.CreateActionAtResource("rs-123", testResourceID, action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResource_CheckHandleError() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", testResourceID, "rs-123").Return(789, nil)
	suite.mockStore.On("CheckActionHandleExists", "rs-123", mock.MatchedBy(func(id *string) bool {
		return id != nil && *id == testResourceID
	}), "test-handle").Return(false, errors.New("database error"))

	result, err := suite.service.CreateActionAtResource("rs-123", testResourceID, action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestCreateActionAtResource_CheckResourceError() {
	action := Action{
		Name:   "test-action",
		Handle: "test-handle",
	}

	suite.mockStore.On("CheckResourceServerExistAndGetInternalID", "rs-123").Return(123, nil)
	suite.mockStore.On("CheckResourceExistAndGetInternalID", testResourceID, "rs-123").
		Return(0, errors.New("database error"))

	result, err := suite.service.CreateActionAtResource("rs-123", testResourceID, action)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResourceServer_Success() {
	expectedAction := Action{
		ID:         "action-123",
		Name:       "test-action",
		ResourceID: nil,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", (*string)(nil)).
		Return(expectedAction, nil)

	result, err := suite.service.GetActionAtResourceServer("rs-123", "action-123")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("action-123", result.ID)
	suite.Nil(result.ResourceID)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResourceServer_MissingID() {
	result, err := suite.service.GetActionAtResourceServer("", "action-123")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	result, err = suite.service.GetActionAtResourceServer("rs-123", "")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResourceServer_NotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", (*string)(nil)).
		Return(Action{}, ErrActionNotFound)

	result, err := suite.service.GetActionAtResourceServer("rs-123", "action-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorActionNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionListAtResourceServer_Success() {
	actions := []Action{
		{ID: "action-1", Name: "Action 1"},
		{ID: "action-2", Name: "Action 2"},
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetActionListCountAtResourceServer", "rs-123").Return(2, nil)
	suite.mockStore.On("GetActionListAtResourceServer", "rs-123", 30, 0).Return(actions, nil)

	result, err := suite.service.GetActionListAtResourceServer("rs-123", 30, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, len(result.Actions))
}

// TestGetActionListAtResourceServer_InvalidPagination removed - redundant with TestValidatePaginationParams

func (suite *ResourceServiceTestSuite) TestGetActionListAtResourceServer_ResourceServerNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.GetActionListAtResourceServer("rs-123", 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionListAtResourceServer_CountError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetActionListCountAtResourceServer", "rs-123").Return(0, errors.New("database error"))

	result, err := suite.service.GetActionListAtResourceServer("rs-123", 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionListAtResourceServer_ListError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetActionListCountAtResourceServer", "rs-123").Return(2, nil)
	suite.mockStore.On("GetActionListAtResourceServer", "rs-123", 30, 0).Return(nil, errors.New("database error"))

	result, err := suite.service.GetActionListAtResourceServer("rs-123", 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceServerList_CountError() {
	suite.mockStore.On("GetResourceServerListCount").Return(0, errors.New("database error"))

	result, err := suite.service.GetResourceServerList(30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetResourceServerList_ListError() {
	suite.mockStore.On("GetResourceServerListCount").Return(2, nil)
	suite.mockStore.On("GetResourceServerList", 30, 0).Return(nil, errors.New("database error"))

	result, err := suite.service.GetResourceServerList(30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResourceServer_ResourceServerNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.GetActionAtResourceServer("rs-123", "action-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResourceServer_StoreError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", (*string)(nil)).
		Return(Action{}, errors.New("database error"))

	result, err := suite.service.GetActionAtResourceServer("rs-123", "action-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResourceServer_Success() {
	currentAction := Action{
		ID:          "action-123",
		Name:        testOriginalName,
		Handle:      testOriginalHandle,
		Description: "old description",
	}

	updateReq := Action{
		Name:        testUpdatedName,
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", (*string)(nil)).
		Return(currentAction, nil)
	suite.mockStore.On("UpdateAction", "action-123", "rs-123", (*string)(nil), mock.MatchedBy(func(a Action) bool {
		return a.Name == testUpdatedName && a.Handle == testOriginalHandle && a.Description == testNewDescription
	})).Return(nil)

	result, err := suite.service.UpdateActionAtResourceServer("rs-123", "action-123", updateReq)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(testUpdatedName, result.Name)
	suite.Equal(testOriginalHandle, result.Handle) // Handle is immutable
	suite.Equal(testNewDescription, result.Description)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResourceServer_MissingID() {
	updateReq := Action{
		Description: testNewDescription,
	}

	result, err := suite.service.UpdateActionAtResourceServer("", "action-123", updateReq)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	result, err = suite.service.UpdateActionAtResourceServer("rs-123", "", updateReq)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResourceServer_NotFound() {
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", (*string)(nil)).
		Return(Action{}, ErrActionNotFound)

	result, err := suite.service.UpdateActionAtResourceServer("rs-123", "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorActionNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResourceServer_StoreError() {
	currentAction := Action{
		ID:          "action-123",
		Name:        testOriginalName,
		Handle:      testOriginalHandle,
		Description: "old description",
	}

	updateReq := Action{
		Name:        testUpdatedName,
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", (*string)(nil)).
		Return(currentAction, nil)
	suite.mockStore.On("UpdateAction", "action-123", "rs-123", (*string)(nil), mock.Anything).
		Return(errors.New("database error"))

	result, err := suite.service.UpdateActionAtResourceServer("rs-123", "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResourceServer_CheckServerError() {
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.UpdateActionAtResourceServer("rs-123", "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResourceServer_GetActionError() {
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", (*string)(nil)).
		Return(Action{}, errors.New("database error"))

	result, err := suite.service.UpdateActionAtResourceServer("rs-123", "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResourceServer_ResourceServerNotFound() {
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.UpdateActionAtResourceServer("rs-123", "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

// UpdateActionAtResource Tests

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResource_Success() {
	resourceID := testResourceID
	currentAction := Action{
		ID:          "action-123",
		Name:        testOriginalName,
		Handle:      testOriginalHandle,
		Description: "old description",
		ResourceID:  &resourceID,
	}

	updateReq := Action{
		Name:        testUpdatedName,
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", &resourceID).
		Return(currentAction, nil)
	suite.mockStore.On("UpdateAction", "action-123", "rs-123", &resourceID, mock.MatchedBy(func(a Action) bool {
		return a.Name == testUpdatedName && a.Handle == testOriginalHandle && a.Description == testNewDescription
	})).Return(nil)

	result, err := suite.service.UpdateActionAtResource("rs-123", testResourceID, "action-123", updateReq)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(testUpdatedName, result.Name)
	suite.Equal(testOriginalHandle, result.Handle) // Handle is immutable
	suite.Equal(testNewDescription, result.Description)
	suite.NotNil(result.ResourceID)
	suite.Equal(testResourceID, *result.ResourceID)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResource_MissingID() {
	updateReq := Action{
		Description: testNewDescription,
	}

	result, err := suite.service.UpdateActionAtResource("", testResourceID, "action-123", updateReq)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	result, err = suite.service.UpdateActionAtResource("rs-123", "", "action-123", updateReq)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	result, err = suite.service.UpdateActionAtResource("rs-123", testResourceID, "", updateReq)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResource_ResourceServerNotFound() {
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.UpdateActionAtResource("rs-123", testResourceID, "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResource_ResourceNotFound() {
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(false, nil)

	result, err := suite.service.UpdateActionAtResource("rs-123", testResourceID, "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResource_ActionNotFound() {
	resourceID := testResourceID
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", &resourceID).
		Return(Action{}, ErrActionNotFound)

	result, err := suite.service.UpdateActionAtResource("rs-123", testResourceID, "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorActionNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResource_StoreError() {
	resourceID := testResourceID
	currentAction := Action{
		ID:          "action-123",
		Name:        testOriginalName,
		Handle:      testOriginalHandle,
		Description: "old description",
		ResourceID:  &resourceID,
	}

	updateReq := Action{
		Name:        testUpdatedName,
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", &resourceID).
		Return(currentAction, nil)
	suite.mockStore.On("UpdateAction", "action-123", "rs-123", &resourceID, mock.Anything).
		Return(errors.New("database error"))

	result, err := suite.service.UpdateActionAtResource("rs-123", testResourceID, "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResource_CheckServerError() {
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.UpdateActionAtResource("rs-123", testResourceID, "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResource_CheckResourceError() {
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.UpdateActionAtResource("rs-123", testResourceID, "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestUpdateActionAtResource_GetActionError() {
	resourceID := testResourceID
	updateReq := Action{
		Description: testNewDescription,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", &resourceID).
		Return(Action{}, errors.New("database error"))

	result, err := suite.service.UpdateActionAtResource("rs-123", testResourceID, "action-123", updateReq)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResourceServer_Success() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsActionExist", "action-123", "rs-123", (*string)(nil)).Return(true, nil)
	suite.mockStore.On("DeleteAction", "action-123", "rs-123", (*string)(nil)).Return(nil)

	err := suite.service.DeleteActionAtResourceServer("rs-123", "action-123")

	suite.Nil(err)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResourceServer_MissingID() {
	err := suite.service.DeleteActionAtResourceServer("", "action-123")
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	err = suite.service.DeleteActionAtResourceServer("rs-123", "")
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResourceServer_ResourceServerNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	err := suite.service.DeleteActionAtResourceServer("rs-123", "action-123")

	suite.Nil(err) // Idempotent delete
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResourceServer_StoreError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsActionExist", "action-123", "rs-123", (*string)(nil)).Return(true, nil)
	suite.mockStore.On("DeleteAction", "action-123", "rs-123", (*string)(nil)).Return(errors.New("database error"))

	err := suite.service.DeleteActionAtResourceServer("rs-123", "action-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResourceServer_CheckServerError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, errors.New("database error"))

	err := suite.service.DeleteActionAtResourceServer("rs-123", "action-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResourceServer_ActionNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsActionExist", "action-123", "rs-123", (*string)(nil)).Return(false, nil)

	err := suite.service.DeleteActionAtResourceServer("rs-123", "action-123")

	suite.Nil(err) // Idempotent delete
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResourceServer_CheckActionExistError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsActionExist", "action-123", "rs-123", (*string)(nil)).Return(false, errors.New("database error"))

	err := suite.service.DeleteActionAtResourceServer("rs-123", "action-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResource_Success() {
	resourceID := testResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("IsActionExist", "action-123", "rs-123", &resourceID).Return(true, nil)
	suite.mockStore.On("DeleteAction", "action-123", "rs-123", &resourceID).Return(nil)

	err := suite.service.DeleteActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(err)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResource_MissingID() {
	err := suite.service.DeleteActionAtResource("", testResourceID, "action-123")
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	err = suite.service.DeleteActionAtResource("rs-123", "", "action-123")
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	err = suite.service.DeleteActionAtResource("rs-123", testResourceID, "")
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResource_ResourceServerNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	err := suite.service.DeleteActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(err) // Idempotent delete
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResource_ResourceNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(false, nil)

	err := suite.service.DeleteActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(err) // Idempotent delete
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResource_StoreError() {
	resourceID := testResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("IsActionExist", "action-123", "rs-123", &resourceID).Return(true, nil)
	suite.mockStore.On("DeleteAction", "action-123", "rs-123", &resourceID).Return(errors.New("database error"))

	err := suite.service.DeleteActionAtResource("rs-123", testResourceID, "action-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResource_CheckServerError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, errors.New("database error"))

	err := suite.service.DeleteActionAtResource("rs-123", testResourceID, "action-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResource_CheckResourceError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(false, errors.New("database error"))

	err := suite.service.DeleteActionAtResource("rs-123", testResourceID, "action-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResource_ActionNotFound() {
	resourceID := testResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("IsActionExist", "action-123", "rs-123", &resourceID).Return(false, nil)

	err := suite.service.DeleteActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(err) // Idempotent delete
}

func (suite *ResourceServiceTestSuite) TestDeleteActionAtResource_CheckActionExistError() {
	resourceID := testResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("IsActionExist", "action-123", "rs-123", &resourceID).Return(false, errors.New("database error"))

	err := suite.service.DeleteActionAtResource("rs-123", testResourceID, "action-123")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// GetActionAtResource Tests

func (suite *ResourceServiceTestSuite) TestGetActionAtResource_Success() {
	resourceID := testResourceID
	expectedAction := Action{
		ID:         "action-123",
		Name:       "test-action",
		ResourceID: &resourceID,
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", &resourceID).
		Return(expectedAction, nil)

	result, err := suite.service.GetActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("action-123", result.ID)
	suite.NotNil(result.ResourceID)
	suite.Equal(testResourceID, *result.ResourceID)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResource_MissingID() {
	result, err := suite.service.GetActionAtResource("", testResourceID, "action-123")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	result, err = suite.service.GetActionAtResource("rs-123", "", "action-123")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)

	result, err = suite.service.GetActionAtResource("rs-123", testResourceID, "")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingID.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResource_ResourceServerNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.GetActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResource_ResourceNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(false, nil)

	result, err := suite.service.GetActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResource_ActionNotFound() {
	resourceID := testResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", &resourceID).
		Return(Action{}, ErrActionNotFound)

	result, err := suite.service.GetActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorActionNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResource_StoreError() {
	resourceID := testResourceID
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetAction", "action-123", "rs-123", &resourceID).
		Return(Action{}, errors.New("database error"))

	result, err := suite.service.GetActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResource_CheckServerError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.GetActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionAtResource_CheckResourceError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.GetActionAtResource("rs-123", testResourceID, "action-123")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// GetActionListAtResource Tests

func (suite *ResourceServiceTestSuite) TestGetActionListAtResource_Success() {
	actions := []Action{
		{ID: "action-1", Name: "Action 1"},
		{ID: "action-2", Name: "Action 2"},
	}

	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetActionListCountAtResource", "rs-123", testResourceID).Return(2, nil)
	suite.mockStore.On("GetActionListAtResource", "rs-123", testResourceID, 30, 0).Return(actions, nil)

	result, err := suite.service.GetActionListAtResource("rs-123", testResourceID, 30, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, len(result.Actions))
}

// TestGetActionListAtResource_InvalidPagination removed - redundant with TestValidatePaginationParams

func (suite *ResourceServiceTestSuite) TestGetActionListAtResource_ResourceServerNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, nil)

	result, err := suite.service.GetActionListAtResource("rs-123", testResourceID, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceServerNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionListAtResource_ResourceNotFound() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(false, nil)

	result, err := suite.service.GetActionListAtResource("rs-123", testResourceID, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorResourceNotFound.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionListAtResource_EmptyList() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetActionListCountAtResource", "rs-123", testResourceID).Return(0, nil)
	suite.mockStore.On("GetActionListAtResource", "rs-123", testResourceID, 30, 0).Return([]Action{}, nil)

	result, err := suite.service.GetActionListAtResource("rs-123", testResourceID, 30, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(0, result.TotalResults)
	suite.Equal(0, len(result.Actions))
}

func (suite *ResourceServiceTestSuite) TestGetActionListAtResource_StoreError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(true, nil)
	suite.mockStore.On("GetActionListCountAtResource", "rs-123", testResourceID).
		Return(0, errors.New("database error"))

	result, err := suite.service.GetActionListAtResource("rs-123", testResourceID, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionListAtResource_CheckResourceServerError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.GetActionListAtResource("rs-123", testResourceID, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestGetActionListAtResource_CheckResourceError() {
	suite.mockStore.On("IsResourceServerExist", "rs-123").Return(true, nil)
	suite.mockStore.On("IsResourceExist", testResourceID, "rs-123").Return(false, errors.New("database error"))

	result, err := suite.service.GetActionListAtResource("rs-123", testResourceID, 30, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// Validation Helper Tests

func (suite *ResourceServiceTestSuite) TestValidatePaginationParams() {
	// Valid params
	err := validatePaginationParams(30, 0)
	suite.Nil(err)

	// Invalid limit - too small
	err = validatePaginationParams(0, 0)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidLimit.Code, err.Code)

	// Invalid limit - too large (assuming MaxPageSize is defined)
	err = validatePaginationParams(10000, 0)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidLimit.Code, err.Code)

	// Invalid offset
	err = validatePaginationParams(30, -1)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidOffset.Code, err.Code)
}

func (suite *ResourceServiceTestSuite) TestBuildPaginationLinks() {
	// Test with offset 0 (first page)
	links := buildPaginationLinks("/test", 10, 0, 25)
	suite.NotNil(links)
	// Should have next and last links
	suite.True(len(links) >= 1)

	// Test with middle page
	links = buildPaginationLinks("/test", 10, 10, 25)
	suite.NotNil(links)
	// Should have first, prev, next, and potentially last links

	// Test with last page
	links = buildPaginationLinks("/test", 10, 20, 25)
	suite.NotNil(links)
	// Should have first and prev links, no next
}
