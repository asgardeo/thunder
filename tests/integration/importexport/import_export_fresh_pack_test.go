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
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package importexport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

type exportRequest struct {
	Applications      []string `json:"applications,omitempty"`
	OrganizationUnits []string `json:"organizationUnits,omitempty"`
	Flows             []string `json:"flows,omitempty"`
	Themes            []string `json:"themes,omitempty"`
	Layouts           []string `json:"layouts,omitempty"`
}

type importRequest struct {
	Content string                 `json:"content"`
	DryRun  bool                   `json:"dryRun,omitempty"`
	Options importOptions          `json:"options"`
	Vars    map[string]interface{} `json:"variables,omitempty"`
}

type importOptions struct {
	Upsert          bool   `json:"upsert"`
	ContinueOnError bool   `json:"continueOnError"`
	Target          string `json:"target"`
}

type importResponse struct {
	Summary importSummary `json:"summary"`
	Results []importItem  `json:"results"`
}

type importSummary struct {
	TotalDocuments int `json:"totalDocuments"`
	Imported       int `json:"imported"`
	Failed         int `json:"failed"`
}

type importItem struct {
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId,omitempty"`
	ResourceName string `json:"resourceName,omitempty"`
	Operation    string `json:"operation,omitempty"`
	Status       string `json:"status"`
	Code         string `json:"code,omitempty"`
	Message      string `json:"message,omitempty"`
}

type createThemeRequest struct {
	Handle      string                 `json:"handle"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	Theme       map[string]interface{} `json:"theme"`
}

type createLayoutRequest struct {
	Handle      string                 `json:"handle"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	Layout      map[string]interface{} `json:"layout"`
}

type exportFile struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

type ImportExportFreshPackSuite struct {
	suite.Suite
}

func TestImportExportFreshPackSuite(t *testing.T) {
	suite.Run(t, new(ImportExportFreshPackSuite))
}

func (suite *ImportExportFreshPackSuite) SetupSuite() {
	if os.Getenv("THUNDER_EXTRACTED_HOME") == "" {
		suite.T().Skip("requires integration harness context (THUNDER_EXTRACTED_HOME is not set)")
	}
}

func (suite *ImportExportFreshPackSuite) TearDownSuite() {
	if os.Getenv("THUNDER_EXTRACTED_HOME") == "" {
		return
	}
	if testutils.GetDBType() != "sqlite" {
		return
	}

	err := suite.resetToFreshPack()
	if err != nil {
		suite.T().Logf("failed to restore fresh pack in teardown: %v", err)
	}
}

func (suite *ImportExportFreshPackSuite) TestExportImportAcrossFreshPack() {
	if testutils.GetDBType() != "sqlite" {
		suite.T().Skip("fresh-pack reset integration test currently supports sqlite only")
	}

	now := time.Now().UnixNano()
	handleSuffix := fmt.Sprintf("%d", now)

	flowID, err := testutils.CreateFlow(testutils.Flow{
		Name:     "Import Export Auth Flow " + handleSuffix,
		FlowType: "AUTHENTICATION",
		Handle:   "import-export-auth-flow-" + handleSuffix,
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
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
	})
	suite.Require().NoError(err)

	themeID, err := suite.createTheme(createThemeRequest{
		Handle:      "import-export-theme-" + handleSuffix,
		DisplayName: "Import Export Theme " + handleSuffix,
		Description: "Theme for export/import fresh-pack test",
		Theme: map[string]interface{}{
			"palette": map[string]interface{}{
				"primary": "#0F766E",
				"accent":  "#FB923C",
			},
		},
	})
	suite.Require().NoError(err)

	layoutID, err := suite.createLayout(createLayoutRequest{
		Handle:      "import-export-layout-" + handleSuffix,
		DisplayName: "Import Export Layout " + handleSuffix,
		Description: "Layout for export/import fresh-pack test",
		Layout: map[string]interface{}{
			"layout": map[string]interface{}{
				"version": 1,
			},
		},
	})
	suite.Require().NoError(err)

	ouID, err := testutils.CreateOrganizationUnit(testutils.OrganizationUnit{
		Handle:      "import-export-ou-" + handleSuffix,
		Name:        "Import Export OU " + handleSuffix,
		Description: "OU for export/import fresh-pack test",
		Parent:      nil,
	})
	suite.Require().NoError(err)

	yamlContent, err := suite.exportResources(exportRequest{
		OrganizationUnits: []string{ouID},
		Flows:             []string{flowID},
		Themes:            []string{themeID},
		Layouts:           []string{layoutID},
	})
	suite.Require().NoError(err)
	suite.Require().NotEmpty(yamlContent)

	err = suite.resetToFreshPack()
	suite.Require().NoError(err)

	suite.assertNotFound("/organization-units/" + ouID)
	suite.assertNotFound("/flows/" + flowID)
	suite.assertNotFound("/design/themes/" + themeID)
	suite.assertNotFound("/design/layouts/" + layoutID)

	importResp, err := suite.importResources(importRequest{
		Content: yamlContent,
		DryRun:  false,
		Options: importOptions{
			Upsert:          false,
			ContinueOnError: true,
			Target:          "runtime",
		},
		Vars: map[string]interface{}{},
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(importResp)

	suite.Equal(4, importResp.Summary.TotalDocuments)
	suite.Equal(4, importResp.Summary.Imported)
	suite.Equal(0, importResp.Summary.Failed)
	suite.Len(importResp.Results, 4)

	resourceTypeToPath := map[string]string{
		"organization_unit": "/organization-units/%s",
		"flow":              "/flows/%s",
		"theme":             "/design/themes/%s",
		"layout":            "/design/layouts/%s",
	}

	seenTypes := map[string]bool{}
	for _, result := range importResp.Results {
		suite.Equal("success", result.Status)
		suite.Equal("create", result.Operation)
		suite.NotEmpty(result.ResourceType)
		suite.NotEmpty(result.ResourceID)

		pathPattern, ok := resourceTypeToPath[result.ResourceType]
		suite.True(ok, "unexpected resourceType in import response: %s", result.ResourceType)
		suite.assertFound(fmt.Sprintf(pathPattern, result.ResourceID))
		seenTypes[result.ResourceType] = true
	}

	suite.True(seenTypes["organization_unit"])
	suite.True(seenTypes["flow"])
	suite.True(seenTypes["theme"])
	suite.True(seenTypes["layout"])
}

func (suite *ImportExportFreshPackSuite) exportResources(reqBody exportRequest) (string, error) {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal export request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		testutils.TestServerURL+"/export",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create export request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-yaml")

	resp, err := testutils.GetHTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send export request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read export response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("export request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var exportResp struct {
		Resources string       `json:"resources"`
		Files     []exportFile `json:"files"`
	}
	if err := json.Unmarshal(body, &exportResp); err != nil {
		return "", fmt.Errorf("failed to parse export response: %w", err)
	}

	if exportResp.Resources != "" {
		return exportResp.Resources, nil
	}

	if len(exportResp.Files) > 0 {
		combined := ""
		for i, file := range exportResp.Files {
			if i > 0 {
				combined += "\n---\n"
			}
			combined += "# File: " + file.FileName + "\n" + file.Content
		}
		return combined, nil
	}

	return "", fmt.Errorf("export response does not contain resources or files")
}

func (suite *ImportExportFreshPackSuite) importResources(reqBody importRequest) (*importResponse, error) {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal import request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		testutils.TestServerURL+"/import",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create import request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := testutils.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send import request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read import response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("import request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var parsed importResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse import response: %w", err)
	}

	return &parsed, nil
}

func (suite *ImportExportFreshPackSuite) createTheme(request createThemeRequest) (string, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal theme request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, testutils.TestServerURL+"/design/themes", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create theme request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := testutils.GetHTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send theme request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read theme response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("theme creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var created map[string]interface{}
	if err := json.Unmarshal(body, &created); err != nil {
		return "", fmt.Errorf("failed to parse theme response: %w", err)
	}

	themeID, _ := created["id"].(string)
	if themeID == "" {
		return "", fmt.Errorf("theme response does not contain id")
	}

	return themeID, nil
}

func (suite *ImportExportFreshPackSuite) createLayout(request createLayoutRequest) (string, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal layout request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, testutils.TestServerURL+"/design/layouts", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create layout request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := testutils.GetHTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send layout request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read layout response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("layout creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var created map[string]interface{}
	if err := json.Unmarshal(body, &created); err != nil {
		return "", fmt.Errorf("failed to parse layout response: %w", err)
	}

	layoutID, _ := created["id"].(string)
	if layoutID == "" {
		return "", fmt.Errorf("layout response does not contain id")
	}

	return layoutID, nil
}

func (suite *ImportExportFreshPackSuite) assertFound(path string) {
	req, err := http.NewRequest(http.MethodGet, testutils.TestServerURL+path, nil)
	suite.Require().NoError(err)

	resp, err := testutils.GetHTTPClient().Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode, "expected resource at %s", path)
}

func (suite *ImportExportFreshPackSuite) assertNotFound(path string) {
	req, err := http.NewRequest(http.MethodGet, testutils.TestServerURL+path, nil)
	suite.Require().NoError(err)

	resp, err := testutils.GetHTTPClient().Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNotFound, resp.StatusCode, "expected resource to be absent at %s", path)
}

func (suite *ImportExportFreshPackSuite) resetToFreshPack() error {
	testutils.StopServer()

	if err := testutils.RunInitScript(testutils.GetZipFilePattern()); err != nil {
		return fmt.Errorf("failed to run init script: %w", err)
	}

	if err := testutils.RunSetupScript(); err != nil {
		return fmt.Errorf("failed to run setup script: %w", err)
	}

	if err := testutils.RestartServer(); err != nil {
		return fmt.Errorf("failed to restart server: %w", err)
	}

	if err := testutils.ObtainAdminAccessToken(); err != nil {
		return fmt.Errorf("failed to re-obtain admin token: %w", err)
	}

	return nil
}
