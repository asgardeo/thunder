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

package role

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

const handlerLoggerComponentName = "RoleHandler"

// roleHandler is the handler for role management operations.
type roleHandler struct {
	roleService RoleServiceInterface
}

// newRoleHandler creates a new instance of roleHandler
func newRoleHandler(roleService RoleServiceInterface) *roleHandler {
	return &roleHandler{
		roleService: roleService,
	}
}

// HandleRoleListRequest handles the list roles request.
func (rh *roleHandler) HandleRoleListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		rh.handleError(w, logger, svcErr)
		return
	}

	roleListResponse, svcErr := rh.roleService.GetRoleList(limit, offset)
	if svcErr != nil {
		rh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(roleListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully listed roles with pagination",
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", roleListResponse.TotalResults),
		log.Int("count", roleListResponse.Count))
}

// HandleRolePostRequest handles the create role request.
func (rh *roleHandler) HandleRolePostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	createRequest, err := sysutils.DecodeJSONBody[CreateRoleRequest](r)
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

	sanitizedRequest := rh.sanitizeCreateRoleRequest(createRequest)
	createdRole, svcErr := rh.roleService.CreateRole(sanitizedRequest)
	if svcErr != nil {
		rh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdRole); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully created role", log.String("role id", createdRole.ID))
}

// HandleRoleGetRequest handles the get role by id request.
func (rh *roleHandler) HandleRoleGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorMissingRoleID.Code,
			Message:     ErrorMissingRoleID.Error,
			Description: ErrorMissingRoleID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	role, svcErr := rh.roleService.GetRole(id)
	if svcErr != nil {
		rh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(role); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully retrieved role", log.String("role id", id))
}

// HandleRolePutRequest handles the update role request.
func (rh *roleHandler) HandleRolePutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorMissingRoleID.Code,
			Message:     ErrorMissingRoleID.Error,
			Description: ErrorMissingRoleID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	updateRequest, err := sysutils.DecodeJSONBody[UpdateRoleRequest](r)
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

	sanitizedRequest := rh.sanitizeUpdateRoleRequest(updateRequest)
	role, svcErr := rh.roleService.UpdateRole(id, sanitizedRequest)
	if svcErr != nil {
		rh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(role); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully updated role", log.String("role id", id))
}

// HandleRoleDeleteRequest handles the delete role request.
func (rh *roleHandler) HandleRoleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorMissingRoleID.Code,
			Message:     ErrorMissingRoleID.Error,
			Description: ErrorMissingRoleID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	svcErr := rh.roleService.DeleteRole(id)
	if svcErr != nil {
		rh.handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Debug("Successfully deleted role", log.String("role id", id))
}

// HandleRoleAssignmentsGetRequest handles the get role assignments request.
func (rh *roleHandler) HandleRoleAssignmentsGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorMissingRoleID.Code,
			Message:     ErrorMissingRoleID.Error,
			Description: ErrorMissingRoleID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		rh.handleError(w, logger, svcErr)
		return
	}

	assignmentListResponse, svcErr := rh.roleService.GetRoleAssignments(id, limit, offset)
	if svcErr != nil {
		rh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(assignmentListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully retrieved role assignments", log.String("role id", id),
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", assignmentListResponse.TotalResults),
		log.Int("count", assignmentListResponse.Count))
}

// HandleRoleAddAssignmentsRequest handles the add assignments to role request.
func (rh *roleHandler) HandleRoleAddAssignmentsRequest(w http.ResponseWriter, r *http.Request) {
	rh.handleAssignmentsRequest(w, r, true)
}

// HandleRoleRemoveAssignmentsRequest handles the remove assignments from role request.
func (rh *roleHandler) HandleRoleRemoveAssignmentsRequest(w http.ResponseWriter, r *http.Request) {
	rh.handleAssignmentsRequest(w, r, false)
}

// handleAssignmentsRequest is a helper function for add/remove assignments operations.
func (rh *roleHandler) handleAssignmentsRequest(w http.ResponseWriter, r *http.Request, isAdd bool) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorMissingRoleID.Code,
			Message:     ErrorMissingRoleID.Error,
			Description: ErrorMissingRoleID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	assignmentsRequest, err := sysutils.DecodeJSONBody[AssignmentsRequest](r)
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

	sanitizedRequest := rh.sanitizeAssignmentsRequest(assignmentsRequest)

	var svcErr *serviceerror.ServiceError
	var successMsg string
	if isAdd {
		svcErr = rh.roleService.AddAssignments(id, sanitizedRequest)
		successMsg = "Successfully added assignments to role"
	} else {
		svcErr = rh.roleService.RemoveAssignments(id, sanitizedRequest)
		successMsg = "Successfully removed assignments from role"
	}

	if svcErr != nil {
		rh.handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Debug(successMsg, log.String("role id", id))
}

// handleError handles service errors and returns appropriate HTTP responses.
func (rh *roleHandler) handleError(w http.ResponseWriter, logger *log.Logger,
	svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case ErrorRoleNotFound.Code:
			statusCode = http.StatusNotFound
		case ErrorOrganizationUnitNotFound.Code, ErrorCannotDeleteRole.Code,
			ErrorInvalidRequestFormat.Code, ErrorMissingRoleID.Code,
			ErrorInvalidLimit.Code, ErrorInvalidOffset.Code,
			ErrorEmptyPermissions.Code, ErrorEmptyAssignments.Code,
			ErrorInvalidAssignmentID.Code:
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusBadRequest
		}
	}

	if statusCode == http.StatusInternalServerError {
		logger.Error("Internal server error occurred", log.String("error", svcErr.Error),
			log.String("description", svcErr.ErrorDescription))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Error encoding error response", log.Error(err))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// sanitizeCreateRoleRequest sanitizes the create role request input.
func (rh *roleHandler) sanitizeCreateRoleRequest(request *CreateRoleRequest) CreateRoleRequest {
	sanitized := CreateRoleRequest{
		Name:               sysutils.SanitizeString(request.Name),
		Description:        sysutils.SanitizeString(request.Description),
		OrganizationUnitID: sysutils.SanitizeString(request.OrganizationUnitID),
	}

	if request.Permissions != nil {
		sanitized.Permissions = make([]string, len(request.Permissions))
		for i, permission := range request.Permissions {
			sanitized.Permissions[i] = sysutils.SanitizeString(permission)
		}
	}

	if request.Assignments != nil {
		sanitized.Assignments = make([]Assignment, len(request.Assignments))
		for i, assignment := range request.Assignments {
			sanitized.Assignments[i] = Assignment{
				ID:   sysutils.SanitizeString(assignment.ID),
				Type: assignment.Type,
			}
		}
	}

	return sanitized
}

// sanitizeUpdateRoleRequest sanitizes the update role request input.
func (rh *roleHandler) sanitizeUpdateRoleRequest(request *UpdateRoleRequest) UpdateRoleRequest {
	sanitized := UpdateRoleRequest{
		Name:               sysutils.SanitizeString(request.Name),
		Description:        sysutils.SanitizeString(request.Description),
		OrganizationUnitID: sysutils.SanitizeString(request.OrganizationUnitID),
	}

	if request.Permissions != nil {
		sanitized.Permissions = make([]string, len(request.Permissions))
		for i, permission := range request.Permissions {
			sanitized.Permissions[i] = sysutils.SanitizeString(permission)
		}
	}

	if request.Assignments != nil {
		sanitized.Assignments = make([]Assignment, len(request.Assignments))
		for i, assignment := range request.Assignments {
			sanitized.Assignments[i] = Assignment{
				ID:   sysutils.SanitizeString(assignment.ID),
				Type: assignment.Type,
			}
		}
	}

	return sanitized
}

// sanitizeAssignmentsRequest sanitizes the assignments request input.
func (rh *roleHandler) sanitizeAssignmentsRequest(request *AssignmentsRequest) AssignmentsRequest {
	sanitized := AssignmentsRequest{}

	if request.Assignments != nil {
		sanitized.Assignments = make([]Assignment, len(request.Assignments))
		for i, assignment := range request.Assignments {
			sanitized.Assignments[i] = Assignment{
				ID:   sysutils.SanitizeString(assignment.ID),
				Type: assignment.Type,
			}
		}
	}

	return sanitized
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

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	return limit, offset, nil
}
