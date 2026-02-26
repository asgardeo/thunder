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

package granthandlers

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/user"
)

// Initialize initializes the grant handler provider with the given services.
func Initialize(
	mux *http.ServeMux,
	jwtService jwt.JWTServiceInterface,
	userService user.UserServiceInterface,
	applicationService application.ApplicationServiceInterface,
	flowExecService flowexec.FlowExecServiceInterface,
	tokenBuilder tokenservice.TokenBuilderInterface,
	tokenValidator tokenservice.TokenValidatorInterface,
) GrantHandlerProviderInterface {
	authzService := authz.Initialize(mux, applicationService, jwtService, flowExecService)
	grantHandlerProvider := newGrantHandlerProvider(
		jwtService, userService, authzService, tokenBuilder, tokenValidator)
	return grantHandlerProvider
}
