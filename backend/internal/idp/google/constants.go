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

package google

// Constants for Google OAuth2 authentication.
const (
	// Google OAuth2 endpoints (also defined in service.go for reference)
	AuthorizeEndpoint = "https://accounts.google.com/o/oauth2/v2/auth"
	TokenEndpoint     = "https://oauth2.googleapis.com/token" // #nosec G101
	UserInfoEndpoint  = "https://openidconnect.googleapis.com/v1/userinfo"
	JwksEndpoint      = "https://www.googleapis.com/oauth2/v3/certs"

	// Google OAuth2 issuer values
	Issuer1 = "accounts.google.com"
	Issuer2 = "https://accounts.google.com"
)
