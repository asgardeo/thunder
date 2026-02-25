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
	"net/http"

	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// ApplicationHandler defines the handler for managing application API requests.
type applicationHandler struct {
	service ApplicationServiceInterface
}

func newApplicationHandler(service ApplicationServiceInterface) *applicationHandler {
	return &applicationHandler{
		service: service,
	}
}

// HandleApplicationPostRequest handles the application request.
func (ah *applicationHandler) HandleApplicationPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	appRequest, err := sysutils.DecodeJSONBody[model.ApplicationRequest](r)
	if err != nil {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: ErrorInvalidRequestFormat.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return
	}

	appDTO := model.ApplicationDTO{
		Name:                      appRequest.Name,
		Description:               appRequest.Description,
		AuthFlowID:                appRequest.AuthFlowID,
		RegistrationFlowID:        appRequest.RegistrationFlowID,
		IsRegistrationFlowEnabled: appRequest.IsRegistrationFlowEnabled,
		ThemeID:                   appRequest.ThemeID,
		LayoutID:                  appRequest.LayoutID,
		Template:                  appRequest.Template,
		URL:                       appRequest.URL,
		LogoURL:                   appRequest.LogoURL,
		Assertion:                 appRequest.Assertion,
		Certificate:               appRequest.Certificate,
		TosURI:                    appRequest.TosURI,
		PolicyURI:                 appRequest.PolicyURI,
		Contacts:                  appRequest.Contacts,
		AllowedUserTypes:          appRequest.AllowedUserTypes,
		Metadata:                  appRequest.Metadata,
	}
	appDTO.InboundAuthConfig = ah.processInboundAuthConfigFromRequest(appRequest.InboundAuthConfig)

	// Create the app using the application service.
	createdAppDTO, svcErr := ah.service.CreateApplication(&appDTO)
	if svcErr != nil {
		ah.handleError(w, svcErr)
		return
	}

	returnApp := model.ApplicationCompleteResponse{
		ID:                        createdAppDTO.ID,
		Name:                      createdAppDTO.Name,
		Description:               createdAppDTO.Description,
		AuthFlowID:                createdAppDTO.AuthFlowID,
		RegistrationFlowID:        createdAppDTO.RegistrationFlowID,
		IsRegistrationFlowEnabled: createdAppDTO.IsRegistrationFlowEnabled,
		ThemeID:                   createdAppDTO.ThemeID,
		LayoutID:                  createdAppDTO.LayoutID,
		Template:                  createdAppDTO.Template,
		URL:                       createdAppDTO.URL,
		LogoURL:                   createdAppDTO.LogoURL,
		Assertion:                 createdAppDTO.Assertion,
		Certificate:               createdAppDTO.Certificate,
		TosURI:                    createdAppDTO.TosURI,
		PolicyURI:                 createdAppDTO.PolicyURI,
		Contacts:                  createdAppDTO.Contacts,
		AllowedUserTypes:          createdAppDTO.AllowedUserTypes,
		Metadata:                  createdAppDTO.Metadata,
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(createdAppDTO.InboundAuthConfig) > 0 {
		success := ah.processInboundAuthConfig(logger, createdAppDTO, &returnApp)
		if !success {
			errResp := apierror.ErrorResponse{
				Code:        ErrorInternalServerError.Code,
				Message:     ErrorInternalServerError.Error,
				Description: ErrorInternalServerError.ErrorDescription,
			}
			sysutils.WriteErrorResponse(w, http.StatusInternalServerError, errResp)
			return
		}
	}

	sysutils.WriteSuccessResponse(w, http.StatusCreated, returnApp)
}

// HandleApplicationListRequest handles the application request.
func (ah *applicationHandler) HandleApplicationListRequest(w http.ResponseWriter, r *http.Request) {
	listResponse, svcErr := ah.service.GetApplicationList()
	if svcErr != nil {
		ah.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, listResponse)
}

// HandleApplicationGetRequest handles the application request.
func (ah *applicationHandler) HandleApplicationGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	id := r.PathValue("id")
	if id == "" {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidApplicationID.Code,
			Message:     ErrorInvalidApplicationID.Error,
			Description: ErrorInvalidApplicationID.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return
	}

	appDTO, svcErr := ah.service.GetApplication(id)
	if svcErr != nil {
		ah.handleError(w, svcErr)
		return
	}

	returnApp := model.ApplicationGetResponse{
		ID:                        appDTO.ID,
		Name:                      appDTO.Name,
		Description:               appDTO.Description,
		AuthFlowID:                appDTO.AuthFlowID,
		RegistrationFlowID:        appDTO.RegistrationFlowID,
		IsRegistrationFlowEnabled: appDTO.IsRegistrationFlowEnabled,
		ThemeID:                   appDTO.ThemeID,
		LayoutID:                  appDTO.LayoutID,
		Template:                  appDTO.Template,
		URL:                       appDTO.URL,
		LogoURL:                   appDTO.LogoURL,
		Assertion:                 appDTO.Assertion,
		Certificate:               appDTO.Certificate,
		TosURI:                    appDTO.TosURI,
		PolicyURI:                 appDTO.PolicyURI,
		Contacts:                  appDTO.Contacts,
		AllowedUserTypes:          appDTO.AllowedUserTypes,
		Metadata:                  appDTO.Metadata,
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(appDTO.InboundAuthConfig) > 0 {
		if appDTO.InboundAuthConfig[0].Type != model.OAuthInboundAuthType {
			logger.Error("Unsupported inbound authentication type returned",
				log.String("type", string(appDTO.InboundAuthConfig[0].Type)))

			errResp := apierror.ErrorResponse{
				Code:        ErrorInternalServerError.Code,
				Message:     ErrorInternalServerError.Error,
				Description: ErrorInternalServerError.ErrorDescription,
			}
			sysutils.WriteErrorResponse(w, http.StatusInternalServerError, errResp)
			return
		}

		returnInboundAuthConfig := appDTO.InboundAuthConfig[0]
		if returnInboundAuthConfig.OAuthAppConfig == nil {
			logger.Error("OAuth application configuration is nil")

			errResp := apierror.ErrorResponse{
				Code:        ErrorInternalServerError.Code,
				Message:     ErrorInternalServerError.Error,
				Description: ErrorInternalServerError.ErrorDescription,
			}
			sysutils.WriteErrorResponse(w, http.StatusInternalServerError, errResp)
			return
		}

		redirectURIs := returnInboundAuthConfig.OAuthAppConfig.RedirectURIs
		if len(redirectURIs) == 0 {
			redirectURIs = []string{}
		}
		grantTypes := returnInboundAuthConfig.OAuthAppConfig.GrantTypes
		if len(grantTypes) == 0 {
			grantTypes = []oauth2const.GrantType{}
		}
		responseTypes := returnInboundAuthConfig.OAuthAppConfig.ResponseTypes
		if len(responseTypes) == 0 {
			responseTypes = []oauth2const.ResponseType{}
		}
		tokenAuthMethod := returnInboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod

		returnInboundAuthConfigs := make([]model.InboundAuthConfig, 0)
		for _, config := range appDTO.InboundAuthConfig {
			oAuthAppConfig := model.OAuthAppConfig{
				ClientID:                config.OAuthAppConfig.ClientID,
				RedirectURIs:            redirectURIs,
				GrantTypes:              grantTypes,
				ResponseTypes:           responseTypes,
				TokenEndpointAuthMethod: tokenAuthMethod,
				PKCERequired:            config.OAuthAppConfig.PKCERequired,
				PublicClient:            config.OAuthAppConfig.PublicClient,
				Token:                   config.OAuthAppConfig.Token,
				Scopes:                  config.OAuthAppConfig.Scopes,
				UserInfo:                config.OAuthAppConfig.UserInfo,
				ScopeClaims:             config.OAuthAppConfig.ScopeClaims,
			}
			returnInboundAuthConfigs = append(returnInboundAuthConfigs, model.InboundAuthConfig{
				Type:           config.Type,
				OAuthAppConfig: &oAuthAppConfig,
			})
		}
		returnApp.InboundAuthConfig = returnInboundAuthConfigs
		returnApp.ClientID = appDTO.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, returnApp)
}

// HandleApplicationPutRequest handles the application request.
func (ah *applicationHandler) HandleApplicationPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	id := r.PathValue("id")
	if id == "" {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidApplicationID.Code,
			Message:     ErrorInvalidApplicationID.Error,
			Description: ErrorInvalidApplicationID.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return
	}

	appRequest, err := sysutils.DecodeJSONBody[model.ApplicationRequest](r)
	if err != nil {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: ErrorInvalidRequestFormat.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return
	}

	updateReqAppDTO := model.ApplicationDTO{
		ID:                        id,
		Name:                      appRequest.Name,
		Description:               appRequest.Description,
		AuthFlowID:                appRequest.AuthFlowID,
		RegistrationFlowID:        appRequest.RegistrationFlowID,
		IsRegistrationFlowEnabled: appRequest.IsRegistrationFlowEnabled,
		ThemeID:                   appRequest.ThemeID,
		LayoutID:                  appRequest.LayoutID,
		Template:                  appRequest.Template,
		URL:                       appRequest.URL,
		LogoURL:                   appRequest.LogoURL,
		Assertion:                 appRequest.Assertion,
		Certificate:               appRequest.Certificate,
		TosURI:                    appRequest.TosURI,
		PolicyURI:                 appRequest.PolicyURI,
		Contacts:                  appRequest.Contacts,
		AllowedUserTypes:          appRequest.AllowedUserTypes,
		Metadata:                  appRequest.Metadata,
	}
	updateReqAppDTO.InboundAuthConfig = ah.processInboundAuthConfigFromRequest(appRequest.InboundAuthConfig)

	// Update the application using the application service.
	updatedAppDTO, svcErr := ah.service.UpdateApplication(id, &updateReqAppDTO)
	if svcErr != nil {
		ah.handleError(w, svcErr)
		return
	}

	returnApp := model.ApplicationCompleteResponse{
		ID:                        updatedAppDTO.ID,
		Name:                      updatedAppDTO.Name,
		Description:               updatedAppDTO.Description,
		AuthFlowID:                updatedAppDTO.AuthFlowID,
		RegistrationFlowID:        updatedAppDTO.RegistrationFlowID,
		IsRegistrationFlowEnabled: updatedAppDTO.IsRegistrationFlowEnabled,
		ThemeID:                   updatedAppDTO.ThemeID,
		LayoutID:                  updatedAppDTO.LayoutID,
		Template:                  updatedAppDTO.Template,
		URL:                       updatedAppDTO.URL,
		LogoURL:                   updatedAppDTO.LogoURL,
		Assertion:                 updatedAppDTO.Assertion,
		Certificate:               updatedAppDTO.Certificate,
		TosURI:                    updatedAppDTO.TosURI,
		PolicyURI:                 updatedAppDTO.PolicyURI,
		Contacts:                  updatedAppDTO.Contacts,
		AllowedUserTypes:          updatedAppDTO.AllowedUserTypes,
		Metadata:                  updatedAppDTO.Metadata,
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(updatedAppDTO.InboundAuthConfig) > 0 {
		success := ah.processInboundAuthConfig(logger, updatedAppDTO, &returnApp)
		if !success {
			errResp := apierror.ErrorResponse{
				Code:        ErrorInternalServerError.Code,
				Message:     ErrorInternalServerError.Error,
				Description: ErrorInternalServerError.ErrorDescription,
			}
			sysutils.WriteErrorResponse(w, http.StatusInternalServerError, errResp)
			return
		}
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, returnApp)
}

// HandleApplicationDeleteRequest handles the application request.
func (ah *applicationHandler) HandleApplicationDeleteRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidApplicationID.Code,
			Message:     ErrorInvalidApplicationID.Error,
			Description: ErrorInvalidApplicationID.ErrorDescription,
		}
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, errResp)
		return
	}

	svcErr := ah.service.DeleteApplication(id)
	if svcErr != nil {
		ah.handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusNoContent, nil)
}

// processInboundAuthConfig prepares the response for OAuth app configuration.
func (ah *applicationHandler) processInboundAuthConfig(logger *log.Logger, appDTO *model.ApplicationDTO,
	returnApp *model.ApplicationCompleteResponse) bool {
	if len(appDTO.InboundAuthConfig) > 0 {
		if appDTO.InboundAuthConfig[0].Type != model.OAuthInboundAuthType {
			logger.Error("Unsupported inbound authentication type returned",
				log.String("type", string(appDTO.InboundAuthConfig[0].Type)))

			return false
		}

		returnInboundAuthConfig := appDTO.InboundAuthConfig[0]
		if returnInboundAuthConfig.OAuthAppConfig == nil {
			logger.Error("OAuth application configuration is nil")
			return false
		}

		redirectURIs := returnInboundAuthConfig.OAuthAppConfig.RedirectURIs
		if len(redirectURIs) == 0 {
			redirectURIs = []string{}
		}
		grantTypes := returnInboundAuthConfig.OAuthAppConfig.GrantTypes
		if len(grantTypes) == 0 {
			grantTypes = []oauth2const.GrantType{}
		}
		responseTypes := returnInboundAuthConfig.OAuthAppConfig.ResponseTypes
		if len(responseTypes) == 0 {
			responseTypes = []oauth2const.ResponseType{}
		}
		tokenAuthMethod := returnInboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod

		returnInboundAuthConfigs := make([]model.InboundAuthConfigComplete, 0)
		for _, config := range appDTO.InboundAuthConfig {
			oAuthAppConfig := model.OAuthAppConfigComplete{
				ClientID:                config.OAuthAppConfig.ClientID,
				ClientSecret:            config.OAuthAppConfig.ClientSecret,
				RedirectURIs:            redirectURIs,
				GrantTypes:              grantTypes,
				ResponseTypes:           responseTypes,
				TokenEndpointAuthMethod: tokenAuthMethod,
				PKCERequired:            config.OAuthAppConfig.PKCERequired,
				PublicClient:            config.OAuthAppConfig.PublicClient,
				Token:                   config.OAuthAppConfig.Token,
				Scopes:                  config.OAuthAppConfig.Scopes,
				UserInfo:                config.OAuthAppConfig.UserInfo,
				ScopeClaims:             config.OAuthAppConfig.ScopeClaims,
			}
			returnInboundAuthConfigs = append(returnInboundAuthConfigs, model.InboundAuthConfigComplete{
				Type:           config.Type,
				OAuthAppConfig: &oAuthAppConfig,
			})
		}
		returnApp.InboundAuthConfig = returnInboundAuthConfigs
		returnApp.ClientID = appDTO.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}

	return true
}

// handleError handles service errors and returns appropriate HTTP responses.
func (ah *applicationHandler) handleError(w http.ResponseWriter,
	svcErr *serviceerror.ServiceError) {
	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		if svcErr.Code == ErrorApplicationNotFound.Code {
			statusCode = http.StatusNotFound
		} else {
			statusCode = http.StatusBadRequest
		}
	}

	sysutils.WriteErrorResponse(w, statusCode, errResp)
}

// processInboundAuthConfigFromRequest processes inbound auth config from request to DTO.
func (ah *applicationHandler) processInboundAuthConfigFromRequest(
	configs []model.InboundAuthConfigComplete) []model.InboundAuthConfigDTO {
	if len(configs) == 0 {
		return nil
	}

	inboundAuthConfigDTOs := make([]model.InboundAuthConfigDTO, 0)
	for _, config := range configs {
		if config.Type != model.OAuthInboundAuthType || config.OAuthAppConfig == nil {
			continue
		}

		inboundAuthConfigDTO := model.InboundAuthConfigDTO{
			Type: config.Type,
			OAuthAppConfig: &model.OAuthAppConfigDTO{
				ClientID:                config.OAuthAppConfig.ClientID,
				ClientSecret:            config.OAuthAppConfig.ClientSecret,
				RedirectURIs:            config.OAuthAppConfig.RedirectURIs,
				GrantTypes:              config.OAuthAppConfig.GrantTypes,
				ResponseTypes:           config.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: config.OAuthAppConfig.TokenEndpointAuthMethod,
				PKCERequired:            config.OAuthAppConfig.PKCERequired,
				PublicClient:            config.OAuthAppConfig.PublicClient,
				Token:                   config.OAuthAppConfig.Token,
				Scopes:                  config.OAuthAppConfig.Scopes,
				UserInfo:                config.OAuthAppConfig.UserInfo,
				ScopeClaims:             config.OAuthAppConfig.ScopeClaims,
			},
		}
		inboundAuthConfigDTOs = append(inboundAuthConfigDTOs, inboundAuthConfigDTO)
	}
	return inboundAuthConfigDTOs
}
