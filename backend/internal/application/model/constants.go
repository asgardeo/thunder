/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package model

import "errors"

const (
	// AuthFlowGraphPrefix defines the prefix for authentication flow graph IDs.
	AuthFlowGraphPrefix = "auth_flow_config_"
	// RegistrationFlowGraphPrefix defines the prefix for registration flow graph IDs.
	RegistrationFlowGraphPrefix = "registration_flow_config_"
)

// InboundAuthType represents the type of inbound authentication.
type InboundAuthType string

const (
	// OAuthInboundAuthType represents the OAuth 2.0 inbound authentication type.
	OAuthInboundAuthType InboundAuthType = "oauth2"
)

// ApplicationNotFoundError is the error returned when an application is not found.
var ApplicationNotFoundError error = errors.New("application not found")

// ApplicationDataCorruptedError is the error returned when application data is corrupted.
var ApplicationDataCorruptedError error = errors.New("application data is corrupted")
