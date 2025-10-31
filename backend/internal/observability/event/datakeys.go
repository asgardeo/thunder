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

package event

// DataKey provides standardized keys for Event.Data map.
// Using these constants prevents typos and makes refactoring easier.
//
// Usage:
//
//	evt.WithData(event.DataKey.ClientID, "client123")
//	evt.WithData(event.DataKey.UserID, "user456")
var DataKey = struct {
	// Identity & User Keys
	UserID   string
	Username string
	ClientID string
	AppID    string

	// OAuth/OIDC Keys
	GrantType    string
	Scope        string
	TokenType    string
	RedirectURI  string
	ResponseType string
	State        string

	// PKCE Keys
	PKCEEnabled         string
	CodeChallenge       string
	CodeChallengeMethod string
	CodeVerifier        string

	// Token Keys
	HasRefreshToken     string
	HasIDToken          string
	ValidityPeriod      string
	CodeLifetimeSeconds string

	// Request Context Keys
	IPAddress      string
	UserAgent      string
	SessionID      string
	SessionDataKey string

	// Flow Execution Keys
	FlowID     string
	NodeID     string
	RedirectTo string

	// Event Metadata Keys
	Message    string
	Error      string
	ErrorCode  string
	DurationMs string
	LatencyUs  string

	// Testing Keys
	Key   string
	Value string
}{
	// Identity & User Keys
	UserID:   "user_id",
	Username: "username",
	ClientID: "client_id",
	AppID:    "app_id",

	// OAuth/OIDC Keys
	GrantType:    "grant_type",
	Scope:        "scope",
	TokenType:    "token_type",
	RedirectURI:  "redirect_uri",
	ResponseType: "response_type",
	State:        "state",

	// PKCE Keys
	PKCEEnabled:         "pkce_enabled",
	CodeChallenge:       "code_challenge",
	CodeChallengeMethod: "code_challenge_method",
	CodeVerifier:        "code_verifier",

	// Token Keys
	HasRefreshToken:     "has_refresh_token",
	HasIDToken:          "has_id_token",
	ValidityPeriod:      "validity_period",
	CodeLifetimeSeconds: "code_lifetime_seconds",

	// Request Context Keys
	IPAddress:      "ip_address",
	UserAgent:      "user_agent",
	SessionID:      "session_id",
	SessionDataKey: "session_data_key",

	// Flow Execution Keys
	FlowID:     "flow_id",
	NodeID:     "node_id",
	RedirectTo: "redirect_to",

	// Event Metadata Keys
	Message:    "message",
	Error:      "error",
	ErrorCode:  "error_code",
	DurationMs: "duration_ms",
	LatencyUs:  "latency_us",

	// Testing Keys
	Key:   "key",
	Value: "value",
}
