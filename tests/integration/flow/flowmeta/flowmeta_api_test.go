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

package flowmeta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	testServerURL    = "https://localhost:8095"
	flowMetaEndpoint = "/flow/meta"
)

var (
	testOU = testutils.OrganizationUnit{
		Handle:          "flowmeta-test-ou",
		Name:            "FlowMeta Test Organization Unit",
		Description:     "Organization unit for flow metadata integration testing",
		Parent:          nil,
		LogoURL:         "https://example.com/ou-logo.png",
		TosURI:          "https://example.com/tos",
		PolicyURI:       "https://example.com/privacy",
		CookiePolicyURI: "https://example.com/cookie-policy",
	}

	testApp = testutils.Application{
		Name:                      "FlowMeta Test Application",
		Description:               "Application for flow metadata integration testing",
		IsRegistrationFlowEnabled: true,
		ClientID:                  "flowmeta_test_client",
		ClientSecret:              "flowmeta_test_secret",
		RedirectURIs:              []string{"https://localhost:8095/callback"},
	}

	// localeTestApp is used for locale-resolution tests; it has name#fr registered.
	localeTestApp = testutils.Application{
		Name:                      "Base App",
		Description:               "Application with localised variants for flow metadata locale testing",
		IsRegistrationFlowEnabled: true,
		ClientID:                  "flowmeta_locale_client",
		ClientSecret:              "flowmeta_locale_secret",
		RedirectURIs:              []string{"https://localhost:8095/callback"},
	}
)

type FlowMetaAPITestSuite struct {
	suite.Suite
	appID       string
	ouID        string
	localeAppID string // app with name#fr registered, used for locale-resolution tests
}

func TestFlowMetaAPITestSuite(t *testing.T) {
	suite.Run(t, new(FlowMetaAPITestSuite))
}

func (suite *FlowMetaAPITestSuite) SetupSuite() {
	// Create OU
	ouID, err := testutils.CreateOrganizationUnit(testOU)
	suite.Require().NoError(err, "Failed to create OU during setup")
	suite.ouID = ouID

	// Create Application
	testApp.OUID = suite.ouID
	appID, err := testutils.CreateApplication(testApp)
	suite.Require().NoError(err, "Failed to create application during setup")
	suite.appID = appID

	// Create localised test application with name#fr variant
	localeAppID, err := createLocalisedApplication(localeTestApp, map[string]string{
		"name#fr": "App FR",
	})
	suite.Require().NoError(err, "Failed to create localised application during setup")
	suite.localeAppID = localeAppID
}

func (suite *FlowMetaAPITestSuite) TearDownSuite() {
	if suite.appID != "" {
		err := testutils.DeleteApplication(suite.appID)
		if err != nil {
			suite.T().Logf("Failed to delete application during teardown: %v", err)
		}
	}

	if suite.localeAppID != "" {
		err := testutils.DeleteApplication(suite.localeAppID)
		if err != nil {
			suite.T().Logf("Failed to delete localised application during teardown: %v", err)
		}
	}

	if suite.ouID != "" {
		err := testutils.DeleteOrganizationUnit(suite.ouID)
		if err != nil {
			suite.T().Logf("Failed to delete OU during teardown: %v", err)
		}
	}
}

// createLocalisedApplication creates an application with additional localised fields (e.g., name#fr).
// extraFields is a map of raw JSON keys (like "name#fr") to their values.
func createLocalisedApplication(app testutils.Application, extraFields map[string]string) (string, error) {
	redirectURIs := app.RedirectURIs
	if len(redirectURIs) == 0 {
		redirectURIs = []string{"http://localhost:8080/callback"}
	}

	appData := map[string]interface{}{
		"name":                      app.Name,
		"description":               app.Description,
		"isRegistrationFlowEnabled": app.IsRegistrationFlowEnabled,
		"certificate": map[string]interface{}{
			"type":  "NONE",
			"value": "",
		},
		"inboundAuthConfig": []interface{}{
			map[string]interface{}{
				"type": "oauth2",
				"config": map[string]interface{}{
					"clientId":     app.ClientID,
					"clientSecret": app.ClientSecret,
					"redirectUris": redirectURIs,
				},
			},
		},
	}
	for k, v := range extraFields {
		appData[k] = v
	}

	appJSON, err := json.Marshal(appData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal application: %w", err)
	}

	client := testutils.GetHTTPClient()
	req, err := http.NewRequest("POST", testServerURL+"/applications", bytes.NewReader(appJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("expected status 201, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	appID, _ := result["id"].(string)
	if appID == "" {
		return "", fmt.Errorf("response does not contain id")
	}
	return appID, nil
}

// TestGetFlowMetadataWithAppType tests GET /flow/meta?type=APP&id={appID}
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithAppType() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=APP&id=%s", testServerURL, flowMetaEndpoint, suite.appID), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var metadata FlowMetadataResponse
	err = json.Unmarshal(body, &metadata)
	suite.Require().NoError(err)

	// Verify application metadata
	suite.NotNil(metadata.Application)
	suite.Equal(suite.appID, metadata.Application.ID)
	suite.Equal(testApp.Name, metadata.Application.Name)
	suite.Equal(testApp.Description, metadata.Application.Description)
	suite.True(metadata.IsRegistrationFlowEnabled)

	// Verify design metadata is present (even if empty)
	suite.NotNil(metadata.Design.Theme)
	suite.NotNil(metadata.Design.Layout)

	// Verify i18n metadata is present
	suite.NotNil(metadata.I18n.Languages)
	suite.NotEmpty(metadata.I18n.Languages)
}

// TestGetFlowMetadataWithOUType tests GET /flow/meta?type=OU&id={ouID}
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithOUType() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=OU&id=%s", testServerURL, flowMetaEndpoint, suite.ouID), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var metadata FlowMetadataResponse
	err = json.Unmarshal(body, &metadata)
	suite.Require().NoError(err)

	// Verify OU metadata
	suite.NotNil(metadata.OU)
	suite.Equal(suite.ouID, metadata.OU.ID)
	suite.Equal(testOU.Name, metadata.OU.Name)
	suite.Equal(testOU.Handle, metadata.OU.Handle)
	suite.Equal(testOU.Description, metadata.OU.Description)
	suite.Equal(testOU.LogoURL, metadata.OU.LogoURL)
	suite.Equal(testOU.TosURI, metadata.OU.TosURI)
	suite.Equal(testOU.PolicyURI, metadata.OU.PolicyURI)
	suite.Equal(testOU.CookiePolicyURI, metadata.OU.CookiePolicyURI)

	// Application should be nil for OU type
	suite.Nil(metadata.Application)
	suite.False(metadata.IsRegistrationFlowEnabled)

	// Verify design metadata is present (even if empty)
	suite.NotNil(metadata.Design.Theme)
	suite.NotNil(metadata.Design.Layout)

	// Verify i18n metadata is present
	suite.NotNil(metadata.I18n.Languages)
	suite.NotEmpty(metadata.I18n.Languages)
}

// TestGetFlowMetadataWithLanguageParam tests GET /flow/meta with language parameter
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithLanguageParam() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=APP&id=%s&language=en", testServerURL, flowMetaEndpoint, suite.appID), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var metadata FlowMetadataResponse
	err = json.Unmarshal(body, &metadata)
	suite.Require().NoError(err)

	// Verify application metadata is present
	suite.NotNil(metadata.Application)
	suite.Equal(suite.appID, metadata.Application.ID)
}

// TestGetFlowMetadataWithNamespaceParam tests GET /flow/meta with namespace parameter
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithNamespaceParam() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=APP&id=%s&language=en&namespace=common",
			testServerURL, flowMetaEndpoint, suite.appID), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var metadata FlowMetadataResponse
	err = json.Unmarshal(body, &metadata)
	suite.Require().NoError(err)

	suite.NotNil(metadata.Application)
}

// TestGetFlowMetadataMissingType tests GET /flow/meta without type parameter
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataMissingType() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?id=%s", testServerURL, flowMetaEndpoint, suite.appID), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("FM-1004", errorResp.Code)
	suite.Equal("Missing required parameter", errorResp.Message)
}

// TestGetFlowMetadataMissingID tests GET /flow/meta without id parameter
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataMissingID() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=APP", testServerURL, flowMetaEndpoint), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("FM-1005", errorResp.Code)
	suite.Equal("Missing required parameter", errorResp.Message)
}

// TestGetFlowMetadataMissingBothParams tests GET /flow/meta without any parameters.
// When neither type nor id is provided, the endpoint returns i18n metadata only (system flow).
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataMissingBothParams() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s", testServerURL, flowMetaEndpoint), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var metadata FlowMetadataResponse
	err = json.Unmarshal(body, &metadata)
	suite.Require().NoError(err)

	// No type/id means system flow: only i18n metadata is returned
	suite.Nil(metadata.Application)
	suite.Nil(metadata.OU)
	suite.NotNil(metadata.I18n.Languages)
	suite.NotEmpty(metadata.I18n.Languages)
}

// TestGetFlowMetadataInvalidType tests GET /flow/meta with invalid type parameter
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataInvalidType() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=INVALID&id=%s", testServerURL, flowMetaEndpoint, suite.appID), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("FM-1001", errorResp.Code)
	suite.Equal("Invalid request", errorResp.Message)
}

// TestGetFlowMetadataAppNotFound tests GET /flow/meta with non-existent application ID
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataAppNotFound() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=APP&id=non-existent-app-id", testServerURL, flowMetaEndpoint), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("FM-1002", errorResp.Code)
	suite.Equal("Resource not found", errorResp.Message)
}

// TestGetFlowMetadataOUNotFound tests GET /flow/meta with non-existent OU ID
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataOUNotFound() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=OU&id=non-existent-ou-id", testServerURL, flowMetaEndpoint), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("FM-1003", errorResp.Code)
	suite.Equal("Resource not found", errorResp.Message)
}

// TestGetFlowMetadataDesignDefaults tests that design metadata returns defaults when no design is configured
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataDesignDefaults() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=OU&id=%s", testServerURL, flowMetaEndpoint, suite.ouID), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var metadata FlowMetadataResponse
	err = json.Unmarshal(body, &metadata)
	suite.Require().NoError(err)

	// Design should default to empty JSON objects when not configured
	suite.NotNil(metadata.Design.Theme)
	suite.NotNil(metadata.Design.Layout)

	// Verify they are valid JSON
	var theme map[string]interface{}
	err = json.Unmarshal(metadata.Design.Theme, &theme)
	suite.NoError(err)

	var layout map[string]interface{}
	err = json.Unmarshal(metadata.Design.Layout, &layout)
	suite.NoError(err)
}

// TestGetFlowMetadataI18nDefaults tests that i18n metadata returns defaults
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataI18nDefaults() {
	client := testutils.GetHTTPClient()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=APP&id=%s", testServerURL, flowMetaEndpoint, suite.appID), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var metadata FlowMetadataResponse
	err = json.Unmarshal(body, &metadata)
	suite.Require().NoError(err)

	// i18n should have at least "en-US" in the languages list
	suite.NotNil(metadata.I18n.Languages)
	suite.Contains(metadata.I18n.Languages, "en-US")

	// Translations map should not be nil
	suite.NotNil(metadata.I18n.Translations)
}

// TestGetFlowMetadataCaseSensitiveType tests that type parameter is case-sensitive
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataCaseSensitiveType() {
	client := testutils.GetHTTPClient()

	// Lowercase "app" should be invalid
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s%s?type=app&id=%s", testServerURL, flowMetaEndpoint, suite.appID), nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("FM-1001", errorResp.Code)
}

// ─── Localised Client Metadata Tests (AC-15–AC-24) ────────────────────────────

// initiateFlowWithUILocale starts an authorization flow for the localeApp with the given ui_locales
// value and returns the flowId from the redirect Location header.
func (suite *FlowMetaAPITestSuite) initiateFlowWithUILocale(uiLocales string) string {
	resp, err := testutils.InitiateAuthorizationFlowWithUILocale(
		localeTestApp.ClientID,
		localeTestApp.RedirectURIs[0],
		"code",
		"openid",
		"test-state",
		uiLocales,
	)
	suite.Require().NoError(err, "Failed to initiate authorization flow with ui_locales=%q", uiLocales)
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	suite.Require().NotEmpty(location, "Expected a redirect Location header from authorize endpoint")

	_, flowID, err := testutils.ExtractAuthData(location)
	suite.Require().NoError(err, "Failed to extract flowId from redirect: %s", location)
	return flowID
}

// getFlowMetaForLocaleApp calls /flow/meta for the locale test app with an optional flowId.
func (suite *FlowMetaAPITestSuite) getFlowMetaForLocaleApp(flowID string) FlowMetadataResponse {
	url := fmt.Sprintf("%s%s?type=APP&id=%s", testServerURL, flowMetaEndpoint, suite.localeAppID)
	if flowID != "" {
		url += "&flowId=" + flowID
	}

	client := testutils.GetHTTPClient()
	req, err := http.NewRequest("GET", url, nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var metadata FlowMetadataResponse
	suite.Require().NoError(json.Unmarshal(body, &metadata))
	return metadata
}

// TestGetFlowMetadataWithLocalisedApp_BaseValues_NoFlowID verifies that without a flowId the base
// (untagged) name is returned and uiLocale is empty (AC-18, AC-24).
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithLocalisedApp_BaseValues_NoFlowID() {
	metadata := suite.getFlowMetaForLocaleApp("")

	suite.NotNil(metadata.Application)
	suite.Equal("Base App", metadata.Application.Name)
	suite.Empty(metadata.UILocale)
}

// TestGetFlowMetadataWithLocalisedApp_ExactLocaleMatch verifies that ui_locales=fr resolves to the
// name#fr variant and uiLocale is returned in the response (AC-15, AC-22).
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithLocalisedApp_ExactLocaleMatch() {
	flowID := suite.initiateFlowWithUILocale("fr")
	metadata := suite.getFlowMetaForLocaleApp(flowID)

	suite.NotNil(metadata.Application)
	suite.Equal("App FR", metadata.Application.Name)
	suite.Equal("fr", metadata.UILocale)
}

// TestGetFlowMetadataWithLocalisedApp_SubtagFallback verifies that ui_locales=fr-CA falls back to
// the registered name#fr variant (AC-16).
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithLocalisedApp_SubtagFallback() {
	flowID := suite.initiateFlowWithUILocale("fr-CA")
	metadata := suite.getFlowMetaForLocaleApp(flowID)

	suite.NotNil(metadata.Application)
	suite.Equal("App FR", metadata.Application.Name, "fr-CA should fall back to the fr variant")
}

// TestGetFlowMetadataWithLocalisedApp_BaseFallback_NoMatchingVariant verifies that ui_locales=de
// falls back to the base name when no name#de is registered (AC-17).
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithLocalisedApp_BaseFallback_NoMatchingVariant() {
	flowID := suite.initiateFlowWithUILocale("de")
	metadata := suite.getFlowMetaForLocaleApp(flowID)

	suite.NotNil(metadata.Application)
	suite.Equal("Base App", metadata.Application.Name, "should fall back to base name when no de variant exists")
	suite.Equal("de", metadata.UILocale)
}

// TestGetFlowMetadataWithLocalisedApp_ExpiredOrUnknownFlowID verifies that an unknown flowId
// degrades gracefully to base values without a 4xx/5xx (AC-19).
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithLocalisedApp_ExpiredOrUnknownFlowID() {
	metadata := suite.getFlowMetaForLocaleApp("00000000-0000-0000-0000-000000000000")

	suite.NotNil(metadata.Application)
	suite.Equal("Base App", metadata.Application.Name)
}

// TestGetFlowMetadataWithLocalisedApp_SpaceSeparatedUILocale_FirstMatchWins verifies that
// "de fr" (space-separated) picks the first matching variant — de has none, fr does (AC-21).
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithLocalisedApp_SpaceSeparatedUILocale_FirstMatchWins() {
	flowID := suite.initiateFlowWithUILocale("de fr")
	metadata := suite.getFlowMetaForLocaleApp(flowID)

	suite.NotNil(metadata.Application)
	suite.Equal("App FR", metadata.Application.Name, "should resolve to fr when de has no variant")
}

// TestGetFlowMetadataWithLocalisedApp_UILocaleReturnedInResponse verifies that the uiLocale field
// in the /flow/meta response matches the ui_locales stored in the flow context (AC-22).
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataWithLocalisedApp_UILocaleReturnedInResponse() {
	flowID := suite.initiateFlowWithUILocale("fr")
	metadata := suite.getFlowMetaForLocaleApp(flowID)

	suite.Equal("fr", metadata.UILocale)
}

// TestGetFlowMetadataBackwardsCompatibility_AppWithNoTaggedVariants verifies that querying
// /flow/meta with a ui_locales flow against a standard (non-localized) app returns base values
// without error (AC-23, AC-24).
func (suite *FlowMetaAPITestSuite) TestGetFlowMetadataBackwardsCompatibility_AppWithNoTaggedVariants() {
	// Initiate a flow for the standard (non-localized) testApp
	resp, err := testutils.InitiateAuthorizationFlowWithUILocale(
		testApp.ClientID,
		testApp.RedirectURIs[0],
		"code",
		"openid",
		"test-state",
		"fr",
	)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	suite.Require().NotEmpty(location)

	_, flowID, err := testutils.ExtractAuthData(location)
	suite.Require().NoError(err)

	// Query /flow/meta for the standard app with that flowId
	url := fmt.Sprintf("%s%s?type=APP&id=%s&flowId=%s", testServerURL, flowMetaEndpoint, suite.appID, flowID)
	client := testutils.GetHTTPClient()
	req, err := http.NewRequest("GET", url, nil)
	suite.Require().NoError(err)

	httpResp, err := client.Do(req)
	suite.Require().NoError(err)
	defer httpResp.Body.Close()

	suite.Equal(http.StatusOK, httpResp.StatusCode)

	body, _ := io.ReadAll(httpResp.Body)
	var metadata FlowMetadataResponse
	suite.Require().NoError(json.Unmarshal(body, &metadata))

	suite.NotNil(metadata.Application)
	suite.Equal(testApp.Name, metadata.Application.Name, "standard app should return its base name unchanged")
}
