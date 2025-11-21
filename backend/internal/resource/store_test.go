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

//nolint:lll // Test file with long mock setup calls
package resource

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/tests/mocks/database/clientmock"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

const (
	testParentID1   = "parent1"
	testResourceID1 = "res1"
)

// ResourceStoreTestSuite is the test suite for resourceStore.
type ResourceStoreTestSuite struct {
	suite.Suite
	mockDBProvider *providermock.DBProviderInterfaceMock
	mockDBClient   *clientmock.DBClientInterfaceMock
	store          *resourceStore
}

// TestResourceStoreTestSuite runs the test suite.
func TestResourceStoreTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceStoreTestSuite))
}

// SetupTest sets up the test suite.
func (suite *ResourceStoreTestSuite) SetupTest() {
	suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.mockDBClient = clientmock.NewDBClientInterfaceMock(suite.T())
	suite.store = &resourceStore{
		dbProvider: suite.mockDBProvider,
	}
}

// Resource Server Tests

func (suite *ResourceStoreTestSuite) TestCreateResourceServer_Success() {
	rs := ResourceServer{
		OrganizationUnitID: "ou1",
		Name:               "Test Server",
		Description:        "Test Description",
		Identifier:         "test-identifier",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResourceServer, "rs1", "ou1", "Test Server", "Test Description", "test-identifier", "{}").
		Return(int64(1), nil)

	err := suite.store.CreateResourceServer("rs1", rs)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestCreateResourceServer_ExecuteError() {
	rs := ResourceServer{
		OrganizationUnitID: "ou1",
		Name:               "Test Server",
		Description:        "Test Description",
		Identifier:         "test-identifier",
	}

	execError := errors.New("insert failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResourceServer, "rs1", "ou1", "Test Server", "Test Description", "test-identifier", "{}").
		Return(int64(0), execError)

	err := suite.store.CreateResourceServer("rs1", rs)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to create resource server")
}

func (suite *ResourceStoreTestSuite) TestGetResourceServer_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerByID, "rs1").Return([]map[string]interface{}{
		{
			"resource_server_id": "rs1",
			"ou_id":              "ou1",
			"name":               "Test Server",
			"description":        "Test Description",
			"identifier":         "test-identifier",
		},
	}, nil)

	rs, err := suite.store.GetResourceServer("rs1")

	suite.NoError(err)
	suite.Equal("rs1", rs.ID)
	suite.Equal("ou1", rs.OrganizationUnitID)
	suite.Equal("Test Server", rs.Name)
	suite.Equal("Test Description", rs.Description)
	suite.Equal("test-identifier", rs.Identifier)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServer_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerByID, "nonexistent").Return([]map[string]interface{}{}, nil)

	rs, err := suite.store.GetResourceServer("nonexistent")

	suite.Error(err)
	suite.Equal(ErrResourceServerNotFound, err)
	suite.Empty(rs.ID)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServer_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerByID, "rs1").Return(nil, queryError)

	rs, err := suite.store.GetResourceServer("rs1")

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get resource server")
	suite.Empty(rs.ID)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServerList_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerList, 10, 0).Return([]map[string]interface{}{
		{
			"resource_server_id": "rs1",
			"ou_id":              "ou1",
			"name":               "Server 1",
			"description":        "Description 1",
			"identifier":         "identifier-1",
		},
		{
			"resource_server_id": "rs2",
			"ou_id":              "ou1",
			"name":               "Server 2",
			"description":        "Description 2",
			"identifier":         "identifier-2",
		},
	}, nil)

	servers, err := suite.store.GetResourceServerList(10, 0)

	suite.NoError(err)
	suite.Len(servers, 2)
	suite.Equal("rs1", servers[0].ID)
	suite.Equal("Server 1", servers[0].Name)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServerList_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerList, 10, 0).Return(nil, queryError)

	servers, err := suite.store.GetResourceServerList(10, 0)

	suite.Error(err)
	suite.Nil(servers)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServerList_InvalidRowData() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerList, 10, 0).Return([]map[string]interface{}{
		{
			"resource_server_id": 123, // Invalid type
			"ou_id":              "ou1",
			"name":               "Server 1",
		},
	}, nil)

	servers, err := suite.store.GetResourceServerList(10, 0)

	suite.Error(err)
	suite.Nil(servers)
	suite.Contains(err.Error(), "failed to build resource server")
}

func (suite *ResourceStoreTestSuite) TestGetResourceServerListCount_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerListCount).Return([]map[string]interface{}{
		{"total": int64(5)},
	}, nil)

	count, err := suite.store.GetResourceServerListCount()

	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServerListCount_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerListCount).Return(nil, queryError)

	count, err := suite.store.GetResourceServerListCount()

	suite.Error(err)
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestUpdateResourceServer_Success() {
	rs := ResourceServer{
		OrganizationUnitID: "ou1",
		Name:               "Updated Server",
		Description:        "Updated Description",
		Identifier:         "updated-identifier",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateResourceServer, "ou1", "Updated Server", "Updated Description", "updated-identifier", "{}", "rs1").
		Return(int64(1), nil)

	err := suite.store.UpdateResourceServer("rs1", rs)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestUpdateResourceServer_ExecuteError() {
	rs := ResourceServer{
		OrganizationUnitID: "ou1",
		Name:               "Updated Server",
		Description:        "Updated Description",
		Identifier:         "updated-identifier",
	}

	execError := errors.New("update failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateResourceServer, "ou1", "Updated Server", "Updated Description", "updated-identifier", "{}", "rs1").
		Return(int64(0), execError)

	err := suite.store.UpdateResourceServer("rs1", rs)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update resource server")
}

func (suite *ResourceStoreTestSuite) TestDeleteResourceServer_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteResourceServer, "rs1").Return(int64(1), nil)

	err := suite.store.DeleteResourceServer("rs1")

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestDeleteResourceServer_ExecuteError() {
	execError := errors.New("delete failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteResourceServer, "rs1").Return(int64(0), execError)

	err := suite.store.DeleteResourceServer("rs1")

	suite.Error(err)
	suite.Contains(err.Error(), "failed to delete resource server")
}

func (suite *ResourceStoreTestSuite) TestIsResourceServerExist_Exists() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerExists, "rs1").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.IsResourceServerExist("rs1")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestIsResourceServerExist_DoesNotExist() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerExists, "nonexistent").Return([]map[string]interface{}{
		{"count": int64(0)},
	}, nil)

	exists, err := suite.store.IsResourceServerExist("nonexistent")

	suite.NoError(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestIsResourceServerExist_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerExists, "rs1").Return(nil, queryError)

	exists, err := suite.store.IsResourceServerExist("rs1")

	suite.Error(err)
	suite.False(exists)
	suite.Contains(err.Error(), "failed to check resource server existence")
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerNameExists_Exists() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerNameExists, "Test Server").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.CheckResourceServerNameExists("Test Server")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerNameExists_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerNameExists, "Test Server").Return(nil, queryError)

	exists, err := suite.store.CheckResourceServerNameExists("Test Server")

	suite.Error(err)
	suite.False(exists)
	suite.Contains(err.Error(), "failed to check resource server name")
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerNameExistsExcludingID_DoesNotExist() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerNameExistsExcludingID, "Test Server", "rs1").
		Return([]map[string]interface{}{
			{"count": int64(0)},
		}, nil)

	exists, err := suite.store.CheckResourceServerNameExistsExcludingID("Test Server", "rs1")

	suite.NoError(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerNameExistsExcludingID_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerNameExistsExcludingID, "Test Server", "rs1").
		Return(nil, queryError)

	exists, err := suite.store.CheckResourceServerNameExistsExcludingID("Test Server", "rs1")

	suite.Error(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerIdentifierExists_Exists() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerIdentifierExists, "test-identifier").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.CheckResourceServerIdentifierExists("test-identifier")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerIdentifierExists_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerIdentifierExists, "test-identifier").Return(nil, queryError)

	exists, err := suite.store.CheckResourceServerIdentifierExists("test-identifier")

	suite.Error(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerIdentifierExistsExcludingID_DoesNotExist() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerIdentifierExistsExcludingID, "test-identifier", "rs1").
		Return([]map[string]interface{}{
			{"count": int64(0)},
		}, nil)

	exists, err := suite.store.CheckResourceServerIdentifierExistsExcludingID("test-identifier", "rs1")

	suite.NoError(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerIdentifierExistsExcludingID_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerIdentifierExistsExcludingID, "test-identifier", "rs1").
		Return(nil, queryError)

	exists, err := suite.store.CheckResourceServerIdentifierExistsExcludingID("test-identifier", "rs1")

	suite.Error(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerHasDependencies_HasDependencies() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerHasDependencies, "rs1").Return([]map[string]interface{}{
		{"count": int64(3)},
	}, nil)

	hasDeps, err := suite.store.CheckResourceServerHasDependencies("rs1")

	suite.NoError(err)
	suite.True(hasDeps)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerHasDependencies_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerHasDependencies, "rs1").Return(nil, queryError)

	hasDeps, err := suite.store.CheckResourceServerHasDependencies("rs1")

	suite.Error(err)
	suite.False(hasDeps)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerExistAndGetInternalID_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerInternalID, "rs1").Return([]map[string]interface{}{
		{"id": int64(42)},
	}, nil)

	id, err := suite.store.CheckResourceServerExistAndGetInternalID("rs1")

	suite.NoError(err)
	suite.Equal(42, id)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerExistAndGetInternalID_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerInternalID, "nonexistent").Return([]map[string]interface{}{}, nil)

	id, err := suite.store.CheckResourceServerExistAndGetInternalID("nonexistent")

	suite.Error(err)
	suite.Equal(ErrResourceServerNotFound, err)
	suite.Equal(0, id)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerExistAndGetInternalID_IntType() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerInternalID, "rs1").Return([]map[string]interface{}{
		{"id": int(42)},
	}, nil)

	id, err := suite.store.CheckResourceServerExistAndGetInternalID("rs1")

	suite.NoError(err)
	suite.Equal(42, id)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerExistAndGetInternalID_Float64Type() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerInternalID, "rs1").Return([]map[string]interface{}{
		{"id": float64(42)},
	}, nil)

	id, err := suite.store.CheckResourceServerExistAndGetInternalID("rs1")

	suite.NoError(err)
	suite.Equal(42, id)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerExistAndGetInternalID_InvalidType() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerInternalID, "rs1").Return([]map[string]interface{}{
		{"id": "invalid"},
	}, nil)

	id, err := suite.store.CheckResourceServerExistAndGetInternalID("rs1")

	suite.Error(err)
	suite.Contains(err.Error(), "unexpected internal ID type")
	suite.Equal(0, id)
}

// Resource Tests

func (suite *ResourceStoreTestSuite) TestCreateResource_Success() {
	res := Resource{
		Name:        "Test Resource",
		Handle:      "test-handle",
		Description: "Test Description",
	}
	parentID := 10

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResource, "res1", 5, "Test Resource", "test-handle", "Test Description", "{}", &parentID).
		Return(int64(1), nil)

	err := suite.store.CreateResource("res1", 5, &parentID, res)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestCreateResource_NullParent() {
	res := Resource{
		Name:        "Test Resource",
		Handle:      "test-handle",
		Description: "Test Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResource, "res1", 5, "Test Resource", "test-handle", "Test Description", "{}", (*int)(nil)).
		Return(int64(1), nil)

	err := suite.store.CreateResource("res1", 5, nil, res)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestCreateResource_ExecuteError() {
	res := Resource{
		Name:        "Test Resource",
		Handle:      "test-handle",
		Description: "Test Description",
	}

	execError := errors.New("insert failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResource, "res1", 5, "Test Resource", "test-handle", "Test Description", "{}", (*int)(nil)).
		Return(int64(0), execError)

	err := suite.store.CreateResource("res1", 5, nil, res)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to create resource")
}

func (suite *ResourceStoreTestSuite) TestGetResource_Success() {
	parentID := testParentID1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceByID, testResourceID1, "rs1").Return([]map[string]interface{}{
		{
			"resource_id":        "res1",
			"resource_server_id": "rs1",
			"name":               "Test Resource",
			"handle":             "test-handle",
			"description":        "Test Description",
			"parent_resource_id": parentID,
		},
	}, nil)

	res, err := suite.store.GetResource("res1", "rs1")

	suite.NoError(err)
	suite.Equal("res1", res.ID)
	suite.Equal("Test Resource", res.Name)
	suite.Equal("test-handle", res.Handle)
	suite.Equal("Test Description", res.Description)
	suite.NotNil(res.Parent)
	suite.Equal(parentID, *res.Parent)
}

func (suite *ResourceStoreTestSuite) TestGetResource_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceByID, "nonexistent", "rs1").Return([]map[string]interface{}{}, nil)

	res, err := suite.store.GetResource("nonexistent", "rs1")

	suite.Error(err)
	suite.Equal(ErrResourceNotFound, err)
	suite.Empty(res.ID)
}

func (suite *ResourceStoreTestSuite) TestGetResourceList_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceList, "rs1", 10, 0).Return([]map[string]interface{}{
		{
			"resource_id":        "res1",
			"resource_server_id": "rs1",
			"name":               "Resource 1",
			"handle":             "resource-1",
			"description":        "Description 1",
		},
		{
			"resource_id":        "res2",
			"resource_server_id": "rs1",
			"name":               "Resource 2",
			"handle":             "resource-2",
			"description":        "Description 2",
		},
	}, nil)

	resources, err := suite.store.GetResourceList("rs1", 10, 0)

	suite.NoError(err)
	suite.Len(resources, 2)
	suite.Equal("res1", resources[0].ID)
	suite.Equal("Resource 1", resources[0].Name)
	suite.Equal("resource-1", resources[0].Handle)
}

func (suite *ResourceStoreTestSuite) TestGetResourceList_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceList, "rs1", 10, 0).Return(nil, queryError)

	resources, err := suite.store.GetResourceList("rs1", 10, 0)

	suite.Error(err)
	suite.Nil(resources)
	suite.Contains(err.Error(), "failed to get resource list")
}

func (suite *ResourceStoreTestSuite) TestGetResourceList_InvalidRowData() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceList, "rs1", 10, 0).Return([]map[string]interface{}{
		{
			"resource_id":        123, // Invalid type
			"resource_server_id": "rs1",
			"name":               "Resource 1",
		},
	}, nil)

	resources, err := suite.store.GetResourceList("rs1", 10, 0)

	suite.Error(err)
	suite.Nil(resources)
	suite.Contains(err.Error(), "failed to build resource")
}

func (suite *ResourceStoreTestSuite) TestGetResourceListByParent_NullParent() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListByNullParent, "rs1", 10, 0).Return([]map[string]interface{}{
		{
			"resource_id":        "res1",
			"resource_server_id": "rs1",
			"name":               "Resource 1",
			"handle":             "resource-1",
			"description":        "Description 1",
		},
	}, nil)

	resources, err := suite.store.GetResourceListByParent("rs1", nil, 10, 0)

	suite.NoError(err)
	suite.Len(resources, 1)
	suite.Equal("res1", resources[0].ID)
	suite.Equal("resource-1", resources[0].Handle)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListByParent_WithParent() {
	parentID := testParentID1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListByParent, "rs1", parentID, 10, 0).Return([]map[string]interface{}{
		{
			"resource_id":        "res1",
			"resource_server_id": "rs1",
			"name":               "Resource 1",
			"handle":             "resource-1",
			"description":        "Description 1",
			"parent_resource_id": parentID,
		},
	}, nil)

	resources, err := suite.store.GetResourceListByParent("rs1", &parentID, 10, 0)

	suite.NoError(err)
	suite.Len(resources, 1)
	suite.Equal("res1", resources[0].ID)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListByParent_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListByNullParent, "rs1", 10, 0).Return(nil, queryError)

	resources, err := suite.store.GetResourceListByParent("rs1", nil, 10, 0)

	suite.Error(err)
	suite.Nil(resources)
	suite.Contains(err.Error(), "failed to get resource list by parent")
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCount_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCount, "rs1").Return([]map[string]interface{}{
		{"total": int64(10)},
	}, nil)

	count, err := suite.store.GetResourceListCount("rs1")

	suite.NoError(err)
	suite.Equal(10, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCount_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCount, "rs1").Return(nil, queryError)

	count, err := suite.store.GetResourceListCount("rs1")

	suite.Error(err)
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCountByParent_NullParent() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCountByNullParent, "rs1").Return([]map[string]interface{}{
		{"total": int64(5)},
	}, nil)

	count, err := suite.store.GetResourceListCountByParent("rs1", nil)

	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCountByParent_WithParent() {
	parentID := testParentID1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCountByParent, "rs1", parentID).Return([]map[string]interface{}{
		{"total": int64(3)},
	}, nil)

	count, err := suite.store.GetResourceListCountByParent("rs1", &parentID)

	suite.NoError(err)
	suite.Equal(3, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCountByParent_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCountByNullParent, "rs1").Return(nil, queryError)

	count, err := suite.store.GetResourceListCountByParent("rs1", nil)

	suite.Error(err)
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestUpdateResource_Success() {
	res := Resource{
		Name:        "Updated Resource",
		Description: "Updated Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateResource, "Updated Resource", "Updated Description", "{}",
		"res1", "rs1").Return(int64(1), nil)

	err := suite.store.UpdateResource("res1", "rs1", res)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestUpdateResource_NullParent() {
	res := Resource{
		Name:        "Updated Resource",
		Description: "Updated Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateResource, "Updated Resource", "Updated Description", "{}",
		"res1", "rs1").Return(int64(1), nil)

	err := suite.store.UpdateResource("res1", "rs1", res)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestUpdateResource_ParentNotFound() {
	res := Resource{
		Name:        "Updated Name",
		Description: "Updated Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateResource, "Updated Name", "Updated Description", "{}",
		"nonexistent", "rs1").Return(int64(0), ErrResourceNotFound)

	err := suite.store.UpdateResource("nonexistent", "rs1", res)

	suite.Error(err)
}

func (suite *ResourceStoreTestSuite) TestUpdateResource_ExecuteError() {
	res := Resource{
		Name:        "Updated Name",
		Description: "Updated Description",
	}

	execError := errors.New("update failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateResource, "Updated Name", "Updated Description", "{}",
		"res1", "rs1").Return(int64(0), execError)

	err := suite.store.UpdateResource("res1", "rs1", res)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update resource")
}

func (suite *ResourceStoreTestSuite) TestDeleteResource_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteResource, "res1", "rs1").Return(int64(1), nil)

	err := suite.store.DeleteResource("res1", "rs1")

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestDeleteResource_ExecuteError() {
	execError := errors.New("delete error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteResource, "res1", "rs1").Return(int64(0), execError)

	err := suite.store.DeleteResource("res1", "rs1")

	suite.Error(err)
	suite.Contains(err.Error(), "failed to delete resource")
}

func (suite *ResourceStoreTestSuite) TestIsResourceExist_Exists() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceExists, "res1", "rs1").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.IsResourceExist("res1", "rs1")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestIsResourceExist_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceExists, "res1", "rs1").Return(nil, queryError)

	exists, err := suite.store.IsResourceExist("res1", "rs1")

	suite.Error(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHandleExistsUnderParent_NullParent() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHandleExistsUnderNullParent, "rs1", "Test Resource").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.CheckResourceHandleExistsUnderParent("rs1", "Test Resource", nil)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHandleExistsUnderParent_WithParent() {
	parentID := testParentID1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHandleExistsUnderParent, "rs1", "Test Resource", parentID).Return([]map[string]interface{}{
		{"count": int64(0)},
	}, nil)

	exists, err := suite.store.CheckResourceHandleExistsUnderParent("rs1", "Test Resource", &parentID)

	suite.NoError(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHandleExistsUnderParent_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHandleExistsUnderNullParent, "rs1", "Test Resource").Return(nil, queryError)

	exists, err := suite.store.CheckResourceHandleExistsUnderParent("rs1", "Test Resource", nil)

	suite.Error(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHasDependencies_HasDependencies() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHasDependencies, "res1").Return([]map[string]interface{}{
		{"count": int64(2)},
	}, nil)

	hasDeps, err := suite.store.CheckResourceHasDependencies("res1")

	suite.NoError(err)
	suite.True(hasDeps)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHasDependencies_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHasDependencies, "res1").Return(nil, queryError)

	hasDeps, err := suite.store.CheckResourceHasDependencies("res1")

	suite.Error(err)
	suite.False(hasDeps)
}

func (suite *ResourceStoreTestSuite) TestCheckCircularDependency_HasCircular() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckCircularDependency, "parent1", "res1").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	hasCircular, err := suite.store.CheckCircularDependency("res1", "parent1")

	suite.NoError(err)
	suite.True(hasCircular)
}

func (suite *ResourceStoreTestSuite) TestCheckCircularDependency_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckCircularDependency, "parent1", "res1").Return(nil, queryError)

	hasCircular, err := suite.store.CheckCircularDependency("res1", "parent1")

	suite.Error(err)
	suite.False(hasCircular)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceExistAndGetInternalID_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceInternalID, "res1", "rs1").Return([]map[string]interface{}{
		{"id": int64(100)},
	}, nil)

	id, err := suite.store.CheckResourceExistAndGetInternalID("res1", "rs1")

	suite.NoError(err)
	suite.Equal(100, id)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceExistAndGetInternalID_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceInternalID, "nonexistent", "rs1").Return([]map[string]interface{}{}, nil)

	id, err := suite.store.CheckResourceExistAndGetInternalID("nonexistent", "rs1")

	suite.Error(err)
	suite.Equal(ErrResourceNotFound, err)
	suite.Equal(0, id)
}

// Action Tests

func (suite *ResourceStoreTestSuite) TestCreateAction_Success() {
	action := Action{
		Name:        "Test Action",
		Handle:      "test-handle",
		Description: "Test Description",
	}
	resourceID := 10

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateAction, "action1", 5, &resourceID, "Test Action", "test-handle", "Test Description", "{}").
		Return(int64(1), nil)

	err := suite.store.CreateAction("action1", 5, &resourceID, action)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestCreateAction_NullResource() {
	action := Action{
		Name:        "Test Action",
		Handle:      "test-handle",
		Description: "Test Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateAction, "action1", 5, (*int)(nil), "Test Action", "test-handle", "Test Description", "{}").
		Return(int64(1), nil)

	err := suite.store.CreateAction("action1", 5, nil, action)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestCreateAction_ExecuteError() {
	action := Action{
		Name:        "Test Action",
		Handle:      "test-handle",
		Description: "Test Description",
	}

	execError := errors.New("insert error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateAction, "action1", 5, (*int)(nil), "Test Action", "test-handle", "Test Description", "{}").
		Return(int64(0), execError)

	err := suite.store.CreateAction("action1", 5, nil, action)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to create action")
}

func (suite *ResourceStoreTestSuite) TestGetAction_AtResourceServer() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionByID, "action1", "rs1").Return([]map[string]interface{}{
		{
			"action_id":          "action1",
			"resource_server_id": "rs1",
			"name":               "Test Action",
			"handle":             "test-handle",
			"description":        "Test Description",
		},
	}, nil)

	action, err := suite.store.GetAction("action1", "rs1", nil)

	suite.NoError(err)
	suite.Equal("action1", action.ID)
	suite.Equal("Test Action", action.Name)
	suite.Equal("test-handle", action.Handle)
	suite.Nil(action.ResourceID)
}

func (suite *ResourceStoreTestSuite) TestGetAction_AtResource() {
	resourceID := testResourceID1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionByIDAtResource, "action1", "rs1", resourceID).Return([]map[string]interface{}{
		{
			"action_id":          "action1",
			"resource_server_id": "rs1",
			"resource_id":        resourceID,
			"name":               "Test Action",
			"handle":             "test-handle",
			"description":        "Test Description",
		},
	}, nil)

	action, err := suite.store.GetAction("action1", "rs1", &resourceID)

	suite.NoError(err)
	suite.Equal("action1", action.ID)
	suite.Equal("test-handle", action.Handle)
	suite.NotNil(action.ResourceID)
	suite.Equal(resourceID, *action.ResourceID)
}

func (suite *ResourceStoreTestSuite) TestGetAction_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionByID, "nonexistent", "rs1").Return([]map[string]interface{}{}, nil)

	action, err := suite.store.GetAction("nonexistent", "rs1", nil)

	suite.Error(err)
	suite.Equal(ErrActionNotFound, err)
	suite.Empty(action.ID)
}

func (suite *ResourceStoreTestSuite) TestGetAction_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionByID, "action1", "rs1").Return(nil, queryError)

	action, err := suite.store.GetAction("action1", "rs1", nil)

	suite.Error(err)
	suite.Empty(action.ID)
	suite.Contains(err.Error(), "failed to get action")
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResourceServer_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListAtResourceServer, "rs1", 10, 0).Return([]map[string]interface{}{
		{
			"action_id":          "action1",
			"resource_server_id": "rs1",
			"name":               "Action 1",
			"handle":             "action-1",
			"description":        "Description 1",
		},
		{
			"action_id":          "action2",
			"resource_server_id": "rs1",
			"name":               "Action 2",
			"handle":             "action-2",
			"description":        "Description 2",
		},
	}, nil)

	actions, err := suite.store.GetActionListAtResourceServer("rs1", 10, 0)

	suite.NoError(err)
	suite.Len(actions, 2)
	suite.Equal("action1", actions[0].ID)
	suite.Equal("Action 1", actions[0].Name)
	suite.Equal("action-1", actions[0].Handle)
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResourceServer_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListAtResourceServer, "rs1", 10, 0).Return(nil, queryError)

	actions, err := suite.store.GetActionListAtResourceServer("rs1", 10, 0)

	suite.Error(err)
	suite.Nil(actions)
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResourceServer_InvalidRowData() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListAtResourceServer, "rs1", 10, 0).Return([]map[string]interface{}{
		{
			"action_id": 123, // Invalid type
			"name":      "Action 1",
		},
	}, nil)

	actions, err := suite.store.GetActionListAtResourceServer("rs1", 10, 0)

	suite.Error(err)
	suite.Nil(actions)
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResource_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListAtResource, "rs1", "res1", 10, 0).Return([]map[string]interface{}{
		{
			"action_id":          "action1",
			"resource_server_id": "rs1",
			"resource_id":        "res1",
			"name":               "Action 1",
			"handle":             "action-1",
			"description":        "Description 1",
		},
	}, nil)

	actions, err := suite.store.GetActionListAtResource("rs1", "res1", 10, 0)

	suite.NoError(err)
	suite.Len(actions, 1)
	suite.Equal("action1", actions[0].ID)
	suite.Equal("action-1", actions[0].Handle)
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResource_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListAtResource, "rs1", "res1", 10, 0).Return(nil, queryError)

	actions, err := suite.store.GetActionListAtResource("rs1", "res1", 10, 0)

	suite.Error(err)
	suite.Nil(actions)
}

func (suite *ResourceStoreTestSuite) TestGetActionListCountAtResourceServer_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListCountAtResourceServer, "rs1").Return([]map[string]interface{}{
		{"total": int64(15)},
	}, nil)

	count, err := suite.store.GetActionListCountAtResourceServer("rs1")

	suite.NoError(err)
	suite.Equal(15, count)
}

func (suite *ResourceStoreTestSuite) TestGetActionListCountAtResourceServer_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListCountAtResourceServer, "rs1").Return(nil, queryError)

	count, err := suite.store.GetActionListCountAtResourceServer("rs1")

	suite.Error(err)
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestGetActionListCountAtResource_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListCountAtResource, "rs1", "res1").Return([]map[string]interface{}{
		{"total": int64(5)},
	}, nil)

	count, err := suite.store.GetActionListCountAtResource("rs1", "res1")

	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *ResourceStoreTestSuite) TestGetActionListCountAtResource_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListCountAtResource, "rs1", "res1").Return(nil, queryError)

	count, err := suite.store.GetActionListCountAtResource("rs1", "res1")

	suite.Error(err)
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestUpdateAction_Success() {
	action := Action{
		Name:        "Updated Action",
		Description: "Updated Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateAction, "Updated Action", "Updated Description", "{}", "action1", "rs1").
		Return(int64(1), nil)

	err := suite.store.UpdateAction("action1", "rs1", nil, action)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestUpdateAction_WithResourceID() {
	resourceID := testResourceID1
	action := Action{
		Name:        "Updated Action",
		Description: "Updated Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateActionAtResource, "Updated Action", "Updated Description", "{}", "action1", "rs1", resourceID).
		Return(int64(1), nil)

	err := suite.store.UpdateAction("action1", "rs1", &resourceID, action)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestUpdateAction_ExecuteError() {
	action := Action{
		Name:        "Updated Action",
		Description: "Updated Description",
	}

	execError := errors.New("update error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateAction, "Updated Action", "Updated Description", "{}", "action1", "rs1").
		Return(int64(0), execError)

	err := suite.store.UpdateAction("action1", "rs1", nil, action)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update action")
}

func (suite *ResourceStoreTestSuite) TestDeleteAction_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteAction, "action1", "rs1").Return(int64(1), nil)

	err := suite.store.DeleteAction("action1", "rs1", nil)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestDeleteAction_WithResourceID() {
	resourceID := testResourceID1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteActionAtResource, "action1", "rs1", resourceID).Return(int64(1), nil)

	err := suite.store.DeleteAction("action1", "rs1", &resourceID)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestDeleteAction_ExecuteError() {
	execError := errors.New("delete error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteAction, "action1", "rs1").Return(int64(0), execError)

	err := suite.store.DeleteAction("action1", "rs1", nil)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to delete action")
}

func (suite *ResourceStoreTestSuite) TestIsActionExist_AtResourceServer() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionExists, "action1", "rs1").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.IsActionExist("action1", "rs1", nil)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestIsActionExist_AtResource() {
	resourceID := testResourceID1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionExistsAtResource, "action1", "rs1", resourceID).Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.IsActionExist("action1", "rs1", &resourceID)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckActionHandleExists_AtResourceServer() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionHandleExistsAtResourceServer, "rs1", "Test Action").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.CheckActionHandleExists("rs1", nil, "Test Action")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckActionHandleExists_AtResource() {
	resourceID := testResourceID1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionHandleExistsAtResource, "rs1", resourceID, "Test Action").Return([]map[string]interface{}{
		{"count": int64(0)},
	}, nil)

	exists, err := suite.store.CheckActionHandleExists("rs1", &resourceID, "Test Action")

	suite.NoError(err)
	suite.False(exists)
}

// Helper Function Tests

func (suite *ResourceStoreTestSuite) TestGetIdentityDBClient_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)

	client, err := suite.store.getIdentityDBClient()

	suite.NoError(err)
	suite.NotNil(client)
	suite.Equal(suite.mockDBClient, client)
}

func (suite *ResourceStoreTestSuite) TestGetIdentityDBClient_Error() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	client, err := suite.store.getIdentityDBClient()

	suite.Error(err)
	suite.Nil(client)
	suite.Contains(err.Error(), "failed to get identity DB client")
}

func (suite *ResourceStoreTestSuite) TestParseCountResult_Success() {
	results := []map[string]interface{}{
		{"total": int64(42)},
	}

	count, err := parseCountResult(results)

	suite.NoError(err)
	suite.Equal(42, count)
}

func (suite *ResourceStoreTestSuite) TestParseCountResult_CountField() {
	results := []map[string]interface{}{
		{"count": int64(42)},
	}

	count, err := parseCountResult(results)

	suite.NoError(err)
	suite.Equal(42, count)
}

func (suite *ResourceStoreTestSuite) TestParseCountResult_IntType() {
	results := []map[string]interface{}{
		{"total": int(42)},
	}

	count, err := parseCountResult(results)

	suite.NoError(err)
	suite.Equal(42, count)
}

func (suite *ResourceStoreTestSuite) TestParseCountResult_Float64Type() {
	results := []map[string]interface{}{
		{"total": float64(42)},
	}

	count, err := parseCountResult(results)

	suite.NoError(err)
	suite.Equal(42, count)
}

func (suite *ResourceStoreTestSuite) TestParseCountResult_EmptyResults() {
	results := []map[string]interface{}{}

	count, err := parseCountResult(results)

	suite.Error(err)
	suite.Contains(err.Error(), "no count result returned")
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestParseCountResult_MissingField() {
	results := []map[string]interface{}{
		{"other": int64(42)},
	}

	count, err := parseCountResult(results)

	suite.Error(err)
	suite.Contains(err.Error(), "count field not found")
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestParseCountResult_InvalidType() {
	results := []map[string]interface{}{
		{"total": "not_a_number"},
	}

	count, err := parseCountResult(results)

	suite.Error(err)
	suite.Contains(err.Error(), "unexpected count type")
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestParseBoolFromCount_True() {
	results := []map[string]interface{}{
		{"count": int64(5)},
	}

	exists, err := parseBoolFromCount(results)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestParseBoolFromCount_False() {
	results := []map[string]interface{}{
		{"count": int64(0)},
	}

	exists, err := parseBoolFromCount(results)

	suite.NoError(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestParseBoolFromCount_Error() {
	results := []map[string]interface{}{}

	exists, err := parseBoolFromCount(results)

	suite.Error(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestBuildResourceServerFromResultRow_Success() {
	row := map[string]interface{}{
		"resource_server_id": "rs1",
		"ou_id":              "ou1",
		"name":               "Test Server",
		"description":        "Test Description",
		"identifier":         "test-identifier",
	}

	rs, err := buildResourceServerFromResultRow(row)

	suite.NoError(err)
	suite.Equal("rs1", rs.ID)
	suite.Equal("ou1", rs.OrganizationUnitID)
	suite.Equal("Test Server", rs.Name)
	suite.Equal("Test Description", rs.Description)
	suite.Equal("test-identifier", rs.Identifier)
}

func (suite *ResourceStoreTestSuite) TestBuildResourceServerFromResultRow_MissingID() {
	row := map[string]interface{}{
		"ou_id": "ou1",
		"name":  "Test Server",
	}

	rs, err := buildResourceServerFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "resource_server_id")
	suite.Empty(rs.ID)
}

func (suite *ResourceStoreTestSuite) TestBuildResourceServerFromResultRow_InvalidType() {
	row := map[string]interface{}{
		"resource_server_id": 123,
		"ou_id":              "ou1",
		"name":               "Test Server",
	}

	rs, err := buildResourceServerFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "resource_server_id")
	suite.Empty(rs.ID)
}

func (suite *ResourceStoreTestSuite) TestBuildResourceFromResultRow_Success() {
	parentID := testParentID1
	row := map[string]interface{}{
		"resource_id":        "res1",
		"resource_server_id": "rs1",
		"name":               "Test Resource",
		"handle":             "test-handle",
		"description":        "Test Description",
		"parent_resource_id": parentID,
	}

	res, err := buildResourceFromResultRow(row)

	suite.NoError(err)
	suite.Equal("res1", res.ID)
	suite.Equal("Test Resource", res.Name)
	suite.Equal("test-handle", res.Handle)
	suite.Equal("Test Description", res.Description)
	suite.NotNil(res.Parent)
	suite.Equal(parentID, *res.Parent)
}

func (suite *ResourceStoreTestSuite) TestBuildResourceFromResultRow_NullParent() {
	row := map[string]interface{}{
		"resource_id":        "res1",
		"resource_server_id": "rs1",
		"name":               "Test Resource",
		"handle":             "test-handle",
		"description":        "Test Description",
		"parent_resource_id": "",
	}

	res, err := buildResourceFromResultRow(row)

	suite.NoError(err)
	suite.Equal("res1", res.ID)
	suite.Equal("test-handle", res.Handle)
	suite.Nil(res.Parent)
}

func (suite *ResourceStoreTestSuite) TestBuildResourceFromResultRow_MissingField() {
	row := map[string]interface{}{
		"resource_id": "res1",
		"name":        "Test Resource",
	}

	_, err := buildResourceFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "handle")
}

func (suite *ResourceStoreTestSuite) TestBuildActionFromResultRow_Success() {
	resourceID := testResourceID1
	row := map[string]interface{}{
		"action_id":          "action1",
		"resource_server_id": "rs1",
		"resource_id":        resourceID,
		"name":               "Test Action",
		"handle":             "test-handle",
		"description":        "Test Description",
	}

	action, err := buildActionFromResultRow(row)

	suite.NoError(err)
	suite.Equal("action1", action.ID)
	suite.Equal("Test Action", action.Name)
	suite.Equal("test-handle", action.Handle)
	suite.Equal("Test Description", action.Description)
	suite.NotNil(action.ResourceID)
	suite.Equal(resourceID, *action.ResourceID)
}

func (suite *ResourceStoreTestSuite) TestBuildActionFromResultRow_NullResource() {
	row := map[string]interface{}{
		"action_id":          "action1",
		"resource_server_id": "rs1",
		"resource_id":        "",
		"name":               "Test Action",
		"handle":             "test-handle",
		"description":        "Test Description",
	}

	action, err := buildActionFromResultRow(row)

	suite.NoError(err)
	suite.Equal("action1", action.ID)
	suite.Equal("test-handle", action.Handle)
	suite.Nil(action.ResourceID)
}

func (suite *ResourceStoreTestSuite) TestBuildActionFromResultRow_MissingField() {
	row := map[string]interface{}{
		"action_id": "action1",
		"name":      "Test Action",
	}

	_, err := buildActionFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "handle")
}
