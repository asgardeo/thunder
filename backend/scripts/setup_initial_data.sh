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

# Script to set up initial admin user and DEVELOP application

# Default settings - can be overridden by command line arguments
BACKEND_PORT=${BACKEND_PORT:-8090}
TIMEOUT=30
RETRY_DELAY=2
CONFIG_FILE=""

# Default config file paths to check (relative to script location)
DEFAULT_CONFIG_PATHS=(
  "./repository/conf/deployment.yaml"
)

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_help() {
  echo ""
  echo "Usage:"
  echo "  $0 [OPTIONS]"
  echo ""
  echo "Options:"
  printf "  %-15s %s\n" "-port, --port" "Thunder server port (default: 8090 or from config)"
  printf "  %-15s %s\n" "-config, --config" "Path to Thunder configuration file"
  printf "  %-15s %s\n" "-timeout, --timeout" "Timeout for server readiness check (default: 30 seconds)"
  printf "  %-15s %s\n" "-h, --help" "Show this help message and exit"
  echo ""
  echo "Environment Variables:"
  printf "  %-15s %s\n" "BACKEND_PORT" "Thunder server port (can be overridden by --port)"
  echo ""
  echo "Configuration:"
  echo "  The script attempts to auto-detect server settings from:"
  echo "  - Specified config file (--config)"
  echo "  - Default locations: ../cmd/server/repository/conf/deployment.yaml"
  echo "  - Command line arguments (--port)"
  echo "  - Environment variables (BACKEND_PORT)"
  echo ""
}

parse_args() {
  while [[ "$#" -gt 0 ]]; do
    case "$1" in
      -port|--port) BACKEND_PORT="$2"; shift 2;;
      -timeout|--timeout) TIMEOUT="$2"; shift 2;;
      -config|--config) CONFIG_FILE="$2"; shift 2;;
      -h|--help) print_help; exit 0;;
      *) echo -e "${RED}Unknown parameter passed: $1${NC}"; exit 1;;
    esac
  done
}

# Function to find and read the Thunder configuration file
read_thunder_config() {
  local config_file="$CONFIG_FILE"
  
  # If no config file specified, try to find it
  if [ -z "$config_file" ]; then
    for path in "${DEFAULT_CONFIG_PATHS[@]}"; do
      local full_path
      full_path="$(dirname "$0")/$path"
      if [ -f "$full_path" ]; then
        config_file="$full_path"
        break
      fi
    done
  fi
  
  if [ -z "$config_file" ] || [ ! -f "$config_file" ]; then
    log_warning "Thunder configuration file not found. Using default settings."
    return 1
  fi
  
  log_info "Reading configuration from: $config_file"
  
  # Parse YAML using yq if available, otherwise fall back to grep/awk
  if command -v yq >/dev/null 2>&1; then
    local hostname port http_only
    hostname=$(yq eval '.server.hostname // "localhost"' "$config_file" 2>/dev/null)
    port=$(yq eval '.server.port // 8090' "$config_file" 2>/dev/null)
    http_only=$(yq eval '.server.http_only // false' "$config_file" 2>/dev/null)
    
    # Override port if provided via command line or environment
    if [ -n "$BACKEND_PORT" ] && [ "$BACKEND_PORT" != "8090" ]; then
      port="$BACKEND_PORT"
    fi
    
    # Determine protocol
    local protocol="https"
    if [ "$http_only" = "true" ]; then
      protocol="http"
    fi
    
    echo "${protocol}://${hostname}:${port}"
    return 0
  else
    log_warning "yq not found. Using basic YAML parsing."
    # Basic fallback parsing using grep and awk
    local hostname port
    hostname=$(grep -E '^\s*hostname:' "$config_file" | awk -F':' '{gsub(/[[:space:]"'\'']/,"",$2); print $2}' | head -1)
    port=$(grep -E '^\s*port:' "$config_file" | awk -F':' '{gsub(/[[:space:]]/,"",$2); print $2}' | head -1)
    
    # Use defaults if not found
    hostname=${hostname:-localhost}
    port=${port:-8090}
    
    # Override port if provided via command line or environment
    if [ -n "$BACKEND_PORT" ] && [ "$BACKEND_PORT" != "8090" ]; then
      port="$BACKEND_PORT"
    fi
    
    # Check for http_only setting (basic parsing)
    if grep -q 'http_only.*true' "$config_file" 2>/dev/null; then
      echo "http://${hostname}:${port}"
    else
      echo "https://${hostname}:${port}"
    fi
    return 0
  fi
}

# Function to construct base URL
construct_base_url() {
  local config_url
  config_url=$(read_thunder_config)
  
  if [ $? -eq 0 ] && [ -n "$config_url" ]; then
    echo "$config_url"
  else
    # Fallback to environment/command line settings
    local port="${BACKEND_PORT:-8090}"
    echo "https://localhost:${port}"
  fi
}

log_info() {
  echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
  echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
  echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
  echo -e "${RED}❌ $1${NC}"
}

wait_for_server() {
  log_info "Waiting for Thunder server to be ready at ${BASE_URL}..."
  
  # Try multiple health endpoints
  local health_endpoints=("/health/readiness" "/health/liveness" "/healthcheck")
  
  local elapsed=0
  while [ $elapsed -lt $TIMEOUT ]; do
    for endpoint in "${health_endpoints[@]}"; do
      if curl -k -s "${BASE_URL}${endpoint}" > /dev/null 2>&1; then
        log_success "Server is ready! (responded to ${endpoint})"
        return 0
      fi
    done
    
    sleep $RETRY_DELAY
    elapsed=$((elapsed + RETRY_DELAY))
    printf "."
  done
  
  echo ""
  log_error "Server is not ready after ${TIMEOUT} seconds"
  log_info "Tried endpoints: ${health_endpoints[*]}"
  log_info "Make sure Thunder server is running at ${BASE_URL}"
  exit 1
}

create_user_schema() {
  log_info "Creating Default user schema..."
  
  local response
  response=$(curl -k -s -w "%{http_code}" -X POST \
    "${BASE_URL}/user-schemas" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "default-user-schema",
      "schema": {
        "sub": {
          "type": "string",
          "required": true,
          "unique": true
        },
        "email": {
          "type": "string",
          "required": true,
          "unique": true
        },
        "email_verified": {
          "type": "boolean",
          "required": false
        },
        "name": {
          "type": "string",
          "required": false
        },
        "given_name": {
          "type": "string",
          "required": false
        },
        "family_name": {
          "type": "string",
          "required": false
        },
        "picture": {
          "type": "string",
          "required": false
        },
        "phone_number": {
          "type": "string",
          "required": false
        },
        "phone_number_verified": {
          "type": "boolean",
          "required": false
        }
      }
    }')
  
  local http_code="${response: -3}"
  local body="${response%???}"
  
  if [[ "$http_code" == "201" ]] || [[ "$http_code" == "200" ]]; then
    log_success "User schema created successfully"
    return 0
  elif [[ "$http_code" == "409" ]]; then
    log_warning "User schema already exists, skipping creation"
    return 0
  else
    log_error "Failed to create user schema. HTTP status: $http_code"
    echo "Response: $body"
    return 1
  fi
}

# Organization unit is not needed - removed get_default_ou function

create_admin_user() {
  log_info "Creating admin user..."
  
  local response
  response=$(curl -k -s -w "%{http_code}" -X POST \
    "${BASE_URL}/users" \
    -H "Content-Type: application/json" \
    -d "{
      \"type\": \"default-user-schema\",
      \"attributes\": {
        \"username\": \"admin\",
        \"password\": \"admin\",
        \"sub\": \"admin\",
        \"email\": \"admin@thunder.dev\",
        \"email_verified\": true,
        \"name\": \"Administrator\",
        \"given_name\": \"Admin\",
        \"family_name\": \"User\",
        \"picture\": \"https://example.com/avatar.jpg\",
        \"phone_number\": \"+12345678920\",
        \"phone_number_verified\": true
      }
    }")
  
  local http_code="${response: -3}"
  local body="${response%???}"
  
  if [[ "$http_code" == "201" ]] || [[ "$http_code" == "200" ]]; then
    log_success "Admin user created successfully"
    log_info "Username: admin"
    log_info "Password: admin"
    return 0
  elif [[ "$http_code" == "409" ]]; then
    log_warning "Admin user already exists, skipping creation"
    return 0
  else
    log_error "Failed to create admin user. HTTP status: $http_code"
    echo "Response: $body"
    return 1
  fi
}

create_develop_app() {
  log_info "Creating DEVELOP application..."
  
  local response
  response=$(curl -k -s -w "%{http_code}" -X POST \
    "${BASE_URL}/applications" \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"Develop\",
      \"description\": \"Developer application for Thunder\",
      \"url\": \"${BASE_URL}/develop\",
      \"logo_url\": \"${BASE_URL}/develop/assets/images/asgardeo-trifacta.svg\",
      \"auth_flow_graph_id\": \"auth_flow_config_basic\",
      \"registration_flow_graph_id\": \"registration_flow_config_basic\",
      \"is_registration_flow_enabled\": true,
      \"user_attributes\": [\"given_name\",\"family_name\",\"email\",\"groups\", \"name\"],
      \"inbound_auth_config\": [{
        \"type\": \"oauth2\",
        \"config\": {
          \"client_id\": \"DEVELOP\",
          \"redirect_uris\": [\"https://localhost:5191\"],
          \"grant_types\": [\"authorization_code\"],
          \"response_types\": [\"code\"],
          \"pkce_required\": false,
          \"token_endpoint_auth_method\": \"none\",
          \"public_client\": true,
          \"token\": {
            \"issuer\": \"${BASE_URL}/oauth2/token\",
            \"access_token\": {
              \"validity_period\": 3600,
              \"user_attributes\": [\"given_name\",\"family_name\",\"email\",\"groups\", \"name\"]
            },
            \"id_token\": {
              \"validity_period\": 3600,
              \"user_attributes\": [\"given_name\",\"family_name\",\"email\",\"groups\", \"name\"],
              \"scope_claims\": {
                \"profile\": [\"name\",\"given_name\",\"family_name\",\"picture\"],
                \"email\": [\"email\",\"email_verified\"],
                \"phone\": [\"phone_number\",\"phone_number_verified\"],
                \"group\": [\"groups\"]
              }
            }
          }
        }
      }]
    }")
  
  local http_code="${response: -3}"
  local body="${response%???}"
  
  if [[ "$http_code" == "201" ]] || [[ "$http_code" == "200" ]]; then
    log_success "DEVELOP application created successfully"
    log_info "Application URL: ${BASE_URL}/develop"
    log_info "Client ID: DEVELOP"
    return 0
  elif [[ "$http_code" == "409" ]]; then
    log_warning "DEVELOP application already exists, skipping creation"
    return 0
  else
    log_error "Failed to create DEVELOP application. HTTP status: $http_code"
    echo "Response: $body"
    return 1
  fi
}

main() {
  echo "🚀 Thunder Initial Data Setup Script"
  echo "===================================="
  echo ""
  
  parse_args "$@"
  
  # Construct base URL dynamically from configuration
  BASE_URL=$(construct_base_url)
  
  log_info "Using Thunder server at: ${BASE_URL}"
  
  # Show configuration source
  if [ -n "${BACKEND_PORT}" ] && [ "${BACKEND_PORT}" != "8090" ]; then
    log_info "Port override detected: ${BACKEND_PORT}"
  fi
  
  echo ""
  
  # Wait for server to be ready
  wait_for_server
  echo ""
  
  # Create user schema
  create_user_schema
  if [ $? -ne 0 ]; then
    log_error "Failed to create user schema. Aborting."
    exit 1
  fi
  echo ""
  
  # Create admin user
  create_admin_user
  if [ $? -ne 0 ]; then
    log_error "Failed to create admin user. Aborting."
    exit 1
  fi
  echo ""
  
  # Create DEVELOP application
  create_develop_app
  if [ $? -ne 0 ]; then
    log_error "Failed to create DEVELOP application. Aborting."
    exit 1
  fi
  echo ""
  
  log_success "Initial data setup completed successfully!"
  echo ""
  echo "📱 You can now access:"
  echo "   🚪 Gate (Login/Register): ${BASE_URL}/signin"
  echo "   🛠️  Develop (Admin Console): ${BASE_URL}/develop"
  echo ""
  echo "👤 Admin credentials:"
  echo "   Username: admin"
  echo "   Password: admin"
  echo ""
}

main "$@"
