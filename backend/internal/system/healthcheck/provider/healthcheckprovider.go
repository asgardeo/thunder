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

// Package provider provides functionality for managing health check service instances.
package provider

import (
	"github.com/asgardeo/thunder/internal/system/healthcheck/service"
)

// HealthCheckProviderInterface defines the interface for the health check provider.
type HealthCheckProviderInterface interface {
	GetHealthCheckService() service.HealthCheckServiceInterface
}

// HealthCheckProvider is the default implementation of the HealthCheckProviderInterface.
type HealthCheckProvider struct{}

// NewHealthCheckProvider creates a new instance of HealthCheckProvider.
func NewHealthCheckProvider() HealthCheckProviderInterface {
	return &HealthCheckProvider{}
}

// GetHealthCheckService returns the health check service instance.
func (hcp *HealthCheckProvider) GetHealthCheckService() service.HealthCheckServiceInterface {
	return service.GetHealthCheckService()
}
