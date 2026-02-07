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

package design

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	resolveBasePath = "/design/resolve"
)

type ResolveAPITestSuite struct {
	suite.Suite
	client *http.Client
}

func TestResolveAPITestSuite(t *testing.T) {
	suite.Run(t, new(ResolveAPITestSuite))
}

func (suite *ResolveAPITestSuite) SetupSuite() {
	// Create HTTP client that skips TLS verification for testing
	suite.client = testutils.GetHTTPClient()
}

// Helper function to resolve design configuration
func (suite *ResolveAPITestSuite) resolveDesign(resolveType, id string) (*DesignResolveResponse, int, error) {
	url := fmt.Sprintf("%s%s?type=%s&id=%s", testServerURL, resolveBasePath, resolveType, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return nil, resp.StatusCode, fmt.Errorf("expected status 200, got %d. Code: %s, Message: %s", resp.StatusCode, errResp.Code, errResp.Message)
		}
		return nil, resp.StatusCode, fmt.Errorf("expected status 200, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var resolveResponse DesignResolveResponse
	if err := json.Unmarshal(bodyBytes, &resolveResponse); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to parse response body: %w. Response: %s", err, string(bodyBytes))
	}

	return &resolveResponse, resp.StatusCode, nil
}

// Test Resolve Design - Missing Type Parameter
func (suite *ResolveAPITestSuite) TestResolveDesign_MissingType() {
	url := fmt.Sprintf("%s%s?id=00000000-0000-0000-0000-000000000000", testServerURL, resolveBasePath)

	req, err := http.NewRequest("GET", url, nil)
	suite.Require().NoError(err)

	resp, err := suite.client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errResp)
	suite.Require().NoError(err)
	suite.Equal("DSR-1001", errResp.Code)
}

// Test Resolve Design - Missing ID Parameter
func (suite *ResolveAPITestSuite) TestResolveDesign_MissingID() {
	url := fmt.Sprintf("%s%s?type=APP", testServerURL, resolveBasePath)

	req, err := http.NewRequest("GET", url, nil)
	suite.Require().NoError(err)

	resp, err := suite.client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errResp)
	suite.Require().NoError(err)
	suite.Equal("DSR-1002", errResp.Code)
}

// Test Resolve Design - Unsupported Type
func (suite *ResolveAPITestSuite) TestResolveDesign_UnsupportedType() {
	_, statusCode, err := suite.resolveDesign("OU", "00000000-0000-0000-0000-000000000000")

	suite.Error(err)
	suite.Equal(http.StatusBadRequest, statusCode)
	suite.Contains(err.Error(), "DSR-1003")
}

// Test Resolve Design - Application Not Found
func (suite *ResolveAPITestSuite) TestResolveDesign_ApplicationNotFound() {
	_, statusCode, err := suite.resolveDesign("APP", "00000000-0000-0000-0000-000000000000")

	suite.Error(err)
	suite.Equal(http.StatusNotFound, statusCode)
	suite.Contains(err.Error(), "DSR-1004")
}

// Test Resolve Design - Invalid ID Format
func (suite *ResolveAPITestSuite) TestResolveDesign_InvalidIDFormat() {
	url := fmt.Sprintf("%s%s?type=APP&id=invalid-uuid", testServerURL, resolveBasePath)

	req, err := http.NewRequest("GET", url, nil)
	suite.Require().NoError(err)

	resp, err := suite.client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	// Should return bad request for invalid UUID format
	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errResp)
	suite.Require().NoError(err)
	suite.Equal("DSR-1002", errResp.Code)
}

// Test Resolve Design - Success Case
// Note: This test requires an actual application with theme and layout configured
// For a complete test, you would need to:
// 1. Create a theme
// 2. Create a layout
// 3. Create an application with theme and layout
// 4. Call resolve with the application ID
// 5. Verify the merged design configuration
func (suite *ResolveAPITestSuite) TestResolveDesign_Success() {
	// This is a placeholder test that demonstrates the expected flow
	// In a real test environment, you would:
	//
	// 1. Create theme
	// theme := CreateThemeRequest{
	//     DisplayName: "Test Theme",
	//     Theme: themeConfig,
	// }
	// createdTheme, _ := createTheme(theme)
	//
	// 2. Create layout
	// layout := CreateLayoutRequest{
	//     DisplayName: "Test Layout",
	//     Layout: layoutConfig,
	// }
	// createdLayout, _ := createLayout(layout)
	//
	// 3. Create application with theme and layout
	// app := CreateApplicationRequest{
	//     Name: "Test App",
	//     ThemeID: createdTheme.ID,
	//     LayoutID: createdLayout.ID,
	// }
	// createdApp, _ := createApplication(app)
	//
	// 4. Resolve design for the application
	// resolvedDesign, statusCode, err := suite.resolveDesign("APP", createdApp.ID)
	// suite.NoError(err)
	// suite.Equal(http.StatusOK, statusCode)
	// suite.NotNil(resolvedDesign)
	// suite.NotEmpty(resolvedDesign.Theme)
	// suite.NotEmpty(resolvedDesign.Layout)
	//
	// 5. Verify that both theme and layout are returned separately
	// Verify theme contains expected theme configuration
	// Verify layout contains expected layout configuration
	//
	// 6. Cleanup
	// deleteApplication(createdApp.ID)
	// deleteTheme(createdTheme.ID)
	// deleteLayout(createdLayout.ID)

	suite.T().Skip("Requires application creation with theme and layout - implement when application API is available")
}

// Test Resolve Design - Application Without Theme and Layout
// This test verifies behavior when an application exists but has no design configuration
func (suite *ResolveAPITestSuite) TestResolveDesign_ApplicationWithoutDesign() {
	// This is a placeholder test
	// In a real test environment:
	//
	// 1. Create application without theme and layout
	// app := CreateApplicationRequest{
	//     Name: "Test App Without Design",
	// }
	// createdApp, _ := createApplication(app)
	//
	// 2. Try to resolve design
	// _, statusCode, err := suite.resolveDesign("APP", createdApp.ID)
	// suite.Error(err)
	// suite.Equal(http.StatusNotFound, statusCode)
	// suite.Contains(err.Error(), "DSR-1005") // No design configuration found
	//
	// 3. Cleanup
	// deleteApplication(createdApp.ID)

	suite.T().Skip("Requires application creation without design - implement when application API is available")
}

// Test Resolve Design - Only Theme Configured
func (suite *ResolveAPITestSuite) TestResolveDesign_OnlyTheme() {
	// This is a placeholder test
	// In a real test environment:
	//
	// 1. Create theme
	// 2. Create application with only theme (no layout)
	// 3. Resolve design
	// 4. Verify Theme field contains theme data and Layout field is empty
	// 5. Cleanup

	suite.T().Skip("Requires application creation with only theme - implement when application API is available")
}

// Test Resolve Design - Only Layout Configured
func (suite *ResolveAPITestSuite) TestResolveDesign_OnlyLayout() {
	// This is a placeholder test
	// In a real test environment:
	//
	// 1. Create layout
	// 2. Create application with only layout (no theme)
	// 3. Resolve design
	// 4. Verify Layout field contains layout data and Theme field is empty
	// 5. Cleanup

	suite.T().Skip("Requires application creation with only layout - implement when application API is available")
}
