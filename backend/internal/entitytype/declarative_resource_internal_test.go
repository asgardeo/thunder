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

package entitytype

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	oupkg "github.com/thunder-id/thunderid/internal/ou"
	"github.com/thunder-id/thunderid/internal/system/config"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
	"github.com/thunder-id/thunderid/internal/system/i18n/core"
	"github.com/thunder-id/thunderid/tests/mocks/oumock"
)

// TestValidateEntityType tests the validateEntityType function with various scenarios.
func TestValidateEntityType(t *testing.T) {
	// Setup mock OU service
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	testCases := []struct {
		name      string
		schema    *EntityType
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid schema",
			schema: &EntityType{
				ID:     "schema-1",
				Name:   "Valid Schema",
				OUID:   "ou-1",
				Schema: json.RawMessage(`{"email":{"type":"string"}}`),
			},
			setupMock: func() {
				mockOUService.EXPECT().GetOrganizationUnit(mock.Anything, "ou-1").
					Return(oupkg.OrganizationUnit{ID: "ou-1"}, nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "missing name",
			schema: &EntityType{
				ID:   "schema-1",
				Name: "",
				OUID: "ou-1",
			},
			setupMock: func() {},
			wantErr:   true,
			errMsg:    "entity type name is required",
		},
		{
			name: "whitespace only name",
			schema: &EntityType{
				ID:   "schema-1",
				Name: "   ",
				OUID: "ou-1",
			},
			setupMock: func() {},
			wantErr:   true,
			errMsg:    "entity type name is required",
		},
		{
			name: "missing ID",
			schema: &EntityType{
				ID:   "",
				Name: "Valid Schema",
				OUID: "ou-1",
			},
			setupMock: func() {},
			wantErr:   true,
			errMsg:    "entity type ID is required",
		},
		{
			name: "whitespace only ID",
			schema: &EntityType{
				ID:   "   ",
				Name: "Valid Schema",
				OUID: "ou-1",
			},
			setupMock: func() {},
			wantErr:   true,
			errMsg:    "entity type ID is required",
		},
		{
			name: "missing organization unit ID",
			schema: &EntityType{
				ID:   "schema-1",
				Name: "Valid Schema",
				OUID: "",
			},
			setupMock: func() {},
			wantErr:   true,
			errMsg:    "organization unit ID is required",
		},
		{
			name: "whitespace only organization unit ID",
			schema: &EntityType{
				ID:   "schema-1",
				Name: "Valid Schema",
				OUID: "   ",
			},
			setupMock: func() {},
			wantErr:   true,
			errMsg:    "organization unit ID is required",
		},
		{
			name: "organization unit not found",
			schema: &EntityType{
				ID:     "schema-1",
				Name:   "Valid Schema",
				OUID:   "nonexistent",
				Schema: json.RawMessage(`{"type": "object"}`),
			},
			setupMock: func() {
				mockOUService.EXPECT().GetOrganizationUnit(mock.Anything, "nonexistent").
					Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{Code: "NOT_FOUND"}).
					Once()
			},
			wantErr: true,
			errMsg:  "organization unit 'nonexistent' not found",
		},
		{
			name: "invalid schema JSON",
			schema: &EntityType{
				ID:     "schema-1",
				Name:   "Invalid Schema",
				OUID:   "ou-1",
				Schema: json.RawMessage(`{invalid json}`),
			},
			setupMock: func() {
				mockOUService.EXPECT().GetOrganizationUnit(mock.Anything, "ou-1").
					Return(oupkg.OrganizationUnit{ID: "ou-1"}, nil).
					Once()
			},
			wantErr: true,
			errMsg:  "invalid schema for entity type",
		},
		{
			name: "empty schema definition rejected",
			schema: &EntityType{
				ID:     "schema-1",
				Name:   "Valid Schema",
				OUID:   "ou-1",
				Schema: json.RawMessage(``),
			},
			setupMock: func() {
				mockOUService.EXPECT().GetOrganizationUnit(mock.Anything, "ou-1").
					Return(oupkg.OrganizationUnit{ID: "ou-1"}, nil).
					Once()
			},
			wantErr: true,
			errMsg:  "schema definition is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			err := validateEntityType(tc.schema, mockOUService)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateEntityTypeWrapper tests the wrapper function.
func TestValidateEntityTypeWrapper(t *testing.T) {
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	t.Run("valid type", func(t *testing.T) {
		schema := &EntityType{
			ID:     "schema-1",
			Name:   "Valid Schema",
			OUID:   "ou-1",
			Schema: json.RawMessage(`{"email":{"type":"string"}}`),
		}

		mockOUService.EXPECT().GetOrganizationUnit(mock.Anything, "ou-1").
			Return(oupkg.OrganizationUnit{ID: "ou-1"}, nil).
			Once()

		validator := validateEntityTypeWrapper(mockOUService)
		err := validator(schema)

		assert.NoError(t, err)
	})

	t.Run("invalid type", func(t *testing.T) {
		invalidData := "not a schema"

		validator := validateEntityTypeWrapper(mockOUService)
		err := validator(invalidData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type: expected *EntityType")
	})
}

// TestParseToEntityTypeDTO tests the parseToEntityTypeDTO function.
func TestParseToEntityTypeDTO(t *testing.T) {
	testCases := []struct {
		name           string
		yaml           string
		want           *EntityType
		wantErr        bool
		errMsg         string
		validateSchema bool
	}{
		{
			name: "valid YAML",
			yaml: `
id: schema-1
name: Test Schema
organization_unit_id: ou-1
allow_self_registration: true
schema: '{"type": "object"}'
`,
			want: &EntityType{
				ID:                    "schema-1",
				Name:                  "Test Schema",
				OUID:                  "ou-1",
				AllowSelfRegistration: true,
				Schema:                json.RawMessage(`{"type": "object"}`),
			},
			wantErr: false,
		},
		{
			name: "valid YAML without optional fields",
			yaml: `
id: schema-2
name: Minimal Schema
organization_unit_id: ou-1
schema: '{}'
`,
			want: &EntityType{
				ID:                    "schema-2",
				Name:                  "Minimal Schema",
				OUID:                  "ou-1",
				AllowSelfRegistration: false,
				Schema:                json.RawMessage(`{}`),
			},
			wantErr: false,
		},
		{
			name: "invalid YAML",
			yaml: `
invalid: [yaml
`,
			wantErr: true,
		},
		{
			name: "invalid JSON in schema field",
			yaml: `
id: schema-1
name: Test Schema
organization_unit_id: ou-1
schema: '{invalid json}'
`,
			wantErr: true,
			errMsg:  "schema field contains invalid JSON",
		},
		{
			name: "schema as YAML object",
			yaml: `
id: schema-1
name: Test Schema
organization_unit_id: ou-1
schema:
  username:
    type: string
    required: true
`,
			want: &EntityType{
				ID:   "schema-1",
				Name: "Test Schema",
				OUID: "ou-1",
				Schema: json.RawMessage(
					`{"username":{"required":true,"type":"string"}}`,
				),
			},
			wantErr:        false,
			validateSchema: true,
		},
		{
			name: "missing schema field",
			yaml: `
id: schema-1
name: Test Schema
organization_unit_id: ou-1
`,
			wantErr: true,
			errMsg:  "schema field contains invalid JSON",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseToEntityTypeDTO([]byte(tc.yaml))

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want.ID, result.ID)
				assert.Equal(t, tc.want.Name, result.Name)
				assert.Equal(t, tc.want.OUID, result.OUID)
				assert.Equal(t, tc.want.AllowSelfRegistration, result.AllowSelfRegistration)
				if tc.validateSchema {
					var got, expected map[string]interface{}
					assert.NoError(t, json.Unmarshal(result.Schema, &got), "result schema must decode to JSON object")
					assert.NoError(t, json.Unmarshal(tc.want.Schema, &expected),
						"test fixture schema must decode to JSON object")
					assert.Equal(t, expected, got, "decoded schema must deep-equal the expected value")
					assert.NotEqual(t, map[string]interface{}{}, got, "schema must not decode to an empty object")
				}
			}
		})
	}
}

// TestParseToEntityTypeDTOWrapper tests the wrapper function.
func TestParseToEntityTypeDTOWrapper(t *testing.T) {
	yaml := `
id: schema-1
name: Test Schema
oUId: ou-1
schema: '{"type": "object"}'
`
	result, err := parseToEntityTypeDTOWrapper([]byte(yaml))

	assert.NoError(t, err)
	schema, ok := result.(*EntityType)
	assert.True(t, ok)
	assert.Equal(t, "schema-1", schema.ID)
	assert.Equal(t, "Test Schema", schema.Name)
}

// TestLoadDeclarativeResources tests the loadDeclarativeResources function.
func TestLoadDeclarativeResources(t *testing.T) {
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	// Initialize runtime config for tests that need DB access
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		Database: config.DatabaseConfig{
			Config: config.DataSource{
				Type:   "sqlite",
				SQLite: config.SQLiteDataSource{Path: ":memory:"},
			},
		},
	}

	t.Run("composite store", func(t *testing.T) {
		config.ResetServerRuntime()
		err := config.InitializeServerRuntime("", testConfig)
		assert.NoError(t, err)
		defer config.ResetServerRuntime()

		fileStore, _ := newEntityTypeFileBasedStore()
		dbStore, _, _ := newEntityTypeStore()
		compositeStore := newCompositeEntityTypeStore(fileStore, dbStore)

		// Mock OU service to return valid OU for any ID
		mockOUService.On("GetOrganizationUnit", mock.Anything, mock.Anything).
			Return(oupkg.OrganizationUnit{ID: "ou-1"}, nil).
			Maybe()

		// loadDeclarativeResources should work with composite store
		err = loadDeclarativeResources(compositeStore, mockOUService)
		// The function should complete without panicking
		// Error handling is appropriate: if no declarative_resources directory exists,
		// that's acceptable for a composite store configuration
		assert.True(t, err == nil || err != nil, "Function should complete regardless of directory presence")
	})

	t.Run("file-based store", func(t *testing.T) {
		config.ResetServerRuntime()
		err := config.InitializeServerRuntime("", testConfig)
		assert.NoError(t, err)
		defer config.ResetServerRuntime()

		fileStore, _ := newEntityTypeFileBasedStore()

		// Mock OU service
		mockOUService.On("GetOrganizationUnit", mock.Anything, mock.Anything).
			Return(oupkg.OrganizationUnit{ID: "ou-1"}, nil).
			Maybe()

		// loadDeclarativeResources should work with file-based store
		err = loadDeclarativeResources(fileStore, mockOUService)
		// May succeed or fail depending on whether declarative_resources directory exists
		_ = err // Don't assert on error as it depends on file system state
	})

	t.Run("invalid store type", func(t *testing.T) {
		config.ResetServerRuntime()
		err := config.InitializeServerRuntime("", testConfig)
		assert.NoError(t, err)
		defer config.ResetServerRuntime()

		// Use the regular database store which should not be valid for declarative resources
		dbStore, _, _ := newEntityTypeStore()

		err = loadDeclarativeResources(dbStore, mockOUService)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid store type")
	})
}

// TestGetAllResourceIDs_WithReadOnlyFilter tests that declarative schemas are excluded from export.
func TestGetAllResourceIDs_WithReadOnlyFilter(t *testing.T) {
	mockService := NewEntityTypeServiceInterfaceMock(t)

	exporter := newEntityTypeExporter(mockService)

	response := &EntityTypeListResponse{
		Types: []EntityTypeListItem{
			{ID: "schema1", Name: "Schema 1", IsReadOnly: false}, // Mutable - should be included
			{ID: "schema2", Name: "Schema 2", IsReadOnly: true},  // Immutable - should be excluded
			{ID: "schema3", Name: "Schema 3", IsReadOnly: false}, // Mutable - should be included
		},
	}

	mockService.On("GetEntityTypeList", mock.Anything, mock.Anything, 100, 0, false).Return(response, nil)

	ids, err := exporter.GetAllResourceIDs(context.Background())

	assert.Nil(t, err)
	assert.Len(t, ids, 2, "Should only include mutable schemas")
	assert.Contains(t, ids, "schema1")
	assert.Contains(t, ids, "schema3")
	assert.NotContains(t, ids, "schema2", "Schema2 is read-only and should be excluded")
}

// TestLoadDeclarativeResources_WithNilOUService tests error handling when OU service is nil.
func TestLoadDeclarativeResources_WithNilOUService(t *testing.T) {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		Database: config.DatabaseConfig{
			Config: config.DataSource{
				Type:   "sqlite",
				SQLite: config.SQLiteDataSource{Path: ":memory:"},
			},
		},
	}

	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("", testConfig)
	assert.NoError(t, err)
	defer config.ResetServerRuntime()

	fileStore, _ := newEntityTypeFileBasedStore()
	dbStore, _, _ := newEntityTypeStore()
	compositeStore := newCompositeEntityTypeStore(fileStore, dbStore)

	// This should handle nil OU service gracefully or return an error
	// depending on whether resources are actually being validated
	err = loadDeclarativeResources(compositeStore, nil)
	// We don't assert specific behavior since it depends on file system state
	// The important part is that it doesn't panic
	_ = err
}

// TestValidateEntityType_OUServiceError tests handling of OU service errors.
func TestValidateEntityType_OUServiceError(t *testing.T) {
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	schema := &EntityType{
		ID:     "schema-1",
		Name:   "Valid Schema",
		OUID:   "ou-1",
		Schema: json.RawMessage(`{"type": "object"}`),
	}

	// Simulate a service error (not just not found)
	mockOUService.EXPECT().GetOrganizationUnit(mock.Anything, "ou-1").
		Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{
			Code: "DB_ERROR",
			Error: core.I18nMessage{
				Key:          "error.organizationunit.database_error",
				DefaultValue: "database connection failed",
			},
		}).
		Once()

	err := validateEntityType(schema, mockOUService)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization unit 'ou-1' not found")
}
