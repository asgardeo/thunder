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
	"github.com/asgardeo/thunder/internal/flow/flowmgt"
	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// ApplicationServiceInterface defines the interface for the application service.
type ApplicationServiceInterface interface {
	GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessed, *serviceerror.ServiceError)
	CreateApplication(app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	GetApplicationList() (*model.ApplicationListResponse, *serviceerror.ServiceError)
	GetApplicationByID(appID string) (*model.ApplicationProcessedDTO, *serviceerror.ServiceError)
	GetApplicationByName(appName string) (*model.ApplicationProcessedDTO, *serviceerror.ServiceError)
	UpdateApplication(appID string, app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	DeleteApplication(appID string) *serviceerror.ServiceError
}

// ApplicationService is the default implementation of the ApplicationServiceInterface.
type ApplicationService struct {
	AppByIDCacheManager   cache.CacheManagerInterface[*model.ApplicationProcessedDTO]
	AppByNameCacheManager cache.CacheManagerInterface[*model.ApplicationProcessedDTO]
	OAuthAppCacheManager  cache.CacheManagerInterface[*model.OAuthAppConfigProcessed]
	AppStore              store.ApplicationStoreInterface
	CertService           cert.CertificateServiceInterface
}

// GetApplicationService creates a new instance of ApplicationService.
func GetApplicationService() ApplicationServiceInterface {
	return &ApplicationService{
		AppByIDCacheManager:   cache.GetCacheManager[*model.ApplicationProcessedDTO]("ApplicationByIDCache"),
		AppByNameCacheManager: cache.GetCacheManager[*model.ApplicationProcessedDTO]("ApplicationByNameCache"),
		OAuthAppCacheManager:  cache.GetCacheManager[*model.OAuthAppConfigProcessed]("OAuthAppCache"),
		AppStore:              store.NewApplicationStore(),
		CertService:           cert.NewCertificateService(),
	}
}

// GetOAuthApplication retrieves the OAuth application based on the client id.
func (as *ApplicationService) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessed,
	*serviceerror.ServiceError) {
	if clientID == "" {
		return nil, &constants.ErrorInvalidClientID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// Check if the OAuth application is cached.
	cacheKey := cache.CacheKey{
		Key: clientID,
	}
	cachedApp, ok := as.OAuthAppCacheManager.Get(cacheKey)
	if ok {
		return cachedApp, nil
	}

	oauthApp, err := as.AppStore.GetOAuthApplication(clientID)
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
	as.cacheOAuthApplication(oauthApp)

	return oauthApp, nil
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

	// Check if a application with the same name already exists
	existingApp, appCheckErr := as.AppStore.GetApplicationByName(app.Name)
	if appCheckErr != nil && !errors.Is(appCheckErr, constants.ApplicationNotFoundError) {
		logger.Error("Failed to check existing application by name", log.Error(appCheckErr),
			log.String("appName", app.Name))
		return nil, &constants.ErrorInternalServerError
	}
	if existingApp != nil {
		return nil, &constants.ErrorApplicationAlreadyExistsWithName
	}

	inboundAuthConfig, err := validateOAuthParamsForCreateAndUpdate(app)
	if err != nil {
		return nil, err
	}

	// Check if a application with the same client ID already exists
	clientID := inboundAuthConfig.OAuthAppConfig.ClientID
	if clientID != "" {
		existingAppWithClientID, clientCheckErr := as.AppStore.GetOAuthApplication(clientID)
		if clientCheckErr != nil && !errors.Is(clientCheckErr, constants.ApplicationNotFoundError) {
			logger.Error("Failed to check existing application by client ID", log.Error(clientCheckErr),
				log.String("clientID", clientID))
			return nil, &constants.ErrorInternalServerError
		}
		if existingAppWithClientID != nil {
			return nil, &constants.ErrorApplicationAlreadyExistsWithClientID
		}
	}

	if err := validateAuthFlowGraphID(app); err != nil {
		return nil, err
	}
	if err := validateRegistrationFlowGraphID(app); err != nil {
		return nil, err
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, &constants.ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, &constants.ErrorInvalidLogoURL
	}

	appID := sysutils.GenerateUUID()

	// Validate and prepare the certificate if provided.
	cert, err := as.getValidatedCertificateForCreate(appID, app)
	if err != nil {
		return nil, err
	}

	processedInboundAuthConfig := model.InboundAuthConfigProcessed{
		Type: constants.OAuthInboundAuthType,
		OAuthAppConfig: &model.OAuthAppConfigProcessed{
			AppID:                   appID,
			ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
			HashedClientSecret:      hash.HashString(inboundAuthConfig.OAuthAppConfig.ClientSecret),
			RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
			GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
			ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
			TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
		},
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
		InboundAuthConfig:         []model.InboundAuthConfigProcessed{processedInboundAuthConfig},
	}

	// Create the application certificate if provided.
	returnCert, err := as.createApplicationCertificate(cert)
	if err != nil {
		return nil, err
	}

	// Create the application.
	storeErr := as.AppStore.CreateApplication(*processedDTO)
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
		Certificate:               returnCert,
		InboundAuthConfig:         []model.InboundAuthConfig{*inboundAuthConfig},
	}

	as.cacheApplication(processedDTO)

	return returnApp, nil
}

// GetApplicationList list the applications.
func (as *ApplicationService) GetApplicationList() (*model.ApplicationListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	totalCount, err := as.AppStore.GetTotalApplicationCount()
	if err != nil {
		logger.Error("Failed to retrieve total application count", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	applications, err := as.AppStore.GetApplicationList()
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

// GetApplicationByID get the application for given app id.
func (as *ApplicationService) GetApplicationByID(appID string) (*model.ApplicationProcessedDTO,
	*serviceerror.ServiceError) {
	if appID == "" {
		return nil, &constants.ErrorInvalidApplicationID
	}

	// Check if the application is cached.
	cacheKey := cache.CacheKey{
		Key: appID,
	}
	cachedApp, ok := as.AppByIDCacheManager.Get(cacheKey)
	if ok {
		return as.enrichApplicationWithCertificate(cachedApp)
	}

	application, err := as.AppStore.GetApplicationByID(appID)
	if err != nil {
		return nil, as.handleApplicationRetrievalError(err)
	}
	as.cacheApplication(application)

	return as.enrichApplicationWithCertificate(application)
}

// GetApplicationByName retrieves the application by its name.
func (as *ApplicationService) GetApplicationByName(appName string) (*model.ApplicationProcessedDTO,
	*serviceerror.ServiceError) {
	if appName == "" {
		return nil, &constants.ErrorInvalidApplicationName
	}

	// Check if the application is cached by name.
	cacheKey := cache.CacheKey{
		Key: appName,
	}
	cachedApp, ok := as.AppByNameCacheManager.Get(cacheKey)
	if ok {
		return as.enrichApplicationWithCertificate(cachedApp)
	}

	application, err := as.AppStore.GetApplicationByName(appName)
	if err != nil {
		return nil, as.handleApplicationRetrievalError(err)
	}
	as.cacheApplication(application)

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

	// Check if the application exists.
	// Check if the application is cached by ID.
	cacheKey := cache.CacheKey{
		Key: appID,
	}
	existingApp, ok := as.AppByIDCacheManager.Get(cacheKey)

	// If the application is not cached, retrieve it from the store.
	if !ok {
		var appCheckErr error
		existingApp, appCheckErr = as.AppStore.GetApplicationByID(appID)
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
	}

	// If the application name is changed, check if a application with the new name already exists.
	if existingApp.Name != app.Name {
		existingAppWithName, appCheckErr := as.AppStore.GetApplicationByName(app.Name)
		if appCheckErr != nil && !errors.Is(appCheckErr, constants.ApplicationNotFoundError) {
			logger.Error("Failed to check existing application by name", log.Error(appCheckErr),
				log.String("appName", app.Name))
			return nil, &constants.ErrorInternalServerError
		}
		if existingAppWithName != nil {
			return nil, &constants.ErrorApplicationAlreadyExistsWithName
		}
	}

	inboundAuthConfig, svcErr := validateOAuthParamsForCreateAndUpdate(app)
	if svcErr != nil {
		return nil, svcErr
	}

	// If the client ID is changed, check if a application with the new client ID already exists.
	newClientID := inboundAuthConfig.OAuthAppConfig.ClientID
	existingClientID := existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
	if newClientID != existingClientID {
		existingAppWithClientID, clientCheckErr := as.AppStore.GetOAuthApplication(newClientID)
		if clientCheckErr != nil && !errors.Is(clientCheckErr, constants.ApplicationNotFoundError) {
			logger.Error("Failed to check existing application by client ID", log.Error(clientCheckErr),
				log.String("clientID", newClientID))
			return nil, &constants.ErrorInternalServerError
		}
		if existingAppWithClientID != nil {
			return nil, &constants.ErrorApplicationAlreadyExistsWithClientID
		}
	}

	if svcErr := validateAuthFlowGraphID(app); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := validateRegistrationFlowGraphID(app); svcErr != nil {
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

	processedInboundAuthConfig := model.InboundAuthConfigProcessed{
		Type: constants.OAuthInboundAuthType,
		OAuthAppConfig: &model.OAuthAppConfigProcessed{
			AppID:                   appID,
			ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
			HashedClientSecret:      hash.HashString(inboundAuthConfig.OAuthAppConfig.ClientSecret),
			RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
			GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
			ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
			TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
		},
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
		InboundAuthConfig:         []model.InboundAuthConfigProcessed{processedInboundAuthConfig},
	}

	storeErr := as.AppStore.UpdateApplication(processedDTO)
	if storeErr != nil {
		logger.Error("Failed to update application", log.Error(storeErr), log.String("appID", appID))

		rollbackErr := as.rollbackApplicationCertificateUpdate(appID, existingCert, updatedCert)
		if rollbackErr != nil {
			return nil, rollbackErr
		}

		return nil, &constants.ErrorInternalServerError
	}

	// Invalidate the application caches.
	existingAppClientID := ""
	if len(existingApp.InboundAuthConfig) > 0 && existingApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		existingAppClientID = existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}
	as.invalidateApplicationCache(appID, existingApp.Name, existingAppClientID)

	// Cache the updated application.
	as.cacheApplication(processedDTO)
	as.cacheOAuthApplication(processedInboundAuthConfig.OAuthAppConfig)

	returnApp := &model.ApplicationDTO{
		ID:                        appID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Certificate:               returnCert,
		InboundAuthConfig:         []model.InboundAuthConfig{*inboundAuthConfig},
	}

	return returnApp, nil
}

// DeleteApplication delete the application for given app id.
func (as *ApplicationService) DeleteApplication(appID string) *serviceerror.ServiceError {
	if appID == "" {
		return &constants.ErrorInvalidApplicationID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// Get the application to find app details for cache invalidation.
	cacheKey := cache.CacheKey{
		Key: appID,
	}
	existingApp, ok := as.AppByIDCacheManager.Get(cacheKey)
	if !ok {
		var appErr error
		existingApp, appErr = as.AppStore.GetApplicationByID(appID)
		if appErr != nil {
			if errors.Is(appErr, constants.ApplicationNotFoundError) {
				logger.Debug("Application not found before deletion", log.String("appID", appID))
				return nil
			}
			logger.Error("Failed to get application before deletion", log.Error(appErr), log.String("appID", appID))
			return &constants.ErrorInternalServerError
		}
	}
	if existingApp == nil {
		logger.Debug("Application not found before deletion", log.String("appID", appID))
		return nil
	}

	// Delete the application from the store
	err := as.AppStore.DeleteApplication(appID)
	if err != nil {
		return &constants.ErrorInternalServerError
	}

	// Delete the application certificate
	svcErr := as.deleteApplicationCertificate(appID)
	if svcErr != nil {
		return svcErr
	}

	// Invalidate the application caches.
	clientID := ""
	if len(existingApp.InboundAuthConfig) > 0 && existingApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		clientID = existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}
	as.invalidateApplicationCache(appID, existingApp.Name, clientID)

	return nil
}

// validateAuthFlowGraphID validates the auth flow graph ID for the application.
// If the graph ID is not provided, it sets the default authentication flow graph ID.
func validateAuthFlowGraphID(app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.AuthFlowGraphID != "" {
		isValidFlowGraphID := flowmgt.GetFlowMgtService().IsValidGraphID(app.AuthFlowGraphID)
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
func validateRegistrationFlowGraphID(app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.RegistrationFlowGraphID != "" {
		isValidFlowGraphID := flowmgt.GetFlowMgtService().IsValidGraphID(app.RegistrationFlowGraphID)
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
func validateOAuthParamsForCreateAndUpdate(app *model.ApplicationDTO) (*model.InboundAuthConfig,
	*serviceerror.ServiceError) {
	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(app.InboundAuthConfig) == 0 || app.InboundAuthConfig[0].Type != constants.OAuthInboundAuthType {
		return nil, &constants.ErrorInvalidInboundAuthConfig
	}
	inboundAuthConfig := app.InboundAuthConfig[0]
	if inboundAuthConfig.OAuthAppConfig == nil {
		return nil, &constants.ErrorInvalidInboundAuthConfig
	}

	oauthAppConfig := inboundAuthConfig.OAuthAppConfig
	if oauthAppConfig.ClientID == "" {
		return nil, &constants.ErrorInvalidClientID
	}
	if oauthAppConfig.ClientSecret == "" {
		return nil, &constants.ErrorInvalidClientSecret
	}
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
		_, svcErr := as.CertService.CreateCertificate(cert)
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

	deleteErr := as.CertService.DeleteCertificateByReference(certconst.CertificateReferenceTypeApplication, appID)
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

	if certErr := as.CertService.DeleteCertificateByReference(
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

	cert, certErr := as.CertService.GetCertificateByReference(
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

	existingCert, certErr := as.CertService.GetCertificateByReference(
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
			_, svcErr := as.CertService.UpdateCertificateByID(existingCert.ID, updatedCert)
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
			_, svcErr := as.CertService.CreateCertificate(updatedCert)
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
			deleteErr := as.CertService.DeleteCertificateByReference(
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
			_, svcErr := as.CertService.UpdateCertificateByID(existingCert.ID, existingCert)
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
			deleteErr := as.CertService.DeleteCertificateByReference(
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
			_, svcErr := as.CertService.CreateCertificate(existingCert)
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

// cacheApplication caches the application and OAuth application configuration if it exists.
func (as *ApplicationService) cacheApplication(app *model.ApplicationProcessedDTO) {
	if app == nil {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if app.ID != "" {
		appByIDCacheKey := cache.CacheKey{
			Key: app.ID,
		}
		if err := as.AppByIDCacheManager.Set(appByIDCacheKey, app); err != nil {
			logger.Error("Failed to cache application by ID", log.String("appID", app.ID), log.Error(err))
		}
	}
	if app.Name != "" {
		appByNameCacheKey := cache.CacheKey{
			Key: app.Name,
		}
		if err := as.AppByNameCacheManager.Set(appByNameCacheKey, app); err != nil {
			logger.Error("Failed to cache application by name", log.String("appName", app.Name),
				log.Error(err))
		}
	}

	// Cache the OAuth application configuration if it exists.
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil {
		as.cacheOAuthApplication(app.InboundAuthConfig[0].OAuthAppConfig)
	}
}

// cacheOAuthApplication caches the OAuth application configuration if it exists.
func (as *ApplicationService) cacheOAuthApplication(oAuthAppConfig *model.OAuthAppConfigProcessed) {
	if oAuthAppConfig == nil || oAuthAppConfig.ClientID == "" {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	oauthCacheKey := cache.CacheKey{
		Key: oAuthAppConfig.ClientID,
	}
	if err := as.OAuthAppCacheManager.Set(oauthCacheKey, oAuthAppConfig); err != nil {
		logger.Error("Failed to cache OAuth application", log.String("clientID", oAuthAppConfig.ClientID),
			log.Error(err))
	}
}

// InvalidateApplicationCache invalidates all application caches.
func (as *ApplicationService) invalidateApplicationCache(appID, appName, clientID string) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if appID != "" {
		cacheKey := cache.CacheKey{
			Key: appID,
		}
		err := as.AppByIDCacheManager.Delete(cacheKey)
		if err != nil {
			logger.Error("Failed to delete application cache by ID", log.String("appID", appID), log.Error(err))
		}
	}
	if appName != "" {
		cacheKey := cache.CacheKey{
			Key: appName,
		}
		err := as.AppByNameCacheManager.Delete(cacheKey)
		if err != nil {
			logger.Error("Failed to delete application cache by name", log.String("appName", appName), log.Error(err))
		}
	}
	if clientID != "" {
		oauthCacheKey := cache.CacheKey{
			Key: clientID,
		}
		err := as.OAuthAppCacheManager.Delete(oauthCacheKey)
		if err != nil {
			logger.Error("Failed to delete OAuth application cache", log.String("clientID", clientID), log.Error(err))
		}
	}
}
