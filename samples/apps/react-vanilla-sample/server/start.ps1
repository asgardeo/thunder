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

Write-Host "⚡ Starting App Server ..."

# Start process and keep a handle to it (let Start-Process error if server not present)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$serverPath = Join-Path $scriptDir 'server'
$env:SERVER_PORT = $SERVER_PORT
$proc = Start-Process -FilePath $serverPath -PassThru -WorkingDirectory $scriptDir -NoNewWindow

Write-Host ""
Write-Host "🚀 App Server running. PID: $($proc.Id)"

Wait-Process -Id $proc.Id

Write-Host "`n🛑 Stopping server..."
if ($proc -and -not $proc.HasExited) {
    try { Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue } 
    catch {
        Write-Host "Unable to kill the process $($proc.Id)"
    }
}
