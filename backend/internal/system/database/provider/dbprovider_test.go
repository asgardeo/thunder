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

package provider

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/model"
)

type DBProviderTestSuite struct {
	suite.Suite
	mockDB sqlmock.Sqlmock
}

func TestDBProviderTestSuite(t *testing.T) {
	suite.Run(t, new(DBProviderTestSuite))
}

func (suite *DBProviderTestSuite) SetupTest() {
	_, mock, err := sqlmock.New()
	suite.Require().NoError(err)
	suite.mockDB = mock

	// Reset global config before each test
	config.ResetThunderRuntime()

	// Initialize a dummy config
	dummyConfig := &config.Config{
		Database: config.DatabaseConfig{
			Identity: config.DataSource{Name: "identity", Type: "postgres"},
			Runtime:  config.DataSource{Name: "runtime", Type: "postgres"},
			User:     config.DataSource{Name: "user", Type: "postgres"},
		},
	}
	err = config.InitializeThunderRuntime(".", dummyConfig)
	suite.Require().NoError(err)
}

func (suite *DBProviderTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func (suite *DBProviderTestSuite) TestGetUserDBTransactioner_Success() {
	// Create a mock DB connection
	db, _, err := sqlmock.New()
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// Manually construct the provider with an initialized client
	provider := &dbProvider{
		userClient: NewDBClient(model.NewDB(db), "postgres", "user"),
	}

	// Test getting the transactioner
	txer, err := provider.GetUserDBTransactioner()
	suite.NoError(err)
	suite.NotNil(txer)
}

func (suite *DBProviderTestSuite) TestGetRuntimeDBTransactioner_Success() {
	// Create a mock DB connection
	db, _, err := sqlmock.New()
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// Manually construct the provider with an initialized client
	provider := &dbProvider{
		runtimeClient: NewDBClient(model.NewDB(db), "postgres", "runtime"),
	}

	// Test getting the transactioner
	txer, err := provider.GetRuntimeDBTransactioner()
	suite.NoError(err)
	suite.NotNil(txer)
}
