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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
)

const testServerURL = "https://localhost:8095"

// Helper function to initiate the registration flow
func initiateRegistrationFlow(appID string, inputs map[string]string, action ...string) (*FlowStep, error) {
	flowReqBody := map[string]interface{}{
		"applicationId": appID,
		"flowType":      "REGISTRATION",
	}
	if len(inputs) > 0 {
		flowReqBody["inputs"] = inputs
	}
	if len(action) > 0 && action[0] != "" {
		flowReqBody["action"] = action[0]
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

	client := testutils.GetHTTPClient()

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
		flowReqBody["action"] = actionID
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

	client := testutils.GetHTTPClient()

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

// getAppConfig retrieves the current application configuration
func getAppConfig(appID string) (map[string]interface{}, error) {
	client := testutils.GetHTTPClient()

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

	client := testutils.GetHTTPClient()

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

	client := testutils.GetHTTPClient()

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
func ValidateRequiredInputs(inputs []Inputs, required []string) bool {
	requiredMap := make(map[string]bool)
	for _, req := range required {
		requiredMap[req] = false
	}

	for _, input := range inputs {
		if input.Required {
			if _, exists := requiredMap[input.Identifier]; exists {
				requiredMap[input.Identifier] = true
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
func HasInput(inputs []Inputs, inputName string) bool {
	for _, input := range inputs {
		if input.Identifier == inputName {
			return true
		}
	}
	return false
}

// Helper function to generate unique usernames
func generateUniqueUsername(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
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

	client := testutils.GetHTTPClient()

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

	client := testutils.GetHTTPClient()

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
