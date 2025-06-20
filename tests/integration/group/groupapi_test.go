/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
)

var (
	testOU = "456e8400-e29b-41d4-a716-446655440001"

	groupToCreate = CreateGroupRequest{
		Name: "Test Group",
		Parent: Parent{
			Type: ParentTypeOrganizationUnit,
			Id:   testOU,
		},
		Users: []string{"550e8400-e29b-41d4-a716-446655440000"},
	}
)

var createdGroupID string

type GroupAPITestSuite struct {
	suite.Suite
}

func (suite *GroupAPITestSuite) SetupSuite() {
	id, err := createGroup(suite)
	if err != nil {
		suite.T().Fatalf("Failed to create group during setup: %v", err)
	} else {
		createdGroupID = id
	}
}

func (suite *GroupAPITestSuite) TearDownSuite() {
	if createdGroupID != "" {
		err := deleteGroup(createdGroupID)
		if err != nil {
			suite.T().Fatalf("Failed to delete group during teardown: %v", err)
		}
	}
}

func (suite *GroupAPITestSuite) TestGetGroup() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for retrieval")
	}

	// Get the created group
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups/"+createdGroupID, nil)
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
	if err != nil {
		suite.T().Fatalf("Failed to read response body: %v", err)
	}

	var retrievedGroup Group
	err = json.Unmarshal(body, &retrievedGroup)
	suite.Require().NoError(err)

	// Verify the retrieved group
	createdGroup := buildCreatedGroup()
	suite.Equal(createdGroup.Id, retrievedGroup.Id)
	suite.Equal(createdGroup.Name, retrievedGroup.Name)
	suite.Equal(createdGroup.Parent.Type, retrievedGroup.Parent.Type)
	suite.Equal(createdGroup.Parent.Id, retrievedGroup.Parent.Id)
}

func (suite *GroupAPITestSuite) TestListGroups() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available, group creation failed in setup")
	}

	// List groups
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.T().Fatalf("Failed to read response body: %v", err)
	}

	var groups []GroupBasic
	err = json.Unmarshal(body, &groups)
	suite.Require().NoError(err)

	// Verify the list contains our created group
	found := false
	createdGroup := buildCreatedGroup()
	for _, group := range groups {
		if group.Id == createdGroup.Id {
			found = true
			suite.Equal(createdGroup.Name, group.Name)
			break
		}
	}
	suite.True(found, "Created group should be in the list")
}

func (suite *GroupAPITestSuite) TestUpdateGroup() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for update")
	}

	// Update the group
	updateRequest := map[string]interface{}{
		"name": "Updated Test Group",
		"parent": map[string]string{
			"type": string(ParentTypeOrganizationUnit),
			"id":   testOU,
		},
		"users":  []string{},
		"groups": []string{},
	}

	jsonData, err := json.Marshal(updateRequest)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("PUT", testServerURL+"/groups/"+createdGroupID, bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var updatedGroup Group
	err = json.Unmarshal(body, &updatedGroup)
	suite.Require().NoError(err)

	// Verify the update
	suite.Equal(createdGroupID, updatedGroup.Id)
	suite.Equal("Updated Test Group", updatedGroup.Name)
}

func (suite *GroupAPITestSuite) TestDeleteGroup() {
	// Create a temporary group for this test since we don't want to delete the main test group
	tempGroupToCreate := CreateGroupRequest{
		Name: "Temp Test Group",
		Parent: Parent{
			Type: ParentTypeOrganizationUnit,
			Id:   testOU,
		},
		Users: []string{},
	}

	jsonData, err := json.Marshal(tempGroupToCreate)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Create temporary group
	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	var tempGroup Group
	err = json.NewDecoder(resp.Body).Decode(&tempGroup)
	suite.Require().NoError(err)

	// Delete the temporary group
	req, err = http.NewRequest("DELETE", testServerURL+"/groups/"+tempGroup.Id, nil)
	suite.Require().NoError(err)

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify the group is deleted by trying to get it
	getReq, err := http.NewRequest("GET", testServerURL+"/groups/"+tempGroup.Id, nil)
	suite.Require().NoError(err)

	getResp, err := client.Do(getReq)
	if err != nil {
		suite.T().Fatalf("Failed to execute GET request: %v", err)
	}
	defer getResp.Body.Close()

	suite.Equal(http.StatusNotFound, getResp.StatusCode)
}

func (suite *GroupAPITestSuite) TestGetNonExistentGroup() {
	// Try to get a non-existent group
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups/non-existent-id", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	if err != nil {
		suite.T().Fatalf("Failed to execute GET request: %v", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *GroupAPITestSuite) TestCreateGroupWithInvalidData() {
	// Try to create a group with invalid data (missing name)
	invalidGroup := map[string]interface{}{
		"parent": map[string]string{
			"type": string(ParentTypeOrganizationUnit),
			"id":   testOU,
		},
	}

	jsonData, err := json.Marshal(invalidGroup)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		suite.T().Fatalf("Failed to execute POST request: %v", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func createGroup(ts *GroupAPITestSuite) (string, error) {
	jsonData, err := json.Marshal(groupToCreate)
	if err != nil {
		return "", fmt.Errorf("failed to marshal groupToCreate: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdGroup Group
	err = json.NewDecoder(resp.Body).Decode(&createdGroup)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	return createdGroup.Id, nil
}

func deleteGroup(groupID string) error {
	req, err := http.NewRequest("DELETE", testServerURL+"/groups/"+groupID, nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("expected status 204, got %d", resp.StatusCode)
	}
	return nil
}

func buildCreatedGroup() Group {
	return Group{
		GroupBasic: GroupBasic{
			Id:   createdGroupID,
			Name: groupToCreate.Name,
			Parent: Parent{
				Type: groupToCreate.Parent.Type,
				Id:   groupToCreate.Parent.Id,
			},
		},
		Users: groupToCreate.Users,
	}
}

func TestGroupAPITestSuite(t *testing.T) {
	suite.Run(t, new(GroupAPITestSuite))
}
