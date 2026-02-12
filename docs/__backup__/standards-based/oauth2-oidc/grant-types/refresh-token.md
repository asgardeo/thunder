# Refresh Token

The Refresh Token grant type allows clients to obtain new access tokens using a refresh token, without requiring the user to re-authenticate. This is essential for maintaining long-lived sessions and improving user experience.

## Overview

When an access token expires, instead of redirecting the user through the full authentication process again, you can use a refresh token to obtain a new access token.

**Process Steps:**
1. Client has a refresh token (obtained from initial authorization)
2. Client requests new access token using refresh token
3. Authorization server validates refresh token and issues new access token

## Use Cases

- **Long-lived sessions**: Applications requiring persistent user sessions
- **Mobile applications**: Reducing authentication frequency
- **Reduced user friction**: Seamless token renewal
- **Offline access**: Applications that work offline

## Prerequisites

1. **OAuth Application** with `refresh_token` grant type enabled
2. **Refresh token** obtained from initial authorization (Authorization Code grant type)

## Step-by-Step Guide

### Step 1: Obtain Refresh Token

First, obtain a refresh token through the Authorization Code grant type. The refresh token is included in the token response:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "scope": "openid profile email"
}
```

### Step 2: Request New Access Token

Use the refresh token to obtain a new access token:

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'grant_type=refresh_token' \
  -d 'refresh_token=REFRESH_TOKEN'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "scope": "openid profile email"
}
```

**With Scope Downscoping:**

You can request a subset of the original scopes:

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'grant_type=refresh_token' \
  -d 'refresh_token=REFRESH_TOKEN' \
  -d 'scope=profile email'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "scope": "profile email"
}
```

**Note:** Only scopes that were originally granted can be requested. If you request scopes not in the original token, they will be ignored.

**Refresh Token Renewal:**

Thunder supports configurable refresh token renewal. If `renew_on_grant` is enabled in configuration, a new refresh token is issued with each refresh. Otherwise, the same refresh token is returned.

## Scope Downscoping

Thunder implements OAuth 2.0 scope downscoping. When requesting a new access token:

- **No scope specified**: All original scopes are granted
- **Subset of scopes**: Only the intersection of requested and original scopes is granted
- **New scopes**: Scopes not in the original token are ignored

**Example:**

Original token scopes: `["openid", "profile", "email", "api:read"]`

Request with `scope=profile email`:
- Granted: `["profile", "email"]`

Request with `scope=api:write`:
- Granted: `[]` (api:write not in original)

## Configuration

Refresh token behavior can be configured in Thunder's configuration:

```yaml
oauth:
  refresh_token:
    renew_on_grant: false  # Set to true to issue new refresh token on each grant
    validity_period: 86400  # Refresh token validity in seconds (default: 24 hours)
```

## Error Handling

Common errors and solutions:

| Error | Description | Solution |
|-------|-------------|----------|
| `invalid_grant` | Invalid or expired refresh token | Obtain a new refresh token through authorization grant type |
| `invalid_client` | Invalid client credentials | Verify client ID and secret |
| `invalid_request` | Missing required parameters | Include grant_type and refresh_token |

## Related Documentation

- [Authorization Code](authorization-code.md) - How to obtain initial refresh token
- [Token Endpoint](../endpoints/token-endpoint.md) - Token endpoint details

