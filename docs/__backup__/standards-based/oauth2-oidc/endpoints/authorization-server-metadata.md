# OAuth 2.0 Authorization Server Metadata

The OAuth 2.0 Authorization Server Metadata endpoint (RFC 8414) provides automatic client configuration information.

## Endpoint

```
GET /.well-known/oauth-authorization-server
```

## Request

```bash
curl -kL https://localhost:8090/.well-known/oauth-authorization-server
```

## Response

Returns OAuth 2.0 authorization server metadata including:

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
  "code_challenge_methods_supported": ["S256"],
  "scopes_supported": ["openid", "profile", "email"],
  "introspection_endpoint": "https://localhost:8090/oauth2/introspect",
  "registration_endpoint": "https://localhost:8090/oauth2/dcr/register"
}
```

## Usage

Clients can use this endpoint to automatically discover:
- Authorization and token endpoints
- Supported grant types and response types
- Supported authentication methods
- JWKS endpoint location
- Supported scopes
- Dynamic Client Registration endpoint
- Token introspection endpoint

## Related Documentation

- [OpenID Connect Discovery](oidc-discovery.md) - OIDC discovery endpoint

