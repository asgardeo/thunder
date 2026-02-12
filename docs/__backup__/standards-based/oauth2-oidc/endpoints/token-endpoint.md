# Token Endpoint

The Token endpoint (`/oauth2/token`) is used to exchange authorization codes, refresh tokens, or client credentials for access tokens.

## Endpoint

```
POST /oauth2/token
```

## Request Format

The token endpoint accepts `application/x-www-form-urlencoded` requests.

### Client Authentication

Thunder supports multiple client authentication methods:

**1. Client Secret Basic (Recommended):**
```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'grant_type=authorization_code' \
  -d 'code=AUTHORIZATION_CODE' \
  -d 'redirect_uri=https://localhost:3000/callback'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_xyz123",
  "scope": "openid profile email",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**2. Client Secret Post:**
```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=authorization_code' \
  -d 'client_id=client_id' \
  -d 'client_secret=client_secret' \
  -d 'code=AUTHORIZATION_CODE' \
  -d 'redirect_uri=https://localhost:3000/callback'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_xyz123",
  "scope": "openid profile email",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Grant Types

### Authorization Code

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'grant_type=authorization_code' \
  -d 'code=AUTHORIZATION_CODE' \
  -d 'redirect_uri=https://localhost:3000/callback' \
  -d 'code_verifier=CODE_VERIFIER'  # If PKCE is used
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_xyz123",
  "scope": "openid profile email",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Client Credentials

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'grant_type=client_credentials' \
  -d 'scope=api:read' \
  -d 'resource=https://api.example.com/resource'
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

### Refresh Token

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'grant_type=refresh_token' \
  -d 'refresh_token=REFRESH_TOKEN' \
  -d 'scope=profile email'  # Optional: scope downscoping
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "new_refresh_token_xyz123",
  "scope": "profile email"
}
```

### Token Exchange

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'grant_type=urn:ietf:params:oauth:grant-type:token-exchange' \
  -d 'subject_token=SUBJECT_TOKEN' \
  -d 'subject_token_type=urn:ietf:params:oauth:token-type:access_token' \
  -d 'resource=https://api.example.com/resource'
```

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

## Error Responses

**Error Response (400 Bad Request):**

```json
{
  "error": "invalid_grant",
  "error_description": "Invalid authorization code"
}
```

**Error Response (401 Unauthorized):**

```json
{
  "error": "invalid_client",
  "error_description": "Client authentication failed"
}
```

## Response Headers

- `Content-Type: application/json`
- `Cache-Control: no-store` (per RFC 6749)

## Error Codes

| Error Code | Description |
|------------|-------------|
| `invalid_request` | Missing or invalid parameters |
| `invalid_client` | Invalid client credentials |
| `invalid_grant` | Invalid authorization code, refresh token, or subject token |
| `invalid_scope` | Invalid scope requested |
| `unauthorized_client` | Client not authorized for grant type |
| `unsupported_grant_type` | Grant type not supported |

## Related Documentation

- [Authorization Code](../grant-types/authorization-code.md)
- [Client Credentials](../grant-types/client-credentials.md)
- [Refresh Token](../grant-types/refresh-token.md)

