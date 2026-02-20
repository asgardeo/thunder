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

package idp

import (
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
)

// compositeIDPStore implements a composite store that combines file-based (immutable) and database (mutable) stores.
// - Read operations query both stores and merge results
// - Write operations (Create/Update/Delete) only affect the database store
// - Declarative IDPs (from YAML files) cannot be modified or deleted
type compositeIDPStore struct {
	fileStore idpStoreInterface
	dbStore   idpStoreInterface
}

// newCompositeIDPStore creates a new composite store with both file-based and database stores.
func newCompositeIDPStore(fileStore, dbStore idpStoreInterface) *compositeIDPStore {
	return &compositeIDPStore{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

// CreateIdentityProvider creates a new identity provider in the database store only.
func (c *compositeIDPStore) CreateIdentityProvider(idp IDPDTO) error {
	return c.dbStore.CreateIdentityProvider(idp)
}

// GetIdentityProviderList retrieves identity providers from both stores and merges the results.
// Database IDPs are marked as mutable (IsReadOnly=false), file-based IDPs as immutable (IsReadOnly=true).
func (c *compositeIDPStore) GetIdentityProviderList() ([]BasicIDPDTO, error) {
	dbIDPs, err := c.dbStore.GetIdentityProviderList()
	if err != nil {
		return nil, err
	}

	fileIDPs, err := c.fileStore.GetIdentityProviderList()
	if err != nil {
		return nil, err
	}

	return mergeAndDeduplicateIDPs(dbIDPs, fileIDPs), nil
}

// GetIdentityProvider retrieves an identity provider by ID from either store.
// Checks database store first, then falls back to file store.
func (c *compositeIDPStore) GetIdentityProvider(idpID string) (*IDPDTO, error) {
	return declarativeresource.CompositeGetHelper(
		func() (*IDPDTO, error) { return c.dbStore.GetIdentityProvider(idpID) },
		func() (*IDPDTO, error) { return c.fileStore.GetIdentityProvider(idpID) },
		ErrIDPNotFound,
	)
}

// GetIdentityProviderByName retrieves an identity provider by name from either store.
// Checks database store first, then falls back to file store.
func (c *compositeIDPStore) GetIdentityProviderByName(idpName string) (*IDPDTO, error) {
	return declarativeresource.CompositeGetHelper(
		func() (*IDPDTO, error) { return c.dbStore.GetIdentityProviderByName(idpName) },
		func() (*IDPDTO, error) { return c.fileStore.GetIdentityProviderByName(idpName) },
		ErrIDPNotFound,
	)
}

// UpdateIdentityProvider updates an identity provider in the database store only.
// Immutability checks are handled at the service layer.
func (c *compositeIDPStore) UpdateIdentityProvider(idp *IDPDTO) error {
	return c.dbStore.UpdateIdentityProvider(idp)
}

// DeleteIdentityProvider deletes an identity provider from the database store only.
// Immutability checks are handled at the service layer.
func (c *compositeIDPStore) DeleteIdentityProvider(idpID string) error {
	return c.dbStore.DeleteIdentityProvider(idpID)
}

// mergeAndDeduplicateIDPs merges IDPs from both stores and removes duplicates by ID.
// Database IDPs are marked as mutable (IsReadOnly=false), file-based IDPs as immutable (IsReadOnly=true).
// While duplicates shouldn't exist by design, this provides defensive programming.
func mergeAndDeduplicateIDPs(dbIDPs, fileIDPs []BasicIDPDTO) []BasicIDPDTO {
	seen := make(map[string]bool)
	result := make([]BasicIDPDTO, 0, len(dbIDPs)+len(fileIDPs))

	// Add DB IDPs first (they take precedence) - mark as mutable (IsReadOnly=false)
	for i := range dbIDPs {
		if !seen[dbIDPs[i].ID] {
			seen[dbIDPs[i].ID] = true
			dbIDPs[i].IsReadOnly = false
			result = append(result, dbIDPs[i])
		}
	}

	// Add file IDPs if not already present - mark as immutable (IsReadOnly=true)
	for i := range fileIDPs {
		if !seen[fileIDPs[i].ID] {
			seen[fileIDPs[i].ID] = true
			fileIDPs[i].IsReadOnly = true
			result = append(result, fileIDPs[i])
		}
	}

	return result
}
