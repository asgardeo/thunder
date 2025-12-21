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

package model

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DBQueryTestSuite struct {
	suite.Suite
}

func TestDBQueryTestSuite(t *testing.T) {
	suite.Run(t, new(DBQueryTestSuite))
}

func (suite *DBQueryTestSuite) TestGetID() {
	query := DBQuery{
		ID:    "TEST-001",
		Query: "SELECT * FROM users",
	}

	suite.Equal("TEST-001", query.GetID())
}

func (suite *DBQueryTestSuite) TestGetQuery_DefaultQuery() {
	query := DBQuery{
		ID:    "TEST-002",
		Query: "SELECT * FROM users",
	}

	suite.Equal("SELECT * FROM users", query.GetQuery("postgres"))
	suite.Equal("SELECT * FROM users", query.GetQuery("sqlite"))
	suite.Equal("SELECT * FROM users", query.GetQuery("mysql"))
}

func (suite *DBQueryTestSuite) TestGetQuery_PostgresQuery() {
	query := DBQuery{
		ID:            "TEST-003",
		Query:         "SELECT * FROM users WHERE id = ?",
		PostgresQuery: "SELECT * FROM users WHERE id = $1",
	}

	suite.Equal("SELECT * FROM users WHERE id = $1", query.GetQuery("postgres"))
	suite.Equal("SELECT * FROM users WHERE id = ?", query.GetQuery("sqlite"))
}

func (suite *DBQueryTestSuite) TestGetQuery_SQLiteQuery() {
	query := DBQuery{
		ID:          "TEST-004",
		Query:       "SELECT * FROM users WHERE id = $1",
		SQLiteQuery: "SELECT * FROM users WHERE id = ?",
	}

	suite.Equal("SELECT * FROM users WHERE id = $1", query.GetQuery("postgres"))
	suite.Equal("SELECT * FROM users WHERE id = ?", query.GetQuery("sqlite"))
}

func (suite *DBQueryTestSuite) TestGetQuery_BothSpecificQueries() {
	query := DBQuery{
		ID:            "TEST-005",
		Query:         "SELECT * FROM users",
		PostgresQuery: "SELECT * FROM users WHERE id = $1",
		SQLiteQuery:   "SELECT * FROM users WHERE id = ?",
	}

	suite.Equal("SELECT * FROM users WHERE id = $1", query.GetQuery("postgres"))
	suite.Equal("SELECT * FROM users WHERE id = ?", query.GetQuery("sqlite"))
	suite.Equal("SELECT * FROM users", query.GetQuery("mysql"))
}

func (suite *DBQueryTestSuite) TestGetQuery_EmptySpecificQueries() {
	query := DBQuery{
		ID:            "TEST-006",
		Query:         "SELECT * FROM users",
		PostgresQuery: "",
		SQLiteQuery:   "",
	}

	suite.Equal("SELECT * FROM users", query.GetQuery("postgres"))
	suite.Equal("SELECT * FROM users", query.GetQuery("sqlite"))
}

func (suite *DBQueryTestSuite) TestGetQuery_UnknownDBType() {
	query := DBQuery{
		ID:            "TEST-007",
		Query:         "SELECT * FROM users",
		PostgresQuery: "SELECT * FROM users WHERE id = $1",
		SQLiteQuery:   "SELECT * FROM users WHERE id = ?",
	}

	suite.Equal("SELECT * FROM users", query.GetQuery("unknown"))
	suite.Equal("SELECT * FROM users", query.GetQuery(""))
}

func (suite *DBQueryTestSuite) TestDBQuery_ImplementsInterface() {
	var _ DBQueryInterface = (*DBQuery)(nil)
}
