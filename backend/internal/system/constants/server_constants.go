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

// Package constants defines global constants used across the system module.
package constants

const (
	// LogLevelEnvironmentVariable is the environment variable name for the log level.
	LogLevelEnvironmentVariable = "LOG_LEVEL"
	// DefaultLogLevel is the default log level used if not specified.
	DefaultLogLevel = "info"
)

// AuthorizationHeaderName is the name of the authorization header used in HTTP requests.
const AuthorizationHeaderName = "Authorization"

// AcceptHeaderName is the name of the accept header used in HTTP requests.
const AcceptHeaderName = "Accept"

// ContentTypeHeaderName is the name of the content type header used in HTTP requests.
const ContentTypeHeaderName = "Content-Type"

// TokenTypeBearer is the token type used in bearer authentication.
const TokenTypeBearer = "Bearer"

// AuthSchemeBasic is the authentication scheme prefix used in HTTP Basic authentication.
const AuthSchemeBasic = "Basic "

// ContentTypeJSON is the content type for JSON data.
const ContentTypeJSON = "application/json"

// ContentTypeFormURLEncoded is the content type for form-urlencoded data.
const ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"

// CacheControlHeaderName is the name of the cache-control header used in HTTP responses.
const CacheControlHeaderName = "Cache-Control"

// CacheControlNoCache is the cache-control directive to force revalidation.
const CacheControlNoCache = "no-cache"

// CacheControlNoStore is the cache-control directive to prevent caching.
const CacheControlNoStore = "no-store"

// CacheControlMustRevalidate is the cache-control directive to require revalidation of stale cache entries.
const CacheControlMustRevalidate = "must-revalidate"

// PragmaHeaderName is the name of the pragma header used in HTTP responses.
const PragmaHeaderName = "Pragma"

// PragmaNoCache is the pragma value to prevent caching.
const PragmaNoCache = "no-cache"

// ExpiresHeaderName is the name of the expires header used in HTTP responses.
const ExpiresHeaderName = "Expires"

// CacheControlNoCacheComposite is the combined cache-control directive to prevent caching and require revalidation.
const CacheControlNoCacheComposite = "no-cache, no-store, must-revalidate"

// ExpiresZero is the expires value to indicate immediate expiration.
const ExpiresZero = "0"

// DefaultPageSize is the default limit for pagination when not specified.
const DefaultPageSize = 30

// MaxPageSize is the maximum allowed limit for pagination.
const MaxPageSize = 100
