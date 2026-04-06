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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	acrClientID     = "acr_authz_test_client"
	acrClientSecret = "acr_authz_test_secret"
	acrAppName      = "ACRAuthzTestApp"
	acrRedirectURI  = "https://localhost:3000/acr-callback"
)

// AcrValuesAuthzTestSuite tests acr_values behaviour in the authorization endpoint.
type AcrValuesAuthzTestSuite struct {
	suite.Suite
	client        *http.Client
	applicationID string
	authFlowID    string
}

func TestAcrValuesAuthzTestSuite(t *testing.T) {
	suite.Run(t, new(AcrValuesAuthzTestSuite))
}

func (ts *AcrValuesAuthzTestSuite) SetupSuite() {
	ts.client = testutils.GetHTTPClient()

	// Patch deployment config with ACR-AMR mapping so that default_acr_values are accepted.
	err := testutils.PatchDeploymentConfig(map[string]interface{}{
		"acr_amr_mapping": map[string]interface{}{
			"amr": map[string]interface{}{
				"Password":    map[string]interface{}{"type": "PWD"},
				"OTP":         map[string]interface{}{"type": "OTP"},
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

	// Create authentication flow
	flowID, err := testutils.CreateFlow(testAuthFlow)
	ts.Require().NoError(err, "failed to create auth flow for ACR values test")
	ts.authFlowID = flowID

	// Create application with DefaultAcrValues configured
	app := map[string]interface{}{
		"name":                      acrAppName,
		"description":               "Application for acr_values authorization integration tests",
		"authFlowId":                ts.authFlowID,
		"isRegistrationFlowEnabled": false,
		"inboundAuthConfig": []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"clientId":     acrClientID,
					"clientSecret": acrClientSecret,
					"redirectUris": []string{acrRedirectURI},
					"grantTypes":   []string{"authorization_code"},
					"responseTypes": []string{
						"code",
					},
					"tokenEndpointAuthMethod": "client_secret_basic",
					"defaultAcrValues": []string{
						"mosip:idp:acr:password",
						"mosip:idp:acr:generated-code",
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(app)
	ts.Require().NoError(err)

	req, err := http.NewRequest(http.MethodPost, testutils.TestServerURL+"/applications", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("failed to create ACR test application: status=%d body=%s", resp.StatusCode, body)
	}

	var respData map[string]interface{}
	ts.Require().NoError(json.NewDecoder(resp.Body).Decode(&respData))
	ts.applicationID = respData["id"].(string)
}

func (ts *AcrValuesAuthzTestSuite) TearDownSuite() {
	if ts.applicationID != "" {
		req, err := http.NewRequest(http.MethodDelete,
			fmt.Sprintf("%s/applications/%s", testutils.TestServerURL, ts.applicationID), nil)
		if err != nil {
			ts.T().Logf("failed to build delete request: %v", err)
			return
		}
		resp, err := ts.client.Do(req)
		if err != nil {
			ts.T().Logf("failed to delete ACR test application: %v", err)
			return
		}
		resp.Body.Close()
	}

	if ts.authFlowID != "" {
		if err := testutils.DeleteFlow(ts.authFlowID); err != nil {
			ts.T().Logf("failed to delete ACR test auth flow: %v", err)
		}
	}

	// Restore deployment config by removing the ACR-AMR mapping patch.
	if err := testutils.PatchDeploymentConfig(map[string]interface{}{
		"acr_amr_mapping": nil,
	}); err != nil {
		ts.T().Logf("failed to restore ACR-AMR mapping in deployment config: %v", err)
	}
	if err := testutils.RestartServer(); err != nil {
		ts.T().Logf("failed to restart server after config restore: %v", err)
	}
}

// initiateAuthorizeWithAcrValues sends a GET /oauth2/authorize request with acr_values.
func (ts *AcrValuesAuthzTestSuite) initiateAuthorizeWithAcrValues(acrValues string) *http.Response {
	params := url.Values{}
	params.Set("client_id", acrClientID)
	params.Set("redirect_uri", acrRedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", "acr_test_state")
	if acrValues != "" {
		params.Set("acr_values", acrValues)
	}

	req, err := http.NewRequest(http.MethodGet,
		testutils.TestServerURL+"/oauth2/authorize?"+params.Encode(), nil)
	ts.Require().NoError(err)

	resp, err := testutils.GetNoRedirectHTTPClient().Do(req)
	ts.Require().NoError(err)
	return resp
}

// TestAcrValues_WithMatchingValues verifies that an acr_values param whose values are all
// present in the app's default_acr_values list results in a valid authorization redirect.
func (ts *AcrValuesAuthzTestSuite) TestAcrValues_WithMatchingValues() {
	resp := ts.initiateAuthorizeWithAcrValues("mosip:idp:acr:password")
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	ts.Assert().NotEmpty(location, "expected redirect to login page")

	authID, flowID, err := testutils.ExtractAuthData(location)
	ts.Assert().NoError(err, "expected valid flow redirect for matching acr_values")
	ts.Assert().NotEmpty(authID)
	ts.Assert().NotEmpty(flowID)
}

// TestAcrValues_WithNoDefaults_PassThrough verifies that when no acr_values are requested
// the authorization request succeeds for an app with default_acr_values configured.
func (ts *AcrValuesAuthzTestSuite) TestAcrValues_WithNoDefaults_PassThrough() {
	resp := ts.initiateAuthorizeWithAcrValues("")
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	ts.Assert().NotEmpty(location)

	authID, flowID, err := testutils.ExtractAuthData(location)
	ts.Assert().NoError(err, "expected valid flow redirect")
	ts.Assert().NotEmpty(authID)
	ts.Assert().NotEmpty(flowID)
}

// TestAcrValues_WithNoneInDefaults_FallsBackToDefaults verifies that when none of the
// requested acr_values match the app's default_acr_values list, the authorization
// request still succeeds (falls back to all defaults).
func (ts *AcrValuesAuthzTestSuite) TestAcrValues_WithNoneInDefaults_FallsBackToDefaults() {
	// "mosip:idp:acr:biometrics" is not in the app's default_acr_values list.
	resp := ts.initiateAuthorizeWithAcrValues("mosip:idp:acr:biometrics")
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	ts.Assert().NotEmpty(location)

	authID, flowID, err := testutils.ExtractAuthData(location)
	ts.Assert().NoError(err, "expected valid flow redirect even when requested ACR not in defaults")
	ts.Assert().NotEmpty(authID)
	ts.Assert().NotEmpty(flowID)
}

// TestAcrValues_PartialMatchWithDefaults verifies that when only some of the requested
// acr_values are in the app's default_acr_values list, the authorization succeeds.
func (ts *AcrValuesAuthzTestSuite) TestAcrValues_PartialMatchWithDefaults() {
	// "mosip:idp:acr:password" is in the list; "mosip:idp:acr:biometrics" is not.
	resp := ts.initiateAuthorizeWithAcrValues("mosip:idp:acr:password mosip:idp:acr:biometrics")
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	ts.Assert().NotEmpty(location)

	authID, flowID, err := testutils.ExtractAuthData(location)
	ts.Assert().NoError(err, "expected valid flow redirect for partial ACR match")
	ts.Assert().NotEmpty(authID)
	ts.Assert().NotEmpty(flowID)
}
