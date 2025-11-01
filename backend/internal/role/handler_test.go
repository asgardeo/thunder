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

package role

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/log"
)

type RoleHandlerTestSuite struct {
	suite.Suite
	mockService *RoleServiceInterfaceMock
	handler     *roleHandler
}

func TestRoleHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(RoleHandlerTestSuite))
}

func (suite *RoleHandlerTestSuite) SetupTest() {
	suite.mockService = NewRoleServiceInterfaceMock(suite.T())
	suite.handler = newRoleHandler(suite.mockService)
}

// HandleRoleListRequest Tests
func (suite *RoleHandlerTestSuite) TestHandleRoleListRequest_Success() {
	expectedResponse := &RoleList{
		TotalResults: 2,
		StartIndex:   1,
		Count:        2,
		Roles: []Role{
			{ID: "role1", Name: "Admin"},
			{ID: "role2", Name: "User"},
		},
		Links: []Link{},
	}

	suite.mockService.On("GetRoleList", 10, 0).Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleRoleListRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response RoleListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal(2, response.TotalResults)
	suite.Equal(2, len(response.Roles))
}

func (suite *RoleHandlerTestSuite) TestHandleRoleListRequest_DefaultPagination() {
	expectedResponse := &RoleList{
		TotalResults: 1,
		StartIndex:   1,
		Count:        1,
		Roles:        []Role{{ID: "role1", Name: "Admin"}},
		Links:        []Link{},
	}

	suite.mockService.On("GetRoleList", 30, 0).Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleRoleListRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleListRequest_ServiceError() {
	suite.mockService.On("GetRoleList", 10, 0).Return(nil, &ErrorInvalidLimit)

	req := httptest.NewRequest(http.MethodGet, "/roles?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleRoleListRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

// HandleRolePostRequest Tests
func (suite *RoleHandlerTestSuite) TestHandleRolePostRequest_Success() {
	request := CreateRoleRequest{
		Name:               "Test Role",
		Description:        "Description",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1"},
	}

	expectedRole := &RoleWithPermissionsAndAssignments{
		ID:                 "role1",
		Name:               "Test Role",
		Description:        "Description",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1"},
	}

	suite.mockService.On("CreateRole", mock.AnythingOfType("RoleCreationDetail")).Return(expectedRole, nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.HandleRolePostRequest(w, req)

	suite.Equal(http.StatusCreated, w.Code)

	var response CreateRoleResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("role1", response.ID)
	suite.Equal("Test Role", response.Name)
}

func (suite *RoleHandlerTestSuite) TestHandleRolePostRequest_InvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.HandleRolePostRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRolePostRequest_ServiceError() {
	request := CreateRoleRequest{
		Name:               "Test Role",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1"},
	}

	suite.mockService.On("CreateRole", mock.AnythingOfType("RoleCreationDetail")).
		Return(nil, &ErrorOrganizationUnitNotFound)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.HandleRolePostRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

// HandleRoleGetRequest Tests
func (suite *RoleHandlerTestSuite) TestHandleRoleGetRequest_Success() {
	expectedRole := &RoleWithPermissions{
		ID:                 "role1",
		Name:               "Admin",
		Description:        "Admin role",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1"},
	}

	suite.mockService.On("GetRoleWithPermissions", "role1").Return(expectedRole, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles/role1", nil)
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleGetRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response RoleResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("role1", response.ID)
	suite.Equal("Admin", response.Name)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleGetRequest_MissingID() {
	suite.mockService.On("GetRoleWithPermissions", "").Return(nil, &ErrorMissingRoleID)

	req := httptest.NewRequest(http.MethodGet, "/roles/", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleRoleGetRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleGetRequest_NotFound() {
	suite.mockService.On("GetRoleWithPermissions", "nonexistent").Return(nil, &ErrorRoleNotFound)

	req := httptest.NewRequest(http.MethodGet, "/roles/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleGetRequest(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
}

// HandleRolePutRequest Tests
func (suite *RoleHandlerTestSuite) TestHandleRolePutRequest_Success() {
	request := UpdateRoleRequest{
		Name:               "Updated Role",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1", "perm2"},
	}

	updatedRole := &RoleWithPermissions{
		ID:                 "role1",
		Name:               "Updated Role",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1", "perm2"},
	}

	suite.mockService.On("UpdateRoleWithPermissions", "role1",
		mock.AnythingOfType("RoleUpdateDetail")).Return(updatedRole, nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPut, "/roles/role1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRolePutRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response RoleResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Updated Role", response.Name)
}

func (suite *RoleHandlerTestSuite) TestHandleRolePutRequest_InvalidJSON() {
	req := httptest.NewRequest(http.MethodPut, "/roles/role1", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRolePutRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

// HandleRoleDeleteRequest Tests
func (suite *RoleHandlerTestSuite) TestHandleRoleDeleteRequest_Success() {
	suite.mockService.On("DeleteRole", "role1").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/roles/role1", nil)
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleDeleteRequest(w, req)

	suite.Equal(http.StatusNoContent, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleDeleteRequest_RoleHasAssignments() {
	suite.mockService.On("DeleteRole", "role1").Return(&ErrorCannotDeleteRole)

	req := httptest.NewRequest(http.MethodDelete, "/roles/role1", nil)
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleDeleteRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

// HandleRoleAssignmentsGetRequest Tests
func (suite *RoleHandlerTestSuite) TestHandleRoleAssignmentsGetRequest_Success() {
	expectedResponse := &AssignmentList{
		TotalResults: 2,
		StartIndex:   1,
		Count:        2,
		Assignments: []RoleAssignmentWithDisplay{
			{ID: "user1", Type: AssigneeTypeUser},
			{ID: "group1", Type: AssigneeTypeGroup},
		},
		Links: []Link{},
	}

	suite.mockService.On("GetRoleAssignments", "role1", 10, 0, false).Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles/role1/assignments?limit=10&offset=0", nil)
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleAssignmentsGetRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response AssignmentListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal(2, response.TotalResults)
	suite.Equal(2, len(response.Assignments))
}

func (suite *RoleHandlerTestSuite) TestHandleRoleAssignmentsGetRequest_RoleNotFound() {
	suite.mockService.On("GetRoleAssignments", "nonexistent", 30, 0, false).Return(nil, &ErrorRoleNotFound)

	req := httptest.NewRequest(http.MethodGet, "/roles/nonexistent/assignments", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleAssignmentsGetRequest(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
}

// HandleRoleAddAssignmentsRequest Tests
func (suite *RoleHandlerTestSuite) TestHandleRoleAddAssignmentsRequest_Success() {
	request := AssignmentsRequest{
		Assignments: []AssignmentRequest{
			{ID: "user1", Type: AssigneeTypeUser},
		},
	}

	suite.mockService.On("AddAssignments", "role1", mock.AnythingOfType("[]role.RoleAssignment")).Return(nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles/role1/add-assignments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleAddAssignmentsRequest(w, req)

	suite.Equal(http.StatusNoContent, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleAddAssignmentsRequest_InvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/roles/role1/add-assignments", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleAddAssignmentsRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleAddAssignmentsRequest_ServiceError() {
	request := AssignmentsRequest{
		Assignments: []AssignmentRequest{
			{ID: "invalid_user", Type: AssigneeTypeUser},
		},
	}

	suite.mockService.On("AddAssignments", "role1", mock.AnythingOfType("[]role.RoleAssignment")).
		Return(&ErrorInvalidAssignmentID)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles/role1/add-assignments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleAddAssignmentsRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

// HandleRoleRemoveAssignmentsRequest Tests
func (suite *RoleHandlerTestSuite) TestHandleRoleRemoveAssignmentsRequest_Success() {
	request := AssignmentsRequest{
		Assignments: []AssignmentRequest{
			{ID: "user1", Type: AssigneeTypeUser},
		},
	}

	suite.mockService.On("RemoveAssignments", "role1", mock.AnythingOfType("[]role.RoleAssignment")).Return(nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles/role1/remove-assignments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleRemoveAssignmentsRequest(w, req)

	suite.Equal(http.StatusNoContent, w.Code)
}

// ParsePaginationParams Tests
func (suite *RoleHandlerTestSuite) TestParsePaginationParams() {
	testCases := []struct {
		name           string
		queryString    string
		expectedLimit  int
		expectedOffset int
		expectError    bool
	}{
		{
			name:           "ValidParams",
			queryString:    "limit=20&offset=10",
			expectedLimit:  20,
			expectedOffset: 10,
			expectError:    false,
		},
		{
			name:           "DefaultLimit",
			queryString:    "offset=5",
			expectedLimit:  30,
			expectedOffset: 5,
			expectError:    false,
		},
		{
			name:           "NoParams",
			queryString:    "",
			expectedLimit:  30,
			expectedOffset: 0,
			expectError:    false,
		},
		{
			name:           "InvalidLimit",
			queryString:    "limit=abc",
			expectedLimit:  0,
			expectedOffset: 0,
			expectError:    true,
		},
		{
			name:           "InvalidOffset",
			queryString:    "offset=xyz",
			expectedLimit:  0,
			expectedOffset: 0,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			query, _ := url.ParseQuery(tc.queryString)
			limit, offset, err := parsePaginationParams(query)

			if tc.expectError {
				suite.NotNil(err)
			} else {
				suite.Nil(err)
				suite.Equal(tc.expectedLimit, limit)
				suite.Equal(tc.expectedOffset, offset)
			}
		})
	}
}

// HandleRolePutRequest additional tests
func (suite *RoleHandlerTestSuite) TestHandleRolePutRequest_MissingID() {
	request := UpdateRoleRequest{
		Name:               "Updated Role",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1"},
	}

	suite.mockService.On("UpdateRoleWithPermissions", "", mock.AnythingOfType("RoleUpdateDetail")).
		Return(nil, &ErrorMissingRoleID)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPut, "/roles/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.HandleRolePutRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRolePutRequest_RoleNotFound() {
	request := UpdateRoleRequest{
		Name:               "Updated Role",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1"},
	}

	suite.mockService.On("UpdateRoleWithPermissions", "nonexistent", mock.AnythingOfType("RoleUpdateDetail")).
		Return(nil, &ErrorRoleNotFound)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPut, "/roles/nonexistent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	suite.handler.HandleRolePutRequest(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
}

// HandleRoleDeleteRequest additional tests
func (suite *RoleHandlerTestSuite) TestHandleRoleDeleteRequest_MissingID() {
	suite.mockService.On("DeleteRole", "").Return(&ErrorMissingRoleID)

	req := httptest.NewRequest(http.MethodDelete, "/roles/", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleRoleDeleteRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleDeleteRequest_RoleNotFound() {
	suite.mockService.On("DeleteRole", "nonexistent").Return(&ErrorRoleNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/roles/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleDeleteRequest(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
}

// HandleRoleAssignmentsGetRequest additional tests
func (suite *RoleHandlerTestSuite) TestHandleRoleAssignmentsGetRequest_MissingID() {
	suite.mockService.On("GetRoleAssignments", "", 30, 0, false).Return(nil, &ErrorMissingRoleID)

	req := httptest.NewRequest(http.MethodGet, "/roles//assignments", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleRoleAssignmentsGetRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleAssignmentsGetRequest_InvalidPagination() {
	req := httptest.NewRequest(http.MethodGet, "/roles/role1/assignments?limit=invalid", nil)
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleAssignmentsGetRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

// HandleRoleAddAssignmentsRequest additional tests
func (suite *RoleHandlerTestSuite) TestHandleRoleAddAssignmentsRequest_MissingID() {
	request := AssignmentsRequest{
		Assignments: []AssignmentRequest{
			{ID: "user1", Type: AssigneeTypeUser},
		},
	}

	suite.mockService.On("AddAssignments", "", mock.AnythingOfType("[]role.RoleAssignment")).
		Return(&ErrorMissingRoleID)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles//add-assignments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleAddAssignmentsRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleAddAssignmentsRequest_RoleNotFound() {
	request := AssignmentsRequest{
		Assignments: []AssignmentRequest{
			{ID: "user1", Type: AssigneeTypeUser},
		},
	}

	suite.mockService.On("AddAssignments", "nonexistent", mock.AnythingOfType("[]role.RoleAssignment")).
		Return(&ErrorRoleNotFound)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles/nonexistent/add-assignments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleAddAssignmentsRequest(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
}

// HandleRoleRemoveAssignmentsRequest additional tests
func (suite *RoleHandlerTestSuite) TestHandleRoleRemoveAssignmentsRequest_MissingID() {
	request := AssignmentsRequest{
		Assignments: []AssignmentRequest{
			{ID: "user1", Type: AssigneeTypeUser},
		},
	}

	suite.mockService.On("RemoveAssignments", "", mock.AnythingOfType("[]role.RoleAssignment")).
		Return(&ErrorMissingRoleID)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles//remove-assignments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleRemoveAssignmentsRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleRemoveAssignmentsRequest_InvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/roles/role1/remove-assignments", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleRemoveAssignmentsRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleRemoveAssignmentsRequest_RoleNotFound() {
	request := AssignmentsRequest{
		Assignments: []AssignmentRequest{
			{ID: "user1", Type: AssigneeTypeUser},
		},
	}

	suite.mockService.On("RemoveAssignments", "nonexistent", mock.AnythingOfType("[]role.RoleAssignment")).
		Return(&ErrorRoleNotFound)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles/nonexistent/remove-assignments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleRemoveAssignmentsRequest(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
}

func (suite *RoleHandlerTestSuite) TestHandleRoleRemoveAssignmentsRequest_ServiceError() {
	request := AssignmentsRequest{
		Assignments: []AssignmentRequest{
			{ID: "user1", Type: AssigneeTypeUser},
		},
	}

	suite.mockService.On("RemoveAssignments", "role1", mock.AnythingOfType("[]role.RoleAssignment")).
		Return(&ErrorInvalidAssignmentID)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/roles/role1/remove-assignments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "role1")
	w := httptest.NewRecorder()

	suite.handler.HandleRoleRemoveAssignmentsRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

// Sanitization Tests
func (suite *RoleHandlerTestSuite) TestSanitizeCreateRoleRequest() {
	request := &CreateRoleRequest{
		Name:               "  Test Role  ",
		Description:        "  Description  ",
		OrganizationUnitID: "  ou1  ",
		Permissions:        []string{"  perm1  ", "  perm2  "},
		Assignments: []AssignmentRequest{
			{ID: "  user1  ", Type: AssigneeTypeUser},
		},
	}

	sanitized := suite.handler.sanitizeCreateRoleRequest(request)

	suite.Equal("Test Role", sanitized.Name)
	suite.Equal("Description", sanitized.Description)
	suite.Equal("ou1", sanitized.OrganizationUnitID)
	suite.Equal("perm1", sanitized.Permissions[0])
	suite.Equal("user1", sanitized.Assignments[0].ID)
}

func (suite *RoleHandlerTestSuite) TestSanitizeUpdateRoleRequest() {
	request := &UpdateRoleRequest{
		Name:               "  Updated Name  ",
		OrganizationUnitID: "  ou2  ",
		Permissions:        []string{"  perm3  "},
	}

	sanitized := suite.handler.sanitizeUpdateRoleRequest(request)

	suite.Equal("Updated Name", sanitized.Name)
	suite.Equal("ou2", sanitized.OrganizationUnitID)
	suite.Equal("perm3", sanitized.Permissions[0])
}

func (suite *RoleHandlerTestSuite) TestSanitizeAssignmentsRequest() {
	request := &AssignmentsRequest{
		Assignments: []AssignmentRequest{
			{ID: "  group1  ", Type: AssigneeTypeGroup},
		},
	}

	sanitized := suite.handler.sanitizeAssignmentsRequest(request)

	suite.Equal("group1", sanitized.Assignments[0].ID)
	suite.Equal(AssigneeTypeGroup, sanitized.Assignments[0].Type)
}

func (suite *RoleHandlerTestSuite) TestwriteToResponse_Success() {
	response := &RoleResponse{
		ID:                 "role1",
		Name:               "Role 1",
		Description:        "A sample role",
		OrganizationUnitID: "ou1",
	}

	w := httptest.NewRecorder()
	isErr := writeToResponse(w, response, log.GetLogger())
	suite.False(isErr)
}

func (suite *RoleHandlerTestSuite) TestwriteToResponse_Error() {
	// Use a function which cannot be marshaled to JSON to cause encoding error
	response := func() {}
	w := httptest.NewRecorder()
	isErr := writeToResponse(w, response, log.GetLogger())
	suite.True(isErr)
}
