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

// Component name constants for event sources.
// These identify which component/module is emitting the event.
const (
	// ComponentFlowEngine identifies events from the flow execution engine.
	ComponentFlowEngine = "FlowEngine"

	// ComponentAuthHandler identifies events from authentication handlers.
	ComponentAuthHandler = "AuthHandler"
)

// Authentication and Authorization Event Types
const (

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
)
