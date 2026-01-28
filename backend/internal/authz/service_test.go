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

package authz

import (
	"context"
	"errors"
	"testing"

	enginemock "github.com/asgardeo/thunder/tests/mocks/authz/engine"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// AuthorizationServiceTestSuite is the test suite for authorization service.
type AuthorizationServiceTestSuite struct {
	suite.Suite
	mockEngine *enginemock.AuthorizationEngineMock
	service    AuthorizationServiceInterface
}

func TestAuthorizationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationServiceTestSuite))
}

func (suite *AuthorizationServiceTestSuite) SetupTest() {
	suite.mockEngine = enginemock.NewAuthorizationEngineMock(suite.T())
	suite.service = newAuthorizationService(suite.mockEngine)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_Success() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "user1",
		GroupIDs:             []string{"group1", "group2"},
		RequestedPermissions: []string{"perm1", "perm2", "perm3"},
	}
	expectedPermissions := []string{"perm1", "perm3"}

	suite.mockEngine.On("GetAuthorizedPermissions", mock.Anything, request.UserID, request.GroupIDs, request.RequestedPermissions).
		Return(expectedPermissions, nil)

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(err)
	suite.NotNil(response)
	suite.Equal(expectedPermissions, response.AuthorizedPermissions)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_MissingBothUserAndGroups() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "",
		GroupIDs:             []string{},
		RequestedPermissions: []string{"perm1", "perm2"},
	}

	// Mock engine to return error (validation happens in underlying service)
	suite.mockEngine.On("GetAuthorizedPermissions", mock.Anything, request.UserID, request.GroupIDs, request.RequestedPermissions).
		Return(nil, errors.New("role service error: Either userId or groups must be provided"))

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(response)
	suite.NotNil(err)
	suite.Equal(ErrorAuthorizationFailed.Code, err.Code)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_MissingBothUserAndNilGroups() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "",
		GroupIDs:             nil,
		RequestedPermissions: []string{"perm1", "perm2"},
	}

	// Mock engine to return error (validation happens in underlying service)
	suite.mockEngine.On("GetAuthorizedPermissions", mock.Anything, request.UserID, []string{}, request.RequestedPermissions).
		Return(nil, errors.New("role service error: Either userId or groups must be provided"))

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(response)
	suite.NotNil(err)
	suite.Equal(ErrorAuthorizationFailed.Code, err.Code)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_EmptyRequestedPermissions() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "user1",
		GroupIDs:             []string{"group1"},
		RequestedPermissions: []string{},
	}

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(err)
	suite.NotNil(response)
	suite.Empty(response.AuthorizedPermissions)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_NilRequestedPermissions() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "user1",
		GroupIDs:             []string{"group1"},
		RequestedPermissions: nil,
	}

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(err)
	suite.NotNil(response)
	suite.Empty(response.AuthorizedPermissions)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_UserOnly() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "user1",
		GroupIDs:             []string{},
		RequestedPermissions: []string{"perm1", "perm2"},
	}
	expectedPermissions := []string{"perm1"}

	suite.mockEngine.On("GetAuthorizedPermissions", mock.Anything, request.UserID, request.GroupIDs, request.RequestedPermissions).
		Return(expectedPermissions, nil)

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(err)
	suite.NotNil(response)
	suite.Equal(expectedPermissions, response.AuthorizedPermissions)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_GroupsOnly() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "",
		GroupIDs:             []string{"group1", "group2"},
		RequestedPermissions: []string{"perm1", "perm2"},
	}
	expectedPermissions := []string{"perm2"}

	suite.mockEngine.On("GetAuthorizedPermissions", mock.Anything, request.UserID, request.GroupIDs, request.RequestedPermissions).
		Return(expectedPermissions, nil)

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(err)
	suite.NotNil(response)
	suite.Equal(expectedPermissions, response.AuthorizedPermissions)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_NilGroups() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "user1",
		GroupIDs:             nil,
		RequestedPermissions: []string{"perm1", "perm2"},
	}
	expectedPermissions := []string{"perm1"}

	suite.mockEngine.On("GetAuthorizedPermissions", mock.Anything, request.UserID, []string{}, request.RequestedPermissions).
		Return(expectedPermissions, nil)

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(err)
	suite.NotNil(response)
	suite.Equal(expectedPermissions, response.AuthorizedPermissions)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_EngineError() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "user1",
		GroupIDs:             []string{"group1"},
		RequestedPermissions: []string{"perm1", "perm2"},
	}

	suite.mockEngine.On("GetAuthorizedPermissions", mock.Anything, request.UserID, request.GroupIDs, request.RequestedPermissions).
		Return(nil, errors.New("engine failed"))

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(response)
	suite.NotNil(err)
	suite.Equal(ErrorAuthorizationFailed.Code, err.Code)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_NoAuthorizedPermissions() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "user1",
		GroupIDs:             []string{"group1"},
		RequestedPermissions: []string{"perm1", "perm2"},
	}

	suite.mockEngine.On("GetAuthorizedPermissions", mock.Anything, request.UserID, request.GroupIDs, request.RequestedPermissions).
		Return([]string{}, nil)

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(err)
	suite.NotNil(response)
	suite.Empty(response.AuthorizedPermissions)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizedPermissions_AllPermissionsAuthorized() {
	request := GetAuthorizedPermissionsRequest{
		UserID:               "user1",
		GroupIDs:             []string{"group1"},
		RequestedPermissions: []string{"perm1", "perm2"},
	}

	suite.mockEngine.On("GetAuthorizedPermissions", mock.Anything, request.UserID, request.GroupIDs, request.RequestedPermissions).
		Return(request.RequestedPermissions, nil)

	response, err := suite.service.GetAuthorizedPermissions(context.TODO(), request)

	suite.Nil(err)
	suite.NotNil(response)
	suite.Equal(request.RequestedPermissions, response.AuthorizedPermissions)
}
