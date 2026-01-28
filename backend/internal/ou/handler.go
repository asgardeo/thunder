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

package ou

import (
	"net/http"
	"net/url"
	"strconv"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "OrganizationUnitHandler"

// organizationUnitHandler is the handler for organization unit management operations.
type organizationUnitHandler struct {
	service OrganizationUnitServiceInterface
}

// newOrganizationUnitHandler creates a new instance of organizationUnitHandler
func newOrganizationUnitHandler(service OrganizationUnitServiceInterface) *organizationUnitHandler {
	return &organizationUnitHandler{
		service: service,
	}
}

// HandleOUListRequest handles the list organization units request.
func (ouh *organizationUnitHandler) HandleOUListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	ouListResponse, svcErr := ouh.service.GetOrganizationUnitList(limit, offset)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, ouListResponse)

	logger.Debug("Successfully listed organization units with pagination",
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", ouListResponse.TotalResults),
		log.Int("count", ouListResponse.Count))
}

// HandleOUPostRequest handles the create organization unit request.
func (ouh *organizationUnitHandler) HandleOUPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	createRequest, err := sysutils.DecodeJSONBody[OrganizationUnitRequest](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body: " + err.Error(),
		})
		return
	}

	sanitizedRequest := ouh.sanitizeOrganizationUnitRequest(*createRequest)

	createdOU, svcErr := ouh.service.CreateOrganizationUnit(sanitizedRequest)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusCreated, createdOU)

	logger.Debug("Successfully created organization unit", log.String("ouId", createdOU.ID))
}

// HandleOUGetRequest handles the get organization unit by id request.
func (ouh *organizationUnitHandler) HandleOUGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id, idValidateFailed := extractAndValidateID(w, r)
	if idValidateFailed {
		return
	}

	ou, svcErr := ouh.service.GetOrganizationUnit(r.Context(), id)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, ou)

	logger.Debug("Successfully retrieved organization unit", log.String("ouId", id))
}

// HandleOUPutRequest handles the update organization unit request.
func (ouh *organizationUnitHandler) HandleOUPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id, idValidateFailed := extractAndValidateID(w, r)
	if idValidateFailed {
		return
	}

	sanitizedRequest, requestValidationFailed := validateUpdateRequest(w, r, ouh)
	if requestValidationFailed {
		return
	}

	ou, svcErr := ouh.service.UpdateOrganizationUnit(id, sanitizedRequest)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, ou)

	logger.Debug("Successfully updated organization unit", log.String("ouId", id))
}

// HandleOUDeleteRequest handles the delete organization unit request.
func (ouh *organizationUnitHandler) HandleOUDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id, idValidateFailed := extractAndValidateID(w, r)
	if idValidateFailed {
		return
	}

	svcErr := ouh.service.DeleteOrganizationUnit(id)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusNoContent, nil)
	logger.Debug("Successfully deleted organization unit", log.String("ouId", id))
}

// HandleOUChildrenListRequest handles the list child organization units request.
func (ouh *organizationUnitHandler) HandleOUChildrenListRequest(w http.ResponseWriter, r *http.Request) {
	ouh.handleResourceListRequest(w, r, "child organization units",
		func(id string, limit, offset int) (interface{}, *serviceerror.ServiceError) {
			return ouh.service.GetOrganizationUnitChildren(r.Context(), id, limit, offset)
		})
}

// HandleOUUsersListRequest handles the list users in organization unit request.
func (ouh *organizationUnitHandler) HandleOUUsersListRequest(w http.ResponseWriter, r *http.Request) {
	ouh.handleResourceListRequest(w, r, "users",
		func(id string, limit, offset int) (interface{}, *serviceerror.ServiceError) {
			return ouh.service.GetOrganizationUnitUsers(id, limit, offset)
		})
}

// HandleOUGroupsListRequest handles the list groups in organization unit request.
func (ouh *organizationUnitHandler) HandleOUGroupsListRequest(w http.ResponseWriter, r *http.Request) {
	ouh.handleResourceListRequest(w, r, "groups",
		func(id string, limit, offset int) (interface{}, *serviceerror.ServiceError) {
			return ouh.service.GetOrganizationUnitGroups(id, limit, offset)
		})
}

// handleError handles service errors and returns appropriate HTTP responses.
func (ouh *organizationUnitHandler) handleError(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	var statusCode int
	switch svcErr.Type {
	case serviceerror.ClientErrorType:
		statusCode = http.StatusBadRequest
		if svcErr.Code == ErrorOrganizationUnitNotFound.Code {
			statusCode = http.StatusNotFound
		} else if svcErr.Code == ErrorOrganizationUnitNameConflict.Code ||
			svcErr.Code == ErrorOrganizationUnitHandleConflict.Code {
			statusCode = http.StatusConflict
		} else if svcErr.Code == ErrorInvalidLimit.Code ||
			svcErr.Code == ErrorInvalidOffset.Code ||
			svcErr.Code == ErrorInvalidHandlePath.Code {
			statusCode = http.StatusBadRequest
		}
	default:
		statusCode = http.StatusInternalServerError
	}

	sysutils.WriteErrorResponse(w, statusCode, apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	})
}

// sanitizeOrganizationUnitRequest sanitizes the create organization unit request input.
func (ouh *organizationUnitHandler) sanitizeOrganizationUnitRequest(
	request OrganizationUnitRequest,
) OrganizationUnitRequest {
	return OrganizationUnitRequest{
		Handle:      sysutils.SanitizeString(request.Handle),
		Name:        sysutils.SanitizeString(request.Name),
		Description: sysutils.SanitizeString(request.Description),
		Parent:      request.Parent,
	}
}

func extractAndValidateID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if id == "" {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, apierror.ErrorResponse{
			Code:        ErrorMissingOrganizationUnitID.Code,
			Message:     ErrorMissingOrganizationUnitID.Error,
			Description: ErrorMissingOrganizationUnitID.ErrorDescription,
		})
		return "", true
	}
	return id, false
}

func validateUpdateRequest(
	w http.ResponseWriter, r *http.Request, ouh *organizationUnitHandler,
) (OrganizationUnitRequest, bool) {
	updateRequest, err := sysutils.DecodeJSONBody[OrganizationUnitRequest](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body: " + err.Error(),
		})
		return OrganizationUnitRequest{}, true
	}

	sanitizedRequest := ouh.sanitizeOrganizationUnitRequest(*updateRequest)
	return sanitizedRequest, false
}

// parsePaginationParams parses limit and offset query parameters from the request.
func parsePaginationParams(query url.Values) (int, int, *serviceerror.ServiceError) {
	limit := 0
	offset := 0

	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
			return 0, 0, &ErrorInvalidLimit
		} else {
			limit = parsedLimit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err != nil {
			return 0, 0, &ErrorInvalidOffset
		} else {
			offset = parsedOffset
		}
	}

	return limit, offset, nil
}

// handleResourceListRequest is a generic handler for listing resources under an organization unit.
func (ouh *organizationUnitHandler) handleResourceListRequest(
	w http.ResponseWriter, r *http.Request, resourceType string,
	serviceFunc func(string, int, int) (interface{}, *serviceerror.ServiceError),
) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id, idValidateFailed := extractAndValidateID(w, r)
	if idValidateFailed {
		return
	}

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	response, svcErr := serviceFunc(id, limit, offset)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, response)

	// Extract pagination info for logging based on response type
	var totalResults, count int
	switch resp := response.(type) {
	case *OrganizationUnitListResponse:
		totalResults = resp.TotalResults
		count = resp.Count
	case *UserListResponse:
		totalResults = resp.TotalResults
		count = resp.Count
	case *GroupListResponse:
		totalResults = resp.TotalResults
		count = resp.Count
	}

	logger.Debug("Successfully listed resources in organization unit", log.String("resourceType", resourceType),
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", totalResults),
		log.Int("count", count))
}

// HandleOUGetByPathRequest handles the get organization unit by hierarchical handle path request.
func (ouh *organizationUnitHandler) HandleOUGetByPathRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	path, pathValidationFailed := extractAndValidatePath(w, r)
	if pathValidationFailed {
		return
	}

	ou, svcErr := ouh.service.GetOrganizationUnitByPath(r.Context(), path)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, ou)

	logger.Debug("Successfully retrieved organization unit by path", log.String("path", path))
}

// HandleOUPutByPathRequest handles the update organization unit by hierarchical handle path request.
func (ouh *organizationUnitHandler) HandleOUPutByPathRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	path, pathValidationFailed := extractAndValidatePath(w, r)
	if pathValidationFailed {
		return
	}

	sanitizedRequest, requestValidationFailed := validateUpdateRequest(w, r, ouh)
	if requestValidationFailed {
		return
	}

	ou, svcErr := ouh.service.UpdateOrganizationUnitByPath(path, sanitizedRequest)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, ou)

	logger.Debug("Successfully updated organization unit by path", log.String("path", path))
}

// HandleOUDeleteByPathRequest handles the delete organization unit by hierarchical handle path request.
func (ouh *organizationUnitHandler) HandleOUDeleteByPathRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	path, pathValidationFailed := extractAndValidatePath(w, r)
	if pathValidationFailed {
		return
	}

	svcErr := ouh.service.DeleteOrganizationUnitByPath(path)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusNoContent, nil)
	logger.Debug("Successfully deleted organization unit by path", log.String("path", path))
}

// handleResourceListByPathRequest is a generic handler for listing resources under an organization unit by path.
func (ouh *organizationUnitHandler) handleResourceListByPathRequest(
	w http.ResponseWriter, r *http.Request, resourceType string,
	serviceFunc func(string, int, int) (interface{}, *serviceerror.ServiceError),
) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	path, pathValidationFailed := extractAndValidatePath(w, r)
	if pathValidationFailed {
		return
	}

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	response, svcErr := serviceFunc(path, limit, offset)
	if svcErr != nil {
		ouh.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, response)

	if logger.IsDebugEnabled() {
		var totalResults, count int
		switch resp := response.(type) {
		case *OrganizationUnitListResponse:
			totalResults = resp.TotalResults
			count = resp.Count
		case *UserListResponse:
			totalResults = resp.TotalResults
			count = resp.Count
		case *GroupListResponse:
			totalResults = resp.TotalResults
			count = resp.Count
		}

		logger.Debug("Successfully listed resources in organization unit by path",
			log.String("resourceType", resourceType), log.String("path", path),
			log.Int("limit", limit), log.Int("offset", offset),
			log.Int("totalResults", totalResults), log.Int("count", count))
	}
}

// HandleOUChildrenListByPathRequest handles the list child organization units by path request.
func (ouh *organizationUnitHandler) HandleOUChildrenListByPathRequest(w http.ResponseWriter, r *http.Request) {
	ouh.handleResourceListByPathRequest(w, r, "child organization units",
		func(path string, limit, offset int) (interface{}, *serviceerror.ServiceError) {
			return ouh.service.GetOrganizationUnitChildrenByPath(r.Context(), path, limit, offset)
		})
}

// HandleOUUsersListByPathRequest handles the list users in organization unit by path request.
func (ouh *organizationUnitHandler) HandleOUUsersListByPathRequest(w http.ResponseWriter, r *http.Request) {
	ouh.handleResourceListByPathRequest(w, r, "users",
		func(path string, limit, offset int) (interface{}, *serviceerror.ServiceError) {
			return ouh.service.GetOrganizationUnitUsersByPath(path, limit, offset)
		})
}

// HandleOUGroupsListByPathRequest handles the list groups in organization unit by path request.
func (ouh *organizationUnitHandler) HandleOUGroupsListByPathRequest(w http.ResponseWriter, r *http.Request) {
	ouh.handleResourceListByPathRequest(w, r, "groups",
		func(path string, limit, offset int) (interface{}, *serviceerror.ServiceError) {
			return ouh.service.GetOrganizationUnitGroupsByPath(path, limit, offset)
		})
}

func extractAndValidatePath(w http.ResponseWriter, r *http.Request) (string, bool) {
	path := r.PathValue("path")
	if path == "" {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, apierror.ErrorResponse{
			Code:        ErrorInvalidHandlePath.Code,
			Message:     ErrorInvalidHandlePath.Error,
			Description: ErrorInvalidHandlePath.ErrorDescription,
		})
		return "", true
	}
	return path, false
}
