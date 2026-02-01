# OAuth 2.0 & OIDC Endpoints

Thunder provides standard OAuth 2.0 and OpenID Connect (OIDC) endpoints for authorization, token management, and user information.

## OAuth 2.0 Endpoints

- **[Authorization Endpoint](authorization-endpoint.md)** - `/oauth2/authorize` - Initiate authorization request
- **[Token Endpoint](token-endpoint.md)** - `/oauth2/token` - Exchange codes/tokens for access tokens
- **[Introspection Endpoint](introspection-endpoint.md)** - `/oauth2/introspect` - Validate and introspect tokens (RFC 7662)
- **[JWKS Endpoint](jwks-endpoint.md)** - `/oauth2/jwks` - JSON Web Key Set for token validation
- **[OAuth 2.0 Authorization Server Metadata](authorization-server-metadata.md)** - `/.well-known/oauth-authorization-server` - OAuth 2.0 discovery (RFC 8414)

## OIDC Endpoints

- **[UserInfo Endpoint](userinfo-endpoint.md)** - `/oauth2/userinfo` - Retrieve user identity information
- **[OpenID Connect Discovery](oidc-discovery.md)** - `/.well-known/openid-configuration` - OIDC discovery

## Base URL

All OAuth 2.0 and OIDC endpoints are available at:
```
https://localhost:8090/oauth2/{endpoint}
```

Discovery endpoints are available at:
```
https://localhost:8090/.well-known/{endpoint}
```

## Related Documentation

- [OAuth 2.0 & OIDC Grant Types](../grant-types/) - How to use these endpoints
- [OAuth 2.0 & OIDC Features](../features/) - Advanced features
