# Declarative Configuration Mode

This guide explains how to run Thunder with declarative configurations loaded from YAML files, enabling GitOps workflows and declarative infrastructure management.

## Overview

Declarative Configuration Mode allows Thunder to load resource configurations from YAML files at startup instead of using the database. This approach provides several benefits:

- **GitOps Friendly:** Manage configurations as code in version control
- **Environment Consistency:** Use the same configuration files across environments
- **Declarative Management:** Define infrastructure as code
- **Audit Trail:** Track all configuration changes through git history
- **Quick Deployment:** Fast server startup with pre-defined configurations

## How It Works

When declarative configuration mode is enabled:

1. Thunder starts and reads YAML files from `repository/resources/`
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
│  (Declarative Mode)       │
│                         │
│  • Load configs at boot │
│  • Read-only via API    │
│  • In-memory storage    │
└─────────────────────────┘
```

## Enabling Declarative Configuration Mode

### 1. Global Configuration

Edit `repository/conf/deployment.yaml`:

```yaml
declarative_resources:
  enabled: true
```

This enables declarative mode for **all** supported resources.

### 2. Service-Level Configuration (Organization Units)

Organization Units support fine-grained control with three store modes:

```yaml
organization_unit:
  store: "mutable"      # Options: "mutable", "immutable", "composite"
```

#### Store Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `mutable` | Database-only storage. Full CRUD operations. | Development, dynamic environments |
| `immutable` | File-based storage only (YAML). Read-only via API. | Production, GitOps workflows |
| `composite` (hybrid) | Both file-based (immutable) + database (mutable). Reads merge both stores, writes to database only. | Mixed environments with predefined OUs + runtime OUs |

#### Examples

**Mutable Mode (Default):**
```yaml
# All OUs stored in database, full CRUD via API
organization_unit:
  store: "mutable"
```

**Declarative Mode:**
```yaml
# All OUs loaded from YAML files, read-only via API
organization_unit:
  store: "immutable"
```

**Composite Mode:**
```yaml
# Predefined OUs from YAML (immutable) + runtime OUs in database (mutable)
organization_unit:
  store: "composite"
```

#### Configuration Fallback

If `organization_unit.store` is not specified, it falls back to the global `declarative_resources.enabled` setting:

- If `declarative_resources.enabled = true` → behaves as **immutable** mode
- If `declarative_resources.enabled = false` → behaves as **mutable** mode

**Example:**
```yaml
# Global declarative mode
declarative_resources:
  enabled: true

# OU will use "immutable" mode (fallback)
# organization_unit:
#   store: not specified

# To override for OUs specifically:
organization_unit:
  store: "composite"  # OUs use composite mode despite global immutable=true
```

#### Composite Mode Behavior

In composite mode:

1. **Reads:** Merge results from both file-based and database stores
2. **List operations:** Return all OUs (both immutable from YAML and mutable from DB)
3. **Create operations:** New OUs go to database store
4. **Update operations:** 
   - File-based OUs (immutable): Returns error "Cannot update immutable OU"
   - Database OUs (mutable): Update succeeds
5. **Delete operations:**
   - File-based OUs (immutable): Returns error "Cannot delete immutable OU"
   - Database OUs (mutable): Delete succeeds

**Example Composite Setup:**
```yaml
organization_unit:
  store: "composite"
```

```
repository/resources/
└── organization_units/
    ├── production.yaml      # Immutable OU from YAML
    ├── staging.yaml         # Immutable OU from YAML
    └── development.yaml     # Immutable OU from YAML
```

At runtime:
- YAML OUs (`production`, `staging`, `development`) are **read-only**
- New OUs created via API are stored in **database** and **mutable**
- List API returns all OUs from both sources

## Directory Structure

Place configuration files in the `repository/resources/` directory:

```
repository/resources/
├── applications/
│   ├── my-web-app.yaml
│   ├── mobile-app.yaml
│   └── admin-portal.yaml
├── identity_providers/
│   ├── google-idp.yaml
│   ├── github-idp.yaml
│   └── oidc-idp.yaml
├── organization_units/
│   ├── production.yaml
│   ├── staging.yaml
│   └── development.yaml
├── flows/
│   ├── auth-flow-basic.yaml
│   ├── auth-flow-mfa.yaml
│   └── registration-flow.yaml
└── notification_senders/
    └── smtp-sender.yaml
```

### Supported Resource Types

| Resource Type | Directory | Store Modes | Status |
|---------------|-----------|-------------|--------|
| Applications | `applications/` | Global only | ✅ Supported |
| Identity Providers | `identity_providers/` | Global only | ✅ Supported |
| Organization Units | `organization_units/` | mutable / immutable / composite | ✅ Supported |
| Flow Graphs | `flows/` | Global only | ✅ Supported |
| Notification Senders | `notification_senders/` | Global only | ✅ Supported |

## Creating Configuration Files

### Using Export API (Recommended)

The easiest way to create configuration files is to export them from a running Thunder instance:

```bash
# Export an application
curl -X POST https://localhost:8090/export \
  -H "Content-Type: application/json" \
  -d '{
    "applications": ["<application-id>"]
  }' > repository/resources/applications/my-app.yaml

# Export an identity provider
curl -X POST https://localhost:8090/export \
  -H "Content-Type: application/json" \
  -d '{
    "identity_providers": ["<idp-id>"]
  }' > repository/resources/identity_providers/google-idp.yaml

# Export specific organization units
curl -X POST https://localhost:8090/export \
  -H "Content-Type: application/json" \
  -d '{
    "organization_units": ["<ou-id-1>", "<ou-id-2>"]
  }' > repository/resources/organization_units/my-ous.yaml
```

See the [Export Configurations Guide](./export-configurations.md) for detailed export instructions.

**⚠️ Organization Units Export Limitation:**

When using wildcard export for organization units (`"organization_units": ["*"]`), only the first 100 organization units will be retrieved due to pagination limits. If you have more than 100 OUs at the root level, they will not be included in the export. 

To export all organization units:
- Export specific OUs by ID rather than using wildcards
- Make multiple export requests with specific OU IDs
- Consider implementing pagination loop logic in your export script to fetch all pages

### Manual Creation

You can also create YAML files manually.

#### Application Configuration Example

```yaml
# repository/resources/applications/my-app.yaml
name: My Application
description: Production web application
url: https://myapp.example.com
logo_url: https://myapp.example.com/logo.png
auth_flow_id: edc013d0-e893-4dc0-990c-3e1d203e005b
registration_flow_id: 80024fb3-29ed-4c33-aa48-8aee5e96d522
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

#### Flow Graph Configuration Example

Flow graphs define authentication and registration flows. Here's an example authentication flow:

```yaml
# repository/resources/flows/auth-flow-basic.yaml
id: "auth-flow-001"
handle: "basic-auth-flow"
name: "Basic Authentication Flow"
flowType: "AUTHENTICATION"
activeVersion: 1
nodes:
  - id: "start"
    type: "START"
    onSuccess: "prompt_credentials"
  
  - id: "prompt_credentials"
    type: "PROMPT"
    meta:
      components:
        - type: "TEXT"
          id: "text_001"
          label: "Sign In"
          variant: "HEADING_1"
        - type: "BLOCK"
          id: "block_001"
          components:
            - id: "input_001"
              ref: "username"
              type: "TEXT_INPUT"
              label: "Username"
              required: true
              placeholder: "Enter your username"
            - id: "input_002"
              ref: "password"
              type: "PASSWORD_INPUT"
              label: "Password"
              required: true
              placeholder: "Enter your password"
            - type: "ACTION"
              id: "action_001"
              label: "Sign In"
              variant: "PRIMARY"
              eventType: "SUBMIT"
    inputs:
      - ref: "input_001"
        identifier: "username"
        type: "TEXT_INPUT"
        required: true
      - ref: "input_002"
        identifier: "password"
        type: "PASSWORD_INPUT"
        required: true
    actions:
      - ref: "action_001"
        nextNode: "basic_auth"
  
  - id: "basic_auth"
    type: "TASK_EXECUTION"
    executor:
      name: "BasicAuthExecutor"
    onSuccess: "authorization_check"
  
  - id: "authorization_check"
    type: "TASK_EXECUTION"
    executor:
      name: "AuthorizationExecutor"
    onSuccess: "auth_assert"
  
  - id: "auth_assert"
    type: "TASK_EXECUTION"
    executor:
      name: "AuthAssertExecutor"
    onSuccess: "end"
  
  - id: "end"
    type: "END"
```

**Flow Graph Node Types:**

- `START` - Entry point of the flow
- `PROMPT` - User interface component for collecting input
- `TASK_EXECUTION` - Execute a specific task (authentication, authorization, etc.)
- `END` - Terminal node of the flow

**Common Flow Types:**

- `AUTHENTICATION` - User login flows
- `REGISTRATION` - User registration flows

**Executors:**

Executors are the business logic components that process authentication steps:

- `BasicAuthExecutor` - Username/password authentication
- `AuthorizationExecutor` - Check user authorization
- `AuthAssertExecutor` - Final authentication assertion
- `TOTPAuthExecutor` - Time-based one-time password (MFA)
- `SMSOTPExecutor` - SMS-based OTP
- `EmailOTPExecutor` - Email-based OTP

#### Multi-Factor Authentication Flow Example

```yaml
# repository/resources/flows/auth-flow-mfa.yaml
id: "auth-flow-mfa-001"
handle: "mfa-auth-flow"
name: "Multi-Factor Authentication Flow"
flowType: "AUTHENTICATION"
activeVersion: 1
nodes:
  - id: "start"
    type: "START"
    onSuccess: "basic_auth"
  
  - id: "basic_auth"
    type: "TASK_EXECUTION"
    executor:
      name: "BasicAuthExecutor"
    onSuccess: "totp_prompt"
  
  - id: "totp_prompt"
    type: "PROMPT"
    meta:
      components:
        - type: "TEXT"
          label: "Enter Verification Code"
        - type: "BLOCK"
          components:
            - id: "totp_input"
              ref: "totp_code"
              type: "TEXT_INPUT"
              label: "TOTP Code"
              required: true
    inputs:
      - ref: "totp_input"
        identifier: "totp_code"
        type: "TEXT_INPUT"
        required: true
    actions:
      - nextNode: "totp_verify"
  
  - id: "totp_verify"
    type: "TASK_EXECUTION"
    executor:
      name: "TOTPAuthExecutor"
    onSuccess: "authorization_check"
  
  - id: "authorization_check"
    type: "TASK_EXECUTION"
    executor:
      name: "AuthorizationExecutor"
    onSuccess: "end"
  
  - id: "end"
    type: "END"
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
  -v $(pwd)/declarative_resources:/app/repository/resources \
  thunder:latest
```

Or use an env file:

```bash
docker run --env-file production.env \
  -v $(pwd)/declarative_resources:/app/repository/resources \
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

## API Behavior in Declarative Mode

When declarative configuration mode is enabled:

### Read Operations (Allowed)

✅ **GET /applications** - List applications  
✅ **GET /applications/\{id\}** - Get application details  
✅ **GET /flows** - List flow graphs  
✅ **GET /flows/\{id\}** - Get flow graph details  

### Write Operations (Disabled)

❌ **POST /applications** - Returns error  
❌ **PUT /applications/\{id\}** - Returns error  
❌ **DELETE /applications/\{id\}** - Returns error  
❌ **POST /flows** - Returns error  
❌ **PUT /flows/\{id\}** - Returns error  
❌ **DELETE /flows/\{id\}** - Returns error


## Best Practices

### 1. Version Control

Store configuration files in git:

```bash
git add repository/resources/
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
yamllint repository/resources/**/*.yaml

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

**Symptom:** Application or flow graph not found after startup.

**Cause:** File not in correct directory or invalid YAML.

**Solution:**
1. Verify file location:
   - Applications: `repository/resources/applications/`
   - Flow graphs: `repository/resources/flows/`
   - Identity providers: `repository/resources/identity_providers/`
2. Check YAML syntax: `yamllint my-config.yaml`
3. Check server logs for parsing errors
4. Ensure the `id` and `handle` fields are unique

### Cannot Create Applications or Flow Graphs

**Symptom:** POST /applications or POST /flows returns error.

**Cause:** Immutable mode is enabled.

**Solution:**
This is expected behavior. To add new resources:
1. Create a new YAML file in the appropriate directory:
   - Applications: `repository/resources/applications/`
   - Flow graphs: `repository/resources/flows/`
2. Restart Thunder
3. Or disable declarative mode to use API

### Flow Graph Validation Errors

**Symptom:** Flow graph fails to load with validation error.

**Cause:** Invalid flow structure or missing required nodes.

**Solution:**
1. Ensure flow has both `START` and `END` nodes
2. Verify all node connections (`onSuccess`, `nextNode`) reference valid node IDs
3. Check that PROMPT nodes have matching `inputs` and `actions`
4. Ensure `flowType` is either `AUTHENTICATION` or `REGISTRATION`
5. Validate executor names match available executors

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
   chmod 700 repository/resources/
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
- **Documentation:** [Thunder Guides](../README.md)
