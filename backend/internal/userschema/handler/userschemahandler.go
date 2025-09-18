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

// Package handler provides the implementation for user schema management operations.
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema/constants"
	"github.com/asgardeo/thunder/internal/userschema/model"
	"github.com/asgardeo/thunder/internal/userschema/service"
)

const userSchemaLoggerComponentName = "UserSchemaHandler"

// UserSchemaHandler is the handler for user schema management operations.
type UserSchemaHandler struct {
	userSchemaService service.UserSchemaServiceInterface
}

// NewUserSchemaHandler creates a new instance of UserSchemaHandler with dependency injection.
func NewUserSchemaHandler() *UserSchemaHandler {
	return &UserSchemaHandler{
		userSchemaService: service.GetUserSchemaService(),
	}
}

// HandleUserSchemaListRequest handles the user schema list request.
func (ash *UserSchemaHandler) HandleUserSchemaListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	userSchemaListResponse, svcErr := ash.userSchemaService.GetUserSchemaList(limit, offset)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userSchemaListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully listed user schemas with pagination",
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", userSchemaListResponse.TotalResults),
		log.Int("count", userSchemaListResponse.Count))
}

// HandleUserSchemaPostRequest handles the user schema creation request.
func (ash *UserSchemaHandler) HandleUserSchemaPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	createRequest, err := sysutils.DecodeJSONBody[model.CreateUserSchemaRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidRequestFormat.Code,
			Message:     constants.ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body",
		}

		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	sanitizedRequest := ash.sanitizeCreateUserSchemaRequest(*createRequest)

	createdUserSchema, svcErr := ash.userSchemaService.CreateUserSchema(sanitizedRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	if !buildUserSchemaResponse(w, createdUserSchema, logger, http.StatusCreated) {
		return
	}

	logger.Debug("Successfully created user schema",
		log.String("schemaID", createdUserSchema.ID), log.String("name", createdUserSchema.Name))
}

// HandleUserSchemaGetRequest handles the user schema get request.
func (ash *UserSchemaHandler) HandleUserSchemaGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	schemaID, idValidationFailed := extractAndValidateSchemaID(w, r, logger)
	if idValidationFailed {
		return
	}

	userSchema, svcErr := ash.userSchemaService.GetUserSchema(schemaID)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	if !buildUserSchemaResponse(w, userSchema, logger, http.StatusOK) {
		return
	}

	logger.Debug("Successfully retrieved user schema", log.String("schemaID", schemaID))
}

// HandleUserSchemaPutRequest handles the user schema update request.
func (ash *UserSchemaHandler) HandleUserSchemaPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	schemaID, idValidationFailed := extractAndValidateSchemaID(w, r, logger)
	if idValidationFailed {
		return
	}

	sanitizedRequest, requestValidationFailed := validateUpdateUserSchemaRequest(w, r, logger, ash)
	if requestValidationFailed {
		return
	}

	updatedUserSchema, svcErr := ash.userSchemaService.UpdateUserSchema(schemaID, sanitizedRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	if !buildUserSchemaResponse(w, updatedUserSchema, logger, http.StatusOK) {
		return
	}

	logger.Debug("Successfully updated user schema",
		log.String("schemaID", schemaID), log.String("name", updatedUserSchema.Name))
}

// HandleUserSchemaDeleteRequest handles the user schema delete request.
func (ash *UserSchemaHandler) HandleUserSchemaDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	schemaID, idValidationFailed := extractAndValidateSchemaID(w, r, logger)
	if idValidationFailed {
		return
	}

	svcErr := ash.userSchemaService.DeleteUserSchema(schemaID)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Debug("Successfully deleted user schema", log.String("schemaID", schemaID))
}

// parsePaginationParams parses limit and offset from query parameters.
func parsePaginationParams(query map[string][]string) (int, int, *serviceerror.ServiceError) {
	var limit, offset int
	var err error

	if limitStr := query["limit"]; len(limitStr) > 0 && limitStr[0] != "" {
		sanitizedLimit := sysutils.SanitizeString(limitStr[0])
		limit, err = strconv.Atoi(sanitizedLimit)
		if err != nil || limit <= 0 {
			return 0, 0, &constants.ErrorInvalidLimit
		}
	}

	if offsetStr := query["offset"]; len(offsetStr) > 0 && offsetStr[0] != "" {
		sanitizedOffset := sysutils.SanitizeString(offsetStr[0])
		offset, err = strconv.Atoi(sanitizedOffset)
		if err != nil || offset < 0 {
			return 0, 0, &constants.ErrorInvalidOffset
		}
	}

	return limit, offset, nil
}

// handleError handles service errors and converts them to appropriate HTTP responses.
func handleError(w http.ResponseWriter, logger *log.Logger, svcErr *serviceerror.ServiceError) {
	var statusCode int
	if svcErr.Type == serviceerror.ClientErrorType {
		statusCode = http.StatusBadRequest
		if svcErr.Code == constants.ErrorUserSchemaNotFound.Code {
			statusCode = http.StatusNotFound
		} else if svcErr.Code == constants.ErrorUserSchemaNameConflict.Code {
			statusCode = http.StatusConflict
		}
	} else {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(statusCode)

	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Error encoding error response", log.Error(err))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// extractAndValidateSchemaID extracts and validates the schema ID from the URL path.
func extractAndValidateSchemaID(w http.ResponseWriter, r *http.Request, logger *log.Logger) (string, bool) {
	schemaID := r.PathValue("id")
	if schemaID == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidUserSchemaRequest.Code,
			Message:     constants.ErrorInvalidUserSchemaRequest.Error,
			Description: constants.ErrorInvalidUserSchemaRequest.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return "", true
	}

	return schemaID, false
}

func validateUpdateUserSchemaRequest(
	w http.ResponseWriter, r *http.Request, logger *log.Logger, ash *UserSchemaHandler,
) (model.UpdateUserSchemaRequest, bool) {
	updateRequest, err := sysutils.DecodeJSONBody[model.UpdateUserSchemaRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidRequestFormat.Code,
			Message:     constants.ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body",
		}

		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return model.UpdateUserSchemaRequest{}, true
	}

	sanitizedRequest := ash.sanitizeUpdateUserSchemaRequest(*updateRequest)
	return sanitizedRequest, false
}

func buildUserSchemaResponse(
	w http.ResponseWriter, userSchema *model.UserSchema, logger *log.Logger, statusCode int) bool {
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(userSchema); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return false
	}
	return true
}

// sanitizeCreateUserSchemaRequest sanitizes the create user schema request input.
func (ash *UserSchemaHandler) sanitizeCreateUserSchemaRequest(
	request model.CreateUserSchemaRequest,
) model.CreateUserSchemaRequest {
	sanitizedName := sysutils.SanitizeString(request.Name)

	return model.CreateUserSchemaRequest{
		Name:   sanitizedName,
		Schema: request.Schema,
	}
}

// sanitizeUpdateUserSchemaRequest sanitizes the update user schema request input.
func (ash *UserSchemaHandler) sanitizeUpdateUserSchemaRequest(
	request model.UpdateUserSchemaRequest,
) model.UpdateUserSchemaRequest {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	originalName := request.Name
	sanitizedName := sysutils.SanitizeString(request.Name)

	if originalName != sanitizedName {
		logger.Debug("Sanitized user schema name in update request",
			log.String("original", log.MaskString(originalName)),
			log.String("sanitized", log.MaskString(sanitizedName)))
	}

	return model.UpdateUserSchemaRequest{
		Name:   sanitizedName,
		Schema: request.Schema, // Schema is validated separately, no need to sanitize JSON
	}
}
