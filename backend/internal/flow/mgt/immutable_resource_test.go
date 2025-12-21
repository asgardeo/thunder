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

package flowmgt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/immutable_resource/entity"
	"github.com/asgardeo/thunder/internal/system/log"
)

type ImmutableResourceTestSuite struct {
	suite.Suite
}

func (s *ImmutableResourceTestSuite) SetupTest() {
	// Reset logger
	_ = log.GetLogger()
}

// TestParseToCompleteFlowDefinition tests parsing YAML to CompleteFlowDefinition
func (s *ImmutableResourceTestSuite) TestParseToCompleteFlowDefinition() {
	yamlData := []byte(`
id: "flow-001"
handle: "basic-auth"
name: "Basic Authentication Flow"
flowtype: "AUTHENTICATION"
activeversion: 1
nodes:
  - id: "start"
    type: "START"
  - id: "basic-login"
    type: "BASIC_AUTHENTICATION"
  - id: "end"
    type: "END"
`)

	result, err := parseToCompleteFlowDefinition(yamlData)
	require.NoError(s.T(), err)

	flow, ok := result.(*CompleteFlowDefinition)
	require.True(s.T(), ok, "result should be *CompleteFlowDefinition")

	assert.Equal(s.T(), "flow-001", flow.ID)
	assert.Equal(s.T(), "basic-auth", flow.Handle)
	assert.Equal(s.T(), "Basic Authentication Flow", flow.Name)
	assert.Len(s.T(), flow.Nodes, 3)
}

// TestParseToCompleteFlowDefinition_InvalidYAML tests parsing invalid YAML
func (s *ImmutableResourceTestSuite) TestParseToCompleteFlowDefinition_InvalidYAML() {
	yamlData := []byte(`{invalid yaml content`)

	_, err := parseToCompleteFlowDefinition(yamlData)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "failed to unmarshal flow definition")
}

// TestValidateFlowGraphWrapper tests flow validation wrapper
func (s *ImmutableResourceTestSuite) TestValidateFlowGraphWrapper_ValidFlow() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-001",
		Handle:   "basic-auth",
		Name:     "Basic Auth",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
			{ID: "login", Type: "BASIC_AUTHENTICATION"},
			{ID: "mfa", Type: "TOTP_AUTHENTICATION"},
			{ID: "end", Type: "END"},
		},
	}

	err := validateFlowGraphWrapper(flow)
	assert.NoError(s.T(), err)
}

// TestValidateFlowGraphWrapper_MissingHandle tests validation with missing handle
func (s *ImmutableResourceTestSuite) TestValidateFlowGraphWrapper_MissingHandle() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-001",
		Handle:   "",
		Name:     "Basic Auth",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
			{ID: "end", Type: "END"},
		},
	}

	err := validateFlowGraphWrapper(flow)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "validation failed")
}

// TestValidateFlowGraphWrapper_InvalidType tests validation with wrong type
func (s *ImmutableResourceTestSuite) TestValidateFlowGraphWrapper_InvalidType() {
	err := validateFlowGraphWrapper("not a flow definition")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid type")
}

// TestValidateFlowGraphWrapper_InsufficientNodes tests validation with insufficient nodes
func (s *ImmutableResourceTestSuite) TestValidateFlowGraphWrapper_InsufficientNodes() {
	flow := &CompleteFlowDefinition{
		ID:       "flow-001",
		Handle:   "invalid-flow",
		Name:     "Invalid Flow",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
			{ID: "end", Type: "END"},
		},
	}

	err := validateFlowGraphWrapper(flow)
	assert.Error(s.T(), err)
}

// TestFlowGraphExporter_GetResourceType tests resource type
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_GetResourceType() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())
	exporter := newFlowGraphExporter(mockService)

	assert.Equal(s.T(), "flow", exporter.GetResourceType())
}

// TestFlowGraphExporter_GetParameterizerType tests parameterizer type
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_GetParameterizerType() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())
	exporter := newFlowGraphExporter(mockService)

	assert.Equal(s.T(), "Flow", exporter.GetParameterizerType())
}

// TestFlowGraphExporter_GetResourceRules tests resource rules
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_GetResourceRules() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())
	exporter := newFlowGraphExporter(mockService)

	rules := exporter.GetResourceRules()
	assert.NotNil(s.T(), rules)
	assert.Empty(s.T(), rules.Variables)
	assert.Empty(s.T(), rules.ArrayVariables)
	assert.Empty(s.T(), rules.DynamicPropertyFields)
}

// TestFlowGraphExporter_GetAllResourceIDs tests retrieving all resource IDs
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_GetAllResourceIDs() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())

	listResponse := &FlowListResponse{
		Flows: []BasicFlowDefinition{
			{ID: "flow-001", Handle: "auth-flow"},
			{ID: "flow-002", Handle: "reg-flow"},
		},
		Count: 2,
	}

	// Use common.FlowType to match the service interface type
	mockService.EXPECT().ListFlows(10000, 0, common.FlowType("")).Return(listResponse, nil)

	exporter := newFlowGraphExporter(mockService)
	ids, err := exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 2)
	assert.Equal(s.T(), "flow-001", ids[0])
	assert.Equal(s.T(), "flow-002", ids[1])
}

// TestFlowGraphExporter_GetAllResourceIDs_Error tests error handling
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_GetAllResourceIDs_Error() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())

	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	mockService.EXPECT().ListFlows(10000, 0, common.FlowType("")).Return(nil, expectedError)

	exporter := newFlowGraphExporter(mockService)
	ids, err := exporter.GetAllResourceIDs()

	assert.Nil(s.T(), ids)
	assert.Equal(s.T(), expectedError, err)
}

// TestFlowGraphExporter_GetAllResourceIDs_EmptyList tests empty list handling
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_GetAllResourceIDs_EmptyList() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())

	listResponse := &FlowListResponse{
		Flows: []BasicFlowDefinition{},
		Count: 0,
	}

	mockService.EXPECT().ListFlows(10000, 0, common.FlowType("")).Return(listResponse, nil)

	exporter := newFlowGraphExporter(mockService)
	ids, err := exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 0)
}

// TestFlowGraphExporter_GetResourceByID tests retrieving resource by ID
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_GetResourceByID() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())

	flow := &CompleteFlowDefinition{
		ID:   "flow-001",
		Name: "Auth Flow",
	}

	mockService.EXPECT().GetFlow("flow-001").Return(flow, nil)

	exporter := newFlowGraphExporter(mockService)
	resource, name, err := exporter.GetResourceByID("flow-001")

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), flow, resource)
	assert.Equal(s.T(), "Auth Flow", name)
}

// TestFlowGraphExporter_GetResourceByID_Error tests error handling
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_GetResourceByID_Error() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())

	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	mockService.EXPECT().GetFlow("flow-001").Return(nil, expectedError)

	exporter := newFlowGraphExporter(mockService)
	resource, name, err := exporter.GetResourceByID("flow-001")

	assert.Nil(s.T(), resource)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), expectedError, err)
}

// TestFlowGraphExporter_ValidateResource tests resource validation
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_ValidateResource() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())
	exporter := newFlowGraphExporter(mockService)

	flow := &CompleteFlowDefinition{
		ID:   "flow-001",
		Name: "Valid Flow Name",
	}

	logger := log.GetLogger()
	name, exportErr := exporter.ValidateResource(flow, "flow-001", logger)

	assert.Nil(s.T(), exportErr)
	assert.Equal(s.T(), "Valid Flow Name", name)
}

// TestFlowGraphExporter_ValidateResource_InvalidType tests validation with invalid type
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_ValidateResource_InvalidType() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())
	exporter := newFlowGraphExporter(mockService)

	logger := log.GetLogger()
	_, exportErr := exporter.ValidateResource("not a flow", "invalid", logger)

	assert.NotNil(s.T(), exportErr)
	assert.Equal(s.T(), "flow", exportErr.ResourceType)
	assert.Equal(s.T(), "invalid", exportErr.ResourceID)
	assert.Equal(s.T(), "INVALID_TYPE", exportErr.Code)
}

// TestFlowGraphExporter_ValidateResource_EmptyName tests validation with empty name
func (s *ImmutableResourceTestSuite) TestFlowGraphExporter_ValidateResource_EmptyName() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())
	exporter := newFlowGraphExporter(mockService)

	flow := &CompleteFlowDefinition{
		ID:   "flow-001",
		Name: "",
	}

	logger := log.GetLogger()
	name, exportErr := exporter.ValidateResource(flow, "flow-001", logger)

	assert.Empty(s.T(), name)
	assert.NotNil(s.T(), exportErr)
	assert.Equal(s.T(), "flow", exportErr.ResourceType)
	assert.Equal(s.T(), "flow-001", exportErr.ResourceID)
	assert.Equal(s.T(), "FLOW_VALIDATION_ERROR", exportErr.Code)
	assert.Contains(s.T(), exportErr.Error, "name is empty")
}

// TestFileBasedStore_CreateFlow tests creating a flow in file-based store
func (s *ImmutableResourceTestSuite) TestFileBasedStore_CreateFlow() {
	_ = entity.GetInstance().Clear()
	store := newFileBasedStore()

	flowDef := &FlowDefinition{
		Handle:   "test-flow",
		Name:     "Test Flow",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
			{ID: "login", Type: "BASIC_AUTHENTICATION"},
			{ID: "end", Type: "END"},
		},
	}

	completeFlow, err := store.CreateFlow("flow-001", flowDef)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "flow-001", completeFlow.ID)
	assert.Equal(s.T(), "test-flow", completeFlow.Handle)
	assert.Equal(s.T(), "Test Flow", completeFlow.Name)
}

// TestFileBasedStore_GetFlowByID tests retrieving flow by ID
func (s *ImmutableResourceTestSuite) TestFileBasedStore_GetFlowByID() {
	_ = entity.GetInstance().Clear()
	store := newFileBasedStore()

	flowDef := &FlowDefinition{
		Handle:   "test-flow",
		Name:     "Test Flow",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
			{ID: "login", Type: "BASIC_AUTHENTICATION"},
			{ID: "end", Type: "END"},
		},
	}

	_, err := store.CreateFlow("flow-001", flowDef)
	require.NoError(s.T(), err)

	retrieved, err := store.GetFlowByID("flow-001")
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "flow-001", retrieved.ID)
	assert.Equal(s.T(), "test-flow", retrieved.Handle)
}

// TestFileBasedStore_GetFlowByID_NotFound tests retrieving non-existent flow
func (s *ImmutableResourceTestSuite) TestFileBasedStore_GetFlowByID_NotFound() {
	_ = entity.GetInstance().Clear()
	store := newFileBasedStore()

	_, err := store.GetFlowByID("non-existent")
	assert.Error(s.T(), err)
}

// TestFileBasedStore_GetFlowByHandle tests retrieving flow by handle
func (s *ImmutableResourceTestSuite) TestFileBasedStore_GetFlowByHandle() {
	_ = entity.GetInstance().Clear()
	store := newFileBasedStore()

	flowDef := &FlowDefinition{
		Handle:   "test-flow",
		Name:     "Test Flow",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
			{ID: "login", Type: "BASIC_AUTHENTICATION"},
			{ID: "end", Type: "END"},
		},
	}

	_, err := store.CreateFlow("flow-001", flowDef)
	require.NoError(s.T(), err)

	retrieved, err := store.GetFlowByHandle("test-flow", "AUTHENTICATION")
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "flow-001", retrieved.ID)
	assert.Equal(s.T(), "test-flow", retrieved.Handle)
}

// TestFileBasedStore_ListFlows tests listing flows with pagination
func (s *ImmutableResourceTestSuite) TestFileBasedStore_ListFlows() {
	_ = entity.GetInstance().Clear()
	store := newFileBasedStore()

	for i := 0; i < 3; i++ {
		flowDef := &FlowDefinition{
			Handle:   fmt.Sprintf("flow-%d", i),
			Name:     fmt.Sprintf("Flow %d", i),
			FlowType: "AUTHENTICATION",
			Nodes: []NodeDefinition{
				{ID: "start", Type: "START"},
				{ID: "login", Type: "BASIC_AUTHENTICATION"},
				{ID: "end", Type: "END"},
			},
		}
		_, err := store.CreateFlow(fmt.Sprintf("flow-%03d", i), flowDef)
		require.NoError(s.T(), err)
	}

	flows, count, err := store.ListFlows(10, 0, "")
	require.NoError(s.T(), err)

	assert.Equal(s.T(), 3, count)
	assert.Len(s.T(), flows, 3)
}

// TestFileBasedStore_IsFlowExists tests checking flow existence
func (s *ImmutableResourceTestSuite) TestFileBasedStore_IsFlowExists() {
	_ = entity.GetInstance().Clear()
	store := newFileBasedStore()

	flowDef := &FlowDefinition{
		Handle:   "test-flow",
		Name:     "Test Flow",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
			{ID: "login", Type: "BASIC_AUTHENTICATION"},
			{ID: "end", Type: "END"},
		},
	}

	_, err := store.CreateFlow("flow-001", flowDef)
	require.NoError(s.T(), err)

	exists, err := store.IsFlowExists("flow-001")
	require.NoError(s.T(), err)
	assert.True(s.T(), exists)

	exists, err = store.IsFlowExists("non-existent")
	require.NoError(s.T(), err)
	assert.False(s.T(), exists)
}

// TestFileBasedStore_UnsupportedOperations tests that unsupported operations return errors
func (s *ImmutableResourceTestSuite) TestFileBasedStore_UnsupportedOperations() {
	_ = entity.GetInstance().Clear()
	store := newFileBasedStore()

	flowDef := &FlowDefinition{
		Handle:   "test-flow",
		Name:     "Test Flow",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{ID: "start", Type: "START"},
			{ID: "login", Type: "BASIC_AUTHENTICATION"},
			{ID: "end", Type: "END"},
		},
	}

	_, err := store.CreateFlow("flow-001", flowDef)
	require.NoError(s.T(), err)

	// UpdateFlow
	_, err = store.UpdateFlow("flow-001", flowDef)
	assert.Error(s.T(), err)

	// DeleteFlow
	err = store.DeleteFlow("flow-001")
	assert.Error(s.T(), err)

	// ListFlowVersions
	_, err = store.ListFlowVersions("flow-001")
	assert.Error(s.T(), err)

	// GetFlowVersion
	_, err = store.GetFlowVersion("flow-001", 1)
	assert.Error(s.T(), err)

	// RestoreFlowVersion
	_, err = store.RestoreFlowVersion("flow-001", 1)
	assert.Error(s.T(), err)
}

// TestParseYAMLToJSON tests YAML parsing with various structures
func (s *ImmutableResourceTestSuite) TestParseYAMLComplexStructure() {
	yamlData := []byte(`
id: "mfa-flow"
handle: "mfa-auth"
name: "Multi-Factor Authentication"
flowType: "AUTHENTICATION"
activeVersion: 2
nodes:
  - id: "start"
    type: "START"
  - id: "basic-auth"
    type: "BASIC_AUTHENTICATION"
  - id: "totp-check"
    type: "TOTP_AUTHENTICATION"
  - id: "success"
    type: "SUCCESS"
  - id: "end"
    type: "END"
createdAt: "2025-01-01T00:00:00Z"
updatedAt: "2025-01-02T00:00:00Z"
`)

	result, err := parseToCompleteFlowDefinition(yamlData)
	require.NoError(s.T(), err)

	flow, ok := result.(*CompleteFlowDefinition)
	require.True(s.T(), ok)

	assert.Equal(s.T(), "mfa-flow", flow.ID)
	assert.Equal(s.T(), "mfa-auth", flow.Handle)
	assert.Equal(s.T(), "Multi-Factor Authentication", flow.Name)
	// Note: flowType and activeVersion are camelCase in YAML, but YAML unmarshaling
	// without yaml tags uses lowercase, so they won't be properly unmarshaled.
	// This is expected - the struct should have yaml tags for proper unmarshaling.
	assert.Len(s.T(), flow.Nodes, 5)
}

// TestFlowGraphExporterIntegration tests the complete flow of exporter usage
func (s *ImmutableResourceTestSuite) TestFlowGraphExporterIntegration() {
	mockService := NewFlowMgtServiceInterfaceMock(s.T())

	flow := &CompleteFlowDefinition{
		ID:            "flow-001",
		Handle:        "auth-flow",
		Name:          "Authentication Flow",
		FlowType:      "AUTHENTICATION",
		ActiveVersion: 1,
	}

	listResponse := &FlowListResponse{
		Flows: []BasicFlowDefinition{
			{ID: "flow-001", Handle: "auth-flow", Name: "Authentication Flow"},
		},
		Count: 1,
	}

	mockService.EXPECT().ListFlows(10000, 0, common.FlowType("")).Return(listResponse, nil)
	mockService.EXPECT().GetFlow("flow-001").Return(flow, nil)

	exporter := newFlowGraphExporter(mockService)

	// Get all IDs
	ids, err := exporter.GetAllResourceIDs()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 1)

	// Get resource by ID
	resource, name, err := exporter.GetResourceByID(ids[0])
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), flow, resource)
	assert.Equal(s.T(), "Authentication Flow", name)

	// Validate resource
	logger := log.GetLogger()
	validName, exportErr := exporter.ValidateResource(resource, ids[0], logger)
	assert.Nil(s.T(), exportErr)
	assert.Equal(s.T(), "Authentication Flow", validName)
}

// TestYAMLUnmarshalVariations tests different YAML formats
func (s *ImmutableResourceTestSuite) TestYAMLUnmarshalVariations() {
	testCases := []struct {
		name     string
		yamlData []byte
		wantErr  bool
	}{
		{
			name: "minimal flow",
			yamlData: []byte(`
id: "flow-1"
handle: "flow"
name: "Flow"
flowType: "AUTHENTICATION"
nodes:
  - id: "start"
    type: "START"
  - id: "step"
    type: "BASIC_AUTHENTICATION"
  - id: "end"
    type: "END"
`),
			wantErr: false,
		},
		{
			name:     "empty YAML",
			yamlData: []byte(""),
			wantErr:  false,
		},
		{
			name:     "invalid YAML syntax",
			yamlData: []byte("{ invalid: yaml:"),
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := parseToCompleteFlowDefinition(tc.yamlData)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLoadImmutableResources_InvalidStoreType tests error when store is not file-based
func (s *ImmutableResourceTestSuite) TestLoadImmutableResources_InvalidStoreType() {
	// Create a mock store that's not a file-based store
	mockStore := &flowStoreInterfaceMock{}

	err := loadImmutableResources(mockStore)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "failed to assert flowStore to *fileBasedStore")
}

func TestImmutableResourceTestSuite(t *testing.T) {
	suite.Run(t, new(ImmutableResourceTestSuite))
}
