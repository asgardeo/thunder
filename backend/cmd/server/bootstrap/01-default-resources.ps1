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

# Parse command line arguments for custom redirect URIs
param(
    [string]$DevelopRedirectUris = ""
)

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

# Dot-source common functions from the same directory as this script
. "$PSScriptRoot/common.ps1"

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
    # Get existing OU ID by handle to ensure we get the correct "default" OU
    $response = Invoke-ThunderApi -Method GET -Endpoint "/organization-units/tree/default"

    if ($response.StatusCode -eq 200) {
        $body = $response.Body | ConvertFrom-Json
        $DEFAULT_OU_ID = $body.id
        if ($DEFAULT_OU_ID) {
            Log-Success "Found OU ID: $DEFAULT_OU_ID"
        }
        else {
            Log-Error "Could not find OU ID in response"
            exit 1
        }
    }
    else {
        Log-Error "Failed to fetch organization unit by handle 'default' (HTTP $($response.StatusCode))"
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
        password = @{
            type = "string"
            required = $true
            credential = $true
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
# Create Default Flows
# ============================================================================

Log-Info "Creating default flows..."

# Path to flow definitions directories
$AUTH_FLOWS_DIR = Join-Path $PSScriptRoot "flows" "authentication"
$REG_FLOWS_DIR = Join-Path $PSScriptRoot "flows" "registration"
$USER_ONBOARDING_FLOWS_DIR = Join-Path $PSScriptRoot "flows" "user_onboarding"

# Check if flows directories exist
if (-not (Test-Path $AUTH_FLOWS_DIR) -and -not (Test-Path $REG_FLOWS_DIR) -and -not (Test-Path $USER_ONBOARDING_FLOWS_DIR)) {
    Log-Warning "Flow definitions directories not found, skipping flow creation"
}
else {
    $flowCount = 0
    $flowSuccess = 0
    $flowSkipped = 0

    # Process authentication flows
    if (Test-Path $AUTH_FLOWS_DIR) {
        $authFlowFiles = Get-ChildItem -Path $AUTH_FLOWS_DIR -Filter "*.json" -File -ErrorAction SilentlyContinue
        
        if ($authFlowFiles.Count -gt 0) {
            Log-Info "Processing authentication flows..."
            
            # Fetch existing auth flows
            $listResponse = Invoke-ThunderApi -Method GET -Endpoint "/flows?flowType=AUTHENTICATION&limit=200"
            
            # Store existing auth flows by handle in a hashtable
            $existingAuthFlows = @{}
            if ($listResponse.StatusCode -eq 200) {
                $listBody = $listResponse.Body | ConvertFrom-Json
                foreach ($flow in $listBody.flows) {
                    $existingAuthFlows[$flow.handle] = $flow.id
                }
            }
            
            foreach ($flowFile in $authFlowFiles) {
                $flowCount++
                
                # Get flow handle and name from file
                $flowContent = Get-Content -Path $flowFile.FullName -Raw | ConvertFrom-Json
                $flowHandle = $flowContent.handle
                $flowName = $flowContent.name
                
                # Check if flow exists by handle
                if ($existingAuthFlows.ContainsKey($flowHandle)) {
                    # Update existing flow
                    $flowId = $existingAuthFlows[$flowHandle]
                    Log-Info "Updating existing auth flow: $flowName (handle: $flowHandle)"
                    $result = Update-Flow -FlowId $flowId -FlowFilePath $flowFile.FullName
                    if ($result) {
                        $flowSuccess++
                    }
                }
                else {
                    # Create new flow
                    $flowId = Create-Flow -FlowFilePath $flowFile.FullName
                    if ($flowId) {
                        $flowSuccess++
                    }
                    elseif ($flowId -eq "") {
                        $flowSkipped++
                    }
                }
            }
        }
        else {
            Log-Info "No authentication flow files found"
        }
    }

    # Process registration flows
    if (Test-Path $REG_FLOWS_DIR) {
        $regFlowFiles = Get-ChildItem -Path $REG_FLOWS_DIR -Filter "*.json" -File -ErrorAction SilentlyContinue
        
        if ($regFlowFiles.Count -gt 0) {
            Log-Info "Processing registration flows..."
            
            # Fetch existing registration flows
            $listResponse = Invoke-ThunderApi -Method GET -Endpoint "/flows?flowType=REGISTRATION&limit=200"
            
            # Store existing registration flows by handle in a hashtable
            $existingRegFlows = @{}
            if ($listResponse.StatusCode -eq 200) {
                $listBody = $listResponse.Body | ConvertFrom-Json
                foreach ($flow in $listBody.flows) {
                    $existingRegFlows[$flow.handle] = $flow.id
                }
            }

            foreach ($flowFile in $regFlowFiles) {
                $flowCount++
                
                # Get flow handle and name from file
                $flowContent = Get-Content -Path $flowFile.FullName -Raw | ConvertFrom-Json
                $flowHandle = $flowContent.handle
                $flowName = $flowContent.name
                
                # Check if flow exists by handle
                if ($existingRegFlows.ContainsKey($flowHandle)) {
                    # Update existing flow
                    $flowId = $existingRegFlows[$flowHandle]
                    Log-Info "Updating existing registration flow: $flowName (handle: $flowHandle)"
                    $result = Update-Flow -FlowId $flowId -FlowFilePath $flowFile.FullName
                    if ($result) {
                        $flowSuccess++
                    }
                }
                else {
                    # Create new flow
                    $flowId = Create-Flow -FlowFilePath $flowFile.FullName
                    if ($flowId) {
                        $flowSuccess++
                    }
                    elseif ($flowId -eq "") {
                        $flowSkipped++
                    }
                }
            }
        }
        else {
            Log-Info "No registration flow files found"
        }
    }

    # Process user onboarding flows
    if (Test-Path $USER_ONBOARDING_FLOWS_DIR) {
        $onboardingFlowFiles = Get-ChildItem -Path $USER_ONBOARDING_FLOWS_DIR -Filter "*.json" -File -ErrorAction SilentlyContinue
        
        if ($onboardingFlowFiles.Count -gt 0) {
            Log-Info "Processing user onboarding flows..."
            
            # Fetch existing user onboarding flows
            $listResponse = Invoke-ThunderApi -Method GET -Endpoint "/flows?flowType=USER_ONBOARDING&limit=200"
            
            # Store existing onboarding flows by handle in a hashtable
            $existingOnboardingFlows = @{}
            if ($listResponse.StatusCode -eq 200) {
                $listBody = $listResponse.Body | ConvertFrom-Json
                foreach ($flow in $listBody.flows) {
                    $existingOnboardingFlows[$flow.handle] = $flow.id
                }
            }
            
            foreach ($flowFile in $onboardingFlowFiles) {
                $flowCount++
                
                # Get flow handle and name from file
                $flowContent = Get-Content -Path $flowFile.FullName -Raw | ConvertFrom-Json
                $flowHandle = $flowContent.handle
                $flowName = $flowContent.name
                
                # Check if flow exists by handle
                if ($existingOnboardingFlows.ContainsKey($flowHandle)) {
                    # Update existing flow
                    $flowId = $existingOnboardingFlows[$flowHandle]
                    Log-Info "Updating existing user onboarding flow: $flowName (handle: $flowHandle)"
                    $result = Update-Flow -FlowId $flowId -FlowFilePath $flowFile.FullName
                    if ($result) {
                        $flowSuccess++
                    }
                }
                else {
                    # Create new flow
                    $flowId = Create-Flow -FlowFilePath $flowFile.FullName
                    if ($flowId) {
                        $flowSuccess++
                    }
                    elseif ($flowId -eq "") {
                        $flowSkipped++
                    }
                }
            }
        }
        else {
            Log-Info "No user onboarding flow files found"
        }
    }

    if ($flowCount -gt 0) {
        Log-Info "Flow creation summary: $flowSuccess created/updated, $flowSkipped skipped, $($flowCount - $flowSuccess - $flowSkipped) failed"
    }
}

Write-Host ""

# ============================================================================
# Create Application-Specific Flows
# ============================================================================

Log-Info "Creating application-specific flows..."

$APPS_FLOWS_DIR = Join-Path $PSScriptRoot "flows" "apps"

# Store application flow IDs in a hashtable
$APP_FLOW_IDS = @{}

if (Test-Path $APPS_FLOWS_DIR) {
    # Fetch all existing flows once
    Log-Info "Fetching existing flows for application flow processing..."
    
    # Get auth flows
    $authResponse = Invoke-ThunderApi -Method GET -Endpoint "/flows?flowType=AUTHENTICATION&limit=200"
    $existingAppAuthFlows = @{}
    if ($authResponse.StatusCode -eq 200) {
        $authBody = $authResponse.Body | ConvertFrom-Json
        foreach ($flow in $authBody.flows) {
            $existingAppAuthFlows[$flow.handle] = $flow.id
        }
    }
    
    # Get registration flows
    $regResponse = Invoke-ThunderApi -Method GET -Endpoint "/flows?flowType=REGISTRATION&limit=200"
    $existingAppRegFlows = @{}
    if ($regResponse.StatusCode -eq 200) {
        $regBody = $regResponse.Body | ConvertFrom-Json
        foreach ($flow in $regBody.flows) {
            $existingAppRegFlows[$flow.handle] = $flow.id
        }
    }

    $appDirs = Get-ChildItem -Path $APPS_FLOWS_DIR -Directory -ErrorAction SilentlyContinue
    
    foreach ($appDir in $appDirs) {
        $appName = $appDir.Name
        $appAuthFlowId = ""
        $appRegFlowId = ""
        
        Log-Info "Processing flows for application: $appName"
        
        # Process authentication flow for app
        $authFlowFiles = Get-ChildItem -Path $appDir.FullName -Filter "auth_*.json" -File -ErrorAction SilentlyContinue
        
        if ($authFlowFiles.Count -gt 0) {
            $authFlowFile = $authFlowFiles[0]
            $flowContent = Get-Content -Path $authFlowFile.FullName -Raw | ConvertFrom-Json
            $flowHandle = $flowContent.handle
            $flowName = $flowContent.name
            
            # Check if auth flow exists by handle
            if ($existingAppAuthFlows.ContainsKey($flowHandle)) {
                # Update existing flow
                $appAuthFlowId = $existingAppAuthFlows[$flowHandle]
                Log-Info "Updating existing auth flow: $flowName (handle: $flowHandle)"
                Update-Flow -FlowId $appAuthFlowId -FlowFilePath $authFlowFile.FullName
            }
            else {
                # Create new flow
                $appAuthFlowId = Create-Flow -FlowFilePath $authFlowFile.FullName
            }
            
            # Re-fetch registration flows after creating auth flow
            if ($appAuthFlowId) {
                $response = Invoke-ThunderApi -Method GET -Endpoint "/flows?flowType=REGISTRATION&limit=200"
                if ($response.StatusCode -eq 200) {
                    $existingAppRegFlows = @{}
                    $flows = ($response.Body | ConvertFrom-Json).flows
                    foreach ($flow in $flows) {
                        $existingAppRegFlows[$flow.handle] = $flow.id
                    }
                }
            }
        }
        else {
            Log-Warning "No authentication flow file found for app: $appName"
        }

        # Process registration flow for app
        $regFlowFiles = Get-ChildItem -Path $appDir.FullName -Filter "registration_*.json" -File -ErrorAction SilentlyContinue
        
        if ($regFlowFiles.Count -gt 0) {
            $regFlowFile = $regFlowFiles[0]
            $flowContent = Get-Content -Path $regFlowFile.FullName -Raw | ConvertFrom-Json
            $flowHandle = $flowContent.handle
            $flowName = $flowContent.name
            
            # Check if registration flow exists by handle
            if ($existingAppRegFlows.ContainsKey($flowHandle)) {
                # Update existing flow
                $appRegFlowId = $existingAppRegFlows[$flowHandle]
                Log-Info "Updating existing registration flow: $flowName (handle: $flowHandle)"
                Update-Flow -FlowId $appRegFlowId -FlowFilePath $regFlowFile.FullName
            }
            else {
                # Create new flow
                $appRegFlowId = Create-Flow -FlowFilePath $regFlowFile.FullName
            }
        }
        else {
            Log-Warning "No registration flow file found for app: $appName"
        }
        
        # Store the flow IDs for this app
        $APP_FLOW_IDS[$appName] = @{
            authFlowId = $appAuthFlowId
            regFlowId = $appRegFlowId
        }
    }
}
else {
    Log-Warning "Application flows directory not found at $APPS_FLOWS_DIR"
}

Write-Host ""

# ============================================================================
# Create DEVELOP Application
# ============================================================================

Log-Info "Creating DEVELOP application..."

# Get flow IDs for develop app from the APP_FLOW_IDS created/found during flow processing
$DEVELOP_AUTH_FLOW_ID = ""
$DEVELOP_REG_FLOW_ID = ""

if ($APP_FLOW_IDS.ContainsKey("develop")) {
    $DEVELOP_AUTH_FLOW_ID = $APP_FLOW_IDS["develop"].authFlowId
    $DEVELOP_REG_FLOW_ID = $APP_FLOW_IDS["develop"].regFlowId
}

# Validate that flow IDs are available
if (-not $DEVELOP_AUTH_FLOW_ID) {
    Log-Error "Develop authentication flow ID not found, cannot create DEVELOP application"
    Log-Error "Make sure flows/apps/develop/auth_flow_develop.json exists"
    exit 1
}
if (-not $DEVELOP_REG_FLOW_ID) {
    Log-Error "Develop registration flow ID not found, cannot create DEVELOP application"
    Log-Error "Make sure flows/apps/develop/registration_flow_develop.json exists"
    exit 1
}

# Use THUNDER_PUBLIC_URL for redirect URIs, fallback to THUNDER_API_BASE if not set
$PUBLIC_URL = if ($env:THUNDER_PUBLIC_URL) { $env:THUNDER_PUBLIC_URL } else { $env:THUNDER_API_BASE }

# Build redirect URIs array - default + custom if provided
$redirectUrisList = @("$PUBLIC_URL/develop")
if ($DevelopRedirectUris) {
    Log-Info "Adding custom redirect URIs: $DevelopRedirectUris"
    # Split comma-separated URIs and append to array
    $customUris = $DevelopRedirectUris -split ',' | ForEach-Object { $_.Trim() }
    $redirectUrisList += $customUris
}

$appData = @{
    name = "Develop"
    description = "Developer application for Thunder"
    url = "$PUBLIC_URL/develop"
    logo_url = "$PUBLIC_URL/develop/assets/images/logo-mini.svg"
    auth_flow_id = $DEVELOP_AUTH_FLOW_ID
    registration_flow_id = $DEVELOP_REG_FLOW_ID
    is_registration_flow_enabled = $false
    allowed_user_types = @("Person")
    user_attributes = @("given_name", "family_name", "email", "groups", "name")
    inbound_auth_config = @(
        @{
            type = "oauth2"
            config = @{
                client_id = "DEVELOP"
                redirect_uris = $redirectUrisList
                grant_types = @("authorization_code")
                response_types = @("code")
                pkce_required = $true
                token_endpoint_auth_method = "none"
                public_client = $true
                token = @{
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
# Create Themes
# ============================================================================

Log-Info "Creating themes..."

# Get the script directory to locate theme files
$themesDir = Join-Path $PSScriptRoot "themes"

# Check if themes directory exists
if (-not (Test-Path $themesDir)) {
    Log-Warning "Themes directory not found at $themesDir, skipping theme creation"
}
else {
    $themeFiles = Get-ChildItem -Path $themesDir -Filter "*.json" -File -ErrorAction SilentlyContinue
    
    if ($themeFiles.Count -gt 0) {
        Log-Info "Processing themes from $themesDir..."
        
        $themeCount = 0
        $themeSuccess = 0
        $themeSkipped = 0
        
        foreach ($themeFile in $themeFiles) {
            $themeCount++
            
            # Get theme name from file content
            $themeContent = Get-Content -Path $themeFile.FullName -Raw | ConvertFrom-Json
            $themeName = if ($themeContent.displayName) { $themeContent.displayName } else { $themeFile.BaseName }
            
            Log-Info "Creating theme: $themeName (from $($themeFile.Name))"
            $themePayload = Get-Content $themeFile.FullName -Raw
            
            $response = Invoke-ThunderApi -Method POST -Endpoint "/design/themes" -Data $themePayload
            
            if ($response.StatusCode -in 200, 201) {
                Log-Success "Theme '$themeName' created successfully"
                $body = $response.Body | ConvertFrom-Json
                $themeId = $body.id
                if ($themeId) {
                    Log-Info "Theme ID: $themeId"
                }
                $themeSuccess++
            }
            elseif ($response.StatusCode -eq 409) {
                Log-Warning "Theme '$themeName' already exists, skipping"
                $themeSkipped++
            }
            else {
                Log-Error "Failed to create theme '$themeName' (HTTP $($response.StatusCode))"
                Write-Host "Response: $($response.Body)"
                exit 1
            }
        }
        
        Write-Host ""
        Log-Info "Theme creation summary: $themeSuccess created, $themeSkipped skipped (Total: $themeCount)"
    }
    else {
        Log-Warning "No theme files found in $themesDir"
    }
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
