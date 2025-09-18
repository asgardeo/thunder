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

package testutils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	TestServerURL = "https://localhost:8095"
)

// getHTTPClient returns a configured HTTP client for test requests
func getHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// CreateUser creates a user via API and returns the user ID
func CreateUser(user User) (string, error) {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user: %w", err)
	}

	req, err := http.NewRequest("POST", TestServerURL+"/users", bytes.NewReader(userJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdUser map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdUser)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	userID, ok := createdUser["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	return userID, nil
}

// DeleteUser deletes a user by ID
func DeleteUser(userID string) error {
	req, err := http.NewRequest("DELETE", TestServerURL+"/users/"+userID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("expected status 204, got %d", resp.StatusCode)
	}
	return nil
}

// CreateMultipleUsers creates multiple users and returns their IDs
func CreateMultipleUsers(users ...User) ([]string, error) {
	var userIDs []string

	for i, user := range users {
		userID, err := CreateUser(user)
		if err != nil {
			// Cleanup already created users on failure
			for _, createdID := range userIDs {
				DeleteUser(createdID)
			}
			return nil, fmt.Errorf("failed to create user %d: %w", i, err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// CleanupUsers deletes multiple users
func CleanupUsers(userIDs []string) error {
	var errs []error

	for _, userID := range userIDs {
		if userID != "" {
			if err := DeleteUser(userID); err != nil {
				errs = append(errs, fmt.Errorf("failed to delete user %s: %w", userID, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// CreateApplication creates an application via API and returns the application ID
func CreateApplication(app Application) (string, error) {
	// Convert Application to the format expected by the application API
	appData := map[string]interface{}{
		"name":                         app.Name,
		"description":                  app.Description,
		"is_registration_flow_enabled": app.IsRegistrationFlowEnabled,
		"auth_flow_graph_id":           app.AuthFlowGraphID,
		"registration_flow_graph_id":   app.RegistrationFlowGraphID,
		"certificate": map[string]interface{}{
			"type":  "NONE",
			"value": "",
		},
		"inbound_auth_config": []map[string]interface{}{
			{
				"type": "oauth2",
				"oauth_app_config": map[string]interface{}{
					"client_id":     app.ClientID,
					"client_secret": app.ClientSecret,
					"redirect_uris": app.RedirectURIs,
				},
			},
		},
	}

	appJSON, err := json.Marshal(appData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal application: %w", err)
	}

	req, err := http.NewRequest("POST", TestServerURL+"/applications", bytes.NewReader(appJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	var createdApp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdApp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	appID, ok := createdApp["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	return appID, nil
}

// DeleteApplication deletes an application by ID
func DeleteApplication(appID string) error {
	req, err := http.NewRequest("DELETE", TestServerURL+"/applications/"+appID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status 204, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}
	return nil
}

// CreateOrganizationUnit creates an organization unit via API and returns the OU ID
func CreateOrganizationUnit(ou OrganizationUnit) (string, error) {
	ouJSON, err := json.Marshal(ou)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OU request: %w", err)
	}

	req, err := http.NewRequest("POST", TestServerURL+"/organization-units", bytes.NewReader(ouJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	var createdOU map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdOU)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	ouID, ok := createdOU["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	return ouID, nil
}

// DeleteOrganizationUnit deletes an organization unit by ID
func DeleteOrganizationUnit(ouID string) error {
	req, err := http.NewRequest("DELETE", TestServerURL+"/organization-units/"+ouID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete organization unit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status 200 or 204, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}
	return nil
}

// CreateIDP creates an identity provider via API and returns the IDP ID
func CreateIDP(idp IDP) (string, error) {
	idpJSON, err := json.Marshal(idp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal IDP: %w", err)
	}

	req, err := http.NewRequest("POST", TestServerURL+"/identity-providers", bytes.NewReader(idpJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	var createdIDP map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdIDP)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	idpID, ok := createdIDP["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	return idpID, nil
}

// DeleteIDP deletes an identity provider by ID
func DeleteIDP(idpID string) error {
	req, err := http.NewRequest("DELETE", TestServerURL+"/identity-providers/"+idpID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete identity provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status 200 or 204, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}
	return nil
}

// CreateLocalIDP creates a Local identity provider for testing
func CreateLocalIDP() (string, error) {
	localIDP := IDP{
		Name:        "Local",
		Description: "Local Identity Provider for testing",
		Properties: []IDPProperty{
			{Name: "type", Value: "local", IsSecret: false},
		},
	}
	return CreateIDP(localIDP)
}

// GetUserAttributes extracts user attributes from JSON into a map
func GetUserAttributes(user User) (map[string]interface{}, error) {
	var userAttrs map[string]interface{}
	err := json.Unmarshal(user.Attributes, &userAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user attributes: %w", err)
	}
	return userAttrs, nil
}

// FindUserByAttribute retrieves all users and returns the user with a matching attribute key and value
func FindUserByAttribute(key, value string) (*User, error) {
	client := getHTTPClient()

	req, err := http.NewRequest("GET", TestServerURL+"/users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user list request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send user list request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user list, status: %d", resp.StatusCode)
	}

	var userListResponse UserListResponse
	err = json.NewDecoder(resp.Body).Decode(&userListResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user list response: %w", err)
	}

	for _, user := range userListResponse.Users {
		attrs, err := GetUserAttributes(user)

		if err != nil {
			continue
		}
		if v, ok := attrs[key]; ok && v == value {
			return &user, nil
		}
	}
	return nil, nil
}
