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
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

const testDeploymentID = "test-deployment-id"

type OrganizationUnitStoreTestSuite struct {
	suite.Suite
	providerMock *providermock.DBProviderInterfaceMock
	dbClientMock *providermock.DBClientInterfaceMock
	store        *organizationUnitStore
}

func TestOrganizationUnitStoreTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationUnitStoreTestSuite))
}

func (suite *OrganizationUnitStoreTestSuite) SetupTest() {
	suite.providerMock = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.dbClientMock = providermock.NewDBClientInterfaceMock(suite.T())
	suite.store = &organizationUnitStore{
		dbProvider:   suite.providerMock,
		deploymentID: testDeploymentID,
	}
}

func (suite *OrganizationUnitStoreTestSuite) expectDBClient() {
	suite.providerMock.On("GetUserDBClient").Return(suite.dbClientMock, nil)
}

type conflictTestCase struct {
	name      string
	hasParent bool
	parentVal string
	setup     func(parent *string)
	want      bool
	wantErr   string
}

// runConflictTestCases centralizes the repeated assertion flow for parent-aware conflict checks.
func (suite *OrganizationUnitStoreTestSuite) runConflictTestCases(
	tests []conflictTestCase,
	invoke func(parent *string) (bool, error),
) {
	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			var parent *string
			if tc.hasParent {
				p := tc.parentVal
				parent = &p
			}
			if tc.setup != nil {
				tc.setup(parent)
			}

			result, err := invoke(parent)

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
			suite.Equal(tc.want, result)
		})
	}
}

type countTestCase struct {
	name    string
	setup   func()
	want    int
	wantErr string
}

// runCountTestCases removes duplicated boilerplate around count store methods.
func (suite *OrganizationUnitStoreTestSuite) runCountTestCases(
	tests []countTestCase,
	invoke func() (int, error),
) {
	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			if tc.setup != nil {
				tc.setup()
			}

			result, err := invoke()

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
			suite.Equal(tc.want, result)
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) runConflictQueryScenario(
	withParentQueryID, withoutParentQueryID interface{}, value string,
	withParentCount, withoutParentCount int64,
	invoke func(parent *string) (bool, error),
) {
	parentLabel := "with parent"
	if withParentCount > 0 {
		parentLabel += " conflict"
	} else {
		parentLabel += " no conflict"
	}

	rootLabel := "without parent"
	if withoutParentCount > 0 {
		rootLabel += " conflict"
	} else {
		rootLabel += " no conflict"
	}

	suite.runConflictTestCases(
		[]conflictTestCase{
			{
				name:      parentLabel,
				hasParent: true,
				parentVal: testParentID,
				setup: func(parent *string) {
					suite.expectDBClient()
					suite.dbClientMock.
						On("Query", withParentQueryID, value, *parent, testDeploymentID).
						Return([]map[string]interface{}{{"count": withParentCount}}, nil).
						Once()
				},
				want: withParentCount > 0,
			},
			{
				name: rootLabel,
				setup: func(_ *string) {
					suite.expectDBClient()
					suite.dbClientMock.
						On("Query", withoutParentQueryID, value, testDeploymentID).
						Return([]map[string]interface{}{{"count": withoutParentCount}}, nil).
						Once()
				},
				want: withoutParentCount > 0,
			},
			{
				name: "query error",
				setup: func(_ *string) {
					suite.expectDBClient()
					suite.dbClientMock.
						On("Query", withoutParentQueryID, value, testDeploymentID).
						Return(nil, errors.New("query err")).
						Once()
				},
				wantErr: "failed to execute query",
			},
			{
				name: "db client error",
				setup: func(_ *string) {
					suite.providerMock.
						On("GetUserDBClient").
						Return(nil, errors.New("db err")).
						Once()
				},
				wantErr: "failed to get database client",
			},
		},
		invoke,
	)
}

func (suite *OrganizationUnitStoreTestSuite) runCountQueryScenario(
	queryID interface{}, arg string, deploymentID string,
	successCount int,
	invoke func() (int, error),
) {
	suite.runCountTestCases(
		[]countTestCase{
			{
				name: "success",
				setup: func() {
					suite.expectDBClient()
					suite.dbClientMock.
						On("Query", queryID, arg, deploymentID).
						Return([]map[string]interface{}{{"total": int64(successCount)}}, nil).
						Once()
				},
				want: successCount,
			},
			{
				name: "empty result",
				setup: func() {
					suite.expectDBClient()
					suite.dbClientMock.
						On("Query", queryID, arg, deploymentID).
						Return([]map[string]interface{}{}, nil).
						Once()
				},
				want: 0,
			},
			{
				name: "invalid type",
				setup: func() {
					suite.expectDBClient()
					suite.dbClientMock.
						On("Query", queryID, arg, deploymentID).
						Return([]map[string]interface{}{{"total": "bad"}}, nil).
						Once()
				},
				wantErr: "failed to parse count result",
			},
			{
				name: "query error",
				setup: func() {
					suite.expectDBClient()
					suite.dbClientMock.
						On("Query", queryID, arg, deploymentID).
						Return(nil, errors.New("query err")).
						Once()
				},
				wantErr: "failed to execute count query",
			},
			{
				name: "db client error",
				setup: func() {
					suite.providerMock.
						On("GetUserDBClient").
						Return(nil, errors.New("db err")).
						Once()
				},
				wantErr: "failed to get database client",
			},
		},
		invoke,
	)
}

func makeOUResultRow(id, handle, name, description string, parent *string) map[string]interface{} {
	row := map[string]interface{}{
		"ou_id":       id,
		"handle":      handle,
		"name":        name,
		"description": description,
	}
	if parent != nil {
		row["parent_id"] = *parent
	} else {
		row["parent_id"] = nil
	}
	return row
}

func TestBuildOrganizationUnitBasicFromResultRow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		row := map[string]interface{}{
			"ou_id":       "ou1",
			"handle":      "root",
			"name":        "Root",
			"description": "desc",
		}

		ou, err := buildOrganizationUnitBasicFromResultRow(row)

		require.NoError(t, err)
		require.Equal(t, "ou1", ou.ID)
		require.Equal(t, "desc", ou.Description)
	})

	t.Run("success with design fields", func(t *testing.T) {
		row := map[string]interface{}{
			"ou_id":       "ou1",
			"handle":      "root",
			"name":        "Root",
			"description": "desc",
			"theme_id":    "theme-123",
			"layout_id":   "layout-456",
			"logo_url":    "https://example.com/logo.png",
		}

		ou, err := buildOrganizationUnitBasicFromResultRow(row)

		require.NoError(t, err)
		require.Equal(t, "ou1", ou.ID)
		require.Equal(t, "https://example.com/logo.png", ou.LogoURL)
	})

	t.Run("success with nil design fields", func(t *testing.T) {
		row := map[string]interface{}{
			"ou_id":       "ou1",
			"handle":      "root",
			"name":        "Root",
			"description": "desc",
			"theme_id":    nil,
			"layout_id":   nil,
			"logo_url":    nil,
		}

		ou, err := buildOrganizationUnitBasicFromResultRow(row)

		require.NoError(t, err)
		require.Equal(t, "ou1", ou.ID)
		require.Equal(t, "", ou.LogoURL)
	})

	tests := []struct {
		name string
		row  map[string]interface{}
		want string
	}{
		{
			name: "missing ou id",
			row: map[string]interface{}{
				"name":   "Root",
				"handle": "root",
			},
			want: "ou_id is not a string",
		},
		{
			name: "missing name",
			row: map[string]interface{}{
				"ou_id":  "ou1",
				"handle": "root",
			},
			want: "name is not a string",
		},
		{
			name: "missing handle",
			row: map[string]interface{}{
				"ou_id": "ou1",
				"name":  "Root",
			},
			want: "handle is not a string",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildOrganizationUnitBasicFromResultRow(tc.row)

			require.Error(t, err)
			require.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestBuildOrganizationUnitFromResultRow(t *testing.T) {
	parentID := testParentID
	row := map[string]interface{}{
		"ou_id":       "child",
		"handle":      "child",
		"name":        "Child",
		"description": "",
		"parent_id":   parentID,
	}

	ou, err := buildOrganizationUnitFromResultRow(row)

	require.NoError(t, err)
	require.NotNil(t, ou.Parent)
	require.Equal(t, parentID, *ou.Parent)

	t.Run("with design fields", func(t *testing.T) {
		row := map[string]interface{}{
			"ou_id":       "ou1",
			"handle":      "root",
			"name":        "Root",
			"description": "desc",
			"parent_id":   nil,
			"theme_id":    "theme-abc",
			"layout_id":   "layout-def",
			"metadata": `{"logo_url":"https://example.com/logo.png","tos_uri":""` +
				`,"policy_uri":"","cookie_policy_uri":""}`,
		}

		ou, err := buildOrganizationUnitFromResultRow(row)

		require.NoError(t, err)
		require.Nil(t, ou.Parent)
		require.Equal(t, "theme-abc", ou.ThemeID)
		require.Equal(t, "layout-def", ou.LayoutID)
		require.Equal(t, "https://example.com/logo.png", ou.LogoURL)
	})

	t.Run("invalid parent type", func(t *testing.T) {
		row := map[string]interface{}{
			"ou_id":       "ou1",
			"handle":      "root",
			"name":        "Root",
			"description": "",
			"parent_id":   123,
		}

		ou, err := buildOrganizationUnitFromResultRow(row)

		require.NoError(t, err)
		require.Nil(t, ou.Parent)
	})

	t.Run("builder error", func(t *testing.T) {
		row := map[string]interface{}{
			"handle": "root",
			"name":   "Root",
		}

		_, err := buildOrganizationUnitFromResultRow(row)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to build organization unit")
	})
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_CheckOrganizationUnitHasChildResources() {
	tests := []struct {
		name    string
		setup   func()
		want    bool
		wantErr string
	}{
		{
			name: "true",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryCheckOrganizationUnitHasUsersOrGroups, "ou1", testDeploymentID).
					Return([]map[string]interface{}{{"count": int64(1)}}, nil).
					Once()
			},
			want: true,
		},
		{
			name: "false",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryCheckOrganizationUnitHasUsersOrGroups, "ou1", testDeploymentID).
					Return([]map[string]interface{}{{"count": int64(0)}}, nil).
					Once()
			},
			want: false,
		},
		{
			name: "query error",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryCheckOrganizationUnitHasUsersOrGroups, "ou1", testDeploymentID).
					Return(nil, errors.New("query err")).
					Once()
			},
			wantErr: "failed to execute query",
		},
		{
			name: "db client error",
			setup: func() {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()

			hasChildren, err := suite.store.CheckOrganizationUnitHasChildResources("ou1")

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
			suite.Equal(tc.want, hasChildren)
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_CheckOrganizationUnitNameConflict() {
	suite.runConflictQueryScenario(
		queryCheckOrganizationUnitNameConflict,
		queryCheckOrganizationUnitNameConflictRoot,
		"Finance",
		1,
		0,
		func(parent *string) (bool, error) {
			return suite.store.CheckOrganizationUnitNameConflict("Finance", parent)
		},
	)
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_CheckOrganizationUnitHandleConflict() {
	suite.runConflictQueryScenario(
		queryCheckOrganizationUnitHandleConflict,
		queryCheckOrganizationUnitHandleConflictRoot,
		"finance",
		0,
		2,
		func(parent *string) (bool, error) {
			return suite.store.CheckOrganizationUnitHandleConflict("finance", parent)
		},
	)
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitGroupsCount() {
	tests := []struct {
		name    string
		setup   func()
		want    int
		wantErr string
	}{
		{
			name: "success",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitGroupsCount, "ou1", testDeploymentID).
					Return([]map[string]interface{}{{"total": int64(2)}}, nil).
					Once()
			},
			want: 2,
		},
		{
			name: "empty result",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitGroupsCount, "ou1", testDeploymentID).
					Return([]map[string]interface{}{}, nil).
					Once()
			},
			want: 0,
		},
		{
			name: "invalid type",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitGroupsCount, "ou1", testDeploymentID).
					Return([]map[string]interface{}{{"total": "bad"}}, nil).
					Once()
			},
			wantErr: "failed to parse count result",
		},
		{
			name: "query error",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitGroupsCount, "ou1", testDeploymentID).
					Return(nil, errors.New("query err")).
					Once()
			},
			wantErr: "failed to execute count query",
		},
		{
			name: "db client error",
			setup: func() {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()

			count, err := suite.store.GetOrganizationUnitGroupsCount("ou1")

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
			suite.Equal(tc.want, count)
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitGroupsList() {
	tests := []struct {
		name    string
		limit   int
		offset  int
		setup   func(limit, offset int)
		assert  func(groups []Group)
		wantErr string
	}{
		{
			name:   "success",
			limit:  10,
			offset: 0,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitGroupsList, "ou1", limit, offset, testDeploymentID).
					Return([]map[string]interface{}{
						{"group_id": "grp1", "name": "Group 1"},
						{"group_id": "grp2", "name": "Group 2"},
					}, nil).
					Once()
			},
			assert: func(groups []Group) {
				suite.Len(groups, 2)
				suite.Equal("grp1", groups[0].ID)
				suite.Equal("Group 2", groups[1].Name)
			},
		},
		{
			name:   "invalid group id",
			limit:  1,
			offset: 0,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitGroupsList, "ou1", limit, offset, testDeploymentID).
					Return([]map[string]interface{}{
						{"group_id": 123},
					}, nil).
					Once()
			},
			wantErr: "expected group_id to be a string",
		},
		{
			name:   "invalid name",
			limit:  1,
			offset: 0,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitGroupsList, "ou1", limit, offset, testDeploymentID).
					Return([]map[string]interface{}{
						{"group_id": "grp1", "name": 5},
					}, nil).
					Once()
			},
			wantErr: "expected name to be a string",
		},
		{
			name:   "query error",
			limit:  5,
			offset: 5,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitGroupsList, "ou1", limit, offset, testDeploymentID).
					Return(nil, errors.New("query err")).
					Once()
			},
			wantErr: "failed to execute query",
		},
		{
			name:   "db client error",
			limit:  1,
			offset: 0,
			setup: func(limit, offset int) {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup(tc.limit, tc.offset)

			groups, err := suite.store.GetOrganizationUnitGroupsList("ou1", tc.limit, tc.offset)

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
			if tc.assert != nil {
				tc.assert(groups)
			}
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitUsersCount() {
	suite.runCountQueryScenario(
		queryGetOrganizationUnitUsersCount,
		"ou1",
		testDeploymentID,
		4,
		func() (int, error) {
			return suite.store.GetOrganizationUnitUsersCount("ou1")
		},
	)
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitUsersList() {
	tests := []struct {
		name    string
		limit   int
		offset  int
		setup   func(limit, offset int)
		assert  func(users []User)
		wantErr string
	}{
		{
			name:   "success",
			limit:  10,
			offset: 0,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitUsersList, "ou1", limit, offset, testDeploymentID).
					Return([]map[string]interface{}{
						{"user_id": "user1"},
						{"user_id": "user2"},
					}, nil).
					Once()
			},
			assert: func(users []User) {
				suite.Len(users, 2)
				suite.Equal("user1", users[0].ID)
			},
		},
		{
			name:   "invalid type",
			limit:  1,
			offset: 0,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitUsersList, "ou1", limit, offset, testDeploymentID).
					Return([]map[string]interface{}{
						{"user_id": 123},
					}, nil).
					Once()
			},
			wantErr: "expected user_id to be a string",
		},
		{
			name:   "query error",
			limit:  5,
			offset: 5,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitUsersList, "ou1", limit, offset, testDeploymentID).
					Return(nil, errors.New("query err")).
					Once()
			},
			wantErr: "failed to execute query",
		},
		{
			name:   "db client error",
			limit:  1,
			offset: 0,
			setup: func(limit, offset int) {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup(tc.limit, tc.offset)

			users, err := suite.store.GetOrganizationUnitUsersList("ou1", tc.limit, tc.offset)

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
			if tc.assert != nil {
				tc.assert(users)
			}
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitChildrenCount() {
	suite.runCountQueryScenario(
		queryGetOrganizationUnitChildrenCount,
		"root",
		testDeploymentID,
		5,
		func() (int, error) {
			return suite.store.GetOrganizationUnitChildrenCount("root")
		},
	)
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitChildrenList() {
	tests := []struct {
		name    string
		parent  string
		limit   int
		offset  int
		setup   func(parent string, limit, offset int)
		assert  func(children []OrganizationUnitBasic)
		wantErr string
	}{
		{
			name:   "success",
			parent: "root",
			limit:  5,
			offset: 10,
			setup: func(parent string, limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitChildrenList, parent, limit, offset, testDeploymentID).
					Return([]map[string]interface{}{
						makeOUResultRow("child1", "child1", "Child 1", "", &parent),
						makeOUResultRow("child2", "child2", "Child 2", "desc", &parent),
					}, nil).
					Once()
			},
			assert: func(children []OrganizationUnitBasic) {
				suite.Len(children, 2)
				suite.Equal("child1", children[0].ID)
			},
		},
		{
			name:   "query error",
			parent: "root",
			limit:  1,
			offset: 0,
			setup: func(parent string, limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitChildrenList, parent, limit, offset, testDeploymentID).
					Return(nil, errors.New("query err")).
					Once()
			},
			wantErr: "failed to execute query",
		},
		{
			name:   "builder error",
			parent: "root",
			limit:  1,
			offset: 0,
			setup: func(parent string, limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitChildrenList, parent, limit, offset, testDeploymentID).
					Return([]map[string]interface{}{{"ou_id": 1}}, nil).
					Once()
			},
			wantErr: "failed to build organization unit basic",
		},
		{
			name:   "db client error",
			parent: "root",
			limit:  1,
			offset: 0,
			setup: func(parent string, limit, offset int) {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup(tc.parent, tc.limit, tc.offset)

			children, err := suite.store.GetOrganizationUnitChildrenList(tc.parent, tc.limit, tc.offset)

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
			if tc.assert != nil {
				tc.assert(children)
			}
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_UpdateOrganizationUnit() {
	tests := []struct {
		name    string
		ou      OrganizationUnit
		setup   func(ou OrganizationUnit)
		wantErr string
	}{
		{
			name: "success",
			ou: func() OrganizationUnit {
				parent := "parent1"
				return OrganizationUnit{
					ID:          "ou1",
					Parent:      &parent,
					Handle:      "root",
					Name:        "Root",
					Description: "desc",
				}
			}(),
			setup: func(ou OrganizationUnit) {
				suite.expectDBClient()
				suite.dbClientMock.
					On(
						"Execute",
						queryUpdateOrganizationUnit,
						ou.ID,
						ou.Parent,
						ou.Handle,
						ou.Name,
						ou.Description,
						ou.ThemeID,
						ou.LayoutID,
						`{"cookie_policy_uri":"","logo_url":"","policy_uri":"","tos_uri":""}`,
						testDeploymentID,
					).
					Return(int64(1), nil).
					Once()
			},
		},
		{
			name: "success with design fields",
			ou: func() OrganizationUnit {
				parent := "parent1"
				return OrganizationUnit{
					ID:          "ou1",
					Parent:      &parent,
					Handle:      "root",
					Name:        "Root",
					Description: "desc",
					ThemeID:     "theme-123",
					LayoutID:    "layout-456",
					LogoURL:     "https://example.com/logo.png",
				}
			}(),
			setup: func(ou OrganizationUnit) {
				suite.expectDBClient()
				suite.dbClientMock.
					On(
						"Execute",
						queryUpdateOrganizationUnit,
						ou.ID,
						ou.Parent,
						ou.Handle,
						ou.Name,
						ou.Description,
						ou.ThemeID,
						ou.LayoutID,
						`{"cookie_policy_uri":"","logo_url":"https://example.com/logo.png",`+
							`"policy_uri":"","tos_uri":""}`,
						testDeploymentID,
					).
					Return(int64(1), nil).
					Once()
			},
		},
		{
			name: "execute error",
			ou:   OrganizationUnit{ID: "ou1"},
			setup: func(ou OrganizationUnit) {
				suite.expectDBClient()
				suite.dbClientMock.
					On(
						"Execute",
						queryUpdateOrganizationUnit,
						ou.ID,
						ou.Parent,
						ou.Handle,
						ou.Name,
						ou.Description,
						ou.ThemeID,
						ou.LayoutID,
						`{"cookie_policy_uri":"","logo_url":"","policy_uri":"","tos_uri":""}`,
						testDeploymentID,
					).
					Return(int64(0), errors.New("update failed")).
					Once()
			},
			wantErr: "failed to execute query",
		},
		{
			name: "db client error",
			ou:   OrganizationUnit{ID: "ou1"},
			setup: func(ou OrganizationUnit) {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup(tc.ou)

			err := suite.store.UpdateOrganizationUnit(tc.ou)

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_DeleteOrganizationUnit() {
	tests := []struct {
		name    string
		setup   func()
		wantErr string
	}{
		{
			name: "success",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Execute", queryDeleteOrganizationUnit, "ou1", testDeploymentID).
					Return(int64(1), nil).
					Once()
			},
		},
		{
			name: "execute error",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Execute", queryDeleteOrganizationUnit, "ou1", testDeploymentID).
					Return(int64(0), errors.New("delete failed")).
					Once()
			},
			wantErr: "failed to execute query",
		},
		{
			name: "db client error",
			setup: func() {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()

			err := suite.store.DeleteOrganizationUnit("ou1")

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_IsOrganizationUnitExists() {
	tests := []struct {
		name    string
		setup   func()
		want    bool
		wantErr string
	}{
		{
			name: "true",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryCheckOrganizationUnitExists, "ou1", testDeploymentID).
					Return([]map[string]interface{}{{"count": int64(1)}}, nil).
					Once()
			},
			want: true,
		},
		{
			name: "false on empty result",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryCheckOrganizationUnitExists, "ou1", testDeploymentID).
					Return([]map[string]interface{}{}, nil).
					Once()
			},
			want: false,
		},
		{
			name: "false on zero count",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryCheckOrganizationUnitExists, "ou1", testDeploymentID).
					Return([]map[string]interface{}{{"count": int64(0)}}, nil).
					Once()
			},
			want: false,
		},
		{
			name: "invalid type",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryCheckOrganizationUnitExists, "ou1", testDeploymentID).
					Return([]map[string]interface{}{{"count": "bad"}}, nil).
					Once()
			},
			wantErr: "failed to parse existence check result",
		},
		{
			name: "query error",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryCheckOrganizationUnitExists, "ou1", testDeploymentID).
					Return(nil, errors.New("query err")).
					Once()
			},
			wantErr: "failed to execute existence check query",
		},
		{
			name: "db client error",
			setup: func() {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db fail")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()

			exists, err := suite.store.IsOrganizationUnitExists("ou1")

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
			suite.Equal(tc.want, exists)
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitByPath() {
	tests := []struct {
		name          string
		path          []string
		setup         func(path []string)
		assert        func(ou OrganizationUnit)
		wantErr       error
		wantErrString string
		after         func()
	}{
		{
			name: "success",
			path: []string{"root", "child"},
			setup: func(_ []string) {
				rootID := "root-id"
				childID := "child-id"
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitByHandle, "root", testDeploymentID).
					Return([]map[string]interface{}{
						makeOUResultRow(rootID, "root", "Root", "desc", nil),
					}, nil).
					Once()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitByHandle, "child", rootID, testDeploymentID).
					Return([]map[string]interface{}{
						makeOUResultRow(childID, "child", "Child", "", &rootID),
					}, nil).
					Once()
			},
			assert: func(ou OrganizationUnit) {
				suite.Equal("child-id", ou.ID)
				suite.NotNil(ou.Parent)
				suite.Equal("root-id", *ou.Parent)
			},
		},
		{
			name:    "empty path",
			path:    []string{},
			wantErr: ErrOrganizationUnitNotFound,
			after: func() {
				suite.providerMock.AssertNotCalled(suite.T(), "GetUserDBClient", mock.Anything)
			},
		},
		{
			name: "db client error",
			path: []string{"root"},
			setup: func(_ []string) {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErrString: "failed to get database client",
		},
		{
			name: "query error root",
			path: []string{"root"},
			setup: func(_ []string) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitByHandle, "root", testDeploymentID).
					Return(nil, errors.New("query")).
					Once()
			},
			wantErrString: "failed to execute query for handle root",
		},
		{
			name: "not found",
			path: []string{"root"},
			setup: func(_ []string) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitByHandle, "root", testDeploymentID).
					Return([]map[string]interface{}{}, nil).
					Once()
			},
			wantErr: ErrOrganizationUnitNotFound,
		},
		{
			name: "builder error",
			path: []string{"root"},
			setup: func(_ []string) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitByHandle, "root", testDeploymentID).
					Return([]map[string]interface{}{{"ou_id": 1}}, nil).
					Once()
			},
			wantErrString: "failed to build organization unit for handle root",
		},
		{
			name: "child query error",
			path: []string{"root", "child"},
			setup: func(_ []string) {
				rootID := "root"
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitByHandle, "root", testDeploymentID).
					Return([]map[string]interface{}{makeOUResultRow(rootID, "root", "Root", "", nil)}, nil).
					Once()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitByHandle, "child", rootID, testDeploymentID).
					Return(nil, errors.New("child query failed")).
					Once()
			},
			wantErrString: "failed to execute query for handle child",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			if tc.setup != nil {
				tc.setup(tc.path)
			}

			ou, err := suite.store.GetOrganizationUnitByPath(tc.path)

			switch {
			case tc.wantErr != nil:
				suite.Require().ErrorIs(err, tc.wantErr)
			case tc.wantErrString != "":
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErrString)
			default:
				suite.Require().NoError(err)
				if tc.assert != nil {
					tc.assert(ou)
				}
			}

			if tc.after != nil {
				tc.after()
			}
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnit() {
	tests := []struct {
		name          string
		id            string
		setup         func(id string)
		assert        func(ou OrganizationUnit)
		wantErr       error
		wantErrString string
	}{
		{
			name: "success",
			id:   "ou1",
			setup: func(id string) {
				parentID := testParentID
				row := makeOUResultRow(id, "root", "Root", "desc", &parentID)
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitByID, id, testDeploymentID).
					Return([]map[string]interface{}{row}, nil).
					Once()
			},
			assert: func(ou OrganizationUnit) {
				suite.Equal("ou1", ou.ID)
				suite.NotNil(ou.Parent)
				suite.Equal(testParentID, *ou.Parent)
			},
		},
		{
			name: "query error",
			id:   "ou1",
			setup: func(id string) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitByID, id, testDeploymentID).
					Return(nil, errors.New("query err")).
					Once()
			},
			wantErrString: "failed to execute query",
		},
		{
			name: "not found",
			id:   "missing",
			setup: func(id string) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitByID, id, testDeploymentID).
					Return([]map[string]interface{}{}, nil).
					Once()
			},
			wantErr: ErrOrganizationUnitNotFound,
		},
		{
			name: "builder error",
			id:   "ou1",
			setup: func(id string) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetOrganizationUnitByID, id, testDeploymentID).
					Return([]map[string]interface{}{{"ou_id": 2}}, nil).
					Once()
			},
			wantErrString: "failed to build organization unit",
		},
		{
			name: "db client error",
			id:   "ou1",
			setup: func(id string) {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErrString: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup(tc.id)

			ou, err := suite.store.GetOrganizationUnit(tc.id)

			switch {
			case tc.wantErr != nil:
				suite.Require().ErrorIs(err, tc.wantErr)
			case tc.wantErrString != "":
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErrString)
			default:
				suite.Require().NoError(err)
				if tc.assert != nil {
					tc.assert(ou)
				}
			}
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_CreateOrganizationUnit() {
	tests := []struct {
		name    string
		ou      OrganizationUnit
		setup   func(ou OrganizationUnit)
		wantErr string
	}{
		{
			name: "success",
			ou: OrganizationUnit{
				ID:          "ou1",
				Handle:      "root",
				Name:        "Root",
				Description: "desc",
			},
			setup: func(ou OrganizationUnit) {
				suite.expectDBClient()
				suite.dbClientMock.
					On(
						"Execute",
						queryCreateOrganizationUnit,
						ou.ID,
						ou.Parent,
						ou.Handle,
						ou.Name,
						ou.Description,
						ou.ThemeID,
						ou.LayoutID,
						`{"cookie_policy_uri":"","logo_url":"","policy_uri":"","tos_uri":""}`,
						testDeploymentID,
					).
					Return(int64(1), nil).
					Once()
			},
		},
		{
			name: "success with design fields",
			ou: OrganizationUnit{
				ID:          "ou1",
				Handle:      "root",
				Name:        "Root",
				Description: "desc",
				ThemeID:     "theme-123",
				LayoutID:    "layout-456",
				LogoURL:     "https://example.com/logo.png",
			},
			setup: func(ou OrganizationUnit) {
				suite.expectDBClient()
				suite.dbClientMock.
					On(
						"Execute",
						queryCreateOrganizationUnit,
						ou.ID,
						ou.Parent,
						ou.Handle,
						ou.Name,
						ou.Description,
						ou.ThemeID,
						ou.LayoutID,
						`{"cookie_policy_uri":"","logo_url":"https://example.com/logo.png",`+
							`"policy_uri":"","tos_uri":""}`,
						testDeploymentID,
					).
					Return(int64(1), nil).
					Once()
			},
		},
		{
			name: "execute error",
			ou: OrganizationUnit{
				ID:          "ou-err",
				Handle:      "root",
				Name:        "Root",
				Description: "desc",
			},
			setup: func(ou OrganizationUnit) {
				suite.expectDBClient()
				suite.dbClientMock.
					On(
						"Execute",
						queryCreateOrganizationUnit,
						ou.ID,
						ou.Parent,
						ou.Handle,
						ou.Name,
						ou.Description,
						ou.ThemeID,
						ou.LayoutID,
						`{"cookie_policy_uri":"","logo_url":"","policy_uri":"","tos_uri":""}`,
						testDeploymentID,
					).
					Return(int64(0), errors.New("insert failed")).
					Once()
			},
			wantErr: "failed to execute query",
		},
		{
			name: "db client error",
			ou:   OrganizationUnit{ID: "ou1"},
			setup: func(ou OrganizationUnit) {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db init failed")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			if tc.setup != nil {
				tc.setup(tc.ou)
			}

			err := suite.store.CreateOrganizationUnit(tc.ou)

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitList() {
	tests := []struct {
		name          string
		limit         int
		offset        int
		setup         func(limit, offset int)
		assert        func(ous []OrganizationUnitBasic)
		wantErrString string
	}{
		{
			name:   "success",
			limit:  2,
			offset: 0,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				rows := []map[string]interface{}{
					makeOUResultRow("root", "root", "Root", "desc", nil),
					makeOUResultRow("child", "child", "Child", "", nil),
				}
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitList, limit, offset, testDeploymentID).
					Return(rows, nil).
					Once()
			},
			assert: func(ous []OrganizationUnitBasic) {
				suite.Len(ous, 2)
				suite.Equal("root", ous[0].ID)
				suite.Equal("child", ous[1].Handle)
			},
		},
		{
			name:   "query error",
			limit:  10,
			offset: 5,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitList, limit, offset, testDeploymentID).
					Return(nil, errors.New("query error")).
					Once()
			},
			wantErrString: "failed to execute query",
		},
		{
			name:   "builder error",
			limit:  1,
			offset: 0,
			setup: func(limit, offset int) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitList, limit, offset, testDeploymentID).
					Return([]map[string]interface{}{{"ou_id": 123}}, nil).
					Once()
			},
			wantErrString: "failed to build organization unit basic",
		},
		{
			name:   "db client error",
			limit:  1,
			offset: 0,
			setup: func(limit, offset int) {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErrString: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup(tc.limit, tc.offset)

			ous, err := suite.store.GetOrganizationUnitList(tc.limit, tc.offset)

			if tc.wantErrString != "" {
				suite.Require().Error(err)
				suite.Nil(ous)
				suite.Contains(err.Error(), tc.wantErrString)
				return
			}

			suite.Require().NoError(err)
			if tc.assert != nil {
				tc.assert(ous)
			}
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitListCount() {
	tests := []struct {
		name    string
		setup   func()
		want    int
		wantErr string
	}{
		{
			name: "success",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitListCount, testDeploymentID).
					Return([]map[string]interface{}{{"total": int64(3)}}, nil).
					Once()
			},
			want: 3,
		},
		{
			name: "query error",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitListCount, testDeploymentID).
					Return(nil, errors.New("boom")).
					Once()
			},
			wantErr: "failed to execute count query",
		},
		{
			name: "unexpected type",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", queryGetRootOrganizationUnitListCount, testDeploymentID).
					Return([]map[string]interface{}{{"total": "3"}}, nil).
					Once()
			},
			wantErr: "unexpected type for total",
		},
		{
			name: "db client error",
			setup: func() {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("no db")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()

			count, err := suite.store.GetOrganizationUnitListCount()

			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Zero(count)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.Require().NoError(err)
			suite.Equal(tc.want, count)
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_GetOrganizationUnitsByIDs() {
	tests := []struct {
		name          string
		ids           []string
		setup         func(ids []string)
		assert        func(ous []OrganizationUnitBasic)
		wantErrString string
	}{
		{
			name: "success",
			ids:  []string{"ou1", "ou2"},
			setup: func(ids []string) {
				suite.expectDBClient()
				rows := []map[string]interface{}{
					makeOUResultRow("ou1", "root1", "Root1", "desc1", nil),
					makeOUResultRow("ou2", "root2", "Root2", "desc2", nil),
				}
				suite.dbClientMock.
					On("Query", mock.AnythingOfType("model.DBQuery"), mock.Anything, mock.Anything, mock.Anything).
					Return(rows, nil).
					Once()
			},
			assert: func(ous []OrganizationUnitBasic) {
				suite.Len(ous, 2)
				suite.Equal("ou1", ous[0].ID)
				suite.Equal("ou2", ous[1].ID)
			},
		},
		{
			name: "empty ids",
			ids:  []string{},
			assert: func(ous []OrganizationUnitBasic) {
				suite.Len(ous, 0)
			},
		},
		{
			name: "query error",
			ids:  []string{"ou1"},
			setup: func(ids []string) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", mock.AnythingOfType("model.DBQuery"), mock.Anything, mock.Anything).
					Return(nil, errors.New("query error")).
					Once()
			},
			wantErrString: "failed to execute query",
		},
		{
			name: "builder error",
			ids:  []string{"ou1"},
			setup: func(ids []string) {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Query", mock.AnythingOfType("model.DBQuery"), mock.Anything, mock.Anything).
					Return([]map[string]interface{}{{"ou_id": 123}}, nil).
					Once()
			},
			wantErrString: "failed to build organization unit basic",
		},
		{
			name: "db client error",
			ids:  []string{"ou1"},
			setup: func(ids []string) {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErrString: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			if tc.setup != nil {
				tc.setup(tc.ids)
			}

			ous, err := suite.store.GetOrganizationUnitsByIDs(tc.ids)

			if tc.wantErrString != "" {
				suite.Require().Error(err)
				suite.Nil(ous)
				suite.Contains(err.Error(), tc.wantErrString)
				return
			}

			suite.Require().NoError(err)
			if tc.assert != nil {
				tc.assert(ous)
			}
		})
	}
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_IsOrganizationUnitDeclarative() {
	suite.Run("returns false", func() {
		suite.SetupTest()
		res := suite.store.IsOrganizationUnitDeclarative("ou1")
		suite.Require().False(res)
	})
}

func (suite *OrganizationUnitStoreTestSuite) TestOUStore_buildGetOrganizationUnitsByIDsQuery() {
	suite.Run("builds query with correct placeholders", func() {
		ids := []string{"id1", "id2", "id3"}
		query := buildGetOrganizationUnitsByIDsQuery(ids)

		suite.Require().Equal("OUQ-OU_MGT-21", query.ID)
		suite.Require().Contains(query.PostgresQuery, "$1, $2, $3")
		suite.Require().Contains(query.PostgresQuery, "DEPLOYMENT_ID = $4")
		suite.Require().Contains(query.SQLiteQuery, "?, ?, ?")
		suite.Require().Contains(query.SQLiteQuery, "DEPLOYMENT_ID = ?")
	})
}
