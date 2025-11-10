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

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/authn"
	"github.com/asgardeo/thunder/internal/authz"
	"github.com/asgardeo/thunder/internal/branding"
	"github.com/asgardeo/thunder/internal/cert"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	"github.com/asgardeo/thunder/internal/flow/flowmgt"
	"github.com/asgardeo/thunder/internal/group"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/notification"
	"github.com/asgardeo/thunder/internal/oauth"
	"github.com/asgardeo/thunder/internal/observability"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/role"
	"github.com/asgardeo/thunder/internal/system/export"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/services"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/internal/userschema"
)

// ObservabilitySvc is the observability service instance. This is used in main.go for graceful shutdown.
var ObservabilitySvc observability.ObservabilityServiceInterface

// registerServices registers all the services with the provided HTTP multiplexer.
// Returns the observability service for use in shutdown.
func registerServices(
	mux *http.ServeMux,
	jwtService jwt.JWTServiceInterface,
) {
	logger := log.GetLogger()

	observabilitySvc := observability.Initialize()

	ouService := ou.Initialize(mux)
	userSchemaService := userschema.Initialize(mux)
	userService := user.Initialize(mux, ouService, userSchemaService)
	groupService := group.Initialize(mux, ouService, userService)
	roleService := role.Initialize(mux, userService, groupService, ouService)
	authZService := authz.Initialize(roleService)

	idpService := idp.Initialize(mux)
	_, otpService := notification.Initialize(mux, jwtService)

	// Initialize authentication services.
	_, authSvcRegistry := authn.Initialize(mux, idpService, jwtService, userService, otpService)

	// Initialize flow and executor services.
	flowFactory := flowcore.Initialize()
	execRegistry := executor.Initialize(flowFactory, userService, ouService,
		idpService, otpService, jwtService, authSvcRegistry, authZService)

	flowMgtService, err := flowmgt.Initialize(flowFactory, execRegistry)
	if err != nil {
		logger.Fatal("Failed to initialize FlowMgtService", log.Error(err))
	}
	certservice := cert.Initialize()
	brandingService := branding.Initialize(mux)
	applicationService := application.Initialize(mux, certservice, flowMgtService, brandingService, userSchemaService)

	// Initialize export service with application service dependency
	_ = export.Initialize(mux, applicationService)

	flowExecService := flowexec.Initialize(mux, flowMgtService, applicationService, execRegistry, observabilitySvc)

	// Initialize OAuth services.
	oauth.Initialize(mux, applicationService, userService, jwtService, flowExecService)

	// TODO: Legacy way of initializing services. These need to be refactored in the future aligning to the
	// dependency injection pattern used above.

	// Register the health service.
	services.NewHealthCheckService(mux)
}
