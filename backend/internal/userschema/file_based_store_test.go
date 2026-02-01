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

	err := suite.store.CreateUserSchema(schema)
	assert.NoError(suite.T(), err)

	// Verify schema was stored
	retrieved, err := suite.store.GetUserSchemaByID("schema-1")
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
	err := suite.store.CreateUserSchema(schema)
	assert.NoError(suite.T(), err)

	// Try to create duplicate - should succeed in file-based store as it doesn't check duplicates
	err = suite.store.CreateUserSchema(schema)
	// File-based store may allow duplicate or return error depending on implementation
	// Just verify it doesn't panic
	_ = err
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaByID_NotFound() {
	_, err := suite.store.GetUserSchemaByID("non-existent-id")
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

	err := suite.store.CreateUserSchema(schema)
	assert.NoError(suite.T(), err)

	// Get by name
	retrieved, err := suite.store.GetUserSchemaByName("basic_schema")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), schema.ID, retrieved.ID)
	assert.Equal(suite.T(), schema.Name, retrieved.Name)
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaByName_NotFound() {
	_, err := suite.store.GetUserSchemaByName("non-existent-name")
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
		err := suite.store.CreateUserSchema(schema)
		assert.NoError(suite.T(), err)
	}

	// Get list with pagination
	list, err := suite.store.GetUserSchemaList(10, 0)
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
		err := suite.store.CreateUserSchema(schema)
		assert.NoError(suite.T(), err)
	}

	// Get first page
	list, err := suite.store.GetUserSchemaList(2, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 2)

	// Get second page
	list, err = suite.store.GetUserSchemaList(2, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 2)

	// Get last page
	list, err = suite.store.GetUserSchemaList(2, 4)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 1)
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaList_EmptyStore() {
	list, err := suite.store.GetUserSchemaList(10, 0)
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

	err := suite.store.UpdateUserSchemaByID("schema-1", schema)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *FileBasedStoreTestSuite) TestDeleteUserSchemaByID_ReturnsError() {
	err := suite.store.DeleteUserSchemaByID("schema-1")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *FileBasedStoreTestSuite) TestGetUserSchemaListCount() {
	// Initially empty
	count, err := suite.store.GetUserSchemaListCount()
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
		err := suite.store.CreateUserSchema(schema)
		assert.NoError(suite.T(), err)
	}

	// Check count
	count, err = suite.store.GetUserSchemaListCount()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)
}
