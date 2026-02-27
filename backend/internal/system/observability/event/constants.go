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
	// Token Issuance Events

	// EventTypeTokenIssuanceStarted is triggered when token issuance begins.
	EventTypeTokenIssuanceStarted EventType = "TOKEN_ISSUANCE_STARTED" //nolint:gosec

	// EventTypeTokenIssued is triggered when a token is successfully issued.
	EventTypeTokenIssued EventType = "TOKEN_ISSUED"

	// EventTypeTokenIssuanceFailed is triggered when token issuance fails.
	EventTypeTokenIssuanceFailed EventType = "TOKEN_ISSUANCE_FAILED" //nolint:gosec

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
