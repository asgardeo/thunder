# OpenChoreo Deployment Guide

This guide provides comprehensive instructions for deploying Thunder on OpenChoreo platform using Helm charts, covering everything from prerequisites to production configurations.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Installation Methods](#installation-methods)
- [Database Setup](#database-setup)
- [Environment Management](#environment-management)

## Prerequisites

### Infrastructure Requirements

- **Kubernetes Cluster**: A running Kubernetes cluster (v1.19+) with OpenChoreo installed
  - [OpenChoreo v0.3.2](https://github.com/openchoreo/openchoreo) installed and configured
  - Proper RBAC permissions for OpenChoreo custom resources
- **Database**: PostgreSQL database (in-cluster or external)

### Required Tools

| Tool          | Installation Guide | Version Check Command|
|---------------|--------------------|-----------------------|
| Git           | [Install Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) | `git --version` |
| Helm          | [Install Helm](https://helm.sh/docs/intro/install/) | `helm version` |
| kubectl       | [Install kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) | `kubectl version` |
| OpenChoreo | [Install OpenChoreo](https://openchoreo.dev/docs/getting-started/quick-start-guide/) | `helm list -n openchoreo-control-plane -o json \| jq -r '.[] \| "\(.name) \(.app_version)"'` |

### Verify Prerequisites

```bash
# Check Kubernetes cluster access
kubectl cluster-info

# Verify Helm installation
helm version

# Check OpenChoreo installation
kubectl get crd | grep openchoreo
```

## Quick Start

Deploy Thunder on OpenChoreo with default settings in under 10 minutes:

### 1. Set Database Configuration

```bash
# Configure database connection (required)
export DB_HOST="postgres.default.svc.cluster.local"  # Your database host
export DB_USER="thunder_user"                        # Your database username
export DB_PASS="secure_password"                     # Your database password
```

### 2. Install Thunder

Clone the Thunder repository if you haven't already:

```bash
# Install Thunder using the local Helm chart
helm install thunder install/openchoreo/helm/ \
  --namespace identity-platform \
  --create-namespace \
  --set database.host="$DB_HOST" \
  --set database.identity.username="$DB_USER" \
  --set database.identity.password="$DB_PASS" \
  --set database.runtime.username="$DB_USER" \
  --set database.runtime.password="$DB_PASS" \
  --set organization.name="identity-platform"
```

### 3. Verify Installation

```bash
# Check OpenChoreo resources
kubectl get components,workloads,services -n identity-platform

# Check deployment status
kubectl get pods -n identity-platform

# Check organization and platform resources
kubectl get organizations,projects,deploymentpipelines,environments
```

## Installation Methods

### Method 1: Inline Value Overrides

```bash
# Install with custom database configuration
helm upgrade --install thunder install/openchoreo/helm/ \
  --namespace identity-platform \
  --create-namespace \
  --set database.host="postgres.example.com" \
  --set database.identity.username="thunder_user" \
  --set database.identity.password="secure_password" \
  --set database.runtime.username="thunder_user" \
  --set database.runtime.password="secure_password" \
  --set database.identity.sslmode="require" \
  --set database.runtime.sslmode="require" \
  --set organization.name="my-organization"
```

### Method 2: Custom Values File

```bash
# Create custom-values.yaml
cat > custom-values.yaml << EOF
# Component configuration
componentName: thunder-identity
pipelineName: identity-platform-pipeline

# Container image configuration
image:
  repository: ghcr.io/asgardeo/thunder
  tag: "0.11.0"

# Database configuration
database:
  host: postgres.example.com
  port: 5432
  identity:
    database: thunderdb
    username: thunder_user
    password: secure_identity_password
    type: postgres
    sslmode: require
  runtime:
    database: runtimedb
    username: thunder_user
    password: secure_runtime_password
    type: postgres
    sslmode: require
  user:
    database: userdb
    username: thunder_user
    password: secure_identity_password
    type: postgres
    sslmode: require

# JWT configuration
jwt:
  issuer: thunder-identity-platform
  validity: 7200  # 2 hours

# OAuth configuration
oauth:
  refresh_token_validity: 604800  # 7 days

# Cache configuration
cache:
  type: memory
  size: 50000
  ttl: 7200  # 2 hours

# CORS configuration
cors:
  allowed_origins:
    - "https://dev.your-domain.com"
    - "https://staging.your-domain.com"
    - "https://prod.your-domain.com"

# Gateway configuration
gateway:
  dnsPrefixDev: dev
  dnsPrefixStaging: staging
  dnsPrefixProd: prod

# Platform resources
organization:
  name: identity-platform
  displayName: Identity Platform Organization
  description: Thunder-powered identity management platform

# Cluster-scoped resources (only created in non-default namespaces)
serviceClass:
  name: default
  create: true

apiClass:
  name: default
  create: true
EOF

# Install with custom values
helm upgrade --install thunder install/openchoreo/helm/ \
  --namespace identity-platform \
  --create-namespace \
  -f custom-values.yaml
```

## Database Setup

Thunder requires PostgreSQL databases for both identity and runtime data.

### PostgreSQL Configuration

Before deploying Thunder, ensure you have:

1. **Created databases**:
   ```sql
   CREATE DATABASE thunderdb;
   CREATE DATABASE runtimedb;
   CREATE DATABASE userdb;
   ```
2. **Run database scripts**: Use the scripts in `backend/dbscripts` to initialize the schema.

3. **Created user**:
   ```sql
   CREATE USER thunder_user WITH PASSWORD 'secure_password';
   ```
4. Grant necessary privileges in all databases:
   ```sql
   ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO thunder_user;
   GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO thunder_user;
   ```

For detailed database setup using Bitnami PostgreSQL, refer to: [Bitnami PostgreSQL Helm Chart](https://bitnami.com/stacks/postgresql)

### Database Configuration Examples

#### External PostgreSQL

```yaml
database:
  host: postgres.example.com
  port: 5432
  identity:
    database: thunderdb
    username: thunder_user
    password: secure_password
    type: postgres
    sslmode: require
  runtime:
    database: runtimedb
    username: thunder_user
    password: secure_password
    type: postgres
    sslmode: require
  user:
    database: userdb
    username: thunder_user
    password: secure_password
    type: postgres
    sslmode: require
```

#### In-Cluster PostgreSQL

```yaml
database:
  host: postgres.default.svc.cluster.local
  port: 5432
  identity:
    database: thunderdb
    username: thunder_user
    password: secure_password
    type: postgres
    sslmode: disable
  runtime:
    database: runtimedb
    username: thunder_user
    password: secure_password
    type: postgres
    sslmode: disable
  user:
    database: userdb
    username: thunder_user
    password: secure_password
    type: postgres
    sslmode: disable
```

## Environment Management

OpenChoreo provides built-in environment management with promotion workflows:

### Available Environments

1. **Development**: For development and testing
2. **Staging**: For pre-production validation
3. **Production**: For live deployments

For additional help with OpenChoreo-specific features, refer to the [OpenChoreo](https://github.com/openchoreo/openchoreo) or open a discussion on [GitHub](https://github.com/asgardeo/thunder/discussions).
