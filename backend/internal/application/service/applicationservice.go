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

// Package service provides application-related business logic and operations.
package service

import (
	"errors"
	"strings"

	"github.com/asgardeo/thunder/internal/application/constants"
	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/application/store"
	"github.com/asgardeo/thunder/internal/cert"
	certconst "github.com/asgardeo/thunder/internal/cert/constants"
	certmodel "github.com/asgardeo/thunder/internal/cert/model"
	"github.com/asgardeo/thunder/internal/flowmgt"
	oauthutils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// ApplicationServiceInterface defines the interface for the application service.
type ApplicationServiceInterface interface {
	CreateApplication(app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	GetApplicationList() (*model.ApplicationListResponse, *serviceerror.ServiceError)
	GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO, *serviceerror.ServiceError)
	GetApplication(appID string) (*model.ApplicationProcessedDTO, *serviceerror.ServiceError)
	UpdateApplication(appID string, app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	DeleteApplication(appID string) *serviceerror.ServiceError
}

// ApplicationService is the default implementation of the ApplicationServiceInterface.
type ApplicationService struct {
	flowMgtService flowmgt.FlowMgtServiceInterface
	appStore       store.ApplicationStoreInterface
	certService    cert.CertificateServiceInterface
}

// GetApplicationService creates a new instance of ApplicationService.
func GetApplicationService() ApplicationServiceInterface {
	flowMgtService, err := flowmgt.GetFlowMgtService()
	if err != nil {
		panic("Failed to initialize Flow Management Service: " + err.Error())
	}
	return &ApplicationService{
		flowMgtService: flowMgtService,
		appStore:       store.NewCachedBackedApplicationStore(),
		certService:    cert.NewCertificateService(),
	}
}

func NewApplicationService(flowMgtService flowmgt.FlowMgtServiceInterface, certService cert.CertificateServiceInterface) ApplicationServiceInterface {
	return &ApplicationService{
		flowMgtService: flowMgtService,
		appStore:       store.NewCachedBackedApplicationStore(),
		certService:    certService,
	}
}

// CreateApplication creates the application.
func (as *ApplicationService) CreateApplication(app *model.ApplicationDTO) (*model.ApplicationDTO,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if app == nil {
		return nil, &constants.ErrorApplicationNil
	}
	if app.Name == "" {
		return nil, &constants.ErrorInvalidApplicationName
	}

	// Check if an application with the same name already exists
	existingApp, appCheckErr := as.appStore.GetApplicationByName(app.Name)
	if appCheckErr != nil && !errors.Is(appCheckErr, constants.ApplicationNotFoundError) {
		logger.Error("Failed to check existing application by name", log.Error(appCheckErr),
			log.String("appName", app.Name))
		return nil, &constants.ErrorInternalServerError
	}
	if existingApp != nil {
		return nil, &constants.ErrorApplicationAlreadyExistsWithName
	}

	inboundAuthConfig, svcErr := validateAndProcessInboundAuthConfig(as.appStore, app, nil, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	if svcErr := as.validateAuthFlowGraphID(app); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := as.validateRegistrationFlowGraphID(app); svcErr != nil {
		return nil, svcErr
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, &constants.ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, &constants.ErrorInvalidLogoURL
	}

	appID := sysutils.GenerateUUID()

	// Process token configuration
	rootToken, finalOAuthToken := processTokenConfiguration(app)

	// Validate and prepare the certificate if provided.
	cert, svcErr := as.getValidatedCertificateForCreate(appID, app)
	if svcErr != nil {
		return nil, svcErr
	}

	processedDTO := &model.ApplicationProcessedDTO{
		ID:                        appID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Token:                     rootToken,
	}
	if inboundAuthConfig != nil {
		// Wrap the finalOAuthToken in OAuthTokenConfig structure
		var oAuthTokenConfig *model.OAuthTokenConfig
		if finalOAuthToken != nil {
			oAuthTokenConfig = &model.OAuthTokenConfig{
				AccessToken: finalOAuthToken,
			}
		}

		processedInboundAuthConfig := model.InboundAuthConfigProcessedDTO{
			Type: constants.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				HashedClientSecret:      hash.GenerateThumbprintFromString(inboundAuthConfig.OAuthAppConfig.ClientSecret),
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				Token:                   oAuthTokenConfig,
			},
		}
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
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

		return nil, &constants.ErrorInternalServerError
	}

	returnApp := &model.ApplicationDTO{
		ID:                        appID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Token:                     rootToken,
		Certificate:               returnCert,
	}
	if inboundAuthConfig != nil {
		// Construct the return DTO with processed token configuration
		var returnTokenConfig *model.OAuthTokenConfig
		if finalOAuthToken != nil {
			returnTokenConfig = &model.OAuthTokenConfig{
				AccessToken: finalOAuthToken,
			}
		}

		returnInboundAuthConfig := model.InboundAuthConfigDTO{
			Type: constants.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				ClientSecret:            inboundAuthConfig.OAuthAppConfig.ClientSecret,
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				Token:                   returnTokenConfig,
			},
		}
		returnApp.InboundAuthConfig = []model.InboundAuthConfigDTO{returnInboundAuthConfig}
	}

	return returnApp, nil
}

// GetApplicationList list the applications.
func (as *ApplicationService) GetApplicationList() (*model.ApplicationListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	totalCount, err := as.appStore.GetTotalApplicationCount()
	if err != nil {
		logger.Error("Failed to retrieve total application count", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	applications, err := as.appStore.GetApplicationList()
	if err != nil {
		logger.Error("Failed to retrieve application list", log.Error(err))
		return nil, &constants.ErrorInternalServerError
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
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
	}
}

// GetOAuthApplication retrieves the OAuth application based on the client id.
func (as *ApplicationService) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO,
	*serviceerror.ServiceError) {
	if clientID == "" {
		return nil, &constants.ErrorInvalidClientID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	oauthApp, err := as.appStore.GetOAuthApplication(clientID)
	if err != nil {
		if errors.Is(err, constants.ApplicationNotFoundError) {
			return nil, &constants.ErrorApplicationNotFound
		}

		logger.Error("Failed to retrieve OAuth application", log.Error(err),
			log.String("clientID", log.MaskString(clientID)))
		return nil, &constants.ErrorInternalServerError
	}
	if oauthApp == nil {
		return nil, &constants.ErrorApplicationNotFound
	}

	return oauthApp, nil
}

// GetApplication get the application for given app id.
func (as *ApplicationService) GetApplication(appID string) (*model.ApplicationProcessedDTO,
	*serviceerror.ServiceError) {
	if appID == "" {
		return nil, &constants.ErrorInvalidApplicationID
	}

	application, err := as.appStore.GetApplicationByID(appID)
	if err != nil {
		return nil, as.handleApplicationRetrievalError(err)
	}

	return as.enrichApplicationWithCertificate(application)
}

// handleApplicationRetrievalError handles common error scenarios when retrieving applications from the store.
func (as *ApplicationService) handleApplicationRetrievalError(err error) *serviceerror.ServiceError {
	if errors.Is(err, constants.ApplicationNotFoundError) {
		return &constants.ErrorApplicationNotFound
	}
	return &constants.ErrorInternalServerError
}

// enrichApplicationWithCertificate retrieves and adds the certificate to the application.
func (as *ApplicationService) enrichApplicationWithCertificate(application *model.ApplicationProcessedDTO) (
	*model.ApplicationProcessedDTO, *serviceerror.ServiceError) {
	cert, certErr := as.getApplicationCertificate(application.ID)
	if certErr != nil {
		return nil, certErr
	}
	application.Certificate = cert

	return application, nil
}

// UpdateApplication update the application for given app id.
func (as *ApplicationService) UpdateApplication(appID string, app *model.ApplicationDTO) (
	*model.ApplicationDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if appID == "" {
		return nil, &constants.ErrorInvalidApplicationID
	}
	if app == nil {
		return nil, &constants.ErrorApplicationNil
	}
	if app.Name == "" {
		return nil, &constants.ErrorInvalidApplicationName
	}

	existingApp, appCheckErr := as.appStore.GetApplicationByID(appID)
	if appCheckErr != nil {
		if errors.Is(appCheckErr, constants.ApplicationNotFoundError) {
			return nil, &constants.ErrorApplicationNotFound
		}
		logger.Error("Failed to get existing application", log.Error(appCheckErr), log.String("appID", appID))
		return nil, &constants.ErrorInternalServerError
	}
	if existingApp == nil {
		logger.Debug("Application not found for update", log.String("appID", appID))
		return nil, &constants.ErrorApplicationNotFound
	}

	// If the application name is changed, check if an application with the new name already exists.
	if existingApp.Name != app.Name {
		existingAppWithName, appCheckErr := as.appStore.GetApplicationByName(app.Name)
		if appCheckErr != nil && !errors.Is(appCheckErr, constants.ApplicationNotFoundError) {
			logger.Error("Failed to check existing application by name", log.Error(appCheckErr),
				log.String("appName", app.Name))
			return nil, &constants.ErrorInternalServerError
		}
		if existingAppWithName != nil {
			return nil, &constants.ErrorApplicationAlreadyExistsWithName
		}
	}

	inboundAuthConfig, svcErr := validateAndProcessInboundAuthConfig(as.appStore, app, existingApp, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	if svcErr := as.validateAuthFlowGraphID(app); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := as.validateRegistrationFlowGraphID(app); svcErr != nil {
		return nil, svcErr
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, &constants.ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, &constants.ErrorInvalidLogoURL
	}

	existingCert, updatedCert, returnCert, svcErr := as.updateApplicationCertificate(app)
	if svcErr != nil {
		return nil, svcErr
	}

	// Process token configuration
	rootToken, finalOAuthToken := processTokenConfiguration(app)

	processedDTO := &model.ApplicationProcessedDTO{
		ID:                        appID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Token:                     rootToken,
	}
	if inboundAuthConfig != nil {
		// Wrap the finalOAuthToken in OAuthTokenConfig structure
		var oAuthTokenConfig *model.OAuthTokenConfig
		if finalOAuthToken != nil {
			oAuthTokenConfig = &model.OAuthTokenConfig{
				AccessToken: finalOAuthToken,
			}
		}

		processedInboundAuthConfig := model.InboundAuthConfigProcessedDTO{
			Type: constants.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				HashedClientSecret:      hash.GenerateThumbprintFromString(inboundAuthConfig.OAuthAppConfig.ClientSecret),
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				Token:                   oAuthTokenConfig,
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

		return nil, &constants.ErrorInternalServerError
	}

	returnApp := &model.ApplicationDTO{
		ID:                        appID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Token:                     rootToken,
		Certificate:               returnCert,
	}
	if inboundAuthConfig != nil {
		// Construct the return DTO with processed token configuration
		var returnTokenConfig *model.OAuthTokenConfig
		if finalOAuthToken != nil {
			returnTokenConfig = &model.OAuthTokenConfig{
				AccessToken: finalOAuthToken,
			}
		}

		returnInboundAuthConfig := model.InboundAuthConfigDTO{
			Type: constants.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				ClientSecret:            inboundAuthConfig.OAuthAppConfig.ClientSecret,
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				Token:                   returnTokenConfig,
			},
		}
		returnApp.InboundAuthConfig = []model.InboundAuthConfigDTO{returnInboundAuthConfig}
	}

	return returnApp, nil
}

// DeleteApplication delete the application for given app id.
func (as *ApplicationService) DeleteApplication(appID string) *serviceerror.ServiceError {
	if appID == "" {
		return &constants.ErrorInvalidApplicationID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// Delete the application from the store
	appErr := as.appStore.DeleteApplication(appID)
	if appErr != nil {
		if errors.Is(appErr, constants.ApplicationNotFoundError) {
			logger.Debug("Application not found for the deletion", log.String("appID", appID))
			return nil
		}
		logger.Error("Error while deleting the application", log.Error(appErr), log.String("appID", appID))
		return &constants.ErrorInternalServerError
	}

	// Delete the application certificate
	svcErr := as.deleteApplicationCertificate(appID)
	if svcErr != nil {
		return svcErr
	}

	return nil
}

// validateAuthFlowGraphID validates the auth flow graph ID for the application.
// If the graph ID is not provided, it sets the default authentication flow graph ID.
func (as *ApplicationService) validateAuthFlowGraphID(app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.AuthFlowGraphID != "" {
		isValidFlowGraphID := as.flowMgtService.IsValidGraphID(app.AuthFlowGraphID)
		if !isValidFlowGraphID {
			return &constants.ErrorInvalidAuthFlowGraphID
		}
	} else {
		app.AuthFlowGraphID = getDefaultAuthFlowGraphID()
	}

	return nil
}

// validateRegistrationFlowGraphID validates the registration flow graph ID for the application.
// If the graph ID is not provided, it attempts to infer it from the auth flow graph ID.
func (as *ApplicationService) validateRegistrationFlowGraphID(app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.RegistrationFlowGraphID != "" {
		isValidFlowGraphID := as.flowMgtService.IsValidGraphID(app.RegistrationFlowGraphID)
		if !isValidFlowGraphID {
			return &constants.ErrorInvalidRegistrationFlowGraphID
		}
	} else {
		if strings.HasPrefix(app.AuthFlowGraphID, constants.AuthFlowGraphPrefix) {
			suffix := strings.TrimPrefix(app.AuthFlowGraphID, constants.AuthFlowGraphPrefix)
			app.RegistrationFlowGraphID = constants.RegistrationFlowGraphPrefix + suffix
		} else {
			return &constants.ErrorInvalidRegistrationFlowGraphID
		}
	}

	return nil
}

// validateOAuthParamsForCreateAndUpdate validates the OAuth parameters for creating or updating an application.
func validateOAuthParamsForCreateAndUpdate(app *model.ApplicationDTO) (*model.InboundAuthConfigDTO,
	*serviceerror.ServiceError) {
	// TODO: Validate the logic here whether it is okay to generate client id/ secret or set empty.
	if len(app.InboundAuthConfig) == 0 {
		return nil, nil
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if app.InboundAuthConfig[0].Type != constants.OAuthInboundAuthType {
		return nil, &constants.ErrorInvalidInboundAuthConfig
	}
	inboundAuthConfig := app.InboundAuthConfig[0]
	if inboundAuthConfig.OAuthAppConfig == nil {
		return nil, &constants.ErrorInvalidInboundAuthConfig
	}

	oauthAppConfig := inboundAuthConfig.OAuthAppConfig
	if len(oauthAppConfig.RedirectURIs) == 0 {
		return nil, &constants.ErrorInvalidRedirectURIs
	}

	// Validate the redirect URIs.
	for _, redirectURI := range oauthAppConfig.RedirectURIs {
		if !sysutils.IsValidURI(redirectURI) {
			return nil, &constants.ErrorInvalidRedirectURI
		}
	}

	// Validate the grant types.
	for _, grantType := range oauthAppConfig.GrantTypes {
		if !grantType.IsValid() {
			return nil, &constants.ErrorInvalidGrantType
		}
	}

	// Validate the response types.
	for _, responseType := range oauthAppConfig.ResponseTypes {
		if !responseType.IsValid() {
			return nil, &constants.ErrorInvalidResponseType
		}
	}

	// Validate the token endpoint authentication methods.
	for _, authMethod := range oauthAppConfig.TokenEndpointAuthMethod {
		if !authMethod.IsValid() {
			return nil, &constants.ErrorInvalidTokenEndpointAuthMethod
		}
	}

	return &inboundAuthConfig, nil
}

// validateAndProcessInboundAuthConfig validates and processes inbound auth configuration for
// creating or updating an application.
func validateAndProcessInboundAuthConfig(appStore store.ApplicationStoreInterface, app *model.ApplicationDTO,
	existingApp *model.ApplicationProcessedDTO, logger *log.Logger) (
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
	if existingApp != nil && len(existingApp.InboundAuthConfig) > 0 {
		existingClientID := existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID

		if clientID == "" {
			// Generate OAuth 2.0 compliant client ID with proper entropy and URL-safe format
			generatedClientID, err := oauthutils.GenerateOAuth2ClientID()
			if err != nil {
				logger.Error("Failed to generate OAuth client ID", log.Error(err))
				return nil, &constants.ErrorInternalServerError
			}
			inboundAuthConfig.OAuthAppConfig.ClientID = generatedClientID
		} else if clientID != existingClientID {
			existingAppWithClientID, clientCheckErr := appStore.GetOAuthApplication(clientID)
			if clientCheckErr != nil && !errors.Is(clientCheckErr, constants.ApplicationNotFoundError) {
				logger.Error("Failed to check existing application by client ID", log.Error(clientCheckErr),
					log.String("clientID", clientID))
				return nil, &constants.ErrorInternalServerError
			}
			if existingAppWithClientID != nil {
				return nil, &constants.ErrorApplicationAlreadyExistsWithClientID
			}
		}
	} else { // For create operation
		if clientID == "" {
			// Generate OAuth 2.0 compliant client ID with proper entropy and URL-safe format
			generatedClientID, err := oauthutils.GenerateOAuth2ClientID()
			if err != nil {
				logger.Error("Failed to generate OAuth client ID", log.Error(err))
				return nil, &constants.ErrorInternalServerError
			}
			inboundAuthConfig.OAuthAppConfig.ClientID = generatedClientID
		} else {
			existingAppWithClientID, clientCheckErr := appStore.GetOAuthApplication(clientID)
			if clientCheckErr != nil && !errors.Is(clientCheckErr, constants.ApplicationNotFoundError) {
				logger.Error("Failed to check existing application by client ID", log.Error(clientCheckErr),
					log.String("clientID", clientID))
				return nil, &constants.ErrorInternalServerError
			}
			if existingAppWithClientID != nil {
				return nil, &constants.ErrorApplicationAlreadyExistsWithClientID
			}
		}
	}

	// Generate OAuth 2.0 compliant client secret with high entropy for security
	if inboundAuthConfig.OAuthAppConfig.ClientSecret == "" {
		generatedClientSecret, err := oauthutils.GenerateOAuth2ClientSecret()
		if err != nil {
			logger.Error("Failed to generate OAuth client secret", log.Error(err))
			return nil, &constants.ErrorInternalServerError
		}
		inboundAuthConfig.OAuthAppConfig.ClientSecret = generatedClientSecret
	}

	return inboundAuthConfig, nil
}

// getDefaultAuthFlowGraphID returns the configured default authentication flow graph ID.
func getDefaultAuthFlowGraphID() string {
	authFlowConfig := config.GetThunderRuntime().Config.Flow.Authn
	return authFlowConfig.DefaultFlow
}

// getValidatedCertificateForCreate validates and returns the certificate for the application during creation.
func (as *ApplicationService) getValidatedCertificateForCreate(appID string, app *model.ApplicationDTO) (
	*certmodel.Certificate, *serviceerror.ServiceError) {
	if app.Certificate == nil || app.Certificate.Type == "" || app.Certificate.Type == certconst.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(appID, "", app)
}

// getValidatedCertificateForUpdate validates and returns the certificate for the application during update.
func (as *ApplicationService) getValidatedCertificateForUpdate(certID string, app *model.ApplicationDTO) (
	*certmodel.Certificate, *serviceerror.ServiceError) {
	if app.Certificate == nil || app.Certificate.Type == "" || app.Certificate.Type == certconst.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(app.ID, certID, app)
}

// getValidatedCertificateInput is a helper method that validates and returns the certificate.
func getValidatedCertificateInput(appID, certID string, app *model.ApplicationDTO) (*certmodel.Certificate,
	*serviceerror.ServiceError) {
	switch app.Certificate.Type {
	case certconst.CertificateTypeJWKS:
		if app.Certificate.Value == "" {
			return nil, &constants.ErrorInvalidCertificateValue
		}
		return &certmodel.Certificate{
			ID:      certID,
			RefType: certconst.CertificateReferenceTypeApplication,
			RefID:   appID,
			Type:    certconst.CertificateTypeJWKS,
			Value:   app.Certificate.Value,
		}, nil
	case certconst.CertificateTypeJWKSURI:
		if !sysutils.IsValidURI(app.Certificate.Value) {
			return nil, &constants.ErrorInvalidJWKSURI
		}
		return &certmodel.Certificate{
			ID:      certID,
			RefType: certconst.CertificateReferenceTypeApplication,
			RefID:   appID,
			Type:    certconst.CertificateTypeJWKSURI,
			Value:   app.Certificate.Value,
		}, nil
	default:
		return nil, &constants.ErrorInvalidCertificateType
	}
}

// createApplicationCertificate creates a certificate for the application.
func (as *ApplicationService) createApplicationCertificate(cert *certmodel.Certificate) (
	*model.ApplicationCertificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	var returnCert *model.ApplicationCertificate
	if cert != nil {
		_, svcErr := as.certService.CreateCertificate(cert)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ClientErrorType {
				retErr := serviceerror.ServiceError{
					Type:             constants.ErrorCertificateClientError.Type,
					Code:             constants.ErrorCertificateClientError.Code,
					Error:            constants.ErrorCertificateClientError.Error,
					ErrorDescription: "Failed to create application certificate: " + svcErr.ErrorDescription,
				}
				return nil, &retErr
			}
			logger.Error("Failed to create application certificate", log.Any("serviceError", svcErr))
			return nil, &constants.ErrorCertificateServerError
		}

		returnCert = &model.ApplicationCertificate{
			Type:  cert.Type,
			Value: cert.Value,
		}
	} else {
		returnCert = &model.ApplicationCertificate{
			Type:  certconst.CertificateTypeNone,
			Value: "",
		}
	}

	return returnCert, nil
}

// rollbackAppCertificateCreation rolls back the application certificate creation in case of an error during
// application creation.
func (as *ApplicationService) rollbackAppCertificateCreation(appID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	deleteErr := as.certService.DeleteCertificateByReference(certconst.CertificateReferenceTypeApplication, appID)
	if deleteErr != nil {
		if deleteErr.Type == serviceerror.ClientErrorType {
			retErr := serviceerror.ServiceError{
				Type:             constants.ErrorCertificateClientError.Type,
				Code:             constants.ErrorCertificateClientError.Code,
				Error:            constants.ErrorCertificateClientError.Error,
				ErrorDescription: "Failed to delete application certificate: " + deleteErr.ErrorDescription,
			}
			return &retErr
		}

		logger.Error("Failed to delete application certificate", log.String("appID", appID),
			log.Any("serviceError", deleteErr))
		return &constants.ErrorCertificateServerError
	}

	return nil
}

// deleteApplicationCertificate deletes the certificate associated with the application.
func (as *ApplicationService) deleteApplicationCertificate(appID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if certErr := as.certService.DeleteCertificateByReference(
		certconst.CertificateReferenceTypeApplication, appID); certErr != nil {
		if certErr.Type == serviceerror.ClientErrorType {
			retErr := serviceerror.ServiceError{
				Type:             constants.ErrorCertificateClientError.Type,
				Code:             constants.ErrorCertificateClientError.Code,
				Error:            constants.ErrorCertificateClientError.Error,
				ErrorDescription: "Failed to delete application certificate: " + certErr.ErrorDescription,
			}
			return &retErr
		}
		logger.Error("Failed to delete application certificate", log.String("appID", appID),
			log.Any("serviceError", certErr))
		return &constants.ErrorCertificateServerError
	}

	return nil
}

// getApplicationCertificate retrieves the certificate associated with the application.
func (as *ApplicationService) getApplicationCertificate(appID string) (*model.ApplicationCertificate,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	cert, certErr := as.certService.GetCertificateByReference(
		certconst.CertificateReferenceTypeApplication, appID)

	if certErr != nil {
		if certErr.Code == certconst.ErrorCertificateNotFound.Code {
			return &model.ApplicationCertificate{
				Type:  certconst.CertificateTypeNone,
				Value: "",
			}, nil
		}

		if certErr.Type == serviceerror.ClientErrorType {
			retErr := serviceerror.ServiceError{
				Type:             constants.ErrorCertificateClientError.Type,
				Code:             constants.ErrorCertificateClientError.Code,
				Error:            constants.ErrorCertificateClientError.Error,
				ErrorDescription: "Failed to retrieve application certificate: " + certErr.ErrorDescription,
			}
			return nil, &retErr
		}
		logger.Error("Failed to retrieve application certificate", log.Any("serviceError", certErr),
			log.String("appID", appID))
		return nil, &constants.ErrorCertificateServerError
	}

	if cert == nil {
		return &model.ApplicationCertificate{
			Type:  certconst.CertificateTypeNone,
			Value: "",
		}, nil
	}

	return &model.ApplicationCertificate{
		Type:  cert.Type,
		Value: cert.Value,
	}, nil
}

// updateApplicationCertificate updates the certificate for the application.
// It returns the existing certificate, the updated certificate, and the return application certificate details.
func (as *ApplicationService) updateApplicationCertificate(app *model.ApplicationDTO) (
	*certmodel.Certificate, *certmodel.Certificate, *model.ApplicationCertificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))
	appID := app.ID

	existingCert, certErr := as.certService.GetCertificateByReference(
		certconst.CertificateReferenceTypeApplication, appID)
	if certErr != nil && certErr.Code != certconst.ErrorCertificateNotFound.Code {
		if certErr.Type == serviceerror.ClientErrorType {
			retErr := serviceerror.ServiceError{
				Type:             constants.ErrorCertificateClientError.Type,
				Code:             constants.ErrorCertificateClientError.Code,
				Error:            constants.ErrorCertificateClientError.Error,
				ErrorDescription: "Failed to retrieve application certificate: " + certErr.ErrorDescription,
			}
			return nil, nil, nil, &retErr
		}
		logger.Error("Failed to retrieve application certificate", log.Any("serviceError", certErr),
			log.String("appID", appID))
		return nil, nil, nil, &constants.ErrorCertificateServerError
	}

	var updatedCert *certmodel.Certificate
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
			_, svcErr := as.certService.UpdateCertificateByID(existingCert.ID, updatedCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					retErr := serviceerror.ServiceError{
						Type:             constants.ErrorCertificateClientError.Type,
						Code:             constants.ErrorCertificateClientError.Code,
						Error:            constants.ErrorCertificateClientError.Error,
						ErrorDescription: "Failed to update application certificate: " + svcErr.ErrorDescription,
					}
					return nil, nil, nil, &retErr
				}
				logger.Error("Failed to update application certificate", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return nil, nil, nil, &constants.ErrorCertificateServerError
			}
		} else {
			_, svcErr := as.certService.CreateCertificate(updatedCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					retErr := serviceerror.ServiceError{
						Type:             constants.ErrorCertificateClientError.Type,
						Code:             constants.ErrorCertificateClientError.Code,
						Error:            constants.ErrorCertificateClientError.Error,
						ErrorDescription: "Failed to create application certificate: " + svcErr.ErrorDescription,
					}
					return nil, nil, nil, &retErr
				}
				logger.Error("Failed to create application certificate", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return nil, nil, nil, &constants.ErrorCertificateServerError
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
				certconst.CertificateReferenceTypeApplication, appID)
			if deleteErr != nil {
				if deleteErr.Type == serviceerror.ClientErrorType {
					retErr := serviceerror.ServiceError{
						Type:             constants.ErrorCertificateClientError.Type,
						Code:             constants.ErrorCertificateClientError.Code,
						Error:            constants.ErrorCertificateClientError.Error,
						ErrorDescription: "Failed to delete application certificate: " + deleteErr.ErrorDescription,
					}
					return nil, nil, nil, &retErr
				}
				logger.Error("Failed to delete application certificate", log.Any("serviceError", deleteErr),
					log.String("appID", appID))
				return nil, nil, nil, &constants.ErrorCertificateServerError
			}
		}

		returnCert = &model.ApplicationCertificate{
			Type:  certconst.CertificateTypeNone,
			Value: "",
		}
	}

	return existingCert, updatedCert, returnCert, nil
}

// rollbackApplicationCertificateUpdate rolls back the certificate update for the application in case of an error.
func (as *ApplicationService) rollbackApplicationCertificateUpdate(appID string,
	existingCert, updatedCert *certmodel.Certificate) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if updatedCert != nil {
		if existingCert != nil {
			// Update to the previously existed certificate.
			_, svcErr := as.certService.UpdateCertificateByID(existingCert.ID, existingCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					retErr := serviceerror.ServiceError{
						Type:  constants.ErrorCertificateClientError.Type,
						Code:  constants.ErrorCertificateClientError.Code,
						Error: constants.ErrorCertificateClientError.Error,
						ErrorDescription: "Failed to revert application certificate update: " +
							svcErr.ErrorDescription,
					}
					return &retErr
				}

				logger.Error("Failed to revert application certificate update", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return &constants.ErrorCertificateServerError
			}
		} else { // Delete the newly created certificate.
			deleteErr := as.certService.DeleteCertificateByReference(
				certconst.CertificateReferenceTypeApplication, appID)
			if deleteErr != nil {
				if deleteErr.Type == serviceerror.ClientErrorType {
					retErr := serviceerror.ServiceError{
						Type:  constants.ErrorCertificateClientError.Type,
						Code:  constants.ErrorCertificateClientError.Code,
						Error: constants.ErrorCertificateClientError.Error,
						ErrorDescription: "Failed to delete application certificate after update failure: " +
							deleteErr.ErrorDescription,
					}
					return &retErr
				}
				logger.Error("Failed to delete application certificate after update failure",
					log.Any("serviceError", deleteErr), log.String("appID", appID))
				return &constants.ErrorCertificateServerError
			}
		}
	} else {
		if existingCert != nil { // Create the previously existed certificate.
			_, svcErr := as.certService.CreateCertificate(existingCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					retErr := serviceerror.ServiceError{
						Type:  constants.ErrorCertificateClientError.Type,
						Code:  constants.ErrorCertificateClientError.Code,
						Error: constants.ErrorCertificateClientError.Error,
						ErrorDescription: "Failed to revert application certificate creation: " +
							svcErr.ErrorDescription,
					}
					return &retErr
				}

				logger.Error("Failed to revert application certificate creation", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return &constants.ErrorCertificateServerError
			}
		}
	}

	return nil
}

// getDefaultTokenConfigFromDeployment creates a default token configuration from deployment settings.
func getDefaultTokenConfigFromDeployment() *model.TokenConfig {
	jwtConfig := config.GetThunderRuntime().Config.OAuth.JWT
	tokenConfig := &model.TokenConfig{
		Issuer:         jwtConfig.Issuer,
		ValidityPeriod: jwtConfig.ValidityPeriod,
	}

	return tokenConfig
}

// processTokenConfiguration processes token configuration for an application, applying defaults where necessary.
func processTokenConfiguration(app *model.ApplicationDTO) (*model.TokenConfig, *model.TokenConfig) {
	// Resolve root token config
	var rootToken *model.TokenConfig
	if app.Token != nil {
		rootToken = &model.TokenConfig{
			Issuer:         app.Token.Issuer,
			ValidityPeriod: app.Token.ValidityPeriod,
			UserAttributes: app.Token.UserAttributes,
		}

		deploymentDefaults := getDefaultTokenConfigFromDeployment()
		if rootToken.Issuer == "" {
			rootToken.Issuer = deploymentDefaults.Issuer
		}
		if rootToken.ValidityPeriod == 0 {
			rootToken.ValidityPeriod = deploymentDefaults.ValidityPeriod
		}
	} else {
		rootToken = getDefaultTokenConfigFromDeployment()
	}
	if rootToken.UserAttributes == nil {
		rootToken.UserAttributes = make([]string, 0)
	}

	// Resolve OAuth token config
	var oauthAccessToken *model.TokenConfig
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken != nil {
		oauthAccessToken = app.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken
	}

	if oauthAccessToken != nil {
		if oauthAccessToken.Issuer == "" {
			oauthAccessToken.Issuer = rootToken.Issuer
		}
		if oauthAccessToken.ValidityPeriod == 0 {
			oauthAccessToken.ValidityPeriod = rootToken.ValidityPeriod
		}
		if oauthAccessToken.UserAttributes == nil {
			oauthAccessToken.UserAttributes = make([]string, 0)
		}
	} else {
		oauthAccessToken = &model.TokenConfig{
			Issuer:         rootToken.Issuer,
			ValidityPeriod: rootToken.ValidityPeriod,
			UserAttributes: rootToken.UserAttributes,
		}
	}

	return rootToken, oauthAccessToken
}
