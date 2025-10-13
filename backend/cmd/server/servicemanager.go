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

// Package managers provides functionality for managing and registering system services.
package main

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/flow"
	"github.com/asgardeo/thunder/internal/group"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/notification"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/services"
	"github.com/asgardeo/thunder/internal/userschema"
)

// registerServices registers all the services with the provided HTTP multiplexer.
func registerServices(mux *http.ServeMux) {
	logger := log.GetLogger()

	// Load the server's private key for signing JWTs.
	jwtService := jwt.GetJWTService()
	if err := jwtService.Init(); err != nil { // TODO: Two-Phase Initialization is anti-pattern. Refactor this.
		logger.Fatal("Failed to load private key", log.Error(err))
	}

	_ = userschema.Initialize(mux)
	ouService := ou.Initialize(mux)
	_ = group.Initialize(mux, ouService)

	_ = idp.Initialize(mux)
	_ = notification.Initialize(mux, jwtService)

	// TODO: Legacy way of initializing services. These need to be refactored in the future aligning to the
	// dependency injection pattern used above.

	// Register the health service.
	services.NewHealthCheckService(mux)

	// Register the token service.
	services.NewTokenService(mux)

	// Register the authorization service.
	services.NewAuthorizationService(mux)

	// Register the JWKS service.
	services.NewJWKSAPIService(mux)

	// Register the introspection service.
	services.NewIntrospectionAPIService(mux)

	// Register the User service.
	services.NewUserService(mux)

	// Register the Application service.
	services.NewApplicationService(mux)

	// Register the flow execution service.
	services.NewFlowExecutionService(mux)

	// Register the authentication service.
	services.NewAuthenticationService(mux)

	svc := flow.GetFlowExecService()
	if err := svc.Init(); err != nil {
		logger.Fatal("Failed to initialize flow service", log.Error(err))
	}
}
