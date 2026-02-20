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
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/authz"
	"github.com/asgardeo/thunder/internal/cert"
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	"github.com/asgardeo/thunder/internal/design/resolve"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	"github.com/asgardeo/thunder/internal/flow/flowmeta"
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
	"github.com/asgardeo/thunder/internal/system/crypto/pki"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/export"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/jose"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/mcp"
	"github.com/asgardeo/thunder/internal/system/services"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/internal/userschema"
)

// observabilitySvc is the observability service instance. This is used for graceful shutdown.
var observabilitySvc observability.ObservabilityServiceInterface

// registerServices registers all the services with the provided HTTP multiplexer.
func registerServices(mux *http.ServeMux) jwt.JWTServiceInterface {
	logger := log.GetLogger()

	// Load the server's private key for signing JWTs.
	pkiService, err := pki.Initialize()
	if err != nil {
		logger.Fatal("Failed to initialize certificate service", log.Error(err))
	}

	jwtService, _, err := jose.Initialize(pkiService)
	if err != nil {
		logger.Fatal("Failed to initialize JOSE services", log.Error(err))
	}

	observabilitySvc = observability.Initialize()

	// List to collect exporters from each package
	var exporters []declarativeresource.ResourceExporter

	// Initialize i18n service for internationalization support.
	i18nService, i18nExporter, err := i18nmgt.Initialize(mux)
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
	groupService, err := group.Initialize(mux, ouService, userService)
	if err != nil {
		logger.Fatal("Failed to initialize GroupService", log.Error(err))
	}

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

	// Initialize MCP server
	mcpServer := mcp.Initialize(mux, jwtService)

	// Initialize authn provider
	authnProvider := authnprovider.InitializeAuthnProvider(userService)

	// Initialize user provider based on configuration
	userProvider := userprovider.InitializeUserProvider(userService)

	// Initialize authentication services.
	_, authSvcRegistry := authn.Initialize(
		mux, mcpServer, idpService, jwtService, userService,
		userProvider, otpService, authnProvider,
	)

	// Initialize flow and executor services.
	flowFactory, graphCache := flowcore.Initialize()
	execRegistry := executor.Initialize(flowFactory, ouService,
		idpService, otpService, jwtService, authSvcRegistry, authZService, userSchemaService, observabilitySvc,
		groupService, roleService, userProvider, authnProvider)

	flowMgtService, flowMgtExporter, err := flowmgt.Initialize(mux, mcpServer, flowFactory, execRegistry, graphCache)
	if err != nil {
		logger.Fatal("Failed to initialize FlowMgtService", log.Error(err))
	}
	exporters = append(exporters, flowMgtExporter)
	certservice, err := cert.Initialize()
	if err != nil {
		logger.Fatal("Failed to initialize CertificateService", log.Error(err))
	}

	// Initialize theme and layout services
	themeMgtService, themeExporter, err := thememgt.Initialize(mux)
	if err != nil {
		logger.Fatal("Failed to initialize ThemeMgtService", log.Error(err))
	}
	exporters = append(exporters, themeExporter)

	layoutMgtService, layoutExporter, err := layoutmgt.Initialize(mux)
	if err != nil {
		logger.Fatal("Failed to initialize LayoutMgtService", log.Error(err))
	}
	exporters = append(exporters, layoutExporter)

	applicationService, applicationExporter, err := application.Initialize(
		mux, mcpServer, certservice, flowMgtService, themeMgtService, layoutMgtService, userSchemaService)
	if err != nil {
		logger.Fatal("Failed to initialize ApplicationService", log.Error(err))
	}
	exporters = append(exporters, applicationExporter)

	// Initialize design resolve service for theme and layout resolution
	designResolveService := resolve.Initialize(mux, themeMgtService, layoutMgtService, applicationService)

	// Initialize flow metadata service
	_ = flowmeta.Initialize(mux, applicationService, ouService, designResolveService, i18nService)

	// Initialize export service with collected exporters
	_ = export.Initialize(mux, exporters)

	flowExecService := flowexec.Initialize(mux, flowMgtService, applicationService, execRegistry,
		observabilitySvc)

	// Initialize OAuth services.
	oauth.Initialize(mux, applicationService, userService, jwtService, flowExecService, observabilitySvc,
		pkiService, ouService)

	// TODO: Legacy way of initializing services. These need to be refactored in the future aligning to the
	// dependency injection pattern used above.

	// Register the health service.
	services.NewHealthCheckService(mux)

	return jwtService
}

// unregisterServices unregisters all services that require cleanup during shutdown.
func unregisterServices() {
	observabilitySvc.Shutdown()
}
