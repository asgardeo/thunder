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

$BACKEND_PORT = if ($env:BACKEND_PORT) { [int]$env:BACKEND_PORT } else { 8090 }
$DEBUG_PORT = if ($env:DEBUG_PORT) { [int]$env:DEBUG_PORT } else { 2345 }
$DEBUG_MODE = $false

# Parse command line arguments
$i = 0
while ($i -lt $args.Count) {
    switch ($args[$i]) {
        '--debug' {
            $DEBUG_MODE = $true
            $i++
            break
        }
        '--debug-port' {
            $i++
            if ($i -lt $args.Count) {
                $DEBUG_PORT = [int]$args[$i]
                $i++
            }
            else {
                Write-Host "Missing value for --debug-port" -ForegroundColor Red
                exit 1
            }
            break
        }
        '--port' {
            $i++
            if ($i -lt $args.Count) {
                $BACKEND_PORT = [int]$args[$i]
                $i++
            }
            else {
                Write-Host "Missing value for --port" -ForegroundColor Red
                exit 1
            }
            break
        }
        '--help' {
            Write-Host "Thunder Server Startup Script"
            Write-Host ""
            Write-Host "Usage: .\start.ps1 [options]"
            Write-Host ""
            Write-Host "Options:"
            Write-Host "  --debug              Enable debug mode with remote debugging"
            Write-Host "  --port PORT          Set application port (default: 8090)"
            Write-Host "  --debug-port PORT    Set debug port (default: 2345)"
            Write-Host "  --help               Show this help message"
            Write-Host ""
            Write-Host "First-Time Setup:"
            Write-Host "  For initial setup, use the setup script:"
            Write-Host "    .\setup.ps1"
            Write-Host ""
            Write-Host "  Then start the server normally:"
            Write-Host "    .\start.ps1"
            Write-Host ""
            Write-Host "Examples:"
            Write-Host "  .\start.ps1                    Start server normally"
            Write-Host "  .\start.ps1 --debug            Start in debug mode"
            Write-Host "  .\start.ps1 --port 9090        Start on custom port"
            exit 0
        }
        default {
            Write-Host "Unknown option: $($args[$i])" -ForegroundColor Yellow
            Write-Host "Use --help for usage information"
            exit 1
        }
    }
}

# Exit on any error
$ErrorActionPreference = 'Stop'

function Stop-PortListener {
    param (
        [int]$port
    )

    Write-Host "Checking for processes listening on TCP port $port..."

    # Try Get-NetTCPConnection first (Windows 8/Server 2012+)
    try {
        $pids = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction Stop | Select-Object -ExpandProperty OwningProcess -Unique
    }
    catch {
        # Fallback to netstat parsing
        $pids = @()
        try {
            $netstat = & netstat -ano 2>$null | Select-String ":$port"
            foreach ($line in $netstat) {
                $parts = ($line -split '\s+') | Where-Object { $_ -ne '' }
                if ($parts.Count -ge 5) {
                    $procId = $parts[-1]
                    if ([int]::TryParse($procId, [ref]$null)) { $pids += [int]$procId }
                }
            }
        }
        catch { }
    }

    $pids = $pids | Where-Object { $_ -and ($_ -ne 0) } | Select-Object -Unique
    foreach ($procId in $pids) {
        try {
            Write-Host "Killing PID $procId that is listening on port $port"
            Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
        }
        catch {
            Write-Host "Unable to kill PID $procId : $_" -ForegroundColor Yellow
        }
    }
}

# Kill ports before binding
Stop-PortListener -port $BACKEND_PORT
if ($DEBUG_MODE) { Stop-PortListener -port $DEBUG_PORT }
Start-Sleep -Seconds 1

# Check if Delve is available for debug mode
if ($DEBUG_MODE) {
    if (-not (Get-Command dlv -ErrorAction SilentlyContinue)) {
        Write-Host "[ERROR] Debug mode requires Delve debugger" -ForegroundColor Red
        Write-Host ""
        Write-Host "[INFO] Install Delve using:" -ForegroundColor Cyan
        Write-Host "   go install github.com/go-delve/delve/cmd/dlv@latest" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "[INFO] Add Delve to PATH and re-run this script with --debug" -ForegroundColor Cyan
        exit 1
    }
}

# Resolve thunder executable path
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$possible = @(
    (Join-Path $scriptDir 'thunder.exe'),
    (Join-Path $scriptDir 'thunder')
)
$thunderPath = $possible | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $thunderPath) {
    # Fallback to ./thunder (will work if PATH or current dir has it)
    $thunderPath = Join-Path $scriptDir 'thunder'
}

$proc = $null
try {
    if ($DEBUG_MODE) {
        Write-Host "[INFO] Starting Thunder Server in DEBUG mode..."
        Write-Host "[INFO] Application will run on: https://localhost:$BACKEND_PORT"
        Write-Host "[INFO] Remote debugger will listen on: localhost:$DEBUG_PORT"
        Write-Host ""
        Write-Host "[INFO] Connect using remote debugging configuration:" -ForegroundColor Gray
        Write-Host "   Host: 127.0.0.1, Port: $DEBUG_PORT" -ForegroundColor Gray
        Write-Host ""

        # Start Delve in headless mode
        $dlvArgs =  @(
            'exec'
            "--listen=:$DEBUG_PORT"
            '--headless=true'
            '--api-version=2'
            '--accept-multiclient'
            '--continue'
            $thunderPath
        )
        $proc = Start-Process -FilePath dlv -ArgumentList $dlvArgs -WorkingDirectory $scriptDir -NoNewWindow -PassThru
    }
    else {
        Write-Host "[INFO] Starting Thunder Server ..."

        # Export BACKEND_PORT for the child process
        $env:BACKEND_PORT = $BACKEND_PORT
        $proc = Start-Process -FilePath $thunderPath -WorkingDirectory $scriptDir -NoNewWindow -PassThru
    }

    Write-Host ""
    Write-Host "[INFO] Server running. PID: $($proc.Id)"
    Write-Host ""
    Write-Host "[INFO] Frontend Apps:"
    Write-Host "   [GATE] Gate (Login/Register): $BACKEND_PORT/gate"
    Write-Host "   [DEV]  Develop (Admin Console): $BACKEND_PORT/develop"
    Write-Host ""

    Write-Host "Press Ctrl+C to stop the server."

    # Wait for the background process. This will block until the process exits.
    Wait-Process -Id $proc.Id
}
finally {
    Write-Host "`n[STOP] Stopping server..."
    if ($proc -and -not $proc.HasExited) {
        try { Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue } catch { }
    }
}
