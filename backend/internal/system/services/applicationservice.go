/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package services provides HTTP service implementations for various domain operations.
//
//nolint:dupl // ApplicationService has similar structure to GroupService but they serve different domains
package services

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application/handler"
	"github.com/asgardeo/thunder/internal/system/server"
)

// ApplicationService defines the service for handling application-related requests.
type ApplicationService struct {
	ServerOpsService   server.ServerOperationServiceInterface
	applicationHandler *handler.ApplicationHandler
}

// NewApplicationService creates a new instance of ApplicationService.
func NewApplicationService(mux *http.ServeMux) ServiceInterface {
	instance := &ApplicationService{
		ServerOpsService:   server.NewServerOperationService(),
		applicationHandler: handler.NewApplicationHandler(),
	}
	instance.RegisterRoutes(mux)

	return instance
}

// RegisterRoutes registers the routes for the ApplicationService.
func (s *ApplicationService) RegisterRoutes(mux *http.ServeMux) {
	opts1 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, POST",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	s.ServerOpsService.WrapHandleFunction(mux, "POST /applications", &opts1,
		s.applicationHandler.HandleApplicationPostRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "GET /applications", &opts1,
		s.applicationHandler.HandleApplicationListRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /applications", &opts1,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

	opts2 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, PUT, DELETE",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	s.ServerOpsService.WrapHandleFunction(mux, "GET /applications/", &opts2,
		s.applicationHandler.HandleApplicationGetRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "PUT /applications/", &opts2,
		s.applicationHandler.HandleApplicationPutRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "DELETE /applications/", &opts2,
		s.applicationHandler.HandleApplicationDeleteRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /applications/", &opts2,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
}
