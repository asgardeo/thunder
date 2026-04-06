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

package application

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

type AcrValuesAPITestSuite struct {
	suite.Suite
	previousAcrAmrMapping interface{}
}

func TestAcrValuesAPITestSuite(t *testing.T) {
	suite.Run(t, new(AcrValuesAPITestSuite))
}

func (ts *AcrValuesAPITestSuite) SetupSuite() {
	prev, err := testutils.ReadDeploymentConfigKey("acr_amr_mapping")
	ts.Require().NoError(err, "failed to read existing acr_amr_mapping from deployment config")
	ts.previousAcrAmrMapping = prev

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
}

func (ts *AcrValuesAPITestSuite) TearDownSuite() {
	if err := testutils.PatchDeploymentConfig(map[string]interface{}{
		"acr_amr_mapping": ts.previousAcrAmrMapping,
	}); err != nil {
		ts.T().Logf("failed to restore ACR-AMR mapping in deployment config: %v", err)
	}
	if err := testutils.RestartServer(); err != nil {
		ts.T().Logf("failed to restart server after config restore: %v", err)
	}
}

// TestCreateApplicationWithValidDefaultAcrValues verifies that an application can be created
// with a valid default_acr_values list and that the values are returned by GET.
func (ts *AcrValuesAPITestSuite) TestCreateApplicationWithValidDefaultAcrValues() {
	app := Application{
		Name:                      "ACR Values Test App",
		Description:               "Application for testing default_acr_values",
		IsRegistrationFlowEnabled: false,
		AuthFlowID:                defaultAuthFlowID,
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "acr_values_test_client",
					ClientSecret:            "acr_values_test_secret",
					RedirectURIs:            []string{"http://localhost/acr_values_test/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					DefaultAcrValues: []string{
						"mosip:idp:acr:password",
						"mosip:idp:acr:generated-code",
					},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err, "failed to create application with default_acr_values")
	defer func() {
		if err := deleteApplication(appID); err != nil {
			ts.T().Logf("Failed to delete test application: %v", err)
		}
	}()

	retrieved, err := getApplicationByID(appID)
	ts.Require().NoError(err, "failed to retrieve application")
	ts.Require().NotNil(retrieved.InboundAuthConfig)
	ts.Require().Len(retrieved.InboundAuthConfig, 1)
	ts.Require().NotNil(retrieved.InboundAuthConfig[0].OAuthAppConfig)

	ts.Assert().ElementsMatch(
		[]string{"mosip:idp:acr:password", "mosip:idp:acr:generated-code"},
		retrieved.InboundAuthConfig[0].OAuthAppConfig.DefaultAcrValues,
		"default_acr_values should be persisted and returned",
	)
}

// TestCreateApplicationWithInvalidDefaultAcrValues verifies that creating an application
// with an unrecognised ACR value is rejected with 400.
func (ts *AcrValuesAPITestSuite) TestCreateApplicationWithInvalidDefaultAcrValues() {
	app := Application{
		Name:                      "Invalid ACR App",
		Description:               "Application with invalid default_acr_values",
		IsRegistrationFlowEnabled: false,
		AuthFlowID:                defaultAuthFlowID,
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "invalid_acr_client",
					ClientSecret:            "invalid_acr_secret",
					RedirectURIs:            []string{"http://localhost/invalid_acr/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					DefaultAcrValues:        []string{"urn:unknown:acr:value"},
				},
			},
		},
	}

	appJSON, err := json.Marshal(app)
	ts.Require().NoError(err)

	client := testutils.GetHTTPClient()
	req, err := http.NewRequest(http.MethodPost, testServerURL+"/applications", bytes.NewReader(appJSON))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusBadRequest, resp.StatusCode,
		"unrecognised ACR value should cause 400 Bad Request")
}

// TestUpdateApplicationDefaultAcrValues verifies that default_acr_values can be added to an
// existing application via PUT and that the updated values are returned by GET.
func (ts *AcrValuesAPITestSuite) TestUpdateApplicationDefaultAcrValues() {
	app := Application{
		Name:                      "ACR Update Test App",
		Description:               "Application for testing default_acr_values update",
		IsRegistrationFlowEnabled: false,
		AuthFlowID:                defaultAuthFlowID,
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "acr_update_test_client",
					ClientSecret:            "acr_update_test_secret",
					RedirectURIs:            []string{"http://localhost/acr_update/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err, "failed to create test application")
	defer func() {
		if err := deleteApplication(appID); err != nil {
			ts.T().Logf("Failed to delete test application: %v", err)
		}
	}()

	// Update: add DefaultAcrValues
	updated := app
	updated.InboundAuthConfig = []InboundAuthConfig{
		{
			Type: "oauth2",
			OAuthAppConfig: &OAuthAppConfig{
				ClientID:                "acr_update_test_client",
				RedirectURIs:            []string{"http://localhost/acr_update/callback"},
				GrantTypes:              []string{"authorization_code"},
				ResponseTypes:           []string{"code"},
				TokenEndpointAuthMethod: "client_secret_basic",
				DefaultAcrValues:        []string{"mosip:idp:acr:biometrics"},
			},
		},
	}

	err = updateApplication(appID, updated)
	ts.Require().NoError(err, "failed to update application")

	retrieved, err := getApplicationByID(appID)
	ts.Require().NoError(err, "failed to retrieve updated application")
	ts.Require().NotNil(retrieved.InboundAuthConfig[0].OAuthAppConfig)

	ts.Assert().Equal(
		[]string{"mosip:idp:acr:biometrics"},
		retrieved.InboundAuthConfig[0].OAuthAppConfig.DefaultAcrValues,
		"default_acr_values should reflect the updated value",
	)
}

// TestClearApplicationDefaultAcrValues verifies that default_acr_values can be removed from
// an application by updating without the field.
func (ts *AcrValuesAPITestSuite) TestClearApplicationDefaultAcrValues() {
	app := Application{
		Name:                      "ACR Clear Test App",
		Description:               "Application for testing clearing of default_acr_values",
		IsRegistrationFlowEnabled: false,
		AuthFlowID:                defaultAuthFlowID,
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "acr_clear_test_client",
					ClientSecret:            "acr_clear_test_secret",
					RedirectURIs:            []string{"http://localhost/acr_clear/callback"},
					GrantTypes:              []string{"authorization_code"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					DefaultAcrValues:        []string{"mosip:idp:acr:password"},
				},
			},
		},
	}

	appID, err := createApplication(app)
	ts.Require().NoError(err, "failed to create test application")
	defer func() {
		if err := deleteApplication(appID); err != nil {
			ts.T().Logf("Failed to delete test application: %v", err)
		}
	}()

	// Update: omit DefaultAcrValues to clear it
	cleared := app
	cleared.InboundAuthConfig = []InboundAuthConfig{
		{
			Type: "oauth2",
			OAuthAppConfig: &OAuthAppConfig{
				ClientID:                "acr_clear_test_client",
				RedirectURIs:            []string{"http://localhost/acr_clear/callback"},
				GrantTypes:              []string{"authorization_code"},
				ResponseTypes:           []string{"code"},
				TokenEndpointAuthMethod: "client_secret_basic",
			},
		},
	}

	err = updateApplication(appID, cleared)
	ts.Require().NoError(err, "failed to update application to clear default_acr_values")

	retrieved, err := getApplicationByID(appID)
	ts.Require().NoError(err, "failed to retrieve application after clearing")
	ts.Require().NotNil(retrieved.InboundAuthConfig[0].OAuthAppConfig)

	ts.Assert().Empty(retrieved.InboundAuthConfig[0].OAuthAppConfig.DefaultAcrValues,
		"default_acr_values should be empty after clearing")
}
