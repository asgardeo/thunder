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

package service

import (
	"errors"
	"sync"
	"testing"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/healthcheck/model"

	dbprovidermock "github.com/asgardeo/thunder/tests/mocks/database/providermock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HealthCheckServiceTestSuite struct {
	suite.Suite
	service        HealthCheckServiceInterface
	mockDBProvider *dbprovidermock.DBProviderInterfaceMock
	mockIdentityDB *dbprovidermock.DBClientInterfaceMock
	mockRuntimeDB  *dbprovidermock.DBClientInterfaceMock
	mockUserDB     *dbprovidermock.DBClientInterfaceMock
}

func TestHealthCheckServiceSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckServiceTestSuite))
}

func (suite *HealthCheckServiceTestSuite) SetupTest() {
	testConfig := &config.Config{
		Database: config.DatabaseConfig{
			Identity: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
			Runtime: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
			User: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	instance = nil
	once = sync.Once{}
	suite.service = GetHealthCheckService()
}

func (suite *HealthCheckServiceTestSuite) BeforeTest(suiteName, testName string) {
	dbClientIdentity := &dbprovidermock.DBClientInterfaceMock{}
	suite.mockIdentityDB = dbClientIdentity

	dbClientRuntime := &dbprovidermock.DBClientInterfaceMock{}
	suite.mockRuntimeDB = dbClientRuntime

	dbClientUser := &dbprovidermock.DBClientInterfaceMock{}
	suite.mockUserDB = dbClientUser

	dbProvider := &dbprovidermock.DBProviderInterfaceMock{}
	dbProvider.On("GetConfigDBClient").Return(dbClientIdentity, nil)
	dbProvider.On("GetRuntimeDBClient").Return(dbClientRuntime, nil)
	dbProvider.On("GetUserDBClient").Return(dbClientUser, nil)
	suite.mockDBProvider = dbProvider
	suite.service.(*HealthCheckService).DBProvider = dbProvider
}

func (suite *HealthCheckServiceTestSuite) TestCheckReadiness() {
	const (
		tcAllDBUp        = "AllDatabasesUp"
		tcIdentityDBDown = "IdentityDBDown"
		tcRuntimeDBDown  = "RuntimeDBDown"
		tcUserDBDown     = "UserDBDown"
		tcAllThreeDBDown = "AllThreeDBsDown"
	)
	testCases := []struct {
		name                 string
		setupIdentityDB      func()
		setupRuntimeDB       func()
		setupUserDB          func()
		expectedStatus       model.Status
		expectedServiceCount int
	}{
		{
			name: tcAllDBUp,
			setupIdentityDB: func() {
				suite.mockIdentityDB.On("Query", queryConfigDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			setupRuntimeDB: func() {
				suite.mockRuntimeDB.On("Query", queryRuntimeDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			setupUserDB: func() {
				suite.mockUserDB.On("Query", queryUserDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			expectedStatus:       model.StatusUp,
			expectedServiceCount: 3,
		},
		{
			name: tcIdentityDBDown,
			setupIdentityDB: func() {
				suite.mockIdentityDB.On("Query", queryConfigDBTable).Return(nil, errors.New("database error"))
			},
			setupRuntimeDB: func() {
				suite.mockRuntimeDB.On("Query", queryRuntimeDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			setupUserDB: func() {
				suite.mockUserDB.On("Query", queryUserDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			expectedStatus:       model.StatusDown,
			expectedServiceCount: 3,
		},
		{
			name: tcRuntimeDBDown,
			setupIdentityDB: func() {
				suite.mockIdentityDB.On("Query", queryConfigDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			setupRuntimeDB: func() {
				suite.mockRuntimeDB.On("Query", queryRuntimeDBTable).Return(nil, errors.New("database error"))
			},
			setupUserDB: func() {
				suite.mockUserDB.On("Query", queryUserDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			expectedStatus:       model.StatusDown,
			expectedServiceCount: 3,
		},
		{
			name: tcUserDBDown,
			setupIdentityDB: func() {
				suite.mockIdentityDB.On("Query", queryConfigDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			setupRuntimeDB: func() {
				suite.mockRuntimeDB.On("Query", queryRuntimeDBTable).Return(nil, errors.New("database error"))
			},
			setupUserDB: func() {
				suite.mockUserDB.On("Query", queryUserDBTable).Return(nil, errors.New("database error"))
			},
			expectedStatus:       model.StatusDown,
			expectedServiceCount: 3,
		},
		{
			name: tcAllThreeDBDown,
			setupIdentityDB: func() {
				suite.mockIdentityDB.On("Query", queryConfigDBTable).Return(nil, errors.New("database error"))
			},
			setupRuntimeDB: func() {
				suite.mockRuntimeDB.On("Query", queryRuntimeDBTable).Return(nil, errors.New("database error"))
			},
			setupUserDB: func() {
				suite.mockUserDB.On("Query", queryUserDBTable).Return(nil, errors.New("database error"))
			},
			expectedStatus:       model.StatusDown,
			expectedServiceCount: 3,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mock expectations
			suite.mockIdentityDB.ExpectedCalls = nil
			suite.mockRuntimeDB.ExpectedCalls = nil
			suite.mockUserDB.ExpectedCalls = nil

			// Setup database mocks
			if tc.setupIdentityDB != nil {
				tc.setupIdentityDB()
			}
			if tc.setupRuntimeDB != nil {
				tc.setupRuntimeDB()
			}
			if tc.setupUserDB != nil {
				tc.setupUserDB()
			}

			// Execute the method being tested
			serverStatus := suite.service.CheckReadiness()

			// Assertions
			assert.Equal(t, tc.expectedStatus, serverStatus.Status, "Server status should match expected")
			assert.Equal(t, tc.expectedServiceCount, len(serverStatus.ServiceStatus),
				"Service status count should match expected")

			serviceNames := make(map[string]bool)
			for _, status := range serverStatus.ServiceStatus {
				serviceNames[status.ServiceName] = true
			}
			assert.True(t, serviceNames["IdentityDB"], "IdentityDB service status should be present")
			assert.True(t, serviceNames["RuntimeDB"], "RuntimeDB service status should be present")
			assert.True(t, serviceNames["UserDB"], "UserDB service status should be present")

			// If identity DB is expected down, verify it's reported as down
			if tc.name == tcIdentityDBDown || tc.name == "IdentityDBClientError" || tc.name == tcAllThreeDBDown {
				for _, status := range serverStatus.ServiceStatus {
					if status.ServiceName == "IdentityDB" {
						assert.Equal(t, model.StatusDown, status.Status, "IdentityDB should be DOWN")
					}
				}
			}

			// If runtime DB is expected down, verify it's reported as down
			if tc.name == tcRuntimeDBDown || tc.name == "RuntimeDBClientError" || tc.name == tcAllThreeDBDown {
				for _, status := range serverStatus.ServiceStatus {
					if status.ServiceName == "RuntimeDB" {
						assert.Equal(t, model.StatusDown, status.Status, "RuntimeDB should be DOWN")
					}
				}
			}

			// If user DB is expected down, verify it's reported as down
			if tc.name == tcUserDBDown || tc.name == "UserDBClientError" || tc.name == tcAllThreeDBDown {
				for _, status := range serverStatus.ServiceStatus {
					if status.ServiceName == "UserDB" {
						assert.Equal(t, model.StatusDown, status.Status, "UserDB should be DOWN")
					}
				}
			}

			// Verify that the mock expectations were met
			suite.mockDBProvider.AssertExpectations(t)
			suite.mockIdentityDB.AssertExpectations(t)
			suite.mockRuntimeDB.AssertExpectations(t)
		})
	}
}

func (suite *HealthCheckServiceTestSuite) TestCheckReadiness_DBRetrievalError() {
	suite.mockDBProvider.ExpectedCalls = nil
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, errors.New("failed to get identity DB client"))
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(nil, errors.New("failed to get runtime DB client"))
	suite.mockDBProvider.On("GetUserDBClient").Return(nil, errors.New("failed to get user DB client"))

	// Execute the method being tested
	serverStatus := suite.service.CheckReadiness()

	// Assertions
	assert.Equal(suite.T(), model.StatusDown, serverStatus.Status, "Server status should be DOWN")
	assert.Len(suite.T(), serverStatus.ServiceStatus, 3, "There should be three service statuses reported")

	for _, status := range serverStatus.ServiceStatus {
		if status.ServiceName == "IdentityDB" {
			assert.Equal(suite.T(), model.StatusDown, status.Status, "IdentityDB should be DOWN")
		} else if status.ServiceName == "RuntimeDB" {
			assert.Equal(suite.T(), model.StatusDown, status.Status, "RuntimeDB should be DOWN")
		} else if status.ServiceName == "UserDB" {
			assert.Equal(suite.T(), model.StatusDown, status.Status, "UserDB should be DOWN")
		}
	}

	suite.mockDBProvider.AssertExpectations(suite.T())
}
