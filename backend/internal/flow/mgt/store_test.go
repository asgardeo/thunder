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
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/database/modelmock"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

type FlowStoreTestSuite struct {
	suite.Suite
	mockDBProvider *providermock.DBProviderInterfaceMock
	mockDBClient   *providermock.DBClientInterfaceMock
	store          *flowStore
}

func TestFlowStoreTestSuite(t *testing.T) {
	suite.Run(t, new(FlowStoreTestSuite))
}

func (s *FlowStoreTestSuite) SetupTest() {
	_ = config.InitializeThunderRuntime("test", &config.Config{
		Server: config.ServerConfig{Identifier: "test-deployment"},
		Flow:   config.FlowConfig{MaxVersionHistory: 5},
	})

	s.mockDBProvider = providermock.NewDBProviderInterfaceMock(s.T())
	s.mockDBClient = providermock.NewDBClientInterfaceMock(s.T())
	s.store = &flowStore{
		dbProvider:        s.mockDBProvider,
		deploymentID:      "test-deployment",
		maxVersionHistory: 5,
		logger:            log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowStore")),
	}
}

// ListFlows Tests

func (s *FlowStoreTestSuite) TestListFlowsDBClientError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(nil, errors.New("connection error"))

	flows, count, err := s.store.ListFlows(10, 0, "")

	s.Error(err)
	s.Contains(err.Error(), "failed to get database client")
	s.Equal(0, count)
	s.Nil(flows)
}

func (s *FlowStoreTestSuite) TestListFlowsCountQueryError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCountFlows, "test-deployment").
		Return(nil, errors.New("query error")).Once()

	flows, count, err := s.store.ListFlows(10, 0, "")

	s.Error(err)
	s.Contains(err.Error(), "failed to count flows")
	s.Equal(0, count)
	s.Nil(flows)
}

func (s *FlowStoreTestSuite) TestListFlowsQueryError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCountFlows, "test-deployment").
		Return([]map[string]interface{}{{colCount: int64(1)}}, nil).Once()
	s.mockDBClient.EXPECT().Query(queryListFlows, "test-deployment", 10, 0).
		Return(nil, errors.New("query error")).Once()

	flows, count, err := s.store.ListFlows(10, 0, "")

	s.Error(err)
	s.Contains(err.Error(), "failed to list flows")
	s.Equal(0, count)
	s.Nil(flows)
}

func (s *FlowStoreTestSuite) TestListFlowsSuccess() {
	flowsData := []map[string]interface{}{
		{
			colFlowID:        "flow-1",
			colHandle:        "handle-1",
			colName:          "Flow 1",
			colFlowType:      string(common.FlowTypeAuthentication),
			colActiveVersion: int64(1),
			colCreatedAt:     "2025-01-01T00:00:00Z",
			colUpdatedAt:     "2025-01-01T00:00:00Z",
		},
		{
			colFlowID:        "flow-2",
			colHandle:        "handle-2",
			colName:          "Flow 2",
			colFlowType:      string(common.FlowTypeRegistration),
			colActiveVersion: int64(2),
			colCreatedAt:     "2025-01-02T00:00:00Z",
			colUpdatedAt:     "2025-01-02T00:00:00Z",
		},
	}

	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCountFlows, "test-deployment").
		Return([]map[string]interface{}{{colCount: int64(2)}}, nil).Once()
	s.mockDBClient.EXPECT().Query(queryListFlows, "test-deployment", 10, 0).
		Return(flowsData, nil).Once()

	flows, count, err := s.store.ListFlows(10, 0, "")

	s.NoError(err)
	s.Equal(2, count)
	s.Len(flows, 2)
	s.Equal("flow-1", flows[0].ID)
	s.Equal("Flow 1", flows[0].Name)
	s.Equal("flow-2", flows[1].ID)
	s.Equal("Flow 2", flows[1].Name)
}

func (s *FlowStoreTestSuite) TestListFlowsWithTypeSuccess() {
	flowsData := []map[string]interface{}{
		{
			colFlowID:        "flow-1",
			colHandle:        "auth-handle",
			colName:          "Auth Flow",
			colFlowType:      string(common.FlowTypeAuthentication),
			colActiveVersion: int64(1),
			colCreatedAt:     "2025-01-01T00:00:00Z",
			colUpdatedAt:     "2025-01-01T00:00:00Z",
		},
	}

	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCountFlowsWithType, string(common.FlowTypeAuthentication), "test-deployment").
		Return([]map[string]interface{}{{colCount: int64(1)}}, nil).Once()
	s.mockDBClient.EXPECT().
		Query(queryListFlowsWithType, string(common.FlowTypeAuthentication), "test-deployment", 10, 0).
		Return(flowsData, nil).Once()

	flows, count, err := s.store.ListFlows(10, 0, string(common.FlowTypeAuthentication))

	s.NoError(err)
	s.Equal(1, count)
	s.Len(flows, 1)
	s.Equal("flow-1", flows[0].ID)
	s.Equal(common.FlowTypeAuthentication, flows[0].FlowType)
}

// GetFlowByID Tests

func (s *FlowStoreTestSuite) TestGetFlowByIDNotFound() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryGetFlow, "non-existent", "test-deployment").
		Return([]map[string]interface{}{}, nil).Once()

	flow, err := s.store.GetFlowByID("non-existent")

	s.Error(err)
	s.ErrorIs(err, errFlowNotFound)
	s.Nil(flow)
}

func (s *FlowStoreTestSuite) TestGetFlowByIDSuccess() {
	flowData := map[string]interface{}{
		colFlowID:        "flow-123",
		colHandle:        "test-handle",
		colName:          "Test Flow",
		colFlowType:      string(common.FlowTypeAuthentication),
		colActiveVersion: int64(1),
		colNodes:         `[{"id":"START","type":"START"}]`,
		colCreatedAt:     "2025-01-01T00:00:00Z",
		colUpdatedAt:     "2025-01-01T00:00:00Z",
	}

	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryGetFlow, "flow-123", "test-deployment").
		Return([]map[string]interface{}{flowData}, nil).Once()

	flow, err := s.store.GetFlowByID("flow-123")

	s.NoError(err)
	s.NotNil(flow)
	s.Equal("flow-123", flow.ID)
	s.Equal("test-handle", flow.Handle)
	s.Equal("Test Flow", flow.Name)
	s.Equal(common.FlowTypeAuthentication, flow.FlowType)
	s.Equal(1, flow.ActiveVersion)
	s.Len(flow.Nodes, 1)
}

func (s *FlowStoreTestSuite) TestGetFlowByIDDBError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(nil, errors.New("connection error"))

	flow, err := s.store.GetFlowByID("flow-1")

	s.Error(err)
	s.Contains(err.Error(), "failed to get database client")
	s.Nil(flow)
}

func (s *FlowStoreTestSuite) TestGetFlowByIDQueryError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryGetFlow, "flow-1", "test-deployment").
		Return(nil, errors.New("query error")).Once()

	flow, err := s.store.GetFlowByID("flow-1")

	s.Error(err)
	s.Contains(err.Error(), "failed to get flow")
	s.Nil(flow)
}

// DeleteFlow Tests

func (s *FlowStoreTestSuite) TestDeleteFlowSuccess() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Execute(queryDeleteFlow, "flow-1", "test-deployment").
		Return(int64(1), nil).Once()

	err := s.store.DeleteFlow("flow-1")

	s.NoError(err)
}

func (s *FlowStoreTestSuite) TestDeleteFlowDBError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(nil, errors.New("connection error"))

	err := s.store.DeleteFlow("flow-1")

	s.Error(err)
	s.Contains(err.Error(), "failed to get database client")
}

func (s *FlowStoreTestSuite) TestDeleteFlowExecuteError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Execute(queryDeleteFlow, "flow-1", "test-deployment").
		Return(int64(0), errors.New("delete failed")).Once()

	err := s.store.DeleteFlow("flow-1")

	s.Error(err)
	s.Contains(err.Error(), "failed to delete flow")
}

// IsFlowExists Tests

func (s *FlowStoreTestSuite) TestIsFlowExistsSuccess() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCheckFlowExistsByID, "flow-1", "test-deployment").
		Return([]map[string]interface{}{{"exists": 1}}, nil).Once()

	exists, err := s.store.IsFlowExists("flow-1")

	s.NoError(err)
	s.True(exists)
}

func (s *FlowStoreTestSuite) TestIsFlowExistsNotFound() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCheckFlowExistsByID, "non-existent", "test-deployment").
		Return([]map[string]interface{}{}, nil).Once()

	exists, err := s.store.IsFlowExists("non-existent")

	s.NoError(err)
	s.False(exists)
}

func (s *FlowStoreTestSuite) TestIsFlowExistsDBError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(nil, errors.New("connection error"))

	exists, err := s.store.IsFlowExists("flow-1")

	s.Error(err)
	s.Contains(err.Error(), "failed to get database client")
	s.False(exists)
}

func (s *FlowStoreTestSuite) TestIsFlowExistsQueryError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCheckFlowExistsByID, "flow-1", "test-deployment").
		Return(nil, errors.New("query error")).Once()

	exists, err := s.store.IsFlowExists("flow-1")

	s.Error(err)
	s.Contains(err.Error(), "failed to check flow existence")
	s.False(exists)
}

// GetFlowByHandle Tests

func (s *FlowStoreTestSuite) TestGetFlowByHandleSuccess() {
	flowData := map[string]interface{}{
		colFlowID:        "flow-123",
		colHandle:        "test-handle",
		colName:          "Test Flow",
		colFlowType:      string(common.FlowTypeAuthentication),
		colActiveVersion: int64(1),
		colNodes:         `[{"id":"START","type":"START"}]`,
		colCreatedAt:     time.Now().Format(time.RFC3339),
		colUpdatedAt:     time.Now().Format(time.RFC3339),
	}

	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryGetFlowByHandle, "test-handle",
		string(common.FlowTypeAuthentication), "test-deployment").Return(
		[]map[string]interface{}{flowData}, nil).Once()

	flow, err := s.store.GetFlowByHandle("test-handle", common.FlowTypeAuthentication)

	s.NoError(err)
	s.NotNil(flow)
	s.Equal("flow-123", flow.ID)
	s.Equal("test-handle", flow.Handle)
	s.Equal("Test Flow", flow.Name)
}

func (s *FlowStoreTestSuite) TestGetFlowByHandleNotFound() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryGetFlowByHandle, "non-existent",
		string(common.FlowTypeAuthentication), "test-deployment").Return(
		[]map[string]interface{}{}, nil).Once()

	flow, err := s.store.GetFlowByHandle("non-existent", common.FlowTypeAuthentication)

	s.Error(err)
	s.ErrorIs(err, errFlowNotFound)
	s.Nil(flow)
}

func (s *FlowStoreTestSuite) TestGetFlowByHandleDBError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(nil, errors.New("connection error"))

	flow, err := s.store.GetFlowByHandle("test-handle", common.FlowTypeAuthentication)

	s.Error(err)
	s.Contains(err.Error(), "failed to get database client")
	s.Nil(flow)
}

func (s *FlowStoreTestSuite) TestGetFlowByHandleQueryError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryGetFlowByHandle, "test-handle",
		string(common.FlowTypeAuthentication), "test-deployment").Return(
		nil, errors.New("query error")).Once()

	flow, err := s.store.GetFlowByHandle("test-handle", common.FlowTypeAuthentication)

	s.Error(err)
	s.Contains(err.Error(), "failed to get flow by handle")
	s.Nil(flow)
}

// IsFlowExistsByHandle Tests

func (s *FlowStoreTestSuite) TestIsFlowExistsByHandleSuccess() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCheckFlowExistsByHandle, "test-handle",
		string(common.FlowTypeAuthentication), "test-deployment").Return(
		[]map[string]interface{}{{"exists": 1}}, nil).Once()

	exists, err := s.store.IsFlowExistsByHandle("test-handle", common.FlowTypeAuthentication)

	s.NoError(err)
	s.True(exists)
}

func (s *FlowStoreTestSuite) TestIsFlowExistsByHandleNotFound() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCheckFlowExistsByHandle, "non-existent",
		string(common.FlowTypeAuthentication), "test-deployment").Return(
		[]map[string]interface{}{}, nil).Once()

	exists, err := s.store.IsFlowExistsByHandle("non-existent", common.FlowTypeAuthentication)

	s.NoError(err)
	s.False(exists)
}

func (s *FlowStoreTestSuite) TestIsFlowExistsByHandleDBError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(nil, errors.New("connection error"))

	exists, err := s.store.IsFlowExistsByHandle("test-handle", common.FlowTypeAuthentication)

	s.Error(err)
	s.Contains(err.Error(), "failed to get database client")
	s.False(exists)
}

func (s *FlowStoreTestSuite) TestIsFlowExistsByHandleQueryError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryCheckFlowExistsByHandle, "test-handle",
		string(common.FlowTypeAuthentication), "test-deployment").Return(
		nil, errors.New("query error")).Once()

	exists, err := s.store.IsFlowExistsByHandle("test-handle", common.FlowTypeAuthentication)

	s.Error(err)
	s.Contains(err.Error(), "failed to check flow existence by handle")
	s.False(exists)
}

// ListFlowVersions Tests

func (s *FlowStoreTestSuite) TestListFlowVersionsDBError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(nil, errors.New("connection error"))

	versions, err := s.store.ListFlowVersions("flow-1")

	s.Error(err)
	s.Contains(err.Error(), "failed to get database client")
	s.Nil(versions)
}

func (s *FlowStoreTestSuite) TestListFlowVersionsSuccess() {
	versionData := []map[string]interface{}{
		{
			colVersion:       int64(2),
			colCreatedAt:     "2025-01-02T00:00:00Z",
			colActiveVersion: int64(2),
		},
		{
			colVersion:       int64(1),
			colCreatedAt:     "2025-01-01T00:00:00Z",
			colActiveVersion: int64(2),
		},
	}

	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryListFlowVersions, "flow-123", s.store.deploymentID).
		Return(versionData, nil).Once()

	versions, err := s.store.ListFlowVersions("flow-123")

	s.NoError(err)
	s.Len(versions, 2)
	s.Equal(2, versions[0].Version)
	s.True(versions[0].IsActive)
	s.Equal(1, versions[1].Version)
	s.False(versions[1].IsActive)
}

// GetFlowVersion Tests

func (s *FlowStoreTestSuite) TestGetFlowVersionNotFound() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryGetFlowVersionWithMetadata, "flow-1", 99, "test-deployment").
		Return([]map[string]interface{}{}, nil).Once()

	version, err := s.store.GetFlowVersion("flow-1", 99)

	s.Error(err)
	s.ErrorIs(err, errVersionNotFound)
	s.Nil(version)
}

func (s *FlowStoreTestSuite) TestGetFlowVersionSuccess() {
	versionData := map[string]interface{}{
		colFlowID:        "flow-123",
		colHandle:        "test-handle",
		colName:          "Test Flow",
		colFlowType:      string(common.FlowTypeAuthentication),
		colVersion:       int64(2),
		colActiveVersion: int64(3),
		colNodes:         `[{"id":"node-1","type":"basic-auth"}]`,
		colCreatedAt:     "2025-01-02T00:00:00Z",
	}

	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().Query(queryGetFlowVersionWithMetadata, "flow-123", 2, "test-deployment").
		Return([]map[string]interface{}{versionData}, nil).Once()

	version, err := s.store.GetFlowVersion("flow-123", 2)

	s.NoError(err)
	s.NotNil(version)
	s.Equal("flow-123", version.ID)
	s.Equal(2, version.Version)
	s.False(version.IsActive) // Version 2, but active is 3
	s.Len(version.Nodes, 1)
}

func (s *FlowStoreTestSuite) TestGetFlowVersionDBError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(nil, errors.New("connection error"))

	version, err := s.store.GetFlowVersion("flow-1", 1)

	s.Error(err)
	s.Contains(err.Error(), "failed to get database client")
	s.Nil(version)
}

func (s *FlowStoreTestSuite) TestListFlowsWithTypeCountQueryError() {
	expectedError := errors.New("count query failed")
	s.mockDBClient.EXPECT().Query(queryCountFlowsWithType, "authentication", s.store.deploymentID).Return(
		nil, expectedError)
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)

	flows, count, err := s.store.ListFlows(10, 0, "authentication")

	s.Error(err)
	s.Nil(flows)
	s.Equal(0, count)
	s.Contains(err.Error(), "failed to count flows")
}

func (s *FlowStoreTestSuite) TestListFlowsWithTypeQueryError() {
	s.mockDBClient.EXPECT().Query(queryCountFlowsWithType, "authentication", s.store.deploymentID).Return(
		[]map[string]interface{}{{colCount: int64(5)}}, nil)
	expectedError := errors.New("list query failed")
	s.mockDBClient.EXPECT().Query(queryListFlowsWithType, "authentication", s.store.deploymentID, 10, 0).Return(
		nil, expectedError)
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)

	flows, count, err := s.store.ListFlows(10, 0, "authentication")

	s.Error(err)
	s.Nil(flows)
	s.Equal(0, count)
	s.Contains(err.Error(), "failed to list flows")
}

func (s *FlowStoreTestSuite) TestListFlowsBuildFlowError() {
	s.mockDBClient.EXPECT().Query(queryCountFlows, s.store.deploymentID).Return(
		[]map[string]interface{}{{colCount: int64(1)}}, nil)
	s.mockDBClient.EXPECT().Query(queryListFlows, s.store.deploymentID, 10, 0).Return(
		[]map[string]interface{}{
			{colFlowID: "flow-1"}, // Missing name field
		}, nil)
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)

	flows, count, err := s.store.ListFlows(10, 0, "")

	s.Error(err)
	s.Nil(flows)
	s.Equal(0, count)
	s.Contains(err.Error(), "failed to build flow")
}

func (s *FlowStoreTestSuite) TestListFlowVersionsQueryError() {
	expectedError := errors.New("query failed")
	s.mockDBClient.EXPECT().Query(queryListFlowVersions, "flow-123", s.store.deploymentID).Return(
		nil, expectedError)
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)

	versions, err := s.store.ListFlowVersions("flow-123")

	s.Error(err)
	s.Nil(versions)
	s.Contains(err.Error(), "failed to list")
}

func (s *FlowStoreTestSuite) TestListFlowVersionsBuildVersionError() {
	s.mockDBClient.EXPECT().Query(queryListFlowVersions, "flow-123", s.store.deploymentID).Return(
		[]map[string]interface{}{
			{colVersion: "invalid"}, // Invalid version type
		}, nil)
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)

	versions, err := s.store.ListFlowVersions("flow-123")

	s.Error(err)
	s.Empty(versions) // Returns empty slice on error, not nil
	s.Contains(err.Error(), "version field")
}

func (s *FlowStoreTestSuite) TestGetFlowVersionQueryError() {
	expectedError := errors.New("query failed")
	s.mockDBClient.EXPECT().Query(queryGetFlowVersionWithMetadata, "flow-123", 5, s.store.deploymentID).Return(
		nil, expectedError)
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)

	version, err := s.store.GetFlowVersion("flow-123", 5)

	s.Error(err)
	s.Nil(version)
	s.Contains(err.Error(), "failed to get")
}

func (s *FlowStoreTestSuite) TestGetFlowVersionBuildError() {
	s.mockDBClient.EXPECT().Query(queryGetFlowVersionWithMetadata, "flow-123", 5, s.store.deploymentID).Return(
		[]map[string]interface{}{
			{colFlowID: 123}, // Invalid type - should be string
		}, nil)
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)

	version, err := s.store.GetFlowVersion("flow-123", 5)

	s.Error(err)
	s.Nil(version)
	s.Contains(err.Error(), "flow_id field")
}

func (s *FlowStoreTestSuite) TestBuildBasicFlowDefinitionFromRowInvalidActiveVersion() {
	row := map[string]interface{}{
		colFlowID:        "flow-1",
		colHandle:        "test-handle",
		colName:          "Test Flow",
		colFlowType:      "authentication",
		colActiveVersion: "not-an-int", // Invalid type
		colCreatedAt:     "2024-01-01T00:00:00Z",
		colUpdatedAt:     "2024-01-02T00:00:00Z",
	}

	flow, err := s.store.buildBasicFlowDefinitionFromRow(row)

	s.Error(err)
	s.Equal(BasicFlowDefinition{}, flow)
	s.Contains(err.Error(), "active_version field is missing or invalid")
}

func (s *FlowStoreTestSuite) TestBuildCompleteFlowDefinitionFromRowInvalidActiveVersion() {
	row := map[string]interface{}{
		colFlowID:        "flow-1",
		colHandle:        "test-handle",
		colName:          "Test Flow",
		colFlowType:      "authentication",
		colActiveVersion: "invalid", // Invalid type
		colNodes:         "{}",
		colCreatedAt:     "2024-01-01T00:00:00Z",
		colUpdatedAt:     "2024-01-02T00:00:00Z",
	}

	flow, err := s.store.buildCompleteFlowDefinitionFromRow(row)

	s.Error(err)
	s.Nil(flow)
	s.Contains(err.Error(), "active_version field is missing or invalid")
}

func (s *FlowStoreTestSuite) TestBuildBasicFlowVersionFromRowInvalidVersion() {
	row := map[string]interface{}{
		colVersion:       "not-an-int", // Invalid type
		colCreatedAt:     "2024-01-01T00:00:00Z",
		colActiveVersion: int64(1),
	}

	version, err := s.store.buildBasicFlowVersionFromRow(row)

	s.Error(err)
	s.Equal(BasicFlowVersion{}, version)
	s.Contains(err.Error(), "version field is missing or invalid")
}

func (s *FlowStoreTestSuite) TestBuildBasicFlowVersionFromRowInvalidActiveVersion() {
	row := map[string]interface{}{
		colVersion:       int64(1),
		colCreatedAt:     "2024-01-01T00:00:00Z",
		colActiveVersion: "not-an-int", // Invalid type
	}

	version, err := s.store.buildBasicFlowVersionFromRow(row)

	s.Error(err)
	s.Equal(BasicFlowVersion{}, version)
	s.Contains(err.Error(), "active_version field is missing or invalid")
}

func (s *FlowStoreTestSuite) TestBuildFlowVersionFromRowInvalidVersion() {
	row := map[string]interface{}{
		colFlowID:    "flow-1",
		colHandle:    "test-handle",
		colName:      "Test",
		colFlowType:  "authentication",
		colVersion:   "not-an-int", // Invalid type
		colNodes:     "{}",
		colCreatedAt: "2024-01-01T00:00:00Z",
	}

	version, err := s.store.buildFlowVersionFromRow(row)

	s.Error(err)
	s.Nil(version)
	s.Contains(err.Error(), "version field is missing or invalid")
}

func (s *FlowStoreTestSuite) TestBuildFlowVersionFromRowInvalidFlowID() {
	row := map[string]interface{}{
		colFlowID:    123, // Invalid type - should be string
		colName:      "Test",
		colFlowType:  "authentication",
		colVersion:   int64(1),
		colNodes:     "{}",
		colCreatedAt: "2024-01-01T00:00:00Z",
	}

	version, err := s.store.buildFlowVersionFromRow(row)

	s.Error(err)
	s.Nil(version)
	s.Contains(err.Error(), "flow_id field is missing or invalid")
}

// Transaction-based Method Tests

func (s *FlowStoreTestSuite) TestCreateFlow_BeginTxError() {
	flowDef := &FlowDefinition{
		Handle:   "login-handle",
		Name:     "Login Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start", ID: "node1"}},
	}

	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(nil, errors.New("tx error"))

	result, err := s.store.CreateFlow("flow-1", flowDef)

	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to begin transaction")
}

func (s *FlowStoreTestSuite) TestCreateFlow_ExecError() {
	flowDef := &FlowDefinition{
		Handle:   "login-handle",
		Name:     "Login Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start", ID: "node1"}},
	}

	mockTx := modelmock.NewTxInterfaceMock(s.T())
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(mockTx, nil)
	mockTx.EXPECT().Exec(queryCreateFlow, "flow-1", "login-handle", "Login Flow", common.FlowTypeAuthentication,
		int64(1), s.store.deploymentID).Return(nil, errors.New("insert error"))
	mockTx.EXPECT().Rollback().Return(nil)

	result, err := s.store.CreateFlow("flow-1", flowDef)

	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to create flow")
}

func (s *FlowStoreTestSuite) TestUpdateFlow_BeginTxError() {
	flowDef := &FlowDefinition{
		Handle:   "updated-handle",
		Name:     "Updated Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{},
	}

	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(nil, errors.New("tx error"))

	result, err := s.store.UpdateFlow("flow-1", flowDef)

	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to begin transaction")
}

func (s *FlowStoreTestSuite) TestRestoreFlowVersion_BeginTxError() {
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(nil, errors.New("tx error"))

	result, err := s.store.RestoreFlowVersion("flow-1", 1)

	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to begin transaction")
}

// Helper Function Tests

func (s *FlowStoreTestSuite) TestParseCountResult() {
	tests := []struct {
		name          string
		results       []map[string]interface{}
		expectedCount int
		expectError   bool
	}{
		{"Parse int", []map[string]interface{}{{colCount: 5}}, 5, false},
		{"Parse int64", []map[string]interface{}{{colCount: int64(10)}}, 10, false},
		{"Parse float64", []map[string]interface{}{{colCount: float64(15)}}, 15, false},
		{"Empty results", []map[string]interface{}{}, 0, false},
		{"Missing count field", []map[string]interface{}{{"other": 5}}, 0, true},
		{"Invalid type", []map[string]interface{}{{colCount: "invalid"}}, 0, true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			count, err := s.store.parseCountResult(tt.results)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expectedCount, count)
			}
		})
	}
}

func (s *FlowStoreTestSuite) TestGetString() {
	tests := []struct {
		name        string
		row         map[string]interface{}
		key         string
		expected    string
		expectError bool
	}{
		{"Valid string", map[string]interface{}{"key": "value"}, "key", "value", false},
		{"Missing key", map[string]interface{}{}, "key", "", true},
		{"Invalid type", map[string]interface{}{"key": 123}, "key", "", true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			value, err := s.store.getString(tt.row, tt.key)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expected, value)
			}
		})
	}
}

func (s *FlowStoreTestSuite) TestGetTimestamp() {
	testTime := time.Date(2025, 12, 13, 10, 30, 0, 0, time.UTC)
	expectedTimeStr := testTime.Format(time.RFC3339)

	tests := []struct {
		name        string
		row         map[string]interface{}
		key         string
		expected    string
		expectError bool
	}{
		{"Valid string", map[string]interface{}{"key": "2025-12-13T10:30:00Z"}, "key", "2025-12-13T10:30:00Z", false},
		{"Valid time.Time", map[string]interface{}{"key": testTime}, "key", expectedTimeStr, false},
		{"Missing key", map[string]interface{}{}, "key", "", true},
		{"Invalid type", map[string]interface{}{"key": 123}, "key", "", true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			value, err := s.store.getTimestamp(tt.row, tt.key)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expected, value)
			}
		})
	}
}

func (s *FlowStoreTestSuite) TestGetInt64() {
	tests := []struct {
		name        string
		row         map[string]interface{}
		key         string
		expected    int64
		expectError bool
	}{
		{"Valid int64", map[string]interface{}{"key": int64(123)}, "key", int64(123), false},
		{"Missing key", map[string]interface{}{}, "key", 0, true},
		{"Invalid type", map[string]interface{}{"key": "string"}, "key", 0, true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			value, err := s.store.getInt64(tt.row, tt.key)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expected, value)
			}
		})
	}
}

func (s *FlowStoreTestSuite) TestBuildBasicFlowDefinitionFromRow() {
	validRow := map[string]interface{}{
		colFlowID:        "flow-1",
		colHandle:        "test-handle",
		colName:          "Test Flow",
		colFlowType:      string(common.FlowTypeAuthentication),
		colActiveVersion: int64(1),
		colCreatedAt:     "2025-01-01T00:00:00Z",
		colUpdatedAt:     "2025-01-01T00:00:00Z",
	}

	flow, err := s.store.buildBasicFlowDefinitionFromRow(validRow)

	s.NoError(err)
	s.Equal("flow-1", flow.ID)
	s.Equal("Test Flow", flow.Name)
	s.Equal(common.FlowTypeAuthentication, flow.FlowType)
	s.Equal(1, flow.ActiveVersion)
}

func (s *FlowStoreTestSuite) TestBuildBasicFlowDefinitionFromRowMissingField() {
	invalidRow := map[string]interface{}{
		colFlowID: "flow-1",
	}

	flow, err := s.store.buildBasicFlowDefinitionFromRow(invalidRow)

	s.Error(err)
	s.Empty(flow.ID)
}

func (s *FlowStoreTestSuite) TestBuildCompleteFlowDefinitionFromRow() {
	nodesJSON := `[{"id":"node-1","type":"basic-auth"}]`

	validRow := map[string]interface{}{
		colFlowID:        "flow-1",
		colHandle:        "test-handle",
		colName:          "Test Flow",
		colFlowType:      string(common.FlowTypeAuthentication),
		colActiveVersion: int64(1),
		colNodes:         nodesJSON,
		colCreatedAt:     "2025-01-01T00:00:00Z",
		colUpdatedAt:     "2025-01-01T00:00:00Z",
	}

	flow, err := s.store.buildCompleteFlowDefinitionFromRow(validRow)

	s.NoError(err)
	s.NotNil(flow)
	s.Equal("flow-1", flow.ID)
	s.Equal("Test Flow", flow.Name)
	s.Len(flow.Nodes, 1)
	s.Equal("node-1", flow.Nodes[0].ID)
}

func (s *FlowStoreTestSuite) TestBuildCompleteFlowDefinitionFromRowInvalidJSON() {
	invalidRow := map[string]interface{}{
		colFlowID:        "flow-1",
		colHandle:        "test-handle",
		colName:          "Test Flow",
		colFlowType:      string(common.FlowTypeAuthentication),
		colActiveVersion: int64(1),
		colNodes:         "invalid-json",
		colCreatedAt:     "2025-01-01T00:00:00Z",
		colUpdatedAt:     "2025-01-01T00:00:00Z",
	}

	flow, err := s.store.buildCompleteFlowDefinitionFromRow(invalidRow)

	s.Error(err)
	s.Nil(flow)
	s.Contains(err.Error(), "failed to unmarshal nodes")
}

func (s *FlowStoreTestSuite) TestBuildBasicFlowVersionFromRow() {
	validRow := map[string]interface{}{
		colVersion:       int64(2),
		colCreatedAt:     "2025-01-02T00:00:00Z",
		colActiveVersion: int64(3),
	}

	version, err := s.store.buildBasicFlowVersionFromRow(validRow)

	s.NoError(err)
	s.Equal(2, version.Version)
	s.Equal("2025-01-02T00:00:00Z", version.CreatedAt)
	s.False(version.IsActive)
}

func (s *FlowStoreTestSuite) TestBuildFlowVersionFromRow() {
	nodesJSON := `[{"id":"node-1","type":"basic-auth"}]`

	validRow := map[string]interface{}{
		colFlowID:        "flow-1",
		colHandle:        "test-handle",
		colName:          "Test Flow",
		colFlowType:      string(common.FlowTypeAuthentication),
		colVersion:       int64(2),
		colActiveVersion: int64(2),
		colNodes:         nodesJSON,
		colCreatedAt:     "2025-01-02T00:00:00Z",
	}

	version, err := s.store.buildFlowVersionFromRow(validRow)

	s.NoError(err)
	s.NotNil(version)
	s.Equal("flow-1", version.ID)
	s.Equal(2, version.Version)
	s.True(version.IsActive)
	s.Len(version.Nodes, 1)
}

func (s *FlowStoreTestSuite) TestGetConfigDBClientError() {
	mockProvider := providermock.NewDBProviderInterfaceMock(s.T())
	mockProvider.EXPECT().GetConfigDBClient().Return(nil, errors.New("database connection failed"))

	s.store.dbProvider = mockProvider

	_, err := s.store.getConfigDBClient()

	s.Error(err)
	s.Contains(err.Error(), "failed to get database client")
}

func (s *FlowStoreTestSuite) TestCreateFlow_InsertFlowVersionError() {
	flowDef := &FlowDefinition{
		Handle:   "login-handle",
		Name:     "Login Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{Type: "start", ID: "node1"}},
	}

	nodesJSON := `[{"id":"node1","type":"start"}]`
	mockTx := modelmock.NewTxInterfaceMock(s.T())
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(mockTx, nil)
	mockTx.EXPECT().Exec(queryCreateFlow, "flow-1", "login-handle", "Login Flow", common.FlowTypeAuthentication,
		int64(1), s.store.deploymentID).Return(nil, nil)
	mockTx.EXPECT().Exec(queryInsertFlowVersion, "flow-1", 1, nodesJSON, s.store.deploymentID).
		Return(nil, errors.New("version insert error"))
	mockTx.EXPECT().Rollback().Return(nil)

	result, err := s.store.CreateFlow("flow-1", flowDef)

	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to create flow version")
}

// UpdateFlow - Flow Not Found (store.go line 222-223): scanFlowMetadata returns err on empty rows
func (s *FlowStoreTestSuite) TestUpdateFlow_FlowNotFound() {
	flowDef := &FlowDefinition{
		Handle:   "updated-handle",
		Name:     "Updated Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{},
	}

	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	flowCols := []string{
		"flow_id", "handle", "name", "flow_type", "active_version", "nodes", "created_at", "updated_at"}
	emptyRows := mock.NewRows(flowCols)
	mock.ExpectQuery("SELECT").WillReturnRows(emptyRows)
	flowSQLRows, _ := db.Query("SELECT")

	mockTx := modelmock.NewTxInterfaceMock(s.T())
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(mockTx, nil)
	mockTx.EXPECT().Query(queryGetFlow, "flow-1", s.store.deploymentID).Return(flowSQLRows, nil)
	mockTx.EXPECT().Rollback().Return(nil)

	result, err := s.store.UpdateFlow("flow-1", flowDef)

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, errFlowNotFound)
}

func (s *FlowStoreTestSuite) TestUpdateFlow_PushToVersionStackError() {
	flowDef := &FlowDefinition{
		Handle:   "updated-handle",
		Name:     "Updated Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{},
	}

	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	flowCols := []string{
		"flow_id", "handle", "name", "flow_type", "active_version", "nodes", "created_at", "updated_at"}
	flowRows := mock.NewRows(flowCols).
		AddRow("flow-1", "updated-handle", "Updated Flow", "authentication",
			int64(3), "[]", "2025-01-01T00:00:00Z", "2025-01-01T00:00:00Z")
	mock.ExpectQuery("SELECT").WillReturnRows(flowRows)
	flowSQLRows, _ := db.Query("SELECT")

	mockTx := modelmock.NewTxInterfaceMock(s.T())
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(mockTx, nil)

	mockTx.EXPECT().Query(queryGetFlow, "flow-1", s.store.deploymentID).Return(flowSQLRows, nil)

	mockTx.EXPECT().Exec(queryInsertFlowVersion, "flow-1", 4, "[]", s.store.deploymentID).
		Return(nil, errors.New("insert version error"))
	mockTx.EXPECT().Rollback().Return(nil)

	result, err := s.store.UpdateFlow("flow-1", flowDef)

	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to insert flow version")
}

func (s *FlowStoreTestSuite) TestRestoreFlowVersion_FlowNotFound() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	flowCols := []string{
		"flow_id", "handle", "name", "flow_type", "active_version", "nodes", "created_at", "updated_at"}
	emptyRows := mock.NewRows(flowCols)
	mock.ExpectQuery("SELECT").WillReturnRows(emptyRows)
	flowSQLRows, _ := db.Query("SELECT")

	mockTx := modelmock.NewTxInterfaceMock(s.T())
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(mockTx, nil)
	mockTx.EXPECT().Query(queryGetFlow, "flow-1", s.store.deploymentID).Return(flowSQLRows, nil)
	mockTx.EXPECT().Rollback().Return(nil)

	result, err := s.store.RestoreFlowVersion("flow-1", 1)

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, errFlowNotFound)
}

func (s *FlowStoreTestSuite) TestRestoreFlowVersion_GetVersionQueryError() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	flowCols := []string{
		"flow_id", "handle", "name", "flow_type", "active_version", "nodes", "created_at", "updated_at"}
	flowRows := mock.NewRows(flowCols).
		AddRow("flow-2", "some-handle", "Some Flow", "authentication",
			int64(2), "[]", "2025-01-01T00:00:00Z", "2025-01-01T00:00:00Z")
	mock.ExpectQuery("SELECT").WillReturnRows(flowRows)
	flowSQLRows, _ := db.Query("SELECT")

	mockTx := modelmock.NewTxInterfaceMock(s.T())
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(mockTx, nil)

	mockTx.EXPECT().Query(queryGetFlow, "flow-2", s.store.deploymentID).Return(flowSQLRows, nil)
	mockTx.EXPECT().Query(queryGetFlowVersion, "flow-2", 2, s.store.deploymentID).
		Return(nil, errors.New("version query error"))
	mockTx.EXPECT().Rollback().Return(nil)

	result, err := s.store.RestoreFlowVersion("flow-2", 2)

	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to get version to restore")
}

func (s *FlowStoreTestSuite) TestRestoreFlowVersion_PushToVersionStackError() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	flowCols := []string{
		"flow_id", "handle", "name", "flow_type", "active_version", "nodes", "created_at", "updated_at"}
	flowRows := mock.NewRows(flowCols).
		AddRow("flow-3", "some-handle", "Some Flow", "authentication",
			int64(1), "[]", "2025-01-01T00:00:00Z", "2025-01-01T00:00:00Z")
	mock.ExpectQuery("SELECT flow").WillReturnRows(flowRows)
	flowSQLRows, _ := db.Query("SELECT flow")

	db2, mock2, err2 := sqlmock.New()
	s.Require().NoError(err2)
	versionRows := mock2.NewRows([]string{"version", "nodes", "created_at"}).
		AddRow(int64(1), `[]`, "2025-01-01T00:00:00Z")
	mock2.ExpectQuery("SELECT").WillReturnRows(versionRows)
	versionSQLRows, _ := db2.Query("SELECT")

	mockTx := modelmock.NewTxInterfaceMock(s.T())
	s.mockDBProvider.EXPECT().GetConfigDBClient().Return(s.mockDBClient, nil)
	s.mockDBClient.EXPECT().BeginTx().Return(mockTx, nil)
	mockTx.EXPECT().Query(queryGetFlow, "flow-3", s.store.deploymentID).Return(flowSQLRows, nil)
	mockTx.EXPECT().Query(queryGetFlowVersion, "flow-3", 1, s.store.deploymentID).Return(versionSQLRows, nil)
	mockTx.EXPECT().Exec(queryInsertFlowVersion, "flow-3", 2, "[]", s.store.deploymentID).
		Return(nil, errors.New("insert error"))
	mockTx.EXPECT().Rollback().Return(nil)

	result, err := s.store.RestoreFlowVersion("flow-3", 1)

	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to insert flow version")
}

func (s *FlowStoreTestSuite) TestPushToVersionStack_CountVersionsQueryError() {
	mockTx := modelmock.NewTxInterfaceMock(s.T())

	mockTx.EXPECT().Exec(queryInsertFlowVersion, "flow-1", 2, `[]`, s.store.deploymentID).
		Return(nil, nil)
	mockTx.EXPECT().Query(queryCountFlowVersions, "flow-1", s.store.deploymentID).
		Return(nil, errors.New("count query error"))

	err := s.store.pushToVersionStack(mockTx, "flow-1", 2, `[]`)

	s.Error(err)
	s.Contains(err.Error(), "failed to count versions")
}

func (s *FlowStoreTestSuite) TestPushToVersionStack_DeleteOldestVersionError() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	countRows := mock.NewRows([]string{"count"}).AddRow(int64(6))
	mock.ExpectQuery("SELECT").WillReturnRows(countRows)
	countSQLRows, _ := db.Query("SELECT")

	mockTx := modelmock.NewTxInterfaceMock(s.T())
	mockTx.EXPECT().Exec(queryInsertFlowVersion, "flow-1", 2, `[]`, s.store.deploymentID).
		Return(nil, nil)
	mockTx.EXPECT().Query(queryCountFlowVersions, "flow-1", s.store.deploymentID).Return(countSQLRows, nil)
	mockTx.EXPECT().Exec(queryDeleteOldestVersion, "flow-1", s.store.deploymentID).
		Return(nil, errors.New("delete error"))

	err = s.store.pushToVersionStack(mockTx, "flow-1", 2, `[]`)

	s.Error(err)
	s.Contains(err.Error(), "failed to delete oldest version")
}

func (s *FlowStoreTestSuite) TestPushToVersionStack_InsertVersionError() {
	mockTx := modelmock.NewTxInterfaceMock(s.T())

	mockTx.EXPECT().Exec(queryInsertFlowVersion, "flow-1", 2, `[]`, s.store.deploymentID).
		Return(nil, errors.New("insert error"))

	err := s.store.pushToVersionStack(mockTx, "flow-1", 2, `[]`)

	s.Error(err)
	s.Contains(err.Error(), "failed to insert flow version")
}

func (s *FlowStoreTestSuite) TestGetMaxVersionHistory() {
	tests := []struct {
		name     string
		config   int
		expected int
	}{
		{"Default value", 0, defaultVersionHistory},
		{"Valid value", 20, 20},
		{"Exceeds max", 200, maxAllowedVersionHistory},
		{"Negative value", -5, defaultVersionHistory},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Reset before reinitializing config for each test case
			config.ResetThunderRuntime()
			err := config.InitializeThunderRuntime("test", &config.Config{
				Server: config.ServerConfig{Identifier: "test-deployment"},
				Flow:   config.FlowConfig{MaxVersionHistory: tt.config},
			})
			s.NoError(err)

			result := getMaxVersionHistory()

			s.Equal(tt.expected, result)
		})
	}
}
