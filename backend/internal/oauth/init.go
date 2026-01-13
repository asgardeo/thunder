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

// Package oauth provides centralized initialization for all OAuth-related services.
package oauth

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	"github.com/asgardeo/thunder/internal/oauth/jwks"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/dcr"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/discovery"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/granthandlers"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/introspect"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/token"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/userinfo"
	"github.com/asgardeo/thunder/internal/oauth/scope"
	"github.com/asgardeo/thunder/internal/observability"
	"github.com/asgardeo/thunder/internal/system/crypto/pki"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/user"
)

// Initialize initializes all OAuth-related services and registers their routes.
func Initialize(
	mux *http.ServeMux,
	applicationService application.ApplicationServiceInterface,
	userService user.UserServiceInterface,
	jwtService jwt.JWTServiceInterface,
	flowExecService flowexec.FlowExecServiceInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
	pkiService pki.PKIServiceInterface,
) {
	jwks.Initialize(mux, pkiService)
	grantHandlerProvider := granthandlers.Initialize(
		mux, jwtService, userService, applicationService, flowExecService)
	scopeValidator := scope.Initialize()
	token.Initialize(mux, applicationService, grantHandlerProvider, scopeValidator, observabilitySvc)
	introspect.Initialize(mux, jwtService)
	userinfo.Initialize(mux, jwtService, applicationService, userService)
	discovery.Initialize(mux)
	dcr.Initialize(mux, applicationService)
}
