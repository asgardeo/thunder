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

package usertype

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/oumock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testCryptoKey = "0579f866ac7c9273580d0ff163fa01a7b2401a7ff3ddc3e3b14ae3136fa6025e"
)

// InitTestSuite contains comprehensive tests for the init.go file.
type InitTestSuite struct {
	suite.Suite
	mockOUService *oumock.OrganizationUnitServiceInterfaceMock
	mux           *http.ServeMux
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (suite *InitTestSuite) SetupTest() {
	suite.mockOUService = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())
	suite.mux = http.NewServeMux()
}

func (suite *InitTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

// TestInitialize tests the Initialize function
func (suite *InitTestSuite) TestInitialize() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	service, _, err := Initialize(suite.mux, suite.mockOUService)
	assert.NoError(suite.T(), err)

	suite.NotNil(service)
	suite.Implements((*UserTypeServiceInterface)(nil), service)
}

// TestRegisterRoutes_ListEndpoint tests that the list endpoint is registered
func (suite *InitTestSuite) TestRegisterRoutes_ListEndpoint() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	_, _, err = Initialize(suite.mux, suite.mockOUService)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, "/user-types", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

// TestRegisterRoutes_CreateEndpoint tests that the create endpoint is registered
func (suite *InitTestSuite) TestRegisterRoutes_CreateEndpoint() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	_, _, err = Initialize(suite.mux, suite.mockOUService)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/user-types", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

// TestRegisterRoutes_GetByIDEndpoint tests that the get by ID endpoint is registered
func (suite *InitTestSuite) TestRegisterRoutes_GetByIDEndpoint() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	_, _, err = Initialize(suite.mux, suite.mockOUService)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, "/user-types/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

// TestRegisterRoutes_UpdateEndpoint tests that the update endpoint is registered
func (suite *InitTestSuite) TestRegisterRoutes_UpdateEndpoint() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	_, _, err = Initialize(suite.mux, suite.mockOUService)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPut, "/user-types/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

// TestRegisterRoutes_DeleteEndpoint tests that the delete endpoint is registered
func (suite *InitTestSuite) TestRegisterRoutes_DeleteEndpoint() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	_, _, err = Initialize(suite.mux, suite.mockOUService)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodDelete, "/user-types/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

// TestRegisterRoutes_CORSPreflight tests that CORS preflight requests are handled
func (suite *InitTestSuite) TestRegisterRoutes_CORSPreflight() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	_, _, err = Initialize(suite.mux, suite.mockOUService)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodOptions, "/user-types", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.Equal(http.StatusNoContent, w.Code)
}

// TestRegisterRoutes_CORSPreflightByID tests that CORS preflight requests for ID endpoint are handled
func (suite *InitTestSuite) TestRegisterRoutes_CORSPreflightByID() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	_, _, err = Initialize(suite.mux, suite.mockOUService)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodOptions, "/user-types/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.Equal(http.StatusNoContent, w.Code)
}

// TestParseToUserTypeDTO_ValidYAML tests parsing a valid YAML configuration
func (suite *InitTestSuite) TestParseToUserTypeDTO_ValidYAML() {
	yamlData := `
id: "schema-001"
name: "Employee Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
allow_self_registration: true
schema: |
  {
    "type": "object",
    "properties": {
      "email": {"type": "string"},
      "username": {"type": "string"}
    },
    "required": ["email", "username"]
  }
`

	schemaDTO, err := parseToUserTypeDTO([]byte(yamlData))

	suite.NoError(err)
	suite.NotNil(schemaDTO)
	suite.Equal("schema-001", schemaDTO.ID)
	suite.Equal("Employee Schema", schemaDTO.Name)
	suite.Equal("550e8400-e29b-41d4-a716-446655440000", schemaDTO.OrganizationUnitID)
	suite.True(schemaDTO.AllowSelfRegistration)
	suite.NotEmpty(schemaDTO.Schema)
}

// TestParseToUserTypeDTO_MinimalYAML tests parsing minimal YAML configuration
func (suite *InitTestSuite) TestParseToUserTypeDTO_MinimalYAML() {
	yamlData := `
id: "minimal-schema"
name: "Minimal Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
schema: |
  {
    "type": "object",
    "properties": {
      "email": {"type": "string"}
    }
  }
`

	schemaDTO, err := parseToUserTypeDTO([]byte(yamlData))

	suite.NoError(err)
	suite.NotNil(schemaDTO)
	suite.Equal("minimal-schema", schemaDTO.ID)
	suite.Equal("Minimal Schema", schemaDTO.Name)
	suite.Equal("550e8400-e29b-41d4-a716-446655440000", schemaDTO.OrganizationUnitID)
	suite.False(schemaDTO.AllowSelfRegistration)
	suite.NotEmpty(schemaDTO.Schema)
}

// TestParseToUserTypeDTO_InvalidYAML tests parsing invalid YAML
func (suite *InitTestSuite) TestParseToUserTypeDTO_InvalidYAML() {
	yamlData := `
invalid yaml content
  - this is not valid
`

	schemaDTO, err := parseToUserTypeDTO([]byte(yamlData))

	suite.Error(err)
	suite.Nil(schemaDTO)
}

// TestParseToUserTypeDTO_ComplexSchema tests parsing with complex schema
func (suite *InitTestSuite) TestParseToUserTypeDTO_ComplexSchema() {
	yamlData := `
id: "complex-schema"
name: "Complex Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
allow_self_registration: true
schema: |
  {
    "type": "object",
    "properties": {
      "email": {
        "type": "string",
        "format": "email"
      },
      "username": {
        "type": "string",
        "minLength": 3,
        "maxLength": 20
      },
      "age": {
        "type": "number",
        "minimum": 18
      },
      "address": {
        "type": "object",
        "properties": {
          "street": {"type": "string"},
          "city": {"type": "string"}
        }
      }
    },
    "required": ["email", "username"]
  }
`

	schemaDTO, err := parseToUserTypeDTO([]byte(yamlData))

	suite.NoError(err)
	suite.NotNil(schemaDTO)
	suite.Equal("complex-schema", schemaDTO.ID)
	suite.Equal("Complex Schema", schemaDTO.Name)
	suite.True(schemaDTO.AllowSelfRegistration)
	suite.NotEmpty(schemaDTO.Schema)
}

// BenchmarkParseToUserTypeDTO benchmarks the YAML parsing performance
func BenchmarkParseToUserTypeDTO(b *testing.B) {
	yamlData := `
id: "benchmark-schema"
name: "Benchmark Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
schema: |
  {
    "type": "object",
    "properties": {
      "email": {"type": "string"}
    }
  }
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseToUserTypeDTO([]byte(yamlData))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestParseToUserTypeDTO_Standalone tests YAML parsing without suite dependencies
func TestParseToUserTypeDTO_Standalone(t *testing.T) {
	yamlData := `
id: "standalone-schema"
name: "Standalone Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
allow_self_registration: false
schema: |
  {
    "type": "object",
    "properties": {
      "email": {"type": "string"}
    }
  }
`

	schemaDTO, err := parseToUserTypeDTO([]byte(yamlData))

	assert.NoError(t, err)
	assert.NotNil(t, schemaDTO)
	assert.Equal(t, "standalone-schema", schemaDTO.ID)
	assert.Equal(t, "Standalone Schema", schemaDTO.Name)
	assert.False(t, schemaDTO.AllowSelfRegistration)
	assert.NotEmpty(t, schemaDTO.Schema)
}

// TestRegisterRoutes_Standalone tests route registration without suite dependencies
func TestRegisterRoutes_Standalone(t *testing.T) {
	mux := http.NewServeMux()
	mockHandler := &userTypeHandler{}

	assert.NotPanics(t, func() {
		registerRoutes(mux, mockHandler)
	})
}

// TestInitialize_Standalone tests Initialize function without suite dependencies
func TestInitialize_Standalone(t *testing.T) {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}

	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(t, err)

	defer config.ResetThunderRuntime()

	mux := http.NewServeMux()
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	service, _, err := Initialize(mux, mockOUService)
	assert.NoError(t, err)

	assert.NotNil(t, service)
	assert.Implements(t, (*UserTypeServiceInterface)(nil), service)
}

// TestParseToUserTypeDTO_InvalidJSONSchema tests parsing with invalid JSON in schema field
func TestParseToUserTypeDTO_InvalidJSONSchema(t *testing.T) {
	yamlData := `
id: "invalid-json-schema"
name: "Invalid JSON Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
schema: |
  {invalid json here}
`

	schemaDTO, err := parseToUserTypeDTO([]byte(yamlData))

	assert.Error(t, err)
	assert.Nil(t, schemaDTO)
	assert.Contains(t, err.Error(), "invalid JSON")
}

// TestParseToUserTypeDTO_EmptySchemaField tests parsing with empty schema field
func TestParseToUserTypeDTO_EmptySchemaField(t *testing.T) {
	yamlData := `
id: "empty-schema"
name: "Empty Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
schema: ""
`

	schemaDTO, err := parseToUserTypeDTO([]byte(yamlData))

	assert.Error(t, err)
	assert.Nil(t, schemaDTO)
	assert.Contains(t, err.Error(), "invalid JSON")
}

// TestValidateUserTypeWithOUCheck tests the validation logic that would be used during initialization
// This tests the same validation path that occurs before the OU service call in Initialize()
func TestValidateUserTypeWithOUCheck(t *testing.T) {
	testCases := []struct {
		name          string
		schema        UserType
		shouldBeValid bool
		errorContains string
	}{
		{
			name: "Valid schema with valid OU ID",
			schema: UserType{
				ID:                 "valid-schema-001",
				Name:               "Valid Schema",
				OrganizationUnitID: "550e8400-e29b-41d4-a716-446655440000",
				Schema:             []byte(`{"email":{"type":"string","required":true}}`),
			},
			shouldBeValid: true,
		},
		{
			name: "Invalid schema - empty name",
			schema: UserType{
				ID:                 "invalid-001",
				Name:               "",
				OrganizationUnitID: "550e8400-e29b-41d4-a716-446655440000",
				Schema:             []byte(`{"email":{"type":"string"}}`),
			},
			shouldBeValid: false,
			errorContains: "user type name must not be empty",
		},
		{
			name: "Invalid schema - empty OU ID",
			schema: UserType{
				ID:                 "invalid-002",
				Name:               "Test Schema",
				OrganizationUnitID: "",
				Schema:             []byte(`{"email":{"type":"string"}}`),
			},
			shouldBeValid: false,
			errorContains: "organization unit id must not be empty",
		},
		{
			name: "Invalid schema - malformed OU ID",
			schema: UserType{
				ID:                 "invalid-003",
				Name:               "Test Schema",
				OrganizationUnitID: "not-a-valid-uuid",
				Schema:             []byte(`{"email":{"type":"string"}}`),
			},
			shouldBeValid: false,
			errorContains: "organization unit id is not a valid UUID",
		},
		{
			name: "Invalid schema - empty schema definition",
			schema: UserType{
				ID:                 "invalid-004",
				Name:               "Test Schema",
				OrganizationUnitID: "550e8400-e29b-41d4-a716-446655440000",
				Schema:             []byte{},
			},
			shouldBeValid: false,
			errorContains: "schema definition must not be empty",
		},
		{
			name: "Invalid schema - malformed schema definition",
			schema: UserType{
				ID:                 "invalid-005",
				Name:               "Test Schema",
				OrganizationUnitID: "550e8400-e29b-41d4-a716-446655440000",
				Schema:             []byte(`{"email":"not-an-object"}`),
			},
			shouldBeValid: false,
			errorContains: "property definition must be an object",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUserTypeDefinition(tc.schema)

			if tc.shouldBeValid {
				assert.Nil(t, err, "Expected schema to be valid but got error: %v", err)
			} else {
				assert.NotNil(t, err, "Expected validation to fail")
				if err != nil {
					assert.Contains(t, err.ErrorDescription, tc.errorContains,
						"Error message should contain expected text")
					assert.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
				}
			}
		})
	}
}

// TestOUServiceInteractionDuringValidation tests that the OU service would be called correctly
// This validates the logic flow that occurs in Initialize() when checking OU existence
func TestOUServiceInteractionDuringValidation(t *testing.T) {
	testCases := []struct {
		name           string
		ouID           string
		ouExists       bool
		ouServiceError *serviceerror.ServiceError
		expectedResult string
	}{
		{
			name:           "OU exists - should pass",
			ouID:           "550e8400-e29b-41d4-a716-446655440000",
			ouExists:       true,
			ouServiceError: nil,
			expectedResult: "success",
		},
		{
			name:           "OU does not exist - should fail",
			ouID:           "550e8400-e29b-41d4-a716-446655440001",
			ouExists:       false,
			ouServiceError: nil,
			expectedResult: "ou_not_found",
		},
		{
			name:     "OU service returns error - should fail",
			ouID:     "550e8400-e29b-41d4-a716-446655440002",
			ouExists: false,
			ouServiceError: &serviceerror.ServiceError{
				Code:             "OUS-5000",
				Type:             serviceerror.ServerErrorType,
				Error:            "Internal server error",
				ErrorDescription: "Failed to query organization unit",
			},
			expectedResult: "service_error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

			// Mock the GetOrganizationUnit call that happens in Initialize()
			if tc.ouServiceError != nil {
				mockOUService.On("GetOrganizationUnit", tc.ouID).
					Return(oupkg.OrganizationUnit{}, tc.ouServiceError).Once()
			} else if tc.ouExists {
				mockOUService.On("GetOrganizationUnit", tc.ouID).
					Return(oupkg.OrganizationUnit{ID: tc.ouID}, (*serviceerror.ServiceError)(nil)).Once()
			} else {
				mockOUService.On("GetOrganizationUnit", tc.ouID).
					Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{
						Code:             "OUS-1002",
						Type:             serviceerror.ClientErrorType,
						Error:            "Organization unit not found",
						ErrorDescription: "The organization unit does not exist",
					}).Once()
			}

			// Simulate the OU validation logic from Initialize()
			_, svcErr := mockOUService.GetOrganizationUnit(tc.ouID)

			switch tc.expectedResult {
			case "success":
				assert.Nil(t, svcErr, "Expected no error when OU exists")
			case "ou_not_found":
				assert.NotNil(t, svcErr, "Expected error when OU does not exist")
				assert.Equal(t, "OUS-1002", svcErr.Code)
			case "service_error":
				assert.NotNil(t, svcErr, "Expected error when OU service fails")
				assert.Equal(t, "OUS-5000", svcErr.Code)
			}

			mockOUService.AssertExpectations(t)
		})
	}
}

// TestParseAndValidateUserTypeFlow tests the complete flow of parsing and validating
// This simulates what happens in Initialize() before the OU check
func TestParseAndValidateUserTypeFlow(t *testing.T) {
	testCases := []struct {
		name          string
		yamlData      string
		expectParseOK bool
		expectValidOK bool
		errorContains string
	}{
		{
			name: "Valid YAML and schema",
			yamlData: `
id: "flow-test-001"
name: "Flow Test Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
schema: |
  {
    "email": {"type": "string", "required": true}
  }
`,
			expectParseOK: true,
			expectValidOK: true,
		},
		{
			name: "Valid YAML but invalid schema definition",
			yamlData: `
id: "flow-test-002"
name: "Invalid Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
schema: |
  {
    "email": {"required": true}
  }
`,
			expectParseOK: true,
			expectValidOK: false,
			errorContains: "missing required 'type' field",
		},
		{
			name: "Valid YAML but empty schema name",
			yamlData: `
id: "flow-test-003"
name: ""
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
schema: |
  {
    "email": {"type": "string"}
  }
`,
			expectParseOK: true,
			expectValidOK: false,
			errorContains: "user type name must not be empty",
		},
		{
			name: "Invalid YAML structure",
			yamlData: `
this is not valid yaml:
  - broken structure
`,
			expectParseOK: false,
			expectValidOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Step 1: Parse YAML (as done in Initialize)
			schemaDTO, parseErr := parseToUserTypeDTO([]byte(tc.yamlData))

			if tc.expectParseOK {
				assert.NoError(t, parseErr, "Expected YAML parsing to succeed")
				assert.NotNil(t, schemaDTO)

				// Step 2: Validate schema (as done in Initialize before OU check)
				validationErr := validateUserTypeDefinition(*schemaDTO)

				if tc.expectValidOK {
					assert.Nil(t, validationErr, "Expected validation to succeed")
				} else {
					assert.NotNil(t, validationErr, "Expected validation to fail")
					if validationErr != nil && tc.errorContains != "" {
						assert.Contains(t, validationErr.ErrorDescription, tc.errorContains)
					}
				}
			} else {
				assert.Error(t, parseErr, "Expected YAML parsing to fail")
				assert.Nil(t, schemaDTO)
			}
		})
	}
}

// TestInitialize_WithDeclarativeResourcesEnabled_InvalidYAML tests Initialize with invalid YAML files
//
//nolint:dupl // Similar test setup required for different error scenarios
func TestInitialize_WithDeclarativeResourcesEnabled_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	confDir := tmpDir + "/repository/resources"
	schemaDir := confDir + "/user_types"

	err := os.MkdirAll(schemaDir, 0750)
	assert.NoError(t, err)

	// Create an invalid YAML file
	invalidYAML := `invalid yaml content
  - this is not: valid
`
	err = os.WriteFile(schemaDir+"/invalid-schema.yaml", []byte(invalidYAML), 0600)
	assert.NoError(t, err)

	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: true,
		},
		Crypto: config.CryptoConfig{
			Encryption: config.EncryptionConfig{
				Key: testCryptoKey,
			},
		},
	}

	config.ResetThunderRuntime()
	err = config.InitializeThunderRuntime(tmpDir, testConfig)
	assert.NoError(t, err)
	defer config.ResetThunderRuntime()

	mux := http.NewServeMux()
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	// Initialize should return an error due to invalid YAML
	_, _, err = Initialize(mux, mockOUService)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load user type resources")
}

// TestInitialize_WithDeclarativeResourcesEnabled_ValidationFailure tests Initialize with validation errors
//
//nolint:dupl // Similar test setup required for different error scenarios
func TestInitialize_WithDeclarativeResourcesEnabled_ValidationFailure(t *testing.T) {
	tmpDir := t.TempDir()
	confDir := tmpDir + "/repository/resources"
	schemaDir := confDir + "/user_types"

	err := os.MkdirAll(schemaDir, 0750)
	assert.NoError(t, err)

	// Create crypto directory
	cryptoDir := tmpDir + "/repository/conf"
	err = os.MkdirAll(cryptoDir, 0750)
	assert.NoError(t, err)

	// Create a YAML file with invalid configuration (empty name)
	invalidSchemaYAML := `id: "invalid-schema"
name: ""
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
schema: |
  {
    "email": {"type": "string"}
  }
`
	err = os.WriteFile(schemaDir+"/invalid-schema.yaml", []byte(invalidSchemaYAML), 0600)
	assert.NoError(t, err)

	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: true,
		},
		Crypto: config.CryptoConfig{
			Encryption: config.EncryptionConfig{
				Key: testCryptoKey,
			},
		},
	}

	config.ResetThunderRuntime()
	err = config.InitializeThunderRuntime(tmpDir, testConfig)
	assert.NoError(t, err)
	defer config.ResetThunderRuntime()

	mux := http.NewServeMux()
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	// Initialize should return an error due to validation failure
	_, _, err = Initialize(mux, mockOUService)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load user type resources")
}

// TestInitialize_WithDeclarativeResourcesEnabled_OUServiceError tests Initialize when OU service fails
func TestInitialize_WithDeclarativeResourcesEnabled_OUServiceError(t *testing.T) {
	tmpDir := t.TempDir()
	confDir := tmpDir + "/repository/resources"
	schemaDir := confDir + "/user_types"

	err := os.MkdirAll(schemaDir, 0750)
	assert.NoError(t, err)

	// Create crypto directory
	cryptoDir := tmpDir + "/repository/conf"
	err = os.MkdirAll(cryptoDir, 0750)
	assert.NoError(t, err)

	// Create a valid YAML file
	validSchemaYAML := `id: "test-schema"
name: "Test Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
allow_self_registration: true
schema: |
  {
    "email": {"type": "string", "required": true}
  }
`
	err = os.WriteFile(schemaDir+"/test-schema.yaml", []byte(validSchemaYAML), 0600)
	assert.NoError(t, err)

	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: true,
		},
		Crypto: config.CryptoConfig{
			Encryption: config.EncryptionConfig{
				Key: testCryptoKey,
			},
		},
	}

	config.ResetThunderRuntime()
	err = config.InitializeThunderRuntime(tmpDir, testConfig)
	assert.NoError(t, err)
	defer config.ResetThunderRuntime()

	mux := http.NewServeMux()
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	// Mock OU service to return an error
	mockOUService.On("GetOrganizationUnit", "550e8400-e29b-41d4-a716-446655440000").
		Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{
			Code:             "OUS-1002",
			Type:             serviceerror.ClientErrorType,
			Error:            "Organization unit not found",
			ErrorDescription: "The organization unit does not exist",
		}).Once()

	// Initialize should return an error due to OU service failure
	_, _, err = Initialize(mux, mockOUService)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load user type resources")

	mockOUService.AssertExpectations(t)
}

// TestInitialize_WithDeclarativeResourcesEnabled_InvalidJSONSchema tests Initialize with invalid JSON in schema
//
//nolint:dupl // Similar test setup required for different error scenarios
func TestInitialize_WithDeclarativeResourcesEnabled_InvalidJSONSchema(t *testing.T) {
	tmpDir := t.TempDir()
	confDir := tmpDir + "/repository/resources"
	schemaDir := confDir + "/user_types"

	err := os.MkdirAll(schemaDir, 0750)
	assert.NoError(t, err)

	// Create crypto directory
	cryptoDir := tmpDir + "/repository/conf"
	err = os.MkdirAll(cryptoDir, 0750)
	assert.NoError(t, err)

	// Create a YAML file with invalid JSON in schema field
	invalidJSONYAML := `id: "invalid-json-schema"
name: "Invalid JSON Schema"
organization_unit_id: "550e8400-e29b-41d4-a716-446655440000"
schema: |
  {invalid json here}
`
	err = os.WriteFile(schemaDir+"/invalid-json.yaml", []byte(invalidJSONYAML), 0600)
	assert.NoError(t, err)

	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: true,
		},
		Crypto: config.CryptoConfig{
			Encryption: config.EncryptionConfig{
				Key: testCryptoKey,
			},
		},
	}

	config.ResetThunderRuntime()
	err = config.InitializeThunderRuntime(tmpDir, testConfig)
	assert.NoError(t, err)
	defer config.ResetThunderRuntime()

	mux := http.NewServeMux()
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	// Initialize should return an error due to invalid JSON
	_, _, err = Initialize(mux, mockOUService)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load user type resources")
}
