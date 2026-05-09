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

# Default settings
PRODUCT_NAME="ThunderID"
PRODUCT_NAME_LOWERCASE="$(echo "$PRODUCT_NAME" | tr '[:upper:]' '[:lower:]')"
BINARY_NAME="${PRODUCT_NAME_LOWERCASE}"
BACKEND_PORT=${BACKEND_PORT:-8090}
DEBUG_PORT=${DEBUG_PORT:-2345}
DEBUG_MODE=${DEBUG_MODE:-false}
WITH_CONSENT=${WITH_CONSENT:-true}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --debug)
            DEBUG_MODE=true
            shift
            ;;
        --debug-port)
            DEBUG_PORT="$2"
            shift 2
            ;;
        --port)
            BACKEND_PORT="$2"
            shift 2
            ;;
        --without-consent)
            WITH_CONSENT=false
            shift
            ;;
        --help)
            echo "${PRODUCT_NAME} Server Startup Script"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --debug              Enable debug mode with remote debugging"
            echo "  --port PORT          Set application port (default: 8090)"
            echo "  --debug-port PORT    Set debug port (default: 2345)"
            echo "  --without-consent    Disable the bundled consent server"
            echo "  --help               Show this help message"
            echo ""
            echo "First-Time Setup:"
            echo "  For initial setup, use the setup script:"
            echo "    ./setup.sh"
            echo ""
            echo "  Then start the server normally:"
            echo "    ./start.sh"
            echo ""
            echo "Examples:"
            echo "  $0                   Start server normally"
            echo "  $0 --debug           Start in debug mode"
            echo "  $0 --port 9090       Start on custom port"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

set -e  # Exit immediately if a command exits with a non-zero status

# Check for port conflicts
check_port() {
    local port=$1
    local port_name=$2
    if lsof -ti tcp:$port >/dev/null 2>&1; then
        echo ""
        echo "❌ Port $port is already in use"
        echo "   $port_name cannot start because another process is using port $port"
        echo ""
        echo "💡 To find the process using this port:"
        echo "   lsof -i tcp:$port"
        echo ""
        echo "💡 To stop the process:"
        echo "   kill -9 \$(lsof -ti tcp:$port)"
        echo ""
        exit 1
    fi
}

# Check if ports are available
check_port $BACKEND_PORT "${PRODUCT_NAME} server"
if [ "$DEBUG_MODE" = "true" ]; then
    check_port $DEBUG_PORT "Debug server"
fi

# Check if Delve is available for debug mode
if [ "$DEBUG_MODE" = "true" ]; then
    # Check for dlv in PATH
    if ! command -v dlv &> /dev/null; then
        echo "❌ Debug mode requires Delve debugger"
        echo ""
        echo "💡 Install Delve using:"
        echo "   go install github.com/go-delve/delve/cmd/dlv@latest"
        echo ""
        echo "🔧 Add Delve to PATH"
        echo ""
        echo "🔧 After installation, run: $0 --debug"
        exit 1
    fi
fi

# Cleanup function
CONSENT_PID=""
SERVER_PID=""
cleanup() {
    echo -e "\n🛑 Stopping server..."
    if [ -n "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    if [ -n "$CONSENT_PID" ]; then
        pkill -P $CONSENT_PID 2>/dev/null || true
        kill $CONSENT_PID 2>/dev/null || true
    fi
}
trap cleanup SIGINT SIGTERM EXIT

# Start consent server if enabled
CONSENT_SERVER_PORT="${CONSENT_SERVER_PORT:-9090}"
if [ "$WITH_CONSENT" = "true" ]; then
    CONSENT_SCRIPT="$(dirname "$0")/consent/start.sh"
    if [ ! -x "$CONSENT_SCRIPT" ]; then
        echo "Error: Consent server is enabled but consent/start.sh is missing or not executable"
        exit 1
    fi
    echo "Starting Consent Server..."
    (cd "$(dirname "$0")/consent" && ./start.sh) &
    CONSENT_PID=$!
    CONSENT_TIMEOUT=30
    CONSENT_ELAPSED=0
    while [ $CONSENT_ELAPSED -lt $CONSENT_TIMEOUT ]; do
        if ! kill -0 "$CONSENT_PID" 2>/dev/null; then
            echo "Error: Consent server process exited unexpectedly"
            exit 1
        fi
        if curl -s -f "http://localhost:${CONSENT_SERVER_PORT}/health/readiness" > /dev/null 2>&1; then
            echo "Consent server is ready"
            break
        fi
        sleep 1
        CONSENT_ELAPSED=$((CONSENT_ELAPSED + 1))
    done
    if [ $CONSENT_ELAPSED -ge $CONSENT_TIMEOUT ]; then
        echo "Error: Consent server failed to become ready within ${CONSENT_TIMEOUT}s"
        exit 1
    fi
fi

# Run the Server
if [ "$DEBUG_MODE" = "true" ]; then
    echo "⚡ Starting ${PRODUCT_NAME} Server in DEBUG mode..."
    echo "📝 Application will run on: https://localhost:$BACKEND_PORT"
    echo "🐛 Remote debugger will listen on: localhost:$DEBUG_PORT"
    echo ""
    echo "💡 Connect using remote debugging configuration:"
    echo "   Host: 127.0.0.1, Port: $DEBUG_PORT"
    echo ""

    # Run debugger
    dlv exec --listen=:$DEBUG_PORT --headless=true --api-version=2 --accept-multiclient --continue ./${BINARY_NAME} &
    SERVER_PID=$!
else
    echo "⚡ Starting ${PRODUCT_NAME} Server ..."

    BACKEND_PORT=$BACKEND_PORT ./${BINARY_NAME} &
    SERVER_PID=$!
fi

# Status
echo ""
echo "🚀 Server running"
echo "Press Ctrl+C to stop the server."

# Wait for background processes
wait $SERVER_PID
