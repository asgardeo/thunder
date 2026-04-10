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

package dcr

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauthutils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/transaction"
)

// DCRServiceInterface defines the interface for the DCR service.
type DCRServiceInterface interface {
	RegisterClient(
		ctx context.Context, request *DCRRegistrationRequest,
	) (*DCRRegistrationResponse, *serviceerror.ServiceError)
	GetClient(
		ctx context.Context, appID string,
	) (*DCRRegistrationResponse, *serviceerror.ServiceError)
}

// dcrService is the default implementation of DCRServiceInterface.
type dcrService struct {
	appService    application.ApplicationServiceInterface
	ouService     ou.OrganizationUnitServiceInterface
	i18nService   i18nmgt.I18nServiceInterface
	transactioner transaction.Transactioner
}

// newDCRService creates a new instance of dcrService.
func newDCRService(
	appService application.ApplicationServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	i18nService i18nmgt.I18nServiceInterface,
	transactioner transaction.Transactioner,
) DCRServiceInterface {
	return &dcrService{
		appService:    appService,
		ouService:     ouService,
		i18nService:   i18nService,
		transactioner: transactioner,
	}
}

// RegisterClient registers a new OAuth client using Dynamic Client Registration.
func (ds *dcrService) RegisterClient(ctx context.Context, request *DCRRegistrationRequest) (
	*DCRRegistrationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DCRService"))

	if request == nil {
		return nil, &ErrorInvalidRequestFormat
	}

	if request.JWKSUri != "" && len(request.JWKS) > 0 {
		return nil, &ErrorJWKSConfigurationConflict
	}

	// TODO: Revisit OU for DCR apps
	if request.OUID == "" {
		rootOUs, svcErr := ds.ouService.GetOrganizationUnitList(ctx, 1, 0)
		if svcErr != nil {
			logger.Error("Failed to retrieve root organization units for DCR", log.String("error", svcErr.Error))
			return nil, &ErrorServerError
		}
		if rootOUs == nil || rootOUs.TotalResults == 0 || len(rootOUs.OrganizationUnits) == 0 {
			logger.Error("No root organization unit available for DCR registration")
			return nil, &ErrorServerError
		}
		request.OUID = rootOUs.OrganizationUnits[0].ID
	}

	appDTO, svcErr := ds.convertDCRToApplication(request)
	if svcErr != nil {
		logger.Error("Failed to convert DCR request to application DTO", log.String("error", svcErr.Error))
		return nil, &ErrorServerError
	}

	var response *DCRRegistrationResponse
	var capturedErr *serviceerror.ServiceError
	var createdAppID string

	err := ds.transactioner.Transact(ctx, func(txCtx context.Context) error {
		createdApp, svcErr := ds.appService.CreateApplication(txCtx, appDTO)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ServerErrorType {
				logger.Error("Failed to create application via Application service",
					log.String("error_code", svcErr.Code))
				capturedErr = &ErrorServerError
				return errors.New("failed to create application")
			}
			logger.Debug("Failed to create application via Application service",
				log.String("error_code", svcErr.Code))
			capturedErr = ds.mapApplicationErrorToDCRError(svcErr)
			return errors.New("failed to create application")
		}

		createdAppID = createdApp.ID

		var convErr *serviceerror.ServiceError
		response, convErr = ds.convertApplicationToDCRResponse(createdApp, request.ClientName)
		if convErr != nil {
			logger.Error("Failed to convert application to DCR response", log.String("error", convErr.Error))
			capturedErr = convErr
			return errors.New("conversion failed")
		}

		return nil
	})

	if err != nil {
		if capturedErr != nil {
			return nil, capturedErr
		}
		return nil, &ErrorServerError
	}

	// Write localized variants after the runtime-DB transaction commits.
	// The i18n store uses a separate DB (configdb) and cannot participate in the
	// runtime transaction. If writing fails, clean up any partial i18n rows and
	// compensate by deleting the created app.
	if writeErr := ds.writeLocalizedVariants(createdAppID, request); writeErr != nil {
		logger.Error("Failed to write localized variants for DCR client; compensating by deleting app",
			log.String("appID", createdAppID), log.String("error", writeErr.Error))
		if ds.i18nService != nil {
			if cleanErr := ds.i18nService.DeleteTranslationsByNamespace("app." + createdAppID); cleanErr != nil {
				logger.Error("Failed to clean up partial i18n rows after write failure",
					log.String("appID", createdAppID))
			}
		}
		if delSvcErr := ds.appService.DeleteApplication(ctx, createdAppID); delSvcErr != nil {
			logger.Error("Compensation delete failed after i18n write failure",
				log.String("appID", createdAppID))
		}
		return nil, &ErrorServerError
	}

	// Echo registered localized variants back in the response (AC-03).
	response.LocalizedClientName = request.LocalizedClientName
	response.LocalizedLogoURI = request.LocalizedLogoURI
	response.LocalizedTosURI = request.LocalizedTosURI
	response.LocalizedPolicyURI = request.LocalizedPolicyURI

	return response, nil
}

// convertDCRToApplication converts DCR registration request to Application DTO.
func (ds *dcrService) convertDCRToApplication(request *DCRRegistrationRequest) (
	*model.ApplicationDTO, *serviceerror.ServiceError) {
	isPublicClient := request.TokenEndpointAuthMethod == oauth2const.TokenEndpointAuthMethodNone

	// Map JWKS/JWKS_URI to application-level certificate
	var appCertificate *model.ApplicationCertificate
	if request.JWKSUri != "" {
		appCertificate = &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKSURI,
			Value: request.JWKSUri,
		}
	} else if len(request.JWKS) > 0 {
		jwksBytes, err := json.Marshal(request.JWKS)
		if err == nil {
			appCertificate = &model.ApplicationCertificate{
				Type:  cert.CertificateTypeJWKS,
				Value: string(jwksBytes),
			}
		}
	}

	var scopes []string
	if request.Scope != "" {
		scopes = strings.Fields(request.Scope)
	}

	// Generate client ID if client_name is not provided and use it as both app name and client ID
	var clientID string
	appName := request.ClientName
	if appName == "" {
		generatedClientID, err := oauthutils.GenerateOAuth2ClientID()
		if err != nil {
			return nil, &ErrorServerError
		}
		clientID = generatedClientID
		appName = clientID
	}

	oauthAppConfig := &model.OAuthAppConfigDTO{
		ClientID:                clientID,
		RedirectURIs:            request.RedirectURIs,
		GrantTypes:              request.GrantTypes,
		ResponseTypes:           request.ResponseTypes,
		TokenEndpointAuthMethod: request.TokenEndpointAuthMethod,
		PublicClient:            isPublicClient,
		PKCERequired:            isPublicClient,
		Scopes:                  scopes,
	}

	inboundAuthConfig := []model.InboundAuthConfigDTO{
		{
			Type:           model.OAuthInboundAuthType,
			OAuthAppConfig: oauthAppConfig,
		},
	}

	appDTO := &model.ApplicationDTO{
		OUID:              request.OUID,
		Name:              appName,
		URL:               request.ClientURI,
		LogoURL:           request.LogoURI,
		InboundAuthConfig: inboundAuthConfig,
		TosURI:            request.TosURI,
		PolicyURI:         request.PolicyURI,
		Contacts:          request.Contacts,
		Certificate:       appCertificate,
	}

	return appDTO, nil
}

// convertApplicationToDCRResponse converts Application DTO to DCR registration response.
func (ds *dcrService) convertApplicationToDCRResponse(appDTO *model.ApplicationDTO, originalClientName string) (
	*DCRRegistrationResponse, *serviceerror.ServiceError) {
	if len(appDTO.InboundAuthConfig) == 0 || appDTO.InboundAuthConfig[0].OAuthAppConfig == nil {
		return &DCRRegistrationResponse{}, nil
	}

	oauthConfig := appDTO.InboundAuthConfig[0].OAuthAppConfig

	clientName := originalClientName
	if clientName == "" {
		clientName = oauthConfig.ClientID
	}

	var jwksURI string
	var jwks map[string]interface{}
	if appDTO.Certificate != nil {
		switch appDTO.Certificate.Type {
		case cert.CertificateTypeJWKSURI:
			jwksURI = appDTO.Certificate.Value
		case cert.CertificateTypeJWKS:
			if err := json.Unmarshal([]byte(appDTO.Certificate.Value), &jwks); err != nil {
				return nil, &ErrorServerError
			}
		}
	}

	scopeString := strings.Join(oauthConfig.Scopes, " ")

	response := &DCRRegistrationResponse{
		ClientID:                oauthConfig.ClientID,
		ClientSecret:            oauthConfig.ClientSecret,
		ClientSecretExpiresAt:   ClientSecretExpiresAtNever,
		RedirectURIs:            oauthConfig.RedirectURIs,
		GrantTypes:              oauthConfig.GrantTypes,
		ResponseTypes:           oauthConfig.ResponseTypes,
		ClientName:              clientName,
		ClientURI:               appDTO.URL,
		LogoURI:                 appDTO.LogoURL,
		TokenEndpointAuthMethod: oauthConfig.TokenEndpointAuthMethod,
		JWKSUri:                 jwksURI,
		JWKS:                    jwks,
		Scope:                   scopeString,
		TosURI:                  appDTO.TosURI,
		PolicyURI:               appDTO.PolicyURI,
		Contacts:                appDTO.Contacts,
		AppID:                   oauthConfig.AppID,
	}

	return response, nil
}

// writeLocalizedVariants persists all localized variants from a DCR request to the i18n table.
// Called after the application record has been committed to the runtime DB.
func (ds *dcrService) writeLocalizedVariants(
	appID string, request *DCRRegistrationRequest) *serviceerror.ServiceError {
	if ds.i18nService == nil {
		return nil
	}
	ns := "app." + appID
	type fieldSpec struct {
		variants map[string]string
		key      string
	}
	fields := []fieldSpec{
		{request.LocalizedClientName, "name"},
		{request.LocalizedLogoURI, "logo_uri"},
		{request.LocalizedTosURI, "tos_uri"},
		{request.LocalizedPolicyURI, "policy_uri"},
	}
	for _, f := range fields {
		for tag, val := range f.variants {
			if _, svcErr := ds.i18nService.SetTranslationOverrideForKey(tag, ns, f.key, val); svcErr != nil {
				return &ErrorServerError
			}
		}
	}
	return nil
}

// GetClient retrieves a registered OAuth client by application ID, including all localized variants (AC-25).
func (ds *dcrService) GetClient(ctx context.Context, appID string) (
	*DCRRegistrationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DCRService"))

	app, svcErr := ds.appService.GetApplication(ctx, appID)
	if svcErr != nil {
		if svcErr.Code == application.ErrorApplicationNotFound.Code {
			return nil, &ErrorClientNotFound
		}
		logger.Error("Failed to get application for DCR GET", log.String("appID", appID),
			log.String("error", svcErr.Error))
		return nil, &ErrorServerError
	}

	if len(app.InboundAuthConfig) == 0 || app.InboundAuthConfig[0].OAuthAppConfig == nil {
		return nil, &ErrorClientNotFound
	}

	oauthCfg := app.InboundAuthConfig[0].OAuthAppConfig
	scopeString := strings.Join(oauthCfg.Scopes, " ")

	var jwksURI string
	var jwks map[string]interface{}
	if app.Certificate != nil {
		switch app.Certificate.Type {
		case cert.CertificateTypeJWKSURI:
			jwksURI = app.Certificate.Value
		case cert.CertificateTypeJWKS:
			if err := json.Unmarshal([]byte(app.Certificate.Value), &jwks); err != nil {
				return nil, &ErrorServerError
			}
		}
	}

	response := &DCRRegistrationResponse{
		ClientID:                oauthCfg.ClientID,
		ClientSecret:            oauthCfg.ClientSecret,
		ClientSecretExpiresAt:   ClientSecretExpiresAtNever,
		RedirectURIs:            oauthCfg.RedirectURIs,
		GrantTypes:              oauthCfg.GrantTypes,
		ResponseTypes:           oauthCfg.ResponseTypes,
		ClientName:              app.Name,
		ClientURI:               app.URL,
		LogoURI:                 app.LogoURL,
		TokenEndpointAuthMethod: oauthCfg.TokenEndpointAuthMethod,
		JWKSUri:                 jwksURI,
		JWKS:                    jwks,
		Scope:                   scopeString,
		TosURI:                  app.TosURI,
		PolicyURI:               app.PolicyURI,
		Contacts:                app.Contacts,
		AppID:                   appID,
	}

	// Load all localized variants in a single query (AC-25).
	if ds.i18nService != nil {
		ns := "app." + appID
		localized, i18nErr := ds.i18nService.GetTranslationsByNamespace(ns)
		if i18nErr == nil && localized != nil {
			if m := localized["name"]; len(m) > 0 {
				response.LocalizedClientName = m
			}
			if m := localized["logo_uri"]; len(m) > 0 {
				response.LocalizedLogoURI = m
			}
			if m := localized["tos_uri"]; len(m) > 0 {
				response.LocalizedTosURI = m
			}
			if m := localized["policy_uri"]; len(m) > 0 {
				response.LocalizedPolicyURI = m
			}
		}
	}

	return response, nil
}

// mapApplicationErrorToDCRError maps Application service errors to DCR standard errors.
func (ds *dcrService) mapApplicationErrorToDCRError(appErr *serviceerror.ServiceError) *serviceerror.ServiceError {
	dcrErr := &serviceerror.ServiceError{
		Type:             appErr.Type,
		Error:            appErr.Error,
		ErrorDescription: appErr.ErrorDescription,
	}

	switch appErr.Code {
	// Redirect URI related errors
	case "APP-1014", "APP-1015":
		dcrErr.Code = ErrorInvalidRedirectURI.Code
	// Server errors
	case "APP-5001", "APP-5002":
		dcrErr.Code = ErrorServerError.Code
	// Default fallback for all other client errors
	default:
		dcrErr.Code = ErrorInvalidClientMetadata.Code
	}

	return dcrErr
}
