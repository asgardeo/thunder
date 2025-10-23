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
	"fmt"
	"time"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

// DCRServiceInterface defines the interface for the DCR service.
type DCRServiceInterface interface {
	RegisterClient(request *DCRRegistrationRequest) (*DCRRegistrationResponse, *serviceerror.ServiceError)
}

// dcrService is the default implementation of DCRServiceInterface.
type dcrService struct {
	appService application.ApplicationServiceInterface
}

// newDCRService creates a new instance of dcrService.
func newDCRService(appService application.ApplicationServiceInterface) DCRServiceInterface {
	return &dcrService{
		appService: appService,
	}
}

// RegisterClient registers a new OAuth client using Dynamic Client Registration.
func (ds *dcrService) RegisterClient(request *DCRRegistrationRequest) (
	*DCRRegistrationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DCRService"))

	if request == nil {
		return nil, &ErrorInvalidRequestFormat
	}

	appDTO := ds.convertDCRToApplicationDTO(request)

	createdApp, err := ds.appService.CreateApplication(appDTO)
	if err != nil {
		logger.Error("Failed to create application via Application service", log.String("error", err.Error))
		return nil, ds.mapApplicationErrorToDCRError(err)
	}

	response := ds.convertApplicationToDCRResponse(createdApp)

	return response, nil
}

// convertDCRToApplicationDTO converts DCR registration request to Application DTO.
func (ds *dcrService) convertDCRToApplicationDTO(request *DCRRegistrationRequest) *model.ApplicationDTO {
	isPublicClient := request.TokenEndpointAuthMethod == oauth2const.TokenEndpointAuthMethodNone

	oauthAppConfig := &model.OAuthAppConfigDTO{
		RedirectURIs:            request.RedirectURIs,
		GrantTypes:              request.GrantTypes,
		ResponseTypes:           request.ResponseTypes,
		TokenEndpointAuthMethod: request.TokenEndpointAuthMethod,
		PublicClient:            isPublicClient,
		JWKSUri:                 request.JWKSUri,
		JWKS:                    request.JWKS,
		Scope:                   request.Scope,
	}

	inboundAuthConfig := []model.InboundAuthConfigDTO{
		{
			Type:           model.OAuthInboundAuthType,
			OAuthAppConfig: oauthAppConfig,
		},
	}

	appName := request.ClientName
	if appName == "" {
		appName = ds.generateDefaultClientName()
	}

	appDTO := &model.ApplicationDTO{
		Name:              appName,
		URL:               request.ClientURI,
		LogoURL:           request.LogoURI,
		InboundAuthConfig: inboundAuthConfig,
		TosURI:            request.TosURI,
		PolicyURI:         request.PolicyURI,
		Contacts:          request.Contacts,
	}

	return appDTO
}

// convertApplicationToDCRResponse converts Application DTO to DCR registration response.
func (ds *dcrService) convertApplicationToDCRResponse(appDTO *model.ApplicationDTO) *DCRRegistrationResponse {
	if len(appDTO.InboundAuthConfig) == 0 || appDTO.InboundAuthConfig[0].OAuthAppConfig == nil {
		return &DCRRegistrationResponse{}
	}

	oauthConfig := appDTO.InboundAuthConfig[0].OAuthAppConfig

	response := &DCRRegistrationResponse{
		ClientID:                oauthConfig.ClientID,
		ClientSecret:            oauthConfig.ClientSecret,
		ClientSecretExpiresAt:   ClientSecretExpiresAtNever,
		RedirectURIs:            oauthConfig.RedirectURIs,
		GrantTypes:              oauthConfig.GrantTypes,
		ResponseTypes:           oauthConfig.ResponseTypes,
		ClientName:              appDTO.Name,
		ClientURI:               appDTO.URL,
		LogoURI:                 appDTO.LogoURL,
		TokenEndpointAuthMethod: oauthConfig.TokenEndpointAuthMethod,
		JWKSUri:                 oauthConfig.JWKSUri,
		JWKS:                    oauthConfig.JWKS,
		Scope:                   oauthConfig.Scope,
		TosURI:                  appDTO.TosURI,
		PolicyURI:               appDTO.PolicyURI,
		Contacts:                appDTO.Contacts,
		AppID:                   oauthConfig.AppID,
	}

	return response
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
	case "APP-1006", "APP-1014", "APP-1015":
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

// generateDefaultClientName generates a unique default client name.
// This is used when client_name is not provided in the DCR request.
func (ds *dcrService) generateDefaultClientName() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("OAuth2 Client %d", timestamp)
}
