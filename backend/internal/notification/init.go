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
	"net/http"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize creates and configures the notification service components.
func Initialize(mux *http.ServeMux, jwtService jwt.JWTServiceInterface) (
	NotificationSenderMgtSvcInterface, OTPServiceInterface, declarativeresource.ResourceExporter, error) {
	var notificationStore notificationStoreInterface
	var tx transaction.Transactioner

	if config.GetThunderRuntime().Config.DeclarativeResources.Enabled {
		notificationStore = newNotificationFileBasedStore()
	} else {
		notificationStore = newNotificationStore()
		client, err := provider.GetDBProvider().GetConfigDBClient()
		if err != nil {
			log.GetLogger().Error("Failed to initialize database client for notification service",
				log.Error(err))
			return nil, nil, nil, err
		}

		var txErr error
		tx, txErr = client.GetTransactioner()
		if txErr != nil {
			log.GetLogger().Error("Failed to initialize database transactioner for notification service",
				log.Error(txErr))
			return nil, nil, nil, txErr
		}
	}

	mgtService := newNotificationSenderMgtService(notificationStore, tx)

	if config.GetThunderRuntime().Config.DeclarativeResources.Enabled {
		if err := loadDeclarativeResources(notificationStore); err != nil {
			return nil, nil, nil, err
		}
	}

	otpService := newOTPService(mgtService, jwtService)
	handler := newMessageNotificationSenderHandler(mgtService, otpService)
	registerRoutes(mux, handler)

	// Create and return exporter
	exporter := newNotificationSenderExporter(mgtService)
	return mgtService, otpService, exporter, nil
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
	mux.HandleFunc(middleware.WithCORS("OPTIONS /notification-senders/message",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts1))

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
	mux.HandleFunc(middleware.WithCORS("OPTIONS /notification-senders/message/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts2))

	opts3 := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /notification-senders/otp/send",
		handler.HandleOTPSendRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /notification-senders/otp/send",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3))
	mux.HandleFunc(middleware.WithCORS("POST /notification-senders/otp/verify",
		handler.HandleOTPVerifyRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /notification-senders/otp/verify",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3))
}
