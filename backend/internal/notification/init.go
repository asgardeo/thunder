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

package notification

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	filebasedruntime "github.com/asgardeo/thunder/internal/system/file_based_runtime"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"

	"gopkg.in/yaml.v3"
)

// Initialize creates and configures the notification service components.
func Initialize(mux *http.ServeMux, jwtService jwt.JWTServiceInterface) (
	NotificationSenderMgtSvcInterface, OTPServiceInterface) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationInit"))
	var notificationStore notificationStoreInterface
	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		notificationStore = newNotificationFileBasedStore()
	} else {
		notificationStore = newNotificationStore()
	}

	mgtService := newNotificationSenderMgtService(notificationStore)

	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		configs, err := filebasedruntime.GetConfigs("notification-senders")
		if err != nil {
			logger.Fatal("Failed to read notification sender configs from file-based runtime", log.Error(err))
		}
		for _, cfg := range configs {
			senderDTO, err := parseToNotificationSenderDTO(cfg)
			if err != nil {
				logger.Fatal("Error parsing notification sender config", log.Error(err))
			}
			svcErr := validateNotificationSenderForInit(senderDTO)
			if svcErr != nil {
				logger.Fatal("Error validating notification sender",
					log.String("senderName", senderDTO.Name), log.Any("serviceError", svcErr))
			}

			err = notificationStore.createSender(*senderDTO)
			if err != nil {
				logger.Fatal("Failed to store notification sender in file-based store",
					log.String("senderName", senderDTO.Name), log.Error(err))
			}
		}
	}

	otpService := newOTPService(mgtService, jwtService)
	handler := newMessageNotificationSenderHandler(mgtService, otpService)
	registerRoutes(mux, handler)
	return mgtService, otpService
}

func parseToNotificationSenderDTO(data []byte) (*common.NotificationSenderDTO, error) {
	var senderRequest common.NotificationSenderRequestWithID
	err := yaml.Unmarshal(data, &senderRequest)
	if err != nil {
		return nil, err
	}

	senderDTO := &common.NotificationSenderDTO{
		ID:          senderRequest.ID,
		Name:        senderRequest.Name,
		Description: senderRequest.Description,
		Type:        common.NotificationSenderTypeMessage,
	}

	// Parse provider type
	provider, err := parseProviderType(senderRequest.Provider)
	if err != nil {
		return nil, err
	}
	senderDTO.Provider = provider

	// Convert PropertyDTO to Property
	if len(senderRequest.Properties) > 0 {
		properties := make([]cmodels.Property, 0, len(senderRequest.Properties))
		for _, propDTO := range senderRequest.Properties {
			prop, err := cmodels.NewProperty(propDTO.Name, propDTO.Value, propDTO.IsSecret)
			if err != nil {
				return nil, err
			}
			properties = append(properties, *prop)
		}
		senderDTO.Properties = properties
	}

	return senderDTO, nil
}

func parseProviderType(providerStr string) (common.MessageProviderType, error) {
	// Convert string to lowercase for case-insensitive matching
	providerStrLower := common.MessageProviderType(strings.ToLower(providerStr))

	// Check if it's a valid provider
	supportedProviders := []common.MessageProviderType{
		common.MessageProviderTypeVonage,
		common.MessageProviderTypeTwilio,
		common.MessageProviderTypeCustom,
	}

	for _, supportedProvider := range supportedProviders {
		if supportedProvider == providerStrLower {
			return supportedProvider, nil
		}
	}

	return "", fmt.Errorf("unsupported provider type: %s", providerStr)
}

func validateNotificationSenderForInit(sender *common.NotificationSenderDTO) *serviceerror.ServiceError {
	if sender == nil {
		return &ErrorInvalidSenderName
	}
	if strings.TrimSpace(sender.Name) == "" {
		return &ErrorInvalidSenderName
	}

	// Validate provider type
	if strings.TrimSpace(string(sender.Provider)) == "" {
		return &ErrorInvalidProvider
	}
	supportedProviders := []common.MessageProviderType{
		common.MessageProviderTypeVonage,
		common.MessageProviderTypeTwilio,
		common.MessageProviderTypeCustom,
	}
	isValidProvider := slices.Contains(supportedProviders, sender.Provider)
	if !isValidProvider {
		return &ErrorInvalidProvider
	}

	// Validate properties based on provider
	if len(sender.Properties) == 0 {
		svcErr := ErrorInvalidRequestFormat
		svcErr.ErrorDescription = "notification sender properties cannot be empty"
		return &svcErr
	}

	var err error
	switch sender.Provider {
	case common.MessageProviderTypeTwilio:
		err = validateTwilioPropertiesForInit(sender.Properties)
	case common.MessageProviderTypeVonage:
		err = validateVonagePropertiesForInit(sender.Properties)
	case common.MessageProviderTypeCustom:
		err = validateCustomPropertiesForInit(sender.Properties)
	default:
		svcErr := ErrorInvalidProvider
		return &svcErr
	}

	if err != nil {
		svcErr := ErrorInvalidRequestFormat
		svcErr.ErrorDescription = err.Error()
		return &svcErr
	}

	return nil
}

func validateTwilioPropertiesForInit(properties []cmodels.Property) error {
	requiredProps := map[string]bool{
		"account_sid": false,
		"auth_token":  false,
		"sender_id":   false,
	}
	return validateSenderPropertiesForInit(properties, requiredProps)
}

func validateVonagePropertiesForInit(properties []cmodels.Property) error {
	requiredProps := map[string]bool{
		"api_key":    false,
		"api_secret": false,
		"sender_id":  false,
	}
	return validateSenderPropertiesForInit(properties, requiredProps)
}

func validateCustomPropertiesForInit(properties []cmodels.Property) error {
	requiredProps := map[string]bool{
		"url": false,
	}
	return validateSenderPropertiesForInit(properties, requiredProps)
}

func validateSenderPropertiesForInit(properties []cmodels.Property, requiredProps map[string]bool) error {
	for _, prop := range properties {
		if _, exists := requiredProps[prop.GetName()]; exists {
			requiredProps[prop.GetName()] = true
		}
	}

	for propName, found := range requiredProps {
		if !found {
			return fmt.Errorf("required property '%s' is missing", propName)
		}
	}

	return nil
}

// registerRoutes registers the HTTP routes for notification services.
func registerRoutes(mux *http.ServeMux, handler *messageNotificationSenderHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /notification-senders/message",
		handler.HandleSenderListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("POST /notification-senders/message",
		handler.HandleSenderCreateRequest, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /notification-senders/message/{id}",
		handler.HandleSenderGetRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /notification-senders/message/{id}",
		handler.HandleSenderUpdateRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /notification-senders/message/{id}",
		handler.HandleSenderDeleteRequest, opts2))

	opts3 := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /notification-senders/otp/send",
		handler.HandleOTPSendRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("POST /notification-senders/otp/verify",
		handler.HandleOTPVerifyRequest, opts3))
}
