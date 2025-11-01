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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
	rolesBasePath = "/roles"
)

var (
	testOU = testutils.OrganizationUnit{
		Handle:      "test-role-ou",
		Name:        "Test Organization Unit for Roles",
		Description: "Organization unit created for role API testing",
		Parent:      nil,
	}

	testUser1 = testutils.User{
		Type: "person",
		Attributes: json.RawMessage(`{
			"email": "roleuser1@example.com",
			"firstName": "Role",
			"lastName": "User1",
			"password": "TestPassword123!"
		}`),
	}

	testUser2 = testutils.User{
		Type: "person",
		Attributes: json.RawMessage(`{
			"email": "roleuser2@example.com",
			"firstName": "Role",
			"lastName": "User2",
			"password": "TestPassword123!"
		}`),
	}

	testGroup = testutils.Group{
		Name:        "Test Role Group",
		Description: "Group created for role API testing",
	}

	testRole = CreateRoleRequest{
		Name:        "Test Admin Role",
		Description: "Admin role for testing",
		Permissions: []string{"read:users", "write:users", "delete:users"},
	}
)

var (
	testOUID     string
	testUserID1  string
	testUserID2  string
	testGroupID  string
	sharedRoleID string // Shared role created in SetupSuite for tests that need a pre-existing role
)

type RoleAPITestSuite struct {
	suite.Suite
	client *http.Client
}

func TestRoleAPITestSuite(t *testing.T) {
	suite.Run(t, new(RoleAPITestSuite))
}

func (suite *RoleAPITestSuite) SetupSuite() {
	// Create HTTP client that skips TLS verification for testing
	suite.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Create test organization unit
	ouID, err := testutils.CreateOrganizationUnit(testOU)
	suite.Require().NoError(err, "Failed to create test organization unit")
	testOUID = ouID

	// Create test users
	user1 := testUser1
	user1.OrganizationUnit = testOUID
	userID1, err := testutils.CreateUser(user1)
	suite.Require().NoError(err, "Failed to create test user 1")
	testUserID1 = userID1

	user2 := testUser2
	user2.OrganizationUnit = testOUID
	userID2, err := testutils.CreateUser(user2)
	suite.Require().NoError(err, "Failed to create test user 2")
	testUserID2 = userID2

	// Create test group
	groupToCreate := testGroup
	groupToCreate.OrganizationUnitId = testOUID
	groupID, err := testutils.CreateGroup(groupToCreate)
	suite.Require().NoError(err, "Failed to create test group")
	testGroupID = groupID

	// Create a shared role that can be used by multiple tests
	sharedRole := testRole
	sharedRole.OrganizationUnitID = testOUID
	role, err := suite.createRole(sharedRole)
	suite.Require().NoError(err, "Failed to create shared role")
	sharedRoleID = role.ID
}

func (suite *RoleAPITestSuite) TearDownSuite() {
	// Cleanup in reverse order
	if sharedRoleID != "" {
		_ = suite.deleteRole(sharedRoleID)
	}
	if testGroupID != "" {
		_ = testutils.DeleteGroup(testGroupID)
	}
	if testUserID2 != "" {
		_ = testutils.DeleteUser(testUserID2)
	}
	if testUserID1 != "" {
		_ = testutils.DeleteUser(testUserID1)
	}
	if testOUID != "" {
		_ = testutils.DeleteOrganizationUnit(testOUID)
	}
}

// Test 1: Create Role
func (suite *RoleAPITestSuite) TestCreateRole_Success() {
	roleRequest := CreateRoleRequest{
		Name:               "Test Create Role Success",
		Description:        "Test role created in TestCreateRole_Success",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data", "write:data"},
	}

	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	suite.Require().NotNil(role)

	suite.NotEmpty(role.ID)
	suite.Equal(roleRequest.Name, role.Name)
	suite.Equal(roleRequest.Description, role.Description)
	suite.Equal(roleRequest.OrganizationUnitID, role.OrganizationUnitID)
	suite.Equal(len(roleRequest.Permissions), len(role.Permissions))

	// Cleanup
	_ = suite.deleteRole(role.ID)
}

// Test 2: Create Role with Assignments
func (suite *RoleAPITestSuite) TestCreateRole_WithAssignments() {
	roleRequest := CreateRoleRequest{
		Name:               "Test Role With Assignments",
		Description:        "Role with initial assignments",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data"},
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
		},
	}

	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	suite.Require().NotNil(role)

	suite.Equal(1, len(role.Assignments))
	suite.Equal(testUserID1, role.Assignments[0].ID)
	suite.Equal(AssigneeTypeUser, role.Assignments[0].Type)

	// Cleanup
	_ = suite.deleteRole(role.ID)
}

// Test 3: Create Role without Permissions
func (suite *RoleAPITestSuite) TestCreateRole_WithoutPermissions() {
	roleRequest := CreateRoleRequest{
		Name:               "Test Role Without Permissions",
		Description:        "Role without permissions",
		OrganizationUnitID: testOUID,
		Permissions:        []string{},
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
		},
	}

	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	suite.Require().NotNil(role)

	suite.Equal(1, len(role.Assignments))
	suite.Equal(testUserID1, role.Assignments[0].ID)
	suite.Equal(AssigneeTypeUser, role.Assignments[0].Type)

	// Cleanup
	_ = suite.deleteRole(role.ID)
}

// Test 4: Create Role - Validation Errors
func (suite *RoleAPITestSuite) TestCreateRole_ValidationErrors() {
	testCases := []struct {
		name        string
		roleRequest CreateRoleRequest
		expectedErr string
	}{
		{
			name: "Missing Name",
			roleRequest: CreateRoleRequest{
				OrganizationUnitID: testOUID,
				Permissions:        []string{"perm1"},
			},
			expectedErr: "ROL-1001",
		},
		{
			name: "Missing OrganizationUnitID",
			roleRequest: CreateRoleRequest{
				Name:        "Test Role",
				Permissions: []string{"perm1"},
			},
			expectedErr: "ROL-1001",
		},
		{
			name: "Invalid Organization Unit",
			roleRequest: CreateRoleRequest{
				Name:               "Test Role",
				OrganizationUnitID: "nonexistent-ou",
				Permissions:        []string{"perm1"},
			},
			expectedErr: "ROL-1005",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			role, err := suite.createRole(tc.roleRequest)
			suite.Error(err)
			suite.Nil(role)
			suite.Contains(err.Error(), tc.expectedErr)
		})
	}
}

// Test 5: Get Role
func (suite *RoleAPITestSuite) TestGetRole_Success() {
	suite.Require().NotEmpty(sharedRoleID, "Shared role must be created in SetupSuite")

	role, err := suite.getRole(sharedRoleID)
	suite.Require().NoError(err)
	suite.Require().NotNil(role)

	suite.Equal(sharedRoleID, role.ID)
	suite.Equal(testRole.Name, role.Name)
	suite.Equal(testRole.Description, role.Description)
}

// Test 6: Get Role - Not Found
func (suite *RoleAPITestSuite) TestGetRole_NotFound() {
	role, err := suite.getRole("nonexistent-role-id")
	suite.Error(err)
	suite.Nil(role)
	suite.Contains(err.Error(), "ROL-1003")
}

// Test 7: List Roles
func (suite *RoleAPITestSuite) TestListRoles_Success() {
	suite.Require().NotEmpty(sharedRoleID, "Shared role must be created in SetupSuite")

	response, err := suite.listRoles(0, 30)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	suite.GreaterOrEqual(response.TotalResults, 1)
	suite.GreaterOrEqual(response.Count, 1)
	suite.NotEmpty(response.Roles)

	// Verify our shared role is in the list
	found := false
	for _, role := range response.Roles {
		if role.ID == sharedRoleID {
			found = true
			suite.Equal(testRole.Name, role.Name)
			break
		}
	}
	suite.True(found, "Shared role should be in the list")
}

// Test 8: List Roles - Pagination
func (suite *RoleAPITestSuite) TestListRoles_Pagination() {
	// Create additional roles for pagination testing
	role1Request := CreateRoleRequest{
		Name:               "Pagination Test Role 1",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"perm1"},
	}
	role2Request := CreateRoleRequest{
		Name:               "Pagination Test Role 2",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"perm2"},
	}

	role1, err := suite.createRole(role1Request)
	suite.Require().NoError(err)
	defer suite.deleteRole(role1.ID)

	role2, err := suite.createRole(role2Request)
	suite.Require().NoError(err)
	defer suite.deleteRole(role2.ID)

	// Test pagination with limit
	response, err := suite.listRoles(0, 2)
	suite.Require().NoError(err)
	suite.LessOrEqual(response.Count, 2)

	// Test with offset
	response2, err := suite.listRoles(1, 2)
	suite.Require().NoError(err)
	suite.NotNil(response2)
}

// Test 9: Update Role
func (suite *RoleAPITestSuite) TestUpdateRole_Success() {
	suite.Require().NotEmpty(sharedRoleID, "Shared role must be created in SetupSuite")

	updateRequest := UpdateRoleRequest{
		Name:               "Updated Admin Role",
		Description:        "Updated description",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:users", "write:users", "delete:users", "admin:all"},
	}

	role, err := suite.updateRole(sharedRoleID, updateRequest)
	suite.Require().NoError(err)
	suite.Require().NotNil(role)

	suite.Equal(sharedRoleID, role.ID)
	suite.Equal(updateRequest.Name, role.Name)
	suite.Equal(updateRequest.Description, role.Description)
	suite.Equal(4, len(role.Permissions))
}

// Test 10: Update Role - Not Found
func (suite *RoleAPITestSuite) TestUpdateRole_NotFound() {
	updateRequest := UpdateRoleRequest{
		Name:               "Updated Role",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"perm1"},
	}

	role, err := suite.updateRole("nonexistent-role-id", updateRequest)
	suite.Error(err)
	suite.Nil(role)
	suite.Contains(err.Error(), "ROL-1003")
}

// Test 11: Add Assignments - User
func (suite *RoleAPITestSuite) TestAddAssignments_User() {
	// Create a role for this test
	roleRequest := CreateRoleRequest{
		Name:               "Test Role for User Assignment",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data"},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	defer suite.deleteRole(role.ID)

	assignmentsRequest := AssignmentsRequest{
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
		},
	}

	err = suite.addAssignments(role.ID, assignmentsRequest)
	suite.Require().NoError(err)

	// Verify assignments were added
	assignments, err := suite.getRoleAssignments(role.ID, 0, 30)
	suite.Require().NoError(err)
	suite.Equal(1, assignments.TotalResults)
	suite.Equal(testUserID1, assignments.Assignments[0].ID)
	suite.Equal(AssigneeTypeUser, assignments.Assignments[0].Type)
}

// Test 12: Add Assignments - Group
func (suite *RoleAPITestSuite) TestAddAssignments_Group() {
	// Create a role for this test
	roleRequest := CreateRoleRequest{
		Name:               "Test Role for Group Assignment",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data"},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	defer suite.deleteRole(role.ID)

	assignmentsRequest := AssignmentsRequest{
		Assignments: []Assignment{
			{ID: testGroupID, Type: AssigneeTypeGroup},
		},
	}

	err = suite.addAssignments(role.ID, assignmentsRequest)
	suite.Require().NoError(err)

	// Verify assignments
	assignments, err := suite.getRoleAssignments(role.ID, 0, 30)
	suite.Require().NoError(err)
	suite.Equal(1, assignments.TotalResults) // Group only

	// Check group assignment exists
	groupFound := false
	for _, assignment := range assignments.Assignments {
		if assignment.ID == testGroupID && assignment.Type == AssigneeTypeGroup {
			groupFound = true
			break
		}
	}
	suite.True(groupFound, "Group assignment should exist")
}

// Test 13: Add Assignments - Multiple
func (suite *RoleAPITestSuite) TestAddAssignments_Multiple() {
	// Create a new role for this test
	roleRequest := CreateRoleRequest{
		Name:               "Multi Assignment Role",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data"},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	defer suite.deleteRole(role.ID)

	assignmentsRequest := AssignmentsRequest{
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
			{ID: testUserID2, Type: AssigneeTypeUser},
			{ID: testGroupID, Type: AssigneeTypeGroup},
		},
	}

	err = suite.addAssignments(role.ID, assignmentsRequest)
	suite.Require().NoError(err)

	// Verify all assignments
	assignments, err := suite.getRoleAssignments(role.ID, 0, 30)
	suite.Require().NoError(err)
	suite.Equal(3, assignments.TotalResults)
}

// Test 14: Add Assignments - Invalid User
func (suite *RoleAPITestSuite) TestAddAssignments_InvalidUser() {
	// Create a role for this test
	roleRequest := CreateRoleRequest{
		Name:               "Test Role for Invalid Assignment",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data"},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	defer suite.deleteRole(role.ID)

	assignmentsRequest := AssignmentsRequest{
		Assignments: []Assignment{
			{ID: "nonexistent-user-id", Type: AssigneeTypeUser},
		},
	}

	err = suite.addAssignments(role.ID, assignmentsRequest)
	suite.Error(err)
	suite.Contains(err.Error(), "ROL-1007")
}

// Test 15: Get Role Assignments
func (suite *RoleAPITestSuite) TestGetRoleAssignments_Success() {
	// Create a role with an assignment for this test
	roleRequest := CreateRoleRequest{
		Name:               "Test Role for Get Assignments",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data"},
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
		},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	defer suite.deleteRole(role.ID)

	assignments, err := suite.getRoleAssignments(role.ID, 0, 30)
	suite.Require().NoError(err)
	suite.Require().NotNil(assignments)
	suite.GreaterOrEqual(assignments.TotalResults, 0)
}

// Test 16: Get Role Assignments - Pagination
func (suite *RoleAPITestSuite) TestGetRoleAssignments_Pagination() {
	// Create a role with multiple assignments for pagination testing
	roleRequest := CreateRoleRequest{
		Name:               "Test Role for Pagination",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data"},
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
			{ID: testUserID2, Type: AssigneeTypeUser},
		},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	defer suite.deleteRole(role.ID)

	// Test with small page size
	assignments, err := suite.getRoleAssignments(role.ID, 0, 1)
	suite.Require().NoError(err)
	suite.LessOrEqual(assignments.Count, 1)

	// Test with offset
	if assignments.TotalResults > 1 {
		assignments2, err := suite.getRoleAssignments(role.ID, 1, 1)
		suite.Require().NoError(err)
		suite.NotNil(assignments2)
	}
}

// Test 17: Remove Assignments
func (suite *RoleAPITestSuite) TestRemoveAssignments_Success() {
	// Create a role with assignments for this test
	roleRequest := CreateRoleRequest{
		Name:               "Test Role for Remove Assignments",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data"},
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
			{ID: testUserID2, Type: AssigneeTypeUser},
		},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	defer suite.deleteRole(role.ID)

	// Get current assignments
	beforeAssignments, err := suite.getRoleAssignments(role.ID, 0, 30)
	suite.Require().NoError(err)
	initialCount := beforeAssignments.TotalResults

	suite.Require().Greater(initialCount, 0, "Should have assignments to remove")

	// Remove first assignment
	assignmentToRemove := beforeAssignments.Assignments[0]
	removeRequest := AssignmentsRequest{
		Assignments: []Assignment{assignmentToRemove},
	}

	err = suite.removeAssignments(role.ID, removeRequest)
	suite.Require().NoError(err)

	// Verify assignment was removed
	afterAssignments, err := suite.getRoleAssignments(role.ID, 0, 30)
	suite.Require().NoError(err)
	suite.Equal(initialCount-1, afterAssignments.TotalResults)
}

// Test 18: Delete Role with Assignments
func (suite *RoleAPITestSuite) TestDeleteRole_WithAssignments() {
	// Create a role with assignments
	roleRequest := CreateRoleRequest{
		Name:               "Role to Delete with Assignments",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"perm1"},
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
		},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)

	// Try to delete - should fail because it has assignments
	err = suite.deleteRole(role.ID)
	suite.Require().Error(err, "Delete should fail when role has assignments")
	suite.Contains(err.Error(), "ROL-1006", "Should return cannot delete role error")

	// Remove assignments first
	removeRequest := AssignmentsRequest{
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
		},
	}
	err = suite.removeAssignments(role.ID, removeRequest)
	suite.Require().NoError(err)

	// Now delete should succeed
	err = suite.deleteRole(role.ID)
	suite.NoError(err)
}

// Test 19: Delete Role - Success
func (suite *RoleAPITestSuite) TestDeleteRole_Success() {
	// Create a role without assignments
	roleRequest := CreateRoleRequest{
		Name:               "Role to Delete",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"perm1"},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)

	// Delete the role
	err = suite.deleteRole(role.ID)
	suite.NoError(err)

	// Verify role is deleted
	deletedRole, err := suite.getRole(role.ID)
	suite.Error(err)
	suite.Nil(deletedRole)
	suite.Contains(err.Error(), "ROL-1003")
}

// Test 20: Delete Role - Not Found (Should return success for idempotency)
func (suite *RoleAPITestSuite) TestDeleteRole_NotFound() {
	err := suite.deleteRole("nonexistent-role-id")
	// As per service implementation, delete returns nil for non-existent roles
	suite.NoError(err)
}

// Test 21: Get Role Assignments with Display Names
func (suite *RoleAPITestSuite) TestGetRoleAssignments_WithDisplay() {
	// Create a role with both user and group assignments
	roleRequest := CreateRoleRequest{
		Name:               "Test Role for Display Names",
		OrganizationUnitID: testOUID,
		Permissions:        []string{"read:data"},
		Assignments: []Assignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
			{ID: testGroupID, Type: AssigneeTypeGroup},
		},
	}
	role, err := suite.createRole(roleRequest)
	suite.Require().NoError(err)
	defer suite.deleteRole(role.ID)

	// Get assignments without display parameter
	assignmentsWithoutDisplay, err := suite.getRoleAssignmentsWithInclude(role.ID, 0, 30, "")
	suite.Require().NoError(err)
	suite.Require().NotNil(assignmentsWithoutDisplay)
	suite.Equal(2, assignmentsWithoutDisplay.TotalResults)

	// Verify display names are not included
	for _, assignment := range assignmentsWithoutDisplay.Assignments {
		suite.Empty(assignment.Display, "Display field should be empty without include=display parameter")
	}

	// Get assignments with include=display parameter
	assignmentsWithDisplay, err := suite.getRoleAssignmentsWithInclude(role.ID, 0, 30, "display")
	suite.Require().NoError(err)
	suite.Require().NotNil(assignmentsWithDisplay)
	suite.Equal(2, assignmentsWithDisplay.TotalResults)

	// Verify display names are included
	userFound := false
	groupFound := false
	for _, assignment := range assignmentsWithDisplay.Assignments {
		suite.NotEmpty(assignment.Display, "Display field should be populated with include=display parameter")

		if assignment.Type == AssigneeTypeUser && assignment.ID == testUserID1 {
			userFound = true
			// Display name for user should be the user ID (as per implementation)
			suite.Equal(testUserID1, assignment.Display)
		}

		if assignment.Type == AssigneeTypeGroup && assignment.ID == testGroupID {
			groupFound = true
			// Display name for group should be the group name
			suite.Equal(testGroup.Name, assignment.Display)
		}
	}

	suite.True(userFound, "User assignment should be found")
	suite.True(groupFound, "Group assignment should be found")
}

// Helper methods

func (suite *RoleAPITestSuite) createRole(request CreateRoleRequest) (*Role, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", testServerURL+rolesBasePath, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("failed to create role: %s - %s", errResp.Code, errResp.Message)
	}

	var role Role
	if err := json.Unmarshal(respBody, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

func (suite *RoleAPITestSuite) getRole(roleID string) (*Role, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testServerURL, rolesBasePath, roleID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("failed to get role: %s - %s", errResp.Code, errResp.Message)
	}

	var role Role
	if err := json.Unmarshal(respBody, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

func (suite *RoleAPITestSuite) listRoles(offset, limit int) (*RoleListResponse, error) {
	url := fmt.Sprintf("%s%s?offset=%d&limit=%d", testServerURL, rolesBasePath, offset, limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("failed to list roles: %s - %s", errResp.Code, errResp.Message)
	}

	var response RoleListResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (suite *RoleAPITestSuite) updateRole(roleID string, request UpdateRoleRequest) (*Role, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s%s/%s", testServerURL, rolesBasePath, roleID),
		bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("failed to update role: %s - %s", errResp.Code, errResp.Message)
	}

	var role Role
	if err := json.Unmarshal(respBody, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

func (suite *RoleAPITestSuite) deleteRole(roleID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s/%s", testServerURL, rolesBasePath, roleID), nil)
	if err != nil {
		return err
	}

	resp, err := suite.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return fmt.Errorf("failed to delete role: %s - %s", errResp.Code, errResp.Message)
	}

	return nil
}

func (suite *RoleAPITestSuite) addAssignments(roleID string, request AssignmentsRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s/%s/assignments/add", testServerURL, rolesBasePath, roleID),
		bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return fmt.Errorf("failed to add assignments: %s - %s", errResp.Code, errResp.Message)
	}

	return nil
}

func (suite *RoleAPITestSuite) removeAssignments(roleID string, request AssignmentsRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s/%s/assignments/remove", testServerURL, rolesBasePath, roleID),
		bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return fmt.Errorf("failed to remove assignments: %s - %s", errResp.Code, errResp.Message)
	}

	return nil
}

func (suite *RoleAPITestSuite) getRoleAssignments(roleID string, offset, limit int) (*AssignmentListResponse, error) {
	return suite.getRoleAssignmentsWithInclude(roleID, offset, limit, "")
}

func (suite *RoleAPITestSuite) getRoleAssignmentsWithInclude(roleID string, offset, limit int,
	include string) (*AssignmentListResponse, error) {
	url := fmt.Sprintf("%s%s/%s/assignments?offset=%d&limit=%d", testServerURL, rolesBasePath, roleID, offset, limit)
	if include != "" {
		url = fmt.Sprintf("%s&include=%s", url, include)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("failed to get role assignments: %s - %s", errResp.Code, errResp.Message)
	}

	var response AssignmentListResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
