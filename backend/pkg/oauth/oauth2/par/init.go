/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package par

import (
	"context"
	"net/http"

	"github.com/asgardeo/thunder/internal/resource"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/middleware"
	authnprovidermgr "github.com/asgardeo/thunder/pkg/authnprovider/manager"
	"github.com/asgardeo/thunder/pkg/inboundclient"
	"github.com/asgardeo/thunder/pkg/oauth/oauth2/clientauth"
	"github.com/asgardeo/thunder/pkg/oauth/oauth2/discovery"
)

// Initialize initializes the PAR handler and registers its routes.
// Returns the PARServiceInterface so the authorization endpoint can resolve request_uri parameters.
func Initialize(
	mux *http.ServeMux,
	inboundClient inboundclient.InboundClientServiceInterface,
	authnProvider authnprovidermgr.AuthnProviderManagerInterface,
	jwtService jwt.JWTServiceInterface,
	discoveryService discovery.DiscoveryServiceInterface,
	resourceService resource.ResourceServiceInterface,
) PARServiceInterface {
	store := initializePARStore()
	parSvc := newPARService(store, resourceService)
	handler := newPARHandler(parSvc)
	registerRoutes(mux, handler, inboundClient, authnProvider, jwtService, discoveryService)
	return parSvc
}

// initializePARStore selects the PAR store implementation based on the configured runtime DB type.
func initializePARStore() parStoreInterface {
	deploymentID := config.GetServerRuntime().Config.Server.Identifier

	if config.GetServerRuntime().Config.Database.Runtime.Type == provider.DataSourceTypeRedis {
		return newRedisPARRequestStore(provider.GetRedisProvider(), deploymentID)
	}
	return newPARRequestStore(deploymentID)
}

// registerRoutes registers the PAR endpoint route with client authentication middleware.
func registerRoutes(
	mux *http.ServeMux,
	handler parHandlerInterface,
	inboundClient inboundclient.InboundClientServiceInterface,
	authnProvider authnprovidermgr.AuthnProviderManagerInterface,
	jwtService jwt.JWTServiceInterface,
	discoveryService discovery.DiscoveryServiceInterface,
) {
	corsOpts := middleware.CORSOptions{
		AllowedMethods:   []string{"POST"},
		AllowedHeaders:   middleware.DefaultAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           600,
	}

	metadata := discoveryService.GetOAuth2AuthorizationServerMetadata(context.Background())
	endpointURL := metadata.PushedAuthorizationRequestEndpoint
	clientAuthMiddleware := clientauth.ClientAuthMiddleware(inboundClient, authnProvider, jwtService, endpointURL)
	wrappedHandler := clientAuthMiddleware(http.HandlerFunc(handler.HandlePARRequest))

	pattern, corsHandler := middleware.WithCORS(
		"POST /oauth2/par",
		wrappedHandler.ServeHTTP,
		corsOpts,
	)

	mux.HandleFunc(pattern, corsHandler)
}
