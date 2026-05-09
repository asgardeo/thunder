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
	"github.com/asgardeo/thunder/internal/attributecache"
	authnprovidermgr "github.com/asgardeo/thunder/internal/authnprovider/manager"
	"github.com/asgardeo/thunder/internal/authz"
	"github.com/asgardeo/thunder/internal/entityprovider"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/inboundclient"
	"github.com/asgardeo/thunder/internal/oauth/jwks"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/dcr"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/discovery"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/granthandlers"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/introspect"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/jwksresolver"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/par"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/token"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/userinfo"
	"github.com/asgardeo/thunder/internal/oauth/scope"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/resource"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	syshttp "github.com/asgardeo/thunder/internal/system/http"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/jose/jwe"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/kmprovider/defaultkm/pkiservice"
	"github.com/asgardeo/thunder/internal/system/observability"
)

// Initialize initializes all OAuth-related services and registers their routes.
func Initialize(
	mux *http.ServeMux,
	applicationService application.ApplicationServiceInterface,
	inboundClient inboundclient.InboundClientServiceInterface,
	authnProvider authnprovidermgr.AuthnProviderManagerInterface,
	jwtService jwt.JWTServiceInterface,
	jweService jwe.JWEServiceInterface,
	flowExecService flowexec.FlowExecServiceInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
	pkiService pkiservice.PKIServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	attributeCacheSvc attributecache.AttributeCacheServiceInterface,
	authzService authz.AuthorizationServiceInterface,
	entityProvider entityprovider.EntityProviderInterface,
	resourceService resource.ResourceServiceInterface,
	i18nService i18nmgt.I18nServiceInterface,
	idpService idp.IDPServiceInterface,
) error {
	// Fetch runtime transactioner for OAuth services.
	transactioner, err := provider.GetDBProvider().GetRuntimeDBTransactioner()
	if err != nil {
		return err
	}

	jwks.Initialize(mux, pkiService)
	httpClient := syshttp.NewHTTPClientWithCheckRedirect(func(req *http.Request, _ []*http.Request) error {
		return syshttp.IsSSRFSafeURL(req.URL.String())
	})
	resolver := jwksresolver.Initialize(httpClient)
	tokenBuilder, tokenValidator := tokenservice.Initialize(jwtService, jweService, resolver, idpService)
	scopeValidator := scope.Initialize()
	discoveryService := discovery.Initialize(mux, pkiService)
	parService := par.Initialize(mux, inboundClient, authnProvider, jwtService, discoveryService,
		resourceService)
	grantHandlerProvider, err := granthandlers.Initialize(
		mux, jwtService, inboundClient, flowExecService, tokenBuilder, tokenValidator,
		attributeCacheSvc, ouService, authzService, entityProvider, resourceService, parService)
	if err != nil {
		return err
	}
	token.Initialize(mux, jwtService, inboundClient, authnProvider, grantHandlerProvider,
		scopeValidator, observabilitySvc, discoveryService, transactioner)
	introspect.Initialize(mux, jwtService, inboundClient, authnProvider, discoveryService)
	userinfo.Initialize(mux, jwtService, jweService, resolver,
		tokenValidator, inboundClient, ouService, attributeCacheSvc, transactioner)
	dcr.Initialize(mux, applicationService, ouService, i18nService, transactioner)
	return nil
}
