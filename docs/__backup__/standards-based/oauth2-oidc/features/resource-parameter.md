# Resource Parameter (RFC 8707)

The Resource Parameter allows clients to specify the target resource or audience for an access token, enabling fine-grained access control and multi-resource scenarios.

## Overview

The resource parameter (RFC 8707) enables:
- **Audience Specification**: Set the `aud` claim in access tokens
- **Multi-Resource Support**: Same client accessing multiple resources
- **Resource-Specific Tokens**: Tokens scoped to specific resources

## Usage

### Authorization Request

Include the `resource` parameter in authorization requests:

```
https://localhost:8090/oauth2/authorize?
  response_type=code&
  client_id=client_id&
  redirect_uri=https://localhost:3000/callback&
  scope=openid%20profile&
  resource=https://api.example.com/resource
```

### Token Request

Include the `resource` parameter in token requests:

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'grant_type=client_credentials' \
  -d 'resource=https://api.example.com/resource' \
  -d 'scope=api:read'
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

## Requirements

The resource parameter must:
- Be an **absolute URI** (include scheme)
- **Not contain a fragment** component
- Be a valid URI format

**Valid Examples:**
- `https://api.example.com/resource`
- `https://mcp.example.com/mcp`
- `urn:example:resource`

**Invalid Examples:**
- `api.example.com/resource` (not absolute)
- `https://api.example.com/resource#fragment` (contains fragment)

## Token Response

When a resource parameter is provided, the access token's `aud` (audience) claim is set to the resource value:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**Decoded Token Claims:**

```json
{
  "sub": "client_id",
  "aud": "https://api.example.com/resource",
  "scope": "api:read",
  "exp": 1234567890,
  "iat": 1234564290
}
```

## Use Cases

### 1. MCP Server Securing

Specify MCP server as the target resource:

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'mcp_client:mcp_secret' \
  -d 'grant_type=client_credentials' \
  -d 'resource=https://mcp.example.com/mcp' \
  -d 'scope=mcp:read'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "mcp:read"
}
```

The access token will have `"aud": "https://mcp.example.com/mcp"` in its claims.

### 2. Multi-Resource Access

Same client accessing multiple resources:

```bash
# Token for API 1
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client:secret' \
  -d 'grant_type=client_credentials' \
  -d 'resource=https://api1.example.com' \
  -d 'scope=api:read'
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

```bash
# Token for API 2
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client:secret' \
  -d 'grant_type=client_credentials' \
  -d 'resource=https://api2.example.com' \
  -d 'scope=api:read'
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

## Resource Validation

Resources receiving tokens should:
1. Extract the `aud` claim from the access token
2. Verify it matches the resource's URL
3. Reject tokens with mismatched audiences

## Related Documentation

- [Client Credentials](../grant-types/client-credentials.md) - Using resource parameter
- [MCP Server Securing](mcp-server-securing.md) - Resource parameter for MCP

