# Authentication

Thunder provides flexible authentication capabilities to secure your applications. Whether you're building a web application, mobile app, or service-to-service integration, Thunder offers multiple authentication approaches to meet your needs.

## 🔐 Authentication Approaches

Thunder supports three main authentication approaches:

### 1. Standards-Based Authentication

Industry-standard OAuth 2.0 and OpenID Connect (OIDC) protocols for secure authentication and authorization.

**Use this when:**
- Building web applications that redirect users for authentication
- Implementing machine-to-machine (M2M) communication
- You need standards-compliant authentication
- Working with frameworks that support OAuth/OIDC

**Supported flows:**
- **Client Credentials** - Service-to-service authentication without user interaction
- **Authorization Code** - User authentication with secure server-side token exchange
- **Refresh Token** - Obtain new access tokens without re-authentication

[Learn more about Standards-Based Authentication →](./standards-based/oauth-authentication.md)

### 2. Server Orchestrated Flows (App Native Authentication)

REST API-driven authentication where Thunder orchestrates the entire authentication process step-by-step.

**Use this when:**
- Building native mobile applications (iOS, Android)
- Developing single-page applications (SPAs) with custom UI
- Implementing multi-step authentication flows

**Supported authentication methods:**
- **Username and Password** - Traditional credentials-based login
- **SMS OTP** - One-time password authentication via SMS
- **Social Login** - Google, GitHub authentication
- **Multi-factor Authentication** - Combine multiple authenticators

[Learn more about Server Orchestrated Flows →](./server-orchestrated-flow/authentication.md)
[Learn more about flow customization →](./server-orchestrated-flow/customize-auth-flow.md)

### 3. Client Orchestrated Flows

Client-side authentication control where the application manages the authentication flow.

**Use this when:**
- You need complete control over the authentication logic
- Building custom authentication experiences
- Implementing advanced security patterns

[Learn more about Client Orchestrated Flows →](./client-orchestrated-flow/authentication.md)
