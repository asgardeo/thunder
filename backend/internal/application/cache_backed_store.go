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

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/log"
)

// CachedBackedApplicationStore is the implementation of ApplicationStoreInterface that uses caching.
type cachedBackedApplicationStore struct {
	AppByIDCache   cache.CacheInterface[*model.ApplicationProcessedDTO]
	AppByNameCache cache.CacheInterface[*model.ApplicationProcessedDTO]
	OAuthAppCache  cache.CacheInterface[*model.OAuthAppConfigProcessedDTO]
	Store          applicationStoreInterface
}

// newCachedBackedApplicationStore creates a new instance of CachedBackedApplicationStore.
func newCachedBackedApplicationStore(store applicationStoreInterface) applicationStoreInterface {
	return &cachedBackedApplicationStore{
		AppByIDCache:   cache.GetCache[*model.ApplicationProcessedDTO]("ApplicationByIDCache"),
		AppByNameCache: cache.GetCache[*model.ApplicationProcessedDTO]("ApplicationByNameCache"),
		OAuthAppCache:  cache.GetCache[*model.OAuthAppConfigProcessedDTO]("OAuthAppCache"),
		Store:          store,
	}
}

// CreateApplication creates a new application and caches it.
func (as *cachedBackedApplicationStore) CreateApplication(app model.ApplicationProcessedDTO) error {
	if err := as.Store.CreateApplication(app); err != nil {
		return err
	}
	as.cacheApplication(&app)
	return nil
}

// GetTotalApplicationCount returns the total count of applications.
func (as *cachedBackedApplicationStore) GetTotalApplicationCount() (int, error) {
	return as.Store.GetTotalApplicationCount()
}

// GetApplicationList returns a list of basic application DTOs.
func (as *cachedBackedApplicationStore) GetApplicationList() ([]model.BasicApplicationDTO, error) {
	return as.Store.GetApplicationList()
}

// GetOAuthApplication retrieves an OAuth application by client ID, using cache if available.
func (as *cachedBackedApplicationStore) GetOAuthApplication(clientID string) (
	*model.OAuthAppConfigProcessedDTO, error) {
	cacheKey := cache.CacheKey{
		Key: clientID,
	}
	cachedApp, ok := as.OAuthAppCache.Get(cacheKey)
	if ok {
		return cachedApp, nil
	}

	oauthApp, err := as.Store.GetOAuthApplication(clientID)
	if err != nil || oauthApp == nil {
		return oauthApp, err
	}
	as.cacheOAuthApplication(oauthApp)

	return oauthApp, nil
}

// GetApplicationByID retrieves an application by ID, using cache if available.
func (as *cachedBackedApplicationStore) GetApplicationByID(id string) (*model.ApplicationProcessedDTO, error) {
	cacheKey := cache.CacheKey{
		Key: id,
	}
	cachedApp, ok := as.AppByIDCache.Get(cacheKey)
	if ok {
		return cachedApp, nil
	}

	app, err := as.Store.GetApplicationByID(id)
	if err != nil || app == nil {
		return app, err
	}
	as.cacheApplication(app)

	return app, nil
}

// GetApplicationByName retrieves an application by name, using cache if available.
func (as *cachedBackedApplicationStore) GetApplicationByName(name string) (*model.ApplicationProcessedDTO, error) {
	cacheKey := cache.CacheKey{
		Key: name,
	}
	cachedApp, ok := as.AppByNameCache.Get(cacheKey)
	if ok {
		return cachedApp, nil
	}

	app, err := as.Store.GetApplicationByName(name)
	if err != nil || app == nil {
		return app, err
	}
	as.cacheApplication(app)

	return app, nil
}

// UpdateApplication updates an existing application and caches the updated version.
func (as *cachedBackedApplicationStore) UpdateApplication(existingApp,
	updatedApp *model.ApplicationProcessedDTO) error {
	if err := as.Store.UpdateApplication(existingApp, updatedApp); err != nil {
		return err
	}

	existingAppClientID := ""
	if len(existingApp.InboundAuthConfig) > 0 && existingApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		existingAppClientID = existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}
	as.invalidateApplicationCache(existingApp.ID, existingApp.Name, existingAppClientID)

	as.cacheApplication(updatedApp)

	if len(updatedApp.InboundAuthConfig) > 0 && updatedApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		as.cacheOAuthApplication(updatedApp.InboundAuthConfig[0].OAuthAppConfig)
	}

	return nil
}

// DeleteApplication deletes an application by ID and invalidates the caches.
func (as *cachedBackedApplicationStore) DeleteApplication(id string) error {
	cacheKey := cache.CacheKey{
		Key: id,
	}
	existingApp, ok := as.AppByIDCache.Get(cacheKey)
	if !ok {
		var err error
		existingApp, err = as.Store.GetApplicationByID(id)
		if err != nil {
			if errors.Is(err, model.ApplicationNotFoundError) {
				return nil
			}
			return err
		}
	}
	if existingApp == nil {
		return nil
	}

	if err := as.Store.DeleteApplication(id); err != nil {
		return err
	}

	existingAppClientID := ""
	if len(existingApp.InboundAuthConfig) > 0 && existingApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		existingAppClientID = existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}
	as.invalidateApplicationCache(id, existingApp.Name, existingAppClientID)

	return nil
}

// cacheApplication caches the application and OAuth application configuration if it exists.
func (as *cachedBackedApplicationStore) cacheApplication(app *model.ApplicationProcessedDTO) {
	if app == nil {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))

	if app.ID != "" {
		appByIDCacheKey := cache.CacheKey{
			Key: app.ID,
		}
		if err := as.AppByIDCache.Set(appByIDCacheKey, app); err != nil {
			logger.Error("Failed to cache application by ID", log.String("appID", app.ID), log.Error(err))
		}
	}
	if app.Name != "" {
		appByNameCacheKey := cache.CacheKey{
			Key: app.Name,
		}
		if err := as.AppByNameCache.Set(appByNameCacheKey, app); err != nil {
			logger.Error("Failed to cache application by name", log.String("appName", app.Name),
				log.Error(err))
		}
	}

	// Cache the OAuth application configuration if it exists.
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil {
		as.cacheOAuthApplication(app.InboundAuthConfig[0].OAuthAppConfig)
	}
}

// cacheOAuthApplication caches the OAuth application configuration if it exists.
func (as *cachedBackedApplicationStore) cacheOAuthApplication(oAuthAppConfig *model.OAuthAppConfigProcessedDTO) {
	if oAuthAppConfig == nil || oAuthAppConfig.ClientID == "" {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))

	oauthCacheKey := cache.CacheKey{
		Key: oAuthAppConfig.ClientID,
	}
	if err := as.OAuthAppCache.Set(oauthCacheKey, oAuthAppConfig); err != nil {
		logger.Error("Failed to cache OAuth application", log.String("clientID", oAuthAppConfig.ClientID),
			log.Error(err))
	}
}

// InvalidateApplicationCache invalidates all application caches.
func (as *cachedBackedApplicationStore) invalidateApplicationCache(appID, appName, clientID string) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))

	if appID != "" {
		cacheKey := cache.CacheKey{
			Key: appID,
		}
		err := as.AppByIDCache.Delete(cacheKey)
		if err != nil {
			logger.Error("Failed to delete application cache by ID", log.String("appID", appID), log.Error(err))
		}
	}
	if appName != "" {
		cacheKey := cache.CacheKey{
			Key: appName,
		}
		err := as.AppByNameCache.Delete(cacheKey)
		if err != nil {
			logger.Error("Failed to delete application cache by name", log.String("appName", appName), log.Error(err))
		}
	}
	if clientID != "" {
		oauthCacheKey := cache.CacheKey{
			Key: clientID,
		}
		err := as.OAuthAppCache.Delete(oauthCacheKey)
		if err != nil {
			logger.Error("Failed to delete OAuth application cache", log.String("clientID", clientID), log.Error(err))
		}
	}
}
