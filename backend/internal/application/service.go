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
	"slices"
	"strings"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	flowcommon "github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauthutils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema"
)

// ApplicationServiceInterface defines the interface for the application service.
type ApplicationServiceInterface interface {
	CreateApplication(app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	ValidateApplication(app *model.ApplicationDTO) (
		*model.ApplicationProcessedDTO, *model.InboundAuthConfigDTO, *serviceerror.ServiceError)
	GetApplicationList() (*model.ApplicationListResponse, *serviceerror.ServiceError)
	GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO, *serviceerror.ServiceError)
	GetApplication(appID string) (*model.Application, *serviceerror.ServiceError)
	UpdateApplication(appID string, app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	DeleteApplication(appID string) *serviceerror.ServiceError
}

// ApplicationService is the default implementation of the ApplicationServiceInterface.
type applicationService struct {
	appStore          applicationStoreInterface
	certService       cert.CertificateServiceInterface
	flowMgtService    flowmgt.FlowMgtServiceInterface
	themeMgtService   thememgt.ThemeMgtServiceInterface
	layoutMgtService  layoutmgt.LayoutMgtServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
}

// newApplicationService creates a new instance of ApplicationService.
func newApplicationService(
	appStore applicationStoreInterface,
	certService cert.CertificateServiceInterface,
	flowMgtService flowmgt.FlowMgtServiceInterface,
	themeMgtService thememgt.ThemeMgtServiceInterface,
	layoutMgtService layoutmgt.LayoutMgtServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
) ApplicationServiceInterface {
	return &applicationService{
		appStore:          appStore,
		certService:       certService,
		flowMgtService:    flowMgtService,
		themeMgtService:   themeMgtService,
		layoutMgtService:  layoutMgtService,
		userSchemaService: userSchemaService,
	}
}

// CreateApplication creates the application.
func (as *applicationService) CreateApplication(app *model.ApplicationDTO) (*model.ApplicationDTO,
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
		if as.appStore.IsApplicationDeclarative(app.ID) {
			return nil, &ErrorCannotModifyDeclarativeResource
		}
	}

	processedDTO, inboundAuthConfig, svcErr := as.ValidateApplication(app)
	if svcErr != nil {
		return nil, svcErr
	}

	appID := processedDTO.ID
	assertion := processedDTO.Assertion

	// Validate and prepare the certificate if provided.
	cert, svcErr := as.getValidatedCertificateForCreate(appID, app)
	if svcErr != nil {
		return nil, svcErr
	}

	// Create the application certificate if provided.
	returnCert, svcErr := as.createApplicationCertificate(cert)
	if svcErr != nil {
		return nil, svcErr
	}

	// Create the application.
	storeErr := as.appStore.CreateApplication(*processedDTO)
	if storeErr != nil {
		logger.Error("Failed to create application", log.Error(storeErr), log.String("appID", appID))

		// Rollback the certificate creation if it was successful.
		if cert != nil {
			deleteErr := as.rollbackAppCertificateCreation(appID)
			if deleteErr != nil {
				return nil, deleteErr
			}
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
			},
		}
		returnApp.InboundAuthConfig = []model.InboundAuthConfigDTO{returnInboundAuthConfig}
	}

	return returnApp, nil
}

// ValidateApplication validates the application data transfer object.
func (as *applicationService) ValidateApplication(app *model.ApplicationDTO) (
	*model.ApplicationProcessedDTO, *model.InboundAuthConfigDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if app == nil {
		return nil, nil, &ErrorApplicationNil
	}
	if app.Name == "" {
		return nil, nil, &ErrorInvalidApplicationName
	}
	existingApp, appCheckErr := as.appStore.GetApplicationByName(app.Name)
	if appCheckErr != nil && !errors.Is(appCheckErr, model.ApplicationNotFoundError) {
		logger.Error("Failed to check existing application by name", log.Error(appCheckErr),
			log.String("appName", app.Name))
		return nil, nil, &ErrorInternalServerError
	}
	if existingApp != nil {
		return nil, nil, &ErrorApplicationAlreadyExistsWithName
	}

	inboundAuthConfig, svcErr := as.processInboundAuthConfig(app, nil)
	if svcErr != nil {
		return nil, nil, svcErr
	}

	if svcErr := as.validateAuthFlowID(app); svcErr != nil {
		return nil, nil, svcErr
	}
	if svcErr := as.validateRegistrationFlowID(app); svcErr != nil {
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

	if svcErr := as.validateAllowedUserTypes(app.AllowedUserTypes); svcErr != nil {
		return nil, nil, svcErr
	}

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

// GetApplicationList list the applications.
func (as *applicationService) GetApplicationList() (*model.ApplicationListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	totalCount, err := as.appStore.GetTotalApplicationCount()
	if err != nil {
		logger.Error("Failed to retrieve total application count", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	applications, err := as.appStore.GetApplicationList()
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
func (as *applicationService) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO,
	*serviceerror.ServiceError) {
	if clientID == "" {
		return nil, &ErrorInvalidClientID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	oauthApp, err := as.appStore.GetOAuthApplication(clientID)
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

	return oauthApp, nil
}

// GetApplication get the application for given app id.
func (as *applicationService) GetApplication(appID string) (*model.Application,
	*serviceerror.ServiceError) {
	if appID == "" {
		return nil, &ErrorInvalidApplicationID
	}

	applicationDTO, err := as.appStore.GetApplicationByID(appID)
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

	return as.enrichApplicationWithCertificate(application)
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
func (as *applicationService) enrichApplicationWithCertificate(application *model.Application) (
	*model.Application, *serviceerror.ServiceError) {
	cert, certErr := as.getApplicationCertificate(application.ID)
	if certErr != nil {
		return nil, certErr
	}
	application.Certificate = cert

	return application, nil
}

// UpdateApplication update the application for given app id.
func (as *applicationService) UpdateApplication(appID string, app *model.ApplicationDTO) (
	*model.ApplicationDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if appID == "" {
		return nil, &ErrorInvalidApplicationID
	}
	if app == nil {
		return nil, &ErrorApplicationNil
	}
	if app.Name == "" {
		return nil, &ErrorInvalidApplicationName
	}

	// Check if the application is declarative (read-only)
	if as.appStore.IsApplicationDeclarative(appID) {
		return nil, &ErrorCannotModifyDeclarativeResource
	}

	existingApp, appCheckErr := as.appStore.GetApplicationByID(appID)
	if appCheckErr != nil {
		if errors.Is(appCheckErr, model.ApplicationNotFoundError) {
			return nil, &ErrorApplicationNotFound
		}
		logger.Error("Failed to get existing application", log.Error(appCheckErr), log.String("appID", appID))
		return nil, &ErrorInternalServerError
	}
	if existingApp == nil {
		logger.Debug("Application not found for update", log.String("appID", appID))
		return nil, &ErrorApplicationNotFound
	}

	// If the application name is changed, check if an application with the new name already exists.
	if existingApp.Name != app.Name {
		existingAppWithName, appCheckErr := as.appStore.GetApplicationByName(app.Name)
		if appCheckErr != nil && !errors.Is(appCheckErr, model.ApplicationNotFoundError) {
			logger.Error("Failed to check existing application by name", log.Error(appCheckErr),
				log.String("appName", app.Name))
			return nil, &ErrorInternalServerError
		}
		if existingAppWithName != nil {
			return nil, &ErrorApplicationAlreadyExistsWithName
		}
	}

	inboundAuthConfig, svcErr := as.processInboundAuthConfig(app, existingApp)
	if svcErr != nil {
		return nil, svcErr
	}

	if svcErr := as.validateAuthFlowID(app); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := as.validateRegistrationFlowID(app); svcErr != nil {
		return nil, svcErr
	}

	if svcErr := as.validateThemeID(app.ThemeID); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := as.validateLayoutID(app.LayoutID); svcErr != nil {
		return nil, svcErr
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, &ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, &ErrorInvalidLogoURL
	}

	if svcErr := as.validateAllowedUserTypes(app.AllowedUserTypes); svcErr != nil {
		return nil, svcErr
	}

	existingCert, updatedCert, returnCert, svcErr := as.updateApplicationCertificate(app)
	if svcErr != nil {
		return nil, svcErr
	}

	// Process token configuration
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
			},
		}
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
	}

	storeErr := as.appStore.UpdateApplication(existingApp, processedDTO)
	if storeErr != nil {
		logger.Error("Failed to update application", log.Error(storeErr), log.String("appID", appID))

		rollbackErr := as.rollbackApplicationCertificateUpdate(appID, existingCert, updatedCert)
		if rollbackErr != nil {
			return nil, rollbackErr
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
			},
		}
		returnApp.InboundAuthConfig = []model.InboundAuthConfigDTO{returnInboundAuthConfig}
	}

	return returnApp, nil
}

// DeleteApplication delete the application for given app id.
func (as *applicationService) DeleteApplication(appID string) *serviceerror.ServiceError {
	if appID == "" {
		return &ErrorInvalidApplicationID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// Check if the application is declarative (read-only)
	if as.appStore.IsApplicationDeclarative(appID) {
		return &ErrorCannotModifyDeclarativeResource
	}

	// Delete the application from the store
	appErr := as.appStore.DeleteApplication(appID)
	if appErr != nil {
		if errors.Is(appErr, model.ApplicationNotFoundError) {
			logger.Debug("Application not found for the deletion", log.String("appID", appID))
			return nil
		}
		logger.Error("Error while deleting the application", log.Error(appErr), log.String("appID", appID))
		return &ErrorInternalServerError
	}

	// Delete the application certificate
	svcErr := as.deleteApplicationCertificate(appID)
	if svcErr != nil {
		return svcErr
	}

	return nil
}

// validateAuthFlowID validates the auth flow ID for the application.
// If the flow ID is not provided, it sets the default authentication flow ID.
func (as *applicationService) validateAuthFlowID(app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.AuthFlowID != "" {
		isValidFlow := as.flowMgtService.IsValidFlow(app.AuthFlowID)
		if !isValidFlow {
			return &ErrorInvalidAuthFlowID
		}
	} else {
		defaultFlowID, svcErr := as.getDefaultAuthFlowID()
		if svcErr != nil {
			return svcErr
		}
		app.AuthFlowID = defaultFlowID
	}

	return nil
}

// validateRegistrationFlowID validates the registration flow ID for the application.
// If the ID is not provided, it attempts to infer it from the equivalent auth flow ID.
func (as *applicationService) validateRegistrationFlowID(app *model.ApplicationDTO) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if app.RegistrationFlowID != "" {
		isValidFlow := as.flowMgtService.IsValidFlow(app.RegistrationFlowID)
		if !isValidFlow {
			return &ErrorInvalidRegistrationFlowID
		}
	} else {
		// Try to get the equivalent registration flow for the auth flow
		authFlow, svcErr := as.flowMgtService.GetFlow(app.AuthFlowID)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ServerErrorType {
				logger.Error("Error while retrieving auth flow definition",
					log.String("flowID", app.AuthFlowID), log.Any("error", svcErr))
				return &serviceerror.InternalServerError
			}
			return &ErrorWhileRetrievingFlowDefinition
		}

		registrationFlow, svcErr := as.flowMgtService.GetFlowByHandle(
			authFlow.Handle, flowcommon.FlowTypeRegistration)
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
func (as *applicationService) validateAllowedUserTypes(allowedUserTypes []string) *serviceerror.ServiceError {
	if len(allowedUserTypes) == 0 {
		return nil
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// Get all user schemas to check if the provided user types exist
	existingUserTypes := make(map[string]bool)
	limit := serverconst.MaxPageSize
	offset := 0

	for {
		// TODO: Pass context from the caller
		userSchemaList, svcErr := as.userSchemaService.GetUserSchemaList(context.TODO(), limit, offset)
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
func (as *applicationService) processInboundAuthConfig(app *model.ApplicationDTO,
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
			existingAppWithClientID, clientCheckErr := as.appStore.GetOAuthApplication(clientID)
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
			existingAppWithClientID, clientCheckErr := as.appStore.GetOAuthApplication(clientID)
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
func (as *applicationService) getDefaultAuthFlowID() (string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	defaultAuthFlowHandle := config.GetThunderRuntime().Config.Flow.DefaultAuthFlowHandle
	defaultAuthFlow, svcErr := as.flowMgtService.GetFlowByHandle(
		defaultAuthFlowHandle, flowcommon.FlowTypeAuthentication)

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
func (as *applicationService) getValidatedCertificateForCreate(appID string, app *model.ApplicationDTO) (
	*cert.Certificate, *serviceerror.ServiceError) {
	if app.Certificate == nil || app.Certificate.Type == "" || app.Certificate.Type == cert.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(appID, "", app)
}

// getValidatedCertificateForUpdate validates and returns the certificate for the application during update.
func (as *applicationService) getValidatedCertificateForUpdate(certID string, app *model.ApplicationDTO) (
	*cert.Certificate, *serviceerror.ServiceError) {
	if app.Certificate == nil || app.Certificate.Type == "" || app.Certificate.Type == cert.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(app.ID, certID, app)
}

// getValidatedCertificateInput is a helper method that validates and returns the certificate.
func getValidatedCertificateInput(appID, certID string, app *model.ApplicationDTO) (*cert.Certificate,
	*serviceerror.ServiceError) {
	switch app.Certificate.Type {
	case cert.CertificateTypeJWKS:
		if app.Certificate.Value == "" {
			return nil, &ErrorInvalidCertificateValue
		}
		return &cert.Certificate{
			ID:      certID,
			RefType: cert.CertificateReferenceTypeApplication,
			RefID:   appID,
			Type:    cert.CertificateTypeJWKS,
			Value:   app.Certificate.Value,
		}, nil
	case cert.CertificateTypeJWKSURI:
		if !sysutils.IsValidURI(app.Certificate.Value) {
			return nil, &ErrorInvalidJWKSURI
		}
		return &cert.Certificate{
			ID:      certID,
			RefType: cert.CertificateReferenceTypeApplication,
			RefID:   appID,
			Type:    cert.CertificateTypeJWKSURI,
			Value:   app.Certificate.Value,
		}, nil
	default:
		return nil, &ErrorInvalidCertificateType
	}
}

// createApplicationCertificate creates a certificate for the application.
func (as *applicationService) createApplicationCertificate(certificate *cert.Certificate) (
	*model.ApplicationCertificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	var returnCert *model.ApplicationCertificate
	if certificate != nil {
		_, svcErr := as.certService.CreateCertificate(context.TODO(), certificate)
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

// rollbackAppCertificateCreation rolls back the application certificate creation in case of an error during
// application creation.
func (as *applicationService) rollbackAppCertificateCreation(appID string) *serviceerror.ServiceError {
	deleteErr := as.certService.DeleteCertificateByReference(context.TODO(),
		cert.CertificateReferenceTypeApplication, appID)
	if deleteErr != nil {
		if deleteErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to rollback application certificate creation: " +
				deleteErr.ErrorDescription
			return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		return &ErrorCertificateServerError
	}

	return nil
}

// deleteApplicationCertificate deletes the certificate associated with the application.
func (as *applicationService) deleteApplicationCertificate(appID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if certErr := as.certService.DeleteCertificateByReference(
		context.TODO(), cert.CertificateReferenceTypeApplication, appID); certErr != nil {
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

// getApplicationCertificate retrieves the certificate associated with the application.
func (as *applicationService) getApplicationCertificate(appID string) (*model.ApplicationCertificate,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	certificate, certErr := as.certService.GetCertificateByReference(
		context.TODO(), cert.CertificateReferenceTypeApplication, appID)

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
// It returns the existing certificate, the updated certificate, and the return application certificate details.
func (as *applicationService) updateApplicationCertificate(app *model.ApplicationDTO) (
	*cert.Certificate, *cert.Certificate, *model.ApplicationCertificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))
	appID := app.ID

	existingCert, certErr := as.certService.GetCertificateByReference(
		context.TODO(), cert.CertificateReferenceTypeApplication, appID)
	if certErr != nil && certErr.Code != cert.ErrorCertificateNotFound.Code {
		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to retrieve application certificate: " +
				certErr.ErrorDescription
			return nil, nil, nil, serviceerror.CustomServiceError(
				ErrorCertificateClientError, errorDescription)
		}
		logger.Error("Failed to retrieve application certificate", log.Any("serviceError", certErr),
			log.String("appID", appID))
		return nil, nil, nil, &ErrorCertificateServerError
	}

	var updatedCert *cert.Certificate
	var err *serviceerror.ServiceError
	if existingCert != nil {
		updatedCert, err = as.getValidatedCertificateForUpdate(existingCert.ID, app)
	} else {
		updatedCert, err = as.getValidatedCertificateForUpdate("", app)
	}
	if err != nil {
		return nil, nil, nil, err
	}

	// Update the certificate if provided.
	var returnCert *model.ApplicationCertificate
	if updatedCert != nil {
		if existingCert != nil {
			_, svcErr := as.certService.UpdateCertificateByID(context.TODO(), existingCert.ID, updatedCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					errorDescription := "Failed to update application certificate: " +
						svcErr.ErrorDescription
					return nil, nil, nil, serviceerror.CustomServiceError(
						ErrorCertificateClientError, errorDescription)
				}
				logger.Error("Failed to update application certificate", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return nil, nil, nil, &ErrorCertificateServerError
			}
		} else {
			_, svcErr := as.certService.CreateCertificate(context.TODO(), updatedCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					errorDescription := "Failed to create application certificate: " +
						svcErr.ErrorDescription
					return nil, nil, nil, serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
				}
				logger.Error("Failed to create application certificate", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return nil, nil, nil, &ErrorCertificateServerError
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
				context.TODO(), cert.CertificateReferenceTypeApplication, appID)
			if deleteErr != nil {
				if deleteErr.Type == serviceerror.ClientErrorType {
					errorDescription := "Failed to delete application certificate: " + deleteErr.ErrorDescription
					return nil, nil, nil, serviceerror.CustomServiceError(
						ErrorCertificateClientError, errorDescription)
				}
				logger.Error("Failed to delete application certificate", log.Any("serviceError", deleteErr),
					log.String("appID", appID))
				return nil, nil, nil, &ErrorCertificateServerError
			}
		}

		returnCert = &model.ApplicationCertificate{
			Type:  cert.CertificateTypeNone,
			Value: "",
		}
	}

	return existingCert, updatedCert, returnCert, nil
}

// rollbackApplicationCertificateUpdate rolls back the certificate update for the application in case of an error.
func (as *applicationService) rollbackApplicationCertificateUpdate(appID string,
	existingCert, updatedCert *cert.Certificate) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if updatedCert != nil {
		if existingCert != nil {
			// Update to the previously existed certificate.
			_, svcErr := as.certService.UpdateCertificateByID(context.TODO(), existingCert.ID, existingCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					errorDescription := "Failed to revert application certificate update: " +
						svcErr.ErrorDescription
					return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
				}
				logger.Error("Failed to revert application certificate update", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return &ErrorCertificateServerError
			}
		} else { // Delete the newly created certificate.
			deleteErr := as.certService.DeleteCertificateByReference(
				context.TODO(), cert.CertificateReferenceTypeApplication, appID)
			if deleteErr != nil {
				if deleteErr.Type == serviceerror.ClientErrorType {
					errorDescription := "Failed to delete application certificate " +
						"after update failure: " + deleteErr.ErrorDescription
					return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
				}
				logger.Error("Failed to delete application certificate after update failure",
					log.Any("serviceError", deleteErr), log.String("appID", appID))
				return &ErrorCertificateServerError
			}
		}
	} else {
		if existingCert != nil { // Create the previously existed certificate.
			_, svcErr := as.certService.CreateCertificate(context.TODO(), existingCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					errorDescription := "Failed to revert application certificate creation: " +
						svcErr.ErrorDescription
					return serviceerror.CustomServiceError(ErrorCertificateClientError,
						errorDescription)
				}
				logger.Error("Failed to revert application certificate creation", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return &ErrorCertificateServerError
			}
		}
	}

	return nil
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

	if slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeClientCredentials) &&
		oauthConfig.TokenEndpointAuthMethod == oauth2const.TokenEndpointAuthMethodNone {
		return serviceerror.CustomServiceError(
			ErrorInvalidOAuthConfiguration,
			"client_credentials grant type cannot use 'none' authentication method",
		)
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

	if slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeClientCredentials) {
		return serviceerror.CustomServiceError(
			ErrorInvalidPublicClientConfiguration,
			"Public clients cannot use the client_credentials grant type",
		)
	}

	if oauthConfig.ClientSecret != "" {
		return serviceerror.CustomServiceError(
			ErrorInvalidPublicClientConfiguration,
			"Public clients cannot have client secrets",
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
	if oauthConfig.PublicClient {
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
	if newOAuthConfig.PublicClient {
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

	// Only process confidential clients that don't have a secret provided
	if inboundAuthConfig.OAuthAppConfig.ClientSecret != "" || inboundAuthConfig.OAuthAppConfig.PublicClient {
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
