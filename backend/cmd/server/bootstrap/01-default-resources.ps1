#!/usr/bin/env pwsh
# ----------------------------------------------------------------------------
# Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
#
# WSO2 LLC. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.
# ----------------------------------------------------------------------------

# Check for PowerShell Version Compatibility
if ($PSVersionTable.PSVersion.Major -lt 7) {
    Write-Host ""
    Write-Host "================================================================" -ForegroundColor Red
    Write-Host " [ERROR] UNSUPPORTED POWERSHELL VERSION" -ForegroundColor Red
    Write-Host "================================================================" -ForegroundColor Red
    Write-Host ""
    Write-Host " You are currently running PowerShell $($PSVersionTable.PSVersion.ToString())" -ForegroundColor Yellow
    Write-Host " Thunder requires PowerShell 7 (Core) or later." -ForegroundColor Yellow
    Write-Host ""
    Write-Host " Please install the latest version from:"
    Write-Host " https://github.com/PowerShell/PowerShell" -ForegroundColor Cyan
    Write-Host ""
    exit 1
}

# Bootstrap Script: Default Resources Setup
# Creates default organization unit, user schema, admin user, system resource server, system action, admin role, and DEVELOP application

$ErrorActionPreference = 'Stop'

Log-Info "Creating default Thunder resources..."
Write-Host ""

# ============================================================================
# Create Default Organization Unit
# ============================================================================

Log-Info "Creating default organization unit..."

$response = Invoke-ThunderApi -Method POST -Endpoint "/organization-units" -Data '{
  "handle": "default",
  "name": "Default",
  "description": "Default organization unit"
}'

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "Organization unit created successfully"
    $body = $response.Body | ConvertFrom-Json
    $DEFAULT_OU_ID = $body.id
    if ($DEFAULT_OU_ID) {
        Log-Info "Default OU ID: $DEFAULT_OU_ID"
    }
    else {
        Log-Error "Could not extract OU ID from response"
        exit 1
    }
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "Organization unit already exists, retrieving OU ID..."
    # Get existing OU ID
    $response = Invoke-ThunderApi -Method GET -Endpoint "/organization-units"

    if ($response.StatusCode -eq 200) {
        $body = $response.Body | ConvertFrom-Json
        $DEFAULT_OU_ID = $body.organizationUnits[0].id
        if ($DEFAULT_OU_ID) {
            Log-Success "Found OU ID: $DEFAULT_OU_ID"
        }
        else {
            Log-Error "Could not find OU ID in response"
            exit 1
        }
    }
    else {
        Log-Error "Failed to fetch organization units (HTTP $($response.StatusCode))"
        exit 1
    }
}
else {
    Log-Error "Failed to create organization unit (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

Write-Host ""

# ============================================================================
# Create Default User Schema
# ============================================================================

Log-Info "Creating default user schema (person)..."

$userSchemaData = ([ordered]@{
    name = "Person"
    ouId = $DEFAULT_OU_ID
    schema = [ordered]@{
        username = @{
            type = "string"
            required = $true
            unique = $true
        }
        email = @{
            type = "string"
            required = $true
            unique = $true
        }
        email_verified = @{
            type = "boolean"
            required = $false
        }
        given_name = @{
            type = "string"
            required = $false
        }
        family_name = @{
            type = "string"
            required = $false
        }
        phone_number = @{
            type = "string"
            required = $false
        }
        phone_number_verified = @{
            type = "boolean"
            required = $false
        }
    }
} | ConvertTo-Json -Depth 5)

$response = Invoke-ThunderApi -Method POST -Endpoint "/user-schemas" -Data $userSchemaData

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "User schema created successfully"
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "User schema already exists, skipping"
}
else {
    Log-Error "Failed to create user schema (HTTP $($response.StatusCode))"
    exit 1
}

Write-Host ""

# ============================================================================
# Create Admin User
# ============================================================================

Log-Info "Creating admin user..."

$adminUserData = ([ordered]@{
    type = "Person"
    organizationUnit = $DEFAULT_OU_ID
    attributes = @{
        username = "admin"
        password = "admin"
        sub = "admin"
        email = "admin@thunder.dev"
        email_verified = $true
        name = "Administrator"
        given_name = "Admin"
        family_name = "User"
        picture = "https://example.com/avatar.jpg"
        phone_number = "+12345678920"
        phone_number_verified = $true
    }
} | ConvertTo-Json -Depth 5)

$response = Invoke-ThunderApi -Method POST -Endpoint "/users" -Data $adminUserData

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "Admin user created successfully"
    Log-Info "Username: admin"
    Log-Info "Password: admin"

    # Extract admin user ID
    $body = $response.Body | ConvertFrom-Json
    $ADMIN_USER_ID = $body.id
    if (-not $ADMIN_USER_ID) {
        Log-Warning "Could not extract admin user ID from response"
    }
    else {
        Log-Info "Admin user ID: $ADMIN_USER_ID"
    }
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "Admin user already exists, retrieving user ID..."

    # Get existing admin user ID
    $response = Invoke-ThunderApi -Method GET -Endpoint "/users"

    if ($response.StatusCode -eq 200) {
        # Parse JSON to find admin user
        $body = $response.Body | ConvertFrom-Json
        $adminUser = $body.users | Where-Object { $_.attributes.username -eq "admin" } | Select-Object -First 1

        if ($adminUser) {
            $ADMIN_USER_ID = $adminUser.id
            Log-Success "Found admin user ID: $ADMIN_USER_ID"
        }
        else {
            Log-Error "Could not find admin user in response"
            exit 1
        }
    }
    else {
        Log-Error "Failed to fetch users (HTTP $($response.StatusCode))"
        exit 1
    }
}
else {
    Log-Error "Failed to create admin user (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

Write-Host ""

# ============================================================================
# Create System Resource Server
# ============================================================================

Log-Info "Creating system resource server..."

if (-not $DEFAULT_OU_ID) {
    Log-Error "Default OU ID is not available. Cannot create resource server."
    exit 1
}

$resourceServerData = @{
    name = "System"
    description = "System resource server"
    identifier = "system"
    ouId = $DEFAULT_OU_ID
} | ConvertTo-Json -Depth 10

$response = Invoke-ThunderApi -Method POST -Endpoint "/resource-servers" -Data $resourceServerData

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "Resource server created successfully"
    $body = $response.Body | ConvertFrom-Json
    $SYSTEM_RS_ID = $body.id
    if ($SYSTEM_RS_ID) {
        Log-Info "System resource server ID: $SYSTEM_RS_ID"
    }
    else {
        Log-Error "Could not extract resource server ID from response"
        exit 1
    }
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "Resource server already exists, retrieving ID..."
    # Get existing resource server ID
    $response = Invoke-ThunderApi -Method GET -Endpoint "/resource-servers"

    if ($response.StatusCode -eq 200) {
        $body = $response.Body | ConvertFrom-Json
        $systemRS = $body.resourceServers | Where-Object { $_.identifier -eq "system" } | Select-Object -First 1

        if ($systemRS) {
            $SYSTEM_RS_ID = $systemRS.id
            Log-Success "Found resource server ID: $SYSTEM_RS_ID"
        }
        else {
            Log-Error "Could not find resource server ID in response"
            exit 1
        }
    }
    else {
        Log-Error "Failed to fetch resource servers (HTTP $($response.StatusCode))"
        exit 1
    }
}
else {
    Log-Error "Failed to create resource server (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

Write-Host ""

# ============================================================================
# Create System Action
# ============================================================================

Log-Info "Creating 'system' action on resource server..."

if (-not $SYSTEM_RS_ID) {
    Log-Error "System resource server ID is not available. Cannot create action."
    exit 1
}

$actionData = @{
    name = "System Access"
    description = "Full system access permission"
    handle = "system"
} | ConvertTo-Json -Depth 10

$response = Invoke-ThunderApi -Method POST -Endpoint "/resource-servers/$SYSTEM_RS_ID/actions" -Data $actionData

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "System action created successfully"
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "System action already exists, skipping"
}
else {
    Log-Error "Failed to create system action (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

Write-Host ""

# ============================================================================
# Create Admin Role
# ============================================================================

Log-Info "Creating admin role with 'system' permission..."

if (-not $ADMIN_USER_ID) {
    Log-Error "Admin user ID is not available. Cannot create role."
    exit 1
}

if (-not $DEFAULT_OU_ID) {
    Log-Error "Default OU ID is not available. Cannot create role."
    exit 1
}

if (-not $SYSTEM_RS_ID) {
    Log-Error "System resource server ID is not available. Cannot create role."
    exit 1
}

$roleData = @{
    name = "Administrator"
    description = "System administrator role with full permissions"
    ouId = $DEFAULT_OU_ID
    permissions = @(
        @{
            resourceServerId = $SYSTEM_RS_ID
            permissions = @("system")
        }
    )
    assignments = @(
        @{
            id = $ADMIN_USER_ID
            type = "user"
        }
    )
} | ConvertTo-Json -Depth 10

$response = Invoke-ThunderApi -Method POST -Endpoint "/roles" -Data $roleData

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "Admin role created and assigned to admin user"
    $body = $response.Body | ConvertFrom-Json
    $ADMIN_ROLE_ID = $body.id
    if ($ADMIN_ROLE_ID) {
        Log-Info "Admin role ID: $ADMIN_ROLE_ID"
    }
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "Admin role already exists"
}
else {
    Log-Error "Failed to create admin role (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

Write-Host ""

# ============================================================================
# Create DEVELOP Application
# ============================================================================

Log-Info "Creating DEVELOP application..."

$appData = @{
    name = "Develop"
    description = "Developer application for Thunder"
    url = "$env:THUNDER_API_BASE/develop"
    logo_url = "$env:THUNDER_API_BASE/develop/assets/images/trifacta.svg"
    auth_flow_graph_id = "auth_flow_config_basic"
    registration_flow_graph_id = "registration_flow_config_basic"
    is_registration_flow_enabled = $true
    user_attributes = @("given_name", "family_name", "email", "groups", "name")
    inbound_auth_config = @(
        @{
            type = "oauth2"
            config = @{
                client_id = "DEVELOP"
                redirect_uris = @("$env:THUNDER_API_BASE/develop")
                grant_types = @("authorization_code")
                response_types = @("code")
                pkce_required = $true
                token_endpoint_auth_method = "none"
                public_client = $true
                token = @{
                    issuer = "$env:THUNDER_API_BASE/oauth2/token"
                    access_token = @{
                        validity_period = 3600
                        user_attributes = @("given_name", "family_name", "email", "groups", "name")
                    }
                    id_token = @{
                        validity_period = 3600
                        user_attributes = @("given_name", "family_name", "email", "groups", "name")
                        scope_claims = @{
                            profile = @("name", "given_name", "family_name", "picture")
                            email = @("email", "email_verified")
                            phone = @("phone_number", "phone_number_verified")
                            group = @("groups")
                        }
                    }
                }
            }
        }
    )
} | ConvertTo-Json -Depth 10

$response = Invoke-ThunderApi -Method POST -Endpoint "/applications" -Data $appData

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "DEVELOP application created successfully"
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "DEVELOP application already exists, skipping"
}
elseif ($response.StatusCode -eq 400 -and ($response.Body -match "Application already exists|APP-1022")) {
    Log-Warning "DEVELOP application already exists, skipping"
}
else {
    Log-Error "Failed to create DEVELOP application (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

Write-Host ""

# ============================================================================
# Summary
# ============================================================================

Log-Success "Default resources setup completed successfully!"
Write-Host ""
Log-Info "👤 Admin credentials:"
Log-Info "   Username: admin"
Log-Info "   Password: admin"
Log-Info "   Role: Administrator (system permission)"
Write-Host ""
