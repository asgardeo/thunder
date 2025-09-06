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

package flowauthn

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const testServerURL = "https://localhost:8095"

// Helper function to initiate the authentication flow
func initiateAuthFlow(appID string, inputs map[string]string) (*FlowStep, error) {
	flowReqBody := map[string]interface{}{
		"applicationId": appID,
		"flowType":      "AUTHENTICATION",
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

// Helper function to initiate the authentication flow with error handling
func initiateAuthFlowWithError(appID string, inputs map[string]string) (*ErrorResponse, error) {
	flowReqBody := map[string]interface{}{
		"applicationId": appID,
		"flowType":      "AUTHENTICATION",
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

	if resp.StatusCode != http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var errorResponse ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse error response body: %w", err)
	}

	return &errorResponse, nil
}

// Helper function to complete the authentication flow
func completeAuthFlow(flowID string, actionID string, inputs map[string]string) (*FlowStep, error) {
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

// Helper function to complete the authentication flow with error handling
func completeAuthFlowWithError(flowID string, inputs map[string]string) (*ErrorResponse, error) {
	flowReqBody := map[string]interface{}{
		"flowId": flowID,
		"inputs": inputs,
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

	if resp.StatusCode != http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var errorResponse ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse error response body: %w", err)
	}

	return &errorResponse, nil
}

// Helper function to create a user
func createUser(user User) (string, error) {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user: %w", err)
	}

	reqBody := bytes.NewReader(userJSON)
	req, err := http.NewRequest("POST", testServerURL+"/users", reqBody)
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

// Helper function to delete a user
func deleteUser(userID string) error {
	req, err := http.NewRequest("DELETE", testServerURL+"/users/"+userID, nil)
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
		return fmt.Errorf("failed to delete user, status code: %d", resp.StatusCode)
	}
	return nil
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
func updateAppConfig(appID string, authFlowGraphID string) error {
	appConfig, err := getAppConfig(appID)
	if err != nil {
		return fmt.Errorf("failed to get current app config: %w", err)
	}

	appConfig["auth_flow_graph_id"] = authFlowGraphID
	appConfig["client_secret"] = "secret123"

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

// TestSuiteConfig holds common configuration for test suites
type TestSuiteConfig struct {
	CreatedUserIDs    []string
	CreatedSenderID   string
	OriginalAppConfig map[string]interface{}
	MockServer        interface{} // Can be cast to specific mock server type
}

// CreateNotificationSender creates a custom notification sender for testing
func CreateNotificationSender(serverPort int, senderName string) (string, error) {
	return CreateNotificationSenderWithURL(fmt.Sprintf("http://localhost:%d/send-sms", serverPort), senderName)
}

// CreateNotificationSenderWithURL creates a custom notification sender with a specific URL
func CreateNotificationSenderWithURL(senderURL, senderName string) (string, error) {
	senderRequest := map[string]interface{}{
		"name":        senderName,
		"description": "Custom SMS sender for integration tests",
		"provider":    "custom",
		"properties": []map[string]interface{}{
			{
				"name":      "url",
				"value":     senderURL,
				"is_secret": false,
			},
			{
				"name":      "http_method",
				"value":     "POST",
				"is_secret": false,
			},
			{
				"name":      "content_type",
				"value":     "JSON",
				"is_secret": false,
			},
		},
	}

	jsonPayload, err := json.Marshal(senderRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sender request: %w", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", testServerURL+"/notification-senders/message", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create sender request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sender request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("sender creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var sender map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&sender); err != nil {
		return "", fmt.Errorf("failed to parse sender response: %w", err)
	}

	senderID, ok := sender["id"].(string)
	if !ok {
		return "", fmt.Errorf("sender ID not found in response")
	}

	return senderID, nil
}

// DeleteNotificationSender deletes a notification sender
func DeleteNotificationSender(senderID string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("DELETE", testServerURL+"/notification-senders/message/"+senderID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CreateMultipleUsers creates multiple test users and returns their IDs
func CreateMultipleUsers(users ...User) ([]string, error) {
	var userIDs []string

	for i, user := range users {
		id, err := createUser(user)
		if err != nil {
			// Cleanup already created users on failure
			for _, createdID := range userIDs {
				deleteUser(createdID)
			}
			return nil, fmt.Errorf("failed to create user %d: %w", i, err)
		}
		userIDs = append(userIDs, id)
	}

	return userIDs, nil
}

// CleanupUsers deletes multiple users
func CleanupUsers(userIDs []string) error {
	var errs []error

	for _, userID := range userIDs {
		if userID != "" {
			if err := deleteUser(userID); err != nil {
				errs = append(errs, fmt.Errorf("failed to delete user %s: %w", userID, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// RestoreAppConfig restores the original application configuration
func RestoreAppConfig(appID string, originalConfig map[string]interface{}) error {
	if originalConfig == nil {
		return fmt.Errorf("no original config to restore")
	}

	// Add client secret to original config for restoration
	originalConfig["client_secret"] = "secret123"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	jsonPayload, err := json.Marshal(originalConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal original config: %w", err)
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/applications/%s", testServerURL, appID),
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return fmt.Errorf("failed to create restore request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("restore request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("restore failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ValidateRequiredInputs validates that the expected input names are present in the flow data
func ValidateRequiredInputs(actualInputs []InputData, expectedInputNames []string) bool {
	inputMap := make(map[string]bool)
	for _, input := range actualInputs {
		inputMap[input.Name] = true
	}

	for _, expectedName := range expectedInputNames {
		if !inputMap[expectedName] {
			return false
		}
	}

	return true
}

// ValidateRequiredActions validates that the expected action IDs are present in the flow data
func ValidateRequiredActions(actualActions []FlowAction, expectedActionIDs []string) bool {
	actionMap := make(map[string]bool)
	for _, action := range actualActions {
		actionMap[action.ID] = true
	}

	for _, expectedID := range expectedActionIDs {
		if !actionMap[expectedID] {
			return false
		}
	}

	return true
}

// HasInput checks if a specific input is present in the flow data
func HasInput(inputs []InputData, inputName string) bool {
	for _, input := range inputs {
		if input.Name == inputName {
			return true
		}
	}
	return false
}

// HasAction checks if a specific action is present in the flow data
func HasAction(actions []FlowAction, actionID string) bool {
	for _, action := range actions {
		if action.ID == actionID {
			return true
		}
	}
	return false
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

// WaitAndValidateNotification waits for a notification to be sent and validates it
// This is a generic helper that can be used with different mock server types
func WaitAndValidateNotification(mockServer interface{}, expectedCount int, timeoutSeconds int) error {
	// This would need to be implemented based on the specific mock server interface
	// For now, we'll return a placeholder
	return fmt.Errorf("notification validation not implemented - should be customized per mock server type")
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

// deleteIdp deletes an identity provider via API
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
