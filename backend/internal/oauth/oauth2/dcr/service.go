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
	inboundmodel "github.com/asgardeo/thunder/internal/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauthutils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/transaction"
)

// DCRServiceInterface defines the interface for the DCR service.
type DCRServiceInterface interface {
	RegisterClient(
		ctx context.Context, request *DCRRegistrationRequest,
	) (*DCRRegistrationResponse, *serviceerror.ServiceError)
}

// dcrService is the default implementation of DCRServiceInterface.
type dcrService struct {
	appService    application.ApplicationServiceInterface
	ouService     ou.OrganizationUnitServiceInterface
	transactioner transaction.Transactioner
}

// newDCRService creates a new instance of dcrService.
func newDCRService(
	appService application.ApplicationServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	transactioner transaction.Transactioner,
) DCRServiceInterface {
	return &dcrService{
		appService:    appService,
		ouService:     ouService,
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
			logger.Error("Failed to retrieve root organization units for DCR",
				log.String("error", svcErr.Error.DefaultValue))
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
		logger.Error("Failed to convert DCR request to application DTO", log.String("error", svcErr.Error.DefaultValue))
		return nil, &ErrorServerError
	}

	var response *DCRRegistrationResponse
	var capturedErr *serviceerror.ServiceError

	err := ds.transactioner.Transact(ctx, func(ctx context.Context) error {
		createdApp, svcErr := ds.appService.CreateApplication(ctx, appDTO)
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

		var convErr *serviceerror.ServiceError
		response, convErr = ds.convertApplicationToDCRResponse(createdApp, request.ClientName)
		if convErr != nil {
			logger.Error("Failed to convert application to DCR response",
				log.String("error", convErr.Error.DefaultValue))
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

	return response, nil
}

// convertDCRToApplication converts DCR registration request to Application DTO.
func (ds *dcrService) convertDCRToApplication(request *DCRRegistrationRequest) (
	*model.ApplicationDTO, *serviceerror.ServiceError) {
	isPublicClient := request.TokenEndpointAuthMethod == oauth2const.TokenEndpointAuthMethodNone

	// Map JWKS/JWKS_URI to application-level certificate
	var appCertificate *inboundmodel.Certificate
	if request.JWKSUri != "" {
		appCertificate = &inboundmodel.Certificate{
			Type:  cert.CertificateTypeJWKSURI,
			Value: request.JWKSUri,
		}
	} else if len(request.JWKS) > 0 {
		jwksBytes, err := json.Marshal(request.JWKS)
		if err != nil {
			return nil, &ErrorServerError
		}
		appCertificate = &inboundmodel.Certificate{
			Type:  cert.CertificateTypeJWKS,
			Value: string(jwksBytes),
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
		ClientID:                           clientID,
		RedirectURIs:                       request.RedirectURIs,
		GrantTypes:                         request.GrantTypes,
		ResponseTypes:                      request.ResponseTypes,
		TokenEndpointAuthMethod:            request.TokenEndpointAuthMethod,
		PublicClient:                       isPublicClient,
		PKCERequired:                       isPublicClient,
		RequirePushedAuthorizationRequests: request.RequirePushedAuthorizationRequests,
		Scopes:                             scopes,
		UserInfo:                           buildUserInfoConfig(request),
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

// buildUserInfoConfig maps UserInfo alg fields from a DCR request to a UserInfoConfig.
// ResponseType is derived from the algorithm fields per OIDC DCR conventions.
func buildUserInfoConfig(request *DCRRegistrationRequest) *inboundmodel.UserInfoConfig {
	if request.UserInfoSignedResponseAlg == "" && request.UserInfoEncryptedResponseAlg == "" &&
		request.UserInfoEncryptedResponseEnc == "" {
		return nil
	}
	hasSign := request.UserInfoSignedResponseAlg != ""
	hasEnc := request.UserInfoEncryptedResponseAlg != ""
	var responseType inboundmodel.UserInfoResponseType
	switch {
	case hasSign && hasEnc:
		responseType = inboundmodel.UserInfoResponseTypeNESTEDJWT
	case hasEnc:
		responseType = inboundmodel.UserInfoResponseTypeJWE
	case hasSign:
		responseType = inboundmodel.UserInfoResponseTypeJWS
	default:
		responseType = inboundmodel.UserInfoResponseTypeJSON
	}
	return &inboundmodel.UserInfoConfig{
		ResponseType:  responseType,
		SigningAlg:    request.UserInfoSignedResponseAlg,
		EncryptionAlg: request.UserInfoEncryptedResponseAlg,
		EncryptionEnc: request.UserInfoEncryptedResponseEnc,
	}
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

	var userInfoSignedAlg, userInfoEncryptedAlg, userInfoEncryptedEnc string
	if oauthConfig.UserInfo != nil {
		userInfoSignedAlg = oauthConfig.UserInfo.SigningAlg
		userInfoEncryptedAlg = oauthConfig.UserInfo.EncryptionAlg
		userInfoEncryptedEnc = oauthConfig.UserInfo.EncryptionEnc
	}

	response := &DCRRegistrationResponse{
		ClientID:                           oauthConfig.ClientID,
		ClientSecret:                       oauthConfig.ClientSecret,
		ClientSecretExpiresAt:              ClientSecretExpiresAtNever,
		RedirectURIs:                       oauthConfig.RedirectURIs,
		GrantTypes:                         oauthConfig.GrantTypes,
		ResponseTypes:                      oauthConfig.ResponseTypes,
		ClientName:                         clientName,
		ClientURI:                          appDTO.URL,
		LogoURI:                            appDTO.LogoURL,
		TokenEndpointAuthMethod:            oauthConfig.TokenEndpointAuthMethod,
		JWKSUri:                            jwksURI,
		JWKS:                               jwks,
		Scope:                              scopeString,
		TosURI:                             appDTO.TosURI,
		PolicyURI:                          appDTO.PolicyURI,
		Contacts:                           appDTO.Contacts,
		AppID:                              oauthConfig.AppID,
		RequirePushedAuthorizationRequests: oauthConfig.RequirePushedAuthorizationRequests,
		UserInfoSignedResponseAlg:          userInfoSignedAlg,
		UserInfoEncryptedResponseAlg:       userInfoEncryptedAlg,
		UserInfoEncryptedResponseEnc:       userInfoEncryptedEnc,
	}

	return response, nil
}

// mapApplicationErrorToDCRError maps Application service errors to DCR standard errors.
func (ds *dcrService) mapApplicationErrorToDCRError(
	appErr *serviceerror.ServiceError) *serviceerror.ServiceError {
	dcrErr := &serviceerror.ServiceError{
		Type:             appErr.Type,
		Error:            appErr.Error,
		ErrorDescription: appErr.ErrorDescription,
	}

	switch appErr.Code {
	// Redirect URI validation errors
	case "APP-1012":
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
