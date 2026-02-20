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

# Thunder Setup Script
# Orchestrates the complete setup lifecycle:
# 1. Starts Thunder server with security disabled
# 2. Executes bootstrap scripts (built-in + custom)
# 3. Stops Thunder server
# 4. Exits cleanly

# Ensure the script runs with Bash even if invoked via `sh`
if [ -z "${BASH_VERSION:-}" ]; then
    exec /usr/bin/env bash "$0" "$@"
fi

# Guard against Bash POSIX mode (e.g., Bash started with --posix or POSIXLY_CORRECT);
# use a guard to avoid re-exec loops if POSIX stays enabled.
if [ -z "${SETUP_SH_POSIX_REEXEC:-}" ] && set -o | grep -q 'posix.*on'; then
    SETUP_SH_POSIX_REEXEC=1 exec /usr/bin/env bash "$0" "$@"
fi

set -e

# Default settings
DEBUG_PORT=${DEBUG_PORT:-2345}
DEBUG_MODE=${DEBUG_MODE:-false}
BOOTSTRAP_FAIL_FAST=${BOOTSTRAP_FAIL_FAST:-true}
BOOTSTRAP_SKIP_PATTERN="${BOOTSTRAP_SKIP_PATTERN:-}"
BOOTSTRAP_ONLY_PATTERN="${BOOTSTRAP_ONLY_PATTERN:-}"
BOOTSTRAP_DIR="${BOOTSTRAP_DIR:-./bootstrap}"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# ============================================================================
# Logging Functions
# ============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} ✓ $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} ⚠ $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} ✗ $1"
}

log_debug() {
    if [ "${DEBUG:-false}" = "true" ]; then
        echo -e "${CYAN}[DEBUG]${NC} $1"
    fi
}

# ============================================================================
# API Call Helper Function
# ============================================================================

thunder_api_call() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"

    local url="${THUNDER_API_BASE}${endpoint}"

    log_debug "API Call: $method $url"

    if [ -z "$data" ]; then
        curl -k -s -w "\n%{http_code}" -X "$method" \
            "$url" \
            -H "Content-Type: application/json" 2>/dev/null || echo "000"
    else
        curl -k -s -w "\n%{http_code}" -X "$method" \
            "$url" \
            -H "Content-Type: application/json" \
            -d "$data" 2>/dev/null || echo "000"
    fi
}

# ============================================================================
# Help Function
# ============================================================================

print_help() {
    echo ""
    echo "Thunder Setup Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --debug                  Enable debug mode with remote debugging"
    echo "  --debug-port PORT        Set debug port (default: 2345)"
    echo "  --help                   Show this help message"
    echo ""
    echo "Description:"
    echo "  This script performs initial setup by:"
    echo "  1. Starting Thunder server temporarily with security disabled"
    echo "  2. Running bootstrap scripts to create default resources"
    echo "  3. Stopping the server cleanly"
    echo ""
    echo "  After setup completes, use './start.sh' to start Thunder normally."
    echo ""
}

# ============================================================================
# Parse Command Line Arguments
# ============================================================================

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
        --help)
            print_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# ============================================================================
# Read Configuration from deployment.yaml
# ============================================================================

CONFIG_FILE="./repository/conf/deployment.yaml"

# Function to read config with fallback
read_config() {
    local config_file="$CONFIG_FILE"

    if [ ! -f "$config_file" ]; then
        # Try alternative path (for packaged distribution)
        config_file="./backend/cmd/server/repository/conf/deployment.yaml"
    fi

    if [ ! -f "$config_file" ]; then
        log_warning "Configuration file not found, using defaults"
        return 1
    fi

    log_debug "Reading configuration from: $config_file"

    # Try yq first (YAML parser)
    if command -v yq >/dev/null 2>&1; then
        HOSTNAME=$(yq eval '.server.hostname // "localhost"' "$config_file" 2>/dev/null)
        PORT=$(yq eval '.server.port // 8090' "$config_file" 2>/dev/null)
        HTTP_ONLY=$(yq eval '.server.http_only // false' "$config_file" 2>/dev/null)
        PUBLIC_URL=$(yq eval '.server.public_url // ""' "$config_file" 2>/dev/null)
    else
        # Fallback: basic parsing with grep/awk
        HOSTNAME=$(grep -E '^\s*hostname:' "$config_file" | sed 's/#.*//' | awk -F':' '{gsub(/[[:space:]"'\'']/,"",$2); print $2}' | head -1)
        PORT=$(grep -E '^\s*port:' "$config_file" | sed 's/#.*//' | awk -F':' '{gsub(/[[:space:]]/,"",$2); print $2}' | head -1)
        PUBLIC_URL=$(grep -E '^\s*public_url:' "$config_file" | sed 's/#.*//' | grep -o '"[^"]*"' | tr -d '"' | head -1)

        # Check for http_only
        if grep -q 'http_only.*true' "$config_file" 2>/dev/null; then
            HTTP_ONLY="true"
        else
            HTTP_ONLY="false"
        fi

        # Use defaults if not found
        HOSTNAME=${HOSTNAME:-localhost}
        PORT=${PORT:-8090}
    fi

    # Determine protocol
    if [ "$HTTP_ONLY" = "true" ]; then
        PROTOCOL="http"
    else
        PROTOCOL="https"
    fi
    return 0
}

# Read configuration
read_config

# Construct base URL (internal API endpoint)
BASE_URL="${PROTOCOL}://${HOSTNAME}:${PORT}"

# Construct public URL (external/redirect URLs)
PUBLIC_URL="${PUBLIC_URL:-$BASE_URL}"

echo ""
echo "========================================="
echo "   Thunder Setup"
echo "========================================="
echo ""
echo -e "${BLUE}Server URL:${NC} $BASE_URL"
echo -e "${BLUE}Public URL:${NC} $PUBLIC_URL"
if [ "$DEBUG_MODE" = "true" ]; then
    echo -e "${BLUE}Debug:${NC} Enabled (port $DEBUG_PORT)"
fi
echo ""

# ============================================================================
# Check for Port Conflicts
# ============================================================================

check_port() {
    local port=$1
    local port_name=$2
    if lsof -ti tcp:$port >/dev/null 2>&1; then
        echo ""
        echo -e "${RED}❌ Port $port is already in use${NC}"
        echo -e "${RED}   $port_name cannot start because another process is using port $port${NC}"
        echo ""
        echo -e "${YELLOW}💡 To find the process using this port:${NC}"
        echo "   lsof -i tcp:$port"
        echo ""
        echo -e "${YELLOW}💡 To stop the process:${NC}"
        echo "   kill -9 \$(lsof -ti tcp:$port)"
        echo ""
        exit 1
    fi
}

# Check if ports are available
check_port $PORT "Thunder server"
if [ "$DEBUG_MODE" = "true" ]; then
    check_port $DEBUG_PORT "Debug server"
fi

# Check for Delve if debug mode is enabled
if [ "$DEBUG_MODE" = "true" ] && ! command -v dlv &> /dev/null; then
    echo -e "${RED}❌ Debug mode requires Delve debugger${NC}"
    echo ""
    echo "💡 Install Delve using:"
    echo "   go install github.com/go-delve/delve/cmd/dlv@latest"
    exit 1
fi

# ============================================================================
# Start Thunder Server with Security Disabled
# ============================================================================

echo -e "${YELLOW}⚠️  Starting temporary server with security disabled...${NC}"
echo ""

# Export environment variable to skip security
export THUNDER_SKIP_SECURITY=true

if [ "$DEBUG_MODE" = "true" ]; then
    dlv exec --listen=:$DEBUG_PORT --headless=true --api-version=2 --accept-multiclient --continue ./thunder &
    THUNDER_PID=$!
else
    ./thunder &
    THUNDER_PID=$!
fi

# Cleanup function
cleanup() {
    echo ""
    echo -e "${CYAN}🛑 Stopping temporary server...${NC}"
    if [ -n "$THUNDER_PID" ]; then
        kill $THUNDER_PID 2>/dev/null || true
        wait $THUNDER_PID 2>/dev/null || true
    fi
}

# Register cleanup on exit
trap cleanup EXIT INT TERM

# ============================================================================
# Wait for Server to be Ready
# ============================================================================

echo -e "${BLUE}⏳ Waiting for server to be ready...${NC}"
TIMEOUT=60
ELAPSED=0
RETRY_DELAY=2

while [ $ELAPSED -lt $TIMEOUT ]; do
    if curl -k -s "${BASE_URL}/health/readiness" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Server is ready${NC}"
        echo ""
        break
    fi
    sleep $RETRY_DELAY
    ELAPSED=$((ELAPSED + RETRY_DELAY))
    printf "."
done

if [ $ELAPSED -ge $TIMEOUT ]; then
    echo ""
    echo -e "${RED}❌ Server failed to start within ${TIMEOUT} seconds${NC}"
    echo -e "${RED}Expected server at: ${BASE_URL}${NC}"
    exit 1
fi

# ============================================================================
# Run Bootstrap Scripts
# ============================================================================

# Export variables to be used in scripts
export THUNDER_API_BASE="${BASE_URL}"
export THUNDER_PUBLIC_URL="${PUBLIC_URL}"

# Check if bootstrap directory exists
if [ ! -d "$BOOTSTRAP_DIR" ]; then
    log_warning "Bootstrap directory not found: $BOOTSTRAP_DIR"
    log_info "Skipping bootstrap execution"
else
    log_info "========================================="
    log_info "Thunder Bootstrap Process"
    log_info "========================================="
    log_info "Bootstrap directory: $BOOTSTRAP_DIR"
    log_info "Fail fast: $BOOTSTRAP_FAIL_FAST"
    log_info "Started at: $(date)"
    echo ""

    # Collect all scripts from bootstrap directory
    SCRIPTS=()

    # Find scripts in bootstrap directory (exclude common.sh)
    if [ -d "$BOOTSTRAP_DIR" ]; then
        for script in "$BOOTSTRAP_DIR"/*.sh "$BOOTSTRAP_DIR"/*.bash; do
            [ ! -e "$script" ] && continue
            if [[ "$(basename "$script")" == "common.sh" ]]; then
                continue
            fi
            SCRIPTS+=("$script")
        done
    fi

    # Sort scripts by filename (numeric prefix determines order)
    IFS=$'\n' SORTED_SCRIPTS=($(printf '%s\n' "${SCRIPTS[@]}" | sort))
    unset IFS

    if [ ${#SORTED_SCRIPTS[@]} -eq 0 ]; then
        log_warning "No bootstrap scripts found"
    else
        log_info "Discovered ${#SORTED_SCRIPTS[@]} script(s)"
        echo ""

        # Execute scripts
        SCRIPT_COUNT=0
        SUCCESS_COUNT=0
        FAILED_COUNT=0
        SKIPPED_COUNT=0

        for script in "${SORTED_SCRIPTS[@]}"; do
            script_name=$(basename "$script")

            # Skip if matches skip pattern
            if [ -n "$BOOTSTRAP_SKIP_PATTERN" ] && [[ "$script_name" =~ $BOOTSTRAP_SKIP_PATTERN ]]; then
                log_info "⊘ Skipping $script_name (matches skip pattern)"
                SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
                continue
            fi

            # Skip if doesn't match only pattern
            if [ -n "$BOOTSTRAP_ONLY_PATTERN" ] && ! [[ "$script_name" =~ $BOOTSTRAP_ONLY_PATTERN ]]; then
                log_info "⊘ Skipping $script_name (doesn't match only pattern)"
                SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
                continue
            fi

            # Check if executable
            if [ ! -x "$script" ]; then
                log_warning "$script_name is not executable, setting permissions..."
                chmod +x "$script" || {
                    log_error "Failed to make $script_name executable"
                    FAILED_COUNT=$((FAILED_COUNT + 1))
                    if [ "$BOOTSTRAP_FAIL_FAST" = "true" ]; then
                        exit 1
                    fi
                    continue
                }
            fi

            log_info "▶ Executing: $script_name"
            SCRIPT_COUNT=$((SCRIPT_COUNT + 1))

            # Execute script
            START_TIME=$(date +%s)

            set +e  # Temporarily disable exit on error to catch errors
            (
                set -e  # Re-enable in subshell to catch script errors
                source "$script"
            )
            EXIT_CODE=$?
            set -e  # Re-enable exit on error

            END_TIME=$(date +%s)
            DURATION=$((END_TIME - START_TIME))

            if [ $EXIT_CODE -eq 0 ]; then
                log_success "$script_name completed (${DURATION}s)"
                SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
            else
                log_error "$script_name failed with exit code $EXIT_CODE (${DURATION}s)"
                FAILED_COUNT=$((FAILED_COUNT + 1))

                # Check if we should fail fast
                if [ "$BOOTSTRAP_FAIL_FAST" = "true" ]; then
                    log_error "Stopping bootstrap (BOOTSTRAP_FAIL_FAST=true)"
                    exit 1
                fi
            fi
            echo ""
        done

        # Summary
        echo ""
        log_info "========================================="
        log_info "Bootstrap Summary"
        log_info "========================================="
        log_info "Total scripts discovered: ${#SORTED_SCRIPTS[@]}"
        log_info "Executed: $SCRIPT_COUNT"
        log_success "Successful: $SUCCESS_COUNT"

        if [ $FAILED_COUNT -gt 0 ]; then
            log_error "Failed: $FAILED_COUNT"
        fi

        if [ $SKIPPED_COUNT -gt 0 ]; then
            log_info "Skipped: $SKIPPED_COUNT"
        fi

        log_info "Completed at: $(date)"
        log_info "========================================="

        if [ $FAILED_COUNT -gt 0 ]; then
            exit 1
        fi

        log_success "Bootstrap completed successfully!"
    fi
fi

# ============================================================================
# Setup Completed
# ============================================================================

echo ""
echo "========================================="
echo -e "${GREEN}✅ Setup completed successfully!${NC}"
echo "========================================="
echo ""

# Cleanup will be called automatically via trap
exit 0
