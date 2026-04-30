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

// Package middleware provides HTTP middleware functions for request processing.
package middleware

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// CORSOptions represents the CORS configuration for HTTP requests.
type CORSOptions struct {
	AllowedMethods   string
	AllowedHeaders   string
	AllowCredentials bool
}

// WithCORS wraps an HTTP handler with CORS headers based on the provided options.
// It returns the pattern and wrapped handler that can be registered with http.ServeMux.
func WithCORS(pattern string, handler http.HandlerFunc, opts CORSOptions) (string, http.HandlerFunc) {
	return pattern, func(w http.ResponseWriter, r *http.Request) {
		applyCORSHeaders(w, r, opts)
		handler(w, r)
	}
}

// applyCORSHeaders sets the CORS headers for the response based on the configured allowed origins.
func applyCORSHeaders(w http.ResponseWriter, r *http.Request, opts CORSOptions) {
	allowedOrigins := getAllowedOrigins()

	requestOrigin := r.Header.Get("Origin")
	if requestOrigin == "" {
		return
	}

	// Set the CORS headers if allowed origins are configured.
	allowedOrigin := utils.GetAllowedOrigin(allowedOrigins, requestOrigin)
	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)

		if opts.AllowedMethods != "" {
			w.Header().Set("Access-Control-Allow-Methods", opts.AllowedMethods)
		}
		if opts.AllowedHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", opts.AllowedHeaders)
		}
		if opts.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}
}

// getAllowedOrigins retrieves the list of allowed origins from configuration.
func getAllowedOrigins() []string {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CORSMiddleware"))

	// Get origins from configuration
	runtimeConfig := config.GetThunderRuntime()
	originList := runtimeConfig.Config.CORS.AllowedOrigins

	if len(originList) == 0 {
		logger.Debug("No allowed origins configured in deployment.yaml")
		return []string{} // Return empty list if no origins configured
	}

	logger.Debug("Using allowed origins from configuration", log.Int("count", len(originList)))
	return originList
}
