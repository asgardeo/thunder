# Flow Execution

This guide covers executing flows and understanding response modes.

## Flow Execution API

The flow execution API processes a given flow step by step.

### Initial Request

Start a new flow by providing the application ID and flow type:

#### Request Fields

| Field | Required | Description |
|-------|----------|-------------|
| `applicationId` | Yes | UUID of the application |
| `flowType` | Yes | Either `AUTHENTICATION` or `REGISTRATION` |
| `verbose` | No | Enable verbose mode for full UI metadata (default: false) |
| `inputs` | No | Initial inputs (e.g., `requested_permissions`) |

#### Example

```bash
curl -X POST https://localhost:8090/flow/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "applicationId": "<app-uuid>",
    "flowType": "AUTHENTICATION"
  }'
```

### Subsequent Requests

Continue the flow by providing the flow ID, selected action, and user inputs:

#### Request Fields

| Field | Required | Description |
|-------|----------|-------------|
| `flowId` | Yes | Flow session ID from previous response |
| `action` | No | Reference to the selected action |
| `inputs` | No | User-provided input values |

#### Example

```bash
curl -X POST https://localhost:8090/flow/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "flowId": "<flow-uuid>",
    "action": "action_001",
    "inputs": {
      "username": "user@example.com",
      "password": "secret123"
    }
  }'
```

---

## Verbose vs Non-Verbose Mode

| Mode | Description | Use Case |
|------|-------------|----------|
| Non-Verbose | Returns only logical data (inputs, actions, redirects) | Custom UI, mobile apps, SPAs |
| Verbose | Includes full UI metadata (`meta` object) | Server-rendered UI, Thunder login SDK |

**Non-verbose is the default.** To enable verbose mode, set `verbose: true` in the initial request:

```bash
curl -X POST https://localhost:8090/flow/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "applicationId": "<app-uuid>",
    "flowType": "AUTHENTICATION",
    "verbose": true
  }'
```

### Non-Verbose Response

```json
{
  "flowId": "abc-123",
  "flowStatus": "PROMPT_ONLY",
  "stepId": "node_001",
  "type": "VIEW",
  "data": {
    "inputs": [
      { 
        "identifier": "username", 
        "type": "TEXT_INPUT", 
        "required": true 
      },
      { 
        "identifier": "password", 
        "type": "PASSWORD_INPUT", 
        "required": true 
      }
    ],
    "actions": [
      { 
        "ref": "action_001", 
        "nextNode": "basic_auth" 
      }
    ]
  }
}
```

### Verbose Response

```json
{
  "flowId": "abc-123",
  "flowStatus": "PROMPT_ONLY",
  "stepId": "node_001",
  "type": "VIEW",
  "data": {
    "inputs": [...],
    "actions": [...],
    "meta": {
      "components": [
        { "type": "TEXT", "label": "Sign In", "variant": "HEADING_1" },
        { "type": "TEXT_INPUT", "id": "input_001", "label": "Username" },
        { "type": "PASSWORD_INPUT", "id": "input_002", "label": "Password" },
        { "type": "ACTION", "id": "action_001", "label": "Submit", "variant": "PRIMARY" }
      ]
    }
  }
}
```

---

## Response Fields

| Field | Description |
|-------|-------------|
| `flowId` | Session identifier to use in subsequent requests |
| `flowStatus` | Current flow state (see Flow Status below) |
| `stepId` | Current node/step identifier |
| `type` | Response type (VIEW or REDIRECTION) |
| `data.inputs` | Required inputs the client should collect |
| `data.actions` | Available actions the user can take |
| `data.meta` | UI metadata (verbose mode only) |
| `data.redirectURL` | External redirect URL (for REDIRECTION type) |
| `assertion` | JWT token (only when flowStatus is COMPLETE) |

---

## Flow Status

| Status | Description |
|--------|-------------|
| PROMPT_ONLY | Flow requires user interaction |
| COMPLETE | Flow completed successfully |
| ERROR | Flow encountered an error |

---

## Response Types

| Type | Description |
|------|-------------|
| VIEW | Display a UI prompt |
| REDIRECTION | Redirect to external URL (e.g., social login) |

---

## Completed Flow

On successful completion, the response includes a JWT assertion:

```json
{
  "flowId": "abc-123",
  "flowStatus": "COMPLETE",
  "assertion": "<jwt-token>"
}
```

---

## API Reference

See the [Flow Execution API](/api/flow-execution.yaml) for the complete API specification.
