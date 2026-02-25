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

# Bootstrap Script: Default Resources Setup
# Creates default organization unit, user schema, admin user, system resource server, system action, admin role, and DEVELOP application

set -e

# Parse command line arguments for custom redirect URIs
CUSTOM_DEVELOP_REDIRECT_URIS=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --develop-redirect-uris)
            CUSTOM_DEVELOP_REDIRECT_URIS="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

# Source common functions from the same directory as this script
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
source "${SCRIPT_DIR}/common.sh"

log_info "Creating default Thunder resources..."
echo ""

# ============================================================================
# Create Default Organization Unit
# ============================================================================

log_info "Creating default organization unit..."

RESPONSE=$(thunder_api_call POST "/organization-units" '{
  "handle": "default",
  "name": "Default",
  "description": "Default organization unit"
}')

HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "Organization unit created successfully"
    DEFAULT_OU_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [[ -n "$DEFAULT_OU_ID" ]]; then
        log_info "Default OU ID: $DEFAULT_OU_ID"
    else
        log_error "Could not extract OU ID from response"
        exit 1
    fi
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "Organization unit already exists, retrieving OU ID..."
    # Get existing OU ID by handle to ensure we get the correct "default" OU
    RESPONSE=$(thunder_api_call GET "/organization-units/tree/default")
    HTTP_CODE="${RESPONSE: -3}"
    BODY="${RESPONSE%???}"

    if [[ "$HTTP_CODE" == "200" ]]; then
        DEFAULT_OU_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        if [[ -n "$DEFAULT_OU_ID" ]]; then
            log_success "Found OU ID: $DEFAULT_OU_ID"
        else
            log_error "Could not find OU ID in response"
            exit 1
        fi
    else
        log_error "Failed to fetch organization unit by handle 'default' (HTTP $HTTP_CODE)"
        exit 1
    fi
else
    log_error "Failed to create organization unit (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi

echo ""

# ============================================================================
# Create Default User Schema
# ============================================================================

log_info "Creating default user schema (person)..."

RESPONSE=$(thunder_api_call POST "/user-schemas" '{
  "name": "Person",
  "ouId": "'${DEFAULT_OU_ID}'",
  "schema": {
    "username": {
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
    "given_name": {
      "type": "string",
      "required": false
    },
    "family_name": {
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
    },
    "password": {
      "type": "string",
      "required": true,
      "credential": true
    }
  }
}')

HTTP_CODE="${RESPONSE: -3}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "User schema created successfully"
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "User schema already exists, skipping"
else
    log_error "Failed to create user schema (HTTP $HTTP_CODE)"
    exit 1
fi

echo ""

# ============================================================================
# Create Admin User
# ============================================================================

log_info "Creating admin user..."

RESPONSE=$(thunder_api_call POST "/users" '{
  "type": "Person",
  "organizationUnit": "'${DEFAULT_OU_ID}'",
  "attributes": {
    "username": "admin",
    "password": "admin",
    "sub": "admin",
    "email": "admin@thunder.dev",
    "email_verified": true,
    "name": "Administrator",
    "given_name": "Admin",
    "family_name": "User",
    "picture": "https://example.com/avatar.jpg",
    "phone_number": "+12345678920",
    "phone_number_verified": true
  }
}')

HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "Admin user created successfully"
    log_info "Username: admin"
    log_info "Password: admin"

    # Extract admin user ID
    ADMIN_USER_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [[ -z "$ADMIN_USER_ID" ]]; then
        log_warning "Could not extract admin user ID from response"
    else
        log_info "Admin user ID: $ADMIN_USER_ID"
    fi
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "Admin user already exists, retrieving user ID..."

    # Get existing admin user ID
    RESPONSE=$(thunder_api_call GET "/users")
    HTTP_CODE="${RESPONSE: -3}"
    BODY="${RESPONSE%???}"

    if [[ "$HTTP_CODE" == "200" ]]; then
        # Parse JSON to find admin user
        ADMIN_USER_ID=$(echo "$BODY" | grep -o '"id":"[^"]*","[^"]*":"[^"]*","attributes":{[^}]*"username":"admin"' | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

        # Fallback parsing
        if [[ -z "$ADMIN_USER_ID" ]]; then
            ADMIN_USER_ID=$(echo "$BODY" | sed 's/},{/}\n{/g' | grep '"username":"admin"' | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        fi

        if [[ -n "$ADMIN_USER_ID" ]]; then
            log_success "Found admin user ID: $ADMIN_USER_ID"
        else
            log_error "Could not find admin user in response"
            exit 1
        fi
    else
        log_error "Failed to fetch users (HTTP $HTTP_CODE)"
        exit 1
    fi
else
    log_error "Failed to create admin user (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi

echo ""

# ============================================================================
# Create System Resource Server
# ============================================================================

log_info "Creating system resource server..."

if [[ -z "$DEFAULT_OU_ID" ]]; then
    log_error "Default OU ID is not available. Cannot create resource server."
    exit 1
fi

RESPONSE=$(thunder_api_call POST "/resource-servers" "{
  \"name\": \"System\",
  \"description\": \"System resource server\",
  \"identifier\": \"system\",
  \"ouId\": \"${DEFAULT_OU_ID}\"
}")

HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "Resource server created successfully"
    SYSTEM_RS_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [[ -n "$SYSTEM_RS_ID" ]]; then
        log_info "System resource server ID: $SYSTEM_RS_ID"
    else
        log_error "Could not extract resource server ID from response"
        exit 1
    fi
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "Resource server already exists, retrieving ID..."
    # Get existing resource server ID
    RESPONSE=$(thunder_api_call GET "/resource-servers")
    HTTP_CODE="${RESPONSE: -3}"
    BODY="${RESPONSE%???}"

    if [[ "$HTTP_CODE" == "200" ]]; then
        SYSTEM_RS_ID=$(echo "$BODY" | grep -o '"id":"[^"]*","[^"]*":"System"' | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

        # Fallback parsing
        if [[ -z "$SYSTEM_RS_ID" ]]; then
            SYSTEM_RS_ID=$(echo "$BODY" | sed 's/},{/}\n{/g' | grep '"identifier":"system"' | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        fi

        if [[ -n "$SYSTEM_RS_ID" ]]; then
            log_success "Found resource server ID: $SYSTEM_RS_ID"
        else
            log_error "Could not find resource server ID in response"
            exit 1
        fi
    else
        log_error "Failed to fetch resource servers (HTTP $HTTP_CODE)"
        exit 1
    fi
else
    log_error "Failed to create resource server (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi

echo ""

# ============================================================================
# Create System Action
# ============================================================================

log_info "Creating 'system' action on resource server..."

if [[ -z "$SYSTEM_RS_ID" ]]; then
    log_error "System resource server ID is not available. Cannot create action."
    exit 1
fi

RESPONSE=$(thunder_api_call POST "/resource-servers/${SYSTEM_RS_ID}/actions" '{
  "name": "System Access",
  "description": "Full system access permission",
  "handle": "system"
}')

HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "System action created successfully"
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "System action already exists, skipping"
else
    log_error "Failed to create system action (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi

echo ""

# ============================================================================
# Create Admin Role
# ============================================================================

log_info "Creating admin role with 'system' permission..."

if [[ -z "$ADMIN_USER_ID" ]]; then
    log_error "Admin user ID is not available. Cannot create role."
    exit 1
fi

if [[ -z "$DEFAULT_OU_ID" ]]; then
    log_error "Default OU ID is not available. Cannot create role."
    exit 1
fi

if [[ -z "$SYSTEM_RS_ID" ]]; then
    log_error "System resource server ID is not available. Cannot create role."
    exit 1
fi

RESPONSE=$(thunder_api_call POST "/roles" "{
  \"name\": \"Administrator\",
  \"description\": \"System administrator role with full permissions\",
  \"ouId\": \"${DEFAULT_OU_ID}\",
  \"permissions\": [
    {
      \"resourceServerId\": \"${SYSTEM_RS_ID}\",
      \"permissions\": [\"system\"]
    }
  ],
  \"assignments\": [
    {
      \"id\": \"${ADMIN_USER_ID}\",
      \"type\": \"user\"
    }
  ]
}")

HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "Admin role created and assigned to admin user"
    ADMIN_ROLE_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [[ -n "$ADMIN_ROLE_ID" ]]; then
        log_info "Admin role ID: $ADMIN_ROLE_ID"
    fi
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "Admin role already exists"
else
    log_error "Failed to create admin role (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi

echo ""

# ============================================================================
# Create Default Flows
# ============================================================================

log_info "Creating default flows..."

# Path to flow definitions directories
AUTH_FLOWS_DIR="${SCRIPT_DIR}/flows/authentication"
REG_FLOWS_DIR="${SCRIPT_DIR}/flows/registration"
USER_ONBOARDING_FLOWS_DIR="${SCRIPT_DIR}/flows/user_onboarding"

# Check if flows directory exists
if [[ ! -d "$AUTH_FLOWS_DIR" ]] && [[ ! -d "$REG_FLOWS_DIR" ]] && [[ ! -d "$USER_ONBOARDING_FLOWS_DIR" ]]; then
    log_warning "Flow definition directories not found, skipping flow creation"
else
    FLOW_COUNT=0
    FLOW_SUCCESS=0
    FLOW_SKIPPED=0

    # Process authentication flows
    if [[ -d "$AUTH_FLOWS_DIR" ]]; then
        shopt -s nullglob
        AUTH_FILES=("$AUTH_FLOWS_DIR"/*.json)
        shopt -u nullglob

        if [[ ${#AUTH_FILES[@]} -gt 0 ]]; then
            log_info "Processing authentication flows..."
            
            # Fetch existing auth flows
            RESPONSE=$(thunder_api_call GET "/flows?flowType=AUTHENTICATION&limit=200")
            HTTP_CODE="${RESPONSE: -3}"
            BODY="${RESPONSE%???}"

            # Store existing auth flows as "handle|id" pairs
            EXISTING_AUTH_FLOWS=""
            if [[ "$HTTP_CODE" == "200" ]]; then
                while IFS= read -r line; do
                    FLOW_ID=$(echo "$line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
                    FLOW_HANDLE=$(echo "$line" | grep -o '"handle":"[^"]*"' | cut -d'"' -f4)
                    if [[ -n "$FLOW_ID" ]] && [[ -n "$FLOW_HANDLE" ]]; then
                        EXISTING_AUTH_FLOWS="${EXISTING_AUTH_FLOWS}${FLOW_HANDLE}|${FLOW_ID}"$'\n'
                        log_debug "Found existing auth flow: handle=$FLOW_HANDLE (ID: $FLOW_ID)"
                    fi
                done < <(echo "$BODY" | grep -o '{[^}]*"id":"[^"]*"[^}]*"handle":"[^"]*"[^}]*}')
            fi
            
            log_debug "Total existing auth flows found: $(echo "$EXISTING_AUTH_FLOWS" | grep -c '|' || echo 0)"
            
            for FLOW_FILE in "$AUTH_FLOWS_DIR"/*.json; do
                [[ ! -f "$FLOW_FILE" ]] && continue

                FLOW_COUNT=$((FLOW_COUNT + 1))
                FLOW_HANDLE=$(grep -o '"handle"[[:space:]]*:[[:space:]]*"[^"]*"' "$FLOW_FILE" | head -1 | sed 's/"handle"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
                FLOW_NAME=$(grep -o '"name"[[:space:]]*:[[:space:]]*"[^"]*"' "$FLOW_FILE" | head -1 | sed 's/"name"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
                log_debug "Processing flow file: $FLOW_FILE with handle: $FLOW_HANDLE, name: $FLOW_NAME"
                
                # Check if flow exists by handle
                if echo "$EXISTING_AUTH_FLOWS" | grep -q "^${FLOW_HANDLE}|"; then
                    # Update existing flow
                    FLOW_ID=$(echo "$EXISTING_AUTH_FLOWS" | grep "^${FLOW_HANDLE}|" | cut -d'|' -f2)
                    log_info "Updating existing auth flow: $FLOW_NAME (handle: $FLOW_HANDLE)"
                    update_flow "$FLOW_ID" "$FLOW_FILE"
                    RESULT=$?
                    if [[ $RESULT -eq 0 ]]; then
                        FLOW_SUCCESS=$((FLOW_SUCCESS + 1))
                    fi
                else
                    # Create new flow
                    create_flow "$FLOW_FILE"
                    RESULT=$?
                    if [[ $RESULT -eq 0 ]]; then
                        FLOW_SUCCESS=$((FLOW_SUCCESS + 1))
                    elif [[ $RESULT -eq 2 ]]; then
                        FLOW_SKIPPED=$((FLOW_SKIPPED + 1))
                    fi
                fi
            done
        else
            log_warning "No authentication flow files found"
        fi
    fi

    # Process registration flows
    if [[ -d "$REG_FLOWS_DIR" ]]; then
        shopt -s nullglob
        REG_FILES=("$REG_FLOWS_DIR"/*.json)
        shopt -u nullglob
        
        if [[ ${#REG_FILES[@]} -gt 0 ]]; then
            log_info "Processing registration flows..."
            
            # Fetch existing registration flows
            RESPONSE=$(thunder_api_call GET "/flows?flowType=REGISTRATION&limit=200")
            HTTP_CODE="${RESPONSE: -3}"
            BODY="${RESPONSE%???}"

            # Store existing registration flows as "handle|id" pairs
            EXISTING_REG_FLOWS=""
            if [[ "$HTTP_CODE" == "200" ]]; then
                while IFS= read -r line; do
                    FLOW_ID=$(echo "$line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
                    FLOW_HANDLE=$(echo "$line" | grep -o '"handle":"[^"]*"' | cut -d'"' -f4)
                    if [[ -n "$FLOW_ID" ]] && [[ -n "$FLOW_HANDLE" ]]; then
                        EXISTING_REG_FLOWS="${EXISTING_REG_FLOWS}${FLOW_HANDLE}|${FLOW_ID}"$'\n'
                    fi
                done < <(echo "$BODY" | grep -o '{[^}]*"id":"[^"]*"[^}]*"handle":"[^"]*"[^}]*}')
            fi

            for FLOW_FILE in "$REG_FLOWS_DIR"/*.json; do
                [[ ! -f "$FLOW_FILE" ]] && continue

                FLOW_COUNT=$((FLOW_COUNT + 1))
                FLOW_HANDLE=$(grep -o '"handle"[[:space:]]*:[[:space:]]*"[^"]*"' "$FLOW_FILE" | head -1 | sed 's/"handle"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
                FLOW_NAME=$(grep -o '"name"[[:space:]]*:[[:space:]]*"[^"]*"' "$FLOW_FILE" | head -1 | sed 's/"name"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
                
                # Check if flow exists by handle
                if echo "$EXISTING_REG_FLOWS" | grep -q "^${FLOW_HANDLE}|"; then
                    # Update existing flow
                    FLOW_ID=$(echo "$EXISTING_REG_FLOWS" | grep "^${FLOW_HANDLE}|" | cut -d'|' -f2)
                    log_info "Updating existing registration flow: $FLOW_NAME (handle: $FLOW_HANDLE)"
                    update_flow "$FLOW_ID" "$FLOW_FILE"
                    RESULT=$?
                    if [[ $RESULT -eq 0 ]]; then
                        FLOW_SUCCESS=$((FLOW_SUCCESS + 1))
                    fi
                else
                    # Create new flow
                    create_flow "$FLOW_FILE"
                    RESULT=$?
                    if [[ $RESULT -eq 0 ]]; then
                        FLOW_SUCCESS=$((FLOW_SUCCESS + 1))
                    elif [[ $RESULT -eq 2 ]]; then
                        FLOW_SKIPPED=$((FLOW_SKIPPED + 1))
                    fi
                fi
            done
        else
            log_warning "No registration flow files found"
        fi
    fi

    # Process user onboarding flows
    if [[ -d "$USER_ONBOARDING_FLOWS_DIR" ]]; then
        shopt -s nullglob
        INVITE_FILES=("$USER_ONBOARDING_FLOWS_DIR"/*.json)
        shopt -u nullglob
        
        if [[ ${#INVITE_FILES[@]} -gt 0 ]]; then
            log_info "Processing user onboarding flows..."
            
            # Fetch existing user onboarding flows
            RESPONSE=$(thunder_api_call GET "/flows?flowType=USER_ONBOARDING&limit=200")
            HTTP_CODE="${RESPONSE: -3}"
            BODY="${RESPONSE%???}"

            # Store existing user onboarding flows as "handle|id" pairs
            EXISTING_INVITE_FLOWS=""
            if [[ "$HTTP_CODE" == "200" ]]; then
                while IFS= read -r line; do
                    FLOW_ID=$(echo "$line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
                    FLOW_HANDLE=$(echo "$line" | grep -o '"handle":"[^"]*"' | cut -d'"' -f4)
                    if [[ -n "$FLOW_ID" ]] && [[ -n "$FLOW_HANDLE" ]]; then
                        EXISTING_INVITE_FLOWS="${EXISTING_INVITE_FLOWS}${FLOW_HANDLE}|${FLOW_ID}"$'\n'
                    fi
                done < <(echo "$BODY" | grep -o '{[^}]*"id":"[^"]*"[^}]*"handle":"[^"]*"[^}]*}')
            fi

            for FLOW_FILE in "$USER_ONBOARDING_FLOWS_DIR"/*.json; do
                [[ ! -f "$FLOW_FILE" ]] && continue

                FLOW_COUNT=$((FLOW_COUNT + 1))
                FLOW_HANDLE=$(grep -o '"handle"[[:space:]]*:[[:space:]]*"[^"]*"' "$FLOW_FILE" | head -1 | sed 's/"handle"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
                FLOW_NAME=$(grep -o '"name"[[:space:]]*:[[:space:]]*"[^"]*"' "$FLOW_FILE" | head -1 | sed 's/"name"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
                
                # Check if flow exists by handle
                if echo "$EXISTING_INVITE_FLOWS" | grep -q "^${FLOW_HANDLE}|"; then
                    # Update existing flow
                    FLOW_ID=$(echo "$EXISTING_INVITE_FLOWS" | grep "^${FLOW_HANDLE}|" | cut -d'|' -f2)
                    log_info "Updating existing user onboarding flow: $FLOW_NAME (handle: $FLOW_HANDLE)"
                    update_flow "$FLOW_ID" "$FLOW_FILE"
                    RESULT=$?
                    if [[ $RESULT -eq 0 ]]; then
                        FLOW_SUCCESS=$((FLOW_SUCCESS + 1))
                    fi
                else
                    # Create new flow
                    create_flow "$FLOW_FILE"
                    RESULT=$?
                    if [[ $RESULT -eq 0 ]]; then
                        FLOW_SUCCESS=$((FLOW_SUCCESS + 1))
                    elif [[ $RESULT -eq 2 ]]; then
                        FLOW_SKIPPED=$((FLOW_SKIPPED + 1))
                    fi
                fi
            done
        else
            log_debug "No user onboarding flow files found"
        fi
    fi

    if [[ $FLOW_COUNT -gt 0 ]]; then
        log_info "Flow creation summary: $FLOW_SUCCESS created/updated, $FLOW_SKIPPED skipped, $((FLOW_COUNT - FLOW_SUCCESS - FLOW_SKIPPED)) failed"
    fi
fi

echo ""

# ============================================================================
# Create Application-Specific Flows
# ============================================================================

log_info "Creating application-specific flows..."

APPS_FLOWS_DIR="${SCRIPT_DIR}/flows/apps"

# Store application flow IDs as "app_name|auth_flow_id|reg_flow_id" pairs
APP_FLOW_IDS=""

if [[ -d "$APPS_FLOWS_DIR" ]]; then
    # Fetch all existing flows once
    log_info "Fetching existing flows for application flow processing..."
    
    # Get auth flows
    RESPONSE=$(thunder_api_call GET "/flows?flowType=AUTHENTICATION&limit=200")
    HTTP_CODE="${RESPONSE: -3}"
    BODY="${RESPONSE%???}"
    EXISTING_APP_AUTH_FLOWS=""
    if [[ "$HTTP_CODE" == "200" ]]; then
        while IFS= read -r line; do
            FLOW_ID=$(echo "$line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
            FLOW_HANDLE=$(echo "$line" | grep -o '"handle":"[^"]*"' | cut -d'"' -f4)
            if [[ -n "$FLOW_ID" ]] && [[ -n "$FLOW_HANDLE" ]]; then
                EXISTING_APP_AUTH_FLOWS="${EXISTING_APP_AUTH_FLOWS}${FLOW_HANDLE}|${FLOW_ID}"$'\n'
            fi
        done < <(echo "$BODY" | grep -o '{[^}]*"id":"[^"]*"[^}]*"handle":"[^"]*"[^}]*}')
    fi
    
    # Get registration flows
    RESPONSE=$(thunder_api_call GET "/flows?flowType=REGISTRATION&limit=200")
    HTTP_CODE="${RESPONSE: -3}"
    BODY="${RESPONSE%???}"
    EXISTING_APP_REG_FLOWS=""
    if [[ "$HTTP_CODE" == "200" ]]; then
        while IFS= read -r line; do
            FLOW_ID=$(echo "$line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
            FLOW_HANDLE=$(echo "$line" | grep -o '"handle":"[^"]*"' | cut -d'"' -f4)
            if [[ -n "$FLOW_ID" ]] && [[ -n "$FLOW_HANDLE" ]]; then
                EXISTING_APP_REG_FLOWS="${EXISTING_APP_REG_FLOWS}${FLOW_HANDLE}|${FLOW_ID}"$'\n'
            fi
        done < <(echo "$BODY" | grep -o '{[^}]*"id":"[^"]*"[^}]*"handle":"[^"]*"[^}]*}')
    fi

    # Process each application directory
    for APP_DIR in "$APPS_FLOWS_DIR"/*; do
        [[ ! -d "$APP_DIR" ]] && continue
        
        APP_NAME=$(basename "$APP_DIR")
        APP_AUTH_FLOW_ID=""
        APP_REG_FLOW_ID=""
        
        log_info "Processing flows for application: $APP_NAME"
        
        # Process authentication flow for app
        shopt -s nullglob
        AUTH_FLOW_FILES=("$APP_DIR"/auth_*.json)
        shopt -u nullglob
        
        if [[ ${#AUTH_FLOW_FILES[@]} -gt 0 ]]; then
            AUTH_FLOW_FILE="${AUTH_FLOW_FILES[0]}"
            FLOW_HANDLE=$(grep -o '"handle"[[:space:]]*:[[:space:]]*"[^"]*"' "$AUTH_FLOW_FILE" | head -1 | sed 's/"handle"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
            FLOW_NAME=$(grep -o '"name"[[:space:]]*:[[:space:]]*"[^"]*"' "$AUTH_FLOW_FILE" | head -1 | sed 's/"name"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
            
            # Check if auth flow exists by handle
            if echo "$EXISTING_APP_AUTH_FLOWS" | grep -q "^${FLOW_HANDLE}|"; then
                # Update existing flow
                APP_AUTH_FLOW_ID=$(echo "$EXISTING_APP_AUTH_FLOWS" | grep "^${FLOW_HANDLE}|" | cut -d'|' -f2)
                log_info "Updating existing auth flow: $FLOW_NAME (handle: $FLOW_HANDLE)"
                update_flow "$APP_AUTH_FLOW_ID" "$AUTH_FLOW_FILE"
            else
                # Create new flow
                APP_AUTH_FLOW_ID=$(create_flow "$AUTH_FLOW_FILE")
            fi
            
            # Re-fetch registration flows after creating auth flow
            if [[ -n "$APP_AUTH_FLOW_ID" ]]; then
                RESPONSE=$(thunder_api_call GET "/flows?flowType=REGISTRATION&limit=200")
                HTTP_CODE="${RESPONSE: -3}"
                BODY="${RESPONSE%???}"
                EXISTING_APP_REG_FLOWS=""
                if [[ "$HTTP_CODE" == "200" ]]; then
                    while IFS= read -r line; do
                        FLOW_ID=$(echo "$line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
                        FLOW_HANDLE_TEMP=$(echo "$line" | grep -o '"handle":"[^"]*"' | cut -d'"' -f4)
                        if [[ -n "$FLOW_ID" ]] && [[ -n "$FLOW_HANDLE_TEMP" ]]; then
                            EXISTING_APP_REG_FLOWS="${EXISTING_APP_REG_FLOWS}${FLOW_HANDLE_TEMP}|${FLOW_ID}"$'\n'
                        fi
                    done < <(echo "$BODY" | grep -o '{[^}]*"id":"[^"]*"[^}]*"handle":"[^"]*"[^}]*}')
                fi
            fi
        else
            log_warning "No authentication flow file found for app: $APP_NAME"
        fi

        # Process registration flow for app
        shopt -s nullglob
        REG_FLOW_FILES=("$APP_DIR"/registration_*.json)
        shopt -u nullglob
        
        if [[ ${#REG_FLOW_FILES[@]} -gt 0 ]]; then
            REG_FLOW_FILE="${REG_FLOW_FILES[0]}"
            FLOW_HANDLE=$(grep -o '"handle"[[:space:]]*:[[:space:]]*"[^"]*"' "$REG_FLOW_FILE" | head -1 | sed 's/"handle"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
            FLOW_NAME=$(grep -o '"name"[[:space:]]*:[[:space:]]*"[^"]*"' "$REG_FLOW_FILE" | head -1 | sed 's/"name"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
            
            # Check if registration flow exists by handle
            if echo "$EXISTING_APP_REG_FLOWS" | grep -q "^${FLOW_HANDLE}|"; then
                # Update existing flow
                APP_REG_FLOW_ID=$(echo "$EXISTING_APP_REG_FLOWS" | grep "^${FLOW_HANDLE}|" | cut -d'|' -f2)
                log_info "Updating existing registration flow: $FLOW_NAME (handle: $FLOW_HANDLE)"
                update_flow "$APP_REG_FLOW_ID" "$REG_FLOW_FILE"
            else
                # Create new flow
                APP_REG_FLOW_ID=$(create_flow "$REG_FLOW_FILE")
            fi
        else
            log_warning "No registration flow file found for app: $APP_NAME"
        fi
        
        # Store the flow IDs for this app
        log_debug "Storing flow IDs for $APP_NAME: auth=$APP_AUTH_FLOW_ID, reg=$APP_REG_FLOW_ID"
        APP_FLOW_IDS="${APP_FLOW_IDS}${APP_NAME}|${APP_AUTH_FLOW_ID}|${APP_REG_FLOW_ID}"$'\n'
    done
else
    log_warning "Application flows directory not found at $APPS_FLOWS_DIR"
fi

echo ""

# ============================================================================
# Create DEVELOP Application
# ============================================================================

log_info "Creating DEVELOP application..."

# Get flow IDs for develop app from the APP_FLOW_IDS created/found during flow processing
DEVELOP_AUTH_FLOW_ID=$(echo "$APP_FLOW_IDS" | grep "^develop|" | cut -d'|' -f2)
DEVELOP_REG_FLOW_ID=$(echo "$APP_FLOW_IDS" | grep "^develop|" | cut -d'|' -f3)
log_debug "Extracted flow IDs: auth=$DEVELOP_AUTH_FLOW_ID, reg=$DEVELOP_REG_FLOW_ID"

# Validate that flow IDs are available
if [[ -z "$DEVELOP_AUTH_FLOW_ID" ]]; then
    log_error "Develop authentication flow ID not found, cannot create DEVELOP application"
    exit 1
fi
if [[ -z "$DEVELOP_REG_FLOW_ID" ]]; then
    log_error "Develop registration flow ID not found, cannot create DEVELOP application"
    exit 1
fi

# Use THUNDER_PUBLIC_URL for redirect URIs, fallback to THUNDER_API_BASE if not set
PUBLIC_URL="${THUNDER_PUBLIC_URL:-$THUNDER_API_BASE}"

# Build redirect URIs array - default + custom if provided
REDIRECT_URIS="\"${PUBLIC_URL}/develop\""
if [[ -n "$CUSTOM_DEVELOP_REDIRECT_URIS" ]]; then
    log_info "Adding custom redirect URIs: $CUSTOM_DEVELOP_REDIRECT_URIS"
    # Split comma-separated URIs and append to array
    IFS=',' read -ra URI_ARRAY <<< "$CUSTOM_DEVELOP_REDIRECT_URIS"
    for uri in "${URI_ARRAY[@]}"; do
        # Trim whitespace
        uri=$(echo "$uri" | xargs)
        REDIRECT_URIS="${REDIRECT_URIS},\"${uri}\""
    done
fi

RESPONSE=$(thunder_api_call POST "/applications" "{
  \"name\": \"Develop\",
  \"description\": \"Developer application for Thunder\",
  \"url\": \"${PUBLIC_URL}/develop\",
  \"logo_url\": \"${PUBLIC_URL}/develop/assets/images/logo-mini.svg\",
  \"auth_flow_id\": \"${DEVELOP_AUTH_FLOW_ID}\",
  \"registration_flow_id\": \"${DEVELOP_REG_FLOW_ID}\",
  \"is_registration_flow_enabled\": false,
  \"allowed_user_types\": [\"Person\"],
  \"user_attributes\": [\"given_name\",\"family_name\",\"email\",\"groups\", \"name\"],
  \"inbound_auth_config\": [{
    \"type\": \"oauth2\",
    \"config\": {
      \"client_id\": \"DEVELOP\",
      \"redirect_uris\": [${REDIRECT_URIS}],
      \"grant_types\": [\"authorization_code\"],
      \"response_types\": [\"code\"],
      \"pkce_required\": true,
      \"token_endpoint_auth_method\": \"none\",
      \"public_client\": true,
      \"token\": {
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

HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "DEVELOP application created successfully"
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "DEVELOP application already exists, skipping"
elif [[ "$HTTP_CODE" == "400" ]] && [[ "$BODY" =~ (Application already exists|APP-1022) ]]; then
    log_warning "DEVELOP application already exists, skipping"
else
    log_error "Failed to create DEVELOP application (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi

echo ""

# ============================================================================
# Create Themes
# ============================================================================

log_info "Creating themes..."

# Get the script directory to locate theme files
THEMES_DIR="${SCRIPT_DIR}/themes"

# Check if themes directory exists
if [[ ! -d "$THEMES_DIR" ]]; then
    log_warning "Themes directory not found at ${THEMES_DIR}, skipping theme creation"
else
    shopt -s nullglob
    THEME_FILES=("$THEMES_DIR"/*.json)
    shopt -u nullglob
    
    if [[ ${#THEME_FILES[@]} -gt 0 ]]; then
        log_info "Processing themes from ${THEMES_DIR}..."
        
        THEME_COUNT=0
        THEME_SUCCESS=0
        THEME_SKIPPED=0
        
        for THEME_FILE in "${THEME_FILES[@]}"; do
            [[ ! -f "$THEME_FILE" ]] && continue
            
            THEME_COUNT=$((THEME_COUNT + 1))
            
            # Get theme name from file content or use filename
            THEME_NAME=$(grep -o '"displayName"[[:space:]]*:[[:space:]]*"[^"]*"' "$THEME_FILE" | head -1 | sed 's/"displayName"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
            if [[ -z "$THEME_NAME" ]]; then
                THEME_NAME=$(basename "$THEME_FILE" .json)
            fi
            
            log_info "Creating theme: ${THEME_NAME} (from $(basename "$THEME_FILE"))"
            THEME_PAYLOAD=$(cat "$THEME_FILE")
            
            RESPONSE=$(thunder_api_call POST "/design/themes" "${THEME_PAYLOAD}")
            HTTP_CODE="${RESPONSE: -3}"
            BODY="${RESPONSE%???}"
            
            if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
                log_success "Theme '${THEME_NAME}' created successfully"
                THEME_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
                if [[ -n "$THEME_ID" ]]; then
                    log_info "Theme ID: $THEME_ID"
                fi
                THEME_SUCCESS=$((THEME_SUCCESS + 1))
            elif [[ "$HTTP_CODE" == "409" ]]; then
                log_warning "Theme '${THEME_NAME}' already exists, skipping"
                THEME_SKIPPED=$((THEME_SKIPPED + 1))
            else
                log_error "Failed to create theme '${THEME_NAME}' (HTTP $HTTP_CODE)"
                echo "Response: $BODY"
                exit 1
            fi
        done
        
        echo ""
        log_info "Theme creation summary: ${THEME_SUCCESS} created, ${THEME_SKIPPED} skipped (Total: ${THEME_COUNT})"
    else
        log_warning "No theme files found in ${THEMES_DIR}"
    fi
fi

echo ""

# ============================================================================
# Summary
# ============================================================================

log_success "Default resources setup completed successfully!"
echo ""
log_info "👤 Admin credentials:"
log_info "   Username: admin"
log_info "   Password: admin"
log_info "   Role: Administrator (system permission)"
echo ""
