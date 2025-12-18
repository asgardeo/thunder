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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	scopeTestClientID     = "scope_authz_test_client_456"
	scopeTestClientSecret = "scope_authz_test_secret_456"
	scopeTestAppName      = "ScopeAuthzTestApp"
	scopeTestRedirectURI  = "https://localhost:3000/callback"
)

var (
	scopeTestOUID           string
	scopeTestRoleID         string
	scopeUserWithRole       string
	scopeUserNoRole         string
	scopeUserSchemaID       string
	scopeTestResourceServer string
	scopeTestUserSchema     = testutils.UserSchema{
		Name: "authz-test-person",
		Schema: map[string]interface{}{
			"username": map[string]interface{}{
				"type": "string",
			},
			"password": map[string]interface{}{
				"type": "string",
			},
			"email": map[string]interface{}{
				"type": "string",
			},
			"firstName": map[string]interface{}{
				"type": "string",
			},
			"lastName": map[string]interface{}{
				"type": "string",
			},
		},
	}
)

type OAuthAuthzScopeTestSuite struct {
	suite.Suite
	client        *http.Client
	applicationID string
}

func TestOAuthAuthzScopeTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthAuthzScopeTestSuite))
}

func (ts *OAuthAuthzScopeTestSuite) SetupSuite() {
	var err error
	// Setup HTTP client
	ts.client = testutils.GetHTTPClient()
	// Create test organization unit
	ou := testutils.OrganizationUnit{
		Handle:      "oauth-scope-authz-test-ou",
		Name:        "OAuth Scope Authorization Test OU",
		Description: "Organization unit for OAuth scope authorization testing",
		Parent:      nil,
	}
	scopeTestOUID, err = testutils.CreateOrganizationUnit(ou)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit: %v", err)
	}
	scopeTestUserSchema.OrganizationUnitId = scopeTestOUID

	// Create user schema
	scopeUserSchemaID, err = testutils.CreateUserType(scopeTestUserSchema)
	if err != nil {
		ts.T().Fatalf("Failed to create user schema: %v", err)
	}

	// We need to use the inbound_auth_config approach for OAuth apps
	appID, err := ts.createOAuthApplication()
	if err != nil {
		ts.T().Fatalf("Failed to create OAuth application: %v", err)
	}
	ts.applicationID = appID

	// Create user with role
	userWithRole := testutils.User{
		OrganizationUnit: scopeTestOUID,
		Type:             "authz-test-person",
		Attributes: json.RawMessage(`{
			"username": "oauth_authorized_user",
			"password": "SecurePass123!",
			"email": "oauth_authorized@test.com",
			"firstName": "OAuth",
			"lastName": "Authorized"
		}`),
	}
	scopeUserWithRole, err = testutils.CreateUser(userWithRole)
	if err != nil {
		ts.T().Fatalf("Failed to create user with role: %v", err)
	}

	// Create user without role
	userNoRole := testutils.User{
		OrganizationUnit: scopeTestOUID,
		Type:             "authz-test-person",
		Attributes: json.RawMessage(`{
			"username": "oauth_unauthorized_user",
			"password": "SecurePass123!",
			"email": "oauth_unauthorized@test.com",
			"firstName": "OAuth",
			"lastName": "Unauthorized"
		}`),
	}
	scopeUserNoRole, err = testutils.CreateUser(userNoRole)
	if err != nil {
		ts.T().Fatalf("Failed to create user without role: %v", err)
	}

	// Create resource server with actions for permissions
	resourceServer := testutils.ResourceServer{
		Name:               "OAuth Document Management System",
		Description:        "System for managing documents via OAuth",
		Identifier:         "oauth-document-mgmt",
		OrganizationUnitID: scopeTestOUID,
	}
	actions := []testutils.Action{
		{
			Name:        "Read Documents",
			Handle:      "read",
			Description: "Permission to read documents",
		},
		{
			Name:        "Write Documents",
			Handle:      "write",
			Description: "Permission to write documents",
		},
	}
	scopeTestResourceServer, err := testutils.CreateResourceServerWithActions(resourceServer, actions)
	if err != nil {
		ts.T().Fatalf("Failed to create resource server with actions: %v", err)
	}

	// Create role with permissions and assign to first user
	role := testutils.Role{
		Name:               "OAuth_DocumentEditor",
		Description:        "Can read and write documents (OAuth test)",
		OrganizationUnitID: scopeTestOUID,
		Permissions: []testutils.ResourcePermissions{
			{
				ResourceServerID: scopeTestResourceServer,
				Permissions:      []string{"read", "write"},
			},
		},
		Assignments: []testutils.Assignment{
			{ID: scopeUserWithRole, Type: "user"},
		},
	}
	scopeTestRoleID, err = testutils.CreateRole(role)
	if err != nil {
		ts.T().Fatalf("Failed to create test role: %v", err)
	}
}

func (ts *OAuthAuthzScopeTestSuite) TearDownSuite() {
	// Cleanup in reverse order
	if scopeTestRoleID != "" {
		if err := testutils.DeleteRole(scopeTestRoleID); err != nil {
			ts.T().Logf("Failed to delete test role: %v", err)
		}
	}

	if scopeTestResourceServer != "" {
		if err := testutils.DeleteResourceServer(scopeTestResourceServer); err != nil {
			ts.T().Logf("Failed to delete test resource server: %v", err)
		}
	}

	if scopeUserNoRole != "" {
		if err := testutils.DeleteUser(scopeUserNoRole); err != nil {
			ts.T().Logf("Failed to delete user without role: %v", err)
		}
	}

	if scopeUserWithRole != "" {
		if err := testutils.DeleteUser(scopeUserWithRole); err != nil {
			ts.T().Logf("Failed to delete user with role: %v", err)
		}
	}

	if ts.applicationID != "" {
		if err := testutils.DeleteApplication(ts.applicationID); err != nil {
			ts.T().Logf("Failed to delete application: %v", err)
		}
	}

	if scopeTestOUID != "" {
		if err := testutils.DeleteOrganizationUnit(scopeTestOUID); err != nil {
			ts.T().Logf("Failed to delete organization unit: %v", err)
		}
	}

	if scopeUserSchemaID != "" {
		if err := testutils.DeleteUserType(scopeUserSchemaID); err != nil {
			ts.T().Logf("Failed to delete user schema: %v", err)
		}
	}
}

// TestOAuthAuthzFlow_WithAuthorizedScopes tests complete OAuth flow with authorized scopes
func (ts *OAuthAuthzScopeTestSuite) TestOAuthAuthzFlow_WithAuthorizedScopes() {
	// Step 1: Initiate OAuth authorization request
	scope := "openid read write"
	state := "test_state_with_scopes"

	authResp, err := testutils.InitiateAuthorizationFlow(scopeTestClientID, scopeTestRedirectURI, "code", scope, state)
	ts.Require().NoError(err, "Failed to initiate authorization")
	ts.Require().NotNil(authResp, "Authorization response should not be nil")

	// Step 2: Extract auth ID and flow ID from redirect
	location := authResp.Header.Get("Location")
	ts.Require().NotEmpty(location, "Location header should not be empty")

	authId, flowID, err := testutils.ExtractAuthData(location)
	ts.Require().NoError(err, "Failed to extract auth ID")
	ts.Require().NotEmpty(authId, "Auth ID should not be empty")
	ts.Require().NotEmpty(flowID, "Flow ID should not be empty")

	// Step 3: Execute authentication flow with authorized user
	flowInputs := map[string]string{
		"username": "oauth_authorized_user",
		"password": "SecurePass123!",
	}

	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, flowInputs)
	ts.Require().NoError(err, "Failed to execute authentication flow")
	ts.Require().NotNil(flowStep, "Flow step should not be nil")
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Flow should be complete")
	ts.Require().NotEmpty(flowStep.Assertion, "Assertion should not be empty")

	// Step 4: Complete authorization with assertion
	authzResponse, err := testutils.CompleteAuthorization(authId, flowStep.Assertion)
	ts.Require().NoError(err, "Failed to complete authorization")
	ts.Require().NotNil(authzResponse, "Authorization response should not be nil")
	ts.Require().NotEmpty(authzResponse.RedirectURI, "Redirect URI should not be empty")

	// Step 5: Extract authorization code
	code, err := testutils.ExtractAuthorizationCode(authzResponse.RedirectURI)
	ts.Require().NoError(err, "Failed to extract authorization code")
	ts.Require().NotEmpty(code, "Authorization code should not be empty")

	// Step 6: Exchange code for token
	tokenResult, err := testutils.RequestToken(scopeTestClientID, scopeTestClientSecret, code, scopeTestRedirectURI, "authorization_code")
	ts.Require().NoError(err, "Failed to exchange code for token")
	ts.Require().NotNil(tokenResult, "Token result should not be nil")
	ts.Require().Equal(http.StatusOK, tokenResult.StatusCode, "Token endpoint should return 200")
	ts.Require().NotNil(tokenResult.Token, "Token response should not be nil")
	ts.Require().NotEmpty(tokenResult.Token.AccessToken, "Access token should not be empty")

	// Step 7: Decode access token and verify scopes
	claims, err := testutils.DecodeJWT(tokenResult.Token.AccessToken)
	ts.Require().NoError(err, "Failed to decode access token")
	ts.Require().NotNil(claims, "Claims should not be nil")

	// Verify scope claim contains authorized scopes
	scopeRaw, ok := claims.Additional["scope"]
	ts.Require().True(ok, "scope claim should be present in access token")

	scopeStr, ok := scopeRaw.(string)
	ts.Require().True(ok, "scope claim should be a string")

	// Parse scopes (space-separated)
	scopes := strings.Split(scopeStr, " ")
	ts.Require().Contains(scopes, "openid", "Token should contain openid scope")
	ts.Require().Contains(scopes, "read", "Token should contain read scope")
	ts.Require().Contains(scopes, "write", "Token should contain write scope")
}

// TestOAuthAuthzFlow_WithNoAuthorizedScopes tests OAuth flow when user has no custom scopes
func (ts *OAuthAuthzScopeTestSuite) TestOAuthAuthzFlow_WithNoAuthorizedScopes() {
	// Step 1: Initiate OAuth authorization request
	scope := "openid read write"
	state := "test_state_no_scopes"

	authResp, err := testutils.InitiateAuthorizationFlow(scopeTestClientID, scopeTestRedirectURI, "code", scope, state)
	ts.Require().NoError(err, "Failed to initiate authorization")
	ts.Require().NotNil(authResp, "Authorization response should not be nil")

	// Step 2: Extract auth ID and flow ID from redirect
	location := authResp.Header.Get("Location")
	ts.Require().NotEmpty(location, "Location header should not be empty")

	authId, flowID, err := testutils.ExtractAuthData(location)
	ts.Require().NoError(err, "Failed to extract auth ID")
	ts.Require().NotEmpty(authId, "Auth ID should not be empty")
	ts.Require().NotEmpty(flowID, "Flow ID should not be empty")

	// Step 3: Execute authentication flow with unauthorized user (no role)
	flowInputs := map[string]string{
		"username": "oauth_unauthorized_user",
		"password": "SecurePass123!",
	}

	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, flowInputs)
	ts.Require().NoError(err, "Failed to execute authentication flow")
	ts.Require().NotNil(flowStep, "Flow step should not be nil")
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Flow should be complete")
	ts.Require().NotEmpty(flowStep.Assertion, "Assertion should not be empty")

	// Step 4: Complete authorization with assertion
	authzResponse, err := testutils.CompleteAuthorization(authId, flowStep.Assertion)
	ts.Require().NoError(err, "Failed to complete authorization")
	ts.Require().NotNil(authzResponse, "Authorization response should not be nil")
	ts.Require().NotEmpty(authzResponse.RedirectURI, "Redirect URI should not be empty")

	// Step 5: Extract authorization code
	code, err := testutils.ExtractAuthorizationCode(authzResponse.RedirectURI)
	ts.Require().NoError(err, "Failed to extract authorization code")
	ts.Require().NotEmpty(code, "Authorization code should not be empty")

	// Step 6: Exchange code for token
	tokenResult, err := testutils.RequestToken(scopeTestClientID, scopeTestClientSecret, code, scopeTestRedirectURI, "authorization_code")
	ts.Require().NoError(err, "Failed to exchange code for token")
	ts.Require().NotNil(tokenResult, "Token result should not be nil")
	ts.Require().Equal(http.StatusOK, tokenResult.StatusCode, "Token endpoint should return 200")
	ts.Require().NotNil(tokenResult.Token, "Token response should not be nil")
	ts.Require().NotEmpty(tokenResult.Token.AccessToken, "Access token should not be empty")

	// Step 7: Decode access token and verify scopes
	claims, err := testutils.DecodeJWT(tokenResult.Token.AccessToken)
	ts.Require().NoError(err, "Failed to decode access token")
	ts.Require().NotNil(claims, "Claims should not be nil")

	// Verify scope claim contains ONLY OIDC scopes (no custom scopes)
	scopeRaw, ok := claims.Additional["scope"]
	ts.Require().True(ok, "scope claim should be present in access token")

	scopeStr, ok := scopeRaw.(string)
	ts.Require().True(ok, "scope claim should be a string")

	// Parse scopes (space-separated)
	scopes := strings.Split(scopeStr, " ")
	ts.Require().Contains(scopes, "openid", "Token should contain openid scope")
	ts.Require().NotContains(scopes, "read", "Token should NOT contain read scope")
	ts.Require().NotContains(scopes, "write", "Token should NOT contain write scope")

	// Verify only OIDC scopes are present
	for _, scope := range scopes {
		// OIDC scopes: openid, profile, email, address, phone, offline_access
		isOIDCScope := scope == "openid" || scope == "profile" || scope == "email" ||
			scope == "address" || scope == "phone" || scope == "offline_access"
		ts.Require().True(isOIDCScope, "Scope '%s' should be an OIDC scope", scope)
	}
}

// createOAuthApplication creates an OAuth application using the low-level API
func (ts *OAuthAuthzScopeTestSuite) createOAuthApplication() (string, error) {
	app := map[string]interface{}{
		"name":                         scopeTestAppName,
		"description":                  "OAuth application for scope authorization testing",
		"auth_flow_graph_id":           "auth_flow_config_basic",
		"registration_flow_graph_id":   "registration_flow_config_basic",
		"is_registration_flow_enabled": false,
		"inbound_auth_config": []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"client_id":                  scopeTestClientID,
					"client_secret":              scopeTestClientSecret,
					"redirect_uris":              []string{scopeTestRedirectURI},
					"grant_types":                []string{"authorization_code", "refresh_token"},
					"response_types":             []string{"code"},
					"token_endpoint_auth_method": "client_secret_basic",
				},
			},
		},
	}

	return ts.createApplicationRaw(app)
}

// createApplicationRaw creates an application using raw HTTP request
func (ts *OAuthAuthzScopeTestSuite) createApplicationRaw(app map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(app)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", testutils.TestServerURL+"/applications", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create application, status: %d", resp.StatusCode)
	}

	var respData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", err
	}

	return respData["id"].(string), nil
}
