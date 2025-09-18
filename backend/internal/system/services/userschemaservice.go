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

	"github.com/asgardeo/thunder/internal/system/server"
	"github.com/asgardeo/thunder/internal/userschema/handler"
)

// UserSchemaService is the service for user schema management operations.
type UserSchemaService struct {
	ServerOpsService  server.ServerOperationServiceInterface
	userSchemaHandler *handler.UserSchemaHandler
}

// NewUserSchemaService creates a new instance of UserSchemaService.
func NewUserSchemaService(mux *http.ServeMux) ServiceInterface {
	instance := &UserSchemaService{
		ServerOpsService:  server.NewServerOperationService(),
		userSchemaHandler: handler.NewUserSchemaHandler(),
	}
	instance.RegisterRoutes(mux)

	return instance
}

// RegisterRoutes registers the routes for user schema management operations.
func (s *UserSchemaService) RegisterRoutes(mux *http.ServeMux) {
	opts1 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, POST",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	s.ServerOpsService.WrapHandleFunction(mux, "POST /user-schemas", &opts1,
		s.userSchemaHandler.HandleUserSchemaPostRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "GET /user-schemas", &opts1,
		s.userSchemaHandler.HandleUserSchemaListRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /user-schemas", &opts1,
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
	s.ServerOpsService.WrapHandleFunction(mux, "GET /user-schemas/{id}", &opts2,
		s.userSchemaHandler.HandleUserSchemaGetRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "PUT /user-schemas/{id}", &opts2,
		s.userSchemaHandler.HandleUserSchemaPutRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "DELETE /user-schemas/{id}", &opts2,
		s.userSchemaHandler.HandleUserSchemaDeleteRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /user-schemas/{id}", &opts2,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
}
