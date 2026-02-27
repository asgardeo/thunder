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

// Package flowmgt provides flow definition management functionality.
package flowmgt

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/system/config"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "FlowMgtService"

// handleFormatRegex matches valid handle format:
// - starts with lowercase letter or digit
// - contains only lowercase letters, digits, underscores, or dashes
// - ends with lowercase letter or digit
var handleFormatRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*[a-z0-9]$|^[a-z0-9]$`)

// FlowMgtServiceInterface defines the interface for the flow management service.
type FlowMgtServiceInterface interface {
	ListFlows(limit, offset int, flowType common.FlowType) (*FlowListResponse, *serviceerror.ServiceError)
	CreateFlow(flowDef *FlowDefinition) (*CompleteFlowDefinition, *serviceerror.ServiceError)
	GetFlow(flowID string) (*CompleteFlowDefinition, *serviceerror.ServiceError)
	GetFlowByHandle(handle string, flowType common.FlowType) (*CompleteFlowDefinition, *serviceerror.ServiceError)
	UpdateFlow(flowID string, flowDef *FlowDefinition) (*CompleteFlowDefinition, *serviceerror.ServiceError)
	DeleteFlow(flowID string) *serviceerror.ServiceError
	ListFlowVersions(flowID string) (*FlowVersionListResponse, *serviceerror.ServiceError)
	GetFlowVersion(flowID string, version int) (*FlowVersion, *serviceerror.ServiceError)
	RestoreFlowVersion(flowID string, version int) (*CompleteFlowDefinition, *serviceerror.ServiceError)
	GetGraph(flowID string) (core.GraphInterface, *serviceerror.ServiceError)
	IsValidFlow(flowID string) bool
}

// flowMgtService is the default implementation of the FlowMgtServiceInterface.
type flowMgtService struct {
	store            flowStoreInterface
	inferenceService flowInferenceServiceInterface
	graphBuilder     graphBuilderInterface
	executorRegistry executor.ExecutorRegistryInterface
	compositeStore   *compositeFlowStore
	logger           *log.Logger
}

// newFlowMgtService creates a new instance of flowMgtService.
func newFlowMgtService(
	store flowStoreInterface,
	inferenceService flowInferenceServiceInterface,
	graphBuilder graphBuilderInterface,
	executorRegistry executor.ExecutorRegistryInterface,
	compositeStore *compositeFlowStore,
) FlowMgtServiceInterface {
	return &flowMgtService{
		store:            store,
		inferenceService: inferenceService,
		graphBuilder:     graphBuilder,
		executorRegistry: executorRegistry,
		compositeStore:   compositeStore,
		logger:           log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
}

// Flow management methods

// ListFlows retrieves a paginated list of flow definitions. Supports optional filtering by flow type.
func (s *flowMgtService) ListFlows(limit, offset int, flowType common.FlowType) (
	*FlowListResponse, *serviceerror.ServiceError) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	if offset < 0 {
		offset = 0
	}

	if flowType != "" && !isValidFlowType(flowType) {
		return nil, &ErrorInvalidFlowType
	}

	flows, totalCount, err := s.store.ListFlows(limit, offset, string(flowType))
	if err != nil {
		s.logger.Error("Failed to list flows", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	listResponse := &FlowListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(flows),
		Flows:        flows,
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}

	return listResponse, nil
}

// CreateFlow creates a new flow definition with version 1.
func (s *flowMgtService) CreateFlow(flowDef *FlowDefinition) (
	*CompleteFlowDefinition, *serviceerror.ServiceError) {
	if err := declarativeresource.CheckDeclarativeCreate(); err != nil {
		return nil, err
	}

	if err := validateFlowDefinition(flowDef); err != nil {
		return nil, err
	}

	// Check if a flow with the same handle and type already exists
	exists, err := s.store.IsFlowExistsByHandle(flowDef.Handle, flowDef.FlowType)
	if err != nil {
		s.logger.Error("Failed to check flow existence by handle", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if exists {
		return nil, &ErrorDuplicateFlowHandle
	}

	flowID, genErr := utils.GenerateUUIDv7()
	if genErr != nil {
		s.logger.Error("Failed to generate UUID v7", log.Error(genErr))
		return nil, &serviceerror.InternalServerError
	}

	createdFlow, storeErr := s.store.CreateFlow(flowID, flowDef)
	if storeErr != nil {
		s.logger.Error("Failed to create flow", log.Error(storeErr))
		return nil, &serviceerror.InternalServerError
	}

	s.logger.Debug("Flow created successfully", log.String(logKeyFlowID, flowID))

	s.tryInferRegistrationFlow(flowID, flowDef)

	return createdFlow, nil
}

// GetFlow retrieves a flow definition by its ID.
func (s *flowMgtService) GetFlow(flowID string) (*CompleteFlowDefinition, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}

	flow, err := s.store.GetFlowByID(flowID)
	if err != nil {
		if errors.Is(err, errFlowNotFound) {
			return nil, &ErrorFlowNotFound
		}
		s.logger.Error("Failed to get flow", log.String(logKeyFlowID, flowID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return flow, nil
}

// GetFlowByHandle retrieves a flow definition by its handle and type.
func (s *flowMgtService) GetFlowByHandle(handle string, flowType common.FlowType) (
	*CompleteFlowDefinition, *serviceerror.ServiceError) {
	if handle == "" {
		return nil, &ErrorMissingFlowHandle
	}
	if !isValidFlowType(flowType) {
		return nil, &ErrorInvalidFlowType
	}

	flow, err := s.store.GetFlowByHandle(handle, flowType)
	if err != nil {
		if errors.Is(err, errFlowNotFound) {
			return nil, &ErrorFlowNotFound
		}
		s.logger.Error("Failed to get flow by handle", log.String("handle", handle),
			log.String("flowType", string(flowType)), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return flow, nil
}

// UpdateFlow updates an existing flow definition with the incremented version.
// Old versions are retained up to the configured max_version_history limit.
func (s *flowMgtService) UpdateFlow(flowID string, flowDef *FlowDefinition) (
	*CompleteFlowDefinition, *serviceerror.ServiceError) {
	if err := declarativeresource.CheckDeclarativeUpdate(); err != nil {
		return nil, err
	}

	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}
	if err := validateFlowDefinition(flowDef); err != nil {
		return nil, err
	}

	logger := s.logger.With(log.String(logKeyFlowID, flowID))

	// Verify the flow exists before updating
	existingFlow, err := s.store.GetFlowByID(flowID)
	if err != nil {
		if errors.Is(err, errFlowNotFound) {
			return nil, &ErrorFlowNotFound
		}
		logger.Error("Failed to get existing flow", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Prevent updating declarative (immutable) flows
	if s.isFlowDeclarative(flowID) {
		return nil, &ErrorFlowDeclarativeReadOnly
	}

	// Prevent changing the flow type
	if existingFlow.FlowType != flowDef.FlowType {
		return nil, &ErrorCannotUpdateFlowType
	}

	// Prevent changing the handle
	if existingFlow.Handle != flowDef.Handle {
		return nil, &ErrorHandleUpdateNotAllowed
	}

	updatedFlow, err := s.store.UpdateFlow(flowID, flowDef)
	if err != nil {
		logger.Error("Failed to update flow", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Flow updated successfully")

	// Invalidate the cached graph since the flow has been updated
	s.graphBuilder.InvalidateCache(flowID)

	return updatedFlow, nil
}

// DeleteFlow deletes a flow definition and all its version history.
func (s *flowMgtService) DeleteFlow(flowID string) *serviceerror.ServiceError {
	if err := declarativeresource.CheckDeclarativeDelete(); err != nil {
		return err
	}

	if flowID == "" {
		return &ErrorMissingFlowID
	}

	logger := s.logger.With(log.String(logKeyFlowID, flowID))

	_, err := s.store.GetFlowByID(flowID)
	if err != nil {
		if errors.Is(err, errFlowNotFound) {
			// Silently return if the flow does not exist
			return nil
		}
		logger.Error("Failed to get existing flow", log.Error(err))
		return &serviceerror.InternalServerError
	}

	// Prevent deleting declarative (immutable) flows
	if s.isFlowDeclarative(flowID) {
		return &ErrorFlowDeclarativeReadOnly
	}

	err = s.store.DeleteFlow(flowID)
	if err != nil {
		logger.Error("Failed to delete flow", log.Error(err))
		return &serviceerror.InternalServerError
	}

	logger.Debug("Flow deleted successfully")

	// Invalidate the cached graph since the flow has been deleted
	s.graphBuilder.InvalidateCache(flowID)

	return nil
}

// Flow version management methods

// ListFlowVersions retrieves all versions of a flow definition.
func (s *flowMgtService) ListFlowVersions(flowID string) (
	*FlowVersionListResponse, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}

	logger := s.logger.With(log.String(logKeyFlowID, flowID))

	_, err := s.store.GetFlowByID(flowID)
	if err != nil {
		if errors.Is(err, errFlowNotFound) {
			return nil, &ErrorFlowNotFound
		}
		logger.Error("Failed to get existing flow", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	versions, err := s.store.ListFlowVersions(flowID)
	if err != nil {
		logger.Error("Failed to list flow versions", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	response := &FlowVersionListResponse{
		TotalVersions: len(versions),
		Versions:      versions,
	}

	return response, nil
}

// GetFlowVersion retrieves a specific version of a flow definition.
func (s *flowMgtService) GetFlowVersion(flowID string, version int) (
	*FlowVersion, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}
	if version <= 0 {
		return nil, &ErrorInvalidVersion
	}

	flowVersion, err := s.store.GetFlowVersion(flowID, version)
	if err != nil {
		if errors.Is(err, errFlowNotFound) {
			return nil, &ErrorFlowNotFound
		}
		if errors.Is(err, errVersionNotFound) {
			return nil, &ErrorVersionNotFound
		}
		s.logger.Error("Failed to get flow version", log.String(logKeyFlowID, flowID),
			log.Int(logKeyVersion, version), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return flowVersion, nil
}

// RestoreFlowVersion restores a specific version as the active version.
// Creates a new version by copying the configuration from the specified version.
func (s *flowMgtService) RestoreFlowVersion(flowID string, version int) (
	*CompleteFlowDefinition, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}
	if version <= 0 {
		return nil, &ErrorInvalidVersion
	}

	logger := s.logger.With(log.String(logKeyFlowID, flowID), log.Int(logKeyVersion, version))

	_, err := s.store.GetFlowVersion(flowID, version)
	if err != nil {
		if errors.Is(err, errFlowNotFound) {
			return nil, &ErrorFlowNotFound
		}
		if errors.Is(err, errVersionNotFound) {
			return nil, &ErrorVersionNotFound
		}
		logger.Error("Failed to get flow version for restore", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	restoredFlow, err := s.store.RestoreFlowVersion(flowID, version)
	if err != nil {
		logger.Error("Failed to restore flow version", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Flow version restored successfully")

	// Invalidate the cached graph since a version has been restored
	s.graphBuilder.InvalidateCache(flowID)

	return restoredFlow, nil
}

// Graph building methods

// GetGraph retrieves or builds a graph for the given flow ID.
func (s *flowMgtService) GetGraph(flowID string) (core.GraphInterface, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}

	// Fetch flow definition from store
	flow, err := s.store.GetFlowByID(flowID)
	if err != nil {
		if errors.Is(err, errFlowNotFound) {
			return nil, &ErrorFlowNotFound
		}
		s.logger.Error("Failed to get flow for graph building", log.String(logKeyFlowID, flowID),
			log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return s.graphBuilder.GetGraph(flow)
}

// IsValidFlow checks if a valid flow exists for the given flow ID.
func (s *flowMgtService) IsValidFlow(flowID string) bool {
	if flowID == "" {
		return false
	}

	exists, err := s.store.IsFlowExists(flowID)
	if err != nil {
		s.logger.Error("Failed to check flow existence", log.String(logKeyFlowID, flowID), log.Error(err))
		return false
	}

	return exists
}

// Helper functions

// isValidFlowType checks if the provided flow type is valid.
func isValidFlowType(flowType common.FlowType) bool {
	return flowType == common.FlowTypeAuthentication ||
		flowType == common.FlowTypeRegistration ||
		flowType == common.FlowTypeUserOnboarding
}

// buildPaginationLinks constructs pagination links for the flow list response.
func buildPaginationLinks(limit, offset, totalCount int) []Link {
	links := make([]Link, 0)

	// Add first and previous links if not on first page
	if offset > 0 {
		links = append(links, Link{
			Href: fmt.Sprintf("/flows?offset=0&limit=%d", limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, Link{
			Href: fmt.Sprintf("/flows?offset=%d&limit=%d", prevOffset, limit),
			Rel:  "prev",
		})
	}

	// Add next link if there are more results
	if offset+limit < totalCount {
		nextOffset := offset + limit
		links = append(links, Link{
			Href: fmt.Sprintf("/flows?offset=%d&limit=%d", nextOffset, limit),
			Rel:  "next",
		})
	}

	// Add last link if not on last page
	lastPageOffset := ((totalCount - 1) / limit) * limit
	if totalCount > 0 && offset < lastPageOffset {
		links = append(links, Link{
			Href: fmt.Sprintf("/flows?offset=%d&limit=%d", lastPageOffset, limit),
			Rel:  "last",
		})
	}

	return links
}

// validateFlowDefinition validates the flow definition request.
func validateFlowDefinition(flowDef *FlowDefinition) *serviceerror.ServiceError {
	if flowDef == nil {
		return &ErrorInvalidRequestFormat
	}
	if flowDef.Handle == "" {
		return &ErrorMissingFlowHandle
	}
	if !isValidHandleFormat(flowDef.Handle) {
		return &ErrorInvalidFlowHandleFormat
	}
	if flowDef.Name == "" {
		return &ErrorMissingFlowName
	}
	if !isValidFlowType(flowDef.FlowType) {
		return &ErrorInvalidFlowType
	}

	if len(flowDef.Nodes) < 2 {
		return serviceerror.CustomServiceError(ErrorInvalidFlowData,
			"Flow definition must contain at least a start and an end node")
	} else if len(flowDef.Nodes) == 2 {
		return serviceerror.CustomServiceError(ErrorInvalidFlowData,
			"Flow definition must contain nodes between start and end nodes")
	}

	return nil
}

// isValidHandleFormat validates that the handle follows the required format:
// - all lowercase
// - alphanumeric characters
// - can contain underscores (_) or dashes (-)
// - cannot start or end with underscore or dash
func isValidHandleFormat(handle string) bool {
	return handleFormatRegex.MatchString(handle)
}

// tryInferRegistrationFlow attempts to infer and create a registration flow from an authentication flow
func (s *flowMgtService) tryInferRegistrationFlow(authFlowID string, authFlowDef *FlowDefinition) {
	logger := s.logger.With(log.String("authFlowID", authFlowID))

	if !config.GetThunderRuntime().Config.Flow.AutoInferRegistration {
		logger.Debug("Automatic registration flow inference is disabled")
		return
	}

	if authFlowDef.FlowType != common.FlowTypeAuthentication {
		logger.Debug("Flow is not an authentication flow, skipping registration inference",
			log.String("flowType", string(authFlowDef.FlowType)))
		return
	}

	// Check if auth flow already contains PasskeyAuthExecutor with registration modes
	// If so, skip registration flow inference as the auth flow handles registration internally
	if s.hasPasskeyRegistrationModes(authFlowDef) {
		logger.Debug("Authentication flow contains PasskeyAuthExecutor with " +
			"register_start and register_finish modes, skipping registration inference")
		return
	}

	logger.Debug("Inferring registration flow from authentication flow",
		log.String("flowName", authFlowDef.Name))

	regFlowDef, inferErr := s.inferenceService.InferRegistrationFlow(authFlowDef)
	if inferErr != nil {
		logger.Error("Failed to infer registration flow", log.Error(inferErr))
		return
	}

	regFlowID, uuidErr := utils.GenerateUUIDv7()
	if uuidErr != nil {
		logger.Error("Failed to generate UUID for inferred registration flow", log.Error(uuidErr))
		return
	}

	_, storeErr := s.store.CreateFlow(regFlowID, regFlowDef)
	if storeErr != nil {
		logger.Error("Failed to create inferred registration flow", log.Error(storeErr))
		return
	}

	logger.Debug("Successfully inferred and created registration flow",
		log.String("authFlowName", authFlowDef.Name), log.String("regFlowID", regFlowID),
		log.String("regFlowName", regFlowDef.Name))
}

// hasPasskeyRegistrationModes checks if the flow contains PasskeyAuthExecutor with both
// register_start and register_finish modes, indicating the auth flow handles passkey registration internally.
func (s *flowMgtService) hasPasskeyRegistrationModes(flowDef *FlowDefinition) bool {
	hasRegStart := false
	hasRegFinish := false

	for _, node := range flowDef.Nodes {
		if node.Executor != nil && node.Executor.Name == executor.ExecutorNamePasskeyAuth {
			switch node.Executor.Mode {
			case "register_start":
				hasRegStart = true
			case "register_finish":
				hasRegFinish = true
			}
		}
		// Early exit if both modes are found
		if hasRegStart && hasRegFinish {
			return true
		}
	}

	return hasRegStart && hasRegFinish
}

// isFlowDeclarative checks if a flow is declarative (read-only from file store).
// Returns true only if the flow exists in the file store but NOT in the database store.
// Database flows take precedence - if a flow exists in both stores, it's treated as mutable (not declarative).
func (s *flowMgtService) isFlowDeclarative(flowID string) bool {
	if s.compositeStore == nil {
		return false
	}

	// Check DB store first - if flow exists in DB, it's mutable (not declarative)
	if s.compositeStore.dbStore != nil {
		_, err := s.compositeStore.dbStore.GetFlowByID(flowID)
		if err == nil {
			// Flow exists in DB, so it's mutable (not declarative)
			return false
		}
	}

	// Flow not in DB, check if it exists in file store
	if s.compositeStore.fileStore != nil {
		_, err := s.compositeStore.fileStore.GetFlowByID(flowID)
		return err == nil
	}

	return false
}
