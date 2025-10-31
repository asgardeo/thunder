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

package authz

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	testServerURL = "https://localhost:8095"
)

// initiateAuthorizationFlow starts the OAuth2 authorization flow
func initiateAuthorizationFlow(clientID, redirectURI, responseType, scope, state string) (*http.Response, error) {
	return initiateAuthorizationFlowWithResource(clientID, redirectURI, responseType, scope, state, "")
}

// initiateAuthorizationFlowWithResource starts the OAuth2 authorization flow with resource parameter
func initiateAuthorizationFlowWithResource(clientID, redirectURI, responseType, scope, state, resource string) (*http.Response, error) {
	authURL := testServerURL + "/oauth2/authorize"
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", responseType)
	params.Set("scope", scope)
	params.Set("state", state)
	if resource != "" {
		params.Set("resource", resource)
	}

	req, err := http.NewRequest("GET", authURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization request: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send authorization request: %w", err)
	}

	return resp, nil
}

// ExecuteAuthenticationFlow executes an authentication flow and returns the flow step
func ExecuteAuthenticationFlow(applicationId string, inputs map[string]string) (*FlowStep, error) {
	flowData := map[string]interface{}{
		"applicationId": applicationId,
		"flowType":      "AUTHENTICATION",
	}
	if len(inputs) > 0 {
		flowData["inputs"] = inputs
	}

	flowJSON, err := json.Marshal(flowData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow data: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/flow/execute", bytes.NewBuffer(flowJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create flow request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute flow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("flow execution failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var flowStep FlowStep
	err = json.NewDecoder(resp.Body).Decode(&flowStep)
	if err != nil {
		return nil, fmt.Errorf("failed to decode flow response: %w", err)
	}

	return &flowStep, nil
}

// completeAuthorization completes the authorization using the assertion
func completeAuthorization(sessionDataKey, assertion string) (*AuthorizationResponse, error) {
	authzData := map[string]interface{}{
		"sessionDataKey": sessionDataKey,
		"assertion":      assertion,
	}

	authzJSON, err := json.Marshal(authzData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authorization data: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/oauth2/authorize", bytes.NewBuffer(authzJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization completion request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to complete authorization: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authorization completion failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var authzResponse AuthorizationResponse
	err = json.NewDecoder(resp.Body).Decode(&authzResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode authorization response: %w", err)
	}

	return &authzResponse, nil
}

// requestToken performs a token request and returns raw HTTP result for both success and failure scenarios.
// grantType, code, and redirectURI are sent in the form body, while client credentials are sent via HTTP Basic Auth header.
func requestToken(clientID, clientSecret, code, redirectURI, grantType string) (*TokenHTTPResult, error) {
	tokenURL := testServerURL + "/oauth2/token"
	tokenData := url.Values{}

	tokenData.Set("grant_type", grantType)
	tokenData.Set("code", code)
	tokenData.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(tokenData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if clientID != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	result := &TokenHTTPResult{
		StatusCode: resp.StatusCode,
		Body:       body,
	}

	// Only try to decode token response if status is 200
	if resp.StatusCode == http.StatusOK {
		var tokenResponse TokenResponse
		if err := json.Unmarshal(body, &tokenResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal token response: %w", err)
		}
		result.Token = &tokenResponse
	}

	return result, nil
}

// extractAuthorizationCode extracts the authorization code from the redirect URI
func extractAuthorizationCode(redirectURI string) (string, error) {
	parsedURL, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("failed to parse redirect URI: %w", err)
	}

	code := parsedURL.Query().Get("code")
	if code == "" {
		return "", fmt.Errorf("authorization code not found in redirect URI")
	}

	return code, nil
}

// extractSessionData extracts session data from the authorization redirect
func extractSessionData(location string) (string, string, error) {
	redirectURL, err := url.Parse(location)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse redirect URL: %w", err)
	}

	sessionDataKey := redirectURL.Query().Get("sessionDataKey")
	if sessionDataKey == "" {
		return "", "", fmt.Errorf("sessionDataKey not found in redirect")
	}

	applicationId := redirectURL.Query().Get("applicationId")
	if applicationId == "" {
		return "", "", fmt.Errorf("applicationId not found in redirect")
	}

	return sessionDataKey, applicationId, nil
}

// validateOAuth2ErrorRedirect validates OAuth2 error redirect responses
func validateOAuth2ErrorRedirect(location string, expectedError string, expectedErrorDescription string) error {
	parsedURL, err := url.Parse(location)
	if err != nil {
		return fmt.Errorf("failed to parse redirect URL: %w", err)
	}

	queryParams := parsedURL.Query()

	// First check for OAuth2 error parameters (error, error_description)
	actualError := queryParams.Get("error")
	if actualError != "" {
		if actualError != expectedError {
			return fmt.Errorf("expected OAuth2 error '%s', got '%s'", expectedError, actualError)
		}

		if expectedErrorDescription != "" {
			actualErrorDescription := queryParams.Get("error_description")
			if actualErrorDescription != expectedErrorDescription {
				return fmt.Errorf("expected error_description '%s', got '%s'", expectedErrorDescription, actualErrorDescription)
			}
		}

		return nil
	}

	// Check for Thunder error page parameters (errorCode, errorMessage)
	actualErrorCode := queryParams.Get("errorCode")
	if actualErrorCode != "" {
		if actualErrorCode != expectedError {
			return fmt.Errorf("expected error code '%s', got '%s'", expectedError, actualErrorCode)
		}

		if expectedErrorDescription != "" {
			actualErrorMessage := queryParams.Get("errorMessage")
			if actualErrorMessage != expectedErrorDescription {
				return fmt.Errorf("expected error message '%s', got '%s'", expectedErrorDescription, actualErrorMessage)
			}
		}

		return nil
	}

	return fmt.Errorf("no error parameters found in redirect URL (neither 'error'/'error_description' nor 'errorCode'/'errorMessage')")
}
