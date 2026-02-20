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
	flowID        string
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

	// Create authentication flow
	authFlowID := ts.createTestAuthenticationFlow()
	ts.NotEmpty(authFlowID, "Authentication flow ID should not be empty")
	ts.flowID = authFlowID

	// We need to use the inbound_auth_config approach for OAuth apps
	appID, err := ts.createOAuthApplication(authFlowID)
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

	if ts.flowID != "" {
		if err := testutils.DeleteFlow(ts.flowID); err != nil {
			ts.T().Logf("Failed to delete authentication flow: %v", err)
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

	authResp, err := testutils.InitiateAuthorizationFlow(scopeTestClientID, scopeTestRedirectURI,
		"code", scope, state)
	ts.Require().NoError(err, "Failed to initiate authorization")
	ts.Require().NotNil(authResp, "Authorization response should not be nil")

	// Step 2: Extract auth ID and flow ID from redirect
	location := authResp.Header.Get("Location")
	ts.Require().NotEmpty(location, "Location header should not be empty")

	authId, flowID, err := testutils.ExtractAuthData(location)
	ts.Require().NoError(err, "Failed to extract auth ID")
	ts.Require().NotEmpty(authId, "Auth ID should not be empty")
	ts.Require().NotEmpty(flowID, "Flow ID should not be empty")

	// Step 3: Initiate authentication flow
	_, err = testutils.ExecuteAuthenticationFlow(flowID, nil, "")
	ts.Require().NoError(err, "Failed to initiate authentication flow")

	// Step 4: Execute authentication flow with authorized user
	flowInputs := map[string]string{
		"username": "oauth_authorized_user",
		"password": "SecurePass123!",
	}

	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, flowInputs, "action_001")
	ts.Require().NoError(err, "Failed to execute authentication flow")
	ts.Require().NotNil(flowStep, "Flow step should not be nil")
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Flow should be complete")
	ts.Require().NotEmpty(flowStep.Assertion, "Assertion should not be empty")

	// Step 5: Complete authorization with assertion
	authzResponse, err := testutils.CompleteAuthorization(authId, flowStep.Assertion)
	ts.Require().NoError(err, "Failed to complete authorization")
	ts.Require().NotNil(authzResponse, "Authorization response should not be nil")
	ts.Require().NotEmpty(authzResponse.RedirectURI, "Redirect URI should not be empty")

	// Step 6: Extract authorization code
	code, err := testutils.ExtractAuthorizationCode(authzResponse.RedirectURI)
	ts.Require().NoError(err, "Failed to extract authorization code")
	ts.Require().NotEmpty(code, "Authorization code should not be empty")

	// Step 7: Exchange code for token
	tokenResult, err := testutils.RequestToken(scopeTestClientID, scopeTestClientSecret, code,
		scopeTestRedirectURI, "authorization_code")
	ts.Require().NoError(err, "Failed to exchange code for token")
	ts.Require().NotNil(tokenResult, "Token result should not be nil")
	ts.Require().Equal(http.StatusOK, tokenResult.StatusCode, "Token endpoint should return 200")
	ts.Require().NotNil(tokenResult.Token, "Token response should not be nil")
	ts.Require().NotEmpty(tokenResult.Token.AccessToken, "Access token should not be empty")

	// Step 8: Decode access token and verify scopes
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

	authResp, err := testutils.InitiateAuthorizationFlow(scopeTestClientID, scopeTestRedirectURI,
		"code", scope, state)
	ts.Require().NoError(err, "Failed to initiate authorization")
	ts.Require().NotNil(authResp, "Authorization response should not be nil")

	// Step 2: Extract auth ID and flow ID from redirect
	location := authResp.Header.Get("Location")
	ts.Require().NotEmpty(location, "Location header should not be empty")

	authId, flowID, err := testutils.ExtractAuthData(location)
	ts.Require().NoError(err, "Failed to extract auth ID")
	ts.Require().NotEmpty(authId, "Auth ID should not be empty")
	ts.Require().NotEmpty(flowID, "Flow ID should not be empty")

	// Step 2: Initiate authentication flow
	_, err = testutils.ExecuteAuthenticationFlow(flowID, nil, "")
	ts.Require().NoError(err, "Failed to initiate authentication flow")

	// Step 4: Execute authentication flow with unauthorized user (no role)
	flowInputs := map[string]string{
		"username": "oauth_unauthorized_user",
		"password": "SecurePass123!",
	}

	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, flowInputs, "action_001")
	ts.Require().NoError(err, "Failed to execute authentication flow")
	ts.Require().NotNil(flowStep, "Flow step should not be nil")
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Flow should be complete")
	ts.Require().NotEmpty(flowStep.Assertion, "Assertion should not be empty")

	// Step 5: Complete authorization with assertion
	authzResponse, err := testutils.CompleteAuthorization(authId, flowStep.Assertion)
	ts.Require().NoError(err, "Failed to complete authorization")
	ts.Require().NotNil(authzResponse, "Authorization response should not be nil")
	ts.Require().NotEmpty(authzResponse.RedirectURI, "Redirect URI should not be empty")

	// Step 6: Extract authorization code
	code, err := testutils.ExtractAuthorizationCode(authzResponse.RedirectURI)
	ts.Require().NoError(err, "Failed to extract authorization code")
	ts.Require().NotEmpty(code, "Authorization code should not be empty")

	// Step 7: Exchange code for token
	tokenResult, err := testutils.RequestToken(scopeTestClientID, scopeTestClientSecret, code,
		scopeTestRedirectURI, "authorization_code")
	ts.Require().NoError(err, "Failed to exchange code for token")
	ts.Require().NotNil(tokenResult, "Token result should not be nil")
	ts.Require().Equal(http.StatusOK, tokenResult.StatusCode, "Token endpoint should return 200")
	ts.Require().NotNil(tokenResult.Token, "Token response should not be nil")
	ts.Require().NotEmpty(tokenResult.Token.AccessToken, "Access token should not be empty")

	// Step 8: Decode access token and verify scopes
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
func (ts *OAuthAuthzScopeTestSuite) createOAuthApplication(authFlowID string) (string, error) {
	app := map[string]interface{}{
		"name":                         scopeTestAppName,
		"description":                  "OAuth application for scope authorization testing",
		"auth_flow_id":                 authFlowID,
		"is_registration_flow_enabled": false,
		"allowed_user_types":           []string{"authz-test-person"},
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

func (ts *OAuthAuthzScopeTestSuite) createTestAuthenticationFlow() string {
	flow := testutils.Flow{
		Name:     "Authz Test Auth Flow",
		FlowType: "AUTHENTICATION",
		Handle:   "authz_test_auth_flow",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "prompt_credentials",
			},
			{
				"id":   "prompt_credentials",
				"type": "PROMPT",
				"prompts": []map[string]interface{}{
					{
						"inputs": []map[string]interface{}{
							{
								"ref":        "input_001",
								"identifier": "username",
								"type":       "TEXT_INPUT",
								"required":   true,
							},
							{
								"ref":        "input_002",
								"identifier": "password",
								"type":       "PASSWORD_INPUT",
								"required":   true,
							},
						},
						"action": map[string]interface{}{
							"ref":      "action_001",
							"nextNode": "basic_auth",
						},
					},
				},
			},
			{
				"id":   "basic_auth",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "BasicAuthExecutor",
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_001",
							"identifier": "username",
							"type":       "TEXT_INPUT",
							"required":   true,
						},
						{
							"ref":        "input_002",
							"identifier": "password",
							"type":       "PASSWORD_INPUT",
							"required":   true,
						},
					},
				},
				"onSuccess": "authorization_check",
			},
			{
				"id":   "authorization_check",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "AuthorizationExecutor",
				},
				"onSuccess": "auth_assert",
			},
			{
				"id":   "auth_assert",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "AuthAssertExecutor",
				},
				"onSuccess": "end",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	}

	flowID, err := testutils.CreateFlow(flow)
	ts.Require().NoError(err, "Failed to create test authentication flow")
	ts.T().Logf("Created test authentication flow with ID: %s", flowID)

	return flowID
}

// TestOAuthAuthzFlow_WithRequiredAttributes tests that required_attributes from scopes and access token config
// are correctly filtered in the auth assertion
func (ts *OAuthAuthzScopeTestSuite) TestOAuthAuthzFlow_WithRequiredAttributes() {
	// Create OAuth app with IDToken and AccessToken configs
	appConfig := map[string]interface{}{
		"name":                       "RequiredAttributesTestApp",
		"client_id":                  "required_attrs_test_client",
		"redirect_uris":              []string{scopeTestRedirectURI},
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "client_secret_basic",
		"auth_flow_id":               ts.flowID,
		"inbound_auth_config": []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"client_id":                  "required_attrs_test_client",
					"client_secret":              "required_attrs_test_secret",
					"redirect_uris":              []string{scopeTestRedirectURI},
					"grant_types":                []string{"authorization_code", "refresh_token"},
					"response_types":             []string{"code"},
					"token_endpoint_auth_method": "client_secret_basic",
					"token": map[string]interface{}{
						"id_token": map[string]interface{}{
							"user_attributes": []string{"sub", "name", "email"}, // Only allow these in ID token
						},
						"access_token": map[string]interface{}{
							"user_attributes": []string{"groups", "roles"}, // Access token attributes
						},
					},
					"scope_claims": map[string]interface{}{
						"profile": []string{"name"}, // Custom mapping
					},
				},
			},
		},
	}

	appID, err := ts.createApplicationRaw(appConfig)
	ts.Require().NoError(err, "Failed to create OAuth application")
	defer func() {
		_ = testutils.DeleteApplication(appID)
	}()

	// Create a user with attributes
	userAttributesJSON := `{
		"username": "requiredattrsuser",
		"password": "TestPassword123!",
		"email": "requiredattrs@test.com",
		"firstName": "Required",
		"lastName": "Attrs",
		"name": "Required Attrs",
		"groups": ["admin", "user"],
		"roles": ["developer"],
		"phone": "+1234567890",
		"customAttr": "customValue"
	}`

	user := testutils.User{
		OrganizationUnit: scopeTestOUID,
		Type:             "authz-test-person",
		Attributes:       json.RawMessage(userAttributesJSON),
	}

	userID, err := testutils.CreateUser(user)
	ts.Require().NoError(err, "Failed to create test user")
	defer func() {
		_ = testutils.DeleteUser(userID)
	}()

	// Step 1: Initiate authorization flow with OIDC scopes
	scope := "openid profile email"
	state := "test-state-123"
	resp, err := testutils.InitiateAuthorizationFlow("required_attrs_test_client", scopeTestRedirectURI, "code", scope, state)
	ts.Require().NoError(err, "Failed to initiate authorization flow")
	ts.Require().Equal(http.StatusFound, resp.StatusCode, "Should redirect to login page")

	// Extract auth ID and flow ID from redirect
	location := resp.Header.Get("Location")
	ts.Require().NotEmpty(location, "Location header should be present")
	authId, flowID, err := testutils.ExtractAuthData(location)
	ts.Require().NoError(err, "Failed to extract auth data")
	ts.Require().NotEmpty(authId, "Auth ID should not be empty")
	ts.Require().NotEmpty(flowID, "Flow ID should not be empty")

	// Step 2: Execute authentication flow
	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, map[string]string{
		"username": "requiredattrsuser",
		"password": "TestPassword123!",
	}, "action_001")
	ts.Require().NoError(err, "Failed to execute authentication flow")
	ts.Require().NotEmpty(flowStep.Assertion, "Assertion should be generated")

	// Step 3: Verify assertion contains only required attributes
	// Expected: sub (from openid), name (from profile scope, filtered by IDToken.UserAttributes),
	// email (from email scope, filtered by IDToken.UserAttributes), groups, roles (from AccessToken.UserAttributes)
	// Should NOT contain: phone, customAttr, email_verified (not in IDToken.UserAttributes)

	// Decode JWT to verify claims
	jwtClaims, err := testutils.DecodeJWT(flowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	claims := jwtClaims.Additional

	// Verify required attributes are present
	ts.Assert().NotNil(claims["sub"], "sub claim should be present")
	ts.Assert().Equal(userID, claims["sub"], "sub should match user ID")
	ts.Assert().NotNil(claims["name"], "name claim should be present (from profile scope)")
	ts.Assert().Equal("Required Attrs", claims["name"], "name should match user attribute")
	ts.Assert().NotNil(claims["email"], "email claim should be present (from email scope)")
	ts.Assert().Equal("requiredattrs@test.com", claims["email"], "email should match user attribute")
	ts.Assert().NotNil(claims["roles"], "roles claim should be present (from access token config)")
	rolesValue, ok := claims["roles"].([]interface{})
	ts.Assert().True(ok, "roles should be an array")
	ts.Assert().Equal([]interface{}{"developer"}, rolesValue, "roles should match user attribute")
	if claims["groups"] != nil {
		groupsValue, ok := claims["groups"].([]interface{})
		ts.Assert().True(ok, "groups should be an array if present")
		ts.Assert().NotEmpty(groupsValue, "groups array should not be empty if present")
	}

	// Verify non-required attributes are NOT present
	ts.Assert().Nil(claims["phone"], "phone should NOT be present (not in required attributes)")
	ts.Assert().Nil(claims["customAttr"], "customAttr should NOT be present (not in required attributes)")
	ts.Assert().Nil(claims["email_verified"], "email_verified should NOT be present (not in IDToken.UserAttributes)")
	ts.Assert().Nil(claims["firstName"], "firstName should NOT be present (not in required attributes)")
	ts.Assert().Nil(claims["lastName"], "lastName should NOT be present (not in required attributes)")
}
