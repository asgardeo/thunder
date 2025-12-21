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
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// InitiateAuthorizationFlow starts the OAuth2 authorization flow
func InitiateAuthorizationFlow(clientID, redirectURI, responseType, scope, state string) (*http.Response, error) {
	return initiateAuthorizationFlow(clientID, redirectURI, responseType, scope, state, "", "", "")
}

// InitiateAuthorizationFlowWithResource starts the OAuth2 authorization flow with resource parameter
func InitiateAuthorizationFlowWithResource(clientID, redirectURI, responseType, scope, state,
	resource string) (*http.Response, error) {
	return initiateAuthorizationFlow(clientID, redirectURI, responseType, scope, state, resource, "", "")
}

// InitiateAuthorizationFlowWithPKCE starts the OAuth2 authorization flow with PKCE parameters
func InitiateAuthorizationFlowWithPKCE(clientID, redirectURI, responseType, scope, state, resource,
	codeChallenge, codeChallengeMethod string) (*http.Response, error) {
	return initiateAuthorizationFlow(clientID, redirectURI, responseType, scope, state, resource,
		codeChallenge, codeChallengeMethod)
}

// initiateAuthorizationFlow starts the OAuth2 authorization flow
// clientID, redirectURI, responseType, scope, and state are required parameters.
// resource, codeChallenge, and codeChallengeMethod are optional parameters.
func initiateAuthorizationFlow(clientID, redirectURI, responseType, scope, state, resource,
	codeChallenge, codeChallengeMethod string) (*http.Response, error) {
	authURL := TestServerURL + "/oauth2/authorize"
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", responseType)
	params.Set("scope", scope)
	params.Set("state", state)
	if resource != "" {
		params.Set("resource", resource)
	}
	if codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", codeChallengeMethod)
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
func ExecuteAuthenticationFlow(flowID string, inputs map[string]string, action string) (*FlowStep, error) {
	flowData := map[string]interface{}{
		"flowId": flowID,
	}

	if len(inputs) > 0 {
		flowData["inputs"] = inputs
	}
	if action != "" {
		flowData["action"] = action
	}

	flowJSON, err := json.Marshal(flowData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow data: %w", err)
	}

	req, err := http.NewRequest("POST", TestServerURL+"/flow/execute", bytes.NewBuffer(flowJSON))
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

	bodyBytes, _ := io.ReadAll(resp.Body)
	var flowStep FlowStep
	err = json.Unmarshal(bodyBytes, &flowStep)
	if err != nil {
		return nil, fmt.Errorf("failed to decode flow response: %w", err)
	}

	return &flowStep, nil
}

// CompleteAuthorization completes the authorization using the assertion
func CompleteAuthorization(authID, assertion string) (*AuthorizationResponse, error) {
	authzData := map[string]interface{}{
		"authId":    authID,
		"assertion": assertion,
	}

	authzJSON, err := json.Marshal(authzData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authorization data: %w", err)
	}

	req, err := http.NewRequest("POST", TestServerURL+"/oauth2/authorize", bytes.NewBuffer(authzJSON))
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
		return nil, fmt.Errorf("authorization completion failed with status %d: %s",
			resp.StatusCode, string(bodyBytes))
	}

	var authzResponse AuthorizationResponse
	err = json.NewDecoder(resp.Body).Decode(&authzResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode authorization response: %w", err)
	}

	return &authzResponse, nil
}

// RequestToken performs a token request and returns raw HTTP result for both success and failure scenarios.
// grantType, code, and redirectURI are sent in the form body, while client credentials are sent via HTTP
// Basic Auth header.
func RequestToken(clientID, clientSecret, code, redirectURI, grantType string) (*TokenHTTPResult, error) {
	return requestToken(clientID, clientSecret, code, redirectURI, grantType, false, "")
}

// RequestTokenWithPKCE performs a token request with PKCE and returns raw HTTP result for both success and
// failure scenarios.
// grantType, code, redirectURI, and codeVerifier are sent in the form body, while client credentials are
// sent via HTTP Basic Auth header.
func RequestTokenWithPKCE(clientID, clientSecret, code, redirectURI, grantType, codeVerifier string) (
	*TokenHTTPResult, error) {
	return requestToken(clientID, clientSecret, code, redirectURI, grantType, true, codeVerifier)
}

// requestToken performs a token request and returns raw HTTP result for both success and failure scenarios.
// grantType, code, and redirectURI are required parameters.
// If tokenAuthInBody is true, client credentials are sent in the request body; otherwise, HTTP Basic Auth
// is used. codeVerifier is required for PKCE token requests.
func requestToken(clientID, clientSecret, code, redirectURI, grantType string, tokenAuthInBody bool,
	codeVerifier string) (*TokenHTTPResult, error) {
	tokenURL := TestServerURL + "/oauth2/token"
	tokenData := url.Values{}

	tokenData.Set("grant_type", grantType)
	tokenData.Set("code", code)
	tokenData.Set("redirect_uri", redirectURI)
	if codeVerifier != "" {
		tokenData.Set("code_verifier", codeVerifier)
	}
	if tokenAuthInBody {
		tokenData.Set("client_id", clientID)
		if clientSecret != "" {
			tokenData.Set("client_secret", clientSecret)
		}
	}

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(tokenData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if !tokenAuthInBody {
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

// ExtractAuthorizationCode extracts the authorization code from the redirect URI
func ExtractAuthorizationCode(redirectURI string) (string, error) {
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

// ExtractAuthData extracts auth ID and flow ID from the authorization redirect
func ExtractAuthData(location string) (string, string, error) {
	redirectURL, err := url.Parse(location)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse redirect URL: %w", err)
	}

	authID := redirectURL.Query().Get("authId")
	if authID == "" {
		return "", "", fmt.Errorf("authId not found in redirect")
	}

	flowId := redirectURL.Query().Get("flowId")
	if flowId == "" {
		return "", "", fmt.Errorf("flowId not found in redirect")
	}

	return authID, flowId, nil
}

// ValidateOAuth2ErrorRedirect validates OAuth2 error redirect responses
func ValidateOAuth2ErrorRedirect(location string, expectedError string,
	expectedErrorDescription string) error {
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

	return fmt.Errorf(
		"no error parameters found in redirect URL (neither 'error'/'error_description' nor " +
			"'errorCode'/'errorMessage')")
}

// ObtainAccessTokenWithPassword performs the complete OAuth authorization code flow with password
// authentication and returns a TokenResponse with the access token and expiry information.
func ObtainAccessTokenWithPassword(clientID, redirectURI, scope, username, password string,
	usePKCE bool) (*TokenResponse, error) {
	var codeVerifier string
	var codeChallenge string

	// Generate PKCE parameters if enabled
	if usePKCE {
		var err error
		codeVerifier, err = generateCodeVerifier()
		if err != nil {
			return nil, fmt.Errorf("failed to generate code verifier: %w", err)
		}
		codeChallenge = generateCodeChallenge(codeVerifier)
		log.Printf("Generated PKCE - Verifier length: %d, Challenge: %s", len(codeVerifier), codeChallenge)
	}

	// Step 1: Initiate authorization flow with PKCE
	resp, err := InitiateAuthorizationFlowWithPKCE(clientID, redirectURI, "code", scope, "test-state", "",
		codeChallenge, "S256")
	if err != nil {
		return nil, fmt.Errorf("failed to initiate authorization: %w", err)
	}
	defer resp.Body.Close()

	// Check for redirect
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther &&
		resp.StatusCode != http.StatusTemporaryRedirect {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("expected redirect response, got status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("no Location header in authorization response")
	}

	log.Printf("Authorization redirect location: %s", location)
	// Step 2: Extract auth ID and flow ID
	authID, flowID, err := ExtractAuthData(location)
	if err != nil {
		return nil, fmt.Errorf("failed to extract auth ID: %w", err)
	}

	// Step 3: Execute initial authentication flow step (to get to the login prompt)
	_, err = ExecuteAuthenticationFlow(flowID, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to execute initial authentication flow: %w", err)
	}

	// Step 4: Execute authentication flow with credentials
	flowStep, err := ExecuteAuthenticationFlow(flowID, map[string]string{
		"username": username,
		"password": password,
	}, "action_001")
	if err != nil {
		return nil, fmt.Errorf("failed to execute authentication flow: %w", err)
	}

	if flowStep.FlowStatus != "COMPLETE" {
		stepJSON, _ := json.Marshal(flowStep)
		return nil, fmt.Errorf("authentication flow not complete: status=%s, failureReason=%s, step=%s",
			flowStep.FlowStatus, flowStep.FailureReason, string(stepJSON))
	}

	if flowStep.Assertion == "" {
		return nil, fmt.Errorf("no assertion returned from authentication flow")
	}

	// Step 5: Complete authorization with assertion
	authzResp, err := CompleteAuthorization(authID, flowStep.Assertion)
	if err != nil {
		return nil, fmt.Errorf("failed to complete authorization: %w", err)
	}

	// Step 6: Extract authorization code
	code, err := ExtractAuthorizationCode(authzResp.RedirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to extract authorization code: %w", err)
	}

	// Step 7: Exchange code for token with PKCE verifier
	tokenResult, err := RequestTokenWithPKCE(clientID, "", code, redirectURI, "authorization_code",
		codeVerifier)
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %w", err)
	}

	if tokenResult.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", tokenResult.StatusCode,
			string(tokenResult.Body))
	}

	if tokenResult.Token == nil {
		return nil, fmt.Errorf("no token in response")
	}

	// Create and return token state
	return tokenResult.Token, nil
}

// RefreshAccessToken uses the refresh token to obtain a new access token
// client credentials are sent via HTTP Basic Auth header.
func RefreshAccessToken(clientID, clientSecret, refreshToken string) (*TokenResponse, error) {
	return refreshAccessToken(clientID, clientSecret, refreshToken, false)
}

// RefreshAccessTokenWithClientCredentialsInBody uses the refresh token to obtain a new access token where
// client credentials are sent in the request body.
func RefreshAccessTokenWithClientCredentialsInBody(clientID, clientSecret, refreshToken string) (
	*TokenResponse, error) {
	return refreshAccessToken(clientID, clientSecret, refreshToken, true)
}

// refreshAccessToken uses the refresh token to obtain a new access token
func refreshAccessToken(clientID, clientSecret, refreshToken string, tokenAuthInBody bool) (
	*TokenResponse, error) {
	tokenURL := TestServerURL + "/oauth2/token"
	tokenData := url.Values{}

	tokenData.Set("grant_type", "refresh_token")
	tokenData.Set("refresh_token", refreshToken)

	if tokenAuthInBody {
		tokenData.Set("client_id", clientID)
		if clientSecret != "" {
			tokenData.Set("client_secret", clientSecret)
		}
	}

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(tokenData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if !tokenAuthInBody {
		if clientID != "" {
			req.SetBasicAuth(clientID, clientSecret)
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send refresh token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	// Preserve the refresh token if not returned in response (common for some OAuth servers)
	if tokenResponse.RefreshToken == "" {
		tokenResponse.RefreshToken = refreshToken
		log.Println("Refresh token not returned in response, preserving existing token")
	}

	return &tokenResponse, nil
}

// generateCodeVerifier generates a cryptographically secure random code verifier for PKCE (RFC 7636).
func generateCodeVerifier() (string, error) {
	// Generate 32 random bytes (will result in 43 characters when base64url encoded)
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Base64 URL encode without padding
	verifier := base64.RawURLEncoding.EncodeToString(bytes)
	return verifier, nil
}

// generateCodeChallenge generates a code challenge from a code verifier using SHA-256
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return challenge
}
