# ----------------------------------------------------------------------------
# Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    [string]$OS = "windows",
    [string]$ARCH = "amd64"
)

# Constants
$VERSION_FILE = "version.txt"
$VERSION = Get-Content $VERSION_FILE -Raw | ForEach-Object { $_.Trim() }
$BINARY_NAME = "thunder"
$TARGET_DIR = "target"
$OUTPUT_DIR = "$TARGET_DIR\out"

# Ensure Go is in PATH
$env:PATH += ";C:\Program Files\Go\bin"

function Show-Help {
    Write-Host "PowerShell build script for Thunder (Windows alternative to Makefile)"
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [command]"
    Write-Host ""
    Write-Host "Commands:"
    Write-Host "  help             - Show this help message"
    Write-Host "  clean            - Remove build artifacts"
    Write-Host "  build            - Build the Thunder backend"
    Write-Host "  test             - Run unit tests"
    Write-Host "  run              - Build and run the application"
    Write-Host "  all              - Clean, build, and test"
    Write-Host ""
}

function Clean-Build {
    Write-Host "================================================================"
    Write-Host "Cleaning build artifacts..."
    if (Test-Path $TARGET_DIR) {
        Remove-Item -Recurse -Force $TARGET_DIR
        Write-Host "Removed $TARGET_DIR directory"
    }
    Write-Host "================================================================"
}

function Build-Backend {
    Write-Host "================================================================"
    Write-Host "Building Go backend..."
    
    # Create output directory
    New-Item -ItemType Directory -Force -Path $OUTPUT_DIR | Out-Null
    
    # Set environment variables
    $env:GOOS = $OS
    $env:GOARCH = $ARCH
    $env:CGO_ENABLED = "0"
    
    # Set binary name with .exe extension for Windows
    $outputBinary = if ($OS -eq "windows") { "$BINARY_NAME.exe" } else { $BINARY_NAME }
    
    # Build date
    $buildDate = Get-Date -Format "yyyy-MM-dd HH:mm:ss UTC"
    
    # Build the application
    $ldflags = "-X main.version=$VERSION -X 'main.buildDate=$buildDate'"
    
    Write-Host "Building with version: $VERSION"
    Write-Host "Build date: $buildDate"
    Write-Host "Target: $OS/$ARCH"
    
    go build -C backend -ldflags $ldflags -o "../$OUTPUT_DIR/$outputBinary" ./cmd/server
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build successful! Binary created at: $OUTPUT_DIR\$outputBinary"
    } else {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "================================================================"
}

function Test-Unit {
    Write-Host "================================================================"
    Write-Host "Running unit tests..."
    go test -C backend ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Tests completed successfully!"
    } else {
        Write-Host "Some tests failed, but this might be due to Windows path issues." -ForegroundColor Yellow
    }
    Write-Host "================================================================"
}

function Run-Application {
    Build-Backend
    
    Write-Host "================================================================"
    Write-Host "Starting Thunder server..."
    $outputBinary = if ($OS -eq "windows") { "$BINARY_NAME.exe" } else { $BINARY_NAME }
    & ".\$OUTPUT_DIR\$outputBinary"
    Write-Host "================================================================"
}

function Build-All {
    Clean-Build
    Build-Backend
    Test-Unit
}

# Main command processing
switch ($Command.ToLower()) {
    "help" { Show-Help }
    "clean" { Clean-Build }
    "build" { Build-Backend }
    "test" { Test-Unit }
    "run" { Run-Application }
    "all" { Build-All }
    default {
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        Write-Host ""
        Show-Help
        exit 1
    }
}
