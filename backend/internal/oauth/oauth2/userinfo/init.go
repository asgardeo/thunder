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

package userinfo

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/user"
)

// Initialize initializes the userinfo handler and registers its routes.
func Initialize(
	mux *http.ServeMux,
	jwtService jwt.JWTServiceInterface,
	applicationService application.ApplicationServiceInterface,
	userService user.UserServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
) userInfoServiceInterface {
	userInfoService := newUserInfoService(jwtService, applicationService, userService, ouService)
	userInfoHandler := newUserInfoHandler(userInfoService)
	registerRoutes(mux, userInfoHandler)
	return userInfoService
}

// registerRoutes registers the routes for the UserInfo endpoint.
func registerRoutes(mux *http.ServeMux, userInfoHandler *userInfoHandler) {
	opts := middleware.CORSOptions{
		AllowedMethods:   "GET, POST, OPTIONS",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET "+constants.OAuth2UserInfoEndpoint,
		userInfoHandler.HandleUserInfo, opts))
	mux.HandleFunc(middleware.WithCORS("POST "+constants.OAuth2UserInfoEndpoint,
		userInfoHandler.HandleUserInfo, opts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS "+constants.OAuth2UserInfoEndpoint,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts))
}
