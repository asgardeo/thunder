# Thunder Documentation

Welcome to the Thunder documentation! This guide will help you understand and integrate Thunder's identity management capabilities into your applications.

## 📖 What is Thunder?

Thunder is a modern, open-source identity management service designed for teams building secure, customizable authentication experiences. It enables developers to design and orchestrate login, registration, and recovery flows using a flexible identity flow designer.

## Accessing the System APIs of Thunder

System APIs of Thunder are secured by default. You need to obtain an access token with `system` scope by authenticating as an admin user. Follow these steps:

1. **Authenticate and obtain an access token:**

   Replace `<application_id>` with the sample app ID generated during "Setup the product."

   ```bash
   curl -k -X POST 'https://localhost:8090/flow/execute' \
     -d '{"applicationId":"<application_id>","flowType":"AUTHENTICATION", "inputs":{"username":"admin","password":"admin", "requested_permissions":"system"}}'
   ```

2. **Extract the assertion from the response:**

   The response will contain an `assertion` field:
   ```json
   {"flowId":"<flow_id>","flowStatus":"COMPLETE","data":{},"assertion":"<assertion>"}
   ```

3. **Use the assertion as a Bearer token:**

   Use this assertion value as the Bearer token in the `Authorization` header when calling Thunder management APIs:
   
   ```bash
   curl -kL -H 'Authorization: Bearer <assertion>' https://localhost:8090/applications
   ```

   > **Note:** In the API examples throughout this documentation, replace `<token>` with the assertion value you obtained from step 2.

## 🗂️ Documentation Structure

This documentation is organized into the following sections:

### [Standards-Based Flows](standards-based/)

Comprehensive documentation for OAuth 2.0 and OpenID Connect (OIDC) standards:

- **[OAuth 2.0 & OpenID Connect (OIDC)](standards-based/)** - Grant types, endpoints, and features for OAuth 2.0 and OIDC with samples

### [Authentication](authentication/)

Learn how to implement authentication in your applications using Thunder:

- **[Standards-Based Authentication](authentication/standards-based/)** - Quick reference for OAuth 2.0 and OIDC (see [Standards-Based Flows](standards-based/) for comprehensive documentation)
- **[Server Orchestrated Flows](authentication/server-orchestrated-flow/)** - App native authentication using server orchestrated flows
- **[Client Orchestrated Flows](authentication/client-orchestrated-flow/)** - App native authentication using client orchestrated flows

### [Registration](registration/)

Implement user registration and self-service onboarding:

- **[Server Orchestrated Flows](registration/server-orchestrated-flow/)** - Self-registration using server orchestrated flows

### [Identity Providers](identity-provider/)

Configure external identity providers for social login and enterprise SSO:

- **[Configure Identity Providers](identity-provider/configure-identity-providers.md)** - Set up Google, GitHub, and other identity providers

### [Notification Senders](notification-sender/)

Configure message senders for OTP and notifications:

- **[Configure Message Senders](notification-sender/configure-message-senders.md)** - Set up message notification senders

### [Immutable Configurations](immutable-configurations/)

Manage Thunder configurations and deployments:

- **[Export Configurations](immutable-configurations/export-configurations.md)** - Export applications and other resources as YAML files for version control and backup
- **[Immutable Configuration Mode](immutable-configurations/immutable-configuration.md)** - Run Thunder with file-based configurations for GitOps workflows

### [Deployment Patterns](./deployment-patterns/)

Deploy Thunder in various environments:

- **[Docker Deployment](./deployment-patterns/docker/docker-deployment.md)** - Guide for running Thunder using Docker
- **[Kubernetes Deployment](./deployment-patterns/kubernetes/kubernetes-deployment.md)** - Complete guide for deploying Thunder on Kubernetes using Helm charts
- **[OpenChoreo Deployment](./deployment-patterns/openchoreo/openchoreo-deployment.md)** - Guide for deploying Thunder on OpenChoreo platform

## 🚀 Getting Started

If you're new to Thunder, we recommend starting with these resources:

1. **[Main README](/README.md)** - Quickstart guide to download and run Thunder
2. **[API Documentation](/api/)** - Explore the RESTful APIs

## 📚 API Documentation

Thunder provides comprehensive RESTful APIs for managing identity and access:

- [User Management API](/api/user.yaml)
- [Application Management API](/api/application.yaml)
- [Authentication APIs](/api/authentication.yaml)
- [Flow Execution API](/api/flow.yaml)
- [Identity Provider Management API](/api/idp.yaml)
- [Notification Sender Management API](/api/notification-sender.yaml)
- [Group Management API](/api/group.yaml)
- [Organization Unit Management API](/api/ou.yaml)

## 🔧 Configuration

Refer to the `backend/cmd/server/repository/conf/deployment.yaml` file for detailed configuration options.

## 💡 Need Help?

- **Issues**: [GitHub Issues](https://github.com/asgardeo/thunder/issues)
- **Contributing**: See [CONTRIBUTING.md](/docs/contributing/README.md)
- **License**: [Apache License 2.0](/LICENSE)

## 🤝 Contributing

We welcome contributions! Please refer to the [Contributing Guide](/docs/contributing/README.md) for guidelines on how to contribute to Thunder.

---

**Note**: This documentation is actively maintained. If you find any issues or have suggestions, please open an issue on [GitHub](https://github.com/asgardeo/thunder/issues).
