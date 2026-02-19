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

package export

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/notification"
	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/usertype"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/notification/notificationmock"
	"github.com/asgardeo/thunder/tests/mocks/usertypemock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// HandlerTestSuite contains comprehensive tests for the export handler functions.
type HandlerTestSuite struct {
	suite.Suite
	mockAppService          *applicationmock.ApplicationServiceInterfaceMock
	mockIDPService          *idpmock.IDPServiceInterfaceMock
	mockNotificationService *notificationmock.NotificationSenderMgtSvcInterfaceMock
	mockUserTypeService     *usertypemock.UserTypeServiceInterfaceMock
	exportService           ExportServiceInterface
	handler                 *exportHandler
}

func (suite *HandlerTestSuite) SetupTest() {
	// Initialize config for tests
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"https://localhost:3000"},
		},
	}
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	suite.Require().NoError(err)

	// Setup services and handler
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockNotificationService = notificationmock.NewNotificationSenderMgtSvcInterfaceMock(suite.T())
	suite.mockUserTypeService = usertypemock.NewUserTypeServiceInterfaceMock(suite.T())
	exporters := []declarativeresource.ResourceExporter{
		application.NewApplicationExporterForTest(suite.mockAppService),
		idp.NewIDPExporterForTest(suite.mockIDPService),
		notification.NewNotificationSenderExporterForTest(suite.mockNotificationService),
		usertype.NewUserTypeExporterForTest(suite.mockUserTypeService),
	}
	parameterizer := newParameterizer(templatingRules{})
	suite.exportService = newExportService(exporters, parameterizer)
	suite.handler = newExportHandler(suite.exportService)
}

func (suite *HandlerTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

// TestGenerateAndSendZipResponse_Success tests successful ZIP generation and response.
func (suite *HandlerTestSuite) TestGenerateAndSendZipResponse_Success() {
	// Setup test data
	exportResponse := &ExportResponse{
		Files: []ExportFile{
			{
				FileName:   "app1.yaml",
				FolderPath: "applications",
				Content:    "name: test-app-1\ndescription: Test Application 1",
				Size:       42,
			},
			{
				FileName:   "app2.yaml",
				FolderPath: "applications",
				Content:    "name: test-app-2\ndescription: Test Application 2",
				Size:       42,
			},
		},
		Summary: &ExportSummary{
			TotalFiles: 2,
			TotalSize:  84,
		},
	}

	// Create test request and response writer
	w := httptest.NewRecorder()
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TestHandler"))

	// Execute
	err := suite.handler.generateAndSendZipResponse(w, logger, exportResponse)

	// Assert no error
	assert.NoError(suite.T(), err)

	// Verify response headers
	assert.Equal(suite.T(), "application/zip", w.Header().Get(serverconst.ContentTypeHeaderName))
	assert.Equal(suite.T(), "attachment; filename=exported_resources.zip", w.Header().Get("Content-Disposition"))
	assert.NotEmpty(suite.T(), w.Header().Get("Content-Length"))
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify ZIP content
	zipBytes := w.Body.Bytes()
	assert.NotEmpty(suite.T(), zipBytes)

	// Read and verify ZIP contents
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), zipReader.File, 2)

	// Verify first file
	file1 := zipReader.File[0]
	assert.Equal(suite.T(), "applications/app1.yaml", file1.Name)
	reader1, err := file1.Open()
	assert.NoError(suite.T(), err)
	content1, err := io.ReadAll(reader1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "name: test-app-1\ndescription: Test Application 1", string(content1))
	err = reader1.Close()
	assert.NoError(suite.T(), err)

	// Verify second file
	file2 := zipReader.File[1]
	assert.Equal(suite.T(), "applications/app2.yaml", file2.Name)
	reader2, err := file2.Open()
	assert.NoError(suite.T(), err)
	content2, err := io.ReadAll(reader2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "name: test-app-2\ndescription: Test Application 2", string(content2))
	err = reader2.Close()
	assert.NoError(suite.T(), err)
}

// Helper function to test ZIP response generation
func (suite *HandlerTestSuite) testZipResponse(
	exportResponse *ExportResponse, expectedFilePath, expectedContent string) {
	// Create test request and response writer
	w := httptest.NewRecorder()
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TestHandler"))

	// Execute
	err := suite.handler.generateAndSendZipResponse(w, logger, exportResponse)

	// Assert no error
	assert.NoError(suite.T(), err)

	// Verify response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify ZIP content
	zipBytes := w.Body.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), zipReader.File, 1)

	// Verify file
	file := zipReader.File[0]
	assert.Equal(suite.T(), expectedFilePath, file.Name)
	reader, err := file.Open()
	assert.NoError(suite.T(), err)
	content, err := io.ReadAll(reader)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedContent, string(content))
	_ = reader.Close()
}

// TestGenerateAndSendZipResponse_SingleFileNoFolder tests ZIP generation with a single file without folder.
func (suite *HandlerTestSuite) TestGenerateAndSendZipResponse_SingleFileNoFolder() {
	// Setup test data with no folder path
	exportResponse := &ExportResponse{
		Files: []ExportFile{
			{
				FileName:   "standalone.yaml",
				FolderPath: "", // No folder path
				Content:    "name: standalone-app\ndescription: Standalone Application",
				Size:       52,
			},
		},
		Summary: &ExportSummary{
			TotalFiles: 1,
			TotalSize:  52,
		},
	}

	suite.testZipResponse(exportResponse, "standalone.yaml",
		"name: standalone-app\ndescription: Standalone Application")
}

// TestGenerateAndSendZipResponse_EmptyFiles tests ZIP generation with empty files list.
func (suite *HandlerTestSuite) TestGenerateAndSendZipResponse_EmptyFiles() {
	// Setup test data with no files
	exportResponse := &ExportResponse{
		Files: []ExportFile{},
		Summary: &ExportSummary{
			TotalFiles: 0,
			TotalSize:  0,
		},
	}

	// Create test request and response writer
	w := httptest.NewRecorder()
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TestHandler"))

	// Execute
	err := suite.handler.generateAndSendZipResponse(w, logger, exportResponse)

	// Assert no error (empty ZIP should be valid)
	assert.NoError(suite.T(), err)

	// Verify response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify empty ZIP content
	zipBytes := w.Body.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), zipReader.File, 0) // Empty ZIP
}

// TestGenerateAndSendZipResponse_LargeContent tests ZIP generation with large content.
func (suite *HandlerTestSuite) TestGenerateAndSendZipResponse_LargeContent() {
	// Generate large content (1MB)
	largeContent := strings.Repeat("# This is a large YAML file with lots of content\n", 20000)

	exportResponse := &ExportResponse{
		Files: []ExportFile{
			{
				FileName:   "large-app.yaml",
				FolderPath: "large",
				Content:    largeContent,
				Size:       int64(len(largeContent)),
			},
		},
		Summary: &ExportSummary{
			TotalFiles: 1,
			TotalSize:  int64(len(largeContent)),
		},
	}

	// Create test request and response writer
	w := httptest.NewRecorder()
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TestHandler"))

	// Execute
	err := suite.handler.generateAndSendZipResponse(w, logger, exportResponse)

	// Assert no error
	assert.NoError(suite.T(), err)

	// Verify response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify ZIP content
	zipBytes := w.Body.Bytes()
	assert.True(suite.T(), len(zipBytes) > 0)
	assert.True(suite.T(), len(zipBytes) < len(largeContent)) // Should be compressed

	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), zipReader.File, 1)

	// Verify compressed file
	file := zipReader.File[0]
	assert.Equal(suite.T(), "large/large-app.yaml", file.Name)
	reader, err := file.Open()
	assert.NoError(suite.T(), err)
	content, err := io.ReadAll(reader)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), largeContent, string(content))
	_ = reader.Close()
}

// TestGenerateAndSendZipResponse_SpecialCharactersInPath tests ZIP generation with special characters in paths.
func (suite *HandlerTestSuite) TestGenerateAndSendZipResponse_SpecialCharactersInPath() {
	exportResponse := &ExportResponse{
		Files: []ExportFile{
			{
				FileName:   "app with spaces.yaml",
				FolderPath: "special-chars/ñañé-çæß",
				Content:    "name: app-with-special-chars\ndata: ñañé-çæß-special",
				Size:       50,
			},
		},
		Summary: &ExportSummary{
			TotalFiles: 1,
			TotalSize:  50,
		},
	}

	suite.testZipResponse(exportResponse, "special-chars/ñañé-çæß/app with spaces.yaml",
		"name: app-with-special-chars\ndata: ñañé-çæß-special")
}

// TestGenerateAndSendZipResponse_DeepFolderStructure tests ZIP generation with deep folder structure.
func (suite *HandlerTestSuite) TestGenerateAndSendZipResponse_DeepFolderStructure() {
	exportResponse := &ExportResponse{
		Files: []ExportFile{
			{
				FileName:   "deep-app.yaml",
				FolderPath: "level1/level2/level3/level4/level5",
				Content:    "name: deep-nested-app\nlocation: very-deep",
				Size:       42,
			},
		},
		Summary: &ExportSummary{
			TotalFiles: 1,
			TotalSize:  42,
		},
	}

	suite.testZipResponse(exportResponse,
		"level1/level2/level3/level4/level5/deep-app.yaml", "name: deep-nested-app\nlocation: very-deep")
}

// Standalone tests for simpler use cases

// TestGenerateAndSendZipResponse_Standalone tests the function without suite dependencies.
func TestGenerateAndSendZipResponse_Standalone(t *testing.T) {
	logger := log.GetLogger()
	// Setup config
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
		},
	}
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	assert.NoError(t, err)
	defer config.ResetThunderRuntime()

	// Setup handler
	mockAppService := applicationmock.NewApplicationServiceInterfaceMock(t)
	mockIDPService := idpmock.NewIDPServiceInterfaceMock(t)
	mockNotificationService := notificationmock.NewNotificationSenderMgtSvcInterfaceMock(t)
	mockUserTypeService := usertypemock.NewUserTypeServiceInterfaceMock(t)
	exporters := []declarativeresource.ResourceExporter{
		application.NewApplicationExporterForTest(mockAppService),
		idp.NewIDPExporterForTest(mockIDPService),
		notification.NewNotificationSenderExporterForTest(mockNotificationService),
		usertype.NewUserTypeExporterForTest(mockUserTypeService),
	}
	parameterizer := newParameterizer(templatingRules{})
	exportService := newExportService(exporters, parameterizer)
	handler := newExportHandler(exportService)

	// Test data
	exportResponse := &ExportResponse{
		Files: []ExportFile{
			{
				FileName:   "test.yaml",
				FolderPath: "test",
				Content:    "test: content",
				Size:       13,
			},
		},
	}

	// Execute
	w := httptest.NewRecorder()
	err = handler.generateAndSendZipResponse(w, logger, exportResponse)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
}

// TestNewExportHandler tests the handler constructor.
func TestNewExportHandler(t *testing.T) {
	mockAppService := applicationmock.NewApplicationServiceInterfaceMock(t)
	mockIDPService := idpmock.NewIDPServiceInterfaceMock(t)
	mockNotificationService := notificationmock.NewNotificationSenderMgtSvcInterfaceMock(t)
	mockUserTypeService := usertypemock.NewUserTypeServiceInterfaceMock(t)
	exporters := []declarativeresource.ResourceExporter{
		application.NewApplicationExporterForTest(mockAppService),
		idp.NewIDPExporterForTest(mockIDPService),
		notification.NewNotificationSenderExporterForTest(mockNotificationService),
		usertype.NewUserTypeExporterForTest(mockUserTypeService),
	}
	parameterizer := newParameterizer(templatingRules{})
	exportService := newExportService(exporters, parameterizer)

	handler := newExportHandler(exportService)

	assert.NotNil(t, handler)
	assert.Equal(t, exportService, handler.service)
}

// Handler Function Tests

// TestHandleExportRequest_Success tests successful YAML export.
func (suite *HandlerTestSuite) TestHandleExportRequest_Success() {
	// Setup mock expectations
	suite.mockAppService.EXPECT().GetApplication("app1").Return(&model.Application{
		ID:          "app1",
		Name:        "Test App 1",
		Description: "Test Application 1",
		URL:         "https://example.com",
	}, nil).Once()

	// Create request body
	requestBody := &ExportRequest{
		Applications: []string{"app1"},
		Options: &ExportOptions{
			Format:          "yaml",
			IncludeMetadata: true,
		},
	}
	requestJSON, _ := json.Marshal(requestBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/export", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportRequest(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/yaml", w.Header().Get("Content-Type"))

	responseBody := w.Body.String()
	assert.NotEmpty(suite.T(), responseBody)
	assert.Contains(suite.T(), responseBody, "# File:")
	assert.Contains(suite.T(), responseBody, "name: Test App 1")
}

// TestHandleExportRequest_InvalidJSON tests invalid JSON request handling.
func (suite *HandlerTestSuite) TestHandleExportRequest_InvalidJSON() {
	// Create malformed JSON request
	req := httptest.NewRequest("POST", "/export", strings.NewReader("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportRequest(w, req)

	// Assert error response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "EXP-1001", errResp["code"])
	assert.Equal(suite.T(), "Invalid export request", errResp["message"])
}

// Helper function to test service error responses
func (suite *HandlerTestSuite) testServiceErrorResponse(
	method, endpoint, appID string, serviceError *serviceerror.ServiceError, expectedErrorCode string) {
	// Setup mock to return service error
	suite.mockAppService.EXPECT().GetApplication(appID).Return(nil, serviceError).Once()

	// Create request body
	requestBody := &ExportRequest{
		Applications: []string{appID},
	}
	requestJSON, _ := json.Marshal(requestBody)

	// Create HTTP request
	req := httptest.NewRequest(method, endpoint, bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute based on endpoint
	switch endpoint {
	case "/export":
		suite.handler.HandleExportRequest(w, req)
	case "/export/json":
		suite.handler.HandleExportJSONRequest(w, req)
	case "/export/zip":
		suite.handler.HandleExportZipRequest(w, req)
	}

	// Assert error response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedErrorCode, errResp["code"])
}

// TestHandleExportRequest_ServiceError tests service error handling.
func (suite *HandlerTestSuite) TestHandleExportRequest_ServiceError() {
	suite.testServiceErrorResponse("POST", "/export", "app1", &ErrorNoResourcesFound, "EXP-1002")
}

// TestHandleExportRequest_MultipleFiles tests YAML export with multiple files.
func (suite *HandlerTestSuite) TestHandleExportRequest_MultipleFiles() {
	// Setup mock expectations for multiple applications
	suite.mockAppService.EXPECT().GetApplication("app1").Return(&model.Application{
		ID:   "app1",
		Name: "App One",
	}, nil).Once()
	suite.mockAppService.EXPECT().GetApplication("app2").Return(&model.Application{
		ID:   "app2",
		Name: "App Two",
	}, nil).Once()

	// Create request body
	requestBody := &ExportRequest{
		Applications: []string{"app1", "app2"},
	}
	requestJSON, _ := json.Marshal(requestBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/export", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportRequest(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	responseBody := w.Body.String()

	// Verify YAML separator between files
	assert.Contains(suite.T(), responseBody, "---")
	assert.Contains(suite.T(), responseBody, "name: App One")
	assert.Contains(suite.T(), responseBody, "name: App Two")

	// Count file headers
	fileHeaders := strings.Count(responseBody, "# File:")
	assert.Equal(suite.T(), 2, fileHeaders)
}

// TestHandleExportJSONRequest_Success tests successful JSON export.
func (suite *HandlerTestSuite) TestHandleExportJSONRequest_Success() {
	// Setup mock expectations
	suite.mockAppService.EXPECT().GetApplication("app1").Return(&model.Application{
		ID:          "app1",
		Name:        "Test App JSON",
		Description: "JSON Test Application",
	}, nil).Once()

	// Create request body
	requestBody := &ExportRequest{
		Applications: []string{"app1"},
		Options: &ExportOptions{
			Format: "json", // Note: JSON format currently falls back to YAML
		},
	}
	requestJSON, _ := json.Marshal(requestBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/export/json", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportJSONRequest(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var response ExportResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response.Files, 1)
	// Note: Service currently generates YAML files even with JSON format (fallback behavior)
	assert.Equal(suite.T(), "Test_App_JSON.yaml", response.Files[0].FileName)
	assert.Contains(suite.T(), response.Files[0].Content, "Test App JSON")
	assert.NotNil(suite.T(), response.Summary)
	assert.Equal(suite.T(), 1, response.Summary.TotalFiles)
}

// TestHandleExportJSONRequest_InvalidJSON tests invalid JSON handling for JSON export.
func (suite *HandlerTestSuite) TestHandleExportJSONRequest_InvalidJSON() {
	// Create malformed JSON request
	req := httptest.NewRequest("POST", "/export/json", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportJSONRequest(w, req)

	// Assert error response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "EXP-1001", errResp["code"])
}

// TestHandleExportJSONRequest_ServiceError tests service error handling for JSON export.
func (suite *HandlerTestSuite) TestHandleExportJSONRequest_ServiceError() {
	// Setup mock to return service error
	suite.testServiceErrorResponse("POST", "/export/json", "app1", &ErrorInternalServerError, "EXP-1002")
}

// TestHandleExportZipRequest_Success tests successful ZIP export.
func (suite *HandlerTestSuite) TestHandleExportZipRequest_Success() {
	// Setup mock expectations
	suite.mockAppService.EXPECT().GetApplication("app1").Return(&model.Application{
		ID:   "app1",
		Name: "ZIP Test App",
	}, nil).Once()

	// Create request body
	requestBody := &ExportRequest{
		Applications: []string{"app1"},
		Options: &ExportOptions{
			FolderStructure: &FolderStructureOptions{
				GroupByType: true,
			},
		},
	}
	requestJSON, _ := json.Marshal(requestBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/export/zip", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportZipRequest(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/zip", w.Header().Get("Content-Type"))
	assert.Contains(suite.T(), w.Header().Get("Content-Disposition"), "attachment")
	assert.Contains(suite.T(), w.Header().Get("Content-Disposition"), "exported_resources.zip")
	assert.NotEmpty(suite.T(), w.Header().Get("Content-Length"))

	// Verify ZIP content
	zipBytes := w.Body.Bytes()
	assert.NotEmpty(suite.T(), zipBytes)

	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), zipReader.File, 1)
	// Note: Service uses name-based file naming
	assert.Equal(suite.T(), "applications/ZIP_Test_App.yaml", zipReader.File[0].Name)
}

// TestHandleExportZipRequest_InvalidJSON tests invalid JSON handling for ZIP export.
func (suite *HandlerTestSuite) TestHandleExportZipRequest_InvalidJSON() {
	// Create malformed JSON request
	req := httptest.NewRequest("POST", "/export/zip", strings.NewReader("}{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportZipRequest(w, req)

	// Assert error response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "EXP-1001", errResp["code"])
}

// TestHandleExportZipRequest_ServiceError tests service error handling for ZIP export.
func (suite *HandlerTestSuite) TestHandleExportZipRequest_ServiceError() {
	suite.testServiceErrorResponse("POST", "/export/zip", "nonexistent", &ErrorNoResourcesFound, "EXP-1002")
}

// TestHandleError_ClientError tests error handling for client errors.
func (suite *HandlerTestSuite) TestHandleError_ClientError() {
	w := httptest.NewRecorder()

	// Create client error
	clientErr := &ErrorNoResourcesFound

	// Execute
	suite.handler.handleError(w, clientErr)

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "EXP-1002", errResp["code"])
	assert.Equal(suite.T(), "No resources found", errResp["message"])
	assert.Equal(suite.T(), "No valid resources found for the provided identifiers", errResp["description"])
}

// TestHandleError_ServerError tests error handling for server errors.
func (suite *HandlerTestSuite) TestHandleError_ServerError() {
	w := httptest.NewRecorder()

	// Create server error
	serverErr := &ErrorInternalServerError

	// Execute
	suite.handler.handleError(w, serverErr)

	// Assert response
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "EXP-5001", errResp["code"])
	assert.Equal(suite.T(), "Internal server error", errResp["message"])
	assert.Equal(suite.T(), "An unexpected error occurred while processing the export request", errResp["description"])
}

// Edge case tests

// TestHandleExportRequest_EmptyBody tests empty request body.
func (suite *HandlerTestSuite) TestHandleExportRequest_EmptyBody() {
	req := httptest.NewRequest("POST", "/export", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportRequest(w, req)

	// Assert error response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))
}

// TestHandleExportRequest_NilOptions tests request with nil options.
func (suite *HandlerTestSuite) TestHandleExportRequest_NilOptions() {
	// Setup mock expectations
	suite.mockAppService.EXPECT().GetApplication("app1").Return(&model.Application{
		ID:   "app1",
		Name: "Test App",
	}, nil).Once()

	// Create request body with nil options
	requestBody := &ExportRequest{
		Applications: []string{"app1"},
		Options:      nil, // Test nil options
	}
	requestJSON, _ := json.Marshal(requestBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/export", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportRequest(w, req)

	// Assert successful response with default behavior
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/yaml", w.Header().Get("Content-Type"))
}

// TestHandleExportJSONRequest_EmptyFiles tests JSON export with no files.
func (suite *HandlerTestSuite) TestHandleExportJSONRequest_EmptyFiles() {
	// Create request body with empty applications
	requestBody := &ExportRequest{
		Applications: []string{}, // No applications
	}
	requestJSON, _ := json.Marshal(requestBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/export/json", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleExportJSONRequest(w, req)

	// Assert error response (empty applications list returns NoResourcesFound error)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "EXP-1002", errResp["code"]) // NoResourcesFound
	assert.Equal(suite.T(), "No resources found", errResp["message"])
}

// Benchmark tests

// BenchmarkGenerateAndSendZipResponse benchmarks ZIP generation performance.
func BenchmarkGenerateAndSendZipResponse(b *testing.B) {
	logger := log.GetLogger()
	// Setup
	config.ResetThunderRuntime()
	testConfig := &config.Config{}
	_ = config.InitializeThunderRuntime("/tmp/test", testConfig)
	defer config.ResetThunderRuntime()

	mockAppService := applicationmock.NewApplicationServiceInterfaceMock(b)
	mockIDPService := idpmock.NewIDPServiceInterfaceMock(b)
	mockNotificationService := notificationmock.NewNotificationSenderMgtSvcInterfaceMock(b)
	mockUserTypeService := usertypemock.NewUserTypeServiceInterfaceMock(b)
	exporters := []declarativeresource.ResourceExporter{
		application.NewApplicationExporterForTest(mockAppService),
		idp.NewIDPExporterForTest(mockIDPService),
		notification.NewNotificationSenderExporterForTest(mockNotificationService),
		usertype.NewUserTypeExporterForTest(mockUserTypeService),
	}
	parameterizer := newParameterizer(templatingRules{})
	exportService := newExportService(exporters, parameterizer)
	handler := newExportHandler(exportService)

	exportResponse := &ExportResponse{
		Files: []ExportFile{
			{
				FileName:   "benchmark.yaml",
				FolderPath: "benchmark",
				Content:    strings.Repeat("data: value\n", 100),
				Size:       1100,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		_ = handler.generateAndSendZipResponse(w, logger, exportResponse)
	}
}

// Helper function for benchmark tests
func setupBenchmarkTest(b *testing.B) (*exportHandler, []byte) {
	// Setup
	config.ResetThunderRuntime()
	testConfig := &config.Config{}
	_ = config.InitializeThunderRuntime("/tmp/test", testConfig)
	b.Cleanup(func() { config.ResetThunderRuntime() })

	mockAppService := applicationmock.NewApplicationServiceInterfaceMock(b)
	mockIDPService := idpmock.NewIDPServiceInterfaceMock(b)
	mockNotificationService := notificationmock.NewNotificationSenderMgtSvcInterfaceMock(b)
	mockUserTypeService := usertypemock.NewUserTypeServiceInterfaceMock(b)
	exporters := []declarativeresource.ResourceExporter{
		application.NewApplicationExporterForTest(mockAppService),
		idp.NewIDPExporterForTest(mockIDPService),
		notification.NewNotificationSenderExporterForTest(mockNotificationService),
		usertype.NewUserTypeExporterForTest(mockUserTypeService),
	}
	parameterizer := newParameterizer(templatingRules{})
	exportService := newExportService(exporters, parameterizer)
	handler := newExportHandler(exportService)

	// Setup mock expectation
	mockAppService.EXPECT().GetApplication("benchmark-app").Return(&model.Application{
		ID:   "benchmark-app",
		Name: "Benchmark Application",
	}, nil).Times(b.N)

	// Create request body
	requestBody := &ExportRequest{
		Applications: []string{"benchmark-app"},
	}
	requestJSON, _ := json.Marshal(requestBody)

	return handler, requestJSON
}

// BenchmarkHandleExportRequest benchmarks YAML export performance.
func BenchmarkHandleExportRequest(b *testing.B) {
	handler, requestJSON := setupBenchmarkTest(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/export", bytes.NewReader(requestJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.HandleExportRequest(w, req)
	}
}

// BenchmarkHandleExportJSONRequest benchmarks JSON export performance.
func BenchmarkHandleExportJSONRequest(b *testing.B) {
	handler, requestJSON := setupBenchmarkTest(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/export/json", bytes.NewReader(requestJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.HandleExportJSONRequest(w, req)
	}
}
