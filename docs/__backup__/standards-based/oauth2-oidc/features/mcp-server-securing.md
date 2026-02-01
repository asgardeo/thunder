# Securing MCP Servers with Thunder

This guide explains how to secure Model Context Protocol (MCP) servers using Thunder's OAuth 2.0 implementation, following the [MCP Authorization Specification](https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization).

## Overview

The Model Context Protocol (MCP) provides authorization capabilities, enabling MCP clients to make requests to restricted MCP servers on behalf of resource owners. This guide demonstrates how Thunder can serve as the authorization server for MCP servers.

## Complete Authorization Process

### Step 1: MCP Server Returns 401 with WWW-Authenticate Header

When an MCP client makes a request without a token, the MCP server **MUST** return HTTP 401 with a `WWW-Authenticate` header indicating the Protected Resource Metadata URL:

```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer realm="mcp.example.com",
  resource_metadata="https://mcp.example.com/.well-known/oauth-protected-resource"
```

### Step 2: MCP Client Discovers Protected Resource Metadata

The MCP client requests the Protected Resource Metadata from the MCP server, which includes Thunder in `authorization_servers`:

```bash
curl -kL https://mcp.example.com/.well-known/oauth-protected-resource
```

**Response:**

```json
{
  "resource": "https://mcp.example.com/mcp",
  "authorization_servers": [
    "https://localhost:8090"
  ],
  "jwks_uri": "https://localhost:8090/oauth2/jwks",
  "scopes_supported": [
    "mcp:read",
    "mcp:write"
  ]
}
```

### Step 3: MCP Client Discovers Authorization Server Metadata

The MCP client requests Authorization Server Metadata from Thunder:

```bash
curl -kL https://localhost:8090/.well-known/oauth-authorization-server
```

**Response:**

```json
{
  "issuer": "https://localhost:8090",
  "authorization_endpoint": "https://localhost:8090/oauth2/authorize",
  "token_endpoint": "https://localhost:8090/oauth2/token",
  "token_endpoint_auth_methods_supported": [
    "client_secret_basic",
    "client_secret_post",
    "none"
  ],
  "jwks_uri": "https://localhost:8090/oauth2/jwks",
  "response_types_supported": ["code"],
  "grant_types_supported": [
    "authorization_code",
    "client_credentials",
    "refresh_token",
    "urn:ietf:params:oauth:grant-type:token-exchange"
  ],
  "code_challenge_methods_supported": ["S256", "plain"],
  "scopes_supported": ["openid", "profile", "email", "mcp:read", "mcp:write"],
  "registration_endpoint": "https://localhost:8090/oauth2/dcr/register",
  "introspection_endpoint": "https://localhost:8090/oauth2/introspect"
}
```

### Step 4: Dynamic Client Registration (Optional but Recommended)

MCP clients **SHOULD** use Dynamic Client Registration to obtain client credentials automatically. This is crucial for MCP because clients may not know all possible MCP servers and their authorization servers in advance.

**Register Client:**

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
    "client_name": "MCP Client Application",
    "token_endpoint_auth_method": "client_secret_basic",
    "scope": "openid mcp:read mcp:write"
  }'
```

**Response:**

```json
{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_secret": "generated_secret_abc123xyz",
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
  "client_name": "MCP Client Application",
  "token_endpoint_auth_method": "client_secret_basic",
  "scope": "openid mcp:read mcp:write",
  "app_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

### Step 5: Authorization Request with PKCE and Resource Parameter

MCP clients **MUST** implement PKCE and **MUST** include the `resource` parameter. Generate PKCE parameters:



**Authorization Request:**

```bash
AUTH_URL="https://localhost:8090/oauth2/authorize?response_type=code&client_id=550e8400-e29b-41d4-a716-446655440000&redirect_uri=http://localhost:3000/callback&scope=openid%20mcp:read%20mcp:write&resource=https://mcp.example.com/mcp&code_challenge=$CODE_CHALLENGE&code_challenge_method=S256&state=$STATE"
```

**User Authorizes → Redirect to Callback:**

```
http://localhost:3000/callback?code=AUTHORIZATION_CODE&state=STATE_VALUE
```

### Step 6: Token Request with Code Verifier and Resource Parameter

Exchange the authorization code for an access token:

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u '550e8400-e29b-41d4-a716-446655440000:generated_secret_abc123xyz' \
  -d "grant_type=authorization_code" \
  -d "code=AUTHORIZATION_CODE" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "code_verifier=$CODE_VERIFIER" \
  -d "resource=https://mcp.example.com/mcp"
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_xyz123",
  "scope": "openid mcp:read mcp:write",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Token Claims (decoded access token):**

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "aud": "https://mcp.example.com/mcp",
  "iss": "https://localhost:8090",
  "exp": 1234567890,
  "iat": 1234564290,
  "scope": "openid mcp:read mcp:write"
}
```

**Important:** The `aud` claim **MUST** match the MCP server URL (`https://mcp.example.com/mcp`) because the `resource` parameter was included in the token request.

### Step 7: Access MCP Server with Token

Use the access token to make requests to the MCP server:

```bash
curl -kL -X POST https://mcp.example.com/mcp \
  -H 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...' \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "id": 1,
    "params": {}
  }'
```

### Step 8: MCP Server Validates Token

The MCP server **MUST** validate the token according to [OAuth 2.1 Section 5.2](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1-13#section-5.2):

1. **Extract Bearer token** from `Authorization` header
2. **Validate token signature** using Thunder's JWKS endpoint
3. **Verify audience** - The `aud` claim **MUST** match the MCP server URL exactly
4. **Check expiration** - Validate `exp` claim
5. **Validate issuer** - Verify `iss` matches Thunder's issuer
6. **Check scopes** - Verify required scopes are present

**Get JWKS for Token Validation:**

```bash
curl -kL https://localhost:8090/oauth2/jwks
```

**Response:**

```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "key-id-1",
      "use": "sig",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

## Alternative: Client Credentials (M2M)

For machine-to-machine scenarios without user interaction, use the Client Credentials grant type:

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u '550e8400-e29b-41d4-a716-446655440000:generated_secret_abc123xyz' \
  -d "grant_type=client_credentials" \
  -d "scope=mcp:read mcp:write" \
  -d "resource=https://mcp.example.com/mcp"
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "mcp:read mcp:write"
}
```

## Related Documentation

- [MCP Authorization Specification](https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization) - Official MCP authorization spec
- [Authorization Code](../grant-types/authorization-code.md) - Detailed authorization code grant type guide
- [Client Credentials](../grant-types/client-credentials.md) - M2M authentication
- [Resource Parameter](resource-parameter.md) - RFC 8707 resource parameter details
- [Dynamic Client Registration](dynamic-client-registration.md) - DCR implementation
- [PKCE](pkce.md) - Proof Key for Code Exchange
- [JWKS Endpoint](../endpoints/jwks-endpoint.md) - Token validation endpoint
