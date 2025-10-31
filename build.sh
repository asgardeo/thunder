#!/bin/bash
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

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# --- Set Default OS and the architecture --- 
# Auto-detect GO OS
DEFAULT_OS=$(go env GOOS 2>/dev/null)
if [ -z "$DEFAULT_OS" ]; then
  UNAME_OS="$(uname -s)"
  case "$UNAME_OS" in
    Darwin) DEFAULT_OS="darwin" ;;
    Linux) DEFAULT_OS="linux" ;;
    MINGW*|MSYS*|CYGWIN*) DEFAULT_OS="windows" ;;
    *) echo "Unsupported OS: $UNAME_OS"; exit 1 ;;
  esac
fi
# Auto-detect GO ARCH
DEFAULT_ARCH=$(go env GOARCH 2>/dev/null)
if [ -z "$DEFAULT_ARCH" ]; then
  UNAME_ARCH="$(uname -m)"
  case "$UNAME_ARCH" in
    x86_64|amd64) DEFAULT_ARCH="amd64" ;;
    arm64|aarch64) DEFAULT_ARCH="arm64" ;;
    *) echo "Unsupported architecture: $UNAME_ARCH"; exit 1 ;;
  esac
fi

GO_OS=${2:-$DEFAULT_OS}
GO_ARCH=${3:-$DEFAULT_ARCH}

echo "Using GO OS: $GO_OS and ARCH: $GO_ARCH"

SAMPLE_DIST_NODE_VERSION=node18
SAMPLE_DIST_OS=${2:-$DEFAULT_OS}
SAMPLE_DIST_ARCH=${3:-$DEFAULT_ARCH}

# Transform OS for node packaging executor
if [ "$SAMPLE_DIST_OS" = "darwin" ]; then
    SAMPLE_DIST_OS=macos
elif [ "$SAMPLE_DIST_OS" = "windows" ]; then
    SAMPLE_DIST_OS="win"
fi

if [ "$SAMPLE_DIST_ARCH" = "amd64" ]; then
    SAMPLE_DIST_ARCH=x64
fi

# --- Thunder Package Distribution details ---
GO_PACKAGE_OS=$GO_OS
GO_PACKAGE_ARCH=$GO_ARCH

# Normalize OS name for distribution packaging
if [ "$GO_OS" = "darwin" ]; then
    GO_PACKAGE_OS=macos
elif [ "$GO_OS" = "windows" ]; then
    GO_PACKAGE_OS="win"
fi

if [ "$GO_ARCH" = "amd64" ]; then
    GO_PACKAGE_ARCH=x64
fi

VERSION_FILE=version.txt
VERSION=$(cat "$VERSION_FILE")
THUNDER_VERSION=${VERSION}
if [[ $THUNDER_VERSION == v* ]]; then
  THUNDER_VERSION="${THUNDER_VERSION#v}"
fi
BINARY_NAME=thunder
PRODUCT_FOLDER=${BINARY_NAME}-${THUNDER_VERSION}-${GO_PACKAGE_OS}-${GO_PACKAGE_ARCH}

# --- Sample App Distribution details ---
SAMPLE_PACKAGE_OS=$SAMPLE_DIST_OS
SAMPLE_PACKAGE_ARCH=$SAMPLE_DIST_ARCH

SAMPLE_APP_SERVER_BINARY_NAME=server
SAMPLE_APP_VERSION=$(grep -o '"version": *"[^"]*"' samples/apps/oauth/package.json | sed 's/"version": *"\(.*\)"/\1/')
SAMPLE_APP_FOLDER="sample-app-${SAMPLE_APP_VERSION}-${SAMPLE_PACKAGE_OS}-${SAMPLE_PACKAGE_ARCH}"

# Server ports
BACKEND_PORT=8090

# Directories
TARGET_DIR=target
OUTPUT_DIR=$TARGET_DIR/out
DIST_DIR=$TARGET_DIR/dist
BUILD_DIR=$OUTPUT_DIR/.build
LOCAL_CERT_DIR=$OUTPUT_DIR/.cert
BACKEND_BASE_DIR=backend
BACKEND_DIR=$BACKEND_BASE_DIR/cmd/server
REPOSITORY_DIR=$BACKEND_BASE_DIR/cmd/server/repository
REPOSITORY_DB_DIR=$REPOSITORY_DIR/database
SERVER_SCRIPTS_DIR=$BACKEND_BASE_DIR/scripts
SERVER_DB_SCRIPTS_DIR=$BACKEND_BASE_DIR/dbscripts
SECURITY_DIR=repository/resources/security
FRONTEND_BASE_DIR=frontend
GATE_APP_DIST_DIR=apps/gate
DEVELOP_APP_DIST_DIR=apps/develop
FRONTEND_GATE_APP_SOURCE_DIR=$FRONTEND_BASE_DIR/apps/thunder-gate
FRONTEND_DEVELOP_APP_SOURCE_DIR=$FRONTEND_BASE_DIR/apps/thunder-develop
SAMPLE_BASE_DIR=samples
SAMPLE_APP_DIR=$SAMPLE_BASE_DIR/apps/oauth
SAMPLE_APP_SERVER_DIR=$SAMPLE_APP_DIR/server

function get_coverage_exclusion_pattern() {
    # Read exclusion patterns (full package paths) from .excludecoverage file
    local coverage_exclude_file
    
    # Check if we're already in the backend directory or need to use relative path
    if [ -f ".excludecoverage" ]; then
        coverage_exclude_file=".excludecoverage"
    elif [ -f "$SCRIPT_DIR/$BACKEND_BASE_DIR/.excludecoverage" ]; then
        coverage_exclude_file="$SCRIPT_DIR/$BACKEND_BASE_DIR/.excludecoverage"
    else
        echo "" >&2
        return
    fi
    
    # Read non-comment, non-empty lines and join with '|' for grep (exact package path matching)
    local pattern=$(awk '!/^#/ && NF {if(count++)printf "|"; printf "%s", $0}' "$coverage_exclude_file")
    echo "$pattern"
}

function clean_all() {
    echo "================================================================"
    echo "Cleaning all build artifacts..."
    rm -rf "$TARGET_DIR"

    echo "Removing certificates in the $BACKEND_DIR/$SECURITY_DIR"
    rm -rf "$BACKEND_DIR/$SECURITY_DIR"

    echo "Removing certificates in the $SAMPLE_APP_DIR"
    rm -f "$SAMPLE_APP_DIR/server.cert"
    rm -f "$SAMPLE_APP_DIR/server.key"
    rm -f "$SAMPLE_APP_SERVER_DIR/server.cert"
    rm -f "$SAMPLE_APP_SERVER_DIR/server.key"
    echo "================================================================"
}

function clean() {
    echo "================================================================"
    echo "Cleaning build artifacts..."
    rm -rf "$OUTPUT_DIR"

    echo "Removing certificates in the $BACKEND_DIR/$SECURITY_DIR"
    rm -rf "$BACKEND_DIR/$SECURITY_DIR"

    echo "Removing certificates in the $SAMPLE_APP_DIR"
    rm -f "$SAMPLE_APP_DIR/server.cert"
    rm -f "$SAMPLE_APP_DIR/server.key"
    echo "================================================================"
}

function build_backend() {
    echo "================================================================"
    echo "Building Go backend..."
    mkdir -p "$BUILD_DIR"

    # Set binary name with .exe extension for Windows
    local output_binary="$BINARY_NAME"
    if [ "$GO_OS" = "windows" ]; then
        output_binary="${BINARY_NAME}.exe"
    fi

    # Check if coverage build is requested via ENABLE_COVERAGE environment variable
    local build_flags="-x"
    if [ "$ENABLE_COVERAGE" = "true" ]; then
        echo "Building with coverage instrumentation enabled..."
        # Build coverage package list
        cd "$BACKEND_BASE_DIR" || exit 1
        local exclude_pattern=$(get_coverage_exclusion_pattern)
        local coverpkg
        if [ -n "$exclude_pattern" ]; then
            echo "Excluding coverage for patterns: $exclude_pattern"
            coverpkg=$(go list ./... | grep -v -E "$exclude_pattern" | tr '\n' ',' | sed 's/,$//')
        else
            coverpkg=$(go list ./... | tr '\n' ',' | sed 's/,$//')
        fi
        cd "$SCRIPT_DIR" || exit 1
        build_flags="$build_flags -cover -coverpkg=$coverpkg"
    fi

    GOOS=$GO_OS GOARCH=$GO_ARCH CGO_ENABLED=0 go build -C "$BACKEND_BASE_DIR" \
    $build_flags -ldflags "-X \"main.version=$VERSION\" \
    -X \"main.buildDate=$$(date -u '+%Y-%m-%d %H:%M:%S UTC')\"" \
    -o "../$BUILD_DIR/$output_binary" ./cmd/server

    echo "Initializing databases..."
    initialize_databases true
    echo "================================================================"
}

function initialize_databases() {
    echo "================================================================"
    local override=$1
    if [[ -z "$override" ]]; then
        override=false
    fi

    echo "Initializing SQLite databases..."

    mkdir -p "$REPOSITORY_DB_DIR"

    db_files=("thunderdb.db" "runtimedb.db")
    script_paths=("thunderdb/sqlite.sql" "runtimedb/sqlite.sql")

    for ((i = 0; i < ${#db_files[@]}; i++)); do
        db_file="${db_files[$i]}"
        script_rel_path="${script_paths[$i]}"
        db_path="$REPOSITORY_DB_DIR/$db_file"
        script_path="$SERVER_DB_SCRIPTS_DIR/$script_rel_path"

        if [[ -f "$script_path" ]]; then
            if [[ -f "$db_path" ]]; then
                if $override; then
                    echo " - Removing existing $db_file as override is true"
                    rm "$db_path"
                else
                    echo " ! Skipping $db_file: DB already exists. Delete the existing and re-run to recreate."
                    continue
                fi
            fi

            echo " - Creating $db_file using $script_path"
            sqlite3 "$db_path" < "$script_path"
        else
            echo " ! Skipping $db_file: SQL script not found at $script_path"
        fi
    done

    echo "SQLite database initialization complete."
    echo "================================================================"
}

function build_frontend() {
    echo "================================================================"
    echo "Building frontend apps..."
    
    # Check if pnpm is installed, if not install it
    if ! command -v pnpm >/dev/null 2>&1; then
        echo "pnpm not found, installing..."
        npm install -g pnpm
    fi
    
    # Navigate to frontend directory and install dependencies
    cd "$FRONTEND_BASE_DIR" || exit 1
    echo "Installing frontend dependencies..."
    pnpm install
    
    echo "Building frontend applications & packages..."
    pnpm build
    
    # Return to script directory
    cd "$SCRIPT_DIR" || exit 1
    echo "================================================================"
}

function prepare_backend_for_packaging() {
    echo "================================================================"
    echo "Copying backend artifacts..."

    # Use appropriate binary name based on OS
    local binary_name="$BINARY_NAME"
    if [ "$GO_OS" = "windows" ]; then
        binary_name="${BINARY_NAME}.exe"
    fi

    cp "$BUILD_DIR/$binary_name" "$DIST_DIR/$PRODUCT_FOLDER/"
    cp -r "$REPOSITORY_DIR" "$DIST_DIR/$PRODUCT_FOLDER/"
    cp "$VERSION_FILE" "$DIST_DIR/$PRODUCT_FOLDER/"
    cp -r "$SERVER_SCRIPTS_DIR" "$DIST_DIR/$PRODUCT_FOLDER/"
    cp -r "$SERVER_DB_SCRIPTS_DIR" "$DIST_DIR/$PRODUCT_FOLDER/"
    mkdir -p "$DIST_DIR/$PRODUCT_FOLDER/$SECURITY_DIR"

    echo "=== Ensuring server certificates exist in the distribution ==="
    ensure_certificates "$DIST_DIR/$PRODUCT_FOLDER/$SECURITY_DIR"
    echo "================================================================"
}

function prepare_frontend_for_packaging() {
    echo "================================================================"
    echo "Copying frontend artifacts..."

    mkdir -p "$DIST_DIR/$PRODUCT_FOLDER/$GATE_APP_DIST_DIR"
    mkdir -p "$DIST_DIR/$PRODUCT_FOLDER/$DEVELOP_APP_DIST_DIR"

    # Copy gate app build output
    if [ -d "$FRONTEND_GATE_APP_SOURCE_DIR/dist" ]; then
        echo "Copying Gate app build output..."
        shopt -s dotglob
        cp -r "$FRONTEND_GATE_APP_SOURCE_DIR/dist/"* "$DIST_DIR/$PRODUCT_FOLDER/$GATE_APP_DIST_DIR"
        shopt -u dotglob
    else
        echo "Warning: Gate app build output not found at $FRONTEND_GATE_APP_SOURCE_DIR/dist"
    fi
    
    # Copy develop app build output
    if [ -d "$FRONTEND_DEVELOP_APP_SOURCE_DIR/dist" ]; then
        echo "Copying Develop app build output..."
        shopt -s dotglob
        cp -r "$FRONTEND_DEVELOP_APP_SOURCE_DIR/dist/"* "$DIST_DIR/$PRODUCT_FOLDER/$DEVELOP_APP_DIST_DIR"
        shopt -u dotglob
    else
        echo "Warning: Develop app build output not found at $FRONTEND_DEVELOP_APP_SOURCE_DIR/dist"
    fi

    echo "================================================================"
}

function package() {
    echo "================================================================"
    echo "Packaging backend & frontend artifacts..."

    mkdir -p "$DIST_DIR/$PRODUCT_FOLDER"

    prepare_frontend_for_packaging
    prepare_backend_for_packaging

    # Copy the appropriate startup script based on the target OS
    if [ "$GO_OS" = "windows" ]; then
            echo "Including Windows start script (start.ps1)..."
            cp -r "start.ps1" "$DIST_DIR/$PRODUCT_FOLDER"
    else
        echo "Including Unix start script (start.sh)..."
        cp -r "start.sh" "$DIST_DIR/$PRODUCT_FOLDER"
    fi

    echo "Creating zip file..."
    (cd "$DIST_DIR" && zip -r "$PRODUCT_FOLDER.zip" "$PRODUCT_FOLDER")
    rm -rf "${DIST_DIR:?}/$PRODUCT_FOLDER" "$BUILD_DIR"
    echo "================================================================"
}

function build_sample_app() {
    echo "================================================================"
    echo "Building sample app..."

    # Ensure certificate exists for the sample app
    echo "=== Ensuring sample app certificates exist ==="
    ensure_certificates "$SAMPLE_APP_DIR"
    
    # Build the application
    cd "$SAMPLE_APP_DIR" || exit 1
    echo "Installing dependencies..."
    npm install
    
    echo "Building the app..."
    npm run build
    
    cd - || exit 1
    
    echo "Sample app built successfully."
    echo "================================================================"
}

function package_sample_app() {
    echo "================================================================"
    echo "Copying sample artifacts..."

    # Use appropriate binary name based on OS
    local binary_name="$SAMPLE_APP_SERVER_BINARY_NAME"
    local executable_name="$SAMPLE_APP_SERVER_BINARY_NAME-$SAMPLE_DIST_OS-$SAMPLE_DIST_ARCH"

    if [ "$SAMPLE_DIST_OS" = "win" ]; then
        binary_name="${SAMPLE_APP_SERVER_BINARY_NAME}.exe"
        executable_name="${SAMPLE_APP_SERVER_BINARY_NAME}-${SAMPLE_DIST_OS}-${SAMPLE_DIST_ARCH}.exe"
    fi
    
    mkdir -p "$DIST_DIR/$SAMPLE_APP_FOLDER"
    
    # Copy the built app files
    cp -r "$SAMPLE_APP_SERVER_DIR/app" "$DIST_DIR/$SAMPLE_APP_FOLDER/"

    cd $SAMPLE_APP_SERVER_DIR

    mkdir -p "executables"

    npx pkg . -t $SAMPLE_DIST_NODE_VERSION-$SAMPLE_DIST_OS-$SAMPLE_DIST_ARCH -o executables/$SAMPLE_APP_SERVER_BINARY_NAME-$SAMPLE_DIST_OS-$SAMPLE_DIST_ARCH

    cd $SCRIPT_DIR

    # Copy the server binary
    cp "$SAMPLE_APP_SERVER_DIR/executables/$executable_name" "$DIST_DIR/$SAMPLE_APP_FOLDER/$binary_name"

    # Ensure the certificates exist in the sample app directory
    echo "=== Ensuring certificates exist in the sample distribution ==="
    ensure_certificates "$DIST_DIR/$SAMPLE_APP_FOLDER"

    # Copy the appropriate startup script based on the target OS
    if [ "$SAMPLE_DIST_OS" = "win" ]; then
            echo "Including Windows start script (start.ps1)..."
            cp -r "$SAMPLE_APP_SERVER_DIR/start.ps1" "$DIST_DIR/$SAMPLE_APP_FOLDER"
    else
        echo "Including Unix start script (start.sh)..."
        cp -r "$SAMPLE_APP_SERVER_DIR/start.sh" "$DIST_DIR/$SAMPLE_APP_FOLDER"
    fi

    echo "Creating zip file..."
    (cd "$DIST_DIR" && zip -r "$SAMPLE_APP_FOLDER.zip" "$SAMPLE_APP_FOLDER")
    rm -rf "${DIST_DIR:?}/$SAMPLE_APP_FOLDER"
    
    echo "Sample app packaged successfully as $DIST_DIR/$SAMPLE_APP_FOLDER.zip"
    echo "================================================================"
}

function test_unit() {
    echo "================================================================"
    echo "Running unit tests with coverage..."
    cd "$BACKEND_BASE_DIR" || exit 1
    
    # Build coverage package list
    local exclude_pattern=$(get_coverage_exclusion_pattern)
    local coverpkg
    if [ -n "$exclude_pattern" ]; then
        echo "Excluding coverage for patterns: $exclude_pattern"
        coverpkg=$(go list ./... | grep -v -E "$exclude_pattern" | tr '\n' ',' | sed 's/,$//')
    else
        echo "No exclusion patterns found, including all packages"
        coverpkg=$(go list ./... | tr '\n' ',' | sed 's/,$//')
    fi
    
    # Run gotestsum if available, otherwise fallback to go test
    if command -v gotestsum &> /dev/null; then
        echo "Running unit tests with coverage using gotestsum..."
        gotestsum -- -v -coverprofile=coverage_unit.out -covermode=atomic -coverpkg="$coverpkg" ./... || { echo "There are unit test failures."; exit 1; }
    else
        echo "Running unit tests with coverage using go test..."
        go test -v -coverprofile=coverage_unit.out -covermode=atomic -coverpkg="$coverpkg" ./... || { echo "There are unit test failures."; exit 1; }
    fi
    
    echo "Unit test coverage profile generated in: backend/coverage_unit.out"
    
    # Generate HTML coverage report for unit tests
    go tool cover -html=coverage_unit.out -o=coverage_unit.html
    echo "Unit test coverage HTML report generated in: backend/coverage_unit.html"
    
    # Display unit test coverage summary
    echo ""
    echo "================================================================"
    echo "Unit Test Coverage Summary:"
    go tool cover -func=coverage_unit.out | tail -n 1
    echo "================================================================"
    echo ""
    
    cd "$SCRIPT_DIR" || exit 1
    echo "================================================================"
}

function test_integration() {
    echo "================================================================"
    echo "Running integration tests..."
    cd "$SCRIPT_DIR" || exit 1
    
    # Set up coverage directory for integration tests
    local coverage_dir="$(pwd)/$OUTPUT_DIR/.test/integration"
    mkdir -p "$coverage_dir"
    
    # Export coverage directory for the server binary to use
    export GOCOVERDIR="$coverage_dir"
    
    echo "Coverage data will be collected in: $coverage_dir"
    go run -C ./tests/integration ./main.go
    test_exit_code=$?
    
    # Process coverage data if tests passed or failed
    if [ -d "$coverage_dir" ] && [ "$(ls -A $coverage_dir 2>/dev/null)" ]; then
        echo "================================================================"
        echo "Processing integration test coverage..."
        
        # Convert binary coverage data to text format
        cd "$BACKEND_BASE_DIR" || exit 1
        go tool covdata textfmt -i="$coverage_dir" -o="../$TARGET_DIR/coverage_integration.out"
        echo "Integration test coverage report generated in: $TARGET_DIR/coverage_integration.out"
        
        # Generate HTML coverage report
        go tool cover -html="../$TARGET_DIR/coverage_integration.out" -o="../$TARGET_DIR/coverage_integration.html"
        echo "Integration test coverage HTML report generated in: $TARGET_DIR/coverage_integration.html"
        
        # Display coverage summary
        echo ""
        echo "================================================================"
        echo "Coverage Summary:"
        go tool cover -func="../$TARGET_DIR/coverage_integration.out" | tail -n 1
        echo "================================================================"
        echo ""
        
        cd "$SCRIPT_DIR" || exit 1
    else
        echo "================================================================"
        echo "No coverage data collected"
    fi
    
    # Exit with the test exit code
    if [ $test_exit_code -ne 0 ]; then
        echo "================================================================"
        echo "Integration tests failed with exit code: $test_exit_code"
        exit $test_exit_code
    fi
    
    echo "================================================================"
}

function merge_coverage() {
    echo "================================================================"
    echo "Merging coverage reports..."
    cd "$SCRIPT_DIR" || exit 1
    
    local unit_coverage="$BACKEND_BASE_DIR/coverage_unit.out"
    local integration_coverage="$TARGET_DIR/coverage_integration.out"
    local combined_coverage="$TARGET_DIR/coverage_combined.out"
    
    # Check if both coverage files exist
    if [ ! -f "$unit_coverage" ]; then
        echo "Warning: Unit test coverage file not found at $unit_coverage"
        echo "Skipping coverage merge."
        return 0
    fi
    
    if [ ! -f "$integration_coverage" ]; then
        echo "Warning: Integration test coverage file not found at $integration_coverage"
        echo "Skipping coverage merge."
        return 0
    fi
    
    echo "Merging unit and integration test coverage..."
    
    # Get the mode from the first file and write to combined coverage
    head -n 1 "$unit_coverage" > "$combined_coverage"
    
    # Combine both files (skip mode lines) and merge overlapping coverage
    { tail -n +2 "$unit_coverage"; tail -n +2 "$integration_coverage"; } | \
        awk '
        {
            key = $1 " " $2
            if (!(key in lines)) {
                lines[key] = $0
                count[key] = $3
            } else {
                # For duplicate entries, take the maximum count
                if ($3 > count[key]) {
                    count[key] = $3
                    lines[key] = $1 " " $2 " " $3
                }
            }
        }
        END {
            for (key in lines) {
                print lines[key]
            }
        }
        ' | sort >> "$combined_coverage"
    
    echo "Combined coverage report generated in: $combined_coverage"
    
    # Generate HTML coverage report for combined coverage
    cd "$BACKEND_BASE_DIR" || exit 1
    go tool cover -html="../$combined_coverage" -o="../$TARGET_DIR/coverage_combined.html"
    echo "Combined coverage HTML report generated in: $TARGET_DIR/coverage_combined.html"
    
    # Display combined coverage summary
    echo ""
    echo "================================================================"
    echo "Combined Test Coverage Summary:"
    go tool cover -func="../$combined_coverage" | tail -n 1
    echo "================================================================"
    echo ""
    
    cd "$SCRIPT_DIR" || exit 1
    echo "================================================================"
}

function ensure_certificates() {
    local cert_dir=$1
    local cert_name_prefix="server"
    local cert_file_name="${cert_name_prefix}.cert"
    local key_file_name="${cert_name_prefix}.key"

    # Generate certificate and key file if don't exists in the cert directory
    local local_cert_file="${LOCAL_CERT_DIR}/${cert_file_name}"
    local local_key_file="${LOCAL_CERT_DIR}/${key_file_name}"
    if [[ ! -f "$local_cert_file" || ! -f "$local_key_file" ]]; then
        mkdir -p "$LOCAL_CERT_DIR"
        echo "Generating SSL certificates in $LOCAL_CERT_DIR..."
        OPENSSL_ERR=$(
            openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                -keyout "$local_key_file" \
                -out "$local_cert_file" \
                -subj "/O=WSO2/OU=Thunder/CN=localhost" \
                > /dev/null 2>&1
        )
        if [[ $? -ne 0 ]]; then
            echo "Error generating SSL certificates: $OPENSSL_ERR"
            exit 1
        fi
        echo "Certificates generated successfully in $LOCAL_CERT_DIR."
    else
        echo "Certificates already exist in $LOCAL_CERT_DIR."
    fi

    # Copy the generated certificates to the specified directory
    local cert_file="$cert_dir/${cert_file_name}"
    local key_file="$cert_dir/${key_file_name}"

    if [[ ! -f "$cert_file" || ! -f "$key_file" ]]; then
        mkdir -p "$cert_dir"
        echo "Copying certificates to $cert_dir..."
        cp "$local_cert_file" "$cert_file"
        cp "$local_key_file" "$key_file"
        echo "Certificates copied successfully to $cert_dir."
    else
        echo "Certificates already exist in $cert_dir."
    fi
}

function run() {
    echo "=== Ensuring server certificates exist ==="
    ensure_certificates "$BACKEND_DIR/$SECURITY_DIR"

    echo "=== Ensuring sample app certificates exist ==="
    ensure_certificates "$SAMPLE_APP_DIR"

    echo "Initializing databases..."
    initialize_databases

    # Kill known ports
    function kill_port() {
        local port=$1
        lsof -ti tcp:$port | xargs kill -9 2>/dev/null || true
    }

    kill_port $BACKEND_PORT

    echo "=== Starting backend on https://localhost:$BACKEND_PORT ==="
    BACKEND_PORT=$BACKEND_PORT go run -C "$BACKEND_DIR" . &
    BACKEND_PID=$!

    echo ""
    echo "🚀 Servers running:"
    echo "👉 Backend : https://localhost:$BACKEND_PORT"
    echo "📱 Frontend Apps:"
    echo "   🚪 Gate (Login/Register): https://localhost:$BACKEND_PORT/signin"
    echo "   🛠️  Develop (Admin Console): https://localhost:$BACKEND_PORT/develop"
    echo "Press Ctrl+C to stop."

    trap 'echo -e "\nStopping servers..."; kill $BACKEND_PID; exit' SIGINT

    wait $BACKEND_PID
}

case "$1" in
    clean)
        clean
        ;;
    clean_all)
        clean_all
        ;;
    build_backend)
        build_backend
        package
        ;;
    build_frontend)
        build_frontend
        ;;
    build_samples)
        build_sample_app
        package_sample_app
        ;;
    package_samples)
        package_sample_app
        ;;
    build)
        build_backend
        build_frontend
        package
        build_sample_app
        package_sample_app
        ;;
    test_unit)
        test_unit
        ;;
    test_integration)
        test_integration
        ;;
    merge_coverage)
        merge_coverage
        ;;
    test)
        test_unit
        test_integration
        ;;
    run)
        run
        ;;
    *)
        echo "Usage: ./build.sh {clean|build|build_backend|build_frontend|test|run} [OS] [ARCH]"
        echo ""
        echo "  clean                    - Clean build artifacts"
        echo "  clean_all                - Clean all build artifacts including distributions"
        echo "  build                    - Build the complete Thunder application (backend + frontend + samples)"
        echo "  build_backend            - Build only the Thunder backend server"
        echo "  build_frontend           - Build only the Next.js frontend applications"
        echo "  build_samples            - Build the sample applications"
        echo "  test_unit                - Run unit tests with coverage"
        echo "  test_integration         - Run integration tests"
        echo "  merge_coverage           - Merge unit and integration test coverage reports"
        echo "  test                     - Run all tests (unit and integration)"
        echo "  run                      - Run the Thunder server for development"
        exit 1
        ;;
esac
