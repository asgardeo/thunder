# Custom Bootstrap Guide

This guide explains how to extend Thunder's setup process with custom bootstrap scripts.

## Overview

Thunder provides an extensible bootstrap system that allows you to add custom initialization logic during setup. The bootstrap scripts are placed in the `bootstrap/` directory and are executed in alphanumeric order based on their filename prefix.

## Quick Start

### 1. Create a Custom Script

Create a new file in the `bootstrap/` directory. You can use either **Bash** (`.sh`) or **PowerShell** (`.ps1`) scripts.

#### Bash Script Example

```bash
cat > bootstrap/30-my-custom-setup.sh << 'EOF'
#!/bin/bash
set -e

# Source common functions (provides log_* and thunder_api_call)
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
source "${SCRIPT_DIR}/common.sh"

log_info "Creating custom user..."

RESPONSE=$(thunder_api_call POST "/users" '{
  "type": "person",
  "attributes": {
    "username": "alice",
    "password": "alice123",
    "sub": "alice",
    "email": "alice@example.com",
    "name": "Alice Johnson"
  }
}')

HTTP_CODE="${RESPONSE: -3}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "User created successfully"
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "User already exists"
else
    log_error "Failed to create user (HTTP $HTTP_CODE)"
    exit 1
fi
EOF

chmod +x bootstrap/30-my-custom-setup.sh
```

#### PowerShell Script Example

```powershell
# bootstrap/30-my-custom-setup.ps1
$ErrorActionPreference = 'Stop'

Log-Info "Creating custom user..."

$response = Invoke-ThunderApi -Method POST -Endpoint "/users" -Data '{
  "type": "person",
  "attributes": {
    "username": "alice",
    "password": "alice123",
    "sub": "alice",
    "email": "alice@example.com",
    "name": "Alice Johnson"
  }
}'

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "User created successfully"
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "User already exists"
}
else {
    Log-Error "Failed to create user (HTTP $($response.StatusCode))"
    exit 1
}
```

### 2. Run Setup

**Linux/macOS:**
```bash
./setup.sh
```

**Windows:**
```powershell
.\setup.ps1
```

The bootstrap system automatically discovers and executes all scripts in the `bootstrap/` directory in alphanumeric order. Both `.sh` and `.ps1` scripts can coexist in the same directory.

## Execution Order

All bootstrap scripts execute in alphanumeric order based on their filename prefix. 

### Recommended Naming

Use descriptive names with numeric prefixes. Both Bash (`.sh`) and PowerShell (`.ps1`) extensions are supported:

- ✅ `30-create-employee-schema.sh`
- ✅ `30-create-employee-schema.ps1`
- ✅ `35-import-users-from-ldap.sh`
- ✅ `40-create-mobile-app.ps1`
- ✅ `99-validate-setup.sh`
- ❌ `script1.sh`
- ❌ `test.sh`

**Note**: All scripts must be placed directly in the `bootstrap/` directory. On Windows with `setup.ps1`, both `.sh` (requires bash) and `.ps1` scripts will be discovered. On Linux/macOS with `setup.sh`, only `.sh` scripts will execute.

## Available Helper Functions

### Sourcing Common Functions (Bash Only)

**Bash scripts** must source the common functions file to access logging and API helper functions:

```bash
#!/bin/bash
set -e

# Source common functions from the bootstrap directory
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
source "${SCRIPT_DIR}/common.sh"

# Now you can use log_* and thunder_api_call functions
log_info "Starting custom setup..."
```

**How it works:**
- `SCRIPT_DIR` gets the directory where your script is located
- `source "${SCRIPT_DIR}/common.sh"` loads the shared functions from the same directory
- All scripts are placed in the `bootstrap/` directory alongside `common.sh`

### Logging Functions

Bootstrap scripts have access to color-coded logging functions:

**Bash:**
```bash
log_info "Informational message"      # Blue
log_success "Success message"          # Green
log_warning "Warning message"          # Yellow
log_error "Error message"              # Red
log_debug "Debug message"              # Cyan (only shown if DEBUG=true)
```

**PowerShell:**
```powershell
Log-Info "Informational message"      # Blue
Log-Success "Success message"          # Green
Log-Warning "Warning message"          # Yellow
Log-Error "Error message"              # Red
Log-Debug "Debug message"              # Cyan (only shown if $env:DEBUG="true")
```

### API Call Function

Make Thunder API calls easily:

**Bash:**
```bash
thunder_api_call METHOD ENDPOINT [JSON_DATA]
```

**PowerShell:**
```powershell
Invoke-ThunderApi -Method METHOD -Endpoint ENDPOINT [-Data JSON_DATA]
```

**Bash Examples**:

```bash
# GET request
RESPONSE=$(thunder_api_call GET "/users")

# POST request with data
RESPONSE=$(thunder_api_call POST "/users" '{
  "type": "person",
  "attributes": {
    "username": "bob",
    "email": "bob@example.com"
  }
}')

# Extract HTTP status code (last 3 characters)
HTTP_CODE="${RESPONSE: -3}"

# Extract response body (everything except last 3 characters)
BODY="${RESPONSE%???}"

# Check result
if [[ "$HTTP_CODE" == "201" ]]; then
    log_success "Resource created"
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "Resource already exists"
else
    log_error "Failed (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi
```

**PowerShell Examples**:

```powershell
# GET request
$response = Invoke-ThunderApi -Method GET -Endpoint "/users"

# POST request with data
$response = Invoke-ThunderApi -Method POST -Endpoint "/users" -Data '{
  "type": "person",
  "attributes": {
    "username": "bob",
    "email": "bob@example.com"
  }
}'

# Check result
if ($response.StatusCode -eq 201) {
    Log-Success "Resource created"
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "Resource already exists"
}
else {
    Log-Error "Failed (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}
```

### Environment Variables

These variables are available in both Bash and PowerShell bootstrap scripts:

| Variable | Description | Default | Access |
|----------|-------------|---------|--------|
| `THUNDER_API_BASE` | Thunder API base URL | `https://localhost:8090` | Bash: `$THUNDER_API_BASE`<br/>PS: `$env:THUNDER_API_BASE` |

## Common Use Cases

### Creating Custom User Types

```bash
#!/bin/bash
set -e

# Source common functions
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
source "${SCRIPT_DIR}/common.sh"

log_info "Creating employee user type..."

RESPONSE=$(thunder_api_call POST "/user-types" '{
  "name": "employee",
  "schema": {
    "sub": {"type": "string", "required": true, "unique": true},
    "email": {"type": "string", "required": true, "unique": true},
    "employee_id": {"type": "string", "required": true, "unique": true},
    "department": {"type": "string", "required": false},
    "job_title": {"type": "string", "required": false}
  }
}')

HTTP_CODE="${RESPONSE: -3}"
[[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "409" ]] || exit 1
log_success "Employee schema created"
```

### Bulk Importing Users

```bash
#!/bin/bash
set -e

# Source common functions
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
source "${SCRIPT_DIR}/common.sh"

log_info "Importing users..."

USERS=(
  '{"type":"person","attributes":{"username":"alice","password":"alice123","sub":"alice","email":"alice@example.com"}}'
  '{"type":"person","attributes":{"username":"bob","password":"bob123","sub":"bob","email":"bob@example.com"}}'
)

for user_data in "${USERS[@]}"; do
    username=$(echo "$user_data" | grep -o '"username":"[^"]*"' | cut -d'"' -f4)
    log_info "Creating user: $username"

    RESPONSE=$(thunder_api_call POST "/users" "$user_data")
    HTTP_CODE="${RESPONSE: -3}"

    if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "409" ]]; then
        log_success "User $username created/exists"
    else
        log_error "Failed to create $username"
    fi
done
```

### Creating OAuth Applications

```bash
#!/bin/bash
set -e

# Source common functions
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
source "${SCRIPT_DIR}/common.sh"

log_info "Creating mobile application..."

RESPONSE=$(thunder_api_call POST "/applications" '{
  "name": "Mobile App",
  "description": "Corporate mobile application",
  "url": "myapp://home",
  "auth_flow_id": "<auth-flow-uuid>",
  "inbound_auth_config": [{
    "type": "oauth2",
    "config": {
      "client_id": "MOBILE_APP",
      "redirect_uris": ["myapp://oauth/callback"],
      "grant_types": ["authorization_code", "refresh_token"],
      "response_types": ["code"],
      "pkce_required": true,
      "public_client": true
    }
  }]
}')

HTTP_CODE="${RESPONSE: -3}"
[[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "409" ]] || exit 1
log_success "Mobile app created"
```

## Containerized Deployments

### Understanding Default Scripts in Containers

Thunder provides default bootstrap scripts in the `/opt/thunder/bootstrap` directory:
- `common.sh` - Helper functions for logging and API calls
- `01-default-resources.sh` - Creates admin user, default organization, and Person user types
- `02-sample-resources.sh` - Creates sample resources for testing

**Important:** When using volume mounts, ConfigMaps, or COPY commands:
- **Mounting/copying to an entire directory** (e.g., `/opt/thunder/bootstrap`) replaces all default scripts.
- **To keep the defaults**, mount/copy individual script files to specific paths (e.g., `/opt/thunder/bootstrap/30-my-script.sh`)
- Your custom scripts can source `common.sh` for helper functions if you preserve it

**Recommended naming for custom scripts:** Use numeric prefixes (e.g., `30-my-users.sh`, `40-my-apps.sh`) to run after default scripts.

---

## Docker Deployment

### Volume Mount (Development)

Mount individual custom script files to preserve Thunder's default scripts (see [Understanding Default Scripts in Containers](#understanding-default-scripts-in-containers)).

**Make scripts executable:**
```bash
chmod +x custom-scripts/*.sh
```

**Run with docker run (mounting individual files):**
```bash
# Run setup with custom scripts (mount each file individually)
docker run -it --rm \
  -v "$(pwd)/custom-scripts/30-my-users.sh:/opt/thunder/bootstrap/30-my-users.sh:ro" \
  -v "$(pwd)/custom-scripts/40-my-apps.sh:/opt/thunder/bootstrap/40-my-apps.sh:ro" \
  ghcr.io/asgardeo/thunder:latest ./setup.sh

# Then start Thunder server
docker run -d \
  -p 8090:8090 \
  --name thunder \
  ghcr.io/asgardeo/thunder:latest
```

**Or use docker-compose (mounting individual files):**
```yaml
# docker-compose.yml
version: '3.8'
services:
  thunder-setup:
    image: ghcr.io/asgardeo/thunder:latest
    command: ./setup.sh
    volumes:
      # Mount each custom script individually to preserve built-in scripts
      - ./custom-scripts/30-my-users.sh:/opt/thunder/bootstrap/30-my-users.sh:ro
      - ./custom-scripts/40-my-apps.sh:/opt/thunder/bootstrap/40-my-apps.sh:ro
    restart: "no"

  thunder:
    image: ghcr.io/asgardeo/thunder:latest
    depends_on:
      thunder-setup:
        condition: service_completed_successfully
    ports:
      - "8090:8090"
```

### Custom Docker Image (Production)

Build a custom image with your scripts embedded. Using the wildcard pattern (`*.sh`) copies only your script files, adding them alongside the default scripts (see [Understanding Default Scripts in Containers](#understanding-default-scripts-in-containers)).

```dockerfile
FROM ghcr.io/asgardeo/thunder:latest

# Copy only custom script files (*.sh pattern adds files without replacing the directory)
COPY custom-scripts/*.sh /opt/thunder/bootstrap/

# Set permissions
USER root
RUN chmod +x /opt/thunder/bootstrap/*.sh && \
    chown -R thunder:thunder /opt/thunder/bootstrap
USER thunder
```

Build and use:

```bash
docker build -t thunder:custom .
docker run --rm thunder:custom ./setup.sh
docker run -d -p 8090:8090 thunder:custom
```

## Kubernetes Deployment

### Using ConfigMap

Create a ConfigMap with your custom scripts:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: thunder-custom-bootstrap
data:
  30-custom-users.sh: |
    #!/bin/bash
    set -e
    SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
    source "${SCRIPT_DIR}/common.sh"
    log_info "Creating custom users..."
    # Your script here

  40-custom-apps.sh: |
    #!/bin/bash
    set -e
    SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
    source "${SCRIPT_DIR}/common.sh"
    log_info "Creating custom applications..."
    # Your script here
```

Mount individual scripts using `subPath` to preserve default scripts (see [Understanding Default Scripts in Containers](#understanding-default-scripts-in-containers)):

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: thunder-setup
spec:
  template:
    spec:
      containers:
      - name: setup
        image: ghcr.io/asgardeo/thunder:latest
        command: ["./setup.sh"]
        volumeMounts:
        # Mount each custom script individually using subPath
        - name: custom-bootstrap
          mountPath: /opt/thunder/bootstrap/30-custom-users.sh
          subPath: 30-custom-users.sh
        - name: custom-bootstrap
          mountPath: /opt/thunder/bootstrap/40-custom-apps.sh
          subPath: 40-custom-apps.sh
      volumes:
      - name: custom-bootstrap
        configMap:
          name: thunder-custom-bootstrap
          defaultMode: 0755
      restartPolicy: OnFailure
```

### Using Helm

Thunder's Helm chart provides three flexible patterns for bootstrap scripts:

#### Pattern 1: Inline Scripts (Preserves Defaults)

Add scripts directly in `values.yaml`:

```yaml
bootstrap:
  scripts:
    30-custom-users.sh: |
      #!/bin/bash
      set -e
      SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
      source "${SCRIPT_DIR}/common.sh"

      log_info "Creating custom users..."
      thunder_api_call POST "/users" '{
        "type": "person",
        "attributes": {
          "username": "alice",
          "password": "alice123",
          "sub": "alice",
          "email": "alice@example.com"
        }
      }'
      log_success "Users created"

    40-custom-apps.sh: |
      #!/bin/bash
      set -e
      SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
      source "${SCRIPT_DIR}/common.sh"

      log_info "Creating custom application..."
      thunder_api_call POST "/applications" '{ ... }'
      log_success "Application created"
```

**Benefits:**
- ✅ Preserves Thunder's default scripts (`common.sh`, `01-*`, `02-*`)
- ✅ Simple for small scripts
- ✅ Version controlled with Helm values

---

#### Pattern 2: External ConfigMap with Specific Files (Preserves Defaults)

For larger scripts or separate management, create a ConfigMap and specify which files to mount:

**Step 1:** Create ConfigMap with your scripts
```bash
kubectl create configmap my-bootstrap \
  --from-file=30-users.sh=./scripts/30-users.sh \
  --from-file=40-apps.sh=./scripts/40-apps.sh
```

**Step 2:** Configure Helm to mount specific files
```yaml
bootstrap:
  configMap:
    name: "my-bootstrap"
    files:
      - 30-users.sh
      - 40-apps.sh
```

**Benefits:**
- ✅ Preserves Thunder's default scripts
- ✅ Scripts managed separately from Helm chart
- ✅ Easy to update scripts without Helm upgrade

---

#### Pattern 3: Replace All Scripts with Complete ConfigMap (Advanced)

⚠️ **WARNING**: Only use this if you need complete control and will provide ALL scripts including `common.sh`.

For complete replacement, create a ConfigMap with all scripts and **omit the files list**:

**Step 1:** Create complete ConfigMap (must include `common.sh`)
```bash
kubectl create configmap complete-bootstrap \
  --from-file=common.sh=./scripts/common.sh \
  --from-file=01-my-setup.sh=./scripts/01-my-setup.sh \
  --from-file=02-my-resources.sh=./scripts/02-my-resources.sh
```

**Step 2:** Configure Helm without specifying files
```yaml
bootstrap:
  configMap:
    name: "complete-bootstrap"
    # No files list = mounts entire ConfigMap (replaces all defaults)
```

**Important:**
- ⚠️ **Removes ALL default scripts** (including `common.sh`, admin user creation, etc.)
- ⚠️ You must provide your own `common.sh` with required helper functions
- ⚠️ No default resources will be created
- ✅ Complete control over bootstrap process

---

## Best Practices

### 1. Make Scripts Idempotent

Scripts should be safe to run multiple times:

```bash
# Check if resource exists before creating
RESPONSE=$(thunder_api_call GET "/user-types")
BODY="${RESPONSE%???}"

if echo "$BODY" | grep -q '"name":"employee"'; then
    log_info "Employee schema already exists, skipping"
    exit 0
fi

# Create the resource
...
```

### 2. Handle Errors Gracefully

```bash
set -e  # Exit on error

# But handle expected errors
if [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "Resource already exists (not an error)"
    exit 0
fi
```

### 3. Never Hardcode Secrets

```bash
# ❌ BAD
PASSWORD="mysecret123"

# ✅ GOOD: From environment
PASSWORD="${ADMIN_PASSWORD:-}"

# ✅ GOOD: From mounted secret
if [ -f "/run/secrets/admin-password" ]; then
    PASSWORD=$(cat /run/secrets/admin-password)
fi

if [ -z "$PASSWORD" ]; then
    log_error "Password not provided"
    exit 1
fi
```

### 4. Log Actions Clearly

```bash
log_info "Creating employee schema with 5 custom fields..."
# ... creation logic ...
log_success "Employee schema created: employee_id, department, job_title, manager_id, hire_date"
```

## Advanced Usage

### Skip Specific Scripts

```bash
# Skip any script with "test" in the name
BOOTSTRAP_SKIP_PATTERN="test" ./setup.sh
```

### Run Only Specific Scripts

```bash
# Only run scripts starting with "30-"
BOOTSTRAP_ONLY_PATTERN="^30-" ./setup.sh
```

### Continue on Errors

```bash
# Don't stop if a script fails
BOOTSTRAP_FAIL_FAST=false ./setup.sh
```

## Troubleshooting

### Script Not Executing

1. **Check permissions** (Linux/macOS only):
   ```bash
   chmod +x bootstrap/your-script.sh
   ```

2. **Check filename** - Must end with:
   - Bash: `.sh` or `.bash`
   - PowerShell: `.ps1`

3. **Check location** - Must be in `bootstrap/` directory

4. **On Windows with `.sh` scripts** - Requires bash (Git Bash, WSL, etc.)

### Script Failing

1. **Enable debug mode**:

   **Bash:**
   ```bash
   DEBUG=true ./setup.sh
   ```

   **PowerShell:**
   ```powershell
   $env:DEBUG = "true"
   .\setup.ps1
   ```

2. **Check logs** for HTTP error codes

3. **Test API calls** manually:

   **Linux/macOS:**
   ```bash
   curl -k -X GET https://localhost:8090/users
   ```

   **Windows:**
   ```powershell
   Invoke-WebRequest -Uri https://localhost:8090/users -SkipCertificateCheck
   ```
