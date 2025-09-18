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

// Package services handles the registration of routes and services for the system.
//
//nolint:dupl // Ignoring false positive duplicate code
package services

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/idp/handler"
	"github.com/asgardeo/thunder/internal/system/server"
)

// IDPService is the service for identity provider management operations.
type IDPService struct {
	ServerOpsService server.ServerOperationServiceInterface
	idpHandler       *handler.IDPHandler
}

// NewIDPService creates a new instance of IDPService.
func NewIDPService(mux *http.ServeMux) ServiceInterface {
	instance := &IDPService{
		ServerOpsService: server.NewServerOperationService(),
		idpHandler:       handler.NewIDPHandler(),
	}
	instance.RegisterRoutes(mux)

	return instance
}

// RegisterRoutes registers the routes for identity provider operations.
func (s *IDPService) RegisterRoutes(mux *http.ServeMux) {
	opts1 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, POST",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	s.ServerOpsService.WrapHandleFunction(mux, "POST /identity-providers", &opts1, s.idpHandler.HandleIDPPostRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "GET /identity-providers", &opts1, s.idpHandler.HandleIDPListRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /identity-providers", &opts1,
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
	s.ServerOpsService.WrapHandleFunction(mux, "GET /identity-providers/{id}", &opts2,
		s.idpHandler.HandleIDPGetRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "PUT /identity-providers/{id}", &opts2,
		s.idpHandler.HandleIDPPutRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "DELETE /identity-providers/{id}", &opts2,
		s.idpHandler.HandleIDPDeleteRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /identity-providers/{id}", &opts2,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
}
