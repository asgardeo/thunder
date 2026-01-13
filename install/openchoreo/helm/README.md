# Thunder Helm Chart

This Helm chart deploys WSO2 Thunder Identity Management Service on OpenChoreo platform. Thunder is a comprehensive identity and access management solution that provides OAuth2, OpenID Connect, and other identity protocols.

## Overview

The chart creates the following OpenChoreo resources:
- **Component**: Defines the Thunder identity management service
- **Workload**: Configures the container deployment
- **Service**: Exposes the Thunder API endpoints
- **Organization**: Creates or references an organization
- **ServiceClass/APIClass**: Cluster-scoped resources for service management

## Configuration Value Types

Thunder's configuration system supports multiple value formats for **any parameter**:

1. **Direct Values** - Static values specified directly in YAML:
   ```yaml
   database:
     host: "postgres.default.svc.cluster.local"
     port: 5432
   cache:
     size: 10000
   ```

2. **Environment Variables** - Use Go template syntax `{{.VARIABLE_NAME}}` to reference environment variables:
   ```yaml
   database:
     identity:
       password: "{{.DB_PASSWORD}}"
   jwt:
     issuer: "{{.JWT_ISSUER}}"
   ```

3. **File References** - Use `file://` protocol to load content from files:
   ```yaml
   crypto:
     encryption:
       key: "file://repository/resources/security/crypto.key"
   ```
   Supports both quoted and unquoted paths:
   - `file://path/to/file` - Unquoted path (no spaces)
   - `file://"path/with spaces"` - Quoted path (with spaces allowed)
   - `file:///absolute/path` - Absolute paths
   - `file://relative/path` - Relative paths (resolved from the Thunder container's home directory)

## Quick Start

### Prerequisites

- Kubernetes cluster with OpenChoreo installed
- Helm 3.x
- PostgreSQL database (in-cluster or external)
- Proper RBAC permissions for OpenChoreo resources

### OpenChoreo version
- 0.3.2

### Basic Installation

1. **Configure database connection** (required):
   ```bash
   export DB_HOST="<your-database-host>"      # Your database host
   export DB_USER="<your-database-username>"  # Your database username
   export DB_PASS="<your-database-password>"  # Your database password
   ```

2. **Install the chart**:
   ```bash
   helm upgrade --install thunder install/openchoreo/helm/ \
     --namespace identity-platform \
     --create-namespace \
     --set database.host="$DB_HOST" \
     --set database.identity.username="$DB_USER" \
     --set database.identity.password="$DB_PASS" \
     --set database.runtime.username="$DB_USER" \
     --set database.runtime.password="$DB_PASS" \
     --set organization.name="identity-platform"
   ```

## Chart Location

- **Chart**: `install/openchoreo/helm`
- **Values**: `install/openchoreo/helm/values.yaml`
- **Templates**: `install/openchoreo/helm/templates/`

## Configuration Values

### Core Settings

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `componentName` | Base name for Component/Workload/Service resources | `thunder-identity` | No |
| `pipelineName` | DeploymentPipeline name (used by platform templates) | `identity-platform-pipeline` | No |
| `image.repository` | Thunder container image repository | `ghcr.io/asgardeo/thunder` | No |
| `image.tag` | Container image tag | `latest` | No |
| `thunder.server.port` | Port on which Thunder server listens | `8090` | No |

### Database Configuration

**⚠️ Required**: Replace placeholder values `<DB_HOST>`, `<DB_USERNAME>`, `<DB_PASSWORD>` with actual values.

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `database.host` | Database hostname/FQDN | `<DB_HOST>` | **Yes** |
| `database.port` | Database port | `5432` | No |
| `database.identity.database` | Identity database name | `thunderdb` | No |
| `database.identity.username` | Identity database username | `<DB_USERNAME>` | **Yes** |
| `database.identity.password` | Identity database password | `<DB_PASSWORD>` | **Yes** |
| `database.runtime.database` | Runtime database name | `runtimedb` | No |
| `database.runtime.username` | Runtime database username | `<DB_USERNAME>` | **Yes** |
| `database.runtime.password` | Runtime database password | `<DB_PASSWORD>` | **Yes** |
| `database.user.database` | User database name | `userdb` | No |
| `database.user.username` | User database username | `<DB_USERNAME>` | **Yes** |
| `database.user.password` | User database password | `<DB_PASSWORD>` | **Yes** |

### Authentication & Security

**⚠️ Required**: For any non-test deployment, override `crypto.encryption.key` with a 32-byte (64 character) hex key. Do not use the default value in production.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `jwt.issuer` | JWT token issuer identifier (derived from server URL if not set) | derived |
| `jwt.validity` | JWT token validity in seconds | `3600` (1 hour) |
| `oauth.refresh_token_validity` | Refresh token validity in seconds | `86400` (24 hours) |
| `crypto.encryption.key` | Crypto encryption key | `file://repository/resources/security/crypto.key` |
| `cors.allowed_origins` | List of allowed CORS origins | See values.yaml |

### Cache Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `cache.type` | Cache type (currently only "memory" supported) | `memory` |
| `cache.size` | Maximum number of cache entries | `10000` |
| `cache.ttl` | Cache entry TTL in seconds | `3600` (1 hour) |

### Platform Resources

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceClass.name` | ServiceClass resource name to reference | `default` |
| `serviceClass.create` | Whether to create ServiceClass resource | `true` |
| `apiClass.name` | APIClass resource name to reference | `default` |
| `apiClass.create` | Whether to create APIClass resource | `true` |
| `organization.name` | Organization name (must match project references) | `identity-platform` |
| `organization.displayName` | Human-readable organization name | `Default Organization` |

### Gateway Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `gateway.dnsPrefixDev` | DNS prefix for development environment | `dev` |
| `gateway.dnsPrefixStaging` | DNS prefix for staging environment | `staging` |
| `gateway.dnsPrefixProd` | DNS prefix for production environment | `prod` |

## Namespace and Resource Management

### Cluster-Scoped Resources

- **ServiceClass** and **APIClass** are cluster-scoped custom resources
- By default, these resources are created only when deploying to a **non-default** namespace
- This prevents conflicts with system-level resources in the default namespace

### Resource Creation Behavior

| Resource | Namespace | Created |
|----------|-----------|---------|
| ServiceClass | `default` | ❌ No (regardless of `create` setting) |
| ServiceClass | `non-default` | ✅ Yes (if `serviceClass.create=true`) |
| APIClass | `default` | ❌ No (regardless of `create` setting) |
| APIClass | `non-default` | ✅ Yes (if `apiClass.create=true`) |

### Using Existing Resources

If you already have cluster-scoped resources, reference them instead of creating new ones:

```bash
--set serviceClass.create=false \
--set serviceClass.name=existing-service-class \
--set apiClass.create=false \
--set apiClass.name=existing-api-class
```

### Template and Validate

```bash
# Render templates locally to inspect generated manifests
helm template thunder install/openchoreo/helm/ \
  --namespace identity-platform \
  --set database.host="$DB_HOST" \
  --set database.identity.username="$DB_USER" \
  --set database.identity.password="$DB_PASS" \
  --set database.runtime.username="$DB_USER" \
  --set database.runtime.password="$DB_PASS" \
  --set organization.name="identity-platform"

# Dry-run installation to check for issues
helm upgrade --install thunder install/openchoreo/helm/ \
  --namespace identity-platform \
  --create-namespace \
  --dry-run \
  --set database.host="$DB_HOST" \
  --set database.identity.username="$DB_USER" \
  --set database.identity.password="$DB_PASS" \
  --set database.runtime.username="$DB_USER" \
  --set database.runtime.password="$DB_PASS" \
  --set organization.name="identity-platform"
```

### Debugging Commands

```bash
# Check pod status and logs
kubectl get pods -n identity-platform

# View logs for a Thunder pod (replace <pod-name> with actual pod name)
kubectl logs <pod-name> -n identity-platform

# Check OpenChoreo resources
kubectl get components,workloads,services -n identity-platform
kubectl get organizations,serviceclasses,apiclasses
```

## Security Considerations

- 🔒 **Never use default passwords in production**
- 🌐 **Configure CORS origins restrictively**
- 🔑 **Use strong JWT and OAuth settings**
- 🛡️ **Enable SSL/TLS for database connections in production**

## Contributing

For questions, support, or to contribute improvements to this Helm chart:

- 📋 Open an issue in the [Thunder GitHub repository](https://github.com/asgardeo/thunder)
- 📖 Refer to the project's [CONTRIBUTING guidelines](../../../CONTRIBUTING.md)  
- 💬 Join the community discussions
- 🐛 Report bugs or security issues through proper channels
