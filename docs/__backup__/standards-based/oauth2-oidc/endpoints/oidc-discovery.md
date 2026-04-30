# OIDC Discovery

Thunder provides OIDC Discovery endpoint for automatic client configuration.

## Endpoint

```
GET /.well-known/openid-configuration
```

## Request

```bash
curl -kL https://localhost:8090/.well-known/openid-configuration
```

## Response

Returns OIDC configuration including:

```json
{
  "issuer": "https://localhost:8090",
  "authorization_endpoint": "https://localhost:8090/oauth2/authorize",
  "token_endpoint": "https://localhost:8090/oauth2/token",
  "userinfo_endpoint": "https://localhost:8090/oauth2/userinfo",
  "jwks_uri": "https://localhost:8090/oauth2/jwks",
  "response_types_supported": ["code"],
  "grant_types_supported": [
    "authorization_code",
    "client_credentials",
    "refresh_token",
    "urn:ietf:params:oauth:grant-type:token-exchange"
  ],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid", "profile", "email"],
  "claims_supported": [
    "sub",
    "iss",
    "aud",
    "exp",
    "iat",
    "auth_time",
    "email",
    "email_verified",
    "name",
    "given_name",
    "family_name"
  ]
}
```

## Usage

Clients can use this endpoint to automatically discover:
- Authorization, token, and UserInfo endpoints
- Supported grant types and response types
- JWKS endpoint location
- Supported scopes and claims
- ID token signing algorithms

## Related Documentation

- [OAuth 2.0 Authorization Server Metadata](authorization-server-metadata.md) - OAuth discovery

