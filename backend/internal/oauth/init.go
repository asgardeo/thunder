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
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/authz"
	"github.com/asgardeo/thunder/internal/entityprovider"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	"github.com/asgardeo/thunder/internal/oauth/jwks"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/dcr"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/discovery"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/granthandlers"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/introspect"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/token"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/userinfo"
	"github.com/asgardeo/thunder/internal/oauth/scope"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/crypto/pki"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/observability"
)

// Initialize initializes all OAuth-related services and registers their routes.
func Initialize(
	mux *http.ServeMux,
	applicationService application.ApplicationServiceInterface,
	authnProvider authnprovider.AuthnProviderInterface,
	jwtService jwt.JWTServiceInterface,
	flowExecService flowexec.FlowExecServiceInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
	pkiService pki.PKIServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	attributeCacheSvc attributecache.AttributeCacheServiceInterface,
	authzService authz.AuthorizationServiceInterface,
	entityProvider entityprovider.EntityProviderInterface,
	i18nService i18nmgt.I18nServiceInterface,
) error {
	// Fetch runtime transactioner for OAuth services.
	transactioner, err := provider.GetDBProvider().GetRuntimeDBTransactioner()
	if err != nil {
		return err
	}

	jwks.Initialize(mux, pkiService)
	tokenBuilder, tokenValidator := tokenservice.Initialize(jwtService)
	grantHandlerProvider, err := granthandlers.Initialize(
		mux, jwtService, applicationService, flowExecService, tokenBuilder, tokenValidator,
		attributeCacheSvc, ouService, authzService, entityProvider)
	if err != nil {
		return err
	}
	scopeValidator := scope.Initialize()
	discoveryService := discovery.Initialize(mux)
	token.Initialize(mux, jwtService, applicationService, authnProvider, grantHandlerProvider,
		scopeValidator, observabilitySvc, discoveryService, transactioner)
	introspect.Initialize(mux, jwtService, applicationService, authnProvider, discoveryService)
	userinfo.Initialize(mux, jwtService, tokenValidator, applicationService, ouService, attributeCacheSvc,
		transactioner)
	dcr.Initialize(mux, applicationService, ouService, i18nService, transactioner)
	return nil
}
