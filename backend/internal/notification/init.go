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

// Package notification handles the initialization of notification services.
package notification

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize creates and configures the notification service components.
func Initialize(mux *http.ServeMux, jwtService jwt.JWTServiceInterface) NotificationSenderMgtSvcInterface {
	mgtService := newNotificationSenderMgtService()
	otpService := newOTPService(mgtService, jwtService)
	handler := newMessageNotificationSenderHandler(mgtService, otpService)
	registerRoutes(mux, handler)
	return mgtService
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
