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
	"context"
	"errors"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

type idpFileBasedStore struct {
	*declarativeresource.GenericFileBasedStore
}

// Create implements declarativeresource.Storer interface for resource loader
func (f *idpFileBasedStore) Create(id string, data interface{}) error {
	idp := data.(*IDPDTO)
	return f.CreateIdentityProvider(context.Background(), *idp)
}

// CreateIdentityProvider implements idpStoreInterface.
func (f *idpFileBasedStore) CreateIdentityProvider(ctx context.Context, idp IDPDTO) error {
	return f.GenericFileBasedStore.Create(idp.ID, &idp)
}

// DeleteIdentityProvider implements idpStoreInterface.
func (f *idpFileBasedStore) DeleteIdentityProvider(ctx context.Context, id string) error {
	return errors.New("DeleteIdentityProvider is not supported in file-based store")
}

// GetIdentityProvider implements idpStoreInterface.
func (f *idpFileBasedStore) GetIdentityProvider(ctx context.Context, idpID string) (*IDPDTO, error) {
	data, err := f.GenericFileBasedStore.Get(idpID)
	if err != nil {
		return nil, ErrIDPNotFound
	}
	idp, ok := data.(*IDPDTO)
	if !ok {
		declarativeresource.LogTypeAssertionError("identity provider", idpID)
		return nil, errors.New("identity provider data corrupted")
	}
	return idp, nil
}

// GetIdentityProviderByName implements idpStoreInterface.
func (f *idpFileBasedStore) GetIdentityProviderByName(ctx context.Context, idpName string) (*IDPDTO, error) {
	data, err := f.GenericFileBasedStore.GetByField(idpName, func(d interface{}) string {
		return d.(*IDPDTO).Name
	})
	if err != nil {
		return nil, ErrIDPNotFound
	}
	return data.(*IDPDTO), nil
}

// GetIdentityProviderList implements idpStoreInterface.
func (f *idpFileBasedStore) GetIdentityProviderList(ctx context.Context) ([]BasicIDPDTO, error) {
	list, err := f.GenericFileBasedStore.List()
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
func (f *idpFileBasedStore) UpdateIdentityProvider(ctx context.Context, idp *IDPDTO) error {
	return errors.New("UpdateIdentityProvider is not supported in file-based store")
}

// newIDPFileBasedStore creates a new instance of a file-based store.
func newIDPFileBasedStore() idpStoreInterface {
	genericStore := declarativeresource.NewGenericFileBasedStore(entity.KeyTypeIDP)
	return &idpFileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}
