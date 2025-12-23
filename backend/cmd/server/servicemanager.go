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
	brandingmgt "github.com/asgardeo/thunder/internal/branding/mgt"
	brandingresolve "github.com/asgardeo/thunder/internal/branding/resolve"
	"github.com/asgardeo/thunder/internal/cert"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/asgardeo/thunder/internal/group"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/notification"
	"github.com/asgardeo/thunder/internal/oauth"
	"github.com/asgardeo/thunder/internal/observability"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/resource"
	"github.com/asgardeo/thunder/internal/role"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/export"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	immutableresource "github.com/asgardeo/thunder/internal/system/immutable_resource"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/services"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/internal/userschema"
)

// observabilitySvc is the observability service instance. This is used for graceful shutdown.
var observabilitySvc observability.ObservabilityServiceInterface

// registerServices registers all the services with the provided HTTP multiplexer.
func registerServices(
	mux *http.ServeMux,
	jwtService jwt.JWTServiceInterface,
) {
	logger := log.GetLogger()

	observabilitySvc = observability.Initialize()

	// List to collect exporters from each package
	var exporters []immutableresource.ResourceExporter

	// Initialize i18n service for internationalization support.
	_, i18nExporter, err := i18nmgt.Initialize(mux)
	if err != nil {
		logger.Fatal("Failed to initialize i18n service", log.Error(err))
	}
	// Add to exporters list (must be done after initializing list)
	exporters = append(exporters, i18nExporter)

	ouService, ouExporter, err := ou.Initialize(mux)
	if err != nil {
		logger.Fatal("Failed to initialize OrganizationUnitService", log.Error(err))
	}
	exporters = append(exporters, ouExporter)

	hashService := hash.Initialize()
	userSchemaService, userSchemaExporter, err := userschema.Initialize(mux, ouService)
	if err != nil {
		logger.Fatal("Failed to initialize UserSchemaService", log.Error(err))
	}
	exporters = append(exporters, userSchemaExporter)

	userService, err := user.Initialize(mux, ouService, userSchemaService, hashService)
	if err != nil {
		logger.Fatal("Failed to initialize UserService", log.Error(err))
	}
	groupService := group.Initialize(mux, ouService, userService)

	resourceService, err := resource.Initialize(mux, ouService)
	if err != nil {
		logger.Fatal("Failed to initialize Resource Service", log.Error(err))
	}
	roleService := role.Initialize(mux, userService, groupService, ouService, resourceService)
	authZService := authz.Initialize(roleService)

	idpService, idpExporter, err := idp.Initialize(mux)
	if err != nil {
		logger.Fatal("Failed to initialize IDPService", log.Error(err))
	}
	exporters = append(exporters, idpExporter)

	_, otpService, notificationExporter, err := notification.Initialize(mux, jwtService)
	if err != nil {
		logger.Fatal("Failed to initialize NotificationService", log.Error(err))
	}
	exporters = append(exporters, notificationExporter)

	// Initialize authentication services.
	_, authSvcRegistry := authn.Initialize(mux, idpService, jwtService, userService, otpService)

	// Initialize flow and executor services.
	flowFactory, graphCache := flowcore.Initialize()
	execRegistry := executor.Initialize(flowFactory, userService, ouService,
		idpService, otpService, jwtService, authSvcRegistry, authZService, userSchemaService, observabilitySvc,
		groupService, roleService)

	flowMgtService, flowMgtExporter, err := flowmgt.Initialize(mux, flowFactory, execRegistry, graphCache)
	if err != nil {
		logger.Fatal("Failed to initialize FlowMgtService", log.Error(err))
	}
	exporters = append(exporters, flowMgtExporter)
	certservice := cert.Initialize()
	brandingMgtService := brandingmgt.Initialize(mux)
	applicationService, applicationExporter, err := application.Initialize(mux, certservice, flowMgtService,
		brandingMgtService, userSchemaService)
	if err != nil {
		logger.Fatal("Failed to initialize ApplicationService", log.Error(err))
	}
	exporters = append(exporters, applicationExporter)

	_ = brandingresolve.Initialize(mux, brandingMgtService, applicationService)

	// Initialize export service with collected exporters
	_ = export.Initialize(mux, exporters)

	flowExecService := flowexec.Initialize(mux, flowMgtService, applicationService, execRegistry, observabilitySvc)

	// Initialize OAuth services.
	oauth.Initialize(mux, applicationService, userService, jwtService, flowExecService, observabilitySvc)

	// TODO: Legacy way of initializing services. These need to be refactored in the future aligning to the
	// dependency injection pattern used above.

	// Register the health service.
	services.NewHealthCheckService(mux)
}

// unregisterServices unregisters all services that require cleanup during shutdown.
func unregisterServices() {
	observabilitySvc.Shutdown()
}
