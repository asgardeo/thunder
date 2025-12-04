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

	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

type OrganizationUnitServiceTestSuite struct {
	suite.Suite
}

const testParentID = "parent"

func TestOUService_OrganizationUnitServiceTestSuite_Run(t *testing.T) {
	suite.Run(t, new(OrganizationUnitServiceTestSuite))
}

func (suite *OrganizationUnitServiceTestSuite) SetupTest() {
	// Initialize ThunderRuntime with immutable mode disabled by default
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		ImmutableResources: config.ImmutableResources{
			Enabled: false,
		},
	}
	_ = config.InitializeThunderRuntime("/tmp/test", testConfig)
}

func (suite *OrganizationUnitServiceTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

type ouListExpectations struct {
	totalResults int
	count        int
	startIndex   int
	handles      []string
	linkRels     []string
	linkHrefs    []string
}

type pathListTestConfig[Resp any] struct {
	invalidPath    string
	validPath      string
	validPathSlice []string
	limit          int
	offset         int
	setupSuccess   func(*organizationUnitStoreInterfaceMock)
	assertSuccess  func(Resp)
	invoke         func(*organizationUnitService, string, int, int) (Resp, *serviceerror.ServiceError)
}

type pathListInvoker[Resp any] func(*organizationUnitService, string, int, int) (Resp, *serviceerror.ServiceError)

// runOUPathListTests de-duplicates the repeated path-based list scenarios across children/users/groups.
func runOUPathListTests[Resp any](suite *OrganizationUnitServiceTestSuite, cfg pathListTestConfig[Resp]) {
	suite.Run("invalid path", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		service := suite.newService(store)

		resp, err := cfg.invoke(service, cfg.invalidPath, cfg.limit, cfg.offset)

		suite.Require().Nil(resp)
		suite.Require().Equal(ErrorInvalidHandlePath, *err)
		store.AssertNotCalled(suite.T(), "GetOrganizationUnitByPath", mock.Anything)
	})

	suite.Run("not found", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnitByPath", cfg.validPathSlice).
			Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
			Once()

		service := suite.newService(store)
		resp, err := cfg.invoke(service, cfg.validPath, cfg.limit, cfg.offset)

		suite.Require().Nil(resp)
		suite.Require().Equal(ErrorOrganizationUnitNotFound, *err)
	})

	suite.Run("store error", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnitByPath", cfg.validPathSlice).
			Return(OrganizationUnit{}, errors.New("boom")).
			Once()

		service := suite.newService(store)
		resp, err := cfg.invoke(service, cfg.validPath, cfg.limit, cfg.offset)

		suite.Require().Nil(resp)
		suite.Require().Equal(ErrorInternalServerError, *err)
	})

	suite.Run("success", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		if cfg.setupSuccess != nil {
			cfg.setupSuccess(store)
		}

		service := suite.newService(store)
		resp, err := cfg.invoke(service, cfg.validPath, cfg.limit, cfg.offset)

		suite.Require().Nil(err)
		if cfg.assertSuccess != nil {
			cfg.assertSuccess(resp)
		} else {
			suite.Require().NotNil(resp)
		}
	})
}

func setupDefaultPathSuccess(
	store *organizationUnitStoreInterfaceMock,
	limit, offset int,
	listMethod string,
	listReturn interface{},
	countMethod string,
	countReturn interface{},
) {
	store.On("GetOrganizationUnitByPath", []string{"root"}).
		Return(OrganizationUnit{ID: "ou-1"}, nil).
		Once()
	store.On("IsOrganizationUnitExists", "ou-1").
		Return(true, nil).
		Once()
	store.On(listMethod, "ou-1", limit, offset).
		Return(listReturn, nil).
		Once()
	store.On(countMethod, "ou-1").
		Return(countReturn, nil).
		Once()
}

func newDefaultPathListConfig[Resp any](
	invalidPath string,
	limit, offset int,
	listMethod string,
	listReturn interface{},
	countMethod string,
	countReturn interface{},
	assert func(Resp),
	invoker pathListInvoker[Resp],
) pathListTestConfig[Resp] {
	return pathListTestConfig[Resp]{
		invalidPath:    invalidPath,
		validPath:      "root",
		validPathSlice: []string{"root"},
		limit:          limit,
		offset:         offset,
		setupSuccess: func(store *organizationUnitStoreInterfaceMock) {
			setupDefaultPathSuccess(store, limit, offset, listMethod, listReturn, countMethod, countReturn)
		},
		assertSuccess: assert,
		invoke:        invoker,
	}
}

func invokeChildrenByPath(
	service *organizationUnitService,
	path string,
	limit, offset int,
) (*OrganizationUnitListResponse, *serviceerror.ServiceError) {
	return service.GetOrganizationUnitChildrenByPath(path, limit, offset)
}

func invokeUsersByPath(
	service *organizationUnitService,
	path string,
	limit, offset int,
) (*UserListResponse, *serviceerror.ServiceError) {
	return service.GetOrganizationUnitUsersByPath(path, limit, offset)
}

func invokeGroupsByPath(
	service *organizationUnitService,
	path string,
	limit, offset int,
) (*GroupListResponse, *serviceerror.ServiceError) {
	return service.GetOrganizationUnitGroupsByPath(path, limit, offset)
}

func (suite *OrganizationUnitServiceTestSuite) newService(
	store *organizationUnitStoreInterfaceMock,
) *organizationUnitService {
	return &organizationUnitService{ouStore: store}
}

func (suite *OrganizationUnitServiceTestSuite) assertOUListResponse(
	resp *OrganizationUnitListResponse,
	expected *ouListExpectations,
) {
	suite.Require().NotNil(resp)
	suite.Require().Equal(expected.totalResults, resp.TotalResults)
	suite.Require().Equal(expected.count, resp.Count)
	suite.Require().Equal(expected.startIndex, resp.StartIndex)
	suite.Require().Len(resp.OrganizationUnits, len(expected.handles))
	for idx, handle := range expected.handles {
		suite.Require().Equal(handle, resp.OrganizationUnits[idx].Handle)
	}
	suite.Require().Len(resp.Links, len(expected.linkRels))
	for idx := range expected.linkRels {
		suite.Require().Equal(expected.linkRels[idx], resp.Links[idx].Rel)
		suite.Require().Equal(expected.linkHrefs[idx], resp.Links[idx].Href)
	}
}

func strPtr(value string) *string {
	return &value
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_GetOrganizationUnitList() {
	testCases := []struct {
		name       string
		limit      int
		offset     int
		setup      func(*organizationUnitStoreInterfaceMock)
		wantErr    *serviceerror.ServiceError
		wantResult *ouListExpectations
	}{
		{
			name:   "success",
			limit:  2,
			offset: 1,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnitListCount").
					Return(3, nil).
					Once()
				store.On("GetOrganizationUnitList", 2, 1).
					Return([]OrganizationUnitBasic{
						{ID: "ou-1", Handle: "root", Name: "Root"},
						{ID: "ou-2", Handle: "child", Name: "Child"},
					}, nil).
					Once()
			},
			wantResult: &ouListExpectations{
				totalResults: 3,
				count:        2,
				startIndex:   2,
				handles:      []string{"root", "child"},
				linkRels:     []string{"first", "prev", "last"},
				linkHrefs: []string{
					"/organization-units?offset=0&limit=2",
					"/organization-units?offset=0&limit=2",
					"/organization-units?offset=2&limit=2",
				},
			},
		},
		{
			name:    "invalid pagination",
			limit:   0,
			offset:  0,
			wantErr: &ErrorInvalidLimit,
		},
		{
			name:   "count failure",
			limit:  5,
			offset: 0,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnitListCount").
					Return(0, errors.New("count failed")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:   "list failure",
			limit:  5,
			offset: 0,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnitListCount").
					Return(10, nil).
					Once()
				store.On("GetOrganizationUnitList", 5, 0).
					Return(nil, errors.New("list failed")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			if tc.setup != nil {
				tc.setup(store)
			}

			service := suite.newService(store)
			resp, err := service.GetOrganizationUnitList(tc.limit, tc.offset)

			if tc.wantErr != nil {
				suite.Require().Nil(resp)
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				suite.assertOUListResponse(resp, tc.wantResult)
			}

			if tc.wantErr == &ErrorInvalidLimit {
				store.AssertNotCalled(suite.T(), "GetOrganizationUnitListCount")
			}
			store.AssertExpectations(suite.T())
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_CreateOrganizationUnit() {
	parentID := "parent-1"
	validRequest := OrganizationUnitRequest{
		Handle:      "finance",
		Name:        "Finance",
		Description: "desc",
	}

	testCases := []struct {
		name    string
		request OrganizationUnitRequest
		setup   func(*organizationUnitStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
	}{
		{
			name:    "invalid name",
			request: OrganizationUnitRequest{Handle: "handle", Name: "  "},
			wantErr: &ErrorInvalidRequestFormat,
		},
		{
			name:    "invalid handle",
			request: OrganizationUnitRequest{Handle: " ", Name: "Finance"},
			wantErr: &ErrorInvalidRequestFormat,
		},
		{
			name: "parent existence check error",
			request: OrganizationUnitRequest{
				Handle: "finance",
				Name:   "Finance",
				Parent: &parentID,
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", parentID).
					Return(false, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "parent not found",
			request: OrganizationUnitRequest{
				Handle: "finance",
				Name:   "Finance",
				Parent: &parentID,
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", parentID).
					Return(false, nil).
					Once()
			},
			wantErr: &ErrorParentOrganizationUnitNotFound,
		},
		{
			name:    "name conflict error",
			request: validRequest,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("CheckOrganizationUnitNameConflict", "Finance", (*string)(nil)).
					Return(true, nil).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNameConflict,
		},
		{
			name:    "name conflict check failure",
			request: validRequest,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("CheckOrganizationUnitNameConflict", "Finance", (*string)(nil)).
					Return(false, errors.New("name check failed")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:    "handle conflict",
			request: validRequest,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("CheckOrganizationUnitNameConflict", "Finance", (*string)(nil)).
					Return(false, nil).
					Once()
				store.On("CheckOrganizationUnitHandleConflict", "finance", (*string)(nil)).
					Return(true, nil).
					Once()
			},
			wantErr: &ErrorOrganizationUnitHandleConflict,
		},
		{
			name:    "handle conflict check failure",
			request: validRequest,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("CheckOrganizationUnitNameConflict", "Finance", (*string)(nil)).
					Return(false, nil).
					Once()
				store.On("CheckOrganizationUnitHandleConflict", "finance", (*string)(nil)).
					Return(false, errors.New("handle check failed")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:    "create failure",
			request: validRequest,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("CheckOrganizationUnitNameConflict", "Finance", (*string)(nil)).
					Return(false, nil).
					Once()
				store.On("CheckOrganizationUnitHandleConflict", "finance", (*string)(nil)).
					Return(false, nil).
					Once()
				store.On("CreateOrganizationUnit", mock.AnythingOfType("ou.OrganizationUnit")).
					Return(errors.New("insert failed")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:    "success",
			request: validRequest,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("CheckOrganizationUnitNameConflict", "Finance", (*string)(nil)).
					Return(false, nil).
					Once()
				store.On("CheckOrganizationUnitHandleConflict", "finance", (*string)(nil)).
					Return(false, nil).
					Once()
				store.On("CreateOrganizationUnit", mock.MatchedBy(func(ou OrganizationUnit) bool {
					return ou.Name == "Finance" && ou.Handle == "finance"
				})).
					Return(nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			if tc.setup != nil {
				tc.setup(store)
			}

			service := suite.newService(store)
			result, err := service.CreateOrganizationUnit(tc.request)

			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				suite.Require().Equal(tc.request.Name, result.Name)
				suite.Require().Equal(tc.request.Handle, result.Handle)
				suite.Require().NotEmpty(result.ID)
			}

			if tc.wantErr == &ErrorInvalidRequestFormat {
				store.AssertNotCalled(suite.T(), "CheckOrganizationUnitNameConflict", mock.Anything, mock.Anything)
			}
			store.AssertExpectations(suite.T())
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_GetOrganizationUnit() {
	testCases := []struct {
		name    string
		setup   func(*organizationUnitStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
	}{
		{
			name: "success",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnit", "ou-1").
					Return(OrganizationUnit{ID: "ou-1", Name: "Root"}, nil).
					Once()
			},
		},
		{
			name: "not found",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnit", "ou-1").
					Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name: "store error",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnit", "ou-1").
					Return(OrganizationUnit{}, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			tc.setup(store)

			service := suite.newService(store)
			result, err := service.GetOrganizationUnit("ou-1")

			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				suite.Require().Equal("ou-1", result.ID)
			}
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_GetOrganizationUnitByPath() {
	testCases := []struct {
		name    string
		path    string
		setup   func(*organizationUnitStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
	}{
		{
			name:    "invalid path",
			path:    "   ",
			wantErr: &ErrorInvalidHandlePath,
		},
		{
			name: "not found",
			path: "/root/child/",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.
					On("GetOrganizationUnitByPath", []string{"root", "child"}).
					Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name: "store error",
			path: "root",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnitByPath", []string{"root"}).
					Return(OrganizationUnit{}, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "success",
			path: "root",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnitByPath", []string{"root"}).
					Return(OrganizationUnit{ID: "ou-1", Handle: "root"}, nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			if tc.setup != nil {
				tc.setup(store)
			}

			service := suite.newService(store)
			result, err := service.GetOrganizationUnitByPath(tc.path)

			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				suite.Require().Equal("ou-1", result.ID)
			}

			if tc.wantErr == &ErrorInvalidHandlePath {
				store.AssertNotCalled(suite.T(), "GetOrganizationUnitByPath", mock.Anything)
			}
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_IsOrganizationUnitExists() {
	testCases := []struct {
		name    string
		setup   func(*organizationUnitStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
		want    bool
	}{
		{
			name: "success",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
			},
			want: true,
		},
		{
			name: "store error",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(false, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			tc.setup(store)

			service := suite.newService(store)
			result, err := service.IsOrganizationUnitExists("ou-1")

			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				suite.Require().Equal(tc.want, result)
			}
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_IsParent() {
	parentID := "parent-1"
	childID := "child-1"

	suite.Run("returns true when IDs are equal", func() {
		service := suite.newService(newOrganizationUnitStoreInterfaceMock(suite.T()))

		result, err := service.IsParent(parentID, parentID)

		suite.Require().True(result)
		suite.Require().Nil(err)
	})

	suite.Run("returns true for direct parent", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnit", childID).
			Return(OrganizationUnit{ID: childID, Parent: strPtr(parentID)}, nil).
			Once()

		service := suite.newService(store)

		result, err := service.IsParent(parentID, childID)

		suite.Require().True(result)
		suite.Require().Nil(err)
	})

	suite.Run("returns true for ancestor", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnit", childID).
			Return(OrganizationUnit{ID: childID, Parent: strPtr("mid-1")}, nil).
			Once()
		store.On("GetOrganizationUnit", "mid-1").
			Return(OrganizationUnit{ID: "mid-1", Parent: strPtr(parentID)}, nil).
			Once()

		service := suite.newService(store)

		result, err := service.IsParent(parentID, childID)

		suite.Require().True(result)
		suite.Require().Nil(err)
	})

	suite.Run("returns false when parent not in hierarchy", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnit", childID).
			Return(OrganizationUnit{ID: childID, Parent: strPtr("mid-1")}, nil).
			Once()
		store.On("GetOrganizationUnit", "mid-1").
			Return(OrganizationUnit{ID: "mid-1"}, nil).
			Once()

		service := suite.newService(store)

		result, err := service.IsParent(parentID, childID)

		suite.Require().False(result)
		suite.Require().Nil(err)
	})

	suite.Run("returns error when child not found", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnit", childID).
			Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
			Once()

		service := suite.newService(store)

		result, err := service.IsParent(parentID, childID)

		suite.Require().False(result)
		suite.Require().Equal(ErrorOrganizationUnitNotFound, *err)
	})

	suite.Run("returns error on store failure", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnit", childID).
			Return(OrganizationUnit{}, errors.New("boom")).
			Once()

		service := suite.newService(store)

		result, err := service.IsParent(parentID, childID)

		suite.Require().False(result)
		suite.Require().Equal(ErrorInternalServerError, *err)
	})
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_UpdateOrganizationUnit() {
	parentID := testParentID
	tests := []struct {
		name    string
		id      string
		request OrganizationUnitRequest
		setup   func(*organizationUnitStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
		assert  func(OrganizationUnit)
	}{
		{
			name: "success",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle:      "root",
				Name:        "Root",
				Description: "updated",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{
					ID:          "ou-1",
					Handle:      "root",
					Name:        "Root",
					Description: "old",
				}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("UpdateOrganizationUnit", OrganizationUnit{
					ID:          "ou-1",
					Handle:      "root",
					Name:        "Root",
					Description: "updated",
					Parent:      nil,
				}).
					Return(nil).
					Once()
			},
			assert: func(ou OrganizationUnit) {
				suite.Equal("updated", ou.Description)
			},
		},
		{
			name: "not found on fetch",
			id:   "missing",
			request: OrganizationUnitRequest{
				Handle: "root",
				Name:   "Root",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnit", "missing").
					Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name: "fetch failure",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "root",
				Name:   "Root",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("GetOrganizationUnit", "ou-1").
					Return(OrganizationUnit{}, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "invalid handle",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: " ",
				Name:   "Root",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
			},
			wantErr: &ErrorInvalidRequestFormat,
		},
		{
			name: "parent existence check failure",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "root",
				Name:   "Root",
				Parent: &parentID,
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("IsOrganizationUnitExists", parentID).
					Return(false, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "parent not found",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "root",
				Name:   "Root",
				Parent: &parentID,
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("IsOrganizationUnitExists", parentID).
					Return(false, nil).
					Once()
			},
			wantErr: &ErrorParentOrganizationUnitNotFound,
		},
		{
			name: "circular dependency",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "root",
				Name:   "Root",
				Parent: strPtr("ou-1"),
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
			},
			wantErr: &ErrorCircularDependency,
		},
		{
			name: "name conflict",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "root",
				Name:   "Finance",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("CheckOrganizationUnitNameConflict", "Finance", (*string)(nil)).
					Return(true, nil).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNameConflict,
		},
		{
			name: "name conflict check failure",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "root",
				Name:   "Finance",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("CheckOrganizationUnitNameConflict", "Finance", (*string)(nil)).
					Return(false, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "handle conflict",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "finance",
				Name:   "Root",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("CheckOrganizationUnitHandleConflict", "finance", (*string)(nil)).
					Return(true, nil).
					Once()
			},
			wantErr: &ErrorOrganizationUnitHandleConflict,
		},
		{
			name: "handle conflict check failure",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "finance",
				Name:   "Root",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("CheckOrganizationUnitHandleConflict", "finance", (*string)(nil)).
					Return(false, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "update returns not found",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "root",
				Name:   "Root",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("UpdateOrganizationUnit", mock.AnythingOfType("ou.OrganizationUnit")).
					Return(ErrOrganizationUnitNotFound).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name: "update failure",
			id:   "ou-1",
			request: OrganizationUnitRequest{
				Handle: "root",
				Name:   "Root",
			},
			setup: func(store *organizationUnitStoreInterfaceMock) {
				existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
				store.On("GetOrganizationUnit", "ou-1").
					Return(existing, nil).
					Once()
				store.On("UpdateOrganizationUnit", mock.AnythingOfType("ou.OrganizationUnit")).
					Return(errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			if tc.setup != nil {
				tc.setup(store)
			}

			service := suite.newService(store)
			result, err := service.UpdateOrganizationUnit(tc.id, tc.request)

			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				if tc.assert != nil {
					tc.assert(result)
				}
			}

			store.AssertExpectations(suite.T())
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_UpdateOrganizationUnitByPath() {
	request := OrganizationUnitRequest{Handle: "root", Name: "Root"}

	suite.Run("invalid path", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		service := suite.newService(store)

		_, err := service.UpdateOrganizationUnitByPath("   ", request)

		suite.Require().Equal(ErrorInvalidHandlePath, *err)
		store.AssertNotCalled(suite.T(), "GetOrganizationUnitByPath", mock.Anything)
	})

	suite.Run("not found", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnitByPath", []string{"root"}).
			Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
			Once()

		service := suite.newService(store)
		_, err := service.UpdateOrganizationUnitByPath("root", request)

		suite.Require().Equal(ErrorOrganizationUnitNotFound, *err)
	})

	suite.Run("get by path error", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnitByPath", []string{"root"}).
			Return(OrganizationUnit{}, errors.New("boom")).
			Once()

		service := suite.newService(store)
		_, err := service.UpdateOrganizationUnitByPath("root", request)

		suite.Require().Equal(ErrorInternalServerError, *err)
	})

	suite.Run("success", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		existing := OrganizationUnit{ID: "ou-1", Handle: "root", Name: "Root"}
		store.On("GetOrganizationUnitByPath", []string{"root"}).
			Return(existing, nil).
			Once()
		store.On("UpdateOrganizationUnit", mock.AnythingOfType("ou.OrganizationUnit")).
			Return(nil).
			Once()

		service := suite.newService(store)
		result, err := service.UpdateOrganizationUnitByPath("root", request)

		suite.Require().Nil(err)
		suite.Require().Equal("ou-1", result.ID)
	})
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_DeleteOrganizationUnit() {
	testCases := []struct {
		name    string
		setup   func(*organizationUnitStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
	}{
		{
			name: "existence check error",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(false, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "not found",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(false, nil).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name: "has children",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("CheckOrganizationUnitHasChildResources", "ou-1").
					Return(true, nil).
					Once()
			},
			wantErr: &ErrorCannotDeleteOrganizationUnit,
		},
		{
			name: "child check failure",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("CheckOrganizationUnitHasChildResources", "ou-1").
					Return(false, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "delete failure",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("CheckOrganizationUnitHasChildResources", "ou-1").
					Return(false, nil).
					Once()
				store.On("DeleteOrganizationUnit", "ou-1").
					Return(errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "delete not found",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("CheckOrganizationUnitHasChildResources", "ou-1").
					Return(false, nil).
					Once()
				store.On("DeleteOrganizationUnit", "ou-1").
					Return(ErrOrganizationUnitNotFound).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name: "success",
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("CheckOrganizationUnitHasChildResources", "ou-1").
					Return(false, nil).
					Once()
				store.On("DeleteOrganizationUnit", "ou-1").
					Return(nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			tc.setup(store)

			service := suite.newService(store)
			err := service.DeleteOrganizationUnit("ou-1")

			if tc.wantErr != nil {
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
			}
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_DeleteOrganizationUnitByPath() {
	suite.Run("invalid path", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		service := suite.newService(store)

		err := service.DeleteOrganizationUnitByPath("  ")

		suite.Require().Equal(ErrorInvalidHandlePath, *err)
		store.AssertNotCalled(suite.T(), "GetOrganizationUnitByPath", mock.Anything)
	})

	suite.Run("not found", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnitByPath", []string{"root"}).
			Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
			Once()

		service := suite.newService(store)
		err := service.DeleteOrganizationUnitByPath("root")

		suite.Require().Equal(ErrorOrganizationUnitNotFound, *err)
	})

	suite.Run("get by path error", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnitByPath", []string{"root"}).
			Return(OrganizationUnit{}, errors.New("boom")).
			Once()

		service := suite.newService(store)
		err := service.DeleteOrganizationUnitByPath("root")

		suite.Require().Equal(ErrorInternalServerError, *err)
	})

	suite.Run("cannot delete", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnitByPath", []string{"root"}).
			Return(OrganizationUnit{ID: "ou-1"}, nil).
			Once()
		store.On("CheckOrganizationUnitHasChildResources", "ou-1").
			Return(true, nil).
			Once()

		service := suite.newService(store)
		err := service.DeleteOrganizationUnitByPath("root")

		suite.Require().Equal(ErrorCannotDeleteOrganizationUnit, *err)
	})

	suite.Run("success", func() {
		store := newOrganizationUnitStoreInterfaceMock(suite.T())
		store.On("GetOrganizationUnitByPath", []string{"root"}).
			Return(OrganizationUnit{ID: "ou-1"}, nil).
			Once()
		store.On("CheckOrganizationUnitHasChildResources", "ou-1").
			Return(false, nil).
			Once()
		store.On("DeleteOrganizationUnit", "ou-1").
			Return(nil).
			Once()

		service := suite.newService(store)
		err := service.DeleteOrganizationUnitByPath("root")

		suite.Require().Nil(err)
	})
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_GetOrganizationUnitChildren() {
	testCases := []struct {
		name       string
		limit      int
		offset     int
		setup      func(*organizationUnitStoreInterfaceMock)
		wantErr    *serviceerror.ServiceError
		wantResult *ouListExpectations
	}{
		{
			name:    "invalid pagination",
			limit:   0,
			offset:  0,
			wantErr: &ErrorInvalidLimit,
		},
		{
			name:  "ou not found",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(false, nil).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name:  "existence check error",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(false, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:  "list failure",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("GetOrganizationUnitChildrenList", "ou-1", 5, 0).
					Return(nil, errors.New("list fail")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:  "count failure",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("GetOrganizationUnitChildrenList", "ou-1", 5, 0).
					Return([]OrganizationUnitBasic{}, nil).
					Once()
				store.On("GetOrganizationUnitChildrenCount", "ou-1").
					Return(0, errors.New("count fail")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:  "success",
			limit: 2,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("GetOrganizationUnitChildrenList", "ou-1", 2, 0).
					Return([]OrganizationUnitBasic{
						{ID: "child-1", Handle: "finance", Name: "Finance"},
					}, nil).
					Once()
				store.On("GetOrganizationUnitChildrenCount", "ou-1").
					Return(1, nil).
					Once()
			},
			wantResult: &ouListExpectations{
				totalResults: 1,
				count:        1,
				startIndex:   1,
				handles:      []string{"finance"},
				linkRels:     []string{},
				linkHrefs:    []string{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			if tc.setup != nil {
				tc.setup(store)
			}

			service := suite.newService(store)
			resp, err := service.GetOrganizationUnitChildren("ou-1", tc.limit, tc.offset)

			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				suite.assertOUListResponse(resp, tc.wantResult)
			}

			if tc.wantErr == &ErrorInvalidLimit {
				store.AssertNotCalled(suite.T(), "IsOrganizationUnitExists", mock.Anything)
			}
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_GetOrganizationUnitChildrenByPath() {
	config := newDefaultPathListConfig[*OrganizationUnitListResponse](
		" ",
		5,
		0,
		"GetOrganizationUnitChildrenList",
		[]OrganizationUnitBasic{},
		"GetOrganizationUnitChildrenCount",
		0,
		func(resp *OrganizationUnitListResponse) {
			suite.Require().NotNil(resp)
		},
		invokeChildrenByPath,
	)
	runOUPathListTests(suite, config)
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_GetOrganizationUnitUsers() {
	testCases := []struct {
		name    string
		limit   int
		offset  int
		setup   func(*organizationUnitStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
	}{
		{
			name:    "invalid pagination",
			limit:   0,
			offset:  0,
			wantErr: &ErrorInvalidLimit,
		},
		{
			name:  "not found",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(false, nil).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name:  "existence error",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(false, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:  "list error",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("GetOrganizationUnitUsersList", "ou-1", 5, 0).
					Return(nil, errors.New("list")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:  "count error",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("GetOrganizationUnitUsersList", "ou-1", 5, 0).
					Return([]User{{ID: "user-1"}}, nil).
					Once()
				store.On("GetOrganizationUnitUsersCount", "ou-1").
					Return(0, errors.New("count")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:  "success",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("GetOrganizationUnitUsersList", "ou-1", 5, 0).
					Return([]User{{ID: "user-1"}, {ID: "user-2"}}, nil).
					Once()
				store.On("GetOrganizationUnitUsersCount", "ou-1").
					Return(2, nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			if tc.setup != nil {
				tc.setup(store)
			}

			service := suite.newService(store)
			resp, err := service.GetOrganizationUnitUsers("ou-1", tc.limit, tc.offset)

			if tc.wantErr != nil {
				suite.Require().Nil(resp)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				suite.Require().Equal(2, resp.TotalResults)
				suite.Require().Len(resp.Users, 2)
			}

			if tc.wantErr == &ErrorInvalidLimit {
				store.AssertNotCalled(suite.T(), "IsOrganizationUnitExists", mock.Anything)
			}
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_GetOrganizationUnitUsersByPath() {
	config := newDefaultPathListConfig[*UserListResponse](
		"   ",
		5,
		0,
		"GetOrganizationUnitUsersList",
		[]User{},
		"GetOrganizationUnitUsersCount",
		0,
		func(resp *UserListResponse) {
			suite.Require().NotNil(resp)
		},
		invokeUsersByPath,
	)
	runOUPathListTests(suite, config)
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_GetOrganizationUnitGroups() {
	testCases := []struct {
		name    string
		limit   int
		offset  int
		setup   func(*organizationUnitStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
		assert  func(*GroupListResponse)
	}{
		{
			name:    "invalid pagination",
			limit:   0,
			wantErr: &ErrorInvalidLimit,
		},
		{
			name:  "list error",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("GetOrganizationUnitGroupsList", "ou-1", 5, 0).
					Return(nil, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:  "success",
			limit: 5,
			setup: func(store *organizationUnitStoreInterfaceMock) {
				store.On("IsOrganizationUnitExists", "ou-1").
					Return(true, nil).
					Once()
				store.On("GetOrganizationUnitGroupsList", "ou-1", 5, 0).
					Return([]Group{{ID: "g1"}}, nil).
					Once()
				store.On("GetOrganizationUnitGroupsCount", "ou-1").
					Return(1, nil).
					Once()
			},
			assert: func(resp *GroupListResponse) {
				suite.Require().Equal(1, resp.TotalResults)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			store := newOrganizationUnitStoreInterfaceMock(suite.T())
			if tc.setup != nil {
				tc.setup(store)
			}

			service := suite.newService(store)
			resp, err := service.GetOrganizationUnitGroups("ou-1", tc.limit, tc.offset)

			if tc.wantErr != nil {
				suite.Require().Nil(resp)
				suite.Require().Equal(*tc.wantErr, *err)
				if tc.wantErr == &ErrorInvalidLimit {
					store.AssertNotCalled(suite.T(), "IsOrganizationUnitExists", mock.Anything)
				}
				return
			}

			suite.Require().Nil(err)
			if tc.assert != nil {
				tc.assert(resp)
			}
		})
	}
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_BuildUserListResponse_InvalidType() {
	resp, err := buildUserListResponse("invalid", 10, 5, 0)
	suite.Nil(resp)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError, *err)
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_BuildGroupListResponse_InvalidType() {
	resp, err := buildGroupListResponse(123, 10, 5, 0)
	suite.Nil(resp)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError, *err)
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_BuildOrganizationUnitListResponse_InvalidType() {
	resp, err := buildOrganizationUnitListResponse(struct{}{}, 10, 5, 0)
	suite.Nil(resp)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError, *err)
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_GetOrganizationUnitGroupsByPath() {
	config := newDefaultPathListConfig[*GroupListResponse](
		"  ",
		5,
		0,
		"GetOrganizationUnitGroupsList",
		[]Group{},
		"GetOrganizationUnitGroupsCount",
		0,
		func(resp *GroupListResponse) {
			suite.Require().NotNil(resp)
		},
		invokeGroupsByPath,
	)
	runOUPathListTests(suite, config)
}

func TestOUService_ValidateAndProcessHandlePath(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		handles, err := validateAndProcessHandlePath("   ")

		require.Nil(t, handles)
		require.Equal(t, &ErrorInvalidHandlePath, err)
	})

	t.Run("only slashes", func(t *testing.T) {
		handles, err := validateAndProcessHandlePath("///")

		require.Nil(t, handles)
		require.Equal(t, &ErrorInvalidHandlePath, err)
	})

	t.Run("success", func(t *testing.T) {
		handles, err := validateAndProcessHandlePath(" /root/ child / ")

		require.Nil(t, err)
		require.Equal(t, []string{"root", "child"}, handles)
	})

	t.Run("ignores empty segments", func(t *testing.T) {
		handles, err := validateAndProcessHandlePath("root//child")

		require.Nil(t, err)
		require.Equal(t, []string{"root", "child"}, handles)
	})
}

func TestOUService_ValidatePaginationParams(t *testing.T) {
	require.Equal(t, &ErrorInvalidLimit, validatePaginationParams(0, 0))
	require.Equal(t, &ErrorInvalidLimit, validatePaginationParams(serverconst.MaxPageSize+1, 0))
	require.Equal(t, &ErrorInvalidOffset, validatePaginationParams(10, -1))
	require.Nil(t, validatePaginationParams(10, 0))
}

func TestOUService_BuildPaginationLinks(t *testing.T) {
	links := buildPaginationLinks(5, 5, 20)
	require.Len(t, links, 4)
	require.Equal(t, "first", links[0].Rel)
	require.Equal(t, "/organization-units?offset=0&limit=5", links[0].Href)

	require.Equal(t, "prev", links[1].Rel)
	require.Equal(t, "/organization-units?offset=0&limit=5", links[1].Href)

	require.Equal(t, "next", links[2].Rel)
	require.Equal(t, "/organization-units?offset=10&limit=5", links[2].Href)

	require.Equal(t, "last", links[3].Rel)
	require.Equal(t, "/organization-units?offset=15&limit=5", links[3].Href)
}

func (suite *OrganizationUnitServiceTestSuite) TestOUService_CheckCircularDependency() {
	store := newOrganizationUnitStoreInterfaceMock(suite.T())
	service := suite.newService(store)
	parentID := testParentID

	store.On("GetOrganizationUnit", parentID).
		Return(OrganizationUnit{ID: parentID, Parent: strPtr("grand")}, nil).
		Once()
	store.On("GetOrganizationUnit", "grand").
		Return(OrganizationUnit{ID: "grand", Parent: strPtr("ou-1")}, nil).
		Once()

	err := service.checkCircularDependency("ou-1", &parentID)

	suite.Require().Equal(&ErrorCircularDependency, err)

	store2 := newOrganizationUnitStoreInterfaceMock(suite.T())
	service2 := suite.newService(store2)
	store2.On("GetOrganizationUnit", parentID).
		Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
		Once()

	err = service2.checkCircularDependency("ou-1", &parentID)
	suite.Require().Nil(err)

	store3 := newOrganizationUnitStoreInterfaceMock(suite.T())
	service3 := suite.newService(store3)
	store3.On("GetOrganizationUnit", parentID).
		Return(OrganizationUnit{}, errors.New("boom")).
		Once()

	err = service3.checkCircularDependency("ou-1", &parentID)
	suite.Require().Equal(&ErrorInternalServerError, err)
}
