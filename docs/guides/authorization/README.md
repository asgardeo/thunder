# Authorization

Thunder provides flexible authorization capabilities to control access to resources in your applications. Whether you're building fine-grained access control for web applications, mobile apps, or service-to-service integrations, Thunder offers role-based access control (RBAC) with multiple integration approaches to meet your needs.

## 📚 Getting Started

Before implementing authorization, you need to understand these core concepts:

### 1. Resource Servers and Permissions

Define your permission model by setting up resource servers, resources, and actions.

**Resource servers** represent your applications or services (e.g., "Booking API", "Payment Service"). Each resource server defines the permissions that can be granted through roles.

[Learn more about Resource Server Management →](./resource-server-management.md)

### 2. Role-Based Access Control (RBAC)

Create roles and assign permissions to users or groups.

**Roles** are collections of permissions grouped by resource server, enabling fine-grained access control across your applications.

[Learn more about Role-Based Access Control →](./role-based-access-control.md)

## 🔒 Authorization Integration

Once you've set up resource servers and roles, integrate authorization into your applications:

### Flow-Based Authorization

Authorization integrated directly into authentication flows, where permissions are evaluated and included in authentication assertions. Ideal for native mobile apps and SPAs.

[Learn more about Flow-Based Authorization →](./flow-based-authorization.md)

### OAuth-Based Authorization with Custom Scopes

Standards-compliant OAuth 2.0 authorization with custom permission scopes that extend beyond OpenID Connect scopes. Perfect for third-party integrations.

[Learn more about OAuth Scopes and Permissions →](./oauth-scopes-permissions.md)

## 🎯 Key Concepts

- **Resource Servers** - Applications or services that define permissions
- **Resources** - Hierarchical organization of entities (optional)
- **Actions** - Operations that can be performed (create, read, update, delete, etc.)
- **Permissions** - Derived strings from resources and actions (e.g., "reservations:create")
- **Roles** - Named sets of permissions grouped by resource server
- **Assignments** - Associations between roles and users/groups
- **Organization Units** - Containers that scope resource servers and roles
