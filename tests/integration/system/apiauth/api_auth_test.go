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

package apiauth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const testServerURL = testutils.TestServerURL

type errorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// APIAuthTestSuite validates both authentication and authorization behavior for protected APIs.
type APIAuthTestSuite struct {
	suite.Suite
	adminClient        *http.Client
	plainClient        *http.Client
	invalidTokenClient *http.Client
	userClient         *http.Client
	ouID               string
	userSchemaID       string
	regularUserID      string
}

func TestAPIAuthTestSuite(t *testing.T) {
	suite.Run(t, new(APIAuthTestSuite))
}

func (suite *APIAuthTestSuite) SetupSuite() {
	suite.adminClient = testutils.GetHTTPClient()
	suite.plainClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	suite.invalidTokenClient = testutils.GetHTTPClientWithToken("invalid-token")

	ouID, err := testutils.CreateOrganizationUnit(testutils.OrganizationUnit{
		Handle: fmt.Sprintf("api-auth-ou-%d", time.Now().UnixNano()),
		Name:   "API Auth Test OU",
	})
	suite.Require().NoError(err)
	suite.ouID = ouID

	userSchema := testutils.UserSchema{
		Name:                  fmt.Sprintf("api-auth-user-%d", time.Now().UnixNano()),
		OrganizationUnitId:    suite.ouID,
		AllowSelfRegistration: true,
		Schema: map[string]interface{}{
			"username":  map[string]interface{}{"type": "string"},
			"password":  map[string]interface{}{"type": "string", "credential": true},
			"email":     map[string]interface{}{"type": "string"},
			"firstName": map[string]interface{}{"type": "string"},
			"lastName":  map[string]interface{}{"type": "string"},
		},
	}

	userSchemaID, err := testutils.CreateUserType(userSchema)
	suite.Require().NoError(err)
	suite.userSchemaID = userSchemaID

	username := fmt.Sprintf("authuser_%d", time.Now().UnixNano())
	password := "ApiAuthTest123!"
	userAttrs := map[string]interface{}{
		"username":  username,
		"password":  password,
		"email":     fmt.Sprintf("%s@example.com", username),
		"firstName": "API",
		"lastName":  "Auth",
	}
	attrBytes, err := json.Marshal(userAttrs)
	suite.Require().NoError(err)

	userID, err := testutils.CreateUser(testutils.User{
		OrganizationUnit: suite.ouID,
		Type:             userSchema.Name,
		Attributes:       attrBytes,
	})
	suite.Require().NoError(err)
	suite.regularUserID = userID

	userClient, err := testutils.GetHTTPClientForUser(username, password)
	suite.Require().NoError(err)
	suite.userClient = userClient
}

func (suite *APIAuthTestSuite) TearDownSuite() {
	if suite.regularUserID != "" {
		if err := testutils.DeleteUser(suite.regularUserID); err != nil {
			suite.T().Logf("Failed to delete regular user: %v", err)
		}
	}

	if suite.userSchemaID != "" {
		if err := testutils.DeleteUserType(suite.userSchemaID); err != nil {
			suite.T().Logf("Failed to delete user schema: %v", err)
		}
	}

	if suite.ouID != "" {
		if err := testutils.DeleteOrganizationUnit(suite.ouID); err != nil {
			suite.T().Logf("Failed to delete organization unit: %v", err)
		}
	}
}

// Authentication: valid system token succeeds.
func (suite *APIAuthTestSuite) TestSystemTokenAuthorized() {
	req, err := http.NewRequest(http.MethodGet, suite.protectedResourceURL(), nil)
	suite.Require().NoError(err)

	resp, err := suite.adminClient.Do(req)
	suite.Require().NoError(err)
	defer closeBodyQuietly(suite.T(), resp.Body)

	suite.Equal(http.StatusOK, resp.StatusCode)

	var ou testutils.OrganizationUnit
	suite.Require().NoError(json.NewDecoder(resp.Body).Decode(&ou))
	suite.Equal(suite.ouID, ou.ID)
}

// Authentication: missing token rejected.
func (suite *APIAuthTestSuite) TestMissingTokenIsUnauthorized() {
	req, err := http.NewRequest(http.MethodGet, suite.protectedResourceURL(), nil)
	suite.Require().NoError(err)

	resp, err := suite.plainClient.Do(req)
	suite.Require().NoError(err)
	defer closeBodyQuietly(suite.T(), resp.Body)

	suite.assertSecurityError(resp, http.StatusUnauthorized, "unauthorized",
		"Authentication is required to access this resource")
	suite.Equal("Bearer", resp.Header.Get("WWW-Authenticate"))
}

// Authentication: malformed token rejected.
func (suite *APIAuthTestSuite) TestInvalidTokenIsUnauthorized() {
	req, err := http.NewRequest(http.MethodGet, suite.protectedResourceURL(), nil)
	suite.Require().NoError(err)

	resp, err := suite.invalidTokenClient.Do(req)
	suite.Require().NoError(err)
	defer closeBodyQuietly(suite.T(), resp.Body)

	suite.assertSecurityError(resp, http.StatusUnauthorized, "unauthorized",
		"Authentication is required to access this resource")
	suite.Equal("Bearer", resp.Header.Get("WWW-Authenticate"))
}

// Authorization: non-system token is forbidden.
func (suite *APIAuthTestSuite) TestNonSystemScopeIsForbidden() {
	suite.Require().NotNil(suite.userClient, "Non-system user client must be available for the test")

	req, err := http.NewRequest(http.MethodGet, suite.protectedResourceURL(), nil)
	suite.Require().NoError(err)

	resp, err := suite.userClient.Do(req)
	suite.Require().NoError(err)
	defer closeBodyQuietly(suite.T(), resp.Body)

	suite.Equal(http.StatusForbidden, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errResp errorResponse
	suite.Require().NoError(json.Unmarshal(bodyBytes, &errResp))

	suite.Equal("forbidden", errResp.Error)
	suite.Equal("You do not have sufficient permissions to access this resource", errResp.ErrorDescription)
}

func (suite *APIAuthTestSuite) assertSecurityError(resp *http.Response, expectedStatus int, expectedCode,
	expectedDescription string) {
	suite.Equal(expectedStatus, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errResp errorResponse
	suite.Require().NoError(json.Unmarshal(bodyBytes, &errResp))

	suite.Equal(expectedCode, errResp.Error)
	suite.Equal(expectedDescription, errResp.ErrorDescription)
}

func (suite *APIAuthTestSuite) protectedResourceURL() string {
	return fmt.Sprintf("%s/organization-units/%s", testServerURL, suite.ouID)
}

func closeBodyQuietly(t *testing.T, body io.ReadCloser) {
	if body != nil {
		if err := body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}
}
