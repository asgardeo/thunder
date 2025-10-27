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

# Script to set up initial admin user and DEVELOP application

[CmdletBinding()]
param(
    [Parameter()]
    [int]$Port,
    
    [Parameter()]
    [string]$Config,
    
    [Parameter()]
    [int]$Timeout = 30,
    
    [Parameter()]
    [switch]$Help
)

# Default settings
$BACKEND_PORT = if ($Port) { $Port } elseif ($env:BACKEND_PORT) { [int]$env:BACKEND_PORT } else { 8090 }
$RETRY_DELAY = 2
$CONFIG_FILE = $Config

# Default config file paths to check (relative to script location)
$DEFAULT_CONFIG_PATHS = @(
    ".\repository\conf\deployment.yaml"
)

$ErrorActionPreference = "Stop"

function Show-Help {
    Write-Host ""
    Write-Host "Usage:"
    Write-Host "  .\setup_initial_data.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Port <port>         Thunder server port (default: 8090 or from config)"
    Write-Host "  -Config <path>       Path to Thunder configuration file"
    Write-Host "  -Timeout <seconds>   Timeout for server readiness check (default: 30 seconds)"
    Write-Host "  -Help                Show this help message and exit"
    Write-Host ""
    Write-Host "Environment Variables:"
    Write-Host "  BACKEND_PORT         Thunder server port (can be overridden by -Port)"
    Write-Host ""
    Write-Host "Configuration:"
    Write-Host "  The script attempts to auto-detect server settings from:"
    Write-Host "  - Specified config file (-Config)"
    Write-Host "  - Default locations: ..\cmd\server\repository\conf\deployment.yaml"
    Write-Host "  - Command line arguments (-Port)"
    Write-Host "  - Environment variables (BACKEND_PORT)"
    Write-Host ""
}

if ($Help) {
    Show-Help
    exit 0
}

# Function to find and read the Thunder configuration file
function Read-ThunderConfig {
    $configFile = $CONFIG_FILE
    
    # If no config file specified, try to find it
    if ([string]::IsNullOrEmpty($configFile)) {
        $scriptDir = Split-Path -Parent $MyInvocation.ScriptName
        foreach ($path in $DEFAULT_CONFIG_PATHS) {
            $fullPath = Join-Path $scriptDir $path
            if (Test-Path $fullPath) {
                $configFile = $fullPath
                break
            }
        }
    }
    
    if ([string]::IsNullOrEmpty($configFile) -or -not (Test-Path $configFile)) {
        Write-Warning "Thunder configuration file not found. Using default settings."
        return $null
    }
    
    Write-Info "Reading configuration from: $configFile"
    
    # Parse YAML using PowerShell-Yaml if available, otherwise fall back to basic parsing
    if (Get-Module -ListAvailable -Name powershell-yaml) {
        try {
            Import-Module powershell-yaml -ErrorAction Stop
            $config = Get-Content $configFile -Raw | ConvertFrom-Yaml
            
            $hostname = if ($config.server.hostname) { $config.server.hostname } else { "localhost" }
            $port = if ($config.server.port) { $config.server.port } else { 8090 }
            $httpOnly = if ($config.server.http_only) { $config.server.http_only } else { $false }
            
            # Override port if provided via command line or environment
            if ($BACKEND_PORT -ne 8090) {
                $port = $BACKEND_PORT
            }
            
            # Determine protocol
            $protocol = if ($httpOnly) { "http" } else { "https" }
            
            return "${protocol}://${hostname}:${port}"
        }
        catch {
            Write-Warning "Failed to parse YAML with PowerShell-Yaml: $_"
        }
    }
    
    Write-Warning "PowerShell-Yaml module not found. Using basic YAML parsing."
    # Basic fallback parsing using regex
    try {
        $content = Get-Content $configFile -Raw
        
        $hostname = "localhost"
        $port = 8090
        $httpOnly = $false
        
        # Extract hostname
        if ($content -match 'hostname:\s*([^\s\r\n]+)') {
            $hostname = $matches[1].Trim('"', "'")
        }
        
        # Extract port
        if ($content -match 'port:\s*(\d+)') {
            $port = [int]$matches[1]
        }
        
        # Check for http_only
        if ($content -match 'http_only:\s*(true|false)') {
            $httpOnly = $matches[1] -eq "true"
        }
        
        # Override port if provided via command line or environment
        if ($BACKEND_PORT -ne 8090) {
            $port = $BACKEND_PORT
        }
        
        # Determine protocol
        $protocol = if ($httpOnly) { "http" } else { "https" }
        
        return "${protocol}://${hostname}:${port}"
    }
    catch {
        Write-Warning "Failed to parse configuration file: $_"
        return $null
    }
}

# Function to construct base URL
function Get-BaseUrl {
    $configUrl = Read-ThunderConfig
    
    if ($configUrl) {
        return $configUrl
    }
    else {
        # Fallback to environment/command line settings
        return "https://localhost:${BACKEND_PORT}"
    }
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ️  $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

function Wait-ForServer {
    param([string]$BaseUrl)
    
    Write-Info "Waiting for Thunder server to be ready at ${BaseUrl}..."
    
    # Try multiple health endpoints
    $healthEndpoints = @("/health/readiness", "/health/liveness", "/healthcheck")
    
    $elapsed = 0
    while ($elapsed -lt $Timeout) {
        foreach ($endpoint in $healthEndpoints) {
            try {
                $response = Invoke-WebRequest -Uri "${BaseUrl}${endpoint}" -Method Get -TimeoutSec 5 -SkipCertificateCheck -ErrorAction Stop
                if ($response.StatusCode -eq 200) {
                    Write-Success "Server is ready! (responded to ${endpoint})"
                    return $true
                }
            }
            catch {
                # Continue trying other endpoints
            }
        }
        
        Start-Sleep -Seconds $RETRY_DELAY
        $elapsed += $RETRY_DELAY
        Write-Host "." -NoNewline
    }
    
    Write-Host ""
    Write-Error "Server is not ready after ${Timeout} seconds"
    Write-Info "Tried endpoints: $($healthEndpoints -join ', ')"
    Write-Info "Make sure Thunder server is running at ${BaseUrl}"
    exit 1
}

function New-UserSchema {
    param([string]$BaseUrl)
    
    Write-Info "Creating Default user schema..."
    
    $body = @{
        name = "default-user-schema"
        schema = @{
            sub = @{
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
            name = @{
                type = "string"
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
            picture = @{
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
    } | ConvertTo-Json -Depth 10
    
    try {
        $response = Invoke-WebRequest -Uri "${BaseUrl}/user-schemas" -Method Post -Body $body -ContentType "application/json" -SkipCertificateCheck -ErrorAction Stop
        
        if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
            Write-Success "User schema created successfully"
            return $true
        }
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 409) {
            Write-Warning "User schema already exists, skipping creation"
            return $true
        }
        else {
            Write-Error "Failed to create user schema. HTTP status: $statusCode"
            Write-Host "Response: $($_.Exception.Message)"
            return $false
        }
    }
    
    return $false
}

function New-AdminUser {
    param([string]$BaseUrl)
    
    Write-Info "Creating admin user..."
    
    $body = @{
        type = "default-user-schema"
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
    } | ConvertTo-Json -Depth 10
    
    try {
        $response = Invoke-WebRequest -Uri "${BaseUrl}/users" -Method Post -Body $body -ContentType "application/json" -SkipCertificateCheck -ErrorAction Stop
        
        if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
            Write-Success "Admin user created successfully"
            Write-Info "Username: admin"
            Write-Info "Password: admin"
            return $true
        }
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 409) {
            Write-Warning "Admin user already exists, skipping creation"
            return $true
        }
        else {
            Write-Error "Failed to create admin user. HTTP status: $statusCode"
            Write-Host "Response: $($_.Exception.Message)"
            return $false
        }
    }
    
    return $false
}

function New-DevelopApp {
    param([string]$BaseUrl)
    
    Write-Info "Creating DEVELOP application..."
    
    $body = @{
        name = "Develop"
        description = "Developer application for Thunder"
        url = "${BaseUrl}/develop"
        logo_url = "${BaseUrl}/develop/assets/images/asgardeo-trifacta.svg"
        auth_flow_graph_id = "auth_flow_config_basic"
        registration_flow_graph_id = "registration_flow_config_basic"
        is_registration_flow_enabled = $true
        user_attributes = @("given_name", "family_name", "email", "groups", "name")
        inbound_auth_config = @(
            @{
                type = "oauth2"
                config = @{
                    client_id = "DEVELOP"
                    redirect_uris = @("https://localhost:5191")
                    grant_types = @("authorization_code")
                    response_types = @("code")
                    pkce_required = $false
                    token_endpoint_auth_method = "none"
                    public_client = $true
                    token = @{
                        issuer = "${BaseUrl}/oauth2/token"
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
    
    try {
        $response = Invoke-WebRequest -Uri "${BaseUrl}/applications" -Method Post -Body $body -ContentType "application/json" -SkipCertificateCheck -ErrorAction Stop
        
        if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
            Write-Success "DEVELOP application created successfully"
            Write-Info "Application URL: ${BaseUrl}/develop"
            Write-Info "Client ID: DEVELOP"
            return $true
        }
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 409) {
            Write-Warning "DEVELOP application already exists, skipping creation"
            return $true
        }
        else {
            Write-Error "Failed to create DEVELOP application. HTTP status: $statusCode"
            Write-Host "Response: $($_.Exception.Message)"
            return $false
        }
    }
    
    return $false
}

function Main {
    Write-Host "🚀 Thunder Initial Data Setup Script"
    Write-Host "===================================="
    Write-Host ""
    
    # Construct base URL dynamically from configuration
    $BASE_URL = Get-BaseUrl
    
    Write-Info "Using Thunder server at: ${BASE_URL}"
    
    # Show configuration source
    if ($BACKEND_PORT -ne 8090) {
        Write-Info "Port override detected: ${BACKEND_PORT}"
    }
    
    Write-Host ""
    
    # Wait for server to be ready
    Wait-ForServer -BaseUrl $BASE_URL
    Write-Host ""
    
    # Create user schema
    $success = New-UserSchema -BaseUrl $BASE_URL
    if (-not $success) {
        Write-Error "Failed to create user schema. Aborting."
        exit 1
    }
    Write-Host ""
    
    # Create admin user
    $success = New-AdminUser -BaseUrl $BASE_URL
    if (-not $success) {
        Write-Error "Failed to create admin user. Aborting."
        exit 1
    }
    Write-Host ""
    
    # Create DEVELOP application
    $success = New-DevelopApp -BaseUrl $BASE_URL
    if (-not $success) {
        Write-Error "Failed to create DEVELOP application. Aborting."
        exit 1
    }
    Write-Host ""
    
    Write-Success "Initial data setup completed successfully!"
    Write-Host ""
    Write-Host "📱 You can now access:"
    Write-Host "   🚪 Gate (Login/Register): ${BASE_URL}/signin"
    Write-Host "   🛠️  Develop (Admin Console): ${BASE_URL}/develop"
    Write-Host ""
    Write-Host "👤 Admin credentials:"
    Write-Host "   Username: admin"
    Write-Host "   Password: admin"
    Write-Host ""
}

# Run main function
Main
