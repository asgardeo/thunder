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

package resource

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

const handlerLoggerComponentName = "ResourceHandler"

// resourceHandler handles HTTP requests for resource management.
type resourceHandler struct {
	resourceService ResourceServiceInterface
}

// newResourceHandler creates a new resource handler.
func newResourceHandler(resourceService ResourceServiceInterface) *resourceHandler {
	return &resourceHandler{
		resourceService: resourceService,
	}
}

// Resource Server Handlers

// HandleResourceServerListRequest handles listing resource servers.
func (h *resourceHandler) HandleResourceServerListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	result, svcErr := h.resourceService.GetResourceServerList(limit, offset)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toResourceServerListResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleResourceServerPostRequest handles creating a resource server.
func (h *resourceHandler) HandleResourceServerPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	req, err := sysutils.DecodeJSONBody[CreateResourceServerRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitized := sanitizeCreateResourceServerRequest(req)
	serviceReq := ResourceServer{
		Name:               sanitized.Name,
		Description:        sanitized.Description,
		Identifier:         sanitized.Identifier,
		OrganizationUnitID: sanitized.OrganizationUnitID,
	}

	result, svcErr := h.resourceService.CreateResourceServer(serviceReq)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toResourceServerResponse(result)
	writeJSONResponse(w, http.StatusCreated, response, logger)
}

// HandleResourceServerGetRequest handles getting a resource server.
func (h *resourceHandler) HandleResourceServerGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	result, svcErr := h.resourceService.GetResourceServer(id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toResourceServerResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleResourceServerPutRequest handles updating a resource server.
func (h *resourceHandler) HandleResourceServerPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	req, err := sysutils.DecodeJSONBody[UpdateResourceServerRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitized := sanitizeUpdateResourceServerRequest(req)
	serviceReq := ResourceServer{
		Name:               sanitized.Name,
		Description:        sanitized.Description,
		Identifier:         sanitized.Identifier,
		OrganizationUnitID: sanitized.OrganizationUnitID,
	}

	result, svcErr := h.resourceService.UpdateResourceServer(id, serviceReq)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toResourceServerResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleResourceServerDeleteRequest handles deleting a resource server.
func (h *resourceHandler) HandleResourceServerDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	id := r.PathValue("id")
	svcErr := h.resourceService.DeleteResourceServer(id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Resource Handlers

// HandleResourceListRequest handles listing resources.
func (h *resourceHandler) HandleResourceListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	// Parse parentId parameter (can be empty string for top-level, or UUID for children)
	// If parentId not in query, parentID remains nil (all resources)
	var parentID *string
	if r.URL.Query().Has("parentId") {
		parentParam := r.URL.Query().Get("parentId")
		parentID = &parentParam
	}

	result, svcErr := h.resourceService.GetResourceList(rsID, parentID, limit, offset)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toResourceListResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleResourcePostRequest handles creating a resource.
func (h *resourceHandler) HandleResourcePostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	req, err := sysutils.DecodeJSONBody[CreateResourceRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitized := sanitizeCreateResourceRequest(req)
	serviceReq := Resource{
		Name:        sanitized.Name,
		Handle:      sanitized.Handle,
		Description: sanitized.Description,
		Parent:      sanitized.Parent,
	}

	result, svcErr := h.resourceService.CreateResource(rsID, serviceReq)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toResourceResponse(result)
	writeJSONResponse(w, http.StatusCreated, response, logger)
}

// HandleResourceGetRequest handles getting a resource.
func (h *resourceHandler) HandleResourceGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	id := r.PathValue("id")

	result, svcErr := h.resourceService.GetResource(rsID, id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toResourceResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleResourcePutRequest handles updating a resource.
func (h *resourceHandler) HandleResourcePutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	id := r.PathValue("id")

	req, err := sysutils.DecodeJSONBody[UpdateResourceRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitized := sanitizeUpdateResourceRequest(req)
	serviceReq := Resource{
		Name:        sanitized.Name,
		Description: sanitized.Description,
	}

	result, svcErr := h.resourceService.UpdateResource(rsID, id, serviceReq)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toResourceResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleResourceDeleteRequest handles deleting a resource.
func (h *resourceHandler) HandleResourceDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	id := r.PathValue("id")

	svcErr := h.resourceService.DeleteResource(rsID, id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Action Handlers (Resource Server Level)

// HandleActionListAtResourceServerRequest handles listing actions at resource server level.
func (h *resourceHandler) HandleActionListAtResourceServerRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	result, svcErr := h.resourceService.GetActionListAtResourceServer(rsID, limit, offset)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toActionListResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleActionPostAtResourceServerRequest handles creating an action at resource server level.
func (h *resourceHandler) HandleActionPostAtResourceServerRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	req, err := sysutils.DecodeJSONBody[CreateActionRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitized := sanitizeCreateActionRequest(req)
	serviceReq := Action{
		Name:        sanitized.Name,
		Handle:      sanitized.Handle,
		Description: sanitized.Description,
	}

	result, svcErr := h.resourceService.CreateActionAtResourceServer(rsID, serviceReq)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toActionResponse(result)
	writeJSONResponse(w, http.StatusCreated, response, logger)
}

// HandleActionGetAtResourceServerRequest handles getting an action at resource server level.
func (h *resourceHandler) HandleActionGetAtResourceServerRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	id := r.PathValue("id")

	result, svcErr := h.resourceService.GetActionAtResourceServer(rsID, id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toActionResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleActionPutAtResourceServerRequest handles updating an action at resource server level.
func (h *resourceHandler) HandleActionPutAtResourceServerRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	id := r.PathValue("id")

	req, err := sysutils.DecodeJSONBody[UpdateActionRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitized := sanitizeUpdateActionRequest(req)
	serviceReq := Action{
		Name:        sanitized.Name,
		Description: sanitized.Description,
	}

	result, svcErr := h.resourceService.UpdateActionAtResourceServer(rsID, id, serviceReq)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toActionResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleActionDeleteAtResourceServerRequest handles deleting an action at resource server level.
func (h *resourceHandler) HandleActionDeleteAtResourceServerRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	id := r.PathValue("id")

	svcErr := h.resourceService.DeleteActionAtResourceServer(rsID, id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Action Handlers (Resource Level)

// HandleActionListAtResourceRequest handles listing actions at resource level.
func (h *resourceHandler) HandleActionListAtResourceRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	resourceID := r.PathValue("resourceId")
	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	result, svcErr := h.resourceService.GetActionListAtResource(rsID, resourceID, limit, offset)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toActionListResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleActionPostAtResourceRequest handles creating an action at resource level.
func (h *resourceHandler) HandleActionPostAtResourceRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	resourceID := r.PathValue("resourceId")

	req, err := sysutils.DecodeJSONBody[CreateActionRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitized := sanitizeCreateActionRequest(req)
	serviceReq := Action{
		Name:        sanitized.Name,
		Handle:      sanitized.Handle,
		Description: sanitized.Description,
	}

	result, svcErr := h.resourceService.CreateActionAtResource(rsID, resourceID, serviceReq)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toActionResponse(result)
	writeJSONResponse(w, http.StatusCreated, response, logger)
}

// HandleActionGetAtResourceRequest handles getting an action at resource level.
func (h *resourceHandler) HandleActionGetAtResourceRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	resourceID := r.PathValue("resourceId")
	id := r.PathValue("id")

	result, svcErr := h.resourceService.GetActionAtResource(rsID, resourceID, id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toActionResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleActionPutAtResourceRequest handles updating an action at resource level.
func (h *resourceHandler) HandleActionPutAtResourceRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	resourceID := r.PathValue("resourceId")
	id := r.PathValue("id")

	req, err := sysutils.DecodeJSONBody[UpdateActionRequest](r)
	if err != nil {
		handleError(w, logger, &ErrorInvalidRequestFormat)
		return
	}

	sanitized := sanitizeUpdateActionRequest(req)
	serviceReq := Action{
		Name:        sanitized.Name,
		Description: sanitized.Description,
	}

	result, svcErr := h.resourceService.UpdateActionAtResource(rsID, resourceID, id, serviceReq)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	response := toActionResponse(result)
	writeJSONResponse(w, http.StatusOK, response, logger)
}

// HandleActionDeleteAtResourceRequest handles deleting an action at resource level.
func (h *resourceHandler) HandleActionDeleteAtResourceRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	rsID := r.PathValue("rsId")
	resourceID := r.PathValue("resourceId")
	id := r.PathValue("id")

	svcErr := h.resourceService.DeleteActionAtResource(rsID, resourceID, id)
	if svcErr != nil {
		handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func parsePaginationParams(query url.Values) (int, int, *serviceerror.ServiceError) {
	limit := serverconst.DefaultPageSize
	offset := 0

	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 {
			return 0, 0, &ErrorInvalidLimit
		}
		limit = parsedLimit
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			return 0, 0, &ErrorInvalidOffset
		}
		offset = parsedOffset
	}

	return limit, offset, nil
}

func handleError(w http.ResponseWriter, logger *log.Logger, svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case ErrorResourceServerNotFound.Code, ErrorResourceNotFound.Code, ErrorActionNotFound.Code:
			statusCode = http.StatusNotFound
		case ErrorNameConflict.Code, ErrorHandleConflict.Code, ErrorIdentifierConflict.Code:
			statusCode = http.StatusConflict
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
	}
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}, logger *log.Logger) {
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Error encoding response", log.Error(err))
	}
}

// Sanitization functions

func sanitizeCreateResourceServerRequest(req *CreateResourceServerRequest) CreateResourceServerRequest {
	return CreateResourceServerRequest{
		Name:               sysutils.SanitizeString(req.Name),
		Description:        sysutils.SanitizeString(req.Description),
		Identifier:         sysutils.SanitizeString(req.Identifier),
		OrganizationUnitID: sysutils.SanitizeString(req.OrganizationUnitID),
	}
}

func sanitizeUpdateResourceServerRequest(req *UpdateResourceServerRequest) UpdateResourceServerRequest {
	return UpdateResourceServerRequest{
		Name:               sysutils.SanitizeString(req.Name),
		Description:        sysutils.SanitizeString(req.Description),
		Identifier:         sysutils.SanitizeString(req.Identifier),
		OrganizationUnitID: sysutils.SanitizeString(req.OrganizationUnitID),
	}
}

func sanitizeCreateResourceRequest(req *CreateResourceRequest) CreateResourceRequest {
	sanitized := CreateResourceRequest{
		Name:        sysutils.SanitizeString(req.Name),
		Handle:      sysutils.SanitizeString(req.Handle),
		Description: sysutils.SanitizeString(req.Description),
		Parent:      nil,
	}

	if req.Parent != nil {
		sanitizedParent := sysutils.SanitizeString(*req.Parent)
		sanitized.Parent = &sanitizedParent
	}

	return sanitized
}

func sanitizeUpdateResourceRequest(req *UpdateResourceRequest) UpdateResourceRequest {
	return UpdateResourceRequest{
		Name:        sysutils.SanitizeString(req.Name),
		Description: sysutils.SanitizeString(req.Description),
	}
}

func sanitizeCreateActionRequest(req *CreateActionRequest) CreateActionRequest {
	return CreateActionRequest{
		Name:        sysutils.SanitizeString(req.Name),
		Handle:      sysutils.SanitizeString(req.Handle),
		Description: sysutils.SanitizeString(req.Description),
	}
}

func sanitizeUpdateActionRequest(req *UpdateActionRequest) UpdateActionRequest {
	return UpdateActionRequest{
		Name:        sysutils.SanitizeString(req.Name),
		Description: sysutils.SanitizeString(req.Description),
	}
}

// Response transformation functions

func toResourceServerResponse(rs *ResourceServer) *ResourceServerResponse {
	return &ResourceServerResponse{
		ID:                 rs.ID,
		Name:               rs.Name,
		Description:        rs.Description,
		Identifier:         rs.Identifier,
		OrganizationUnitID: rs.OrganizationUnitID,
	}
}

func toResourceServerListResponse(list *ResourceServerList) *ResourceServerListResponse {
	resourceServers := make([]ResourceServerResponse, len(list.ResourceServers))
	for i, rs := range list.ResourceServers {
		resourceServers[i] = ResourceServerResponse(rs)
	}

	links := make([]LinkResponse, len(list.Links))
	for i, link := range list.Links {
		links[i] = LinkResponse(link)
	}

	return &ResourceServerListResponse{
		TotalResults:    list.TotalResults,
		StartIndex:      list.StartIndex,
		Count:           list.Count,
		ResourceServers: resourceServers,
		Links:           links,
	}
}

func toResourceResponse(res *Resource) *ResourceResponse {
	return &ResourceResponse{
		ID:          res.ID,
		Name:        res.Name,
		Handle:      res.Handle,
		Description: res.Description,
		Parent:      res.Parent,
	}
}

func toResourceListResponse(list *ResourceList) *ResourceListResponse {
	resources := make([]ResourceResponse, len(list.Resources))
	for i, res := range list.Resources {
		resources[i] = ResourceResponse(res)
	}

	links := make([]LinkResponse, len(list.Links))
	for i, link := range list.Links {
		links[i] = LinkResponse(link)
	}

	return &ResourceListResponse{
		TotalResults: list.TotalResults,
		StartIndex:   list.StartIndex,
		Count:        list.Count,
		Resources:    resources,
		Links:        links,
	}
}

func toActionResponse(action *Action) *ActionResponse {
	return &ActionResponse{
		ID:          action.ID,
		Name:        action.Name,
		Handle:      action.Handle,
		Description: action.Description,
		ResourceID:  action.ResourceID,
	}
}

func toActionListResponse(list *ActionList) *ActionListResponse {
	actions := make([]ActionResponse, len(list.Actions))
	for i, action := range list.Actions {
		actions[i] = ActionResponse(action)
	}

	links := make([]LinkResponse, len(list.Links))
	for i, link := range list.Links {
		links[i] = LinkResponse(link)
	}

	return &ActionListResponse{
		TotalResults: list.TotalResults,
		StartIndex:   list.StartIndex,
		Count:        list.Count,
		Actions:      actions,
		Links:        links,
	}
}
