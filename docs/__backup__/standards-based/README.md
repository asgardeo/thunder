# Standards-Based Protocols

Thunder implements industry-standard OAuth 2.0 and OpenID Connect (OIDC) protocols for secure authentication and authorization. This section provides comprehensive documentation for all standards-based protocol and features.

## 📚 Contents

### [OAuth 2.0 & OpenID Connect (OIDC)](oauth2-oidc/)

OAuth 2.0 is an authorization framework that enables applications to obtain limited access to resources. OpenID Connect (OIDC) is an authentication layer built on top of OAuth that enables clients to verify the identity of users.

**Supported Grant Types:**
- [Authorization Code](oauth2-oidc/grant-types/authorization-code.md) - User authorization with token for access delegation
- [Client Credentials](oauth2-oidc/grant-types/client-credentials.md) - Machine-to-machine communication without user interaction
- [Refresh Token](oauth2-oidc/grant-types/refresh-token.md) - Obtain new access tokens without user interaction
- [Token Exchange](oauth2-oidc/grant-types/token-exchange.md) - Exchange tokens for different use cases

**Endpoints:**
- [Authorization Endpoint](oauth2-oidc/endpoints/authorization-endpoint.md) - `/oauth2/authorize` - Initiate authorization request
- [Token Endpoint](oauth2-oidc/endpoints/token-endpoint.md) - `/oauth2/token` - Exchange codes/tokens for access tokens
- [Introspection Endpoint](oauth2-oidc/endpoints/introspection-endpoint.md) - `/oauth2/introspect` - Validate and introspect tokens
- [JWKS Endpoint](oauth2-oidc/endpoints/jwks-endpoint.md) - `/oauth2/jwks` - JSON Web Key Set for token validation
- [UserInfo Endpoint](oauth2-oidc/endpoints/userinfo-endpoint.md) - `/oauth2/userinfo` - Retrieve user identity information (OIDC)
- [OAuth 2.0 Authorization Server Metadata](oauth2-oidc/endpoints/authorization-server-metadata.md) - `/.well-known/oauth-authorization-server` - OAuth 2.0 discovery
- [OpenID Connect Discovery](oauth2-oidc/endpoints/oidc-discovery.md) - `/.well-known/openid-configuration` - OIDC discovery

**Features:**
- [Scopes and Permissions](oauth2-oidc/features/scopes-permissions.md) - Custom scopes and fine-grained access control
- [Resource Parameter](oauth2-oidc/features/resource-parameter.md) - Specify target resource/audience for tokens
- [Dynamic Client Registration](oauth2-oidc/features/dynamic-client-registration.md) - Register OAuth clients dynamically (`/oauth2/dcr/register`)
- [Public Clients](oauth2-oidc/features/public-clients.md) - Configure public clients for mobile and SPA applications
- [PKCE](oauth2-oidc/features/pkce.md) - Proof Key for Code Exchange for enhanced security


> 💡 **Securing MCP Servers with Thunder**
> 
> Learn how to secure Model Context Protocol (MCP) servers using Thunder's OAuth 2.0 implementation, following the [official MCP Authorization Specification](https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization). This comprehensive guide covers the complete authorization process including Dynamic Client Registration (DCR), PKCE, and resource parameter usage.
> 
> **[📖 Securing MCP Servers with Thunder Guide →](oauth2-oidc/features/mcp-server-securing.md)**

## 🚀 Quick Start

If you're new to OAuth/OIDC, we recommend starting with:

1. **[Authorization Code](oauth2-oidc/grant-types/authorization-code.md)** - Most common grant type for web applications
2. **[Client Credentials](oauth2-oidc/grant-types/client-credentials.md)** - For machine-to-machine communication
3. **[UserInfo Endpoint](oauth2-oidc/endpoints/userinfo-endpoint.md)** - Get user identity information

## 🔗 Related Documentation

- [Authentication Guides](../authentication/) - Server-orchestrated and client-orchestrated flows
- [Authorization Guides](../authorization/) - Flow-based authorization and RBAC
- [API Documentation](/api/) - Complete API reference

## 💡 Need Help?

- **Issues**: [GitHub Issues](https://github.com/asgardeo/thunder/issues)
- **Contributing**: See [CONTRIBUTING.md](../../community/contributing/README.md)
