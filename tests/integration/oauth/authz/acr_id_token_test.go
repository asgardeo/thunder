/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

// ACR ID Token E2E Tests
//
// Verifies that the `acr` claim propagates through the full OAuth2 authorization code flow
// into the issued ID token (AC-16) and that the auth assertion JWT carries the selected ACR
// (AC-17) so the token builder can include it.

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	acrE2EClientID     = "acr_e2e_test_client"
	acrE2EClientSecret = "acr_e2e_test_secret"
	acrE2ERedirectURI  = "https://localhost:3000/acr-e2e-callback"
)

// acrE2EFlow is a flow that includes a login_options prompt node and password authentication,
// followed by the AuthAssertExecutor that records the selected ACR in the assertion JWT.
var acrE2EFlow = testutils.Flow{
	Name:     "ACR E2E ID Token Flow",
	FlowType: "AUTHENTICATION",
	Handle:   "auth_flow_acr_e2e_test",
	Nodes: []map[string]interface{}{
		{
			"id":        "start",
			"type":      "START",
			"onSuccess": "acr_chooser",
		},
		{
			"id":   "acr_chooser",
			"type": "PROMPT",
			"properties": map[string]interface{}{
				"authMethodMapping": map[string]interface{}{
					"mosip:idp:acr:password":       "pwd_action",
					"mosip:idp:acr:generated-code": "code_action",
				},
			},
			"prompts": []map[string]interface{}{
				{
					"action": map[string]interface{}{
						"ref":      "pwd_action",
						"nextNode": "prompt_pwd",
					},
				},
				{
					"action": map[string]interface{}{
						"ref":      "code_action",
						"nextNode": "prompt_code",
					},
				},
			},
		},
		{
			"id":   "prompt_pwd",
			"type": "PROMPT",
			"prompts": []map[string]interface{}{
				{
					"inputs": []map[string]interface{}{
						{"ref": "input_u1", "identifier": "username", "type": "TEXT_INPUT", "required": true},
						{"ref": "input_p1", "identifier": "password", "type": "PASSWORD_INPUT", "required": true},
					},
					"action": map[string]interface{}{
						"ref":      "submit_pwd",
						"nextNode": "basic_auth_pwd",
					},
				},
			},
		},
		{
			"id":   "prompt_code",
			"type": "PROMPT",
			"prompts": []map[string]interface{}{
				{
					"inputs": []map[string]interface{}{
						{"ref": "input_u2", "identifier": "username", "type": "TEXT_INPUT", "required": true},
						{"ref": "input_p2", "identifier": "password", "type": "PASSWORD_INPUT", "required": true},
					},
					"action": map[string]interface{}{
						"ref":      "submit_code",
						"nextNode": "basic_auth_code",
					},
				},
			},
		},
		{
			"id":   "basic_auth_pwd",
			"type": "TASK_EXECUTION",
			"executor": map[string]interface{}{
				"name": "BasicAuthExecutor",
				"inputs": []map[string]interface{}{
					{"ref": "input_u1", "identifier": "username", "type": "TEXT_INPUT", "required": true},
					{"ref": "input_p1", "identifier": "password", "type": "PASSWORD_INPUT", "required": true},
				},
			},
			"onSuccess":    "auth_assert",
			"onIncomplete": "prompt_pwd",
		},
		{
			"id":   "basic_auth_code",
			"type": "TASK_EXECUTION",
			"executor": map[string]interface{}{
				"name": "BasicAuthExecutor",
				"inputs": []map[string]interface{}{
					{"ref": "input_u2", "identifier": "username", "type": "TEXT_INPUT", "required": true},
					{"ref": "input_p2", "identifier": "password", "type": "PASSWORD_INPUT", "required": true},
				},
			},
			"onSuccess":    "auth_assert",
			"onIncomplete": "prompt_code",
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

var acrE2ETestOU = testutils.OrganizationUnit{
	Handle:      "acr-e2e-test-ou",
	Name:        "ACR E2E Test Organization Unit",
	Description: "Organization unit for ACR E2E ID token testing",
	Parent:      nil,
}

var acrE2EUserSchema = testutils.UserSchema{
	Name: "acr_e2e_test_person",
	Schema: map[string]interface{}{
		"username": map[string]interface{}{"type": "string"},
		"password": map[string]interface{}{"type": "string", "credential": true},
		"email":    map[string]interface{}{"type": "string"},
	},
}

var acrE2ETestUser = testutils.User{
	Type: acrE2EUserSchema.Name,
	Attributes: json.RawMessage(`{
		"username": "acre2euser",
		"password": "testpassword",
		"email": "acre2euser@example.com"
	}`),
}

// AcrIDTokenTestSuite verifies that the acr claim is present in the ID token
// issued through the full OAuth2 authorization code flow (AC-16).
type AcrIDTokenTestSuite struct {
	suite.Suite
	client                *http.Client
	applicationID         string
	authFlowID            string
	ouID                  string
	userSchemaID          string
	userIDs               []string
	previousAcrAmrMapping interface{}
}

func TestAcrIDTokenTestSuite(t *testing.T) {
	suite.Run(t, new(AcrIDTokenTestSuite))
}

func (ts *AcrIDTokenTestSuite) SetupSuite() {
	ts.client = testutils.GetHTTPClient()

	// Save existing ACR-AMR mapping so it can be restored in TearDownSuite.
	prev, err := testutils.ReadDeploymentConfigKey("acr_amr_mapping")
	ts.Require().NoError(err, "failed to read existing acr_amr_mapping from deployment config")
	ts.previousAcrAmrMapping = prev

	// Patch deployment config with ACR-AMR mapping so that acr_values are recognized.
	err = testutils.PatchDeploymentConfig(map[string]interface{}{
		"acr_amr_mapping": map[string]interface{}{
			"amr": map[string]interface{}{
				"Password":      map[string]interface{}{"type": "PWD"},
				"OTP":           map[string]interface{}{"type": "OTP"},
				"L1-bio-device": map[string]interface{}{"type": "BIO", "count": 1},
			},
			"acr_amr": map[string]interface{}{
				"mosip:idp:acr:password":       []string{"Password"},
				"mosip:idp:acr:generated-code": []string{"OTP"},
				"mosip:idp:acr:biometrics":     []string{"L1-bio-device"},
			},
		},
	})
	ts.Require().NoError(err, "failed to patch deployment config with ACR-AMR mapping")
	ts.Require().NoError(testutils.RestartServer(), "failed to restart server after config patch")

	ouID, err := testutils.CreateOrganizationUnit(acrE2ETestOU)
	ts.Require().NoError(err, "failed to create OU")
	ts.ouID = ouID

	schema := acrE2EUserSchema
	schema.OUID = ts.ouID
	schemaID, err := testutils.CreateUserType(schema)
	ts.Require().NoError(err, "failed to create user schema")
	ts.userSchemaID = schemaID

	user := acrE2ETestUser
	user.OUID = ts.ouID
	userIDs, err := testutils.CreateMultipleUsers(user)
	ts.Require().NoError(err, "failed to create test user")
	ts.userIDs = userIDs

	flowID, err := testutils.CreateFlow(acrE2EFlow)
	ts.Require().NoError(err, "failed to create ACR E2E flow")
	ts.authFlowID = flowID

	app := testutils.Application{
		Name:                      "ACR E2E ID Token Test App",
		Description:               "Application for ACR E2E ID token tests",
		IsRegistrationFlowEnabled: false,
		AuthFlowID:                ts.authFlowID,
		ClientID:                  acrE2EClientID,
		ClientSecret:              acrE2EClientSecret,
		RedirectURIs:              []string{acrE2ERedirectURI},
		AllowedUserTypes:          []string{acrE2EUserSchema.Name},
		InboundAuthConfig: []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"clientId":                acrE2EClientID,
					"clientSecret":            acrE2EClientSecret,
					"redirectUris":            []string{acrE2ERedirectURI},
					"grantTypes":              []string{"authorization_code"},
					"responseTypes":           []string{"code"},
					"tokenEndpointAuthMethod": "client_secret_basic",
					"defaultAcrValues": []string{
						"mosip:idp:acr:password",
						"mosip:idp:acr:generated-code",
					},
				},
			},
		},
	}

	appID, err := testutils.CreateApplication(app)
	ts.Require().NoError(err, "failed to create ACR E2E test application")
	ts.applicationID = appID
}

func (ts *AcrIDTokenTestSuite) TearDownSuite() {
	if err := testutils.CleanupUsers(ts.userIDs); err != nil {
		ts.T().Logf("failed to cleanup users: %v", err)
	}
	if ts.applicationID != "" {
		if err := testutils.DeleteApplication(ts.applicationID); err != nil {
			ts.T().Logf("failed to delete application: %v", err)
		}
	}
	if ts.authFlowID != "" {
		if err := testutils.DeleteFlow(ts.authFlowID); err != nil {
			ts.T().Logf("failed to delete flow: %v", err)
		}
	}
	if ts.userSchemaID != "" {
		if err := testutils.DeleteUserType(ts.userSchemaID); err != nil {
			ts.T().Logf("failed to delete user schema: %v", err)
		}
	}
	if ts.ouID != "" {
		if err := testutils.DeleteOrganizationUnit(ts.ouID); err != nil {
			ts.T().Logf("failed to delete OU: %v", err)
		}
	}

	// Restore the previous ACR-AMR mapping in deployment config.
	if err := testutils.PatchDeploymentConfig(map[string]interface{}{
		"acr_amr_mapping": ts.previousAcrAmrMapping,
	}); err != nil {
		ts.T().Logf("failed to restore ACR-AMR mapping in deployment config: %v", err)
	}
	if err := testutils.RestartServer(); err != nil {
		ts.T().Logf("failed to restart server after config restore: %v", err)
	}
}

// TestAcrClaimInIDToken performs the full OAuth2 authorization code flow with acr_values
// and verifies that the issued ID token contains the correct acr claim (AC-16).
//
// Flow: authorize(acr_values=password) → auto-select → submit creds → assertion →
//
//	callback → token exchange → decode ID token → assert acr claim.
func (ts *AcrIDTokenTestSuite) TestAcrClaimInIDToken() {
	// Step 1: Initiate authorization with acr_values requesting only password.
	// Since only one ACR matches the app's default list, the login_options node is auto-skipped.
	resp, err := ts.initiateAuthorize("mosip:idp:acr:password")
	ts.Require().NoError(err)
	defer resp.Body.Close()

	ts.Require().Equal(http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	ts.Require().NotEmpty(location)

	authID, flowID, err := testutils.ExtractAuthData(location)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(authID)
	ts.Require().NotEmpty(flowID)

	// Step 2: Resume the flow — auto-selection should forward to the password credentials prompt.
	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, nil, "")
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)

	// Step 3: Submit valid credentials.
	flowStep, err = testutils.ExecuteAuthenticationFlow(flowID, map[string]string{
		"username": "acre2euser",
		"password": "testpassword",
	}, "submit_pwd")
	ts.Require().NoError(err)
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "flow should complete after valid credentials")
	ts.Require().NotEmpty(flowStep.Assertion, "assertion must be present")

	// Step 4: Complete authorization with the assertion to get a redirect with auth code.
	authzResp, err := testutils.CompleteAuthorization(authID, flowStep.Assertion)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(authzResp.RedirectURI)

	code, err := testutils.ExtractAuthorizationCode(authzResp.RedirectURI)
	ts.Require().NoError(err, "authorization code must be in the redirect")
	ts.Require().NotEmpty(code)

	// Step 5: Exchange code for tokens.
	tokenResult, err := testutils.RequestToken(
		acrE2EClientID, acrE2EClientSecret, code, acrE2ERedirectURI, "authorization_code")
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, tokenResult.StatusCode, "token exchange should succeed")
	ts.Require().NotNil(tokenResult.Token)
	ts.Require().NotEmpty(tokenResult.Token.IDToken, "ID token must be present in token response")

	// Step 6: Decode the ID token and verify the acr claim (AC-16).
	idTokenClaims, err := testutils.DecodeJWT(tokenResult.Token.IDToken)
	ts.Require().NoError(err, "ID token must be decodable")
	ts.Require().NotNil(idTokenClaims)

	acrClaim, ok := idTokenClaims.Additional["acr"]
	ts.Require().True(ok, "acr claim must be present in the ID token (AC-16)")
	ts.Require().Equal("mosip:idp:acr:password", acrClaim,
		"acr claim in ID token must reflect the ACR the user satisfied")
}

// TestAcrClaimInIDToken_WithManualSelection performs the full OAuth flow where the user
// manually selects an ACR option from the login_options node, and verifies the acr claim
// in the ID token matches the selected option.
func (ts *AcrIDTokenTestSuite) TestAcrClaimInIDToken_WithManualSelection() {
	// Step 1: Initiate authorization with both ACR values so the chooser is presented.
	resp, err := ts.initiateAuthorize("mosip:idp:acr:generated-code mosip:idp:acr:password")
	ts.Require().NoError(err)
	defer resp.Body.Close()

	ts.Require().Equal(http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	ts.Require().NotEmpty(location)

	authID, flowID, err := testutils.ExtractAuthData(location)
	ts.Require().NoError(err)

	// Step 2: Resume the flow — both ACR options should be presented.
	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, nil, "")
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)

	// Step 3: Select the password ACR action.
	flowStep, err = testutils.ExecuteAuthenticationFlow(flowID, nil, "pwd_action")
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)

	// Step 4: Submit credentials.
	flowStep, err = testutils.ExecuteAuthenticationFlow(flowID, map[string]string{
		"username": "acre2euser",
		"password": "testpassword",
	}, "submit_pwd")
	ts.Require().NoError(err)
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus)
	ts.Require().NotEmpty(flowStep.Assertion)

	// Step 5: Complete authorization and exchange code for tokens.
	authzResp, err := testutils.CompleteAuthorization(authID, flowStep.Assertion)
	ts.Require().NoError(err)

	code, err := testutils.ExtractAuthorizationCode(authzResp.RedirectURI)
	ts.Require().NoError(err)

	tokenResult, err := testutils.RequestToken(
		acrE2EClientID, acrE2EClientSecret, code, acrE2ERedirectURI, "authorization_code")
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, tokenResult.StatusCode)
	ts.Require().NotNil(tokenResult.Token)
	ts.Require().NotEmpty(tokenResult.Token.IDToken)

	// Step 6: Verify the acr claim matches the manually selected ACR.
	idTokenClaims, err := testutils.DecodeJWT(tokenResult.Token.IDToken)
	ts.Require().NoError(err)

	acrClaim, ok := idTokenClaims.Additional["acr"]
	ts.Require().True(ok, "acr claim must be present in the ID token (AC-16)")
	ts.Require().Equal("mosip:idp:acr:password", acrClaim,
		"acr claim must reflect the manually selected authentication class")
}

// TestAcrClaimPresentWithDefaultAcrValues verifies that when no explicit acr_values
// parameter is sent in the authorize request, the server falls back to the application's
// defaultAcrValues and the ID token still contains the correct acr claim.
func (ts *AcrIDTokenTestSuite) TestAcrClaimPresentWithDefaultAcrValues() {
	// Step 1: Initiate authorization without explicit acr_values.
	// The server falls back to the app's defaultAcrValues.
	resp, err := ts.initiateAuthorize("")
	ts.Require().NoError(err)
	defer resp.Body.Close()

	ts.Require().Equal(http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	ts.Require().NotEmpty(location)

	authID, flowID, err := testutils.ExtractAuthData(location)
	ts.Require().NoError(err)

	// Step 2: Resume flow — both default ACR options are presented.
	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, nil, "")
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)

	// Step 3: Select the password option and authenticate.
	flowStep, err = testutils.ExecuteAuthenticationFlow(flowID, nil, "pwd_action")
	ts.Require().NoError(err)

	flowStep, err = testutils.ExecuteAuthenticationFlow(flowID, map[string]string{
		"username": "acre2euser",
		"password": "testpassword",
	}, "submit_pwd")
	ts.Require().NoError(err)
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus)
	ts.Require().NotEmpty(flowStep.Assertion)

	// Step 4: Complete OAuth flow.
	authzResp, err := testutils.CompleteAuthorization(authID, flowStep.Assertion)
	ts.Require().NoError(err)

	code, err := testutils.ExtractAuthorizationCode(authzResp.RedirectURI)
	ts.Require().NoError(err)

	tokenResult, err := testutils.RequestToken(
		acrE2EClientID, acrE2EClientSecret, code, acrE2ERedirectURI, "authorization_code")
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, tokenResult.StatusCode)
	ts.Require().NotNil(tokenResult.Token)
	ts.Require().NotEmpty(tokenResult.Token.IDToken)

	// Step 5: Decode ID token — acr claim should be present because the app's
	// defaultAcrValues are used and the login_options node records the selected ACR.
	idTokenClaims, err := testutils.DecodeJWT(tokenResult.Token.IDToken)
	ts.Require().NoError(err)

	acrClaim, ok := idTokenClaims.Additional["acr"]
	ts.Require().True(ok, "acr claim must be present when defaultAcrValues are configured")
	ts.Require().Equal("mosip:idp:acr:password", acrClaim,
		"acr claim should reflect the selected authentication class")
}

// initiateAuthorize sends GET /oauth2/authorize with the given acr_values parameter.
func (ts *AcrIDTokenTestSuite) initiateAuthorize(acrValues string) (*http.Response, error) {
	params := url.Values{}
	params.Set("client_id", acrE2EClientID)
	params.Set("redirect_uri", acrE2ERedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", "acr_e2e_state")
	if acrValues != "" {
		params.Set("acr_values", acrValues)
	}

	req, err := http.NewRequest(http.MethodGet,
		testutils.TestServerURL+"/oauth2/authorize?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	return testutils.GetNoRedirectHTTPClient().Do(req)
}
