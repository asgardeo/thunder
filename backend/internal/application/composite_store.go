/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

	"github.com/asgardeo/thunder/internal/application/model"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
)

// compositeApplicationStore combines file-based (immutable) and database (mutable) stores.
type compositeApplicationStore struct {
	fileStore applicationStoreInterface
	dbStore   applicationStoreInterface
}

// newCompositeApplicationStore creates a new composite store.
func newCompositeApplicationStore(fileStore, dbStore applicationStoreInterface) *compositeApplicationStore {
	return &compositeApplicationStore{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

func (c *compositeApplicationStore) GetTotalApplicationCount(ctx context.Context) (int, error) {
	return declarativeresource.CompositeMergeCountHelper(
		func() (int, error) { return c.dbStore.GetTotalApplicationCount(ctx) },
		func() (int, error) { return c.fileStore.GetTotalApplicationCount(ctx) },
	)
}

func (c *compositeApplicationStore) GetApplicationList(ctx context.Context) ([]applicationConfigDAO, error) {
	apps, limitExceeded, err := declarativeresource.CompositeMergeListHelperWithLimit(
		func() (int, error) { return c.dbStore.GetTotalApplicationCount(ctx) },
		func() (int, error) { return c.fileStore.GetTotalApplicationCount(ctx) },
		func(limit int) ([]applicationConfigDAO, error) { return c.dbStore.GetApplicationList(ctx) },
		func(limit int) ([]applicationConfigDAO, error) { return c.fileStore.GetApplicationList(ctx) },
		mergeAndDeduplicateAppConfigs,
		100, 0,
		serverconst.MaxCompositeStoreRecords,
	)
	if err != nil {
		return nil, err
	}
	if limitExceeded {
		return nil, errResultLimitExceededInCompositeMode
	}
	return apps, nil
}

func (c *compositeApplicationStore) CreateApplication(ctx context.Context, app applicationConfigDAO) error {
	return c.dbStore.CreateApplication(ctx, app)
}

func (c *compositeApplicationStore) CreateOAuthConfig(ctx context.Context, entityID string,
	oauthConfigJSON json.RawMessage) error {
	return c.dbStore.CreateOAuthConfig(ctx, entityID, oauthConfigJSON)
}

func (c *compositeApplicationStore) GetApplicationByID(
	ctx context.Context, id string) (*applicationConfigDAO, error) {
	return declarativeresource.CompositeGetHelper(
		func() (*applicationConfigDAO, error) { return c.dbStore.GetApplicationByID(ctx, id) },
		func() (*applicationConfigDAO, error) { return c.fileStore.GetApplicationByID(ctx, id) },
		model.ApplicationNotFoundError,
	)
}

func (c *compositeApplicationStore) GetOAuthConfigByEntityID(
	ctx context.Context, entityID string) (*oauthConfigDAO, error) {
	return declarativeresource.CompositeGetHelper(
		func() (*oauthConfigDAO, error) { return c.dbStore.GetOAuthConfigByEntityID(ctx, entityID) },
		func() (*oauthConfigDAO, error) { return c.fileStore.GetOAuthConfigByEntityID(ctx, entityID) },
		model.ApplicationNotFoundError,
	)
}

func (c *compositeApplicationStore) UpdateApplication(ctx context.Context, app applicationConfigDAO) error {
	return c.dbStore.UpdateApplication(ctx, app)
}

func (c *compositeApplicationStore) UpdateOAuthConfig(ctx context.Context, entityID string,
	oauthConfigJSON json.RawMessage) error {
	return c.dbStore.UpdateOAuthConfig(ctx, entityID, oauthConfigJSON)
}

func (c *compositeApplicationStore) DeleteApplication(ctx context.Context, id string) error {
	return c.dbStore.DeleteApplication(ctx, id)
}

func (c *compositeApplicationStore) DeleteOAuthConfig(ctx context.Context, entityID string) error {
	return c.dbStore.DeleteOAuthConfig(ctx, entityID)
}

func (c *compositeApplicationStore) IsApplicationExists(ctx context.Context, id string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.IsApplicationExists(ctx, id) },
		func() (bool, error) { return c.dbStore.IsApplicationExists(ctx, id) },
	)
}

func (c *compositeApplicationStore) IsApplicationDeclarative(ctx context.Context, id string) bool {
	return declarativeresource.CompositeIsDeclarativeHelper(
		id,
		func(id string) (bool, error) { return c.fileStore.IsApplicationExists(ctx, id) },
	)
}

// mergeAndDeduplicateAppConfigs merges configs from both stores, deduplicating by ID.
func mergeAndDeduplicateAppConfigs(dbApps, fileApps []applicationConfigDAO) []applicationConfigDAO {
	seen := make(map[string]bool)
	result := make([]applicationConfigDAO, 0, len(dbApps)+len(fileApps))

	for i := range dbApps {
		if !seen[dbApps[i].ID] {
			seen[dbApps[i].ID] = true
			app := dbApps[i]
			app.IsReadOnly = false
			result = append(result, app)
		}
	}

	for i := range fileApps {
		if !seen[fileApps[i].ID] {
			seen[fileApps[i].ID] = true
			app := fileApps[i]
			app.IsReadOnly = true
			result = append(result, app)
		}
	}

	return result
}
