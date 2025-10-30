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

package event

// Authentication and Authorization Event Types
const (
	// Authorization Flow Events

	// EventTypeAuthorizationStarted is triggered when OAuth authorization request is received.
	EventTypeAuthorizationStarted EventType = "AUTHORIZATION_STARTED"

	// EventTypeAuthorizationValidated is triggered after authorization request validation.
	EventTypeAuthorizationValidated EventType = "AUTHORIZATION_VALIDATED"

	// EventTypeAuthorizationRedirect is triggered when redirecting to authentication.
	EventTypeAuthorizationRedirect EventType = "AUTHORIZATION_REDIRECT"

	// EventTypeAuthorizationCodeGenerated is triggered when authorization code is created.
	EventTypeAuthorizationCodeGenerated EventType = "AUTHORIZATION_CODE_GENERATED"

	// EventTypeAuthorizationCompleted is triggered when authorization flow completes.
	EventTypeAuthorizationCompleted EventType = "AUTHORIZATION_COMPLETED"

	// EventTypeAuthorizationFailed is triggered when authorization fails.
	EventTypeAuthorizationFailed EventType = "AUTHORIZATION_FAILED"

	// Authentication Flow Events

	// EventTypeAuthenticationStarted is triggered when authentication flow begins.
	EventTypeAuthenticationStarted EventType = "AUTHENTICATION_STARTED"

	// EventTypeAuthenticationMethodSelected is triggered when user selects auth method.
	EventTypeAuthenticationMethodSelected EventType = "AUTHENTICATION_METHOD_SELECTED"

	// EventTypeCredentialsAuthStarted is triggered when credentials authentication begins.
	EventTypeCredentialsAuthStarted EventType = "CREDENTIALS_AUTH_STARTED"

	// EventTypeCredentialsAuthCompleted is triggered when credentials authentication succeeds.
	EventTypeCredentialsAuthCompleted EventType = "CREDENTIALS_AUTH_COMPLETED"

	// EventTypeCredentialsAuthFailed is triggered when credentials authentication fails.
	EventTypeCredentialsAuthFailed EventType = "CREDENTIALS_AUTH_FAILED"

	// EventTypeOTPSent is triggered when OTP is sent to user.
	EventTypeOTPSent EventType = "OTP_SENT"

	// EventTypeOTPVerificationStarted is triggered when OTP verification begins.
	EventTypeOTPVerificationStarted EventType = "OTP_VERIFICATION_STARTED"

	// EventTypeOTPVerified is triggered when OTP is successfully verified.
	EventTypeOTPVerified EventType = "OTP_VERIFIED"

	// EventTypeOTPVerificationFailed is triggered when OTP verification fails.
	EventTypeOTPVerificationFailed EventType = "OTP_VERIFICATION_FAILED"

	// EventTypeSocialAuthStarted is triggered when social authentication begins.
	EventTypeSocialAuthStarted EventType = "SOCIAL_AUTH_STARTED"

	// EventTypeSocialAuthCallbackReceived is triggered when social provider calls back.
	EventTypeSocialAuthCallbackReceived EventType = "SOCIAL_AUTH_CALLBACK_RECEIVED"

	// EventTypeSocialAuthCompleted is triggered when social authentication succeeds.
	EventTypeSocialAuthCompleted EventType = "SOCIAL_AUTH_COMPLETED"

	// EventTypeSocialAuthFailed is triggered when social authentication fails.
	EventTypeSocialAuthFailed EventType = "SOCIAL_AUTH_FAILED"

	// EventTypeAuthenticationCompleted is triggered when authentication flow succeeds.
	EventTypeAuthenticationCompleted EventType = "AUTHENTICATION_COMPLETED"

	// EventTypeAuthenticationFailed is triggered when authentication flow fails.
	EventTypeAuthenticationFailed EventType = "AUTHENTICATION_FAILED"

	// Token Flow Events

	// EventTypeTokenRequestReceived is triggered when token endpoint receives request.
	EventTypeTokenRequestReceived EventType = "TOKEN_REQUEST_RECEIVED"

	// EventTypeTokenRequestValidated is triggered after token request validation.
	EventTypeTokenRequestValidated EventType = "TOKEN_REQUEST_VALIDATED"

	// EventTypeAuthorizationCodeValidated is triggered when auth code is validated.
	EventTypeAuthorizationCodeValidated EventType = "AUTHORIZATION_CODE_VALIDATED"

	// EventTypePKCEValidated is triggered when PKCE validation succeeds.
	EventTypePKCEValidated EventType = "PKCE_VALIDATED"

	// EventTypePKCEFailed is triggered when PKCE validation fails.
	EventTypePKCEFailed EventType = "PKCE_FAILED"

	// EventTypeAccessTokenGenerated is triggered when access token is created.
	EventTypeAccessTokenGenerated EventType = "ACCESS_TOKEN_GENERATED"

	// EventTypeIDTokenGenerated is triggered when ID token is created.
	EventTypeIDTokenGenerated EventType = "ID_TOKEN_GENERATED" // #nosec G101 -- Not a credential, event type name

	// EventTypeRefreshTokenGenerated is triggered when refresh token is created.
	// #nosec G101 -- Not a credential, event type name
	EventTypeRefreshTokenGenerated EventType = "REFRESH_TOKEN_GENERATED"

	// EventTypeTokenIssued is triggered when tokens are successfully issued.
	EventTypeTokenIssued EventType = "TOKEN_ISSUED"

	// EventTypeTokenRequestFailed is triggered when token request fails.
	EventTypeTokenRequestFailed EventType = "TOKEN_REQUEST_FAILED"

	// EventTypeRefreshTokenUsed is triggered when refresh token is used.
	EventTypeRefreshTokenUsed EventType = "REFRESH_TOKEN_USED"

	// Flow Execution Events

	// EventTypeFlowStarted is triggered when a flow execution begins.
	EventTypeFlowStarted EventType = "FLOW_STARTED"

	// EventTypeFlowNodeExecutionStarted is triggered when a flow node execution begins.
	EventTypeFlowNodeExecutionStarted EventType = "FLOW_NODE_EXECUTION_STARTED"

	// EventTypeFlowNodeExecutionCompleted is triggered when a flow node completes.
	EventTypeFlowNodeExecutionCompleted EventType = "FLOW_NODE_EXECUTION_COMPLETED"

	// EventTypeFlowNodeExecutionFailed is triggered when a flow node fails.
	EventTypeFlowNodeExecutionFailed EventType = "FLOW_NODE_EXECUTION_FAILED"

	// EventTypeFlowUserInputRequired is triggered when flow requires user input.
	EventTypeFlowUserInputRequired EventType = "FLOW_USER_INPUT_REQUIRED"

	// EventTypeFlowCompleted is triggered when flow execution succeeds.
	EventTypeFlowCompleted EventType = "FLOW_COMPLETED"

	// EventTypeFlowFailed is triggered when flow execution fails.
	EventTypeFlowFailed EventType = "FLOW_FAILED"

	// Registration Events

	// EventTypeRegistrationStarted is triggered when user registration begins.
	EventTypeRegistrationStarted EventType = "REGISTRATION_STARTED"

	// EventTypeUserProvisioned is triggered when user account is created.
	EventTypeUserProvisioned EventType = "USER_PROVISIONED"

	// EventTypeRegistrationCompleted is triggered when registration succeeds.
	EventTypeRegistrationCompleted EventType = "REGISTRATION_COMPLETED"

	// EventTypeRegistrationFailed is triggered when registration fails.
	EventTypeRegistrationFailed EventType = "REGISTRATION_FAILED"

	// Session Events

	// EventTypeSessionCreated is triggered when a new session is created.
	EventTypeSessionCreated EventType = "SESSION_CREATED"

	// EventTypeSessionUpdated is triggered when session is updated.
	EventTypeSessionUpdated EventType = "SESSION_UPDATED"

	// EventTypeSessionExpired is triggered when session expires.
	EventTypeSessionExpired EventType = "SESSION_EXPIRED"

	// EventTypeSessionDestroyed is triggered when session is explicitly terminated.
	EventTypeSessionDestroyed EventType = "SESSION_DESTROYED"
)

// Component names for event sources
const (
	ComponentAuthorizationHandler  = "AuthorizationHandler"
	ComponentTokenHandler          = "TokenHandler"
	ComponentFlowExecutionService  = "FlowExecutionService"
	ComponentFlowEngine            = "FlowEngine"
	ComponentAuthenticationService = "AuthenticationService"
	ComponentBasicAuthExecutor     = "BasicAuthExecutor"
	ComponentOTPAuthExecutor       = "OTPAuthExecutor"
	ComponentGoogleAuthExecutor    = "GoogleAuthExecutor"
	ComponentGithubAuthExecutor    = "GithubAuthExecutor"
	ComponentProvisioningExecutor  = "ProvisioningExecutor"
	ComponentGrantHandler          = "GrantHandler"
	ComponentSessionManager        = "SessionManager"
)
