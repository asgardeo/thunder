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

package usertype

import (
	"net/http"
	"strconv"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

const userTypeHandlerLoggerComponentName = "UserTypeHandler"

// userTypeHandler is the handler for user type management operations.
type userTypeHandler struct {
	userTypeService UserTypeServiceInterface
}

// newUserTypeHandler creates a new instance of userTypeHandler.
func newUserTypeHandler(userTypeService UserTypeServiceInterface) *userTypeHandler {
	return &userTypeHandler{
		userTypeService: userTypeService,
	}
}

// HandleUserTypeListRequest handles the user type list request.
func (h *userTypeHandler) HandleUserTypeListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeHandlerLoggerComponentName))

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	userTypeListResponse, svcErr := h.userTypeService.GetUserTypeList(limit, offset)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, userTypeListResponse)

	logger.Debug("Successfully listed user types with pagination",
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", userTypeListResponse.TotalResults),
		log.Int("count", userTypeListResponse.Count))
}

// HandleUserTypePostRequest handles the user type creation request.
func (h *userTypeHandler) HandleUserTypePostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeHandlerLoggerComponentName))

	createRequest, err := sysutils.DecodeJSONBody[CreateUserTypeRequest](r)
	if err != nil {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body",
		}

		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return
	}

	sanitizedRequest := h.sanitizeCreateUserTypeRequest(*createRequest)

	createdUserType, svcErr := h.userTypeService.CreateUserType(sanitizedRequest)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusCreated, createdUserType)

	logger.Debug("Successfully created user type",
		log.String("schemaID", createdUserType.ID), log.String("name", createdUserType.Name))
}

// HandleUserTypeGetRequest handles the user type get request.
func (h *userTypeHandler) HandleUserTypeGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeHandlerLoggerComponentName))

	schemaID, idValidationFailed := extractAndValidateSchemaID(w, r)
	if idValidationFailed {
		return
	}

	userType, svcErr := h.userTypeService.GetUserType(schemaID)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, userType)

	logger.Debug("Successfully retrieved user type", log.String("schemaID", schemaID))
}

// HandleUserTypePutRequest handles the user type update request.
func (h *userTypeHandler) HandleUserTypePutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeHandlerLoggerComponentName))

	schemaID, idValidationFailed := extractAndValidateSchemaID(w, r)
	if idValidationFailed {
		return
	}

	sanitizedRequest, requestValidationFailed := validateUpdateUserTypeRequest(w, r, h)
	if requestValidationFailed {
		return
	}

	updatedUserType, svcErr := h.userTypeService.UpdateUserType(schemaID, sanitizedRequest)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, updatedUserType)

	logger.Debug("Successfully updated user type",
		log.String("schemaID", schemaID), log.String("name", updatedUserType.Name))
}

// HandleUserTypeDeleteRequest handles the user type delete request.
func (h *userTypeHandler) HandleUserTypeDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeHandlerLoggerComponentName))

	schemaID, idValidationFailed := extractAndValidateSchemaID(w, r)
	if idValidationFailed {
		return
	}

	svcErr := h.userTypeService.DeleteUserType(schemaID)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusNoContent, nil)
	logger.Debug("Successfully deleted user type", log.String("schemaID", schemaID))
}

// parsePaginationParams parses limit and offset from query parameters.
func parsePaginationParams(query map[string][]string) (int, int, *serviceerror.ServiceError) {
	var limit, offset int
	var err error

	if limitStr := query["limit"]; len(limitStr) > 0 && limitStr[0] != "" {
		sanitizedLimit := sysutils.SanitizeString(limitStr[0])
		limit, err = strconv.Atoi(sanitizedLimit)
		if err != nil || limit <= 0 {
			return 0, 0, &ErrorInvalidLimit
		}
	}

	if offsetStr := query["offset"]; len(offsetStr) > 0 && offsetStr[0] != "" {
		sanitizedOffset := sysutils.SanitizeString(offsetStr[0])
		offset, err = strconv.Atoi(sanitizedOffset)
		if err != nil || offset < 0 {
			return 0, 0, &ErrorInvalidOffset
		}
	}

	return limit, offset, nil
}

// handleError handles service errors and converts them to appropriate HTTP responses.
func handleError(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	var statusCode int
	if svcErr.Type == serviceerror.ClientErrorType {
		statusCode = http.StatusBadRequest
		if svcErr.Code == ErrorUserTypeNotFound.Code {
			statusCode = http.StatusNotFound
		} else if svcErr.Code == ErrorUserTypeNameConflict.Code {
			statusCode = http.StatusConflict
		}
	} else {
		statusCode = http.StatusInternalServerError
	}

	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	sysutils.WriteErrorResponse(w, statusCode, errResp)
}

// extractAndValidateSchemaID extracts and validates the schema ID from the URL path.
func extractAndValidateSchemaID(w http.ResponseWriter, r *http.Request) (string, bool) {
	schemaID := r.PathValue("id")
	if schemaID == "" {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidUserTypeRequest.Code,
			Message:     ErrorInvalidUserTypeRequest.Error,
			Description: ErrorInvalidUserTypeRequest.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return "", true
	}

	return schemaID, false
}

func validateUpdateUserTypeRequest(
	w http.ResponseWriter, r *http.Request, h *userTypeHandler,
) (UpdateUserTypeRequest, bool) {
	updateRequest, err := sysutils.DecodeJSONBody[UpdateUserTypeRequest](r)
	if err != nil {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body",
		}

		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return UpdateUserTypeRequest{}, true
	}

	sanitizedRequest := h.sanitizeUpdateUserTypeRequest(*updateRequest)
	return sanitizedRequest, false
}

// sanitizeCreateUserTypeRequest sanitizes the create user type request input.
func (h *userTypeHandler) sanitizeCreateUserTypeRequest(
	request CreateUserTypeRequest,
) CreateUserTypeRequest {
	sanitizedName := sysutils.SanitizeString(request.Name)
	sanitizedOrganizationUnitID := sysutils.SanitizeString(request.OrganizationUnitID)

	return CreateUserTypeRequest{
		Name:                  sanitizedName,
		OrganizationUnitID:    sanitizedOrganizationUnitID,
		AllowSelfRegistration: request.AllowSelfRegistration,
		Schema:                request.Schema,
	}
}

// sanitizeUpdateUserTypeRequest sanitizes the update user type request input.
func (h *userTypeHandler) sanitizeUpdateUserTypeRequest(
	request UpdateUserTypeRequest,
) UpdateUserTypeRequest {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeHandlerLoggerComponentName))

	originalName := request.Name
	sanitizedName := sysutils.SanitizeString(request.Name)
	sanitizedOrganizationUnitID := sysutils.SanitizeString(request.OrganizationUnitID)

	if originalName != sanitizedName {
		logger.Debug("Sanitized user type name in update request",
			log.String("original", log.MaskString(originalName)),
			log.String("sanitized", log.MaskString(sanitizedName)))
	}

	return UpdateUserTypeRequest{
		Name:                  sanitizedName,
		OrganizationUnitID:    sanitizedOrganizationUnitID,
		AllowSelfRegistration: request.AllowSelfRegistration,
		Schema:                request.Schema,
	}
}
