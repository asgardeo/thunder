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

# Bootstrap Script: Sample Resources Setup
# Creates resources required to run the Thunder sample experience

set -e

# Source common functions from the same directory as this script
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]:-$0}")"
source "${SCRIPT_DIR}/common.sh"

log_info "Creating sample Thunder resources..."
echo ""

# ============================================================================
# Create Customers Organization Unit
# ============================================================================

CUSTOMER_OU_HANDLE="customers"

log_info "Creating Customers organization unit..."

read -r -d '' CUSTOMERS_OU_PAYLOAD <<JSON || true
{
  "handle": "${CUSTOMER_OU_HANDLE}",
  "name": "Customers",
  "description": "Organization unit for customer accounts"
}
JSON

RESPONSE=$(thunder_api_call POST "/organization-units" "${CUSTOMERS_OU_PAYLOAD}")
HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "Customers organization unit created successfully"
    CUSTOMER_OU_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "Customers organization unit already exists, retrieving ID..."
    # Get existing OU ID by handle to ensure we get the correct "customers" OU
    RESPONSE=$(thunder_api_call GET "/organization-units/tree/${CUSTOMER_OU_HANDLE}")
    HTTP_CODE="${RESPONSE: -3}"
    BODY="${RESPONSE%???}"

    if [[ "$HTTP_CODE" == "200" ]]; then
        CUSTOMER_OU_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    else
        log_error "Failed to fetch organization unit by handle '${CUSTOMER_OU_HANDLE}' (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
        exit 1
    fi
else
    log_error "Failed to create Customers organization unit (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi

if [[ -z "$CUSTOMER_OU_ID" ]]; then
    log_error "Could not determine Customers organization unit ID"
    exit 1
fi

log_info "Customers OU ID: $CUSTOMER_OU_ID"

echo ""

# ============================================================================
# Create Customer User Type
# ============================================================================

log_info "Creating Customer user type..."

read -r -d '' CUSTOMER_USER_TYPE_PAYLOAD <<JSON || true
{
  "name": "Customer",
  "ouId": "${CUSTOMER_OU_ID}",
  "allowSelfRegistration": true,
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
    "given_name": {
      "type": "string",
      "required": false
    },
    "family_name": {
      "type": "string",
      "required": false
    }
  }
}
JSON

RESPONSE=$(thunder_api_call POST "/user-types" "${CUSTOMER_USER_TYPE_PAYLOAD}")
HTTP_CODE="${RESPONSE: -3}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]]; then
    log_success "Customer user type created successfully"
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "Customer user type already exists, skipping"
else
    log_error "Failed to create Customer user type (HTTP $HTTP_CODE)"
    exit 1
fi

echo ""

# ============================================================================
# Create Sample Application
# ============================================================================

log_info "Creating Sample App application..."

read -r -d '' SAMPLE_APP_PAYLOAD <<JSON || true
{
  "name": "Sample App",
  "description": "Sample application for testing",
  "url": "https://localhost:3000",
  "logo_url": "https://localhost:3000/logo.png",
  "tos_uri": "https://localhost:3000/terms",
  "policy_uri": "https://localhost:3000/privacy",
  "contacts": ["admin@example.com", "support@example.com"],
  "is_registration_flow_enabled": true,
  "user_attributes": ["given_name","family_name","email","groups"],
  "allowed_user_types": ["Customer"],
  "inbound_auth_config": [{
    "type": "oauth2",
    "config": {
      "client_id": "sample_app_client",
      "redirect_uris": ["https://localhost:3000"],
      "grant_types": ["authorization_code"],
      "response_types": ["code"],
      "token_endpoint_auth_method": "none",
      "pkce_required": true,
      "public_client": true,
      "scopes": ["openid", "profile", "email"],
      "token": {
        "issuer": "thunder",
        "access_token": {
          "validity_period": 3600,
          "user_attributes": ["given_name","family_name","email","groups"]
        },
        "id_token": {
          "validity_period": 3600,
          "user_attributes": ["given_name","family_name","email","groups"],
          "scope_claims": {
            "profile": ["name","given_name","family_name","picture"],
            "email": ["email","email_verified"],
            "phone": ["phone_number","phone_number_verified"],
            "group": ["groups"]
          }
        }
      }
    }
  }]
}
JSON

RESPONSE=$(thunder_api_call POST "/applications" "${SAMPLE_APP_PAYLOAD}")
HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]] || [[ "$HTTP_CODE" == "202" ]]; then
    log_success "Sample App created successfully"
    SAMPLE_APP_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [[ -n "$SAMPLE_APP_ID" ]]; then
        log_info "Sample App ID: $SAMPLE_APP_ID"
    else
        log_warning "Could not extract Sample App ID from response"
    fi
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "Sample App already exists, skipping"
elif [[ "$HTTP_CODE" == "400" ]] && [[ "$BODY" =~ (Application already exists|APP-1022) ]]; then
    log_warning "Sample App already exists, skipping"
else
    log_error "Failed to create Sample App (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi

echo ""

# ============================================================================
# Create React SDK Sample Application
# ============================================================================

log_info "Creating React SDK Sample App application..."

read -r -d '' REACT_SDK_APP_PAYLOAD <<JSON || true
{
  "name": "React SDK Sample",
  "description": "Sample React application using Thunder React SDK",
  "client_id": "REACT_SDK_SAMPLE",
  "url": "https://localhost:3000",
  "logo_url": "https://localhost:3000/logo.png",
  "tos_uri": "https://localhost:3000/terms",
  "policy_uri": "https://localhost:3000/privacy",
  "contacts": ["admin@example.com"],
  "is_registration_flow_enabled": true,
  "assertion": {
    "issuer": "thunder",
    "validity_period": 3600,
    "user_attributes": null
  },
  "certificate": {
    "type": "NONE",
    "value": ""
  },
  "user_attributes": ["given_name","family_name","email","groups","name"],
  "allowed_user_types": ["Customer"],
  "inbound_auth_config": [{
    "type": "oauth2",
    "config": {
      "client_id": "REACT_SDK_SAMPLE",
      "redirect_uris": ["https://localhost:3000"],
      "grant_types": ["authorization_code"],
      "response_types": ["code"],
      "token_endpoint_auth_method": "none",
      "pkce_required": true,
      "public_client": true,
      "token": {
        "issuer": "https://localhost:8090",
        "access_token": {
          "validity_period": 3600,
          "user_attributes": ["given_name","family_name","email","groups","name"]
        },
        "id_token": {
          "validity_period": 3600,
          "user_attributes": ["given_name","family_name","email","groups","name"],
          "scope_claims": {
            "email": ["email","email_verified"],
            "group": ["groups"],
            "phone": ["phone_number","phone_number_verified"],
            "profile": ["name","given_name","family_name","picture"]
          }
        }
      }
    }
  }]
}
JSON

RESPONSE=$(thunder_api_call POST "/applications" "${REACT_SDK_APP_PAYLOAD}")
HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [[ "$HTTP_CODE" == "201" ]] || [[ "$HTTP_CODE" == "200" ]] || [[ "$HTTP_CODE" == "202" ]]; then
    log_success "React SDK Sample App created successfully"
    REACT_SDK_APP_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [[ -n "$REACT_SDK_APP_ID" ]]; then
        log_info "React SDK Sample App ID: $REACT_SDK_APP_ID"
    else
        log_warning "Could not extract React SDK Sample App ID from response"
    fi
elif [[ "$HTTP_CODE" == "409" ]]; then
    log_warning "React SDK Sample App already exists, skipping"
elif [[ "$HTTP_CODE" == "400" ]] && [[ "$BODY" =~ (Application already exists|APP-1022) ]]; then
    log_warning "React SDK Sample App already exists, skipping"
else
    log_error "Failed to create React SDK Sample App (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
fi

echo ""

# ============================================================================
# Summary
# ============================================================================

log_success "Sample resources setup completed successfully!"
echo ""
