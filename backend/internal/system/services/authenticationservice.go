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

package services

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/authn"
	"github.com/asgardeo/thunder/internal/system/server"
)

// AuthenticationService defines the service for handling authentication-related requests.
type AuthenticationService struct {
	ServerOpsService server.ServerOperationServiceInterface
	authHandler      *authn.AuthenticationHandler
}

// NewAuthenticationService creates a new instance of AuthenticationService.
func NewAuthenticationService(mux *http.ServeMux) ServiceInterface {
	instance := &AuthenticationService{
		ServerOpsService: server.NewServerOperationService(),
		authHandler:      authn.NewAuthenticationHandler(),
	}
	instance.RegisterRoutes(mux)

	return instance
}

// RegisterRoutes registers the routes for the AuthenticationService.
func (s *AuthenticationService) RegisterRoutes(mux *http.ServeMux) {
	opts := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "POST",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}

	// SMS OTP routes
	s.ServerOpsService.WrapHandleFunction(mux, "POST /auth/otp/sms/send", &opts,
		s.authHandler.HandleSendSMSOTPRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "POST /auth/otp/sms/verify", &opts,
		s.authHandler.HandleVerifySMSOTPRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /auth/otp/sms/send", &opts,
		optionsNoContentHandler)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /auth/otp/sms/verify", &opts,
		optionsNoContentHandler)

	// Google OAuth routes
	s.ServerOpsService.WrapHandleFunction(mux, "POST /auth/oauth/google/start", &opts,
		s.authHandler.HandleGoogleAuthStartRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "POST /auth/oauth/google/finish", &opts,
		s.authHandler.HandleGoogleAuthFinishRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /auth/oauth/google/start", &opts,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /auth/oauth/google/finish", &opts,
		optionsNoContentHandler)

	// GitHub OAuth routes
	s.ServerOpsService.WrapHandleFunction(mux, "POST /auth/oauth/github/start", &opts,
		s.authHandler.HandleGithubAuthStartRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "POST /auth/oauth/github/finish", &opts,
		s.authHandler.HandleGithubAuthFinishRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /auth/oauth/github/start", &opts,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /auth/oauth/github/finish", &opts,
		optionsNoContentHandler)

	// Standard OAuth routes
	s.ServerOpsService.WrapHandleFunction(mux, "POST /auth/oauth/standard/start", &opts,
		s.authHandler.HandleStandardOAuthStartRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "POST /auth/oauth/standard/finish", &opts,
		s.authHandler.HandleStandardOAuthFinishRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /auth/oauth/standard/start", &opts,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /auth/oauth/standard/finish", &opts,
		optionsNoContentHandler)
}

// optionsNoContentHandler handles OPTIONS requests by responding with 204 No Content.
func optionsNoContentHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
