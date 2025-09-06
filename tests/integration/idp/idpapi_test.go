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

package idp

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
	testLocalIdp = IDP{
		Name:        "Test Local IDP",
		Description: "Local Identity Provider for testing",
		Properties:  []IDPProperty{},
	}

	testGithubIdp = IDP{
		Name:        "Test Github IDP",
		Description: "Github Identity Provider for testing",
		Properties: []IDPProperty{
			{
				Name:     "client_id",
				Value:    "test_github_client",
				IsSecret: false,
			},
			{
				Name:     "client_secret",
				Value:    "test_github_secret",
				IsSecret: true,
			},
			{
				Name:     "redirect_uri",
				Value:    "https://localhost:3000/github/callback",
				IsSecret: false,
			},
			{
				Name:     "scopes",
				Value:    "user:email,read:user",
				IsSecret: false,
			},
		},
	}

	testGoogleIdp = IDP{
		Name:        "Test Google IDP",
		Description: "Google Identity Provider for testing",
		Properties: []IDPProperty{
			{
				Name:     "client_id",
				Value:    "test_google_client",
				IsSecret: false,
			},
			{
				Name:     "client_secret",
				Value:    "test_google_secret",
				IsSecret: true,
			},
			{
				Name:     "redirect_uri",
				Value:    "https://localhost:3000/google/callback",
				IsSecret: false,
			},
			{
				Name:     "scopes",
				Value:    "openid,email,profile",
				IsSecret: false,
			},
		},
	}

	idpToCreate = IDP{
		Name:        "Test OIDC IDP",
		Description: "OIDC test identity provider for CRUD operations",
		Properties: []IDPProperty{
			{
				Name:     "client_id",
				Value:    "test_oidc_client",
				IsSecret: false,
			},
			{
				Name:     "client_secret",
				Value:    "test_oidc_secret",
				IsSecret: true,
			},
			{
				Name:     "redirect_uri",
				Value:    "https://localhost:3000/oidc/callback",
				IsSecret: false,
			},
			{
				Name:     "scopes",
				Value:    "openid,email,profile",
				IsSecret: false,
			},
		},
	}

	idpToUpdate = IDP{
		Name:        "Test Updated IDP",
		Description: "Updated test identity provider",
		Properties: []IDPProperty{
			{
				Name:     "client_id",
				Value:    "test_updated_client",
				IsSecret: false,
			},
			{
				Name:     "client_secret",
				Value:    "test_updated_secret",
				IsSecret: true,
			},
			{
				Name:     "redirect_uri",
				Value:    "https://localhost:3000/updated/callback",
				IsSecret: false,
			},
			{
				Name:     "scopes",
				Value:    "user:email,read:user",
				IsSecret: false,
			},
		},
	}
)

var (
	testLocalIdpID  string
	testGithubIdpID string
	testGoogleIdpID string
	createdIdpID   string
	testIdps        []IDP // Track all created IDPs for validation
)

type IdpAPITestSuite struct {
	suite.Suite
}

func TestIdpAPITestSuite(t *testing.T) {

	suite.Run(t, new(IdpAPITestSuite))
}

// SetupSuite creates test IDPs via API
func (ts *IdpAPITestSuite) SetupSuite() {
	// Create all test IDPs
	localId, err := createIdp(ts, testLocalIdp)
	if err != nil {
		ts.T().Fatalf("Failed to create test Local IDP during setup: %v", err)
	}
	testLocalIdpID = localId

	githubId, err := createIdp(ts, testGithubIdp)
	if err != nil {
		ts.T().Fatalf("Failed to create test Github IDP during setup: %v", err)
	}
	testGithubIdpID = githubId

	googleId, err := createIdp(ts, testGoogleIdp)
	if err != nil {
		ts.T().Fatalf("Failed to create test Google IDP during setup: %v", err)
	}
	testGoogleIdpID = googleId

	createId, err := createIdp(ts, idpToCreate)
	if err != nil {
		ts.T().Fatalf("Failed to create test OIDC IDP during setup: %v", err)
	}
	createdIdpID = createId

	// Build the list of created IDPs for test validations
	testIdps = []IDP{
		{ID: testLocalIdpID, Name: testLocalIdp.Name, Description: testLocalIdp.Description, Properties: testLocalIdp.Properties},
		{ID: testGithubIdpID, Name: testGithubIdp.Name, Description: testGithubIdp.Description, Properties: testGithubIdp.Properties},
		{ID: testGoogleIdpID, Name: testGoogleIdp.Name, Description: testGoogleIdp.Description, Properties: testGoogleIdp.Properties},
		{ID: createdIdpID, Name: idpToCreate.Name, Description: idpToCreate.Description, Properties: idpToCreate.Properties},
	}
}

// TearDownSuite cleans up all test IDPs
func (ts *IdpAPITestSuite) TearDownSuite() {
	// Delete all test IDPs
	if testLocalIdpID != "" {
		err := deleteIdp(testLocalIdpID)
		if err != nil {
			ts.T().Logf("Failed to delete test Local IDP during teardown: %v", err)
		}
	}

	if testGithubIdpID != "" {
		err := deleteIdp(testGithubIdpID)
		if err != nil {
			ts.T().Logf("Failed to delete test Github IDP during teardown: %v", err)
		}
	}

	if testGoogleIdpID != "" {
		err := deleteIdp(testGoogleIdpID)
		if err != nil {
			ts.T().Logf("Failed to delete test Google IDP during teardown: %v", err)
		}
	}

	if createdIdpID != "" {
		err := deleteIdp(createdIdpID)
		if err != nil {
			ts.T().Logf("Failed to delete test OIDC IDP during teardown: %v", err)
		}
	}
}

// Test IdP listing
func (ts *IdpAPITestSuite) TestIdpListing() {

	req, err := http.NewRequest("GET", testServerURL+"/identity-providers", nil)
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
	var idps []IDP
	err = json.NewDecoder(resp.Body).Decode(&idps)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	idpListLength := len(idps)
	if idpListLength == 0 {
		ts.T().Fatalf("Response does not contain any identity providers")
	}

	// Verify that all test IDPs we created are present in the list
	for _, expectedIdp := range testIdps {
		found := false
		for _, idp := range idps {
			if idp.equals(expectedIdp) {
				found = true
				break
			}
		}
		if !found {
			ts.T().Fatalf("Test IDP not found in list: %+v", expectedIdp)
		}
	}
}

// Test idP get by ID
func (ts *IdpAPITestSuite) TestIdpGetByID() {

	if createdIdpID == "" {
		ts.T().Fatal("IdP ID is not available for retrieval")
	}
	idp := testIdps[3]
	retrieveAndValidateIdpDetails(ts, idp)
}

// Test idP update
func (ts *IdpAPITestSuite) TestIdpUpdate() {

	if createdIdpID == "" {
		ts.T().Fatal("IdP ID is not available for update")
	}

	idpJSON, err := json.Marshal(idpToUpdate)
	if err != nil {
		ts.T().Fatalf("Failed to marshal idPToUpdate: %v", err)
	}

	reqBody := bytes.NewReader(idpJSON)
	req, err := http.NewRequest("PUT", testServerURL+"/identity-providers/"+createdIdpID, reqBody)
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

	// Validate the update by retrieving the idP
	retrieveAndValidateIdpDetails(ts, IDP{
		ID:          createdIdpID,
		Name:        idpToUpdate.Name,
		Description: idpToUpdate.Description,
		Properties:  idpToUpdate.Properties,
	})
}

func retrieveAndValidateIdpDetails(ts *IdpAPITestSuite, expectedIdp IDP) {

	req, err := http.NewRequest("GET", testServerURL+"/identity-providers/"+expectedIdp.ID, nil)
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

	var idp IDP
	err = json.NewDecoder(resp.Body).Decode(&idp)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	if !idp.equals(expectedIdp) {
		ts.T().Fatalf("IdP mismatch, expected %+v, got %+v", expectedIdp, idp)
	}
}

func createIdp(ts *IdpAPITestSuite, idp IDP) (string, error) {

	idpJSON, err := json.Marshal(idp)
	if err != nil {
		ts.T().Fatalf("Failed to marshal IDP template: %v", err)
	}

	reqBody := bytes.NewReader(idpJSON)
	req, err := http.NewRequest("POST", testServerURL+"/identity-providers", reqBody)
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
	return id, nil
}

func deleteIdp(idpId string) error {

	req, err := http.NewRequest("DELETE", testServerURL+"/identity-providers/"+idpId, nil)
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
