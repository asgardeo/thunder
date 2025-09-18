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

// Package server provides server wide operations and utilities.
package server

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// ServerOperationServiceInterface defines the interface for server operations.
type ServerOperationServiceInterface interface {
	WrapHandleFunction(mux *http.ServeMux, pattern string, options *RequestWrapOptions,
		handlerFunc http.HandlerFunc)
}

// ServerOperationService implements the ServerOperationServiceInterface.
type ServerOperationService struct {
	OriginCache cache.CacheInterface[[]string]
}

// NewServerOperationService creates a new instance of ServerOperationService.
func NewServerOperationService() ServerOperationServiceInterface {
	return &ServerOperationService{
		OriginCache: cache.GetCache[[]string]("OriginCache"),
	}
}

// WrapHandleFunction wraps the provided handler function with pre-request processing and registers it with the mux.
func (ops *ServerOperationService) WrapHandleFunction(mux *http.ServeMux, pattern string, options *RequestWrapOptions,
	handlerFunc http.HandlerFunc) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ops.addAllowedOriginHeaders(w, r, options)

		// Return the handler function
		handlerFunc(w, r)
	})
}

// getAllowedOrigins retrieves the list of allowed origins from configuration.
func (ops *ServerOperationService) getAllowedOrigins() []string {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ServerOperationService"))

	// TODO: Revisit this when adding support for the organization concept.
	cacheKey := cache.CacheKey{
		Key: "origins",
	}
	originList, ok := ops.OriginCache.Get(cacheKey)

	if !ok {
		// Get origins from configuration
		runtimeConfig := config.GetThunderRuntime()
		originList = runtimeConfig.Config.CORS.AllowedOrigins

		if len(originList) == 0 {
			logger.Debug("No allowed origins configured in deployment.yaml")
			originList = []string{} // Return empty list if no origins configured
		} else {
			logger.Debug("Using allowed origins from configuration", log.Int("count", len(originList)))
		}

		if err := ops.OriginCache.Set(cacheKey, originList); err != nil {
			logger.Error("Failed to cache allowed origins", log.Error(err))
		}
	}

	return originList
}

// addAllowedOriginHeaders sets the CORS headers for the response based on the configured allowed origins.
func (ops *ServerOperationService) addAllowedOriginHeaders(w http.ResponseWriter, r *http.Request,
	options *RequestWrapOptions) {
	allowedOrigins := ops.getAllowedOrigins()

	requestOrigin := r.Header.Get("Origin")
	if requestOrigin == "" {
		return
	}

	// Set the CORS headers if allowed origins are configured.
	allowedOrigin := utils.GetAllowedOrigin(allowedOrigins, requestOrigin)
	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)

		if options.Cors.AllowedMethods != "" {
			w.Header().Set("Access-Control-Allow-Methods", options.Cors.AllowedMethods)
		}
		if options.Cors.AllowedHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", options.Cors.AllowedHeaders)
		}
		if options.Cors.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}
}
