/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package handler provides the implementation for group management operations.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/asgardeo/thunder/internal/group/constants"
	"github.com/asgardeo/thunder/internal/group/model"
	"github.com/asgardeo/thunder/internal/group/provider"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "GroupHandler"

// GroupHandler is the handler for group management operations.
type GroupHandler struct{}

// NewGroupHandler creates a new instance of GroupHandler
func NewGroupHandler() *GroupHandler {
	return &GroupHandler{}
}

// HandleGroupListRequest handles the list groups request.
func (gh *GroupHandler) HandleGroupListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	groups, svcErr := groupService.GetGroupList()
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(groups); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully listed groups")
}

// HandleGroupPostRequest handles the create group request.
func (gh *GroupHandler) HandleGroupPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	createRequest, err := sysutils.DecodeJSONBody[model.CreateGroupRequest](r)
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

	sanitizedRequest := gh.sanitizeCreateGroupRequest(createRequest)

	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	createdGroup, svcErr := groupService.CreateGroup(sanitizedRequest)
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

// HandleGroupGetRequest handles the get group by id request.
func (gh *GroupHandler) HandleGroupGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorMissingGroupID.Code,
			Message:     constants.ErrorMissingGroupID.Error,
			Description: constants.ErrorMissingGroupID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	group, svcErr := groupService.GetGroup(id)
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
func (gh *GroupHandler) HandleGroupPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorMissingGroupID.Code,
			Message:     constants.ErrorMissingGroupID.Error,
			Description: constants.ErrorMissingGroupID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	updateRequest, err := sysutils.DecodeJSONBody[model.UpdateGroupRequest](r)
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

	sanitizedRequest := gh.sanitizeUpdateGroupRequest(updateRequest)

	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	group, svcErr := groupService.UpdateGroup(id, sanitizedRequest)
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
func (gh *GroupHandler) HandleGroupDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorMissingGroupID.Code,
			Message:     constants.ErrorMissingGroupID.Error,
			Description: constants.ErrorMissingGroupID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	svcErr := groupService.DeleteGroup(id)
	if svcErr != nil {
		gh.handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Debug("Successfully deleted group", log.String("group id", id))
}

// handleError handles service errors and returns appropriate HTTP responses.
func (gh *GroupHandler) handleError(w http.ResponseWriter, logger *log.Logger,
	svcErr *serviceerror.ServiceError) {
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)

	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case constants.ErrorGroupNotFound.Code:
			statusCode = http.StatusNotFound
		case constants.ErrorGroupNameConflict.Code:
			statusCode = http.StatusConflict
		case constants.ErrorParentNotFound.Code, constants.ErrorCannotDeleteGroup.Code,
			constants.ErrorInvalidRequestFormat.Code, constants.ErrorMissingGroupID.Code:
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusBadRequest
		}
	}
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Error encoding error response", log.Error(err))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// sanitizeCreateGroupRequest sanitizes the create group request input.
func (gh *GroupHandler) sanitizeCreateGroupRequest(request *model.CreateGroupRequest) model.CreateGroupRequest {
	sanitized := model.CreateGroupRequest{
		Name:        sysutils.SanitizeString(request.Name),
		Description: sysutils.SanitizeString(request.Description),
		Parent: model.Parent{
			Type: request.Parent.Type,
			ID:   sysutils.SanitizeString(request.Parent.ID),
		},
	}

	if request.Users != nil {
		sanitized.Users = make([]string, len(request.Users))
		for i, user := range request.Users {
			sanitized.Users[i] = sysutils.SanitizeString(user)
		}
	}

	return sanitized
}

// sanitizeUpdateGroupRequest sanitizes the update group request input.
func (gh *GroupHandler) sanitizeUpdateGroupRequest(request *model.UpdateGroupRequest) model.UpdateGroupRequest {
	sanitized := model.UpdateGroupRequest{
		Name:        sysutils.SanitizeString(request.Name),
		Description: sysutils.SanitizeString(request.Description),
		Parent: model.Parent{
			Type: request.Parent.Type,
			ID:   sysutils.SanitizeString(request.Parent.ID),
		},
	}

	if request.Users != nil {
		sanitized.Users = make([]string, len(request.Users))
		for i, user := range request.Users {
			sanitized.Users[i] = sysutils.SanitizeString(user)
		}
	}

	if request.Groups != nil {
		sanitized.Groups = make([]string, len(request.Groups))
		for i, group := range request.Groups {
			sanitized.Groups[i] = sysutils.SanitizeString(group)
		}
	}

	return sanitized
}
