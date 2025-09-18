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
	"io"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = testutils.TestServerURL
)

var (
	testGithubIdp = testutils.IDP{
		Name:        "Test Github IDP",
		Description: "Github Identity Provider for testing",
		Properties: []testutils.IDPProperty{
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

	testGoogleIdp = testutils.IDP{
		Name:        "Test Google IDP",
		Description: "Google Identity Provider for testing",
		Properties: []testutils.IDPProperty{
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

	idpToCreate = testutils.IDP{
		Name:        "Test OIDC IDP",
		Description: "OIDC test identity provider for CRUD operations",
		Properties: []testutils.IDPProperty{
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

	idpToUpdate = testutils.IDP{
		Name:        "Test Updated IDP",
		Description: "Updated test identity provider",
		Properties: []testutils.IDPProperty{
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
	testGithubIdpID string
	testGoogleIdpID string
	createdIdpID    string
	testIdps        []testutils.IDP // Track all created IDPs for validation
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

	githubId, err := testutils.CreateIDP(testGithubIdp)
	if err != nil {
		ts.T().Fatalf("Failed to create test Github IDP during setup: %v", err)
	}
	testGithubIdpID = githubId

	googleId, err := testutils.CreateIDP(testGoogleIdp)
	if err != nil {
		ts.T().Fatalf("Failed to create test Google IDP during setup: %v", err)
	}
	testGoogleIdpID = googleId

	createId, err := testutils.CreateIDP(idpToCreate)
	if err != nil {
		ts.T().Fatalf("Failed to create test OIDC IDP during setup: %v", err)
	}
	createdIdpID = createId

	// Build the list of created IDPs for test validations
	testIdps = []testutils.IDP{
		{ID: testGithubIdpID, Name: testGithubIdp.Name, Description: testGithubIdp.Description, Properties: testGithubIdp.Properties},
		{ID: testGoogleIdpID, Name: testGoogleIdp.Name, Description: testGoogleIdp.Description, Properties: testGoogleIdp.Properties},
		{ID: createdIdpID, Name: idpToCreate.Name, Description: idpToCreate.Description, Properties: idpToCreate.Properties},
	}
}

// TearDownSuite cleans up all test IDPs
func (ts *IdpAPITestSuite) TearDownSuite() {
	// Delete all test IDPs

	if testGithubIdpID != "" {
		err := testutils.DeleteIDP(testGithubIdpID)
		if err != nil {
			ts.T().Logf("Failed to delete test Github IDP during teardown: %v", err)
		}
	}

	if testGoogleIdpID != "" {
		err := testutils.DeleteIDP(testGoogleIdpID)
		if err != nil {
			ts.T().Logf("Failed to delete test Google IDP during teardown: %v", err)
		}
	}

	if createdIdpID != "" {
		err := testutils.DeleteIDP(createdIdpID)
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

	// Parse the response body - list endpoint returns BasicIdpResponse objects
	var basicIdps []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	err = json.NewDecoder(resp.Body).Decode(&basicIdps)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	idpListLength := len(basicIdps)
	if idpListLength == 0 {
		ts.T().Fatalf("Response does not contain any identity providers")
	}

	// Verify that all test IDPs we created are present in the list (basic info only)
	for _, expectedIdp := range testIdps {
		found := false
		for _, idp := range basicIdps {
			if idp.ID == expectedIdp.ID && idp.Name == expectedIdp.Name && idp.Description == expectedIdp.Description {
				found = true
				break
			}
		}
		if !found {
			ts.T().Fatalf("Test IDP not found in list: ID=%s, Name=%s", expectedIdp.ID, expectedIdp.Name)
		}
	}
}

// Test idP get by ID
func (ts *IdpAPITestSuite) TestIdpGetByID() {

	if createdIdpID == "" {
		ts.T().Fatal("IdP ID is not available for retrieval")
	}
	idp := testIdps[2]
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
	retrieveAndValidateIdpDetails(ts, testutils.IDP{
		ID:          createdIdpID,
		Name:        idpToUpdate.Name,
		Description: idpToUpdate.Description,
		Properties:  idpToUpdate.Properties,
	})
}

func retrieveAndValidateIdpDetails(ts *IdpAPITestSuite, expectedIdp testutils.IDP) {

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

	var idp testutils.IDP
	err = json.NewDecoder(resp.Body).Decode(&idp)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	if !compareIDPs(idp, expectedIdp) {
		ts.T().Fatalf("IdP mismatch, expected %+v, got %+v", expectedIdp, idp)
	}
}

// compareIDPs compares two IDP instances for equality, accounting for secret property masking
func compareIDPs(idp, expectedIdp testutils.IDP) bool {
	if idp.ID != expectedIdp.ID || idp.Name != expectedIdp.Name || idp.Description != expectedIdp.Description {
		return false
	}

	if len(idp.Properties) != len(expectedIdp.Properties) {
		return false
	}

	for _, expProp := range expectedIdp.Properties {
		propFound := false
		for _, p := range idp.Properties {
			if p.Name == expProp.Name {
				propFound = true
				// For secret properties, the API returns "******" so we don't compare values
				if !expProp.IsSecret && p.Value != expProp.Value {
					return false
				}
				// For secret properties, just check that it's masked
				if expProp.IsSecret && p.Value != "******" {
					return false
				}
				// Ensure IsSecret flag matches
				if p.IsSecret != expProp.IsSecret {
					return false
				}
				break
			}
		}
		if !propFound {
			return false
		}
	}

	return true
}
