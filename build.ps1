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

[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [string]$Command,
    
    [Parameter(Position = 1)]
    [string]$GO_OS,
    
    [Parameter(Position = 2)]
    [string]$GO_ARCH
)

$ErrorActionPreference = "Stop"

$SCRIPT_DIR = $PSScriptRoot

# --- Set Default OS and the architecture --- 
# Auto-detect GO OS
if ([string]::IsNullOrEmpty($GO_OS)) {
    try {
        $DEFAULT_OS = & go env GOOS
        if ([string]::IsNullOrEmpty($DEFAULT_OS)) {
            throw "Go environment not found"
        }
    }
    catch {
        $DEFAULT_OS = "windows"
    }
    $GO_OS = $DEFAULT_OS
}

# Auto-detect GO ARCH
if ([string]::IsNullOrEmpty($GO_ARCH)) {
    try {
        $DEFAULT_ARCH = & go env GOARCH
        if ([string]::IsNullOrEmpty($DEFAULT_ARCH)) {
            throw "Go environment not found"
        }
    }
    catch {
        # Use PowerShell to detect architecture
        if ([Environment]::Is64BitOperatingSystem) {
            $DEFAULT_ARCH = "amd64"
        }
        else {
            throw "Unsupported architecture"
        }
    }
    $GO_ARCH = $DEFAULT_ARCH
}

Write-Host "Using GO OS: $GO_OS and ARCH: $GO_ARCH"

$SAMPLE_DIST_NODE_VERSION = "node18"
$SAMPLE_DIST_OS = $GO_OS
$SAMPLE_DIST_ARCH = $GO_ARCH

# Transform OS for node packaging executor
if ($SAMPLE_DIST_OS -eq "darwin") {
    $SAMPLE_DIST_OS = "macos"
}
elseif ($SAMPLE_DIST_OS -eq "windows") {
    $SAMPLE_DIST_OS = "win"
}

if ($SAMPLE_DIST_ARCH -eq "amd64") {
    $SAMPLE_DIST_ARCH = "x64"
}

# --- Thunder Package Distribution details ---
$GO_PACKAGE_OS = $GO_OS
$GO_PACKAGE_ARCH = $GO_ARCH

# Normalize OS name for distribution packaging
if ($GO_OS -eq "darwin") {
    $GO_PACKAGE_OS = "macos"
}
elseif ($GO_OS -eq "windows") {
    $GO_PACKAGE_OS = "win"
}

if ($GO_ARCH -eq "amd64") {
    $GO_PACKAGE_ARCH = "x64"
}

$VERSION_FILE = "version.txt"
$VERSION = Get-Content $VERSION_FILE -Raw
$VERSION = $VERSION.Trim()
$THUNDER_VERSION = $VERSION
if ($THUNDER_VERSION.StartsWith("v")) {
    $THUNDER_VERSION = $THUNDER_VERSION.Substring(1)
}
$BINARY_NAME = "thunder"
$PRODUCT_FOLDER = "${BINARY_NAME}-${THUNDER_VERSION}-${GO_PACKAGE_OS}-${GO_PACKAGE_ARCH}"

# --- Sample App Distribution details ---
$SAMPLE_PACKAGE_OS = $SAMPLE_DIST_OS
$SAMPLE_PACKAGE_ARCH = $SAMPLE_DIST_ARCH

$SAMPLE_APP_SERVER_BINARY_NAME = "server"
$packageJson = Get-Content "samples/apps/oauth/package.json" -Raw | ConvertFrom-Json
$SAMPLE_APP_VERSION = $packageJson.version
$SAMPLE_APP_FOLDER = "sample-app-${SAMPLE_APP_VERSION}-${SAMPLE_PACKAGE_OS}-${SAMPLE_PACKAGE_ARCH}"

# Server ports
$BACKEND_PORT = 8090

# Directories
$TARGET_DIR = "target"
$OUTPUT_DIR = Join-Path $TARGET_DIR "out"
$DIST_DIR = Join-Path $TARGET_DIR "dist"
$BUILD_DIR = Join-Path $OUTPUT_DIR ".build"
$LOCAL_CERT_DIR = Join-Path $OUTPUT_DIR ".cert"
$BACKEND_BASE_DIR = "backend"
$BACKEND_DIR = Join-Path $BACKEND_BASE_DIR "cmd/server"
$REPOSITORY_DIR = Join-Path $BACKEND_BASE_DIR "cmd/server/repository"
$REPOSITORY_DB_DIR = Join-Path $REPOSITORY_DIR "database"
$SERVER_SCRIPTS_DIR = Join-Path $BACKEND_BASE_DIR "scripts"
$SERVER_DB_SCRIPTS_DIR = Join-Path $BACKEND_BASE_DIR "dbscripts"
$SECURITY_DIR = "repository/resources/security"
$FRONTEND_BASE_DIR = "frontend"
$GATE_APP_DIST_DIR = "apps/gate"
$DEVELOP_APP_DIST_DIR = "apps/develop"
$FRONTEND_GATE_APP_SOURCE_DIR = Join-Path $FRONTEND_BASE_DIR "apps/thunder-gate"
$FRONTEND_DEVELOP_APP_SOURCE_DIR = Join-Path $FRONTEND_BASE_DIR "apps/thunder-develop"
$SAMPLE_BASE_DIR = "samples"
$SAMPLE_APP_DIR = Join-Path $SAMPLE_BASE_DIR "apps/oauth"
$SAMPLE_APP_SERVER_DIR = Join-Path $SAMPLE_APP_DIR "server"

function Get-CoverageExclusionPattern {
    # Read exclusion patterns (full package paths) from .excludecoverage file
    # This function can be called from any directory
    
    $coverage_exclude_file = $null
    
    # Check if we're already in the backend directory or need to use relative path
    if (Test-Path ".excludecoverage") {
        $coverage_exclude_file = ".excludecoverage"
    }
    elseif (Test-Path (Join-Path $SCRIPT_DIR $BACKEND_BASE_DIR ".excludecoverage")) {
        $coverage_exclude_file = Join-Path $SCRIPT_DIR $BACKEND_BASE_DIR ".excludecoverage"
    }
    else {
        return ""
    }
    
    # Read non-comment, non-empty lines and join with '|' for regex (exact package path matching)
    $patterns = Get-Content $coverage_exclude_file | Where-Object { 
        $_ -notmatch '^\s*#' -and $_ -notmatch '^\s*$' 
    }
    
    if ($patterns) {
        return ($patterns -join '|')
    }
    
    return ""
}

function Clean-All {
    Write-Host "================================================================"
    Write-Host "Cleaning all build artifacts..."
    if (Test-Path $TARGET_DIR) {
        Remove-Item -Path $TARGET_DIR -Recurse -Force -ErrorAction SilentlyContinue
    }

    Write-Host "Removing certificates in the $BACKEND_DIR/$SECURITY_DIR"
    if (Test-Path (Join-Path $BACKEND_DIR $SECURITY_DIR)) {
        Remove-Item -Path (Join-Path $BACKEND_DIR $SECURITY_DIR) -Recurse -Force -ErrorAction SilentlyContinue
    }

    Write-Host "Removing certificates in the $SAMPLE_APP_DIR"
    Remove-Item -Path (Join-Path $SAMPLE_APP_DIR "server.cert") -Force -ErrorAction SilentlyContinue
    Remove-Item -Path (Join-Path $SAMPLE_APP_DIR "server.key") -Force -ErrorAction SilentlyContinue
    Remove-Item -Path (Join-Path $SAMPLE_APP_SERVER_DIR "server.cert") -Force -ErrorAction SilentlyContinue
    Remove-Item -Path (Join-Path $SAMPLE_APP_SERVER_DIR "server.key") -Force -ErrorAction SilentlyContinue
    Write-Host "================================================================"
}

function Clean {
    Write-Host "================================================================"
    Write-Host "Cleaning build artifacts..."
    if (Test-Path $OUTPUT_DIR) {
        Remove-Item -Path $OUTPUT_DIR -Recurse -Force -ErrorAction SilentlyContinue
    }

    Write-Host "Removing certificates in the $BACKEND_DIR/$SECURITY_DIR"
    if (Test-Path (Join-Path $BACKEND_DIR $SECURITY_DIR)) {
        Remove-Item -Path (Join-Path $BACKEND_DIR $SECURITY_DIR) -Recurse -Force -ErrorAction SilentlyContinue
    }

    Write-Host "Removing certificates in the $SAMPLE_APP_DIR"
    Remove-Item -Path (Join-Path $SAMPLE_APP_DIR "server.cert") -Force -ErrorAction SilentlyContinue
    Remove-Item -Path (Join-Path $SAMPLE_APP_DIR "server.key") -Force -ErrorAction SilentlyContinue
    Write-Host "================================================================"
}

function Build-Backend {
    Write-Host "================================================================"
    Write-Host "Building Go backend..."
    New-Item -Path $BUILD_DIR -ItemType Directory -Force | Out-Null

    # Set binary name with .exe extension for Windows
    $output_binary = $BINARY_NAME
    if ($GO_OS -eq "windows") {
        $output_binary = "${BINARY_NAME}.exe"
    }

    # Prepare build date without spaces to avoid ldflags splitting
    $buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

    $env:GOOS = $GO_OS
    $env:GOARCH = $GO_ARCH
    $env:CGO_ENABLED = "0"

    # Check if coverage build is requested via ENABLE_COVERAGE environment variable
    $buildArgs = @('build', '-x')
    if ($env:ENABLE_COVERAGE -eq "true") {
        Write-Host "Building with coverage instrumentation enabled..."
        
        # Build coverage package list, excluding patterns from .excludecoverage
        Push-Location $BACKEND_BASE_DIR
        try {
            $exclude_pattern = Get-CoverageExclusionPattern
            $coverpkg = ""
            
            if ($exclude_pattern) {
                Write-Host "Excluding coverage for patterns: $exclude_pattern"
                $packages = & go list ./...
                $filtered_packages = $packages | Where-Object { $_ -notmatch $exclude_pattern }
                $coverpkg = $filtered_packages -join ','
            }
            else {
                $packages = & go list ./...
                $coverpkg = $packages -join ','
            }
        }
        finally {
            Pop-Location
        }
        
        $buildArgs += @('-cover', "-coverpkg=$coverpkg")
    }

    # Construct ldflags safely and pass as an argument array to avoid PowerShell splitting
    $ldflags = "-X main.version=$VERSION -X main.buildDate=$buildDate"
    $outputPath = "../$BUILD_DIR/$output_binary"
    $buildArgs += @('-ldflags', $ldflags, '-o', $outputPath, './cmd/server')

    Write-Host "Executing: go $($buildArgs -join ' ')"

    Push-Location $BACKEND_BASE_DIR
    try {
        & go @buildArgs
        if ($LASTEXITCODE -ne 0) {
            throw "Go build failed with exit code $LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }

    Write-Host "Initializing databases..."
    Initialize-Databases -override $true
    Write-Host "================================================================"
}

function Build-Frontend {
    Write-Host "================================================================"
    Write-Host "Building frontend apps..."
    
    # Check if pnpm is installed, if not install it
    if (-not (Get-Command pnpm -ErrorAction SilentlyContinue)) {
        Write-Host "pnpm not found, installing..."
        & npm install -g pnpm
    }
    
    # Navigate to frontend directory and install dependencies
    Push-Location $FRONTEND_BASE_DIR
    try {
        Write-Host "Installing frontend dependencies..."
        & pnpm install
        
        Write-Host "Building frontend applications & packages..."
        & pnpm build
    }
    finally {
        Pop-Location
    }
    
    Write-Host "================================================================"
}

function Initialize-Databases {
    param(
        [bool]$override = $false
    )
    
    Write-Host "================================================================"
    Write-Host "Initializing SQLite databases..."

    # Check for sqlite3 CLI availability
    $sqliteCmd = Get-Command sqlite3 -ErrorAction SilentlyContinue
    if (-not $sqliteCmd) {
        Write-Host ""
        Write-Host "ERROR: 'sqlite3' CLI not found on PATH. The build script uses the sqlite3 command to initialize local SQLite databases."
        Write-Host "On Windows you can install sqlite3 using one of the following methods:"
        Write-Host "  1) Chocolatey (requires admin PowerShell):"
        Write-Host "       choco install sqlite" 
        Write-Host "  2) Scoop (recommended for user installs):"
        Write-Host "       scoop install sqlite" 
        Write-Host "  3) Download prebuilt binaries from https://www.sqlite.org/download.html and add the folder to your PATH."
        Write-Host ""
        Write-Host "Alternatively, skip database initialization and create the DB files manually under '$REPOSITORY_DB_DIR'."
        throw "sqlite3 CLI not found. Install sqlite3 and re-run the build."
    }

    New-Item -Path $REPOSITORY_DB_DIR -ItemType Directory -Force | Out-Null

    $db_files = @("thunderdb.db", "runtimedb.db")
    $script_paths = @("thunderdb/sqlite.sql", "runtimedb/sqlite.sql")

    for ($i = 0; $i -lt $db_files.Length; $i++) {
        $db_file = $db_files[$i]
        $script_rel_path = $script_paths[$i]
        $db_path = Join-Path $REPOSITORY_DB_DIR $db_file
        $script_path = Join-Path $SERVER_DB_SCRIPTS_DIR $script_rel_path

        if (Test-Path $script_path) {
            if (Test-Path $db_path) {
                if ($override) {
                    Write-Host " - Removing existing $db_file as override is true"
                    Remove-Item $db_path -Force
                }
                else {
                    Write-Host " ! Skipping $db_file : DB already exists. Delete the existing and re-run to recreate."
                    continue
                }
            }

            Write-Host " - Creating $db_file using $script_path"
            # Use sqlite3 command line tool
            & sqlite3 $db_path ".read $script_path"
            if ($LASTEXITCODE -ne 0) {
                throw "SQLite operation failed with exit code $LASTEXITCODE"
            }
        }
        else {
            Write-Host " ! Skipping $db_file : SQL script not found at $script_path"
        }
    }

    Write-Host "SQLite database initialization complete."
    Write-Host "================================================================"
}

function Prepare-Backend-For-Packaging {
    Write-Host "================================================================"
    Write-Host "Copying backend artifacts..."

    # Use appropriate binary name based on OS
    $binary_name = $BINARY_NAME
    if ($GO_OS -eq "windows") {
        $binary_name = "${BINARY_NAME}.exe"
    }

    $package_folder = Join-Path $DIST_DIR $PRODUCT_FOLDER
    Copy-Item -Path (Join-Path $BUILD_DIR $binary_name) -Destination $package_folder -Force
    Copy-Item -Path $REPOSITORY_DIR -Destination $package_folder -Recurse -Force
    Copy-Item -Path $VERSION_FILE -Destination $package_folder -Force
    Copy-Item -Path $SERVER_SCRIPTS_DIR -Destination $package_folder -Recurse -Force
    Copy-Item -Path $SERVER_DB_SCRIPTS_DIR -Destination $package_folder -Recurse -Force
    
    $security_dir = Join-Path $package_folder $SECURITY_DIR
    New-Item -Path $security_dir -ItemType Directory -Force | Out-Null

    Write-Host "=== Ensuring server certificates exist in the distribution ==="
    Ensure-Certificates -cert_dir $security_dir
    Write-Host "================================================================"
}

function Prepare-Frontend-For-Packaging {
    Write-Host "================================================================"
    Write-Host "Copying frontend artifacts..."

    $package_folder = Join-Path $DIST_DIR $PRODUCT_FOLDER
    New-Item -Path (Join-Path $package_folder $GATE_APP_DIST_DIR) -ItemType Directory -Force | Out-Null
    New-Item -Path (Join-Path $package_folder $DEVELOP_APP_DIST_DIR) -ItemType Directory -Force | Out-Null

    # Copy gate app build output
    if (Test-Path (Join-Path $FRONTEND_GATE_APP_SOURCE_DIR "dist")) {
        Write-Host "Copying Gate app build output..."
        Get-ChildItem -Path (Join-Path $FRONTEND_GATE_APP_SOURCE_DIR "dist") -Force | ForEach-Object {
            Copy-Item -Path $_.FullName -Destination (Join-Path $package_folder $GATE_APP_DIST_DIR) -Recurse -Force
        }
    }
    else {
        Write-Host "Warning: Gate app build output not found at $((Join-Path $FRONTEND_GATE_APP_SOURCE_DIR "dist"))"
    }
    
    # Copy develop app build output
    if (Test-Path (Join-Path $FRONTEND_DEVELOP_APP_SOURCE_DIR "dist")) {
        Write-Host "Copying Develop app build output..."
        Get-ChildItem -Path (Join-Path $FRONTEND_DEVELOP_APP_SOURCE_DIR "dist") -Force | ForEach-Object {
            Copy-Item -Path $_.FullName -Destination (Join-Path $package_folder $DEVELOP_APP_DIST_DIR) -Recurse -Force
        }
    }
    else {
        Write-Host "Warning: Develop app build output not found at $((Join-Path $FRONTEND_DEVELOP_APP_SOURCE_DIR "dist"))"
    }

    Write-Host "================================================================"
}

function Package {
    Write-Host "================================================================"
    Write-Host "Packaging backend & frontend artifacts..."

    $package_folder = Join-Path $DIST_DIR $PRODUCT_FOLDER
    New-Item -Path $package_folder -ItemType Directory -Force | Out-Null

    Prepare-Frontend-For-Packaging
    Prepare-Backend-For-Packaging

    # Copy the appropriate startup script based on the target OS
    if ($GO_OS -eq "windows") {
        Write-Host "Including Windows start script (start.ps1)..."
        Copy-Item -Path "start.ps1" -Destination $package_folder -Force
    }
    else {
        Write-Host "Including Unix start script (start.sh)..."
        Copy-Item -Path "start.sh" -Destination $package_folder -Force
    }

    Write-Host "Creating zip file..."
    $zipFile = Join-Path $DIST_DIR "$PRODUCT_FOLDER.zip"
    if (Test-Path $zipFile) {
        Remove-Item $zipFile -Force
    }
    
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::CreateFromDirectory($package_folder, $zipFile)
    
    Remove-Item -Path $package_folder -Recurse -Force
    if (Test-Path $BUILD_DIR) {
        Remove-Item -Path $BUILD_DIR -Recurse -Force
    }
    Write-Host "================================================================"
}

function Build-Sample-App {
    Write-Host "================================================================"
    Write-Host "Building sample app..."

    # Ensure certificate exists for the sample app
    Write-Host "=== Ensuring sample app certificates exist ==="
    Ensure-Certificates -cert_dir $SAMPLE_APP_DIR
    
    # Build the application
    Push-Location $SAMPLE_APP_DIR
    try {
        Write-Host "Installing dependencies..."
        & npm install
        if ($LASTEXITCODE -ne 0) {
            throw "npm install failed with exit code $LASTEXITCODE"
        }
        
        Write-Host "Building the app (TypeScript + Vite)..."

        # Use npx to invoke local tsc and vite to avoid shell-specific npm script commands
        Write-Host " - Running TypeScript build (tsc -b)..."
        & npx tsc -b
        if ($LASTEXITCODE -ne 0) {
            throw "tsc build failed with exit code $LASTEXITCODE"
        }

        Write-Host " - Running Vite build..."
        & npx vite build
        if ($LASTEXITCODE -ne 0) {
            throw "vite build failed with exit code $LASTEXITCODE"
        }

        # Replicate npm script: copy dist to server/app and copy certs using PowerShell (cross-platform safe)
        $serverDir = Join-Path $SAMPLE_APP_DIR "server"
        $serverAppDir = Join-Path $serverDir "app"
        if (Test-Path $serverAppDir) { 
            Remove-Item -Path $serverAppDir -Recurse -Force 
        }
        New-Item -Path $serverAppDir -ItemType Directory -Force | Out-Null

        # Resolve absolute dist path to avoid path duplication when Push-Location is used
        # We're already inside $SAMPLE_APP_DIR due to Push-Location, so resolve relative 'dist'
        $distFull = Resolve-Path -Path "dist" | Select-Object -ExpandProperty Path
        Copy-Item -Path (Join-Path $distFull "*") -Destination $serverAppDir -Recurse -Force

        # Copy server certs into server directory
        if (Test-Path (Join-Path $SAMPLE_APP_DIR "server.key")) {
            Copy-Item -Path (Join-Path $SAMPLE_APP_DIR "server.key") -Destination $serverDir -Force
        }
        if (Test-Path (Join-Path $SAMPLE_APP_DIR "server.cert")) {
            Copy-Item -Path (Join-Path $SAMPLE_APP_DIR "server.cert") -Destination $serverDir -Force
        }

        # Install server dependencies
        Push-Location $serverDir
        try {
            Write-Host " - Installing server dependencies..."
            & npm install
            if ($LASTEXITCODE -ne 0) {
                throw "npm install (server) failed with exit code $LASTEXITCODE"
            }
        }
        finally {
            Pop-Location
        }
    }
    finally {
        Pop-Location
    }
    
    Write-Host "Sample app built successfully."
    Write-Host "================================================================"
}

function Package-Sample-App {
    Write-Host "================================================================"
    Write-Host "Copying sample artifacts..."

    # Use appropriate binary name based on OS
    $binary_name = $SAMPLE_APP_SERVER_BINARY_NAME
    $executable_name = "$SAMPLE_APP_SERVER_BINARY_NAME-$SAMPLE_DIST_OS-$SAMPLE_DIST_ARCH"

    if ($SAMPLE_DIST_OS -eq "win") {
        $binary_name = "${SAMPLE_APP_SERVER_BINARY_NAME}.exe"
        $executable_name = "${SAMPLE_APP_SERVER_BINARY_NAME}-${SAMPLE_DIST_OS}-${SAMPLE_DIST_ARCH}.exe"
    }
    
    $sample_app_folder = Join-Path $DIST_DIR $SAMPLE_APP_FOLDER
    # Ensure we have full absolute paths to avoid duplication when CreateFromDirectory is called
    New-Item -Path $sample_app_folder -ItemType Directory -Force | Out-Null
    $sample_app_folder = (Resolve-Path -Path $sample_app_folder).Path
    
    # Copy the built app files. If the server 'app' folder doesn't exist (build step may not have created it),
    # fall back to copying directly from the sample app 'dist' directory.
    $serverAppSource = Join-Path $SAMPLE_APP_SERVER_DIR "app"
    if (-not (Test-Path $serverAppSource)) {
        Write-Host "Server app folder '$serverAppSource' not found; falling back to copying from '$SAMPLE_APP_DIR/dist'..."
        # Ensure server directory exists
        New-Item -Path $SAMPLE_APP_SERVER_DIR -ItemType Directory -Force | Out-Null
        New-Item -Path $serverAppSource -ItemType Directory -Force | Out-Null

        # Resolve dist (relative to sample app dir)
        $distFull = Resolve-Path -Path (Join-Path $SAMPLE_APP_DIR "dist") | Select-Object -ExpandProperty Path
        Copy-Item -Path (Join-Path $distFull "*") -Destination $serverAppSource -Recurse -Force
    }

    Copy-Item -Path $serverAppSource -Destination $sample_app_folder -Recurse -Force

    Push-Location $SAMPLE_APP_SERVER_DIR
    try {
        New-Item -Path "executables" -ItemType Directory -Force | Out-Null

        & npx pkg . -t $SAMPLE_DIST_NODE_VERSION-$SAMPLE_DIST_OS-$SAMPLE_DIST_ARCH -o executables/$SAMPLE_APP_SERVER_BINARY_NAME-$SAMPLE_DIST_OS-$SAMPLE_DIST_ARCH
        if ($LASTEXITCODE -ne 0) {
            throw "npx pkg failed with exit code $LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }

    # Copy the server binary
    Copy-Item -Path (Join-Path $SAMPLE_APP_SERVER_DIR "executables/$executable_name") -Destination (Join-Path $sample_app_folder $binary_name) -Force

    # Ensure the certificates exist in the sample app directory
    Write-Host "=== Ensuring certificates exist in the sample distribution ==="
    Ensure-Certificates -cert_dir $sample_app_folder

    # Copy the appropriate startup script based on the target OS
    if ($SAMPLE_DIST_OS -eq "win") {
        Write-Host "Including Windows start script (start.ps1)..."
        Copy-Item -Path (Join-Path $SAMPLE_APP_SERVER_DIR "start.ps1") -Destination $sample_app_folder -Force
    }
    else {
        Write-Host "Including Unix start script (start.sh)..."
        Copy-Item -Path (Join-Path $SAMPLE_APP_SERVER_DIR "start.sh") -Destination $sample_app_folder -Force
    }

    Write-Host "Creating zip file..."
    # Construct an absolute, well-formed zip path using .NET helpers to avoid accidental
    # concatenation of already-absolute paths (which produced duplicated segments).
    $distAbs = (Resolve-Path -Path $DIST_DIR).Path
    $zipFile = [System.IO.Path]::Combine($distAbs, "$SAMPLE_APP_FOLDER.zip")
    if (Test-Path $zipFile) {
        Remove-Item $zipFile -Force
    }

    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::CreateFromDirectory($sample_app_folder, $zipFile)
    
    Remove-Item -Path $sample_app_folder -Recurse -Force
    
    Write-Host "Sample app packaged successfully as $zipFile"
    Write-Host "================================================================"
}

function Test-Unit {
    Write-Host "================================================================"
    Write-Host "Running unit tests with coverage..."
    
    Push-Location $BACKEND_BASE_DIR
    try {
        # Build coverage package list
        $exclude_pattern = Get-CoverageExclusionPattern
        $coverpkg = ""
        
        if ($exclude_pattern) {
            Write-Host "Excluding coverage for patterns: $exclude_pattern"
            $packages = & go list ./...
            $filtered_packages = $packages | Where-Object { $_ -notmatch $exclude_pattern }
            $coverpkg = $filtered_packages -join ','
        }
        else {
            Write-Host "No exclusion patterns found, including all packages"
            $packages = & go list ./...
            $coverpkg = $packages -join ','
        }
        
        # Check if gotestsum is available
        $gotestsum = Get-Command gotestsum -ErrorAction SilentlyContinue
        
        if ($gotestsum) {
            Write-Host "Running unit tests with coverage using gotestsum..."
            & gotestsum -- -v -coverprofile=coverage_unit.out -covermode=atomic "-coverpkg=$coverpkg" ./...
            if ($LASTEXITCODE -ne 0) {
                Write-Host "There are unit test failures."
                exit 1
            }
        }
        else {
            Write-Host "Running unit tests with coverage using go test..."
            & go test -v -coverprofile=coverage_unit.out -covermode=atomic "-coverpkg=$coverpkg" ./...
            if ($LASTEXITCODE -ne 0) {
                Write-Host "There are unit test failures."
                exit 1
            }
        }
        
        Write-Host "Unit test coverage profile generated in: backend/coverage_unit.out"
        
        # Generate HTML coverage report for unit tests
        & go tool cover -html=coverage_unit.out -o=coverage_unit.html
        Write-Host "Unit test coverage HTML report generated in: backend/coverage_unit.html"
        
        # Display unit test coverage summary
        Write-Host ""
        Write-Host "================================================================"
        Write-Host "Unit Test Coverage Summary:"
        & go tool cover -func=coverage_unit.out | Select-Object -Last 1
        Write-Host "================================================================"
        Write-Host ""
    }
    finally {
        Pop-Location
    }
    
    Write-Host "================================================================"
}

function Test-Integration {
    Write-Host "================================================================"
    Write-Host "Running integration tests..."
    
    Push-Location $SCRIPT_DIR
    try {
        # Set up coverage directory for integration tests
        $coverage_dir = Join-Path (Get-Location) "$OUTPUT_DIR\.test\integration"
        New-Item -Path $coverage_dir -ItemType Directory -Force | Out-Null
        
        # Export coverage directory for the server binary to use
        $env:GOCOVERDIR = $coverage_dir
        
        Write-Host "Coverage data will be collected in: $coverage_dir"
        & go run -C ./tests/integration ./main.go
        $test_exit_code = $LASTEXITCODE
        
        # Process coverage data if tests passed or failed
        if ((Test-Path $coverage_dir) -and ((Get-ChildItem $coverage_dir -ErrorAction SilentlyContinue).Count -gt 0)) {
            Write-Host "================================================================"
            Write-Host "Processing integration test coverage..."
            
            # Convert binary coverage data to text format
            Push-Location $BACKEND_BASE_DIR
            try {
                & go tool covdata textfmt -i="$coverage_dir" -o="../$TARGET_DIR/coverage_integration.out"
                Write-Host "Integration test coverage report generated in: $TARGET_DIR/coverage_integration.out"
                
                # Generate HTML coverage report
                & go tool cover -html="../$TARGET_DIR/coverage_integration.out" -o="../$TARGET_DIR/coverage_integration.html"
                Write-Host "Integration test coverage HTML report generated in: $TARGET_DIR/coverage_integration.html"
                
                # Display coverage summary
                Write-Host ""
                Write-Host "================================================================"
                Write-Host "Coverage Summary:"
                & go tool cover -func="../$TARGET_DIR/coverage_integration.out" | Select-Object -Last 1
                Write-Host "================================================================"
                Write-Host ""
            }
            finally {
                Pop-Location
            }
        }
        else {
            Write-Host "================================================================"
            Write-Host "No coverage data collected"
        }
        
        # Exit with the test exit code
        if ($test_exit_code -ne 0) {
            Write-Host "================================================================"
            Write-Host "Integration tests failed with exit code: $test_exit_code"
            exit $test_exit_code
        }
    }
    finally {
        Pop-Location
    }
    
    Write-Host "================================================================"
}

function Merge-Coverage {
    Write-Host "================================================================"
    Write-Host "Merging coverage reports..."
    
    Push-Location $SCRIPT_DIR
    try {
        $unit_coverage = Join-Path $BACKEND_BASE_DIR "coverage_unit.out"
        $integration_coverage = Join-Path $TARGET_DIR "coverage_integration.out"
        $combined_coverage = Join-Path $TARGET_DIR "coverage_combined.out"
        
        # Check if both coverage files exist
        if (-not (Test-Path $unit_coverage)) {
            Write-Host "Warning: Unit test coverage file not found at $unit_coverage"
            Write-Host "Skipping coverage merge."
            return
        }
        
        if (-not (Test-Path $integration_coverage)) {
            Write-Host "Warning: Integration test coverage file not found at $integration_coverage"
            Write-Host "Skipping coverage merge."
            return
        }
        
        Write-Host "Merging unit and integration test coverage..."
        
        # Get the mode from the first file and write to combined coverage
        $mode_line = Get-Content $unit_coverage -First 1
        $mode_line | Set-Content $combined_coverage
        
        # Read both files (skip mode lines) and merge overlapping coverage
        $unit_lines = Get-Content $unit_coverage | Select-Object -Skip 1
        $integration_lines = Get-Content $integration_coverage | Select-Object -Skip 1
        
        # Combine and process coverage data
        $coverage_map = @{}
        
        foreach ($line in ($unit_lines + $integration_lines)) {
            $parts = $line -split '\s+'
            if ($parts.Count -ge 3) {
                $key = "$($parts[0]) $($parts[1])"
                $count = [int]$parts[2]
                
                if ($coverage_map.ContainsKey($key)) {
                    # For duplicate entries, take the maximum count
                    if ($count -gt $coverage_map[$key]) {
                        $coverage_map[$key] = $count
                    }
                }
                else {
                    $coverage_map[$key] = $count
                }
            }
        }
        
        # Sort and write to combined coverage file
        $sorted_lines = $coverage_map.GetEnumerator() | Sort-Object Key | ForEach-Object {
            "$($_.Key) $($_.Value)"
        }
        
        $sorted_lines | Add-Content $combined_coverage
        
        Write-Host "Combined coverage report generated in: $combined_coverage"
        
        # Generate HTML coverage report for combined coverage
        Push-Location $BACKEND_BASE_DIR
        try {
            & go tool cover -html="../$combined_coverage" -o="../$TARGET_DIR/coverage_combined.html"
            Write-Host "Combined coverage HTML report generated in: $TARGET_DIR/coverage_combined.html"
            
            # Display combined coverage summary
            Write-Host ""
            Write-Host "================================================================"
            Write-Host "Combined Test Coverage Summary:"
            & go tool cover -func="../$combined_coverage" | Select-Object -Last 1
            Write-Host "================================================================"
            Write-Host ""
        }
        finally {
            Pop-Location
        }
    }
    finally {
        Pop-Location
    }
    
    Write-Host "================================================================"
}

function Export-CertificateAndKeyToPem {
    param(
        [System.Security.Cryptography.X509Certificates.X509Certificate2]$cert,
        [string]$certPath,
        [string]$keyPath,
        [System.Security.Cryptography.RSA]$privateRSA = $null
    )
    # Export cert to PEM
    $rawCert = $cert.Export([System.Security.Cryptography.X509Certificates.X509ContentType]::Cert)
    $certBase64 = [System.Convert]::ToBase64String($rawCert)
    $certLines = $certBase64 -split '(.{64})' | Where-Object { $_ -ne '' }
    $certPem = "-----BEGIN CERTIFICATE-----`n" + ($certLines -join "`n") + "`n-----END CERTIFICATE-----`n"
    Set-Content -Path $certPath -Value $certPem -Encoding ascii

    # Obtain RSA private key. If a privateRSA instance was provided by the caller use it
    # (this avoids relying on PFX export/import semantics which can vary across runtimes).
    $rsa = $null
    $reloadCert = $null
    try {
        if ($null -ne $privateRSA) {
            $rsa = $privateRSA
        }
        else {
            # Export as PFX and reload with Exportable flag so we can export the private key
            $pfxBytes = $cert.Export([System.Security.Cryptography.X509Certificates.X509ContentType]::Pfx, '')
            $reloadCert = [System.Security.Cryptography.X509Certificates.X509Certificate2]::new($pfxBytes, '', [System.Security.Cryptography.X509Certificates.X509KeyStorageFlags]::Exportable)

            # Try the modern API first
            try { $rsa = $reloadCert.GetRSAPrivateKey() } catch { $rsa = $null }

            # Fallback: some runtimes expose PrivateKey which can export parameters
            if (-not $rsa -and $null -ne $reloadCert.PrivateKey) {
                try {
                    $privateKey = $reloadCert.PrivateKey
                    $rsaFallback = [System.Security.Cryptography.RSA]::Create()
                    $rsaFallback.ImportParameters($privateKey.ExportParameters($true))
                    $rsa = $rsaFallback
                }
                catch {
                    if ($rsaFallback -is [System.IDisposable]) { $rsaFallback.Dispose() }
                    $rsa = $null
                }
            }
        }

        if (-not $rsa) { throw "Certificate does not contain an RSA private key" }

        # Export private key to PEM (PKCS#8)
        $pkcs8 = $rsa.ExportPkcs8PrivateKey()
        $keyBase64 = [System.Convert]::ToBase64String($pkcs8)
        $pkcs8Lines = $keyBase64 -split '(.{64})' | Where-Object { $_ -ne '' }
        $keyPem = "-----BEGIN PRIVATE KEY-----`n" + ($pkcs8Lines -join "`n") + "`n-----END PRIVATE KEY-----`n"
        Set-Content -Path $keyPath -Value $keyPem -Encoding ascii
    }
    finally {
        # Only dispose RSA if we created it locally (i.e., privateRSA was not passed in)
        if ($null -eq $privateRSA) {
            if ($rsa -is [System.IDisposable]) { $rsa.Dispose() }
            if ($reloadCert -is [System.IDisposable]) { $reloadCert.Dispose() }
        }
    }
}

function Ensure-Certificates {
    param(
        [string]$cert_dir
    )
    
    $cert_name_prefix = "server"
    $cert_file_name = "${cert_name_prefix}.cert"
    $key_file_name = "${cert_name_prefix}.key"

    # Generate certificate and key file if they don't exist in the cert directory
    $local_cert_file = Join-Path $LOCAL_CERT_DIR $cert_file_name
    $local_key_file = Join-Path $LOCAL_CERT_DIR $key_file_name
    
    if (-not (Test-Path $local_cert_file) -or -not (Test-Path $local_key_file)) {
        New-Item -Path $LOCAL_CERT_DIR -ItemType Directory -Force | Out-Null
        
        Write-Host "Generating SSL certificates in $LOCAL_CERT_DIR..."
        try {
            $openssl = Get-Command openssl -ErrorAction SilentlyContinue
            if ($openssl) {
                & openssl req -x509 -nodes -days 365 -newkey rsa:2048 `
                    -keyout $local_key_file `
                    -out $local_cert_file `
                    -subj "/O=WSO2/OU=Thunder/CN=localhost" 2>$null
                if ($LASTEXITCODE -ne 0) {
                    throw "Error generating SSL certificates: OpenSSL failed with exit code $LASTEXITCODE"
                }
                Write-Host "Certificates generated successfully in $LOCAL_CERT_DIR using OpenSSL."
            }
            else {
                Write-Host "OpenSSL not found - generating self-signed cert using .NET CertificateRequest (no UI)."
                # Use .NET CertificateRequest to avoid CertEnroll / smartcard enrollment UI issues.
                try {
                    $rsa = [System.Security.Cryptography.RSA]::Create(2048)

                    $subjectName = New-Object System.Security.Cryptography.X509Certificates.X500DistinguishedName("CN=localhost, O=WSO2, OU=Thunder")
                    $certReq = New-Object System.Security.Cryptography.X509Certificates.CertificateRequest($subjectName, $rsa, [System.Security.Cryptography.HashAlgorithmName]::SHA256, [System.Security.Cryptography.RSASignaturePadding]::Pkcs1)

                    # Add standard server usages
                    $basicConstraints = New-Object System.Security.Cryptography.X509Certificates.X509BasicConstraintsExtension($false, $false, 0, $false)
                    $ku1 = [int][System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]::DigitalSignature
                    $ku2 = [int][System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]::KeyEncipherment
                    $kuFlags = $ku1 -bor $ku2
                    $keyUsage = New-Object System.Security.Cryptography.X509Certificates.X509KeyUsageExtension([System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]$kuFlags, $true)
                    $ekuCollection = New-Object System.Security.Cryptography.OidCollection
                    $serverAuthOid = New-Object System.Security.Cryptography.Oid("1.3.6.1.5.5.7.3.1")
                    [void]$ekuCollection.Add($serverAuthOid)
                    $eku = New-Object System.Security.Cryptography.X509Certificates.X509EnhancedKeyUsageExtension($ekuCollection, $false)

                    $certReq.CertificateExtensions.Add($basicConstraints)
                    $certReq.CertificateExtensions.Add($keyUsage)
                    $certReq.CertificateExtensions.Add($eku)

                    $notBefore = (Get-Date).AddDays(-1)
                    $notAfter = (Get-Date).AddYears(1)

                    $cert = $certReq.CreateSelfSigned($notBefore, $notAfter)

                    # Ensure the generated certificate has the private key associated. Use CopyWithPrivateKey
                    # so that when we export the PFX it includes the private key and can be reloaded as exportable.
                    # Use the RSA extension helper to avoid overload resolution issues in PowerShell.
                    try {
                        $certWithKey = [System.Security.Cryptography.X509Certificates.RSACertificateExtensions]::CopyWithPrivateKey($cert, $rsa)
                    }
                    catch {
                        try {
                            $certWithKey = [System.Security.Cryptography.X509Certificates.RSACertificateExtensions]::CopyWithPrivateKey(([System.Security.Cryptography.X509Certificates.X509Certificate2]::new($cert.RawData)), $rsa)
                        }
                        catch {
                            throw "Failed to associate private key with certificate: $_"
                        }
                    }

                    # Export and reload as exportable so we can extract the private key bytes
                    $pfxBytes = $certWithKey.Export([System.Security.Cryptography.X509Certificates.X509ContentType]::Pfx, '')
                    $exportableCert = [System.Security.Cryptography.X509Certificates.X509Certificate2]::new($pfxBytes, '', [System.Security.Cryptography.X509Certificates.X509KeyStorageFlags]::Exportable)

                    # Pass the RSA instance used to sign the certificate to the exporter so it
                    # can directly export the private key (avoids re-import issues on some runtimes).
                    Export-CertificateAndKeyToPem -cert $exportableCert -certPath $local_cert_file -keyPath $local_key_file -privateRSA $rsa

                    if ($exportableCert -is [System.IDisposable]) { $exportableCert.Dispose() }
                    if ($certWithKey -is [System.IDisposable]) { $certWithKey.Dispose() }
                    if ($cert -is [System.IDisposable]) { $cert.Dispose() }
                    if ($rsa -is [System.IDisposable]) { $rsa.Dispose() }

                    Write-Host "Certificates generated successfully in $LOCAL_CERT_DIR using .NET CertificateRequest." 
                }
                catch {
                    throw "Error creating self-signed certificate using .NET APIs: $_"
                }
            }
        }
        catch {
            Write-Error "Error generating SSL certificates: $_"
            exit 1
        }
    }
    else {
        Write-Host "Certificates already exist in $LOCAL_CERT_DIR."
    }

    # Copy the generated certificates to the specified directory
    $cert_file = Join-Path $cert_dir $cert_file_name
    $key_file = Join-Path $cert_dir $key_file_name

    if (-not (Test-Path $cert_file) -or -not (Test-Path $key_file)) {
        New-Item -Path $cert_dir -ItemType Directory -Force | Out-Null
        
        Write-Host "Copying certificates to $cert_dir..."
        Copy-Item -Path $local_cert_file -Destination $cert_file -Force
        Copy-Item -Path $local_key_file -Destination $key_file -Force
        Write-Host "Certificates copied successfully to $cert_dir."
    }
    else {
        Write-Host "Certificates already exist in $cert_dir."
    }
}

function Run-Server {
    Write-Host "=== Ensuring server certificates exist ==="
    Ensure-Certificates -cert_dir (Join-Path $BACKEND_DIR $SECURITY_DIR)

    Write-Host "=== Ensuring sample app certificates exist ==="
    Ensure-Certificates -cert_dir $SAMPLE_APP_DIR

    Write-Host "Initializing databases..."
    Initialize-Databases

    # Kill processes on known ports
    function Kill-Port {
        param([int]$port)
        
        $processes = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess
        foreach ($process in $processes) {
            Stop-Process -Id $process -Force -ErrorAction SilentlyContinue
        }
    }

    Kill-Port $BACKEND_PORT

    Write-Host "=== Starting backend on https://localhost:$BACKEND_PORT ==="
    $env:BACKEND_PORT = $BACKEND_PORT
    
    Push-Location $BACKEND_DIR
    try {
        $backendProcess = Start-Process -FilePath "go" -ArgumentList "run", "." -PassThru -NoNewWindow
    }
    finally {
        Pop-Location
    }

    Write-Host ""
    Write-Host "🚀 Servers running:"
    Write-Host "👉 Backend : https://localhost:$BACKEND_PORT"
    Write-Host "📱 Frontend Apps:"
    Write-Host "   � Gate (Login/Register): https://localhost:$BACKEND_PORT/signin"
    Write-Host "   �️  Develop (Admin Console): https://localhost:$BACKEND_PORT/develop"
    Write-Host "Press Ctrl+C to stop."
    
    # Wait for user to press Ctrl+C
    try {
        while ($true) {
            Start-Sleep -Seconds 1
        }
    }
    catch [System.Management.Automation.PipelineStoppedException] {
        Write-Host "Stopping servers..."
        if ($backendProcess -and -not $backendProcess.HasExited) {
            Stop-Process -Id $backendProcess.Id -Force -ErrorAction SilentlyContinue
        }
    }
}

# Main script logic
switch ($Command) {
    "clean" {
        Clean
    }
    "clean_all" {
        Clean-All
    }
    "build_backend" {
        Build-Backend
        Package
    }
    "build_frontend" {
        Build-Frontend
    }
    "build_samples" {
        Build-Sample-App
        Package-Sample-App
    }
    "package_samples" {
        Package-Sample-App
    }
    "build" {
        Build-Backend
        Build-Frontend
        Package
        Build-Sample-App
        Package-Sample-App
    }
    "test_unit" {
        Test-Unit
    }
    "test_integration" {
        Test-Integration
    }
    "merge_coverage" {
        Merge-Coverage
    }
    "test" {
        Test-Unit
        Test-Integration
    }
    "run" {
        Run-Server
    }
    default {
        Write-Host "Usage: ./build.ps1 {clean|build|build_backend|build_frontend|test|run} [OS] [ARCH]"
        Write-Host ""
        Write-Host "  clean                    - Clean build artifacts"
        Write-Host "  clean_all                - Clean all build artifacts including distributions"
        Write-Host "  build                    - Build the complete Thunder application (backend + frontend + samples)"
        Write-Host "  build_backend            - Build only the Thunder backend server"
        Write-Host "  build_frontend           - Build only the frontend applications"
        Write-Host "  build_samples            - Build the sample applications"
        Write-Host "  test_unit                - Run unit tests with coverage"
        Write-Host "  test_integration         - Run integration tests"
        Write-Host "  merge_coverage           - Merge unit and integration test coverage reports"
        Write-Host "  test                     - Run all tests (unit and integration)"
        Write-Host "  run                      - Run the Thunder server for development"
        exit 1
    }
}
