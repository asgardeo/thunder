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

package group

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

const handlerLoggerComponentName = "GroupHandler"

// groupHandler is the handler for group management operations.
type groupHandler struct {
	groupService GroupServiceInterface
}

// newGroupHandler creates a new instance of groupHandler
func newGroupHandler(groupService GroupServiceInterface) *groupHandler {
	return &groupHandler{
		groupService: groupService,
	}
}

// HandleGroupListRequest handles the list groups request.
func (gh *groupHandler) HandleGroupListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	groupListResponse, svcErr := gh.groupService.GetGroupList(limit, offset)
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(groupListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully listed groups with pagination",
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", groupListResponse.TotalResults),
		log.Int("count", groupListResponse.Count))
}

// HandleGroupListByPathRequest handles the list groups by OU path request.
func (gh *groupHandler) HandleGroupListByPathRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	path, pathValidationFailed := extractAndValidatePath(w, r, logger)
	if pathValidationFailed {
		return
	}

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	groupListResponse, svcErr := gh.groupService.GetGroupsByPath(path, limit, offset)
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(groupListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully listed groups by path", log.String("path", path),
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", groupListResponse.TotalResults),
		log.Int("count", groupListResponse.Count))
}

// HandleGroupPostRequest handles the create group request.
func (gh *groupHandler) HandleGroupPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	createRequest, err := sysutils.DecodeJSONBody[CreateGroupRequest](r)
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

	sanitizedRequest := gh.sanitizeCreateGroupRequest(createRequest)
	createdGroup, svcErr := gh.groupService.CreateGroup(sanitizedRequest)
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdGroup); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully created group", log.String("group id", createdGroup.ID))
}

// HandleGroupPostByPathRequest handles the create group by OU path request.
func (gh *groupHandler) HandleGroupPostByPathRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	path, pathValidationFailed := extractAndValidatePath(w, r, logger)
	if pathValidationFailed {
		return
	}

	createRequest, err := sysutils.DecodeJSONBody[CreateGroupByPathRequest](r)
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

	group, svcErr := gh.groupService.CreateGroupByPath(path, *createRequest)
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(group); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully created group by path", log.String("path", path), log.String("groupName", group.Name))
}

// HandleGroupGetRequest handles the get group by id request.
func (gh *groupHandler) HandleGroupGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorMissingGroupID.Code,
			Message:     ErrorMissingGroupID.Error,
			Description: ErrorMissingGroupID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	group, svcErr := gh.groupService.GetGroup(id)
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(group); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully retrieved group", log.String("group id", id))
}

// HandleGroupPutRequest handles the update group request.
func (gh *groupHandler) HandleGroupPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorMissingGroupID.Code,
			Message:     ErrorMissingGroupID.Error,
			Description: ErrorMissingGroupID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	updateRequest, err := sysutils.DecodeJSONBody[UpdateGroupRequest](r)
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

	sanitizedRequest := gh.sanitizeUpdateGroupRequest(updateRequest)
	group, svcErr := gh.groupService.UpdateGroup(id, sanitizedRequest)
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(group); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully updated group", log.String("group id", id))
}

// HandleGroupDeleteRequest handles the delete group request.
func (gh *groupHandler) HandleGroupDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorMissingGroupID.Code,
			Message:     ErrorMissingGroupID.Error,
			Description: ErrorMissingGroupID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	svcErr := gh.groupService.DeleteGroup(id)
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Debug("Successfully deleted group", log.String("group id", id))
}

// HandleGroupMembersGetRequest handles the get group members request.
func (gh *groupHandler) HandleGroupMembersGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorMissingGroupID.Code,
			Message:     ErrorMissingGroupID.Error,
			Description: ErrorMissingGroupID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	memberListResponse, svcErr := gh.groupService.GetGroupMembers(id, limit, offset)
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(memberListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully retrieved group members", log.String("group id", id),
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", memberListResponse.TotalResults),
		log.Int("count", memberListResponse.Count))
}

// handleError handles service errors and returns appropriate HTTP responses.
func (gh *groupHandler) handleError(w http.ResponseWriter, logger *log.Logger,
	svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case ErrorGroupNotFound.Code:
			statusCode = http.StatusNotFound
		case ErrorGroupNameConflict.Code:
			statusCode = http.StatusConflict
		case ErrorInvalidOUID.Code, ErrorCannotDeleteGroup.Code,
			ErrorInvalidRequestFormat.Code, ErrorMissingGroupID.Code,
			ErrorInvalidLimit.Code, ErrorInvalidOffset.Code:
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

// sanitizeCreateGroupRequest sanitizes the create group request input.
func (gh *groupHandler) sanitizeCreateGroupRequest(request *CreateGroupRequest) CreateGroupRequest {
	sanitized := CreateGroupRequest{
		Name:               sysutils.SanitizeString(request.Name),
		Description:        sysutils.SanitizeString(request.Description),
		OrganizationUnitID: sysutils.SanitizeString(request.OrganizationUnitID),
	}

	if request.Members != nil {
		sanitized.Members = make([]Member, len(request.Members))
		for i, member := range request.Members {
			sanitized.Members[i] = Member{
				ID:   sysutils.SanitizeString(member.ID),
				Type: member.Type,
			}
		}
	}

	return sanitized
}

// sanitizeUpdateGroupRequest sanitizes the update group request input.
func (gh *groupHandler) sanitizeUpdateGroupRequest(request *UpdateGroupRequest) UpdateGroupRequest {
	sanitized := UpdateGroupRequest{
		Name:               sysutils.SanitizeString(request.Name),
		Description:        sysutils.SanitizeString(request.Description),
		OrganizationUnitID: sysutils.SanitizeString(request.OrganizationUnitID),
	}

	if request.Members != nil {
		sanitized.Members = make([]Member, len(request.Members))
		for i, member := range request.Members {
			sanitized.Members[i] = Member{
				ID:   sysutils.SanitizeString(member.ID),
				Type: member.Type,
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

// extractAndValidatePath extracts and validates the path parameter from the request.
func extractAndValidatePath(w http.ResponseWriter, r *http.Request, logger *log.Logger) (string, bool) {
	path := r.PathValue("path")
	if path == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: "Handle path is required",
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return "", true
	}
	return path, false
}
