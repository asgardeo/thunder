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

package security

// maxPublicPathLength defines the maximum allowed length for a public path.
// This prevents potential DoS attacks via excessively long paths (even with safe regex).
const maxPublicPathLength = 4096

// publicPaths defines the list of public paths using glob patterns.
// - "*": Matches a single path segment (e.g., /a/*/b).
// - "**": Matches zero or more path segments (subpaths) at the end of the path (e.g., /a/**).
// Not allowed in the middle of the path (e.g., /a/**/b is invalid).
var publicPaths = []string{
	"/health/**",
	"/auth/**",
	"/register/passkey/**",
	"/flow/execute/**",
	"/oauth2/**",
	"/.well-known/openid-configuration/**",
	"/.well-known/oauth-authorization-server/**",
	"/.well-known/oauth-protected-resource",
	"/gate/**",
	"/develop/**",
	"/error/**",
	"/branding/resolve/**",
	"/i18n/languages",
	"/i18n/languages/*/translations/resolve",
	"/i18n/languages/*/translations/ns/*/keys/*/resolve",
	"/mcp/**", // MCP authorization is handled at MCP server handler.
}
