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

	"github.com/asgardeo/thunder/internal/system/healthcheck/model"
	"github.com/asgardeo/thunder/tests/mocks/database/clientmock"
	dbprovidermock "github.com/asgardeo/thunder/tests/mocks/database/providermock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HealthCheckServiceTestSuite struct {
	suite.Suite
	service        HealthCheckServiceInterface
	mockDBProvider *dbprovidermock.DBProviderInterfaceMock
	mockIdentityDB *clientmock.DBClientInterfaceMock
	mockRuntimeDB  *clientmock.DBClientInterfaceMock
}

func TestHealthCheckServiceSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckServiceTestSuite))
}

func (suite *HealthCheckServiceTestSuite) SetupTest() {
	instance = nil
	once = sync.Once{}
	suite.service = GetHealthCheckService()
}

func (suite *HealthCheckServiceTestSuite) BeforeTest(suiteName, testName string) {
	dbClientIdentity := &clientmock.DBClientInterfaceMock{}
	suite.mockIdentityDB = dbClientIdentity

	dbClientRuntime := &clientmock.DBClientInterfaceMock{}
	suite.mockRuntimeDB = dbClientRuntime

	dbProvider := &dbprovidermock.DBProviderInterfaceMock{}
	dbProvider.On("GetDBClient", "identity").Return(dbClientIdentity, nil)
	dbProvider.On("GetDBClient", "runtime").Return(dbClientRuntime, nil)
	suite.mockDBProvider = dbProvider
	suite.service.(*HealthCheckService).DBProvider = dbProvider
}

func (suite *HealthCheckServiceTestSuite) TestCheckReadiness() {
	testCases := []struct {
		name                 string
		setupIdentityDB      func()
		setupRuntimeDB       func()
		expectedStatus       model.Status
		expectedServiceCount int
	}{
		{
			name: "AllDatabasesUp",
			setupIdentityDB: func() {
				suite.mockIdentityDB.On("Query", queryConfigDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			setupRuntimeDB: func() {
				suite.mockRuntimeDB.On("Query", queryRuntimeDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			expectedStatus:       model.StatusUp,
			expectedServiceCount: 2,
		},
		{
			name: "IdentityDBDown",
			setupIdentityDB: func() {
				suite.mockIdentityDB.On("Query", queryConfigDBTable).Return(nil, errors.New("database error"))
			},
			setupRuntimeDB: func() {
				suite.mockRuntimeDB.On("Query", queryRuntimeDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			expectedStatus:       model.StatusDown,
			expectedServiceCount: 2,
		},
		{
			name: "RuntimeDBDown",
			setupIdentityDB: func() {
				suite.mockIdentityDB.On("Query", queryConfigDBTable).Return([]map[string]interface{}{
					{"1": 1}}, nil)
			},
			setupRuntimeDB: func() {
				suite.mockRuntimeDB.On("Query", queryRuntimeDBTable).Return(nil, errors.New("database error"))
			},
			expectedStatus:       model.StatusDown,
			expectedServiceCount: 2,
		},
		{
			name: "BothDBsDown",
			setupIdentityDB: func() {
				suite.mockIdentityDB.On("Query", queryConfigDBTable).Return(nil, errors.New("database error"))
			},
			setupRuntimeDB: func() {
				suite.mockRuntimeDB.On("Query", queryRuntimeDBTable).Return(nil, errors.New("database error"))
			},
			expectedStatus:       model.StatusDown,
			expectedServiceCount: 2,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mock expectations
			suite.mockIdentityDB.ExpectedCalls = nil
			suite.mockRuntimeDB.ExpectedCalls = nil
			suite.mockIdentityDB.On("Close").Return(nil)
			suite.mockRuntimeDB.On("Close").Return(nil)

			// Setup database mocks
			if tc.setupIdentityDB != nil {
				tc.setupIdentityDB()
			}
			if tc.setupRuntimeDB != nil {
				tc.setupRuntimeDB()
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

			// If identity DB is expected down, verify it's reported as down
			if tc.name == "IdentityDBDown" || tc.name == "IdentityDBClientError" || tc.name == "BothDBsDown" {
				for _, status := range serverStatus.ServiceStatus {
					if status.ServiceName == "IdentityDB" {
						assert.Equal(t, model.StatusDown, status.Status, "IdentityDB should be DOWN")
					}
				}
			}

			// If runtime DB is expected down, verify it's reported as down
			if tc.name == "RuntimeDBDown" || tc.name == "RuntimeDBClientError" || tc.name == "BothDBsDown" {
				for _, status := range serverStatus.ServiceStatus {
					if status.ServiceName == "RuntimeDB" {
						assert.Equal(t, model.StatusDown, status.Status, "RuntimeDB should be DOWN")
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
	suite.mockDBProvider.On("GetDBClient", "identity").Return(nil, errors.New("failed to get identity DB client"))
	suite.mockDBProvider.On("GetDBClient", "runtime").Return(nil, errors.New("failed to get runtime DB client"))

	// Execute the method being tested
	serverStatus := suite.service.CheckReadiness()

	// Assertions
	assert.Equal(suite.T(), model.StatusDown, serverStatus.Status, "Server status should be DOWN")
	assert.Len(suite.T(), serverStatus.ServiceStatus, 2, "There should be two service statuses reported")

	for _, status := range serverStatus.ServiceStatus {
		if status.ServiceName == "IdentityDB" {
			assert.Equal(suite.T(), model.StatusDown, status.Status, "IdentityDB should be DOWN")
		} else if status.ServiceName == "RuntimeDB" {
			assert.Equal(suite.T(), model.StatusDown, status.Status, "RuntimeDB should be DOWN")
		}
	}

	suite.mockDBProvider.AssertExpectations(suite.T())
}
