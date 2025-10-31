# Customizing Registration Flows

This guide explains how to customize registration flows in WSO2 Thunder using the flow graph system. Thunder uses a flexible graph-based approach to define registration flows, allowing you to create tailored user registration experiences.

## Understanding Registration Flows

In Thunder, registration flows are represented as directed graphs, similar to authentication flows. Each graph consists of:

- **Nodes**: Individual steps in the registration process
- **Edges**: Connections between nodes that define the flow path
- **Executors**: Components that perform specific registration tasks

Registration flows are defined in JSON files and stored in the graph directory configured in your deployment.

## Automatic Registration Flow Creation

When you create an authentication flow, Thunder automatically creates an equivalent registration flow with the following transformations:

- The flow ID is prefixed with `registration_flow_config_` instead of `auth_flow_config_`
- A `ProvisioningExecutor` node is inserted before the final `AUTHENTICATION_SUCCESS` node
- All existing authentication steps are preserved in the registration flow

This ensures that users can register using the same authentication methods configured for login.

Thunder comes with several registration flows corresponding to the pre-configured authentication flows.

## Registration Flow Components

### Node Types

Registration flows support the same node types as authentication flows:

| Node Type | Constant | Description |
|-----------|----------|-------------|
| Task Execution | `TASK_EXECUTION` | Performs a specific registration task using an executor |
| Prompt Only | `PROMPT_ONLY` | Prompts for user input without performing any tasks |
| Decision | `DECISION` | Implements conditional logic in the flow |
| Authentication Success | `AUTHENTICATION_SUCCESS` | Finalizes the registration process and generates the auth assertion. Should be placed after the `ProvisioningExecutor` node |

### Executors

Registration flows can use most authentication executors plus specific registration executors:

| Executor | Description | Input Parameters | Notes |
|----------|-------------|------------------|-------|
| `BasicAuthExecutor` | Validates username and password for password-based registration | `username`, `password` |
| `GoogleOIDCAuthExecutor` | Authenticates users with Google for social registration | `code` (authorization code) |
| `GithubOAuthExecutor` | Authenticates users with GitHub for social registration | `code` (authorization code) |
| `SMSOTPAuthExecutor` | Validates SMS OTP for OTP-based registration | `username`, `otp` |
| `ProvisioningExecutor` | Creates the user account | Various attributes |

## Creating a Custom Registration Flow

To create a custom registration flow:

1. Create a new JSON file in your graphs directory with the flow definition
2. Use the `registration_flow_config_` prefix for automatic detection or any custom name
3. Define the flow nodes, their executors, and connections
4. Include a `ProvisioningExecutor` node to create user accounts
5. Add an `AUTHENTICATION_SUCCESS` node at the end if you need to automatically login the user after registration
6. Restart the server to register the flow in Thunder

### Flow Definition Structure

```json
{
  "id": "registration_flow_config_custom",
  "type": "REGISTRATION",
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
      "next": ["provisioning"]
    },
    {
      "id": "provisioning",
      "type": "TASK_EXECUTION",
      "inputData": [
        {
          "name": "firstName",
          "type": "string",
          "required": true
        },
        {
          "name": "lastName",
          "type": "string",
          "required": true
        }
      ],
      "executor": {
        "name": "ProvisioningExecutor"
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

> * Make sure to use type `REGISTRATION` for the top-level flow object. The `ProvisioningExecutor` should be placed after all authentication and data collection nodes.
> * By default `ProvisioningExecutor` will provision all previously collected attributes. If you want to provision only specific attributes, define the `inputData` in the `ProvisioningExecutor` node.
> * `AUTHENTICATION_SUCCESS` node is optional but recommended if you want to authenticate the user immediately after registration.

### Example 1: Basic Registration with Attribute Collection

Here's an example of a basic registration flow that collects additional user attributes:

```json
{
  "id": "registration_flow_config_basic_with_profile",
  "type": "REGISTRATION",
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
      "next": ["collect_profile"]
    },
    {
      "id": "collect_profile",
      "type": "PROMPT_ONLY",
      "inputData": [
        {
          "name": "firstName",
          "type": "string",
          "required": true
        },
        {
          "name": "lastName",
          "type": "string",
          "required": true
        },
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
      "next": ["provisioning"]
    },
    {
      "id": "provisioning",
      "type": "TASK_EXECUTION",
      "inputData": [
        {
          "name": "username",
          "type": "string",
          "required": true
        },
        {
          "name": "firstName",
          "type": "string",
          "required": true
        },
        {
          "name": "lastName",
          "type": "string",
          "required": true
        },
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
        "name": "ProvisioningExecutor"
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

### Example 2: Social Registration Flow

Here's an example of a registration flow using Google OAuth:

```json
{
  "id": "registration_flow_config_google_with_profile",
  "type": "REGISTRATION",
  "nodes": [
    {
      "id": "google_auth",
      "type": "TASK_EXECUTION",
      "inputData": [
        {
          "name": "code",
          "type": "string",
          "required": true
        },
        {
          "name": "nonce",
          "type": "string",
          "required": false
        }
      ],
      "executor": {
        "name": "GoogleOIDCAuthExecutor",
        "idpName": "Google"
      },
      "next": ["collect_additional_info"]
    },
    {
      "id": "collect_additional_info",
      "type": "PROMPT_ONLY",
      "inputData": [
        {
          "name": "mobileNumber",
          "type": "string",
          "required": true
        },
        {
          "name": "dateOfBirth",
          "type": "string",
          "required": false
        }
      ],
      "next": ["provisioning"]
    },
    {
      "id": "provisioning",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "ProvisioningExecutor"
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

## Setting up a Registration Flow for an Application

When you configure an authentication flow for an application, Thunder automatically assigns the equivalent registration flow. For example, if you set `auth_flow_graph_id` to `auth_flow_config_basic`, the system will automatically use `registration_flow_config_basic` for registration.

To use a custom registration flow, specify both the authentication and registration flow graph IDs when creating or updating an application:

```bash
curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications/<app_id> \
--data '{
    "name": "My Application",
    "description": "Application with custom registration flow",
    "auth_flow_graph_id": "custom_auth_flow",
    "registration_flow_graph_id": "custom_registration_flow",
    "is_registration_flow_enabled": true
}'
```
