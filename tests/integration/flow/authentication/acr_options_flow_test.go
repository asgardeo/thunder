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

package authentication

// ACR Options Flow Integration Tests
//
// Covers acceptance criteria from feature-docs/acr-values.md:
//
//   AC-10  A PROMPT node with authMethodMapping (login_options variant) is valid in the flow
//          graph and the /flow/execute endpoint returns its actions.
//   AC-11  When the next node is a login_options prompt, only actions whose ACR is in the
//          validated requested_auth_classes RuntimeData list are returned.
//   AC-12  The filtered actions are ordered to match the preference order of acr_values.
//   AC-13  When exactly one valid ACR option remains after filtering, the login_options node
//          is skipped and the next node is returned directly.
//   AC-14  The auth executor reads selected_auth_class from RuntimeData and verifies the
//          method's AMR satisfies the selected ACR; a mismatch returns an error.
//   AC-15  A /flow/execute call that bypasses the login_options node and submits credentials
//          for an AMR inconsistent with the selected ACR is rejected.
//   AC-16  The issued ID token contains an acr claim reflecting the ACR the user satisfied.
//   AC-17  At flow completion the satisfied ACR value is recorded in the auth assertion JWT.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/flow/common"
	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

// acrOptionsFlow is a flow that contains a login_options PROMPT node followed by two
// separate password-based authentication paths, each tagged to a distinct ACR value.
//
// Flow graph:
//   START → acr_chooser (login_options)
//     ├─ "pwd_action"  (acr: mosip:idp:acr:password)       → prompt_pwd  → basic_auth_pwd  → auth_assert → END
//     └─ "code_action" (acr: mosip:idp:acr:generated-code)  → prompt_code → basic_auth_code → auth_assert → END
var acrOptionsFlow = testutils.Flow{
	Name:     "ACR Options Auth Flow",
	FlowType: "AUTHENTICATION",
	Handle:   "auth_flow_acr_options_test",
	Nodes: []map[string]interface{}{
		{
			"id":        "start",
			"type":      "START",
			"onSuccess": "acr_chooser",
		},
		// login_options prompt node — the ACR chooser (AC-10)
		{
			"id":   "acr_chooser",
			"type": "PROMPT",
			"properties": map[string]interface{}{
				"authMethodMapping": map[string]interface{}{
					"mosip:idp:acr:password":       "pwd_action",
					"mosip:idp:acr:generated-code": "code_action",
					"mosip:idp:acr:biometrics":     "bio_action",
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
				{
					"action": map[string]interface{}{
						"ref":      "bio_action",
						"nextNode": "prompt_bio",
					},
				},
			},
		},
		// Credentials prompt for the password ACR path
		{
			"id":   "prompt_pwd",
			"type": "PROMPT",
			"prompts": []map[string]interface{}{
				{
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_u1",
							"identifier": "username",
							"type":       "TEXT_INPUT",
							"required":   true,
						},
						{
							"ref":        "input_p1",
							"identifier": "password",
							"type":       "PASSWORD_INPUT",
							"required":   true,
						},
					},
					"action": map[string]interface{}{
						"ref":      "submit_pwd",
						"nextNode": "basic_auth_pwd",
					},
				},
			},
		},
		// Credentials prompt for the generated-code (OTP-style password) ACR path.
		// For simplicity in tests both paths use the BasicAuthExecutor.
		{
			"id":   "prompt_code",
			"type": "PROMPT",
			"prompts": []map[string]interface{}{
				{
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_u2",
							"identifier": "username",
							"type":       "TEXT_INPUT",
							"required":   true,
						},
						{
							"ref":        "input_p2",
							"identifier": "password",
							"type":       "PASSWORD_INPUT",
							"required":   true,
						},
					},
					"action": map[string]interface{}{
						"ref":      "submit_code",
						"nextNode": "basic_auth_code",
					},
				},
			},
		},
		// Credentials prompt for the biometrics ACR path.
		// For simplicity in tests this path also uses the BasicAuthExecutor.
		{
			"id":   "prompt_bio",
			"type": "PROMPT",
			"prompts": []map[string]interface{}{
				{
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_u3",
							"identifier": "username",
							"type":       "TEXT_INPUT",
							"required":   true,
						},
						{
							"ref":        "input_p3",
							"identifier": "password",
							"type":       "PASSWORD_INPUT",
							"required":   true,
						},
					},
					"action": map[string]interface{}{
						"ref":      "submit_bio",
						"nextNode": "basic_auth_bio",
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
			"id":   "basic_auth_bio",
			"type": "TASK_EXECUTION",
			"executor": map[string]interface{}{
				"name": "BasicAuthExecutor",
				"inputs": []map[string]interface{}{
					{"ref": "input_u3", "identifier": "username", "type": "TEXT_INPUT", "required": true},
					{"ref": "input_p3", "identifier": "password", "type": "PASSWORD_INPUT", "required": true},
				},
			},
			"onSuccess":    "auth_assert",
			"onIncomplete": "prompt_bio",
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

// acrOptionsTestApp is a minimal OAuth2 application used across ACR options tests.
var acrOptionsTestApp = testutils.Application{
	Name:                      "ACR Options Flow Test Application",
	Description:               "Application for testing login_options flow behaviour",
	IsRegistrationFlowEnabled: false,
	ClientID:                  "acr_options_flow_test_client",
	ClientSecret:              "acr_options_flow_test_secret",
	RedirectURIs:              []string{"https://localhost:3000/acr-options-callback"},
	AllowedUserTypes:          []string{"acr_options_test_person"},
	InboundAuthConfig: []map[string]interface{}{
		{
			"type": "oauth2",
			"config": map[string]interface{}{
				"clientId":                "acr_options_flow_test_client",
				"clientSecret":            "acr_options_flow_test_secret",
				"redirectUris":            []string{"https://localhost:3000/acr-options-callback"},
				"grantTypes":              []string{"authorization_code"},
				"responseTypes":           []string{"code"},
				"tokenEndpointAuthMethod": "client_secret_basic",
				"defaultAcrValues": []string{
					"mosip:idp:acr:password",
					"mosip:idp:acr:generated-code",
					"mosip:idp:acr:biometrics",
				},
			},
		},
	},
}

var acrOptionsTestOU = testutils.OrganizationUnit{
	Handle:      "acr-options-flow-test-ou",
	Name:        "ACR Options Flow Test Organization Unit",
	Description: "Organization unit for ACR options flow testing",
	Parent:      nil,
}

var acrOptionsUserSchema = testutils.UserSchema{
	Name: "acr_options_test_person",
	Schema: map[string]interface{}{
		"username": map[string]interface{}{"type": "string"},
		"password": map[string]interface{}{"type": "string", "credential": true},
		"email":    map[string]interface{}{"type": "string"},
	},
}

var acrOptionsTestUser = testutils.User{
	Type: acrOptionsUserSchema.Name,
	Attributes: json.RawMessage(`{
		"username": "acroptionsuser",
		"password": "testpassword",
		"email": "acroptionsuser@example.com"
	}`),
}

var (
	acrOptionsTestAppID    string
	acrOptionsTestOUID     string
	acrOptionsFlowID       string
	acrOptionsUserSchemaID string
)

// AcrOptionsFlowTestSuite tests the login_options PROMPT node filtering, ordering,
// auto-selection, AMR validation, and the ACR claim in the auth assertion JWT.
type AcrOptionsFlowTestSuite struct {
	suite.Suite
	config *common.TestSuiteConfig
}

func TestAcrOptionsFlowTestSuite(t *testing.T) {
	suite.Run(t, new(AcrOptionsFlowTestSuite))
}

func (ts *AcrOptionsFlowTestSuite) SetupSuite() {
	ts.config = &common.TestSuiteConfig{}

	err := testutils.PatchDeploymentConfig(map[string]interface{}{
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

	ouID, err := testutils.CreateOrganizationUnit(acrOptionsTestOU)
	ts.Require().NoError(err, "failed to create OU for ACR options tests")
	acrOptionsTestOUID = ouID

	acrOptionsUserSchema.OUID = acrOptionsTestOUID
	schemaID, err := testutils.CreateUserType(acrOptionsUserSchema)
	ts.Require().NoError(err, "failed to create user schema for ACR options tests")
	acrOptionsUserSchemaID = schemaID

	user := acrOptionsTestUser
	user.OUID = acrOptionsTestOUID
	userIDs, err := testutils.CreateMultipleUsers(user)
	ts.Require().NoError(err, "failed to create test user for ACR options tests")
	ts.config.CreatedUserIDs = userIDs

	flowID, err := testutils.CreateFlow(acrOptionsFlow)
	ts.Require().NoError(err, "failed to create ACR options flow")
	acrOptionsFlowID = flowID
	ts.config.CreatedFlowIDs = append(ts.config.CreatedFlowIDs, flowID)

	acrOptionsTestApp.AuthFlowID = flowID
	appID, err := testutils.CreateApplication(acrOptionsTestApp)
	ts.Require().NoError(err, "failed to create ACR options test application")
	acrOptionsTestAppID = appID
}

func (ts *AcrOptionsFlowTestSuite) TearDownSuite() {
	if err := testutils.PatchDeploymentConfig(map[string]interface{}{
		"acr_amr_mapping": nil,
	}); err != nil {
		ts.T().Logf("failed to restore ACR-AMR mapping in deployment config: %v", err)
	}

	if err := testutils.CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("failed to cleanup users: %v", err)
	}
	if acrOptionsTestAppID != "" {
		if err := testutils.DeleteApplication(acrOptionsTestAppID); err != nil {
			ts.T().Logf("failed to delete test application: %v", err)
		}
	}
	for _, id := range ts.config.CreatedFlowIDs {
		if err := testutils.DeleteFlow(id); err != nil {
			ts.T().Logf("failed to delete flow %s: %v", id, err)
		}
	}
	if acrOptionsUserSchemaID != "" {
		if err := testutils.DeleteUserType(acrOptionsUserSchemaID); err != nil {
			ts.T().Logf("failed to delete user schema: %v", err)
		}
	}
	if acrOptionsTestOUID != "" {
		if err := testutils.DeleteOrganizationUnit(acrOptionsTestOUID); err != nil {
			ts.T().Logf("failed to delete OU: %v", err)
		}
	}
}

// ------- AC-10: login_options node is valid in a flow graph ----------

// TestAcrOptions_NodeValidInFlowGraph verifies that a PROMPT node with authMethodMapping
// (login_options variant) is accepted by the flow engine and returns its actions.
func (ts *AcrOptionsFlowTestSuite) TestAcrOptions_NodeValidInFlowGraph() {
	// Initiate flow without any acr_values — engine must present both ACR options.
	flowStep, err := common.InitiateAuthenticationFlow(acrOptionsTestAppID, false, nil, "")
	ts.Require().NoError(err, "flow initiation should succeed")

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)
	ts.Require().Equal("VIEW", flowStep.Type)
	ts.Require().NotEmpty(flowStep.FlowID)

	// Both actions must be present (AC-10).
	ts.Require().NotEmpty(flowStep.Data.Actions, "login_options node should return actions")
	ts.Require().True(
		common.HasAction(flowStep.Data.Actions, "pwd_action"),
		"pwd_action should be present")
	ts.Require().True(
		common.HasAction(flowStep.Data.Actions, "code_action"),
		"code_action should be present")
}

// ------- AC-11: actions filtered to requested ACRs -------------------

// TestAcrOptions_FilteredToRequestedACR verifies that when acr_values is propagated into
// RuntimeData, only the matching actions are returned by the login_options node.
// Three ACRs are in the flow; two are requested so that filtering is exercised while keeping
// more than one option — preventing AC-13 auto-selection from masking the filter behaviour.
func (ts *AcrOptionsFlowTestSuite) TestAcrOptions_FilteredToRequestedACR() {
	// Request password and biometrics ACRs — generated-code must be filtered out.
	// Two options remain so auto-selection (AC-13) does not trigger.
	authID, flowID, err := ts.initiateAuthorizeAndExtract("mosip:idp:acr:password mosip:idp:acr:biometrics")
	ts.Require().NoError(err, "authorization initiation should succeed")
	ts.Require().NotEmpty(authID)
	ts.Require().NotEmpty(flowID)

	// Resume the already-initiated flow — engine should filter to the two requested ACRs.
	flowStep, err := common.ResumeFlow(flowID)
	ts.Require().NoError(err, "flow resumption should succeed")

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)
	// Only the requested ACR actions must be visible; generated-code must be absent (AC-11).
	actionRefs := ts.actionRefs(flowStep.Data.Actions)
	ts.Require().Contains(actionRefs, "pwd_action", "pwd_action should be present")
	ts.Require().Contains(actionRefs, "bio_action", "bio_action should be present")
	ts.Require().NotContains(actionRefs, "code_action", "code_action should be filtered out")
}

// ------- AC-12: preference ordering ----------------------------------

// TestAcrOptions_OrderedByPreference verifies that when multiple ACR values are requested
// the returned actions are ordered to match the preference order of acr_values.
func (ts *AcrOptionsFlowTestSuite) TestAcrOptions_OrderedByPreference() {
	// Request generated-code first, then password — the engine must return them in that order.
	authID, flowID, err := ts.initiateAuthorizeAndExtract("mosip:idp:acr:generated-code mosip:idp:acr:password")
	ts.Require().NoError(err)
	ts.Require().NotEmpty(authID)
	ts.Require().NotEmpty(flowID)

	flowStep, err := common.ResumeFlow(flowID)
	ts.Require().NoError(err)

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)
	ts.Require().Len(flowStep.Data.Actions, 2, "both ACR actions should be present")

	// First action must correspond to the first requested ACR (AC-12).
	ts.Require().Equal("code_action", flowStep.Data.Actions[0].Ref,
		"generated-code action should be first")
	ts.Require().Equal("pwd_action", flowStep.Data.Actions[1].Ref,
		"password action should be second")
}

// ------- AC-13: auto-selection when only one ACR remains -------------

// TestAcrOptions_AutoSelectsWhenSingleACR verifies that when only one valid ACR option
// remains after filtering the login_options node is skipped and the credentials prompt
// for that ACR is returned immediately.
func (ts *AcrOptionsFlowTestSuite) TestAcrOptions_AutoSelectsWhenSingleACR() {
	authID, flowID, err := ts.initiateAuthorizeAndExtract("mosip:idp:acr:password")
	ts.Require().NoError(err)
	ts.Require().NotEmpty(authID)
	ts.Require().NotEmpty(flowID)

	flowStep, err := common.ResumeFlow(flowID)
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)

	// With a single valid ACR the chooser node must be auto-skipped (AC-13).
	// The response must contain the credential inputs for the password path.
	// The chooser-level actions (pwd_action, code_action, bio_action) must be absent;
	// the credential submit action (submit_pwd) from prompt_pwd is expected.
	ts.Require().NotEmpty(flowStep.Data.Inputs, "credential inputs should be returned after auto-selection")
	ts.Require().True(
		common.HasInput(flowStep.Data.Inputs, "username"),
		"username input should be returned for the password path")
	ts.Require().True(
		common.HasInput(flowStep.Data.Inputs, "password"),
		"password input should be returned for the password path")
	ts.Require().False(
		common.HasAction(flowStep.Data.Actions, "pwd_action"),
		"chooser action must not be present after auto-selection")
	ts.Require().False(
		common.HasAction(flowStep.Data.Actions, "code_action"),
		"chooser action must not be present after auto-selection")
	ts.Require().False(
		common.HasAction(flowStep.Data.Actions, "bio_action"),
		"chooser action must not be present after auto-selection")
}

// ------- AC-17: satisfied ACR in auth assertion JWT ------------------

// TestAcrOptions_AcrInAuthAssertionJWT verifies that after a user selects an ACR option
// and authenticates successfully, the satisfied ACR value is recorded in the auth assertion
// JWT (AC-17), which is used by the token builder to populate the id_token acr claim (AC-16).
func (ts *AcrOptionsFlowTestSuite) TestAcrOptions_AcrInAuthAssertionJWT() {
	// Step 1: Start the flow without acr filtering so the chooser is shown.
	flowStep, err := common.InitiateAuthenticationFlow(acrOptionsTestAppID, false, nil, "")
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)
	ts.Require().NotEmpty(flowStep.Data.Actions)

	// Step 2: Select the password ACR action.
	credStep, err := common.CompleteFlow(flowStep.FlowID, map[string]string{}, "pwd_action")
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", credStep.FlowStatus)
	ts.Require().NotEmpty(credStep.Data.Inputs, "credential inputs should be requested")

	// Step 3: Submit valid credentials.
	var userAttrs map[string]interface{}
	ts.Require().NoError(json.Unmarshal(acrOptionsTestUser.Attributes, &userAttrs))
	inputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}
	completeStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "submit_pwd")
	ts.Require().NoError(err)

	ts.Require().Equal("COMPLETE", completeStep.FlowStatus, "flow should complete successfully")
	ts.Require().NotEmpty(completeStep.Assertion, "auth assertion JWT must be present")

	// Decode the assertion and verify the completed_auth_class claim is set.
	jwtClaims, err := testutils.DecodeJWT(completeStep.Assertion)
	ts.Require().NoError(err, "assertion JWT must be decodable")
	ts.Require().NotNil(jwtClaims)

	acrClaim, ok := jwtClaims.Additional["completed_auth_class"]
	ts.Require().True(ok, "completed_auth_class claim must be present in the auth assertion JWT")
	ts.Require().Equal("mosip:idp:acr:password", acrClaim,
		"completed_auth_class claim must reflect the selected authentication class")
}

// ------- AC-14 & AC-15: AMR validation in the auth executor ----------

// TestAcrOptions_AMRMismatchRejected verifies that the auth executor rejects a credential
// submission whose AMR type does not satisfy the ACR that was selected via the
// login_options node (AC-14 / AC-15).
//
// Scenario: the user selects the generated-code ACR option but then submits BasicAuth
// credentials — BasicAuthExecutor's AMR (pwd) does not satisfy the generated-code ACR
// which requires an OTP-class AMR.
func (ts *AcrOptionsFlowTestSuite) TestAcrOptions_AMRMismatchRejected() {
	// Step 1: Initiate flow.
	flowStep, err := common.InitiateAuthenticationFlow(acrOptionsTestAppID, false, nil, "")
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)

	// Step 2: Select the generated-code ACR action.
	credStep, err := common.CompleteFlow(flowStep.FlowID, map[string]string{}, "code_action")
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", credStep.FlowStatus)

	// Step 3: Submit valid credentials through the basic_auth_code path.
	// The BasicAuthExecutor provides AMR "pwd"; the selected ACR (generated-code)
	// requires an OTP-class AMR → executor must reject the request.
	var userAttrs map[string]interface{}
	ts.Require().NoError(json.Unmarshal(acrOptionsTestUser.Attributes, &userAttrs))
	inputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}
	resultStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "submit_code")
	ts.Require().NoError(err)

	// The flow must not complete successfully due to AMR mismatch (AC-14).
	ts.Require().NotEqual("COMPLETE", resultStep.FlowStatus,
		"flow must not complete when AMR does not satisfy selected ACR")
	ts.Require().NotEmpty(resultStep.FailureReason,
		"a failure reason must be provided on AMR mismatch")
}

// TestAcrOptions_BypassACRChooserRejected verifies that submitting credentials directly
// into an executor whose AMR is inconsistent with the selected ACR — even without going
// through the chooser — is rejected (AC-15).
func (ts *AcrOptionsFlowTestSuite) TestAcrOptions_BypassACRChooserRejected() {
	// Trigger an authorize request that pre-sets selected_auth_class to generated-code
	// via acr_values; since only one ACR is requested the engine auto-selects it.
	authID, flowID, err := ts.initiateAuthorizeAndExtract("mosip:idp:acr:generated-code")
	ts.Require().NoError(err)
	ts.Require().NotEmpty(authID)
	ts.Require().NotEmpty(flowID)

	// The auto-selection forwards to the generated-code credentials prompt.
	credStep, err := common.ResumeFlow(flowID)
	ts.Require().NoError(err)
	ts.Require().Equal("INCOMPLETE", credStep.FlowStatus)

	// Submit credentials into the basic_auth_code executor. BasicAuth provides "pwd"
	// AMR which does not satisfy generated-code ACR (AC-15).
	var userAttrs map[string]interface{}
	ts.Require().NoError(json.Unmarshal(acrOptionsTestUser.Attributes, &userAttrs))
	inputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}
	resultStep, err := common.CompleteFlow(flowID, inputs, "submit_code")
	ts.Require().NoError(err)

	ts.Require().NotEqual("COMPLETE", resultStep.FlowStatus,
		"bypassing ACR chooser with an incompatible AMR must be rejected (AC-15)")
	ts.Require().NotEmpty(resultStep.FailureReason,
		"a failure reason must be returned on AMR mismatch after auto-selection")
}

// ------- helpers -------------------------------------------------------

// actionRefs returns the Ref fields of the given actions as a string slice.
func (ts *AcrOptionsFlowTestSuite) actionRefs(actions []common.Action) []string {
	refs := make([]string, len(actions))
	for i, a := range actions {
		refs[i] = a.Ref
	}
	return refs
}

// initiateAuthorizeAndExtract sends GET /oauth2/authorize with the given acr_values and
// returns the authID and flowID extracted from the redirect location.
func (ts *AcrOptionsFlowTestSuite) initiateAuthorizeAndExtract(acrValues string) (string, string, error) {
	params := url.Values{}
	params.Set("client_id", "acr_options_flow_test_client")
	params.Set("redirect_uri", "https://localhost:3000/acr-options-callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", "acr_state")
	if acrValues != "" {
		params.Set("acr_values", acrValues)
	}

	resp, err := testutils.GetNoRedirectHTTPClient().Get(testutils.TestServerURL + "/oauth2/authorize?" + params.Encode())
	if err != nil {
		return "", "", fmt.Errorf("authorize request failed: %w", err)
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		return "", "", fmt.Errorf("no Location header in authorize response (status %d)", resp.StatusCode)
	}

	authID, flowID, err := testutils.ExtractAuthData(location)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract auth data from location %q: %w", location, err)
	}
	return authID, flowID, nil
}
