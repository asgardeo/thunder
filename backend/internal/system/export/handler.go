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
	"fmt"
	"net/http"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// exportHandler defines the handler for managing export API requests.
type exportHandler struct {
	service ExportServiceInterface
}

func newExportHandler(service ExportServiceInterface) *exportHandler {
	return &exportHandler{
		service: service,
	}
}

// HandleExportRequest handles the export request and returns YAML content.
func (eh *exportHandler) HandleExportRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ExportHandler"))

	exportRequest, err := sysutils.DecodeJSONBody[ExportRequest](r)
	if err != nil {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequest.Code,
			Message:     ErrorInvalidRequest.Error,
			Description: ErrorInvalidRequest.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return
	}

	// Export resources using the export service
	exportResponse, svcErr := eh.service.ExportResources(r.Context(), exportRequest)
	if svcErr != nil {
		eh.handleError(w, svcErr)
		return
	}

	// Combine all YAML files into a single response with separators
	var combinedYAML string
	for i, file := range exportResponse.Files {
		if i > 0 {
			combinedYAML += "\n---\n" // YAML document separator
		}
		combinedYAML += "# File: " + file.FileName + "\n"
		combinedYAML += file.Content
	}

	// Return the combined YAML content
	w.Header().Set(serverconst.ContentTypeHeaderName, "application/yaml")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(combinedYAML)); err != nil {
		logger.Error("Error writing YAML response", log.Error(err))
		return
	}
}

// HandleExportJSONRequest handles the export request and returns JSON with files.
func (eh *exportHandler) HandleExportJSONRequest(w http.ResponseWriter, r *http.Request) {
	exportRequest, err := sysutils.DecodeJSONBody[ExportRequest](r)
	if err != nil {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequest.Code,
			Message:     ErrorInvalidRequest.Error,
			Description: ErrorInvalidRequest.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return
	}

	// Export resources using the export service
	exportResponse, svcErr := eh.service.ExportResources(r.Context(), exportRequest)
	if svcErr != nil {
		eh.handleError(w, svcErr)
		return
	}

	// Return the JSON response with files
	sysutils.WriteSuccessResponse(w, http.StatusOK, exportResponse)
}

// HandleExportZipRequest handles the export request and returns a ZIP file containing all resources.
func (eh *exportHandler) HandleExportZipRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ExportHandler"))

	exportRequest, err := sysutils.DecodeJSONBody[ExportRequest](r)
	if err != nil {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequest.Code,
			Message:     ErrorInvalidRequest.Error,
			Description: ErrorInvalidRequest.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return
	}

	// Export resources using the export service
	exportResponse, svcErr := eh.service.ExportResources(r.Context(), exportRequest)
	if svcErr != nil {
		eh.handleError(w, svcErr)
		return
	}

	// Generate ZIP file and send response
	if err := eh.generateAndSendZipResponse(w, logger, exportResponse); err != nil {
		logger.Error("Error generating ZIP response", log.Error(err))
		errResp := apierror.ErrorResponse{
			Code:        ErrorInternalServerError.Code,
			Message:     ErrorInternalServerError.Error,
			Description: ErrorInternalServerError.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusInternalServerError, errResp)
		return
	}
}

// generateAndSendZipResponse creates a ZIP file from export files and sends it as HTTP response.
func (eh *exportHandler) generateAndSendZipResponse(
	w http.ResponseWriter, logger *log.Logger, exportResponse *ExportResponse) error {
	// Create ZIP file in memory
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	// Add each file to the ZIP
	for _, file := range exportResponse.Files {
		// Create the full path within the ZIP
		zipPath := file.FileName
		if file.FolderPath != "" {
			zipPath = file.FolderPath + "/" + file.FileName
		}

		fileWriter, err := zipWriter.Create(zipPath)
		if err != nil {
			logger.Error("Error creating file in ZIP", log.String("zipPath", zipPath), log.Error(err))
			return fmt.Errorf("failed to create file in ZIP: %w", err)
		}

		if _, err := fileWriter.Write([]byte(file.Content)); err != nil {
			logger.Error("Error writing file content to ZIP", log.String("zipPath", zipPath), log.Error(err))
			return fmt.Errorf("failed to write content to ZIP: %w", err)
		}
	}

	// Close the ZIP writer
	if err := zipWriter.Close(); err != nil {
		logger.Error("Error closing ZIP writer", log.Error(err))
		return fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	// Set headers for ZIP file download
	w.Header().Set(serverconst.ContentTypeHeaderName, "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=exported_resources.zip")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", zipBuffer.Len()))
	w.WriteHeader(http.StatusOK)

	// Write the ZIP content
	if _, err := w.Write(zipBuffer.Bytes()); err != nil {
		logger.Error("Error writing ZIP response", log.Error(err))
		return fmt.Errorf("failed to write ZIP response: %w", err)
	}

	return nil
}

// handleError handles service errors and sends appropriate HTTP responses.
func (eh *exportHandler) handleError(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		statusCode = http.StatusBadRequest
	}

	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	sysutils.WriteErrorResponse(w, statusCode, errResp)
}
