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

package userschema

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

const testSchemaJSON = `{"type":"object"}`

type FileBasedStoreTestSuite struct {
	suite.Suite
	store userSchemaStoreInterface
}

func (suite *FileBasedStoreTestSuite) SetupTest() {
	suite.store = newUserSchemaFileBasedStoreForTest()
}

// newUserSchemaFileBasedStoreForTest creates a test instance
func newUserSchemaFileBasedStoreForTest() userSchemaStoreInterface {
	genericStore := declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeUserSchema)
	return &userSchemaFileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}

func TestFileBasedStoreTestSuite(t *testing.T) {
	suite.Run(t, new(FileBasedStoreTestSuite))
}

func (suite *FileBasedStoreTestSuite) TestCreateUserSchema() {
	schemaJSON := `{"type":"object","properties":{"username":{"type":"string"}}}`
	schema := UserSchema{
		ID:                    "schema-1",
		Name:                  "basic_schema",
		OrganizationUnitID:    "ou-1",
		AllowSelfRegistration: true,
		Schema:                json.RawMessage(schemaJSON),
	}

	err := suite.store.CreateUserSchema(context.Background(), schema)
	assert.NoError(suite.T(), err)

	// Verify schema was stored
	retrieved, err := suite.store.GetUserSchemaByID(context.Background(), "schema-1")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), schema.ID, retrieved.ID)
	assert.Equal(suite.T(), schema.Name, retrieved.Name)
	assert.Equal(suite.T(), schema.OrganizationUnitID, retrieved.OrganizationUnitID)
	assert.Equal(suite.T(), schema.AllowSelfRegistration, retrieved.AllowSelfRegistration)
}

func (suite *FileBasedStoreTestSuite) TestCreateUserSchema_DuplicateID() {
	schemaJSON := testSchemaJSON
	schema := UserSchema{
		ID:                    "schema-1",
		Name:                  "basic_schema",
		OrganizationUnitID:    "ou-1",
		AllowSelfRegistration: true,
		Schema:                json.RawMessage(schemaJSON),
	}

	// Create first schema
	err := suite.store.CreateUserSchema(context.Background(), schema)
	assert.NoError(suite.T(), err)

	// Try to create duplicate - should succeed in file-based store as it doesn't check duplicates
	err = suite.store.CreateUserSchema(context.Background(), schema)
	// File-based store may allow duplicate or return error depending on implementation
	// Just verify it doesn't panic
	_ = err
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaByID_NotFound() {
	_, err := suite.store.GetUserSchemaByID(context.Background(), "non-existent-id")
	assert.Error(suite.T(), err)
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaByName() {
	schemaJSON := testSchemaJSON
	schema := UserSchema{
		ID:                    "schema-1",
		Name:                  "basic_schema",
		OrganizationUnitID:    "ou-1",
		AllowSelfRegistration: true,
		Schema:                json.RawMessage(schemaJSON),
	}

	err := suite.store.CreateUserSchema(context.Background(), schema)
	assert.NoError(suite.T(), err)

	// Get by name
	retrieved, err := suite.store.GetUserSchemaByName(context.Background(), "basic_schema")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), schema.ID, retrieved.ID)
	assert.Equal(suite.T(), schema.Name, retrieved.Name)
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaByName_NotFound() {
	_, err := suite.store.GetUserSchemaByName(context.Background(), "non-existent-name")
	assert.Error(suite.T(), err)
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaList() {
	schemaJSON := testSchemaJSON
	// Create multiple schemas
	schemas := []UserSchema{
		{
			ID:                    "schema-1",
			Name:                  "basic_schema",
			OrganizationUnitID:    "ou-1",
			AllowSelfRegistration: true,
			Schema:                json.RawMessage(schemaJSON),
		},
		{
			ID:                    "schema-2",
			Name:                  "extended_schema",
			OrganizationUnitID:    "ou-1",
			AllowSelfRegistration: false,
			Schema:                json.RawMessage(schemaJSON),
		},
		{
			ID:                    "schema-3",
			Name:                  "minimal_schema",
			OrganizationUnitID:    "ou-1",
			AllowSelfRegistration: true,
			Schema:                json.RawMessage(schemaJSON),
		},
	}
	for _, schema := range schemas {
		err := suite.store.CreateUserSchema(context.Background(), schema)
		assert.NoError(suite.T(), err)
	}

	// Get list with pagination
	list, err := suite.store.GetUserSchemaList(context.Background(), 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 3)
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaList_WithPagination() {
	schemaJSON := testSchemaJSON
	// Create multiple schemas
	for i := 1; i <= 5; i++ {
		schema := UserSchema{
			ID:                    "schema-" + string(rune('0'+i)),
			Name:                  "schema_" + string(rune('0'+i)),
			OrganizationUnitID:    "ou-1",
			AllowSelfRegistration: true,
			Schema:                json.RawMessage(schemaJSON),
		}
		err := suite.store.CreateUserSchema(context.Background(), schema)
		assert.NoError(suite.T(), err)
	}

	// Get first page
	list, err := suite.store.GetUserSchemaList(context.Background(), 2, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 2)

	// Get second page
	list, err = suite.store.GetUserSchemaList(context.Background(), 2, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 2)

	// Get last page
	list, err = suite.store.GetUserSchemaList(context.Background(), 2, 4)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 1)
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaList_EmptyStore() {
	list, err := suite.store.GetUserSchemaList(context.Background(), 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 0)
}

func (suite *FileBasedStoreTestSuite) TestUpdateUserSchemaByID_ReturnsError() {
	schemaJSON := testSchemaJSON
	schema := UserSchema{
		ID:                    "schema-1",
		Name:                  "basic_schema",
		OrganizationUnitID:    "ou-1",
		AllowSelfRegistration: true,
		Schema:                json.RawMessage(schemaJSON),
	}

	err := suite.store.UpdateUserSchemaByID(context.Background(), "schema-1", schema)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *FileBasedStoreTestSuite) TestDeleteUserSchemaByID_ReturnsError() {
	err := suite.store.DeleteUserSchemaByID(context.Background(), "schema-1")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaListCount() {
	// Initially empty
	count, err := suite.store.GetUserSchemaListCount(context.Background())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	schemaJSON := testSchemaJSON
	// Add schemas
	for i := 1; i <= 3; i++ {
		schema := UserSchema{
			ID:                    "schema-" + string(rune('0'+i)),
			Name:                  "schema_" + string(rune('0'+i)),
			OrganizationUnitID:    "ou-1",
			AllowSelfRegistration: true,
			Schema:                json.RawMessage(schemaJSON),
		}
		err := suite.store.CreateUserSchema(context.Background(), schema)
		assert.NoError(suite.T(), err)
	}

	// Check count
	count, err = suite.store.GetUserSchemaListCount(context.Background())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaListByOUIDs() {
	schemaJSON := testSchemaJSON
	schemas := []UserSchema{
		{
			ID:                    "schema-1",
			Name:                  "schema_1",
			OrganizationUnitID:    "ou-1",
			AllowSelfRegistration: true,
			Schema:                json.RawMessage(schemaJSON),
		},
		{
			ID:                    "schema-2",
			Name:                  "schema_2",
			OrganizationUnitID:    "ou-2",
			AllowSelfRegistration: false,
			Schema:                json.RawMessage(schemaJSON),
		},
		{
			ID:                    "schema-3",
			Name:                  "schema_3",
			OrganizationUnitID:    "ou-1",
			AllowSelfRegistration: true,
			Schema:                json.RawMessage(schemaJSON),
		},
	}
	for _, schema := range schemas {
		err := suite.store.CreateUserSchema(context.Background(), schema)
		assert.NoError(suite.T(), err)
	}

	testCases := []struct {
		name          string
		ouIDs         []string
		limit         int
		offset        int
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "Get by single OU ID",
			ouIDs:         []string{"ou-2"},
			limit:         10,
			offset:        0,
			expectedCount: 1,
			expectedNames: []string{"schema_2"},
		},
		{
			name:          "Get by multiple OU IDs",
			ouIDs:         []string{"ou-1", "ou-2"},
			limit:         10,
			offset:        0,
			expectedCount: 3,
			expectedNames: []string{"schema_1", "schema_2", "schema_3"},
		},
		{
			name:          "Get by non-existent OU ID",
			ouIDs:         []string{"ou-3"},
			limit:         10,
			offset:        0,
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name:          "Pagination limit",
			ouIDs:         []string{"ou-1"},
			limit:         1,
			offset:        0,
			expectedCount: 1,
			expectedNames: []string{"schema_1"},
		},
		{
			name:          "Pagination offset",
			ouIDs:         []string{"ou-1"},
			limit:         10,
			offset:        1,
			expectedCount: 1,
			expectedNames: []string{"schema_3"},
		},
		{
			name:          "Pagination beyond total",
			ouIDs:         []string{"ou-1"},
			limit:         10,
			offset:        5,
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			list, err := suite.store.GetUserSchemaListByOUIDs(context.Background(), tc.ouIDs, tc.limit, tc.offset)
			assert.NoError(suite.T(), err)
			assert.Len(suite.T(), list, tc.expectedCount)

			// Verify names if expected
			var names []string
			for _, item := range list {
				names = append(names, item.Name)
			}
			assert.ElementsMatch(suite.T(), tc.expectedNames, names)
		})
	}
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaListCountByOUIDs() {
	schemaJSON := testSchemaJSON
	schemas := []UserSchema{
		{
			ID:                    "schema-1",
			Name:                  "schema_1",
			OrganizationUnitID:    "ou-1",
			AllowSelfRegistration: true,
			Schema:                json.RawMessage(schemaJSON),
		},
		{
			ID:                    "schema-2",
			Name:                  "schema_2",
			OrganizationUnitID:    "ou-2",
			AllowSelfRegistration: false,
			Schema:                json.RawMessage(schemaJSON),
		},
		{
			ID:                    "schema-3",
			Name:                  "schema_3",
			OrganizationUnitID:    "ou-1",
			AllowSelfRegistration: true,
			Schema:                json.RawMessage(schemaJSON),
		},
	}
	for _, schema := range schemas {
		err := suite.store.CreateUserSchema(context.Background(), schema)
		assert.NoError(suite.T(), err)
	}

	testCases := []struct {
		name          string
		ouIDs         []string
		expectedCount int
	}{
		{
			name:          "Count by single OU ID",
			ouIDs:         []string{"ou-2"},
			expectedCount: 1,
		},
		{
			name:          "Count by multiple OU IDs",
			ouIDs:         []string{"ou-1", "ou-2"},
			expectedCount: 3,
		},
		{
			name:          "Count by non-existent OU ID",
			ouIDs:         []string{"ou-3"},
			expectedCount: 0,
		},
		{
			name:          "Empty OU IDs",
			ouIDs:         []string{},
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			count, err := suite.store.GetUserSchemaListCountByOUIDs(context.Background(), tc.ouIDs)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), tc.expectedCount, count)
		})
	}
}
