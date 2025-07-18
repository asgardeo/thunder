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

package application

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
	preCreatedApp = Application{
		ID:                  "550e8400-e29b-41d4-a716-446655440000",
		OrgID:               "f5b738ee-8ee4-4kd0-b174-6f8d1eaf32ad",
		Name:                "Test SPA",
		Description:         "Initial testing App",
		ClientID:            "client123",
		CallbackURL:         []string{"https://localhost:3000"},
		SupportedGrantTypes: []string{"client_credentials", "authorization_code"},
	}

	appToCreate = Application{
		Name:                "My App",
		Description:         "A demo application",
		ClientID:            "abc1237",
		ClientSecret:        "s3cret",
		CallbackURL:         []string{"http://localhost/callback"},
		SupportedGrantTypes: []string{"authorization_code", "client_credentials"},
	}

	appToUpdate = Application{
		Name:                "Updated App",
		Description:         "Updated Description",
		ClientID:            "Updated abc1237",
		ClientSecret:        "Updated s3cret",
		CallbackURL:         []string{"http://localhost/callback2"},
		SupportedGrantTypes: []string{"authorization_code2", "client_credentials2"},
	}
)

var createdAppID string

type ApplicationAPITestSuite struct {
	suite.Suite
}

func TestApplicationAPITestSuite(t *testing.T) {

	suite.Run(t, new(ApplicationAPITestSuite))
}

// SetupSuite test application creation
func (ts *ApplicationAPITestSuite) SetupSuite() {

	id, err := createApplication(ts)
	if err != nil {
		ts.T().Fatalf("Failed to create application during setup: %v", err)
	} else {
		createdAppID = id
	}
}

// TearDownSuite test application deletion
func (ts *ApplicationAPITestSuite) TearDownSuite() {

	if createdAppID != "" {
		err := deleteApplication(createdAppID)
		if err != nil {
			ts.T().Fatalf("Failed to de;ete application during teardown: %v", err)
		}
	}
}

// Test application listing
func (ts *ApplicationAPITestSuite) TestApplicationListing() {

	req, err := http.NewRequest("GET", testServerURL+"/applications", nil)
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
	var applications []Application
	err = json.NewDecoder(resp.Body).Decode(&applications)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	applicationListLength := len(applications)
	if applicationListLength == 0 {
		ts.T().Fatalf("Response does not contain any applications")
	}

	if applicationListLength != 2 {
		ts.T().Fatalf("Expected 2 applications, got %d", applicationListLength)
	}

	app1 := applications[0]
	if !app1.equals(preCreatedApp) {
		ts.T().Fatalf("Application mismatch, expected %+v, got %+v", preCreatedApp, app1)
	}

	app2 := applications[1]
	createdApp := buildCreatedApp()
	if !app2.equals(createdApp) {
		ts.T().Fatalf("Application mismatch, expected %+v, got %+v", createdApp, app2)
	}
}

// Test application get by ID
func (ts *ApplicationAPITestSuite) TestApplicationGetByID() {

	if createdAppID == "" {
		ts.T().Fatal("Application ID is not available for retrieval")
	}
	application := buildCreatedApp()
	retrieveAndValidateApplicationDetails(ts, application)
}

// Test application update
func (ts *ApplicationAPITestSuite) TestApplicationUpdate() {

	if createdAppID == "" {
		ts.T().Fatal("Application ID is not available for update")
	}

	appJSON, err := json.Marshal(appToUpdate)
	if err != nil {
		ts.T().Fatalf("Failed to marshal appToUpdate: %v", err)
	}

	reqBody := bytes.NewReader(appJSON)
	req, err := http.NewRequest("PUT", testServerURL+"/applications/"+createdAppID, reqBody)
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

	// Validate the update by retrieving the application
	retrieveAndValidateApplicationDetails(ts, Application{
		ID:                  createdAppID,
		OrgID:               "f5b738ee-8ee4-4kd0-b174-6f8d1eaf32ad",
		Name:                appToUpdate.Name,
		Description:         appToUpdate.Description,
		ClientID:            appToUpdate.ClientID,
		ClientSecret:        appToUpdate.ClientSecret,
		CallbackURL:         appToUpdate.CallbackURL,
		SupportedGrantTypes: appToUpdate.SupportedGrantTypes,
	})
}

func retrieveAndValidateApplicationDetails(ts *ApplicationAPITestSuite, expectedApp Application) {

	req, err := http.NewRequest("GET", testServerURL+"/applications/"+expectedApp.ID, nil)
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

	var app Application
	err = json.NewDecoder(resp.Body).Decode(&app)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	if !app.equals(expectedApp) {
		ts.T().Fatalf("Application mismatch, expected %+v, got %+v", expectedApp, app)
	}
}

func createApplication(ts *ApplicationAPITestSuite) (string, error) {

	appJSON, err := json.Marshal(appToCreate)
	if err != nil {
		ts.T().Fatalf("Failed to marshal appToCreate: %v", err)
	}

	reqBody := bytes.NewReader(appJSON)
	req, err := http.NewRequest("POST", testServerURL+"/applications", reqBody)
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
	createdAppID = id
	return id, nil
}

func deleteApplication(appID string) error {

	req, err := http.NewRequest("DELETE", testServerURL+"/applications/"+appID, nil)
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

func buildCreatedApp() Application {

	return Application{
		ID:                  createdAppID,
		OrgID:               "f5b738ee-8ee4-4kd0-b174-6f8d1eaf32ad",
		Name:                appToCreate.Name,
		Description:         appToCreate.Description,
		ClientID:            appToCreate.ClientID,
		ClientSecret:        appToCreate.ClientSecret,
		CallbackURL:         appToCreate.CallbackURL,
		SupportedGrantTypes: appToCreate.SupportedGrantTypes,
	}
}
