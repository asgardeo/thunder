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
	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/cachemock"
)

type CacheBackedFlowStoreTestSuite struct {
	suite.Suite
	mockStore     *flowStoreInterfaceMock
	flowByIDCache *cachemock.CacheInterfaceMock[*CompleteFlowDefinition]
	cachedStore   *cacheBackedFlowStore
	cacheData     map[string]*CompleteFlowDefinition
}

func TestCacheBackedFlowStoreTestSuite(t *testing.T) {
	suite.Run(t, new(CacheBackedFlowStoreTestSuite))
}

func (s *CacheBackedFlowStoreTestSuite) SetupTest() {
	s.mockStore = newFlowStoreInterfaceMock(s.T())
	s.cacheData = make(map[string]*CompleteFlowDefinition)

	s.flowByIDCache = cachemock.NewCacheInterfaceMock[*CompleteFlowDefinition](s.T())

	s.setupCacheMock()

	s.cachedStore = &cacheBackedFlowStore{
		flowByIDCache: s.flowByIDCache,
		store:         s.mockStore,
		logger:        log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CacheBackedFlowStore")),
	}
}

func (s *CacheBackedFlowStoreTestSuite) setupCacheMock() {
	s.flowByIDCache.EXPECT().Set(mock.Anything, mock.Anything).
		RunAndReturn(func(key cache.CacheKey, value *CompleteFlowDefinition) error {
			s.cacheData[key.Key] = value
			return nil
		}).Maybe()

	s.flowByIDCache.EXPECT().Get(mock.Anything).
		RunAndReturn(func(key cache.CacheKey) (*CompleteFlowDefinition, bool) {
			if val, ok := s.cacheData[key.Key]; ok {
				return val, true
			}
			return nil, false
		}).Maybe()

	s.flowByIDCache.EXPECT().Delete(mock.Anything).
		RunAndReturn(func(key cache.CacheKey) error {
			delete(s.cacheData, key.Key)
			return nil
		}).Maybe()

	s.flowByIDCache.EXPECT().Clear().
		RunAndReturn(func() error {
			for k := range s.cacheData {
				delete(s.cacheData, k)
			}
			return nil
		}).Maybe()

	s.flowByIDCache.EXPECT().GetName().Return("FlowByIDCache").Maybe()
	s.flowByIDCache.EXPECT().CleanupExpired().Maybe()
	s.flowByIDCache.EXPECT().IsEnabled().Return(true).Maybe()
}

func (s *CacheBackedFlowStoreTestSuite) createTestFlow() *CompleteFlowDefinition {
	return &CompleteFlowDefinition{
		ID:            "flow-1",
		Name:          "Test Flow",
		FlowType:      common.FlowTypeAuthentication,
		ActiveVersion: 1,
		Nodes: []NodeDefinition{
			{
				ID:   "node-1",
				Type: "basic-auth",
			},
		},
		CreatedAt: "2025-01-01T00:00:00Z",
		UpdatedAt: "2025-01-01T00:00:00Z",
	}
}

func (s *CacheBackedFlowStoreTestSuite) TestListFlows() {
	flows := []BasicFlowDefinition{
		{
			ID:            "flow-1",
			Name:          "Flow 1",
			FlowType:      common.FlowTypeAuthentication,
			ActiveVersion: 1,
		},
		{
			ID:            "flow-2",
			Name:          "Flow 2",
			FlowType:      common.FlowTypeRegistration,
			ActiveVersion: 1,
		},
	}

	s.mockStore.EXPECT().ListFlows(10, 0, "").Return(flows, 2, nil)

	result, count, err := s.cachedStore.ListFlows(10, 0, "")

	s.NoError(err)
	s.Len(result, 2)
	s.Equal(2, count)
	s.Equal("flow-1", result[0].ID)
}

func (s *CacheBackedFlowStoreTestSuite) TestListFlowsError() {
	s.mockStore.EXPECT().ListFlows(10, 0, "").Return(nil, 0, errors.New("list error"))

	result, count, err := s.cachedStore.ListFlows(10, 0, "")

	s.Error(err)
	s.Nil(result)
	s.Equal(0, count)
}

func (s *CacheBackedFlowStoreTestSuite) TestCreateFlowSuccess() {
	flowDef := &FlowDefinition{
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes: []NodeDefinition{
			{ID: "node-1", Type: "basic-auth"},
		},
	}

	expected := s.createTestFlow()
	s.mockStore.EXPECT().CreateFlow("flow-1", flowDef).Return(expected, nil)

	result, err := s.cachedStore.CreateFlow("flow-1", flowDef)

	s.NoError(err)
	s.NotNil(result)
	s.Equal("flow-1", result.ID)

	cached, ok := s.cacheData["flow-1"]
	s.True(ok)
	s.Equal("flow-1", cached.ID)
}

func (s *CacheBackedFlowStoreTestSuite) TestCreateFlowError() {
	flowDef := &FlowDefinition{
		Name:     "Test Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{ID: "node-1", Type: "basic-auth"}},
	}

	s.mockStore.EXPECT().CreateFlow("flow-1", flowDef).Return(nil, errors.New("create error"))

	result, err := s.cachedStore.CreateFlow("flow-1", flowDef)

	s.Error(err)
	s.Nil(result)

	_, ok := s.cacheData["flow-1"]
	s.False(ok)
}

func (s *CacheBackedFlowStoreTestSuite) TestGetFlowByIDFromCache() {
	expected := s.createTestFlow()
	s.cacheData["flow-1"] = expected

	result, err := s.cachedStore.GetFlowByID("flow-1")

	s.NoError(err)
	s.NotNil(result)
	s.Equal("flow-1", result.ID)
}

func (s *CacheBackedFlowStoreTestSuite) TestGetFlowByIDFromStoreAndCache() {
	expected := s.createTestFlow()
	s.mockStore.EXPECT().GetFlowByID("flow-1").Return(expected, nil)

	result, err := s.cachedStore.GetFlowByID("flow-1")

	s.NoError(err)
	s.NotNil(result)
	s.Equal("flow-1", result.ID)

	cached, ok := s.cacheData["flow-1"]
	s.True(ok)
	s.Equal("flow-1", cached.ID)
}

func (s *CacheBackedFlowStoreTestSuite) TestGetFlowByIDNotFound() {
	s.mockStore.EXPECT().GetFlowByID("nonexistent").Return(nil, errFlowNotFound)

	result, err := s.cachedStore.GetFlowByID("nonexistent")

	s.Error(err)
	s.Nil(result)

	_, ok := s.cacheData["nonexistent"]
	s.False(ok)
}

func (s *CacheBackedFlowStoreTestSuite) TestGetFlowByIDNilFlow() {
	s.mockStore.EXPECT().GetFlowByID("flow-1").Return(nil, nil)

	result, err := s.cachedStore.GetFlowByID("flow-1")

	s.NoError(err)
	s.Nil(result)

	_, ok := s.cacheData["flow-1"]
	s.False(ok)
}

func (s *CacheBackedFlowStoreTestSuite) TestUpdateFlowSuccess() {
	flowDef := &FlowDefinition{
		Name:     "Updated Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{ID: "node-1", Type: "basic-auth"}},
	}

	updated := s.createTestFlow()
	updated.Name = "Updated Flow"
	updated.ActiveVersion = 2

	s.mockStore.EXPECT().UpdateFlow("flow-1", flowDef).Return(updated, nil)

	result, err := s.cachedStore.UpdateFlow("flow-1", flowDef)

	s.NoError(err)
	s.NotNil(result)
	s.Equal("Updated Flow", result.Name)
	s.Equal(2, result.ActiveVersion)

	cached, ok := s.cacheData["flow-1"]
	s.True(ok)
	s.Equal("Updated Flow", cached.Name)
}

func (s *CacheBackedFlowStoreTestSuite) TestUpdateFlowError() {
	flowDef := &FlowDefinition{
		Name:     "Updated Flow",
		FlowType: common.FlowTypeAuthentication,
		Nodes:    []NodeDefinition{{ID: "node-1", Type: "basic-auth"}},
	}

	s.mockStore.EXPECT().UpdateFlow("flow-1", flowDef).Return(nil, errors.New("update error"))

	result, err := s.cachedStore.UpdateFlow("flow-1", flowDef)

	s.Error(err)
	s.Nil(result)
}

func (s *CacheBackedFlowStoreTestSuite) TestDeleteFlowFromCache() {
	flow := s.createTestFlow()
	s.cacheData["flow-1"] = flow

	s.mockStore.EXPECT().DeleteFlow("flow-1").Return(nil)

	err := s.cachedStore.DeleteFlow("flow-1")

	s.NoError(err)

	_, ok := s.cacheData["flow-1"]
	s.False(ok)
}

func (s *CacheBackedFlowStoreTestSuite) TestDeleteFlowFromStore() {
	flow := s.createTestFlow()

	s.mockStore.EXPECT().GetFlowByID("flow-1").Return(flow, nil)
	s.mockStore.EXPECT().DeleteFlow("flow-1").Return(nil)

	err := s.cachedStore.DeleteFlow("flow-1")

	s.NoError(err)

	_, ok := s.cacheData["flow-1"]
	s.False(ok)
}

func (s *CacheBackedFlowStoreTestSuite) TestDeleteFlowNotFound() {
	s.mockStore.EXPECT().GetFlowByID("nonexistent").Return(nil, errFlowNotFound)

	err := s.cachedStore.DeleteFlow("nonexistent")

	s.NoError(err)
}

func (s *CacheBackedFlowStoreTestSuite) TestDeleteFlowGetError() {
	s.mockStore.EXPECT().GetFlowByID("flow-1").Return(nil, errors.New("get error"))

	err := s.cachedStore.DeleteFlow("flow-1")

	s.Error(err)
	s.Contains(err.Error(), "get error")
}

func (s *CacheBackedFlowStoreTestSuite) TestDeleteFlowDeleteError() {
	flow := s.createTestFlow()
	s.cacheData["flow-1"] = flow

	s.mockStore.EXPECT().DeleteFlow("flow-1").Return(errors.New("delete error"))

	err := s.cachedStore.DeleteFlow("flow-1")

	s.Error(err)
	s.Contains(err.Error(), "delete error")
}

func (s *CacheBackedFlowStoreTestSuite) TestDeleteFlowNilFromStore() {
	s.mockStore.EXPECT().GetFlowByID("flow-1").Return(nil, nil)

	err := s.cachedStore.DeleteFlow("flow-1")

	s.NoError(err)
}

func (s *CacheBackedFlowStoreTestSuite) TestListFlowVersions() {
	versions := []BasicFlowVersion{
		{Version: 3, CreatedAt: "2025-01-03T00:00:00Z", IsActive: true},
		{Version: 2, CreatedAt: "2025-01-02T00:00:00Z", IsActive: false},
		{Version: 1, CreatedAt: "2025-01-01T00:00:00Z", IsActive: false},
	}

	s.mockStore.EXPECT().ListFlowVersions("flow-1").Return(versions, nil)

	result, err := s.cachedStore.ListFlowVersions("flow-1")

	s.NoError(err)
	s.Len(result, 3)
	s.Equal(3, result[0].Version)
	s.True(result[0].IsActive)
}

func (s *CacheBackedFlowStoreTestSuite) TestListFlowVersionsError() {
	s.mockStore.EXPECT().ListFlowVersions("flow-1").Return(nil, errors.New("list versions error"))

	result, err := s.cachedStore.ListFlowVersions("flow-1")

	s.Error(err)
	s.Nil(result)
}

func (s *CacheBackedFlowStoreTestSuite) TestGetFlowVersion() {
	version := &FlowVersion{
		ID:        "flow-1",
		Name:      "Test Flow",
		FlowType:  string(common.FlowTypeAuthentication),
		Version:   2,
		IsActive:  false,
		Nodes:     []NodeDefinition{{ID: "node-1", Type: "basic-auth"}},
		CreatedAt: "2025-01-02T00:00:00Z",
	}

	s.mockStore.EXPECT().GetFlowVersion("flow-1", 2).Return(version, nil)

	result, err := s.cachedStore.GetFlowVersion("flow-1", 2)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(2, result.Version)
	s.False(result.IsActive)
}

func (s *CacheBackedFlowStoreTestSuite) TestGetFlowVersionError() {
	s.mockStore.EXPECT().GetFlowVersion("flow-1", 999).Return(nil, errVersionNotFound)

	result, err := s.cachedStore.GetFlowVersion("flow-1", 999)

	s.Error(err)
	s.Nil(result)
}

func (s *CacheBackedFlowStoreTestSuite) TestRestoreFlowVersionSuccess() {
	restored := s.createTestFlow()
	restored.ActiveVersion = 4

	s.mockStore.EXPECT().RestoreFlowVersion("flow-1", 1).Return(restored, nil)

	result, err := s.cachedStore.RestoreFlowVersion("flow-1", 1)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(4, result.ActiveVersion)

	cached, ok := s.cacheData["flow-1"]
	s.True(ok)
	s.Equal(4, cached.ActiveVersion)
}

func (s *CacheBackedFlowStoreTestSuite) TestRestoreFlowVersionError() {
	s.mockStore.EXPECT().RestoreFlowVersion("flow-1", 1).Return(nil, errors.New("restore error"))

	result, err := s.cachedStore.RestoreFlowVersion("flow-1", 1)

	s.Error(err)
	s.Nil(result)

	_, ok := s.cacheData["flow-1"]
	s.False(ok)
}

func (s *CacheBackedFlowStoreTestSuite) TestCacheFlowNil() {
	s.cachedStore.cacheFlow(nil)

	s.Empty(s.cacheData)
}

func (s *CacheBackedFlowStoreTestSuite) TestCacheFlowEmptyID() {
	flow := s.createTestFlow()
	flow.ID = ""

	s.cachedStore.cacheFlow(flow)

	s.Empty(s.cacheData)
}

func (s *CacheBackedFlowStoreTestSuite) TestCacheFlowCacheError() {
	// Create a new cache mock just for this test to override the setupCacheMock expectations
	errorCache := cachemock.NewCacheInterfaceMock[*CompleteFlowDefinition](s.T())
	errorCache.EXPECT().Set(mock.Anything, mock.Anything).
		Return(errors.New("cache error")).Once()
	errorCache.EXPECT().GetName().Return("FlowByIDCache").Maybe()

	// Temporarily replace the cache
	originalCache := s.cachedStore.flowByIDCache
	s.cachedStore.flowByIDCache = errorCache

	flow := s.createTestFlow()
	s.cachedStore.cacheFlow(flow)

	// Restore original cache
	s.cachedStore.flowByIDCache = originalCache

	// Verify the flow was not cached in the original cache data
	_, found := s.cacheData[flow.ID]
	s.False(found)
}

func (s *CacheBackedFlowStoreTestSuite) TestInvalidateFlowCacheEmptyID() {
	s.cachedStore.invalidateFlowCache("")

	s.Empty(s.cacheData)
}

func (s *CacheBackedFlowStoreTestSuite) TestInvalidateFlowCacheError() {
	flow := s.createTestFlow()
	s.cacheData["flow-1"] = flow

	// Create a new cache mock just for this test to override the setupCacheMock expectations
	errorCache := cachemock.NewCacheInterfaceMock[*CompleteFlowDefinition](s.T())
	errorCache.EXPECT().Delete(mock.Anything).
		Return(errors.New("cache error")).Once()
	errorCache.EXPECT().GetName().Return("FlowByIDCache").Maybe()

	// Temporarily replace the cache
	originalCache := s.cachedStore.flowByIDCache
	s.cachedStore.flowByIDCache = errorCache

	s.cachedStore.invalidateFlowCache("flow-1")

	// Restore original cache
	s.cachedStore.flowByIDCache = originalCache

	// The flow should still be in the original cache data since we used error cache
	val, found := s.cacheData[flow.ID]
	s.True(found)
	s.Equal(flow, val)
}
