# OpenChoreo Deployment Guide

This guide provides comprehensive instructions for deploying Thunder on OpenChoreo platform using Helm charts, covering everything from prerequisites to multi-environment promotion.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Database Setup](#database-setup)
- [Configuration](#configuration)
- [Environment Management and Promotion](#environment-management-and-promotion)
- [Accessing Thunder](#accessing-thunder)
- [Debugging](#debugging)

## Prerequisites

### Infrastructure Requirements

- **Kubernetes Cluster**: A running Kubernetes cluster (v1.19+) with OpenChoreo 1.0.0 installed and configured
- **`ClusterDataPlane`**: A `ClusterDataPlane` resource provisioned by the OpenChoreo platform installation — run `kubectl get clusterdataplane` to verify
- **Database**: PostgreSQL database (in-cluster or external)

### Required Tools

| Tool | Installation Guide | Version Check Command |
|------|--------------------|-----------------------|
| Git | [Install Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) | `git --version` |
| Helm | [Install Helm](https://helm.sh/docs/intro/install/) | `helm version` |
| `kubectl` | [Install kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) | `kubectl version` |
| OpenChoreo | [Install OpenChoreo](https://openchoreo.dev/docs/getting-started/quick-start-guide/) | `kubectl get crd \| grep openchoreo` |

### Verify Prerequisites

```bash
# Check Kubernetes cluster access
kubectl cluster-info

# Verify a ClusterDataPlane is available
kubectl get clusterdataplane

# Confirm OpenChoreo CRDs are installed
kubectl get crd | grep openchoreo
```

## Architecture

The chart uses OpenChoreo's **ComponentType** architecture to manage the full lifecycle of Thunder — from installation to multi-environment promotion. A single `helm install` fully deploys Thunder with no manual steps required.

```
ComponentType (thunder-idp)
  └── defines: parameters schema, environmentConfigs schema, K8s resource templates
        │
        ├── Component          — component-level parameters (crypto key, JWT, cache, etc.)
        ├── Workload           — container image
        ├── ComponentRelease   — frozen snapshot (image + parameters), created via post-install hook
        └── ReleaseBinding     — binds release to development environment with per-env configs
              └── generates → ConfigMap + Deployment + Service + HTTPRoute in data plane namespace
```

### Install-Time Hook Order

```
pre-install hooks
  ├── (weight -10) setup-config ConfigMap  ← DB connection config for the setup job
  └── (weight  -5) setup Job               ← runs ./setup.sh to initialise DB schemas

main resources
  └── Namespace, Project, DeploymentPipeline, Environments, ComponentType, Component, Workload

post-install hooks
  ├── (weight  0) ComponentRelease  ← frozen snapshot of image + parameters
  └── (weight  1) ReleaseBinding    ← injects environmentConfigs, triggers reconciliation
        └── OpenChoreo reconciles → Deployment + Service + HTTPRoute in dp-* namespace
```

### Parameters vs Environment Configurations

Thunder configuration is split into two categories:

- **`parameters`** — frozen at release time, identical across all environments. Set via `helm install --set`. Cannot be changed after a `ComponentRelease` is created without cutting a new release.
- **`environmentConfigs`** — per-environment values. Set for `development` by the Helm chart. Configurable at promotion time via the Backstage portal.

Fields visible in the promotion UI:

| Field | Description |
|-------|-------------|
| `configDbHostname` | Config database hostname |
| `configDbName` | Config database name |
| `configDbUsername` | Config database username |
| `configDbPassword` | Config database password |
| `runtimeDbHostname` | Runtime database hostname |
| `runtimeDbName` | Runtime database name |
| `runtimeDbUsername` | Runtime database username |
| `runtimeDbPassword` | Runtime database password |
| `userDbHostname` | User database hostname |
| `userDbName` | User database name |
| `userDbUsername` | User database username |
| `userDbPassword` | User database password |
| `serverPublicUrl` | Thunder public-facing URL |
| `gateClientHostname` | Gate service hostname |
| `gateClientPort` | Gate service port |
| `gateClientScheme` | Gate connection scheme (`http` or `https`) |
| `corsAllowedOrigins` | Allowed CORS origins (array) |
| `replicas` | Number of pod replicas |

## Quick Start

### 1. Set Required Values

```bash
export COMPONENT_NAME="my-thunder"
export DB_HOST="postgres.example.com"
export DB_NAME="thunderdb"
export DB_USER="asgthunder"
export DB_PASS="<your-database-password>"
export SERVER_PUBLIC_URL="http://my-thunder-development.openchoreoapis.localhost:19080"
export GATE_HOSTNAME="my-thunder-development.openchoreoapis.localhost"
```

### 2. Install Thunder

```bash
helm upgrade --install "$COMPONENT_NAME" install/openchoreo/helm/ \
  --namespace identity-platform \
  --create-namespace \
  --set componentName="$COMPONENT_NAME" \
  --set database.host="$DB_HOST" \
  --set database.config.database="$DB_NAME" \
  --set database.config.username="$DB_USER" \
  --set database.config.password="$DB_PASS" \
  --set database.runtime.database="$DB_NAME" \
  --set database.runtime.username="$DB_USER" \
  --set database.runtime.password="$DB_PASS" \
  --set database.user.database="$DB_NAME" \
  --set database.user.username="$DB_USER" \
  --set database.user.password="$DB_PASS" \
  --set serverPublicUrl="$SERVER_PUBLIC_URL" \
  --set gate.hostname="$GATE_HOSTNAME" \
  --set organization.name=identity-platform
```

### 3. Verify Installation

```bash
# Check OpenChoreo resource status
kubectl get componenttype,component,workload,componentrelease,releasebinding -n identity-platform

# Find the Thunder pod (deployed to a data plane namespace)
kubectl get pod -A | grep thunder
```

### Custom Values File

For repeated deployments, create a `custom-values.yaml`:

```yaml
componentName: my-thunder

image:
  repository: ghcr.io/asgardeo/thunder
  tag: "0.11.0"

database:
  host: postgres.example.com
  port: 5432
  config:
    database: configdb
    username: asgthunder
    password: secure_password
    sslmode: disable
  runtime:
    database: runtimedb
    username: asgthunder
    password: secure_password
    sslmode: disable
  user:
    database: userdb
    username: asgthunder
    password: secure_password
    sslmode: disable

serverPublicUrl: "http://my-thunder-development.openchoreoapis.localhost:19080"

gate:
  hostname: "my-thunder-development.openchoreoapis.localhost"
  port: 19080
  scheme: http
  clientBase: "/gate"

console:
  clientBase: "/console"
  clientId: "CONSOLE"
  scopes: '["openid", "profile", "email"]'

jwt:
  issuer: ""
  validity: 3600

oauth:
  refresh_token_validity: 86400

cache:
  size: 10000
  ttl: 3600

cors:
  allowed_origins:
    - "https://thunder-gate.your-domain.com"

organization:
  name: identity-platform

dataPlane:
  name: default

replicas: 1
```

Install with the values file:

```bash
helm upgrade --install my-thunder install/openchoreo/helm/ \
  --namespace identity-platform \
  --create-namespace \
  -f custom-values.yaml
```

## Configuration

For the full configuration reference, see the [Helm chart README](../../../install/openchoreo/helm/README.md).

Key categories:

| Category | Key Parameters |
|----------|----------------|
| Core | `componentName`, `image.repository`, `image.tag`, `thunder.server.port` |
| Database | `database.host`, `database.config.*`, `database.runtime.*`, `database.user.*` |
| Public Access | `serverPublicUrl`, `gate.hostname`, `gate.port`, `gate.scheme` |
| Frontend | `gate.clientBase`, `console.clientBase`, `console.clientId`, `console.scopes` |
| Security | `jwt.issuer`, `jwt.validity`, `crypto.encryption.key` |
| Cache | `cache.size`, `cache.ttl` |
| CORS | `cors.allowed_origins` |
| Platform | `organization.name`, `dataPlane.name`, `replicas` |

## Environment Management and Promotion

### Available Environments

The chart creates three environments in a linear promotion pipeline:

```
development  →  staging  →  production
```

Thunder is automatically deployed to `development` via post-install hooks after `helm install`.

### Promoting to Staging or Production

1. Open the Backstage portal and navigate to the Thunder component.
2. Click **Promote** on the `development` `ReleaseBinding`.
3. In the promotion UI, fill in the environment-specific values for the target environment. All database connection details, URLs, and CORS origins are configurable per environment:
   - Database host names, names, usernames, and passwords for `config`, `runtime`, and `user` databases
   - `serverPublicUrl` — public Thunder URL for the target environment
   - `gateClientHostname`, `gateClientPort`, `gateClientScheme` — gate service location
   - `corsAllowedOrigins` — allowed origins for the target environment
   - `replicas` — desired replica count
4. Confirm the promotion. OpenChoreo will reconcile the resources into the data plane `Namespace` for the target environment.

> **Note**: Setup Job defined in the `ComponentType` runs `./setup.sh` automatically for each promoted environment, initialising schemas before the Deployment starts.

## Accessing Thunder

Once the `ReleaseBinding` is active and the pod is running, Thunder is accessible via the `HTTPRoute` created by OpenChoreo.

The URL pattern is:

```
http://<componentName>-<environmentName>.<gateway-domain>:<gateway-port>
```

For example, with the default local development setup:

```
http://my-thunder-development.openchoreoapis.localhost:19080
```

Verify the `HTTPRoute` is active:

```bash
kubectl get httproute -A | grep thunder
```

Test the health endpoint:

```bash
curl http://my-thunder-development.openchoreoapis.localhost:19080/api/v1/health
```

The Gate and Console frontend applications are served under their configured base paths:

- Gate: `http://my-thunder-development.openchoreoapis.localhost:19080/gate`
- Console: `http://my-thunder-development.openchoreoapis.localhost:19080/console`

## Debugging

```bash
# Check all OpenChoreo resource statuses
kubectl get componenttype,component,workload,componentrelease,releasebinding -n identity-platform

# Find the Thunder pod (in the dp-* data plane namespace)
kubectl get pod -A | grep thunder

# Check Thunder logs
kubectl logs <pod-name> -n <dp-namespace>

# Inspect the rendered Thunder configuration
kubectl get configmap <componentName>-config -n <dp-namespace> \
  -o jsonpath='{.data.deployment\.yaml}'

# Check the setup job logs (pre-install, in control plane namespace)
kubectl logs job/<componentName>-setup -n identity-platform

# Check the setup job logs (in data plane namespace, runs at each promotion)
kubectl logs job/<componentName>-setup -n <dp-namespace>
```

For additional help, refer to the [OpenChoreo documentation](https://github.com/openchoreo/openchoreo) or open a discussion on [GitHub](https://github.com/asgardeo/thunder/discussions).
