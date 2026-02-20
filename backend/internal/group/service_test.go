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

package group

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

// stubTransactioner is a stub implementation of Transactioner for testing.
// It simply executes the function without actual transaction management.
type stubTransactioner struct{}

func (s *stubTransactioner) Transact(ctx context.Context, txFunc func(context.Context) error) error {
	return txFunc(ctx)
}

type GroupServiceTestSuite struct {
	suite.Suite
}

func TestGroupServiceTestSuite(t *testing.T) {
	suite.Run(t, new(GroupServiceTestSuite))
}

type groupRequestValidationTestCase[T any] struct {
	name    string
	request T
	wantErr bool
}

type groupListExpectations struct {
	totalResults int
	count        int
	startIndex   int
	groupNames   []string
	linkRels     []string
	linkHrefs    []string
}

func (suite *GroupServiceTestSuite) assertGroupListResponse(
	response *GroupListResponse,
	expected *groupListExpectations,
) {
	suite.Require().NotNil(response)
	suite.Require().Equal(expected.totalResults, response.TotalResults)
	suite.Require().Equal(expected.count, response.Count)
	suite.Require().Equal(expected.startIndex, response.StartIndex)
	suite.Require().Len(response.Groups, len(expected.groupNames))
	for idx, name := range expected.groupNames {
		suite.Require().Equal(name, response.Groups[idx].Name)
	}
	suite.Require().Len(response.Links, len(expected.linkRels))
	for idx := range expected.linkRels {
		suite.Require().Equal(expected.linkRels[idx], response.Links[idx].Rel)
		suite.Require().Equal(expected.linkHrefs[idx], response.Links[idx].Href)
	}
}

func runGroupRequestValidationTests[T any](
	suite *GroupServiceTestSuite,
	testCases []groupRequestValidationTestCase[T],
	validate func(T) *serviceerror.ServiceError,
) {
	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			err := validate(tc.request)
			if tc.wantErr {
				suite.Require().NotNil(err)
				suite.Require().Equal(ErrorInvalidRequestFormat, *err)
			} else {
				suite.Require().Nil(err)
			}
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_GetGroupList() {
	testCases := []struct {
		name       string
		limit      int
		offset     int
		setup      func(*groupStoreInterfaceMock)
		wantErr    *serviceerror.ServiceError
		wantResult *groupListExpectations
	}{
		{
			name:   "success",
			limit:  2,
			offset: 1,
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroupListCount", mock.Anything).
					Return(3, nil).
					Once()
				storeMock.On("GetGroupList", mock.Anything, 2, 1).
					Return([]GroupBasicDAO{
						{ID: "g1", Name: "group-1", Description: "desc-1", OrganizationUnitID: "ou-1"},
						{ID: "g2", Name: "group-2", Description: "desc-2", OrganizationUnitID: "ou-2"},
					}, nil).
					Once()
			},
			wantResult: &groupListExpectations{
				totalResults: 3,
				count:        2,
				startIndex:   2,
				groupNames:   []string{"group-1", "group-2"},
				linkRels:     []string{"first", "prev", "last"},
				linkHrefs: []string{"/groups?offset=0&limit=2", "/groups?offset=0&limit=2",
					"/groups?offset=2&limit=2"},
			},
		},
		{
			name:    "invalid pagination",
			limit:   0,
			offset:  0,
			wantErr: &ErrorInvalidLimit,
		},
		{
			name:   "count retrieval error",
			limit:  5,
			offset: 0,
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroupListCount", mock.Anything).
					Return(0, errors.New("count failure")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:   "list retrieval error",
			limit:  5,
			offset: 0,
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroupListCount", mock.Anything).
					Return(2, nil).
					Once()
				storeMock.On("GetGroupList", mock.Anything, 5, 0).
					Return(nil, errors.New("list failure")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			storeMock := newGroupStoreInterfaceMock(suite.T())

			if tc.setup != nil {
				tc.setup(storeMock)
			}

			service := &groupService{
				groupStore: storeMock,
			}

			response, err := service.GetGroupList(context.Background(), tc.limit, tc.offset)

			if tc.wantErr != nil {
				suite.Require().Nil(response)
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				suite.assertGroupListResponse(response, tc.wantResult)
			}

			if tc.wantErr == &ErrorInvalidLimit {
				storeMock.AssertNotCalled(suite.T(), "GetGroupListCount", mock.Anything)
			}
			storeMock.AssertExpectations(suite.T())
		})
	}
}
func (suite *GroupServiceTestSuite) TestGroupService_GetGroupsByPath() {
	testCases := []struct {
		name   string
		path   string
		limit  int
		offset int
		setup  func(
			*groupStoreInterfaceMock,
			*oumock.OrganizationUnitServiceInterfaceMock,
		) *serviceerror.ServiceError
		wantErr             *serviceerror.ServiceError
		wantErrFromSetup    bool
		wantResult          *groupListExpectations
		assertStoreCalls    func(*groupStoreInterfaceMock)
		assertOUServiceCall func(*oumock.OrganizationUnitServiceInterfaceMock)
	}{
		{
			name:   "success",
			path:   "root/child",
			limit:  2,
			offset: 0,
			setup: func(
				storeMock *groupStoreInterfaceMock,
				ouMock *oumock.OrganizationUnitServiceInterfaceMock,
			) *serviceerror.ServiceError {
				storeMock.On("GetGroupsByOrganizationUnitCount", mock.Anything, "ou-123").
					Return(4, nil).
					Once()
				storeMock.On("GetGroupsByOrganizationUnit", mock.Anything, "ou-123", 2, 0).
					Return([]GroupBasicDAO{
						{ID: "g1", Name: "group-1", OrganizationUnitID: "ou-123"},
						{ID: "g2", Name: "group-2", OrganizationUnitID: "ou-123"},
					}, nil).
					Once()

				ouMock.On("GetOrganizationUnitByPath", "root/child").
					Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil).
					Once()
				return nil
			},
			wantResult: &groupListExpectations{
				totalResults: 4,
				count:        2,
				startIndex:   1,
				groupNames:   []string{"group-1", "group-2"},
				linkRels:     []string{"next", "last"},
				linkHrefs:    []string{"/groups?offset=2&limit=2", "/groups?offset=2&limit=2"},
			},
		},
		{
			name:    "invalid path",
			path:    "  ",
			limit:   10,
			offset:  0,
			wantErr: &ErrorInvalidRequestFormat,
			assertOUServiceCall: func(ouMock *oumock.OrganizationUnitServiceInterfaceMock) {
				ouMock.AssertNotCalled(suite.T(), "GetOrganizationUnitByPath", mock.Anything)
			},
			assertStoreCalls: func(storeMock *groupStoreInterfaceMock) {
				storeMock.AssertNotCalled(suite.T(), "GetGroupsByOrganizationUnitCount", mock.Anything, mock.Anything)
			},
		},
		{
			name:   "organization unit not found",
			path:   "root/child",
			limit:  10,
			offset: 0,
			setup: func(
				storeMock *groupStoreInterfaceMock,
				ouMock *oumock.OrganizationUnitServiceInterfaceMock,
			) *serviceerror.ServiceError {
				ouMock.On("GetOrganizationUnitByPath", "root/child").
					Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound).
					Once()
				return nil
			},
			wantErr: &ErrorGroupNotFound,
			assertStoreCalls: func(storeMock *groupStoreInterfaceMock) {
				storeMock.AssertNotCalled(suite.T(), "GetGroupsByOrganizationUnitCount", mock.Anything, mock.Anything)
			},
		},
		{
			name:   "organization unit service error",
			path:   "root/child",
			limit:  5,
			offset: 0,
			setup: func(
				storeMock *groupStoreInterfaceMock,
				ouMock *oumock.OrganizationUnitServiceInterfaceMock,
			) *serviceerror.ServiceError {
				expectedErr := &serviceerror.ServiceError{
					Code: "OU-5000",
					Type: serviceerror.ServerErrorType,
				}
				ouMock.On("GetOrganizationUnitByPath", "root/child").
					Return(oupkg.OrganizationUnit{}, expectedErr).
					Once()
				return expectedErr
			},
			wantErrFromSetup: true,
		},
		{
			name:   "invalid pagination",
			path:   "root/child",
			limit:  0,
			offset: 0,
			setup: func(
				storeMock *groupStoreInterfaceMock,
				ouMock *oumock.OrganizationUnitServiceInterfaceMock,
			) *serviceerror.ServiceError {
				ouMock.On("GetOrganizationUnitByPath", "root/child").
					Return(oupkg.OrganizationUnit{ID: "ou-1"}, nil).
					Once()
				return nil
			},
			wantErr: &ErrorInvalidLimit,
			assertStoreCalls: func(storeMock *groupStoreInterfaceMock) {
				storeMock.AssertNotCalled(suite.T(), "GetGroupsByOrganizationUnitCount", mock.Anything, mock.Anything)
			},
		},
		{
			name:   "count retrieval error",
			path:   "root/child",
			limit:  5,
			offset: 0,
			setup: func(
				storeMock *groupStoreInterfaceMock,
				ouMock *oumock.OrganizationUnitServiceInterfaceMock,
			) *serviceerror.ServiceError {
				storeMock.On("GetGroupsByOrganizationUnitCount", mock.Anything, "ou-123").
					Return(0, errors.New("count fail")).
					Once()

				ouMock.On("GetOrganizationUnitByPath", "root/child").
					Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil).
					Once()
				return nil
			},
			wantErr: &ErrorInternalServerError,
			assertStoreCalls: func(storeMock *groupStoreInterfaceMock) {
				storeMock.AssertNotCalled(suite.T(), "GetGroupsByOrganizationUnit",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:   "list retrieval error",
			path:   "root/child",
			limit:  5,
			offset: 0,
			setup: func(
				storeMock *groupStoreInterfaceMock,
				ouMock *oumock.OrganizationUnitServiceInterfaceMock,
			) *serviceerror.ServiceError {
				storeMock.On("GetGroupsByOrganizationUnitCount", mock.Anything, "ou-123").
					Return(1, nil).
					Once()
				storeMock.On("GetGroupsByOrganizationUnit", mock.Anything, "ou-123", 5, 0).
					Return(nil, errors.New("list fail")).
					Once()

				ouMock.On("GetOrganizationUnitByPath", "root/child").
					Return(oupkg.OrganizationUnit{ID: "ou-123"}, nil).
					Once()
				return nil
			},
			wantErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			storeMock := newGroupStoreInterfaceMock(suite.T())
			ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

			var expectedErr *serviceerror.ServiceError
			if tc.setup != nil {
				expectedErr = tc.setup(storeMock, ouServiceMock)
			}

			service := &groupService{
				groupStore: storeMock,
				ouService:  ouServiceMock,
			}

			response, err := service.GetGroupsByPath(context.Background(), tc.path, tc.limit, tc.offset)

			if tc.wantErr != nil || tc.wantErrFromSetup {
				suite.Require().Nil(response)
				suite.Require().NotNil(err)
				if tc.wantErrFromSetup {
					suite.Require().Equal(expectedErr, err)
				} else {
					suite.Require().Equal(*tc.wantErr, *err)
				}
			} else {
				suite.Require().Nil(err)
				suite.assertGroupListResponse(response, tc.wantResult)
			}

			if tc.assertStoreCalls != nil {
				tc.assertStoreCalls(storeMock)
			}
			if tc.assertOUServiceCall != nil {
				tc.assertOUServiceCall(ouServiceMock)
			}

			storeMock.AssertExpectations(suite.T())
			ouServiceMock.AssertExpectations(suite.T())
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_CreateGroup() {
	type setupArgs struct {
		store *groupStoreInterfaceMock
		ou    *oumock.OrganizationUnitServiceInterfaceMock
		user  *usermock.UserServiceInterfaceMock
	}

	testCases := []struct {
		name      string
		request   CreateGroupRequest
		setup     func(*setupArgs)
		expectErr *serviceerror.ServiceError
		expectRes bool
	}{
		{
			name: "success",
			request: CreateGroupRequest{
				Name:               "engineering",
				Description:        "Engineers",
				OrganizationUnitID: "ou-001",
				Members: []Member{
					{ID: "usr-001", Type: MemberTypeUser},
					{ID: "grp-002", Type: MemberTypeGroup},
				},
			},
			setup: func(args *setupArgs) {
				args.store.On("CheckGroupNameConflictForCreate", mock.Anything, "engineering", "ou-001").
					Return(nil).
					Once()
				args.store.On("ValidateGroupIDs", mock.Anything, []string{"grp-002"}).
					Return([]string{}, nil).
					Once()
				args.store.On("CreateGroup", mock.Anything, mock.MatchedBy(func(group GroupDAO) bool {
					return group.Name == "engineering" &&
						group.OrganizationUnitID == "ou-001" &&
						len(group.Members) == 2
				})).
					Return(nil).
					Once()

				args.ou.On("GetOrganizationUnit", "ou-001").
					Return(oupkg.OrganizationUnit{ID: "ou-001"}, nil).
					Once()

				args.user.On("ValidateUserIDs", mock.Anything, []string{"usr-001"}).
					Return([]string{}, nil).
					Once()
			},
			expectRes: true,
		},
		{
			name: "invalid organization unit",
			request: CreateGroupRequest{
				Name:               "engineering",
				OrganizationUnitID: "ou-unknown",
			},
			setup: func(args *setupArgs) {
				args.ou.On("GetOrganizationUnit", "ou-unknown").
					Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound).
					Once()
				args.user.On("ValidateUserIDs", mock.Anything, []string{}).
					Return([]string{}, nil).
					Maybe()
			},
			expectErr: &ErrorInvalidOUID,
		},
		{
			name: "invalid user IDs",
			request: CreateGroupRequest{
				Name:               "engineering",
				OrganizationUnitID: "ou-001",
				Members:            []Member{{ID: "usr-invalid", Type: MemberTypeUser}},
			},
			setup: func(args *setupArgs) {
				args.store.On("CheckGroupNameConflictForCreate", mock.Anything, "engineering", "ou-001").
					Return(nil).
					Maybe()
				args.ou.On("GetOrganizationUnit", "ou-001").
					Return(oupkg.OrganizationUnit{ID: "ou-001"}, nil).
					Once()
				args.user.On("ValidateUserIDs", mock.Anything, []string{"usr-invalid"}).
					Return([]string{"usr-invalid"}, nil).
					Once()
			},
			expectErr: &ErrorInvalidUserMemberID,
		},
		{
			name: "name conflict",
			request: CreateGroupRequest{
				Name:               "engineering",
				OrganizationUnitID: "ou-001",
			},
			setup: func(args *setupArgs) {
				args.store.On("CheckGroupNameConflictForCreate", mock.Anything, "engineering", "ou-001").
					Return(ErrGroupNameConflict).
					Once()
				args.ou.On("GetOrganizationUnit", "ou-001").
					Return(oupkg.OrganizationUnit{ID: "ou-001"}, nil).
					Once()
				args.user.On("ValidateUserIDs", mock.Anything, mock.Anything).
					Return([]string{}, nil).
					Once()
				args.store.On("ValidateGroupIDs", mock.Anything, mock.Anything).
					Return([]string{}, nil).
					Once()
			},
			expectErr: &ErrorGroupNameConflict,
		},
		{
			name: "conflict check error",
			request: CreateGroupRequest{
				Name:               "engineering",
				OrganizationUnitID: "ou-001",
			},
			setup: func(args *setupArgs) {
				args.store.On("CheckGroupNameConflictForCreate", mock.Anything, "engineering", "ou-001").
					Return(errors.New("db failure")).
					Once()
				args.ou.On("GetOrganizationUnit", "ou-001").
					Return(oupkg.OrganizationUnit{ID: "ou-001"}, nil).
					Once()
				args.user.On("ValidateUserIDs", mock.Anything, mock.Anything).
					Return([]string{}, nil).
					Once()
				args.store.On("ValidateGroupIDs", mock.Anything, mock.Anything).
					Return([]string{}, nil).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
		{
			name: "create error",
			request: CreateGroupRequest{
				Name:               "engineering",
				OrganizationUnitID: "ou-001",
			},
			setup: func(args *setupArgs) {
				args.store.On("CheckGroupNameConflictForCreate", mock.Anything, "engineering", "ou-001").
					Return(nil).
					Once()
				args.store.On("ValidateGroupIDs", mock.Anything, mock.Anything).
					Return([]string{}, nil).
					Once()
				args.store.On("CreateGroup", mock.Anything, mock.Anything).
					Return(errors.New("create fail")).
					Once()
				args.ou.On("GetOrganizationUnit", "ou-001").
					Return(oupkg.OrganizationUnit{ID: "ou-001"}, nil).
					Once()
				args.user.On("ValidateUserIDs", mock.Anything, mock.Anything).
					Return([]string{}, nil).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
		{
			name: "organization unit service error",
			request: CreateGroupRequest{
				Name:               "engineering",
				OrganizationUnitID: "ou-001",
			},
			setup: func(args *setupArgs) {
				args.ou.On("GetOrganizationUnit", "ou-001").
					Return(oupkg.OrganizationUnit{},
						&serviceerror.ServiceError{Code: "OU-5000", Type: serviceerror.ServerErrorType}).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			var storeMock *groupStoreInterfaceMock
			var ouServiceMock *oumock.OrganizationUnitServiceInterfaceMock
			var userServiceMock *usermock.UserServiceInterfaceMock

			if tc.setup != nil {
				storeMock = newGroupStoreInterfaceMock(suite.T())
				ouServiceMock = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())
				userServiceMock = usermock.NewUserServiceInterfaceMock(suite.T())
				tc.setup(&setupArgs{store: storeMock, ou: ouServiceMock, user: userServiceMock})
			}

			service := &groupService{
				groupStore:    storeMock,
				ouService:     ouServiceMock,
				userService:   userServiceMock,
				transactioner: &stubTransactioner{},
			}

			group, err := service.CreateGroup(context.Background(), tc.request)

			if tc.expectErr != nil {
				suite.Require().Nil(group)
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.expectErr, *err)
			} else if tc.expectRes {
				suite.Require().Nil(err)
				suite.Require().NotNil(group)
			} else {
				suite.Require().Nil(err)
			}

			if storeMock != nil {
				storeMock.AssertExpectations(suite.T())
			}
			if ouServiceMock != nil {
				ouServiceMock.AssertExpectations(suite.T())
			}
			if userServiceMock != nil {
				userServiceMock.AssertExpectations(suite.T())
			}
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_CreateGroupByPath() {
	type setupArgs struct {
		store *groupStoreInterfaceMock
		ou    *oumock.OrganizationUnitServiceInterfaceMock
		user  *usermock.UserServiceInterfaceMock
	}

	testCases := []struct {
		name      string
		path      string
		request   CreateGroupByPathRequest
		setup     func(*setupArgs) *serviceerror.ServiceError
		expectErr *serviceerror.ServiceError
	}{
		{
			name:      "invalid path",
			path:      " ",
			request:   CreateGroupByPathRequest{Name: "n"},
			expectErr: &ErrorInvalidRequestFormat,
		},
		{
			name:    "organization unit service error",
			path:    "root",
			request: CreateGroupByPathRequest{Name: "n"},
			setup: func(args *setupArgs) *serviceerror.ServiceError {
				expected := &serviceerror.ServiceError{Code: "OU-5000", Type: serviceerror.ServerErrorType}
				args.ou.On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{}, expected).
					Once()
				return expected
			},
			expectErr: &serviceerror.ServiceError{Code: "OU-5000", Type: serviceerror.ServerErrorType},
		},
		{
			name:    "organization unit not found",
			path:    "root",
			request: CreateGroupByPathRequest{Name: "n"},
			setup: func(args *setupArgs) *serviceerror.ServiceError {
				args.ou.On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound).
					Once()
				return nil
			},
			expectErr: &ErrorGroupNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			var storeMock *groupStoreInterfaceMock
			var ouServiceMock *oumock.OrganizationUnitServiceInterfaceMock
			var userServiceMock *usermock.UserServiceInterfaceMock
			var expectedOUError *serviceerror.ServiceError

			if tc.setup != nil {
				storeMock = newGroupStoreInterfaceMock(suite.T())
				ouServiceMock = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())
				userServiceMock = usermock.NewUserServiceInterfaceMock(suite.T())
				expectedOUError = tc.setup(&setupArgs{store: storeMock, ou: ouServiceMock, user: userServiceMock})
			}

			service := &groupService{
				groupStore:    storeMock,
				ouService:     ouServiceMock,
				userService:   userServiceMock,
				transactioner: &stubTransactioner{},
			}

			group, err := service.CreateGroupByPath(context.Background(), tc.path, tc.request)

			if tc.expectErr != nil {
				if expectedOUError != nil {
					suite.Require().Equal(expectedOUError, err)
				} else {
					suite.Require().Nil(group)
					suite.Require().NotNil(err)
					suite.Require().Equal(*tc.expectErr, *err)
				}
			}

			if storeMock != nil {
				storeMock.AssertExpectations(suite.T())
			}
			if ouServiceMock != nil {
				ouServiceMock.AssertExpectations(suite.T())
			}
			if userServiceMock != nil {
				userServiceMock.AssertExpectations(suite.T())
			}
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_GetGroup() {
	testCases := []struct {
		name    string
		id      string
		setup   func(*groupStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
	}{
		{
			name:    "missing id",
			id:      "",
			wantErr: &ErrorMissingGroupID,
		},
		{
			name: "internal error",
			id:   "grp-001",
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{}, errors.New("db error")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "not found",
			id:   "grp-404",
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-404").
					Return(GroupDAO{}, ErrGroupNotFound).
					Once()
			},
			wantErr: &ErrorGroupNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			var storeMock *groupStoreInterfaceMock

			if tc.setup != nil {
				storeMock = newGroupStoreInterfaceMock(suite.T())
				tc.setup(storeMock)
			}

			service := &groupService{groupStore: storeMock}

			group, err := service.GetGroup(context.Background(), tc.id)

			if tc.wantErr != nil {
				suite.Require().Nil(group)
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
			} else {
				suite.Require().Nil(err)
				suite.Require().NotNil(group)
			}

			if storeMock != nil {
				storeMock.AssertExpectations(suite.T())
			}
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_UpdateGroup() {
	type setupArgs struct {
		store *groupStoreInterfaceMock
		ou    *oumock.OrganizationUnitServiceInterfaceMock
		user  *usermock.UserServiceInterfaceMock
	}

	testCases := []struct {
		name        string
		groupID     string
		request     UpdateGroupRequest
		setup       func(*setupArgs)
		expectErr   *serviceerror.ServiceError
		expectGroup bool
	}{
		{
			name:      "missing id",
			groupID:   "",
			expectErr: &ErrorMissingGroupID,
		},
		{
			name:      "invalid request",
			groupID:   "grp-001",
			request:   UpdateGroupRequest{},
			expectErr: &ErrorInvalidRequestFormat,
		},
		{
			name:    "success",
			groupID: "grp-001",
			request: UpdateGroupRequest{
				Name:               "new-name",
				Description:        "New desc",
				OrganizationUnitID: "ou-new",
			},
			setup: func(args *setupArgs) {
				args.store.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "old", Description: "legacy",
						OrganizationUnitID: "ou-old"}, nil).
					Once()
				args.store.On("CheckGroupNameConflictForUpdate", mock.Anything, "new-name", "ou-new", "grp-001").
					Return(nil).
					Once()
				args.store.On("UpdateGroup", mock.Anything, mock.MatchedBy(func(group GroupDAO) bool {
					return group.ID == "grp-001" && group.Name == "new-name" && group.OrganizationUnitID == "ou-new"
				})).
					Return(nil).
					Once()
				args.ou.On("GetOrganizationUnit", "ou-new").
					Return(oupkg.OrganizationUnit{ID: "ou-new"}, nil).
					Once()
			},
			expectGroup: true,
		},
		{
			name:    "name conflict",
			groupID: "grp-001",
			request: UpdateGroupRequest{
				Name:               "new-name",
				OrganizationUnitID: "ou-new",
			},
			setup: func(args *setupArgs) {
				args.store.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "old", OrganizationUnitID: "ou-old"}, nil).
					Once()
				args.store.On("CheckGroupNameConflictForUpdate", mock.Anything, "new-name", "ou-new", "grp-001").
					Return(ErrGroupNameConflict).
					Once()
				args.ou.On("GetOrganizationUnit", "ou-new").
					Return(oupkg.OrganizationUnit{ID: "ou-new"}, nil).
					Once()
			},
			expectErr: &ErrorGroupNameConflict,
		},
		{
			name:    "group not found",
			groupID: "grp-001",
			request: UpdateGroupRequest{
				Name:               "name",
				OrganizationUnitID: "ou",
			},
			setup: func(args *setupArgs) {
				args.store.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{}, ErrGroupNotFound).
					Once()
			},
			expectErr: &ErrorGroupNotFound,
		},
		{
			name:    "get group error",
			groupID: "grp-001",
			request: UpdateGroupRequest{
				Name:               "name",
				OrganizationUnitID: "ou",
			},
			setup: func(args *setupArgs) {
				args.store.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{}, errors.New("db error")).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
		{
			name:    "validate organization unit error",
			groupID: "grp-001",
			request: UpdateGroupRequest{
				Name:               "name",
				OrganizationUnitID: "ou-new",
			},
			setup: func(args *setupArgs) {
				args.store.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "name", OrganizationUnitID: "ou-old"}, nil).
					Once()
				args.ou.On("GetOrganizationUnit", "ou-new").
					Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound).
					Once()
			},
			expectErr: &ErrorInvalidOUID,
		},
		{
			name:    "conflict check error",
			groupID: "grp-001",
			request: UpdateGroupRequest{
				Name:               "new",
				OrganizationUnitID: "ou",
			},
			setup: func(args *setupArgs) {
				args.store.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "old", OrganizationUnitID: "ou"}, nil).
					Once()
				args.store.On("CheckGroupNameConflictForUpdate", mock.Anything, "new", "ou", "grp-001").
					Return(errors.New("db error")).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
		{
			name:    "update error",
			groupID: "grp-001",
			request: UpdateGroupRequest{
				Name:               "new",
				OrganizationUnitID: "ou",
			},
			setup: func(args *setupArgs) {
				args.store.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "old-name", OrganizationUnitID: "ou"}, nil).
					Once()
				args.store.On("CheckGroupNameConflictForUpdate", mock.Anything, "new", "ou", "grp-001").
					Return(nil).
					Once()
				args.store.On("UpdateGroup", mock.Anything, mock.Anything).
					Return(errors.New("update fail")).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			var storeMock *groupStoreInterfaceMock
			var ouServiceMock *oumock.OrganizationUnitServiceInterfaceMock
			var userServiceMock *usermock.UserServiceInterfaceMock

			if tc.setup != nil {
				storeMock = newGroupStoreInterfaceMock(suite.T())
				ouServiceMock = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())
				userServiceMock = usermock.NewUserServiceInterfaceMock(suite.T())
				tc.setup(&setupArgs{store: storeMock, ou: ouServiceMock, user: userServiceMock})
			}

			service := &groupService{
				groupStore:    storeMock,
				ouService:     ouServiceMock,
				userService:   userServiceMock,
				transactioner: &stubTransactioner{},
			}

			group, err := service.UpdateGroup(context.Background(), tc.groupID, tc.request)

			if tc.expectErr != nil {
				suite.Require().Nil(group)
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.expectErr, *err)
			} else if tc.expectGroup {
				suite.Require().Nil(err)
				suite.Require().NotNil(group)
			} else {
				suite.Require().Nil(err)
			}

			if storeMock != nil {
				storeMock.AssertExpectations(suite.T())
			}
			if ouServiceMock != nil {
				ouServiceMock.AssertExpectations(suite.T())
			}
			if userServiceMock != nil {
				userServiceMock.AssertExpectations(suite.T())
			}
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_DeleteGroup() {
	testCases := []struct {
		name      string
		id        string
		setup     func(*groupStoreInterfaceMock)
		expectErr *serviceerror.ServiceError
	}{
		{
			name: "success",
			id:   "grp-001",
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001"}, nil).
					Once()
				storeMock.On("DeleteGroup", mock.Anything, "grp-001").
					Return(nil).
					Once()
			},
		},
		{
			name:      "missing id",
			id:        "",
			expectErr: &ErrorMissingGroupID,
		},
		{
			name: "get group error",
			id:   "grp-001",
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{}, errors.New("db error")).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
		{
			name: "delete error",
			id:   "grp-001",
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001"}, nil).
					Once()
				storeMock.On("DeleteGroup", mock.Anything, "grp-001").
					Return(errors.New("delete fail")).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
		{
			name: "group not found",
			id:   "grp-001",
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{}, ErrGroupNotFound).
					Once()
			},
			expectErr: &ErrorGroupNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			var storeMock *groupStoreInterfaceMock
			if tc.setup != nil {
				storeMock = newGroupStoreInterfaceMock(suite.T())
				tc.setup(storeMock)
			}

			service := &groupService{groupStore: storeMock, transactioner: &stubTransactioner{}}

			err := service.DeleteGroup(context.Background(), tc.id)

			if tc.expectErr != nil {
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.expectErr, *err)
			} else {
				suite.Require().Nil(err)
			}

			if storeMock != nil {
				storeMock.AssertExpectations(suite.T())
			}
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_GetGroupMembers() {
	testCases := []struct {
		name      string
		id        string
		limit     int
		offset    int
		setup     func(*groupStoreInterfaceMock)
		expectErr *serviceerror.ServiceError
		expectRes bool
	}{
		{
			name:   "success",
			id:     "grp-001",
			limit:  2,
			offset: 0,
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001"}, nil).
					Once()
				storeMock.On("GetGroupMemberCount", mock.Anything, "grp-001").
					Return(3, nil).
					Once()
				storeMock.On("GetGroupMembers", mock.Anything, "grp-001", 2, 0).
					Return([]Member{
						{ID: "usr-001", Type: MemberTypeUser},
						{ID: "grp-002", Type: MemberTypeGroup},
					}, nil).
					Once()
			},
			expectRes: true,
		},
		{
			name:   "group not found",
			id:     "grp-001",
			limit:  5,
			offset: 0,
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{}, ErrGroupNotFound).
					Once()
			},
			expectErr: &ErrorGroupNotFound,
		},
		{
			name:      "invalid pagination",
			id:        "grp-001",
			limit:     0,
			offset:    0,
			expectErr: &ErrorInvalidLimit,
		},
		{
			name:      "missing id",
			id:        "",
			limit:     5,
			offset:    0,
			expectErr: &ErrorMissingGroupID,
		},
		{
			name:   "get group error",
			id:     "grp-001",
			limit:  5,
			offset: 0,
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{}, errors.New("db error")).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
		{
			name:   "count error",
			id:     "grp-001",
			limit:  5,
			offset: 0,
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001"}, nil).
					Once()
				storeMock.On("GetGroupMemberCount", mock.Anything, "grp-001").
					Return(0, errors.New("count fail")).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
		{
			name:   "list error",
			id:     "grp-001",
			limit:  5,
			offset: 0,
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001"}, nil).
					Once()
				storeMock.On("GetGroupMemberCount", mock.Anything, "grp-001").
					Return(1, nil).
					Once()
				storeMock.On("GetGroupMembers", mock.Anything, "grp-001", 5, 0).
					Return(nil, errors.New("list fail")).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			var storeMock *groupStoreInterfaceMock
			if tc.setup != nil {
				storeMock = newGroupStoreInterfaceMock(suite.T())
				tc.setup(storeMock)
			}

			service := &groupService{groupStore: storeMock}

			response, err := service.GetGroupMembers(context.Background(), tc.id, tc.limit, tc.offset)

			if tc.expectErr != nil {
				suite.Require().Nil(response)
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.expectErr, *err)
			} else if tc.expectRes {
				suite.Require().Nil(err)
				suite.Require().NotNil(response)
				suite.Require().Equal(3, response.TotalResults)
				suite.Require().Equal(2, response.Count)
				suite.Require().Equal(1, response.StartIndex)
				suite.Require().Len(response.Members, 2)
			} else {
				suite.Require().Nil(err)
			}

			if storeMock != nil {
				storeMock.AssertExpectations(suite.T())
			}
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_ValidateCreateGroupRequest() {
	service := &groupService{}

	testCases := []groupRequestValidationTestCase[CreateGroupRequest]{
		{
			name:    "missing fields",
			request: CreateGroupRequest{},
			wantErr: true,
		},
		{
			name:    "missing organization unit",
			request: CreateGroupRequest{Name: "name"},
			wantErr: true,
		},
		{
			name: "invalid member type",
			request: CreateGroupRequest{
				Name:               "name",
				OrganizationUnitID: "ou",
				Members:            []Member{{ID: "id", Type: "invalid"}},
			},
			wantErr: true,
		},
		{
			name: "missing member id",
			request: CreateGroupRequest{
				Name:               "name",
				OrganizationUnitID: "ou",
				Members:            []Member{{ID: "", Type: MemberTypeUser}},
			},
			wantErr: true,
		},
		{
			name: "valid request",
			request: CreateGroupRequest{
				Name:               "name",
				OrganizationUnitID: "ou",
				Members:            []Member{{ID: "usr-1", Type: MemberTypeUser}},
			},
			wantErr: false,
		},
	}

	runGroupRequestValidationTests(suite, testCases, service.validateCreateGroupRequest)
}

func (suite *GroupServiceTestSuite) TestGroupService_ValidateUpdateGroupRequest() {
	service := &groupService{}

	testCases := []groupRequestValidationTestCase[UpdateGroupRequest]{
		{
			name:    "missing fields",
			request: UpdateGroupRequest{},
			wantErr: true,
		},
		{
			name:    "missing organization unit",
			request: UpdateGroupRequest{Name: "name"},
			wantErr: true,
		},
		{
			name: "valid request",
			request: UpdateGroupRequest{
				Name:               "name",
				OrganizationUnitID: "ou",
			},
			wantErr: false,
		},
	}

	runGroupRequestValidationTests(suite, testCases, service.validateUpdateGroupRequest)
}

func (suite *GroupServiceTestSuite) TestGroupService_ValidateOUHandlesInternalError() {
	t := suite.T()
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	ouServiceMock.On("GetOrganizationUnit", "ou-1").
		Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{
			Code: "OU-5000",
			Type: serviceerror.ServerErrorType,
		}).
		Once()

	service := &groupService{
		ouService: ouServiceMock,
	}

	err := service.validateOU("ou-1")

	require.NotNil(t, err)
	require.Equal(t, ErrorInternalServerError, *err)
}

func (suite *GroupServiceTestSuite) TestGroupService_ValidateAndProcessHandlePath() {
	t := suite.T()
	service := &groupService{}

	testCases := []struct {
		name        string
		handlePath  string
		expectError bool
	}{
		{
			name:        "empty string",
			handlePath:  "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			handlePath:  "   ",
			expectError: true,
		},
		{
			name:        "only slashes",
			handlePath:  "///",
			expectError: true,
		},
		{
			name:        "double slash between handles",
			handlePath:  "root//child",
			expectError: true,
		},
		{
			name:        "single slash",
			handlePath:  "/",
			expectError: true,
		},
		{
			name:        "valid handles",
			handlePath:  "root/child",
			expectError: false,
		},
		{
			name:        "valid handles with surrounding whitespace and slashes",
			handlePath:  "  /root/child/  ",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.validateAndProcessHandlePath(tc.handlePath)
			if tc.expectError {
				require.NotNil(t, err)
				require.Equal(t, ErrorInvalidRequestFormat, *err)
				return
			}

			require.Nil(t, err)
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_ValidateUserIDsHandlesServiceError() {
	t := suite.T()
	userServiceMock := usermock.NewUserServiceInterfaceMock(t)
	userServiceMock.On("ValidateUserIDs", mock.Anything, []string{"usr-001"}).
		Return([]string{}, &serviceerror.ServiceError{
			Code: "USR-5000",
			Type: serviceerror.ServerErrorType,
		}).
		Once()

	service := &groupService{
		userService: userServiceMock,
	}

	err := service.validateUserIDs(context.Background(), []string{"usr-001"})

	require.NotNil(t, err)
	require.Equal(t, ErrorInternalServerError, *err)
}

func (suite *GroupServiceTestSuite) TestGroupService_ValidateGroupIDs() {
	testCases := []struct {
		name      string
		setup     func(*groupStoreInterfaceMock)
		expectErr *serviceerror.ServiceError
	}{
		{
			name: "invalid ids",
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("ValidateGroupIDs", mock.Anything, []string{"grp-001"}).
					Return([]string{"grp-001"}, nil).
					Once()
			},
			expectErr: &ErrorInvalidGroupMemberID,
		},
		{
			name: "store error",
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("ValidateGroupIDs", mock.Anything, []string{"grp-001"}).
					Return(nil, errors.New("db error")).
					Once()
			},
			expectErr: &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			storeMock := newGroupStoreInterfaceMock(suite.T())
			service := &groupService{groupStore: storeMock}
			tc.setup(storeMock)

			err := service.ValidateGroupIDs(context.Background(), []string{"grp-001"})

			suite.Require().NotNil(err)
			suite.Require().Equal(*tc.expectErr, *err)

			storeMock.AssertExpectations(suite.T())
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_AddGroupMembers() {
	testCases := []struct {
		name    string
		groupID string
		members []Member
		setup   func(*groupStoreInterfaceMock, *usermock.UserServiceInterfaceMock)
		wantErr *serviceerror.ServiceError
	}{
		{
			name:    "missing group id",
			groupID: "",
			members: []Member{{ID: "usr-001", Type: MemberTypeUser}},
			wantErr: &ErrorMissingGroupID,
		},
		{
			name:    "empty members list",
			groupID: "grp-001",
			members: []Member{},
			wantErr: &ErrorEmptyMembers,
		},
		{
			name:    "invalid member type",
			groupID: "grp-001",
			members: []Member{{ID: "usr-001", Type: "invalid"}},
			wantErr: &ErrorInvalidRequestFormat,
		},
		{
			name:    "empty member id",
			groupID: "grp-001",
			members: []Member{{ID: "", Type: MemberTypeUser}},
			wantErr: &ErrorInvalidRequestFormat,
		},
		{
			name:    "group not found",
			groupID: "grp-001",
			members: []Member{{ID: "usr-001", Type: MemberTypeUser}},
			setup: func(storeMock *groupStoreInterfaceMock, _ *usermock.UserServiceInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{}, ErrGroupNotFound).Once()
			},
			wantErr: &ErrorGroupNotFound,
		},
		{
			name:    "invalid user member id",
			groupID: "grp-001",
			members: []Member{{ID: "usr-invalid", Type: MemberTypeUser}},
			setup: func(storeMock *groupStoreInterfaceMock, userServiceMock *usermock.UserServiceInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "test"}, nil).Once()
				userServiceMock.On("ValidateUserIDs", mock.Anything, []string{"usr-invalid"}).
					Return([]string{"usr-invalid"}, nil).Once()
			},
			wantErr: &ErrorInvalidUserMemberID,
		},
		{
			name:    "store failure",
			groupID: "grp-001",
			members: []Member{{ID: "usr-001", Type: MemberTypeUser}},
			setup: func(storeMock *groupStoreInterfaceMock, userServiceMock *usermock.UserServiceInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "test"}, nil).Once()
				userServiceMock.On("ValidateUserIDs", mock.Anything, []string{"usr-001"}).
					Return([]string{}, nil).Once()
				storeMock.On("ValidateGroupIDs", mock.Anything, mock.Anything).
					Return([]string{}, nil).Once()
				storeMock.On("AddGroupMembers", mock.Anything, "grp-001", mock.Anything).
					Return(errors.New("db error")).Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:    "success",
			groupID: "grp-001",
			members: []Member{{ID: "usr-001", Type: MemberTypeUser}},
			setup: func(storeMock *groupStoreInterfaceMock, userServiceMock *usermock.UserServiceInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "test"}, nil)
				userServiceMock.On("ValidateUserIDs", mock.Anything, []string{"usr-001"}).
					Return([]string{}, nil).Once()
				storeMock.On("ValidateGroupIDs", mock.Anything, mock.Anything).
					Return([]string{}, nil).Once()
				storeMock.On("AddGroupMembers", mock.Anything, "grp-001",
					[]Member{{ID: "usr-001", Type: MemberTypeUser}}).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			storeMock := newGroupStoreInterfaceMock(suite.T())
			userServiceMock := usermock.NewUserServiceInterfaceMock(suite.T())

			if tc.setup != nil {
				tc.setup(storeMock, userServiceMock)
			}

			service := &groupService{
				groupStore:    storeMock,
				userService:   userServiceMock,
				transactioner: &stubTransactioner{},
			}

			group, err := service.AddGroupMembers(context.Background(), tc.groupID, tc.members)

			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
				suite.Require().Nil(group)
			} else {
				suite.Require().Nil(err)
				suite.Require().NotNil(group)
			}

			storeMock.AssertExpectations(suite.T())
			userServiceMock.AssertExpectations(suite.T())
		})
	}
}

func (suite *GroupServiceTestSuite) TestGroupService_RemoveGroupMembers() {
	testCases := []struct {
		name    string
		groupID string
		members []Member
		setup   func(*groupStoreInterfaceMock)
		wantErr *serviceerror.ServiceError
	}{
		{
			name:    "missing group id",
			groupID: "",
			members: []Member{{ID: "usr-001", Type: MemberTypeUser}},
			wantErr: &ErrorMissingGroupID,
		},
		{
			name:    "empty members list",
			groupID: "grp-001",
			members: []Member{},
			wantErr: &ErrorEmptyMembers,
		},
		{
			name:    "invalid member type",
			groupID: "grp-001",
			members: []Member{{ID: "usr-001", Type: "invalid"}},
			wantErr: &ErrorInvalidRequestFormat,
		},
		{
			name:    "empty member id",
			groupID: "grp-001",
			members: []Member{{ID: "", Type: MemberTypeUser}},
			wantErr: &ErrorInvalidRequestFormat,
		},
		{
			name:    "group not found",
			groupID: "grp-001",
			members: []Member{{ID: "usr-001", Type: MemberTypeUser}},
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{}, ErrGroupNotFound).Once()
			},
			wantErr: &ErrorGroupNotFound,
		},
		{
			name:    "store failure",
			groupID: "grp-001",
			members: []Member{{ID: "usr-001", Type: MemberTypeUser}},
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "test"}, nil).Once()
				storeMock.On("RemoveGroupMembers", mock.Anything, "grp-001", mock.Anything).
					Return(errors.New("db error")).Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:    "success",
			groupID: "grp-001",
			members: []Member{{ID: "usr-001", Type: MemberTypeUser}},
			setup: func(storeMock *groupStoreInterfaceMock) {
				storeMock.On("GetGroup", mock.Anything, "grp-001").
					Return(GroupDAO{ID: "grp-001", Name: "test"}, nil)
				storeMock.On("RemoveGroupMembers", mock.Anything, "grp-001",
					[]Member{{ID: "usr-001", Type: MemberTypeUser}}).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			storeMock := newGroupStoreInterfaceMock(suite.T())

			if tc.setup != nil {
				tc.setup(storeMock)
			}

			service := &groupService{
				groupStore:    storeMock,
				transactioner: &stubTransactioner{},
			}

			group, err := service.RemoveGroupMembers(context.Background(), tc.groupID, tc.members)

			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Require().Equal(*tc.wantErr, *err)
				suite.Require().Nil(group)
			} else {
				suite.Require().Nil(err)
				suite.Require().NotNil(group)
			}

			storeMock.AssertExpectations(suite.T())
		})
	}
}
