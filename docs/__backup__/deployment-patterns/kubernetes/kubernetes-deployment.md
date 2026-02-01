# Kubernetes Deployment Guide

This guide provides comprehensive instructions for deploying Thunder in Kubernetes environments using Helm charts, covering everything from prerequisites to production configurations.

## Architecture Overview

![Thunder Kubernetes Architecture](./assets/images/thunder-kubernetes-diagram.png)

The diagram above shows the Thunder deployment architecture in Kubernetes, including the main application pods, ingress controller, and database configuration options.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Installation Methods](#installation-methods)
- [Database Setup](#database-setup)

## Prerequisites

### Infrastructure Requirements

- **Kubernetes Cluster**: A running Kubernetes cluster (v1.19+)
  - [minikube](https://kubernetes.io/docs/tasks/tools/#minikube) for local development
  - [kind](https://kind.sigs.k8s.io/) for local testing
  - Managed Kubernetes services (EKS, GKE, AKS) for production
- **Ingress Controller**: NGINX Ingress Controller (recommended) or alternative
- **TLS Certificates**: Valid SSL certificates for production deployments

### Required Tools

| Tool          | Installation Guide | Version Check Command|
|---------------|--------------------|-----------------------|
| Git           | [Install Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) | `git --version` |
| Helm          | [Install Helm](https://helm.sh/docs/intro/install/) | `helm version` |
| kubectl       | [Install kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) | `kubectl version` |
| Docker        | [Install Docker](https://docs.docker.com/engine/install/) | `docker --version` |

### Verify Prerequisites

```bash
# Check Kubernetes cluster access
kubectl cluster-info

# Verify Helm installation
helm version

# Check ingress controller (If available)
kubectl get pods -n ingress-nginx
```

## Quick Start

Deploy Thunder with default settings in under 5 minutes:



### 1. Install Thunder

```bash
# Install Thunder from GitHub Container Registry
helm install thunder oci://ghcr.io/asgardeo/helm-charts/thunder

# Or install a specific version
helm install thunder oci://ghcr.io/asgardeo/helm-charts/thunder --version 0.11.0
```

### 2. Verify Installation

```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/name=thunder

# Check services
kubectl get services -l app.kubernetes.io/name=thunder

# Check ingress
kubectl get ingress
```

### 3. Access Thunder

To access Thunder, you need to add the NGINX Ingress Controller's external IP to your local `/etc/hosts` file:

1. Get the external IP of your NGINX Ingress Controller
2. Add an entry to your `/etc/hosts` file mapping the IP to `thunder.local`
3. Access Thunder at: `http://thunder.local`

**Note**: If you're using a cloud provider, the external IP will be provided by the load balancer.

## Installation Methods

### Method 1: Inline Value Overrides

```bash

# Install with SQLite database (for development/testing)
helm install thunder oci://ghcr.io/asgardeo/helm-charts/thunder \
  --set configuration.database.identity.type=sqlite \
  --set configuration.database.runtime.type=sqlite
```

### Method 2: Custom Values File

```bash
# Create custom-values.yaml
cat > custom-values.yaml << EOF
deployment:
  replicaCount: 3
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 2
      memory: 1Gi

ingress:
  hostname: thunder.example.com

configuration:
  database:
    identity:
      type: postgres
      host: postgres.default.svc.cluster.local
      port: 5432
      name: thunderdb
      username: thunder_user
      password: secure_password
      sslmode: require
    runtime:
      type: postgres
      host: postgres.default.svc.cluster.local
      port: 5432
      name: runtimedb
      username: thunder_user
      password: secure_password
      sslmode: require
    user:
      type: postgres
      host: postgres.default.svc.cluster.local
      port: 5432
      name: userdb
      username: thunder_user
      password: secure_password
      sslmode: require
EOF

# Install with custom values
helm install thunder oci://ghcr.io/asgardeo/helm-charts/thunder -f custom-values.yaml
```

## Database Setup

Thunder supports both PostgreSQL and SQLite databases. PostgreSQL is recommended for production.


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


### PostgreSQL Configuration Example

```yaml
configuration:
  database:
    identity:
      type: postgres
      host: postgres.example.com
      port: 5432
      name: thunderdb
      username: thunder_user
      password: secure_password
      sslmode: require
    runtime:
      type: postgres
      host: postgres.example.com
      port: 5432
      name: runtimedb
      username: thunder_user
      password: secure_password
      sslmode: require
    user:
      type: postgres
      host: postgres.example.com
      port: 5432
      name: userdb
      username: thunder_user
      password: secure_password
      sslmode: require
```

### SQLite Configuration

```yaml
configuration:
  database:
    identity:
      type: sqlite
      sqlitePath: repository/database/thunderdb.db
      sqliteOptions: "_journal_mode=WAL&_busy_timeout=5000"
    runtime:
      type: sqlite
      sqlitePath: repository/database/runtimedb.db
      sqliteOptions: "_journal_mode=WAL&_busy_timeout=5000"
    user:
      type: sqlite
      sqlitePath: repository/database/userdb.db
      sqliteOptions: "_journal_mode=WAL&_busy_timeout=5000"  
```

### Update Strategy

```bash
# Rolling update
helm upgrade thunder oci://ghcr.io/asgardeo/helm-charts/thunder \
  --version 0.12.0 \
  -f production-values.yaml

# Rollback if needed
helm rollback thunder 1
```

For additional help, refer to the [Thunder documentation](../../README.md) or open a discussion on [GitHub](https://github.com/asgardeo/thunder/discussions).