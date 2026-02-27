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
	"github.com/asgardeo/thunder/internal/application/model"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
)

// compositeApplicationStore implements a composite store that combines file-based (immutable) and
// database (mutable) stores.
// - Read operations query both stores and merge results
// - Write operations (Create/Update/Delete) only affect the database store
// - Declarative applications (from YAML files) cannot be modified or deleted
type compositeApplicationStore struct {
	fileStore applicationStoreInterface
	dbStore   applicationStoreInterface
}

// newCompositeApplicationStore creates a new composite store with both file-based and database stores.
func newCompositeApplicationStore(fileStore, dbStore applicationStoreInterface) *compositeApplicationStore {
	return &compositeApplicationStore{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

// GetTotalApplicationCount retrieves the total count of applications from both stores.
func (c *compositeApplicationStore) GetTotalApplicationCount() (int, error) {
	return declarativeresource.CompositeMergeCountHelper(
		func() (int, error) { return c.dbStore.GetTotalApplicationCount() },
		func() (int, error) { return c.fileStore.GetTotalApplicationCount() },
	)
}

// GetApplicationList retrieves applications from both stores.
// Note: Application list does not support pagination at the API level, so we don't apply pagination here.
// However, we still apply the 1000-record limit in composite mode to prevent memory exhaustion.
func (c *compositeApplicationStore) GetApplicationList() ([]model.BasicApplicationDTO, error) {
	// Use the helper to fetch, merge, and check limits
	// Since application list doesn't support pagination, we use a fixed limit of 100 and offset=0
	apps, limitExceeded, err := declarativeresource.CompositeMergeListHelperWithLimit(
		func() (int, error) { return c.dbStore.GetTotalApplicationCount() },
		func() (int, error) { return c.fileStore.GetTotalApplicationCount() },
		func(limit int) ([]model.BasicApplicationDTO, error) { return c.dbStore.GetApplicationList() },
		func(limit int) ([]model.BasicApplicationDTO, error) { return c.fileStore.GetApplicationList() },
		mergeAndDeduplicateApplications,
		100, // Setting limit to 100 as pagination is not supported at API level.
		0,   // offset 0 - start from beginning
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

// CreateApplication creates a new application in the database store only.
// Conflict checking is handled at the service layer.
func (c *compositeApplicationStore) CreateApplication(app model.ApplicationProcessedDTO) error {
	return c.dbStore.CreateApplication(app)
}

// GetApplicationByID retrieves an application by ID from either store.
// Checks database store first, then falls back to file store.
func (c *compositeApplicationStore) GetApplicationByID(id string) (*model.ApplicationProcessedDTO, error) {
	app, err := declarativeresource.CompositeGetHelper(
		func() (*model.ApplicationProcessedDTO, error) { return c.dbStore.GetApplicationByID(id) },
		func() (*model.ApplicationProcessedDTO, error) { return c.fileStore.GetApplicationByID(id) },
		model.ApplicationNotFoundError,
	)
	return app, err
}

// GetApplicationByName retrieves an application by name from either store.
// Checks database store first, then falls back to file store.
func (c *compositeApplicationStore) GetApplicationByName(name string) (*model.ApplicationProcessedDTO, error) {
	app, err := declarativeresource.CompositeGetHelper(
		func() (*model.ApplicationProcessedDTO, error) { return c.dbStore.GetApplicationByName(name) },
		func() (*model.ApplicationProcessedDTO, error) { return c.fileStore.GetApplicationByName(name) },
		model.ApplicationNotFoundError,
	)
	return app, err
}

// GetOAuthApplication retrieves an OAuth application by client ID from either store.
// Checks database store first, then falls back to file store.
func (c *compositeApplicationStore) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO, error) {
	config, err := declarativeresource.CompositeGetHelper(
		func() (*model.OAuthAppConfigProcessedDTO, error) { return c.dbStore.GetOAuthApplication(clientID) },
		func() (*model.OAuthAppConfigProcessedDTO, error) { return c.fileStore.GetOAuthApplication(clientID) },
		model.ApplicationNotFoundError,
	)
	return config, err
}

// UpdateApplication updates an application in the database store only.
// Immutability checks are handled at the service layer.
func (c *compositeApplicationStore) UpdateApplication(
	existingApp, updatedApp *model.ApplicationProcessedDTO,
) error {
	return c.dbStore.UpdateApplication(existingApp, updatedApp)
}

// DeleteApplication deletes an application from the database store only.
// Immutability checks are handled at the service layer.
func (c *compositeApplicationStore) DeleteApplication(id string) error {
	return c.dbStore.DeleteApplication(id)
}

// IsApplicationExists checks if an application exists in either store.
// Checks file store first to prevent conflicts with immutable resources.
func (c *compositeApplicationStore) IsApplicationExists(id string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.IsApplicationExists(id) },
		func() (bool, error) { return c.dbStore.IsApplicationExists(id) },
	)
}

// IsApplicationExistsByName checks if an application with the given name exists in either store.
// Checks file store first to prevent conflicts with immutable resources.
func (c *compositeApplicationStore) IsApplicationExistsByName(name string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.IsApplicationExistsByName(name) },
		func() (bool, error) { return c.dbStore.IsApplicationExistsByName(name) },
	)
}

// IsApplicationDeclarative checks if an application is immutable (exists in file store).
func (c *compositeApplicationStore) IsApplicationDeclarative(id string) bool {
	return declarativeresource.CompositeIsDeclarativeHelper(
		id,
		func(id string) (bool, error) { return c.fileStore.IsApplicationExists(id) },
	)
}

// mergeAndDeduplicateApplications merges applications from both stores and removes duplicates by ID.
// While duplicates shouldn't exist by design (an app exists in only one store), this provides
// defensive programming against misconfigurations or bugs.
func mergeAndDeduplicateApplications(dbApps, fileApps []model.BasicApplicationDTO) []model.BasicApplicationDTO {
	seen := make(map[string]bool)
	result := make([]model.BasicApplicationDTO, 0, len(dbApps)+len(fileApps))

	// Add DB apps first (they take precedence) - mark as mutable (isReadOnly=false)
	for i := range dbApps {
		if !seen[dbApps[i].ID] {
			seen[dbApps[i].ID] = true
			appCopy := dbApps[i]
			appCopy.IsReadOnly = false
			result = append(result, appCopy)
		}
	}

	// Add file apps if not already present - mark as immutable (isReadOnly=true)
	for i := range fileApps {
		if !seen[fileApps[i].ID] {
			seen[fileApps[i].ID] = true
			appCopy := fileApps[i]
			appCopy.IsReadOnly = true
			result = append(result, appCopy)
		}
	}

	return result
}
