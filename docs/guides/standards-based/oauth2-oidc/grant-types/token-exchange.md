# Token Exchange

The Token Exchange grant type (RFC 8693) allows clients to exchange tokens for different audiences or scopes, enabling token delegation and impersonation scenarios.

## Overview

Token Exchange enables:
- **Token Delegation**: Exchange a token for a new token with different audience
- **Scope Modification**: Request different scopes for the new token
- **Impersonation**: Exchange tokens on behalf of users (with actor tokens)

## Use Cases

- Service-to-service token delegation
- Token transformation for different audiences
- Impersonation scenarios (with proper authorization)

## Prerequisites

1. **OAuth Application** with `urn:ietf:params:oauth:grant-type:token-exchange` grant type
2. **Subject Token** - The token to exchange (only Thunder issued tokens are supported now)

## Step-by-Step Guide

### Step 1: Obtain Admin Token

First, obtain an admin token to create applications.

#### 1.1: Initiate the Authentication Flow

Run the following command, replacing `<application_id>` with your sample app ID (created during Thunder setup).

```bash
FLOW_RESPONSE=$(curl -k -s -X POST 'https://localhost:8090/flow/execute' \
  -d '{"applicationId":"<application_id>","flowType":"AUTHENTICATION"}')

FLOW_ID=$(echo $FLOW_RESPONSE | jq -r '.flowId')
```

#### 1.2: Submit Admin Credentials

Run the following command with the extracted `flowId`.

```bash
ADMIN_TOKEN_RESPONSE=$(curl -k -s -X POST 'https://localhost:8090/flow/execute' \
  -d '{"flowId":"'$FLOW_ID'", "inputs":{"username":"admin","password":"admin","requested_permissions":"system"},"action": "action_001"}')

ADMIN_TOKEN=$(echo $ADMIN_TOKEN_RESPONSE | jq -r '.assertion')
```

### Step 2: Create OAuth Application

```bash
curl -kL -X POST https://localhost:8090/applications \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "Token Exchange Client",
    "description": "Client for token exchange",
    "auth_flow_id": "<auth-flow-uuid>",
    "inbound_auth_config": [
      {
        "type": "oauth2",
        "config": {
          "client_id": "token_exchange_client",
          "client_secret": "token_exchange_secret",
          "grant_types": [
            "urn:ietf:params:oauth:grant-type:token-exchange"
          ],
          "token_endpoint_auth_method": "client_secret_basic"
        }
      }
    ]
  }'
```

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Token Exchange Client",
  "description": "Client for token exchange",
  "client_id": "token_exchange_client",
  "client_secret": "token_exchange_secret",
  "auth_flow_id": "<auth-flow-uuid>",
  "inbound_auth_config": [
    {
      "type": "oauth2",
      "config": {
        "client_id": "token_exchange_client",
        "client_secret": "token_exchange_secret",
        "grant_types": [
          "urn:ietf:params:oauth:grant-type:token-exchange"
        ],
        "token_endpoint_auth_method": "client_secret_basic"
      }
    }
  ]
}
```

### Step 3: Exchange Token

Exchange a subject token for a new token:

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'token_exchange_client:token_exchange_secret' \
  -d 'grant_type=urn:ietf:params:oauth:grant-type:token-exchange' \
  -d 'subject_token=SUBJECT_ACCESS_TOKEN' \
  -d 'subject_token_type=urn:ietf:params:oauth:token-type:access_token' \
  -d 'resource=https://api.example.com/resource' \
  -d 'scope=api:read'
```

**Parameters:**
- `grant_type`: `urn:ietf:params:oauth:grant-type:token-exchange`
- `subject_token`: The token to exchange
- `subject_token_type`: Token type (e.g., `urn:ietf:params:oauth:token-type:access_token`)
- `resource`: Target resource/audience (optional)
- `audience`: Alternative to resource (optional)
- `scope`: Requested scopes (optional)
- `actor_token`: Actor token for impersonation (optional)
- `actor_token_type`: Actor token type (optional)

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token"
}
```

## Token Types

**Supported input types** (`subject_token_type` and `actor_token_type`):
- `urn:ietf:params:oauth:token-type:access_token` - Access token
- `urn:ietf:params:oauth:token-type:jwt` - JWT token

**Supported output types** (`requested_token_type`):
- `urn:ietf:params:oauth:token-type:access_token` - Access token (default)

**Note:** Thunder currently always issues access tokens in token exchange.

## Scope Handling

- If no scope is requested, original scopes are preserved
- If scope is requested, only intersection with original scopes is granted
- New scopes not in original token are ignored

## Related Documentation

- [Client Credentials](client-credentials.md) - Obtain initial token
- [Resource Parameter](../features/resource-parameter.md) - Resource parameter details

