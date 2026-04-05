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

	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/log"
)

// cachedBackedApplicationStore wraps an applicationStoreInterface with caching.
type cachedBackedApplicationStore struct {
	AppByIDCache  cache.CacheInterface[*applicationConfigDAO]
	OAuthAppCache cache.CacheInterface[*oauthConfigDAO]
	Store         applicationStoreInterface
}

// newCachedBackedApplicationStore creates a new cache-backed application store.
func newCachedBackedApplicationStore(store applicationStoreInterface) applicationStoreInterface {
	return &cachedBackedApplicationStore{
		AppByIDCache:  cache.GetCache[*applicationConfigDAO]("ApplicationByIDCache"),
		OAuthAppCache: cache.GetCache[*oauthConfigDAO]("OAuthAppByEntityIDCache"),
		Store:         store,
	}
}

func (as *cachedBackedApplicationStore) CreateApplication(ctx context.Context, app applicationConfigDAO) error {
	if err := as.Store.CreateApplication(ctx, app); err != nil {
		return err
	}
	as.cacheAppConfig(ctx, &app)
	return nil
}

func (as *cachedBackedApplicationStore) CreateOAuthConfig(ctx context.Context, entityID string,
	oauthConfigJSON json.RawMessage) error {
	return as.Store.CreateOAuthConfig(ctx, entityID, oauthConfigJSON)
}

func (as *cachedBackedApplicationStore) GetApplicationByID(ctx context.Context,
	id string) (*applicationConfigDAO, error) {
	cacheKey := cache.CacheKey{Key: id}
	if cached, ok := as.AppByIDCache.Get(ctx, cacheKey); ok {
		return cached, nil
	}

	app, err := as.Store.GetApplicationByID(ctx, id)
	if err != nil || app == nil {
		return app, err
	}
	as.cacheAppConfig(ctx, app)
	return app, nil
}

func (as *cachedBackedApplicationStore) GetOAuthConfigByEntityID(ctx context.Context,
	entityID string) (*oauthConfigDAO, error) {
	cacheKey := cache.CacheKey{Key: entityID}
	if cached, ok := as.OAuthAppCache.Get(ctx, cacheKey); ok {
		return cached, nil
	}

	cfg, err := as.Store.GetOAuthConfigByEntityID(ctx, entityID)
	if err != nil || cfg == nil {
		return cfg, err
	}
	as.cacheOAuthConfig(ctx, cfg)
	return cfg, nil
}

func (as *cachedBackedApplicationStore) GetApplicationList(ctx context.Context) ([]applicationConfigDAO, error) {
	return as.Store.GetApplicationList(ctx)
}

func (as *cachedBackedApplicationStore) GetTotalApplicationCount(ctx context.Context) (int, error) {
	return as.Store.GetTotalApplicationCount(ctx)
}

func (as *cachedBackedApplicationStore) UpdateApplication(ctx context.Context, app applicationConfigDAO) error {
	if err := as.Store.UpdateApplication(ctx, app); err != nil {
		return err
	}
	as.invalidateAppCache(ctx, app.ID)
	as.cacheAppConfig(ctx, &app)
	return nil
}

func (as *cachedBackedApplicationStore) UpdateOAuthConfig(ctx context.Context, entityID string,
	oauthConfigJSON json.RawMessage) error {
	if err := as.Store.UpdateOAuthConfig(ctx, entityID, oauthConfigJSON); err != nil {
		return err
	}
	as.invalidateOAuthCache(ctx, entityID)
	return nil
}

func (as *cachedBackedApplicationStore) DeleteApplication(ctx context.Context, id string) error {
	if err := as.Store.DeleteApplication(ctx, id); err != nil {
		return err
	}
	as.invalidateAppCache(ctx, id)
	as.invalidateOAuthCache(ctx, id)
	return nil
}

func (as *cachedBackedApplicationStore) DeleteOAuthConfig(ctx context.Context, entityID string) error {
	if err := as.Store.DeleteOAuthConfig(ctx, entityID); err != nil {
		return err
	}
	as.invalidateOAuthCache(ctx, entityID)
	return nil
}

func (as *cachedBackedApplicationStore) IsApplicationExists(ctx context.Context, id string) (bool, error) {
	return as.Store.IsApplicationExists(ctx, id)
}

func (as *cachedBackedApplicationStore) IsApplicationDeclarative(ctx context.Context, id string) bool {
	return as.Store.IsApplicationDeclarative(ctx, id)
}

// --- Cache helpers ---

func (as *cachedBackedApplicationStore) cacheAppConfig(ctx context.Context, app *applicationConfigDAO) {
	if app == nil || app.ID == "" {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))
	if err := as.AppByIDCache.Set(ctx, cache.CacheKey{Key: app.ID}, app); err != nil {
		logger.Error("Failed to cache app config", log.String("appID", app.ID), log.Error(err))
	}
}

func (as *cachedBackedApplicationStore) cacheOAuthConfig(ctx context.Context, cfg *oauthConfigDAO) {
	if cfg == nil || cfg.EntityID == "" {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))
	if err := as.OAuthAppCache.Set(ctx, cache.CacheKey{Key: cfg.EntityID}, cfg); err != nil {
		logger.Error("Failed to cache OAuth config", log.String("entityID", cfg.EntityID), log.Error(err))
	}
}

func (as *cachedBackedApplicationStore) invalidateAppCache(ctx context.Context, appID string) {
	if appID == "" {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))
	if err := as.AppByIDCache.Delete(ctx, cache.CacheKey{Key: appID}); err != nil {
		logger.Error("Failed to invalidate app cache", log.String("appID", appID), log.Error(err))
	}
}

func (as *cachedBackedApplicationStore) invalidateOAuthCache(ctx context.Context, entityID string) {
	if entityID == "" {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))
	if err := as.OAuthAppCache.Delete(ctx, cache.CacheKey{Key: entityID}); err != nil {
		logger.Error("Failed to invalidate OAuth cache", log.String("entityID", entityID), log.Error(err))
	}
}
