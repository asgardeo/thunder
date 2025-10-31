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

package application

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
)

var (
	testApp = Application{
		Name:                      "Test App",
		Description:               "Test application for API testing",
		IsRegistrationFlowEnabled: false,
		URL:                       "https://testapp.example.com",
		LogoURL:                   "https://testapp.example.com/logo.png",
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "test_app_client",
					ClientSecret:            "test_app_secret",
					RedirectURIs:            []string{"http://localhost/testapp/callback"},
					GrantTypes:              []string{"authorization_code", "client_credentials"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					PKCERequired:            false,
					PublicClient:            false,
				},
			},
		},
	}

	appToCreate = Application{
		Name:                      "App To Create",
		Description:               "Application to create for API testing",
		IsRegistrationFlowEnabled: true,
		URL:                       "https://apptocreate.example.com",
		LogoURL:                   "https://apptocreate.example.com/logo.png",
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "app_to_create_client",
					ClientSecret:            "app_to_create_secret",
					RedirectURIs:            []string{"http://localhost/apptocreate/callback"},
					GrantTypes:              []string{"authorization_code", "client_credentials"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					PKCERequired:            false,
					PublicClient:            false,
				},
			},
		},
	}

	appToUpdate = Application{
		Name:                      "Updated App",
		Description:               "Updated Description",
		IsRegistrationFlowEnabled: false,
		URL:                       "https://appToUpdate.example.com",
		LogoURL:                   "https://appToUpdate.example.com/logo.png",
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "updated_client_id",
					ClientSecret:            "updated_secret",
					RedirectURIs:            []string{"http://localhost/callback2"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					PKCERequired:            false,
					PublicClient:            false,
				},
			},
		},
	}
)

var (
	testAppID       string
	testAppInstance Application
)

type ApplicationAPITestSuite struct {
	suite.Suite
}

func TestApplicationAPITestSuite(t *testing.T) {

	suite.Run(t, new(ApplicationAPITestSuite))
}

// SetupSuite creates test applications for the test suite
func (ts *ApplicationAPITestSuite) SetupSuite() {
	// Create test application
	app1ID, err := createApplication(testApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	testAppID = app1ID

	// Build the test app structure for validations
	testAppInstance = testApp
	testAppInstance.ID = testAppID
	if len(testAppInstance.InboundAuthConfig) > 0 && testAppInstance.InboundAuthConfig[0].OAuthAppConfig != nil {
		testAppInstance.ClientID = testAppInstance.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}
}

// TearDownSuite cleans up test applications
func (ts *ApplicationAPITestSuite) TearDownSuite() {
	// Delete the test application
	if testAppID != "" {
		err := deleteApplication(testAppID)
		if err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
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
	var appList ApplicationList
	err = json.NewDecoder(resp.Body).Decode(&appList)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	totalResults := appList.TotalResults
	if totalResults == 0 {
		ts.T().Fatalf("Response does not contain a valid total results count")
	}

	appCount := appList.Count
	if appCount == 0 {
		ts.T().Fatalf("Response does not contain a valid application count")
	}

	applicationListLength := len(appList.Applications)
	if applicationListLength == 0 {
		ts.T().Fatalf("Response does not contain any applications")
	}

	// Verify that the test application is present in the list
	testApps := []Application{testAppInstance}
	for _, expectedApp := range testApps {
		found := false
		for _, app := range appList.Applications {
			if app.ID == expectedApp.ID &&
				app.Name == expectedApp.Name &&
				app.Description == expectedApp.Description &&
				app.ClientID == expectedApp.ClientID {
				found = true
				break
			}
		}
		if !found {
			ts.T().Fatalf("Test application not found in list: %+v", expectedApp)
		}
	}
}

// Test application get by ID
func (ts *ApplicationAPITestSuite) TestApplicationGetByID() {
	// Create an application for get testing
	appID, err := createApplication(appToCreate)
	if err != nil {
		ts.T().Fatalf("Failed to create application for get test: %v", err)
	}
	defer func() {
		// Clean up the created application
		if err := deleteApplication(appID); err != nil {
			ts.T().Logf("Failed to delete application after get test: %v", err)
		}
	}()

	// Build the expected app structure for validation
	expectedApp := appToCreate
	expectedApp.ID = appID
	if len(expectedApp.InboundAuthConfig) > 0 && expectedApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		expectedApp.ClientID = expectedApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}

	retrieveAndValidateApplicationDetails(ts, expectedApp)
}

// Test application update
func (ts *ApplicationAPITestSuite) TestApplicationUpdate() {
	// Create an application for update testing
	appID, err := createApplication(appToCreate)
	if err != nil {
		ts.T().Fatalf("Failed to create application for update test: %v", err)
	}
	defer func() {
		// Clean up the created application
		if err := deleteApplication(appID); err != nil {
			ts.T().Logf("Failed to delete application after update test: %v", err)
		}
	}()

	// Add the ID to the application to update
	appToUpdateWithID := appToUpdate
	appToUpdateWithID.ID = appID

	appJSON, err := json.Marshal(appToUpdateWithID)
	if err != nil {
		ts.T().Fatalf("Failed to marshal appToUpdate: %v", err)
	}

	reqBody := bytes.NewReader(appJSON)
	req, err := http.NewRequest("PUT", testServerURL+"/applications/"+appID, reqBody)
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
		responseBody, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	// For update operations, verify the response directly
	var updatedApp Application
	if err = json.NewDecoder(resp.Body).Decode(&updatedApp); err != nil {
		responseBody, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Failed to decode update response: %v. Response: %s", err, string(responseBody))
	}

	// Client secret should be present in the update response
	if len(updatedApp.InboundAuthConfig) > 0 &&
		updatedApp.InboundAuthConfig[0].OAuthAppConfig != nil &&
		updatedApp.InboundAuthConfig[0].OAuthAppConfig.ClientSecret == "" {
		ts.T().Fatalf("Expected client secret in update response but got empty string")
	}

	// Now validate by getting the application (which should not have client secret)
	// Make sure client ID is properly set in the root level before validation
	if len(appToUpdateWithID.InboundAuthConfig) > 0 &&
		appToUpdateWithID.InboundAuthConfig[0].OAuthAppConfig != nil {
		appToUpdateWithID.ClientID = appToUpdateWithID.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}

	retrieveAndValidateApplicationDetails(ts, appToUpdateWithID)
}

func retrieveAndValidateApplicationDetails(ts *ApplicationAPITestSuite, expectedApp Application) {

	req, err := http.NewRequest("GET", testServerURL+"/applications/"+expectedApp.ID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	// Check if the response Content-Type is application/json
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		ts.T().Fatalf("Expected Content-Type application/json, got %s", contentType)
	}

	var app Application
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &app)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v\nResponse body: %s", err, string(body))
	}

	// For GET operations, client secret should be empty in the response
	// Make sure expectedApp has client secret cleared for proper comparison
	appForComparison := expectedApp
	if len(appForComparison.InboundAuthConfig) > 0 && appForComparison.InboundAuthConfig[0].OAuthAppConfig != nil {
		// Make sure client ID is in root object
		appForComparison.ClientID = appForComparison.InboundAuthConfig[0].OAuthAppConfig.ClientID
		// Remove client secret for GET comparison
		appForComparison.InboundAuthConfig[0].OAuthAppConfig.ClientSecret = ""
	}

	// Ensure certificate is set in expected app if it's null
	if appForComparison.Certificate == nil {
		appForComparison.Certificate = &ApplicationCert{
			Type:  "NONE",
			Value: "",
		}
	}

	// If expected doesn't have Token but API returned one (default), copy it to expected
	// This handles cases where the server provides default token config
	if appForComparison.Token == nil && app.Token != nil {
		appForComparison.Token = app.Token
	}

	if !app.equals(appForComparison) {
		appJSON, _ := json.MarshalIndent(app, "", "  ")
		expectedJSON, _ := json.MarshalIndent(appForComparison, "", "  ")
		ts.T().Fatalf("Application mismatch:\nGot:\n%s\n\nExpected:\n%s", string(appJSON), string(expectedJSON))
	}
}

func createApplication(app Application) (string, error) {
	appJSON, err := json.Marshal(app)
	if err != nil {
		return "", fmt.Errorf("failed to marshal application: %w", err)
	}

	reqBody := bytes.NewReader(appJSON)
	req, err := http.NewRequest("POST", testServerURL+"/applications", reqBody)
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
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	// For create operations, directly parse the response to a full Application
	var createdApp Application
	err = json.NewDecoder(resp.Body).Decode(&createdApp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	// Verify client secret is present in the create response for confidential clients
	if len(createdApp.InboundAuthConfig) > 0 &&
		createdApp.InboundAuthConfig[0].OAuthAppConfig != nil &&
		!createdApp.InboundAuthConfig[0].OAuthAppConfig.PublicClient &&
		createdApp.InboundAuthConfig[0].OAuthAppConfig.ClientSecret == "" {
		return "", fmt.Errorf("expected client secret in create response but got empty string")
	}

	id := createdApp.ID
	if id == "" {
		return "", fmt.Errorf("response does not contain id")
	}
	return id, nil
}

func deleteApplication(appID string) error {
	req, err := http.NewRequest("DELETE", testServerURL+"/applications/"+appID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status 204, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}
	return nil
}

// TestApplicationCreationWithDefaults tests that applications created without grant_types, response_types, or token_endpoint_auth_method get proper defaults
func (ts *ApplicationAPITestSuite) TestApplicationCreationWithDefaults() {
	appWithDefaults := Application{
		Name:                      "App With Defaults",
		Description:               "Application to test default values",
		IsRegistrationFlowEnabled: false,
		URL:                       "https://defaults.example.com",
		LogoURL:                   "https://defaults.example.com/logo.png",
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:     "defaults_app_client",
					ClientSecret: "defaults_app_secret",
					RedirectURIs: []string{"http://localhost/defaults/callback"},
					// Intentionally omitting GrantTypes, ResponseTypes, and TokenEndpointAuthMethod
					PKCERequired: false,
					PublicClient: false,
				},
			},
		},
	}

	appID, err := createApplication(appWithDefaults)
	if err != nil {
		ts.T().Fatalf("Failed to create application: %v", err)
	}

	req, err := http.NewRequest("GET", testServerURL+"/applications/"+appID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to create GET request: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	var retrievedApp Application
	err = json.NewDecoder(resp.Body).Decode(&retrievedApp)
	if err != nil {
		ts.T().Fatalf("Failed to decode response: %v", err)
	}

	// Verify defaults were applied
	if len(retrievedApp.InboundAuthConfig) > 0 && retrievedApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		oauthConfig := retrievedApp.InboundAuthConfig[0].OAuthAppConfig

		ts.Assert().Equal([]string{"authorization_code"}, oauthConfig.GrantTypes, "Default grant_types should be ['authorization_code']")
		ts.Assert().Equal([]string{"code"}, oauthConfig.ResponseTypes, "Default response_types should be ['code']")
		ts.Assert().Equal("client_secret_basic", oauthConfig.TokenEndpointAuthMethod, "Default token_endpoint_auth_method should be 'client_secret_basic'")
	}

	err = deleteApplication(appID)
	if err != nil {
		ts.T().Logf("Failed to delete test application: %v", err)
	}
}

// TestApplicationCreationWithInvalidTokenEndpointAuthMethod tests validation of invalid token_endpoint_auth_method values
func (ts *ApplicationAPITestSuite) TestApplicationCreationWithInvalidTokenEndpointAuthMethod() {
	appWithInvalidAuthMethod := Application{
		Name:                      "App With Invalid Auth Method",
		Description:               "Application to test invalid token endpoint auth method",
		IsRegistrationFlowEnabled: false,
		URL:                       "https://invalid.example.com",
		LogoURL:                   "https://invalid.example.com/logo.png",
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "invalid_auth_app_client",
					ClientSecret:            "invalid_auth_app_secret",
					RedirectURIs:            []string{"http://localhost/invalid/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "invalid_auth_method", // Invalid value
					PKCERequired:            false,
					PublicClient:            false,
				},
			},
		},
	}

	_, err := createApplication(appWithInvalidAuthMethod)
	if err == nil {
		ts.T().Fatalf("Expected validation error for invalid token_endpoint_auth_method, but application was created successfully")
	}

	appWithEmptyAuthMethod := Application{
		Name:                      "App With Empty Auth Method",
		Description:               "Application to test empty token endpoint auth method",
		IsRegistrationFlowEnabled: false,
		URL:                       "https://empty.example.com",
		LogoURL:                   "https://empty.example.com/logo.png",
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "empty_auth_app_client",
					ClientSecret:            "empty_auth_app_secret",
					RedirectURIs:            []string{"http://localhost/empty/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "",
					PKCERequired:            false,
					PublicClient:            false,
				},
			},
		},
	}

	appID, err := createApplication(appWithEmptyAuthMethod)
	if err != nil {
		ts.T().Fatalf("Failed to create application with empty token_endpoint_auth_method: %v", err)
	}

	req, err := http.NewRequest("GET", testServerURL+"/applications/"+appID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to create GET request: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	var retrievedApp Application
	err = json.NewDecoder(resp.Body).Decode(&retrievedApp)
	if err != nil {
		ts.T().Fatalf("Failed to decode response: %v", err)
	}

	if len(retrievedApp.InboundAuthConfig) > 0 && retrievedApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		oauthConfig := retrievedApp.InboundAuthConfig[0].OAuthAppConfig
		ts.Assert().Equal("client_secret_basic", oauthConfig.TokenEndpointAuthMethod, "Empty token_endpoint_auth_method should get default 'client_secret_basic'")
	}

	err = deleteApplication(appID)
	if err != nil {
		ts.T().Logf("Failed to delete test application: %v", err)
	}
}

// TestApplicationCreationWithPartialDefaults tests applications with some fields missing (partial defaults)
func (ts *ApplicationAPITestSuite) TestApplicationCreationWithPartialDefaults() {
	appWithPartialDefaults := Application{
		Name:                      "App With Partial Defaults",
		Description:               "Application to test partial default values",
		IsRegistrationFlowEnabled: false,
		URL:                       "https://partial.example.com",
		LogoURL:                   "https://partial.example.com/logo.png",
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:     "partial_app_client",
					ClientSecret: "partial_app_secret",
					RedirectURIs: []string{"http://localhost/partial/callback"},
					// GrantTypes missing - should get default
					ResponseTypes:           []string{"code"},     // Explicitly set
					TokenEndpointAuthMethod: "client_secret_post", // Explicitly set
					PKCERequired:            false,
					PublicClient:            false,
				},
			},
		},
	}

	appID, err := createApplication(appWithPartialDefaults)
	if err != nil {
		ts.T().Fatalf("Failed to create application: %v", err)
	}

	// Verify that defaults were applied by getting the application
	req, err := http.NewRequest("GET", testServerURL+"/applications/"+appID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to create GET request: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	var retrievedApp Application
	err = json.NewDecoder(resp.Body).Decode(&retrievedApp)
	if err != nil {
		ts.T().Fatalf("Failed to decode response: %v", err)
	}

	if len(retrievedApp.InboundAuthConfig) > 0 && retrievedApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		oauthConfig := retrievedApp.InboundAuthConfig[0].OAuthAppConfig

		ts.Assert().Equal([]string{"authorization_code"}, oauthConfig.GrantTypes, "Missing grant_types should get default ['authorization_code']")
		ts.Assert().Equal([]string{"code"}, oauthConfig.ResponseTypes, "Explicitly set response_types should be preserved")
		ts.Assert().Equal("client_secret_post", oauthConfig.TokenEndpointAuthMethod, "Explicitly set token_endpoint_auth_method should be preserved")
	}

	err = deleteApplication(appID)
	if err != nil {
		ts.T().Logf("Failed to delete test application: %v", err)
	}
}

// TestApplicationWithJWKSURICertificate tests creating application with JWKS_URI certificate.
func (ts *ApplicationAPITestSuite) TestApplicationWithJWKSURICertificate() {
	app := Application{
		Name:        "JWKS URI Certificate Test App",
		Description: "Test application with JWKS_URI certificate",
		URL:         "https://jwksuri.example.com",
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://jwksuri.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile", "email"},
				},
			},
		},
		Certificate: &ApplicationCert{
			Type:  "JWKS_URI",
			Value: "https://jwksuri.example.com/.well-known/jwks.json",
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Require().NotNil(retrievedApp.Certificate)
	ts.Assert().Equal("JWKS_URI", retrievedApp.Certificate.Type)
	ts.Assert().Equal("https://jwksuri.example.com/.well-known/jwks.json", retrievedApp.Certificate.Value)

	err = deleteApplication(appID)
	if err != nil {
		ts.T().Logf("Failed to delete test application: %v", err)
	}
}

// TestApplicationWithJWKSCertificate tests creating application with inline JWKS certificate.
func (ts *ApplicationAPITestSuite) TestApplicationWithJWKSCertificate() {
	jwksJSON := `{"keys":[{"kty":"RSA","use":"sig","kid":"test-key","n":"0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw","e":"AQAB"}]}`

	app := Application{
		Name:        "JWKS Inline Certificate Test App",
		Description: "Test application with inline JWKS certificate",
		URL:         "https://jwks.example.com",
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://jwks.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile"},
				},
			},
		},
		Certificate: &ApplicationCert{
			Type:  "JWKS",
			Value: jwksJSON,
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Require().NotNil(retrievedApp.Certificate)
	ts.Assert().Equal("JWKS", retrievedApp.Certificate.Type)
	ts.Assert().Equal(jwksJSON, retrievedApp.Certificate.Value)

	err = deleteApplication(appID)
	if err != nil {
		ts.T().Logf("Failed to delete test application: %v", err)
	}
}

// TestApplicationScopesAsArray tests that scopes are stored and retrieved as array.
func (ts *ApplicationAPITestSuite) TestApplicationScopesAsArray() {
	expectedScopes := []string{"openid", "profile", "email", "address", "phone"}

	app := Application{
		Name:        "Scopes Array Test App",
		Description: "Test application with scopes as array",
		URL:         "https://scopes.example.com",
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://scopes.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  expectedScopes,
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(appID)

	// Cleanup
	err = deleteApplication(appID)
	if err != nil {
		ts.T().Logf("Failed to delete test application: %v", err)
	}
}

// TestApplicationWithMultipleScopesAndCertificate tests creating application with both scopes and certificate.
func (ts *ApplicationAPITestSuite) TestApplicationWithMultipleScopesAndCertificate() {
	app := Application{
		Name:        "Multi Feature Test App",
		Description: "Test application with certificate and scopes",
		URL:         "https://multi.example.com",
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://multi.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile", "email", "custom:scope"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(appID)

	err = deleteApplication(appID)
	if err != nil {
		ts.T().Logf("Failed to delete test application: %v", err)
	}
}

// TestApplicationRedirectURIFragmentValidation tests that redirect URIs with fragments are rejected.
func (ts *ApplicationAPITestSuite) TestApplicationRedirectURIFragmentValidation() {
	app := Application{
		Name:        "Invalid Redirect URI Test",
		Description: "Test redirect URI validation",
		URL:         "https://invalid.example.com",
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://invalid.example.com/callback#fragment"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
				},
			},
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err)
}

// TestApplicationEmptyScopesArray tests that empty scopes array is accepted.
func (ts *ApplicationAPITestSuite) TestApplicationEmptyScopesArray() {
	app := Application{
		Name:        "Empty Scopes Test App",
		Description: "Test application with empty scopes",
		URL:         "https://emptyscopes.example.com",
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://emptyscopes.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(appID)

	err = deleteApplication(appID)
	if err != nil {
		ts.T().Logf("Failed to delete test application: %v", err)
	}
}

// TestApplicationCertificateUpdate tests updating application certificate.
func (ts *ApplicationAPITestSuite) TestApplicationCertificateUpdate() {
	app := Application{
		Name:        "Certificate Update Test App",
		Description: "Test certificate updates",
		URL:         "https://certupdate.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://certupdate.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Update to add JWKS_URI certificate
	app.Certificate = &ApplicationCert{
		Type:  "JWKS_URI",
		Value: "https://certupdate.example.com/.well-known/jwks.json",
	}

	appJSON, _ := json.Marshal(app)
	req, _ := http.NewRequest("PUT", testServerURL+"/applications/"+appID, bytes.NewReader(appJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)

	// Update to JWKS
	app.Certificate = &ApplicationCert{
		Type:  "JWKS",
		Value: `{"keys":[{"kty":"RSA","use":"sig","kid":"test"}]}`,
	}
	appJSON, _ = json.Marshal(app)
	req, _ = http.NewRequest("PUT", testServerURL+"/applications/"+appID, bytes.NewReader(appJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)
}

// TestOAuthAppCertificateUpdate tests updating OAuth app certificate.
func (ts *ApplicationAPITestSuite) TestOAuthAppCertificateUpdate() {
	app := Application{
		Name:        "OAuth Cert Update Test",
		Description: "Test OAuth certificate updates",
		URL:         "https://oauthcertupdate.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://oauthcertupdate.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Update to add JWKS_URI certificate at application level
	app.Certificate = &ApplicationCert{
		Type:  "JWKS_URI",
		Value: "https://oauthcertupdate.example.com/.well-known/jwks.json",
	}

	appJSON, _ := json.Marshal(app)
	req, _ := http.NewRequest("PUT", testServerURL+"/applications/"+appID, bytes.NewReader(appJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)
}

// TestApplicationInvalidCertificateType tests invalid certificate type rejection.
func (ts *ApplicationAPITestSuite) TestApplicationInvalidCertificateType() {
	app := Application{
		Name:        "Invalid Cert Type Test",
		Description: "Test invalid certificate type",
		URL:         "https://invalidcert.example.com",
		Certificate: &ApplicationCert{Type: "INVALID_TYPE", Value: "some-value"},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://invalidcert.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err)
}

// TestApplicationInvalidJWKSURI tests invalid JWKS_URI rejection.
func (ts *ApplicationAPITestSuite) TestApplicationInvalidJWKSURI() {
	app := Application{
		Name:        "Invalid JWKS URI Test",
		Description: "Test invalid JWKS URI",
		URL:         "https://invalidjwksuri.example.com",
		Certificate: &ApplicationCert{Type: "JWKS_URI", Value: "not-a-valid-uri"},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://invalidjwksuri.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err)
}

// TestApplicationEmptyJWKS tests empty JWKS value rejection.
func (ts *ApplicationAPITestSuite) TestApplicationEmptyJWKS() {
	app := Application{
		Name:        "Empty JWKS Test",
		Description: "Test empty JWKS",
		URL:         "https://emptyjwks.example.com",
		Certificate: &ApplicationCert{Type: "JWKS", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://emptyjwks.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err)
}

// TestApplicationPublicClientValidations tests public client configuration validations.
func (ts *ApplicationAPITestSuite) TestApplicationPublicClientValidations() {
	// Public client with wrong auth method
	app := Application{
		Name:        "Public Client Invalid Auth",
		Description: "Test public client validations",
		URL:         "https://publicclienttest.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://publicclienttest.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					PublicClient:            true,
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err)

	// Public client with invalid grant type
	app.Name = "Public Client Invalid Grant"
	app.InboundAuthConfig[0].OAuthAppConfig.TokenEndpointAuthMethod = "none"
	app.InboundAuthConfig[0].OAuthAppConfig.GrantTypes = []string{"client_credentials"}
	app.InboundAuthConfig[0].OAuthAppConfig.ResponseTypes = []string{}
	_, err = createApplication(app)
	ts.Assert().Error(err)
}

// TestApplicationOAuthConfigValidations tests OAuth configuration validations.
func (ts *ApplicationAPITestSuite) TestApplicationOAuthConfigValidations() {
	// authorization_code without redirect_uris
	app := Application{
		Name:        "OAuth Config No RedirectURIs",
		Description: "Test OAuth config validations",
		URL:         "https://oauthconfigtest.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err)

	// client_credentials with response_types
	app.Name = "OAuth Config Invalid client_credentials"
	app.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs = []string{"https://test.example.com/callback"}
	app.InboundAuthConfig[0].OAuthAppConfig.GrantTypes = []string{"client_credentials"}
	app.InboundAuthConfig[0].OAuthAppConfig.ResponseTypes = []string{"code"}
	_, err = createApplication(app)
	ts.Assert().Error(err)

	// client_credentials with none auth method
	app.Name = "OAuth Config client_credentials with none"
	app.InboundAuthConfig[0].OAuthAppConfig.ResponseTypes = []string{}
	app.InboundAuthConfig[0].OAuthAppConfig.TokenEndpointAuthMethod = "none"
	_, err = createApplication(app)
	ts.Assert().Error(err)
}

// TestApplicationWithTokenConfiguration tests creating and updating applications with token config.
func (ts *ApplicationAPITestSuite) TestApplicationWithTokenConfiguration() {
	app := Application{
		Name:        "Token Config Test App",
		Description: "Test application with token configuration",
		URL:         "https://tokenconfig.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://tokenconfig.example.com/callback"},
					GrantTypes:              []string{"authorization_code", "refresh_token"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile"},
					Token: &OAuthTokenConfig{
						Issuer: "https://tokenconfig.example.com",
						AccessToken: &TokenConfig{
							Issuer:         "https://tokenconfig.example.com",
							ValidityPeriod: 3600,
							UserAttributes: []string{"email", "username"},
						},
						IDToken: &IDTokenConfig{
							ValidityPeriod: 3600,
							UserAttributes: []string{"sub", "email"},
							ScopeClaims: map[string][]string{
								"profile": {"name", "given_name", "family_name"},
								"email":   {"email", "email_verified"},
							},
						},
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(appID)
	defer deleteApplication(appID)

	// Retrieve and verify the token configuration was persisted
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token)
	ts.Assert().Equal("https://tokenconfig.example.com", retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.Issuer)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken)
	ts.Assert().Equal(int64(3600), retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.ValidityPeriod)
}

// TestApplicationWithIDTokenScopeClaims tests ID token scope claims configuration.
func (ts *ApplicationAPITestSuite) TestApplicationWithIDTokenScopeClaims() {
	app := Application{
		Name:        "ID Token Scope Claims Test",
		Description: "Test ID token scope claims",
		URL:         "https://idtokenclaims.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://idtokenclaims.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile", "email", "address"},
					Token: &OAuthTokenConfig{
						IDToken: &IDTokenConfig{
							ValidityPeriod: 7200,
							UserAttributes: []string{"sub", "email", "name"},
							ScopeClaims: map[string][]string{
								"profile": {"name", "given_name", "family_name", "middle_name", "nickname", "preferred_username"},
								"email":   {"email", "email_verified"},
								"address": {"address"},
								"phone":   {"phone_number", "phone_number_verified"},
							},
						},
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Retrieve and verify scope claims
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims)
	ts.Assert().Contains(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims, "profile")
	ts.Assert().Contains(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims["email"], "email")
}

// TestApplicationUpdateWithTokenConfigChanges tests updating token configuration.
func (ts *ApplicationAPITestSuite) TestApplicationUpdateWithTokenConfigChanges() {
	// Create app with basic token config
	app := Application{
		Name:        "Token Config Update Test",
		Description: "Test token config updates",
		URL:         "https://tokenconfigupdate.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://tokenconfigupdate.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
					Token: &OAuthTokenConfig{
						AccessToken: &TokenConfig{
							ValidityPeriod: 1800,
						},
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Update with more complex token config
	app.InboundAuthConfig[0].OAuthAppConfig.Token = &OAuthTokenConfig{
		Issuer: "https://tokenconfigupdate.example.com",
		AccessToken: &TokenConfig{
			Issuer:         "https://tokenconfigupdate.example.com",
			ValidityPeriod: 7200,
			UserAttributes: []string{"email", "username", "role"},
		},
		IDToken: &IDTokenConfig{
			ValidityPeriod: 3600,
			UserAttributes: []string{"sub", "email", "name"},
			ScopeClaims: map[string][]string{
				"profile": {"name", "picture"},
			},
		},
	}

	appJSON, _ := json.Marshal(app)
	req, _ := http.NewRequest("PUT", testServerURL+"/applications/"+appID, bytes.NewReader(appJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)

	// Verify the updated config
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Equal(int64(7200), retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.ValidityPeriod)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken)
}

// TestApplicationWithPKCERequired tests creating application with PKCE requirement.
func (ts *ApplicationAPITestSuite) TestApplicationWithPKCERequired() {
	app := Application{
		Name:        "PKCE Required Test",
		Description: "Test PKCE required configuration",
		URL:         "https://pkce.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://pkce.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "none",
					PublicClient:            true,
					PKCERequired:            true,
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Verify PKCE configuration
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().True(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.PKCERequired)
	ts.Assert().True(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.PublicClient)
}

// TestApplicationListRetrievesMultiple tests listing multiple applications.
func (ts *ApplicationAPITestSuite) TestApplicationListRetrievesMultiple() {
	// Create multiple applications
	appIDs := make([]string, 0)

	for i := 0; i < 3; i++ {
		app := Application{
			Name:        fmt.Sprintf("List Test App %d", i),
			Description: fmt.Sprintf("Test application %d", i),
			URL:         fmt.Sprintf("https://listtest%d.example.com", i),
			Certificate: &ApplicationCert{Type: "NONE", Value: ""},
			InboundAuthConfig: []InboundAuthConfig{
				{
					Type: "oauth2",
					OAuthAppConfig: &OAuthAppConfig{
						RedirectURIs:            []string{fmt.Sprintf("https://listtest%d.example.com/callback", i)},
						GrantTypes:              []string{"authorization_code"},
						ResponseTypes:           []string{"code"},
						TokenEndpointAuthMethod: "client_secret_basic",
						Scopes:                  []string{"openid"},
					},
				},
			},
		}

		appID, err := createApplication(app)
		ts.Require().NoError(err)
		appIDs = append(appIDs, appID)
	}

	// Cleanup
	defer func() {
		for _, appID := range appIDs {
			deleteApplication(appID)
		}
	}()

	// List applications
	req, _ := http.NewRequest("GET", testServerURL+"/applications", nil)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)

	var listResponse ApplicationList
	json.NewDecoder(resp.Body).Decode(&listResponse)
	ts.Assert().GreaterOrEqual(listResponse.TotalResults, 3)
}

// TestApplicationUpdateCompleteOAuthConfig tests updating all OAuth fields.
func (ts *ApplicationAPITestSuite) TestApplicationUpdateCompleteOAuthConfig() {
	// Create with minimal config
	app := Application{
		Name:        "Complete OAuth Update Test",
		Description: "Test complete OAuth config update",
		URL:         "https://completeoauth.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://completeoauth.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Update with complete configuration
	app.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs = []string{
		"https://completeoauth.example.com/callback1",
		"https://completeoauth.example.com/callback2",
	}
	app.InboundAuthConfig[0].OAuthAppConfig.GrantTypes = []string{
		"authorization_code",
		"refresh_token",
	}
	app.InboundAuthConfig[0].OAuthAppConfig.Scopes = []string{
		"openid", "profile", "email", "address", "phone",
	}
	app.InboundAuthConfig[0].OAuthAppConfig.PKCERequired = true
	app.Certificate = &ApplicationCert{
		Type:  "JWKS_URI",
		Value: "https://completeoauth.example.com/.well-known/jwks.json",
	}

	appJSON, _ := json.Marshal(app)
	req, _ := http.NewRequest("PUT", testServerURL+"/applications/"+appID, bytes.NewReader(appJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)

	// Verify all updates
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs, 2)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.GrantTypes, 2)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Scopes, 5)
	ts.Assert().True(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.PKCERequired)
}

// Helper function to get application by ID
func getApplicationByID(appID string) (*Application, error) {
	req, err := http.NewRequest("GET", testServerURL+"/applications/"+appID, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get application: status %d", resp.StatusCode)
	}

	var app Application
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, err
	}

	return &app, nil
}

// TestApplicationWithOnlyAccessToken tests creating application with only AccessToken config.
func (ts *ApplicationAPITestSuite) TestApplicationWithOnlyAccessToken() {
	app := Application{
		Name:        "Only Access Token Test",
		Description: "Test with only access token config",
		URL:         "https://accesstokenonly.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://accesstokenonly.example.com/callback"},
					GrantTypes:              []string{"client_credentials"},
					ResponseTypes:           []string{},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"api.read", "api.write"},
					Token: &OAuthTokenConfig{
						Issuer: "https://accesstokenonly.example.com",
						AccessToken: &TokenConfig{
							Issuer:         "https://accesstokenonly.example.com",
							ValidityPeriod: 7200,
							UserAttributes: []string{"email", "username", "role", "department"},
						},
						// No IDToken
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Retrieve and verify AccessToken is configured properly
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken)
	ts.Assert().Equal(int64(7200), retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.ValidityPeriod)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.UserAttributes, 4)
}

// TestApplicationWithOnlyIDToken tests creating application with only IDToken config.
func (ts *ApplicationAPITestSuite) TestApplicationWithOnlyIDToken() {
	app := Application{
		Name:        "Only ID Token Test",
		Description: "Test with only ID token config",
		URL:         "https://idtokenonly.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://idtokenonly.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile"},
					Token: &OAuthTokenConfig{
						// No AccessToken
						IDToken: &IDTokenConfig{
							ValidityPeriod: 3600,
							UserAttributes: []string{"sub", "email", "name", "picture"},
							ScopeClaims: map[string][]string{
								"profile": {"name", "given_name", "family_name", "middle_name"},
								"email":   {"email", "email_verified"},
							},
						},
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Retrieve and verify IDToken is configured properly
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken)
	ts.Assert().Equal(int64(3600), retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ValidityPeriod)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims, 2)
}

// TestApplicationWithBothTokenTypes tests creating application with both AccessToken and IDToken.
func (ts *ApplicationAPITestSuite) TestApplicationWithBothTokenTypes() {
	app := Application{
		Name:        "Both Token Types Test",
		Description: "Test with both access and ID tokens",
		URL:         "https://bothtokens.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://bothtokens.example.com/callback"},
					GrantTypes:              []string{"authorization_code", "refresh_token"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_post",
					Scopes:                  []string{"openid", "profile", "email"},
					Token: &OAuthTokenConfig{
						Issuer: "https://bothtokens.example.com",
						AccessToken: &TokenConfig{
							Issuer:         "https://bothtokens.example.com",
							ValidityPeriod: 5400,
							UserAttributes: []string{"email", "username"},
						},
						IDToken: &IDTokenConfig{
							ValidityPeriod: 3600,
							UserAttributes: []string{"sub", "email"},
							ScopeClaims: map[string][]string{
								"profile": {"name"},
								"email":   {"email"},
							},
						},
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Retrieve and verify both tokens are present
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken)
	ts.Assert().Equal(int64(5400), retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.ValidityPeriod)
	ts.Assert().Equal(int64(3600), retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ValidityPeriod)
}

// TestApplicationUpdateRemoveOAuthConfig tests removing OAuth config from application.
func (ts *ApplicationAPITestSuite) TestApplicationUpdateRemoveOAuthConfig() {
	// Create app with OAuth config
	app := Application{
		Name:        "Remove OAuth Config Test",
		Description: "Test removing OAuth config",
		URL:         "https://removeoauth.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://removeoauth.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Update to remove OAuth config (empty InboundAuthConfig)
	app.InboundAuthConfig = []InboundAuthConfig{}
	appJSON, _ := json.Marshal(app)
	req, _ := http.NewRequest("PUT", testServerURL+"/applications/"+appID, bytes.NewReader(appJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)

	// Verify OAuth config was removed
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Len(retrievedApp.InboundAuthConfig, 0)
}

// TestApplicationWithMultipleGrantAndResponseTypes tests multiple grant/response types conversion.
func (ts *ApplicationAPITestSuite) TestApplicationWithMultipleGrantAndResponseTypes() {
	app := Application{
		Name:        "Multiple Grant Types Test",
		Description: "Test with multiple grant and response types",
		URL:         "https://multiplegrants.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs: []string{
						"https://multiplegrants.example.com/callback1",
						"https://multiplegrants.example.com/callback2",
						"https://multiplegrants.example.com/callback3",
					},
					GrantTypes: []string{
						"authorization_code",
						"refresh_token",
						"client_credentials",
					},
					ResponseTypes: []string{
						"code",
					},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile", "email", "address", "phone", "offline_access"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Retrieve and verify arrays were properly stored
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs, 3)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.GrantTypes, 3)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Scopes, 6)
}

// TestApplicationWithEmptyTokenIssuer tests token config with empty issuer.
func (ts *ApplicationAPITestSuite) TestApplicationWithEmptyTokenIssuer() {
	app := Application{
		Name:        "Empty Token Issuer Test",
		Description: "Test with empty token issuer",
		URL:         "https://emptyissuer.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://emptyissuer.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
					Token: &OAuthTokenConfig{
						// Empty Issuer
						AccessToken: &TokenConfig{
							ValidityPeriod: 3600,
						},
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Retrieve and verify
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token)
}

// TestApplicationWithMinimalTokenConfig tests minimal token configuration.
func (ts *ApplicationAPITestSuite) TestApplicationWithMinimalTokenConfig() {
	app := Application{
		Name:        "Minimal Token Config Test",
		Description: "Test with minimal token config",
		URL:         "https://minimaltoken.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://minimaltoken.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
					Token: &OAuthTokenConfig{
						Issuer: "https://minimaltoken.example.com",
						// No AccessToken or IDToken
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Retrieve and verify minimal token config
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token)
	ts.Assert().Equal("https://minimaltoken.example.com", retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.Issuer)
}

// TestApplicationWithComplexScopeClaims tests complex scope claims mapping.
func (ts *ApplicationAPITestSuite) TestApplicationWithComplexScopeClaims() {
	app := Application{
		Name:        "Complex Scope Claims Test",
		Description: "Test with complex scope claims",
		URL:         "https://complexscopes.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://complexscopes.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile", "email", "address", "phone", "custom"},
					Token: &OAuthTokenConfig{
						IDToken: &IDTokenConfig{
							ValidityPeriod: 3600,
							UserAttributes: []string{"sub", "email", "name"},
							ScopeClaims: map[string][]string{
								"profile": {
									"name", "given_name", "family_name", "middle_name",
									"nickname", "preferred_username", "profile", "picture",
									"website", "gender", "birthdate", "zoneinfo", "locale",
									"updated_at",
								},
								"email": {"email", "email_verified"},
								"address": {
									"address.formatted", "address.street_address",
									"address.locality", "address.region",
									"address.postal_code", "address.country",
								},
								"phone":  {"phone_number", "phone_number_verified"},
								"custom": {"organization", "department", "employee_id"},
							},
						},
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Retrieve and verify complex scope claims
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims, 5)
	ts.Assert().Contains(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims, "profile")
	ts.Assert().GreaterOrEqual(len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims["profile"]), 10)
}

// TestApplicationCertificateRollbackOnOAuthFail tests certificate rollback when OAuth creation fails.
func (ts *ApplicationAPITestSuite) TestApplicationCertificateRollbackOnOAuthFail() {
	// Try to create app with invalid OAuth config (should trigger rollback)
	app := Application{
		Name:        "Certificate Rollback Test",
		Description: "Test certificate rollback on OAuth failure",
		URL:         "https://rollback.example.com",
		Certificate: &ApplicationCert{
			Type:  "JWKS_URI",
			Value: "https://rollback.example.com/.well-known/jwks.json",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://rollback.example.com/callback#fragment"}, // Invalid - has fragment
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err) // Should fail due to fragment in redirect URI
}

// TestApplicationGetByName tests retrieving application by name.
func (ts *ApplicationAPITestSuite) TestApplicationGetByName() {
	uniqueName := fmt.Sprintf("Get By Name Test %d", time.Now().UnixNano())
	app := Application{
		Name:        uniqueName,
		Description: "Test get by name",
		URL:         "https://getbyname.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://getbyname.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Get by name using query parameter
	req, _ := http.NewRequest("GET", testServerURL+"/applications?name="+url.QueryEscape(uniqueName), nil)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)
}

// TestApplicationWithOAuthCertificateEmptyJWKSURI tests OAuth cert with empty JWKS_URI.
func (ts *ApplicationAPITestSuite) TestApplicationWithOAuthCertificateEmptyJWKSURI() {
	app := Application{
		Name:        "OAuth Empty JWKS URI Test",
		Description: "Test OAuth certificate with empty JWKS_URI",
		URL:         "https://oauthemptyjwksuri.example.com",
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://oauthemptyjwksuri.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
		Certificate: &ApplicationCert{
			Type:  "JWKS_URI",
			Value: "",
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err) // Should fail due to empty JWKS_URI
}

// TestApplicationValidationGrantTypeResponseTypeIncompat tests incompatible grant/response type combinations.
func (ts *ApplicationAPITestSuite) TestApplicationValidationGrantTypeResponseTypeIncompat() {
	// authorization_code without 'code' in response_types
	app := Application{
		Name:        "Grant Response Incompat Test",
		Description: "Test incompatible grant and response types",
		URL:         "https://grantresponseincompat.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://grantresponseincompat.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"token"}, // Wrong response type for authorization_code
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err) // Should fail due to incompatibility
}

// TestApplicationMultipleRedirectURIValidation tests multiple redirect URI validation.
func (ts *ApplicationAPITestSuite) TestApplicationMultipleRedirectURIValidation() {
	app := Application{
		Name:        "Multiple Redirect URI Validation Test",
		Description: "Test validation of multiple redirect URIs",
		URL:         "https://multiredirect.example.com",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs: []string{
						"https://multiredirect.example.com/callback1",
						"invalid-uri", // Invalid
						"https://multiredirect.example.com/callback3",
					},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	_, err := createApplication(app)
	ts.Assert().Error(err) // Should fail due to invalid redirect URI
}

// TestApplicationUpdateAddOAuthConfig tests adding OAuth config to existing app without it.
func (ts *ApplicationAPITestSuite) TestApplicationUpdateAddOAuthConfig() {
	// Create app without OAuth config
	app := Application{
		Name:              "Add OAuth Config Test",
		Description:       "Test adding OAuth config via update",
		URL:               "https://addoauth.example.com",
		Certificate:       &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{}, // No OAuth initially
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Update to add OAuth config
	app.InboundAuthConfig = []InboundAuthConfig{
		{
			Type: "oauth2",
			OAuthAppConfig: &OAuthAppConfig{
				RedirectURIs:            []string{"https://addoauth.example.com/callback"},
				GrantTypes:              []string{"authorization_code"},
				ResponseTypes:           []string{"code"},
				TokenEndpointAuthMethod: "client_secret_basic",
				Scopes:                  []string{"openid"},
			},
		},
	}

	appJSON, _ := json.Marshal(app)
	req, _ := http.NewRequest("PUT", testServerURL+"/applications/"+appID, bytes.NewReader(appJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)

	// Verify OAuth config was added
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Len(retrievedApp.InboundAuthConfig, 1)
	ts.Assert().NotEmpty(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.ClientID)
}

// TestApplicationTotalCountRetrieval tests getting total application count.
func (ts *ApplicationAPITestSuite) TestApplicationTotalCountRetrieval() {
	// Create a few apps
	appIDs := make([]string, 0)
	for i := 0; i < 2; i++ {
		app := Application{
			Name:        fmt.Sprintf("Count Test App %d", i),
			Description: "Test count",
			URL:         fmt.Sprintf("https://counttest%d.example.com", i),
			Certificate: &ApplicationCert{Type: "NONE", Value: ""},
			InboundAuthConfig: []InboundAuthConfig{
				{
					Type: "oauth2",
					OAuthAppConfig: &OAuthAppConfig{
						RedirectURIs:            []string{fmt.Sprintf("https://counttest%d.example.com/cb", i)},
						GrantTypes:              []string{"authorization_code"},
						ResponseTypes:           []string{"code"},
						TokenEndpointAuthMethod: "client_secret_basic",
						Scopes:                  []string{"openid"},
					},
				},
			},
		}
		appID, err := createApplication(app)
		ts.Require().NoError(err)
		appIDs = append(appIDs, appID)
	}

	// Cleanup
	defer func() {
		for _, appID := range appIDs {
			deleteApplication(appID)
		}
	}()

	// Get list to verify count
	req, _ := http.NewRequest("GET", testServerURL+"/applications", nil)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	var listResponse ApplicationList
	json.NewDecoder(resp.Body).Decode(&listResponse)
	ts.Assert().GreaterOrEqual(listResponse.TotalResults, 2)
}

// TestApplicationWithCompleteMetadata tests creating an application with all metadata fields.
func (ts *ApplicationAPITestSuite) TestApplicationWithCompleteMetadata() {
	app := Application{
		Name:        "Complete Metadata App",
		Description: "App with all metadata",
		URL:         "https://completemeta.example.com",
		LogoURL:     "https://completemeta.example.com/logo.png",
		TosURI:      "https://completemeta.example.com/tos",
		PolicyURI:   "https://completemeta.example.com/privacy",
		Contacts:    []string{"admin@completemeta.example.com", "support@completemeta.example.com"},
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		Token: &TokenConfig{
			Issuer:         "https://custom-issuer.example.com",
			ValidityPeriod: 7200,
			UserAttributes: []string{"email", "username", "groups"},
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://completemeta.example.com/callback"},
					GrantTypes:              []string{"authorization_code", "refresh_token"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile", "email"},
					Token: &OAuthTokenConfig{
						Issuer: "https://oauth-issuer.example.com",
						AccessToken: &TokenConfig{
							Issuer:         "https://oauth-issuer.example.com",
							ValidityPeriod: 3600,
							UserAttributes: []string{"sub", "email"},
						},
						IDToken: &IDTokenConfig{
							ValidityPeriod: 3600,
							UserAttributes: []string{"sub", "email", "name"},
							ScopeClaims: map[string][]string{
								"profile": {"name", "given_name", "family_name"},
								"email":   {"email", "email_verified"},
							},
						},
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Retrieve and verify all fields
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)

	// Verify basic fields
	ts.Assert().Equal("Complete Metadata App", retrievedApp.Name)
	ts.Assert().Equal("App with all metadata", retrievedApp.Description)
	ts.Assert().Equal("https://completemeta.example.com", retrievedApp.URL)
	ts.Assert().Equal("https://completemeta.example.com/logo.png", retrievedApp.LogoURL)

	// Verify metadata fields
	ts.Assert().Equal("https://completemeta.example.com/tos", retrievedApp.TosURI)
	ts.Assert().Equal("https://completemeta.example.com/privacy", retrievedApp.PolicyURI)
	ts.Assert().Equal([]string{"admin@completemeta.example.com", "support@completemeta.example.com"}, retrievedApp.Contacts)

	// Verify root token config
	ts.Require().NotNil(retrievedApp.Token)
	ts.Assert().Equal("https://custom-issuer.example.com", retrievedApp.Token.Issuer)
	ts.Assert().Equal(int64(7200), retrievedApp.Token.ValidityPeriod)
	ts.Assert().Equal([]string{"email", "username", "groups"}, retrievedApp.Token.UserAttributes)

	// Verify OAuth config fields
	ts.Require().Len(retrievedApp.InboundAuthConfig, 1)
	ts.Assert().Equal([]string{"https://completemeta.example.com/callback"}, retrievedApp.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs)
	ts.Assert().Equal([]string{"authorization_code", "refresh_token"}, retrievedApp.InboundAuthConfig[0].OAuthAppConfig.GrantTypes)
	ts.Assert().Equal([]string{"code"}, retrievedApp.InboundAuthConfig[0].OAuthAppConfig.ResponseTypes)
	ts.Assert().Equal("client_secret_basic", retrievedApp.InboundAuthConfig[0].OAuthAppConfig.TokenEndpointAuthMethod)
	ts.Assert().Equal([]string{"openid", "profile", "email"}, retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Scopes)

	// Verify OAuth token config
	ts.Require().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token)
	ts.Assert().Equal("https://oauth-issuer.example.com", retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.Issuer)

	// Verify access token config
	ts.Require().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken)
	ts.Assert().Equal(int64(3600), retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.ValidityPeriod)
	ts.Assert().Equal([]string{"sub", "email"}, retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.UserAttributes)

	// Verify ID token config
	ts.Require().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken)
	ts.Assert().Equal(int64(3600), retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ValidityPeriod)
	ts.Assert().Equal([]string{"sub", "email", "name"}, retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.UserAttributes)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims)
	ts.Assert().Equal([]string{"name", "given_name", "family_name"}, retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims["profile"])
}

// TestApplicationWithOnlyRootToken tests app with only root token config.
func (ts *ApplicationAPITestSuite) TestApplicationWithOnlyRootToken() {
	app := Application{
		Name:        "Root Token Only App",
		Description: "App with only root token",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		Token: &TokenConfig{
			Issuer:         "https://root-issuer.example.com",
			ValidityPeriod: 5400,
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Require().NotNil(retrievedApp.Token)
	ts.Assert().Equal("https://root-issuer.example.com", retrievedApp.Token.Issuer)
	ts.Assert().Equal(int64(5400), retrievedApp.Token.ValidityPeriod)
}

// TestApplicationUpdateMetadataFields tests updating metadata fields.
func (ts *ApplicationAPITestSuite) TestApplicationUpdateMetadataFields() {
	// Create initial app
	app := Application{
		Name:        "Update Metadata App",
		Description: "Initial description",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://updatemeta.example.com/cb"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Update with metadata
	app.Description = "Updated description"
	app.TosURI = "https://updatemeta.example.com/tos"
	app.PolicyURI = "https://updatemeta.example.com/privacy"
	app.Contacts = []string{"contact@updatemeta.example.com"}
	app.LogoURL = "https://updatemeta.example.com/logo.png"

	payload, _ := json.Marshal(app)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/applications/%s", testServerURL, appID), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)

	// Verify updates
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Equal("Updated description", retrievedApp.Description)
	ts.Assert().Equal("https://updatemeta.example.com/tos", retrievedApp.TosURI)
	ts.Assert().Equal("https://updatemeta.example.com/privacy", retrievedApp.PolicyURI)
	ts.Assert().Equal([]string{"contact@updatemeta.example.com"}, retrievedApp.Contacts)
	ts.Assert().Equal("https://updatemeta.example.com/logo.png", retrievedApp.LogoURL)
}

// TestApplicationPublicClientWithoutSecret tests public client creation without client secret.
func (ts *ApplicationAPITestSuite) TestApplicationPublicClientWithoutSecret() {
	app := Application{
		Name:        "Public Client No Secret",
		Description: "Public client without secret",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://public-nosecret.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "none",
					PublicClient:            true,
					PKCERequired:            true,
					Scopes:                  []string{"openid", "profile"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().True(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.PublicClient)
	ts.Assert().True(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.PKCERequired)
	ts.Assert().Equal("none", string(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.TokenEndpointAuthMethod))
}

// TestApplicationWithRefreshTokenGrant tests app with refresh_token grant.
func (ts *ApplicationAPITestSuite) TestApplicationWithRefreshTokenGrant() {
	app := Application{
		Name:        "Refresh Token App",
		Description: "App with refresh token grant",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://refreshtoken.example.com/callback"},
					GrantTypes:              []string{"authorization_code", "refresh_token"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "offline_access"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Contains(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.GrantTypes, "authorization_code")
	ts.Assert().Contains(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.GrantTypes, "refresh_token")
}

// TestApplicationUpdateTokenConfiguration tests updating token configuration.
func (ts *ApplicationAPITestSuite) TestApplicationUpdateTokenConfiguration() {
	// Create app with initial token config
	app := Application{
		Name:        "Update Token Config App",
		Description: "App to update token config",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		Token: &TokenConfig{
			Issuer:         "https://initial-issuer.example.com",
			ValidityPeriod: 3600,
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://updatetoken.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	// Update token config
	app.Token.Issuer = "https://updated-issuer.example.com"
	app.Token.ValidityPeriod = 7200
	app.Token.UserAttributes = []string{"email", "username"}

	payload, _ := json.Marshal(app)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/applications/%s", testServerURL, appID), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	ts.Assert().Equal(http.StatusOK, resp.StatusCode)

	// Verify token config update
	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Require().NotNil(retrievedApp.Token)
	ts.Assert().Equal("https://updated-issuer.example.com", retrievedApp.Token.Issuer)
	ts.Assert().Equal(int64(7200), retrievedApp.Token.ValidityPeriod)
	ts.Assert().Equal([]string{"email", "username"}, retrievedApp.Token.UserAttributes)
}

// TestApplicationWithEmptyContacts tests app with empty contacts array.
func (ts *ApplicationAPITestSuite) TestApplicationWithEmptyContacts() {
	app := Application{
		Name:        "Empty Contacts App",
		Description: "App with empty contacts",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		Contacts:    []string{},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Empty(retrievedApp.Contacts)
}

// TestApplicationClientCredentialsGrant tests app with client_credentials grant.
func (ts *ApplicationAPITestSuite) TestApplicationClientCredentialsGrant() {
	app := Application{
		Name:        "Client Credentials App",
		Description: "App with client credentials grant",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					GrantTypes:              []string{"client_credentials"},
					ResponseTypes:           []string{},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"api:read", "api:write"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Equal([]string{"client_credentials"}, retrievedApp.InboundAuthConfig[0].OAuthAppConfig.GrantTypes)
	ts.Assert().Empty(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs)
}

// TestApplicationWithIDTokenScopeClaimsOnly tests app with only ID token scope claims.
func (ts *ApplicationAPITestSuite) TestApplicationWithIDTokenScopeClaimsOnly() {
	app := Application{
		Name:        "ID Token Scope Claims App",
		Description: "App with ID token scope claims only",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://idtoken-scope.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile"},
					Token: &OAuthTokenConfig{
						IDToken: &IDTokenConfig{
							ValidityPeriod: 3600,
							ScopeClaims: map[string][]string{
								"profile": {"name", "picture"},
								"email":   {"email"},
							},
						},
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Require().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token)
	ts.Require().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken)
	ts.Assert().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims)
	ts.Assert().Equal([]string{"name", "picture"}, retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ScopeClaims["profile"])
}

// TestApplicationWithOAuthTokenIssuerOnly tests app with only OAuth token issuer.
func (ts *ApplicationAPITestSuite) TestApplicationWithOAuthTokenIssuerOnly() {
	app := Application{
		Name:        "OAuth Token Issuer Only App",
		Description: "App with only OAuth token issuer",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs:            []string{"https://oauth-issuer-only.example.com/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid"},
					Token: &OAuthTokenConfig{
						Issuer: "https://custom-oauth-issuer.example.com",
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Require().NotNil(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token)
	ts.Assert().Equal("https://custom-oauth-issuer.example.com", retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Token.Issuer)
}

// TestApplicationGetByNonExistentID tests retrieving app by non-existent ID.
func (ts *ApplicationAPITestSuite) TestApplicationGetByNonExistentID() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, err := getApplicationByID(nonExistentID)
	ts.Assert().Error(err)
}

// TestApplicationWithMultipleRedirectURIsAndScopes tests app with multiple redirect URIs and scopes.
func (ts *ApplicationAPITestSuite) TestApplicationWithMultipleRedirectURIsAndScopes() {
	app := Application{
		Name:        "Multiple URIs and Scopes App",
		Description: "App with multiple redirect URIs and scopes",
		Certificate: &ApplicationCert{Type: "NONE", Value: ""},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					RedirectURIs: []string{
						"https://multi-uris.example.com/callback1",
						"https://multi-uris.example.com/callback2",
						"https://multi-uris.example.com/callback3",
					},
					GrantTypes:              []string{"authorization_code", "refresh_token"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					Scopes:                  []string{"openid", "profile", "email", "address", "phone"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err)
	defer deleteApplication(appID)

	retrievedApp, err := getApplicationByID(appID)
	ts.Require().NoError(err)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs, 3)
	ts.Assert().Len(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Scopes, 5)
	ts.Assert().Contains(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Scopes, "address")
	ts.Assert().Contains(retrievedApp.InboundAuthConfig[0].OAuthAppConfig.Scopes, "phone")
}
