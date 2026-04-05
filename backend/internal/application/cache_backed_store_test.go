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

package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/tests/mocks/cachemock"
)

// Test suite
type CacheBackedStoreTestSuite struct {
	suite.Suite
	mockStore     *applicationStoreInterfaceMock
	appByIDCache  *cachemock.CacheInterfaceMock[*applicationConfigDAO]
	oauthAppCache *cachemock.CacheInterfaceMock[*oauthConfigDAO]
	cachedStore   *cachedBackedApplicationStore
	// Helper maps to track cached values for verification
	appByIDData  map[string]*applicationConfigDAO
	oauthAppData map[string]*oauthConfigDAO
}

func TestCacheBackedStoreTestSuite(t *testing.T) {
	suite.Run(t, new(CacheBackedStoreTestSuite))
}

func (suite *CacheBackedStoreTestSuite) SetupTest() {
	suite.mockStore = newApplicationStoreInterfaceMock(suite.T())
	suite.appByIDData = make(map[string]*applicationConfigDAO)
	suite.oauthAppData = make(map[string]*oauthConfigDAO)

	suite.appByIDCache = cachemock.NewCacheInterfaceMock[*applicationConfigDAO](suite.T())
	suite.oauthAppCache = cachemock.NewCacheInterfaceMock[*oauthConfigDAO](suite.T())

	setupCacheMockDAO(suite.appByIDCache, suite.appByIDData)
	setupOAuthCacheMockDAO(suite.oauthAppCache, suite.oauthAppData)

	suite.appByIDCache.EXPECT().IsEnabled().Return(true).Maybe()
	suite.oauthAppCache.EXPECT().IsEnabled().Return(true).Maybe()

	suite.cachedStore = &cachedBackedApplicationStore{
		AppByIDCache:  suite.appByIDCache,
		OAuthAppCache: suite.oauthAppCache,
		Store:         suite.mockStore,
	}
}

func setupGenericCacheMock[T any](
	mockCache *cachemock.CacheInterfaceMock[T],
	data map[string]T,
	cacheName string,
) {
	mockCache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, key cache.CacheKey, value T) error {
			data[key.Key] = value
			return nil
		}).Maybe()

	mockCache.EXPECT().Get(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, key cache.CacheKey) (T, bool) {
			val, ok := data[key.Key]
			return val, ok
		}).Maybe()

	mockCache.EXPECT().Delete(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, key cache.CacheKey) error {
			delete(data, key.Key)
			return nil
		}).Maybe()

	mockCache.EXPECT().Clear(mock.Anything).
		RunAndReturn(func(_ context.Context) error {
			for k := range data {
				delete(data, k)
			}
			return nil
		}).Maybe()

	mockCache.EXPECT().GetName().Return(cacheName).Maybe()
	mockCache.EXPECT().CleanupExpired().Maybe()
}

func setupCacheMockDAO(
	mockCache *cachemock.CacheInterfaceMock[*applicationConfigDAO],
	data map[string]*applicationConfigDAO,
) {
	setupGenericCacheMock(mockCache, data, "mockAppByIDCache")
}

func setupOAuthCacheMockDAO(
	mockCache *cachemock.CacheInterfaceMock[*oauthConfigDAO],
	data map[string]*oauthConfigDAO,
) {
	setupGenericCacheMock(mockCache, data, "mockOAuthAppCache")
}

func (suite *CacheBackedStoreTestSuite) createTestApp() *applicationConfigDAO {
	return &applicationConfigDAO{
		ID:                        "test-app-id",
		AuthFlowID:                "auth-flow-1",
		RegistrationFlowID:        "reg-flow-1",
		IsRegistrationFlowEnabled: true,
	}
}

func (suite *CacheBackedStoreTestSuite) createTestOAuthConfig() *oauthConfigDAO {
	return &oauthConfigDAO{
		EntityID: "test-entity-id",
		OAuthConfig: &oAuthConfig{
			RedirectURIs: []string{"https://example.com/callback"},
			GrantTypes:   []string{"authorization_code"},
		},
	}
}

// TestNewCachedBackedApplicationStore verifies store initialization.
func (suite *CacheBackedStoreTestSuite) TestNewCachedBackedApplicationStore() {
	suite.NotNil(suite.cachedStore)
	suite.IsType(&cachedBackedApplicationStore{}, suite.cachedStore)
	suite.NotNil(suite.cachedStore.AppByIDCache)
	suite.NotNil(suite.cachedStore.OAuthAppCache)
	suite.NotNil(suite.cachedStore.Store)
}

// CreateApplication tests

func (suite *CacheBackedStoreTestSuite) TestCreateApplication_Success() {
	app := suite.createTestApp()

	suite.mockStore.On("CreateApplication", mock.Anything, *app).Return(nil).Once()

	err := suite.cachedStore.CreateApplication(context.Background(), *app)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify application is cached by ID
	cachedByID, ok := suite.appByIDCache.Get(context.Background(), cache.CacheKey{Key: app.ID})
	suite.True(ok)
	suite.Equal(app.ID, cachedByID.ID)
}

func (suite *CacheBackedStoreTestSuite) TestCreateApplication_StoreError() {
	app := suite.createTestApp()
	storeErr := errors.New("store error")

	suite.mockStore.On("CreateApplication", mock.Anything, *app).Return(storeErr).Once()

	err := suite.cachedStore.CreateApplication(context.Background(), *app)
	suite.Equal(storeErr, err)
	suite.mockStore.AssertExpectations(suite.T())

	_, ok := suite.appByIDCache.Get(context.Background(), cache.CacheKey{Key: app.ID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestCreateApplication_CacheSetError() {
	app := suite.createTestApp()
	cacheSetErr := errors.New("cache set error")

	suite.appByIDCache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(cacheSetErr).Maybe()
	suite.mockStore.On("CreateApplication", mock.Anything, *app).Return(nil).Once()

	// Should not fail even if cache set fails
	err := suite.cachedStore.CreateApplication(context.Background(), *app)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

// CreateOAuthConfig tests

func (suite *CacheBackedStoreTestSuite) TestCreateOAuthConfig_Success() {
	rawJSON := json.RawMessage(`{"redirect_uris":["https://example.com/cb"]}`)
	suite.mockStore.On("CreateOAuthConfig", mock.Anything, testEntityID, rawJSON).Return(nil).Once()

	err := suite.cachedStore.CreateOAuthConfig(context.Background(), testEntityID, rawJSON)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestCreateOAuthConfig_StoreError() {
	rawJSON := json.RawMessage(`{}`)
	storeErr := errors.New("store error")
	suite.mockStore.On("CreateOAuthConfig", mock.Anything, testEntityID, rawJSON).Return(storeErr).Once()

	err := suite.cachedStore.CreateOAuthConfig(context.Background(), testEntityID, rawJSON)
	suite.Equal(storeErr, err)
}

// GetTotalApplicationCount tests

func (suite *CacheBackedStoreTestSuite) TestGetTotalApplicationCount_Success() {
	suite.mockStore.On("GetTotalApplicationCount", mock.Anything).Return(10, nil).Once()

	count, err := suite.cachedStore.GetTotalApplicationCount(context.Background())
	suite.Nil(err)
	suite.Equal(10, count)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestGetTotalApplicationCount_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("GetTotalApplicationCount", mock.Anything).Return(0, storeErr).Once()

	count, err := suite.cachedStore.GetTotalApplicationCount(context.Background())
	suite.Equal(storeErr, err)
	suite.Equal(0, count)
	suite.mockStore.AssertExpectations(suite.T())
}

// GetApplicationList tests

func (suite *CacheBackedStoreTestSuite) TestGetApplicationList_Success() {
	expectedList := []applicationConfigDAO{
		{ID: "app-1", AuthFlowID: "flow-1"},
		{ID: "app-2", AuthFlowID: "flow-2"},
	}

	suite.mockStore.On("GetApplicationList", mock.Anything).Return(expectedList, nil).Once()

	list, err := suite.cachedStore.GetApplicationList(context.Background())
	suite.Nil(err)
	suite.Equal(expectedList, list)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationList_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("GetApplicationList", mock.Anything).Return(([]applicationConfigDAO)(nil), storeErr).Once()

	list, err := suite.cachedStore.GetApplicationList(context.Background())
	suite.Equal(storeErr, err)
	suite.Nil(list)
	suite.mockStore.AssertExpectations(suite.T())
}

// GetApplicationByID tests

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByID_CacheHit() {
	app := suite.createTestApp()
	_ = suite.appByIDCache.Set(context.Background(), cache.CacheKey{Key: app.ID}, app)

	result, err := suite.cachedStore.GetApplicationByID(context.Background(), app.ID)
	suite.Nil(err)
	suite.Equal(app, result)

	suite.mockStore.AssertNotCalled(suite.T(), "GetApplicationByID")
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByID_CacheMiss() {
	app := suite.createTestApp()

	suite.mockStore.On("GetApplicationByID", mock.Anything, app.ID).Return(app, nil).Once()

	result, err := suite.cachedStore.GetApplicationByID(context.Background(), app.ID)
	suite.Nil(err)
	suite.Equal(app, result)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify it's now cached
	cachedApp, ok := suite.appByIDCache.Get(context.Background(), cache.CacheKey{Key: app.ID})
	suite.True(ok)
	suite.Equal(app.ID, cachedApp.ID)
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByID_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("GetApplicationByID", mock.Anything, "test-id").
		Return((*applicationConfigDAO)(nil), storeErr).Once()

	result, err := suite.cachedStore.GetApplicationByID(context.Background(), "test-id")
	suite.Equal(storeErr, err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByID_NilResult() {
	suite.mockStore.On("GetApplicationByID", mock.Anything, "test-id").
		Return((*applicationConfigDAO)(nil), nil).Once()

	result, err := suite.cachedStore.GetApplicationByID(context.Background(), "test-id")
	suite.Nil(err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())

	_, ok := suite.appByIDCache.Get(context.Background(), cache.CacheKey{Key: "test-id"})
	suite.False(ok)
}

// GetOAuthConfigByEntityID tests

func (suite *CacheBackedStoreTestSuite) TestGetOAuthConfigByEntityID_CacheHit() {
	cfg := suite.createTestOAuthConfig()
	_ = suite.oauthAppCache.Set(context.Background(), cache.CacheKey{Key: cfg.EntityID}, cfg)

	result, err := suite.cachedStore.GetOAuthConfigByEntityID(context.Background(), cfg.EntityID)
	suite.Nil(err)
	suite.Equal(cfg, result)

	suite.mockStore.AssertNotCalled(suite.T(), "GetOAuthConfigByEntityID")
}

func (suite *CacheBackedStoreTestSuite) TestGetOAuthConfigByEntityID_CacheMiss() {
	cfg := suite.createTestOAuthConfig()

	suite.mockStore.On("GetOAuthConfigByEntityID", mock.Anything, cfg.EntityID).Return(cfg, nil).Once()

	result, err := suite.cachedStore.GetOAuthConfigByEntityID(context.Background(), cfg.EntityID)
	suite.Nil(err)
	suite.Equal(cfg, result)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify it's now cached
	cachedCfg, ok := suite.oauthAppCache.Get(context.Background(), cache.CacheKey{Key: cfg.EntityID})
	suite.True(ok)
	suite.Equal(cfg.EntityID, cachedCfg.EntityID)
}

func (suite *CacheBackedStoreTestSuite) TestGetOAuthConfigByEntityID_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("GetOAuthConfigByEntityID", mock.Anything, "test-entity-id").
		Return((*oauthConfigDAO)(nil), storeErr).Once()

	result, err := suite.cachedStore.GetOAuthConfigByEntityID(context.Background(), "test-entity-id")
	suite.Equal(storeErr, err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestGetOAuthConfigByEntityID_NilResult() {
	suite.mockStore.On("GetOAuthConfigByEntityID", mock.Anything, "test-entity-id").
		Return((*oauthConfigDAO)(nil), nil).Once()

	result, err := suite.cachedStore.GetOAuthConfigByEntityID(context.Background(), "test-entity-id")
	suite.Nil(err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())

	_, ok := suite.oauthAppCache.Get(context.Background(), cache.CacheKey{Key: "test-entity-id"})
	suite.False(ok)
}

// UpdateApplication tests

func (suite *CacheBackedStoreTestSuite) TestUpdateApplication_Success() {
	app := suite.createTestApp()
	app.AuthFlowID = "updated-flow"

	// Pre-cache the old entry
	_ = suite.appByIDCache.Set(context.Background(), cache.CacheKey{Key: app.ID}, suite.createTestApp())

	suite.mockStore.On("UpdateApplication", mock.Anything, *app).Return(nil).Once()

	err := suite.cachedStore.UpdateApplication(context.Background(), *app)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify updated app is cached
	cachedByID, ok := suite.appByIDCache.Get(context.Background(), cache.CacheKey{Key: app.ID})
	suite.True(ok)
	suite.Equal("updated-flow", cachedByID.AuthFlowID)
}

func (suite *CacheBackedStoreTestSuite) TestUpdateApplication_StoreError() {
	app := suite.createTestApp()
	storeErr := errors.New("store error")

	suite.mockStore.On("UpdateApplication", mock.Anything, *app).Return(storeErr).Once()

	err := suite.cachedStore.UpdateApplication(context.Background(), *app)
	suite.Equal(storeErr, err)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestUpdateApplication_CacheInvalidationError() {
	app := suite.createTestApp()
	cacheDelErr := errors.New("cache delete error")

	suite.appByIDCache.EXPECT().Delete(mock.Anything, mock.Anything).Return(cacheDelErr).Maybe()
	suite.mockStore.On("UpdateApplication", mock.Anything, *app).Return(nil).Once()

	// Should not fail even if cache invalidation fails
	err := suite.cachedStore.UpdateApplication(context.Background(), *app)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

// UpdateOAuthConfig tests

func (suite *CacheBackedStoreTestSuite) TestUpdateOAuthConfig_Success() {
	rawJSON := json.RawMessage(`{"redirect_uris":["https://example.com/cb"]}`)
	entityID := testEntityID
	// Pre-cache entry to verify it's invalidated
	_ = suite.oauthAppCache.Set(context.Background(), cache.CacheKey{Key: entityID}, suite.createTestOAuthConfig())

	suite.mockStore.On("UpdateOAuthConfig", mock.Anything, entityID, rawJSON).Return(nil).Once()

	err := suite.cachedStore.UpdateOAuthConfig(context.Background(), entityID, rawJSON)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify OAuth cache is invalidated
	_, ok := suite.oauthAppCache.Get(context.Background(), cache.CacheKey{Key: entityID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestUpdateOAuthConfig_StoreError() {
	rawJSON := json.RawMessage(`{}`)
	storeErr := errors.New("store error")
	suite.mockStore.On("UpdateOAuthConfig", mock.Anything, testEntityID, rawJSON).Return(storeErr).Once()

	err := suite.cachedStore.UpdateOAuthConfig(context.Background(), testEntityID, rawJSON)
	suite.Equal(storeErr, err)
}

// DeleteApplication tests

func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_Success() {
	appID := "test-app-id"
	app := suite.createTestApp()
	cfg := suite.createTestOAuthConfig()
	cfg.EntityID = appID

	// Pre-cache entries
	_ = suite.appByIDCache.Set(context.Background(), cache.CacheKey{Key: appID}, app)
	_ = suite.oauthAppCache.Set(context.Background(), cache.CacheKey{Key: appID}, cfg)

	suite.mockStore.On("DeleteApplication", mock.Anything, appID).Return(nil).Once()

	err := suite.cachedStore.DeleteApplication(context.Background(), appID)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify both caches are invalidated
	_, ok := suite.appByIDCache.Get(context.Background(), cache.CacheKey{Key: appID})
	suite.False(ok)
	_, ok = suite.oauthAppCache.Get(context.Background(), cache.CacheKey{Key: appID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("DeleteApplication", mock.Anything, "test-id").Return(storeErr).Once()

	err := suite.cachedStore.DeleteApplication(context.Background(), "test-id")
	suite.Equal(storeErr, err)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_CacheInvalidationError() {
	cacheDelErr := errors.New("cache delete error")
	suite.appByIDCache.EXPECT().Delete(mock.Anything, mock.Anything).Return(cacheDelErr).Maybe()
	suite.oauthAppCache.EXPECT().Delete(mock.Anything, mock.Anything).Return(cacheDelErr).Maybe()

	suite.mockStore.On("DeleteApplication", mock.Anything, "test-id").Return(nil).Once()

	// Should not fail even if cache invalidation fails
	err := suite.cachedStore.DeleteApplication(context.Background(), "test-id")
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

// DeleteOAuthConfig tests

func (suite *CacheBackedStoreTestSuite) TestDeleteOAuthConfig_Success() {
	entityID := testEntityID
	cfg := suite.createTestOAuthConfig()
	cfg.EntityID = entityID
	_ = suite.oauthAppCache.Set(context.Background(), cache.CacheKey{Key: entityID}, cfg)

	suite.mockStore.On("DeleteOAuthConfig", mock.Anything, entityID).Return(nil).Once()

	err := suite.cachedStore.DeleteOAuthConfig(context.Background(), entityID)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	_, ok := suite.oauthAppCache.Get(context.Background(), cache.CacheKey{Key: entityID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestDeleteOAuthConfig_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("DeleteOAuthConfig", mock.Anything, testEntityID).Return(storeErr).Once()

	err := suite.cachedStore.DeleteOAuthConfig(context.Background(), testEntityID)
	suite.Equal(storeErr, err)
}

// IsApplicationExists tests

func (suite *CacheBackedStoreTestSuite) TestIsApplicationExists_Success() {
	appID := "test-app-123"
	suite.mockStore.EXPECT().IsApplicationExists(mock.Anything, appID).Return(true, nil).Once()

	exists, err := suite.cachedStore.IsApplicationExists(context.Background(), appID)
	suite.NoError(err)
	suite.True(exists)
}

func (suite *CacheBackedStoreTestSuite) TestIsApplicationExists_NotFound() {
	appID := "non-existent-app"
	suite.mockStore.EXPECT().IsApplicationExists(mock.Anything, appID).Return(false, nil).Once()

	exists, err := suite.cachedStore.IsApplicationExists(context.Background(), appID)
	suite.NoError(err)
	suite.False(exists)
}

func (suite *CacheBackedStoreTestSuite) TestIsApplicationExists_Error() {
	appID := "error-app"
	expectedErr := errors.New("database error")
	suite.mockStore.EXPECT().IsApplicationExists(mock.Anything, appID).Return(false, expectedErr).Once()

	exists, err := suite.cachedStore.IsApplicationExists(context.Background(), appID)
	suite.Error(err)
	suite.False(exists)
	suite.Equal(expectedErr, err)
}

// IsApplicationDeclarative tests

func (suite *CacheBackedStoreTestSuite) TestIsApplicationDeclarative_True() {
	suite.mockStore.EXPECT().IsApplicationDeclarative(mock.Anything, "decl-app").Return(true).Once()

	result := suite.cachedStore.IsApplicationDeclarative(context.Background(), "decl-app")
	suite.True(result)
}

func (suite *CacheBackedStoreTestSuite) TestIsApplicationDeclarative_False() {
	suite.mockStore.EXPECT().IsApplicationDeclarative(mock.Anything, "non-decl-app").Return(false).Once()

	result := suite.cachedStore.IsApplicationDeclarative(context.Background(), "non-decl-app")
	suite.False(result)
}

// cacheAppConfig tests (private helper)

func (suite *CacheBackedStoreTestSuite) TestCacheAppConfig_WithNil() {
	suite.cachedStore.cacheAppConfig(context.Background(), nil)
	suite.Equal(0, len(suite.appByIDData))
}

func (suite *CacheBackedStoreTestSuite) TestCacheAppConfig_WithEmptyID() {
	app := suite.createTestApp()
	app.ID = ""

	suite.cachedStore.cacheAppConfig(context.Background(), app)
	suite.Equal(0, len(suite.appByIDData))
}

func (suite *CacheBackedStoreTestSuite) TestCacheAppConfig_Success() {
	app := suite.createTestApp()

	suite.cachedStore.cacheAppConfig(context.Background(), app)

	cachedByID, ok := suite.appByIDCache.Get(context.Background(), cache.CacheKey{Key: app.ID})
	suite.True(ok)
	suite.Equal(app.ID, cachedByID.ID)
}

func (suite *CacheBackedStoreTestSuite) TestCacheAppConfig_CacheSetError() {
	app := suite.createTestApp()
	cacheSetErr := errors.New("cache set error")

	suite.appByIDCache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(cacheSetErr).Maybe()

	// Should not panic or fail
	suite.cachedStore.cacheAppConfig(context.Background(), app)
}

// cacheOAuthConfig tests (private helper)

func (suite *CacheBackedStoreTestSuite) TestCacheOAuthConfig_WithNil() {
	suite.cachedStore.cacheOAuthConfig(context.Background(), nil)
	suite.Equal(0, len(suite.oauthAppData))
}

func (suite *CacheBackedStoreTestSuite) TestCacheOAuthConfig_WithEmptyEntityID() {
	cfg := suite.createTestOAuthConfig()
	cfg.EntityID = ""

	suite.cachedStore.cacheOAuthConfig(context.Background(), cfg)
	suite.Equal(0, len(suite.oauthAppData))
}

func (suite *CacheBackedStoreTestSuite) TestCacheOAuthConfig_Success() {
	cfg := suite.createTestOAuthConfig()

	suite.cachedStore.cacheOAuthConfig(context.Background(), cfg)

	cached, ok := suite.oauthAppCache.Get(context.Background(), cache.CacheKey{Key: cfg.EntityID})
	suite.True(ok)
	suite.Equal(cfg.EntityID, cached.EntityID)
}

func (suite *CacheBackedStoreTestSuite) TestCacheOAuthConfig_CacheSetError() {
	cfg := suite.createTestOAuthConfig()
	cacheSetErr := errors.New("cache set error")

	suite.oauthAppCache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(cacheSetErr).Maybe()

	// Should not panic or fail
	suite.cachedStore.cacheOAuthConfig(context.Background(), cfg)
}

// invalidateAppCache tests (private helper)

func (suite *CacheBackedStoreTestSuite) TestInvalidateAppCache_EmptyID() {
	suite.cachedStore.invalidateAppCache(context.Background(), "")
	suite.Equal(0, len(suite.appByIDData))
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateAppCache_Success() {
	app := suite.createTestApp()
	_ = suite.appByIDCache.Set(context.Background(), cache.CacheKey{Key: app.ID}, app)

	suite.cachedStore.invalidateAppCache(context.Background(), app.ID)

	_, ok := suite.appByIDCache.Get(context.Background(), cache.CacheKey{Key: app.ID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateAppCache_DeleteError() {
	cacheDelErr := errors.New("cache delete error")
	suite.appByIDCache.EXPECT().Delete(mock.Anything, mock.Anything).Return(cacheDelErr).Maybe()

	// Should not panic or fail
	suite.cachedStore.invalidateAppCache(context.Background(), "some-id")
}

// invalidateOAuthCache tests (private helper)

func (suite *CacheBackedStoreTestSuite) TestInvalidateOAuthCache_EmptyEntityID() {
	suite.cachedStore.invalidateOAuthCache(context.Background(), "")
	suite.Equal(0, len(suite.oauthAppData))
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateOAuthCache_Success() {
	cfg := suite.createTestOAuthConfig()
	_ = suite.oauthAppCache.Set(context.Background(), cache.CacheKey{Key: cfg.EntityID}, cfg)

	suite.cachedStore.invalidateOAuthCache(context.Background(), cfg.EntityID)

	_, ok := suite.oauthAppCache.Get(context.Background(), cache.CacheKey{Key: cfg.EntityID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateOAuthCache_DeleteError() {
	cacheDelErr := errors.New("cache delete error")
	suite.oauthAppCache.EXPECT().Delete(mock.Anything, mock.Anything).Return(cacheDelErr).Maybe()

	// Should not panic or fail
	suite.cachedStore.invalidateOAuthCache(context.Background(), "some-entity-id")
}
