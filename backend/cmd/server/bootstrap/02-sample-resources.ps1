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

# Bootstrap Script: Sample Resources Setup
# Creates resources required to run the Thunder sample experience

$ErrorActionPreference = 'Stop'

# Dot-source common functions from the same directory as this script
. "$PSScriptRoot/common.ps1"

Log-Info "Creating sample Thunder resources..."
Write-Host ""

# ============================================================================
# Create Customers Organization Unit
# ============================================================================

$customerOuHandle = "customers"

Log-Info "Creating Customers organization unit..."

$customerOuData = @{
    handle = $customerOuHandle
    name = "Customers"
    description = "Organization unit for customer accounts"
} | ConvertTo-Json -Depth 5

$response = Invoke-ThunderApi -Method POST -Endpoint "/organization-units" -Data $customerOuData

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "Customers organization unit created successfully"
    $body = $response.Body | ConvertFrom-Json
    $CUSTOMER_OU_ID = $body.id
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "Customers organization unit already exists, retrieving ID..."
    # Get existing OU ID by handle to ensure we get the correct "customers" OU
    $response = Invoke-ThunderApi -Method GET -Endpoint "/organization-units/tree/$customerOuHandle"
    if ($response.StatusCode -eq 200) {
        $body = $response.Body | ConvertFrom-Json
        $CUSTOMER_OU_ID = $body.id
    }
    else {
        Log-Error "Failed to fetch organization unit by handle '$customerOuHandle' (HTTP $($response.StatusCode))"
        Write-Host "Response: $($response.Body)"
        exit 1
    }
}
else {
    Log-Error "Failed to create Customers organization unit (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

if (-not $CUSTOMER_OU_ID) {
    Log-Error "Could not determine Customers organization unit ID"
    exit 1
}

Log-Info "Customers OU ID: $CUSTOMER_OU_ID"

Write-Host ""

# ============================================================================
# Create Customer User Type
# ============================================================================

Log-Info "Creating Customer user type..."

$customerUserTypeData = ([ordered]@{
    name = "Customer"
    ouId = $CUSTOMER_OU_ID
    allowSelfRegistration = $true
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
        given_name = @{
            type = "string"
            required = $false
        }
        family_name = @{
            type = "string"
            required = $false
        }
    }
} | ConvertTo-Json -Depth 5)

$response = Invoke-ThunderApi -Method POST -Endpoint "/user-types" -Data $customerUserTypeData

if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
    Log-Success "Customer user type created successfully"
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "Customer user type already exists, skipping"
}
else {
    Log-Error "Failed to create Customer user type (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

Write-Host ""

# ============================================================================
# Create Sample Application
# ============================================================================

Log-Info "Creating Sample App application..."

$appData = @{
    name = "Sample App"
    description = "Sample application for testing"
    url = "https://localhost:3000"
    logo_url = "https://localhost:3000/logo.png"
    tos_uri = "https://localhost:3000/terms"
    policy_uri = "https://localhost:3000/privacy"
    contacts = @("admin@example.com", "support@example.com")
    is_registration_flow_enabled = $true
    user_attributes = @("given_name","family_name","email","groups")
    allowed_user_types = @("Customer")
    inbound_auth_config = @(
        @{
            type = "oauth2"
            config = @{
                client_id = "sample_app_client"
                redirect_uris = @("https://localhost:3000")
                grant_types = @("authorization_code")
                response_types = @("code")
                token_endpoint_auth_method = "none"
                pkce_required = $true
                public_client = $true
                scopes = @("openid", "profile", "email")
                token = @{
                    issuer = "thunder"
                    access_token = @{
                        validity_period = 3600
                        user_attributes = @("given_name","family_name","email","groups")
                    }
                    id_token = @{
                        validity_period = 3600
                        user_attributes = @("given_name","family_name","email","groups")
                        scope_claims = @{
                            profile = @("name","given_name","family_name","picture")
                            email = @("email","email_verified")
                            phone = @("phone_number","phone_number_verified")
                            group = @("groups")
                        }
                    }
                }
            }
        }
    )
} | ConvertTo-Json -Depth 15

$response = Invoke-ThunderApi -Method POST -Endpoint "/applications" -Data $appData

if ($response.StatusCode -in 200, 201, 202) {
    Log-Success "Sample App created successfully"
    $body = $response.Body | ConvertFrom-Json
    $sampleAppId = $body.id
    if ($sampleAppId) {
        Log-Info "Sample App ID: $sampleAppId"
    }
    else {
        Log-Warning "Could not extract Sample App ID from response"
    }
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "Sample App already exists, skipping"
}
elseif ($response.StatusCode -eq 400 -and ($response.Body -match "Application already exists|APP-1022")) {
    Log-Warning "Sample App already exists, skipping"
}
else {
    Log-Error "Failed to create Sample App (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

Write-Host ""

# ============================================================================
# Create React SDK Sample Application
# ============================================================================

Log-Info "Creating React SDK Sample App application..."

$reactSdkAppData = @{
    name = "React SDK Sample"
    description = "Sample React application using Thunder React SDK"
    client_id = "REACT_SDK_SAMPLE"
    url = "https://localhost:3000"
    logo_url = "https://localhost:3000/logo.png"
    tos_uri = "https://localhost:3000/terms"
    policy_uri = "https://localhost:3000/privacy"
    contacts = @("admin@example.com")
    is_registration_flow_enabled = $true
    assertion = @{
        issuer = "thunder"
        validity_period = 3600
        user_attributes = $null
    }
    certificate = @{
        type = "NONE"
        value = ""
    }
    user_attributes = @("given_name","family_name","email","groups","name")
    allowed_user_types = @("Customer")
    inbound_auth_config = @(
        @{
            type = "oauth2"
            config = @{
                client_id = "REACT_SDK_SAMPLE"
                redirect_uris = @("https://localhost:3000")
                grant_types = @("authorization_code")
                response_types = @("code")
                token_endpoint_auth_method = "none"
                pkce_required = $true
                public_client = $true
                token = @{
                    issuer = "https://localhost:8090"
                    access_token = @{
                        validity_period = 3600
                        user_attributes = @("given_name","family_name","email","groups","name")
                    }
                    id_token = @{
                        validity_period = 3600
                        user_attributes = @("given_name","family_name","email","groups","name")
                        scope_claims = @{
                            email = @("email","email_verified")
                            group = @("groups")
                            phone = @("phone_number","phone_number_verified")
                            profile = @("name","given_name","family_name","picture")
                        }
                    }
                }
            }
        }
    )
} | ConvertTo-Json -Depth 15

$response = Invoke-ThunderApi -Method POST -Endpoint "/applications" -Data $reactSdkAppData

if ($response.StatusCode -in 200, 201, 202) {
    Log-Success "React SDK Sample App created successfully"
    $body = $response.Body | ConvertFrom-Json
    $reactSdkAppId = $body.id
    if ($reactSdkAppId) {
        Log-Info "React SDK Sample App ID: $reactSdkAppId"
    }
    else {
        Log-Warning "Could not extract React SDK Sample App ID from response"
    }
}
elseif ($response.StatusCode -eq 409) {
    Log-Warning "React SDK Sample App already exists, skipping"
}
elseif ($response.StatusCode -eq 400 -and ($response.Body -match "Application already exists|APP-1022")) {
    Log-Warning "React SDK Sample App already exists, skipping"
}
else {
    Log-Error "Failed to create React SDK Sample App (HTTP $($response.StatusCode))"
    Write-Host "Response: $($response.Body)"
    exit 1
}

Write-Host ""

# ============================================================================
# Summary
# ============================================================================

Log-Success "Sample resources setup completed successfully!"
Write-Host ""
