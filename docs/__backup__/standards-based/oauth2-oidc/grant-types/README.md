# OAuth 2.0 Grant Types

OAuth 2.0 grant types define how clients obtain access tokens. Thunder supports multiple grant types to accommodate different application types and use cases.

## Supported Grant Types

### [Authorization Code](authorization-code.md)

The Authorization Code grant type is the most secure OAuth 2.0 grant type for web applications. It involves redirecting users to an authorization server, where they authenticate and authorize the application.

**Use Cases:**
- Web applications with server-side components
- Applications that can securely store client secrets
- When user authorization is required
- With OpenID Connect for user authentication

**Key Features:**
- User authorization
- Support for PKCE (Proof Key for Code Exchange)
- Refresh token support

### [Client Credentials](client-credentials.md)

The Client Credentials grant type is used for machine-to-machine (M2M) communication where no user interaction is required.

**Use Cases:**
- Service-to-service communication
- Backend API access
- Microservices authentication
- Automated systems

**Key Features:**
- No user interaction required
- Simple token request
- Fast and efficient
- Suitable for server-to-server scenarios

### [Refresh Token](refresh-token.md)

The Refresh Token grant type allows clients to obtain new access tokens using a refresh token, without requiring the user to re-authenticate.

**Use Cases:**
- Long-lived sessions
- Mobile applications
- Applications requiring persistent access
- Reducing authorization frequency

**Key Features:**
- Extends access without re-authorization
- Scope downscoping support
- Configurable refresh token renewal
- Improved user experience

### [Token Exchange](token-exchange.md)

The Token Exchange grant type (RFC 8693) allows clients to exchange tokens for different audiences or scopes, enabling token delegation and impersonation scenarios.

**Use Cases:**
- Token delegation between services
- Impersonation scenarios
- Cross-service authorization
- Token transformation

**Key Features:**
- Token delegation support
- Audience transformation
- Scope modification
- RFC 8693 compliant

## Choosing the Right Grant Type

| Grant Type | User Interaction | Client Type | Use Case |
|------------|------------------|-------------|----------|
| Authorization Code | Required | Confidential | Web apps, user authorization |
| Client Credentials | Not required | Confidential | M2M, service-to-service |
| Refresh Token | Not required | Any | Token renewal |
| Token Exchange | Not required | Confidential | Token delegation |

## Security Considerations

- **Authorization Code**: Most secure for user-facing applications, especially with PKCE
- **Client Credentials**: Secure for M2M scenarios, requires secure client secret storage
- **Refresh Token**: Must be stored securely, supports scope downscoping
- **Token Exchange**: Requires careful validation of subject tokens and audiences

## 🔗 Related Documentation

- [OAuth 2.0 & OIDC Endpoints](../endpoints/) - Token and authorization endpoints
- [OAuth 2.0 & OIDC Features](../features/) - Scopes, resource parameter, PKCE, best practices, and more
