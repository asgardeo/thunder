# Dynamic Client Registration (DCR)

Thunder supports Dynamic Client Registration (DCR) as specified in [RFC 7591](https://datatracker.ietf.org/doc/html/rfc7591), allowing OAuth clients to register themselves programmatically with the authorization server.

## Endpoint

```
POST /oauth2/dcr/register
```

## Overview

DCR allows clients to register themselves with the authorization server without manual configuration. This is particularly useful for:

- MCP clients connecting to new MCP servers
- Applications that need to register dynamically
- Reducing manual configuration overhead
- Enabling seamless integration with new authorization servers

## Request Format

**Content-Type:** `application/json`

**Request Body:**

```json
{
  "redirect_uris": [
    "https://client.example.com/callback",
    "http://localhost:3000/callback"
  ],
  "grant_types": [
    "authorization_code",
    "refresh_token"
  ],
  "response_types": [
    "code"
  ],
  "client_name": "My OAuth Client",
  "client_uri": "https://client.example.com",
  "logo_uri": "https://client.example.com/logo.png",
  "token_endpoint_auth_method": "client_secret_basic",
  "jwks_uri": "https://client.example.com/.well-known/jwks.json",
  "scope": "openid profile email",
  "contacts": [
    "admin@client.example.com"
  ],
  "tos_uri": "https://client.example.com/terms",
  "policy_uri": "https://client.example.com/privacy"
}
```

**Required Fields:**
- `redirect_uris` - Array of valid redirect URIs required for `authorization_code` grant

**Optional Fields:**
- `grant_types` - Defaults to `["authorization_code"]` if not specified
- `response_types` - Defaults to `["code"]` if not specified
- `client_name` - Human-readable client name
- `client_uri` - Client homepage URL
- `logo_uri` - Client logo URL
- `token_endpoint_auth_method` - Defaults to `"client_secret_basic"`. Can be:
  - `"client_secret_basic"` - HTTP Basic Authentication
  - `"client_secret_post"` - POST body authentication
  - `"none"` - Public client (no authentication)
- `jwks_uri` - URL to client's JWKS endpoint
- `jwks` - Inline JWKS (cannot be used with `jwks_uri`)
- `scope` - Space-separated list of requested scopes
- `contacts` - Array of contact email addresses
- `tos_uri` - Terms of Service URI
- `policy_uri` - Privacy Policy URI

## Complete Example

### Register a Client

```bash
curl -kL -X POST https://localhost:8090/oauth2/dcr/register \
  -H 'Content-Type: application/json' \
  -d '{
    "redirect_uris": [
      "http://localhost:3000/callback"
    ],
    "grant_types": [
      "authorization_code",
      "refresh_token"
    ],
    "response_types": [
      "code"
    ],
    "client_name": "My OAuth Client Application",
    "token_endpoint_auth_method": "client_secret_basic",
    "scope": "openid profile email"
  }'
```

**Response:**

```json
{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_secret": "generated_secret_abc123xyz789",
  "client_secret_expires_at": 0,
  "redirect_uris": [
    "http://localhost:3000/callback"
  ],
  "grant_types": [
    "authorization_code",
    "refresh_token"
  ],
  "response_types": [
    "code"
  ],
  "client_name": "My OAuth Client Application",
  "token_endpoint_auth_method": "client_secret_basic",
  "scope": "openid profile email",
  "app_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Response Fields:**
- `client_id` - OAuth client identifier (UUID format)
- `client_secret` - Client secret (if not a public client)
- `client_secret_expires_at` - Expiration timestamp (0 = never expires)
- `app_id` - Thunder application ID
- Other fields echo back the registered values

### Register a Public Client

```bash
curl -kL -X POST https://localhost:8090/oauth2/dcr/register \
  -H 'Content-Type: application/json' \
  -d '{
    "redirect_uris": [
      "http://localhost:3000/callback"
    ],
    "grant_types": [
      "authorization_code"
    ],
    "response_types": [
      "code"
    ],
    "client_name": "Public SPA Client",
    "token_endpoint_auth_method": "none"
  }'
```

**Response:**

```json
{
  "client_id": "660e8400-e29b-41d4-a716-446655440002",
  "client_secret_expires_at": 0,
  "redirect_uris": [
    "http://localhost:3000/callback"
  ],
  "grant_types": [
    "authorization_code"
  ],
  "response_types": [
    "code"
  ],
  "client_name": "Public SPA Client",
  "token_endpoint_auth_method": "none",
  "app_id": "660e8400-e29b-41d4-a716-446655440003"
}
```

Note: Public clients don't receive a `client_secret`.

### Register with JWKS URI

```bash
curl -kL -X POST https://localhost:8090/oauth2/dcr/register \
  -H 'Content-Type: application/json' \
  -d '{
    "redirect_uris": [
      "https://client.example.com/callback"
    ],
    "grant_types": [
      "authorization_code"
    ],
    "response_types": [
      "code"
    ],
    "client_name": "Client with JWKS",
    "token_endpoint_auth_method": "client_secret_basic",
    "jwks_uri": "https://client.example.com/.well-known/jwks.json"
  }'
```

## Error Responses

### Invalid Redirect URI

**Status:** `400 Bad Request`

```json
{
  "error": "invalid_redirect_uri",
  "error_description": "One or more redirect URIs are invalid"
}
```

### Invalid Client Metadata

**Status:** `400 Bad Request`

```json
{
  "error": "invalid_client_metadata",
  "error_description": "Invalid client metadata provided"
}
```

### JWKS Configuration Conflict

**Status:** `400 Bad Request`

```json
{
  "error": "invalid_client_metadata",
  "error_description": "Cannot specify both jwks_uri and jwks"
}
```

## Related Documentation

- [OAuth 2.0 Grant Types](../grant-types/) - Using registered clients
- [Authorization Code](../grant-types/authorization-code.md) - Authorization code grant type with registered client
- [Public Clients](public-clients.md) - Public client configuration
- [Securing MCP Server](mcp-server-securing.md) - DCR usage in MCP scenarios
