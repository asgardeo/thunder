# Thunder Documentation

Welcome to the Thunder documentation! This guide will help you understand and integrate Thunder's identity management capabilities into your applications.

## 📖 What is Thunder?

Thunder is a modern, open-source identity management service designed for teams building secure, customizable authentication experiences. It enables developers to design and orchestrate login, registration, and recovery flows using a flexible identity flow designer.

## 🗂️ Documentation Structure

This documentation is organized into the following sections:

### [Authentication](./guides/authentication/)

Learn how to implement authentication in your applications using Thunder:

- **[Standards-Based Authentication](./guides/authentication/standards-based/)** - OAuth 2.0, OpenID Connect (OIDC) flows including Authorization Code, Client Credentials, and Refresh Token
- **[Server Orchestrated Flows](./guides/authentication/server-orchestrated-flow/)** - App native authentication using server orchestrated flows
- **[Client Orchestrated Flows](./guides/authentication/client-orchestrated-flow/)** - App native authentication using client orchestrated flows

### [Registration](./guides/registration/)

Implement user registration and self-service onboarding:

- **[Server Orchestrated Flows](./guides/registration/server-orchestrated-flow/)** - Self-registration using server orchestrated flows

### [Identity Providers](./guides/identity-provider/)

Configure external identity providers for social login and enterprise SSO:

- **[Configure Identity Providers](./guides/identity-provider/configure-identity-providers.md)** - Set up Google, GitHub, and other identity providers

### [Notification Senders](./guides/notification-sender/)

Configure message senders for OTP and notifications:

- **[Configure Message Senders](./guides/notification-sender/configure-message-senders.md)** - Set up message notification senders

## 🚀 Getting Started

If you're new to Thunder, we recommend starting with these resources:

1. **[Main README](../README.md)** - Quickstart guide to download and run Thunder
2. **[API Documentation](../api/)** - Explore the RESTful APIs

## 📚 API Documentation

Thunder provides comprehensive RESTful APIs for managing identity and access:

- [User Management API](../api/user.yaml)
- [Application Management API](../api/application.yaml)
- [Authentication APIs](../api/authentication.yaml)
- [Flow Execution API](../api/flow.yaml)
- [Identity Provider Management API](../api/idp.yaml)
- [Notification Sender Management API](../api/notification-sender.yaml)
- [Group Management API](../api/group.yaml)
- [Organization Unit Management API](../api/ou.yaml)
- [Health Check API](../api/healthcheck.yaml)

## 🔧 Configuration

Refer to the `backend/cmd/server/repository/conf/deployment.yaml` file for detailed configuration options.

## 💡 Need Help?

- **Issues**: [GitHub Issues](https://github.com/asgardeo/thunder/issues)
- **Contributing**: See [CONTRIBUTING.md](../CONTRIBUTING.md)
- **License**: [Apache License 2.0](../LICENSE)

## 🤝 Contributing

We welcome contributions! Please refer to the [Contributing Guide](../CONTRIBUTING.md) for guidelines on how to contribute to Thunder.

---

**Note**: This documentation is actively maintained. If you find any issues or have suggestions, please open an issue on [GitHub](https://github.com/asgardeo/thunder/issues).
