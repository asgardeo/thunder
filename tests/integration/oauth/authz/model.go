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

package authz

// TestCase represents a test case for authorization code flow
type TestCase struct {
	Name           string
	ClientID       string
	RedirectURI    string
	ResponseType   string
	Scope          string
	State          string
	Username       string
	Password       string
	ExpectedStatus int
	ExpectedError  string
}

// FlowResponse represents the response from flow execution
type FlowResponse struct {
	FlowID        string    `json:"flowId"`
	FlowStatus    string    `json:"flowStatus"`
	Type          string    `json:"type"`
	Data          *FlowData `json:"data,omitempty"`
	Assertion     string    `json:"assertion,omitempty"`
	FailureReason string    `json:"failureReason,omitempty"`
}

// FlowData represents the data returned by flow execution
type FlowData struct {
	Inputs []FlowInput `json:"inputs,omitempty"`
}

// FlowInput represents an input required by the flow
type FlowInput struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// AuthorizationResponse represents the response from authorization completion
type AuthorizationResponse struct {
	RedirectURI string `json:"redirect_uri"`
}

// TokenResponse represents the response from token exchange
type TokenResponse struct {
	AccessToken  string  `json:"access_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    float64 `json:"expires_in"`
	Scope        string  `json:"scope,omitempty"`
	RefreshToken string  `json:"refresh_token,omitempty"`
}

// TokenHTTPResult captures raw HTTP response details from the token endpoint.
type TokenHTTPResult struct {
	StatusCode int
	Body       []byte
	Token      *TokenResponse
}

// FlowStep represents a single step in a flow execution
type FlowStep struct {
	FlowID        string    `json:"flowId"`
	FlowStatus    string    `json:"flowStatus"`
	Type          string    `json:"type"`
	Data          *FlowData `json:"data,omitempty"`
	Assertion     string    `json:"assertion,omitempty"`
	FailureReason string    `json:"failureReason,omitempty"`
}
