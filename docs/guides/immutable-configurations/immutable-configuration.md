# Immutable Configuration Mode

This guide explains how to run Thunder with immutable configurations loaded from YAML files, enabling GitOps workflows and declarative infrastructure management.

## Overview

Immutable Configuration Mode allows Thunder to load resource configurations from YAML files at startup instead of using the database. This approach provides several benefits:

- **GitOps Friendly:** Manage configurations as code in version control
- **Environment Consistency:** Use the same configuration files across environments
- **Declarative Management:** Define infrastructure as code
- **Audit Trail:** Track all configuration changes through git history
- **Quick Deployment:** Fast server startup with pre-defined configurations

## How It Works

When immutable configuration mode is enabled:

1. Thunder starts and reads YAML files from `repository/conf/immutable_resources/`
2. Configurations are loaded into memory (not the database)
3. Create, Update, and Delete operations are **disabled** via API
4. Applications use the file-based configurations
5. Changes require updating YAML files and restarting Thunder

```
┌─────────────────────────┐
│   YAML Config Files     │
│  (Version Controlled)   │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│   Thunder Server        │
│  (Immutable Mode)       │
│                         │
│  • Load configs at boot │
│  • Read-only via API    │
│  • In-memory storage    │
└─────────────────────────┘
```

## Enabling Immutable Configuration Mode

### 1. Configuration File

Edit `repository/conf/deployment.yaml`:

```yaml
immutable_resources:
  enabled: true
```

## Directory Structure

Place configuration files in the `repository/conf/immutable_resources/` directory:

```
repository/conf/immutable_resources/
├── applications/
│   ├── my-web-app.yaml
│   ├── mobile-app.yaml
│   └── admin-portal.yaml
├── identity-providers/
│   ├── google-idp.yaml
│   ├── github-idp.yaml
│   └── oidc-idp.yaml
└── notification-senders/        # Coming soon
    └── smtp-sender.yaml
```

### Supported Resource Types

| Resource Type | Directory | Status |
|---------------|-----------|--------|
| Applications | `applications/` | ✅ Supported |
| Identity Providers | `identity-providers/` | ✅ Supported |
| Notification Senders | `notification-senders/` | 🔜 Coming Soon |
| Groups | `groups/` | 🔜 Coming Soon |
| Roles | `roles/` | 🔜 Coming Soon |

## Creating Configuration Files

### Using Export API (Recommended)

The easiest way to create configuration files is to export them from a running Thunder instance:

```bash
# Export an application
curl -X POST https://localhost:8090/export \
  -H "Content-Type: application/json" \
  -d '{
    "applications": ["<application-id>"]
  }' > repository/conf/immutable_resources/applications/my-app.yaml

# Export an identity provider
curl -X POST https://localhost:8090/export \
  -H "Content-Type: application/json" \
  -d '{
    "identity_providers": ["<idp-id>"]
  }' > repository/conf/immutable_resources/identity-providers/google-idp.yaml
```

See the [Export Configurations Guide](./export-configurations.md) for detailed export instructions.

### Manual Creation

You can also create YAML files manually. Here's an example application configuration:

```yaml
# repository/conf/immutable_resources/applications/my-app.yaml
name: My Application
description: Production web application
url: https://myapp.example.com
logo_url: https://myapp.example.com/logo.png
auth_flow_graph_id: auth_flow_config_basic
registration_flow_graph_id: registration_flow_config_basic
is_registration_flow_enabled: true
inbound_auth_config:
  - type: oauth2
    config:
      client_id: {{.MY_APPLICATION_CLIENT_ID}}
      client_secret: {{.MY_APPLICATION_CLIENT_SECRET}}
      redirect_uris:
        {{- range .MY_APPLICATION_REDIRECT_URIS}}
        - {{.}}
        {{- end}}
      grant_types:
        - authorization_code
        - refresh_token
      response_types:
        - code
      token_endpoint_auth_method: client_secret_basic
      pkce_required: false
      public_client: false
      token:
        issuer: thunder
        access_token:
          validity_period: 3600
          user_attributes:
            - email
            - name
            - groups
        id_token:
          validity_period: 3600
          user_attributes:
            - email
            - name
```

## Parameterized Variables

Configuration files support Go template syntax for environment-specific values. This allows you to use the same configuration across different environments.

### Variable Substitution

Use the `{{.VARIABLE_NAME}}` syntax for simple values:

```yaml
client_id: {{.MY_APP_CLIENT_ID}}
client_secret: {{.MY_APP_CLIENT_SECRET}}
```

**Environment variables:**
```bash
export MY_APP_CLIENT_ID=prod-client-id
export MY_APP_CLIENT_SECRET=prod-secret
```

### Array Variables

Use the `{{- range .ARRAY_VAR}}` syntax for arrays:

```yaml
redirect_uris:
  {{- range .MY_APP_REDIRECT_URIS}}
  - {{.}}
  {{- end}}
```

**Environment variables (indexed):**
```bash
export MY_APP_REDIRECT_URIS_0=https://app.example.com/callback
export MY_APP_REDIRECT_URIS_1=https://app.example.com/silent-callback
export MY_APP_REDIRECT_URIS_2=https://app.example.com/logout
```

Thunder automatically builds the array by reading `VARNAME_0`, `VARNAME_1`, `VARNAME_2`, etc., until it finds an empty or non-existent variable.

### Supported Parameterization

Currently, the following application fields support parameterization:

| Field | Type | Template Syntax |
|-------|------|-----------------|
| `client_id` | Variable | `{{.APP_NAME_CLIENT_ID}}` |
| `client_secret` | Variable | `{{.APP_NAME_CLIENT_SECRET}}` |
| `redirect_uris` | Array | `{{- range .APP_NAME_REDIRECT_URIS}}` |

**Note:** Variable names are automatically generated during export. For manual configurations, follow the naming convention: `APP_NAME_FIELD_NAME` (uppercase, underscores for spaces).

## Providing Variable Values

There are multiple ways to provide values for parameterized variables:

### 1. Environment Variables (Recommended)

Set environment variables before starting Thunder:

```bash
export MY_APP_CLIENT_ID=my-client-id
export MY_APP_CLIENT_SECRET=my-secret
export MY_APP_REDIRECT_URIS_0=https://app.example.com/callback
export MY_APP_REDIRECT_URIS_1=https://app.example.com/logout

./backend/server
```

### 2. Environment Files

Use `.env` files for easier management:

**production.env:**
```bash
# Application: My Application
MY_APP_CLIENT_ID=prod-client-id
MY_APP_CLIENT_SECRET=prod-secret-value
MY_APP_REDIRECT_URIS_0=https://app.example.com/callback
MY_APP_REDIRECT_URIS_1=https://app.example.com/silent-callback

# Application: Mobile App
MOBILE_APP_CLIENT_ID=mobile-client-id
MOBILE_APP_CLIENT_SECRET=mobile-secret
MOBILE_APP_REDIRECT_URIS_0=myapp://callback
```

Load the environment file before starting:

```bash
source production.env
./backend/server
```

Or use `export` command:

```bash
export $(cat production.env | xargs)
./backend/server
```

### 3. Docker/Container Environments

Pass environment variables via Docker:

```bash
docker run \
  -e MY_APP_CLIENT_ID=client-id \
  -e MY_APP_CLIENT_SECRET=secret \
  -e MY_APP_REDIRECT_URIS_0=https://app.example.com/callback \
  -v $(pwd)/immutable_resources:/app/repository/conf/immutable_resources \
  thunder:latest
```

Or use an env file:

```bash
docker run --env-file production.env \
  -v $(pwd)/immutable_resources:/app/repository/conf/immutable_resources \
  thunder:latest
```

### 4. Kubernetes Secrets

Store sensitive values in Kubernetes secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: thunder-app-secrets
type: Opaque
stringData:
  MY_APP_CLIENT_ID: prod-client-id
  MY_APP_CLIENT_SECRET: prod-secret
  MY_APP_REDIRECT_URIS_0: https://app.example.com/callback
  MY_APP_REDIRECT_URIS_1: https://app.example.com/silent-callback
```

Reference in the deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: thunder
spec:
  template:
    spec:
      containers:
      - name: thunder
        envFrom:
        - secretRef:
            name: thunder-app-secrets
```

### 5. Helm Values

For Helm deployments, use values.yaml:

```yaml
# values.yaml
env:
  - name: MY_APP_CLIENT_ID
    value: "prod-client-id"
  - name: MY_APP_CLIENT_SECRET
    valueFrom:
      secretKeyRef:
        name: thunder-secrets
        key: client-secret
  - name: MY_APP_REDIRECT_URIS_0
    value: "https://app.example.com/callback"
  - name: MY_APP_REDIRECT_URIS_1
    value: "https://app.example.com/silent-callback"
```

## API Behavior in Immutable Mode

When immutable configuration mode is enabled:

### Read Operations (Allowed)

✅ **GET /applications** - List applications  
✅ **GET /applications/{id}** - Get application details  
✅ **GET /oauth2/token** - OAuth endpoints (authentication works normally)

### Write Operations (Disabled)

❌ **POST /applications** - Returns error  
❌ **PUT /applications/{id}** - Returns error  
❌ **DELETE /applications/{id}** - Returns error


## Best Practices

### 1. Version Control

Store configuration files in git:

```bash
git add repository/conf/immutable_resources/
git commit -m "Add production application configs"
git tag v1.0.0
```

### 2. Separate Secrets from Config

**DO:**
- Store YAML configs in git ✅
- Use parameterized variables for secrets ✅
- Store actual secrets in secure vaults ✅

**DON'T:**
- Commit `.env` files with actual secrets ❌
- Hardcode secrets in YAML files ❌

### 3. Environment-Specific Branches

Use git branches for different environments:

```bash
# Development
git checkout develop
./deploy.sh dev

# Production
git checkout main
./deploy.sh prod
```

### 4. Validation Before Deployment

Validate configurations before deployment:

```bash
# Check for syntax errors
yamllint repository/conf/immutable_resources/**/*.yaml

# Verify all variables are set
./scripts/validate-env.sh production.env
```


## Troubleshooting

### Variables Not Substituted

**Symptom:** YAML contains `{{.VARIABLE}}` in application configuration.

**Cause:** Environment variable not set.

**Solution:**
```bash
# Check if variable is set
echo $MY_APP_CLIENT_ID

# Set the variable
export MY_APP_CLIENT_ID=my-value

# Restart Thunder
```

### Array Variables Empty

**Symptom:** Array fields are empty in loaded configuration.

**Cause:** Incorrect array variable indexing.

**Solution:**
Ensure array variables start at `_0` and are sequential:
```bash
export MY_APP_REDIRECT_URIS_0=https://example.com/callback
export MY_APP_REDIRECT_URIS_1=https://example.com/logout
# NOT: MY_APP_REDIRECT_URIS_2 (must be sequential without gaps in numbers)
```

### Configuration File Not Loaded

**Symptom:** Application not found after startup.

**Cause:** File not in correct directory or invalid YAML.

**Solution:**
1. Verify file location: `repository/conf/immutable_resources/applications/`
2. Check YAML syntax: `yamllint my-app.yaml`
3. Check server logs for parsing errors

### Cannot Create Applications

**Symptom:** POST /applications returns error.

**Cause:** Immutable mode is enabled.

**Solution:**
This is expected behavior. To add new applications:
1. Create a new YAML file in `immutable_resources/applications/`
2. Restart Thunder
3. Or disable immutable mode to use API

## Security Considerations

⚠️ **Important:**

1. **Never commit secrets to version control:**
   ```bash
   # .gitignore
   *.env
   secrets/
   .env.*
   ```

2. **Use secret management systems:**
   - AWS Secrets Manager
   - HashiCorp Vault
   - Kubernetes Secrets
   - Azure Key Vault

3. **Restrict file permissions:**
   ```bash
   chmod 600 environments/*.env
   chmod 700 repository/conf/immutable_resources/
   ```

4. **Rotate secrets regularly:**
   Update environment variables and restart Thunder.

5. **Audit configuration changes:**
   Use git history to track all changes to configuration files.

## Next Steps

- Learn about [Exporting Configurations](./export-configurations.md)
- Review [Application Management API](/api/application.yaml)

## Support

For issues or questions:
- **GitHub Issues:** [Report a bug](https://github.com/asgardeo/thunder/issues)
- **Documentation:** [Thunder Guides](/docs/guides/README.md)
