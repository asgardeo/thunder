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
	"encoding/json"
	"net/http"

	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
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
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: ErrorInvalidRequestFormat.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	appDTO := model.ApplicationDTO{
		Name:                      appRequest.Name,
		Description:               appRequest.Description,
		AuthFlowGraphID:           appRequest.AuthFlowGraphID,
		RegistrationFlowGraphID:   appRequest.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: appRequest.IsRegistrationFlowEnabled,
		URL:                       appRequest.URL,
		LogoURL:                   appRequest.LogoURL,
		Token:                     appRequest.Token,
		Certificate:               appRequest.Certificate,
		TosURI:                    appRequest.TosURI,
		PolicyURI:                 appRequest.PolicyURI,
		Contacts:                  appRequest.Contacts,
	}
	appDTO.InboundAuthConfig = ah.processInboundAuthConfigFromRequest(appRequest.InboundAuthConfig)

	// Create the app using the application service.
	createdAppDTO, svcErr := ah.service.CreateApplication(&appDTO)
	if svcErr != nil {
		ah.handleError(w, logger, svcErr)
		return
	}

	returnApp := model.ApplicationCompleteResponse{
		ID:                        createdAppDTO.ID,
		Name:                      createdAppDTO.Name,
		Description:               createdAppDTO.Description,
		AuthFlowGraphID:           createdAppDTO.AuthFlowGraphID,
		RegistrationFlowGraphID:   createdAppDTO.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: createdAppDTO.IsRegistrationFlowEnabled,
		URL:                       createdAppDTO.URL,
		LogoURL:                   createdAppDTO.LogoURL,
		Token:                     createdAppDTO.Token,
		Certificate:               createdAppDTO.Certificate,
		TosURI:                    createdAppDTO.TosURI,
		PolicyURI:                 createdAppDTO.PolicyURI,
		Contacts:                  createdAppDTO.Contacts,
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(createdAppDTO.InboundAuthConfig) > 0 {
		success := ah.processInboundAuthConfig(logger, createdAppDTO, &returnApp)
		if !success {
			w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
			w.WriteHeader(http.StatusInternalServerError)

			errResp := apierror.ErrorResponse{
				Code:        ErrorInternalServerError.Code,
				Message:     ErrorInternalServerError.Error,
				Description: ErrorInternalServerError.ErrorDescription,
			}
			if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
				logger.Error("Error encoding error response", log.Error(encodeErr))
				http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	if encodeErr := json.NewEncoder(w).Encode(returnApp); encodeErr != nil {
		logger.Error("Error encoding response", log.Error(encodeErr))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleApplicationListRequest handles the application request.
func (ah *applicationHandler) HandleApplicationListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	listResponse, svcErr := ah.service.GetApplicationList()
	if svcErr != nil {
		ah.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if encodeErr := json.NewEncoder(w).Encode(listResponse); encodeErr != nil {
		logger.Error("Error encoding response", log.Error(encodeErr))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleApplicationGetRequest handles the application request.
func (ah *applicationHandler) HandleApplicationGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidApplicationID.Code,
			Message:     ErrorInvalidApplicationID.Error,
			Description: ErrorInvalidApplicationID.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	appDTO, svcErr := ah.service.GetApplication(id)
	if svcErr != nil {
		ah.handleError(w, logger, svcErr)
		return
	}

	returnApp := model.ApplicationGetResponse{
		ID:                        appDTO.ID,
		Name:                      appDTO.Name,
		Description:               appDTO.Description,
		AuthFlowGraphID:           appDTO.AuthFlowGraphID,
		RegistrationFlowGraphID:   appDTO.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: appDTO.IsRegistrationFlowEnabled,
		URL:                       appDTO.URL,
		LogoURL:                   appDTO.LogoURL,
		Token:                     appDTO.Token,
		Certificate:               appDTO.Certificate,
		TosURI:                    appDTO.TosURI,
		PolicyURI:                 appDTO.PolicyURI,
		Contacts:                  appDTO.Contacts,
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(appDTO.InboundAuthConfig) > 0 {
		if appDTO.InboundAuthConfig[0].Type != model.OAuthInboundAuthType {
			logger.Error("Unsupported inbound authentication type returned",
				log.String("type", string(appDTO.InboundAuthConfig[0].Type)))

			w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
			w.WriteHeader(http.StatusInternalServerError)

			errResp := apierror.ErrorResponse{
				Code:        ErrorInternalServerError.Code,
				Message:     ErrorInternalServerError.Error,
				Description: ErrorInternalServerError.ErrorDescription,
			}
			if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
				logger.Error("Error encoding error response", log.Error(encodeErr))
				http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
			}
			return
		}

		returnInboundAuthConfig := appDTO.InboundAuthConfig[0]
		if returnInboundAuthConfig.OAuthAppConfig == nil {
			logger.Error("OAuth application configuration is nil")

			w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
			w.WriteHeader(http.StatusInternalServerError)

			errResp := apierror.ErrorResponse{
				Code:        ErrorInternalServerError.Code,
				Message:     ErrorInternalServerError.Error,
				Description: ErrorInternalServerError.ErrorDescription,
			}
			if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
				logger.Error("Error encoding error response", log.Error(encodeErr))
				http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
			}
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
			}
			returnInboundAuthConfigs = append(returnInboundAuthConfigs, model.InboundAuthConfig{
				Type:           config.Type,
				OAuthAppConfig: &oAuthAppConfig,
			})
		}
		returnApp.InboundAuthConfig = returnInboundAuthConfigs
		returnApp.ClientID = appDTO.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if encodeErr := json.NewEncoder(w).Encode(returnApp); encodeErr != nil {
		logger.Error("Error encoding response", log.Error(encodeErr))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleApplicationPutRequest handles the application request.
func (ah *applicationHandler) HandleApplicationPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidApplicationID.Code,
			Message:     ErrorInvalidApplicationID.Error,
			Description: ErrorInvalidApplicationID.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	appRequest, err := sysutils.DecodeJSONBody[model.ApplicationRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: ErrorInvalidRequestFormat.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	updateReqAppDTO := model.ApplicationDTO{
		ID:                        id,
		Name:                      appRequest.Name,
		Description:               appRequest.Description,
		AuthFlowGraphID:           appRequest.AuthFlowGraphID,
		RegistrationFlowGraphID:   appRequest.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: appRequest.IsRegistrationFlowEnabled,
		URL:                       appRequest.URL,
		LogoURL:                   appRequest.LogoURL,
		Token:                     appRequest.Token,
		Certificate:               appRequest.Certificate,
		TosURI:                    appRequest.TosURI,
		PolicyURI:                 appRequest.PolicyURI,
		Contacts:                  appRequest.Contacts,
	}
	updateReqAppDTO.InboundAuthConfig = ah.processInboundAuthConfigFromRequest(appRequest.InboundAuthConfig)

	// Update the application using the application service.
	updatedAppDTO, svcErr := ah.service.UpdateApplication(id, &updateReqAppDTO)
	if svcErr != nil {
		ah.handleError(w, logger, svcErr)
		return
	}

	returnApp := model.ApplicationCompleteResponse{
		ID:                        updatedAppDTO.ID,
		Name:                      updatedAppDTO.Name,
		Description:               updatedAppDTO.Description,
		AuthFlowGraphID:           updatedAppDTO.AuthFlowGraphID,
		RegistrationFlowGraphID:   updatedAppDTO.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: updatedAppDTO.IsRegistrationFlowEnabled,
		URL:                       updatedAppDTO.URL,
		LogoURL:                   updatedAppDTO.LogoURL,
		Token:                     updatedAppDTO.Token,
		Certificate:               updatedAppDTO.Certificate,
		TosURI:                    updatedAppDTO.TosURI,
		PolicyURI:                 updatedAppDTO.PolicyURI,
		Contacts:                  updatedAppDTO.Contacts,
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(updatedAppDTO.InboundAuthConfig) > 0 {
		success := ah.processInboundAuthConfig(logger, updatedAppDTO, &returnApp)
		if !success {
			w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
			w.WriteHeader(http.StatusInternalServerError)

			errResp := apierror.ErrorResponse{
				Code:        ErrorInternalServerError.Code,
				Message:     ErrorInternalServerError.Error,
				Description: ErrorInternalServerError.ErrorDescription,
			}
			if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
				logger.Error("Error encoding error response", log.Error(encodeErr))
				http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if encodeErr := json.NewEncoder(w).Encode(returnApp); encodeErr != nil {
		logger.Error("Error encoding response", log.Error(encodeErr))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleApplicationDeleteRequest handles the application request.
func (ah *applicationHandler) HandleApplicationDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidApplicationID.Code,
			Message:     ErrorInvalidApplicationID.Error,
			Description: ErrorInvalidApplicationID.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	svcErr := ah.service.DeleteApplication(id)
	if svcErr != nil {
		ah.handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
func (ah *applicationHandler) handleError(w http.ResponseWriter, logger *log.Logger,
	svcErr *serviceerror.ServiceError) {
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)

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
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Error encoding error response", log.Error(err))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
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
			},
		}
		inboundAuthConfigDTOs = append(inboundAuthConfigDTOs, inboundAuthConfigDTO)
	}
	return inboundAuthConfigDTOs
}
