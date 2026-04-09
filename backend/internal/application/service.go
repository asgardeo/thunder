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
	"strings"

	"encoding/json"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/consent"
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	"github.com/asgardeo/thunder/internal/entityprovider"
	flowcommon "github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauthutils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/internal/system/transaction"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema"
)

// ApplicationServiceInterface defines the interface for the application service.
type ApplicationServiceInterface interface {
	CreateApplication(
		ctx context.Context, app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	ValidateApplication(ctx context.Context, app *model.ApplicationDTO) (
		*model.ApplicationProcessedDTO, *model.InboundAuthConfigDTO, *serviceerror.ServiceError)
	GetApplicationList(ctx context.Context) (*model.ApplicationListResponse, *serviceerror.ServiceError)
	GetOAuthApplication(
		ctx context.Context, clientID string) (*model.OAuthAppConfigProcessedDTO, *serviceerror.ServiceError)
	GetApplication(ctx context.Context, appID string) (*model.Application, *serviceerror.ServiceError)
	UpdateApplication(
		ctx context.Context, appID string, app *model.ApplicationDTO) (
		*model.ApplicationDTO, *serviceerror.ServiceError)
	DeleteApplication(ctx context.Context, appID string) *serviceerror.ServiceError
}

// ApplicationService is the default implementation of the ApplicationServiceInterface.
type applicationService struct {
	logger            *log.Logger
	appStore          applicationStoreInterface
	entityProvider    entityprovider.EntityProviderInterface
	ouService         oupkg.OrganizationUnitServiceInterface
	certService       cert.CertificateServiceInterface
	flowMgtService    flowmgt.FlowMgtServiceInterface
	themeMgtService   thememgt.ThemeMgtServiceInterface
	layoutMgtService  layoutmgt.LayoutMgtServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
	consentService    consent.ConsentServiceInterface
	i18nService       i18nmgt.I18nServiceInterface
	transactioner     transaction.Transactioner
}

// newApplicationService creates a new instance of ApplicationService.
func newApplicationService(
	appStore applicationStoreInterface,
	entityProvider entityprovider.EntityProviderInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	certService cert.CertificateServiceInterface,
	flowMgtService flowmgt.FlowMgtServiceInterface,
	themeMgtService thememgt.ThemeMgtServiceInterface,
	layoutMgtService layoutmgt.LayoutMgtServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	consentService consent.ConsentServiceInterface,
	i18nService i18nmgt.I18nServiceInterface,
	transactioner transaction.Transactioner,
) ApplicationServiceInterface {
	return &applicationService{
		logger:            log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService")),
		appStore:          appStore,
		entityProvider:    entityProvider,
		ouService:         ouService,
		certService:       certService,
		flowMgtService:    flowMgtService,
		themeMgtService:   themeMgtService,
		layoutMgtService:  layoutMgtService,
		userSchemaService: userSchemaService,
		consentService:    consentService,
		i18nService:       i18nService,
		transactioner:     transactioner,
	}
}

// appI18nNamespace returns the i18n namespace for an application.
func appI18nNamespace(appID string) string {
	return "app." + appID
}

// deleteLocalizedVariants removes all i18n translations for the application namespace.
func (as *applicationService) deleteLocalizedVariants(appID string) {
	if as.i18nService == nil {
		return
	}
	if svcErr := as.i18nService.DeleteTranslationsByNamespace(appI18nNamespace(appID)); svcErr != nil {
		as.logger.Error("Failed to delete localized variants on app deletion",
			log.String("appID", appID), log.String("error", svcErr.Error.String()))
	}
}

func (as *applicationService) deleteEntityCompensation(appID string) {
	if delErr := as.entityProvider.DeleteEntity(appID); delErr != nil {
		as.logger.Error("Failed to delete entity during compensation", log.Error(delErr),
			log.String("appID", appID))
	}
}

// CreateApplication creates the application.
//
//nolint:gocyclo // orchestration function — branching is inherent to multi-step create flow
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
	if app.ID != "" {
		if as.appStore.IsApplicationDeclarative(ctx, app.ID) {
			return nil, &ErrorCannotModifyDeclarativeResource
		}
	}

	processedDTO, inboundAuthConfig, svcErr := as.ValidateApplication(ctx, app)
	if svcErr != nil {
		return nil, svcErr
	}

	appID := processedDTO.ID
	assertion := processedDTO.Assertion

	// Validate and prepare the certificate if provided.
	appCert, svcErr := as.getValidatedCertificateForCreate(appID, app.Certificate,
		cert.CertificateReferenceTypeApplication)
	if svcErr != nil {
		return nil, svcErr
	}

	// Validate and prepare the OAuth certificate if provided.
	var oAuthCert *cert.Certificate
	if inboundAuthConfig != nil && inboundAuthConfig.OAuthAppConfig != nil {
		oAuthCert, svcErr = as.getValidatedCertificateForCreate(inboundAuthConfig.OAuthAppConfig.ClientID,
			inboundAuthConfig.OAuthAppConfig.Certificate, cert.CertificateReferenceTypeOAuthApp)
		if svcErr != nil {
			return nil, svcErr
		}
	}

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
		return nil, &ErrorInternalServerError
	}

	_, epErr := as.entityProvider.CreateEntity(appEntity, sysCredsJSON)
	if epErr != nil {
		if svcErr := mapEntityProviderError(epErr); svcErr != nil {
			return nil, svcErr
		}
		as.logger.Error("Failed to create application entity", log.String("appID", appID), log.Error(epErr))
		return nil, &ErrorInternalServerError
	}

	// Create config(with compensation if it fails).
	var returnCert *model.ApplicationCertificate
	var returnOAuthCert *model.ApplicationCertificate
	var innerSvcErr *serviceerror.ServiceError
	err := as.transactioner.Transact(ctx, func(txCtx context.Context) error {
		var certErr *serviceerror.ServiceError
		returnCert, certErr = as.createApplicationCertificate(txCtx, appCert)
		if certErr != nil {
			innerSvcErr = certErr
			return fmt.Errorf("certificate creation failed")
		}

		if inboundAuthConfig != nil && inboundAuthConfig.OAuthAppConfig != nil {
			returnOAuthCert, certErr = as.createApplicationCertificate(txCtx, oAuthCert)
			if certErr != nil {
				innerSvcErr = certErr
				return fmt.Errorf("OAuth certificate creation failed")
			}
		}

		configDAO := toConfigDAO(processedDTO)
		createErr := as.appStore.CreateApplication(txCtx, configDAO)
		if createErr != nil {
			return createErr
		}

		// Create OAuth config in store if present.
		oauthJSON, oauthErr := toOAuthConfigJSON(processedDTO)
		if oauthErr != nil {
			return oauthErr
		}
		if oauthJSON != nil {
			if err := as.appStore.CreateOAuthConfig(txCtx, appID, oauthJSON); err != nil {
				return err
			}
		}

		// Sync consent purpose for the application creation.
		if as.consentService.IsEnabled() {
			if svcErr := as.syncConsentPurposeOnCreate(txCtx, processedDTO); svcErr != nil {
				innerSvcErr = svcErr
				return fmt.Errorf("consent sync failed")
			}
		}

		return nil
	})

	if innerSvcErr != nil {
		// Compensate: delete entity since config creation failed.
		as.deleteEntityCompensation(appID)
		return nil, innerSvcErr
	}

	if err != nil {
		as.logger.Error("Failed to create application", log.Error(err), log.String("appID", appID))
		// Compensate: delete entity since config creation failed.
		as.deleteEntityCompensation(appID)
		return nil, &ErrorInternalServerError
	}

	var oauthToken *model.OAuthTokenConfig
	var userInfo *model.UserInfoConfig
	var scopeClaims map[string][]string

	if inboundAuthConfig != nil {
		processedOAuthConfig := getOAuthInboundAuthConfigProcessedDTO(
			processedDTO.InboundAuthConfig)
		if processedOAuthConfig != nil && processedOAuthConfig.OAuthAppConfig != nil {
			oauthToken = processedOAuthConfig.OAuthAppConfig.Token
			userInfo = processedOAuthConfig.OAuthAppConfig.UserInfo
			scopeClaims = processedOAuthConfig.OAuthAppConfig.ScopeClaims
		} else {
			inboundAuthConfig = nil
		}
	}
	return buildReturnApplicationDTO(appID, app, assertion, returnCert, processedDTO.Metadata,
		inboundAuthConfig, oauthToken, userInfo, scopeClaims, returnOAuthCert), nil
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
	assertion, finalOAuthAccessToken, finalOAuthIDToken := processTokenConfiguration(app)
	userInfo := processUserInfoConfiguration(app, finalOAuthIDToken)
	scopeClaims := processScopeClaimsConfiguration(app)

	processedDTO := buildBaseApplicationProcessedDTO(appID, app, assertion)
	if inboundAuthConfig != nil {
		processedInboundAuthConfig := buildOAuthInboundAuthConfigProcessedDTO(
			appID, inboundAuthConfig,
			&model.OAuthTokenConfig{AccessToken: finalOAuthAccessToken, IDToken: finalOAuthIDToken},
			userInfo, scopeClaims, nil,
		)
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
	}
	return processedDTO, inboundAuthConfig, nil
}

// GetApplicationList list the applications.
func (as *applicationService) GetApplicationList(
	ctx context.Context) (*model.ApplicationListResponse, *serviceerror.ServiceError) {
	totalCount, err := as.appStore.GetTotalApplicationCount(ctx)
	if err != nil {
		as.logger.Error("Failed to retrieve total application count", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	appConfigs, err := as.appStore.GetApplicationList(ctx)
	if err != nil {
		// Check for composite limit exceeded
		if errors.Is(err, errResultLimitExceededInCompositeMode) {
			return nil, &ErrorResultLimitExceeded
		}
		as.logger.Error("Failed to retrieve application list", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	// Batch-fetch entities to get identity data.
	entityIDs := make([]string, 0, len(appConfigs))
	for _, cfg := range appConfigs {
		entityIDs = append(entityIDs, cfg.ID)
	}

	entityMap := make(map[string]*entityprovider.Entity)
	if len(entityIDs) > 0 {
		entities, epErr := as.entityProvider.GetEntitiesByIDs(entityIDs)
		if epErr != nil {
			as.logger.Warn("Failed to batch-fetch entities for app list", log.Error(epErr))
		} else {
			for i := range entities {
				entityMap[entities[i].ID] = &entities[i]
			}
		}
	}

	applicationList := make([]model.BasicApplicationResponse, 0, len(appConfigs))
	for _, cfg := range appConfigs {
		applicationList = append(applicationList, buildBasicApplicationResponse(cfg, entityMap[cfg.ID]))
	}

	response := &model.ApplicationListResponse{
		TotalResults: totalCount,
		Count:        len(appConfigs),
		Applications: applicationList,
	}

	return response, nil
}

// GetOAuthApplication retrieves the OAuth application based on the client id.
// Resolves clientId → entityId via entity provider, then loads only the OAuth config from app store.
func (as *applicationService) GetOAuthApplication(
	ctx context.Context, clientID string) (*model.OAuthAppConfigProcessedDTO, *serviceerror.ServiceError) {
	if clientID == "" {
		return nil, &ErrorInvalidClientID
	}

	// Resolve clientId → entityId via entity identifier.
	entityID, epErr := as.entityProvider.IdentifyEntity(map[string]interface{}{fieldClientID: clientID})
	if epErr != nil {
		if epErr.Code == entityprovider.ErrorCodeEntityNotFound {
			return nil, &ErrorApplicationNotFound
		}
		as.logger.Error("Failed to resolve clientId to entity", log.Error(epErr),
			log.String("clientID", log.MaskString(clientID)))
		return nil, &ErrorInternalServerError
	}
	if entityID == nil {
		return nil, &ErrorApplicationNotFound
	}

	// Load only the OAuth config.
	oauthDAO, err := as.appStore.GetOAuthConfigByAppID(ctx, *entityID)
	if err != nil {
		return nil, as.mapStoreError(err)
	}
	if oauthDAO == nil || oauthDAO.OAuthConfig == nil {
		return nil, &ErrorApplicationNotFound
	}

	// Resolve the application's organization unit
	var ouID string
	if entity, getErr := as.entityProvider.GetEntity(*entityID); getErr == nil && entity != nil {
		ouID = entity.OrganizationUnitID
	} else if getErr != nil {
		as.logger.Error("Failed to fetch entity for OAuth app",
			log.String("appId", *entityID), log.String("error", getErr.Error()))
		return nil, &ErrorInternalServerError
	}

	oauthProcessed := toOAuthProcessedDTO(*entityID, clientID, ouID, oauthDAO)

	certificate, certErr := as.getApplicationCertificate(ctx, clientID, cert.CertificateReferenceTypeOAuthApp)
	if certErr != nil {
		return nil, certErr
	}

	oauthProcessed.Certificate = certificate

	return oauthProcessed, nil
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
//
//nolint:gocyclo // orchestration function — branching is inherent to multi-step update flow
func (as *applicationService) UpdateApplication(ctx context.Context, appID string, app *model.ApplicationDTO) (
	*model.ApplicationDTO, *serviceerror.ServiceError) {
	existingApp, inboundAuthConfig, svcErr := as.validateApplicationForUpdate(ctx, appID, app)

	if svcErr != nil {
		return nil, svcErr
	}

	// Process token configuration
	assertion, finalOAuthAccessToken, finalOAuthIDToken := processTokenConfiguration(app)
	userInfo := processUserInfoConfiguration(app, finalOAuthIDToken)
	scopeClaims := processScopeClaimsConfiguration(app)

	processedDTO := as.buildProcessedDTOForUpdate(
		appID, app,
		inboundAuthConfig,
		assertion, finalOAuthAccessToken, finalOAuthIDToken,
		userInfo, scopeClaims,
	)

	// Update entity identity data.
	var clientID string
	if inboundAuthConfig != nil && inboundAuthConfig.OAuthAppConfig != nil {
		clientID = inboundAuthConfig.OAuthAppConfig.ClientID
	}
	sysAttrsJSON, marshalErr := buildSystemAttributes(app, clientID)
	if marshalErr != nil {
		as.logger.Error("Failed to build entity system attributes for update", log.Error(marshalErr))
		return nil, &ErrorInternalServerError
	}
	if epErr := as.entityProvider.UpdateSystemAttributes(appID, sysAttrsJSON); epErr != nil {
		if svcErr := mapEntityProviderError(epErr); svcErr != nil {
			return nil, svcErr
		}
		as.logger.Error("Failed to update entity system attributes", log.String("appID", appID), log.Error(epErr))
		return nil, &ErrorInternalServerError
	}

	// Update system credentials if a new client secret is provided.
	if inboundAuthConfig != nil && inboundAuthConfig.OAuthAppConfig != nil &&
		inboundAuthConfig.OAuthAppConfig.ClientSecret != "" {
		sysCredsJSON, _ := json.Marshal(map[string]interface{}{
			fieldClientSecret: inboundAuthConfig.OAuthAppConfig.ClientSecret,
		})
		if epErr := as.entityProvider.UpdateSystemCredentials(appID, sysCredsJSON); epErr != nil {
			if svcErr := mapEntityProviderError(epErr); svcErr != nil {
				return nil, svcErr
			}
			as.logger.Error("Failed to update entity system credentials", log.String("appID", appID), log.Error(epErr))
			return nil, &ErrorInternalServerError
		}
	}

	// Update config.
	var returnCert, returnOAuthCert *model.ApplicationCertificate
	var innerSvcErr *serviceerror.ServiceError
	err := as.transactioner.Transact(ctx, func(txCtx context.Context) error {
		var certErr *serviceerror.ServiceError
		returnCert, certErr = as.updateApplicationCertificate(txCtx, appID,
			app.Certificate, cert.CertificateReferenceTypeApplication)
		if certErr != nil {
			innerSvcErr = certErr
			return fmt.Errorf("application certificate update failed")
		}

		if inboundAuthConfig != nil {
			returnOAuthCert, certErr = as.updateApplicationCertificate(
				txCtx, inboundAuthConfig.OAuthAppConfig.ClientID, inboundAuthConfig.OAuthAppConfig.Certificate,
				cert.CertificateReferenceTypeOAuthApp)
			if certErr != nil {
				innerSvcErr = certErr
				return fmt.Errorf("OAuth certificate update failed")
			}
		}

		configDAO := toConfigDAO(processedDTO)
		storeErr := as.appStore.UpdateApplication(txCtx, configDAO)
		if storeErr != nil {
			return storeErr
		}

		// Sync OAuth config
		existingOAuthDAO, _ := as.appStore.GetOAuthConfigByAppID(txCtx, appID)
		oauthJSON, oauthErr := toOAuthConfigJSON(processedDTO)
		if oauthErr != nil {
			return oauthErr
		}
		if oauthJSON != nil && existingOAuthDAO != nil {
			// Update existing OAuth config.
			if err := as.appStore.UpdateOAuthConfig(txCtx, appID, oauthJSON); err != nil {
				return err
			}
		} else if oauthJSON != nil && existingOAuthDAO == nil {
			// Add new OAuth config.
			if err := as.appStore.CreateOAuthConfig(txCtx, appID, oauthJSON); err != nil {
				return err
			}
		} else if oauthJSON == nil && existingOAuthDAO != nil {
			// Remove OAuth config.
			if err := as.appStore.DeleteOAuthConfig(txCtx, appID); err != nil {
				return err
			}
		}

		// Sync consent purpose for the application update
		if as.consentService.IsEnabled() {
			if svcErr := as.syncConsentPurposeOnUpdate(txCtx, existingApp, processedDTO); svcErr != nil {
				innerSvcErr = svcErr
				return fmt.Errorf("consent sync failed")
			}
		}

		return nil
	})

	if innerSvcErr != nil {
		return nil, innerSvcErr
	}
	if err != nil {
		as.logger.Error("Failed to update application", log.Error(err), log.String("appID", appID))
		return nil, &ErrorInternalServerError
	}

	return buildReturnApplicationDTO(appID, app, assertion, returnCert, processedDTO.Metadata,
		inboundAuthConfig, &model.OAuthTokenConfig{AccessToken: finalOAuthAccessToken, IDToken: finalOAuthIDToken},
		userInfo, scopeClaims, returnOAuthCert), nil
}

// DeleteApplication delete the application for given app id.
func (as *applicationService) DeleteApplication(ctx context.Context, appID string) *serviceerror.ServiceError {
	if appID == "" {
		return &ErrorInvalidApplicationID
	}

	// Check if the application is declarative (read-only)
	if as.appStore.IsApplicationDeclarative(ctx, appID) {
		return &ErrorCannotModifyDeclarativeResource
	}

	// Load full application before deletion for cleanup.
	existingFullApp, loadErr := as.getApplication(ctx, appID)
	if loadErr != nil {
		if loadErr == &ErrorApplicationNotFound {
			return nil
		}
		return loadErr
	}

	// Delete config.
	var appNotFound bool
	var transactionSvcErr *serviceerror.ServiceError
	err := as.transactioner.Transact(ctx, func(txCtx context.Context) error {
		appErr := as.appStore.DeleteApplication(txCtx, appID)
		if appErr != nil {
			if errors.Is(appErr, model.ApplicationNotFoundError) {
				as.logger.Debug("Application not found for the deletion", log.String("appID", appID))
				appNotFound = true
				return nil
			}
			return appErr
		}

		if as.consentService.IsEnabled() {
			if svcErr := as.deleteConsentPurposes(txCtx, appID); svcErr != nil {
				transactionSvcErr = svcErr
				return fmt.Errorf("consent deletion failed")
			}
		}

		if svcErr := as.deleteApplicationCertificate(txCtx, appID); svcErr != nil {
			transactionSvcErr = svcErr
			return fmt.Errorf("application certificate deletion failed")
		}

		for _, inboundConfig := range existingFullApp.InboundAuthConfig {
			if inboundConfig.OAuthAppConfig != nil && inboundConfig.OAuthAppConfig.ClientID != "" {
				if svcErr := as.deleteOAuthAppCertificate(txCtx, inboundConfig.OAuthAppConfig.ClientID); svcErr != nil {
					transactionSvcErr = svcErr
					return fmt.Errorf("OAuth app certificate deletion failed")
				}
			}
		}

		return nil
	})

	if appNotFound {
		return nil
	}

	if transactionSvcErr != nil {
		return transactionSvcErr
	}
	if err != nil {
		as.logger.Error("Failed to delete application", log.Error(err), log.String("appID", appID))
		return &ErrorInternalServerError
	}

	// Delete all localized variants from the i18n table (AC-19).
	as.deleteLocalizedVariants(appID)

	// Delete entity.
	if epErr := as.entityProvider.DeleteEntity(appID); epErr != nil {
		if svcErr := mapEntityProviderError(epErr); svcErr != nil {
			return svcErr
		}
		as.logger.Error("Failed to delete application entity", log.String("appID", appID), log.Error(epErr))
		return &ErrorInternalServerError
	}

	return nil
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
		return false, &ErrorInternalServerError
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
	configDAO, err := as.appStore.GetApplicationByID(ctx, appID)
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
			return nil, &ErrorInternalServerError
		}
	}

	oauthDAO, _ := as.appStore.GetOAuthConfigByAppID(ctx, appID)

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
func toConfigDAO(dto *model.ApplicationProcessedDTO) applicationConfigDAO {
	dao := applicationConfigDAO{
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
	e *entityprovider.Entity, dao *applicationConfigDAO, oauthDAO *oauthConfigDAO,
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
		dto.OUID = e.OrganizationUnitID
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
		if contacts, ok := dao.Properties[propContacts].([]interface{}); ok {
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

	// Merge OAuth config if present.
	if oauthDAO != nil && oauthDAO.OAuthConfig != nil {
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
			ouID = e.OrganizationUnitID
		}
		oauthProcessed := toOAuthProcessedDTO(dao.ID, clientID, ouID, oauthDAO)
		dto.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{
			{Type: model.OAuthInboundAuthType, OAuthAppConfig: oauthProcessed},
		}
	}

	return dto
}

// toOAuthConfigJSON builds the OAuth config JSON from a processed DTO for store persistence.
func toOAuthConfigJSON(processedDTO *model.ApplicationProcessedDTO) (json.RawMessage, error) {
	oauthProcessed := getOAuthInboundAuthConfigProcessedDTO(processedDTO.InboundAuthConfig)
	if oauthProcessed == nil || oauthProcessed.OAuthAppConfig == nil {
		return nil, nil
	}
	return getOAuthConfigJSONBytes(*oauthProcessed)
}

// toOAuthProcessedDTO converts an oauthConfigDAO into an OAuthAppConfigProcessedDTO.
func toOAuthProcessedDTO(
	appID, clientID, ouID string, oauthDAO *oauthConfigDAO,
) *model.OAuthAppConfigProcessedDTO {
	cfg := oauthDAO.OAuthConfig
	dto := &model.OAuthAppConfigProcessedDTO{
		AppID:                   appID,
		OUID:                    ouID,
		ClientID:                clientID,
		RedirectURIs:            cfg.RedirectURIs,
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethod(cfg.TokenEndpointAuthMethod),
		PKCERequired:            cfg.PKCERequired,
		PublicClient:            cfg.PublicClient,
		Scopes:                  cfg.Scopes,
		ScopeClaims:             cfg.ScopeClaims,
	}

	for _, gt := range cfg.GrantTypes {
		dto.GrantTypes = append(dto.GrantTypes, oauth2const.GrantType(gt))
	}
	for _, rt := range cfg.ResponseTypes {
		dto.ResponseTypes = append(dto.ResponseTypes, oauth2const.ResponseType(rt))
	}

	if cfg.Token != nil {
		dto.Token = &model.OAuthTokenConfig{}
		if cfg.Token.AccessToken != nil {
			dto.Token.AccessToken = &model.AccessTokenConfig{
				ValidityPeriod: cfg.Token.AccessToken.ValidityPeriod,
				UserAttributes: cfg.Token.AccessToken.UserAttributes,
			}
		}
		if cfg.Token.IDToken != nil {
			dto.Token.IDToken = &model.IDTokenConfig{
				ValidityPeriod: cfg.Token.IDToken.ValidityPeriod,
				UserAttributes: cfg.Token.IDToken.UserAttributes,
			}
		}
	}

	if cfg.UserInfo != nil {
		dto.UserInfo = &model.UserInfoConfig{
			ResponseType:   cfg.UserInfo.ResponseType,
			UserAttributes: cfg.UserInfo.UserAttributes,
		}
	}

	if cfg.Certificate != nil {
		dto.Certificate = cfg.Certificate
	}

	return dto
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

	var sysCredsJSON json.RawMessage
	if plaintextSecret != "" {
		creds := map[string]interface{}{
			fieldClientSecret: plaintextSecret,
		}
		sysCredsJSON, _ = json.Marshal(creds)
	}

	e := &entityprovider.Entity{
		ID:                 appID,
		Category:           entityprovider.EntityCategoryApp,
		Type:               "application",
		State:              entityprovider.EntityStateActive,
		OrganizationUnitID: app.OUID,
		SystemAttributes:   sysAttrsJSON,
	}
	return e, sysCredsJSON, nil
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

	// Check if the application is declarative (read-only)
	if as.appStore.IsApplicationDeclarative(ctx, appID) {
		return nil, nil, &ErrorCannotModifyDeclarativeResource
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

	if svcErr := as.validateAuthFlowID(ctx, app); svcErr != nil {
		return svcErr
	}
	if svcErr := as.validateRegistrationFlowID(ctx, app); svcErr != nil {
		return svcErr
	}
	if svcErr := as.validateThemeID(app.ThemeID); svcErr != nil {
		return svcErr
	}
	if svcErr := as.validateLayoutID(app.LayoutID); svcErr != nil {
		return svcErr
	}
	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return &ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidLogoURI(app.LogoURL) {
		return &ErrorInvalidLogoURL
	}
	if svcErr := as.validateAllowedUserTypes(ctx, app.AllowedUserTypes); svcErr != nil {
		return svcErr
	}
	as.validateConsentConfig(app)
	return nil
}

// validateAuthFlowID validates the auth flow ID for the application.
// If the flow ID is not provided, it sets the default authentication flow ID.
func (as *applicationService) validateAuthFlowID(
	ctx context.Context, app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.AuthFlowID != "" {
		valid, svcErr := as.flowMgtService.IsValidFlow(ctx, app.AuthFlowID, flowcommon.FlowTypeAuthentication)
		if svcErr != nil {
			return svcErr
		}
		if !valid {
			return &ErrorInvalidAuthFlowID
		}
	} else {
		defaultFlowID, svcErr := as.getDefaultAuthFlowID(ctx)
		if svcErr != nil {
			return svcErr
		}
		app.AuthFlowID = defaultFlowID
	}

	return nil
}

// validateRegistrationFlowID validates the registration flow ID for the application.
// If the ID is not provided, it attempts to infer it from the equivalent auth flow ID.
func (as *applicationService) validateRegistrationFlowID(
	ctx context.Context, app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.RegistrationFlowID != "" {
		valid, svcErr := as.flowMgtService.IsValidFlow(ctx, app.RegistrationFlowID, flowcommon.FlowTypeRegistration)
		if svcErr != nil {
			return svcErr
		}
		if !valid {
			return &ErrorInvalidRegistrationFlowID
		}
	} else {
		// Try to get the equivalent registration flow for the auth flow
		authFlow, svcErr := as.flowMgtService.GetFlow(ctx, app.AuthFlowID)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ServerErrorType {
				as.logger.Error("Error while retrieving auth flow definition",
					log.String("flowID", app.AuthFlowID), log.Any("error", svcErr))
				return &serviceerror.InternalServerError
			}
			return &ErrorWhileRetrievingFlowDefinition
		}

		registrationFlow, svcErr := as.flowMgtService.GetFlowByHandle(
			ctx, authFlow.Handle, flowcommon.FlowTypeRegistration)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ServerErrorType {
				as.logger.Error("Error while retrieving registration flow definition by handle",
					log.String("flowHandle", authFlow.Handle), log.Any("error", svcErr))
				return &serviceerror.InternalServerError
			}
			return &ErrorWhileRetrievingFlowDefinition
		}

		app.RegistrationFlowID = registrationFlow.ID
	}

	return nil
}

// validateThemeID validates the theme ID for the application.
func (as *applicationService) validateThemeID(themeID string) *serviceerror.ServiceError {
	if themeID == "" {
		return nil
	}

	exists, svcErr := as.themeMgtService.IsThemeExist(themeID)
	if svcErr != nil {
		return svcErr
	}
	if !exists {
		return &ErrorThemeNotFound
	}

	return nil
}

// validateLayoutID validates the layout ID for the application.
func (as *applicationService) validateLayoutID(layoutID string) *serviceerror.ServiceError {
	if layoutID == "" {
		return nil
	}

	exists, svcErr := as.layoutMgtService.IsLayoutExist(layoutID)
	if svcErr != nil {
		return svcErr
	}
	if !exists {
		return &ErrorLayoutNotFound
	}

	return nil
}

// validateAllowedUserTypes validates that all user types in allowed_user_types exist in the system.
// TODO: Refine validation logic from user schema service.
func (as *applicationService) validateAllowedUserTypes(
	ctx context.Context, allowedUserTypes []string) *serviceerror.ServiceError {
	if len(allowedUserTypes) == 0 {
		return nil
	}

	// Get all user schemas to check if the provided user types exist
	existingUserTypes := make(map[string]bool)
	limit := serverconst.MaxPageSize
	offset := 0

	for {
		// Runtime context is used to avoid authorization checks when fetching user schemas.
		userSchemaList, svcErr := as.userSchemaService.GetUserSchemaList(
			security.WithRuntimeContext(ctx), limit, offset, false)
		if svcErr != nil {
			as.logger.Error("Failed to retrieve user schema list for validation",
				log.String("error", svcErr.Error), log.String("code", svcErr.Code))
			return &ErrorInternalServerError
		}

		for _, schema := range userSchemaList.Schemas {
			existingUserTypes[schema.Name] = true
		}

		if len(userSchemaList.Schemas) == 0 || offset+len(userSchemaList.Schemas) >= userSchemaList.TotalResults {
			break
		}

		offset += limit
	}

	// Check each provided user type
	var invalidUserTypes []string
	for _, userType := range allowedUserTypes {
		if userType == "" {
			// Empty strings are invalid user types
			invalidUserTypes = append(invalidUserTypes, userType)
			continue
		}
		if !existingUserTypes[userType] {
			invalidUserTypes = append(invalidUserTypes, userType)
		}
	}

	if len(invalidUserTypes) > 0 {
		as.logger.Info("Invalid user types found", log.Any("invalidTypes", invalidUserTypes))
		return &ErrorInvalidUserType
	}

	return nil
}

// validateConsentConfig validates the consent configuration for the application.
func (as *applicationService) validateConsentConfig(appDTO *model.ApplicationDTO) {
	if appDTO.LoginConsent == nil {
		appDTO.LoginConsent = &model.LoginConsentConfig{
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

	// Apply defaults for OAuth configuration if not specified.
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

	// Validate redirect URIs
	if err := validateRedirectURIs(oauthAppConfig); err != nil {
		return nil, err
	}

	// Validate grant types and response types
	if err := validateGrantTypesAndResponseTypes(oauthAppConfig); err != nil {
		return nil, err
	}

	// Validate token endpoint authentication method
	if err := validateTokenEndpointAuthMethod(oauthAppConfig); err != nil {
		return nil, err
	}

	// Validate public client configurations
	if oauthAppConfig.PublicClient {
		if err := validatePublicClientConfiguration(oauthAppConfig); err != nil {
			return nil, err
		}
	}

	return inboundAuthConfig, nil
}

// validateRedirectURIs validates redirect URIs format and requirements.
func validateRedirectURIs(oauthConfig *model.OAuthAppConfigDTO) *serviceerror.ServiceError {
	for _, redirectURI := range oauthConfig.RedirectURIs {
		parsedURI, err := sysutils.ParseURL(redirectURI)
		if err != nil {
			return &ErrorInvalidRedirectURI
		}

		if parsedURI.Scheme == "" || parsedURI.Host == "" {
			return &ErrorInvalidRedirectURI
		}

		if parsedURI.Fragment != "" {
			return serviceerror.CustomServiceError(
				ErrorInvalidRedirectURI,
				"Redirect URIs must not contain a fragment component",
			)
		}
	}

	if slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeAuthorizationCode) &&
		len(oauthConfig.RedirectURIs) == 0 {
		return serviceerror.CustomServiceError(
			ErrorInvalidOAuthConfiguration,
			"authorization_code grant type requires redirect URIs",
		)
	}

	return nil
}

// validateGrantTypesAndResponseTypes validates grant types, response types, and their compatibility.
func validateGrantTypesAndResponseTypes(oauthConfig *model.OAuthAppConfigDTO) *serviceerror.ServiceError {
	for _, grantType := range oauthConfig.GrantTypes {
		if !grantType.IsValid() {
			return &ErrorInvalidGrantType
		}
	}

	for _, responseType := range oauthConfig.ResponseTypes {
		if !responseType.IsValid() {
			return &ErrorInvalidResponseType
		}
	}

	if len(oauthConfig.GrantTypes) == 1 &&
		slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeClientCredentials) &&
		len(oauthConfig.ResponseTypes) > 0 {
		return serviceerror.CustomServiceError(
			ErrorInvalidOAuthConfiguration,
			"client_credentials grant type cannot be used with response types",
		)
	}

	if slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeAuthorizationCode) {
		if len(oauthConfig.ResponseTypes) == 0 ||
			!slices.Contains(oauthConfig.ResponseTypes, oauth2const.ResponseTypeCode) {
			return serviceerror.CustomServiceError(
				ErrorInvalidOAuthConfiguration,
				"authorization_code grant type requires 'code' response type",
			)
		}
	}

	return nil
}

// validateTokenEndpointAuthMethod validates the token endpoint authentication method
// and its compatibility with grant types.
func validateTokenEndpointAuthMethod(oauthConfig *model.OAuthAppConfigDTO) *serviceerror.ServiceError {
	if !oauthConfig.TokenEndpointAuthMethod.IsValid() {
		return &ErrorInvalidTokenEndpointAuthMethod
	}

	hasCert := oauthConfig.Certificate != nil && oauthConfig.Certificate.Type != cert.CertificateTypeNone

	switch oauthConfig.TokenEndpointAuthMethod {
	case oauth2const.TokenEndpointAuthMethodPrivateKeyJWT:
		if !hasCert {
			return serviceerror.CustomServiceError(
				ErrorInvalidOAuthConfiguration,
				"private_key_jwt authentication method requires a certificate",
			)
		}
		if oauthConfig.ClientSecret != "" {
			return serviceerror.CustomServiceError(
				ErrorInvalidOAuthConfiguration,
				"private_key_jwt authentication method cannot have a client secret",
			)
		}
	case oauth2const.TokenEndpointAuthMethodClientSecretBasic, oauth2const.TokenEndpointAuthMethodClientSecretPost:
		if hasCert {
			return serviceerror.CustomServiceError(
				ErrorInvalidOAuthConfiguration,
				"client_secret authentication methods cannot have a certificate",
			)
		}
	case oauth2const.TokenEndpointAuthMethodNone:
		if !oauthConfig.PublicClient {
			return serviceerror.CustomServiceError(
				ErrorInvalidOAuthConfiguration,
				"'none' authentication method requires the client to be a public client",
			)
		}
		if hasCert || oauthConfig.ClientSecret != "" {
			return serviceerror.CustomServiceError(
				ErrorInvalidOAuthConfiguration,
				"'none' authentication method cannot have a certificate or client secret",
			)
		}
		if slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeClientCredentials) {
			return serviceerror.CustomServiceError(
				ErrorInvalidOAuthConfiguration,
				"client_credentials grant type cannot use 'none' authentication method",
			)
		}
	}

	return nil
}

// validatePublicClientConfiguration validates that public client configurations are correct.
func validatePublicClientConfiguration(oauthConfig *model.OAuthAppConfigDTO) *serviceerror.ServiceError {
	if oauthConfig.TokenEndpointAuthMethod != oauth2const.TokenEndpointAuthMethodNone {
		return serviceerror.CustomServiceError(
			ErrorInvalidPublicClientConfiguration,
			"Public clients must use 'none' as token endpoint authentication method",
		)
	}

	// Public clients must always have PKCE required for security
	if !oauthConfig.PKCERequired {
		return serviceerror.CustomServiceError(
			ErrorInvalidPublicClientConfiguration,
			"Public clients must have PKCE required set to true",
		)
	}

	return nil
}

// processInboundAuthConfig validates and processes inbound auth configuration for
// creating or updating an application.
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

	// Resolve client secret for confidential clients
	if svcErr := resolveClientSecret(inboundAuthConfig, existingApp); svcErr != nil {
		return nil, svcErr
	}

	return inboundAuthConfig, nil
}

// getDefaultAuthFlowID retrieves the default authentication flow ID from the configuration.
func (as *applicationService) getDefaultAuthFlowID(ctx context.Context) (string, *serviceerror.ServiceError) {
	defaultAuthFlowHandle := config.GetThunderRuntime().Config.Flow.DefaultAuthFlowHandle
	defaultAuthFlow, svcErr := as.flowMgtService.GetFlowByHandle(
		ctx, defaultAuthFlowHandle, flowcommon.FlowTypeAuthentication)

	if svcErr != nil {
		if svcErr.Type == serviceerror.ServerErrorType {
			as.logger.Error("Error while retrieving default auth flow definition by handle",
				log.String("flowHandle", defaultAuthFlowHandle), log.Any("error", svcErr))
			return "", &serviceerror.InternalServerError
		}
		return "", &ErrorWhileRetrievingFlowDefinition
	}

	return defaultAuthFlow.ID, nil
}

// getDefaultAssertionConfigFromDeployment creates a default assertion configuration from deployment settings.
func getDefaultAssertionConfigFromDeployment() *model.AssertionConfig {
	jwtConfig := config.GetThunderRuntime().Config.JWT
	assertionConfig := &model.AssertionConfig{
		ValidityPeriod: jwtConfig.ValidityPeriod,
	}

	return assertionConfig
}

// processTokenConfiguration processes token configuration for an application, applying defaults where necessary.
func processTokenConfiguration(app *model.ApplicationDTO) (
	*model.AssertionConfig, *model.AccessTokenConfig, *model.IDTokenConfig) {
	// Resolve root assertion config
	var assertion *model.AssertionConfig
	if app.Assertion != nil {
		assertion = &model.AssertionConfig{
			ValidityPeriod: app.Assertion.ValidityPeriod,
			UserAttributes: app.Assertion.UserAttributes,
		}

		deploymentDefaults := getDefaultAssertionConfigFromDeployment()
		if assertion.ValidityPeriod == 0 {
			assertion.ValidityPeriod = deploymentDefaults.ValidityPeriod
		}
	} else {
		assertion = getDefaultAssertionConfigFromDeployment()
	}
	if assertion.UserAttributes == nil {
		assertion.UserAttributes = make([]string, 0)
	}

	// Resolve OAuth access token config
	var oauthAccessToken *model.AccessTokenConfig
	oauthInboundAuth, _ := getOAuthInboundAuthConfigDTO(app.InboundAuthConfig)
	if oauthInboundAuth != nil && oauthInboundAuth.OAuthAppConfig != nil &&
		oauthInboundAuth.OAuthAppConfig.Token != nil &&
		oauthInboundAuth.OAuthAppConfig.Token.AccessToken != nil {
		oauthAccessToken = &model.AccessTokenConfig{
			ValidityPeriod: oauthInboundAuth.OAuthAppConfig.Token.AccessToken.ValidityPeriod,
			UserAttributes: oauthInboundAuth.OAuthAppConfig.Token.AccessToken.UserAttributes,
		}
	}

	if oauthAccessToken != nil {
		if oauthAccessToken.ValidityPeriod == 0 {
			oauthAccessToken.ValidityPeriod = assertion.ValidityPeriod
		}
		if oauthAccessToken.UserAttributes == nil {
			oauthAccessToken.UserAttributes = make([]string, 0)
		}
	} else {
		oauthAccessToken = &model.AccessTokenConfig{
			ValidityPeriod: assertion.ValidityPeriod,
			UserAttributes: assertion.UserAttributes,
		}
	}

	// Resolve OAuth ID token config
	var oauthIDToken *model.IDTokenConfig
	if oauthInboundAuth != nil && oauthInboundAuth.OAuthAppConfig != nil &&
		oauthInboundAuth.OAuthAppConfig.Token != nil &&
		oauthInboundAuth.OAuthAppConfig.Token.IDToken != nil {
		oauthIDToken = &model.IDTokenConfig{
			ValidityPeriod: oauthInboundAuth.OAuthAppConfig.Token.IDToken.ValidityPeriod,
			UserAttributes: oauthInboundAuth.OAuthAppConfig.Token.IDToken.UserAttributes,
		}
	}

	if oauthIDToken != nil {
		if oauthIDToken.ValidityPeriod == 0 {
			oauthIDToken.ValidityPeriod = assertion.ValidityPeriod
		}
		if oauthIDToken.UserAttributes == nil {
			oauthIDToken.UserAttributes = make([]string, 0)
		}
	} else {
		oauthIDToken = &model.IDTokenConfig{
			ValidityPeriod: assertion.ValidityPeriod,
			UserAttributes: assertion.UserAttributes,
		}
	}

	return assertion, oauthAccessToken, oauthIDToken
}

// processUserInfoConfiguration processes user info configuration for an application.
func processUserInfoConfiguration(app *model.ApplicationDTO,
	idTokenConfig *model.IDTokenConfig) *model.UserInfoConfig {
	oauthUserInfo := &model.UserInfoConfig{}

	oauthInboundAuth, _ := getOAuthInboundAuthConfigDTO(app.InboundAuthConfig)
	if oauthInboundAuth != nil && oauthInboundAuth.OAuthAppConfig != nil &&
		oauthInboundAuth.OAuthAppConfig.UserInfo != nil {
		userInfoConfigInput := oauthInboundAuth.OAuthAppConfig.UserInfo
		oauthUserInfo.UserAttributes = userInfoConfigInput.UserAttributes
		responseType := model.UserInfoResponseType(strings.ToUpper(string(userInfoConfigInput.ResponseType)))

		switch responseType {
		case model.UserInfoResponseTypeJWS:
			oauthUserInfo.ResponseType = responseType
		default:
			oauthUserInfo.ResponseType = model.UserInfoResponseTypeJSON
		}
	}
	if oauthUserInfo.UserAttributes == nil {
		oauthUserInfo.UserAttributes = idTokenConfig.UserAttributes
	}
	if oauthUserInfo.ResponseType == "" {
		oauthUserInfo.ResponseType = model.UserInfoResponseTypeJSON
	}

	return oauthUserInfo
}

// processScopeClaimsConfiguration processes scope claims configuration for an application.
func processScopeClaimsConfiguration(app *model.ApplicationDTO) map[string][]string {
	var scopeClaims map[string][]string
	oauthInboundAuth, _ := getOAuthInboundAuthConfigDTO(app.InboundAuthConfig)
	if oauthInboundAuth != nil && oauthInboundAuth.OAuthAppConfig != nil &&
		oauthInboundAuth.OAuthAppConfig.ScopeClaims != nil {
		scopeClaims = oauthInboundAuth.OAuthAppConfig.ScopeClaims
	}
	if scopeClaims == nil {
		scopeClaims = make(map[string][]string)
	}

	return scopeClaims
}

// generateAndAssignClientID generates an OAuth 2.0 compliant client ID and assigns it to the inbound auth config.
func generateAndAssignClientID(inboundAuthConfig *model.InboundAuthConfigDTO) *serviceerror.ServiceError {
	generatedClientID, err := oauthutils.GenerateOAuth2ClientID()
	if err != nil {
		log.GetLogger().Error("Failed to generate OAuth client ID", log.Error(err))
		return &ErrorInternalServerError
	}
	inboundAuthConfig.OAuthAppConfig.ClientID = generatedClientID
	return nil
}

// resolveClientSecret generates a new client secret for confidential clients if needed.
// It preserves existing secrets during update operations unless explicitly provided.
func resolveClientSecret(
	inboundAuthConfig *model.InboundAuthConfigDTO,
	existingApp *model.ApplicationProcessedDTO,
) *serviceerror.ServiceError {
	// Only process confidential clients that use client_secret auth method and don't have a secret provided
	if (inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod !=
		oauth2const.TokenEndpointAuthMethodClientSecretBasic &&
		inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod !=
			oauth2const.TokenEndpointAuthMethodClientSecretPost) ||
		inboundAuthConfig.OAuthAppConfig.ClientSecret != "" {
		return nil
	}

	// Check if we should preserve existing confidential OAuth config secret.
	// If the existing app already uses a secret-based auth method, the entity layer
	// already has the hashed credential — no need to regenerate.
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

	// Generate OAuth 2.0 compliant client secret with high entropy for security
	generatedClientSecret, err := oauthutils.GenerateOAuth2ClientSecret()
	if err != nil {
		log.GetLogger().Error("Failed to generate OAuth client secret", log.Error(err))
		return &ErrorInternalServerError
	}

	inboundAuthConfig.OAuthAppConfig.ClientSecret = generatedClientSecret
	return nil
}

// getValidatedCertificateForCreate validates and returns the certificate for the application during creation.
func (as *applicationService) getValidatedCertificateForCreate(appID string, certificate *model.ApplicationCertificate,
	certRefType cert.CertificateReferenceType) (
	*cert.Certificate, *serviceerror.ServiceError) {
	if certificate == nil || certificate.Type == "" || certificate.Type == cert.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(appID, "", certificate, certRefType)
}

// getValidatedCertificateForUpdate validates and returns the certificate for the application during update.
func (as *applicationService) getValidatedCertificateForUpdate(appID, certID string,
	certificate *model.ApplicationCertificate, certRefType cert.CertificateReferenceType) (
	*cert.Certificate, *serviceerror.ServiceError) {
	if certificate == nil || certificate.Type == "" || certificate.Type == cert.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(appID, certID, certificate, certRefType)
}

// getValidatedCertificateInput is a helper method that validates and returns the certificate.
func getValidatedCertificateInput(appID, certID string, certificate *model.ApplicationCertificate,
	certRefType cert.CertificateReferenceType) (*cert.Certificate, *serviceerror.ServiceError) {
	switch certificate.Type {
	case cert.CertificateTypeJWKS:
		if certificate.Value == "" {
			return nil, &ErrorInvalidCertificateValue
		}
		return &cert.Certificate{
			ID:      certID,
			RefType: certRefType,
			RefID:   appID,
			Type:    cert.CertificateTypeJWKS,
			Value:   certificate.Value,
		}, nil
	case cert.CertificateTypeJWKSURI:
		if !sysutils.IsValidURI(certificate.Value) {
			return nil, &ErrorInvalidJWKSURI
		}
		return &cert.Certificate{
			ID:      certID,
			RefType: certRefType,
			RefID:   appID,
			Type:    cert.CertificateTypeJWKSURI,
			Value:   certificate.Value,
		}, nil
	default:
		return nil, &ErrorInvalidCertificateType
	}
}

// createApplicationCertificate creates a certificate for the application.
func (as *applicationService) createApplicationCertificate(ctx context.Context, certificate *cert.Certificate) (
	*model.ApplicationCertificate, *serviceerror.ServiceError) {
	var returnCert *model.ApplicationCertificate
	if certificate != nil {
		_, svcErr := as.certService.CreateCertificate(ctx, certificate)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ClientErrorType {
				errorDescription := "Failed to create application certificate: " +
					svcErr.ErrorDescription
				return nil, serviceerror.CustomServiceError(
					ErrorCertificateClientError, errorDescription)
			}
			as.logger.Error("Failed to create application certificate", log.Any("serviceError", svcErr))
			return nil, &ErrorCertificateServerError
		}

		returnCert = &model.ApplicationCertificate{
			Type:  certificate.Type,
			Value: certificate.Value,
		}
	} else {
		returnCert = &model.ApplicationCertificate{
			Type:  cert.CertificateTypeNone,
			Value: "",
		}
	}

	return returnCert, nil
}

// deleteApplicationCertificate deletes the certificate associated with the application.
func (as *applicationService) deleteApplicationCertificate(
	ctx context.Context, appID string) *serviceerror.ServiceError {
	if certErr := as.certService.DeleteCertificateByReference(
		ctx, cert.CertificateReferenceTypeApplication, appID); certErr != nil {
		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to delete application certificate: " +
				certErr.ErrorDescription
			return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		as.logger.Error("Failed to delete application certificate", log.String("appID", appID),
			log.Any("serviceError", certErr))
		return &ErrorCertificateServerError
	}

	return nil
}

// deleteOAuthAppCertificate deletes the certificate associated with an OAuth app (by client ID).
func (as *applicationService) deleteOAuthAppCertificate(
	ctx context.Context, clientID string) *serviceerror.ServiceError {
	if certErr := as.certService.DeleteCertificateByReference(
		ctx, cert.CertificateReferenceTypeOAuthApp, clientID); certErr != nil {
		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to delete OAuth app certificate: " +
				certErr.ErrorDescription
			return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		as.logger.Error("Failed to delete OAuth app certificate", log.String("clientID", clientID),
			log.Any("serviceError", certErr))
		return &ErrorCertificateServerError
	}

	return nil
}

// getApplicationCertificate retrieves the certificate associated with the application based
// on the reference type (application or OAuth app).
func (as *applicationService) getApplicationCertificate(ctx context.Context, appID string,
	refType cert.CertificateReferenceType) (*model.ApplicationCertificate, *serviceerror.ServiceError) {
	certificate, certErr := as.certService.GetCertificateByReference(
		ctx, refType, appID)

	if certErr != nil {
		if certErr.Code == cert.ErrorCertificateNotFound.Code {
			return &model.ApplicationCertificate{
				Type:  cert.CertificateTypeNone,
				Value: "",
			}, nil
		}

		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to retrieve application certificate: " +
				certErr.ErrorDescription
			return nil, serviceerror.CustomServiceError(
				ErrorCertificateClientError, errorDescription)
		}
		as.logger.Error("Failed to retrieve application certificate", log.Any("serviceError", certErr),
			log.String("appID", appID))
		return nil, &ErrorCertificateServerError
	}

	if certificate == nil {
		return &model.ApplicationCertificate{
			Type:  cert.CertificateTypeNone,
			Value: "",
		}, nil
	}

	return &model.ApplicationCertificate{
		Type:  certificate.Type,
		Value: certificate.Value,
	}, nil
}

// updateApplicationCertificate updates the certificate for the application.
// It returns the updated application certificate details.
func (as *applicationService) updateApplicationCertificate(ctx context.Context, appID string,
	certificate *model.ApplicationCertificate, refType cert.CertificateReferenceType) (
	*model.ApplicationCertificate, *serviceerror.ServiceError) {
	existingCert, certErr := as.certService.GetCertificateByReference(
		ctx, refType, appID)
	if certErr != nil && certErr.Code != cert.ErrorCertificateNotFound.Code {
		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to retrieve application certificate: " +
				certErr.ErrorDescription
			return nil, serviceerror.CustomServiceError(
				ErrorCertificateClientError, errorDescription)
		}
		as.logger.Error("Failed to retrieve application certificate", log.Any("serviceError", certErr),
			log.String("appID", appID))
		return nil, &ErrorCertificateServerError
	}

	var updatedCert *cert.Certificate
	var err *serviceerror.ServiceError
	if existingCert != nil {
		updatedCert, err = as.getValidatedCertificateForUpdate(appID, existingCert.ID, certificate, refType)
	} else {
		updatedCert, err = as.getValidatedCertificateForUpdate(appID, "", certificate, refType)
	}
	if err != nil {
		return nil, err
	}

	// Update the certificate if provided.
	var returnCert *model.ApplicationCertificate
	if updatedCert != nil {
		if existingCert != nil {
			_, svcErr := as.certService.UpdateCertificateByID(ctx, existingCert.ID, updatedCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					errorDescription := "Failed to update application certificate: " +
						svcErr.ErrorDescription
					return nil, serviceerror.CustomServiceError(
						ErrorCertificateClientError, errorDescription)
				}
				as.logger.Error("Failed to update application certificate", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return nil, &ErrorCertificateServerError
			}
		} else {
			_, svcErr := as.certService.CreateCertificate(ctx, updatedCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					errorDescription := "Failed to create application certificate: " +
						svcErr.ErrorDescription
					return nil, serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
				}
				as.logger.Error("Failed to create application certificate", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return nil, &ErrorCertificateServerError
			}
		}
		returnCert = &model.ApplicationCertificate{
			Type:  updatedCert.Type,
			Value: updatedCert.Value,
		}
	} else {
		if existingCert != nil {
			// If no new certificate is provided, delete the existing certificate.
			deleteErr := as.certService.DeleteCertificateByReference(
				ctx, refType, appID)
			if deleteErr != nil {
				if deleteErr.Type == serviceerror.ClientErrorType {
					errorDescription := "Failed to delete application certificate: " + deleteErr.ErrorDescription
					return nil, serviceerror.CustomServiceError(
						ErrorCertificateClientError, errorDescription)
				}
				as.logger.Error("Failed to delete application certificate", log.Any("serviceError", deleteErr),
					log.String("appID", appID))
				return nil, &ErrorCertificateServerError
			}
		}

		returnCert = &model.ApplicationCertificate{
			Type:  cert.CertificateTypeNone,
			Value: "",
		}
	}

	return returnCert, nil
}

// enrichApplicationWithCertificate retrieves and adds the certificate to the application.
func (as *applicationService) enrichApplicationWithCertificate(ctx context.Context, application *model.Application) (
	*model.Application, *serviceerror.ServiceError) {
	appCert, certErr := as.getApplicationCertificate(ctx, application.ID, cert.CertificateReferenceTypeApplication)
	if certErr != nil {
		return nil, certErr
	}
	application.Certificate = appCert

	// Enrich OAuth config certificate for each inbound auth config.
	for i, inboundAuthConfig := range application.InboundAuthConfig {
		if inboundAuthConfig.Type == model.OAuthInboundAuthType && inboundAuthConfig.OAuthAppConfig != nil {
			oauthCert, oauthCertErr := as.getApplicationCertificate(ctx, inboundAuthConfig.OAuthAppConfig.ClientID,
				cert.CertificateReferenceTypeOAuthApp)
			if oauthCertErr != nil {
				return nil, oauthCertErr
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
					ClientID:                oauthAppConfig.ClientID,
					RedirectURIs:            oauthAppConfig.RedirectURIs,
					GrantTypes:              oauthAppConfig.GrantTypes,
					ResponseTypes:           oauthAppConfig.ResponseTypes,
					TokenEndpointAuthMethod: oauthAppConfig.TokenEndpointAuthMethod,
					PKCERequired:            oauthAppConfig.PKCERequired,
					PublicClient:            oauthAppConfig.PublicClient,
					Token:                   oauthAppConfig.Token,
					Scopes:                  oauthAppConfig.Scopes,
					UserInfo:                oauthAppConfig.UserInfo,
					ScopeClaims:             oauthAppConfig.ScopeClaims,
				},
			})
		}
	}
	application.InboundAuthConfig = inboundAuthConfigs
	return application
}

// buildBasicApplicationResponse builds a BasicApplicationResponse by merging config + entity data.
func buildBasicApplicationResponse(cfg applicationConfigDAO, e *entityprovider.Entity) model.BasicApplicationResponse {
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
	assertion *model.AssertionConfig) *model.ApplicationProcessedDTO {
	return &model.ApplicationProcessedDTO{
		ID:                        appID,
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

// buildProcessedDTOForUpdate constructs the ApplicationProcessedDTO for an application update operation.
func (as *applicationService) buildProcessedDTOForUpdate(appID string, app *model.ApplicationDTO,
	inboundAuthConfig *model.InboundAuthConfigDTO,
	assertion *model.AssertionConfig, finalOAuthAccessToken *model.AccessTokenConfig,
	finalOAuthIDToken *model.IDTokenConfig, userInfo *model.UserInfoConfig,
	scopeClaims map[string][]string) *model.ApplicationProcessedDTO {
	processedDTO := buildBaseApplicationProcessedDTO(appID, app, assertion)

	if inboundAuthConfig != nil {
		processedInboundAuthConfig := buildOAuthInboundAuthConfigProcessedDTO(
			appID, inboundAuthConfig,
			&model.OAuthTokenConfig{AccessToken: finalOAuthAccessToken, IDToken: finalOAuthIDToken},
			userInfo, scopeClaims, inboundAuthConfig.OAuthAppConfig.Certificate,
		)
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
	}

	return processedDTO
}

// buildOAuthInboundAuthConfigProcessedDTO constructs the InboundAuthConfigProcessedDTO for an OAuth application.
func buildOAuthInboundAuthConfigProcessedDTO(
	appID string, inboundAuthConfig *model.InboundAuthConfigDTO,
	oauthToken *model.OAuthTokenConfig, userInfo *model.UserInfoConfig,
	scopeClaims map[string][]string, certificate *model.ApplicationCertificate,
) model.InboundAuthConfigProcessedDTO {
	return model.InboundAuthConfigProcessedDTO{
		Type: model.OAuthInboundAuthType,
		OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
			AppID:                   appID,
			ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
			RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
			GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
			ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
			TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
			PKCERequired:            inboundAuthConfig.OAuthAppConfig.PKCERequired,
			PublicClient:            inboundAuthConfig.OAuthAppConfig.PublicClient,
			Token:                   oauthToken,
			Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
			UserInfo:                userInfo,
			ScopeClaims:             scopeClaims,
			Certificate:             certificate,
		},
	}
}

// buildReturnApplicationDTO constructs the ApplicationDTO returned from create and update operations.
func buildReturnApplicationDTO(
	appID string, app *model.ApplicationDTO, assertion *model.AssertionConfig,
	returnCert *model.ApplicationCertificate, metadata map[string]any,
	inboundAuthConfig *model.InboundAuthConfigDTO, oauthToken *model.OAuthTokenConfig,
	userInfo *model.UserInfoConfig, scopeClaims map[string][]string,
	returnOAuthCert *model.ApplicationCertificate) *model.ApplicationDTO {
	returnApp := &model.ApplicationDTO{
		ID:                        appID,
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
		Certificate:               returnCert,
		TosURI:                    app.TosURI,
		PolicyURI:                 app.PolicyURI,
		Contacts:                  app.Contacts,
		AllowedUserTypes:          app.AllowedUserTypes,
		LoginConsent:              app.LoginConsent,
		Metadata:                  metadata,
	}
	if inboundAuthConfig != nil {
		returnInboundAuthConfig := model.InboundAuthConfigDTO{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				ClientSecret:            inboundAuthConfig.OAuthAppConfig.ClientSecret,
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				PKCERequired:            inboundAuthConfig.OAuthAppConfig.PKCERequired,
				PublicClient:            inboundAuthConfig.OAuthAppConfig.PublicClient,
				Token:                   oauthToken,
				Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
				UserInfo:                userInfo,
				ScopeClaims:             scopeClaims,
				Certificate:             returnOAuthCert,
			},
		}
		returnApp.InboundAuthConfig = []model.InboundAuthConfigDTO{returnInboundAuthConfig}
	}
	return returnApp
}

// mapStoreError handles common error scenarios when retrieving applications from the
// application store. It maps specific errors, such as ApplicationNotFoundError, to corresponding service errors.
func (as *applicationService) mapStoreError(err error) *serviceerror.ServiceError {
	if errors.Is(err, model.ApplicationNotFoundError) {
		return &ErrorApplicationNotFound
	}
	as.logger.Error("Failed to retrieve application", log.Error(err))
	return &ErrorInternalServerError
}

// wrapConsentServiceError converts an I18nServiceError from the consent service into a ServiceError
// for the application service.
func (as *applicationService) wrapConsentServiceError(err *serviceerror.I18nServiceError) *serviceerror.ServiceError {
	if err == nil {
		return nil
	}

	if err.Type == serviceerror.ClientErrorType {
		as.logger.Debug("Failed to sync consent purpose for the application changes", log.Any("error", err))
		return serviceerror.CustomServiceError(ErrorConsentSyncFailed,
			fmt.Sprintf(ErrorConsentSyncFailed.ErrorDescription+" : code - %s", err.Code))
	}

	as.logger.Error("Failed to sync consent purpose for the application changes", log.Any("error", err))
	return &ErrorInternalServerError
}

// extractRequestedAttributes collects all unique user attributes requested by the application
// across various configurations including assertions, token config, and user info.
func extractRequestedAttributes(app *model.ApplicationProcessedDTO) map[string]bool {
	if app == nil {
		return nil
	}

	attrMap := make(map[string]bool)

	// Extract from assertion configuration
	if app.Assertion != nil && len(app.Assertion.UserAttributes) > 0 {
		for _, attr := range app.Assertion.UserAttributes {
			attrMap[attr] = true
		}
	}

	// Extract from inbound authentication configurations
	for _, inbound := range app.InboundAuthConfig {
		if inbound.Type == model.OAuthInboundAuthType && inbound.OAuthAppConfig != nil {
			oauthConfig := inbound.OAuthAppConfig

			// Extract from access token
			if oauthConfig.Token != nil && oauthConfig.Token.AccessToken != nil {
				for _, attr := range oauthConfig.Token.AccessToken.UserAttributes {
					attrMap[attr] = true
				}
			}

			// Extract from ID token
			if oauthConfig.Token != nil && oauthConfig.Token.IDToken != nil {
				for _, attr := range oauthConfig.Token.IDToken.UserAttributes {
					attrMap[attr] = true
				}
			}

			// Extract from user info
			if oauthConfig.UserInfo != nil {
				for _, attr := range oauthConfig.UserInfo.UserAttributes {
					attrMap[attr] = true
				}
			}
		}
	}

	return attrMap
}

// syncConsentPurposeOnCreate creates a consent purpose for the application create.
func (as *applicationService) syncConsentPurposeOnCreate(
	ctx context.Context, appDTO *model.ApplicationProcessedDTO) *serviceerror.ServiceError {
	// TODO: Replace with application's actual OU when OU support is added
	const ouID = "default"

	as.logger.Debug("Attempting to synchronize consent purpose for the newly created application",
		log.String("appID", appDTO.ID))

	attributesMap := extractRequestedAttributes(appDTO)

	// Skip consent purpose creation if there are no user attributes requested by the application
	if len(attributesMap) == 0 {
		as.logger.Debug("No user attributes requested by the application, skipping consent purpose creation",
			log.String("appID", appDTO.ID))
		return nil
	}

	attributes := make([]string, 0, len(attributesMap))
	for attr := range attributesMap {
		attributes = append(attributes, attr)
	}

	// Create missing consent elements in case they're not created during user type creation
	// or by another application. This is to ensure that all required consent elements exist
	// before creating the consent purpose.
	if err := as.createMissingConsentElements(ctx, ouID, attributes); err != nil {
		return err
	}

	as.logger.Debug("Creating consent purpose for the newly created application", log.String("appID", appDTO.ID),
		log.Int("attributesCount", len(attributes)))

	purpose := consent.ConsentPurposeInput{
		Name:        appDTO.Name,
		Description: "Consent purpose for application " + appDTO.Name,
		GroupID:     appDTO.ID,
		Elements:    attributesToPurposeElements(attributesMap),
	}
	if _, err := as.consentService.CreateConsentPurpose(ctx, ouID, &purpose); err != nil {
		return as.wrapConsentServiceError(err)
	}

	return nil
}

// syncConsentPurposeOnUpdate synchronizes the consent purpose when an application is updated.
// It updates the existing consent purpose to match the updated application configuration or
// deletes it if all attributes are removed.
func (as *applicationService) syncConsentPurposeOnUpdate(ctx context.Context,
	existingAppDTO, updatedAppDTO *model.ApplicationProcessedDTO) *serviceerror.ServiceError {
	// TODO: Replace with application's actual OU when OU support is added
	const ouID = "default"

	as.logger.Debug("Attempting to synchronize consent purpose for the updated application",
		log.String("appID", existingAppDTO.ID))

	// Find out the attributes that need to be part of the consent purpose based on the updated application
	newAttributes := extractRequestedAttributes(updatedAppDTO)

	// We need to ensure that consent elements exist for all requested attributes
	// regardless of what existed in the old application configuration because the consent
	// purpose might not have been created previously if consent was disabled.
	requiredAttributes := make([]string, 0, len(newAttributes))
	for attr := range newAttributes {
		requiredAttributes = append(requiredAttributes, attr)
	}

	if len(requiredAttributes) > 0 {
		as.logger.Debug("Ensuring consent elements exist for all requested attributes",
			log.String("appID", existingAppDTO.ID), log.Int("requiredAttributesCount", len(requiredAttributes)))

		if err := as.createMissingConsentElements(ctx, ouID, requiredAttributes); err != nil {
			return err
		}
	}

	// Retrieve the existing consent purposes for the application
	existingPurposes, err := as.consentService.ListConsentPurposes(ctx, ouID, existingAppDTO.ID)
	if err != nil {
		return as.wrapConsentServiceError(err)
	}

	// If there are no existing purposes handle separately
	if len(existingPurposes) == 0 {
		as.logger.Debug("No existing consent purpose found for the application", log.String("appID", existingAppDTO.ID))

		// If attributes exists in the updated payload, create a new consent purpose
		if len(newAttributes) > 0 {
			as.logger.Debug("Creating new consent purpose for the application", log.String("appID", existingAppDTO.ID))

			purpose := consent.ConsentPurposeInput{
				Name:        updatedAppDTO.Name,
				Description: "Consent purpose for application " + updatedAppDTO.Name,
				GroupID:     existingAppDTO.ID,
				Elements:    attributesToPurposeElements(newAttributes),
			}
			if _, err := as.consentService.CreateConsentPurpose(ctx, ouID, &purpose); err != nil {
				return as.wrapConsentServiceError(err)
			}
		}

		return nil
	}

	// If all attributes are removed, and a purpose exists, delete it
	if len(newAttributes) == 0 {
		as.logger.Debug("All user attributes removed from the application", log.String("appID", existingAppDTO.ID))
		if err := as.deleteConsentPurposes(ctx, existingAppDTO.ID); err != nil {
			return err
		}

		return nil
	}

	as.logger.Debug("Existing consent purpose found for the application, updating it with the specified attributes",
		log.String("appID", existingAppDTO.ID))

	// Update existing purpose with the application changes.
	// We assume there is only one consent purpose per application
	updated := consent.ConsentPurposeInput{
		Name:        updatedAppDTO.Name,
		Description: "Consent purpose for application " + updatedAppDTO.Name,
		GroupID:     existingAppDTO.ID,
		Elements:    attributesToPurposeElements(newAttributes),
	}
	if _, err := as.consentService.UpdateConsentPurpose(ctx, ouID,
		existingPurposes[0].ID, &updated); err != nil {
		return as.wrapConsentServiceError(err)
	}

	return nil
}

// deleteConsentPurposes removes all consent purposes associated with an application.
func (as *applicationService) deleteConsentPurposes(ctx context.Context, appID string) *serviceerror.ServiceError {
	// TODO: Replace with application's actual OU when OU support is added
	const ouID = "default"

	as.logger.Debug("Attempting to delete consent purposes for the application", log.String("appID", appID))

	purposes, err := as.consentService.ListConsentPurposes(ctx, ouID, appID)
	if err != nil {
		return as.wrapConsentServiceError(err)
	}

	// If there are no purposes, return early
	if len(purposes) == 0 {
		as.logger.Debug("No consent purposes found for the application", log.String("appID", appID))
		return nil
	}

	// We assume there is only one consent purpose per application
	as.logger.Debug("Deleting consent purpose for the application", log.String("appID", appID),
		log.Int("purposesCount", len(purposes)))
	if err := as.consentService.DeleteConsentPurpose(ctx, ouID, purposes[0].ID); err != nil {
		// TODO: Default consent service implementation doesn't allow deleting consent purposes with existing consents.
		//  We need to handle this case gracefully until the consent service supports force delete or cascade delete
		// for consent purposes.
		if err.Code == consent.ErrorDeletingConsentPurposeWithAssociatedRecords.Code {
			as.logger.Warn("Cannot delete consent purpose due to existing consents. Consent service doesn't support "+
				"deleting consent purposes with existing consents",
				log.String("appID", appID), log.Int("purposesCount", len(purposes)))
			return nil
		}

		return as.wrapConsentServiceError(err)
	}

	return nil
}

// createMissingConsentElements validates a list of consent element names and creates only the missing ones.
// nolint:unparam // ouID is always "default" in current usage but kept for future flexibility
func (as *applicationService) createMissingConsentElements(ctx context.Context,
	ouID string, names []string) *serviceerror.ServiceError {
	if len(names) == 0 {
		as.logger.Debug("No consent elements to create", log.String("ouID", ouID))
		return nil
	}

	validNames, err := as.consentService.ValidateConsentElements(ctx, ouID, names)
	if err != nil {
		return as.wrapConsentServiceError(err)
	}

	// Create a map of existing elements for fast lookup
	existingMap := make(map[string]bool, len(validNames))
	for _, name := range validNames {
		existingMap[name] = true
	}

	// Filter out the existing elements
	var elementsToCreate []consent.ConsentElementInput
	for _, name := range names {
		if !existingMap[name] {
			elementsToCreate = append(elementsToCreate, consent.ConsentElementInput{
				Name:      name,
				Namespace: consent.NamespaceAttribute,
			})
		}
	}

	if len(elementsToCreate) > 0 {
		as.logger.Debug("Creating missing consent elements", log.String("ouID", ouID),
			log.Int("totalRequested", len(names)), log.Int("toCreate", len(elementsToCreate)))

		if _, err := as.consentService.CreateConsentElements(ctx, ouID, elementsToCreate); err != nil {
			return as.wrapConsentServiceError(err)
		}
	}

	return nil
}

// attributesToPurposeElements converts a list of user attribute names to consent PurposeElements.
// For the consent purpose, we assume all user attributes are optional. The mandatory attributes are
// handled in the runtime when generating the consent form.
func attributesToPurposeElements(attributes map[string]bool) []consent.PurposeElement {
	elements := make([]consent.PurposeElement, 0, len(attributes))
	for attr := range attributes {
		elements = append(elements, consent.PurposeElement{
			Name:        attr,
			Namespace:   consent.NamespaceAttribute,
			IsMandatory: false,
		})
	}

	return elements
}
