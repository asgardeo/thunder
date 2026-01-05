# Customizing Authentication Flows

This guide explains how to customize authentication flows in WSO2 Thunder using the flow graph system. Thunder uses a flexible graph-based approach to define authentication flows, allowing you to create tailored login experiences.

## Understanding Authentication Flows

In Thunder, authentication flows are represented as directed graphs. Each graph consists of:

- **Nodes**: Individual steps in the authentication process
- **Edges**: Connections between nodes that define the flow path
- **Executors**: Components that perform specific authentication tasks

Flows are defined in JSON files and stored in the graph directory configured in your deployment.

## Preconfigured Authentication Flows

Thunder comes with several preconfigured authentication flows:

| Flow ID | Description |
|---------|-------------|
| `auth_flow_config_basic` | Username and password authentication |
| `auth_flow_config_google` | Authentication with Google |
| `auth_flow_config_github` | Authentication with GitHub |
| `auth_flow_config_sms` | SMS OTP authentication with mobile number |
| `auth_flow_config_sms_with_username` | SMS OTP authentication with username |
| `auth_flow_config_basic_with_prompt` | Username and password authentication with prompt nodes |
| `auth_flow_config_basic_google` | Authentication with username/password or Google |
| `auth_flow_config_basic_google_github` | Authentication with username/password, Google, or GitHub |
| `auth_flow_config_basic_google_github_sms` | Authentication with username/password, Google, GitHub, or SMS OTP |

These graphs are located in the `<THUNDER_HOME>/backend/cmd/server/repository/resources/graphs` directory.

## Authentication Flow Components

### Node Types

Thunder supports the following node types:

| Node Type | Constant | Description |
|-----------|----------|-------------|
| Authentication Success | `AUTHENTICATION_SUCCESS` | Finalizes the authentication process and generates the assertion |
| Task Execution | `TASK_EXECUTION` | Performs a specific authentication task using an executor |
| Prompt Only | `PROMPT_ONLY` | Prompts for user input without performing any tasks |
| Decision | `DECISION` | Implements conditional logic in the flow |

### Executors

Thunder provides several executor implementations:

| Executor | Description | Input Parameters | Properties |
|----------|-------------|------------------|----------|
| `BasicAuthExecutor` | Authenticates users with username and password | `username`, `password` | |
| `GoogleOAuthExecutor` | Authenticates users with Google | `code` (authorization code) | `idpId` (Identity Provider id) |
| `GithubOAuthExecutor` | Authenticates users with GitHub | `code` (authorization code) | `idpId` (Identity Provider id) |
| `SMSOTPAuthExecutor` | Authenticates users with SMS OTP | `username`, `otp` | `senderId` (SMS sender ID) |
| `AttributeCollector` | Collects additional user attributes. Can only be used when there's a authenticated user | `email`, `mobileNumber`, etc |
| `AuthAssertExecutor` | Creates the auth assertion | None |

## Creating a Custom Authentication Flow

To create a custom authentication flow:

1. Create a new JSON file in your graphs directory with the flow definition
2. Define the flow nodes, their executors, and connections
3. Restart the server to register the flow in Thunder

### Flow Definition Structure

```json
{
  "id": "custom_auth_flow",
  "type": "AUTHENTICATION",
  "nodes": [
    {
      "id": "node_id",
      "type": "TASK_EXECUTION",
      "inputData": [
        {
          "name": "param_name",
          "type": "string",
          "required": true
        }
      ],
      "executor": {
        "name": "ExecutorName"
      },
      "next": ["next_node_id"]
    },
    {
      "id": "next_node_id",
      "type": "AUTHENTICATION_SUCCESS"
    }
  ]
}
```

> Make sure to use type `AUTHENTICATION` for the top-level flow object. Each node must have a unique `id`, and the `next` field specifies the next node(s) to transition to after execution.

### Example 1: Multi-factor Authentication Flow

Here's an example of a multi-factor authentication flow that combines basic authentication and SMS OTP:

```json
{
  "id": "auth_flow_config_mfa",
  "type": "AUTHENTICATION",
  "nodes": [
    {
      "id": "basic_auth",
      "type": "TASK_EXECUTION",
      "inputData": [
        {
          "name": "username",
          "type": "string",
          "required": true
        },
        {
          "name": "password",
          "type": "string",
          "required": true
        }
      ],
      "executor": {
        "name": "BasicAuthExecutor"
      },
      "next": ["sms_auth"]
    },
    {
      "id": "sms_auth",
      "type": "TASK_EXECUTION",
      "inputData": [
        {
          "name": "otp",
          "type": "string",
          "required": true
        }
      ],
      "executor": {
        "name": "SMSOTPAuthExecutor"
      },
      "next": ["authenticated"]
    },
    {
      "id": "authenticated",
      "type": "AUTHENTICATION_SUCCESS"
    }
  ]
}
```

Save this file in the graphs directory (e.g., `<THUNDER_HOME>/backend/cmd/server/repository/resources/graphs/auth_flow_config_mfa.json`).

### Example 2: Basic Authentication with Attribute Collection

Here's an example of a basic authentication flow that collects additional user attributes:

```json
{
  "id": "auth_flow_config_basic_with_attributes",
  "type": "AUTHENTICATION",
  "nodes": [
    {
      "id": "basic_auth",
      "type": "TASK_EXECUTION",
      "inputData": [
        {
          "name": "username",
          "type": "string",
          "required": true
        },
        {
          "name": "password",
          "type": "string",
          "required": true
        }
      ],
      "executor": {
        "name": "BasicAuthExecutor"
      },
      "next": ["collect_attributes"]
    },
    {
        "id": "collect_attributes",
        "type": "TASK_EXECUTION",
        "inputData": [
            {
                "name": "email",
                "type": "string",
                "required": true
            },
            {
                "name": "mobileNumber",
                "type": "string",
                "required": false
            }
        ],
        "executor": {
            "name": "AttributeCollector"
        },
        "next": [
            "authenticated"
        ]
    },
    {
      "id": "authenticated",
      "type": "AUTHENTICATION_SUCCESS"
    }
  ]
}
```

Save this file in the graphs directory (e.g., `<THUNDER_HOME>/backend/cmd/server/repository/resources/graphs/auth_flow_config_basic_with_attributes.json`).

## Setting an Authentication Flow for an Application

To set an authentication flow for an application, create a new application or update an existing one using the Thunder API. The `auth_flow_id` field should reference your custom flow.

```bash
curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications/<app_id> \
--data '{
    "name": "My Application",
    "description": "Application with custom auth flow",
    "auth_flow_id": "custom_auth_flow"
}'
```
