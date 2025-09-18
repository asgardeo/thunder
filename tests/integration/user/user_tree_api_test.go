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

package user

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

var (
	pathTestOU   OrganizationUnit
	pathTestOUID string
)

// CreateOURequest represents the request body for creating an organization unit.
type CreateOURequest struct {
	Handle      string  `json:"handle"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent,omitempty"`
}

// OrganizationUnit represents an organization unit.
type OrganizationUnit struct {
	ID          string  `json:"id"`
	Handle      string  `json:"handle"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent"`
}

// CreateUserByPathRequest represents the request body for creating a user by path.
type CreateUserByPathRequest struct {
	Type       string          `json:"type"`
	Groups     []string        `json:"groups,omitempty"`
	Attributes json.RawMessage `json:"attributes,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

type UserTreeAPITestSuite struct {
	suite.Suite
}

func TestUserTreeAPITestSuite(t *testing.T) {
	suite.Run(t, new(UserTreeAPITestSuite))
}

func (suite *UserTreeAPITestSuite) SetupSuite() {
	// Create a test organization unit for path-based tests
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	ouRequest := CreateOURequest{
		Handle:      "test-ou-for-users",
		Name:        "Test OU for Users",
		Description: "Test organization unit for user path-based operations",
	}

	ouJSON, err := json.Marshal(ouRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+"/organization-units", bytes.NewBuffer(ouJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	err = json.Unmarshal(body, &pathTestOU)
	suite.Require().NoError(err)

	pathTestOUID = pathTestOU.ID
	suite.T().Logf("Created test OU with ID: %s and handle: %s", pathTestOUID, pathTestOU.Handle)
}

func (suite *UserTreeAPITestSuite) TearDownSuite() {
	// Clean up the test organization unit
	if pathTestOUID != "" {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		req, err := http.NewRequest("DELETE", testServerURL+"/organization-units/"+pathTestOUID, nil)
		if err != nil {
			suite.T().Logf("Failed to create delete request: %v", err)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			suite.T().Logf("Failed to delete test OU: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			suite.T().Logf("Failed to delete test OU, status: %d", resp.StatusCode)
		}
	}
}

// TestGetUsersByPath tests retrieving users by organization unit handle path
func (suite *UserTreeAPITestSuite) TestGetUsersByPath() {
	if pathTestOUID == "" {
		suite.T().Fatal("OU ID is not available for path-based user retrieval")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/users/tree/"+pathTestOU.Handle, nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Failed to close response body: %v", err)
		}
	}()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Failed to read response body: %v", err)

	var userListResponse testutils.UserListResponse
	err = json.Unmarshal(body, &userListResponse)
	suite.Require().NoError(err)

	// Verify the response structure
	suite.GreaterOrEqual(userListResponse.TotalResults, 0)
	suite.Equal(userListResponse.StartIndex, 1)
	suite.Equal(userListResponse.Count, len(userListResponse.Users))
}

// TestCreateUserByPath tests creating a user by organization unit handle path
func (suite *UserTreeAPITestSuite) TestCreateUserByPath() {
	if pathTestOUID == "" {
		suite.T().Fatal("OU ID is not available for path-based user creation")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	createRequest := CreateUserByPathRequest{
		Type:       "employee",
		Attributes: json.RawMessage(`{"username": "test.user", "email": "test.user@example.com", "department": "Engineering"}`),
	}

	requestJSON, err := json.Marshal(createRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+"/users/tree/"+pathTestOU.Handle, bytes.NewBuffer(requestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Failed to close response body: %v", err)
		}
	}()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var createdUser testutils.User
	err = json.Unmarshal(body, &createdUser)
	suite.Require().NoError(err)

	// Verify the created user
	suite.NotEmpty(createdUser.ID)
	suite.Equal(pathTestOUID, createdUser.OrganizationUnit)
	suite.Equal("employee", createdUser.Type)
	suite.NotEmpty(createdUser.Attributes)

	// Clean up: delete the created user
	deleteReq, err := http.NewRequest("DELETE", testServerURL+"/users/"+createdUser.ID, nil)
	if err != nil {
		suite.T().Logf("Failed to create delete request for user: %v", err)
		return
	}

	deleteResp, err := client.Do(deleteReq)
	if err != nil {
		suite.T().Logf("Failed to delete created user: %v", err)
		return
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusNoContent {
		suite.T().Logf("Failed to delete created user, status: %d", deleteResp.StatusCode)
	}
}

// TestGetUsersByInvalidPath tests retrieving users by invalid organization unit handle path
func (suite *UserTreeAPITestSuite) TestGetUsersByInvalidPath() {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/users/tree/nonexistent-ou", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Failed to close response body: %v", err)
		}
	}()

	suite.Equal(http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("USR-1005", errorResp.Code)
	suite.Equal("Organization unit not found", errorResp.Message)
}

// TestGetUsersByPathWithPagination tests retrieving users by path with pagination parameters
func (suite *UserTreeAPITestSuite) TestGetUsersByPathWithPagination() {
	if pathTestOUID == "" {
		suite.T().Fatal("OU ID is not available for pagination test")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/users/tree/"+pathTestOU.Handle+"?limit=5&offset=0", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Failed to close response body: %v", err)
		}
	}()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var userListResponse testutils.UserListResponse
	err = json.Unmarshal(body, &userListResponse)
	suite.Require().NoError(err)

	// Verify pagination parameters
	suite.GreaterOrEqual(userListResponse.TotalResults, 0)
	suite.Equal(userListResponse.StartIndex, 1)
	suite.LessOrEqual(userListResponse.Count, 5)
}
