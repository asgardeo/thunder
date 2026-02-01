# Observability Event Reference

This document defines the contract for observability events published by Thunder. It serves as a reference for understanding the data available in logs and traces.

## Event Structure

Every event published by Thunder follows a standard schema (`event.Event`).

| Field | Type | Description |
|-------|------|-------------|
| `trace_id` | `string` (UUID) | **Required**. The unique ID linking all events in a single request transaction. |
| `event_id` | `string` (UUID) | **Required**. Unique ID for this specific event instance. |
| `type` | `string` | **Required**. The semantic name of the event (e.g., `TOKEN_ISSUED`). |
| `timestamp` | `string` (ISO8601) | **Required**. When the event occurred. |
| `component` | `string` | **Required**. Source component (e.g., `FlowEngine`, `AuthHandler`). |
| `status` | `string` | Outcome status: `success`, `failure`, `in_progress`, `pending`. |
| `data` | `map[string]any` | Context-specific metadata (see event definitions below). |

## Event Definitions

Events are grouped by their functional domain.

### Authentication & Token Events

These events track the lifecycle of token issuance and general authentication operations.

**Component**: `AuthHandler`

#### `TOKEN_ISSUANCE_STARTED`
Triggered when a request for a token is received and processing begins.

- **Status**: `in_progress`
- **Data Keys**:
  - `client_id`: The ID of the client requesting the token.
  - `grant_type`: The OAuth 2.0 grant type (e.g., `authorization_code`).
  - `scope`: The requested scopes (space-separated).

#### `TOKEN_ISSUED`
Triggered when a token is successfully generated and returned to the client.

- **Status**: `success`
- **Data Keys**:
  - `client_id`: The ID of the client.
  - `user_id`: The ID of the authenticated user (if applicable).
  - `grant_type`: The OAuth 2.0 grant type.
  - `scope`: The granted scopes.
  - `latency_us`: Processing time in microseconds.

#### `TOKEN_ISSUANCE_FAILED`
Triggered when token issuance fails (e.g., invalid grant, system error).

- **Status**: `failure`
- **Data Keys**:
  - `client_id`: The ID of the client.
  - `error`: Technical error message.
  - `error_code`: Error code (e.g., `invalid_grant`).
  - `failure_reason`: Human-readable reason for failure.
  - `latency_us`: Processing time until failure.

---

### Flow Execution Events

These events track the execution of authentication flows (e.g., login flows, registration flows) driven by the Flow Engine.

**Component**: `FlowEngine`

#### `FLOW_STARTED`
Triggered when a new flow execution session is initialized.

- **Status**: `in_progress`
- **Data Keys**:
  - `flow_id`: Unique ID for the flow session.
  - `flow_type`: Type of flow (e.g., `login`, `register`).
  - `client_id`: The client application initiating the flow.

#### `FLOW_NODE_EXECUTION_STARTED`
Triggered when the engine begins executing a specific step (node) in the flow.

- **Status**: `in_progress`
- **Data Keys**:
  - `flow_id`: The flow session ID.
  - `node_id`: ID of the node being executed.
  - `node_type`: Type of the node (e.g., `authenticator`, `conditional`).
  - `step_number`: The step index in the flow.
  - `trace_parent`: (Optional) ID of the parent span for hierarchical tracing.

#### `FLOW_NODE_EXECUTION_COMPLETED`
Triggered when a node successfully completes its execution.

- **Status**: `success`
- **Data Keys**:
  - `flow_id`: The flow session ID.
  - `node_id`: ID of the node.
  - `duration_ms`: Execution time in milliseconds.

#### `FLOW_NODE_EXECUTION_FAILED`
Triggered when a node execution encounters an error.

- **Status**: `failure`
- **Data Keys**:
  - `flow_id`: The flow session ID.
  - `node_id`: ID of the node.
  - `error`: Error message.
  - `failure_reason`: Reason for failure.

#### `FLOW_USER_INPUT_REQUIRED`
Triggered when the flow pauses to wait for user interaction (e.g., displaying a login page).

- **Status**: `pending`
- **Data Keys**:
  - `flow_id`: The flow session ID.
  - `redirect_to`: The URL or path where the user is being redirected (if applicable).
  - `step_number`: Current step number.

#### `FLOW_COMPLETED`
Triggered when the entire flow executes successfully.

- **Status**: `success`
- **Data Keys**:
  - `flow_id`: The flow session ID.
  - `user_id`: The authenticated user ID (if flow resulted in authentication).
  - `duration_ms`: Total flow duration.

#### `FLOW_FAILED`
Triggered when the flow terminates due to an unrecoverable error.

- **Status**: `failure`
- **Data Keys**:
  - `flow_id`: The flow session ID.
  - `error`: Error details.
  - `failure_reason`: Reason for flow failure.

## Data Dictionary

Common data keys used across events (defined in `event.DataKey`).

| Key | JSON Field | Description |
|-----|------------|-------------|
| `UserID` | `user_id` | Unique identifier for the user. |
| `ClientID` | `client_id` | Unique identifier for the OAuth client/app. |
| `FlowID` | `flow_id` | Correlation ID for a specific flow execution session. |
| `TraceParent` | `trace_parent` | Used for linking spans in distributed tracing. |
| `Error` | `error` | Technical error description or stack trace. |
| `FailureReason` | `failure_reason` | Functional reason for a failure. |
| `LatencyUs` | `latency_us` | Duration in microseconds. |
| `DurationMs` | `duration_ms` | Duration in milliseconds. |
