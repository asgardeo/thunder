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

package executor

import (
	"github.com/asgardeo/thunder/internal/authn"
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/authz"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/group"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/notification"
	"github.com/asgardeo/thunder/internal/observability"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/role"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/internal/userschema"
)

// Initialize registers available executors and returns the executor registry.
func Initialize(
	flowFactory core.FlowFactoryInterface,
	ouService ou.OrganizationUnitServiceInterface,
	idpService idp.IDPServiceInterface,
	otpService notification.OTPServiceInterface,
	jwtService jwt.JWTServiceInterface,
	authRegistry *authn.AuthServiceRegistry,
	authZService authz.AuthorizationServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
	groupService group.GroupServiceInterface,
	roleService role.RoleServiceInterface,
	userProvider userprovider.UserProviderInterface,
	authnProvider authnprovider.AuthnProviderInterface,
) ExecutorRegistryInterface {
	reg := newExecutorRegistry()
	reg.RegisterExecutor(ExecutorNameBasicAuth, newBasicAuthExecutor(
		flowFactory, userProvider, authRegistry.CredentialsAuthnService, observabilitySvc))
	reg.RegisterExecutor(ExecutorNameSMSAuth, newSMSOTPAuthExecutor(
		flowFactory, otpService, observabilitySvc, userProvider))
	reg.RegisterExecutor(ExecutorNamePasskeyAuth, newPasskeyAuthExecutor(
		flowFactory, authRegistry.PasskeyService, observabilitySvc, userProvider))

	reg.RegisterExecutor(ExecutorNameOAuth, newOAuthExecutor(
		"", []common.Input{}, []common.Input{}, flowFactory, idpService, userSchemaService,
		authRegistry.OAuthAuthnService))
	reg.RegisterExecutor(ExecutorNameOIDCAuth, newOIDCAuthExecutor(
		"", []common.Input{}, []common.Input{}, flowFactory, idpService, userSchemaService,
		authRegistry.OIDCAuthnService))
	reg.RegisterExecutor(ExecutorNameGitHubAuth, newGithubOAuthExecutor(
		flowFactory, idpService, userSchemaService, authRegistry.GithubOAuthAuthnService))
	reg.RegisterExecutor(ExecutorNameGoogleAuth, newGoogleOIDCAuthExecutor(
		flowFactory, idpService, userSchemaService, authRegistry.GoogleOIDCAuthnService))

	reg.RegisterExecutor(ExecutorNameProvisioning, newProvisioningExecutor(flowFactory,
		groupService, roleService, userProvider))
	reg.RegisterExecutor(ExecutorNameOUCreation, newOUExecutor(flowFactory, ouService))

	reg.RegisterExecutor(ExecutorNameAttributeCollect, newAttributeCollector(flowFactory, userProvider))
	reg.RegisterExecutor(ExecutorNameAuthAssert, newAuthAssertExecutor(flowFactory, jwtService,
		ouService, authRegistry.AuthAssertGenerator, authnProvider, userProvider))
	reg.RegisterExecutor(ExecutorNameAuthorization, newAuthorizationExecutor(flowFactory, authZService))
	reg.RegisterExecutor(ExecutorNameHTTPRequest, newHTTPRequestExecutor(flowFactory))
	reg.RegisterExecutor(ExecutorNameUserTypeResolver, newUserTypeResolver(flowFactory, userSchemaService))
	reg.RegisterExecutor(ExecutorNameInviteExecutor, newInviteExecutor(flowFactory))
	reg.RegisterExecutor(ExecutorNameCredentialSetter, newCredentialSetter(flowFactory, userProvider))
	reg.RegisterExecutor(ExecutorNamePermissionValidator, newPermissionValidator(flowFactory))
	reg.RegisterExecutor(ExecutorNameIdentifying, newIdentifyingExecutor(
		"", []common.Input{{Identifier: userAttributeUsername, Type: "string", Required: true}}, []common.Input{},
		flowFactory, userProvider))

	return reg
}
