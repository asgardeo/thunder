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

// Package managers provides functionality for managing and registering system services.
package managers

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/services"
)

// ServiceManagerInterface defines the interface for managing services.
type ServiceManagerInterface interface {
	RegisterServices() error
}

// ServiceManager implements the ServiceManagerInterface and is responsible for registering services.
type ServiceManager struct {
	mux *http.ServeMux
}

// NewServiceManager creates a new instance of ServiceManager.
func NewServiceManager(mux *http.ServeMux) ServiceManagerInterface {
	return &ServiceManager{
		mux: mux,
	}
}

// RegisterServices registers all the services with the provided HTTP multiplexer.
func (sm *ServiceManager) RegisterServices() error {
	// Register the health service.
	services.NewHealthCheckService(sm.mux)

	// Register the token service.
	services.NewTokenService(sm.mux)

	// Register the authorization service.
	services.NewAuthorizationService(sm.mux)

	// Register the JWKS service.
	services.NewJWKSAPIService(sm.mux)

	// Register the introspection service.
	services.NewIntrospectionAPIService(sm.mux)

	// Register the Organization Unit service.
	services.NewOrganizationUnitService(sm.mux)

	// Register the User service.
	services.NewUserService(sm.mux)

	// Register the User Schema service.
	services.NewUserSchemaService(sm.mux)

	// Register the Group service.
	services.NewGroupService(sm.mux)

	// Register the Application service.
	services.NewApplicationService(sm.mux)

	// Register the identity provider service.
	services.NewIDPService(sm.mux)

	// Register the flow execution service.
	services.NewFlowExecutionService(sm.mux)

	// Register the notification sender service.
	services.NewNotificationSenderService(sm.mux)

	return nil
}
