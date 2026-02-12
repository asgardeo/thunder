# OAuth Standards-Based Authentication

Thunder supports OAuth 2.0 and OpenID Connect (OIDC) standards for authentication and authorization. This page provides a quick overview of available grant types. For comprehensive documentation, examples, and detailed guides, see the [Standards-Based documentation](../../standards-based/).

## Available OAuth Grant Types

Thunder supports the following OAuth 2.0/OIDC grant types:

- **[Authorization Code](../../standards-based/oauth2-oidc/grant-types/authorization-code.md)** - User authentication with secure server-side token exchange
- **[Client Credentials](../../standards-based/oauth2-oidc/grant-types/client-credentials.md)** - Machine-to-machine communication without user interaction
- **[Refresh Token](../../standards-based/oauth2-oidc/grant-types/refresh-token.md)** - Obtain new access tokens without re-authentication
- **[Token Exchange](../../standards-based/oauth2-oidc/grant-types/token-exchange.md)** - Exchange tokens for different audiences or scopes

## Additional Resources

For complete documentation on OAuth 2.0 and OIDC, including:

- **Endpoints**: Authorization, Token, Introspection, JWKS, UserInfo, and Discovery endpoints
- **Features**: Scopes, PKCE, Dynamic Client Registration, Public Clients, Resource Parameter
- **Examples**: Step-by-step guides with cURL commands and code samples

👉 **[View Complete Standards-Based Documentation →](../../standards-based/)**
