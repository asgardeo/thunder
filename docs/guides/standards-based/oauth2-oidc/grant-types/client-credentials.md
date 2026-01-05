# Client Credentials

The Client Credentials grant type is used for machine-to-machine (M2M) communication where no user interaction is required. It's ideal for service-to-service authentication and backend API access.

## Overview

The Client Credentials grant type allows applications to authenticate using their client credentials (client ID and secret) to obtain an access token directly, without user involvement.

**Process Steps:**
1. Client authenticates with client ID and secret
2. Client requests access token
3. Authorization server issues access token

## Use Cases

- **Service-to-service communication**: Microservices authenticating to each other
- **Backend API access**: Server-side applications accessing protected APIs
- **Automated systems**: Scripts and scheduled jobs requiring API access
- **Machine-to-machine**: IoT devices and automated processes

## Prerequisites

1. **Create an OAuth Application** with `client_credentials` grant type
2. **Client credentials** (client ID and secret)

## Step-by-Step Guide

### Step 1: Obtain Admin Token

First, obtain an admin token to create applications:

```bash
# Replace <application_id> with your sample app ID (created during Thunder setup)
ADMIN_TOKEN_RESPONSE=$(curl -k -s -X POST 'https://localhost:8090/flow/execute' \
  -H 'Content-Type: application/json' \
  -d '{
    "applicationId": "<application_id>",
    "flowType": "AUTHENTICATION",
    "inputs": {
      "username": "admin",
      "password": "admin",
      "requested_permissions": "system"
    }
  }')

ADMIN_TOKEN=$(echo $ADMIN_TOKEN_RESPONSE | jq -r '.assertion')
```

### Step 2: Create an OAuth Application

Create an application with OAuth 2.0 configuration supporting the client credentials grant type:

```bash
curl -kL -X POST https://localhost:8090/applications \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "Backend Service",
    "description": "Backend service using client credentials grant type",
    "auth_flow_id": "auth_flow_config_basic",
    "inbound_auth_config": [
      {
        "type": "oauth2",
        "config": {
          "client_id": "backend_service_client",
          "client_secret": "backend_service_secret",
          "redirect_uris": [
            "https://localhost:3000"
          ],
          "grant_types": [
            "client_credentials"
          ],
          "token_endpoint_auth_method": "client_secret_basic",
          "pkce_required": false,
          "public_client": false,
          "scopes": [
            "api:read",
            "api:write"
          ],
          "token": {
            "issuer": "thunder",
            "access_token": {
              "validity_period": 3600
            }
          }
        }
      }
    ]
  }'
```

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Backend Service",
  "description": "Backend service using client credentials grant type",
  "client_id": "backend_service_client",
  "client_secret": "backend_service_secret",
  "auth_flow_id": "auth_flow_config_basic",
  "inbound_auth_config": [
    {
      "type": "oauth2",
      "config": {
        "client_id": "backend_service_client",
        "client_secret": "backend_service_secret",
        "grant_types": [
          "client_credentials"
        ],
        "token_endpoint_auth_method": "client_secret_basic",
        "scopes": [
          "api:read",
          "api:write"
        ]
      }
    }
  ]
}
```

**Key Configuration:**
- `grant_types`: Must include `"client_credentials"`
- `token_endpoint_auth_method`: `"client_secret_basic"` or `"client_secret_post"`
- `scopes`: Custom scopes for your API (optional)
- No `redirect_uris` required

### Step 3: Request Access Token

Request an access token using client credentials:

**Using Basic Authentication (Recommended):**

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'backend_service_client:backend_service_secret' \
  -d 'grant_type=client_credentials'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read api:write"
}
```

**Using POST Body Authentication:**

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=client_credentials' \
  -d 'client_id=backend_service_client' \
  -d 'client_secret=backend_service_secret'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read api:write"
}
```

**With Scopes:**

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'backend_service_client:backend_service_secret' \
  -d 'grant_type=client_credentials' \
  -d 'scope=api:read api:write'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read api:write"
}
```

**With Resource Parameter (RFC 8707):**

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'backend_service_client:backend_service_secret' \
  -d 'grant_type=client_credentials' \
  -d 'scope=api:read' \
  -d 'resource=https://api.example.com'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read"
}
```

The access token will have `"aud": "https://api.example.com"` in its claims.

### Step 4: Use Access Token

Use the access token to access protected resources:

```bash
curl -kL -X GET https://api.example.com/protected-resource \
  -H 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...'
```

**Response:**

```json
{
  "data": "Protected resource data"
}
```

## Token Structure

The access token is a JWT containing:
- `sub`: Subject (client ID)
- `aud`: Audience (client ID or resource if specified)
- `scope`: Granted scopes
- `exp`: Expiration time
- `iat`: Issued at time
- `client_id`: Client identifier

**Note:** The granted scopes depend on:
- The resources the application has access to
- Scopes requested in the token request
- Any scope restrictions

## Resource Parameter (RFC 8707)

The resource parameter allows you to specify the target resource/audience for the token. See the [Resource Parameter](../features/resource-parameter.md) guide for detailed information.

## Error Handling

Common errors and solutions:

| Error | Description | Solution |
|-------|-------------|----------|
| `invalid_client` | Invalid client credentials | Verify client ID and secret |
| `invalid_grant` | Invalid grant type | Ensure `client_credentials` is enabled |
| `invalid_scope` | Invalid scope requested | Check allowed scopes |
| `unauthorized_client` | Client not authorized | Verify grant type is enabled for the client |

## Related Documentation

- [Resource Parameter](../features/resource-parameter.md) - Detailed resource parameter guide
- [Token Endpoint](../endpoints/token-endpoint.md) - Token endpoint details
- [Scopes and Permissions](../features/scopes-permissions.md) - Custom scopes configuration

