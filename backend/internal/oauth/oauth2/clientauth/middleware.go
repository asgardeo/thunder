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

package clientauth

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/discovery"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// ClientAuthMiddleware authenticates OAuth2 clients and attaches client info to request context.
func ClientAuthMiddleware(appService application.ApplicationServiceInterface, jwtService jwt.JWTServiceInterface,
	discoveryService discovery.DiscoveryServiceInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Authenticate client
			clientInfo, authErr := authenticate(r, appService, jwtService, discoveryService)
			if authErr != nil {
				// Convert headers to the format expected by WriteJSONError
				var respHeaders []map[string]string
				if len(authErr.ResponseHeaders) > 0 {
					respHeaders = []map[string]string{authErr.ResponseHeaders}
				}
				// Write error response
				utils.WriteJSONError(
					w,
					authErr.ErrorCode,
					authErr.ErrorDescription,
					authErr.StatusCode,
					respHeaders,
				)
				return
			}

			// Attach client info to context
			ctx := withOAuthClient(r.Context(), clientInfo)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
