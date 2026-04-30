# Thunder Helm Chart

This Helm chart deploys WSO2 Thunder Identity Management Service on the OpenChoreo platform. Thunder provides OAuth2, OpenID Connect, and other identity protocols.

## Overview

### Chart Architecture

The umbrella chart contains two independent sub-charts:

| Sub-chart | Purpose | Who installs |
|-----------|---------|--------------|
| `thunder-oc-componenttype` | Registers the `ClusterComponentType` (or namespace-scoped `ComponentType`) that defines Thunder's full schema and Kubernetes resource templates | Platform admins — once per cluster |
| `thunder-component` | Deploys the `Component`, `Workload`, `ComponentRelease`, `ReleaseBinding`, and all platform resources for a Thunder instance | Teams — once per Thunder deployment |

Install both together (default):
```bash
helm install thunder .
```

Install each independently:
```bash
# Platform admin: register the ComponentType once
helm install thunder-type charts/thunder-oc-componenttype

# Team: deploy a Thunder instance (ComponentType must already exist)
helm install thunder-app charts/thunder-component -f my-values.yaml
```

### OpenChoreo Resource Flow

```
thunder-oc-componenttype (ClusterComponentType)
  └── defines: parameters schema, environmentConfigs schema, K8s resource templates
        │
thunder-component
  ├── Component          ← component-level parameters (image, dbType, crypto key, etc.)
  ├── Workload           ← container image + env vars
  ├── ComponentRelease   ← frozen snapshot (post-install hook, weight 0)
  └── ReleaseBinding     ← binds release to development with per-env configs (weight 1)
        └── OpenChoreo reconciles → resources rendered in the data plane namespace
```

### Install-Time Lifecycle

```
main resources
  └── Namespace, Project, DeploymentPipeline, Environments
  └── ComponentType/ClusterComponentType, Component, Workload

post-install hooks (weight 0 → 1)
  └── ComponentRelease        ← frozen snapshot of image + parameters
  └── ReleaseBinding          ← binds release to development, injects environmentConfigs
        └── OpenChoreo reconciles → data plane namespace
              ├── sqlite-pvc        ← PVC for SQLite data (always created)
              ├── setup-job         ← initialises DB schemas, writes .setup-complete marker
              ├── thunder-config    ← deployment.yaml ConfigMap (sqlite or postgres variant)
              ├── gate-config       ← Gate frontend config.js
              ├── console-config    ← Console frontend config.js
              ├── file-config (×N)  ← one ConfigMap per attached configuration file
              ├── Deployment        ← wait-for-setup init container + thunder container
              ├── Service           ← ClusterIP on the Thunder server port
              └── HTTPRoute         ← external ingress (when endpointVisibility: external)
```

### ComponentType Managed Resources

| Resource ID | Kind | Description |
|-------------|------|-------------|
| `sqlite-pvc` | `v1/PersistentVolumeClaim` | Shared PVC for SQLite data files and setup markers. Always created regardless of `dbType`. |
| `setup-job` | `batch/v1/Job` | Runs `./setup.sh` to initialise DB schemas. Writes `.setup-complete` marker on success. `backoffLimit: 3`, cleaned up after 300s. |
| `thunder-config` | `v1/ConfigMap` | Thunder `deployment.yaml` with SQLite paths. Rendered when `dbType: sqlite`. |
| `thunder-config-pg` | `v1/ConfigMap` | Thunder `deployment.yaml` with PostgreSQL connection details. Rendered when `dbType: postgres`. |
| `gate-config` | `v1/ConfigMap` | Gate frontend `config.js` with server public URL. |
| `console-config` | `v1/ConfigMap` | Console frontend `config.js` with client ID, scopes, and server URL. |
| `file-config` | `v1/ConfigMap` (×N) | One ConfigMap per file attached via `configurations`. Dynamically rendered. |
| `deployment` | `apps/v1/Deployment` | Thunder pod. Includes `wait-for-setup` init container that blocks until `.setup-complete` exists. |
| `service` | `v1/Service` | ClusterIP service on the Thunder server port. |
| `httproute-external` | `gateway.networking.k8s.io/v1/HTTPRoute` | External ingress route. Created when `endpointVisibility: external`. |

### Parameters vs Environment Configurations

- **`parameters`** — frozen at `ComponentRelease` creation time, identical across all environments. Cannot be changed without cutting a new release.
- **`environmentConfigs`** — per-environment values injected via `ReleaseBinding`. Configurable per environment at promotion time.

#### Parameters Schema

| Field | Description | Default |
|-------|-------------|---------|
| `image` | Thunder container image (`repository:tag`) | `ghcr.io/asgardeo/thunder:latest` |
| `runtime.imagePullPolicy` | Container image pull policy | `Always` |
| `runtime.dbType` | Database engine: `sqlite` or `postgres` | `sqlite` |
| `runtime.dbStorageSize` | PVC size for SQLite data files | `1Gi` |
| `runtime.gate.clientBase` | Gate frontend base path | `/gate` |
| `runtime.console.clientBase` | Console frontend base path | `/console` |
| `runtime.console.clientId` | Console OAuth client ID | `CONSOLE` |
| `runtime.console.scopes` | Console OAuth scopes (JSON array string) | `["openid", "profile", "email"]` |

#### Environment Configurations Schema

| Field | Description | Default |
|-------|-------------|---------|
| `replicas` | Number of pod replicas | `1` |
| `endpointVisibility` | `external` (`HTTPRoute`) or `internal` (ClusterIP only) | `external` |
| `serverPublicUrl` | Thunder public-facing URL | `""` |
| `gateClientHostname` | Gate service hostname | `""` |
| `resourceRequestsCpu` | CPU request | `100m` |
| `resourceRequestsMemory` | Memory request | `128Mi` |
| `resourceLimitsCpu` | CPU limit | `500m` |
| `resourceLimitsMemory` | Memory limit | `512Mi` |

## Prerequisites

- Kubernetes cluster with OpenChoreo installed
- Helm 3.x
- A `ClusterDataPlane` resource provisioned (run `kubectl get clusterdataplane` to verify)
- **SQLite (default):** no external database required — data is stored on a PVC
- **PostgreSQL (optional):** an accessible PostgreSQL instance when `database.type: postgres`


## Quick Start

### SQLite (Default — No External DB Needed)

```bash
helm upgrade --install thunder install/openchoreo/helm/ \
  --namespace identity-platform \
  --create-namespace \
  --set thunder-component.serverPublicUrl="http://development-thunder-identity-platform.openchoreoapis.localhost:19080" \
  --set thunder-component.gate.hostname="development-thunder-identity-platform.openchoreoapis.localhost"
```

### PostgreSQL

1. **Export required values**:

   ```bash
   export DB_HOST="postgres.example.com"
   export DB_NAME="thunderdb"
   export DB_USER="thunder"
   export DB_PASS="<your-password>"
   export SERVER_PUBLIC_URL="http://development-thunder-identity-platform.openchoreoapis.localhost:19080"
   export GATE_HOSTNAME="development-thunder-identity-platform.openchoreoapis.localhost"
   ```

2. **Install the chart**:

   ```bash
   helm upgrade --install thunder install/openchoreo/helm/ \
     --namespace identity-platform \
     --create-namespace \
     --set thunder-component.database.type=postgres \
     --set thunder-component.database.host="$DB_HOST" \
     --set thunder-component.database.config.database="$DB_NAME" \
     --set thunder-component.database.config.username="$DB_USER" \
     --set thunder-component.database.config.password="$DB_PASS" \
     --set thunder-component.database.runtime.database="$DB_NAME" \
     --set thunder-component.database.runtime.username="$DB_USER" \
     --set thunder-component.database.runtime.password="$DB_PASS" \
     --set thunder-component.database.user.database="$DB_NAME" \
     --set thunder-component.database.user.username="$DB_USER" \
     --set thunder-component.database.user.password="$DB_PASS" \
     --set thunder-component.serverPublicUrl="$SERVER_PUBLIC_URL" \
     --set thunder-component.gate.hostname="$GATE_HOSTNAME"
   ```

3. **Verify deployment**:

   ```bash
   # Check OpenChoreo resource status
   kubectl get componentrelease,releasebinding -n identity-platform

   # Find the Thunder pod in the data plane namespace
   kubectl get pod -A | grep thunder
   ```

4. **Access Thunder**:

   Once the `ReleaseBinding` is active and the pod is running, Thunder is accessible via the `HTTPRoute` hostname:

   ```
   http://<environmentName>-<componentName>-<componentNamespace>.<gateway-domain>:<port>
   # e.g. http://development-thunder-identity-platform.openchoreoapis.localhost:19080
   ```

## Promotion

To promote Thunder to `staging` or `production`:

1. Open the Backstage portal and navigate to the Thunder component.
2. Click **Promote** on the `development` `ReleaseBinding`.
3. Fill in the environment-specific `environmentConfigs` for the target environment:
   - `serverPublicUrl` — public URL for the target environment
   - `gateClientHostname` — gate service hostname
   - `replicas` — desired replica count
   - `endpointVisibility` — `external` or `internal`
   - resource requests/limits
4. Confirm the promotion.

## Configuration Reference

### Core

| Parameter | Description | Default |
|-----------|-------------|---------|
| `thunder-component.componentName` | Base name for all OpenChoreo resources | `thunder` |
| `thunder-component.image.repository` | Thunder container image repository | `ghcr.io/asgardeo/thunder` |
| `thunder-component.image.tag` | Container image tag | `latest` |
| `thunder-component.thunder.server.port` | Port on which Thunder server listens | `8090` |
| `thunder-component.serverPublicUrl` | Thunder public-facing URL | `<SERVER_PUBLIC_URL>` |
| `thunder-component.project.name` | OpenChoreo project and Kubernetes `namespace` name | `identity-platform` |
| `thunder-component.dataPlane.name` | `ClusterDataPlane` resource to bind environments to | `default` |
| `thunder-component.replicas` | Pod replicas in the development environment | `1` |

### Database

| Parameter | Description | Default |
|-----------|-------------|---------|
| `thunder-component.database.type` | Database engine: `sqlite` or `postgres` | `sqlite` |
| `thunder-component.database.storageSize` | PVC size for SQLite files | `1Gi` |
| `thunder-component.database.config.path` | SQLite config DB path (relative to Thunder working directory) | `repository/database/configdb.db` |
| `thunder-component.database.runtime.path` | SQLite runtime DB path | `repository/database/runtimedb.db` |
| `thunder-component.database.user.path` | SQLite user DB path | `repository/database/userdb.db` |
| `thunder-component.database.host` | PostgreSQL hostname (`postgres` only) | — |
| `thunder-component.database.port` | PostgreSQL port — rendered as an integer in the ConfigMap (`postgres` only) | `5432` |
| `thunder-component.database.config.database` | Config DB name (`postgres` only) | `thunderdb` |
| `thunder-component.database.config.username` | Config DB username (`postgres` only) | — |
| `thunder-component.database.config.password` | Config DB password (`postgres` only) | — |
| `thunder-component.database.config.sslmode` | Config DB SSL mode (`postgres` only) | `disable` |
| `thunder-component.database.runtime.*` | Runtime DB settings — same fields as `config` | — |
| `thunder-component.database.user.*` | User DB settings — same fields as `config` | — |

### Gate and Console

| Parameter | Description | Default |
|-----------|-------------|---------|
| `thunder-component.gate.hostname` | Gate hostname (used in development `ReleaseBinding`) | `<GATE_HOSTNAME>` |
| `thunder-component.gate.port` | Gate port | `19080` |
| `thunder-component.gate.scheme` | Gate scheme (`http` or `https`) | `http` |
| `thunder-component.gate.clientBase` | Gate frontend base path | `/gate` |
| `thunder-component.console.clientBase` | Console frontend base path | `/console` |
| `thunder-component.console.clientId` | Console OAuth client ID | `CONSOLE` |
| `thunder-component.console.scopes` | Console OAuth scopes (JSON array string) | `'["openid", "profile", "email", "system"]'` |

### Security

> **Warning**: Replace `crypto.encryption.key` with a 32-byte (64 hex character) key in production.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `thunder-component.jwt.validity` | JWT token validity in seconds | `3600` |
| `thunder-component.oauth.refresh_token_validity` | Refresh token validity in seconds | `86400` |
| `thunder-component.crypto.encryption.key` | Crypto encryption key (path or raw value) | `file://repository/resources/security/crypto.key` |

### Cache and Consent

| Parameter | Description | Default |
|-----------|-------------|---------|
| `thunder-component.cache.size` | Maximum number of cache entries | `10000` |
| `thunder-component.cache.ttl` | Cache entry TTL in seconds | `3600` |
| `thunder-component.consent.enabled` | Enable consent server integration | `false` |
| `thunder-component.consent.baseUrl` | Consent server base URL | `http://localhost:9090/api/v1` |

### CORS

| Parameter | Description | Default |
|-----------|-------------|---------|
| `thunder-component.cors.allowed_origins` | Allowed CORS origins for the development environment | `["https://thunder-gate.your-domain.com", "https://localhost:3000"]` |

### Resource Controls

| Parameter | Description | Default |
|-----------|-------------|---------|
| `thunder-component.namespace.create` | Create the Kubernetes `namespace` | `true` |
| `thunder-component.project.create` | Create the OpenChoreo Project | `true` |
| `thunder-component.pipeline.create` | Create the DeploymentPipeline | `true` |
| `thunder-component.environments.development.create` | Create the development Environment | `true` |
| `thunder-component.environments.staging.create` | Create the staging Environment | `true` |
| `thunder-component.environments.production.create` | Create the production Environment | `true` |
| `thunder-component.releaseBinding.create` | Create the development ReleaseBinding (auto-deploy) | `true` |

### ComponentType Scope

| Parameter | Description | Default |
|-----------|-------------|---------|
| `thunder-oc-componenttype.componentType.cluster` | `true` → `ClusterComponentType` (cluster-scoped); `false` → namespace-scoped `ComponentType` | `true` |
| `thunder-component.componentType.cluster` | Must match the value above | `true` |

## OpenChoreo UI — Environment Variables

When deploying or promoting Thunder through the OpenChoreo UI, set the following environment variables on the Workload.

### Endpoint

Create an endpoint with the following settings:

| Field | Value |
|-------|-------|
| Name | `api` |
| Port | `SERVER_PORT` (default `8090`) |
| Type | `HTTP` |

The endpoint name `api` is required — the ComponentType references it as `workload.endpoints["api"]` to wire the Service and `HTTPRoute`.

The `HTTPRoute` hostname is constructed as:

```
<environmentName>-<componentName>-<componentNamespace>.<gateway-domain>
```

For example, with `componentName: thunder`, `namespace` `identity-platform`, environment `development`, and gateway domain `openchoreoapis.localhost`:

```
development-thunder-identity-platform.openchoreoapis.localhost
```

Use this to derive the following values when configuring the Component via the UI:

| Field | Derived value |
|-------|---------------|
| `Server Public URL` | `http://<hostname>:<gateway-port>` e.g. `http://development-thunder-identity-platform.openchoreoapis.localhost:19080` |
| `Gate Client Hostname` | `<hostname>` e.g. `development-thunder-identity-platform.openchoreoapis.localhost` |

### General

| Environment Variable | How to get |
|---------|------------|
| `SERVER_PORT` | Fixed — port Thunder listens on inside the container (`8090`) |
| `CRYPTO_ENCRYPTION_KEY` | Generate a 32-byte hex key: `openssl rand -hex 32` |
| `JWT_VALIDITY` | Token lifetime in seconds; adjust per security policy |
| `OAUTH_REFRESH_TOKEN_VALIDITY` | Refresh token lifetime in seconds |
| `CACHE_SIZE` | Maximum number of in-memory cache entries |
| `CACHE_TTL` | Cache entry TTL in seconds |
| `CORS_ALLOWED_ORIGINS` | JSON array of allowed origins e.g. `["https://app.example.com"]` |

### Gate Client

| Environment Variable | How to get |
|---------|------------|
| `GATE_CLIENT_BASE` | Base path for the Gate frontend |
| `GATE_CLIENT_PORT` | Port of the gateway or Gate service |
| `GATE_CLIENT_SCHEME` | `http` or `https` depending on your gateway TLS setup |

### Console

| Environment Variable | How to get |
|---------|------------|
| `CONSOLE_CLIENT_BASE` | Base path for the Console frontend |
| `CONSOLE_CLIENT_ID` | OAuth client ID registered for the Console |
| `CONSOLE_SCOPES` | JSON array of OAuth scopes the Console requests |

### Consent

| Environment Variable | How to get |
|---------|------------|
| `CONSENT_ENABLED` | Set `true` to enable consent server integration |
| `CONSENT_BASE_URL` | Base URL of the consent server |

### Database — SQLite (`dbType: sqlite`)

| Environment Variable | How to get |
|---------|------------|
| `DB_CONFIG_PATH` | Path relative to Thunder working directory inside the container |
| `DB_RUNTIME_PATH` | Path relative to Thunder working directory inside the container |
| `DB_USER_PATH` | Path relative to Thunder working directory inside the container |

### Database — PostgreSQL (`dbType: postgres`)

Create the database and grant privileges before setting these values:

```sql
CREATE DATABASE thunderdb OWNER asgthunder;
GRANT ALL PRIVILEGES ON DATABASE thunderdb TO asgthunder;
```

| Environment Variable | How to get |
|---------|------------|
| `DB_PORT` | PostgreSQL port — must be an integer (not quoted) |
| `DB_CONFIG_HOSTNAME` | PostgreSQL hostname e.g. `postgres-postgresql.identity-platform.svc.cluster.local` |
| `DB_CONFIG_NAME` | Database name created in the SQL above |
| `DB_CONFIG_USERNAME` | PostgreSQL application user e.g. `asgthunder` |
| `DB_CONFIG_PASSWORD` | Password for the application user |
| `DB_CONFIG_SSLMODE` | `disable`, `require`, or `verify-full` (recommended for production) |
| `DB_RUNTIME_HOSTNAME` | Same as `DB_CONFIG_HOSTNAME` unless using separate DB hosts |
| `DB_RUNTIME_NAME` | Same as `DB_CONFIG_NAME` unless using separate databases |
| `DB_RUNTIME_USERNAME` | Same as `DB_CONFIG_USERNAME` unless using separate users |
| `DB_RUNTIME_PASSWORD` | Same as `DB_CONFIG_PASSWORD` unless using separate users |
| `DB_RUNTIME_SSLMODE` | Same as `DB_CONFIG_SSLMODE` |
| `DB_USER_HOSTNAME` | Same as `DB_CONFIG_HOSTNAME` unless using separate DB hosts |
| `DB_USER_NAME` | Same as `DB_CONFIG_NAME` unless using separate databases |
| `DB_USER_USERNAME` | Same as `DB_CONFIG_USERNAME` unless using separate users |
| `DB_USER_PASSWORD` | Same as `DB_CONFIG_PASSWORD` unless using separate users |
| `DB_USER_SSLMODE` | Same as `DB_CONFIG_SSLMODE` |

## Chart Structure

```
install/openchoreo/helm/
├── Chart.yaml                          # Umbrella chart definition
├── values.yaml                         # Top-level defaults (sub-chart overrides)
├── README.md
└── charts/
    ├── thunder-oc-componenttype/
    │   ├── Chart.yaml
    │   ├── values.yaml                 # ComponentType scope toggle
    │   └── templates/
    │       └── thunder-componenttype.yaml  # ClusterComponentType / ComponentType
    └── thunder-component/
        ├── Chart.yaml
        ├── values.yaml                 # All Thunder instance defaults
        └── templates/
            ├── namespace.yaml          # Kubernetes Namespace
            ├── thunder-platform.yaml   # Project, DeploymentPipeline, Environments
            ├── thunder-component.yaml  # Component and Workload
            └── thunder-release.yaml    # ComponentRelease and ReleaseBinding (post-install hooks)
```

## Debugging

```bash
# Check all OpenChoreo resources
kubectl get clustercomponenttype,component,workload,componentrelease,releasebinding -n identity-platform

# Find all Thunder pods across data plane namespaces
kubectl get pod -A | grep thunder

# Check Thunder logs
kubectl logs <pod-name> -n <dp-namespace>

# Inspect the rendered Thunder configuration
kubectl get configmap <componentName>-config -n <dp-namespace> -o jsonpath='{.data.deployment\.yaml}'

# Check setup job logs
kubectl logs job/<componentName>-setup -n <dp-namespace>

# Check wait-for-setup init container logs
kubectl logs <pod-name> -n <dp-namespace> -c wait-for-setup

# Render templates locally without installing
helm template thunder install/openchoreo/helm/ \
  --namespace identity-platform \
  --set thunder-component.serverPublicUrl="http://thunder-dev.example.com" \
  --set thunder-component.gate.hostname="thunder-dev.example.com"
```

## Security Considerations

- Never use default passwords in production
- Replace `crypto.encryption.key` with a strong 32-byte hex key in production
- Configure `cors.allowed_origins` restrictively — avoid wildcards
- Enable SSL/TLS for PostgreSQL connections in production (`sslmode: verify-full`)
- Use specific image tags instead of `latest` in production
- Set `thunder-oc-componenttype.componentType.cluster: false` if you need namespace-scoped isolation

## Contributing

- Open an issue in the [Thunder GitHub repository](https://github.com/asgardeo/thunder)
- Refer to the project's [CONTRIBUTING guidelines](../../../CONTRIBUTING.md)
