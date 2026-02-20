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
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/tests/mocks/cachemock"
)

// Test suite
type CacheBackedStoreTestSuite struct {
	suite.Suite
	mockStore      *applicationStoreInterfaceMock
	appByIDCache   *cachemock.CacheInterfaceMock[*model.ApplicationProcessedDTO]
	appByNameCache *cachemock.CacheInterfaceMock[*model.ApplicationProcessedDTO]
	oauthAppCache  *cachemock.CacheInterfaceMock[*model.OAuthAppConfigProcessedDTO]
	cachedStore    *cachedBackedApplicationStore
	// Helper maps to track cached values for verification (one per cache type)
	appByIDData   map[string]*model.ApplicationProcessedDTO
	appByNameData map[string]*model.ApplicationProcessedDTO
	oauthAppData  map[string]*model.OAuthAppConfigProcessedDTO
}

func TestCacheBackedStoreTestSuite(t *testing.T) {
	suite.Run(t, new(CacheBackedStoreTestSuite))
}

func (suite *CacheBackedStoreTestSuite) SetupTest() {
	suite.mockStore = newApplicationStoreInterfaceMock(suite.T())
	suite.appByIDData = make(map[string]*model.ApplicationProcessedDTO)
	suite.appByNameData = make(map[string]*model.ApplicationProcessedDTO)
	suite.oauthAppData = make(map[string]*model.OAuthAppConfigProcessedDTO)

	// Create mockery-generated cache mocks
	suite.appByIDCache = cachemock.NewCacheInterfaceMock[*model.ApplicationProcessedDTO](suite.T())
	suite.appByNameCache = cachemock.NewCacheInterfaceMock[*model.ApplicationProcessedDTO](suite.T())
	suite.oauthAppCache = cachemock.NewCacheInterfaceMock[*model.OAuthAppConfigProcessedDTO](suite.T())

	// Configure mocks to track Set operations and return values on Get
	suite.setupAppCacheMock(suite.appByIDCache, suite.appByIDData)
	suite.setupAppCacheMock(suite.appByNameCache, suite.appByNameData)
	suite.setupOAuthCacheMock(suite.oauthAppCache, suite.oauthAppData)

	// Set up IsEnabled to return true by default
	suite.appByIDCache.EXPECT().IsEnabled().Return(true).Maybe()
	suite.appByNameCache.EXPECT().IsEnabled().Return(true).Maybe()
	suite.oauthAppCache.EXPECT().IsEnabled().Return(true).Maybe()

	suite.cachedStore = &cachedBackedApplicationStore{
		AppByIDCache:   suite.appByIDCache,
		AppByNameCache: suite.appByNameCache,
		OAuthAppCache:  suite.oauthAppCache,
		Store:          suite.mockStore,
	}
}

// setupCacheMock configures a cache mock to track Set operations and return values on Get
func setupCacheMock[T any](
	mockCache *cachemock.CacheInterfaceMock[T],
	data map[string]T,
) {
	// Set up Set to track values
	mockCache.EXPECT().Set(mock.Anything, mock.Anything).
		RunAndReturn(func(key cache.CacheKey, value T) error {
			data[key.Key] = value
			return nil
		}).Maybe()

	// Set up Get to return tracked values
	mockCache.EXPECT().Get(mock.Anything).
		RunAndReturn(func(key cache.CacheKey) (T, bool) {
			if val, ok := data[key.Key]; ok {
				return val, true
			}
			var zero T
			return zero, false
		}).Maybe()

	// Set up Delete to remove from tracked values
	mockCache.EXPECT().Delete(mock.Anything).
		RunAndReturn(func(key cache.CacheKey) error {
			delete(data, key.Key)
			return nil
		}).Maybe()

	// Set up Clear to clear tracked values
	mockCache.EXPECT().Clear().
		RunAndReturn(func() error {
			for k := range data {
				delete(data, k)
			}
			return nil
		}).Maybe()

	// Set up GetName
	mockCache.EXPECT().GetName().Return("mockCache").Maybe()

	// Set up CleanupExpired
	mockCache.EXPECT().CleanupExpired().Maybe()
}

// setupAppCacheMock configures an Application cache mock to track Set operations and return values on Get
func (suite *CacheBackedStoreTestSuite) setupAppCacheMock(
	mockCache *cachemock.CacheInterfaceMock[*model.ApplicationProcessedDTO],
	data map[string]*model.ApplicationProcessedDTO,
) {
	setupCacheMock(mockCache, data)
}

// setupOAuthCacheMock configures an OAuth cache mock to track Set operations and return values on Get
func (suite *CacheBackedStoreTestSuite) setupOAuthCacheMock(
	mockCache *cachemock.CacheInterfaceMock[*model.OAuthAppConfigProcessedDTO],
	data map[string]*model.OAuthAppConfigProcessedDTO,
) {
	setupCacheMock(mockCache, data)
}

// Helper functions
func (suite *CacheBackedStoreTestSuite) createTestApp() *model.ApplicationProcessedDTO {
	return &model.ApplicationProcessedDTO{
		ID:                        "test-app-id",
		Name:                      "Test App",
		Description:               "Test Description",
		AuthFlowID:                "auth-flow-1",
		RegistrationFlowID:        "reg-flow-1",
		IsRegistrationFlowEnabled: true,
		URL:                       "https://example.com",
		LogoURL:                   "https://example.com/logo.png",
		Assertion: &model.AssertionConfig{
			ValidityPeriod: 3600,
			UserAttributes: []string{"email", "name"},
		},
	}
}

func (suite *CacheBackedStoreTestSuite) createTestAppWithOAuth() *model.ApplicationProcessedDTO {
	app := suite.createTestApp()
	app.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{
		{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
				AppID:              app.ID,
				ClientID:           "test-client-id",
				HashedClientSecret: "hashed-secret",
				RedirectURIs:       []string{"https://example.com/callback"},
				GrantTypes:         []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
				ResponseTypes:      []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
				PKCERequired:       true,
				PublicClient:       false,
			},
		},
	}
	return app
}

func (suite *CacheBackedStoreTestSuite) createTestOAuthApp() *model.OAuthAppConfigProcessedDTO {
	return &model.OAuthAppConfigProcessedDTO{
		AppID:              "test-app-id",
		ClientID:           "test-client-id",
		HashedClientSecret: "hashed-secret",
		RedirectURIs:       []string{"https://example.com/callback"},
		GrantTypes:         []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ResponseTypes:      []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
		PKCERequired:       true,
		PublicClient:       false,
	}
}

// Test newCachedBackedApplicationStore
func (suite *CacheBackedStoreTestSuite) TestNewCachedBackedApplicationStore() {
	// This test verifies that the suite setup works correctly
	// The actual newCachedBackedApplicationStore function requires Thunder runtime
	// which is initialized in the suite setup via mock caches
	suite.NotNil(suite.cachedStore)
	suite.IsType(&cachedBackedApplicationStore{}, suite.cachedStore)
	suite.NotNil(suite.cachedStore.AppByIDCache)
	suite.NotNil(suite.cachedStore.AppByNameCache)
	suite.NotNil(suite.cachedStore.OAuthAppCache)
	suite.NotNil(suite.cachedStore.Store)
}

// CreateApplication tests
func (suite *CacheBackedStoreTestSuite) TestCreateApplication_Success() {
	app := suite.createTestApp()

	suite.mockStore.On("CreateApplication", *app).Return(nil).Once()

	err := suite.cachedStore.CreateApplication(*app)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify application is cached by ID
	cachedByID, ok := suite.appByIDCache.Get(cache.CacheKey{Key: app.ID})
	suite.True(ok)
	suite.Equal(app.ID, cachedByID.ID)

	// Verify application is cached by name
	cachedByName, ok := suite.appByNameCache.Get(cache.CacheKey{Key: app.Name})
	suite.True(ok)
	suite.Equal(app.Name, cachedByName.Name)
}

func (suite *CacheBackedStoreTestSuite) TestCreateApplication_WithOAuth() {
	app := suite.createTestAppWithOAuth()

	suite.mockStore.On("CreateApplication", *app).Return(nil).Once()

	err := suite.cachedStore.CreateApplication(*app)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify OAuth app is cached
	cachedOAuth, ok := suite.oauthAppCache.Get(cache.CacheKey{Key: "test-client-id"})
	suite.True(ok)
	suite.Equal("test-client-id", cachedOAuth.ClientID)
}

func (suite *CacheBackedStoreTestSuite) TestCreateApplication_StoreError() {
	app := suite.createTestApp()
	storeErr := errors.New("store error")

	suite.mockStore.On("CreateApplication", *app).Return(storeErr).Once()

	err := suite.cachedStore.CreateApplication(*app)
	suite.Equal(storeErr, err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify nothing is cached on error
	_, ok := suite.appByIDCache.Get(cache.CacheKey{Key: app.ID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestCreateApplication_CacheSetError() {
	app := suite.createTestApp()
	cacheSetErr := errors.New("cache set error")

	// Override Set expectation to return error
	suite.appByIDCache.EXPECT().Set(mock.Anything, mock.Anything).Return(cacheSetErr).Maybe()
	suite.appByNameCache.EXPECT().Set(mock.Anything, mock.Anything).Return(cacheSetErr).Maybe()

	suite.mockStore.On("CreateApplication", *app).Return(nil).Once()

	// Should not fail even if cache set fails
	err := suite.cachedStore.CreateApplication(*app)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

// GetTotalApplicationCount tests
func (suite *CacheBackedStoreTestSuite) TestGetTotalApplicationCount_Success() {
	suite.mockStore.On("GetTotalApplicationCount").Return(10, nil).Once()

	count, err := suite.cachedStore.GetTotalApplicationCount()
	suite.Nil(err)
	suite.Equal(10, count)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestGetTotalApplicationCount_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("GetTotalApplicationCount").Return(0, storeErr).Once()

	count, err := suite.cachedStore.GetTotalApplicationCount()
	suite.Equal(storeErr, err)
	suite.Equal(0, count)
	suite.mockStore.AssertExpectations(suite.T())
}

// GetApplicationList tests
func (suite *CacheBackedStoreTestSuite) TestGetApplicationList_Success() {
	expectedList := []model.BasicApplicationDTO{
		{
			ID:          "app-1",
			Name:        "App 1",
			Description: "Description 1",
		},
		{
			ID:          "app-2",
			Name:        "App 2",
			Description: "Description 2",
		},
	}

	suite.mockStore.On("GetApplicationList").Return(expectedList, nil).Once()

	list, err := suite.cachedStore.GetApplicationList()
	suite.Nil(err)
	suite.Equal(expectedList, list)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationList_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("GetApplicationList").Return(([]model.BasicApplicationDTO)(nil), storeErr).Once()

	list, err := suite.cachedStore.GetApplicationList()
	suite.Equal(storeErr, err)
	suite.Nil(list)
	suite.mockStore.AssertExpectations(suite.T())
}

// GetOAuthApplication tests
func (suite *CacheBackedStoreTestSuite) TestGetOAuthApplication_CacheHit() {
	oauthApp := suite.createTestOAuthApp()
	_ = suite.oauthAppCache.Set(cache.CacheKey{Key: "test-client-id"}, oauthApp)

	result, err := suite.cachedStore.GetOAuthApplication("test-client-id")
	suite.Nil(err)
	suite.Equal(oauthApp, result)

	// Store should not be called
	suite.mockStore.AssertNotCalled(suite.T(), "GetOAuthApplication")
}

func (suite *CacheBackedStoreTestSuite) TestGetOAuthApplication_CacheMiss() {
	oauthApp := suite.createTestOAuthApp()

	suite.mockStore.On("GetOAuthApplication", "test-client-id").Return(oauthApp, nil).Once()

	result, err := suite.cachedStore.GetOAuthApplication("test-client-id")
	suite.Nil(err)
	suite.Equal(oauthApp, result)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify it's now cached
	cachedOAuth, ok := suite.oauthAppCache.Get(cache.CacheKey{Key: "test-client-id"})
	suite.True(ok)
	suite.Equal(oauthApp.ClientID, cachedOAuth.ClientID)
}

func (suite *CacheBackedStoreTestSuite) TestGetOAuthApplication_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("GetOAuthApplication", "test-client-id").
		Return((*model.OAuthAppConfigProcessedDTO)(nil), storeErr).Once()

	result, err := suite.cachedStore.GetOAuthApplication("test-client-id")
	suite.Equal(storeErr, err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestGetOAuthApplication_NilResult() {
	suite.mockStore.On("GetOAuthApplication", "test-client-id").
		Return((*model.OAuthAppConfigProcessedDTO)(nil), nil).Once()

	result, err := suite.cachedStore.GetOAuthApplication("test-client-id")
	suite.Nil(err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify nothing is cached
	_, ok := suite.oauthAppCache.Get(cache.CacheKey{Key: "test-client-id"})
	suite.False(ok)
}

// GetApplicationByID tests
func (suite *CacheBackedStoreTestSuite) TestGetApplicationByID_CacheHit() {
	app := suite.createTestApp()
	_ = suite.appByIDCache.Set(cache.CacheKey{Key: app.ID}, app)

	result, err := suite.cachedStore.GetApplicationByID(app.ID)
	suite.Nil(err)
	suite.Equal(app, result)

	// Store should not be called
	suite.mockStore.AssertNotCalled(suite.T(), "GetApplicationByID")
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByID_CacheMiss() {
	app := suite.createTestApp()

	suite.mockStore.On("GetApplicationByID", app.ID).Return(app, nil).Once()

	result, err := suite.cachedStore.GetApplicationByID(app.ID)
	suite.Nil(err)
	suite.Equal(app, result)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify it's now cached
	cachedApp, ok := suite.appByIDCache.Get(cache.CacheKey{Key: app.ID})
	suite.True(ok)
	suite.Equal(app.ID, cachedApp.ID)
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByID_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("GetApplicationByID", "test-id").Return((*model.ApplicationProcessedDTO)(nil), storeErr).Once()

	result, err := suite.cachedStore.GetApplicationByID("test-id")
	suite.Equal(storeErr, err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByID_NilResult() {
	suite.mockStore.On("GetApplicationByID", "test-id").Return((*model.ApplicationProcessedDTO)(nil), nil).Once()

	result, err := suite.cachedStore.GetApplicationByID("test-id")
	suite.Nil(err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify nothing is cached
	_, ok := suite.appByIDCache.Get(cache.CacheKey{Key: "test-id"})
	suite.False(ok)
}

// GetApplicationByName tests
func (suite *CacheBackedStoreTestSuite) TestGetApplicationByName_CacheHit() {
	app := suite.createTestApp()
	_ = suite.appByNameCache.Set(cache.CacheKey{Key: app.Name}, app)

	result, err := suite.cachedStore.GetApplicationByName(app.Name)
	suite.Nil(err)
	suite.Equal(app, result)

	// Store should not be called
	suite.mockStore.AssertNotCalled(suite.T(), "GetApplicationByName")
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByName_CacheMiss() {
	app := suite.createTestApp()

	suite.mockStore.On("GetApplicationByName", app.Name).Return(app, nil).Once()

	result, err := suite.cachedStore.GetApplicationByName(app.Name)
	suite.Nil(err)
	suite.Equal(app, result)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify it's now cached
	cachedApp, ok := suite.appByNameCache.Get(cache.CacheKey{Key: app.Name})
	suite.True(ok)
	suite.Equal(app.Name, cachedApp.Name)
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByName_StoreError() {
	storeErr := errors.New("store error")
	suite.mockStore.On("GetApplicationByName", "test-name").
		Return((*model.ApplicationProcessedDTO)(nil), storeErr).Once()

	result, err := suite.cachedStore.GetApplicationByName("test-name")
	suite.Equal(storeErr, err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestGetApplicationByName_NilResult() {
	suite.mockStore.On("GetApplicationByName", "test-name").Return((*model.ApplicationProcessedDTO)(nil), nil).Once()

	result, err := suite.cachedStore.GetApplicationByName("test-name")
	suite.Nil(err)
	suite.Nil(result)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify nothing is cached
	_, ok := suite.appByNameCache.Get(cache.CacheKey{Key: "test-name"})
	suite.False(ok)
}

// UpdateApplication tests
func (suite *CacheBackedStoreTestSuite) TestUpdateApplication_Success() {
	existingApp := suite.createTestApp()
	updatedApp := suite.createTestApp()
	updatedApp.Name = "Updated App"
	updatedApp.Description = "Updated Description"

	suite.mockStore.On("UpdateApplication", existingApp, updatedApp).Return(nil).Once()

	err := suite.cachedStore.UpdateApplication(existingApp, updatedApp)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify updated app is cached
	cachedByID, ok := suite.appByIDCache.Get(cache.CacheKey{Key: updatedApp.ID})
	suite.True(ok)
	suite.Equal(updatedApp.Name, cachedByID.Name)
}

func (suite *CacheBackedStoreTestSuite) TestUpdateApplication_WithOAuth() {
	existingApp := suite.createTestAppWithOAuth()
	updatedApp := suite.createTestAppWithOAuth()
	updatedApp.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs = []string{"https://new.example.com/callback"}

	// Pre-cache the existing app
	_ = suite.appByIDCache.Set(cache.CacheKey{Key: existingApp.ID}, existingApp)
	_ = suite.appByNameCache.Set(cache.CacheKey{Key: existingApp.Name}, existingApp)
	_ = suite.oauthAppCache.Set(
		cache.CacheKey{Key: existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID},
		existingApp.InboundAuthConfig[0].OAuthAppConfig,
	)

	suite.mockStore.On("UpdateApplication", existingApp, updatedApp).Return(nil).Once()

	err := suite.cachedStore.UpdateApplication(existingApp, updatedApp)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify old caches are invalidated and new app is cached
	clientID := updatedApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
	cachedOAuth, ok := suite.oauthAppCache.Get(cache.CacheKey{Key: clientID})
	suite.True(ok)
	suite.Equal(updatedApp.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs, cachedOAuth.RedirectURIs)
}

func (suite *CacheBackedStoreTestSuite) TestUpdateApplication_StoreError() {
	existingApp := suite.createTestApp()
	updatedApp := suite.createTestApp()
	storeErr := errors.New("store error")

	suite.mockStore.On("UpdateApplication", existingApp, updatedApp).Return(storeErr).Once()

	err := suite.cachedStore.UpdateApplication(existingApp, updatedApp)
	suite.Equal(storeErr, err)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestUpdateApplication_CacheInvalidationError() {
	existingApp := suite.createTestApp()
	updatedApp := suite.createTestApp()
	cacheDelErr := errors.New("cache delete error")

	// Override Delete expectation to return error
	suite.appByIDCache.EXPECT().Delete(mock.Anything).Return(cacheDelErr).Maybe()
	suite.appByNameCache.EXPECT().Delete(mock.Anything).Return(cacheDelErr).Maybe()

	suite.mockStore.On("UpdateApplication", existingApp, updatedApp).Return(nil).Once()

	// Should not fail even if cache invalidation fails
	err := suite.cachedStore.UpdateApplication(existingApp, updatedApp)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

// DeleteApplication tests
func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_FoundInCache() {
	app := suite.createTestAppWithOAuth()
	_ = suite.appByIDCache.Set(cache.CacheKey{Key: app.ID}, app)

	suite.mockStore.On("DeleteApplication", app.ID).Return(nil).Once()

	err := suite.cachedStore.DeleteApplication(app.ID)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())

	// Verify all caches are invalidated
	_, ok := suite.appByIDCache.Get(cache.CacheKey{Key: app.ID})
	suite.False(ok)
	_, ok = suite.appByNameCache.Get(cache.CacheKey{Key: app.Name})
	suite.False(ok)
	_, ok = suite.oauthAppCache.Get(cache.CacheKey{Key: app.InboundAuthConfig[0].OAuthAppConfig.ClientID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_NotFoundInCache() {
	app := suite.createTestApp()

	suite.mockStore.On("GetApplicationByID", app.ID).Return(app, nil).Once()
	suite.mockStore.On("DeleteApplication", app.ID).Return(nil).Once()

	err := suite.cachedStore.DeleteApplication(app.ID)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_GetApplicationError() {
	storeErr := errors.New("store error")

	suite.mockStore.On("GetApplicationByID", "test-id").
		Return((*model.ApplicationProcessedDTO)(nil), storeErr).Once()

	err := suite.cachedStore.DeleteApplication("test-id")
	suite.Equal(storeErr, err)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_ApplicationNotFoundError() {
	suite.mockStore.On("GetApplicationByID", "test-id").
		Return((*model.ApplicationProcessedDTO)(nil), model.ApplicationNotFoundError).Once()

	err := suite.cachedStore.DeleteApplication("test-id")
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_NilApplication() {
	suite.mockStore.On("GetApplicationByID", "test-id").Return((*model.ApplicationProcessedDTO)(nil), nil).Once()

	err := suite.cachedStore.DeleteApplication("test-id")
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_StoreError() {
	app := suite.createTestApp()
	_ = suite.appByIDCache.Set(cache.CacheKey{Key: app.ID}, app)
	storeErr := errors.New("store error")

	suite.mockStore.On("DeleteApplication", app.ID).Return(storeErr).Once()

	err := suite.cachedStore.DeleteApplication(app.ID)
	suite.Equal(storeErr, err)
	suite.mockStore.AssertExpectations(suite.T())
}

func (suite *CacheBackedStoreTestSuite) TestDeleteApplication_CacheInvalidationError() {
	app := suite.createTestApp()
	_ = suite.appByIDCache.Set(cache.CacheKey{Key: app.ID}, app)
	cacheDelErr := errors.New("cache delete error")

	// Override Delete expectation to return error
	suite.appByIDCache.EXPECT().Delete(mock.Anything).Return(cacheDelErr).Maybe()
	suite.appByNameCache.EXPECT().Delete(mock.Anything).Return(cacheDelErr).Maybe()

	suite.mockStore.On("DeleteApplication", app.ID).Return(nil).Once()

	// Should not fail even if cache invalidation fails
	err := suite.cachedStore.DeleteApplication(app.ID)
	suite.Nil(err)
	suite.mockStore.AssertExpectations(suite.T())
}

// cacheApplication tests
func (suite *CacheBackedStoreTestSuite) TestCacheApplication_WithNil() {
	// Should not panic with nil
	suite.cachedStore.cacheApplication(nil)

	// Verify nothing is cached
	suite.Equal(0, len(suite.appByIDData))
	suite.Equal(0, len(suite.appByNameData))
}

func (suite *CacheBackedStoreTestSuite) TestCacheApplication_WithEmptyID() {
	app := suite.createTestApp()
	app.ID = ""

	suite.cachedStore.cacheApplication(app)

	// Verify only name is cached
	_, ok := suite.appByIDCache.Get(cache.CacheKey{Key: ""})
	suite.False(ok)
	cachedByName, ok := suite.appByNameCache.Get(cache.CacheKey{Key: app.Name})
	suite.True(ok)
	suite.Equal(app.Name, cachedByName.Name)
}

func (suite *CacheBackedStoreTestSuite) TestCacheApplication_WithEmptyName() {
	app := suite.createTestApp()
	app.Name = ""

	suite.cachedStore.cacheApplication(app)

	// Verify only ID is cached
	cachedByID, ok := suite.appByIDCache.Get(cache.CacheKey{Key: app.ID})
	suite.True(ok)
	suite.Equal(app.ID, cachedByID.ID)
	_, ok = suite.appByNameCache.Get(cache.CacheKey{Key: ""})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestCacheApplication_WithOAuthConfig() {
	app := suite.createTestAppWithOAuth()

	suite.cachedStore.cacheApplication(app)

	// Verify all caches are populated
	_, ok := suite.appByIDCache.Get(cache.CacheKey{Key: app.ID})
	suite.True(ok)
	_, ok = suite.appByNameCache.Get(cache.CacheKey{Key: app.Name})
	suite.True(ok)
	_, ok = suite.oauthAppCache.Get(cache.CacheKey{Key: app.InboundAuthConfig[0].OAuthAppConfig.ClientID})
	suite.True(ok)
}

func (suite *CacheBackedStoreTestSuite) TestCacheApplication_CacheSetError() {
	app := suite.createTestApp()
	cacheSetErr := errors.New("cache set error")

	// Override Set expectation for ID cache to return error
	suite.appByIDCache.EXPECT().Set(mock.Anything, mock.Anything).Return(cacheSetErr).Maybe()

	// Should not panic or fail
	suite.cachedStore.cacheApplication(app)

	// Name cache should still work
	cachedByName, ok := suite.appByNameCache.Get(cache.CacheKey{Key: app.Name})
	suite.True(ok)
	suite.Equal(app.Name, cachedByName.Name)
}

func (suite *CacheBackedStoreTestSuite) TestCacheApplication_NameCacheSetError() {
	app := suite.createTestApp()
	cacheSetErr := errors.New("cache set error")

	// Override Set expectation for name cache to return error
	suite.appByNameCache.EXPECT().Set(mock.Anything, mock.Anything).Return(cacheSetErr).Maybe()

	// Should not panic or fail
	suite.cachedStore.cacheApplication(app)

	// ID cache should still work
	cachedByID, ok := suite.appByIDCache.Get(cache.CacheKey{Key: app.ID})
	suite.True(ok)
	suite.Equal(app.ID, cachedByID.ID)
}

// cacheOAuthApplication tests
func (suite *CacheBackedStoreTestSuite) TestCacheOAuthApplication_WithNil() {
	// Should not panic with nil
	suite.cachedStore.cacheOAuthApplication(nil)

	// Verify nothing is cached
	suite.Equal(0, len(suite.oauthAppData))
}

func (suite *CacheBackedStoreTestSuite) TestCacheOAuthApplication_WithEmptyClientID() {
	oauthApp := suite.createTestOAuthApp()
	oauthApp.ClientID = ""

	suite.cachedStore.cacheOAuthApplication(oauthApp)

	// Verify nothing is cached
	suite.Equal(0, len(suite.oauthAppData))
}

func (suite *CacheBackedStoreTestSuite) TestCacheOAuthApplication_Success() {
	oauthApp := suite.createTestOAuthApp()

	suite.cachedStore.cacheOAuthApplication(oauthApp)

	// Verify it's cached
	cached, ok := suite.oauthAppCache.Get(cache.CacheKey{Key: oauthApp.ClientID})
	suite.True(ok)
	suite.Equal(oauthApp.ClientID, cached.ClientID)
}

func (suite *CacheBackedStoreTestSuite) TestCacheOAuthApplication_CacheSetError() {
	oauthApp := suite.createTestOAuthApp()
	cacheSetErr := errors.New("cache set error")

	// Override Set expectation to return error
	suite.oauthAppCache.EXPECT().Set(mock.Anything, mock.Anything).Return(cacheSetErr).Maybe()

	// Should not panic or fail
	suite.cachedStore.cacheOAuthApplication(oauthApp)
}

// invalidateApplicationCache tests
func (suite *CacheBackedStoreTestSuite) TestInvalidateApplicationCache_WithEmptyValues() {
	// Should not panic with empty values
	suite.cachedStore.invalidateApplicationCache("", "", "")

	// Verify nothing happens
	suite.Equal(0, len(suite.appByIDData))
	suite.Equal(0, len(suite.appByNameData))
	suite.Equal(0, len(suite.oauthAppData))
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateApplicationCache_WithAllValues() {
	app := suite.createTestAppWithOAuth()
	clientID := app.InboundAuthConfig[0].OAuthAppConfig.ClientID

	// Pre-cache the data
	_ = suite.appByIDCache.Set(cache.CacheKey{Key: app.ID}, app)
	_ = suite.appByNameCache.Set(cache.CacheKey{Key: app.Name}, app)
	_ = suite.oauthAppCache.Set(cache.CacheKey{Key: clientID}, app.InboundAuthConfig[0].OAuthAppConfig)

	suite.cachedStore.invalidateApplicationCache(app.ID, app.Name, clientID)

	// Verify all caches are invalidated
	_, ok := suite.appByIDCache.Get(cache.CacheKey{Key: app.ID})
	suite.False(ok)
	_, ok = suite.appByNameCache.Get(cache.CacheKey{Key: app.Name})
	suite.False(ok)
	_, ok = suite.oauthAppCache.Get(cache.CacheKey{Key: clientID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateApplicationCache_OnlyAppID() {
	app := suite.createTestApp()
	_ = suite.appByIDCache.Set(cache.CacheKey{Key: app.ID}, app)

	suite.cachedStore.invalidateApplicationCache(app.ID, "", "")

	// Verify only ID cache is invalidated
	_, ok := suite.appByIDCache.Get(cache.CacheKey{Key: app.ID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateApplicationCache_OnlyAppName() {
	app := suite.createTestApp()
	_ = suite.appByNameCache.Set(cache.CacheKey{Key: app.Name}, app)

	suite.cachedStore.invalidateApplicationCache("", app.Name, "")

	// Verify only name cache is invalidated
	_, ok := suite.appByNameCache.Get(cache.CacheKey{Key: app.Name})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateApplicationCache_OnlyClientID() {
	oauthApp := suite.createTestOAuthApp()
	_ = suite.oauthAppCache.Set(cache.CacheKey{Key: oauthApp.ClientID}, oauthApp)

	suite.cachedStore.invalidateApplicationCache("", "", oauthApp.ClientID)

	// Verify only OAuth cache is invalidated
	_, ok := suite.oauthAppCache.Get(cache.CacheKey{Key: oauthApp.ClientID})
	suite.False(ok)
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateApplicationCache_CacheDeleteErrors() {
	app := suite.createTestApp()
	_ = suite.appByIDCache.Set(cache.CacheKey{Key: app.ID}, app)
	_ = suite.appByNameCache.Set(cache.CacheKey{Key: app.Name}, app)
	cacheDelErr := errors.New("cache delete error")

	// Override Delete expectation to return error
	suite.appByIDCache.EXPECT().Delete(mock.Anything).Return(cacheDelErr).Maybe()
	suite.appByNameCache.EXPECT().Delete(mock.Anything).Return(cacheDelErr).Maybe()

	// Should not panic or fail
	suite.cachedStore.invalidateApplicationCache(app.ID, app.Name, "")
}

func (suite *CacheBackedStoreTestSuite) TestInvalidateApplicationCache_OAuthCacheDeleteError() {
	oauthApp := suite.createTestOAuthApp()
	_ = suite.oauthAppCache.Set(cache.CacheKey{Key: oauthApp.ClientID}, oauthApp)
	cacheDelErr := errors.New("cache delete error")

	// Override Delete expectation to return error
	suite.oauthAppCache.EXPECT().Delete(mock.Anything).Return(cacheDelErr).Maybe()

	// Should not panic or fail
	suite.cachedStore.invalidateApplicationCache("", "", oauthApp.ClientID)
}

func (suite *CacheBackedStoreTestSuite) TestIsApplicationExists_Success() {
	appID := "test-app-123"

	suite.mockStore.EXPECT().IsApplicationExists(appID).Return(true, nil).Once()

	exists, err := suite.cachedStore.IsApplicationExists(appID)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *CacheBackedStoreTestSuite) TestIsApplicationExists_NotFound() {
	appID := "non-existent-app"

	suite.mockStore.EXPECT().IsApplicationExists(appID).Return(false, nil).Once()

	exists, err := suite.cachedStore.IsApplicationExists(appID)

	suite.NoError(err)
	suite.False(exists)
}

func (suite *CacheBackedStoreTestSuite) TestIsApplicationExists_Error() {
	appID := "error-app"
	expectedErr := errors.New("database error")

	suite.mockStore.EXPECT().IsApplicationExists(appID).Return(false, expectedErr).Once()

	exists, err := suite.cachedStore.IsApplicationExists(appID)

	suite.Error(err)
	suite.False(exists)
	suite.Equal(expectedErr, err)
}

func (suite *CacheBackedStoreTestSuite) TestIsApplicationExistsByName_Success() {
	appName := "Test Application"

	suite.mockStore.EXPECT().IsApplicationExistsByName(appName).Return(true, nil).Once()

	exists, err := suite.cachedStore.IsApplicationExistsByName(appName)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *CacheBackedStoreTestSuite) TestIsApplicationExistsByName_NotFound() {
	appName := "Non-Existent App"

	suite.mockStore.EXPECT().IsApplicationExistsByName(appName).Return(false, nil).Once()

	exists, err := suite.cachedStore.IsApplicationExistsByName(appName)

	suite.NoError(err)
	suite.False(exists)
}

func (suite *CacheBackedStoreTestSuite) TestIsApplicationExistsByName_Error() {
	appName := "Error App"
	expectedErr := errors.New("database connection error")

	suite.mockStore.EXPECT().IsApplicationExistsByName(appName).Return(false, expectedErr).Once()

	exists, err := suite.cachedStore.IsApplicationExistsByName(appName)

	suite.Error(err)
	suite.False(exists)
	suite.Equal(expectedErr, err)
}
