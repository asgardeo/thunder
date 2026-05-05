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

// Package app provides the application builder (orchestrator) for ThunderID.
//
// External consumers can construct a fully wired ThunderID application by injecting
// custom implementations of the core extension-point interfaces:
//
//	app, err := appbuilder.New().
//	    WithEntityProvider(myEntityProvider).
//	    WithAuthnProvider(myAuthnProvider).
//	    WithConsentService(myConsentService).
//	    WithExecutor("my_executor", myExecutor).
//	    Build(mux)
//
// Dependencies not injected fall back to ThunderID's built-in defaults.
package app

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/agent"
	"github.com/asgardeo/thunder/internal/attributecache"
	"github.com/asgardeo/thunder/internal/authn"
	authnAssert "github.com/asgardeo/thunder/internal/authn/assert"
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authnConsent "github.com/asgardeo/thunder/internal/authn/consent"
	"github.com/asgardeo/thunder/internal/authn/github"
	"github.com/asgardeo/thunder/internal/authn/google"
	authnOAuth "github.com/asgardeo/thunder/internal/authn/oauth"
	authnOIDC "github.com/asgardeo/thunder/internal/authn/oidc"
	"github.com/asgardeo/thunder/internal/authn/otp"
	"github.com/asgardeo/thunder/internal/authn/passkey"
	"github.com/asgardeo/thunder/internal/authz"
	"github.com/asgardeo/thunder/internal/cert"
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	"github.com/asgardeo/thunder/internal/design/resolve"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	"github.com/asgardeo/thunder/internal/entity"
	"github.com/asgardeo/thunder/internal/group"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/notification"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/resource"
	"github.com/asgardeo/thunder/internal/role"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/crypto/pki"
	dbprovider "github.com/asgardeo/thunder/internal/system/database/provider"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/email"
	"github.com/asgardeo/thunder/internal/system/export"
	healthcheckservice "github.com/asgardeo/thunder/internal/system/healthcheck/service"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/importer"
	"github.com/asgardeo/thunder/internal/system/jose"
	"github.com/asgardeo/thunder/internal/system/mcp"
	"github.com/asgardeo/thunder/internal/system/observability"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/internal/system/services"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
	"github.com/asgardeo/thunder/internal/system/template"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/pkg/application"
	authnprovidermgr "github.com/asgardeo/thunder/pkg/authnprovider/manager"
	"github.com/asgardeo/thunder/pkg/consent"
	"github.com/asgardeo/thunder/pkg/entityprovider"
	flowcore "github.com/asgardeo/thunder/pkg/flow/core"
	"github.com/asgardeo/thunder/pkg/flow/executor"
	"github.com/asgardeo/thunder/pkg/flow/flowexec"
	"github.com/asgardeo/thunder/pkg/flow/flowmeta"
	flowmgt "github.com/asgardeo/thunder/pkg/flow/mgt"
	"github.com/asgardeo/thunder/pkg/inboundclient"
	"github.com/asgardeo/thunder/pkg/oauth"
)

// ThunderApp is the wired application returned by Build. It exposes the HTTP mux
// with all routes registered and provides lifecycle management.
type ThunderApp struct {
	mux                *http.ServeMux
	securityMiddleware func(http.Handler) http.Handler
	observabilitySvc   observability.ObservabilityServiceInterface
}

// Mux returns the HTTP multiplexer with all application routes registered.
func (a *ThunderApp) Mux() *http.ServeMux { return a.mux }

// SecurityMiddleware returns the pre-wired security middleware. Callers should wrap their
// handler with it: securityMiddleware(myHandler).
func (a *ThunderApp) SecurityMiddleware() func(http.Handler) http.Handler {
	return a.securityMiddleware
}

// Shutdown tears down long-lived background services (observability, etc.).
func (a *ThunderApp) Shutdown() {
	if a.observabilitySvc != nil {
		a.observabilitySvc.Shutdown()
	}
}

// Builder constructs a ThunderApp with optional dependency overrides.
type Builder struct {
	entityProvider   entityprovider.EntityProviderInterface
	authnProvider    authnprovidermgr.AuthnProviderManagerInterface
	consentService   consent.ConsentServiceInterface
	extraExecutors   []extraExecutor
	observabilitySvc observability.ObservabilityServiceInterface
	emailClient      email.EmailClientInterface
}

type extraExecutor struct {
	name string
	exec flowcore.ExecutorInterface
}

// New returns a Builder with no overrides — all defaults are used.
func New() *Builder {
	return &Builder{}
}

// WithEntityProvider overrides the default entity provider.
func (b *Builder) WithEntityProvider(ep entityprovider.EntityProviderInterface) *Builder {
	b.entityProvider = ep
	return b
}

// WithAuthnProvider overrides the default authentication provider manager.
func (b *Builder) WithAuthnProvider(ap authnprovidermgr.AuthnProviderManagerInterface) *Builder {
	b.authnProvider = ap
	return b
}

// WithConsentService overrides the default consent service.
func (b *Builder) WithConsentService(cs consent.ConsentServiceInterface) *Builder {
	b.consentService = cs
	return b
}

// WithExecutor registers an additional executor into the flow engine's executor registry.
func (b *Builder) WithExecutor(name string, ex flowcore.ExecutorInterface) *Builder {
	b.extraExecutors = append(b.extraExecutors, extraExecutor{name: name, exec: ex})
	return b
}

// WithObservabilityService overrides the default observability service.
func (b *Builder) WithObservabilityService(obs observability.ObservabilityServiceInterface) *Builder {
	b.observabilitySvc = obs
	return b
}

// WithEmailClient overrides the default SMTP email client.
func (b *Builder) WithEmailClient(client email.EmailClientInterface) *Builder {
	b.emailClient = client
	return b
}

// Build wires all ThunderID services and registers HTTP routes on mux.
// If mux is nil, a new http.ServeMux is created.
func (b *Builder) Build(mux *http.ServeMux) (*ThunderApp, error) { //nolint:gocyclo
	if mux == nil {
		mux = http.NewServeMux()
	}

	pkiService, err := pki.Initialize()
	if err != nil {
		return nil, err
	}

	jwtService, jweService, err := jose.Initialize(pkiService)
	if err != nil {
		return nil, err
	}

	observabilitySvc := b.observabilitySvc
	if observabilitySvc == nil {
		observabilitySvc = observability.Initialize()
	}

	var exporters []declarativeresource.ResourceExporter

	i18nService, i18nExporter, err := i18nmgt.Initialize(mux)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, i18nExporter)

	ouAuthzService, err := sysauthz.Initialize()
	if err != nil {
		return nil, err
	}

	ouService, ouHierarchyResolver, ouExporter, err := ou.Initialize(mux, ouAuthzService)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, ouExporter)
	ouAuthzService.SetOUHierarchyResolver(ouHierarchyResolver)

	hashService, err := hash.Initialize()
	if err != nil {
		return nil, err
	}

	consentService := b.consentService
	if consentService == nil {
		consentService = consent.Initialize()
	}

	userSchemaService, userSchemaExporter, err := userschema.Initialize(
		mux, ouService, ouAuthzService, consentService)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, userSchemaExporter)

	entityService, err := entity.Initialize(hashService, userSchemaService, ouService)
	if err != nil {
		return nil, err
	}

	entityProv := b.entityProvider
	if entityProv == nil {
		entityProv = entityprovider.InitializeEntityProvider(entityService)
	}

	userService, ouUserResolver, userExporter, err := user.Initialize(
		mux, entityService, ouService, userSchemaService, ouAuthzService,
	)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, userExporter)

	groupService, ouGroupResolver, err := group.Initialize(
		mux, dbprovider.GetDBProvider(), ouService, entityService, userSchemaService, ouAuthzService,
	)
	if err != nil {
		return nil, err
	}

	ouService.SetOUUserResolver(ouUserResolver)
	ouService.SetOUGroupResolver(ouGroupResolver)

	resourceService, resourceExporter, err := resource.Initialize(mux, ouService)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, resourceExporter)

	roleService, roleExporter, err := role.Initialize(
		mux, entityService, groupService, ouService, resourceService, userSchemaService,
	)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, roleExporter)

	authZService := authz.Initialize(roleService)

	idpService, idpExporter, err := idp.Initialize(mux)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, idpExporter)

	templateService, err := template.Initialize()
	if err != nil {
		return nil, err
	}

	_, otpService, notifSenderSvc, notificationExporter, err := notification.Initialize(
		mux, jwtService, templateService)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, notificationExporter)

	mcpServer := mcp.Initialize(mux, jwtService)

	passkeyService := passkey.Initialize(entityService)
	otpCoreService := otp.Initialize(otpService, entityProv)

	oauthAuthnService := authnOAuth.Initialize(idpService, entityProv)
	oidcAuthnService := authnOIDC.Initialize(oauthAuthnService, jwtService)
	googleAuthnService := google.Initialize(oidcAuthnService, jwtService)
	githubAuthnService := github.Initialize(oauthAuthnService)

	federatedAuths := map[idp.IDPType]authncm.FederatedAuthenticator{
		idp.IDPTypeOAuth:  oauthAuthnService,
		idp.IDPTypeOIDC:   oidcAuthnService,
		idp.IDPTypeGoogle: googleAuthnService,
		idp.IDPTypeGitHub: githubAuthnService,
	}

	authnProv := b.authnProvider
	if authnProv == nil {
		authnProv = authnprovidermgr.InitializeAuthnProviderManager(
			entityService, passkeyService, otpCoreService, federatedAuths)
	}

	authAssertGen := authnAssert.Initialize()
	consentEnforcer := authnConsent.Initialize(consentService, jwtService)

	authn.Initialize(mux, mcpServer, idpService, jwtService, authnProv, authAssertGen, passkeyService,
		otpCoreService, oauthAuthnService, oidcAuthnService, googleAuthnService, githubAuthnService)

	attributeCacheService := attributecache.Initialize()

	flowFactory, graphCache := flowcore.Initialize()

	emailClient := b.emailClient
	if emailClient == nil {
		if emailClient, err = email.Initialize(); err != nil {
			emailClient = nil
		}
	}

	execRegistry := executor.Initialize(flowFactory, ouService, idpService, notifSenderSvc, jwtService,
		authAssertGen, consentEnforcer, authnProv, otpCoreService, passkeyService, authZService,
		userSchemaService, observabilitySvc, groupService, roleService, entityProv, attributeCacheService,
		emailClient, templateService, oauthAuthnService, oidcAuthnService, githubAuthnService, googleAuthnService)

	for _, ex := range b.extraExecutors {
		execRegistry.RegisterExecutor(ex.name, ex.exec)
	}

	flowMgtService, flowMgtExporter, err := flowmgt.Initialize(mux, mcpServer, flowFactory, execRegistry, graphCache)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, flowMgtExporter)

	certservice, err := cert.Initialize(dbprovider.GetDBProvider())
	if err != nil {
		return nil, err
	}

	themeMgtService, themeExporter, err := thememgt.Initialize(mux)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, themeExporter)

	layoutMgtService, layoutExporter, err := layoutmgt.Initialize(mux)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, layoutExporter)

	inboundClientService, err := inboundclient.Initialize(
		certservice, entityProv, themeMgtService, layoutMgtService, flowMgtService,
		userSchemaService, consentService)
	if err != nil {
		return nil, err
	}

	// TODO: Remove entityService dependency after finalizing declarative resource loading pattern
	applicationService, applicationExporter, err := application.Initialize(
		mux, mcpServer, entityProv, entityService, inboundClientService, ouService, i18nService)
	if err != nil {
		return nil, err
	}
	exporters = append(exporters, applicationExporter)

	if _, err := agent.Initialize(mux, entityService, inboundClientService, ouService); err != nil {
		return nil, err
	}

	designResolveService := resolve.Initialize(mux, themeMgtService, layoutMgtService, applicationService)

	_ = flowmeta.Initialize(mux, applicationService, ouService, designResolveService, i18nService)

	_ = export.Initialize(mux, exporters)

	_ = importer.Initialize(
		mux,
		applicationService,
		idpService,
		flowMgtService,
		ouService,
		userSchemaService,
		roleService,
		resourceService,
		themeMgtService,
		layoutMgtService,
		userService,
		i18nService,
	)

	flowExecService, err := flowexec.Initialize(mux, flowMgtService, applicationService, execRegistry,
		observabilitySvc)
	if err != nil {
		return nil, err
	}

	err = oauth.Initialize(mux, applicationService, inboundClientService, authnProv, jwtService, jweService,
		flowExecService, observabilitySvc, pkiService, ouService, attributeCacheService, authZService, entityProv,
		resourceService, i18nService)
	if err != nil {
		return nil, err
	}

	healthSvc := healthcheckservice.Initialize(dbprovider.GetDBProvider(), dbprovider.GetRedisProvider())
	services.NewHealthCheckService(mux, healthSvc)

	secMw, err := security.Initialize(jwtService)
	if err != nil {
		return nil, err
	}

	return &ThunderApp{
		mux:                mux,
		securityMiddleware: secMw,
		observabilitySvc:   observabilitySvc,
	}, nil
}
