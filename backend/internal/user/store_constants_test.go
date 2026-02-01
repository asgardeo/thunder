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

package user

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const testDeploymentIDForConstants = "test-deployment-id"

// StoreConstantsTestSuite is the test suite for store_constants.go functions.
type StoreConstantsTestSuite struct {
	suite.Suite
}

// TestStoreConstantsTestSuite runs the test suite.
func TestStoreConstantsTestSuite(t *testing.T) {
	suite.Run(t, new(StoreConstantsTestSuite))
}

// Test buildIdentifyQueryFromIndexedAttributes

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryFromIndexedAttributes_EmptyFilters() {
	query, args, err := buildIdentifyQueryFromIndexedAttributes(map[string]interface{}{}, testDeploymentIDForConstants)

	suite.Error(err)
	suite.Contains(err.Error(), "filters cannot be empty")
	suite.Nil(args)
	suite.Empty(query.Query)
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryFromIndexedAttributes_SingleFilter() {
	filters := map[string]interface{}{
		"username": "john.doe",
	}

	query, args, err := buildIdentifyQueryFromIndexedAttributes(filters, testDeploymentIDForConstants)

	suite.NoError(err)
	suite.Equal("ASQ-USER_MGT-16", query.ID)
	suite.Contains(query.Query, "SELECT DISTINCT ia1.USER_ID FROM USER_INDEXED_ATTRIBUTES ia1")
	suite.Contains(query.Query, "WHERE ia1.ATTRIBUTE_NAME = $1 AND ia1.ATTRIBUTE_VALUE = $2")
	suite.Contains(query.Query, "AND ia1.DEPLOYMENT_ID = $3")
	suite.Len(args, 3)
	suite.Equal("username", args[0])
	suite.Equal("john.doe", args[1])
	suite.Equal(testDeploymentIDForConstants, args[2])
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryFromIndexedAttributes_MultipleFilters() {
	filters := map[string]interface{}{
		"username": "john.doe",
		"email":    "john@example.com",
	}

	query, args, err := buildIdentifyQueryFromIndexedAttributes(filters, testDeploymentIDForConstants)

	suite.NoError(err)
	suite.Equal("ASQ-USER_MGT-16", query.ID)
	suite.Contains(query.Query, "SELECT DISTINCT ia1.USER_ID FROM USER_INDEXED_ATTRIBUTES ia1")
	suite.Contains(query.Query, "INNER JOIN USER_INDEXED_ATTRIBUTES ia2 ON ia1.USER_ID = ia2.USER_ID")
	suite.Contains(query.Query, "ia1.DEPLOYMENT_ID = ia2.DEPLOYMENT_ID")
	suite.Contains(query.Query, "WHERE")
	suite.Contains(query.Query, "AND ia1.DEPLOYMENT_ID = $5")
	suite.Len(args, 5)
	suite.Equal(testDeploymentIDForConstants, args[4])
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryFromIndexedAttributes_ThreeFilters() {
	filters := map[string]interface{}{
		"username":     "john.doe",
		"email":        "john@example.com",
		"mobileNumber": "1234567890",
	}

	query, args, err := buildIdentifyQueryFromIndexedAttributes(filters, testDeploymentIDForConstants)

	suite.NoError(err)
	suite.Equal("ASQ-USER_MGT-16", query.ID)
	suite.Contains(query.Query, "INNER JOIN USER_INDEXED_ATTRIBUTES ia2")
	suite.Contains(query.Query, "INNER JOIN USER_INDEXED_ATTRIBUTES ia3")
	suite.Len(args, 7) // 3 filters * 2 args each + 1 deployment ID
	suite.Equal(testDeploymentIDForConstants, args[6])
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryFromIndexedAttributes_DeterministicOrder() {
	filters := map[string]interface{}{
		"email":        "john@example.com",
		"username":     "john.doe",
		"mobileNumber": "1234567890",
	}

	query1, args1, err1 := buildIdentifyQueryFromIndexedAttributes(filters, testDeploymentIDForConstants)
	query2, args2, err2 := buildIdentifyQueryFromIndexedAttributes(filters, testDeploymentIDForConstants)

	suite.NoError(err1)
	suite.NoError(err2)
	suite.Equal(query1.Query, query2.Query, "Query should be deterministic")
	suite.Equal(args1, args2, "Args should be deterministic")
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryFromIndexedAttributes_IntegerValue() {
	filters := map[string]interface{}{
		"age": 30,
	}

	_, args, err := buildIdentifyQueryFromIndexedAttributes(filters, testDeploymentIDForConstants)

	suite.NoError(err)
	suite.Len(args, 3)
	suite.Equal("age", args[0])
	suite.Equal("30", args[1]) // Should be converted to string
}

// Test buildIdentifyQueryHybrid

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryHybrid_EmptyIndexedFilters() {
	indexedFilters := map[string]interface{}{}
	nonIndexedFilters := map[string]interface{}{
		"age": 30,
	}

	query, args, err := buildIdentifyQueryHybrid(indexedFilters, nonIndexedFilters, testDeploymentIDForConstants)

	suite.Error(err)
	suite.Contains(err.Error(), "indexed filters cannot be empty")
	suite.Nil(args)
	suite.Empty(query.Query)
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryHybrid_OnlyIndexedFilters() {
	indexedFilters := map[string]interface{}{
		"username": "john.doe",
	}
	nonIndexedFilters := map[string]interface{}{}

	query, args, err := buildIdentifyQueryHybrid(indexedFilters, nonIndexedFilters, testDeploymentIDForConstants)

	suite.NoError(err)
	suite.Equal("ASQ-USER_MGT-17", query.ID)
	suite.Contains(query.Query, "SELECT DISTINCT u.USER_ID FROM \"USER\" u")
	suite.Contains(query.Query, "INNER JOIN USER_INDEXED_ATTRIBUTES ia1")
	suite.Contains(query.Query, "WHERE ia1.ATTRIBUTE_NAME = $1 AND ia1.ATTRIBUTE_VALUE = $2")
	suite.Len(args, 3) // username (name + value) + deployment ID
	suite.Equal("username", args[0])
	suite.Equal("john.doe", args[1])
	suite.Equal(testDeploymentIDForConstants, args[2])
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryHybrid_IndexedAndNonIndexed() {
	indexedFilters := map[string]interface{}{
		"username": "john.doe",
	}
	nonIndexedFilters := map[string]interface{}{
		"age": "30",
	}

	query, args, err := buildIdentifyQueryHybrid(indexedFilters, nonIndexedFilters, testDeploymentIDForConstants)

	suite.NoError(err)
	suite.Equal("ASQ-USER_MGT-17", query.ID)

	// Check PostgreSQL query
	suite.Contains(query.PostgresQuery, "SELECT DISTINCT u.USER_ID FROM \"USER\" u")
	suite.Contains(query.PostgresQuery, "INNER JOIN USER_INDEXED_ATTRIBUTES ia1")
	suite.Contains(query.PostgresQuery, "u.ATTRIBUTES->>'age' = $3")
	suite.Contains(query.PostgresQuery, "u.DEPLOYMENT_ID = $4")

	// Check SQLite query
	suite.Contains(query.SQLiteQuery, "SELECT DISTINCT u.USER_ID FROM \"USER\" u")
	suite.Contains(query.SQLiteQuery, "json_extract(u.ATTRIBUTES, '$.age') = ?")
	suite.Contains(query.SQLiteQuery, "u.DEPLOYMENT_ID = ?")

	suite.Len(args, 4) // indexed (name + value) + non-indexed value + deployment ID
	suite.Equal("username", args[0])
	suite.Equal("john.doe", args[1])
	suite.Equal("30", args[2])
	suite.Equal(testDeploymentIDForConstants, args[3])
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryHybrid_MultipleIndexedAndNonIndexed() {
	indexedFilters := map[string]interface{}{
		"username": "john.doe",
		"email":    "john@example.com",
	}
	nonIndexedFilters := map[string]interface{}{
		"age":  "30",
		"city": "New York",
	}

	query, args, err := buildIdentifyQueryHybrid(indexedFilters, nonIndexedFilters, testDeploymentIDForConstants)

	suite.NoError(err)
	suite.Equal("ASQ-USER_MGT-17", query.ID)
	suite.Contains(query.Query, "INNER JOIN USER_INDEXED_ATTRIBUTES ia1")
	suite.Contains(query.Query, "INNER JOIN USER_INDEXED_ATTRIBUTES ia2")

	// Both PostgreSQL and SQLite should have conditions for non-indexed attributes
	suite.Contains(query.PostgresQuery, "u.ATTRIBUTES->>'age'")
	suite.Contains(query.PostgresQuery, "u.ATTRIBUTES->>'city'")
	suite.Contains(query.SQLiteQuery, "json_extract(u.ATTRIBUTES, '$.age')")
	suite.Contains(query.SQLiteQuery, "json_extract(u.ATTRIBUTES, '$.city')")

	suite.Len(args, 7) // 2 indexed * 2 + 2 non-indexed + 1 deployment ID
	suite.Equal(testDeploymentIDForConstants, args[6])
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryHybrid_NestedNonIndexedAttribute() {
	indexedFilters := map[string]interface{}{
		"username": "john.doe",
	}
	nonIndexedFilters := map[string]interface{}{
		"address.city": "New York",
	}

	query, args, err := buildIdentifyQueryHybrid(indexedFilters, nonIndexedFilters, testDeploymentIDForConstants)

	suite.NoError(err)

	// PostgreSQL should use #>> for nested paths
	suite.Contains(query.PostgresQuery, "u.ATTRIBUTES#>>'{address,city}' = $3")

	// SQLite should use json_extract with dot notation
	suite.Contains(query.SQLiteQuery, "json_extract(u.ATTRIBUTES, '$.address.city') = ?")

	suite.Len(args, 4)
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryHybrid_InvalidNonIndexedKey() {
	indexedFilters := map[string]interface{}{
		"username": "john.doe",
	}
	nonIndexedFilters := map[string]interface{}{
		"invalid-key": "value", // Contains hyphen
	}

	query, args, err := buildIdentifyQueryHybrid(indexedFilters, nonIndexedFilters, testDeploymentIDForConstants)

	suite.Error(err)
	suite.Contains(err.Error(), "invalid non-indexed filter key")
	suite.Nil(args)
	suite.Empty(query.Query)
}

func (suite *StoreConstantsTestSuite) TestBuildIdentifyQueryHybrid_DeterministicOrder() {
	indexedFilters := map[string]interface{}{
		"email":    "john@example.com",
		"username": "john.doe",
	}
	nonIndexedFilters := map[string]interface{}{
		"city": "New York",
		"age":  "30",
	}

	query1, args1, err1 := buildIdentifyQueryHybrid(indexedFilters, nonIndexedFilters, testDeploymentIDForConstants)
	query2, args2, err2 := buildIdentifyQueryHybrid(indexedFilters, nonIndexedFilters, testDeploymentIDForConstants)

	suite.NoError(err1)
	suite.NoError(err2)
	suite.Equal(query1.Query, query2.Query, "Query should be deterministic")
	suite.Equal(args1, args2, "Args should be deterministic")
}
func (suite *StoreConstantsTestSuite) TestBuildUserListQuery_MultipleFilters() {
	filters := map[string]interface{}{
		"username": "user1",
		"email":    "user1@example.com",
	}
	query, args, err := buildUserListQuery(filters, 10, 0, "test-id")
	suite.NoError(err)
	suite.Contains(query.PostgresQuery, "WHERE")
	suite.Contains(query.PostgresQuery, "AND")
	suite.NotEmpty(args)
}

func (suite *StoreConstantsTestSuite) TestBuildUserCountQuery() {
	filters := map[string]interface{}{"username": "user1"}
	query, args, err := buildUserCountQuery(filters, "test-id")
	suite.NoError(err)
	suite.Contains(query.PostgresQuery, "SELECT COUNT")
	suite.Contains(query.PostgresQuery, "WHERE")
	suite.NotEmpty(args)
}
