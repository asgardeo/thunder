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
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
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
	ListFlows(ctx context.Context, limit, offset int, flowType common.FlowType) (*FlowListResponse,
		*serviceerror.ServiceError)
	CreateFlow(ctx context.Context, flowDef *FlowDefinition) (*CompleteFlowDefinition, *serviceerror.ServiceError)
	GetFlow(ctx context.Context, flowID string) (*CompleteFlowDefinition, *serviceerror.ServiceError)
	GetFlowByHandle(ctx context.Context, handle string, flowType common.FlowType) (*CompleteFlowDefinition,
		*serviceerror.ServiceError)
	UpdateFlow(ctx context.Context, flowID string, flowDef *FlowDefinition) (*CompleteFlowDefinition,
		*serviceerror.ServiceError)
	DeleteFlow(ctx context.Context, flowID string) *serviceerror.ServiceError
	ListFlowVersions(ctx context.Context, flowID string) (*FlowVersionListResponse, *serviceerror.ServiceError)
	GetFlowVersion(ctx context.Context, flowID string, version int) (*FlowVersion, *serviceerror.ServiceError)
	RestoreFlowVersion(ctx context.Context, flowID string, version int) (*CompleteFlowDefinition,
		*serviceerror.ServiceError)
	GetGraph(ctx context.Context, flowID string) (core.GraphInterface, *serviceerror.ServiceError)
	IsValidFlow(ctx context.Context, flowID string) bool
}

// flowMgtService is the default implementation of the FlowMgtServiceInterface.
type flowMgtService struct {
	store            flowStoreInterface
	inferenceService flowInferenceServiceInterface
	graphBuilder     graphBuilderInterface
	executorRegistry executor.ExecutorRegistryInterface
	transactioner    transaction.Transactioner
	logger           *log.Logger
}

// newFlowMgtService creates a new instance of flowMgtService.
func newFlowMgtService(
	store flowStoreInterface,
	inferenceService flowInferenceServiceInterface,
	graphBuilder graphBuilderInterface,
	executorRegistry executor.ExecutorRegistryInterface,
	transactioner transaction.Transactioner,
) FlowMgtServiceInterface {
	return &flowMgtService{
		store:            store,
		inferenceService: inferenceService,
		graphBuilder:     graphBuilder,
		executorRegistry: executorRegistry,
		transactioner:    transactioner,
		logger:           log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
}

// Flow management methods

// ListFlows retrieves a paginated list of flow definitions. Supports optional filtering by flow type.
func (s *flowMgtService) ListFlows(ctx context.Context, limit, offset int, flowType common.FlowType) (
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

	flows, totalCount, err := s.store.ListFlows(ctx, limit, offset, string(flowType))
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
func (s *flowMgtService) CreateFlow(ctx context.Context, flowDef *FlowDefinition) (
	*CompleteFlowDefinition, *serviceerror.ServiceError) {
	if err := declarativeresource.CheckDeclarativeCreate(); err != nil {
		return nil, err
	}

	if err := validateFlowDefinition(flowDef); err != nil {
		return nil, err
	}

	var createdFlow *CompleteFlowDefinition
	var capturedSvcErr *serviceerror.ServiceError

	err := s.transactioner.Transact(ctx, func(ctx context.Context) error {
		// Check if a flow with the same handle and type already exists
		exists, err := s.store.IsFlowExistsByHandle(ctx, flowDef.Handle, flowDef.FlowType)
		if err != nil {
			s.logger.Error("Failed to check flow existence by handle", log.Error(err))
			return err
		}
		if exists {
			capturedSvcErr = &ErrorDuplicateFlowHandle
			return errors.New("rollback")
		}

		svcErr := s.applyExecutorDefaultMeta(flowDef)
		if svcErr != nil {
			capturedSvcErr = svcErr
			return errors.New("rollback")
		}

		flowID, genErr := utils.GenerateUUIDv7()
		if genErr != nil {
			s.logger.Error("Failed to generate UUID v7", log.Error(genErr))
			return genErr
		}

		createdFlow, err = s.store.CreateFlow(ctx, flowID, flowDef)
		if err != nil {
			s.logger.Error("Failed to create flow", log.Error(err))
			return err
		}

		s.logger.Debug("Flow created successfully", log.String(logKeyFlowID, flowID))

		s.tryInferRegistrationFlow(ctx, flowID, flowDef)
		return nil
	})

	if capturedSvcErr != nil {
		return nil, capturedSvcErr
	}

	if err != nil {
		return nil, &serviceerror.InternalServerError
	}

	return createdFlow, nil
}

// GetFlow retrieves a flow definition by its ID.
func (s *flowMgtService) GetFlow(ctx context.Context, flowID string) (*CompleteFlowDefinition,
	*serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}

	flow, err := s.store.GetFlowByID(ctx, flowID)
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
func (s *flowMgtService) GetFlowByHandle(ctx context.Context, handle string, flowType common.FlowType) (
	*CompleteFlowDefinition, *serviceerror.ServiceError) {
	if handle == "" {
		return nil, &ErrorMissingFlowHandle
	}
	if !isValidFlowType(flowType) {
		return nil, &ErrorInvalidFlowType
	}

	flow, err := s.store.GetFlowByHandle(ctx, handle, flowType)
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
func (s *flowMgtService) UpdateFlow(ctx context.Context, flowID string, flowDef *FlowDefinition) (
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

	var updatedFlow *CompleteFlowDefinition
	var capturedSvcErr *serviceerror.ServiceError

	errTx := s.transactioner.Transact(ctx, func(ctx context.Context) error {
		// Verify the flow exists before updating
		existingFlow, err := s.store.GetFlowByID(ctx, flowID)
		if err != nil {
			if errors.Is(err, errFlowNotFound) {
				capturedSvcErr = &ErrorFlowNotFound
				return errors.New("rollback")
			}
			logger.Error("Failed to get existing flow", log.Error(err))
			return err
		}

		// Prevent changing the flow type
		if existingFlow.FlowType != flowDef.FlowType {
			capturedSvcErr = &ErrorCannotUpdateFlowType
			return errors.New("rollback")
		}

		// Prevent changing the handle
		if existingFlow.Handle != flowDef.Handle {
			capturedSvcErr = &ErrorHandleUpdateNotAllowed
			return errors.New("rollback")
		}

		svcErr := s.applyExecutorDefaultMeta(flowDef)
		if svcErr != nil {
			capturedSvcErr = svcErr
			return errors.New("rollback")
		}

		var errStore error
		updatedFlow, errStore = s.store.UpdateFlow(ctx, flowID, flowDef)
		if errStore != nil {
			logger.Error("Failed to update flow", log.Error(errStore))
			return errStore
		}

		logger.Debug("Flow updated successfully")

		// Invalidate the cached graph since the flow has been updated
		s.graphBuilder.InvalidateCache(flowID)

		return nil
	})

	if capturedSvcErr != nil {
		return nil, capturedSvcErr
	}

	if errTx != nil {
		return nil, &serviceerror.InternalServerError
	}

	return updatedFlow, nil
}

// DeleteFlow deletes a flow definition and all its version history.
func (s *flowMgtService) DeleteFlow(ctx context.Context, flowID string) *serviceerror.ServiceError {
	if err := declarativeresource.CheckDeclarativeDelete(); err != nil {
		return err
	}

	if flowID == "" {
		return &ErrorMissingFlowID
	}

	logger := s.logger.With(log.String(logKeyFlowID, flowID))

	errTx := s.transactioner.Transact(ctx, func(ctx context.Context) error {
		_, err := s.store.GetFlowByID(ctx, flowID)
		if err != nil {
			if errors.Is(err, errFlowNotFound) {
				// Silently return if the flow does not exist
				return nil
			}
			logger.Error("Failed to get existing flow", log.Error(err))
			return err
		}

		err = s.store.DeleteFlow(ctx, flowID)
		if err != nil {
			logger.Error("Failed to delete flow", log.Error(err))
			return err
		}

		logger.Debug("Flow deleted successfully")

		// Invalidate the cached graph since the flow has been deleted
		s.graphBuilder.InvalidateCache(flowID)
		return nil
	})

	if errTx != nil {
		return &serviceerror.InternalServerError
	}

	return nil
}

// Flow version management methods

// ListFlowVersions retrieves all versions of a flow definition.
func (s *flowMgtService) ListFlowVersions(ctx context.Context, flowID string) (
	*FlowVersionListResponse, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}

	logger := s.logger.With(log.String(logKeyFlowID, flowID))

	_, err := s.store.GetFlowByID(ctx, flowID)
	if err != nil {
		if errors.Is(err, errFlowNotFound) {
			return nil, &ErrorFlowNotFound
		}
		logger.Error("Failed to get existing flow", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	versions, err := s.store.ListFlowVersions(ctx, flowID)
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
func (s *flowMgtService) GetFlowVersion(ctx context.Context, flowID string, version int) (
	*FlowVersion, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}
	if version <= 0 {
		return nil, &ErrorInvalidVersion
	}

	flowVersion, err := s.store.GetFlowVersion(ctx, flowID, version)
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
func (s *flowMgtService) RestoreFlowVersion(ctx context.Context, flowID string, version int) (
	*CompleteFlowDefinition, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}
	if version <= 0 {
		return nil, &ErrorInvalidVersion
	}

	logger := s.logger.With(log.String(logKeyFlowID, flowID), log.Int(logKeyVersion, version))

	var restoredFlow *CompleteFlowDefinition
	var capturedSvcErr *serviceerror.ServiceError

	errTx := s.transactioner.Transact(ctx, func(ctx context.Context) error {
		_, err := s.store.GetFlowVersion(ctx, flowID, version)
		if err != nil {
			if errors.Is(err, errFlowNotFound) {
				capturedSvcErr = &ErrorFlowNotFound
				return errors.New("rollback")
			}
			if errors.Is(err, errVersionNotFound) {
				capturedSvcErr = &ErrorVersionNotFound
				return errors.New("rollback")
			}
			logger.Error("Failed to get flow version for restore", log.Error(err))
			return err
		}

		var errStore error
		restoredFlow, errStore = s.store.RestoreFlowVersion(ctx, flowID, version)
		if errStore != nil {
			logger.Error("Failed to restore flow version", log.Error(errStore))
			return errStore
		}

		logger.Debug("Flow version restored successfully")

		// Invalidate the cached graph since a version has been restored
		s.graphBuilder.InvalidateCache(flowID)
		return nil
	})

	if capturedSvcErr != nil {
		return nil, capturedSvcErr
	}

	if errTx != nil {
		return nil, &serviceerror.InternalServerError
	}

	return restoredFlow, nil
}

// Graph building methods

// GetGraph retrieves or builds a graph for the given flow ID.
func (s *flowMgtService) GetGraph(ctx context.Context, flowID string) (core.GraphInterface,
	*serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorMissingFlowID
	}

	// Fetch flow definition from store
	flow, err := s.store.GetFlowByID(ctx, flowID)
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
func (s *flowMgtService) IsValidFlow(ctx context.Context, flowID string) bool {
	if flowID == "" {
		return false
	}

	exists, err := s.store.IsFlowExists(ctx, flowID)
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
func (s *flowMgtService) tryInferRegistrationFlow(ctx context.Context, authFlowID string, authFlowDef *FlowDefinition) {
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

	logger.Debug("Inferring registration flow from authentication flow",
		log.String("flowName", authFlowDef.Name))

	regFlowDef, inferErr := s.inferenceService.InferRegistrationFlow(authFlowDef)
	if inferErr != nil {
		logger.Error("Failed to infer registration flow", log.Error(inferErr))
		return
	}

	metaErr := s.applyExecutorDefaultMeta(regFlowDef)
	if metaErr != nil {
		logger.Error("Failed to apply executor default meta to inferred registration flow",
			log.String("error", metaErr.Code))
		return
	}

	regFlowID, uuidErr := utils.GenerateUUIDv7()
	if uuidErr != nil {
		logger.Error("Failed to generate UUID for inferred registration flow", log.Error(uuidErr))
		return
	}

	_, storeErr := s.store.CreateFlow(ctx, regFlowID, regFlowDef)
	if storeErr != nil {
		logger.Error("Failed to create inferred registration flow", log.Error(storeErr))
		return
	}

	logger.Debug("Successfully inferred and created registration flow",
		log.String("authFlowName", authFlowDef.Name), log.String("regFlowID", regFlowID),
		log.String("regFlowName", regFlowDef.Name))
}

// applyExecutorDefaultMeta applies default meta from executors to TASK_EXECUTION nodes.
func (s *flowMgtService) applyExecutorDefaultMeta(flowDef *FlowDefinition) *serviceerror.ServiceError {
	if s.executorRegistry == nil {
		s.logger.Error("Executor registry is nil, cannot apply default meta")
		return &serviceerror.InternalServerError
	}

	for i := range flowDef.Nodes {
		node := &flowDef.Nodes[i]

		if node.Type != string(common.NodeTypeTaskExecution) || node.Executor == nil {
			continue
		}
		if node.Meta != nil {
			s.logger.Debug("Node already has meta, skipping default meta application",
				log.String("nodeID", node.ID), log.String("executorName", node.Executor.Name))
			continue
		}

		exec, err := s.executorRegistry.GetExecutor(node.Executor.Name)
		if err != nil {
			s.logger.Error("Failed to get executor for default meta application",
				log.String("nodeID", node.ID), log.String("executorName", node.Executor.Name), log.Error(err))
			return &serviceerror.InternalServerError
		}

		meta := exec.GetDefaultMeta()
		if meta != nil {
			node.Meta = meta
		}
	}

	return nil
}
