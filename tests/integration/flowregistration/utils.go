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

package flowregistration

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const testServerURL = "https://localhost:8095"

// Helper function to initiate the registration flow
func initiateRegistrationFlow(appID string, inputs map[string]string) (*FlowStep, error) {
	flowReqBody := map[string]interface{}{
		"applicationId": appID,
		"flowType":      "REGISTRATION",
	}
	if len(inputs) > 0 {
		flowReqBody["inputs"] = inputs
	}

	reqBody, err := json.Marshal(flowReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/flow/execute", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create flow request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send flow request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var flowStep FlowStep
	err = json.NewDecoder(resp.Body).Decode(&flowStep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &flowStep, nil
}

// Helper function to complete the registration flow
func completeRegistrationFlow(flowID string, actionID string, inputs map[string]string) (*FlowStep, error) {
	flowReqBody := map[string]interface{}{
		"flowId": flowID,
		"inputs": inputs,
	}
	if actionID != "" {
		flowReqBody["actionId"] = actionID
	}

	reqBody, err := json.Marshal(flowReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/flow/execute", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create flow request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send flow request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var flowStep FlowStep
	err = json.NewDecoder(resp.Body).Decode(&flowStep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &flowStep, nil
}

// Helper function to create multiple users
func CreateMultipleUsers(users ...User) ([]string, error) {
	var userIDs []string

	for _, user := range users {
		userID, err := createUser(user)
		if err != nil {
			// If error occurs, cleanup already created users
			CleanupUsers(userIDs)
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// Helper function to create a single user
func createUser(user User) (string, error) {
	reqBody, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user request body: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/users", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create user request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send user creation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create user, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdUser User
	err = json.NewDecoder(resp.Body).Decode(&createdUser)
	if err != nil {
		return "", fmt.Errorf("failed to parse user response body: %w", err)
	}

	return createdUser.Id, nil
}

// Helper function to cleanup users
func CleanupUsers(userIDs []string) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var lastErr error
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}

		req, err := http.NewRequest("DELETE", testServerURL+"/users/"+userID, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			lastErr = fmt.Errorf("failed to delete user %s, status: %d", userID, resp.StatusCode)
		}
	}

	return lastErr
}

// getAppConfig retrieves the current application configuration
func getAppConfig(appID string) (map[string]interface{}, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/applications/%s", testServerURL, appID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var appConfig map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&appConfig); err != nil {
		return nil, fmt.Errorf("failed to parse app config: %w", err)
	}

	return appConfig, nil
}

// updateAppConfig updates the application configuration with the specified auth flow graph ID
func updateAppConfig(appID string, authFlowGraphID, regFlowGraphID string) error {
	appConfig, err := getAppConfig(appID)
	if err != nil {
		return fmt.Errorf("failed to get current app config: %w", err)
	}

	appConfig["auth_flow_graph_id"] = authFlowGraphID
	appConfig["client_secret"] = "secret123"
	if regFlowGraphID != "" {
		appConfig["registration_flow_graph_id"] = regFlowGraphID
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	jsonPayload, err := json.Marshal(appConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/applications/%s", testServerURL, appID),
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Helper function to restore application configuration
func RestoreAppConfig(appID string, originalConfig map[string]interface{}) error {
	if originalConfig == nil {
		return fmt.Errorf("no original config to restore")
	}

	// Add client secret to original config for restoration
	originalConfig["client_secret"] = "secret123"

	reqBody, err := json.Marshal(originalConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal original app config: %w", err)
	}

	req, err := http.NewRequest("PUT", testServerURL+"/applications/"+appID, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create app config request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send app config request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to restore app config, status: %d", resp.StatusCode)
	}

	return nil
}

// Helper function to validate required inputs
func ValidateRequiredInputs(inputs []InputData, required []string) bool {
	requiredMap := make(map[string]bool)
	for _, req := range required {
		requiredMap[req] = false
	}

	for _, input := range inputs {
		if input.Required {
			if _, exists := requiredMap[input.Name]; exists {
				requiredMap[input.Name] = true
			}
		}
	}

	for _, found := range requiredMap {
		if !found {
			return false
		}
	}

	return true
}

// Helper function to check if a specific input exists
func HasInput(inputs []InputData, inputName string) bool {
	for _, input := range inputs {
		if input.Name == inputName {
			return true
		}
	}
	return false
}

// Helper function to generate unique usernames
func generateUniqueUsername(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// Helper function to get user attributes
func GetUserAttributes(user User) (map[string]interface{}, error) {
	var attrs map[string]interface{}
	err := json.Unmarshal(user.Attributes, &attrs)
	return attrs, err
}

// NotificationSenderRequest represents the request to create a message notification sender.
type NotificationSenderRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Provider    string           `json:"provider"`
	Properties  []SenderProperty `json:"properties"`
}

// NotificationSender represents a message notification sender.
type NotificationSender struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Provider    string           `json:"provider"`
	Properties  []SenderProperty `json:"properties"`
}

// SenderProperty represents a key-value property for a message notification sender.
type SenderProperty struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// Helper function to create a notification sender
func CreateNotificationSender(mockServerPort int, senderName string) (string, error) {
	senderReq := NotificationSenderRequest{
		Name:        senderName,
		Description: "SMS sender for registration tests",
		Provider:    "custom",
		Properties: []SenderProperty{
			{
				Name:     "url",
				Value:    fmt.Sprintf("http://localhost:%d/send-sms", mockServerPort),
				IsSecret: false,
			},
			{
				Name:     "http_method",
				Value:    "POST",
				IsSecret: false,
			},
			{
				Name:     "content_type",
				Value:    "JSON",
				IsSecret: false,
			},
		},
	}

	reqBody, err := json.Marshal(senderReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sender request body: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/notification-senders/message", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create sender request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send sender creation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create sender, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdSender NotificationSender
	err = json.NewDecoder(resp.Body).Decode(&createdSender)
	if err != nil {
		return "", fmt.Errorf("failed to parse sender response body: %w", err)
	}

	return createdSender.ID, nil
}

// Helper function to delete a notification sender
func DeleteNotificationSender(senderID string) error {
	req, err := http.NewRequest("DELETE", testServerURL+"/notification-senders/message/"+senderID, nil)
	if err != nil {
		return fmt.Errorf("failed to create sender deletion request: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send sender deletion request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete sender, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// FindUserByAttribute retrieves all users and returns the user with a matching attribute key and value
func FindUserByAttribute(key, value string) (*User, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/users", nil)
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

// Helper function to create an organization unit
func createOrganizationUnit(ouRequest TestOrganizationUnit) (string, error) {
	ouJSON, err := json.Marshal(ouRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OU request: %w", err)
	}

	reqBody := bytes.NewReader(ouJSON)
	req, err := http.NewRequest("POST", testServerURL+"/organization-units", reqBody)
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

	var createdOU map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdOU)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	id, ok := createdOU["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	return id, nil
}

// Helper function to delete an organization unit
func deleteOrganizationUnit(ouID string) error {
	req, err := http.NewRequest("DELETE", testServerURL+"/organization-units/"+ouID, nil)
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("expected status 200 or 204, got %d", resp.StatusCode)
	}
	return nil
}

// Helper function to create an application
func createApplication(appRequest TestApplication) (string, error) {
	// Convert TestApplication to the format expected by the application API
	appData := map[string]interface{}{
		"name":                         appRequest.Name,
		"description":                  appRequest.Description,
		"is_registration_flow_enabled": appRequest.IsRegistrationFlowEnabled,
		"auth_flow_graph_id":           appRequest.AuthFlowGraphID,
		"registration_flow_graph_id":   appRequest.RegistrationFlowGraphID,
		"certificate": map[string]interface{}{
			"type":  "NONE",
			"value": "",
		},
		"inbound_auth_config": []map[string]interface{}{
			{
				"type": "oauth2",
				"oauth_app_config": map[string]interface{}{
					"client_id":     appRequest.ClientID,
					"client_secret": appRequest.ClientSecret,
					"redirect_uris": appRequest.RedirectURIs,
				},
			},
		},
	}

	appJSON, err := json.Marshal(appData)
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

	var createdApp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdApp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	id, ok := createdApp["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	return id, nil
}

// Helper function to delete an application
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

	if resp.StatusCode != http.StatusNoContent {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status 204, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}
	return nil
}

// createIdp creates an identity provider via API
func createIdp(idp IDP) (string, error) {
	idpJSON, err := json.Marshal(idp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal IDP: %w", err)
	}

	reqBody := bytes.NewReader(idpJSON)
	req, err := http.NewRequest("POST", testServerURL+"/identity-providers", reqBody)
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

	var createdIdp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdIdp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	id, ok := createdIdp["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	return id, nil
}

// deleteIdp deletes an identity provider via API
func deleteIdp(idpID string) error {
	req, err := http.NewRequest("DELETE", testServerURL+"/identity-providers/"+idpID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status 200 or 204, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}
	return nil
}

// createLocalIdp creates a Local identity provider for testing
func createLocalIdp() (string, error) {
	localIdp := IDP{
		Name:        "Local",
		Description: "Local Identity Provider for testing",
		Properties: []IDPProperty{
			{Name: "type", Value: "local", IsSecret: false},
		},
	}
	return createIdp(localIdp)
}
