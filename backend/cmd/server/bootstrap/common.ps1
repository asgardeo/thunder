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

# Common functions for bootstrap scripts
# Dot-source this file at the beginning of each bootstrap script

# Logging Functions
function Log-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Log-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] ✓ $Message" -ForegroundColor Green
}

function Log-Warning {
    param([string]$Message)
    Write-Host "[WARNING] ⚠ $Message" -ForegroundColor Yellow
}

function Log-Error {
    param([string]$Message)
    Write-Host "[ERROR] ✗ $Message" -ForegroundColor Red
}

function Log-Debug {
    param([string]$Message)
    if ($env:DEBUG -eq "true") {
        Write-Host "[DEBUG] $Message" -ForegroundColor Cyan
    }
}

# API Call Helper Function
function Invoke-ThunderApi {
    param(
        [Parameter(Mandatory=$true)]
        [string]$Method,
        [Parameter(Mandatory=$true)]
        [string]$Endpoint,
        [Parameter(Mandatory=$false)]
        [string]$Data = $null
    )

    $url = "$($env:THUNDER_API_BASE)$Endpoint"
    
    Log-Debug "API Call: $Method $url"

    try {
        $headers = @{
            "Content-Type" = "application/json"
        }

        $params = @{
            Uri = $url
            Method = $Method
            Headers = $headers
            SkipCertificateCheck = $true
        }

        if ($Data) {
            $params["Body"] = $Data
        }

        $response = Invoke-WebRequest @params -ErrorAction Stop
        
        return @{
            StatusCode = $response.StatusCode
            Body = $response.Content
        }
    }
    catch {
        $statusCode = 500
        if ($_.Exception.Response) {
            $statusCode = [int]$_.Exception.Response.StatusCode
        }
        
        $body = ""
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $body = $reader.ReadToEnd()
            $reader.Close()
        }
        
        return @{
            StatusCode = $statusCode
            Body = $body
        }
    }
}

# Helper function to create a flow and return its ID
# Returns: Flow ID on success, empty string on failure
function Create-Flow {
    param(
        [Parameter(Mandatory=$true)]
        [string]$FlowFilePath
    )
    
    $flowPayload = Get-Content -Path $FlowFilePath -Raw
    $flowJson = $flowPayload | ConvertFrom-Json
    $flowDisplayName = $flowJson.name
    
    if (-not $flowDisplayName) {
        Log-Warning "Could not extract flow name from $(Split-Path $FlowFilePath -Leaf), skipping"
        return ""
    }
    
    Log-Info "Creating flow: $flowDisplayName"
    
    $response = Invoke-ThunderApi -Method POST -Endpoint "/flows" -Data $flowPayload
    
    if ($response.StatusCode -eq 201 -or $response.StatusCode -eq 200) {
        $body = $response.Body | ConvertFrom-Json
        $flowId = $body.id
        Log-Success "Flow '$flowDisplayName' created successfully (ID: $flowId)"
        return $flowId
    }
    elseif ($response.StatusCode -eq 409) {
        Log-Warning "Flow '$flowDisplayName' already exists, skipping"
        return ""
    }
    else {
        Log-Error "Failed to create flow '$flowDisplayName' (HTTP $($response.StatusCode))"
        Log-Error "Response: $($response.Body)"
        return ""
    }
}

# Helper function to update a flow
# Returns: $true on success, $false on failure
function Update-Flow {
    param(
        [Parameter(Mandatory=$true)]
        [string]$FlowId,
        [Parameter(Mandatory=$true)]
        [string]$FlowFilePath
    )
    
    $flowPayload = Get-Content -Path $FlowFilePath -Raw
    $flowJson = $flowPayload | ConvertFrom-Json
    $flowDisplayName = $flowJson.name
    
    if (-not $flowDisplayName) {
        Log-Warning "Could not extract flow name from $(Split-Path $FlowFilePath -Leaf), skipping"
        return $false
    }
    
    Log-Info "Updating existing flow: $flowDisplayName (ID: $FlowId)"
    
    $response = Invoke-ThunderApi -Method PUT -Endpoint "/flows/$FlowId" -Data $flowPayload
    
    if ($response.StatusCode -eq 200) {
        Log-Success "Flow '$flowDisplayName' updated successfully"
        return $true
    }
    else {
        Log-Error "Failed to update flow '$flowDisplayName' (HTTP $($response.StatusCode))"
        Log-Error "Response: $($response.Body)"
        return $false
    }
}
