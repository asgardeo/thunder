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

package user

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
	preCreatedUser User = User{
		Id:         "550e8400-e29b-41d4-a716-446655440000",
		OrgId:      "456e8400-e29b-41d4-a716-446655440001",
		Type:       "person",
		Attributes: json.RawMessage(`{"age": 30, "roles": ["admin", "user"], "address": {"city": "Colombo", "zip": "00100"}}`),
	}

	userToCreate = User{
		OrgId:      "456e8400-e29b-41d4-a716-446655440001",
		Type:       "person",
		Attributes: json.RawMessage(`{"age": 25, "roles": ["viewer"], "address": {"city": "Seattle", "zip": "98101"}}`),
	}

	userToUpdate = User{
		OrgId:      "456e8400-e29b-41d4-a716-446655440001",
		Type:       "person",
		Attributes: json.RawMessage(`{"age": 35, "roles": ["admin"], "address": {"city": "Colombo", "zip": "10300"}}`),
	}
)

var createdUserID string

type UserAPITestSuite struct {
	suite.Suite
}

func TestUserAPITestSuite(t *testing.T) {

	suite.Run(t, new(UserAPITestSuite))
}

// SetupSuite test user creation
func (ts *UserAPITestSuite) SetupSuite() {

	id, err := createUser(ts)
	if err != nil {
		ts.T().Fatalf("Failed to create user during setup: %v", err)
	} else {
		createdUserID = id
	}
}

// TearDownSuite test user deletion
func (ts *UserAPITestSuite) TearDownSuite() {

	if createdUserID != "" {
		err := deleteUser(createdUserID)
		if err != nil {
			ts.T().Fatalf("Failed to delete user during teardown: %v", err)
		}
	}
}

// Test user listing
func (ts *UserAPITestSuite) TestUserListing() {

	req, err := http.NewRequest("GET", testServerURL+"/users", nil)
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}

	// Configure the HTTP client to skip TLS verification
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Skip certificate verification
		},
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Validate the response
	if resp.StatusCode != http.StatusOK {
		ts.T().Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Parse the response body
	var users []User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	userListLength := len(users)
	if userListLength == 0 {
		ts.T().Fatalf("Response does not contain any users")
	}

	if userListLength != 2 {
		ts.T().Fatalf("Expected 2 users, got %d", userListLength)
	}

	user1 := users[0]
	if !user1.equals(preCreatedUser) {
		ts.T().Fatalf("User mismatch, expected %+v, got %+v", preCreatedUser, user1)
	}

	user2 := users[1]
	createdUser := buildCreatedUser()
	if !user2.equals(createdUser) {
		ts.T().Fatalf("User mismatch, expected %+v, got %+v", createdUser, user2)
	}
}

// Test user get by ID
func (ts *UserAPITestSuite) TestUserGetByID() {

	if createdUserID == "" {
		ts.T().Fatal("user ID is not available for retrieval")
	}
	user := buildCreatedUser()
	retrieveAndValidateUserDetails(ts, user)
}

// Test user update
func (ts *UserAPITestSuite) TestUserUpdate() {

	if createdUserID == "" {
		ts.T().Fatal("User ID is not available for update")
	}

	userJSON, err := json.Marshal(userToUpdate)
	if err != nil {
		ts.T().Fatalf("Failed to marshal userToUpdate: %v", err)
	}

	reqBody := bytes.NewReader(userJSON)
	req, err := http.NewRequest("PUT", testServerURL+"/users/"+createdUserID, reqBody)
	if err != nil {
		ts.T().Fatalf("Failed to create update request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send update request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ts.T().Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Validate the update by retrieving the user
	retrieveAndValidateUserDetails(ts, User{
		Id:         createdUserID,
		OrgId:      userToUpdate.OrgId,
		Type:       userToUpdate.Type,
		Attributes: userToUpdate.Attributes,
	})
}

func retrieveAndValidateUserDetails(ts *UserAPITestSuite, expectedUser User) {

	req, err := http.NewRequest("GET", testServerURL+"/users/"+expectedUser.Id, nil)
	if err != nil {
		ts.T().Fatalf("Failed to create get request: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send get request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ts.T().Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check if the response Content-Type is application/json
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		rawBody, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Unexpected Content-Type: %s. Raw body: %s", contentType, string(rawBody))
	}

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	if !user.equals(expectedUser) {
		ts.T().Fatalf("User mismatch, expected %+v, got %+v", expectedUser, user)
	}
}

func createUser(ts *UserAPITestSuite) (string, error) {

	userJSON, err := json.Marshal(userToCreate)
	if err != nil {
		ts.T().Fatalf("Failed to marshal userToCreate: %v", err)
	}

	reqBody := bytes.NewReader(userJSON)
	req, err := http.NewRequest("POST", testServerURL+"/users", reqBody)
	if err != nil {
		// print error
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
		return "", fmt.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var respBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	id, ok := respBody["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	createdUserID = id
	return id, nil
}

func deleteUser(userId string) error {

	req, err := http.NewRequest("DELETE", testServerURL+"/users/"+userId, nil)
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

	if resp.StatusCode != http.StatusOK {
		return err
	}
	return nil
}

func buildCreatedUser() User {

	return User{
		Id:         createdUserID,
		OrgId:      userToCreate.OrgId,
		Type:       userToCreate.Type,
		Attributes: userToCreate.Attributes,
	}
}
