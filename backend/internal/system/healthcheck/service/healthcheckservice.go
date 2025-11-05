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

// Package service provides health check-related business logic and operations.
package service

import (
	"sync"

	"github.com/asgardeo/thunder/internal/system/database"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/healthcheck/model"
	"github.com/asgardeo/thunder/internal/system/log"
)

var (
	instance *HealthCheckService
	once     sync.Once
)

// HealthCheckServiceInterface defines the interface for the health check service.
type HealthCheckServiceInterface interface {
	CheckReadiness() model.ServerStatus
}

// HealthCheckService is the default implementation of the HealthCheckServiceInterface.
type HealthCheckService struct {
	DBProvider provider.DBProviderInterface
}

// GetHealthCheckService returns a singleton instance of HealthCheckService.
func GetHealthCheckService() HealthCheckServiceInterface {
	once.Do(func() {
		instance = &HealthCheckService{
			DBProvider: provider.GetDBProvider(),
		}
	})
	return instance
}

// CheckReadiness checks the readiness of the server and its dependencies.
func (hcs *HealthCheckService) CheckReadiness() model.ServerStatus {
	configDBStatus := model.ServiceStatus{
		ServiceName: "ConfigDB",
		Status:      hcs.checkDatabaseStatus(database.ConfigDBClientName, queryConfigDBTable),
	}

	runtimeDBStatus := model.ServiceStatus{
		ServiceName: "RuntimeDB",
		Status:      hcs.checkDatabaseStatus(database.RuntimeDBClientName, queryRuntimeDBTable),
	}

	userDBStatus := model.ServiceStatus{
		ServiceName: "UserDB",
		Status:      hcs.checkDatabaseStatus(database.UserDBClientName, queryUserDBTable),
	}

	status := model.StatusUp
	if configDBStatus.Status == model.StatusDown ||
		runtimeDBStatus.Status == model.StatusDown ||
		userDBStatus.Status == model.StatusDown {
		status = model.StatusDown
	}
	return model.ServerStatus{
		Status: status,
		ServiceStatus: []model.ServiceStatus{
			configDBStatus,
			runtimeDBStatus,
			userDBStatus,
		},
	}
}

// checkDatabaseStatus checks the status of the specified database with the specified query.
func (hcs *HealthCheckService) checkDatabaseStatus(dbname string, query dbmodel.DBQuery) model.Status {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "HealthCheckService"))
	var dbClient provider.DBClientInterface
	var err error
	switch dbname {
	case database.ConfigDBClientName:
		dbClient, err = hcs.DBProvider.GetConfigDBClient()
		if err != nil {
			logger.Error("Failed to get database client", log.Error(err))
			return model.StatusDown
		}

	case database.RuntimeDBClientName:
		dbClient, err = hcs.DBProvider.GetRuntimeDBClient()
		if err != nil {
			logger.Error("Failed to get database client", log.Error(err))
			return model.StatusDown
		}
	case database.UserDBClientName:
		dbClient, err = hcs.DBProvider.GetUserDBClient()
		if err != nil {
			logger.Error("Failed to get database client", log.Error(err))
			return model.StatusDown
		}
	default:
		return model.StatusDown
	}

	_, err = dbClient.Query(query)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return model.StatusDown
	}
	return model.StatusUp
}
