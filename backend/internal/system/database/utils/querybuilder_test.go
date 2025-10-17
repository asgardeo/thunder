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

package utils

import (
	"testing"

	"github.com/asgardeo/thunder/internal/system/database/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testBaseQuery  = "SELECT * FROM users"
	testColumnName = "attributes"
)

type QueryBuilderTestSuite struct {
	suite.Suite
}

func TestQueryBuilderSuite(t *testing.T) {
	suite.Run(t, new(QueryBuilderTestSuite))
}

func (suite *QueryBuilderTestSuite) TestBuildFilterQuery() {
	queryID := "test_query"
	baseQuery := testBaseQuery
	columnName := testColumnName
	filters := map[string]interface{}{
		"role": "admin",
		"age":  30,
	}

	query, args, err := BuildFilterQuery(queryID, baseQuery, columnName, filters, nil)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), queryID, query.ID)
	assert.Len(suite.T(), args, 2)

	// Verify args order due to sorting of keys
	assert.Equal(suite.T(), int(30), args[0])
	assert.Equal(suite.T(), "admin", args[1])

	// Test Postgres query
	postgresQuery := query.GetQuery("postgres")
	assert.Contains(suite.T(), postgresQuery, baseQuery)
	assert.Contains(suite.T(), postgresQuery, "attributes->>'age' = $1")
	assert.Contains(suite.T(), postgresQuery, "attributes->>'role' = $2")

	// Test SQLite query
	sqliteQuery := query.GetQuery("sqlite")
	assert.Contains(suite.T(), sqliteQuery, baseQuery)
	assert.Contains(suite.T(), sqliteQuery, "json_extract(attributes, '$.age') = ?")
	assert.Contains(suite.T(), sqliteQuery, "json_extract(attributes, '$.role') = ?")

	// Test default query (should return PostgreSQL query)
	defaultQuery := query.GetQuery("unknown")
	assert.Equal(suite.T(), postgresQuery, defaultQuery)
}

func (suite *QueryBuilderTestSuite) TestBuildFilterQueryWithEmptyFilters() {
	queryID := "empty_filters"
	baseQuery := testBaseQuery
	columnName := testColumnName
	filters := map[string]interface{}{}

	query, args, err := BuildFilterQuery(queryID, baseQuery, columnName, filters, nil)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), queryID, query.ID)
	assert.Empty(suite.T(), args)

	// Both Postgres and SQLite queries should be the same as base query when no filters
	postgresQuery := query.GetQuery("postgres")
	sqliteQuery := query.GetQuery("sqlite")
	assert.Equal(suite.T(), baseQuery, postgresQuery)
	assert.Equal(suite.T(), baseQuery, sqliteQuery)
	assert.Equal(suite.T(), baseQuery, query.Query)
}

func (suite *QueryBuilderTestSuite) TestBuildFilterQueryWithInvalidColumnName() {
	queryID := "invalid_column"
	baseQuery := testBaseQuery
	columnName := "attributes;DROP TABLE users"
	filters := map[string]interface{}{
		"role": "admin",
	}

	query, args, err := BuildFilterQuery(queryID, baseQuery, columnName, filters, nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid column name")
	assert.Equal(suite.T(), model.DBQuery{}, query)
	assert.Nil(suite.T(), args)
}

func (suite *QueryBuilderTestSuite) TestBuildFilterQueryWithInvalidFilterKey() {
	queryID := "invalid_filter_key"
	baseQuery := testBaseQuery
	columnName := testColumnName
	filters := map[string]interface{}{
		"valid":              "value",
		"invalid-filter-key": "value", // Contains invalid character '-'
	}

	query, args, err := BuildFilterQuery(queryID, baseQuery, columnName, filters, nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid filter key")
	assert.Equal(suite.T(), model.DBQuery{}, query)
	assert.Nil(suite.T(), args)
}

func (suite *QueryBuilderTestSuite) TestValidateKey() {
	validKeys := []string{
		"name",
		"user_id",
		"role123",
		"UPPERCASE",
		"mixedCASE",
		"with_underscore",
		"_leading_underscore",
		"trailing_underscore_",
	}

	for _, key := range validKeys {
		err := validateKey(key)
		assert.NoError(suite.T(), err, "Key should be valid: %s", key)
	}
}

func (suite *QueryBuilderTestSuite) TestValidateKeyInvalid() {
	invalidKeys := []string{
		"space key",
		"hyphen-key",
		"special!char",
		"sql;injection",
		"quote'test",
		"double\"quote",
	}

	for _, key := range invalidKeys {
		err := validateKey(key)
		assert.Error(suite.T(), err, "Key should be invalid: %s", key)
		assert.Contains(suite.T(), err.Error(), "invalid characters")
	}
}

func (suite *QueryBuilderTestSuite) TestBuildFilterQueryDatabaseSpecificQueries() {
	queryID := "db_specific_test"
	baseQuery := "SELECT USER_ID FROM \"USER\" WHERE 1=1"
	columnName := "ATTRIBUTES"
	filters := map[string]interface{}{
		"email": "test@example.com",
		"name":  "John Doe",
	}

	query, args, err := BuildFilterQuery(queryID, baseQuery, columnName, filters, nil)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), queryID, query.ID)
	assert.Len(suite.T(), args, 2)

	// Verify arguments are in sorted order (email, name)
	assert.Equal(suite.T(), "test@example.com", args[0])
	assert.Equal(suite.T(), "John Doe", args[1])

	// Test PostgreSQL-specific query
	postgresQuery := query.GetQuery("postgres")
	expectedPostgres := "SELECT USER_ID FROM \"USER\" WHERE 1=1" +
		" AND ATTRIBUTES->>'email' = $1" +
		" AND ATTRIBUTES->>'name' = $2"
	assert.Equal(suite.T(), expectedPostgres, postgresQuery)

	// Test SQLite-specific query
	sqliteQuery := query.GetQuery("sqlite")
	expectedSQLite := "SELECT USER_ID FROM \"USER\" WHERE 1=1" +
		" AND json_extract(ATTRIBUTES, '$.email') = ?" +
		" AND json_extract(ATTRIBUTES, '$.name') = ?"
	assert.Equal(suite.T(), expectedSQLite, sqliteQuery)

	// Test that both queries are stored in the struct
	assert.Equal(suite.T(), expectedPostgres, query.PostgresQuery)
	assert.Equal(suite.T(), expectedSQLite, query.SQLiteQuery)
	assert.Equal(suite.T(), expectedPostgres, query.Query) // Default should be PostgreSQL
}

func (suite *QueryBuilderTestSuite) TestBuildFilterQuerySingleFilter() {
	queryID := "single_filter"
	baseQuery := "SELECT * FROM users WHERE active = true"
	columnName := "metadata"
	filters := map[string]interface{}{
		"department": "engineering",
	}

	query, args, err := BuildFilterQuery(queryID, baseQuery, columnName, filters, nil)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), args, 1)
	assert.Equal(suite.T(), "engineering", args[0])

	// PostgreSQL query
	postgresQuery := query.GetQuery("postgres")
	expectedPostgres := "SELECT * FROM users WHERE active = true" +
		" AND metadata->>'department' = $1"
	assert.Equal(suite.T(), expectedPostgres, postgresQuery)

	// SQLite query
	sqliteQuery := query.GetQuery("sqlite")
	expectedSQLite := "SELECT * FROM users WHERE active = true" +
		" AND json_extract(metadata, '$.department') = ?"
	assert.Equal(suite.T(), expectedSQLite, sqliteQuery)
}
