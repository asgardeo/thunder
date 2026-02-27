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
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	adminTokenState *TokenResponse
	tokenInitOnce   sync.Once
)

// authTransport wraps http.RoundTripper to inject authorization headers
type authTransport struct {
	base     http.RoundTripper
	getToken func() (string, error)
}

// RoundTrip implements http.RoundTripper interface
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Skip auth for public endpoints
	if isPublicEndpoint(req.URL.Path) {
		return t.base.RoundTrip(req)
	}

	// Get token (auto-refreshes if needed)
	token, err := t.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Clone request and add auth header
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+token)

	return t.base.RoundTrip(reqClone)
}

// isPublicEndpoint determines if an endpoint requires authentication
func isPublicEndpoint(path string) bool {
	publicPrefixes := []string{
		"/health/",
		"/auth/",
		"/flow/execute",
		"/flow/meta",
		"/oauth2/",
		"/.well-known/openid-configuration",
		"/.well-known/oauth-authorization-server",
		"/gate/",    // Gate application (login UI)
		"/develop/", // Develop application
		"/error",
	}

	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// NewHTTPClientWithTokenProvider builds an HTTP client that injects Authorization headers using the provided token
// provider and skips TLS verification to work with local test servers.
func NewHTTPClientWithTokenProvider(getToken func() (string, error)) *http.Client {
	return &http.Client{
		Transport: &authTransport{
			base: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			getToken: getToken,
		},
	}
}

// GetHTTPClientWithToken returns an HTTP client that always uses the provided bearer token.
func GetHTTPClientWithToken(token string) *http.Client {
	return NewHTTPClientWithTokenProvider(func() (string, error) {
		if token == "" {
			return "", fmt.Errorf("token is empty")
		}
		return token, nil
	})
}

// GetHTTPClientForUser obtains a token using password grant (via DEVELOP app) and returns an HTTP client that
// injects that token. This keeps token generation out of individual tests.
func GetHTTPClientForUser(username, password string) (*http.Client, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	tokenResp, err := ObtainAccessTokenWithPassword(
		"DEVELOP",
		"https://localhost:8095/develop",
		"openid",
		username,
		password,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain token for user %s: %w", username, err)
	}

	if tokenResp == nil || tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token returned for user %s", username)
	}

	return GetHTTPClientWithToken(tokenResp.AccessToken), nil
}

// ObtainAdminAccessToken obtains an admin access token using the DEVELOP app and stores it globally
func ObtainAdminAccessToken() error {
	log.Println("Obtaining admin access token...")
	var err error
	adminTokenState, err = ObtainAccessTokenWithPassword(
		"DEVELOP",
		"https://localhost:8095/develop",
		"system",
		"admin",
		"admin",
		true,
	)
	if err != nil {
		return fmt.Errorf("failed to obtain access token: %w", err)
	}
	now := time.Now()
	adminTokenState.ExpiresAt = now.Add(time.Duration(adminTokenState.ExpiresIn) * time.Second)

	// Export complete token state to environment variable for other test packages
	if err := exportTokenStateToEnv(); err != nil {
		return fmt.Errorf("failed to export token state to environment: %w", err)
	}

	log.Printf("Access token obtained successfully")
	return nil
}

// GetAccessToken returns the current access token, refreshing it if necessary
func GetAccessToken() (string, error) {
	// First try to load complete token from environment (set by main runner)
	// This allows token refresh to work across test packages
	if adminTokenState == nil {
		tokenState, err := loadTokenStateFromEnv()
		if err != nil {
			return "", fmt.Errorf("failed to load token state from environment: %w", err)
		}
		if tokenState != nil {
			adminTokenState = tokenState
		}
	}

	// Fallback: Initialize token if not available (for running individual test packages)
	if adminTokenState == nil {
		// Use sync.Once to ensure token is obtained only once even with concurrent calls
		var initErr error
		tokenInitOnce.Do(func() {
			log.Println("No token available, obtaining access token automatically...")
			initErr = ObtainAdminAccessToken()
		})
		if initErr != nil {
			return "", fmt.Errorf("failed to obtain access token: %w", initErr)
		}
	}

	// Check if token needs refresh
	if err := RefreshTokenIfNeeded(); err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	return adminTokenState.AccessToken, nil
}

// RefreshTokenIfNeeded checks if the token is expired or expiring soon and refreshes it
func RefreshTokenIfNeeded() error {
	if adminTokenState == nil {
		return nil // No token to refresh
	}

	// Check if refresh is needed (expired or within buffer time)
	if !shouldRefresh(adminTokenState) {
		return nil // Token is still valid
	}

	refreshToken := adminTokenState.RefreshToken

	log.Println("Token expired or expiring soon, refreshing...")

	// Refresh the token
	var err error
	adminTokenState, err = RefreshAccessTokenWithClientCredentialsInBody("DEVELOP", "", refreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh access token: %w", err)
	}

	now := time.Now()
	adminTokenState.ExpiresAt = now.Add(time.Duration(adminTokenState.ExpiresIn) * time.Second)

	// Update environment variable so other test packages can use refreshed token
	if err := exportTokenStateToEnv(); err != nil {
		log.Printf("Warning: Failed to update token state in environment: %v\n", err)
		// Don't fail - the refresh was successful, just the env update failed
	}

	log.Printf("Access token refreshed successfully")
	return nil
}

func shouldRefresh(tokenState *TokenResponse) bool {
	if tokenState == nil {
		return false
	}
	now := time.Now()
	// Refresh if within 5 minutes of expiry
	return now.After(tokenState.ExpiresAt.Add(-5 * time.Minute))
}

// exportTokenStateToEnv serializes the current global token state and exports it to environment
func exportTokenStateToEnv() error {
	if adminTokenState == nil {
		return fmt.Errorf("no token state available to export")
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(adminTokenState)
	if err != nil {
		return fmt.Errorf("failed to serialize token state: %w", err)
	}

	// Encode as base64 for safe environment variable storage
	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	os.Setenv("THUNDER_TEST_ADMIN_TOKEN", encoded)

	log.Printf("Token state exported to environment")
	return nil
}

// loadTokenStateFromEnv deserializes token state from environment variable
func loadTokenStateFromEnv() (*TokenResponse, error) {
	encoded := os.Getenv("THUNDER_TEST_ADMIN_TOKEN")
	if encoded == "" {
		return nil, nil // No token state in environment
	}

	jsonBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token state: %w", err)
	}

	// Deserialize from JSON
	var tokenState TokenResponse
	if err := json.Unmarshal(jsonBytes, &tokenState); err != nil {
		return nil, fmt.Errorf("failed to deserialize token state: %w", err)
	}

	log.Printf("Token state loaded from environment")
	return &tokenState, nil
}
