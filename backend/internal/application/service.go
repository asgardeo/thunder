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

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/consent"
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	flowcommon "github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauthutils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
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
	appStore          applicationStoreInterface
	certService       cert.CertificateServiceInterface
	flowMgtService    flowmgt.FlowMgtServiceInterface
	themeMgtService   thememgt.ThemeMgtServiceInterface
	layoutMgtService  layoutmgt.LayoutMgtServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
	consentService    consent.ConsentServiceInterface
	transactioner     transaction.Transactioner
}

// newApplicationService creates a new instance of ApplicationService.
func newApplicationService(
	appStore applicationStoreInterface,
	certService cert.CertificateServiceInterface,
	flowMgtService flowmgt.FlowMgtServiceInterface,
	themeMgtService thememgt.ThemeMgtServiceInterface,
	layoutMgtService layoutmgt.LayoutMgtServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	consentService consent.ConsentServiceInterface,
	transactioner transaction.Transactioner,
) ApplicationServiceInterface {
	return &applicationService{
		appStore:          appStore,
		certService:       certService,
		flowMgtService:    flowMgtService,
		themeMgtService:   themeMgtService,
		layoutMgtService:  layoutMgtService,
		userSchemaService: userSchemaService,
		consentService:    consentService,
		transactioner:     transactioner,
	}
}

// CreateApplication creates the application.
func (as *applicationService) CreateApplication(ctx context.Context, app *model.ApplicationDTO) (*model.ApplicationDTO,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))
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

	// Create certificates, application, and consent purpose atomically within a transaction.
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

		createErr := as.appStore.CreateApplication(txCtx, *processedDTO)
		if createErr != nil {
			logger.Error("Failed to create application", log.Error(createErr), log.String("appID", appID))
			return createErr
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

	if err != nil {
		if innerSvcErr != nil {
			return nil, innerSvcErr
		}
		return nil, &ErrorInternalServerError
	}

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
		Metadata:                  processedDTO.Metadata,
	}
	if inboundAuthConfig != nil && len(processedDTO.InboundAuthConfig) > 0 {
		processedTokenConfig := processedDTO.InboundAuthConfig[0].OAuthAppConfig.Token

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
				Token:                   processedTokenConfig,
				Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
				UserInfo:                processedDTO.InboundAuthConfig[0].OAuthAppConfig.UserInfo,
				ScopeClaims:             processedDTO.InboundAuthConfig[0].OAuthAppConfig.ScopeClaims,
				Certificate:             returnOAuthCert,
			},
		}
		returnApp.InboundAuthConfig = []model.InboundAuthConfigDTO{returnInboundAuthConfig}
	}

	return returnApp, nil
}

// ValidateApplication validates the application data transfer object.
func (as *applicationService) ValidateApplication(ctx context.Context, app *model.ApplicationDTO) (
	*model.ApplicationProcessedDTO, *model.InboundAuthConfigDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if app == nil {
		return nil, nil, &ErrorApplicationNil
	}
	if app.Name == "" {
		return nil, nil, &ErrorInvalidApplicationName
	}
	existingApp, appCheckErr := as.appStore.GetApplicationByName(ctx, app.Name)
	if appCheckErr != nil && !errors.Is(appCheckErr, model.ApplicationNotFoundError) {
		logger.Error("Failed to check existing application by name", log.Error(appCheckErr),
			log.String("appName", app.Name))
		return nil, nil, &ErrorInternalServerError
	}
	if existingApp != nil {
		return nil, nil, &ErrorApplicationAlreadyExistsWithName
	}

	inboundAuthConfig, svcErr := as.processInboundAuthConfig(ctx, app, nil)
	if svcErr != nil {
		return nil, nil, svcErr
	}

	if svcErr := as.validateAuthFlowID(ctx, app); svcErr != nil {
		return nil, nil, svcErr
	}
	if svcErr := as.validateRegistrationFlowID(ctx, app); svcErr != nil {
		return nil, nil, svcErr
	}

	if svcErr := as.validateThemeID(app.ThemeID); svcErr != nil {
		return nil, nil, svcErr
	}
	if svcErr := as.validateLayoutID(app.LayoutID); svcErr != nil {
		return nil, nil, svcErr
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, nil, &ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, nil, &ErrorInvalidLogoURL
	}

	if svcErr := as.validateAllowedUserTypes(ctx, app.AllowedUserTypes); svcErr != nil {
		return nil, nil, svcErr
	}

	as.validateConsentConfig(app)

	appID := app.ID
	if appID == "" {
		var err error
		appID, err = sysutils.GenerateUUIDv7()
		if err != nil {
			logger.Error("Failed to generate UUID", log.Error(err))
			return nil, nil, &serviceerror.InternalServerError
		}
	}
	assertion, finalOAuthAccessToken, finalOAuthIDToken := processTokenConfiguration(app)
	userInfo := processUserInfoConfiguration(app, finalOAuthIDToken)
	scopeClaims := processScopeClaimsConfiguration(app)

	processedDTO := &model.ApplicationProcessedDTO{
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
	if inboundAuthConfig != nil {
		// Construct the return DTO with processed token configuration
		returnTokenConfig := &model.OAuthTokenConfig{
			AccessToken: finalOAuthAccessToken,
			IDToken:     finalOAuthIDToken,
		}

		processedInboundAuthConfig := model.InboundAuthConfigProcessedDTO{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				HashedClientSecret:      getProcessedClientSecret(inboundAuthConfig.OAuthAppConfig),
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				PKCERequired:            inboundAuthConfig.OAuthAppConfig.PKCERequired,
				PublicClient:            inboundAuthConfig.OAuthAppConfig.PublicClient,
				Token:                   returnTokenConfig,
				Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
				UserInfo:                userInfo,
				ScopeClaims:             scopeClaims,
			},
		}
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
	}
	return processedDTO, inboundAuthConfig, nil
}

func (as *applicationService) ValidateApplicationForUpdate(
	ctx context.Context, appID string, app *model.ApplicationDTO) (
	*model.ApplicationProcessedDTO, *model.InboundAuthConfigDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

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

	existingApp, appCheckErr := as.appStore.GetApplicationByID(ctx, appID)
	if appCheckErr != nil {
		if errors.Is(appCheckErr, model.ApplicationNotFoundError) {
			return nil, nil, &ErrorApplicationNotFound
		}
		logger.Debug("Failed to get existing application", log.Error(appCheckErr), log.String("appID", appID))
		return nil, nil, &ErrorInternalServerError
	}
	if existingApp == nil {
		logger.Debug("Application not found for update", log.String("appID", appID))
		return nil, nil, &ErrorApplicationNotFound
	}

	// If the application name is changed, check if an application with the new name already exists.
	if existingApp.Name != app.Name {
		existingAppWithName, appCheckErr := as.appStore.GetApplicationByName(ctx, app.Name)
		if appCheckErr != nil && !errors.Is(appCheckErr, model.ApplicationNotFoundError) {
			logger.Debug("Failed to check existing application by name", log.Error(appCheckErr),
				log.String("appName", app.Name))
			return nil, nil, &ErrorInternalServerError
		}
		if existingAppWithName != nil {
			return nil, nil, &ErrorApplicationAlreadyExistsWithName
		}
	}

	if svcErr := as.validateAuthFlowID(ctx, app); svcErr != nil {
		return nil, nil, svcErr
	}
	if svcErr := as.validateRegistrationFlowID(ctx, app); svcErr != nil {
		return nil, nil, svcErr
	}

	if svcErr := as.validateThemeID(app.ThemeID); svcErr != nil {
		return nil, nil, svcErr
	}
	if svcErr := as.validateLayoutID(app.LayoutID); svcErr != nil {
		return nil, nil, svcErr
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, nil, &ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, nil, &ErrorInvalidLogoURL
	}

	if svcErr := as.validateAllowedUserTypes(ctx, app.AllowedUserTypes); svcErr != nil {
		return nil, nil, svcErr
	}

	as.validateConsentConfig(app)

	inboundAuthConfig, svcErr := as.processInboundAuthConfig(ctx, app, existingApp)
	if svcErr != nil {
		return nil, nil, svcErr
	}

	return existingApp, inboundAuthConfig, nil
}

// GetApplicationList list the applications.
func (as *applicationService) GetApplicationList(
	ctx context.Context) (*model.ApplicationListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	totalCount, err := as.appStore.GetTotalApplicationCount(ctx)
	if err != nil {
		logger.Error("Failed to retrieve total application count", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	applications, err := as.appStore.GetApplicationList(ctx)
	if err != nil {
		// Check for composite limit exceeded
		if errors.Is(err, errResultLimitExceededInCompositeMode) {
			return nil, &ErrorResultLimitExceeded
		}
		logger.Error("Failed to retrieve application list", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	applicationList := make([]model.BasicApplicationResponse, 0, len(applications))
	for _, app := range applications {
		applicationList = append(applicationList, buildBasicApplicationResponse(app))
	}

	response := &model.ApplicationListResponse{
		TotalResults: totalCount,
		Count:        len(applications),
		Applications: applicationList,
	}

	return response, nil
}

// buildBasicApplicationResponse builds a basic application response from the processed application DTO.
func buildBasicApplicationResponse(app model.BasicApplicationDTO) model.BasicApplicationResponse {
	return model.BasicApplicationResponse{
		ID:                        app.ID,
		Name:                      app.Name,
		Description:               app.Description,
		ClientID:                  app.ClientID,
		LogoURL:                   app.LogoURL,
		AuthFlowID:                app.AuthFlowID,
		RegistrationFlowID:        app.RegistrationFlowID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		ThemeID:                   app.ThemeID,
		LayoutID:                  app.LayoutID,
		Template:                  app.Template,
		IsReadOnly:                app.IsReadOnly,
	}
}

// GetOAuthApplication retrieves the OAuth application based on the client id.
func (as *applicationService) GetOAuthApplication(
	ctx context.Context, clientID string) (*model.OAuthAppConfigProcessedDTO, *serviceerror.ServiceError) {
	if clientID == "" {
		return nil, &ErrorInvalidClientID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	oauthApp, err := as.appStore.GetOAuthApplication(ctx, clientID)
	if err != nil {
		if errors.Is(err, model.ApplicationNotFoundError) {
			return nil, &ErrorApplicationNotFound
		}

		logger.Error("Failed to retrieve OAuth application", log.Error(err),
			log.String("clientID", log.MaskString(clientID)))
		return nil, &ErrorInternalServerError
	}
	if oauthApp == nil {
		return nil, &ErrorApplicationNotFound
	}

	certificate, certErr := as.getApplicationCertificate(ctx, clientID, cert.CertificateReferenceTypeOAuthApp)
	if certErr != nil {
		return nil, certErr
	}

	oauthApp.Certificate = certificate

	return oauthApp, nil
}

// GetApplication get the application for given app id.
func (as *applicationService) GetApplication(ctx context.Context, appID string) (*model.Application,
	*serviceerror.ServiceError) {
	if appID == "" {
		return nil, &ErrorInvalidApplicationID
	}

	applicationDTO, err := as.appStore.GetApplicationByID(ctx, appID)
	if err != nil {
		return nil, as.handleApplicationRetrievalError(err)
	}

	application := &model.Application{
		ID:                        applicationDTO.ID,
		Name:                      applicationDTO.Name,
		Description:               applicationDTO.Description,
		AuthFlowID:                applicationDTO.AuthFlowID,
		RegistrationFlowID:        applicationDTO.RegistrationFlowID,
		IsRegistrationFlowEnabled: applicationDTO.IsRegistrationFlowEnabled,
		ThemeID:                   applicationDTO.ThemeID,
		LayoutID:                  applicationDTO.LayoutID,
		Template:                  applicationDTO.Template,
		URL:                       applicationDTO.URL,
		LogoURL:                   applicationDTO.LogoURL,
		TosURI:                    applicationDTO.TosURI,
		PolicyURI:                 applicationDTO.PolicyURI,
		Assertion:                 applicationDTO.Assertion,
		Contacts:                  applicationDTO.Contacts,
		Certificate:               applicationDTO.Certificate,
		AllowedUserTypes:          applicationDTO.AllowedUserTypes,
		LoginConsent:              applicationDTO.LoginConsent,
		Metadata:                  applicationDTO.Metadata,
	}

	if len(applicationDTO.InboundAuthConfig) > 0 {
		inboundAuthConfigs := make([]model.InboundAuthConfigComplete, 0, len(applicationDTO.InboundAuthConfig))
		for _, inboundAuthConfigDTO := range applicationDTO.InboundAuthConfig {
			if inboundAuthConfigDTO.Type == model.OAuthInboundAuthType && inboundAuthConfigDTO.OAuthAppConfig != nil {
				oauthAppConfig := inboundAuthConfigDTO.OAuthAppConfig
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
	}

	return as.enrichApplicationWithCertificate(ctx, application)
}

// handleApplicationRetrievalError handles common error scenarios when retrieving applications from the
// application store. It maps specific errors, such as ApplicationNotFoundError, to corresponding service errors.
func (as *applicationService) handleApplicationRetrievalError(err error) *serviceerror.ServiceError {
	if errors.Is(err, model.ApplicationNotFoundError) {
		return &ErrorApplicationNotFound
	}
	return &ErrorInternalServerError
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

// UpdateApplication update the application for given app id.
func (as *applicationService) UpdateApplication(ctx context.Context, appID string, app *model.ApplicationDTO) (
	*model.ApplicationDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	existingApp, inboundAuthConfig, svcErr := as.ValidateApplicationForUpdate(ctx, appID, app)

	if svcErr != nil {
		return nil, svcErr
	}

	// Process token configuration
	assertion, finalOAuthAccessToken, finalOAuthIDToken := processTokenConfiguration(app)
	userInfo := processUserInfoConfiguration(app, finalOAuthIDToken)
	scopeClaims := processScopeClaimsConfiguration(app)

	processedDTO := as.buildProcessedDTOForUpdate(
		appID, app,
		inboundAuthConfig, existingApp,
		assertion, finalOAuthAccessToken, finalOAuthIDToken,
		userInfo, scopeClaims,
	)

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

		storeErr := as.appStore.UpdateApplication(txCtx, existingApp, processedDTO)
		if storeErr != nil {
			logger.Error("Failed to update application", log.Error(storeErr), log.String("appID", appID))
			return storeErr
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

	if err != nil {
		if innerSvcErr != nil {
			return nil, innerSvcErr
		}
		return nil, &ErrorInternalServerError
	}

	return as.buildReturnDTOForUpdate(
		appID, app,
		inboundAuthConfig, processedDTO,
		assertion, finalOAuthAccessToken, finalOAuthIDToken,
		userInfo, scopeClaims, returnCert, returnOAuthCert,
	), nil
}

// buildProcessedDTOForUpdate constructs the ApplicationProcessedDTO for an application update operation.
func (as *applicationService) buildProcessedDTOForUpdate(appID string, app *model.ApplicationDTO,
	inboundAuthConfig *model.InboundAuthConfigDTO, existingApp *model.ApplicationProcessedDTO,
	assertion *model.AssertionConfig, finalOAuthAccessToken *model.AccessTokenConfig,
	finalOAuthIDToken *model.IDTokenConfig, userInfo *model.UserInfoConfig,
	scopeClaims map[string][]string) *model.ApplicationProcessedDTO {
	processedDTO := &model.ApplicationProcessedDTO{
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

	if inboundAuthConfig != nil {
		// Wrap the finalOAuthAccessToken and finalOAuthIDToken in OAuthTokenConfig structure
		oAuthTokenConfig := &model.OAuthTokenConfig{
			AccessToken: finalOAuthAccessToken,
			IDToken:     finalOAuthIDToken,
		}

		var existingOAuthConfig *model.OAuthAppConfigProcessedDTO
		if len(existingApp.InboundAuthConfig) > 0 {
			existingOAuthConfig = existingApp.InboundAuthConfig[0].OAuthAppConfig
		}

		processedInboundAuthConfig := model.InboundAuthConfigProcessedDTO{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
				AppID:    appID,
				ClientID: inboundAuthConfig.OAuthAppConfig.ClientID,
				HashedClientSecret: getProcessedClientSecretForUpdate(
					inboundAuthConfig.OAuthAppConfig, existingOAuthConfig),
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				PKCERequired:            inboundAuthConfig.OAuthAppConfig.PKCERequired,
				PublicClient:            inboundAuthConfig.OAuthAppConfig.PublicClient,
				Token:                   oAuthTokenConfig,
				Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
				UserInfo:                userInfo,
				ScopeClaims:             scopeClaims,
				Certificate:             inboundAuthConfig.OAuthAppConfig.Certificate,
			},
		}
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
	}

	return processedDTO
}

// buildReturnDTOForUpdate constructs the ApplicationDTO to return from an application update operation.
func (as *applicationService) buildReturnDTOForUpdate(appID string, app *model.ApplicationDTO,
	inboundAuthConfig *model.InboundAuthConfigDTO, processedDTO *model.ApplicationProcessedDTO,
	assertion *model.AssertionConfig, finalOAuthAccessToken *model.AccessTokenConfig,
	finalOAuthIDToken *model.IDTokenConfig, userInfo *model.UserInfoConfig, scopeClaims map[string][]string,
	returnCert *model.ApplicationCertificate, returnOAuthCert *model.ApplicationCertificate) *model.ApplicationDTO {
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
		Metadata:                  processedDTO.Metadata,
	}
	if inboundAuthConfig != nil {
		// Construct the return DTO with processed token configuration
		returnTokenConfig := &model.OAuthTokenConfig{
			AccessToken: finalOAuthAccessToken,
			IDToken:     finalOAuthIDToken,
		}

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
				Token:                   returnTokenConfig,
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

// DeleteApplication delete the application for given app id.
func (as *applicationService) DeleteApplication(ctx context.Context, appID string) *serviceerror.ServiceError {
	if appID == "" {
		return &ErrorInvalidApplicationID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// Check if the application is declarative (read-only)
	if as.appStore.IsApplicationDeclarative(ctx, appID) {
		return &ErrorCannotModifyDeclarativeResource
	}

	var appNotFound bool
	var transactionSvcErr *serviceerror.ServiceError
	err := as.transactioner.Transact(ctx, func(txCtx context.Context) error {
		existingApp, fetchErr := as.appStore.GetApplicationByID(txCtx, appID)
		if fetchErr != nil {
			if errors.Is(fetchErr, model.ApplicationNotFoundError) {
				appNotFound = true
				return nil
			}
			logger.Error("Error while retrieving the application before deletion", log.Error(fetchErr),
				log.String("appID", appID))
			return fetchErr
		}

		appErr := as.appStore.DeleteApplication(txCtx, appID)
		if appErr != nil {
			if errors.Is(appErr, model.ApplicationNotFoundError) {
				logger.Debug("Application not found for the deletion", log.String("appID", appID))
				appNotFound = true
				return nil
			}
			logger.Error("Error while deleting the application", log.Error(appErr), log.String("appID", appID))
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

		for _, inboundConfig := range existingApp.InboundAuthConfig {
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

	if err != nil {
		if transactionSvcErr != nil {
			return transactionSvcErr
		}
		return &ErrorInternalServerError
	}

	return nil
}

// validateAuthFlowID validates the auth flow ID for the application.
// If the flow ID is not provided, it sets the default authentication flow ID.
func (as *applicationService) validateAuthFlowID(
	ctx context.Context, app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.AuthFlowID != "" {
		isValidFlow := as.flowMgtService.IsValidFlow(ctx, app.AuthFlowID)
		if !isValidFlow {
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
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if app.RegistrationFlowID != "" {
		isValidFlow := as.flowMgtService.IsValidFlow(ctx, app.RegistrationFlowID)
		if !isValidFlow {
			return &ErrorInvalidRegistrationFlowID
		}
	} else {
		// Try to get the equivalent registration flow for the auth flow
		authFlow, svcErr := as.flowMgtService.GetFlow(ctx, app.AuthFlowID)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ServerErrorType {
				logger.Error("Error while retrieving auth flow definition",
					log.String("flowID", app.AuthFlowID), log.Any("error", svcErr))
				return &serviceerror.InternalServerError
			}
			return &ErrorWhileRetrievingFlowDefinition
		}

		registrationFlow, svcErr := as.flowMgtService.GetFlowByHandle(
			ctx, authFlow.Handle, flowcommon.FlowTypeRegistration)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ServerErrorType {
				logger.Error("Error while retrieving registration flow definition by handle",
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

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// Get all user schemas to check if the provided user types exist
	existingUserTypes := make(map[string]bool)
	limit := serverconst.MaxPageSize
	offset := 0

	for {
		// Runtime context is used to avoid authorization checks when fetching user schemas.
		userSchemaList, svcErr := as.userSchemaService.GetUserSchemaList(
			security.WithRuntimeContext(ctx), limit, offset)
		if svcErr != nil {
			logger.Error("Failed to retrieve user schema list for validation",
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
		logger.Info("Invalid user types found", log.Any("invalidTypes", invalidUserTypes))
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

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if app.InboundAuthConfig[0].Type != model.OAuthInboundAuthType {
		return nil, &ErrorInvalidInboundAuthConfig
	}
	inboundAuthConfig := app.InboundAuthConfig[0]
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

	return &inboundAuthConfig, nil
}

// processInboundAuthConfig validates and processes inbound auth configuration for
// creating or updating an application.
func (as *applicationService) processInboundAuthConfig(ctx context.Context, app *model.ApplicationDTO,
	existingApp *model.ApplicationProcessedDTO) (
	*model.InboundAuthConfigDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))
	inboundAuthConfig, err := validateOAuthParamsForCreateAndUpdate(app)
	if err != nil {
		return nil, err
	}

	if inboundAuthConfig == nil {
		return nil, nil
	}

	clientID := inboundAuthConfig.OAuthAppConfig.ClientID

	// For update operation
	if existingApp != nil && len(existingApp.InboundAuthConfig) > 0 {
		existingClientID := existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID

		if clientID == "" {
			// Generate OAuth 2.0 compliant client ID with proper entropy and URL-safe format
			generatedClientID, err := oauthutils.GenerateOAuth2ClientID()
			if err != nil {
				logger.Error("Failed to generate OAuth client ID", log.Error(err))
				return nil, &ErrorInternalServerError
			}
			inboundAuthConfig.OAuthAppConfig.ClientID = generatedClientID
		} else if clientID != existingClientID {
			existingAppWithClientID, clientCheckErr := as.appStore.GetOAuthApplication(ctx, clientID)
			if clientCheckErr != nil && !errors.Is(clientCheckErr, model.ApplicationNotFoundError) {
				logger.Error("Failed to check existing application by client ID", log.Error(clientCheckErr),
					log.String("clientID", clientID))
				return nil, &ErrorInternalServerError
			}
			if existingAppWithClientID != nil {
				return nil, &ErrorApplicationAlreadyExistsWithClientID
			}
		}
	} else { // For create operation
		if clientID == "" {
			// Generate OAuth 2.0 compliant client ID with proper entropy and URL-safe format
			generatedClientID, err := oauthutils.GenerateOAuth2ClientID()
			if err != nil {
				logger.Error("Failed to generate OAuth client ID", log.Error(err))
				return nil, &ErrorInternalServerError
			}
			inboundAuthConfig.OAuthAppConfig.ClientID = generatedClientID
		} else {
			existingAppWithClientID, clientCheckErr := as.appStore.GetOAuthApplication(ctx, clientID)
			if clientCheckErr != nil && !errors.Is(clientCheckErr, model.ApplicationNotFoundError) {
				logger.Error("Failed to check existing application by client ID", log.Error(clientCheckErr),
					log.String("clientID", clientID))
				return nil, &ErrorInternalServerError
			}
			if existingAppWithClientID != nil {
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
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	defaultAuthFlowHandle := config.GetThunderRuntime().Config.Flow.DefaultAuthFlowHandle
	defaultAuthFlow, svcErr := as.flowMgtService.GetFlowByHandle(
		ctx, defaultAuthFlowHandle, flowcommon.FlowTypeAuthentication)

	if svcErr != nil {
		if svcErr.Type == serviceerror.ServerErrorType {
			logger.Error("Error while retrieving default auth flow definition by handle",
				log.String("flowHandle", defaultAuthFlowHandle), log.Any("error", svcErr))
			return "", &serviceerror.InternalServerError
		}
		return "", &ErrorWhileRetrievingFlowDefinition
	}

	return defaultAuthFlow.ID, nil
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
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

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
			logger.Error("Failed to create application certificate", log.Any("serviceError", svcErr))
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
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if certErr := as.certService.DeleteCertificateByReference(
		ctx, cert.CertificateReferenceTypeApplication, appID); certErr != nil {
		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to delete application certificate: " +
				certErr.ErrorDescription
			return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		logger.Error("Failed to delete application certificate", log.String("appID", appID),
			log.Any("serviceError", certErr))
		return &ErrorCertificateServerError
	}

	return nil
}

// deleteOAuthAppCertificate deletes the certificate associated with an OAuth app (by client ID).
func (as *applicationService) deleteOAuthAppCertificate(
	ctx context.Context, clientID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if certErr := as.certService.DeleteCertificateByReference(
		ctx, cert.CertificateReferenceTypeOAuthApp, clientID); certErr != nil {
		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to delete OAuth app certificate: " +
				certErr.ErrorDescription
			return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		logger.Error("Failed to delete OAuth app certificate", log.String("clientID", clientID),
			log.Any("serviceError", certErr))
		return &ErrorCertificateServerError
	}

	return nil
}

// getApplicationCertificate retrieves the certificate associated with the application based
// on the reference type (application or OAuth app).
func (as *applicationService) getApplicationCertificate(ctx context.Context, appID string,
	refType cert.CertificateReferenceType) (*model.ApplicationCertificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

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
		logger.Error("Failed to retrieve application certificate", log.Any("serviceError", certErr),
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
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	existingCert, certErr := as.certService.GetCertificateByReference(
		ctx, refType, appID)
	if certErr != nil && certErr.Code != cert.ErrorCertificateNotFound.Code {
		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to retrieve application certificate: " +
				certErr.ErrorDescription
			return nil, serviceerror.CustomServiceError(
				ErrorCertificateClientError, errorDescription)
		}
		logger.Error("Failed to retrieve application certificate", log.Any("serviceError", certErr),
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
				logger.Error("Failed to update application certificate", log.Any("serviceError", svcErr),
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
				logger.Error("Failed to create application certificate", log.Any("serviceError", svcErr),
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
				logger.Error("Failed to delete application certificate", log.Any("serviceError", deleteErr),
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
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken != nil {
		oauthAccessToken = &model.AccessTokenConfig{
			ValidityPeriod: app.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.ValidityPeriod,
			UserAttributes: app.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.UserAttributes,
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
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken != nil {
		oauthIDToken = &model.IDTokenConfig{
			ValidityPeriod: app.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ValidityPeriod,
			UserAttributes: app.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.UserAttributes,
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

	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.UserInfo != nil {
		userInfoConfigInput := app.InboundAuthConfig[0].OAuthAppConfig.UserInfo
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
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.ScopeClaims != nil {
		scopeClaims = app.InboundAuthConfig[0].OAuthAppConfig.ScopeClaims
	}
	if scopeClaims == nil {
		scopeClaims = make(map[string][]string)
	}

	return scopeClaims
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

// getProcessedClientSecret returns the hashed client secret for confidential clients, empty string for public clients.
func getProcessedClientSecret(oauthConfig *model.OAuthAppConfigDTO) string {
	if oauthConfig.TokenEndpointAuthMethod != oauth2const.TokenEndpointAuthMethodClientSecretBasic &&
		oauthConfig.TokenEndpointAuthMethod != oauth2const.TokenEndpointAuthMethodClientSecretPost {
		return ""
	}
	return hash.GenerateThumbprintFromString(oauthConfig.ClientSecret)
}

// getProcessedClientSecretForUpdate returns the hashed client secret for update operations.
// If a new secret is provided, it hashes it. Otherwise, it preserves the existing hashed secret.
func getProcessedClientSecretForUpdate(
	newOAuthConfig *model.OAuthAppConfigDTO,
	existingOAuthConfig *model.OAuthAppConfigProcessedDTO,
) string {
	// Public clients don't have secrets
	if newOAuthConfig.TokenEndpointAuthMethod != oauth2const.TokenEndpointAuthMethodClientSecretBasic &&
		newOAuthConfig.TokenEndpointAuthMethod != oauth2const.TokenEndpointAuthMethodClientSecretPost {
		return ""
	}

	// If a new secret is provided, hash it
	if newOAuthConfig.ClientSecret != "" {
		return hash.GenerateThumbprintFromString(newOAuthConfig.ClientSecret)
	}

	// For updates with no new secret, preserve existing hashed secret
	if existingOAuthConfig != nil && existingOAuthConfig.HashedClientSecret != "" {
		return existingOAuthConfig.HashedClientSecret
	}

	return ""
}

// resolveClientSecret generates a new client secret for confidential clients if needed.
// It preserves existing secrets during update operations unless explicitly provided.
func resolveClientSecret(
	inboundAuthConfig *model.InboundAuthConfigDTO,
	existingApp *model.ApplicationProcessedDTO,
) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// Only process confidential clients that use client_secret auth method and don't have a secret provided
	if (inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod !=
		oauth2const.TokenEndpointAuthMethodClientSecretBasic &&
		inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod !=
			oauth2const.TokenEndpointAuthMethodClientSecretPost) ||
		inboundAuthConfig.OAuthAppConfig.ClientSecret != "" {
		return nil
	}

	// Check if we should preserve existing confidential OAuth config secret
	shouldPreserveSecret := existingApp != nil &&
		len(existingApp.InboundAuthConfig) > 0 &&
		existingApp.InboundAuthConfig[0].OAuthAppConfig != nil &&
		existingApp.InboundAuthConfig[0].OAuthAppConfig.HashedClientSecret != "" &&
		!existingApp.InboundAuthConfig[0].OAuthAppConfig.PublicClient

	if shouldPreserveSecret {
		return nil
	}

	// Generate OAuth 2.0 compliant client secret with high entropy for security
	generatedClientSecret, err := oauthutils.GenerateOAuth2ClientSecret()
	if err != nil {
		logger.Error("Failed to generate OAuth client secret", log.Error(err))
		return &ErrorInternalServerError
	}

	inboundAuthConfig.OAuthAppConfig.ClientSecret = generatedClientSecret
	return nil
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
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// TODO: Replace with application's actual OU when OU support is added
	const ouID = "default"

	logger.Debug("Attempting to synchronize consent purpose for the newly created application",
		log.String("appID", appDTO.ID))

	attributesMap := extractRequestedAttributes(appDTO)

	// Skip consent purpose creation if there are no user attributes requested by the application
	if len(attributesMap) == 0 {
		logger.Debug("No user attributes requested by the application, skipping consent purpose creation",
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

	logger.Debug("Creating consent purpose for the newly created application", log.String("appID", appDTO.ID),
		log.Int("attributesCount", len(attributes)))

	purpose := consent.ConsentPurposeInput{
		Name:        appDTO.Name,
		Description: "Consent purpose for application " + appDTO.Name,
		GroupID:     appDTO.ID,
		Elements:    attributesToPurposeElements(attributesMap),
	}
	if _, err := as.consentService.CreateConsentPurpose(ctx, ouID, &purpose); err != nil {
		return wrapConsentServiceError(err)
	}

	return nil
}

// // syncConsentPurposeOnUpdate synchronizes the consent purpose when an application is updated.
// // It updates the existing consent purpose to match the updated application configuration or
// // deletes it if all attributes are removed.
func (as *applicationService) syncConsentPurposeOnUpdate(ctx context.Context,
	existingAppDTO, updatedAppDTO *model.ApplicationProcessedDTO) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// TODO: Replace with application's actual OU when OU support is added
	const ouID = "default"

	logger.Debug("Attempting to synchronize consent purpose for the updated application",
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
		logger.Debug("Ensuring consent elements exist for all requested attributes",
			log.String("appID", existingAppDTO.ID), log.Int("requiredAttributesCount", len(requiredAttributes)))

		if err := as.createMissingConsentElements(ctx, ouID, requiredAttributes); err != nil {
			return err
		}
	}

	// Retrieve the existing consent purposes for the application
	existingPurposes, err := as.consentService.ListConsentPurposes(ctx, ouID, existingAppDTO.ID)
	if err != nil {
		return wrapConsentServiceError(err)
	}

	// If there are no existing purposes handle separately
	if len(existingPurposes) == 0 {
		logger.Debug("No existing consent purpose found for the application", log.String("appID", existingAppDTO.ID))

		// If attributes exists in the updated payload, create a new consent purpose
		if len(newAttributes) > 0 {
			logger.Debug("Creating new consent purpose for the application", log.String("appID", existingAppDTO.ID))

			purpose := consent.ConsentPurposeInput{
				Name:        updatedAppDTO.Name,
				Description: "Consent purpose for application " + updatedAppDTO.Name,
				GroupID:     existingAppDTO.ID,
				Elements:    attributesToPurposeElements(newAttributes),
			}
			if _, err := as.consentService.CreateConsentPurpose(ctx, ouID, &purpose); err != nil {
				return wrapConsentServiceError(err)
			}
		}

		return nil
	}

	// If all attributes are removed, and a purpose exists, delete it
	if len(newAttributes) == 0 {
		logger.Debug("All user attributes removed from the application", log.String("appID", existingAppDTO.ID))
		if err := as.deleteConsentPurposes(ctx, existingAppDTO.ID); err != nil {
			return err
		}

		return nil
	}

	logger.Debug("Existing consent purpose found for the application, updating it with the specified attributes",
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
		return wrapConsentServiceError(err)
	}

	return nil
}

// deleteConsentPurposes removes all consent purposes associated with an application.
func (as *applicationService) deleteConsentPurposes(ctx context.Context, appID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// TODO: Replace with application's actual OU when OU support is added
	const ouID = "default"

	logger.Debug("Attempting to delete consent purposes for the application", log.String("appID", appID))

	purposes, err := as.consentService.ListConsentPurposes(ctx, ouID, appID)
	if err != nil {
		return wrapConsentServiceError(err)
	}

	// If there are no purposes, return early
	if len(purposes) == 0 {
		logger.Debug("No consent purposes found for the application", log.String("appID", appID))
		return nil
	}

	// We assume there is only one consent purpose per application
	logger.Debug("Deleting consent purpose for the application", log.String("appID", appID),
		log.Int("purposesCount", len(purposes)))
	if err := as.consentService.DeleteConsentPurpose(ctx, ouID, purposes[0].ID); err != nil {
		// TODO: Default consent service implementation doesn't allow deleting consent purposes with existing consents.
		//  We need to handle this case gracefully until the consent service supports force delete or cascade delete
		// for consent purposes.
		if err.Code == consent.ErrorDeletingConsentPurposeWithAssociatedRecords.Code {
			logger.Warn("Cannot delete consent purpose due to existing consents. Consent service doesn't support "+
				"deleting consent purposes with existing consents",
				log.String("appID", appID), log.Int("purposesCount", len(purposes)))
			return nil
		}

		return wrapConsentServiceError(err)
	}

	return nil
}

// createMissingConsentElements validates a list of consent element names and creates only the missing ones.
// nolint:unparam // ouID is always "default" in current usage but kept for future flexibility
func (as *applicationService) createMissingConsentElements(ctx context.Context,
	ouID string, names []string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if len(names) == 0 {
		logger.Debug("No consent elements to create", log.String("ouID", ouID))
		return nil
	}

	validNames, err := as.consentService.ValidateConsentElements(ctx, ouID, names)
	if err != nil {
		return wrapConsentServiceError(err)
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
		logger.Debug("Creating missing consent elements", log.String("ouID", ouID),
			log.Int("totalRequested", len(names)), log.Int("toCreate", len(elementsToCreate)))

		if _, err := as.consentService.CreateConsentElements(ctx, ouID, elementsToCreate); err != nil {
			return wrapConsentServiceError(err)
		}
	}

	return nil
}

// wrapConsentServiceError converts an I18nServiceError from the consent service into a ServiceError
// for the application service.
func wrapConsentServiceError(err *serviceerror.I18nServiceError) *serviceerror.ServiceError {
	if err == nil {
		return nil
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if err.Type == serviceerror.ClientErrorType {
		logger.Debug("Failed to sync consent purpose for the application changes", log.Any("error", err))
		return serviceerror.CustomServiceError(ErrorConsentSyncFailed,
			fmt.Sprintf(ErrorConsentSyncFailed.ErrorDescription+" : code - %s", err.Code))
	}

	logger.Error("Failed to sync consent purpose for the application changes", log.Any("error", err))
	return &ErrorInternalServerError
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
