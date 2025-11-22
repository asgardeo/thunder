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
	"errors"

	"github.com/asgardeo/thunder/internal/system/file_based_runtime/entity"
	"github.com/asgardeo/thunder/internal/system/log"
)

type idpFileBasedStore struct {
	storage entity.StoreInterface
}

// CreateIdentityProvider implements idpStoreInterface.
func (f *idpFileBasedStore) CreateIdentityProvider(idp IDPDTO) error {
	idpKey := entity.NewCompositeKey(idp.ID, entity.KeyTypeIDP)
	return f.storage.Set(idpKey, &idp)
}

// DeleteIdentityProvider implements idpStoreInterface.
func (f *idpFileBasedStore) DeleteIdentityProvider(id string) error {
	return errors.New("DeleteIdentityProvider is not supported in file-based store")
}

// GetIdentityProvider implements idpStoreInterface.
func (f *idpFileBasedStore) GetIdentityProvider(idpID string) (*IDPDTO, error) {
	entity, err := f.storage.Get(entity.NewCompositeKey(idpID, entity.KeyTypeIDP))
	if err != nil {
		return nil, ErrIDPNotFound
	}
	idp, ok := entity.Data.(*IDPDTO)
	if !ok {
		log.GetLogger().Error("Type assertion failed while retrieving identity provider by ID",
			log.String("idpID", idpID))
		return nil, errors.New("identity provider data corrupted")
	}
	return idp, nil
}

// GetIdentityProviderByName implements idpStoreInterface.
func (f *idpFileBasedStore) GetIdentityProviderByName(idpName string) (*IDPDTO, error) {
	list, err := f.storage.ListByType(entity.KeyTypeIDP)
	if err != nil {
		return nil, err
	}

	for _, item := range list {
		if idp, ok := item.Data.(*IDPDTO); ok && idp.Name == idpName {
			return idp, nil
		}
	}

	return nil, ErrIDPNotFound
}

// GetIdentityProviderList implements idpStoreInterface.
func (f *idpFileBasedStore) GetIdentityProviderList() ([]BasicIDPDTO, error) {
	list, err := f.storage.ListByType(entity.KeyTypeIDP)
	if err != nil {
		return nil, err
	}

	var idpList []BasicIDPDTO
	for _, item := range list {
		if idp, ok := item.Data.(*IDPDTO); ok {
			basicIDP := BasicIDPDTO{
				ID:          idp.ID,
				Name:        idp.Name,
				Description: idp.Description,
				Type:        idp.Type,
			}
			idpList = append(idpList, basicIDP)
		}
	}
	return idpList, nil
}

// UpdateIdentityProvider implements idpStoreInterface.
func (f *idpFileBasedStore) UpdateIdentityProvider(idp *IDPDTO) error {
	return errors.New("UpdateIdentityProvider is not supported in file-based store")
}

// newIDPFileBasedStore creates a new instance of a file-based store.
func newIDPFileBasedStore() idpStoreInterface {
	store := entity.GetInstance()
	return &idpFileBasedStore{
		storage: store,
	}
}
