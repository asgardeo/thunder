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
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/application/constants"
	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/application/store"
	"github.com/asgardeo/thunder/internal/cert"
	certconst "github.com/asgardeo/thunder/internal/cert/constants"
	certmodel "github.com/asgardeo/thunder/internal/cert/model"
	"github.com/asgardeo/thunder/internal/flow/graphservice"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// ApplicationServiceInterface defines the interface for the application service.
type ApplicationServiceInterface interface {
	GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessed, error)
	CreateApplication(app *model.ApplicationDTO) (*model.ApplicationDTO, error)
	GetApplicationList() ([]model.BasicApplicationDTO, error)
	GetApplication(appID string) (*model.ApplicationProcessedDTO, error)
	UpdateApplication(appID string, app *model.ApplicationDTO) (*model.ApplicationDTO, error)
	DeleteApplication(appID string) error
}

// ApplicationService is the default implementation of the ApplicationServiceInterface.
type ApplicationService struct {
	AppStore    store.ApplicationStoreInterface
	CertService cert.CertificateServiceInterface
}

// GetApplicationService creates a new instance of ApplicationService.
func GetApplicationService() ApplicationServiceInterface {
	return &ApplicationService{
		AppStore:    store.NewApplicationStore(),
		CertService: cert.NewCertificateService(),
	}
}

// GetOAuthApplication retrieves the OAuth application based on the client id.
func (as *ApplicationService) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessed, error) {
	if clientID == "" {
		return nil, errors.New("client ID cannot be empty")
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))
	oauthApp, err := as.AppStore.GetOAuthApplication(clientID)
	if err != nil {
		logger.Error("Failed to retrieve OAuth application", log.Error(err), log.String("clientID", clientID))
		return nil, err
	}

	return oauthApp, nil
}

// CreateApplication creates the application.
func (as *ApplicationService) CreateApplication(app *model.ApplicationDTO) (*model.ApplicationDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if app == nil {
		return nil, errors.New("application is nil")
	}
	if app.Name == "" {
		return nil, errors.New("application name cannot be empty")
	}

	inboundAuthConfig, err := validateOAuthParamsForCreateAndUpdate(app)
	if err != nil {
		return nil, err
	}

	if err := validateAuthFlowGraphID(app); err != nil {
		return nil, err
	}
	if err := validateRegistrationFlowGraphID(app); err != nil {
		return nil, err
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, errors.New("application URL is not a valid URI")
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, errors.New("application logo URL is not a valid URI")
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
	err = as.AppStore.CreateApplication(*processedDTO)
	if err != nil {
		logger.Error("Failed to create application", log.Error(err), log.String("appID", appID))

		// Rollback the certificate creation if it was successful.
		if cert != nil {
			deleteErr := as.rollbackAppCertificateCreation(appID)
			if deleteErr != nil {
				logger.Error("Failed to delete application certificate after application creation failure",
					log.Error(deleteErr), log.String("appID", appID))
			}
		}

		return nil, err
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

	return returnApp, nil
}

// GetApplicationList list the applications.
func (as *ApplicationService) GetApplicationList() ([]model.BasicApplicationDTO, error) {
	applications, err := as.AppStore.GetApplicationList()
	if err != nil {
		return nil, err
	}

	return applications, nil
}

// GetApplication get the application for given app id.
func (as *ApplicationService) GetApplication(appID string) (*model.ApplicationProcessedDTO, error) {
	if appID == "" {
		return nil, errors.New("application ID is empty")
	}

	application, err := as.AppStore.GetApplication(appID)
	if err != nil {
		return nil, err
	}

	cert, certErr := as.getApplicationCertificate(appID)
	if certErr != nil {
		return nil, certErr
	}
	application.Certificate = cert

	return &application, nil
}

// UpdateApplication update the application for given app id.
func (as *ApplicationService) UpdateApplication(appID string, app *model.ApplicationDTO) (
	*model.ApplicationDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if appID == "" {
		return nil, errors.New("application ID is empty")
	}
	if app == nil {
		return nil, errors.New("application is nil")
	}
	if app.Name == "" {
		return nil, errors.New("application name cannot be empty")
	}

	inboundAuthConfig, err := validateOAuthParamsForCreateAndUpdate(app)
	if err != nil {
		return nil, err
	}

	if err := validateAuthFlowGraphID(app); err != nil {
		return nil, err
	}
	if err := validateRegistrationFlowGraphID(app); err != nil {
		return nil, err
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, errors.New("application URL is not a valid URI")
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, errors.New("application logo URL is not a valid URI")
	}

	existingCert, updatedCert, returnCert, err := as.updateApplicationCertificate(app)
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

	err = as.AppStore.UpdateApplication(processedDTO)
	if err != nil {
		logger.Error("Failed to update application", log.Error(err), log.String("appID", appID))

		rollbackErr := as.rollbackApplicationCertificateUpdate(appID, existingCert, updatedCert)
		if rollbackErr != nil {
			return nil, fmt.Errorf("failed to rollback application certificate update: %w", rollbackErr)
		}

		return nil, err
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
	return returnApp, nil
}

// DeleteApplication delete the application for given app id.
func (as *ApplicationService) DeleteApplication(appID string) error {
	if appID == "" {
		return errors.New("application ID is empty")
	}

	err := as.AppStore.DeleteApplication(appID)
	if err != nil {
		return err
	}

	err = as.deleteApplicationCertificate(appID)
	if err != nil {
		return err
	}

	return nil
}

// validateAuthFlowGraphID validates the auth flow graph ID for the application.
// If the graph ID is not provided, it sets the default authentication flow graph ID.
func validateAuthFlowGraphID(app *model.ApplicationDTO) error {
	if app.AuthFlowGraphID != "" {
		isValidFlowGraphID := graphservice.GetGraphService().IsValidGraphID(app.AuthFlowGraphID)
		if !isValidFlowGraphID {
			return fmt.Errorf("invalid auth flow graph ID: %s", app.AuthFlowGraphID)
		}
	} else {
		app.AuthFlowGraphID = getDefaultAuthFlowGraphID()
	}

	return nil
}

// validateRegistrationFlowGraphID validates the registration flow graph ID for the application.
// If the graph ID is not provided, it attempts to infer it from the auth flow graph ID.
func validateRegistrationFlowGraphID(app *model.ApplicationDTO) error {
	if app.RegistrationFlowGraphID != "" {
		isValidFlowGraphID := graphservice.GetGraphService().IsValidGraphID(app.RegistrationFlowGraphID)
		if !isValidFlowGraphID {
			return fmt.Errorf("invalid registration flow graph ID: %s", app.RegistrationFlowGraphID)
		}
	} else {
		if strings.HasPrefix(app.AuthFlowGraphID, constants.AuthFlowGraphPrefix) {
			suffix := strings.TrimPrefix(app.AuthFlowGraphID, constants.AuthFlowGraphPrefix)
			app.RegistrationFlowGraphID = constants.RegistrationFlowGraphPrefix + suffix
		} else {
			return fmt.Errorf("cannot infer registration flow graph ID from auth flow graph ID: %s",
				app.AuthFlowGraphID)
		}
	}

	return nil
}

// validateOAuthParamsForCreateAndUpdate validates the OAuth parameters for creating or updating an application.
func validateOAuthParamsForCreateAndUpdate(app *model.ApplicationDTO) (*model.InboundAuthConfig, error) {
	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(app.InboundAuthConfig) == 0 || app.InboundAuthConfig[0].Type != constants.OAuthInboundAuthType {
		return nil, errors.New("invalid inbound authentication configuration")
	}
	inboundAuthConfig := app.InboundAuthConfig[0]
	if inboundAuthConfig.OAuthAppConfig == nil {
		return nil, errors.New("OAuth application configuration is nil")
	}

	oauthAppConfig := inboundAuthConfig.OAuthAppConfig
	if oauthAppConfig.ClientID == "" {
		return nil, errors.New("client ID cannot be empty")
	}
	if oauthAppConfig.ClientSecret == "" {
		return nil, errors.New("client secret cannot be empty")
	}
	if len(oauthAppConfig.RedirectURIs) == 0 {
		return nil, errors.New("at least one Redirect URI is required")
	}

	// Validate the redirect URIs.
	for _, redirectURI := range oauthAppConfig.RedirectURIs {
		if !sysutils.IsValidURI(redirectURI) {
			return nil, fmt.Errorf("redirect URI is not a valid URI: %s", redirectURI)
		}
	}

	// Validate the grant types.
	for _, grantType := range oauthAppConfig.GrantTypes {
		if !grantType.IsValid() {
			return nil, fmt.Errorf("invalid grant type: %s", grantType)
		}
	}

	// Validate the response types.
	for _, responseType := range oauthAppConfig.ResponseTypes {
		if !responseType.IsValid() {
			return nil, fmt.Errorf("invalid response type: %s", responseType)
		}
	}

	// Validate the token endpoint authentication methods.
	for _, authMethod := range oauthAppConfig.TokenEndpointAuthMethod {
		if !authMethod.IsValid() {
			return nil, fmt.Errorf("invalid token endpoint authentication method: %s", authMethod)
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
	*certmodel.Certificate, error) {
	if app.Certificate == nil || app.Certificate.Type == "" || app.Certificate.Type == certconst.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(appID, "", app)
}

// getValidatedCertificateForUpdate validates and returns the certificate for the application during update.
func (as *ApplicationService) getValidatedCertificateForUpdate(certID string, app *model.ApplicationDTO) (
	*certmodel.Certificate, error) {
	if app.Certificate == nil || app.Certificate.Type == "" || app.Certificate.Type == certconst.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(app.ID, certID, app)
}

// getValidatedCertificateInput is a helper method that validates and returns the certificate.
func getValidatedCertificateInput(appID, certID string, app *model.ApplicationDTO) (*certmodel.Certificate, error) {
	switch app.Certificate.Type {
	case certconst.CertificateTypeJWKS:
		if app.Certificate.Value == "" {
			return nil, errors.New("JWKS certificate value cannot be empty")
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
			return nil, errors.New("JWKS URI certificate value is not a valid URI")
		}
		return &certmodel.Certificate{
			ID:      certID,
			RefType: certconst.CertificateReferenceTypeApplication,
			RefID:   appID,
			Type:    certconst.CertificateTypeJWKSURI,
			Value:   app.Certificate.Value,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported certificate type: %s", app.Certificate.Type)
	}
}

// createApplicationCertificate creates a certificate for the application.
func (as *ApplicationService) createApplicationCertificate(cert *certmodel.Certificate) (
	*model.ApplicationCertificate, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	var returnCert *model.ApplicationCertificate
	if cert != nil {
		_, svcErr := as.CertService.CreateCertificate(cert)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ClientErrorType {
				return nil, fmt.Errorf("failed to create application certificate: %s", svcErr.ErrorDescription)
			}
			logger.Error("Failed to create application certificate", log.Any("serviceError", svcErr))
			return nil, fmt.Errorf("server error while creating application certificate")
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
func (as *ApplicationService) rollbackAppCertificateCreation(appID string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	deleteErr := as.CertService.DeleteCertificateByReference(certconst.CertificateReferenceTypeApplication, appID)
	if deleteErr != nil {
		if deleteErr.Type == serviceerror.ClientErrorType {
			return fmt.Errorf("failed to delete application certificate: %s", deleteErr.ErrorDescription)
		}
		logger.Error("Failed to delete application certificate", log.String("appID", appID),
			log.Any("serviceError", deleteErr))
		return errors.New("server error while deleting application certificate")
	}

	return nil
}

// deleteApplicationCertificate deletes the certificate associated with the application.
func (as *ApplicationService) deleteApplicationCertificate(appID string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if certErr := as.CertService.DeleteCertificateByReference(
		certconst.CertificateReferenceTypeApplication, appID); certErr != nil {
		logger.Error("Failed to delete application certificate", log.Any("serviceError", certErr),
			log.String("appID", appID))
		return fmt.Errorf("error while deleting application certificate: %s", certErr.ErrorDescription)
	}

	return nil
}

// getApplicationCertificate retrieves the certificate associated with the application.
func (as *ApplicationService) getApplicationCertificate(appID string) (*model.ApplicationCertificate, error) {
	cert, certErr := as.CertService.GetCertificateByReference(
		certconst.CertificateReferenceTypeApplication, appID)

	if certErr != nil {
		if certErr.Code == certconst.ErrorCertificateNotFound.Code {
			return &model.ApplicationCertificate{
				Type:  certconst.CertificateTypeNone,
				Value: "",
			}, nil
		}

		return nil, fmt.Errorf("failed to get application certificate: %s", certErr.ErrorDescription)
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
	*certmodel.Certificate, *certmodel.Certificate, *model.ApplicationCertificate, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))
	appID := app.ID

	existingCert, certErr := as.CertService.GetCertificateByReference(
		certconst.CertificateReferenceTypeApplication, appID)
	if certErr != nil && certErr.Code != certconst.ErrorCertificateNotFound.Code {
		if certErr.Type == serviceerror.ClientErrorType {
			return nil, nil, nil, fmt.Errorf("failed to get application certificate: %s", certErr.ErrorDescription)
		}
		logger.Error("Failed to retrieve application certificate", log.Any("serviceError", certErr),
			log.String("appID", appID))
		return nil, nil, nil, fmt.Errorf("server error while retrieving application certificate: %s",
			certErr.ErrorDescription)
	}

	var updatedCert *certmodel.Certificate
	var err error
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
					return nil, nil, nil, fmt.Errorf("failed to update application certificate: %s",
						svcErr.ErrorDescription)
				}
				logger.Error("Failed to update application certificate", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return nil, nil, nil, fmt.Errorf("server error while updating application certificate: %s",
					svcErr.ErrorDescription)
			}
		} else {
			_, svcErr := as.CertService.CreateCertificate(updatedCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					return nil, nil, nil, fmt.Errorf("failed to create application certificate: %s",
						svcErr.ErrorDescription)
				}
				logger.Error("Failed to create application certificate", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return nil, nil, nil, fmt.Errorf("server error while creating application certificate: %s",
					svcErr.ErrorDescription)
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
					return nil, nil, nil, fmt.Errorf("failed to delete application certificate: %s",
						deleteErr.ErrorDescription)
				}
				logger.Error("Failed to delete application certificate", log.Any("serviceError", deleteErr),
					log.String("appID", appID))
				return nil, nil, nil, fmt.Errorf("server error while deleting application certificate: %s",
					deleteErr.ErrorDescription)
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
	existingCert, updatedCert *certmodel.Certificate) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if updatedCert != nil {
		if existingCert != nil {
			// Update to the previously existed certificate.
			_, svcErr := as.CertService.UpdateCertificateByID(existingCert.ID, existingCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					return fmt.Errorf("failed to revert application certificate update: %s",
						svcErr.ErrorDescription)
				}
				logger.Error("Failed to revert application certificate update", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return fmt.Errorf("server error while reverting application certificate update: %s",
					svcErr.ErrorDescription)
			}
		} else {
			// Delete the newly created certificate.
			deleteErr := as.CertService.DeleteCertificateByReference(
				certconst.CertificateReferenceTypeApplication, appID)
			if deleteErr != nil {
				if deleteErr.Type == serviceerror.ClientErrorType {
					return fmt.Errorf("failed to delete application certificate: %s",
						deleteErr.ErrorDescription)
				}
				logger.Error("Failed to delete application certificate after update failure",
					log.Any("serviceError", deleteErr), log.String("appID", appID))
				return fmt.Errorf("server error while deleting application certificate: %s",
					deleteErr.ErrorDescription)
			}
		}
	} else {
		if existingCert != nil {
			// Create the previously existed certificate.
			_, svcErr := as.CertService.CreateCertificate(existingCert)
			if svcErr != nil {
				if svcErr.Type == serviceerror.ClientErrorType {
					return fmt.Errorf("failed to revert application certificate creation: %s",
						svcErr.ErrorDescription)
				}
				logger.Error("Failed to revert application certificate creation", log.Any("serviceError", svcErr),
					log.String("appID", appID))
				return fmt.Errorf("server error while reverting application certificate creation: %s",
					svcErr.ErrorDescription)
			}
		}
	}

	return nil
}
