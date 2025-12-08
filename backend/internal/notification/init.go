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
	"strings"

	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/immutableresource"
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
		// Type assert to access Storer interface for resource loading
		fileBasedStore, ok := notificationStore.(*notificationFileBasedStore)
		if !ok {
			logger.Fatal("Failed to assert notificationStore to *notificationFileBasedStore")
		}

		resourceConfig := immutableresource.ResourceConfig{
			ResourceType:  "NotificationSender",
			DirectoryName: "notification_senders",
			Parser:        parseToNotificationSenderDTOWrapper,
			Validator:     validateNotificationSenderWrapper,
			IDExtractor: func(data interface{}) string {
				return data.(*common.NotificationSenderDTO).ID
			},
		}

		loader := immutableresource.NewResourceLoader(resourceConfig, fileBasedStore)
		if err := loader.LoadResources(); err != nil {
			logger.Fatal("Failed to load notification sender resources", log.Error(err))
		}
	}

	otpService := newOTPService(mgtService, jwtService)
	handler := newMessageNotificationSenderHandler(mgtService, otpService)
	registerRoutes(mux, handler)
	return mgtService, otpService
}

// parseToNotificationSenderDTOWrapper wraps parseToNotificationSenderDTO to match ResourceConfig.Parser signature
func parseToNotificationSenderDTOWrapper(data []byte) (interface{}, error) {
	return parseToNotificationSenderDTO(data)
}

// validateNotificationSenderWrapper wraps validation logic to match ResourceConfig.Validator signature
func validateNotificationSenderWrapper(data interface{}) error {
	sender := data.(*common.NotificationSenderDTO)

	// Validate notification sender
	if validationErr := validateNotificationSender(*sender); validationErr != nil {
		return fmt.Errorf("invalid notification sender configuration for '%s': %s - %s",
			sender.Name, validationErr.Error, validationErr.ErrorDescription)
	}

	return nil
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
