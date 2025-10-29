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
	"fmt"
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
		handleError(w, logger, svcErr)
		return
	}

	roleList, svcErr := rh.roleService.GetRoleList(limit, offset)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	// Convert service response to HTTP response
	roles := make([]RoleSummaryResponse, 0, len(roleList.Roles))
	for _, role := range roleList.Roles {
		roles = append(roles, RoleSummaryResponse(role))
	}

	roleListResponse := &RoleListResponse{
		TotalResults: roleList.TotalResults,
		StartIndex:   roleList.StartIndex,
		Count:        roleList.Count,
		Roles:        roles,
		Links:        toHTTPLinks(roleList.Links),
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	isErr := writeToResponse(w, roleListResponse, logger)
	if isErr {
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
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitizedRequest := rh.sanitizeCreateRoleRequest(createRequest)

	// Convert HTTP request to service request
	serviceRequest := rh.toRoleCreationDetail(sanitizedRequest)

	serviceRole, svcErr := rh.roleService.CreateRole(serviceRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	// Convert service response to HTTP response
	createdRole := rh.toHTTPCreateRoleResponse(serviceRole)

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	isErr := writeToResponse(w, createdRole, logger)
	if isErr {
		return
	}

	logger.Debug("Successfully created role", log.String("roleId", createdRole.ID))
}

// HandleRoleGetRequest handles the get role by id request.
func (rh *roleHandler) HandleRoleGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	serviceRole, svcErr := rh.roleService.GetRoleWithPermissions(id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	// Convert service response to HTTP response
	role := rh.toHTTPRoleResponse(serviceRole)

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	isErr := writeToResponse(w, role, logger)
	if isErr {
		return
	}

	logger.Debug("Successfully retrieved role", log.String("role id", id))
}

// HandleRolePutRequest handles the update role request.
func (rh *roleHandler) HandleRolePutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	updateRequest, err := sysutils.DecodeJSONBody[UpdateRoleRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitizedRequest := rh.sanitizeUpdateRoleRequest(updateRequest)

	// Convert HTTP request to service request
	serviceRequest := RoleUpdateDetail(sanitizedRequest)

	serviceRole, svcErr := rh.roleService.UpdateRoleWithPermissions(id, serviceRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	// Convert service response to HTTP response
	role := rh.toHTTPRoleResponse(serviceRole)

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	isErr := writeToResponse(w, role, logger)
	if isErr {
		return
	}

	logger.Debug("Successfully updated role", log.String("role id", id))
}

// HandleRoleDeleteRequest handles the delete role request.
func (rh *roleHandler) HandleRoleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	svcErr := rh.roleService.DeleteRole(id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Debug("Successfully deleted role", log.String("role id", id))
}

// HandleRoleAssignmentsGetRequest handles the get role assignments request.
func (rh *roleHandler) HandleRoleAssignmentsGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	// Parse include parameter to check if display names should be included
	includeDisplay := r.URL.Query().Get("include") == "display"

	serviceResponse, svcErr := rh.roleService.GetRoleAssignments(id, limit, offset, includeDisplay)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	// Convert service response to HTTP response
	httpAssignments := make([]AssignmentResponse, len(serviceResponse.Assignments))
	for i, sa := range serviceResponse.Assignments {
		httpAssignments[i] = AssignmentResponse(sa)
	}

	assignmentListResponse := &AssignmentListResponse{
		TotalResults: serviceResponse.TotalResults,
		StartIndex:   serviceResponse.StartIndex,
		Count:        serviceResponse.Count,
		Assignments:  httpAssignments,
		Links:        toHTTPLinks(serviceResponse.Links),
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	isErr := writeToResponse(w, assignmentListResponse, logger)
	if isErr {
		return
	}

	logger.Debug("Successfully retrieved role assignments", log.String("role id", id),
		log.Int("limit", limit), log.Int("offset", offset),
		log.Bool("includeDisplay", includeDisplay),
		log.Int("totalResults", assignmentListResponse.TotalResults),
		log.Int("count", assignmentListResponse.Count))
}

// HandleRoleAddAssignmentsRequest handles the add assignments to role request.
func (rh *roleHandler) HandleRoleAddAssignmentsRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	assignmentsRequest, err := sysutils.DecodeJSONBody[AssignmentsRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitizedRequest := rh.sanitizeAssignmentsRequest(assignmentsRequest)

	// Convert HTTP request to service request
	serviceRequest := rh.toRoleAssignments(sanitizedRequest)

	svcErr := rh.roleService.AddAssignments(id, serviceRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Debug("Successfully added assignments to role", log.String("role id", id))
}

// HandleRoleRemoveAssignmentsRequest handles the remove assignments from role request.
func (rh *roleHandler) HandleRoleRemoveAssignmentsRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	assignmentsRequest, err := sysutils.DecodeJSONBody[AssignmentsRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitizedRequest := rh.sanitizeAssignmentsRequest(assignmentsRequest)

	// Convert HTTP request to service request
	serviceRequest := rh.toRoleAssignments(sanitizedRequest)

	svcErr := rh.roleService.RemoveAssignments(id, serviceRequest)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Debug("Successfully removed assignments from role", log.String("role id", id))
}

func writeToResponse(w http.ResponseWriter, response any, logger *log.Logger) bool {
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		handleEncodingError(w)
		return true
	}
	return false
}

// handleError handles service errors and returns appropriate HTTP responses.
func handleError(w http.ResponseWriter, logger *log.Logger,
	svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case ErrorRoleNotFound.Code:
			statusCode = http.StatusNotFound
		case ErrorRoleNameConflict.Code:
			statusCode = http.StatusConflict
		case ErrorOrganizationUnitNotFound.Code, ErrorCannotDeleteRole.Code,
			ErrorInvalidRequestFormat.Code, ErrorMissingRoleID.Code,
			ErrorInvalidLimit.Code, ErrorInvalidOffset.Code,
			ErrorEmptyAssignments.Code,
			ErrorInvalidAssignmentID.Code:
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusBadRequest
		}
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
		handleEncodingError(w)
		return
	}
}

// handleEncodingError handles errors that occur during response encoding.
func handleEncodingError(w http.ResponseWriter) {
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintln(w, serviceerror.ErrorEncodingError)
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
		sanitized.Assignments = make([]AssignmentRequest, len(request.Assignments))
		for i, assignment := range request.Assignments {
			sanitized.Assignments[i] = AssignmentRequest{
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

	return sanitized
}

// sanitizeAssignmentsRequest sanitizes the assignments request input.
func (rh *roleHandler) sanitizeAssignmentsRequest(request *AssignmentsRequest) AssignmentsRequest {
	sanitized := AssignmentsRequest{}

	if request.Assignments != nil {
		sanitized.Assignments = make([]AssignmentRequest, len(request.Assignments))
		for i, assignment := range request.Assignments {
			sanitized.Assignments[i] = AssignmentRequest{
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

// Conversion helper methods between HTTP and service layer structs

// toRoleCreationDetail converts HTTP CreateRoleRequest to service layer RoleCreationDetail.
func (rh *roleHandler) toRoleCreationDetail(req CreateRoleRequest) RoleCreationDetail {
	serviceAssignments := make([]RoleAssignment, len(req.Assignments))
	for i, a := range req.Assignments {
		serviceAssignments[i] = RoleAssignment(a)
	}

	return RoleCreationDetail{
		Name:               req.Name,
		Description:        req.Description,
		OrganizationUnitID: req.OrganizationUnitID,
		Permissions:        req.Permissions,
		Assignments:        serviceAssignments,
	}
}

// toHTTPRole converts service layer RoleWithPermissions to HTTP Role.
func (rh *roleHandler) toHTTPRoleResponse(role *RoleWithPermissions) *RoleResponse {
	return &RoleResponse{
		ID:                 role.ID,
		Name:               role.Name,
		Description:        role.Description,
		OrganizationUnitID: role.OrganizationUnitID,
		Permissions:        role.Permissions,
	}
}

// toHTTPCreateRoleResponse converts service layer RoleDetails to HTTP CreateRoleResponse.
func (rh *roleHandler) toHTTPCreateRoleResponse(role *RoleWithPermissionsAndAssignments) *CreateRoleResponse {
	httpAssignments := make([]AssignmentResponse, len(role.Assignments))
	for i, sa := range role.Assignments {
		httpAssignments[i] = AssignmentResponse{
			ID:   sa.ID,
			Type: sa.Type,
		}
	}

	return &CreateRoleResponse{
		ID:                 role.ID,
		Name:               role.Name,
		Description:        role.Description,
		OrganizationUnitID: role.OrganizationUnitID,
		Permissions:        role.Permissions,
		Assignments:        httpAssignments,
	}
}

func toHTTPLinks(links []Link) []LinkResponse {
	httpLinks := make([]LinkResponse, len(links))
	for i, link := range links {
		httpLinks[i] = LinkResponse(link)
	}
	return httpLinks
}

// toRoleAssignments converts HTTP AssignmentsRequest to service layer RoleAssignments.
func (rh *roleHandler) toRoleAssignments(req AssignmentsRequest) []RoleAssignment {
	serviceAssignments := make([]RoleAssignment, len(req.Assignments))
	for i, a := range req.Assignments {
		serviceAssignments[i] = RoleAssignment(a)
	}
	return serviceAssignments
}
