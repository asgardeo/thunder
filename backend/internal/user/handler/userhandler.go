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

// Package handler provides the implementation for user management operations.
package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/user/constants"
	"github.com/asgardeo/thunder/internal/user/model"
	"github.com/asgardeo/thunder/internal/user/service"
)

const loggerComponentName = "UserHandler"

// UserHandler is the handler for user management operations.
type UserHandler struct {
	userService service.UserServiceInterface
}

// NewUserHandler creates a new instance of UserHandler with dependency injection.
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.GetUserService(),
	}
}

// HandleUserListRequest handles the user list request.
func (ah *UserHandler) HandleUserListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	// Get the user list using the user service.
	userListResponse, svcErr := ah.userService.GetUserList(limit, offset)
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
		log.Int("count", userListResponse.Count))
}

// HandleUserPostRequest handles the user request.
func (ah *UserHandler) HandleUserPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserHandler"))

	createRequest, err := sysutils.DecodeJSONBody[model.User](r)
	if err != nil {
		http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		return
	}

	// Create the user using the user service.
	createdUser, svcErr := ah.userService.CreateUser(createRequest)
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
func (ah *UserHandler) HandleUserGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "Bad Request: Missing user id.", http.StatusBadRequest)
		return
	}

	// Get the user using the user service.
	user, svcErr := ah.userService.GetUser(id)
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

// HandleUserPutRequest handles the user request.
func (ah *UserHandler) HandleUserPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "Bad Request: Missing user id.", http.StatusBadRequest)
		return
	}

	updateRequest, err := sysutils.DecodeJSONBody[model.User](r)
	if err != nil {
		http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		return
	}
	updateRequest.ID = id

	// Update the user using the user service.
	user, svcErr := ah.userService.UpdateUser(id, updateRequest)
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
func (ah *UserHandler) HandleUserDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "Bad Request: Missing user id.", http.StatusBadRequest)
		return
	}

	// Delete the user using the user service.
	svcErr := ah.userService.DeleteUser(id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log the user response.
	logger.Debug("User DELETE response sent", log.String("user id", id))
}

// HandleUserListByPathRequest handles the list users by OU path request.
func (ah *UserHandler) HandleUserListByPathRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

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

	userListResponse, svcErr := ah.userService.GetUsersByPath(path, limit, offset)
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
		log.Int("count", userListResponse.Count))
}

// HandleUserPostByPathRequest handles the create user by OU path request.
func (ah *UserHandler) HandleUserPostByPathRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	path, pathValidationFailed := extractAndValidatePath(w, r, logger)
	if pathValidationFailed {
		return
	}

	createRequest, err := sysutils.DecodeJSONBody[model.CreateUserByPathRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidRequestFormat.Code,
			Message:     constants.ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body: " + err.Error(),
		}

		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	user, svcErr := ah.userService.CreateUserByPath(path, *createRequest)
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

// HandleUserAuthenticateRequest handles the user authentication request.
func (ah *UserHandler) HandleUserAuthenticateRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	authenticateRequest, err := sysutils.DecodeJSONBody[model.AuthenticateUserRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidRequestFormat.Code,
			Message:     constants.ErrorInvalidRequestFormat.Error,
			Description: "The request body is malformed or contains invalid data",
		}

		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	authResponse, svcErr := ah.userService.AuthenticateUser(*authenticateRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(authResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("User authentication successful", log.String("userID", authResponse.ID))
}

// parsePaginationParams parses limit and offset query parameters from the request.
func parsePaginationParams(query url.Values) (int, int, *serviceerror.ServiceError) {
	limit := 0
	offset := 0

	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
			return 0, 0, &constants.ErrorInvalidLimit
		} else if parsedLimit <= 0 {
			return 0, 0, &constants.ErrorInvalidLimit
		} else {
			limit = parsedLimit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err != nil {
			return 0, 0, &constants.ErrorInvalidOffset
		} else if parsedOffset < 0 {
			return 0, 0, &constants.ErrorInvalidOffset
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
		case constants.ErrorMissingUserID.Code,
			constants.ErrorUserNotFound.Code,
			constants.ErrorOrganizationUnitNotFound.Code:
			statusCode = http.StatusNotFound
		case constants.ErrorAttributeConflict.Code:
			statusCode = http.StatusConflict
		case constants.ErrorHandlePathRequired.Code,
			constants.ErrorInvalidHandlePath.Code,
			constants.ErrorMissingRequiredFields.Code,
			constants.ErrorMissingCredentials.Code:
			statusCode = http.StatusBadRequest
		case constants.ErrorAuthenticationFailed.Code:
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
			Code:        constants.ErrorHandlePathRequired.Code,
			Message:     constants.ErrorHandlePathRequired.Error,
			Description: constants.ErrorHandlePathRequired.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return "", true
	}
	return path, false
}
