# PKCE (Proof Key for Code Exchange)

PKCE (RFC 7636) is a security extension for OAuth 2.0 that enhances the security of the Authorization Code grant type, especially for public clients.

## Overview

PKCE protects against authorization code interception attacks by:
- Generating a code verifier (secret) on the client
- Creating a code challenge from the verifier
- Sending the challenge during authorization
- Verifying the verifier during token exchange

## Why Use PKCE?

- **Enhanced Security**: Protects against authorization code interception
- **Public Client Support**: Enables secure OAuth for mobile and SPA apps
- **Best Practice**: Recommended for all authorization code grant types

## Implementation

### Step 1: Generate Code Verifier

Generate a random code verifier (43-128 characters, URL-safe)

### Step 2: Generate Code Challenge

Create code challenge using S256 (recommended)

### Step 3: Authorization Request

Include code challenge in authorization request:

```
https://localhost:8090/oauth2/authorize?
  response_type=code&
  client_id=client_id&
  redirect_uri=https://localhost:3000/callback&
  scope=openid%20profile&
  code_challenge=CODE_CHALLENGE&
  code_challenge_method=S256&
  state=random_state
```

### Step 4: Token Request

Include code verifier in token request:

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'grant_type=authorization_code' \
  -d 'code=AUTHORIZATION_CODE' \
  -d 'redirect_uri=https://localhost:3000/callback' \
  -d 'code_verifier=CODE_VERIFIER'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_xyz123",
  "scope": "openid profile",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Code Challenge Methods

Thunder supports:
- **S256**: SHA256 hash (required, per OAuth 2.0 Security Best Current Practice)

The `plain` method is not supported as it does not provide adequate security. Per RFC 9700 (OAuth 2.0 Security Best Current Practice), clients MUST use S256 as the code challenge method.

## Application Configuration

Enable PKCE requirement for an application:

```json
{
  "inbound_auth_config": [{
    "type": "oauth2",
    "config": {
      "pkce_required": true
    }
  }]
}
```

When `pkce_required` is `true`, PKCE parameters are mandatory. Additionally PKCE is checked mandatorily for `public_clients` during runtime.

## Related Documentation

- [Authorization Code](../grant-types/authorization-code.md) - Using PKCE with authorization code grant type

