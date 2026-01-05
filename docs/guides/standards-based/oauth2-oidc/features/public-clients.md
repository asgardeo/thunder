# Public Clients

Thunder supports public clients for mobile and single-page applications that cannot securely store client secrets.

## Overview

Public clients are OAuth 2.0 applications that cannot securely store client secrets. They are typically:
- Mobile applications
- Single Page Applications (SPAs)
- Desktop applications

**Security Requirements:**
- **PKCE is enforced** - Public clients cannot authorize without PKCE at runtime, adhering to OAuth 2.0, regardless of the `pkce_required` setting.
- Must use `none` as token endpoint authentication method
- Cannot use client credentials grant type
- Cannot have client secrets

## Configuration

### Step 1: Obtain Admin Token

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

### Step 2: Create Public Client Application

```bash
curl -kL -X POST https://localhost:8090/applications \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "My Public Client App",
    "description": "Mobile/SPA application using public client",
    "auth_flow_id": "auth_flow_config_basic",
    "inbound_auth_config": [
      {
        "type": "oauth2",
        "config": {
          "redirect_uris": [
            "https://localhost:3000/callback",
            "http://localhost:3000/callback"
          ],
          "grant_types": [
            "authorization_code",
            "refresh_token"
          ],
          "response_types": [
            "code"
          ],
          "token_endpoint_auth_method": "none",
          "public_client": true,
          "pkce_required": true,
          "scopes": [
            "openid",
            "profile",
            "email"
          ],
          "token": {
            "issuer": "thunder",
            "access_token": {
              "validity_period": 3600,
              "user_attributes": [
                "email",
                "username"
              ]
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
  "name": "My Public Client App",
  "description": "Mobile/SPA application using public client",
  "client_id": "auto-generated-client-id",
  "auth_flow_id": "auth_flow_config_basic",
  "inbound_auth_config": [
    {
      "type": "oauth2",
      "config": {
        "client_id": "auto-generated-client-id",
        "redirect_uris": [
          "https://localhost:3000/callback",
          "http://localhost:3000/callback"
        ],
        "grant_types": [
          "authorization_code",
          "refresh_token"
        ],
        "response_types": [
          "code"
        ],
        "token_endpoint_auth_method": "none",
        "public_client": true,
        "pkce_required": true,
        "scopes": [
          "openid",
          "profile",
          "email"
        ]
      }
    }
  ]
}
```

## Required Fields

**Required:**
- `redirect_uris` - Array of valid redirect URIs for your application
- `grant_types` - Must include `authorization_code` (defaults to `authorization_code` if not specified)
- `response_types` - Must include `code` (defaults to `code` if authorization_code grant is used)
- `token_endpoint_auth_method` - Must be `"none"` for public clients
- `public_client` - Must be `true`

**Recommended:**
- `pkce_required` - Can be set to `true` for explicit documentation, but PKCE is automatically enforced for public clients at runtime regardless of this setting
- `scopes` - OIDC scopes like `openid`, `profile`, `email`

**Not Allowed:**
- `client_secret` - Must not be provided (or empty)
- `client_credentials` grant type - Not allowed for public clients

## Security Considerations

- **PKCE is automatically enforced**: Public clients cannot complete authorization without PKCE. If a public client attempts to authorize without providing `code_challenge` and `code_challenge_method` parameters, the authorization request will be rejected with an `invalid_request` error.
- **No client secret**: Public clients don't use client secrets
- **Token endpoint auth**: Must use `none` authentication method
- **Code verifier required**: When exchanging the authorization code for tokens, public clients must provide the `code_verifier` parameter

## Related Documentation

- [PKCE](pkce.md) - Proof Key for Code Exchange
- [Authorization Code](../grant-types/authorization-code.md) - Using with public clients

