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
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const testFlowIDService = "test-flow-id"

type FlowMgtServiceTestSuite struct {
	suite.Suite
	service          FlowMgtServiceInterface
	mockStore        *flowStoreInterfaceMock
	mockInference    *flowInferenceServiceInterfaceMock
	mockGraphBuilder *graphBuilderInterfaceMock
}

func TestFlowMgtServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FlowMgtServiceTestSuite))
}

func (s *FlowMgtServiceTestSuite) SetupTest() {
	s.mockStore = newFlowStoreInterfaceMock(s.T())
	s.mockInference = newFlowInferenceServiceInterfaceMock(s.T())
	s.mockGraphBuilder = newGraphBuilderInterfaceMock(s.T())
	s.service = newFlowMgtService(s.mockStore, s.mockInference, s.mockGraphBuilder)

	testConfig := &config.Config{
		Flow: config.FlowConfig{
			AutoInferRegistration: false,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (s *FlowMgtServiceTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

// ListFlows tests

func (s *FlowMgtServiceTestSuite) TestListFlows_Success() {
	expectedFlows := []BasicFlowDefinition{
		{ID: "flow1", Handle: "test-handle", Name: "Flow 1", FlowType: common.FlowTypeAuthentication},
	}
	s.mockStore.EXPECT().ListFlows(30, 0, "").Return(expectedFlows, 1, nil)

	result, err := s.service.ListFlows(30, 0, "")

	s.Nil(err)
	s.NotNil(result)
	s.Equal(1, result.Count)
	s.Equal(1, result.TotalResults)
	s.Len(result.Flows, 1)
}

func (s *FlowMgtServiceTestSuite) TestListFlows_DefaultLimit() {
	s.mockStore.EXPECT().ListFlows(defaultPageSize, 0, "").Return([]BasicFlowDefinition{}, 0, nil)

	result, err := s.service.ListFlows(0, 0, "")

	s.Nil(err)
	s.NotNil(result)
}

func (s *FlowMgtServiceTestSuite) TestListFlows_MaxLimitExceeded() {
	s.mockStore.EXPECT().ListFlows(maxPageSize, 0, "").Return([]BasicFlowDefinition{}, 0, nil)

	result, err := s.service.ListFlows(1000, 0, "")

	s.Nil(err)
	s.NotNil(result)
}

func (s *FlowMgtServiceTestSuite) TestListFlows_NegativeOffset() {
	s.mockStore.EXPECT().ListFlows(30, 0, "").Return([]BasicFlowDefinition{}, 0, nil)

	result, err := s.service.ListFlows(30, -10, "")

	s.Nil(err)
	s.NotNil(result)
}

func (s *FlowMgtServiceTestSuite) TestListFlows_WithFlowType() {
	s.mockStore.EXPECT().ListFlows(30, 0, string(common.FlowTypeAuthentication)).
		Return([]BasicFlowDefinition{}, 0, nil)

	result, err := s.service.ListFlows(30, 0, common.FlowTypeAuthentication)

	s.Nil(err)
	s.NotNil(result)
}

func (s *FlowMgtServiceTestSuite) TestListFlows_InvalidFlowType() {
	result, err := s.service.ListFlows(30, 0, "invalid")

	s.Nil(result)
	s.Equal(&ErrorInvalidFlowType, err)
}

func (s *FlowMgtServiceTestSuite) TestListFlows_StoreError() {
	s.mockStore.EXPECT().ListFlows(30, 0, "").Return(nil, 0, errors.New("db error"))

	result, err := s.service.ListFlows(30, 0, "")

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

func (s *FlowMgtServiceTestSuite) TestListFlows_PaginationLinks() {
	s.mockStore.EXPECT().ListFlows(10, 20, "").Return([]BasicFlowDefinition{}, 100, nil)

	result, err := s.service.ListFlows(10, 20, "")

	s.Nil(err)
	s.NotNil(result)
	// Should have first, prev, next, last links
	s.Len(result.Links, 4)
}

func (s *FlowMgtServiceTestSuite) TestListFlows_PaginationLinksFirstPage() {
	s.mockStore.EXPECT().ListFlows(10, 0, "").Return([]BasicFlowDefinition{}, 100, nil)

	result, err := s.service.ListFlows(10, 0, "")

	s.Nil(err)
	s.NotNil(result)
	// Should only have next and last links (no first/prev on first page)
	s.Len(result.Links, 2)
}

func (s *FlowMgtServiceTestSuite) TestListFlows_PaginationLinksLastPage() {
	s.mockStore.EXPECT().ListFlows(10, 90, "").Return([]BasicFlowDefinition{}, 100, nil)

	result, err := s.service.ListFlows(10, 90, "")

	s.Nil(err)
	s.NotNil(result)
	// Should only have first and prev links (no next/last on last page)
	s.Len(result.Links, 2)
}

// CreateFlow tests

func (s *FlowMgtServiceTestSuite) TestCreateFlow_Success() {
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{Type: "start"},
			{Type: "action"},
			{Type: "end"},
		},
	}
	expectedFlow := &CompleteFlowDefinition{
		Handle:        "test-handle",
		Name:          "Test Flow",
		FlowType:      common.FlowTypeAuthentication,
		ActiveVersion: 1,
	}
	s.mockStore.EXPECT().IsFlowExistsByHandle("test-handle", common.FlowTypeAuthentication).Return(false, nil)
	s.mockStore.EXPECT().CreateFlow(mock.Anything, flowDef).Return(expectedFlow, nil)

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(err)
	s.NotNil(result)
	s.Equal("Test Flow", result.Name)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_ValidationError() {
	flowDef := &FlowDefinition{
		Handle:   "",
		Name:     "",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "end"}},
	}

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&ErrorMissingFlowHandle, err)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_InvalidHandleFormat_Uppercase() {
	flowDef := &FlowDefinition{
		Handle:   "Test-Handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&ErrorInvalidFlowHandleFormat, err)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_InvalidHandleFormat_Spaces() {
	flowDef := &FlowDefinition{
		Handle:   "test handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&ErrorInvalidFlowHandleFormat, err)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_InvalidHandleFormat_SpecialChars() {
	flowDef := &FlowDefinition{
		Handle:   "test@handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&ErrorInvalidFlowHandleFormat, err)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_InvalidHandleFormat_StartsWithDash() {
	flowDef := &FlowDefinition{
		Handle:   "-test-handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&ErrorInvalidFlowHandleFormat, err)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_InvalidHandleFormat_EndsWithUnderscore() {
	flowDef := &FlowDefinition{
		Handle:   "test_handle_",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&ErrorInvalidFlowHandleFormat, err)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_ValidHandleFormats() {
	testCases := []struct {
		name   string
		handle string
	}{
		{
			name:   "With dashes and numbers",
			handle: "test-handle-123",
		},
		{
			name:   "With underscores",
			handle: "test_handle_456",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			flowDef := &FlowDefinition{
				Handle:   tc.handle,
				Name:     "Test",
				FlowType: common.FlowTypeAuthentication,
				Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
			}

			flowID, _ := utils.GenerateUUIDv7()
			expectedFlow := &CompleteFlowDefinition{
				ID:            flowID,
				Handle:        flowDef.Handle,
				Name:          flowDef.Name,
				FlowType:      flowDef.FlowType,
				ActiveVersion: 1,
				Nodes:         flowDef.Nodes,
			}

			s.mockStore.EXPECT().IsFlowExistsByHandle(tc.handle, common.FlowTypeAuthentication).Return(false, nil)
			s.mockStore.EXPECT().CreateFlow(mock.Anything, flowDef).Return(expectedFlow, nil)

			result, err := s.service.CreateFlow(flowDef)

			s.Nil(err)
			s.NotNil(result)
			s.Equal(tc.handle, result.Handle)
		})
	}
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_InvalidFlowType() {
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Test",
		FlowType: "invalid",
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&ErrorInvalidFlowType, err)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_InsufficientNodes() {
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}},
	}

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(ErrorInvalidFlowData.Code, err.Code)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_OnlyStartAndEnd() {
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "end"}},
	}

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(ErrorInvalidFlowData.Code, err.Code)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_StoreError() {
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	s.mockStore.EXPECT().IsFlowExistsByHandle("test-handle", common.FlowTypeAuthentication).Return(false, nil)
	s.mockStore.EXPECT().CreateFlow(mock.Anything, flowDef).Return(nil, errors.New("db error"))

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_WithAutoInference() {
	// Enable auto-inference for this test
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			AutoInferRegistration: true,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
	defer config.ResetThunderRuntime()

	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Auth Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	expectedFlow := &CompleteFlowDefinition{
		Handle:        "test-handle",
		Name:          "Auth Flow",
		FlowType:      common.FlowTypeAuthentication,
		ActiveVersion: 1,
	}
	inferredRegFlow := &FlowDefinition{
		Handle:   "test-handle-reg",
		Name:     "Auth Flow - Registration",
		FlowType: common.FlowTypeRegistration,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}

	s.mockStore.EXPECT().IsFlowExistsByHandle("test-handle", common.FlowTypeAuthentication).Return(false, nil)
	s.mockStore.EXPECT().CreateFlow(mock.Anything, flowDef).Return(expectedFlow, nil)
	s.mockInference.EXPECT().InferRegistrationFlow(flowDef).Return(inferredRegFlow, nil)
	s.mockStore.EXPECT().CreateFlow(mock.Anything, inferredRegFlow).Return(nil, nil)

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(err)
	s.NotNil(result)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_AutoInferenceFailure() {
	// Enable auto-inference for this test
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			AutoInferRegistration: true,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
	defer config.ResetThunderRuntime()

	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Auth Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	expectedFlow := &CompleteFlowDefinition{
		Handle:        "test-handle",
		Name:          "Auth Flow",
		FlowType:      common.FlowTypeAuthentication,
		ActiveVersion: 1,
	}

	// Mock expectations in the correct order of execution
	s.mockStore.EXPECT().IsFlowExistsByHandle("test-handle", common.FlowTypeAuthentication).Return(false, nil)
	s.mockStore.EXPECT().CreateFlow(mock.Anything, flowDef).Return(expectedFlow, nil)
	s.mockInference.EXPECT().InferRegistrationFlow(flowDef).Return(nil, errors.New("inference error"))

	// Should still succeed even if inference fails
	result, err := s.service.CreateFlow(flowDef)

	s.Nil(err)
	s.NotNil(result)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_DuplicateHandle() {
	flowDef := &FlowDefinition{
		Handle:   "existing-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	s.mockStore.EXPECT().IsFlowExistsByHandle("existing-handle", common.FlowTypeAuthentication).Return(
		true, nil)

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&ErrorDuplicateFlowHandle, err)
}

func (s *FlowMgtServiceTestSuite) TestCreateFlow_DuplicateHandleCheckError() {
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	s.mockStore.EXPECT().IsFlowExistsByHandle("test-handle", common.FlowTypeAuthentication).Return(
		false, errors.New("db error"))

	result, err := s.service.CreateFlow(flowDef)

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

// GetFlow tests

func (s *FlowMgtServiceTestSuite) TestGetFlow_Success() {
	expectedFlow := &CompleteFlowDefinition{
		ID:     testFlowIDService,
		Handle: "test-handle",
		Name:   "Test",
	}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(expectedFlow, nil)

	result, err := s.service.GetFlow(testFlowIDService)

	s.Nil(err)
	s.Equal(expectedFlow, result)
}

func (s *FlowMgtServiceTestSuite) TestGetFlow_EmptyID() {
	result, err := s.service.GetFlow("")

	s.Nil(result)
	s.Equal(&ErrorMissingFlowID, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlow_NotFound() {
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(nil, errFlowNotFound)

	result, err := s.service.GetFlow(testFlowIDService)

	s.Nil(result)
	s.Equal(&ErrorFlowNotFound, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlow_StoreError() {
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(nil, errors.New("db error"))

	result, err := s.service.GetFlow(testFlowIDService)

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

// GetFlowByHandle tests

func (s *FlowMgtServiceTestSuite) TestGetFlowByHandle_Success() {
	expectedFlow := &CompleteFlowDefinition{
		ID:       testFlowIDService,
		Handle:   "test-auth-flow",
		Name:     "Test Auth Flow",
		FlowType: common.FlowTypeAuthentication,
	}
	s.mockStore.EXPECT().GetFlowByHandle("test-auth-flow", common.FlowTypeAuthentication).
		Return(expectedFlow, nil)

	result, err := s.service.GetFlowByHandle("test-auth-flow", common.FlowTypeAuthentication)

	s.Nil(err)
	s.Equal(expectedFlow, result)
	s.Equal("test-auth-flow", result.Handle)
	s.Equal(common.FlowTypeAuthentication, result.FlowType)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowByHandle_SuccessRegistrationFlow() {
	expectedFlow := &CompleteFlowDefinition{
		ID:       "flow-reg-id",
		Handle:   "test-reg-flow",
		Name:     "Test Registration Flow",
		FlowType: common.FlowTypeRegistration,
	}
	s.mockStore.EXPECT().GetFlowByHandle("test-reg-flow", common.FlowTypeRegistration).
		Return(expectedFlow, nil)

	result, err := s.service.GetFlowByHandle("test-reg-flow", common.FlowTypeRegistration)

	s.Nil(err)
	s.Equal(expectedFlow, result)
	s.Equal("test-reg-flow", result.Handle)
	s.Equal(common.FlowTypeRegistration, result.FlowType)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowByHandle_EmptyHandle() {
	result, err := s.service.GetFlowByHandle("", common.FlowTypeAuthentication)

	s.Nil(result)
	s.Equal(&ErrorMissingFlowHandle, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowByHandle_InvalidFlowType() {
	result, err := s.service.GetFlowByHandle("test-handle", "INVALID_TYPE")

	s.Nil(result)
	s.Equal(&ErrorInvalidFlowType, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowByHandle_EmptyFlowType() {
	result, err := s.service.GetFlowByHandle("test-handle", "")

	s.Nil(result)
	s.Equal(&ErrorInvalidFlowType, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowByHandle_NotFound() {
	s.mockStore.EXPECT().GetFlowByHandle("non-existent-handle", common.FlowTypeAuthentication).
		Return(nil, errFlowNotFound)

	result, err := s.service.GetFlowByHandle("non-existent-handle", common.FlowTypeAuthentication)

	s.Nil(result)
	s.Equal(&ErrorFlowNotFound, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowByHandle_StoreError() {
	s.mockStore.EXPECT().GetFlowByHandle("test-handle", common.FlowTypeAuthentication).
		Return(nil, errors.New("database connection error"))

	result, err := s.service.GetFlowByHandle("test-handle", common.FlowTypeAuthentication)

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

// UpdateFlow tests

func (s *FlowMgtServiceTestSuite) TestUpdateFlow_Success() {
	existingFlow := &CompleteFlowDefinition{
		ID:       testFlowIDService,
		Handle:   "test-handle",
		FlowType: common.FlowTypeAuthentication,
	}
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Updated",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	updatedFlow := &CompleteFlowDefinition{
		Handle:        "test-handle",
		Name:          "Updated",
		ActiveVersion: 2,
	}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(existingFlow, nil)
	s.mockStore.EXPECT().UpdateFlow(testFlowIDService, flowDef).Return(updatedFlow, nil)
	s.mockGraphBuilder.EXPECT().InvalidateCache(testFlowIDService)

	result, err := s.service.UpdateFlow(testFlowIDService, flowDef)

	s.Nil(err)
	s.Equal(updatedFlow, result)
}

func (s *FlowMgtServiceTestSuite) TestUpdateFlow_EmptyID() {
	flowDef := &FlowDefinition{Name: "Test", FlowType: common.FlowTypeAuthentication}

	result, err := s.service.UpdateFlow("", flowDef)

	s.Nil(result)
	s.Equal(&ErrorMissingFlowID, err)
}

func (s *FlowMgtServiceTestSuite) TestUpdateFlow_ValidationError() {
	flowDef := &FlowDefinition{Handle: "", Name: "", FlowType: common.FlowTypeAuthentication}

	result, err := s.service.UpdateFlow(testFlowIDService, flowDef)

	s.Nil(result)
	s.Equal(&ErrorMissingFlowHandle, err)
}

func (s *FlowMgtServiceTestSuite) TestUpdateFlow_FlowNotFound() {
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(nil, errFlowNotFound)

	result, err := s.service.UpdateFlow(testFlowIDService, flowDef)

	s.Nil(result)
	s.Equal(&ErrorFlowNotFound, err)
}

func (s *FlowMgtServiceTestSuite) TestUpdateFlow_CannotChangeFlowType() {
	existingFlow := &CompleteFlowDefinition{
		ID:       testFlowIDService,
		Handle:   "test-handle",
		FlowType: common.FlowTypeAuthentication,
	}
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Test",
		FlowType: common.FlowTypeRegistration,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(existingFlow, nil)

	result, err := s.service.UpdateFlow(testFlowIDService, flowDef)

	s.Nil(result)
	s.Equal(&ErrorCannotUpdateFlowType, err)
}

func (s *FlowMgtServiceTestSuite) TestUpdateFlow_CannotChangeHandle() {
	existingFlow := &CompleteFlowDefinition{
		ID:       testFlowIDService,
		Handle:   "original-handle",
		FlowType: common.FlowTypeAuthentication,
	}
	flowDef := &FlowDefinition{
		Handle:   "new-handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(existingFlow, nil)

	result, err := s.service.UpdateFlow(testFlowIDService, flowDef)

	s.Nil(result)
	s.Equal(&ErrorHandleUpdateNotAllowed, err)
}

func (s *FlowMgtServiceTestSuite) TestUpdateFlow_StoreError() {
	existingFlow := &CompleteFlowDefinition{
		ID:       testFlowIDService,
		Handle:   "test-handle",
		FlowType: common.FlowTypeAuthentication,
	}
	flowDef := &FlowDefinition{
		Handle:   "test-handle",
		Name:     "Test",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start"}, {Type: "action"}, {Type: "end"}},
	}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(existingFlow, nil)
	s.mockStore.EXPECT().UpdateFlow(testFlowIDService, flowDef).Return(nil, errors.New("db error"))

	result, err := s.service.UpdateFlow(testFlowIDService, flowDef)

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

// DeleteFlow tests

func (s *FlowMgtServiceTestSuite) TestDeleteFlow_Success() {
	existingFlow := &CompleteFlowDefinition{ID: testFlowIDService, Handle: "test-handle"}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(existingFlow, nil)
	s.mockStore.EXPECT().DeleteFlow(testFlowIDService).Return(nil)
	s.mockGraphBuilder.EXPECT().InvalidateCache(testFlowIDService)

	err := s.service.DeleteFlow(testFlowIDService)

	s.Nil(err)
}

func (s *FlowMgtServiceTestSuite) TestDeleteFlow_EmptyID() {
	err := s.service.DeleteFlow("")

	s.Equal(&ErrorMissingFlowID, err)
}

func (s *FlowMgtServiceTestSuite) TestDeleteFlow_NotFound() {
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(nil, errFlowNotFound)

	err := s.service.DeleteFlow(testFlowIDService)

	s.Nil(err)
}

func (s *FlowMgtServiceTestSuite) TestDeleteFlow_GetError() {
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(nil, errors.New("db error"))

	err := s.service.DeleteFlow(testFlowIDService)

	s.Equal(&serviceerror.InternalServerError, err)
}

func (s *FlowMgtServiceTestSuite) TestDeleteFlow_StoreError() {
	existingFlow := &CompleteFlowDefinition{ID: testFlowIDService, Handle: "test-handle"}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(existingFlow, nil)
	s.mockStore.EXPECT().DeleteFlow(testFlowIDService).Return(errors.New("db error"))

	err := s.service.DeleteFlow(testFlowIDService)

	s.Equal(&serviceerror.InternalServerError, err)
}

// ListFlowVersions tests

func (s *FlowMgtServiceTestSuite) TestListFlowVersions_Success() {
	existingFlow := &CompleteFlowDefinition{ID: testFlowIDService, Handle: "test-handle"}
	versions := []BasicFlowVersion{{Version: 1}, {Version: 2}}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(existingFlow, nil)
	s.mockStore.EXPECT().ListFlowVersions(testFlowIDService).Return(versions, nil)

	result, err := s.service.ListFlowVersions(testFlowIDService)

	s.Nil(err)
	s.NotNil(result)
	s.Equal(2, result.TotalVersions)
	s.Len(result.Versions, 2)
}

func (s *FlowMgtServiceTestSuite) TestListFlowVersions_EmptyID() {
	result, err := s.service.ListFlowVersions("")

	s.Nil(result)
	s.Equal(&ErrorMissingFlowID, err)
}

func (s *FlowMgtServiceTestSuite) TestListFlowVersions_FlowNotFound() {
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(nil, errFlowNotFound)

	result, err := s.service.ListFlowVersions(testFlowIDService)

	s.Nil(result)
	s.Equal(&ErrorFlowNotFound, err)
}

func (s *FlowMgtServiceTestSuite) TestListFlowVersions_StoreError() {
	existingFlow := &CompleteFlowDefinition{ID: testFlowIDService, Handle: "test-handle"}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(existingFlow, nil)
	s.mockStore.EXPECT().ListFlowVersions(testFlowIDService).Return(nil, errors.New("db error"))

	result, err := s.service.ListFlowVersions(testFlowIDService)

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

// GetFlowVersion tests

func (s *FlowMgtServiceTestSuite) TestGetFlowVersion_Success() {
	expectedVersion := &FlowVersion{Version: 1}
	s.mockStore.EXPECT().GetFlowVersion(testFlowIDService, 1).Return(expectedVersion, nil)

	result, err := s.service.GetFlowVersion(testFlowIDService, 1)

	s.Nil(err)
	s.Equal(expectedVersion, result)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowVersion_EmptyID() {
	result, err := s.service.GetFlowVersion("", 1)

	s.Nil(result)
	s.Equal(&ErrorMissingFlowID, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowVersion_InvalidVersion() {
	result, err := s.service.GetFlowVersion(testFlowIDService, 0)

	s.Nil(result)
	s.Equal(&ErrorInvalidVersion, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowVersion_FlowNotFound() {
	s.mockStore.EXPECT().GetFlowVersion(testFlowIDService, 1).Return(nil, errFlowNotFound)

	result, err := s.service.GetFlowVersion(testFlowIDService, 1)

	s.Nil(result)
	s.Equal(&ErrorFlowNotFound, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowVersion_VersionNotFound() {
	s.mockStore.EXPECT().GetFlowVersion(testFlowIDService, 1).Return(nil, errVersionNotFound)

	result, err := s.service.GetFlowVersion(testFlowIDService, 1)

	s.Nil(result)
	s.Equal(&ErrorVersionNotFound, err)
}

func (s *FlowMgtServiceTestSuite) TestGetFlowVersion_StoreError() {
	s.mockStore.EXPECT().GetFlowVersion(testFlowIDService, 1).Return(nil, errors.New("db error"))

	result, err := s.service.GetFlowVersion(testFlowIDService, 1)

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

// RestoreFlowVersion tests

func (s *FlowMgtServiceTestSuite) TestRestoreFlowVersion_Success() {
	version := &FlowVersion{Version: 1}
	restoredFlow := &CompleteFlowDefinition{ActiveVersion: 2}
	s.mockStore.EXPECT().GetFlowVersion(testFlowIDService, 1).Return(version, nil)
	s.mockStore.EXPECT().RestoreFlowVersion(testFlowIDService, 1).Return(restoredFlow, nil)
	s.mockGraphBuilder.EXPECT().InvalidateCache(testFlowIDService)

	result, err := s.service.RestoreFlowVersion(testFlowIDService, 1)

	s.Nil(err)
	s.Equal(restoredFlow, result)
}

func (s *FlowMgtServiceTestSuite) TestRestoreFlowVersion_EmptyID() {
	result, err := s.service.RestoreFlowVersion("", 1)

	s.Nil(result)
	s.Equal(&ErrorMissingFlowID, err)
}

func (s *FlowMgtServiceTestSuite) TestRestoreFlowVersion_InvalidVersion() {
	result, err := s.service.RestoreFlowVersion(testFlowIDService, 0)

	s.Nil(result)
	s.Equal(&ErrorInvalidVersion, err)
}

func (s *FlowMgtServiceTestSuite) TestRestoreFlowVersion_FlowNotFound() {
	s.mockStore.EXPECT().GetFlowVersion(testFlowIDService, 1).Return(nil, errFlowNotFound)

	result, err := s.service.RestoreFlowVersion(testFlowIDService, 1)

	s.Nil(result)
	s.Equal(&ErrorFlowNotFound, err)
}

func (s *FlowMgtServiceTestSuite) TestRestoreFlowVersion_VersionNotFound() {
	s.mockStore.EXPECT().GetFlowVersion(testFlowIDService, 1).Return(nil, errVersionNotFound)

	result, err := s.service.RestoreFlowVersion(testFlowIDService, 1)

	s.Nil(result)
	s.Equal(&ErrorVersionNotFound, err)
}

func (s *FlowMgtServiceTestSuite) TestRestoreFlowVersion_StoreError() {
	version := &FlowVersion{Version: 1}
	s.mockStore.EXPECT().GetFlowVersion(testFlowIDService, 1).Return(version, nil)
	s.mockStore.EXPECT().RestoreFlowVersion(testFlowIDService, 1).Return(nil, errors.New("db error"))

	result, err := s.service.RestoreFlowVersion(testFlowIDService, 1)

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

// GetGraph tests

func (s *FlowMgtServiceTestSuite) TestGetGraph_Success() {
	flow := &CompleteFlowDefinition{ID: testFlowIDService}
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(flow, nil)
	s.mockGraphBuilder.EXPECT().GetGraph(flow).Return(nil, nil)

	result, err := s.service.GetGraph(testFlowIDService)

	s.Nil(err)
	s.Nil(result)
}

func (s *FlowMgtServiceTestSuite) TestGetGraph_EmptyID() {
	result, err := s.service.GetGraph("")

	s.Nil(result)
	s.Equal(&ErrorMissingFlowID, err)
}

func (s *FlowMgtServiceTestSuite) TestGetGraph_FlowNotFound() {
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(nil, errFlowNotFound)

	result, err := s.service.GetGraph(testFlowIDService)

	s.Nil(result)
	s.Equal(&ErrorFlowNotFound, err)
}

func (s *FlowMgtServiceTestSuite) TestGetGraph_StoreError() {
	s.mockStore.EXPECT().GetFlowByID(testFlowIDService).Return(nil, errors.New("db error"))

	result, err := s.service.GetGraph(testFlowIDService)

	s.Nil(result)
	s.Equal(&serviceerror.InternalServerError, err)
}

// IsValidFlow tests

func (s *FlowMgtServiceTestSuite) TestIsValidFlow_Success() {
	s.mockStore.EXPECT().IsFlowExists(testFlowIDService).Return(true, nil)

	result := s.service.IsValidFlow(testFlowIDService)

	s.True(result)
}

func (s *FlowMgtServiceTestSuite) TestIsValidFlow_NotFound() {
	s.mockStore.EXPECT().IsFlowExists(testFlowIDService).Return(false, nil)

	result := s.service.IsValidFlow(testFlowIDService)

	s.False(result)
}

func (s *FlowMgtServiceTestSuite) TestIsValidFlow_EmptyID() {
	result := s.service.IsValidFlow("")

	s.False(result)
}

func (s *FlowMgtServiceTestSuite) TestIsValidFlow_StoreError() {
	s.mockStore.EXPECT().IsFlowExists(testFlowIDService).Return(false, errors.New("db error"))

	result := s.service.IsValidFlow(testFlowIDService)

	s.False(result)
}
