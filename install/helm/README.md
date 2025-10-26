# Thunder Helm Chart

This repository contains the Helm chart for WSO2 Thunder, a lightweight user and identity management system designed for modern application development.

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

### 1. Clone the Thunder repository

```bash
git clone https://github.com/asgardeo/thunder.git
cd thunder/install/helm
```

### 2. Install the Thunder Helm chart

You can install the Thunder Helm chart with the release name `my-thunder` as follows:

```bash
helm install my-thunder .
```

If you want to customize the installation, create a `custom-values.yaml` file with your configurations and use:

```bash
helm install my-thunder . -f custom-values.yaml
```

The command deploys Thunder on the Kubernetes cluster with the default configuration. The [Parameters](#parameters) section lists the available parameters that can be configured during installation.

### 3. Access Thunder

### 4. Obtain the External IP

After deploying WSO2 Identity Server, you need to find its external IP address to access it outside the cluster. Run the following command to list the Ingress resources:

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
| `deployment.securityContext.enableRunAsUser` | Enable running as non-root user                                                    | `true`                         |
| `deployment.securityContext.runAsUser`  | User ID to run the container                                                            | `802`                          |
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

### Thunder Configuration Parameters

| Name                                   | Description                                                     | Default                      |
| -------------------------------------- | --------------------------------------------------------------- | ---------------------------- |
| `configuration.server.hostname`        | Thunder server hostname                                         | `0.0.0.0`                    |
| `configuration.server.port`            | Thunder server port                                             | `8090`                       |
| `configuration.gateClient.hostname`    | Gate client hostname                                            | `0.0.0.0`                    |
| `configuration.gateClient.port`        | Gate client port                                                | `8090`                       |
| `configuration.gateClient.scheme`      | Gate client scheme                                              | `https`                      |
| `configuration.gateClient.loginPath`   | Gate client login path                                          | `/signin`                     |
| `configuration.gateClient.errorPath`   | Gate client error path                                          | `/error`                     |
| `configuration.security.certFile`      | Server certificate file path                                    | `repository/resources/security/server.cert` |
| `configuration.security.keyFile`       | Server key file path                                            | `repository/resources/security/server.key`  |
| `configuration.database.identity.type` | Identity database type (postgres or sqlite)                     | `postgres`                   |
| `configuration.database.identity.sqlitePath` | SQLite database path (for sqlite only)                    | `repository/database/thunderdb.db` |
| `configuration.database.identity.sqliteOptions` | SQLite options (for sqlite only)                       | `_journal_mode=WAL&_busy_timeout=5000` |
| `configuration.database.identity.name` | Postgres database name (for postgres only)                      | `thunderdb`                  |
| `configuration.database.identity.host` | Postgres host (for postgres only)                               | `localhost` |
| `configuration.database.identity.port` | Postgres port (for postgres only)                               | `5432`                       |
| `configuration.database.identity.username` | Postgres username (for postgres only)                       | `asgthunder`                   |
| `configuration.database.identity.password` | Postgres password (for postgres only)                       | `asgthunder`              |
| `configuration.database.identity.sslmode` | Postgres SSL mode (for postgres only)                        | `require`                    |
| `configuration.database.runtime.type`  | Runtime database type (postgres or sqlite)                      | `postgres`                   |
| `configuration.database.runtime.sqlitePath` | SQLite database path (for sqlite only)                     | `repository/database/runtimedb.db` |
| `configuration.database.runtime.sqliteOptions` | SQLite options (for sqlite only)                        | `_journal_mode=WAL&_busy_timeout=5000` |
| `configuration.database.runtime.name`  | Postgres database name (for postgres only)                      | `runtimedb`                  |
| `configuration.database.runtime.host`  | Postgres host (for postgres only)                               | `localhost` |
| `configuration.database.runtime.port`  | Postgres port (for postgres only)                               | `5432`                       |
| `configuration.database.runtime.username` | Postgres username (for postgres only)                        | `asgthunder`                   |
| `configuration.database.runtime.password` | Postgres password (for postgres only)                        | `asgthunder`              |
| `configuration.database.runtime.sslmode` | Postgres SSL mode (for postgres only)                         | `require`                    |
| `configuration.cache.disabled`         | Disable cache                                                   | `false`                      |
| `configuration.cache.type`             | Cache type                                                      | `inmemory`                   |
| `configuration.cache.size`             | Cache size                                                      | `1000`                       |
| `configuration.cache.ttl`              | Cache TTL in seconds                                            | `3600`                       |
| `configuration.cache.evictionPolicy`   | Cache eviction policy                                           | `LRU`                        |
| `configuration.cache.cleanupInterval`  | Cache cleanup interval in seconds                               | `300`                        |
| `configuration.jwt.issuer`             | JWT issuer                                                      | `thunder`                    |
| `configuration.jwt.validityPeriod`     | JWT validity period in seconds                                  | `3600`                       |
| `configuration.oauth.refreshToken.renewOnGrant` | Renew refresh token on grant                           | `false`                      |
| `configuration.oauth.refreshToken.validityPeriod` | Refresh token validity period in seconds             | `86400`                      |
| `configuration.flow.graphDirectory`    | Flow graph directory                                            | `repository/resources/graphs/` |
| `configuration.flow.authn.defaultFlow` | Default authentication flow                                     | `auth_flow_config_basic`     |
| `configuration.cors.allowedOrigins`    | CORS allowed origins                                            | See values.yaml              |

### Custom Configuration

The Thunder configuration file (deployment.yaml) can be customized by overriding the default values in the values.yaml file.
Alternatively, you can directly update the values in conf/deployment.yaml before deploying the Helm chart.

### Database Configuration

Thunder supports both sqlite and postgres databases. By default, postgres is configured.

Make sure to create the necessary databases and users in your Postgres instance before deploying Thunder. The values.yaml should be overridden with the required database configurations for the DB created.

Note: Use sqlite only if you are running a single pod.
