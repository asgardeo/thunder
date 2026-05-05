/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

// Package agent provides functionality for managing agents
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/agent/model"
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/entity"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/i18n/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/pkg/inboundclient"
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
	oauthutils "github.com/asgardeo/thunder/pkg/oauth/oauth2/utils"
)

// AgentServiceInterface defines the operations exposed by the agent service.
type AgentServiceInterface interface {
	CreateAgent(ctx context.Context, req *model.CreateAgentRequest) (*model.AgentCompleteResponse,
		*serviceerror.ServiceError)
	GetAgent(ctx context.Context, agentID string, includeDisplay bool) (*model.AgentGetResponse,
		*serviceerror.ServiceError)
	GetAgentList(ctx context.Context, limit, offset int, filters map[string]interface{},
		includeDisplay bool) (*model.AgentListResponse, *serviceerror.ServiceError)
	UpdateAgent(ctx context.Context, agentID string, req *model.UpdateAgentRequest) (
		*model.AgentCompleteResponse, *serviceerror.ServiceError)
	DeleteAgent(ctx context.Context, agentID string) *serviceerror.ServiceError
	GetAgentGroups(ctx context.Context, agentID string, limit, offset int) (
		*model.AgentGroupListResponse, *serviceerror.ServiceError)
}

type agentService struct {
	logger               *log.Logger
	entityService        entity.EntityServiceInterface
	inboundClientService inboundclient.InboundClientServiceInterface
	ouService            oupkg.OrganizationUnitServiceInterface
}

func newAgentService(
	entityService entity.EntityServiceInterface,
	inboundClientService inboundclient.InboundClientServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
) AgentServiceInterface {
	return &agentService{
		logger:               log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AgentService")),
		entityService:        entityService,
		inboundClientService: inboundClientService,
		ouService:            ouService,
	}
}

// CreateAgent creates an agent entity with optional inbound auth profile.
func (s *agentService) CreateAgent(ctx context.Context, req *model.CreateAgentRequest) (
	*model.AgentCompleteResponse, *serviceerror.ServiceError) {
	if req == nil {
		return nil, &ErrorInvalidRequestFormat
	}
	if svcErr := validateBaseFields(req.Name, req.Type); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := s.validateOUExists(ctx, req.OUID); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := s.validateNameUnique(ctx, req.Name, ""); svcErr != nil {
		return nil, svcErr
	}

	agentID, err := sysutils.GenerateUUIDv7()
	if err != nil {
		s.logger.Error("Failed to generate agent ID", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	normalizeLoginConsent(req.LoginConsent)

	clientID, clientSecret, svcErr := s.resolveOAuthCredentials(ctx, req.InboundAuthConfig, "", "")
	if svcErr != nil {
		return nil, svcErr
	}

	owner := req.Owner
	if owner == "" {
		owner = security.GetSubject(ctx)
	}

	e, sysCredsJSON, buildErr := buildAgentEntity(agentID, req.Type, req.OUID, req.Attributes,
		req.Name, req.Description, owner, clientID, clientSecret)
	if buildErr != nil {
		s.logger.Error("Failed to build agent entity", log.Error(buildErr))
		return nil, &serviceerror.InternalServerError
	}

	createdEntity, entErr := s.entityService.CreateEntity(ctx, e, sysCredsJSON)
	if entErr != nil {
		if mapped := mapEntityError(entErr); mapped != nil {
			return nil, mapped
		}
		s.logger.Error("Failed to create agent entity", log.String("agentID", agentID), log.Error(entErr))
		return nil, &serviceerror.InternalServerError
	}

	authFlowID, regFlowID := req.AuthFlowID, req.RegistrationFlowID
	assertion, loginConsent := req.Assertion, req.LoginConsent
	var inboundConfigs []model.InboundAuthConfig

	if needsInboundClient(req) {
		resolvedClient, resolvedOAuth, svcErr := s.createInboundForAgent(ctx, agentID, req, clientSecret)
		if svcErr != nil {
			s.deleteEntityCompensation(ctx, agentID)
			return nil, svcErr
		}
		authFlowID = resolvedClient.AuthFlowID
		regFlowID = resolvedClient.RegistrationFlowID
		assertion = resolvedClient.Assertion
		loginConsent = resolvedClient.LoginConsent
		if resolvedOAuth != nil {
			inboundConfigs = []model.InboundAuthConfig{{
				Type:   model.OAuthInboundAuthType,
				Config: oauthProfileToConfig(clientID, resolvedOAuth),
			}}
		}
	}

	resp := buildCompleteResponse(agentID, owner, clientID, clientSecret,
		req.Type, req.Name, req.Description, createdEntity.Attributes,
		authFlowID, regFlowID, req.IsRegistrationFlowEnabled,
		req.ThemeID, req.LayoutID, assertion, loginConsent,
		req.AllowedUserTypes, req.Certificate, inboundConfigs)
	resp.OUID = req.OUID
	s.populateOUHandleForComplete(ctx, resp)
	return resp, nil
}

// GetAgent returns a single agent by ID.
func (s *agentService) GetAgent(ctx context.Context, agentID string, includeDisplay bool) (
	*model.AgentGetResponse, *serviceerror.ServiceError) {
	if agentID == "" {
		return nil, &ErrorMissingAgentID
	}

	e, err := s.entityService.GetEntity(ctx, agentID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, &ErrorAgentNotFound
		}
		s.logger.Error("Failed to retrieve agent entity", log.String("agentID", agentID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if e.Category != entity.EntityCategoryAgent {
		return nil, &ErrorAgentNotFound
	}

	resp, svcErr := s.composeGetResponse(ctx, e)
	if svcErr != nil {
		return nil, svcErr
	}

	if includeDisplay {
		s.populateOUHandleForGet(ctx, resp)
	}

	return resp, nil
}

// GetAgentList returns a paginated list of agents.
func (s *agentService) GetAgentList(ctx context.Context, limit, offset int,
	filters map[string]interface{}, includeDisplay bool) (
	*model.AgentListResponse, *serviceerror.ServiceError) {
	if svcErr := validatePaginationParams(limit, offset); svcErr != nil {
		return nil, svcErr
	}
	if limit == 0 {
		limit = 30
	}

	totalCount, err := s.entityService.GetEntityListCount(ctx, entity.EntityCategoryAgent, filters)
	if err != nil {
		s.logger.Error("Failed to get agent list count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	entities, err := s.entityService.GetEntityList(ctx, entity.EntityCategoryAgent, limit, offset, filters)
	if err != nil {
		s.logger.Error("Failed to get agent list", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return s.buildListResponse(ctx, entities, totalCount, limit, offset, includeDisplay), nil
}

// UpdateAgent applies a full-replacement update to the agent.
func (s *agentService) UpdateAgent(ctx context.Context, agentID string,
	req *model.UpdateAgentRequest) (*model.AgentCompleteResponse, *serviceerror.ServiceError) {
	if agentID == "" {
		return nil, &ErrorMissingAgentID
	}
	if req == nil {
		return nil, &ErrorInvalidRequestFormat
	}
	if svcErr := validateBaseFields(req.Name, req.Type); svcErr != nil {
		return nil, svcErr
	}
	existing, err := s.entityService.GetEntity(ctx, agentID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, &ErrorAgentNotFound
		}
		s.logger.Error("Failed to retrieve agent entity for update",
			log.String("agentID", agentID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if existing.Category != entity.EntityCategoryAgent {
		return nil, &ErrorAgentNotFound
	}
	if existing.IsReadOnly {
		return nil, &ErrorCannotModifyDeclarativeResource
	}

	currentName, _, currentOwner, currentClientID := readSystemAttributes(existing.SystemAttributes)
	if req.Name != currentName {
		if svcErr := s.validateNameUnique(ctx, req.Name, agentID); svcErr != nil {
			return nil, svcErr
		}
	}

	normalizeLoginConsent(req.LoginConsent)

	var existingOAuthMethod string
	existingOAuth, oauthErr := s.inboundClientService.GetOAuthProfileByEntityID(ctx, agentID)
	if oauthErr != nil && !errors.Is(oauthErr, inboundclient.ErrInboundClientNotFound) {
		s.logger.Error("Failed to load existing OAuth profile",
			log.String("agentID", agentID), log.Error(oauthErr))
		return nil, &serviceerror.InternalServerError
	}
	if existingOAuth != nil && existingOAuth.OAuthProfile != nil {
		existingOAuthMethod = existingOAuth.OAuthProfile.TokenEndpointAuthMethod
	}

	clientID, clientSecret, svcErr := s.resolveOAuthCredentials(
		ctx, req.InboundAuthConfig, currentClientID, existingOAuthMethod)
	if svcErr != nil {
		return nil, svcErr
	}

	owner := req.Owner
	if owner == "" {
		owner = currentOwner
	}

	ouID := req.OUID
	if ouID == "" {
		ouID = existing.OUID
	} else if ouID != existing.OUID {
		if svcErr := s.validateOUExists(ctx, ouID); svcErr != nil {
			return nil, svcErr
		}
	}

	resolvedClient, resolvedOAuth, svcErr := s.reconcileInboundForUpdate(
		ctx, agentID, req, clientID, clientSecret, currentName, req.Name)
	if svcErr != nil {
		return nil, svcErr
	}

	if resolvedOAuth == nil {
		clientID = ""
		clientSecret = ""
	}

	updatedEntity := &entity.Entity{
		ID:         agentID,
		Category:   entity.EntityCategoryAgent,
		Type:       req.Type,
		State:      entity.EntityStateActive,
		OUID:       ouID,
		Attributes: req.Attributes,
	}
	sysAttrsJSON, marshalErr := buildSystemAttributesJSON(req.Name, req.Description, owner, clientID)
	if marshalErr != nil {
		s.logger.Error("Failed to build system attributes for update", log.Error(marshalErr))
		return nil, &serviceerror.InternalServerError
	}
	updatedEntity.SystemAttributes = sysAttrsJSON

	if _, err := s.entityService.UpdateEntity(ctx, agentID, updatedEntity); err != nil {
		if mapped := mapEntityError(err); mapped != nil {
			return nil, mapped
		}
		s.logger.Error("Failed to update agent entity", log.String("agentID", agentID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if clientSecret != "" {
		sysCredsJSON, credErr := buildSystemCredentialsJSON(clientSecret)
		if credErr == nil && sysCredsJSON != nil {
			if err := s.entityService.UpdateSystemCredentials(ctx, agentID, sysCredsJSON); err != nil {
				s.logger.Error("Failed to update agent system credentials",
					log.String("agentID", agentID), log.Error(err))
				return nil, &serviceerror.InternalServerError
			}
		}
	}

	authFlowID := resolvedClient.AuthFlowID
	regFlowID := resolvedClient.RegistrationFlowID
	assertion := resolvedClient.Assertion
	loginConsent := resolvedClient.LoginConsent
	var inboundConfigs []model.InboundAuthConfig
	if resolvedOAuth != nil {
		inboundConfigs = []model.InboundAuthConfig{{
			Type:   model.OAuthInboundAuthType,
			Config: oauthProfileToConfig(clientID, resolvedOAuth),
		}}
	}

	resp := buildCompleteResponse(agentID, owner, clientID, clientSecret,
		req.Type, req.Name, req.Description, req.Attributes,
		authFlowID, regFlowID, resolvedClient.IsRegistrationFlowEnabled,
		req.ThemeID, req.LayoutID, assertion, loginConsent,
		req.AllowedUserTypes, req.Certificate, inboundConfigs)
	resp.OUID = ouID
	s.populateOUHandleForComplete(ctx, resp)
	return resp, nil
}

// DeleteAgent removes the agent and its associated inbound client.
func (s *agentService) DeleteAgent(ctx context.Context, agentID string) *serviceerror.ServiceError {
	if agentID == "" {
		return &ErrorMissingAgentID
	}

	existing, err := s.entityService.GetEntity(ctx, agentID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return &ErrorAgentNotFound
		}
		s.logger.Error("Failed to retrieve agent for delete", log.String("agentID", agentID), log.Error(err))
		return &serviceerror.InternalServerError
	}
	if existing.Category != entity.EntityCategoryAgent {
		return &ErrorAgentNotFound
	}
	if existing.IsReadOnly {
		return &ErrorCannotModifyDeclarativeResource
	}

	if err := s.inboundClientService.DeleteInboundClient(ctx, agentID); err != nil &&
		!errors.Is(err, inboundclient.ErrInboundClientNotFound) {
		if errors.Is(err, inboundclient.ErrCannotModifyDeclarative) {
			return &ErrorCannotModifyDeclarativeResource
		}
		s.logger.Error("Failed to delete inbound client for agent",
			log.String("agentID", agentID), log.Error(err))
		return &serviceerror.InternalServerError
	}

	if err := s.entityService.DeleteEntity(ctx, agentID); err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil
		}
		s.logger.Error("Failed to delete agent entity", log.String("agentID", agentID), log.Error(err))
		return &serviceerror.InternalServerError
	}
	return nil
}

// GetAgentGroups returns the groups the agent belongs to.
func (s *agentService) GetAgentGroups(ctx context.Context, agentID string, limit, offset int) (
	*model.AgentGroupListResponse, *serviceerror.ServiceError) {
	if agentID == "" {
		return nil, &ErrorMissingAgentID
	}
	if svcErr := validatePaginationParams(limit, offset); svcErr != nil {
		return nil, svcErr
	}
	if limit == 0 {
		limit = 30
	}

	existing, err := s.entityService.GetEntity(ctx, agentID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, &ErrorAgentNotFound
		}
		s.logger.Error("Failed to retrieve agent for groups", log.String("agentID", agentID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if existing.Category != entity.EntityCategoryAgent {
		return nil, &ErrorAgentNotFound
	}

	totalCount, err := s.entityService.GetGroupCountForEntity(ctx, agentID)
	if err != nil {
		s.logger.Error("Failed to get agent group count", log.String("agentID", agentID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	groups, err := s.entityService.GetEntityGroups(ctx, agentID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get agent groups", log.String("agentID", agentID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	out := make([]model.AgentGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, model.AgentGroup{ID: g.ID, Name: g.Name, OUID: g.OUID})
	}

	resp := &model.AgentGroupListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(out),
		Groups:       out,
		Links: sysutils.BuildPaginationLinks(
			fmt.Sprintf("%s/%s/groups", agentBasePath, agentID), limit, offset, totalCount, ""),
	}
	return resp, nil
}

// deleteEntityCompensation deletes the entity row as a best-effort rollback after a failed downstream operation.
func (s *agentService) deleteEntityCompensation(ctx context.Context, agentID string) {
	if err := s.entityService.DeleteEntity(ctx, agentID); err != nil {
		s.logger.Error("Failed to delete entity during compensation",
			log.String("agentID", agentID), log.Error(err))
	}
}

// validateOUExists returns an error if the given OU is empty or does not exist.
func (s *agentService) validateOUExists(ctx context.Context, ouID string) *serviceerror.ServiceError {
	if ouID == "" {
		return &ErrorOrganizationUnitNotFound
	}
	exists, err := s.ouService.IsOrganizationUnitExists(ctx, ouID)
	if err != nil {
		if err.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return &ErrorOrganizationUnitNotFound
		}
		s.logger.Error("Failed to verify OU existence", log.String("ouID", ouID), log.Any("error", err))
		return &serviceerror.InternalServerError
	}
	if !exists {
		return &ErrorOrganizationUnitNotFound
	}
	return nil
}

// validateNameUnique returns an error if another agent already uses the given name (excludeID is exempt on updates).
func (s *agentService) validateNameUnique(ctx context.Context, name, excludeID string) *serviceerror.ServiceError {
	if name == "" {
		return &ErrorInvalidAgentName
	}
	id, err := s.entityService.IdentifyEntity(ctx, map[string]interface{}{fieldName: name})
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil
		}
		if errors.Is(err, entity.ErrAmbiguousEntity) {
			return &ErrorAgentAlreadyExistsWithName
		}
		s.logger.Error("Failed to verify agent name uniqueness", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if id == nil || *id == "" {
		return nil
	}
	if excludeID != "" && *id == excludeID {
		return nil
	}
	// Verify the found entity is actually an agent before treating it as a name conflict.
	// IdentifyEntity searches across all entity categories; apps also store their name in
	// system attributes under the same key.
	found, getErr := s.entityService.GetEntity(ctx, *id)
	if getErr != nil || found.Category != entity.EntityCategoryAgent {
		return nil
	}
	return &ErrorAgentAlreadyExistsWithName
}

// resolveOAuthCredentials resolves the clientID and clientSecret for an agent OAuth profile.
func (s *agentService) resolveOAuthCredentials(ctx context.Context,
	configs []model.InboundAuthConfig, existingClientID, existingOAuthMethod string,
) (string, string, *serviceerror.ServiceError) {
	oauthCfg := pickOAuthConfig(configs)
	if oauthCfg == nil {
		return existingClientID, "", nil
	}

	clientID := oauthCfg.ClientID
	if clientID == "" {
		clientID = existingClientID
	}
	if clientID == "" {
		generated, err := oauthutils.GenerateOAuth2ClientID()
		if err != nil {
			s.logger.Error("Failed to generate client ID", log.Error(err))
			return "", "", &serviceerror.InternalServerError
		}
		clientID = generated
	} else if clientID != existingClientID {
		taken, svcErr := s.isClientIDTaken(ctx, clientID, existingClientID)
		if svcErr != nil {
			return "", "", svcErr
		}
		if taken {
			return "", "", &ErrorAgentAlreadyExistsWithClientID
		}
	}

	clientSecret := oauthCfg.ClientSecret
	requiresSecret := requiresClientSecret(oauthCfg)
	if requiresSecret && clientSecret == "" {
		existingWasSecretBased := existingOAuthMethod == string(oauth2const.TokenEndpointAuthMethodClientSecretBasic) ||
			existingOAuthMethod == string(oauth2const.TokenEndpointAuthMethodClientSecretPost)
		if !existingWasSecretBased {
			generated, err := oauthutils.GenerateOAuth2ClientSecret()
			if err != nil {
				s.logger.Error("Failed to generate client secret", log.Error(err))
				return "", "", &serviceerror.InternalServerError
			}
			clientSecret = generated
		}
	}
	return clientID, clientSecret, nil
}

// isClientIDTaken reports whether the given clientID is already used by a different entity.
func (s *agentService) isClientIDTaken(
	ctx context.Context, clientID, excludeID string) (bool, *serviceerror.ServiceError) {
	id, err := s.entityService.IdentifyEntity(ctx, map[string]interface{}{fieldClientID: clientID})
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return false, nil
		}
		s.logger.Error("Failed to check client ID availability", log.MaskedString("clientID", clientID),
			log.Error(err))
		return false, &serviceerror.InternalServerError
	}
	if id == nil || *id == "" {
		return false, nil
	}
	if excludeID != "" && *id == excludeID {
		return false, nil
	}
	return true, nil
}

// createInboundForAgent creates the inbound client row; applies server defaults via CreateInboundClient.
func (s *agentService) createInboundForAgent(ctx context.Context, agentID string,
	req *model.CreateAgentRequest, clientSecret string) (
	inboundmodel.InboundClient, *inboundmodel.OAuthProfileData, *serviceerror.ServiceError) {
	client := buildInboundClientRecord(agentID, req.AuthFlowID, req.RegistrationFlowID,
		req.IsRegistrationFlowEnabled, req.ThemeID, req.LayoutID, req.Assertion,
		req.LoginConsent, req.AllowedUserTypes)

	oauthData := buildOAuthProfileData(req.InboundAuthConfig)

	hasSecret := clientSecret != ""
	if err := s.inboundClientService.CreateInboundClient(ctx, &client, req.Certificate,
		oauthData, hasSecret, req.Name); err != nil {
		if mapped := s.translateInboundClientError(err); mapped != nil {
			return inboundmodel.InboundClient{}, nil, mapped
		}
		s.logger.Error("Failed to create inbound client for agent",
			log.String("agentID", agentID), log.Error(err))
		return inboundmodel.InboundClient{}, nil, &serviceerror.InternalServerError
	}
	return client, oauthData, nil
}

// reconcileInboundForUpdate creates, updates, or removes the inbound client row and returns the mutated structs.
func (s *agentService) reconcileInboundForUpdate(ctx context.Context, agentID string,
	req *model.UpdateAgentRequest, clientID, clientSecret, oldName, newName string,
) (inboundmodel.InboundClient, *inboundmodel.OAuthProfileData, *serviceerror.ServiceError) {
	wantsInbound := updateNeedsInboundClient(req)

	existingClient, getErr := s.inboundClientService.GetInboundClientByEntityID(ctx, agentID)
	hasExisting := getErr == nil && existingClient != nil

	if !wantsInbound {
		if hasExisting {
			if err := s.inboundClientService.DeleteInboundClient(ctx, agentID); err != nil &&
				!errors.Is(err, inboundclient.ErrInboundClientNotFound) {
				s.logger.Error("Failed to remove inbound client during update",
					log.String("agentID", agentID), log.Error(err))
				return inboundmodel.InboundClient{}, nil, &serviceerror.InternalServerError
			}
		}
		return inboundmodel.InboundClient{}, nil, nil
	}

	client := buildInboundClientRecord(agentID, req.AuthFlowID, req.RegistrationFlowID,
		req.IsRegistrationFlowEnabled, req.ThemeID, req.LayoutID, req.Assertion,
		req.LoginConsent, req.AllowedUserTypes)
	oauthData := buildOAuthProfileData(req.InboundAuthConfig)
	hasSecret := clientSecret != ""

	if hasExisting {
		entityName := newName
		if entityName == "" {
			entityName = oldName
		}
		if err := s.inboundClientService.UpdateInboundClient(ctx, &client, req.Certificate,
			oauthData, hasSecret, clientID, entityName); err != nil {
			if mapped := s.translateInboundClientError(err); mapped != nil {
				return inboundmodel.InboundClient{}, nil, mapped
			}
			s.logger.Error("Failed to update inbound client",
				log.String("agentID", agentID), log.Error(err))
			return inboundmodel.InboundClient{}, nil, &serviceerror.InternalServerError
		}
		return client, oauthData, nil
	}

	if err := s.inboundClientService.CreateInboundClient(ctx, &client, req.Certificate,
		oauthData, hasSecret, newName); err != nil {
		if mapped := s.translateInboundClientError(err); mapped != nil {
			return inboundmodel.InboundClient{}, nil, mapped
		}
		s.logger.Error("Failed to create inbound client during update",
			log.String("agentID", agentID), log.Error(err))
		return inboundmodel.InboundClient{}, nil, &serviceerror.InternalServerError
	}
	return client, oauthData, nil
}

// composeGetResponse builds the GET response by loading inbound client, OAuth profile, and certificates for the entity.
func (s *agentService) composeGetResponse(ctx context.Context, e *entity.Entity) (
	*model.AgentGetResponse, *serviceerror.ServiceError) {
	name, description, owner, clientID := readSystemAttributes(e.SystemAttributes)

	resp := &model.AgentGetResponse{
		ID:          e.ID,
		OUID:        e.OUID,
		OUHandle:    e.OUHandle,
		Type:        e.Type,
		Name:        name,
		Description: description,
		Owner:       owner,
		ClientID:    clientID,
		Attributes:  e.Attributes,
	}

	inbound, err := s.inboundClientService.GetInboundClientByEntityID(ctx, e.ID)
	if err != nil {
		if !errors.Is(err, inboundclient.ErrInboundClientNotFound) {
			s.logger.Error("Failed to load inbound client for agent",
				log.String("agentID", e.ID), log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		return resp, nil
	}

	resp.AuthFlowID = inbound.AuthFlowID
	resp.RegistrationFlowID = inbound.RegistrationFlowID
	resp.IsRegistrationFlowEnabled = inbound.IsRegistrationFlowEnabled
	resp.ThemeID = inbound.ThemeID
	resp.LayoutID = inbound.LayoutID
	resp.Assertion = inbound.Assertion
	resp.LoginConsent = inbound.LoginConsent
	resp.AllowedUserTypes = inbound.AllowedEntityTypes

	oauth, oauthErr := s.inboundClientService.GetOAuthProfileByEntityID(ctx, e.ID)
	if oauthErr != nil && !errors.Is(oauthErr, inboundclient.ErrInboundClientNotFound) {
		s.logger.Error("Failed to load OAuth profile for agent",
			log.String("agentID", e.ID), log.Error(oauthErr))
		return nil, &serviceerror.InternalServerError
	}
	if oauthErr == nil && oauth != nil && oauth.OAuthProfile != nil {
		resp.InboundAuthConfig = removeSecrets([]model.InboundAuthConfig{
			{
				Type:   model.OAuthInboundAuthType,
				Config: oauthProfileToConfig(clientID, oauth.OAuthProfile),
			},
		})
	}

	entityCert, certOpErr := s.inboundClientService.GetCertificate(ctx, cert.CertificateReferenceTypeApplication, e.ID)
	if certOpErr != nil {
		return nil, s.translateCertOperationError(certOpErr)
	}
	resp.Certificate = entityCert

	if clientID != "" {
		oauthCert, oauthCertOpErr := s.inboundClientService.GetCertificate(
			ctx, cert.CertificateReferenceTypeOAuthApp, clientID)
		if oauthCertOpErr != nil {
			return nil, s.translateCertOperationError(oauthCertOpErr)
		}
		if len(resp.InboundAuthConfig) > 0 && resp.InboundAuthConfig[0].Config != nil {
			resp.InboundAuthConfig[0].Config.Certificate = oauthCert
		}
	}

	return resp, nil
}

// buildListResponse builds the paged agent list response from a slice of entities and pagination metadata.
func (s *agentService) buildListResponse(ctx context.Context, entities []entity.Entity,
	totalCount, limit, offset int, includeDisplay bool) *model.AgentListResponse {
	agents := make([]model.BasicAgentResponse, 0, len(entities))
	for i := range entities {
		e := &entities[i]
		name, description, owner, clientID := readSystemAttributes(e.SystemAttributes)
		agents = append(agents, model.BasicAgentResponse{
			ID:          e.ID,
			OUID:        e.OUID,
			OUHandle:    e.OUHandle,
			Type:        e.Type,
			Name:        name,
			Description: description,
			ClientID:    clientID,
			Owner:       owner,
			Attributes:  e.Attributes,
		})
	}

	if includeDisplay {
		s.populateOUHandlesForList(ctx, agents)
	}

	displayQuery := sysutils.DisplayQueryParam(includeDisplay)
	return &model.AgentListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(agents),
		Agents:       agents,
		Links:        sysutils.BuildPaginationLinks(agentBasePath, limit, offset, totalCount, displayQuery),
	}
}

// lookupOUHandle resolves the OU handle for the given OU ID; returns empty string on failure.
func (s *agentService) lookupOUHandle(ctx context.Context, ouID string) string {
	handles, err := s.ouService.GetOrganizationUnitHandlesByIDs(ctx, []string{ouID})
	if err != nil {
		s.logger.Debug("Failed to resolve OU handle for agent", log.String("ouID", ouID), log.Any("error", err))
		return ""
	}
	return handles[ouID]
}

// populateOUHandleForGet resolves and sets OUHandle on a single-agent GET response; silently skips on lookup failure.
func (s *agentService) populateOUHandleForGet(ctx context.Context, resp *model.AgentGetResponse) {
	if resp.OUID == "" || resp.OUHandle != "" {
		return
	}
	resp.OUHandle = s.lookupOUHandle(ctx, resp.OUID)
}

// populateOUHandleForComplete sets OUHandle on a complete-agent response; silently skips on lookup failure.
func (s *agentService) populateOUHandleForComplete(ctx context.Context, resp *model.AgentCompleteResponse) {
	if resp.OUID == "" {
		return
	}
	resp.OUHandle = s.lookupOUHandle(ctx, resp.OUID)
}

// populateOUHandlesForList batch-resolves OU handles for a list of agents, filling in OUHandle where available.
func (s *agentService) populateOUHandlesForList(ctx context.Context, agents []model.BasicAgentResponse) {
	if len(agents) == 0 {
		return
	}
	idSet := make(map[string]struct{}, len(agents))
	ids := make([]string, 0, len(agents))
	for _, a := range agents {
		if a.OUID == "" {
			continue
		}
		if _, ok := idSet[a.OUID]; ok {
			continue
		}
		idSet[a.OUID] = struct{}{}
		ids = append(ids, a.OUID)
	}
	if len(ids) == 0 {
		return
	}
	handles, err := s.ouService.GetOrganizationUnitHandlesByIDs(ctx, ids)
	if err != nil {
		s.logger.Debug("Failed to resolve OU handles for agent list", log.Any("error", err))
		return
	}
	for i := range agents {
		if h, ok := handles[agents[i].OUID]; ok {
			agents[i].OUHandle = h
		}
	}
}

// needsInboundClient reports whether any inbound auth field in the create request requires an inbound client row.
func needsInboundClient(req *model.CreateAgentRequest) bool {
	if req == nil {
		return false
	}
	return req.AuthFlowID != "" ||
		req.RegistrationFlowID != "" ||
		req.IsRegistrationFlowEnabled ||
		req.ThemeID != "" ||
		req.LayoutID != "" ||
		req.Assertion != nil ||
		req.LoginConsent != nil ||
		len(req.AllowedUserTypes) > 0 ||
		req.Certificate != nil ||
		len(req.InboundAuthConfig) > 0
}

// updateNeedsInboundClient reports whether an update request contains any inbound auth field requiring a client row.
func updateNeedsInboundClient(req *model.UpdateAgentRequest) bool {
	if req == nil {
		return false
	}
	return req.AuthFlowID != "" ||
		req.RegistrationFlowID != "" ||
		req.IsRegistrationFlowEnabled ||
		req.ThemeID != "" ||
		req.LayoutID != "" ||
		req.Assertion != nil ||
		req.LoginConsent != nil ||
		len(req.AllowedUserTypes) > 0 ||
		req.Certificate != nil ||
		len(req.InboundAuthConfig) > 0
}

// validateBaseFields validates the mandatory top-level fields required for both create and update.
func validateBaseFields(name, agentType string) *serviceerror.ServiceError {
	if name == "" {
		return &ErrorInvalidAgentName
	}
	if agentType == "" {
		return &ErrorInvalidAgentType
	}
	return nil
}

// validatePaginationParams validates that limit and offset are within acceptable bounds.
func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 0 || limit > 100 {
		return &ErrorInvalidLimit
	}
	if offset < 0 {
		return &ErrorInvalidOffset
	}
	return nil
}

// normalizeLoginConsent clamps a negative ValidityPeriod to zero; leaves a nil config untouched.
func normalizeLoginConsent(lc *inboundmodel.LoginConsentConfig) {
	if lc == nil {
		return
	}
	if lc.ValidityPeriod < 0 {
		lc.ValidityPeriod = 0
	}
}

// pickOAuthConfig returns the first OAuth-typed entry, or nil.
func pickOAuthConfig(configs []model.InboundAuthConfig) *model.OAuthAgentConfig {
	for i := range configs {
		if configs[i].Type == model.OAuthInboundAuthType && configs[i].Config != nil {
			return configs[i].Config
		}
	}
	return nil
}

// requiresClientSecret reports whether the OAuth config implies a confidential client requiring a secret.
func requiresClientSecret(cfg *model.OAuthAgentConfig) bool {
	if cfg == nil {
		return false
	}
	if cfg.PublicClient {
		return false
	}
	switch cfg.TokenEndpointAuthMethod {
	case oauth2const.TokenEndpointAuthMethodClientSecretBasic,
		oauth2const.TokenEndpointAuthMethodClientSecretPost:
		return true
	case oauth2const.TokenEndpointAuthMethodNone,
		oauth2const.TokenEndpointAuthMethodPrivateKeyJWT:
		return false
	}
	// Default to client_secret_basic when unspecified.
	return true
}

// removeSecrets returns a copy of the configs with clientSecret stripped.
func removeSecrets(in []model.InboundAuthConfig) []model.InboundAuthConfig {
	if len(in) == 0 {
		return nil
	}
	out := make([]model.InboundAuthConfig, len(in))
	for i, cfg := range in {
		copyCfg := cfg
		if copyCfg.Config != nil {
			c := *copyCfg.Config
			c.ClientSecret = ""
			copyCfg.Config = &c
		}
		out[i] = copyCfg
	}
	return out
}

// buildAgentEntity constructs the entity row and system credentials JSON for a new or updated agent.
func buildAgentEntity(agentID, agentType, ouID string, attributes json.RawMessage,
	name, description, owner, clientID, clientSecret string) (*entity.Entity, json.RawMessage, error) {
	sysAttrsJSON, err := buildSystemAttributesJSON(name, description, owner, clientID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build agent system attributes: %w", err)
	}

	sysCredsJSON, err := buildSystemCredentialsJSON(clientSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build agent system credentials: %w", err)
	}

	e := &entity.Entity{
		ID:               agentID,
		Category:         entity.EntityCategoryAgent,
		Type:             agentType,
		State:            entity.EntityStateActive,
		OUID:             ouID,
		Attributes:       attributes,
		SystemAttributes: sysAttrsJSON,
	}
	return e, sysCredsJSON, nil
}

// buildSystemAttributesJSON serializes agent name, description, owner, and clientID into the systemAttributes blob.
func buildSystemAttributesJSON(name, description, owner, clientID string) (json.RawMessage, error) {
	attrs := map[string]interface{}{}
	if name != "" {
		attrs[fieldName] = name
	}
	if description != "" {
		attrs[fieldDescription] = description
	}
	if owner != "" {
		attrs[fieldOwner] = owner
	}
	if clientID != "" {
		attrs[fieldClientID] = clientID
	}
	if len(attrs) == 0 {
		return nil, nil
	}
	return json.Marshal(attrs)
}

// buildSystemCredentialsJSON serializes the client secret into the systemCredentials JSON blob; returns nil when empty.
func buildSystemCredentialsJSON(clientSecret string) (json.RawMessage, error) {
	if clientSecret == "" {
		return nil, nil
	}
	return json.Marshal(map[string]interface{}{
		fieldClientSecret: clientSecret,
	})
}

// readSystemAttributes deserializes the systemAttributes JSON blob back into individual string fields.
func readSystemAttributes(raw json.RawMessage) (name, description, owner, clientID string) {
	if len(raw) == 0 {
		return "", "", "", ""
	}
	var attrs map[string]interface{}
	if err := json.Unmarshal(raw, &attrs); err != nil {
		return "", "", "", ""
	}
	if v, ok := attrs[fieldName].(string); ok {
		name = v
	}
	if v, ok := attrs[fieldDescription].(string); ok {
		description = v
	}
	if v, ok := attrs[fieldOwner].(string); ok {
		owner = v
	}
	if v, ok := attrs[fieldClientID].(string); ok {
		clientID = v
	}
	return name, description, owner, clientID
}

// buildInboundClientRecord constructs an InboundClient record from the agent's identity and inbound auth fields.
func buildInboundClientRecord(agentID, authFlowID, regFlowID string, isRegEnabled bool,
	themeID, layoutID string, assertion *inboundmodel.AssertionConfig,
	loginConsent *inboundmodel.LoginConsentConfig, allowedUserTypes []string) inboundmodel.InboundClient {
	return inboundmodel.InboundClient{
		ID:                        agentID,
		AuthFlowID:                authFlowID,
		RegistrationFlowID:        regFlowID,
		IsRegistrationFlowEnabled: isRegEnabled,
		ThemeID:                   themeID,
		LayoutID:                  layoutID,
		Assertion:                 assertion,
		LoginConsent:              loginConsent,
		AllowedEntityTypes:        allowedUserTypes,
	}
}

// buildOAuthProfileData maps the agent OAuth config to the inbound client profile shape.
func buildOAuthProfileData(configs []model.InboundAuthConfig) *inboundmodel.OAuthProfileData {
	cfg := pickOAuthConfig(configs)
	if cfg == nil {
		return nil
	}
	authMethod := cfg.TokenEndpointAuthMethod
	if authMethod == "" {
		authMethod = oauth2const.TokenEndpointAuthMethodClientSecretBasic
	}
	grantTypes := sysutils.ConvertToStringSlice(cfg.GrantTypes)
	if len(grantTypes) == 0 {
		// Default to client_credentials for agents.
		grantTypes = []string{string(oauth2const.GrantTypeClientCredentials)}
	}
	return &inboundmodel.OAuthProfileData{
		RedirectURIs:                       cfg.RedirectURIs,
		GrantTypes:                         grantTypes,
		ResponseTypes:                      sysutils.ConvertToStringSlice(cfg.ResponseTypes),
		TokenEndpointAuthMethod:            string(authMethod),
		PKCERequired:                       cfg.PKCERequired,
		PublicClient:                       cfg.PublicClient,
		RequirePushedAuthorizationRequests: cfg.RequirePushedAuthorizationRequests,
		Certificate:                        cfg.Certificate,
		Token:                              cfg.Token,
		Scopes:                             cfg.Scopes,
		UserInfo:                           cfg.UserInfo,
		ScopeClaims:                        cfg.ScopeClaims,
	}
}

// oauthProfileToConfig converts a stored OAuth profile back into the agent-facing config.
func oauthProfileToConfig(clientID string, p *inboundmodel.OAuthProfileData) *model.OAuthAgentConfig {
	if p == nil {
		return nil
	}
	grants := make([]oauth2const.GrantType, 0, len(p.GrantTypes))
	for _, g := range p.GrantTypes {
		grants = append(grants, oauth2const.GrantType(g))
	}
	respTypes := make([]oauth2const.ResponseType, 0, len(p.ResponseTypes))
	for _, r := range p.ResponseTypes {
		respTypes = append(respTypes, oauth2const.ResponseType(r))
	}
	return &model.OAuthAgentConfig{
		ClientID:                           clientID,
		RedirectURIs:                       p.RedirectURIs,
		GrantTypes:                         grants,
		ResponseTypes:                      respTypes,
		TokenEndpointAuthMethod:            oauth2const.TokenEndpointAuthMethod(p.TokenEndpointAuthMethod),
		PKCERequired:                       p.PKCERequired,
		PublicClient:                       p.PublicClient,
		RequirePushedAuthorizationRequests: p.RequirePushedAuthorizationRequests,
		Certificate:                        p.Certificate,
		Token:                              p.Token,
		Scopes:                             p.Scopes,
		UserInfo:                           p.UserInfo,
		ScopeClaims:                        p.ScopeClaims,
	}
}

// buildCompleteResponse constructs the full create/update response including credentials and all inbound auth fields.
func buildCompleteResponse(agentID, owner, clientID, clientSecret, agentType, name, description string,
	attributes json.RawMessage, authFlowID, regFlowID string, isRegEnabled bool,
	themeID, layoutID string, assertion *inboundmodel.AssertionConfig,
	loginConsent *inboundmodel.LoginConsentConfig, allowedUserTypes []string,
	certificate *inboundmodel.Certificate, inboundAuthConfig []model.InboundAuthConfig,
) *model.AgentCompleteResponse {
	resp := &model.AgentCompleteResponse{
		ID:                        agentID,
		Type:                      agentType,
		Name:                      name,
		Description:               description,
		Owner:                     owner,
		Attributes:                attributes,
		AuthFlowID:                authFlowID,
		RegistrationFlowID:        regFlowID,
		IsRegistrationFlowEnabled: isRegEnabled,
		ThemeID:                   themeID,
		LayoutID:                  layoutID,
		Assertion:                 assertion,
		LoginConsent:              loginConsent,
		AllowedUserTypes:          allowedUserTypes,
		Certificate:               certificate,
	}
	if len(inboundAuthConfig) > 0 {
		resp.InboundAuthConfig = annotateOAuthConfig(inboundAuthConfig, clientID, clientSecret)
	}
	return resp
}

// annotateOAuthConfig stamps clientID and clientSecret onto the OAuth entry.
func annotateOAuthConfig(in []model.InboundAuthConfig, clientID, clientSecret string) []model.InboundAuthConfig {
	out := make([]model.InboundAuthConfig, len(in))
	for i, cfg := range in {
		copyCfg := cfg
		if copyCfg.Type == model.OAuthInboundAuthType && copyCfg.Config != nil {
			c := *copyCfg.Config
			if clientID != "" {
				c.ClientID = clientID
			}
			if clientSecret != "" {
				c.ClientSecret = clientSecret
			}
			copyCfg.Config = &c
		}
		out[i] = copyCfg
	}
	return out
}

// mapEntityError maps entity-layer errors to agent-service errors.
func mapEntityError(err error) *serviceerror.ServiceError {
	switch {
	case errors.Is(err, entity.ErrEntityNotFound):
		return &ErrorAgentNotFound
	case errors.Is(err, entity.ErrSchemaValidationFailed):
		return &ErrorSchemaValidationFailed
	case errors.Is(err, entity.ErrAttributeConflict):
		return &ErrorAttributeConflict
	case errors.Is(err, entity.ErrInvalidCredential):
		return &ErrorInvalidCredential
	}
	return nil
}

// translateInboundClientError maps inbound-client-layer errors to agent-service errors.
func (s *agentService) translateInboundClientError(err error) *serviceerror.ServiceError {
	if err == nil {
		return nil
	}
	if errors.Is(err, inboundclient.ErrCannotModifyDeclarative) {
		return &ErrorCannotModifyDeclarativeResource
	}
	if svcErr := translateInboundClientFKError(err); svcErr != nil {
		return svcErr
	}
	if svcErr := translateOAuthValidationError(err); svcErr != nil {
		return svcErr
	}
	if svcErr := translateCertValidationError(err); svcErr != nil {
		return svcErr
	}
	var opErr *inboundclient.CertOperationError
	if errors.As(err, &opErr) {
		return s.translateCertOperationError(opErr)
	}
	var consentErr *inboundclient.ConsentSyncError
	if errors.As(err, &consentErr) {
		if consentErr.IsClientError() {
			return serviceerror.CustomServiceError(ErrorConsentSyncFailed, core.I18nMessage{
				DefaultValue: "Consent sync failed: " + consentErr.Underlying.Code,
			})
		}
	}
	return nil
}

// translateCertOperationError maps a CertOperationError to a service error, logging server-side failures.
func (s *agentService) translateCertOperationError(err *inboundclient.CertOperationError) *serviceerror.ServiceError {
	if !err.IsClientError() {
		s.logger.Error("Certificate operation failed",
			log.Any("operation", err.Operation),
			log.Any("refType", err.RefType),
			log.Any("serviceError", err.Underlying))
		return &serviceerror.InternalServerError
	}
	var prefix string
	switch err.Operation {
	case inboundclient.CertOpCreate:
		prefix = "Failed to create agent certificate: "
	case inboundclient.CertOpUpdate:
		prefix = "Failed to update agent certificate: "
	case inboundclient.CertOpRetrieve:
		prefix = "Failed to retrieve agent certificate: "
	case inboundclient.CertOpDelete:
		prefix = "Failed to delete agent certificate: "
	default:
		return &serviceerror.InternalServerError
	}
	return serviceerror.CustomServiceError(ErrorCertificateClientError, core.I18nMessage{
		DefaultValue: prefix + err.Underlying.ErrorDescription.DefaultValue,
	})
}

// translateInboundClientFKError maps inbound client foreign-key errors (flows, themes, user types) to service errors.
func translateInboundClientFKError(err error) *serviceerror.ServiceError {
	switch {
	case errors.Is(err, inboundclient.ErrFKInvalidAuthFlow):
		return &ErrorInvalidAuthFlowID
	case errors.Is(err, inboundclient.ErrFKInvalidRegistrationFlow):
		return &ErrorInvalidRegistrationFlowID
	case errors.Is(err, inboundclient.ErrFKFlowDefinitionRetrievalFailed):
		return &ErrorWhileRetrievingFlowDefinition
	case errors.Is(err, inboundclient.ErrFKFlowServerError):
		return &serviceerror.InternalServerError
	case errors.Is(err, inboundclient.ErrFKThemeNotFound):
		return &ErrorThemeNotFound
	case errors.Is(err, inboundclient.ErrFKLayoutNotFound):
		return &ErrorLayoutNotFound
	case errors.Is(err, inboundclient.ErrFKInvalidUserType):
		return &ErrorInvalidUserType
	default:
		return nil
	}
}

// translateOAuthValidationError maps OAuth validation sentinel errors to service errors.
func translateOAuthValidationError(err error) *serviceerror.ServiceError {
	switch {
	case errors.Is(err, inboundclient.ErrOAuthInvalidRedirectURI),
		errors.Is(err, inboundclient.ErrOAuthRedirectURIFragmentNotAllowed),
		errors.Is(err, inboundclient.ErrOAuthAuthCodeRequiresRedirectURIs):
		return &ErrorInvalidRedirectURI
	case errors.Is(err, inboundclient.ErrOAuthInvalidGrantType),
		errors.Is(err, inboundclient.ErrOAuthRefreshTokenCannotBeSoleGrant),
		errors.Is(err, inboundclient.ErrOAuthClientCredentialsCannotUseResponseTypes),
		errors.Is(err, inboundclient.ErrOAuthAuthCodeRequiresCodeResponseType):
		return &ErrorInvalidGrantType
	case errors.Is(err, inboundclient.ErrOAuthInvalidResponseType):
		return &ErrorInvalidResponseType
	case errors.Is(err, inboundclient.ErrOAuthInvalidTokenEndpointAuthMethod):
		return &ErrorInvalidTokenEndpointAuthMethod
	case errors.Is(err, inboundclient.ErrOAuthPublicClientMustUseNoneAuth),
		errors.Is(err, inboundclient.ErrOAuthPublicClientMustHavePKCE):
		return &ErrorInvalidPublicClientConfiguration
	case errors.Is(err, inboundclient.ErrOAuthPKCERequiresAuthCode),
		errors.Is(err, inboundclient.ErrOAuthResponseTypesRequireAuthCode),
		errors.Is(err, inboundclient.ErrOAuthPrivateKeyJWTRequiresCertificate),
		errors.Is(err, inboundclient.ErrOAuthPrivateKeyJWTCannotHaveClientSecret),
		errors.Is(err, inboundclient.ErrOAuthClientSecretCannotHaveCertificate),
		errors.Is(err, inboundclient.ErrOAuthNoneAuthRequiresPublicClient),
		errors.Is(err, inboundclient.ErrOAuthNoneAuthCannotHaveCertOrSecret),
		errors.Is(err, inboundclient.ErrOAuthClientCredentialsCannotUseNoneAuth),
		errors.Is(err, inboundclient.ErrOAuthUserInfoUnsupportedSigningAlg),
		errors.Is(err, inboundclient.ErrOAuthUserInfoUnsupportedEncryptionAlg),
		errors.Is(err, inboundclient.ErrOAuthUserInfoUnsupportedEncryptionEnc),
		errors.Is(err, inboundclient.ErrOAuthUserInfoEncryptionAlgRequiresEnc),
		errors.Is(err, inboundclient.ErrOAuthUserInfoEncryptionEncRequiresAlg),
		errors.Is(err, inboundclient.ErrOAuthUserInfoEncryptionRequiresCertificate),
		errors.Is(err, inboundclient.ErrOAuthUserInfoJWKSURINotSSRFSafe),
		errors.Is(err, inboundclient.ErrOAuthUserInfoUnsupportedResponseType),
		errors.Is(err, inboundclient.ErrOAuthUserInfoJWSRequiresSigningAlg),
		errors.Is(err, inboundclient.ErrOAuthUserInfoJWERequiresEncryption),
		errors.Is(err, inboundclient.ErrOAuthUserInfoNestedJWTRequiresAll):
		return &ErrorInvalidOAuthConfiguration
	default:
		return nil
	}
}

// translateCertValidationError maps certificate validation sentinel errors to service errors.
func translateCertValidationError(err error) *serviceerror.ServiceError {
	switch {
	case errors.Is(err, inboundclient.ErrCertValueRequired):
		return &ErrorInvalidCertificateValue
	case errors.Is(err, inboundclient.ErrCertInvalidJWKSURI):
		return &ErrorInvalidJWKSURI
	case errors.Is(err, inboundclient.ErrCertInvalidType):
		return &ErrorInvalidCertificateType
	default:
		return nil
	}
}
