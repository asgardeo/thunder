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

package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

const handlerLoggerComponentName = "UserHandler"

// userHandler is the handler for user management operations.
type userHandler struct {
	userService UserServiceInterface
}

// newUserHandler creates a new instance of userHandler with dependency injection.
func newUserHandler(userService UserServiceInterface) *userHandler {
	return &userHandler{
		userService: userService,
	}
}

// HandleUserListRequest handles the user list request.
func (uh *userHandler) HandleUserListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	filters, svcErr := parseFilterParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	// Get the user list using the user service.
	userListResponse, svcErr := uh.userService.GetUserList(limit, offset, filters)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully listed users with pagination",
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", userListResponse.TotalResults),
		log.Int("count", userListResponse.Count),
		log.Any("filters", filters))
}

// HandleUserPostRequest handles the user request.
func (uh *userHandler) HandleUserPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	createRequest, err := sysutils.DecodeJSONBody[User](r)
	if err != nil {
		http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		return
	}

	// Create the user using the user service.
	createdUser, svcErr := uh.userService.CreateUser(createRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdUser); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the user creation response.
	logger.Debug("User POST response sent", log.String("user id", createdUser.ID))
}

// HandleUserGetRequest handles the user request.
func (uh *userHandler) HandleUserGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Bad Request: Missing user id.", http.StatusBadRequest)
		return
	}

	// Get the user using the user service.
	user, svcErr := uh.userService.GetUser(id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the user response.
	logger.Debug("User GET response sent", log.String("user id", id))
}

// HandleUserGroupsGetRequest handles the get user groups request.
func (ah *userHandler) HandleUserGroupsGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		handleError(w, logger, &ErrorMissingUserID)
		return
	}

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	groupListResponse, svcErr := ah.userService.GetUserGroups(id, limit, offset)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(groupListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully retrieved user groups", log.String("user id", id),
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", groupListResponse.TotalResults),
		log.Int("count", groupListResponse.Count))
}

// HandleUserPutRequest handles the user request.
func (uh *userHandler) HandleUserPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "Bad Request: Missing user id.", http.StatusBadRequest)
		return
	}

	updateRequest, err := sysutils.DecodeJSONBody[User](r)
	if err != nil {
		http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		return
	}
	updateRequest.ID = id

	// Update the user using the user service.
	user, svcErr := uh.userService.UpdateUser(id, updateRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the user response.
	logger.Debug("User PUT response sent", log.String("user id", id))
}

// HandleUserDeleteRequest handles the delete user request.
func (uh *userHandler) HandleUserDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "Bad Request: Missing user id.", http.StatusBadRequest)
		return
	}

	// Delete the user using the user service.
	svcErr := uh.userService.DeleteUser(id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log the user response.
	logger.Debug("User DELETE response sent", log.String("user id", id))
}

// HandleUserListByPathRequest handles the list users by OU path request.
func (uh *userHandler) HandleUserListByPathRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	path, pathValidationFailed := extractAndValidatePath(w, r, logger)
	if pathValidationFailed {
		return
	}

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	filters, svcErr := parseFilterParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	userListResponse, svcErr := uh.userService.GetUsersByPath(path, limit, offset, filters)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully listed users by path", log.String("path", path),
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", userListResponse.TotalResults),
		log.Int("count", userListResponse.Count),
		log.Any("filters", filters))
}

// HandleUserPostByPathRequest handles the create user by OU path request.
func (uh *userHandler) HandleUserPostByPathRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	path, pathValidationFailed := extractAndValidatePath(w, r, logger)
	if pathValidationFailed {
		return
	}

	createRequest, err := sysutils.DecodeJSONBody[CreateUserByPathRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body: " + err.Error(),
		}

		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	user, svcErr := uh.userService.CreateUserByPath(path, *createRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully created user by path", log.String("path", path), log.String("userType", user.Type))
}

// parsePaginationParams parses limit and offset query parameters from the request.
func parsePaginationParams(query url.Values) (int, int, *serviceerror.ServiceError) {
	limit := 0
	offset := 0

	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
			return 0, 0, &ErrorInvalidLimit
		} else if parsedLimit <= 0 {
			return 0, 0, &ErrorInvalidLimit
		} else {
			limit = parsedLimit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err != nil {
			return 0, 0, &ErrorInvalidOffset
		} else if parsedOffset < 0 {
			return 0, 0, &ErrorInvalidOffset
		} else {
			offset = parsedOffset
		}
	}

	return limit, offset, nil
}

// handleError handles service errors and writes appropriate HTTP responses.
func handleError(w http.ResponseWriter, logger *log.Logger, svcErr *serviceerror.ServiceError) {
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)

	var statusCode int
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case ErrorMissingUserID.Code,
			ErrorUserNotFound.Code,
			ErrorOrganizationUnitNotFound.Code:
			statusCode = http.StatusNotFound
		case ErrorAttributeConflict.Code:
			statusCode = http.StatusConflict
		case ErrorHandlePathRequired.Code,
			ErrorInvalidHandlePath.Code,
			ErrorMissingRequiredFields.Code,
			ErrorMissingCredentials.Code,
			ErrorUserSchemaNotFound.Code:
			statusCode = http.StatusBadRequest
		case ErrorAuthenticationFailed.Code:
			statusCode = http.StatusUnauthorized
		default:
			statusCode = http.StatusBadRequest
		}
	} else {
		statusCode = http.StatusInternalServerError
	}

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

// extractAndValidatePath extracts and validates the path parameter from the request.
func extractAndValidatePath(w http.ResponseWriter, r *http.Request, logger *log.Logger) (string, bool) {
	path := r.PathValue("path")
	if path == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorHandlePathRequired.Code,
			Message:     ErrorHandlePathRequired.Error,
			Description: ErrorHandlePathRequired.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return "", true
	}
	return path, false
}

// parseFilterParams parses and sanitizes filter query parameters from the request.
func parseFilterParams(query url.Values) (map[string]interface{}, *serviceerror.ServiceError) {
	if !query.Has("filter") {
		return make(map[string]interface{}), nil
	}

	filterStr := query.Get("filter")
	filterStr = strings.TrimSpace(filterStr)
	if filterStr == "" {
		return nil, &ErrorInvalidFilter
	}

	parsedFilter, err := parseFilterExpression(filterStr)
	if err != nil {
		return nil, &ErrorInvalidFilter
	}

	sanitized := sanitizeFilter(parsedFilter)

	return sanitized, nil
}

// parseFilterExpression parses filter expressions in the format: attribute eq "value"
func parseFilterExpression(filterStr string) (map[string]interface{}, error) {
	// Regex to match: attribute_name eq "value" or attribute_name eq value
	pattern := `^(\w+(?:\.\w+)*)\s+([A-Za-z]+)\s+(?:"([^"]*)"|([A-Za-z0-9_.+-]+))$`
	regex := regexp.MustCompile(pattern)

	matches := regex.FindStringSubmatch(filterStr)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid filter format")
	}

	attribute := matches[1]
	operator := strings.ToLower(matches[2])

	if operator != "eq" {
		return nil, fmt.Errorf("unsupported operator: %s", operator)
	}

	// Get the value (either quoted string or unquoted value)
	if matches[3] != "" {
		return map[string]interface{}{attribute: matches[3]}, nil
	} else {
		value := matches[4] // Unquoted value
		// Try to convert numeric values
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return map[string]interface{}{attribute: intVal}, nil
		}
		// If not an integer, try to parse as float
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return map[string]interface{}{attribute: floatVal}, nil
		}

		// Check for boolean values
		if bool, err := strconv.ParseBool(value); err == nil {
			return map[string]interface{}{attribute: bool}, nil
		}

		return nil, fmt.Errorf("invalid filter value")
	}
}

// sanitizeFilter performs additional sanitization on parsed filters
func sanitizeFilter(filters map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range filters {
		sanitizedKey := sysutils.SanitizeString(key)

		if strValue, ok := value.(string); ok {
			sanitized[sanitizedKey] = sysutils.SanitizeString(strValue)
		} else {
			sanitized[sanitizedKey] = value
		}
	}

	return sanitized
}
