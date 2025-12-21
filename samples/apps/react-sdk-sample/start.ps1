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


$SERVER_PORT = 3000
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition

function KillPort {
    param([int]$Port)
    # Try modern cmdlet first
    try {
        $conns = Get-NetTCPConnection -LocalPort $Port -ErrorAction Stop
        $pids = $conns | Select-Object -Unique -ExpandProperty OwningProcess
        foreach ($p in $pids) {
            if ($p -and $p -ne $PID) { Stop-Process -Id $p -Force -ErrorAction SilentlyContinue }
        }
    }
    catch {
        # Fallback to netstat parsing
        $lines = netstat -ano 2>$null | Select-String ":$Port"
        foreach ($line in $lines) {
            $parts = ($line -split '\s+') | Where-Object { $_ -ne '' }
            $foundPid = $parts[-1]
            if ($foundPid -and ([int]$foundPid -ne $PID)) { Stop-Process -Id $foundPid -Force -ErrorAction SilentlyContinue }
        }
    }
}

KillPort -Port $SERVER_PORT

# Check if npx is available
$npx = Get-Command npx -ErrorAction SilentlyContinue
if (-not $npx) {
    Write-Host "[ERROR] Error: npx is not installed. Please install Node.js and npm."
    exit 1
}

# Check if certificates exist in dist folder
$distPath = Join-Path $scriptDir 'dist'
$certFile = Join-Path $distPath 'server.cert'
$keyFile = Join-Path $distPath 'server.key'

Write-Host "[INFO] Starting React SDK Sample App Server on port $SERVER_PORT..."
Write-Host ""
Write-Host "[INFO] Serving static files from ./dist directory"
Write-Host ""

# Start the static file server with HTTPS if certificates exist
if ((Test-Path $certFile) -and (Test-Path $keyFile)) {
    Write-Host "[INFO] Using HTTPS with SSL certificates"
    $arguments = @("serve", "-s", $distPath, "-l", $SERVER_PORT, "--ssl-cert", $certFile, "--ssl-key", $keyFile)
    $protocol = "https"
} else {
    Write-Host "[WARN] SSL certificates not found. Running with HTTP"
    Write-Host "     Run the build script to generate certificates"
    $arguments = @("serve", "-s", $distPath, "-l", $SERVER_PORT)
    $protocol = "http"
}

# Start process and keep a handle to it
$proc = Start-Process -FilePath "npx" -ArgumentList $arguments -PassThru -WorkingDirectory $scriptDir -NoNewWindow

Write-Host ""
Write-Host "[INFO] React SDK Sample App running at ${protocol}://localhost:$SERVER_PORT"
Write-Host "Press Ctrl+C to stop the server."
Write-Host ""

# Wait for the process
try {
    Wait-Process -Id $proc.Id
}
catch [System.Management.Automation.PipelineStoppedException] {
    # User pressed Ctrl+C
    Write-Host "`n[STOP] Stopping server..."
}
finally {
    if ($proc -and -not $proc.HasExited) {
        try {
            Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
        }
        catch {
            Write-Host "Unable to kill the process $($proc.Id)"
        }
    }
}
