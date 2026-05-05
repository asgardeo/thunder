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

package application

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"encoding/json"

	"github.com/asgardeo/thunder/internal/cert"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/i18n/core"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/pkg/application/model"
	"github.com/asgardeo/thunder/pkg/entityprovider"
	"github.com/asgardeo/thunder/pkg/inboundclient"
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
	oauthutils "github.com/asgardeo/thunder/pkg/oauth/oauth2/utils"
)

// ApplicationServiceInterface defines the interface for the application service.
type ApplicationServiceInterface interface {
	CreateApplication(
		ctx context.Context, app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	ValidateApplication(ctx context.Context, app *model.ApplicationDTO) (
		*model.ApplicationProcessedDTO, *model.InboundAuthConfigDTO, *serviceerror.ServiceError)
	GetApplicationList(ctx context.Context) (*model.ApplicationListResponse, *serviceerror.ServiceError)
	GetOAuthApplication(
		ctx context.Context, clientID string) (*inboundmodel.OAuthClient, *serviceerror.ServiceError)
	GetApplication(ctx context.Context, appID string) (*model.Application, *serviceerror.ServiceError)
	UpdateApplication(
		ctx context.Context, appID string, app *model.ApplicationDTO) (
		*model.ApplicationDTO, *serviceerror.ServiceError)
	DeleteApplication(ctx context.Context, appID string) *serviceerror.ServiceError
}

// ApplicationService is the default implementation of the ApplicationServiceInterface.
type applicationService struct {
	logger               *log.Logger
	inboundClientService inboundclient.InboundClientServiceInterface
	entityProvider       entityprovider.EntityProviderInterface
	ouService            oupkg.OrganizationUnitServiceInterface
	i18nService          i18nmgt.I18nServiceInterface
}

// newApplicationService creates a new instance of ApplicationService.
func newApplicationService(
	inboundClient inboundclient.InboundClientServiceInterface,
	entityProvider entityprovider.EntityProviderInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	i18nService i18nmgt.I18nServiceInterface,
) ApplicationServiceInterface {
	return &applicationService{
		logger:               log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService")),
		inboundClientService: inboundClient,
		entityProvider:       entityProvider,
		ouService:            ouService,
		i18nService:          i18nService,
	}
}

func (as *applicationService) deleteEntityCompensation(appID string) {
	if delErr := as.entityProvider.DeleteEntity(appID); delErr != nil {
		as.logger.Error("Failed to delete entity during compensation", log.Error(delErr),
			log.String("appID", appID))
	}
}

// CreateApplication creates the application.
func (as *applicationService) CreateApplication(ctx context.Context, app *model.ApplicationDTO) (*model.ApplicationDTO,
	*serviceerror.ServiceError) {
	if app == nil {
		return nil, &ErrorApplicationNil
	}
	// Check if store is in pure declarative mode
	if isDeclarativeModeEnabled() {
		return nil, &ErrorCannotModifyDeclarativeResource
	}

	// Check if an application with the same ID exists and is declarative (in composite mode)
	if app.ID != "" && as.inboundClientService.IsDeclarative(ctx, app.ID) {
		return nil, &ErrorCannotModifyDeclarativeResource
	}

	processedDTO, inboundAuthConfig, svcErr := as.ValidateApplication(ctx, app)
	if svcErr != nil {
		return nil, svcErr
	}

	appID := processedDTO.ID

	configDAO := toConfigDAO(processedDTO)
	oauthData := toOAuthConfigData(processedDTO)

	// Create entity.
	var clientID string
	var clientSecret string
	if inboundAuthConfig != nil && inboundAuthConfig.OAuthAppConfig != nil {
		clientID = inboundAuthConfig.OAuthAppConfig.ClientID
		clientSecret = inboundAuthConfig.OAuthAppConfig.ClientSecret
	}

	appEntity, sysCredsJSON, buildErr := buildAppEntity(appID, app, clientID, clientSecret)
	if buildErr != nil {
		as.logger.Error("Failed to build entity for create", log.Error(buildErr))
		return nil, &serviceerror.InternalServerError
	}

	_, epErr := as.entityProvider.CreateEntity(appEntity, sysCredsJSON)
	if epErr != nil {
		if svcErr := mapEntityProviderError(epErr); svcErr != nil {
			return nil, svcErr
		}
		as.logger.Error("Failed to create application entity", log.String("appID", appID), log.Error(epErr))
		return nil, &serviceerror.InternalServerError
	}

	// Create config (with compensation if it fails).
	if err := as.inboundClientService.CreateInboundClient(ctx, &configDAO, app.Certificate, oauthData,
		clientSecret != "", app.Name); err != nil {
		// Compensate: delete entity since config creation failed.
		as.deleteEntityCompensation(appID)
		if svcErr := translateInboundClientError(err); svcErr != nil {
			return nil, svcErr
		}
		if svcErr := as.translateCertError(err); svcErr != nil {
			return nil, svcErr
		}
		as.logger.Error("Failed to create application", log.Error(err), log.String("appID", appID))
		return nil, &serviceerror.InternalServerError
	}

	appForReturn := *app
	appForReturn.AuthFlowID = configDAO.AuthFlowID
	appForReturn.RegistrationFlowID = configDAO.RegistrationFlowID
	if app.Certificate == nil || app.Certificate.Type == "" {
		appForReturn.Certificate = nil
	}
	var oauthToken *inboundmodel.OAuthTokenConfig
	var userInfo *inboundmodel.UserInfoConfig
	var scopeClaims map[string][]string
	if inboundAuthConfig != nil && oauthData != nil {
		oauthToken = oauthData.Token
		userInfo = oauthData.UserInfo
		scopeClaims = oauthData.ScopeClaims
		oauthCfg := inboundAuthConfig.OAuthAppConfig
		if oauthCfg != nil &&
			(oauthCfg.Certificate == nil || oauthCfg.Certificate.Type == "") {
			oauthCfg.Certificate = nil
		}
	}
	return buildReturnApplicationDTO(appID, &appForReturn, configDAO.Assertion, processedDTO.Metadata,
		inboundAuthConfig, oauthToken, userInfo, scopeClaims), nil
}

// ValidateApplication validates the application data transfer object.
func (as *applicationService) ValidateApplication(ctx context.Context, app *model.ApplicationDTO) (
	*model.ApplicationProcessedDTO, *model.InboundAuthConfigDTO, *serviceerror.ServiceError) {
	if app == nil {
		return nil, nil, &ErrorApplicationNil
	}
	if app.Name == "" {
		return nil, nil, &ErrorInvalidApplicationName
	}
	nameExists, nameCheckErr := as.isIdentifierTaken(fieldName, app.Name, app.ID)
	if nameCheckErr != nil {
		return nil, nil, nameCheckErr
	}
	if nameExists {
		return nil, nil, &ErrorApplicationAlreadyExistsWithName
	}

	inboundAuthConfig, svcErr := as.processInboundAuthConfig(app, nil)
	if svcErr != nil {
		return nil, nil, svcErr
	}

	if svcErr := as.validateApplicationFields(ctx, app); svcErr != nil {
		return nil, nil, svcErr
	}

	appID := app.ID
	if appID == "" {
		var err error
		appID, err = sysutils.GenerateUUIDv7()
		if err != nil {
			as.logger.Error("Failed to generate UUID", log.Error(err))
			return nil, nil, &serviceerror.InternalServerError
		}
	}
	processedDTO := buildBaseApplicationProcessedDTO(appID, app, app.Assertion)
	if inboundAuthConfig != nil {
		oa := inboundAuthConfig.OAuthAppConfig
		processedInboundAuthConfig := buildOAuthInboundAuthConfigProcessedDTO(
			appID, inboundAuthConfig, oa.Token, oa.UserInfo, oa.ScopeClaims, oa.Certificate,
		)
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
	}

	// Validate FK constraints (flow, theme, layout, user-type) and OAuth profile.
	// This runs the same checks as Create/Update so declarative resources are validated consistently.
	configDAO := toConfigDAO(processedDTO)
	oauthValidateData := toOAuthConfigData(processedDTO)
	var hasClientSecret bool
	if inboundAuthConfig != nil && inboundAuthConfig.OAuthAppConfig != nil {
		hasClientSecret = inboundAuthConfig.OAuthAppConfig.ClientSecret != ""
	}
	if err := as.inboundClientService.Validate(ctx, &configDAO, oauthValidateData, hasClientSecret); err != nil {
		if svcErr := translateInboundClientError(err); svcErr != nil {
			return nil, nil, svcErr
		}
		as.logger.Error("Inbound client validation failed", log.Error(err))
		return nil, nil, &serviceerror.InternalServerError
	}
	processedDTO.AuthFlowID = configDAO.AuthFlowID
	processedDTO.RegistrationFlowID = configDAO.RegistrationFlowID

	return processedDTO, inboundAuthConfig, nil
}

// GetApplicationList list the applications.
func (as *applicationService) GetApplicationList(
	ctx context.Context) (*model.ApplicationListResponse, *serviceerror.ServiceError) {
	configs, err := as.inboundClientService.GetInboundClientList(ctx)
	if err != nil {
		if errors.Is(err, inboundclient.ErrCompositeResultLimitExceeded) {
			return nil, &ErrorResultLimitExceeded
		}
		as.logger.Error("Failed to list inbound clients", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if len(configs) == 0 {
		return &model.ApplicationListResponse{
			TotalResults: 0,
			Count:        0,
			Applications: []model.BasicApplicationResponse{},
		}, nil
	}

	entityIDs := make([]string, 0, len(configs))
	for _, cfg := range configs {
		entityIDs = append(entityIDs, cfg.ID)
	}
	entities, epErr := as.entityProvider.GetEntitiesByIDs(entityIDs)
	if epErr != nil {
		as.logger.Error("Failed to retrieve entities for application list", log.Error(epErr))
		return nil, &serviceerror.InternalServerError
	}
	entityMap := make(map[string]*entityprovider.Entity, len(entities))
	for i := range entities {
		entityMap[entities[i].ID] = &entities[i]
	}

	applicationList := make([]model.BasicApplicationResponse, 0, len(configs))
	for i := range configs {
		applicationList = append(applicationList, buildBasicApplicationResponse(configs[i], entityMap[configs[i].ID]))
	}

	return &model.ApplicationListResponse{
		TotalResults: len(configs),
		Count:        len(applicationList),
		Applications: applicationList,
	}, nil
}

// GetOAuthApplication retrieves the OAuth application based on the client id.
func (as *applicationService) GetOAuthApplication(
	ctx context.Context, clientID string) (*inboundmodel.OAuthClient, *serviceerror.ServiceError) {
	if clientID == "" {
		return nil, &ErrorInvalidClientID
	}

	client, err := as.inboundClientService.GetOAuthClientByClientID(ctx, clientID)
	if err != nil {
		as.logger.Error("Failed to retrieve OAuth client", log.Error(err),
			log.MaskedString("clientID", clientID))
		return nil, &serviceerror.InternalServerError
	}
	if client == nil {
		return nil, &ErrorApplicationNotFound
	}
	return client, nil
}

// GetApplication get the application for given app id.
func (as *applicationService) GetApplication(ctx context.Context, appID string) (*model.Application,
	*serviceerror.ServiceError) {
	if appID == "" {
		return nil, &ErrorInvalidApplicationID
	}

	fullApp, svcErr := as.getApplication(ctx, appID)
	if svcErr != nil {
		return nil, svcErr
	}

	return as.enrichApplicationWithCertificate(ctx, buildApplicationResponse(fullApp))
}

// UpdateApplication update the application for given app id.
func (as *applicationService) UpdateApplication(ctx context.Context, appID string, app *model.ApplicationDTO) (
	*model.ApplicationDTO, *serviceerror.ServiceError) {
	if appID == "" {
		return nil, &ErrorInvalidApplicationID
	}
	if as.inboundClientService.IsDeclarative(ctx, appID) {
		return nil, &ErrorCannotModifyDeclarativeResource
	}
	existingApp, inboundAuthConfig, svcErr := as.validateApplicationForUpdate(ctx, appID, app)

	if svcErr != nil {
		return nil, svcErr
	}

	processedDTO := as.buildProcessedDTOForUpdate(appID, app, inboundAuthConfig)

	configDAO := toConfigDAO(processedDTO)
	oauthData := toOAuthConfigData(processedDTO)

	var newOAuthClientID string
	if inboundAuthConfig != nil && inboundAuthConfig.OAuthAppConfig != nil {
		newOAuthClientID = inboundAuthConfig.OAuthAppConfig.ClientID
	}
	oauthSecretSupplied := inboundAuthConfig != nil &&
		inboundAuthConfig.OAuthAppConfig != nil &&
		inboundAuthConfig.OAuthAppConfig.ClientSecret != ""
	// Update config first, while entity attributes still hold the previous client_id so the
	// inbound client service can clean up the old OAuth-app cert.
	if err := as.inboundClientService.UpdateInboundClient(
		ctx, &configDAO, app.Certificate, oauthData, oauthSecretSupplied, newOAuthClientID, app.Name,
	); err != nil {
		if svcErr := translateInboundClientError(err); svcErr != nil {
			return nil, svcErr
		}
		if svcErr := as.translateCertError(err); svcErr != nil {
			return nil, svcErr
		}
		as.logger.Error("Failed to update application", log.Error(err), log.String("appID", appID))
		return nil, &serviceerror.InternalServerError
	}

	if svcErr := as.updateEntityDataForApplicationUpdate(appID, app, inboundAuthConfig); svcErr != nil {
		return nil, svcErr
	}

	if svcErr := as.cleanupStaleI18nKeys(ctx, appID, existingApp, app); svcErr != nil {
		return nil, svcErr
	}

	appForReturn := *app
	appForReturn.AuthFlowID = configDAO.AuthFlowID
	appForReturn.RegistrationFlowID = configDAO.RegistrationFlowID
	if app.Certificate == nil || app.Certificate.Type == "" {
		appForReturn.Certificate = nil
	}
	var oauthToken *inboundmodel.OAuthTokenConfig
	var userInfo *inboundmodel.UserInfoConfig
	var scopeClaims map[string][]string
	if oauthData != nil {
		oauthToken = oauthData.Token
		userInfo = oauthData.UserInfo
		scopeClaims = oauthData.ScopeClaims
	}
	if inboundAuthConfig != nil && inboundAuthConfig.OAuthAppConfig != nil {
		c := inboundAuthConfig.OAuthAppConfig.Certificate
		if c == nil || c.Type == "" {
			inboundAuthConfig.OAuthAppConfig.Certificate = nil
		}
	}
	return buildReturnApplicationDTO(appID, &appForReturn, configDAO.Assertion, processedDTO.Metadata,
		inboundAuthConfig, oauthToken, userInfo, scopeClaims), nil
}

func (as *applicationService) updateEntityDataForApplicationUpdate(
	appID string,
	app *model.ApplicationDTO,
	inboundAuthConfig *model.InboundAuthConfigDTO,
) *serviceerror.ServiceError {
	var clientID string
	if inboundAuthConfig != nil && inboundAuthConfig.OAuthAppConfig != nil {
		clientID = inboundAuthConfig.OAuthAppConfig.ClientID
	}

	sysAttrsJSON, marshalErr := buildSystemAttributes(app, clientID)
	if marshalErr != nil {
		as.logger.Error("Failed to build entity system attributes for update", log.Error(marshalErr))
		return &serviceerror.InternalServerError
	}

	if epErr := as.entityProvider.UpdateSystemAttributes(appID, sysAttrsJSON); epErr != nil {
		if svcErr := mapEntityProviderError(epErr); svcErr != nil {
			return svcErr
		}
		as.logger.Error("Failed to update entity system attributes", log.String("appID", appID), log.Error(epErr))
		return &serviceerror.InternalServerError
	}

	if inboundAuthConfig == nil || inboundAuthConfig.OAuthAppConfig == nil ||
		inboundAuthConfig.OAuthAppConfig.ClientSecret == "" {
		return nil
	}

	sysCredsJSON, marshalErr := buildSystemCredentials(inboundAuthConfig.OAuthAppConfig.ClientSecret)
	if marshalErr != nil {
		as.logger.Error("Failed to build entity system credentials for update", log.Error(marshalErr))
		return &serviceerror.InternalServerError
	}

	if epErr := as.entityProvider.UpdateSystemCredentials(appID, sysCredsJSON); epErr != nil {
		if svcErr := mapEntityProviderError(epErr); svcErr != nil {
			return svcErr
		}
		as.logger.Error("Failed to update entity system credentials", log.String("appID", appID), log.Error(epErr))
		return &serviceerror.InternalServerError
	}

	return nil
}

// DeleteApplication delete the application for given app id.
func (as *applicationService) DeleteApplication(ctx context.Context, appID string) *serviceerror.ServiceError {
	if appID == "" {
		return &ErrorInvalidApplicationID
	}

	// Delete config.
	appErr := as.inboundClientService.DeleteInboundClient(ctx, appID)
	if appErr != nil {
		if errors.Is(appErr, inboundclient.ErrInboundClientNotFound) {
			return nil
		}
		if svcErr := translateInboundClientError(appErr); svcErr != nil {
			return svcErr
		}
		if svcErr := as.translateCertError(appErr); svcErr != nil {
			return svcErr
		}
		as.logger.Error("Failed to delete application", log.Error(appErr), log.String("appID", appID))
		return &serviceerror.InternalServerError
	}

	// Delete entity.
	if epErr := as.entityProvider.DeleteEntity(appID); epErr != nil {
		if svcErr := mapEntityProviderError(epErr); svcErr != nil {
			return svcErr
		}
		as.logger.Error("Failed to delete application entity", log.String("appID", appID), log.Error(epErr))
		return &serviceerror.InternalServerError
	}

	return as.deleteLocalizedVariants(ctx, appID)
}

// isIdentifierTaken checks if an entity with the given identifier already exists.
// If excludeID is non-empty, the entity with that ID is excluded from the check
// (used during declarative loading and updates where the entity already exists).
func (as *applicationService) isIdentifierTaken(key, value, excludeID string) (bool, *serviceerror.ServiceError) {
	entityID, epErr := as.entityProvider.IdentifyEntity(map[string]interface{}{key: value})
	if epErr != nil {
		if epErr.Code == entityprovider.ErrorCodeEntityNotFound {
			return false, nil
		}
		as.logger.Error("Failed to check identifier availability",
			log.String("key", key), log.String("value", value), log.Error(epErr))
		return false, &serviceerror.InternalServerError
	}
	if entityID == nil {
		return false, nil
	}
	if excludeID != "" && *entityID == excludeID {
		return false, nil
	}
	return true, nil
}

// getApplication loads entity + config + OAuth config and merges into ApplicationProcessedDTO.
func (as *applicationService) getApplication(
	ctx context.Context, appID string,
) (*model.ApplicationProcessedDTO, *serviceerror.ServiceError) {
	configDAO, err := as.inboundClientService.GetInboundClientByEntityID(ctx, appID)
	if err != nil {
		return nil, as.mapStoreError(err)
	}
	if configDAO == nil {
		return nil, &ErrorApplicationNotFound
	}

	entity, epErr := as.entityProvider.GetEntity(appID)
	if epErr != nil {
		if epErr.Code == entityprovider.ErrorCodeEntityNotFound {
			entity = nil
		} else {
			as.logger.Error("Failed to get entity for application", log.String("appID", appID), log.Error(epErr))
			return nil, &serviceerror.InternalServerError
		}
	}

	oauthDAO, err := as.inboundClientService.GetOAuthProfileByEntityID(ctx, appID)
	if err != nil && !errors.Is(err, inboundclient.ErrInboundClientNotFound) {
		as.logger.Error("Failed to get OAuth profile for application", log.String("appID", appID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	dto := toProcessedDTO(entity, configDAO, oauthDAO)
	return dto, nil
}

// mapEntityProviderError maps entity provider error codes to application service errors.
func mapEntityProviderError(epErr *entityprovider.EntityProviderError) *serviceerror.ServiceError {
	if epErr == nil {
		return nil
	}
	switch epErr.Code {
	case entityprovider.ErrorCodeEntityNotFound:
		return &ErrorApplicationNotFound
	default:
		return nil
	}
}

// toConfigDAO extracts gateway config fields from a full ApplicationProcessedDTO.
func toConfigDAO(dto *model.ApplicationProcessedDTO) inboundmodel.InboundClient {
	dao := inboundmodel.InboundClient{
		ID:                        dto.ID,
		AuthFlowID:                dto.AuthFlowID,
		RegistrationFlowID:        dto.RegistrationFlowID,
		IsRegistrationFlowEnabled: dto.IsRegistrationFlowEnabled,
		ThemeID:                   dto.ThemeID,
		LayoutID:                  dto.LayoutID,
		Assertion:                 dto.Assertion,
		LoginConsent:              dto.LoginConsent,
		AllowedEntityTypes:        dto.AllowedUserTypes,
	}

	// Pack remaining fields into Properties.
	props := make(map[string]interface{})
	if dto.URL != "" {
		props[propURL] = dto.URL
	}
	if dto.LogoURL != "" {
		props[propLogoURL] = dto.LogoURL
	}
	if dto.TosURI != "" {
		props[propTosURI] = dto.TosURI
	}
	if dto.PolicyURI != "" {
		props[propPolicyURI] = dto.PolicyURI
	}
	if len(dto.Contacts) > 0 {
		props[propContacts] = dto.Contacts
	}
	if dto.Template != "" {
		props[propTemplate] = dto.Template
	}
	if dto.Metadata != nil {
		props[propMetadata] = dto.Metadata
	}
	if len(props) > 0 {
		dao.Properties = props
	}

	return dao
}

// toProcessedDTO merges entity identity data with store config into a full
// ApplicationProcessedDTO.
func toProcessedDTO(
	e *entityprovider.Entity, dao *inboundmodel.InboundClient, oauthDAO *inboundmodel.OAuthProfile,
) *model.ApplicationProcessedDTO {
	dto := &model.ApplicationProcessedDTO{
		ID:                        dao.ID,
		AuthFlowID:                dao.AuthFlowID,
		RegistrationFlowID:        dao.RegistrationFlowID,
		IsRegistrationFlowEnabled: dao.IsRegistrationFlowEnabled,
		ThemeID:                   dao.ThemeID,
		LayoutID:                  dao.LayoutID,
		Assertion:                 dao.Assertion,
		LoginConsent:              dao.LoginConsent,
		AllowedUserTypes:          dao.AllowedEntityTypes,
	}

	// Extract identity fields from entity system attributes.
	if e != nil {
		dto.OUID = e.OUID
		var sysAttrs map[string]interface{}
		if len(e.SystemAttributes) > 0 {
			_ = json.Unmarshal(e.SystemAttributes, &sysAttrs)
		}
		if sysAttrs != nil {
			if name, ok := sysAttrs[fieldName].(string); ok {
				dto.Name = name
			}
			if desc, ok := sysAttrs[fieldDescription].(string); ok {
				dto.Description = desc
			}
		}
	}

	// Extract remaining fields from Properties.
	if dao.Properties != nil {
		if url, ok := dao.Properties[propURL].(string); ok {
			dto.URL = url
		}
		if logoURL, ok := dao.Properties[propLogoURL].(string); ok {
			dto.LogoURL = logoURL
		}
		if tosURI, ok := dao.Properties[propTosURI].(string); ok {
			dto.TosURI = tosURI
		}
		if policyURI, ok := dao.Properties[propPolicyURI].(string); ok {
			dto.PolicyURI = policyURI
		}
		switch contacts := dao.Properties[propContacts].(type) {
		case []string:
			dto.Contacts = append(dto.Contacts, contacts...)
		case []interface{}:
			for _, c := range contacts {
				if s, ok := c.(string); ok {
					dto.Contacts = append(dto.Contacts, s)
				}
			}
		}
		if template, ok := dao.Properties[propTemplate].(string); ok {
			dto.Template = template
		}
		if metadata, ok := dao.Properties[propMetadata].(map[string]interface{}); ok {
			dto.Metadata = metadata
		}
	}

	// Merge OAuth profile if present.
	if oauthDAO != nil && oauthDAO.OAuthProfile != nil {
		var clientID string
		if e != nil {
			var sysAttrs map[string]interface{}
			if len(e.SystemAttributes) > 0 {
				_ = json.Unmarshal(e.SystemAttributes, &sysAttrs)
			}
			if sysAttrs != nil {
				if cid, ok := sysAttrs[fieldClientID].(string); ok {
					clientID = cid
				}
			}
		}

		var ouID string
		if e != nil {
			ouID = e.OUID
		}
		oauthProcessed := inboundclient.BuildOAuthClient(dao.ID, clientID, ouID, oauthDAO)
		dto.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{
			{Type: model.OAuthInboundAuthType, OAuthAppConfig: oauthProcessed},
		}
	}

	return dto
}

// toOAuthConfigData builds the typed OAuth config from a processed DTO for store persistence.
// Returns nil when no OAuth inbound config is present.
func toOAuthConfigData(processedDTO *model.ApplicationProcessedDTO) *inboundmodel.OAuthProfileData {
	oauthProcessed := getOAuthInboundAuthConfigProcessedDTO(processedDTO.InboundAuthConfig)
	if oauthProcessed == nil || oauthProcessed.OAuthAppConfig == nil {
		return nil
	}
	return buildOAuthConfigData(*oauthProcessed)
}

// buildOAuthConfigData builds a typed OAuthConfigData from an InboundAuthConfigProcessedDTO.
// Returns nil if the inbound auth config has no OAuth application config.
func buildOAuthConfigData(inboundAuth model.InboundAuthConfigProcessedDTO) *inboundmodel.OAuthProfileData {
	if inboundAuth.OAuthAppConfig == nil {
		return nil
	}
	oa := inboundAuth.OAuthAppConfig
	return &inboundmodel.OAuthProfileData{
		RedirectURIs:                       oa.RedirectURIs,
		GrantTypes:                         sysutils.ConvertToStringSlice(oa.GrantTypes),
		ResponseTypes:                      sysutils.ConvertToStringSlice(oa.ResponseTypes),
		TokenEndpointAuthMethod:            string(oa.TokenEndpointAuthMethod),
		PKCERequired:                       oa.PKCERequired,
		PublicClient:                       oa.PublicClient,
		RequirePushedAuthorizationRequests: oa.RequirePushedAuthorizationRequests,
		Scopes:                             oa.Scopes,
		ScopeClaims:                        oa.ScopeClaims,
		Token:                              oa.Token,
		UserInfo:                           oa.UserInfo,
		Certificate:                        oa.Certificate,
	}
}

// buildSystemAttributes builds the system attributes JSON for the entity.
func buildSystemAttributes(app *model.ApplicationDTO, clientID string) (json.RawMessage, error) {
	sysAttrs := map[string]interface{}{
		fieldName: app.Name,
	}
	if app.Description != "" {
		sysAttrs[fieldDescription] = app.Description
	}
	if clientID != "" {
		sysAttrs[fieldClientID] = clientID
	}
	return json.Marshal(sysAttrs)
}

// buildAppEntity constructs an entity and system credentials for entity creation.
func buildAppEntity(appID string, app *model.ApplicationDTO, clientID string, plaintextSecret string) (
	*entityprovider.Entity, json.RawMessage, error) {
	sysAttrsJSON, err := buildSystemAttributes(app, clientID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build entity system attributes: %w", err)
	}

	sysCredsJSON, err := buildSystemCredentials(plaintextSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build entity system credentials: %w", err)
	}

	e := &entityprovider.Entity{
		ID:               appID,
		Category:         entityprovider.EntityCategoryApp,
		Type:             "application",
		State:            entityprovider.EntityStateActive,
		OUID:             app.OUID,
		SystemAttributes: sysAttrsJSON,
	}
	return e, sysCredsJSON, nil
}

// buildSystemCredentials builds the system credentials JSON for the entity.
func buildSystemCredentials(clientSecret string) (json.RawMessage, error) {
	if clientSecret == "" {
		return nil, nil
	}

	return json.Marshal(map[string]interface{}{
		fieldClientSecret: clientSecret,
	})
}

// getOAuthInboundAuthConfigDTO returns the single OAuth InboundAuthConfigDTO.
// It returns an error if multiple OAuth configs are found, nil if none exist.
func getOAuthInboundAuthConfigDTO(
	configs []model.InboundAuthConfigDTO,
) (*model.InboundAuthConfigDTO, *serviceerror.ServiceError) {
	var cfg *model.InboundAuthConfigDTO
	for i := range configs {
		if configs[i].Type == model.OAuthInboundAuthType {
			if cfg != nil {
				return nil, &ErrorInvalidInboundAuthConfig
			}
			cfg = &configs[i]
		}
	}
	return cfg, nil
}

// getOAuthInboundAuthConfigProcessedDTO returns the first OAuth InboundAuthConfigProcessedDTO, or nil.
func getOAuthInboundAuthConfigProcessedDTO(
	configs []model.InboundAuthConfigProcessedDTO,
) *model.InboundAuthConfigProcessedDTO {
	for i := range configs {
		if configs[i].Type == model.OAuthInboundAuthType {
			return &configs[i]
		}
	}
	return nil
}

func (as *applicationService) validateApplicationForUpdate(
	ctx context.Context, appID string, app *model.ApplicationDTO) (
	*model.ApplicationProcessedDTO, *model.InboundAuthConfigDTO, *serviceerror.ServiceError) {
	if appID == "" {
		return nil, nil, &ErrorInvalidApplicationID
	}
	if app == nil {
		return nil, nil, &ErrorApplicationNil
	}
	if app.Name == "" {
		return nil, nil, &ErrorInvalidApplicationName
	}

	existingApp, existingAppErr := as.getApplication(ctx, appID)
	if existingAppErr != nil {
		return nil, nil, existingAppErr
	}

	// If the application name is changed, check if an application with the new name already exists.
	if existingApp.Name != app.Name {
		nameExists, nameCheckErr := as.isIdentifierTaken(fieldName, app.Name, appID)
		if nameCheckErr != nil {
			return nil, nil, nameCheckErr
		}
		if nameExists {
			return nil, nil, &ErrorApplicationAlreadyExistsWithName
		}
	}

	if svcErr := as.validateApplicationFields(ctx, app); svcErr != nil {
		return nil, nil, svcErr
	}

	inboundAuthConfig, svcErr := as.processInboundAuthConfig(app, existingApp)
	if svcErr != nil {
		return nil, nil, svcErr
	}

	return existingApp, inboundAuthConfig, nil
}

// validateApplicationFields validates application fields that are common to both create and update operations.
func (as *applicationService) validateApplicationFields(
	ctx context.Context, app *model.ApplicationDTO) *serviceerror.ServiceError {
	// Validate organization unit ID.
	if app.OUID == "" {
		return &ErrorInvalidRequestFormat
	}
	if exists, err := as.ouService.IsOrganizationUnitExists(ctx, app.OUID); err != nil || !exists {
		return &ErrorInvalidRequestFormat
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return &ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidLogoURI(app.LogoURL) {
		return &ErrorInvalidLogoURL
	}
	as.validateConsentConfig(app)
	return nil
}

// validateConsentConfig validates the consent configuration for the application.
func (as *applicationService) validateConsentConfig(appDTO *model.ApplicationDTO) {
	if appDTO.LoginConsent == nil {
		appDTO.LoginConsent = &inboundmodel.LoginConsentConfig{
			ValidityPeriod: 0,
		}

		return
	}

	if appDTO.LoginConsent.ValidityPeriod < 0 {
		appDTO.LoginConsent.ValidityPeriod = 0
	}
}

// validateOAuthParamsForCreateAndUpdate validates the OAuth parameters for creating or updating an application.
func validateOAuthParamsForCreateAndUpdate(app *model.ApplicationDTO) (*model.InboundAuthConfigDTO,
	*serviceerror.ServiceError) {
	if len(app.InboundAuthConfig) == 0 {
		return nil, nil
	}

	inboundAuthConfig, svcErr := getOAuthInboundAuthConfigDTO(app.InboundAuthConfig)
	if svcErr != nil {
		return nil, svcErr
	}
	if inboundAuthConfig == nil {
		return nil, &ErrorInvalidInboundAuthConfig
	}
	if inboundAuthConfig.OAuthAppConfig == nil {
		return nil, &ErrorInvalidInboundAuthConfig
	}

	oauthAppConfig := inboundAuthConfig.OAuthAppConfig

	if len(oauthAppConfig.GrantTypes) == 0 {
		oauthAppConfig.GrantTypes = []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode}
	}
	if len(oauthAppConfig.ResponseTypes) == 0 {
		if slices.Contains(oauthAppConfig.GrantTypes, oauth2const.GrantTypeAuthorizationCode) {
			oauthAppConfig.ResponseTypes = []oauth2const.ResponseType{oauth2const.ResponseTypeCode}
		}
	}
	if oauthAppConfig.TokenEndpointAuthMethod == "" {
		oauthAppConfig.TokenEndpointAuthMethod = oauth2const.TokenEndpointAuthMethodClientSecretBasic
	}

	return inboundAuthConfig, nil
}

func translateOAuthValidationError(err error) *serviceerror.ServiceError {
	switch {
	case errors.Is(err, inboundclient.ErrOAuthInvalidRedirectURI):
		return &ErrorInvalidRedirectURI
	case errors.Is(err, inboundclient.ErrOAuthRedirectURIFragmentNotAllowed):
		return serviceerror.CustomServiceError(ErrorInvalidRedirectURI, core.I18nMessage{
			Key:          "error.applicationservice.redirect_uri_fragment_not_allowed_description",
			DefaultValue: "Redirect URIs must not contain a fragment component",
		})
	case errors.Is(err, inboundclient.ErrOAuthAuthCodeRequiresRedirectURIs):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.auth_code_requires_redirect_uris_description",
			DefaultValue: "authorization_code grant type requires redirect URIs",
		})
	case errors.Is(err, inboundclient.ErrOAuthInvalidGrantType):
		return &ErrorInvalidGrantType
	case errors.Is(err, inboundclient.ErrOAuthInvalidResponseType):
		return &ErrorInvalidResponseType
	case errors.Is(err, inboundclient.ErrOAuthClientCredentialsCannotUseResponseTypes):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.client_credentials_cannot_use_response_types_description",
			DefaultValue: "client_credentials grant type cannot be used with response types",
		})
	case errors.Is(err, inboundclient.ErrOAuthAuthCodeRequiresCodeResponseType):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.auth_code_requires_code_response_type_description",
			DefaultValue: "authorization_code grant type requires 'code' response type",
		})
	case errors.Is(err, inboundclient.ErrOAuthRefreshTokenCannotBeSoleGrant):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.refresh_token_cannot_be_sole_grant_description",
			DefaultValue: "refresh_token grant type cannot be used without another grant type",
		})
	case errors.Is(err, inboundclient.ErrOAuthPKCERequiresAuthCode):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.pkce_requires_authorization_code_description",
			DefaultValue: "PKCE can only be enabled when the authorization_code grant type is selected",
		})
	case errors.Is(err, inboundclient.ErrOAuthResponseTypesRequireAuthCode):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.response_types_require_authorization_code_description",
			DefaultValue: "Response types can only be configured with the authorization_code grant type",
		})
	case errors.Is(err, inboundclient.ErrOAuthInvalidTokenEndpointAuthMethod):
		return &ErrorInvalidTokenEndpointAuthMethod
	case errors.Is(err, inboundclient.ErrOAuthPrivateKeyJWTRequiresCertificate):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.private_key_jwt_requires_certificate_description",
			DefaultValue: "private_key_jwt authentication method requires a certificate",
		})
	case errors.Is(err, inboundclient.ErrOAuthPrivateKeyJWTCannotHaveClientSecret):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.private_key_jwt_cannot_have_client_secret_description",
			DefaultValue: "private_key_jwt authentication method cannot have a client secret",
		})
	case errors.Is(err, inboundclient.ErrOAuthClientSecretCannotHaveCertificate):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.client_secret_cannot_have_certificate_description",
			DefaultValue: "client_secret authentication methods cannot have a certificate",
		})
	case errors.Is(err, inboundclient.ErrOAuthNoneAuthRequiresPublicClient):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.none_auth_method_requires_public_client_description",
			DefaultValue: "'none' authentication method requires the client to be a public client",
		})
	case errors.Is(err, inboundclient.ErrOAuthNoneAuthCannotHaveCertOrSecret):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.none_auth_method_cannot_have_cert_or_secret_description",
			DefaultValue: "'none' authentication method cannot have a certificate or client secret",
		})
	case errors.Is(err, inboundclient.ErrOAuthClientCredentialsCannotUseNoneAuth):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.client_credentials_cannot_use_none_auth_description",
			DefaultValue: "client_credentials grant type cannot use 'none' authentication method",
		})
	case errors.Is(err, inboundclient.ErrOAuthPublicClientMustUseNoneAuth):
		return serviceerror.CustomServiceError(ErrorInvalidPublicClientConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.public_client_must_use_none_auth_description",
			DefaultValue: "Public clients must use 'none' as token endpoint authentication method",
		})
	case errors.Is(err, inboundclient.ErrOAuthPublicClientMustHavePKCE):
		return serviceerror.CustomServiceError(ErrorInvalidPublicClientConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.public_client_must_have_pkce_description",
			DefaultValue: "Public clients must have PKCE required set to true",
		})
	default:
		return translateUserInfoValidationError(err)
	}
}

func translateUserInfoValidationError(err error) *serviceerror.ServiceError {
	switch {
	case errors.Is(err, inboundclient.ErrOAuthUserInfoUnsupportedSigningAlg):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_unsupported_signing_alg_description",
			DefaultValue: "userinfo signing algorithm is not supported",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoUnsupportedEncryptionAlg):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_unsupported_encryption_alg_description",
			DefaultValue: "userinfo encryption algorithm is not supported",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoUnsupportedEncryptionEnc):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_unsupported_encryption_enc_description",
			DefaultValue: "userinfo content-encryption algorithm is not supported",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoEncryptionAlgRequiresEnc):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_encryption_alg_requires_enc_description",
			DefaultValue: "encryptionEnc is required when encryptionAlg is set",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoEncryptionEncRequiresAlg):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_encryption_enc_requires_alg_description",
			DefaultValue: "encryptionAlg is required when encryptionEnc is set",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoEncryptionRequiresCertificate):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_encryption_requires_certificate_description",
			DefaultValue: "a certificate (JWKS or JWKS_URI) is required when userinfo encryption is configured",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoJWKSURINotSSRFSafe):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_jwks_uri_not_ssrf_safe_description",
			DefaultValue: "JWKS URI must be a publicly reachable HTTPS URL",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoUnsupportedResponseType):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_unsupported_response_type_description",
			DefaultValue: "userinfo responseType is not supported",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoJWSRequiresSigningAlg):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_jws_requires_signing_alg_description",
			DefaultValue: "signingAlg is required when responseType is JWS",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoJWERequiresEncryption):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_jwe_requires_encryption_description",
			DefaultValue: "encryptionAlg and encryptionEnc are required when responseType is JWE",
		})
	case errors.Is(err, inboundclient.ErrOAuthUserInfoNestedJWTRequiresAll):
		return serviceerror.CustomServiceError(ErrorInvalidOAuthConfiguration, core.I18nMessage{
			Key:          "error.applicationservice.userinfo_nested_jwt_requires_all_description",
			DefaultValue: "signingAlg, encryptionAlg, and encryptionEnc are required when responseType is NESTED_JWT",
		})
	default:
		return nil
	}
}

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

func translateInboundClientError(err error) *serviceerror.ServiceError {
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
	var consentErr *inboundclient.ConsentSyncError
	if errors.As(err, &consentErr) {
		return translateConsentSyncError(consentErr)
	}
	return nil
}

func translateConsentSyncError(err *inboundclient.ConsentSyncError) *serviceerror.ServiceError {
	if err.IsClientError() {
		return serviceerror.CustomServiceError(ErrorConsentSyncFailed, core.I18nMessage{
			Key: "error.applicationservice.consent_sync_failed_description",
			DefaultValue: fmt.Sprintf(
				ErrorConsentSyncFailed.ErrorDescription.DefaultValue+" : code - %s",
				err.Underlying.Code,
			),
		})
	}
	return &serviceerror.InternalServerError
}

func (as *applicationService) processInboundAuthConfig(app *model.ApplicationDTO,
	existingApp *model.ApplicationProcessedDTO) (
	*model.InboundAuthConfigDTO, *serviceerror.ServiceError) {
	inboundAuthConfig, err := validateOAuthParamsForCreateAndUpdate(app)
	if err != nil {
		return nil, err
	}

	if inboundAuthConfig == nil {
		return nil, nil
	}

	clientID := inboundAuthConfig.OAuthAppConfig.ClientID

	// For update operation
	if existingApp != nil {
		var existingClientID string
		if existingOAuthConfig := getOAuthInboundAuthConfigProcessedDTO(
			existingApp.InboundAuthConfig); existingOAuthConfig != nil &&
			existingOAuthConfig.OAuthAppConfig != nil {
			existingClientID = existingOAuthConfig.OAuthAppConfig.ClientID
		}

		if clientID == "" {
			if svcErr := generateAndAssignClientID(inboundAuthConfig); svcErr != nil {
				return nil, svcErr
			}
		} else if clientID != existingClientID {
			if taken, svcErr := as.isIdentifierTaken(fieldClientID, clientID, existingApp.ID); svcErr != nil {
				return nil, svcErr
			} else if taken {
				return nil, &ErrorApplicationAlreadyExistsWithClientID
			}
		}
	} else { // For create operation
		if clientID == "" {
			if svcErr := generateAndAssignClientID(inboundAuthConfig); svcErr != nil {
				return nil, svcErr
			}
		} else {
			if taken, svcErr := as.isIdentifierTaken(fieldClientID, clientID, app.ID); svcErr != nil {
				return nil, svcErr
			} else if taken {
				return nil, &ErrorApplicationAlreadyExistsWithClientID
			}
		}
	}

	if svcErr := resolveClientSecret(inboundAuthConfig, existingApp); svcErr != nil {
		return nil, svcErr
	}

	return inboundAuthConfig, nil
}

// generateAndAssignClientID generates an OAuth 2.0 compliant client ID and assigns it to the inbound auth config.
func generateAndAssignClientID(inboundAuthConfig *model.InboundAuthConfigDTO) *serviceerror.ServiceError {
	generatedClientID, err := oauthutils.GenerateOAuth2ClientID()
	if err != nil {
		log.GetLogger().Error("Failed to generate OAuth client ID", log.Error(err))
		return &serviceerror.InternalServerError
	}
	inboundAuthConfig.OAuthAppConfig.ClientID = generatedClientID
	return nil
}

func resolveClientSecret(
	inboundAuthConfig *model.InboundAuthConfigDTO,
	existingApp *model.ApplicationProcessedDTO,
) *serviceerror.ServiceError {
	if (inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod !=
		oauth2const.TokenEndpointAuthMethodClientSecretBasic &&
		inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod !=
			oauth2const.TokenEndpointAuthMethodClientSecretPost) ||
		inboundAuthConfig.OAuthAppConfig.ClientSecret != "" {
		return nil
	}

	if existingApp != nil {
		if existingInboundAuth := getOAuthInboundAuthConfigProcessedDTO(
			existingApp.InboundAuthConfig); existingInboundAuth != nil {
			existingOAuth := existingInboundAuth.OAuthAppConfig
			if existingOAuth != nil && !existingOAuth.PublicClient &&
				(existingOAuth.TokenEndpointAuthMethod == oauth2const.TokenEndpointAuthMethodClientSecretBasic ||
					existingOAuth.TokenEndpointAuthMethod == oauth2const.TokenEndpointAuthMethodClientSecretPost) {
				return nil
			}
		}
	}

	generatedClientSecret, err := oauthutils.GenerateOAuth2ClientSecret()
	if err != nil {
		log.GetLogger().Error("Failed to generate OAuth client secret", log.Error(err))
		return &serviceerror.InternalServerError
	}

	inboundAuthConfig.OAuthAppConfig.ClientSecret = generatedClientSecret
	return nil
}

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

func (as *applicationService) translateCertOperationError(
	err *inboundclient.CertOperationError) *serviceerror.ServiceError {
	if !err.IsClientError() {
		as.logger.Error("Certificate operation failed",
			log.Any("operation", err.Operation),
			log.Any("refType", err.RefType),
			log.Any("serviceError", err.Underlying))
		return &serviceerror.InternalServerError
	}

	var key, prefix string
	switch err.Operation {
	case inboundclient.CertOpCreate:
		key, prefix = "error.applicationservice.create_certificate_failed_description",
			"Failed to create application certificate: "
	case inboundclient.CertOpUpdate:
		key, prefix = "error.applicationservice.update_certificate_failed_description",
			"Failed to update application certificate: "
	case inboundclient.CertOpRetrieve:
		key, prefix = "error.applicationservice.retrieve_certificate_failed_description",
			"Failed to retrieve application certificate: "
	case inboundclient.CertOpDelete:
		if err.RefType == cert.CertificateReferenceTypeOAuthApp {
			key, prefix = "error.applicationservice.delete_oauth_certificate_failed_description",
				"Failed to delete OAuth app certificate: "
		} else {
			key, prefix = "error.applicationservice.delete_certificate_failed_description",
				"Failed to delete application certificate: "
		}
	default:
		return &serviceerror.InternalServerError
	}
	return serviceerror.CustomServiceError(ErrorCertificateClientError, core.I18nMessage{
		Key:          key,
		DefaultValue: prefix + err.Underlying.ErrorDescription.DefaultValue,
	})
}

// translateCertError converts a typed cert error returned from inboundclient into a ServiceError.
// Returns nil if the error is not a cert error (caller should handle it separately).
func (as *applicationService) translateCertError(err error) *serviceerror.ServiceError {
	if svcErr := translateCertValidationError(err); svcErr != nil {
		return svcErr
	}
	var opErr *inboundclient.CertOperationError
	if errors.As(err, &opErr) {
		return as.translateCertOperationError(opErr)
	}
	return nil
}

// enrichApplicationWithCertificate retrieves and adds the certificate to the application.
func (as *applicationService) enrichApplicationWithCertificate(ctx context.Context, application *model.Application) (
	*model.Application, *serviceerror.ServiceError) {
	appCert, opErr := as.inboundClientService.GetCertificate(
		ctx, cert.CertificateReferenceTypeApplication, application.ID)
	if opErr != nil {
		return nil, as.translateCertOperationError(opErr)
	}
	application.Certificate = appCert

	// Enrich OAuth config certificate for each inbound auth config.
	for i, inboundAuthConfig := range application.InboundAuthConfig {
		if inboundAuthConfig.Type == model.OAuthInboundAuthType && inboundAuthConfig.OAuthAppConfig != nil {
			oauthCert, oauthCertOpErr := as.inboundClientService.GetCertificate(ctx,
				cert.CertificateReferenceTypeOAuthApp, inboundAuthConfig.OAuthAppConfig.ClientID)
			if oauthCertOpErr != nil {
				return nil, as.translateCertOperationError(oauthCertOpErr)
			}
			application.InboundAuthConfig[i].OAuthAppConfig.Certificate = oauthCert
		}
	}

	return application, nil
}

// buildApplicationResponse maps an ApplicationProcessedDTO to an Application response.
// The returned application's Certificate field is populated separately by enrichApplicationWithCertificate.
func buildApplicationResponse(dto *model.ApplicationProcessedDTO) *model.Application {
	application := &model.Application{
		ID:                        dto.ID,
		OUID:                      dto.OUID,
		Name:                      dto.Name,
		Description:               dto.Description,
		AuthFlowID:                dto.AuthFlowID,
		RegistrationFlowID:        dto.RegistrationFlowID,
		IsRegistrationFlowEnabled: dto.IsRegistrationFlowEnabled,
		ThemeID:                   dto.ThemeID,
		LayoutID:                  dto.LayoutID,
		Template:                  dto.Template,
		URL:                       dto.URL,
		LogoURL:                   dto.LogoURL,
		TosURI:                    dto.TosURI,
		PolicyURI:                 dto.PolicyURI,
		Assertion:                 dto.Assertion,
		Contacts:                  dto.Contacts,
		AllowedUserTypes:          dto.AllowedUserTypes,
		LoginConsent:              dto.LoginConsent,
		Metadata:                  dto.Metadata,
	}
	inboundAuthConfigs := make([]model.InboundAuthConfigComplete, 0, len(dto.InboundAuthConfig))
	for _, config := range dto.InboundAuthConfig {
		if config.Type == model.OAuthInboundAuthType && config.OAuthAppConfig != nil {
			oauthAppConfig := config.OAuthAppConfig
			inboundAuthConfigs = append(inboundAuthConfigs, model.InboundAuthConfigComplete{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID:                           oauthAppConfig.ClientID,
					RedirectURIs:                       oauthAppConfig.RedirectURIs,
					GrantTypes:                         oauthAppConfig.GrantTypes,
					ResponseTypes:                      oauthAppConfig.ResponseTypes,
					TokenEndpointAuthMethod:            oauthAppConfig.TokenEndpointAuthMethod,
					PKCERequired:                       oauthAppConfig.PKCERequired,
					PublicClient:                       oauthAppConfig.PublicClient,
					RequirePushedAuthorizationRequests: oauthAppConfig.RequirePushedAuthorizationRequests,
					Token:                              oauthAppConfig.Token,
					Scopes:                             oauthAppConfig.Scopes,
					UserInfo:                           oauthAppConfig.UserInfo,
					ScopeClaims:                        oauthAppConfig.ScopeClaims,
				},
			})
		}
	}
	application.InboundAuthConfig = inboundAuthConfigs
	return application
}

// buildBasicApplicationResponse builds a BasicApplicationResponse by merging config + entity data.
func buildBasicApplicationResponse(
	cfg inboundmodel.InboundClient, e *entityprovider.Entity,
) model.BasicApplicationResponse {
	resp := model.BasicApplicationResponse{
		ID:                        cfg.ID,
		AuthFlowID:                cfg.AuthFlowID,
		RegistrationFlowID:        cfg.RegistrationFlowID,
		IsRegistrationFlowEnabled: cfg.IsRegistrationFlowEnabled,
		ThemeID:                   cfg.ThemeID,
		LayoutID:                  cfg.LayoutID,
		IsReadOnly:                cfg.IsReadOnly,
	}
	if cfg.Properties != nil {
		if t, ok := cfg.Properties[propTemplate].(string); ok {
			resp.Template = t
		}
		if logoURL, ok := cfg.Properties[propLogoURL].(string); ok {
			resp.LogoURL = logoURL
		}
	}
	// Enrich from entity system attributes.
	if e != nil {
		var sysAttrs map[string]interface{}
		if len(e.SystemAttributes) > 0 {
			_ = json.Unmarshal(e.SystemAttributes, &sysAttrs)
		}
		if sysAttrs != nil {
			if name, ok := sysAttrs[fieldName].(string); ok {
				resp.Name = name
			}
			if desc, ok := sysAttrs[fieldDescription].(string); ok {
				resp.Description = desc
			}
			if clientID, ok := sysAttrs[fieldClientID].(string); ok {
				resp.ClientID = clientID
			}
		}
	}
	return resp
}

// buildBaseApplicationProcessedDTO constructs an ApplicationProcessedDTO with the common base fields.
// Callers are responsible for setting InboundAuthConfig on the returned DTO.
func buildBaseApplicationProcessedDTO(appID string, app *model.ApplicationDTO,
	assertion *inboundmodel.AssertionConfig) *model.ApplicationProcessedDTO {
	return &model.ApplicationProcessedDTO{
		ID:                        appID,
		OUID:                      app.OUID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowID:                app.AuthFlowID,
		RegistrationFlowID:        app.RegistrationFlowID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		ThemeID:                   app.ThemeID,
		LayoutID:                  app.LayoutID,
		Template:                  app.Template,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Assertion:                 assertion,
		TosURI:                    app.TosURI,
		PolicyURI:                 app.PolicyURI,
		Contacts:                  app.Contacts,
		AllowedUserTypes:          app.AllowedUserTypes,
		LoginConsent:              app.LoginConsent,
		Metadata:                  app.Metadata,
	}
}

// buildProcessedDTOForUpdate constructs the ApplicationProcessedDTO for an application
// update operation.
func (as *applicationService) buildProcessedDTOForUpdate(appID string, app *model.ApplicationDTO,
	inboundAuthConfig *model.InboundAuthConfigDTO) *model.ApplicationProcessedDTO {
	processedDTO := buildBaseApplicationProcessedDTO(appID, app, app.Assertion)

	if inboundAuthConfig != nil {
		oa := inboundAuthConfig.OAuthAppConfig
		processedInboundAuthConfig := buildOAuthInboundAuthConfigProcessedDTO(
			appID, inboundAuthConfig, oa.Token, oa.UserInfo, oa.ScopeClaims, oa.Certificate,
		)
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
	}

	return processedDTO
}

// buildOAuthInboundAuthConfigProcessedDTO constructs the InboundAuthConfigProcessedDTO for an OAuth application.
func buildOAuthInboundAuthConfigProcessedDTO(
	appID string, inboundAuthConfig *model.InboundAuthConfigDTO,
	oauthToken *inboundmodel.OAuthTokenConfig, userInfo *inboundmodel.UserInfoConfig,
	scopeClaims map[string][]string, certificate *inboundmodel.Certificate,
) model.InboundAuthConfigProcessedDTO {
	return model.InboundAuthConfigProcessedDTO{
		Type: model.OAuthInboundAuthType,
		OAuthAppConfig: &inboundmodel.OAuthClient{
			AppID:                              appID,
			ClientID:                           inboundAuthConfig.OAuthAppConfig.ClientID,
			RedirectURIs:                       inboundAuthConfig.OAuthAppConfig.RedirectURIs,
			GrantTypes:                         inboundAuthConfig.OAuthAppConfig.GrantTypes,
			ResponseTypes:                      inboundAuthConfig.OAuthAppConfig.ResponseTypes,
			TokenEndpointAuthMethod:            inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
			PKCERequired:                       inboundAuthConfig.OAuthAppConfig.PKCERequired,
			PublicClient:                       inboundAuthConfig.OAuthAppConfig.PublicClient,
			RequirePushedAuthorizationRequests: inboundAuthConfig.OAuthAppConfig.RequirePushedAuthorizationRequests,
			Token:                              oauthToken,
			Scopes:                             inboundAuthConfig.OAuthAppConfig.Scopes,
			UserInfo:                           userInfo,
			ScopeClaims:                        scopeClaims,
			Certificate:                        certificate,
		},
	}
}

// buildReturnApplicationDTO constructs the ApplicationDTO returned from create and update operations.
func buildReturnApplicationDTO(
	appID string, app *model.ApplicationDTO, assertion *inboundmodel.AssertionConfig,
	metadata map[string]any, inboundAuthConfig *model.InboundAuthConfigDTO,
	oauthToken *inboundmodel.OAuthTokenConfig, userInfo *inboundmodel.UserInfoConfig,
	scopeClaims map[string][]string) *model.ApplicationDTO {
	returnApp := &model.ApplicationDTO{
		ID:                        appID,
		OUID:                      app.OUID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowID:                app.AuthFlowID,
		RegistrationFlowID:        app.RegistrationFlowID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		ThemeID:                   app.ThemeID,
		LayoutID:                  app.LayoutID,
		Template:                  app.Template,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Assertion:                 assertion,
		Certificate:               app.Certificate,
		TosURI:                    app.TosURI,
		PolicyURI:                 app.PolicyURI,
		Contacts:                  app.Contacts,
		AllowedUserTypes:          app.AllowedUserTypes,
		LoginConsent:              app.LoginConsent,
		Metadata:                  metadata,
	}
	if inboundAuthConfig != nil {
		var oauthCert *inboundmodel.Certificate
		if inboundAuthConfig.OAuthAppConfig != nil {
			oauthCert = inboundAuthConfig.OAuthAppConfig.Certificate
		}
		returnInboundAuthConfig := model.InboundAuthConfigDTO{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigDTO{
				AppID:                              appID,
				ClientID:                           inboundAuthConfig.OAuthAppConfig.ClientID,
				ClientSecret:                       inboundAuthConfig.OAuthAppConfig.ClientSecret,
				RedirectURIs:                       inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:                         inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:                      inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod:            inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				PKCERequired:                       inboundAuthConfig.OAuthAppConfig.PKCERequired,
				PublicClient:                       inboundAuthConfig.OAuthAppConfig.PublicClient,
				RequirePushedAuthorizationRequests: inboundAuthConfig.OAuthAppConfig.RequirePushedAuthorizationRequests,
				Token:                              oauthToken,
				Scopes:                             inboundAuthConfig.OAuthAppConfig.Scopes,
				UserInfo:                           userInfo,
				ScopeClaims:                        scopeClaims,
				Certificate:                        oauthCert,
			},
		}
		returnApp.InboundAuthConfig = []model.InboundAuthConfigDTO{returnInboundAuthConfig}
	}
	return returnApp
}

// mapStoreError maps inbound client store errors to application service errors.
func (as *applicationService) mapStoreError(err error) *serviceerror.ServiceError {
	if errors.Is(err, inboundclient.ErrInboundClientNotFound) {
		return &ErrorApplicationNotFound
	}
	as.logger.Error("Failed to retrieve application", log.Error(err))
	return &serviceerror.InternalServerError
}

// deleteLocalizedVariants removes all i18n translations for an application's fields.
// All fields are attempted; returns an internal server error if any deletion fails.
func (as *applicationService) deleteLocalizedVariants(ctx context.Context, appID string) *serviceerror.ServiceError {
	if as.i18nService == nil {
		return nil
	}
	var hasErr bool
	for _, field := range []string{"name", "logo_uri", "tos_uri", "policy_uri"} {
		if svcErr := as.i18nService.DeleteTranslationsByKey(
			ctx, AppI18nNamespace(), AppI18nKey(appID, field)); svcErr != nil {
			as.logger.Error("Failed to delete localized variant on app deletion",
				log.String("appID", appID),
				log.String("field", field),
				log.String("namespace", AppI18nNamespace()))
			hasErr = true
		}
	}
	if hasErr {
		return &serviceerror.InternalServerError
	}
	return nil
}

// cleanupStaleI18nKeys removes i18n keys for fields that changed from an i18n ref back to plain text.
// Returns an internal server error if any deletion fails.
func (as *applicationService) cleanupStaleI18nKeys(
	ctx context.Context, appID string,
	existing *model.ApplicationProcessedDTO, updated *model.ApplicationDTO,
) *serviceerror.ServiceError {
	if as.i18nService == nil {
		return nil
	}
	type pair struct{ old, updated, field string }
	fields := []pair{
		{existing.Name, updated.Name, "name"},
		{existing.LogoURL, updated.LogoURL, "logo_uri"},
		{existing.TosURI, updated.TosURI, "tos_uri"},
		{existing.PolicyURI, updated.PolicyURI, "policy_uri"},
	}
	var hasErr bool
	for _, f := range fields {
		if isI18nRef(f.old) && !isI18nRef(f.updated) {
			if svcErr := as.i18nService.DeleteTranslationsByKey(
				ctx, AppI18nNamespace(), AppI18nKey(appID, f.field)); svcErr != nil {
				as.logger.Error("Failed to delete stale i18n key",
					log.String("appID", appID),
					log.String("field", f.field),
					log.String("namespace", AppI18nNamespace()))
				hasErr = true
			}
		}
	}
	if hasErr {
		return &serviceerror.InternalServerError
	}
	return nil
}
