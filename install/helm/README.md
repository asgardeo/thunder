# Thunder Helm Chart

This repository contains the Helm chart for WSO2 Thunder, a lightweight user and identity management system designed for modern application development.

## Configuration Value Types

Thunder's configuration system supports multiple value formats for **any parameter** in the configuration:

1. **Direct Values** - Static values specified directly in YAML:
   ```yaml
   server:
     hostname: "localhost"
     port: 8090
   ```

2. **Environment Variables** - Use Go template syntax `{{.VARIABLE_NAME}}` to reference environment variables:
   ```yaml
   database:
     identity:
       password: "{{.DB_PASSWORD}}"
   server:
     publicUrl: "{{.PUBLIC_URL}}"
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
   - `file://relative/path` - Relative paths (resolved from the Thunder installation directory)

## Prerequisites

### Infrastructure
- Running Kubernetes cluster ([minikube](https://kubernetes.io/docs/tasks/tools/#minikube) or an alternative cluster)
- Kubernetes ingress controller ([NGINX Ingress](https://github.com/kubernetes/ingress-nginx) recommended)

### Tools
| Tool          | Installation Guide | Version Check Command |
|---------------|--------------------|-----------------------|
| Git           | [Install Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) | `git --version` |
| Helm          | [Install Helm](https://helm.sh/docs/intro/install/) | `helm version` |
| Docker        | [Install Docker](https://docs.docker.com/engine/install/) | `docker --version` |
| kubectl       | [Install kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) | `kubectl version` |

## Quick Start Guide

Follow these steps to deploy Thunder in your Kubernetes cluster:

### 1. Install the Thunder Helm chart

```bash
# Pull and install from GitHub Container Registry
helm install my-thunder oci://ghcr.io/asgardeo/helm-charts/thunder
```

If you wish to install another version, use the command below to specify the desired version.

```bash
helm install my-thunder oci://ghcr.io/asgardeo/helm-charts/thunder --version <VERSION>
```

> To see which chart versions are available, you can:
> - Visit the [Thunder Helm Chart Registry](https://github.com/asgardeo/thunder/pkgs/container/helm-charts%2Fthunder) on GitHub Container Registry.

If you want to customize the installation, create a `custom-values.yaml` file with your configurations and use:

```bash
helm install my-thunder oci://ghcr.io/asgardeo/helm-charts/thunder -f custom-values.yaml
```

The command deploys Thunder on the Kubernetes cluster with the default configuration. The [Parameters](#parameters) section lists the available parameters that can be configured during installation.

If you want to install Thunder with SQLite databases, use the following command:

```bash
helm install my-thunder oci://ghcr.io/asgardeo/helm-charts/thunder \
  --set configuration.database.identity.type=sqlite \
  --set configuration.database.runtime.type=sqlite \
  --set configuration.database.user.type=sqlite
```

**Note:** When using SQLite:
- **Persistence is automatically enabled** when any database is configured to use SQLite
- The setup job's init container will automatically copy SQLite databases from the image to a PVC
- Database files will persist across pod restarts

### 2. Obtain the External IP

After deploying Thunder, you need to find its external IP address to access it outside the cluster. Run the following command to list the Ingress resources:

```bash
kubectl get ingress
```
**Output Fields:**

- **HOSTS** – Hostname (e.g., `thunder.local`)
- **ADDRESS** – External IP
- **PORTS** – Exposed ports (usually 80, 443)

After the installation is complete, you can access Thunder via the Ingress hostname.

By default, Thunder will be available at `http://thunder.local`. You may need to add this hostname to your local hosts file or configure your DNS accordingly.

### Uninstalling the Chart

To uninstall/delete the `my-thunder` deployment:

```bash
helm uninstall my-thunder
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

The following table lists the configurable parameters of the Thunder chart and their default values.

### Global Parameters

| Name                      | Description                                     | Default                                                 |
| ------------------------- | ----------------------------------------------- | ------------------------------------------------------- |
| `nameOverride`            | String to partially override common.names.fullname | `""`                                                  |
| `fullnameOverride`        | String to fully override common.names.fullname  | `""`                                                    |

### Deployment Parameters

| Name                                    | Description                                                                             | Default                        |
| --------------------------------------- | --------------------------------------------------------------------------------------- | ------------------------------ |
| `deployment.replicaCount`               | Number of Thunder replicas                                                              | `2`                            |
| `deployment.strategy.rollingUpdate.maxSurge` | Maximum number of pods that can be created over the desired number during an update | `1`                           |
| `deployment.strategy.rollingUpdate.maxUnavailable` | Maximum number of pods that can be unavailable during an update              | `0`                           |
| `deployment.image.registry`             | Thunder image registry                                                                  | `ghcr.io/asgardeo`             |
| `deployment.image.repository`           | Thunder image repository                                                                | `thunder`                      |
| `deployment.image.tag`                  | Thunder image tag                                                                       | `0.7.0`                        |
| `deployment.image.digest`               | Thunder image digest (use either tag or digest)                                         | `""`                           |
| `deployment.image.pullPolicy`           | Thunder image pull policy                                                               | `Always`                       |
| `deployment.terminationGracePeriodSeconds` | Pod termination grace period in seconds                                              | `10`                           |
| `deployment.container.port`             | Thunder container port                                                                  | `8090`                         |
| `deployment.startupProbe.initialDelaySeconds` | Startup probe initial delay seconds                                               | `1`                            |
| `deployment.startupProbe.periodSeconds` | Startup probe period seconds                                                            | `2`                            |
| `deployment.startupProbe.failureThreshold` | Startup probe failure threshold                                                      | `30`                           |
| `deployment.livenessProbe.periodSeconds` | Liveness probe period seconds                                                          | `10`                           |
| `deployment.readinessProbe.initialDelaySeconds` | Readiness probe initial delay seconds                                           | `1`                            |
| `deployment.readinessProbe.periodSeconds` | Readiness probe period seconds                                                        | `10`                           |
| `deployment.resources.limits.cpu`       | CPU resource limits                                                                     | `1.5`                          |
| `deployment.resources.limits.memory`    | Memory resource limits                                                                  | `512Mi`                        |
| `deployment.resources.requests.cpu`     | CPU resource requests                                                                   | `1`                            |
| `deployment.resources.requests.memory`  | Memory resource requests                                                                | `256Mi`                        |
| `deployment.securityContext.readOnlyRootFilesystem` | Enable read-only root filesystem (must be false for SQLite)                     | `true`                         |
| `deployment.securityContext.enableRunAsUser` | Enforce user ID via pod security context                                               | `true`                         |
| `deployment.securityContext.runAsUser`  | User ID to run the container                                                            | `10001`                        |
| `deployment.securityContext.enableRunAsGroup` | Enable setting group ID for the container process                                 | `true`                         |
| `deployment.securityContext.runAsGroup` | Group ID to run the container                                                           | `10001`                        |
| `deployment.securityContext.enableFsGroup` | Enable setting fsGroup for volume ownership                                          | `true`                         |
| `deployment.securityContext.fsGroup`    | Group ID for mounted volumes (fixes SQLite permission issues on cloud platforms)        | `10001`                        |
| `deployment.securityContext.seccompProfile.enabled` | Enable seccomp profile                                                      | `false`                        |
| `deployment.securityContext.seccompProfile.type` | Seccomp profile type                                                           | `RuntimeDefault`               |

### HPA Parameters

| Name                              | Description                                                      | Default                       |
| --------------------------------- | ---------------------------------------------------------------- | ----------------------------- |
| `hpa.enabled`                     | Enable Horizontal Pod Autoscaler                                 | `true`                        |
| `hpa.maxReplicas`                 | Maximum number of replicas                                       | `10`                          |
| `hpa.averageUtilizationCPU`       | Target CPU utilization percentage                                | `65`                          |
| `hpa.averageUtilizationMemory`    | Target Memory utilization percentage                             | `75`                          |

### Service Parameters

| Name                             | Description                                                       | Default                      |
| -------------------------------- | ----------------------------------------------------------------- | ---------------------------- |
| `service.port`                   | Thunder service port                                              | `8090`                       |

### Service Account Parameters

| Name                         | Description                                                | Default                       |
| ---------------------------- | ---------------------------------------------------------- | ----------------------------- |
| `serviceAccount.create`      | Enable creation of ServiceAccount                          | `true`                        |
| `serviceAccount.name`        | Name of the service account to use                         | `thunder-service-account`     |

### PDB Parameters

| Name                        | Description                                                 | Default                       |
| --------------------------- | ----------------------------------------------------------- | ----------------------------- |
| `pdb.minAvailable`          | Minimum number of pods that must be available               | `50%`                         |

### Ingress Parameters

| Name                                  | Description                                                     | Default                      |
| ------------------------------------- | --------------------------------------------------------------- | ---------------------------- |
| `ingress.className`                   | Ingress controller class                                        | `nginx`                      |
| `ingress.hostname`                    | Default host for the ingress resource                           | `thunder.local`              |
| `ingress.paths[0].path`               | Path for the ingress resource                                   | `/`                          |
| `ingress.paths[0].pathType`           | Path type for the ingress resource                              | `Prefix`                     |
| `ingress.tlsSecretsName`              | TLS secret name for HTTPS                                       | `thunder-tls`                |
| `ingress.commonAnnotations`           | Common annotations for ingress                                  | See values.yaml              |
| `ingress.customAnnotations`           | Custom annotations for ingress                                  | `{}`                         |

### Database Password Management

Thunder provides flexible password management for database connections with automatic Kubernetes Secret integration.

#### Security Warning

⚠️ **Storing passwords as plaintext in values.yaml is NOT recommended for production.** Use Kubernetes Secrets or `--set` flags to store sensitive credentials securely.

#### How Password Management Works

Thunder uses intelligent password detection based on the `password` and `passwordRef` fields:

1. **If `passwordRef.key` is set** → Uses external Secret (production pattern)
2. **If `password` has a value but `passwordRef.key` is empty** → Auto-converts to Helm-managed Secret (dev/test pattern)
3. **If both are empty** → No password (SQLite-only deployments)

The auto-created Secret is created as a Helm pre-install/pre-upgrade hook to ensure it exists before the main deployment and setup job run.

#### Pattern 1: Auto-Convert to Helm-Managed Secret (For Development/Testing)

Provide passwords directly in the `password` field. Helm automatically creates a Secret named `<release-name>-db-credentials`:

```yaml
configuration:
  database:
    identity:
      password: "my-secret-password-1"  # Auto-converted to Secret!
    runtime:
      password: "my-secret-password-2"
    user:
      password: "my-secret-password-3"
```

**Best Practice:** Use `--set` flags to avoid committing passwords:
```bash
helm install my-thunder oci://ghcr.io/asgardeo/helm-charts/thunder \
  --set configuration.database.identity.password=mypass1 \
  --set configuration.database.runtime.password=mypass2 \
  --set configuration.database.user.password=mypass3
```

Helm automatically:
- Creates `<release-name>-db-credentials` Secret as a pre-install/pre-upgrade hook
- Injects environment variables (`DB_IDENTITY_PASSWORD`, `DB_RUNTIME_PASSWORD`, `DB_USER_PASSWORD`) into pods
- Updates pods when passwords change (via checksum annotations)

#### Pattern 2: External Secret (For Production - Recommended)

Reference a pre-existing Kubernetes Secret (created manually or by external-secrets-operator):

**Step 1:** Create your Secret:
```bash
kubectl create secret generic my-db-secrets \
  --from-literal=identity-password=secret1 \
  --from-literal=runtime-password=secret2 \
  --from-literal=user-password=secret3
```

**Step 2:** Configure Helm to reference the external Secret:
```yaml
configuration:
  database:
    identity:
      passwordRef:
        name: "my-db-secrets"      # Your Secret name
        key: "identity-password"    # Key within Secret
    runtime:
      passwordRef:
        name: "my-db-secrets"
        key: "runtime-password"
    user:
      passwordRef:
        name: "my-db-secrets"
        key: "user-password"
```

When `passwordRef.key` is set, the `password` field is ignored and Helm uses your external Secret.

**Important:** The checksum annotation used to trigger pod rollouts is only computed for auto-generated Secrets. When you use an external Secret via `passwordRef` (Pattern 2), changes to that Secret will **not** automatically restart pods. You must either manually restart the pods or use a tool to watch for Secret changes and trigger rollouts.

**Important:** When you *do not* use `passwordRef.key` (i.e., you rely on the auto-generated Secret), the Helm chart will
base64-encode the `password` value directly into a Kubernetes Secret. In this mode, values like `"{{.DB_PASSWORD}}"` or
`"file:///secrets/pass"` are stored as literal strings in the Secret and **are not** resolved as environment variables or
file references by Helm. Environment variable placeholders (`{{.VAR}}`) and `file://` references are only resolved when
Thunder reads configuration directly via its application config loader (e.g., from a ConfigMap or file), not when the
value is first converted into a Kubernetes Secret by this chart.

#### Password Field Options
Each database section (`identity`, `runtime`, `user`) supports these fields:
| Field                  | Description                                                                                                                                    | Example                      |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| `password`             | Direct password value. When Thunder reads config directly, this may also be an env var placeholder (`{{.VAR}}`) or file reference (`file://path`). When using the auto-generated Secret, the value is stored **as-is** in the Secret and such placeholders are **not** resolved. | `"mypassword"` or `"{{.DB_PASSWORD}}"` or `"file:///secrets/pass"` |
| `passwordRef.name`     | Kubernetes Secret name (optional, defaults to `<release-name>-db-credentials` for auto-convert)                                               | `"my-db-secrets"`            |
| `passwordRef.key`      | Secret key name. When set, `password` field is ignored and external Secret is used                                                            | `"identity-password"`        |

### Thunder Configuration Parameters

| Name                                              | Description                                                                                           | Default                      |
|---------------------------------------------------|-------------------------------------------------------------------------------------------------------| ---------------------------- |
| `configuration.server.port`                       | Thunder server port                                                                                   | `8090`                       |
| `configuration.server.httpOnly`                   | Whether the server should run in HTTP-only mode                                                       | `false`                      |
| `configuration.server.publicURL`                  | Public URL of the Thunder server                                                                      | `https://thunder.local`      |
| `configuration.gateClient.hostname`               | Gate client hostname                                                                                  | `thunder.local`              |
| `configuration.gateClient.port`                   | Gate client port                                                                                      | `443`                       |
| `configuration.gateClient.scheme`                 | Gate client scheme                                                                                    | `https`                      |
| `configuration.gateClient.path`                   | Gate client base path                                                                                 | `/gate`                      |
| `configuration.developerClient.path`              | Developer client base path                                                                            | `/develop`                 |
| `configuration.developerClient.clientId`          | Developer client ID                                                                                   | `DEVELOP`   |
| `configuration.developerClient.scopes`            | Developer client scopes                                                                               | `['openid', 'profile', 'email', 'system']` |
| `configuration.security.certFile`                 | Server certificate file path                                                                          | `repository/resources/security/server.cert` |
| `configuration.security.keyFile`                  | Server key file path                                                                                  | `repository/resources/security/server.key`  |
| `configuration.crypto.encryption.key`             | Crypto encryption key (change the default key with a 32-byte (64 character) hex string in production) | `file://repository/resources/security/crypto.key` |
| `configuration.database.identity.type`            | Identity database type (postgres or sqlite)                                                           | `postgres`                   |
| `configuration.database.identity.sqlitePath`      | SQLite database path (for sqlite only)                                                                | `repository/database/thunderdb.db` |
| `configuration.database.identity.sqliteOptions`   | SQLite options (for sqlite only)                                                                      | `_journal_mode=WAL&_busy_timeout=5000` |
| `configuration.database.identity.name`            | Postgres database name (for postgres only)                                                            | `thunderdb`                  |
| `configuration.database.identity.host`            | Postgres host (for postgres only)                                                                     | `localhost` |
| `configuration.database.identity.port`            | Postgres port (for postgres only)                                                                     | `5432`                       |
| `configuration.database.identity.username`        | Postgres username (for postgres only)                                                                 | `asgthunder`                   |
| `configuration.database.identity.password`        | Database password - supports plaintext. When `passwordRef.key` is set, this field is ignored and the external Secret is used instead. | `asgthunder`    |
| `configuration.database.identity.passwordRef.name` | Kubernetes Secret name for identity database password. Leave empty to use auto-created `<release-name>-db-credentials` Secret when password field is set | `""`    |
| `configuration.database.identity.passwordRef.key`  | Kubernetes Secret key for identity database password. When set, overrides `password` field and uses external Secret | `""`    |
| `configuration.database.identity.sslmode`         | Postgres SSL mode (for postgres only)                                                                 | `require`                    |
| `configuration.database.identity.max_open_conns`  | Maximum number of open connections to the database                                                    | `500`                        |
| `configuration.database.identity.max_idle_conns`  | Maximum number of idle connections in the pool                                                        | `100`                        |
| `configuration.database.identity.conn_max_lifetime` | Maximum lifetime of a connection in seconds                                                         | `3600`                       |
| `configuration.database.runtime.type`             | Runtime database type (postgres or sqlite)                                                            | `postgres`                   |
| `configuration.database.runtime.sqlitePath`       | SQLite database path (for sqlite only)                                                                | `repository/database/runtimedb.db` |
| `configuration.database.runtime.sqliteOptions`    | SQLite options (for sqlite only)                                                                      | `_journal_mode=WAL&_busy_timeout=5000` |
| `configuration.database.runtime.name`             | Postgres database name (for postgres only)                                                            | `runtimedb`                  |
| `configuration.database.runtime.host`             | Postgres host (for postgres only)                                                                     | `localhost` |
| `configuration.database.runtime.port`             | Postgres port (for postgres only)                                                                     | `5432`                      |
| `configuration.database.runtime.username`         | Postgres username (for postgres only)                                                                 | `asgthunder`                   |
| `configuration.database.runtime.password`         | Database password - supports plaintext. When `passwordRef.key` is set, this field is ignored and the external Secret is used instead. | `asgthunder`     |
| `configuration.database.runtime.passwordRef.name`  | Kubernetes Secret name for runtime database password. Leave empty to use auto-created `<release-name>-db-credentials` Secret when password field is set | `""`    |
| `configuration.database.runtime.passwordRef.key`   | Kubernetes Secret key for runtime database password. When set, overrides `password` field and uses external Secret | `""`    |
| `configuration.database.runtime.sslmode`          | Postgres SSL mode (for postgres only)                                                                 | `require`                    |
| `configuration.database.runtime.max_open_conns`   | Maximum number of open connections to the database                                                    | `500`                        |
| `configuration.database.runtime.max_idle_conns`   | Maximum number of idle connections in the pool                                                        | `100`                        |
| `configuration.database.runtime.conn_max_lifetime` | Maximum lifetime of a connection in seconds                                                          | `3600`                       |
| `configuration.database.user.type`                | User database type (postgres or sqlite)                                                               | `postgres`                   |
| `configuration.database.user.sqlitePath`          | SQLite database path (for sqlite only)                                                                | `repository/database/userdb.db` |
| `configuration.database.user.sqliteOptions`       | SQLite options (for sqlite only)                                                                      | `_journal_mode=WAL&_busy_timeout=5000` |
| `configuration.database.user.name`                | Postgres database name (for postgres only)                                                            | `userdb`                     |
| `configuration.database.user.host`                | Postgres host (for postgres only)                                                                     | `localhost` |
| `configuration.database.user.port`                | Postgres port (for postgres only)                                                                     | `5432`                       |
| `configuration.database.user.username`            | Postgres username (for postgres only)                                                                 | `asgthunder`                   |
| `configuration.database.user.password`            | Database password - supports plaintext. When `passwordRef.key` is set, this field is ignored and the external Secret is used instead. | `asgthunder`        |
| `configuration.database.user.passwordRef.name`     | Kubernetes Secret name for user database password. Leave empty to use auto-created `<release-name>-db-credentials` Secret when password field is set | `""`    |
| `configuration.database.user.passwordRef.key`      | Kubernetes Secret key for user database password. When set, overrides `password` field and uses external Secret | `""`    |
| `configuration.database.user.sslmode`             | Postgres SSL mode (for postgres only)                                                                 | `require`                    |
| `configuration.database.user.max_open_conns`      | Maximum number of open connections to the database                                                    | `500`                        |
| `configuration.database.user.max_idle_conns`      | Maximum number of idle connections in the pool                                                        | `100`                        |
| `configuration.database.user.conn_max_lifetime`   | Maximum lifetime of a connection in seconds                                                           | `3600`                       |
| `configuration.cache.disabled`                    | Disable cache                                                                                         | `false`                      |
| `configuration.cache.type`                        | Cache type                                                                                            | `inmemory`                   |
| `configuration.cache.size`                        | Cache size                                                                                            | `1000`                       |
| `configuration.cache.ttl`                         | Cache TTL in seconds                                                                                  | `3600`                       |
| `configuration.cache.evictionPolicy`              | Cache eviction policy                                                                                 | `LRU`                        |
| `configuration.cache.cleanupInterval`             | Cache cleanup interval in seconds                                                                     | `300`                        |
| `configuration.jwt.issuer`                        | JWT issuer (derived from server.publicUrl if not set)                                                 | derived                      |
| `configuration.jwt.validityPeriod`                | JWT validity period in seconds                                                                        | `3600`                       |
| `configuration.jwt.audience`                      | Default audience for auth assertions                                                                  | `application`                |
| `configuration.oauth.refreshToken.renewOnGrant`   | Renew refresh token on grant                                                                          | `false`                      |
| `configuration.oauth.refreshToken.validityPeriod` | Refresh token validity period in seconds                                                              | `86400`                      |
| `configuration.flow.defaultAuthFlowHandle`        | Default authentication flow handle                                                                    | `default-basic-flow`         |
| `configuration.flow.maxVersionHistory`            | Maximum flow version history to retain                                                                | `3`                          |
| `configuration.flow.autoInferRegistration`        | Enable auto-infer registration flow                                                                   | `true`                       |
| `configuration.cors.allowedOrigins`               | CORS allowed origins                                                                                  | See values.yaml              |
| `configuration.passkey.allowedOrigins`            | Passkey allowed origins                                                                               | `[]`                         |

### Persistence Parameters

Persistence is **automatically enabled** when using SQLite as the database type for any database (identity, runtime, or user). It creates a PersistentVolumeClaim to store SQLite database files.

| Name                                   | Description                                                     | Default                      |
| -------------------------------------- | --------------------------------------------------------------- | ---------------------------- |
| `persistence.enabled`                  | Enable persistence for SQLite databases (auto-enabled for SQLite) | `false`                    |
| `persistence.storageClass`             | Storage class name (use "-" for no storage class)               | `""`                         |
| `persistence.accessMode`               | PVC access mode                                                 | `ReadWriteOnce`              |
| `persistence.size`                     | PVC storage size                                                | `1Gi`                        |
| `persistence.annotations`              | Additional annotations for PVC                                  | `{}`                         |

**Note:** 
- When any database is configured to use SQLite, a PersistentVolumeClaim (PVC) is **always created** to store the database files, regardless of the `persistence.enabled` or `setup.enabled` settings.
- The PVC is mounted by the setup job's init container (if `setup.enabled` is true) to initialize the database, and by the main Thunder deployment for ongoing operation.
- You can customize the storage size and storage class for the PVC using the `persistence.size` and `persistence.storageClass` values.

### Setup Job Parameters

The setup job runs `setup.sh` as a one-time Helm pre-install hook to initialize Thunder with default resources (admin user, organization, etc.).

| Name                                   | Description                                                     | Default                      |
| -------------------------------------- | --------------------------------------------------------------- | ---------------------------- |
| `setup.enabled`                        | Enable setup job (runs on install via Helm hook)                | `true`                       |
| `setup.backoffLimit`                   | Number of retries if setup fails                                | `3`                          |
| `setup.preserveJob`                    | Preserve job after completion (false = delete on success)       | `false`                      |
| `setup.ttlSecondsAfterFinished`        | Time to keep failed jobs (only if preserveJob=false)            | `86400` (24 hours)           |
| `setup.debug`                          | Enable debug mode for setup                                     | `false`                      |
| `setup.args`                           | Additional command-line arguments for setup.sh                  | `[]`                         |
| `setup.env`                            | Additional environment variables for setup job                  | `[]`                         |
| `setup.resources.requests.cpu`         | CPU request for setup job                                       | `250m`                       |
| `setup.resources.requests.memory`      | Memory request for setup job                                    | `128Mi`                      |
| `setup.resources.limits.cpu`           | CPU limit for setup job                                         | `500m`                       |
| `setup.resources.limits.memory`        | Memory limit for setup job                                      | `256Mi`                      |
| `setup.extraVolumeMounts`              | Additional volume mounts for setup job                          | `[]`                         |
| `setup.extraVolumes`                   | Additional volumes for setup job                                | `[]`                         |

**Job Retention Behavior:**
- When `preserveJob=false` (default): Successful jobs are deleted immediately. Failed jobs are kept for `ttlSecondsAfterFinished` (24 hours) to allow debugging.
- When `preserveJob=true`: Job is kept indefinitely regardless of success/failure status. Use this for troubleshooting or audit purposes.

### Bootstrap Script Parameters

Bootstrap scripts extend Thunder's setup process by adding your own initialization logic. These scripts run as part of the setup job.

#### Understanding Default Bootstrap Scripts

Thunder provides these default bootstrap scripts in `/opt/thunder/bootstrap/`:
- **`common.sh`** - Helper functions for logging (`log_info`, `log_success`, `log_warning`, `log_error`) and API calls (`thunder_api_call`)
- **`01-default-resources.sh`** - Creates admin user, default organization, and Person user schema
- **`02-sample-resources.sh`** - Creates sample resources for testing

#### Configuration Parameters

| Name                        | Description                                                                      | Default |
| --------------------------- | -------------------------------------------------------------------------------- | ------- |
| `bootstrap.scripts`         | Inline custom bootstrap scripts (key: filename, value: content)                 | `{}`    |
| `bootstrap.configMap.name`  | Name of external ConfigMap containing bootstrap scripts                          | `""`    |
| `bootstrap.configMap.files` | List of script filenames to mount from ConfigMap (empty = mount entire ConfigMap) | `[]`    |

#### Three Bootstrap Patterns

**Pattern 1: Add Inline Scripts** (Preserves Defaults)

Use `bootstrap.scripts` to define scripts directly in values.yaml. These scripts are added to the default bootstrap scripts.

```yaml
bootstrap:
  scripts:
    30-custom-users.sh: |
      #!/bin/bash
      set -e
      SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
      source "${SCRIPT_DIR}/common.sh"

      log_info "Creating custom user..."
      thunder_api_call POST "/users" '{"type":"person","attributes":{"username":"alice","password":"alice123","sub":"alice","email":"alice@example.com"}}'
      log_success "User created"
```

- ✅ Preserves Thunder's default scripts (`common.sh`, `01-*`, `02-*`)
- ✅ Can use helper functions from `common.sh`
- ✅ No additional configuration needed

---

**Pattern 2: Add External ConfigMap Scripts** (Preserves Defaults)

Use `bootstrap.configMap` with a `files` list to mount specific scripts from an external ConfigMap.

Create your ConfigMap:
```bash
kubectl create configmap my-bootstrap \
  --from-file=30-users.sh=./30-users.sh \
  --from-file=40-apps.sh=./40-apps.sh
```

Configure Helm values:
```yaml
bootstrap:
  configMap:
    name: "my-bootstrap"
    files:
      - 30-users.sh
      - 40-apps.sh
```

- ✅ Preserves Thunder's default scripts
- ✅ Can use helper functions from `common.sh`
- ✅ Scripts managed separately from Helm chart

---

**Pattern 3: Replace All Scripts with ConfigMap** (Complete Replacement)

⚠️ **WARNING**: This completely replaces Thunder's default bootstrap scripts. Use only if you need complete control.

Use `bootstrap.configMap` **without** specifying `files` to mount the entire ConfigMap and replace all defaults.

Create your complete ConfigMap (must include `common.sh`):
```bash
kubectl create configmap complete-bootstrap \
  --from-file=common.sh=./common.sh \
  --from-file=01-my-setup.sh=./01-my-setup.sh
```

Configure Helm values:
```yaml
bootstrap:
  configMap:
    name: "complete-bootstrap"
    # No files list = mounts entire ConfigMap (replaces all defaults)
```

- ⚠️ **Removes ALL default scripts** (`common.sh`, `01-default-resources.sh`, `02-sample-resources.sh`)
- ⚠️ You MUST provide your own `common.sh` with required helper functions
- ⚠️ No default admin user, organization, or schemas will be created
- ✅ Complete control over bootstrap process

**For comprehensive examples, helper function documentation, and best practices, see:** [Custom Bootstrap Guide](../../docs/guides/setup/custom-bootstrap.md)

### Custom Configuration

The Thunder configuration file (deployment.yaml) can be customized by overriding the default values in the values.yaml file.
Alternatively, you can directly update the values in conf/deployment.yaml before deploying the Helm chart.

### Database Configuration

Thunder supports both sqlite and postgres databases. By default, postgres is configured.

Make sure to create the necessary databases and users in your Postgres instance before deploying Thunder. The values.yaml should be overridden with the required database configurations for the DB created.

Note: Use sqlite only if you are running a single pod.
