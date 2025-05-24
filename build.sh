#!/bin/bash
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

set -e

VERSION_FILE=version.txt
BINARY_NAME=thunder
VERSION=$(cat "$VERSION_FILE")
PRODUCT_FOLDER=${BINARY_NAME}-${VERSION}

# Server ports
FRONTEND_PORT=9090
BACKEND_PORT=8090

# Directories
OUTPUT_DIR=target
BUILD_DIR=$OUTPUT_DIR/.build
BACKEND_BASE_DIR=backend
BACKEND_DIR=$BACKEND_BASE_DIR/cmd/server
REPOSITORY_DIR=$BACKEND_BASE_DIR/cmd/server/repository
SERVER_SCRIPTS_DIR=$BACKEND_BASE_DIR/scripts
SERVER_DB_SCRIPTS_DIR=$BACKEND_BASE_DIR/dbscripts
SECURITY_DIR=repository/resources/security
FRONTEND_BASE_DIR=frontend
GATE_APP_DIR=apps/gate
PACKAGES_DIR=packages
FRONTEND_GATE_APP_DIR=$FRONTEND_BASE_DIR/$GATE_APP_DIR
FRONTEND_PACKAGES=$FRONTEND_BASE_DIR/$PACKAGES_DIR
SAMPLE_BASE_DIR=samples
SAMPLE_OAUTH_APP_DIR=$SAMPLE_BASE_DIR/apps/oauth

GOOS=darwin
GOARCH=amd64

function clean() {
    echo "Cleaning build artifacts..."
    rm -rf "$OUTPUT_DIR"
}

function build_backend() {
    echo "Building Go backend..."
    mkdir -p "$BUILD_DIR"

    GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -C "$BACKEND_BASE_DIR" \
    -x -ldflags "-X \"main.version=$VERSION\" \
    -X \"main.buildDate=$$(date -u '+%Y-%m-%d %H:%M:%S UTC')\"" \
    -o "../$BUILD_DIR/$BINARY_NAME" ./cmd/server
}

function build_frontend() {
    echo "Building Next frontend apps..."
    pnpm i
    pnpm --filter gate build
}

function prepare_backend_for_packaging() {
    echo "Copying backend artifacts..."

    cp "$BUILD_DIR/$BINARY_NAME" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    cp -r "$REPOSITORY_DIR" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    cp "$VERSION_FILE" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    cp -r "$SERVER_SCRIPTS_DIR" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    cp -r "$SERVER_DB_SCRIPTS_DIR" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    mkdir -p "$OUTPUT_DIR/$PRODUCT_FOLDER/$SECURITY_DIR"

    echo "=== Ensuring server certificates exist in the distribution ==="
    ensure_certificates "$OUTPUT_DIR/$PRODUCT_FOLDER/$SECURITY_DIR"
}

function prepare_frontend_for_packaging() {
    echo "Copying frontend artifacts..."

    mkdir -p "$OUTPUT_DIR/$PRODUCT_FOLDER/$GATE_APP_DIR"
    cp -r "$FRONTEND_GATE_APP_DIR/dist/"* "$OUTPUT_DIR/$PRODUCT_FOLDER/$GATE_APP_DIR"
}

function package() {
    echo "Packaging artifacts..."

    mkdir -p "$OUTPUT_DIR/$PRODUCT_FOLDER"

    prepare_backend_for_packaging
    prepare_frontend_for_packaging

    cp -r "start.sh" "$OUTPUT_DIR/$PRODUCT_FOLDER"

    echo "Creating zip file..."
    (cd "$OUTPUT_DIR" && zip -r "$PRODUCT_FOLDER.zip" "$PRODUCT_FOLDER")
    rm -rf "${OUTPUT_DIR:?}/$PRODUCT_FOLDER" "$BUILD_DIR"
}

function test_integration() {
    echo "Running integration tests..."
    go run -C ./tests/integration ./main.go
}

function ensure_certificates() {
    local cert_dir=$1
    local cert_name_prefix="server"
    local cert_file="$cert_dir/${cert_name_prefix}.cert"
    local key_file="$cert_dir/${cert_name_prefix}.key"

    if [[ ! -f "$cert_file" || ! -f "$key_file" ]]; then
        mkdir -p "$cert_dir"
        echo "Generating SSL certificates in $cert_dir..."
        OPENSSL_ERR=$(
            openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                -keyout "$key_file" \
                -out "$cert_file" \
                -subj "/O=WSO2/OU=Thunder/CN=localhost" \
                > /dev/null 2>&1
        )
        if [[ $? -ne 0 ]]; then
            echo "Error generating SSL certificates: $OPENSSL_ERR"
            exit 1
        fi
        echo "Certificates generated successfully in $cert_dir."
    else
        echo "Certificates already exist in $cert_dir."
    fi
}

function run() {
    echo "=== Cleaning build output ==="
    clean

    echo "=== Building backend ==="
    build_backend

    echo "=== Building frontend ==="
    build_frontend

    echo "=== Ensuring server certificates exist ==="
    ensure_certificates "$BACKEND_DIR/$SECURITY_DIR"

    echo "=== Ensuring portal certificates exist ==="
    ensure_certificates "$FRONTEND_GATE_APP_DIR"

    echo "=== Ensuring Sample app certificates exist ==="
    ensure_certificates "$SAMPLE_OAUTH_APP_DIR"

    # Kill known ports
    function kill_port() {
        local port=$1
        lsof -ti tcp:$port | xargs kill -9 2>/dev/null || true
    }

    kill_port $FRONTEND_PORT
    kill_port $BACKEND_PORT

    echo "=== Starting frontend on https://localhost:$FRONTEND_PORT ==="
    FRONTEND_PORT=$FRONTEND_PORT pnpm --filter gate start &
    FRONTEND_PID=$!

    echo "=== Starting backend on https://localhost:$BACKEND_PORT ==="
    BACKEND_PORT=$BACKEND_PORT go run -C "$BACKEND_DIR" . &
    BACKEND_PID=$!

    echo ""
    echo "🚀 Servers running:"
    echo "👉 Frontend: https://localhost:$FRONTEND_PORT"
    echo "👉 Backend : https://localhost:$BACKEND_PORT"
    echo "Press Ctrl+C to stop."

    trap 'echo -e "\nStopping servers..."; kill $FRONTEND_PID $BACKEND_PID; exit' SIGINT

    wait $FRONTEND_PID
    wait $BACKEND_PID
}

case "$1" in
    clean)
        clean
        ;;
    build)
        build_backend
        build_frontend
        package
        ;;
    test)
        test_integration
        ;;
    run)
        run
        ;;
    *)
        echo "Usage: ./build.sh {clean|build|test|run}"
        exit 1
        ;;
esac
