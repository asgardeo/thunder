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

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/database/provider"

	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

const (
	testParentID1   = "parent1"
	testResourceID1 = "res1"

	// Test constants for IDs and pagination
	testResourceServerInternalID = 5
	testLimit                    = 10
	testOffset                   = 0
)

// ResourceStoreTestSuite is the test suite for resourceStore.
type ResourceStoreTestSuite struct {
	suite.Suite
	mockDBProvider *providermock.DBProviderInterfaceMock
	mockDBClient   *providermock.DBClientInterfaceMock
	store          *resourceStore
}

// TestResourceStoreTestSuite runs the test suite.
func TestResourceStoreTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceStoreTestSuite))
}

// SetupTest sets up the test suite.
func (suite *ResourceStoreTestSuite) SetupTest() {
	suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
	suite.store = &resourceStore{
		dbProvider:   suite.mockDBProvider,
		deploymentID: "test-deployment",
	}
}

// Resource Server Tests

func (suite *ResourceStoreTestSuite) TestCreateResourceServer_Success() {
	rs := ResourceServer{
		OrganizationUnitID: "ou1",
		Name:               "Test Server",
		Description:        "Test Description",
		Identifier:         "test-identifier",
		Delimiter:          ":",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResourceServer, "rs1", "ou1", "Test Server",
		"Test Description", "test-identifier", []byte(`{"delimiter":":"}`), "test-deployment").
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
		Delimiter:          ":",
	}

	execError := errors.New("insert failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResourceServer, "rs1", "ou1", "Test Server",
		"Test Description", "test-identifier", []byte(`{"delimiter":":"}`), "test-deployment").
		Return(int64(0), execError)

	err := suite.store.CreateResourceServer("rs1", rs)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to create resource server")
}

func (suite *ResourceStoreTestSuite) TestCreateResourceServer_DBClientError() {
	rs := ResourceServer{
		OrganizationUnitID: "ou1",
		Name:               "Test Server",
		Description:        "Test Description",
		Identifier:         "test-identifier",
	}

	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	err := suite.store.CreateResourceServer("rs1", rs)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get identity DB client")
}

func (suite *ResourceStoreTestSuite) TestGetResourceServer_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerByID, "rs1", "test-deployment").
		Return([]map[string]interface{}{
			{
				"id":                 7,
				"resource_server_id": "rs1",
				"ou_id":              "ou1",
				"name":               "Test Server",
				"description":        "Test Description",
				"identifier":         "test-identifier",
				"properties":         []byte(`{"delimiter":"/"}`),
			},
		}, nil)

	internalID, rs, err := suite.store.GetResourceServer("rs1")

	suite.NoError(err)
	suite.Equal(7, internalID)
	suite.Equal("rs1", rs.ID)
	suite.Equal("ou1", rs.OrganizationUnitID)
	suite.Equal("Test Server", rs.Name)
	suite.Equal("Test Description", rs.Description)
	suite.Equal("test-identifier", rs.Identifier)
	suite.Equal("/", rs.Delimiter)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServer_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerByID, "nonexistent",
		"test-deployment").Return([]map[string]interface{}{}, nil)

	internalID, rs, err := suite.store.GetResourceServer("nonexistent")

	suite.Error(err)
	suite.Equal(errResourceServerNotFound, err)
	suite.Equal(0, internalID)
	suite.Empty(rs.ID)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServer_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerByID, "rs1", "test-deployment").Return(nil, queryError)

	internalID, rs, err := suite.store.GetResourceServer("rs1")

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get resource server")
	suite.Equal(0, internalID)
	suite.Empty(rs.ID)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServerList_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerList, 10, 0, "test-deployment").
		Return([]map[string]interface{}{
			{
				"id":                 1,
				"resource_server_id": "rs1",
				"ou_id":              "ou1",
				"name":               "Server 1",
				"description":        "Description 1",
				"identifier":         "identifier-1",
			},
			{
				"id":                 2,
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
	suite.mockDBClient.On("Query", queryGetResourceServerList, 10, 0, "test-deployment").Return(nil, queryError)

	servers, err := suite.store.GetResourceServerList(10, 0)

	suite.Error(err)
	suite.Nil(servers)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServerList_InvalidRowData() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerList, 10, 0, "test-deployment").
		Return([]map[string]interface{}{
			{
				"id":                 3,
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
	suite.mockDBClient.On("Query", queryGetResourceServerListCount, "test-deployment").Return([]map[string]interface{}{
		{"total": int64(5)},
	}, nil)

	count, err := suite.store.GetResourceServerListCount()

	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceServerListCount_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceServerListCount, "test-deployment").Return(nil, queryError)

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
		Delimiter:          "-",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateResourceServer, "ou1", "Updated Server",
		"Updated Description", "updated-identifier", []byte(`{"delimiter":"-"}`), "rs1", "test-deployment").
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
		Delimiter:          "-",
	}

	execError := errors.New("update failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateResourceServer, "ou1", "Updated Server",
		"Updated Description", "updated-identifier", []byte(`{"delimiter":"-"}`), "rs1", "test-deployment").
		Return(int64(0), execError)

	err := suite.store.UpdateResourceServer("rs1", rs)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update resource server")
}

func (suite *ResourceStoreTestSuite) TestDeleteResourceServer_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteResourceServer, "rs1", "test-deployment").Return(int64(1), nil)

	err := suite.store.DeleteResourceServer("rs1")

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestDeleteResourceServer_ExecuteError() {
	execError := errors.New("delete failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteResourceServer, "rs1", "test-deployment").Return(int64(0), execError)

	err := suite.store.DeleteResourceServer("rs1")

	suite.Error(err)
	suite.Contains(err.Error(), "failed to delete resource server")
}
func (suite *ResourceStoreTestSuite) TestCheckResourceServerNameExists_Exists() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerNameExists, "Test Server",
		"test-deployment").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.CheckResourceServerNameExists("Test Server")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerNameExists_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerNameExists, "Test Server",
		"test-deployment").Return(nil, queryError)

	exists, err := suite.store.CheckResourceServerNameExists("Test Server")

	suite.Error(err)
	suite.False(exists)
	suite.Contains(err.Error(), "failed to check resource server name")
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerIdentifierExists_Exists() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerIdentifierExists, "test-identifier",
		"test-deployment").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.CheckResourceServerIdentifierExists("test-identifier")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerIdentifierExists_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerIdentifierExists, "test-identifier",
		"test-deployment").Return(nil, queryError)

	exists, err := suite.store.CheckResourceServerIdentifierExists("test-identifier")

	suite.Error(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerHasDependencies_HasDependencies() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerHasDependencies, 1,
		"test-deployment").Return([]map[string]interface{}{
		{"count": int64(3)},
	}, nil)

	hasDeps, err := suite.store.CheckResourceServerHasDependencies(1)

	suite.NoError(err)
	suite.True(hasDeps)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceServerHasDependencies_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceServerHasDependencies, 1,
		"test-deployment").Return(nil, queryError)

	hasDeps, err := suite.store.CheckResourceServerHasDependencies(1)

	suite.Error(err)
	suite.False(hasDeps)
}

// Resource Tests

func (suite *ResourceStoreTestSuite) TestCreateResource_Success() {
	res := Resource{
		Name:        "Test Resource",
		Handle:      "test-handle",
		Description: "Test Description",
		Permission:  "perm:create",
	}
	parentID := 10

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResource, "res1", 5, "Test Resource",
		"test-handle", "Test Description", "perm:create", "{}", &parentID, "test-deployment").
		Return(int64(1), nil)

	err := suite.store.CreateResource("res1", 5, &parentID, res)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestCreateResource_NullParent() {
	res := Resource{
		Name:        "Test Resource",
		Handle:      "test-handle",
		Description: "Test Description",
		Permission:  "perm:create",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResource, "res1", 5, "Test Resource",
		"test-handle", "Test Description", "perm:create", "{}", (*int)(nil), "test-deployment").
		Return(int64(1), nil)

	err := suite.store.CreateResource("res1", 5, nil, res)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestCreateResource_ExecuteError() {
	res := Resource{
		Name:        "Test Resource",
		Handle:      "test-handle",
		Description: "Test Description",
		Permission:  "perm:create",
	}

	execError := errors.New("insert failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateResource, "res1", 5, "Test Resource",
		"test-handle", "Test Description", "perm:create", "{}", (*int)(nil), "test-deployment").
		Return(int64(0), execError)

	err := suite.store.CreateResource("res1", 5, nil, res)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to create resource")
}

func (suite *ResourceStoreTestSuite) TestGetResource_Success() {
	parentID := testParentID1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceByID, testResourceID1, 1,
		"test-deployment").Return([]map[string]interface{}{
		{
			"id":                 11,
			"resource_id":        "res1",
			"resource_server_id": "rs1",
			"name":               "Test Resource",
			"handle":             "test-handle",
			"description":        "Test Description",
			"parent_resource_id": parentID,
			"permission":         "perm:read",
		},
	}, nil)

	internalID, res, err := suite.store.GetResource("res1", 1)

	suite.NoError(err)
	suite.Equal(11, internalID)
	suite.Equal("res1", res.ID)
	suite.Equal("Test Resource", res.Name)
	suite.Equal("test-handle", res.Handle)
	suite.Equal("Test Description", res.Description)
	suite.Equal("perm:read", res.Permission)
	suite.NotNil(res.Parent)
	suite.Equal(parentID, *res.Parent)
}

func (suite *ResourceStoreTestSuite) TestGetResource_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceByID, "nonexistent", 1,
		"test-deployment").Return([]map[string]interface{}{}, nil)

	internalID, res, err := suite.store.GetResource("nonexistent", 1)

	suite.Error(err)
	suite.Equal(errResourceNotFound, err)
	suite.Equal(0, internalID)
	suite.Empty(res.ID)
}

func (suite *ResourceStoreTestSuite) TestGetResourceList_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceList, 1, 10, 0, "test-deployment").Return([]map[string]interface{}{
		{
			"id":                 21,
			"resource_id":        "res1",
			"resource_server_id": "rs1",
			"name":               "Resource 1",
			"handle":             "resource-1",
			"description":        "Description 1",
			"permission":         "perm:r1",
		},
		{
			"id":                 22,
			"resource_id":        "res2",
			"resource_server_id": "rs1",
			"name":               "Resource 2",
			"handle":             "resource-2",
			"description":        "Description 2",
			"permission":         "perm:r2",
		},
	}, nil)

	resources, err := suite.store.GetResourceList(1, 10, 0)

	suite.NoError(err)
	suite.Len(resources, 2)
	suite.Equal("res1", resources[0].ID)
	suite.Equal("Resource 1", resources[0].Name)
	suite.Equal("resource-1", resources[0].Handle)
}

func (suite *ResourceStoreTestSuite) TestGetResourceList_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceList, 1, 10, 0, "test-deployment").Return(nil, queryError)

	resources, err := suite.store.GetResourceList(1, 10, 0)

	suite.Error(err)
	suite.Nil(resources)
	suite.Contains(err.Error(), "failed to get resource list")
}

func (suite *ResourceStoreTestSuite) TestGetResourceList_InvalidRowData() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceList, 1, 10, 0, "test-deployment").Return([]map[string]interface{}{
		{
			"id":                 23,
			"resource_id":        123, // Invalid type
			"resource_server_id": "rs1",
			"name":               "Resource 1",
		},
	}, nil)

	resources, err := suite.store.GetResourceList(1, 10, 0)

	suite.Error(err)
	suite.Nil(resources)
	suite.Contains(err.Error(), "failed to build resource")
}

func (suite *ResourceStoreTestSuite) TestGetResourceListByParent_NullParent() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListByNullParent, 1, 10, 0,
		"test-deployment").Return([]map[string]interface{}{
		{
			"id":                 31,
			"resource_id":        "res1",
			"resource_server_id": "rs1",
			"name":               "Resource 1",
			"handle":             "resource-1",
			"description":        "Description 1",
		},
	}, nil)

	resources, err := suite.store.GetResourceListByParent(1, nil, 10, 0)

	suite.NoError(err)
	suite.Len(resources, 1)
	suite.Equal("res1", resources[0].ID)
	suite.Equal("resource-1", resources[0].Handle)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListByParent_WithParent() {
	parentID := 1
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListByParent, 1, parentID, 10, 0,
		"test-deployment").Return([]map[string]interface{}{
		{
			"id":                 32,
			"resource_id":        "res1",
			"resource_server_id": "rs1",
			"name":               "Resource 1",
			"handle":             "resource-1",
			"description":        "Description 1",
			"parent_resource_id": testParentID1,
		},
	}, nil)

	resources, err := suite.store.GetResourceListByParent(1, &parentID, 10, 0)

	suite.NoError(err)
	suite.Len(resources, 1)
	suite.Equal("res1", resources[0].ID)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListByParent_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListByNullParent, 1, 10, 0,
		"test-deployment").Return(nil, queryError)

	resources, err := suite.store.GetResourceListByParent(1, nil, 10, 0)

	suite.Error(err)
	suite.Nil(resources)
	suite.Contains(err.Error(), "failed to get resource list by parent")
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCount_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCount, 1, "test-deployment").Return([]map[string]interface{}{
		{"total": int64(10)},
	}, nil)

	count, err := suite.store.GetResourceListCount(1)

	suite.NoError(err)
	suite.Equal(10, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCount_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCount, 1, "test-deployment").Return(nil, queryError)

	count, err := suite.store.GetResourceListCount(1)

	suite.Error(err)
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCountByParent_NullParent() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCountByNullParent, 1,
		"test-deployment").Return([]map[string]interface{}{
		{"total": int64(5)},
	}, nil)

	count, err := suite.store.GetResourceListCountByParent(1, nil)

	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCountByParent_WithParent() {
	parentID := 2
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCountByParent, 1, parentID,
		"test-deployment").Return([]map[string]interface{}{
		{"total": int64(3)},
	}, nil)

	count, err := suite.store.GetResourceListCountByParent(1, &parentID)

	suite.NoError(err)
	suite.Equal(3, count)
}

func (suite *ResourceStoreTestSuite) TestGetResourceListCountByParent_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetResourceListCountByNullParent, 1, "test-deployment").Return(nil, queryError)

	count, err := suite.store.GetResourceListCountByParent(1, nil)

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
		"res1", 1, "test-deployment").Return(int64(1), nil)

	err := suite.store.UpdateResource("res1", 1, res)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestUpdateResource_ParentNotFound() {
	res := Resource{
		Name:        "Updated Name",
		Description: "Updated Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateResource, "Updated Name", "Updated Description", "{}",
		"nonexistent", 1, "test-deployment").Return(int64(0), errResourceNotFound)

	err := suite.store.UpdateResource("nonexistent", 1, res)

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
		"res1", 1, "test-deployment").Return(int64(0), execError)

	err := suite.store.UpdateResource("res1", 1, res)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update resource")
}

func (suite *ResourceStoreTestSuite) TestDeleteResource_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteResource, "res1", 1, "test-deployment").Return(int64(1), nil)

	err := suite.store.DeleteResource("res1", 1)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestDeleteResource_ExecuteError() {
	execError := errors.New("delete error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteResource, "res1", 1, "test-deployment").Return(int64(0), execError)

	err := suite.store.DeleteResource("res1", 1)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to delete resource")
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHandleExistsUnderParent_NullParent() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHandleExistsUnderNullParent, 1,
		"Test Resource", "test-deployment").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.CheckResourceHandleExists(1, "Test Resource", nil)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHandleExistsUnderParent_WithParent() {
	parentID := 10
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHandleExistsUnderParent, 1,
		"Test Resource", parentID, "test-deployment").Return([]map[string]interface{}{
		{"count": int64(0)},
	}, nil)

	exists, err := suite.store.CheckResourceHandleExists(1, "Test Resource", &parentID)

	suite.NoError(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHandleExistsUnderParent_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHandleExistsUnderNullParent, 1,
		"Test Resource", "test-deployment").Return(nil, queryError)

	exists, err := suite.store.CheckResourceHandleExists(1, "Test Resource", nil)

	suite.Error(err)
	suite.False(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHasDependencies_HasDependencies() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHasDependencies, 1,
		"test-deployment").Return([]map[string]interface{}{
		{"count": int64(2)},
	}, nil)

	hasDeps, err := suite.store.CheckResourceHasDependencies(1)

	suite.NoError(err)
	suite.True(hasDeps)
}

func (suite *ResourceStoreTestSuite) TestCheckResourceHasDependencies_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckResourceHasDependencies, 1, "test-deployment").Return(nil, queryError)

	hasDeps, err := suite.store.CheckResourceHasDependencies(1)

	suite.Error(err)
	suite.False(hasDeps)
}

func (suite *ResourceStoreTestSuite) TestCheckCircularDependency_HasCircular() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckCircularDependency, "parent1", "res1",
		"test-deployment").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	hasCircular, err := suite.store.CheckCircularDependency("res1", "parent1")

	suite.NoError(err)
	suite.True(hasCircular)
}

func (suite *ResourceStoreTestSuite) TestCheckCircularDependency_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckCircularDependency, "parent1", "res1",
		"test-deployment").Return(nil, queryError)

	hasCircular, err := suite.store.CheckCircularDependency("res1", "parent1")

	suite.Error(err)
	suite.False(hasCircular)
}

// Action Tests

func (suite *ResourceStoreTestSuite) TestCreateAction_Success() {
	action := Action{
		Name:        "Test Action",
		Handle:      "test-handle",
		Description: "Test Description",
		Permission:  "perm:act",
	}
	resourceID := 10

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateAction, "action1", 5, &resourceID,
		"Test Action", "test-handle", "Test Description", "perm:act", "{}", "test-deployment").
		Return(int64(1), nil)

	err := suite.store.CreateAction("action1", 5, &resourceID, action)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestCreateAction_NullResource() {
	action := Action{
		Name:        "Test Action",
		Handle:      "test-handle",
		Description: "Test Description",
		Permission:  "perm:act",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateAction, "action1", 5, (*int)(nil),
		"Test Action", "test-handle", "Test Description", "perm:act", "{}", "test-deployment").
		Return(int64(1), nil)

	err := suite.store.CreateAction("action1", 5, nil, action)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestCreateAction_ExecuteError() {
	action := Action{
		Name:        "Test Action",
		Handle:      "test-handle",
		Description: "Test Description",
		Permission:  "perm:act",
	}

	execError := errors.New("insert error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateAction, "action1", 5, (*int)(nil),
		"Test Action", "test-handle", "Test Description", "perm:act", "{}", "test-deployment").
		Return(int64(0), execError)

	err := suite.store.CreateAction("action1", 5, nil, action)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to create action")
}

func (suite *ResourceStoreTestSuite) TestGetAction_AtResourceServer() {
	var nilResID *int
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionByID, "action1", 1,
		nilResID, "test-deployment").Return([]map[string]interface{}{
		{
			"action_id":          "action1",
			"resource_server_id": "rs1",
			"name":               "Test Action",
			"handle":             "test-handle",
			"description":        "Test Description",
			"permission":         "perm:a",
		},
	}, nil)

	action, err := suite.store.GetAction("action1", 1, nil)

	suite.NoError(err)
	suite.Equal("action1", action.ID)
	suite.Equal("Test Action", action.Name)
	suite.Equal("test-handle", action.Handle)
	suite.Equal("perm:a", action.Permission)
}

func (suite *ResourceStoreTestSuite) TestGetAction_AtResource() {
	resourceID := 10
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionByID, "action1", 1,
		&resourceID, "test-deployment").Return([]map[string]interface{}{
		{
			"action_id":          "action1",
			"resource_server_id": "rs1",
			"resource_id":        testResourceID1,
			"name":               "Test Action",
			"handle":             "test-handle",
			"description":        "Test Description",
			"permission":         "perm:a",
		},
	}, nil)

	action, err := suite.store.GetAction("action1", 1, &resourceID)

	suite.NoError(err)
	suite.Equal("action1", action.ID)
	suite.Equal("test-handle", action.Handle)
	suite.Equal("perm:a", action.Permission)
}

func (suite *ResourceStoreTestSuite) TestGetAction_NotFound() {
	var nilResID *int
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionByID, "nonexistent", 1,
		nilResID, "test-deployment").Return([]map[string]interface{}{}, nil)

	action, err := suite.store.GetAction("nonexistent", 1, nil)

	suite.Error(err)
	suite.Equal(errActionNotFound, err)
	suite.Empty(action.ID)
}

func (suite *ResourceStoreTestSuite) TestGetAction_QueryError() {
	var nilResID *int
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionByID, "action1", 1,
		nilResID, "test-deployment").Return(nil, queryError)

	action, err := suite.store.GetAction("action1", 1, nil)

	suite.Error(err)
	suite.Empty(action.ID)
	suite.Contains(err.Error(), "failed to get action")
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResourceServer_Success() {
	var nilResID *int
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionList, testResourceServerInternalID, nilResID, testLimit, testOffset,
		"test-deployment").Return([]map[string]interface{}{
		{
			"action_id":          "action1",
			"resource_server_id": "rs1",
			"name":               "Action 1",
			"handle":             "action-1",
			"description":        "Description 1",
			"permission":         "perm:1",
		},
		{
			"action_id":          "action2",
			"resource_server_id": "rs1",
			"name":               "Action 2",
			"handle":             "action-2",
			"description":        "Description 2",
			"permission":         "perm:2",
		},
	}, nil)

	actions, err := suite.store.GetActionList(testResourceServerInternalID, nil, testLimit, testOffset)

	suite.NoError(err)
	suite.Len(actions, 2)
	suite.Equal("action1", actions[0].ID)
	suite.Equal("Action 1", actions[0].Name)
	suite.Equal("action-1", actions[0].Handle)
	suite.Equal("perm:1", actions[0].Permission)
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResourceServer_QueryError() {
	var nilResID *int
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionList, testResourceServerInternalID, nilResID, testLimit, testOffset,
		"test-deployment").Return(nil, queryError)

	actions, err := suite.store.GetActionList(testResourceServerInternalID, nil, testLimit, testOffset)

	suite.Error(err)
	suite.Nil(actions)
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResourceServer_InvalidRowData() {
	var nilResID *int
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionList, testResourceServerInternalID, nilResID, testLimit, testOffset,
		"test-deployment").Return([]map[string]interface{}{
		{
			"action_id": 123, // Invalid type
			"name":      "Action 1",
		},
	}, nil)

	actions, err := suite.store.GetActionList(testResourceServerInternalID, nil, testLimit, testOffset)

	suite.Error(err)
	suite.Nil(actions)
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResource_Success() {
	resourceInternalID := 10
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionList, 5, &resourceInternalID, 10, 0,
		"test-deployment").Return([]map[string]interface{}{
		{
			"action_id":          "action1",
			"resource_server_id": "rs1",
			"resource_id":        "res1",
			"name":               "Action 1",
			"handle":             "action-1",
			"description":        "Description 1",
			"permission":         "perm:r",
		},
	}, nil)

	actions, err := suite.store.GetActionList(5, &resourceInternalID, 10, 0)

	suite.NoError(err)
	suite.Len(actions, 1)
	suite.Equal("action1", actions[0].ID)
	suite.Equal("action-1", actions[0].Handle)
	suite.Equal("perm:r", actions[0].Permission)
}

func (suite *ResourceStoreTestSuite) TestGetActionListAtResource_QueryError() {
	resourceInternalID := 10
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionList, 5, &resourceInternalID, 10, 0,
		"test-deployment").Return(nil, queryError)

	actions, err := suite.store.GetActionList(5, &resourceInternalID, 10, 0)

	suite.Error(err)
	suite.Nil(actions)
}

func (suite *ResourceStoreTestSuite) TestGetActionListCountAtResourceServer_Success() {
	var nilResID *int
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListCount, testResourceServerInternalID,
		nilResID, "test-deployment").Return([]map[string]interface{}{
		{"total": int64(15)},
	}, nil)

	count, err := suite.store.GetActionListCount(testResourceServerInternalID, nil)

	suite.NoError(err)
	suite.Equal(15, count)
}

func (suite *ResourceStoreTestSuite) TestGetActionListCountAtResourceServer_QueryError() {
	var nilResID *int
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListCount, testResourceServerInternalID,
		nilResID, "test-deployment").Return(nil, queryError)

	count, err := suite.store.GetActionListCount(testResourceServerInternalID, nil)

	suite.Error(err)
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestGetActionListCountAtResource_Success() {
	resourceInternalID := 10
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListCount, 5, &resourceInternalID,
		"test-deployment").Return([]map[string]interface{}{
		{"total": int64(5)},
	}, nil)

	count, err := suite.store.GetActionListCount(5, &resourceInternalID)

	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *ResourceStoreTestSuite) TestGetActionListCountAtResource_QueryError() {
	resourceInternalID := 10
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetActionListCount, 5, &resourceInternalID,
		"test-deployment").Return(nil, queryError)

	count, err := suite.store.GetActionListCount(5, &resourceInternalID)

	suite.Error(err)
	suite.Equal(0, count)
}

func (suite *ResourceStoreTestSuite) TestUpdateAction_Success() {
	var nilResID *int
	action := Action{
		Name:        "Updated Action",
		Description: "Updated Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateAction, "Updated Action",
		"Updated Description", "{}", "action1", testResourceServerInternalID, nilResID, "test-deployment").
		Return(int64(1), nil)

	err := suite.store.UpdateAction("action1", testResourceServerInternalID, nil, action)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestUpdateAction_WithResourceID() {
	resourceInternalID := 10
	action := Action{
		Name:        "Updated Action",
		Description: "Updated Description",
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateAction, "Updated Action",
		"Updated Description", "{}", "action1", 5, &resourceInternalID, "test-deployment").
		Return(int64(1), nil)

	err := suite.store.UpdateAction("action1", 5, &resourceInternalID, action)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestUpdateAction_ExecuteError() {
	var nilResID *int
	action := Action{
		Name:        "Updated Action",
		Description: "Updated Description",
	}

	execError := errors.New("update error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateAction, "Updated Action",
		"Updated Description", "{}", "action1", testResourceServerInternalID, nilResID, "test-deployment").
		Return(int64(0), execError)

	err := suite.store.UpdateAction("action1", testResourceServerInternalID, nil, action)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update action")
}

func (suite *ResourceStoreTestSuite) TestDeleteAction_Success() {
	var nilResID *int
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On(
		"Execute", queryDeleteAction, "action1", testResourceServerInternalID, nilResID, "test-deployment",
	).Return(int64(1), nil)

	err := suite.store.DeleteAction("action1", testResourceServerInternalID, nil)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestDeleteAction_WithResourceID() {
	resourceInternalID := 10
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteAction, "action1", 5,
		&resourceInternalID, "test-deployment").Return(int64(1), nil)

	err := suite.store.DeleteAction("action1", 5, &resourceInternalID)

	suite.NoError(err)
}

func (suite *ResourceStoreTestSuite) TestDeleteAction_ExecuteError() {
	var nilResID *int
	execError := errors.New("delete error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On(
		"Execute", queryDeleteAction, "action1", testResourceServerInternalID, nilResID, "test-deployment",
	).Return(int64(0), execError)

	err := suite.store.DeleteAction("action1", testResourceServerInternalID, nil)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to delete action")
}

func (suite *ResourceStoreTestSuite) TestIsActionExist_AtResourceServer() {
	var nilResID *int
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionExists, "action1", testResourceServerInternalID,
		nilResID, "test-deployment").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.IsActionExist("action1", testResourceServerInternalID, nil)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestIsActionExist_QueryError() {
	var nilResID *int
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionExists, "action1", testResourceServerInternalID,
		nilResID, "test-deployment").Return(nil, queryError)

	exists, err := suite.store.IsActionExist("action1", testResourceServerInternalID, nil)

	suite.Error(err)
	suite.False(exists)
	suite.Contains(err.Error(), "failed to check action existence")
}

func (suite *ResourceStoreTestSuite) TestIsActionExist_AtResource() {
	resourceInternalID := 10
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionExists, "action1", 5,
		&resourceInternalID, "test-deployment").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.IsActionExist("action1", 5, &resourceInternalID)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckActionHandleExists_AtResourceServer() {
	var nilResID *int
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionHandleExists, testResourceServerInternalID,
		nilResID, "Test Action", "test-deployment").Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.CheckActionHandleExists(testResourceServerInternalID, nil, "Test Action")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *ResourceStoreTestSuite) TestCheckActionHandleExists_QueryError() {
	var nilResID *int
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionHandleExists, testResourceServerInternalID,
		nilResID, "Test Action", "test-deployment").Return(nil, queryError)

	exists, err := suite.store.CheckActionHandleExists(testResourceServerInternalID, nil, "Test Action")

	suite.Error(err)
	suite.False(exists)
	suite.Contains(err.Error(), "failed to check action handle")
}

func (suite *ResourceStoreTestSuite) TestCheckActionHandleExists_AtResource() {
	resourceInternalID := 10
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckActionHandleExists, 5,
		&resourceInternalID, "Test Action", "test-deployment").Return([]map[string]interface{}{
		{"count": int64(0)},
	}, nil)

	exists, err := suite.store.CheckActionHandleExists(5, &resourceInternalID, "Test Action")

	suite.NoError(err)
	suite.False(exists)
}

// Helper Function Tests

func (suite *ResourceStoreTestSuite) TestGetIdentityDBClient_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)

	client, err := suite.store.getConfigDBClient()

	suite.NoError(err)
	suite.NotNil(client)
	suite.Equal(suite.mockDBClient, client)
}

func (suite *ResourceStoreTestSuite) TestGetIdentityDBClient_Error() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	client, err := suite.store.getConfigDBClient()

	suite.Error(err)
	suite.Nil(client)
	suite.Contains(err.Error(), "failed to get identity DB client")
}

func (suite *ResourceStoreTestSuite) TestWithDBClient_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)

	called := false
	err := suite.store.withDBClient(func(dbClient provider.DBClientInterface) error {
		called = true
		suite.Equal(suite.mockDBClient, dbClient)
		return nil
	})

	suite.NoError(err)
	suite.True(called)
}

func (suite *ResourceStoreTestSuite) TestWithDBClient_DBClientError() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	called := false
	err := suite.store.withDBClient(func(dbClient provider.DBClientInterface) error {
		called = true
		return nil
	})

	suite.Error(err)
	suite.False(called, "Function should not be called when DB client retrieval fails")
	suite.Contains(err.Error(), "failed to get identity DB client")
}

func (suite *ResourceStoreTestSuite) TestWithDBClient_InnerFunctionError() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)

	expectedErr := errors.New("inner function error")
	err := suite.store.withDBClient(func(dbClient provider.DBClientInterface) error {
		return expectedErr
	})

	suite.Error(err)
	suite.Equal(expectedErr, err, "Error from inner function should be propagated")
}

func (suite *ResourceStoreTestSuite) TestParseCountResult() {
	testCases := []struct {
		name          string
		results       []map[string]interface{}
		expectedCount int
		shouldErr     bool
		errContains   string
	}{
		{
			name: "Success_TotalField_Int64",
			results: []map[string]interface{}{
				{"total": int64(42)},
			},
			expectedCount: 42,
			shouldErr:     false,
		},
		{
			name: "Success_CountField_Int64",
			results: []map[string]interface{}{
				{"count": int64(42)},
			},
			expectedCount: 42,
			shouldErr:     false,
		},
		{
			name: "Success_TotalField_Int",
			results: []map[string]interface{}{
				{"total": int(42)},
			},
			expectedCount: 42,
			shouldErr:     false,
		},
		{
			name: "Success_TotalField_Float64",
			results: []map[string]interface{}{
				{"total": float64(42)},
			},
			expectedCount: 42,
			shouldErr:     false,
		},
		{
			name:          "Error_EmptyResults",
			results:       []map[string]interface{}{},
			expectedCount: 0,
			shouldErr:     true,
			errContains:   "no count result returned",
		},
		{
			name: "Error_MissingField",
			results: []map[string]interface{}{
				{"other": int64(42)},
			},
			expectedCount: 0,
			shouldErr:     true,
			errContains:   "count field not found",
		},
		{
			name: "Error_InvalidType",
			results: []map[string]interface{}{
				{"total": "not_a_number"},
			},
			expectedCount: 0,
			shouldErr:     true,
			errContains:   "unexpected count type",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			count, err := parseCountResult(tc.results)

			if tc.shouldErr {
				suite.Error(err)
				suite.Contains(err.Error(), tc.errContains)
				suite.Equal(0, count)
			} else {
				suite.NoError(err)
				suite.Equal(tc.expectedCount, count)
			}
		})
	}
}

func (suite *ResourceStoreTestSuite) TestParseBoolFromCount() {
	testCases := []struct {
		name         string
		results      []map[string]interface{}
		expectedBool bool
		shouldErr    bool
	}{
		{
			name: "Success_True",
			results: []map[string]interface{}{
				{"count": int64(5)},
			},
			expectedBool: true,
			shouldErr:    false,
		},
		{
			name: "Success_False",
			results: []map[string]interface{}{
				{"count": int64(0)},
			},
			expectedBool: false,
			shouldErr:    false,
		},
		{
			name:         "Error_EmptyResults",
			results:      []map[string]interface{}{},
			expectedBool: false,
			shouldErr:    true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			exists, err := parseBoolFromCount(tc.results)

			if tc.shouldErr {
				suite.Error(err)
				suite.False(exists)
			} else {
				suite.NoError(err)
				suite.Equal(tc.expectedBool, exists)
			}
		})
	}
}

func (suite *ResourceStoreTestSuite) TestBuildResourceServerFromResultRow() {
	testCases := []struct {
		name                   string
		row                    map[string]interface{}
		expectedResourceServer ResourceServer
		expectedInternalID     int
		shouldErr              bool
		errContains            string
	}{
		{
			name: "Success_AllFields",
			row: map[string]interface{}{
				"id":                 50,
				"resource_server_id": "rs1",
				"ou_id":              "ou1",
				"name":               "Test Server",
				"description":        "Test Description",
				"identifier":         "test-identifier",
				"properties":         []byte(`{"delimiter":"|"}`),
			},
			expectedResourceServer: ResourceServer{
				ID:                 "rs1",
				OrganizationUnitID: "ou1",
				Name:               "Test Server",
				Description:        "Test Description",
				Identifier:         "test-identifier",
				Delimiter:          "|",
			},
			expectedInternalID: 50,
			shouldErr:          false,
		},
		{
			name: "Success_OptionalFields",
			row: map[string]interface{}{
				"id":                 51,
				"resource_server_id": "rs1",
				"ou_id":              "ou1",
				"name":               "Test Server",
			},
			expectedResourceServer: ResourceServer{
				ID:                 "rs1",
				OrganizationUnitID: "ou1",
				Name:               "Test Server",
				Description:        "",
				Identifier:         "",
			},
			expectedInternalID: 51,
			shouldErr:          false,
		},
		{
			name: "Success_PropertiesString",
			row: map[string]interface{}{
				"id":                 52,
				"resource_server_id": "rs1",
				"ou_id":              "ou1",
				"name":               "Test Server",
				"properties":         `{"delimiter":"."}`,
			},
			expectedResourceServer: ResourceServer{
				ID:                 "rs1",
				OrganizationUnitID: "ou1",
				Name:               "Test Server",
				Delimiter:          ".",
			},
			expectedInternalID: 52,
			shouldErr:          false,
		},
		{
			name: "Error_MissingResourceServerID",
			row: map[string]interface{}{
				"id":    60,
				"ou_id": "ou1",
				"name":  "Test Server",
			},
			shouldErr:   true,
			errContains: "resource_server_id",
		},
		{
			name: "Error_InvalidResourceServerID",
			row: map[string]interface{}{
				"id":                 61,
				"resource_server_id": 123,
				"ou_id":              "ou1",
				"name":               "Test Server",
			},
			shouldErr:   true,
			errContains: "resource_server_id",
		},
		{
			name: "Error_MissingOuID",
			row: map[string]interface{}{
				"id":                 62,
				"resource_server_id": "rs1",
				"name":               "Test Server",
			},
			shouldErr:   true,
			errContains: "ou_id",
		},
		{
			name: "Error_InvalidOuID",
			row: map[string]interface{}{
				"id":                 63,
				"resource_server_id": "rs1",
				"ou_id":              123,
				"name":               "Test Server",
			},
			shouldErr:   true,
			errContains: "ou_id",
		},
		{
			name: "Error_MissingName",
			row: map[string]interface{}{
				"id":                 64,
				"resource_server_id": "rs1",
				"ou_id":              "ou1",
			},
			shouldErr:   true,
			errContains: "name",
		},
		{
			name: "Error_InvalidName",
			row: map[string]interface{}{
				"id":                 65,
				"resource_server_id": "rs1",
				"ou_id":              "ou1",
				"name":               123,
			},
			shouldErr:   true,
			errContains: "name",
		},
		{
			name: "Error_MissingInternalID",
			row: map[string]interface{}{
				"resource_server_id": "rs1",
				"ou_id":              "ou1",
				"name":               "Test Server",
			},
			shouldErr:   true,
			errContains: "unexpected internal ID",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			internalID, rs, err := buildResourceServerFromResultRow(tc.row)

			if tc.shouldErr {
				suite.Error(err)
				suite.Contains(err.Error(), tc.errContains)
			} else {
				suite.NoError(err)
				suite.Equal(tc.expectedInternalID, internalID)
				suite.Equal(tc.expectedResourceServer.ID, rs.ID)
				suite.Equal(tc.expectedResourceServer.OrganizationUnitID, rs.OrganizationUnitID)
				suite.Equal(tc.expectedResourceServer.Name, rs.Name)
				suite.Equal(tc.expectedResourceServer.Description, rs.Description)
				suite.Equal(tc.expectedResourceServer.Identifier, rs.Identifier)
				suite.Equal(tc.expectedResourceServer.Delimiter, rs.Delimiter)
			}
		})
	}
}

func (suite *ResourceStoreTestSuite) TestBuildResourceFromResultRow() {
	parentID := testParentID1
	testCases := []struct {
		name             string
		row              map[string]interface{}
		expectedResource Resource
		expectedID       int
		shouldErr        bool
		errContains      string
	}{
		{
			name: "Success_WithParent",
			row: map[string]interface{}{
				"id":                 70,
				"resource_id":        "res1",
				"resource_server_id": "rs1",
				"name":               "Test Resource",
				"handle":             "test-handle",
				"description":        "Test Description",
				"parent_resource_id": parentID,
				"permission":         "perm:r",
			},
			expectedResource: Resource{
				ID:          "res1",
				Name:        "Test Resource",
				Handle:      "test-handle",
				Description: "Test Description",
				Permission:  "perm:r",
				Parent:      &parentID,
			},
			expectedID: 70,
			shouldErr:  false,
		},
		{
			name: "Success_NullParent",
			row: map[string]interface{}{
				"id":                 71,
				"resource_id":        "res1",
				"resource_server_id": "rs1",
				"name":               "Test Resource",
				"handle":             "test-handle",
				"description":        "Test Description",
				"parent_resource_id": "",
				"permission":         "perm:r",
			},
			expectedResource: Resource{
				ID:          "res1",
				Name:        "Test Resource",
				Handle:      "test-handle",
				Description: "Test Description",
				Permission:  "perm:r",
				Parent:      nil,
			},
			expectedID: 71,
			shouldErr:  false,
		},
		{
			name: "Success_EmptyDescription",
			row: map[string]interface{}{
				"id":          72,
				"resource_id": "res1",
				"name":        "Test Resource",
				"handle":      "test-handle",
				"description": "",
				"permission":  "perm:r",
			},
			expectedResource: Resource{
				ID:          "res1",
				Name:        "Test Resource",
				Handle:      "test-handle",
				Description: "",
				Permission:  "perm:r",
				Parent:      nil,
			},
			expectedID: 72,
			shouldErr:  false,
		},
		{
			name: "Success_MissingDescription",
			row: map[string]interface{}{
				"id":          73,
				"resource_id": "res1",
				"name":        "Test Resource",
				"handle":      "test-handle",
				"permission":  "perm:r",
			},
			expectedResource: Resource{
				ID:          "res1",
				Name:        "Test Resource",
				Handle:      "test-handle",
				Description: "",
				Permission:  "perm:r",
				Parent:      nil,
			},
			expectedID: 73,
			shouldErr:  false,
		},
		{
			name: "Error_MissingResourceID",
			row: map[string]interface{}{
				"id":     80,
				"name":   "Test Resource",
				"handle": "test-handle",
			},
			shouldErr:   true,
			errContains: "resource_id",
		},
		{
			name: "Error_InvalidResourceID",
			row: map[string]interface{}{
				"id":          81,
				"resource_id": 123,
				"name":        "Test Resource",
				"handle":      "test-handle",
			},
			shouldErr:   true,
			errContains: "resource_id",
		},
		{
			name: "Error_MissingName",
			row: map[string]interface{}{
				"id":          82,
				"resource_id": "res1",
				"handle":      "test-handle",
			},
			shouldErr:   true,
			errContains: "name",
		},
		{
			name: "Error_InvalidName",
			row: map[string]interface{}{
				"id":          83,
				"resource_id": "res1",
				"name":        123,
				"handle":      "test-handle",
			},
			shouldErr:   true,
			errContains: "name",
		},
		{
			name: "Error_MissingHandle",
			row: map[string]interface{}{
				"id":          84,
				"resource_id": "res1",
				"name":        "Test Resource",
			},
			shouldErr:   true,
			errContains: "handle",
		},
		{
			name: "Error_InvalidHandle",
			row: map[string]interface{}{
				"id":          85,
				"resource_id": "res1",
				"name":        "Test Resource",
				"handle":      123,
			},
			shouldErr:   true,
			errContains: "handle",
		},
		{
			name: "Error_MissingInternalID",
			row: map[string]interface{}{
				"resource_id": "res1",
				"name":        "Test Resource",
				"handle":      "test-handle",
			},
			shouldErr:   true,
			errContains: "unexpected internal ID type",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			internalID, res, err := buildResourceFromResultRow(tc.row)

			if tc.shouldErr {
				suite.Error(err)
				suite.Contains(err.Error(), tc.errContains)
			} else {
				suite.NoError(err)
				suite.Equal(tc.expectedID, internalID)
				suite.Equal(tc.expectedResource.ID, res.ID)
				suite.Equal(tc.expectedResource.Name, res.Name)
				suite.Equal(tc.expectedResource.Handle, res.Handle)
				suite.Equal(tc.expectedResource.Description, res.Description)
				suite.Equal(tc.expectedResource.Permission, res.Permission)
				if tc.expectedResource.Parent != nil {
					suite.NotNil(res.Parent)
					suite.Equal(*tc.expectedResource.Parent, *res.Parent)
				} else {
					suite.Nil(res.Parent)
				}
			}
		})
	}
}

func (suite *ResourceStoreTestSuite) TestBuildActionFromResultRow() {
	testCases := []struct {
		name           string
		row            map[string]interface{}
		expectedAction Action
		shouldErr      bool
		errContains    string
	}{
		{
			name: "Success_WithResourceID",
			row: map[string]interface{}{
				"action_id":          "action1",
				"resource_server_id": "rs1",
				"resource_id":        testResourceID1,
				"name":               "Test Action",
				"handle":             "test-handle",
				"description":        "Test Description",
			},
			expectedAction: Action{
				ID:          "action1",
				Name:        "Test Action",
				Handle:      "test-handle",
				Description: "Test Description",
			},
			shouldErr: false,
		},
		{
			name: "Success_NullResource",
			row: map[string]interface{}{
				"action_id":          "action1",
				"resource_server_id": "rs1",
				"resource_id":        "",
				"name":               "Test Action",
				"handle":             "test-handle",
				"description":        "Test Description",
			},
			expectedAction: Action{
				ID:          "action1",
				Name:        "Test Action",
				Handle:      "test-handle",
				Description: "Test Description",
			},
			shouldErr: false,
		},
		{
			name: "Success_EmptyDescription",
			row: map[string]interface{}{
				"action_id":   "action1",
				"name":        "Test Action",
				"handle":      "test-handle",
				"description": "",
			},
			expectedAction: Action{
				ID:          "action1",
				Name:        "Test Action",
				Handle:      "test-handle",
				Description: "",
			},
			shouldErr: false,
		},
		{
			name: "Success_MissingDescription",
			row: map[string]interface{}{
				"action_id": "action1",
				"name":      "Test Action",
				"handle":    "test-handle",
			},
			expectedAction: Action{
				ID:          "action1",
				Name:        "Test Action",
				Handle:      "test-handle",
				Description: "",
			},
			shouldErr: false,
		},
		{
			name: "Error_MissingActionID",
			row: map[string]interface{}{
				"name":   "Test Action",
				"handle": "test-handle",
			},
			shouldErr:   true,
			errContains: "action_id",
		},
		{
			name: "Error_InvalidActionID",
			row: map[string]interface{}{
				"action_id": 123,
				"name":      "Test Action",
				"handle":    "test-handle",
			},
			shouldErr:   true,
			errContains: "action_id",
		},
		{
			name: "Error_MissingName",
			row: map[string]interface{}{
				"action_id": "action1",
				"handle":    "test-handle",
			},
			shouldErr:   true,
			errContains: "name",
		},
		{
			name: "Error_InvalidName",
			row: map[string]interface{}{
				"action_id": "action1",
				"name":      123,
				"handle":    "test-handle",
			},
			shouldErr:   true,
			errContains: "name",
		},
		{
			name: "Error_MissingHandle",
			row: map[string]interface{}{
				"action_id": "action1",
				"name":      "Test Action",
			},
			shouldErr:   true,
			errContains: "handle",
		},
		{
			name: "Error_InvalidHandle",
			row: map[string]interface{}{
				"action_id": "action1",
				"name":      "Test Action",
				"handle":    123,
			},
			shouldErr:   true,
			errContains: "handle",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			action, err := buildActionFromResultRow(tc.row)

			if tc.shouldErr {
				suite.Error(err)
				suite.Contains(err.Error(), tc.errContains)
			} else {
				suite.NoError(err)
				suite.Equal(tc.expectedAction.ID, action.ID)
				suite.Equal(tc.expectedAction.Name, action.Name)
				suite.Equal(tc.expectedAction.Handle, action.Handle)
				suite.Equal(tc.expectedAction.Description, action.Description)
			}
		})
	}
}

// resolveInternalID Tests

func (suite *ResourceStoreTestSuite) TestResolveInternalID() {
	testCases := []struct {
		name        string
		row         map[string]interface{}
		expectedID  int
		shouldErr   bool
		errContains string
	}{
		{
			name: "Success_Int",
			row: map[string]interface{}{
				"id": 123,
			},
			expectedID: 123,
			shouldErr:  false,
		},
		{
			name: "Success_Int64",
			row: map[string]interface{}{
				"id": int64(456),
			},
			expectedID: 456,
			shouldErr:  false,
		},
		{
			name: "Success_Float64",
			row: map[string]interface{}{
				"id": float64(789),
			},
			expectedID: 789,
			shouldErr:  false,
		},
		{
			name: "Success_Float64WithDecimals",
			row: map[string]interface{}{
				"id": float64(789.99),
			},
			expectedID: 789, // Truncates to int
			shouldErr:  false,
		},
		{
			name: "Error_UnexpectedType",
			row: map[string]interface{}{
				"id": "not-a-number",
			},
			shouldErr:   true,
			errContains: "unexpected internal ID type",
		},
		{
			name: "Error_MissingID",
			row: map[string]interface{}{
				"other_field": "value",
			},
			shouldErr:   true,
			errContains: "unexpected internal ID type",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			id, err := resolveInternalID(tc.row)

			if tc.shouldErr {
				suite.Error(err)
				suite.Contains(err.Error(), tc.errContains)
			} else {
				suite.NoError(err)
				suite.Equal(tc.expectedID, id)
			}
		})
	}
}

func (suite *ResourceStoreTestSuite) TestResolveIdentifier() {
	testCases := []struct {
		name       string
		identifier string
		expected   interface{}
	}{
		{
			name:       "Success_NonEmptyIdentifier",
			identifier: "https://api.example.com",
			expected:   "https://api.example.com",
		},
		{
			name:       "Success_AnotherNonEmptyIdentifier",
			identifier: "urn:example:resource:server",
			expected:   "urn:example:resource:server",
		},
		{
			name:       "Success_EmptyIdentifier_ReturnsNil",
			identifier: "",
			expected:   nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := resolveIdentifier(tc.identifier)
			suite.Equal(tc.expected, result)
		})
	}
}
