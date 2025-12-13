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

package mgt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
	flowsEndpoint = "/flows"
)

var (
	basicAuthFlow = FlowDefinition{
		Name:     "Basic Authentication Flow",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{
				ID:   "START",
				Type: "START",
			},
			{
				ID:   "basic_auth",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Name: "BasicAuthExecutor",
				},
				OnSuccess: "END",
				OnFailure: "END",
			},
			{
				ID:   "END",
				Type: "END",
			},
		},
	}

	conditionalAuthFlow = FlowDefinition{
		Name:     "Conditional Authentication Flow",
		FlowType: "AUTHENTICATION",
		Nodes: []NodeDefinition{
			{
				ID:   "START",
				Type: "START",
			},
			{
				ID:   "basic_auth",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Name: "BasicAuthExecutor",
				},
				OnSuccess: "ou_node",
				OnFailure: "END",
			},
			{
				ID:   "ou_node",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Name: "OUExecutor",
				},
				Condition: &ConditionDefinition{
					Key:    "{{ context.userEligibleForProvisioning }}",
					Value:  "true",
					OnSkip: "END",
				},
				OnSuccess: "END",
				OnFailure: "END",
			},
			{
				ID:   "END",
				Type: "END",
			},
		},
	}

	basicRegistrationFlow = FlowDefinition{
		Name:     "Basic Registration Flow",
		FlowType: "REGISTRATION",
		Nodes: []NodeDefinition{
			{
				ID:   "START",
				Type: "START",
			},
			{
				ID:   "user_type_resolver",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Name: "UserTypeResolver",
				},
				OnSuccess: "provisioning",
				OnFailure: "END",
			},
			{
				ID:   "provisioning",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Name: "ProvisioningExecutor",
				},
				OnSuccess: "END",
				OnFailure: "END",
			},
			{
				ID:   "END",
				Type: "END",
			},
		},
	}
)

type FlowMgtAPITestSuite struct {
	suite.Suite
	createdFlowIDs []string
}

func TestFlowMgtAPITestSuite(t *testing.T) {
	suite.Run(t, new(FlowMgtAPITestSuite))
}

func (suite *FlowMgtAPITestSuite) TearDownSuite() {
	for _, flowID := range suite.createdFlowIDs {
		suite.deleteFlow(flowID)
	}
}

func (suite *FlowMgtAPITestSuite) trackFlow(flowID string) {
	suite.createdFlowIDs = append(suite.createdFlowIDs, flowID)
}

func (suite *FlowMgtAPITestSuite) TestCreateFlow_Success() {
	testCases := []struct {
		name     string
		flowDef  FlowDefinition
		testFunc func(*CompleteFlowDefinition)
	}{
		{
			name:    "Create basic authentication flow",
			flowDef: basicAuthFlow,
			testFunc: func(response *CompleteFlowDefinition) {
				suite.NotEmpty(response.ID)
				suite.Equal(basicAuthFlow.Name, response.Name)
				suite.Equal(basicAuthFlow.FlowType, response.FlowType)
				suite.Equal(1, response.ActiveVersion)
				suite.Len(response.Nodes, len(basicAuthFlow.Nodes))
				suite.NotEmpty(response.CreatedAt)
				suite.NotEmpty(response.UpdatedAt)
				suite.trackFlow(response.ID)
			},
		},
		{
			name:    "Create conditional authentication flow",
			flowDef: conditionalAuthFlow,
			testFunc: func(response *CompleteFlowDefinition) {
				suite.NotEmpty(response.ID)
				suite.Equal(conditionalAuthFlow.Name, response.Name)
				suite.Equal(conditionalAuthFlow.FlowType, response.FlowType)
				suite.Equal(1, response.ActiveVersion)
				suite.Len(response.Nodes, 4)
				suite.trackFlow(response.ID)
			},
		},
		{
			name:    "Create registration flow",
			flowDef: basicRegistrationFlow,
			testFunc: func(response *CompleteFlowDefinition) {
				suite.NotEmpty(response.ID)
				suite.Equal(basicRegistrationFlow.Name, response.Name)
				suite.Equal(basicRegistrationFlow.FlowType, response.FlowType)
				suite.Equal(1, response.ActiveVersion)
				suite.Len(response.Nodes, len(basicRegistrationFlow.Nodes))
				suite.trackFlow(response.ID)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			response := suite.createFlow(tc.flowDef)
			tc.testFunc(response)
		})
	}
}

func (suite *FlowMgtAPITestSuite) TestCreateFlow_ValidationErrors() {
	testCases := []struct {
		name           string
		flowDef        FlowDefinition
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "Missing flow name",
			flowDef: FlowDefinition{
				Name:     "",
				FlowType: "AUTHENTICATION",
				Nodes:    basicAuthFlow.Nodes,
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "FLM-1010",
		},
		{
			name: "Invalid flow type",
			flowDef: FlowDefinition{
				Name:     "Test Flow",
				FlowType: "INVALID_TYPE",
				Nodes:    basicAuthFlow.Nodes,
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "FLM-1004",
		},
		{
			name: "Missing START node",
			flowDef: FlowDefinition{
				Name:     "Test Flow",
				FlowType: "AUTHENTICATION",
				Nodes: []NodeDefinition{
					{
						ID:   "END",
						Type: "END",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "FLM-1005",
		},
		{
			name: "Missing END node",
			flowDef: FlowDefinition{
				Name:     "Test Flow",
				FlowType: "AUTHENTICATION",
				Nodes: []NodeDefinition{
					{
						ID:   "START",
						Type: "START",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "FLM-1005",
		},
		{
			name: "Duplicate node IDs",
			flowDef: FlowDefinition{
				Name:     "Test Flow With Duplicate IDs",
				FlowType: "AUTHENTICATION",
				Nodes: []NodeDefinition{
					{
						ID:   "START",
						Type: "START",
					},
					{
						ID:   "node1",
						Type: "TASK_EXECUTION",
					},
					{
						ID:   "node1",
						Type: "TASK_EXECUTION",
					},
					{
						ID:   "END",
						Type: "END",
					},
				},
			},
			expectedStatus: http.StatusCreated,
			expectedCode:   "",
		},
		{
			name: "Invalid node reference",
			flowDef: FlowDefinition{
				Name:     "Test Flow With Invalid Ref",
				FlowType: "AUTHENTICATION",
				Nodes: []NodeDefinition{
					{
						ID:   "START",
						Type: "START",
					},
					{
						ID:        "node1",
						Type:      "TASK_EXECUTION",
						OnSuccess: "invalid_node",
					},
					{
						ID:   "END",
						Type: "END",
					},
				},
			},
			expectedStatus: http.StatusCreated,
			expectedCode:   "",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.expectedStatus == http.StatusCreated {
				createdFlow := suite.createFlow(tc.flowDef)
				suite.NotEmpty(createdFlow.ID)
				suite.deleteFlow(createdFlow.ID)
			} else {
				suite.createFlowExpectError(tc.flowDef, tc.expectedStatus, tc.expectedCode)
			}
		})
	}
}

func (suite *FlowMgtAPITestSuite) TestGetFlow_Success() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)
	response := suite.getFlow(createdFlow.ID)

	suite.Equal(createdFlow.ID, response.ID)
	suite.Equal(createdFlow.Name, response.Name)
	suite.Equal(createdFlow.FlowType, response.FlowType)
	suite.Equal(createdFlow.ActiveVersion, response.ActiveVersion)
	suite.Len(response.Nodes, len(createdFlow.Nodes))
}

func (suite *FlowMgtAPITestSuite) TestGetFlow_NotFound() {
	suite.getFlowExpectError("non-existent-id", http.StatusNotFound, "FLM-1003")
}

func (suite *FlowMgtAPITestSuite) TestListFlows_Success() {
	flow1 := suite.createFlow(basicAuthFlow)
	flow2 := suite.createFlow(conditionalAuthFlow)
	defer suite.deleteFlow(flow1.ID)
	defer suite.deleteFlow(flow2.ID)

	response := suite.listFlows(nil)

	suite.GreaterOrEqual(response.TotalResults, 2)
	suite.GreaterOrEqual(response.Count, 2)
	suite.GreaterOrEqual(len(response.Flows), 2)

	foundFlow1 := false
	foundFlow2 := false
	for _, flow := range response.Flows {
		if flow.ID == flow1.ID {
			foundFlow1 = true
			suite.Equal(flow1.Name, flow.Name)
			suite.Equal(flow1.FlowType, flow.FlowType)
		}
		if flow.ID == flow2.ID {
			foundFlow2 = true
			suite.Equal(flow2.Name, flow.Name)
			suite.Equal(flow2.FlowType, flow.FlowType)
		}
	}
	suite.True(foundFlow1)
	suite.True(foundFlow2)
}

func (suite *FlowMgtAPITestSuite) TestListFlows_WithPagination() {
	response := suite.listFlows(map[string]string{
		"limit":  "2",
		"offset": "0",
	})

	suite.LessOrEqual(response.Count, 2)
	suite.NotEmpty(response.Links)
}

func (suite *FlowMgtAPITestSuite) TestListFlows_FilterByFlowType() {
	authFlow := suite.createFlow(basicAuthFlow)
	regFlow := suite.createFlow(basicRegistrationFlow)
	defer suite.deleteFlow(authFlow.ID)
	defer suite.deleteFlow(regFlow.ID)

	response := suite.listFlows(map[string]string{
		"flowType": "AUTHENTICATION",
	})

	suite.GreaterOrEqual(response.TotalResults, 1)
	for _, flow := range response.Flows {
		suite.Equal("AUTHENTICATION", flow.FlowType)
	}
}

func (suite *FlowMgtAPITestSuite) TestUpdateFlow_Success() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)
	updatedFlow := FlowDefinition{
		Name:     "Updated Flow Name",
		FlowType: createdFlow.FlowType,
		Nodes: []NodeDefinition{
			{
				ID:   "START",
				Type: "START",
			},
			{
				ID:   "updated_executor",
				Type: "TASK_EXECUTION",
				Executor: &ExecutorDefinition{
					Name: "UpdatedExecutor",
				},
				OnSuccess: "END",
				OnFailure: "END",
			},
			{
				ID:   "END",
				Type: "END",
			},
		},
	}

	response := suite.updateFlow(createdFlow.ID, updatedFlow)

	suite.Equal(createdFlow.ID, response.ID)
	suite.Equal(updatedFlow.Name, response.Name)
	suite.Equal(updatedFlow.FlowType, response.FlowType)
	suite.Equal(2, response.ActiveVersion)
	suite.Len(response.Nodes, len(updatedFlow.Nodes))
}

func (suite *FlowMgtAPITestSuite) TestUpdateFlow_ChangeFlowType() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)
	updatedFlow := FlowDefinition{
		Name:     "Updated Flow",
		FlowType: "REGISTRATION",
		Nodes:    basicAuthFlow.Nodes,
	}

	suite.updateFlowExpectError(createdFlow.ID, updatedFlow, http.StatusBadRequest, "FLM-1011")
}

func (suite *FlowMgtAPITestSuite) TestUpdateFlow_NotFound() {
	updatedFlow := FlowDefinition{
		Name:     "Updated Flow",
		FlowType: "AUTHENTICATION",
		Nodes:    basicAuthFlow.Nodes,
	}

	suite.updateFlowExpectError("non-existent-id", updatedFlow, http.StatusNotFound, "FLM-1003")
}

func (suite *FlowMgtAPITestSuite) TestDeleteFlow_Success() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.deleteFlow(createdFlow.ID)
	suite.getFlowExpectError(createdFlow.ID, http.StatusNotFound, "FLM-1003")
}

func (suite *FlowMgtAPITestSuite) TestDeleteFlow_NotFound() {
	suite.deleteFlowExpectStatus("non-existent-id", http.StatusNoContent)
}

func (suite *FlowMgtAPITestSuite) TestListFlowVersions_Success() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)
	updatedFlow := FlowDefinition{
		Name:     "Updated Flow V2",
		FlowType: createdFlow.FlowType,
		Nodes:    basicAuthFlow.Nodes,
	}
	suite.updateFlow(createdFlow.ID, updatedFlow)
	updatedFlow.Name = "Updated Flow V3"
	suite.updateFlow(createdFlow.ID, updatedFlow)

	response := suite.listFlowVersions(createdFlow.ID)

	suite.Equal(3, response.TotalVersions)
	suite.Len(response.Versions, 3)

	activeVersionCount := 0
	for _, version := range response.Versions {
		suite.GreaterOrEqual(version.Version, 1)
		suite.LessOrEqual(version.Version, 3)
		suite.NotEmpty(version.CreatedAt)
		if version.IsActive {
			activeVersionCount++
			suite.Equal(3, version.Version)
		}
	}
	suite.Equal(1, activeVersionCount)
}

func (suite *FlowMgtAPITestSuite) TestListFlowVersions_VersionHistoryLimit() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)

	// Create multiple versions to exceed the limit (configured as 3)
	for i := 2; i <= 5; i++ {
		updatedFlow := FlowDefinition{
			Name:     fmt.Sprintf("Updated Flow V%d", i),
			FlowType: createdFlow.FlowType,
			Nodes:    basicAuthFlow.Nodes,
		}
		suite.updateFlow(createdFlow.ID, updatedFlow)
	}

	response := suite.listFlowVersions(createdFlow.ID)

	// Should only keep the most recent 3 versions (versions 3, 4, 5)
	suite.Equal(3, response.TotalVersions)
	suite.Len(response.Versions, 3)

	versions := make(map[int]bool)
	for _, version := range response.Versions {
		versions[version.Version] = true
		suite.GreaterOrEqual(version.Version, 3)
		suite.LessOrEqual(version.Version, 5)
	}
	suite.True(versions[3])
	suite.True(versions[4])
	suite.True(versions[5])

	// Verify version 5 is active
	activeVersionCount := 0
	for _, version := range response.Versions {
		if version.IsActive {
			activeVersionCount++
			suite.Equal(5, version.Version)
		}
	}
	suite.Equal(1, activeVersionCount)
}

func (suite *FlowMgtAPITestSuite) TestGetFlowVersion_Success() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)
	updatedFlow := FlowDefinition{
		Name:     "Updated Flow V2",
		FlowType: createdFlow.FlowType,
		Nodes:    basicAuthFlow.Nodes,
	}
	suite.updateFlow(createdFlow.ID, updatedFlow)

	response := suite.getFlowVersion(createdFlow.ID, 1)

	suite.Equal(createdFlow.ID, response.ID)
	suite.Equal("Updated Flow V2", response.Name)
	suite.Equal(1, response.Version)
	suite.False(response.IsActive)
	suite.Len(response.Nodes, len(basicAuthFlow.Nodes))
}

func (suite *FlowMgtAPITestSuite) TestGetFlowVersion_ActiveVersion() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)

	response := suite.getFlowVersion(createdFlow.ID, 1)

	suite.Equal(createdFlow.ID, response.ID)
	suite.Equal(1, response.Version)
	suite.True(response.IsActive)
}

func (suite *FlowMgtAPITestSuite) TestGetFlowVersion_NotFound() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)

	suite.getFlowVersionExpectError(createdFlow.ID, 999, http.StatusNotFound, "FLM-1008")
}

func (suite *FlowMgtAPITestSuite) TestRestoreFlowVersion_Success() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)
	updatedFlow := FlowDefinition{
		Name:     "Updated Flow V2",
		FlowType: createdFlow.FlowType,
		Nodes:    basicAuthFlow.Nodes,
	}
	suite.updateFlow(createdFlow.ID, updatedFlow)
	updatedFlow.Name = "Updated Flow V3"
	suite.updateFlow(createdFlow.ID, updatedFlow)

	response := suite.restoreFlowVersion(createdFlow.ID, 1)

	suite.Equal(createdFlow.ID, response.ID)
	suite.Equal(4, response.ActiveVersion)
	suite.Equal("Updated Flow V3", response.Name)

	versions := suite.listFlowVersions(createdFlow.ID)
	suite.Equal(3, versions.TotalVersions)
	for _, version := range versions.Versions {
		if version.Version == 4 {
			suite.True(version.IsActive)
		} else {
			suite.False(version.IsActive)
		}
	}
}

func (suite *FlowMgtAPITestSuite) TestRestoreFlowVersion_InvalidVersion() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)
	restoreReq := RestoreVersionRequest{Version: 0}

	suite.restoreFlowVersionExpectError(createdFlow.ID, restoreReq, http.StatusBadRequest, "FLM-1009")
}

func (suite *FlowMgtAPITestSuite) TestRestoreFlowVersion_VersionNotFound() {
	createdFlow := suite.createFlow(basicAuthFlow)
	suite.trackFlow(createdFlow.ID)
	restoreReq := RestoreVersionRequest{Version: 999}

	suite.restoreFlowVersionExpectError(createdFlow.ID, restoreReq, http.StatusNotFound, "FLM-1008")
}

func (suite *FlowMgtAPITestSuite) TestRestoreFlowVersion_FlowNotFound() {
	restoreReq := RestoreVersionRequest{Version: 1}

	suite.restoreFlowVersionExpectError("non-existent-id", restoreReq, http.StatusNotFound, "FLM-1008")
}

// Helper methods for API interactions

func (suite *FlowMgtAPITestSuite) createFlow(flowDef FlowDefinition) *CompleteFlowDefinition {
	body, _ := json.Marshal(flowDef)
	req, _ := http.NewRequest(http.MethodPost, testServerURL+flowsEndpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response CompleteFlowDefinition
	err = json.Unmarshal(bodyBytes, &response)
	suite.NoError(err)

	return &response
}

func (suite *FlowMgtAPITestSuite) createFlowExpectError(
	flowDef FlowDefinition, expectedStatus int, expectedCode string) {
	body, _ := json.Marshal(flowDef)
	req, _ := http.NewRequest(http.MethodPost, testServerURL+flowsEndpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(expectedStatus, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var errorResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResp)
	suite.NoError(err)
	suite.Equal(expectedCode, errorResp.Code)
}

func (suite *FlowMgtAPITestSuite) getFlow(flowID string) *CompleteFlowDefinition {
	req, _ := http.NewRequest(http.MethodGet, testServerURL+flowsEndpoint+"/"+flowID, nil)

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response CompleteFlowDefinition
	err = json.Unmarshal(bodyBytes, &response)
	suite.NoError(err)

	return &response
}

func (suite *FlowMgtAPITestSuite) getFlowExpectError(flowID string, expectedStatus int, expectedCode string) {
	req, _ := http.NewRequest(http.MethodGet, testServerURL+flowsEndpoint+"/"+flowID, nil)

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(expectedStatus, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var errorResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResp)
	suite.NoError(err)
	suite.Equal(expectedCode, errorResp.Code)
}

func (suite *FlowMgtAPITestSuite) listFlows(params map[string]string) *FlowListResponse {
	reqURL := testServerURL + flowsEndpoint
	if len(params) > 0 {
		query := ""
		for k, v := range params {
			if query != "" {
				query += "&"
			}
			query += fmt.Sprintf("%s=%s", k, v)
		}
		if query != "" {
			reqURL += "?" + query
		}
	}

	req, _ := http.NewRequest(http.MethodGet, reqURL, nil)

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response FlowListResponse
	err = json.Unmarshal(bodyBytes, &response)
	suite.NoError(err)

	return &response
}

func (suite *FlowMgtAPITestSuite) updateFlow(flowID string, flowDef FlowDefinition) *CompleteFlowDefinition {
	body, _ := json.Marshal(flowDef)
	req, _ := http.NewRequest(http.MethodPut, testServerURL+flowsEndpoint+"/"+flowID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response CompleteFlowDefinition
	err = json.Unmarshal(bodyBytes, &response)
	suite.NoError(err)

	return &response
}

func (suite *FlowMgtAPITestSuite) updateFlowExpectError(
	flowID string, flowDef FlowDefinition, expectedStatus int, expectedCode string) {
	body, _ := json.Marshal(flowDef)
	req, _ := http.NewRequest(http.MethodPut, testServerURL+flowsEndpoint+"/"+flowID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(expectedStatus, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var errorResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResp)
	suite.NoError(err)
	suite.Equal(expectedCode, errorResp.Code)
}

func (suite *FlowMgtAPITestSuite) deleteFlow(flowID string) {
	req, _ := http.NewRequest(http.MethodDelete, testServerURL+flowsEndpoint+"/"+flowID, nil)

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNoContent, resp.StatusCode)
}

func (suite *FlowMgtAPITestSuite) deleteFlowExpectStatus(flowID string, expectedStatus int) {
	req, _ := http.NewRequest(http.MethodDelete, testServerURL+flowsEndpoint+"/"+flowID, nil)

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(expectedStatus, resp.StatusCode)
}

func (suite *FlowMgtAPITestSuite) listFlowVersions(flowID string) *FlowVersionListResponse {
	req, _ := http.NewRequest(http.MethodGet, testServerURL+flowsEndpoint+"/"+flowID+"/versions", nil)

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response FlowVersionListResponse
	err = json.Unmarshal(bodyBytes, &response)
	suite.NoError(err)

	return &response
}

func (suite *FlowMgtAPITestSuite) getFlowVersion(flowID string, version int) *FlowVersion {
	reqURL := fmt.Sprintf("%s%s/%s/versions/%d", testServerURL, flowsEndpoint, flowID, version)
	req, _ := http.NewRequest(http.MethodGet, reqURL, nil)

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response FlowVersion
	err = json.Unmarshal(bodyBytes, &response)
	suite.NoError(err)

	return &response
}

func (suite *FlowMgtAPITestSuite) getFlowVersionExpectError(
	flowID string, version int, expectedStatus int, expectedCode string) {
	reqURL := fmt.Sprintf("%s%s/%s/versions/%d", testServerURL, flowsEndpoint, flowID, version)
	req, _ := http.NewRequest(http.MethodGet, reqURL, nil)

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(expectedStatus, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var errorResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResp)
	suite.NoError(err)
	suite.Equal(expectedCode, errorResp.Code)
}

func (suite *FlowMgtAPITestSuite) restoreFlowVersion(flowID string, version int) *CompleteFlowDefinition {
	restoreReq := RestoreVersionRequest{Version: version}
	body, _ := json.Marshal(restoreReq)
	req, _ := http.NewRequest(http.MethodPost, testServerURL+flowsEndpoint+"/"+flowID+"/restore",
		bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response CompleteFlowDefinition
	err = json.Unmarshal(bodyBytes, &response)
	suite.NoError(err)

	return &response
}

func (suite *FlowMgtAPITestSuite) restoreFlowVersionExpectError(
	flowID string, restoreReq RestoreVersionRequest, expectedStatus int, expectedCode string) {
	body, _ := json.Marshal(restoreReq)
	req, _ := http.NewRequest(http.MethodPost, testServerURL+flowsEndpoint+"/"+flowID+"/restore",
		bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := testutils.GetHTTPClient()
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(expectedStatus, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var errorResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResp)
	suite.NoError(err)
	suite.Equal(expectedCode, errorResp.Code)
}
