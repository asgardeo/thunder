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

package flowexec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/flowmgtmock"
)

func TestInitiateFlowNilContext(t *testing.T) {
	// Setup
	service := &flowExecService{}

	// Execute
	flowID, err := service.InitiateFlow(nil)

	// Assert
	assert.NotNil(t, err)
	assert.Empty(t, flowID)
	assert.Equal(t, "FES-1008", err.Code)
}

func TestInitiateFlowEmptyApplicationID(t *testing.T) {
	// Setup
	service := &flowExecService{}

	initContext := &FlowInitContext{
		ApplicationID: "",
		FlowType:      "AUTHENTICATION",
		RuntimeData:   map[string]string{},
	}

	// Execute
	flowID, err := service.InitiateFlow(initContext)

	// Assert
	assert.NotNil(t, err)
	assert.Empty(t, flowID)
	assert.Equal(t, "FES-1008", err.Code)
}

func TestInitiateFlowEmptyFlowType(t *testing.T) {
	// Setup
	service := &flowExecService{}

	initContext := &FlowInitContext{
		ApplicationID: "test-app",
		FlowType:      "",
		RuntimeData:   map[string]string{},
	}

	// Execute
	flowID, err := service.InitiateFlow(initContext)

	// Assert
	assert.NotNil(t, err)
	assert.Empty(t, flowID)
	assert.Equal(t, "FES-1008", err.Code)
}

func TestInitiateFlowInvalidFlowType(t *testing.T) {
	// Setup
	service := &flowExecService{}

	initContext := &FlowInitContext{
		ApplicationID: "test-app",
		FlowType:      "INVALID_TYPE",
		RuntimeData:   map[string]string{},
	}

	// Execute
	flowID, err := service.InitiateFlow(initContext)

	// Assert
	assert.NotNil(t, err)
	assert.Empty(t, flowID)
	assert.Equal(t, "FES-1005", err.Code) // ErrorInvalidFlowType
}

func TestInitiateFlowSuccessScenarios(t *testing.T) {
	appID := "test-app-123"

	testConfig := &config.Config{}
	_ = config.InitializeThunderRuntime("/tmp/test", testConfig)

	flowFactory, _ := core.Initialize()
	testGraph := flowFactory.CreateGraph("auth-graph-1", common.FlowTypeAuthentication)

	// Mock application and graph - shared across all test cases
	mockApp := &appmodel.Application{
		ID:              "app-id-123",
		AuthFlowGraphID: "auth-graph-1",
	}

	tests := []struct {
		name                     string
		runtimeData              map[string]string
		setRuntimeDataField      bool // whether to explicitly set the RuntimeData field
		expectedRuntimeDataCheck func(ctx EngineContext) bool
	}{
		{
			name: "with runtime data",
			runtimeData: map[string]string{
				"permissions": "perm1 perm2 perm3",
				"state":       "random-state-value",
				"type":        "code",
			},
			setRuntimeDataField: true,
			expectedRuntimeDataCheck: func(ctx EngineContext) bool {
				// Verify RuntimeData is preserved
				return ctx.RuntimeData != nil &&
					ctx.RuntimeData["permissions"] == "perm1 perm2 perm3" &&
					ctx.RuntimeData["state"] == "random-state-value" &&
					ctx.RuntimeData["type"] == "code"
			},
		},
		{
			name:                "with nil runtime data",
			runtimeData:         nil,
			setRuntimeDataField: true,
			expectedRuntimeDataCheck: func(ctx EngineContext) bool {
				// Verify RuntimeData is nil (since initContext.RuntimeData is nil and len > 0 check fails)
				return ctx.RuntimeData == nil
			},
		},
		{
			name:                "with empty runtime data",
			runtimeData:         map[string]string{},
			setRuntimeDataField: true,
			expectedRuntimeDataCheck: func(ctx EngineContext) bool {
				// Verify RuntimeData is not nil and empty
				return ctx.RuntimeData != nil && len(ctx.RuntimeData) == 0
			},
		},
		{
			name:                "without runtime data field",
			runtimeData:         nil, // This won't be used since setRuntimeDataField is false
			setRuntimeDataField: false,
			expectedRuntimeDataCheck: func(ctx EngineContext) bool {
				// Verify RuntimeData is nil (since initContext.RuntimeData is nil and len > 0 check fails)
				return ctx.RuntimeData == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockStore := newFlowStoreInterfaceMock(t)
			mockAppService := applicationmock.NewApplicationServiceInterfaceMock(t)
			mockFlowMgtSvc := flowmgtmock.NewFlowMgtServiceInterfaceMock(t)

			// Create service with mocked dependencies
			service := &flowExecService{
				flowMgtService: mockFlowMgtSvc,
				flowStore:      mockStore,
				appService:     mockAppService,
				flowEngine:     nil,
			}

			initContext := &FlowInitContext{
				ApplicationID: appID,
				FlowType:      "AUTHENTICATION",
			}

			// Set RuntimeData field only if specified in test case
			if tt.setRuntimeDataField {
				initContext.RuntimeData = tt.runtimeData
			}

			// Setup expectations
			mockAppService.EXPECT().GetApplication(appID).Return(mockApp, nil)
			mockFlowMgtSvc.EXPECT().GetGraph("auth-graph-1").Return(testGraph, nil)
			mockStore.EXPECT().StoreFlowContext(mock.MatchedBy(func(ctx EngineContext) bool {
				// Verify flowID is generated
				if ctx.FlowID == "" {
					return false
				}
				// Verify runtime data according to test case expectation
				if !tt.expectedRuntimeDataCheck(ctx) {
					return false
				}
				// Verify AppID and FlowType
				if ctx.AppID != appID {
					return false
				}
				if ctx.FlowType != common.FlowTypeAuthentication {
					return false
				}
				return true
			})).Return(nil)

			// Execute
			flowID, svcErr := service.InitiateFlow(initContext)

			// Assert
			assert.NotEmpty(t, flowID)
			assert.Nil(t, svcErr)

			// All mocks automatically verified by mockery
		})
	}
}

func TestInitiateFlowErrorScenarios(t *testing.T) {
	appID := "test-app-123"

	testConfig := &config.Config{}
	_ = config.InitializeThunderRuntime("/tmp/test", testConfig)

	flowFactory, _ := core.Initialize()

	tests := []struct {
		name       string
		setupMocks func(
			*flowStoreInterfaceMock,
			*applicationmock.ApplicationServiceInterfaceMock,
			*flowmgtmock.FlowMgtServiceInterfaceMock,
		)
		expectedErrorCode        string
		expectedErrorDescription string
	}{
		{
			name: "error from getApplication - application not found",
			setupMocks: func(
				mockStore *flowStoreInterfaceMock,
				mockAppService *applicationmock.ApplicationServiceInterfaceMock,
				mockFlowMgtSvc *flowmgtmock.FlowMgtServiceInterfaceMock,
			) {
				// Import application package for its error constants
				appNotFoundErr := &serviceerror.ServiceError{
					Type:             serviceerror.ClientErrorType,
					Code:             "APP-1001", // ErrorApplicationNotFound.Code
					Error:            "Application not found",
					ErrorDescription: "The requested application could not be found",
				}
				mockAppService.EXPECT().GetApplication(appID).Return(nil, appNotFoundErr)
				// No other mocks needed as it fails early
			},
			expectedErrorCode: "FES-1003", // ErrorInvalidAppID (converted from application not found)
		},
		{
			name: "error from getApplication - other client error",
			setupMocks: func(
				mockStore *flowStoreInterfaceMock,
				mockAppService *applicationmock.ApplicationServiceInterfaceMock,
				mockFlowMgtSvc *flowmgtmock.FlowMgtServiceInterfaceMock,
			) {
				// Mock application service to return a different client error
				mockAppService.EXPECT().GetApplication(appID).Return(nil, &ErrorApplicationRetrievalClientError)
				// No other mocks needed as it fails early
			},
			expectedErrorCode: "FES-1007", // ErrorApplicationRetrievalClientError
		},
		{
			name: "error from flowMgtService.GetGraph - graph not found",
			setupMocks: func(
				mockStore *flowStoreInterfaceMock,
				mockAppService *applicationmock.ApplicationServiceInterfaceMock,
				mockFlowMgtSvc *flowmgtmock.FlowMgtServiceInterfaceMock,
			) {
				// Mock application service to return valid app
				mockApp := &appmodel.Application{
					ID:              "app-id-123",
					AuthFlowGraphID: "auth-graph-1",
				}
				mockAppService.EXPECT().GetApplication(appID).Return(mockApp, nil)

				// Mock flow management service to return error (graph not found)
				mockFlowMgtSvc.EXPECT().GetGraph("auth-graph-1").Return(nil, &serviceerror.InternalServerError)
				// No store mock needed as it fails before storing
			},
			expectedErrorCode: serviceerror.InternalServerError.Code,
		},
		{
			name: "error from storeContext - store failure",
			setupMocks: func(
				mockStore *flowStoreInterfaceMock,
				mockAppService *applicationmock.ApplicationServiceInterfaceMock,
				mockFlowMgtSvc *flowmgtmock.FlowMgtServiceInterfaceMock,
			) {
				// Mock application service to return valid app
				mockApp := &appmodel.Application{
					ID:              "app-id-123",
					AuthFlowGraphID: "auth-graph-1",
				}
				mockAppService.EXPECT().GetApplication(appID).Return(mockApp, nil)

				// Mock flow management service to return valid graph
				testGraph := flowFactory.CreateGraph("auth-graph-1", common.FlowTypeAuthentication)
				mockFlowMgtSvc.EXPECT().GetGraph("auth-graph-1").Return(testGraph, nil)

				// Mock store to return error
				mockStore.EXPECT().StoreFlowContext(mock.AnythingOfType("EngineContext")).Return(assert.AnError)
			},
			expectedErrorCode: serviceerror.InternalServerError.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockStore := newFlowStoreInterfaceMock(t)
			mockAppService := applicationmock.NewApplicationServiceInterfaceMock(t)
			mockFlowMgtSvc := flowmgtmock.NewFlowMgtServiceInterfaceMock(t)

			// Create service with mocked dependencies
			service := &flowExecService{
				flowMgtService: mockFlowMgtSvc,
				flowStore:      mockStore,
				appService:     mockAppService,
				flowEngine:     nil,
			}

			initContext := &FlowInitContext{
				ApplicationID: appID,
				FlowType:      "AUTHENTICATION",
				RuntimeData: map[string]string{
					"test": "data",
				},
			}

			// Setup test-specific mocks
			tt.setupMocks(mockStore, mockAppService, mockFlowMgtSvc)

			// Execute
			flowID, svcErr := service.InitiateFlow(initContext)

			// Assert
			assert.Empty(t, flowID)
			assert.NotNil(t, svcErr)
			assert.Equal(t, tt.expectedErrorCode, svcErr.Code)

			// All mocks automatically verified by mockery
		})
	}
}
