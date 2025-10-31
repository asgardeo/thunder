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

import (
	"time"
)

// OAuthMessage represents the OAuth message.
type OAuthMessage struct {
	RequestType        string
	SessionDataKey     string
	RequestQueryParams map[string]string
	RequestBodyParams  map[string]string
}

// AuthorizationCode represents the authorization code.
type AuthorizationCode struct {
	CodeID              string
	Code                string
	ClientID            string
	RedirectURI         string
	AuthorizedUserID    string
	TimeCreated         time.Time
	ExpiryTime          time.Time
	Scopes              string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string
}

// AuthZPostRequest represents the request body for the authorization POST request.
type AuthZPostRequest struct {
	SessionDataKey string `json:"sessionDataKey"`
	Assertion      string `json:"assertion"`
}

// AuthZPostResponse represents the response body for the authorization POST request.
type AuthZPostResponse struct {
	RedirectURI string `json:"redirect_uri"`
}
