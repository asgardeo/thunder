/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

	"github.com/stretchr/testify/suite"
)

type StoreTestSuite struct {
	suite.Suite
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (suite *StoreTestSuite) TestParseUserSchemaFromRowWithDisplayAttribute() {
	row := map[string]interface{}{
		"schema_id":               "test-id",
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
		"display_attribute":       "email",
		"schema_def":              `{"email":{"type":"string"}}`,
	}

	schema, err := parseUserSchemaFromRow(row)

	suite.NoError(err)
	suite.Equal("test-id", schema.ID)
	suite.Equal("employee", schema.Name)
	suite.Equal(testOUID1, schema.OrganizationUnitID)
	suite.True(schema.AllowSelfRegistration)
	suite.Equal("email", schema.DisplayAttribute)
	suite.Equal(json.RawMessage(`{"email":{"type":"string"}}`), schema.Schema)
}

func (suite *StoreTestSuite) TestParseUserSchemaFromRowWithNilDisplayAttribute() {
	row := map[string]interface{}{
		"schema_id":               "test-id",
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
		"display_attribute":       nil,
		"schema_def":              `{"email":{"type":"string"}}`,
	}

	schema, err := parseUserSchemaFromRow(row)

	suite.NoError(err)
	suite.Equal("", schema.DisplayAttribute)
}

func (suite *StoreTestSuite) TestParseUserSchemaFromRowWithMissingDisplayAttribute() {
	row := map[string]interface{}{
		"schema_id":               "test-id",
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
		"schema_def":              `{"email":{"type":"string"}}`,
	}

	schema, err := parseUserSchemaFromRow(row)

	suite.NoError(err)
	suite.Equal("", schema.DisplayAttribute)
}

func (suite *StoreTestSuite) TestParseUserSchemaFromRowWithByteSliceSchemaDef() {
	row := map[string]interface{}{
		"schema_id":               "test-id",
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
		"display_attribute":       "email",
		"schema_def":              []byte(`{"email":{"type":"string"}}`),
	}

	schema, err := parseUserSchemaFromRow(row)

	suite.NoError(err)
	suite.Equal(json.RawMessage(`{"email":{"type":"string"}}`), schema.Schema)
}

func (suite *StoreTestSuite) TestParseUserSchemaFromRowReturnsErrorForMissingSchemaID() {
	row := map[string]interface{}{
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
		"schema_def":              `{"email":{"type":"string"}}`,
	}

	_, err := parseUserSchemaFromRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "schema_id")
}

func (suite *StoreTestSuite) TestParseUserSchemaFromRowReturnsErrorForMissingName() {
	row := map[string]interface{}{
		"schema_id":               "test-id",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
		"schema_def":              `{"email":{"type":"string"}}`,
	}

	_, err := parseUserSchemaFromRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "name")
}

func (suite *StoreTestSuite) TestParseUserSchemaFromRowReturnsErrorForInvalidSchemaDef() {
	row := map[string]interface{}{
		"schema_id":               "test-id",
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
		"schema_def":              12345,
	}

	_, err := parseUserSchemaFromRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "schema_def")
}

func (suite *StoreTestSuite) TestParseUserSchemaListItemFromRowWithDisplayAttribute() {
	row := map[string]interface{}{
		"schema_id":               "test-id",
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
		"display_attribute":       "email",
	}

	item, err := parseUserSchemaListItemFromRow(row)

	suite.NoError(err)
	suite.Equal("test-id", item.ID)
	suite.Equal("employee", item.Name)
	suite.Equal(testOUID1, item.OrganizationUnitID)
	suite.True(item.AllowSelfRegistration)
	suite.Equal("email", item.DisplayAttribute)
}

func (suite *StoreTestSuite) TestParseUserSchemaListItemFromRowWithNilDisplayAttribute() {
	row := map[string]interface{}{
		"schema_id":               "test-id",
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
		"display_attribute":       nil,
	}

	item, err := parseUserSchemaListItemFromRow(row)

	suite.NoError(err)
	suite.Equal("", item.DisplayAttribute)
}

func (suite *StoreTestSuite) TestParseUserSchemaListItemFromRowWithMissingDisplayAttribute() {
	row := map[string]interface{}{
		"schema_id":               "test-id",
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
	}

	item, err := parseUserSchemaListItemFromRow(row)

	suite.NoError(err)
	suite.Equal("", item.DisplayAttribute)
}

func (suite *StoreTestSuite) TestParseUserSchemaListItemFromRowReturnsErrorForMissingSchemaID() {
	row := map[string]interface{}{
		"name":                    "employee",
		"ou_id":                   testOUID1,
		"allow_self_registration": true,
	}

	_, err := parseUserSchemaListItemFromRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "schema_id")
}

func (suite *StoreTestSuite) TestParseBoolHandlesVariousTypes() {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"int64 1", int64(1), true},
		{"int64 0", int64(0), false},
		{"float64 1", float64(1), true},
		{"float64 0", float64(0), false},
		{"string true", "true", true},
		{"string false", "false", false},
		{"string 1", "1", true},
		{"byte slice true", []byte("true"), true},
		{"byte slice 1", []byte("1"), true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := parseBool(tt.value, "test_field")
			suite.NoError(err)
			suite.Equal(tt.expected, result)
		})
	}
}

func (suite *StoreTestSuite) TestParseBoolReturnsErrorForNil() {
	_, err := parseBool(nil, "test_field")
	suite.Error(err)
	suite.Contains(err.Error(), "test_field")
}

func (suite *StoreTestSuite) TestParseBoolReturnsErrorForUnsupportedType() {
	_, err := parseBool([]int{1, 2, 3}, "test_field")
	suite.Error(err)
	suite.Contains(err.Error(), "test_field")
}
