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
	"fmt"

	"github.com/asgardeo/thunder/internal/application/model"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
	"github.com/asgardeo/thunder/internal/system/transaction"
)

type fileBasedStore struct {
	*declarativeresource.GenericFileBasedStore
}

// Create implements declarativeresource.Storer interface for resource loader.
func (f *fileBasedStore) Create(id string, data interface{}) error {
	dto, ok := data.(*model.ApplicationProcessedDTO)
	if !ok {
		return fmt.Errorf("unexpected data type: %T", data)
	}
	dao := toConfigDAO(dto)
	return f.CreateApplication(context.Background(), dao)
}

// CreateApplication implements applicationStoreInterface.
func (f *fileBasedStore) CreateApplication(_ context.Context, app applicationConfigDAO) error {
	return f.GenericFileBasedStore.Create(app.ID, &app)
}

// CreateOAuthConfig is not supported in file-based store — OAuth config is embedded in the app config.
func (f *fileBasedStore) CreateOAuthConfig(_ context.Context, _ string, _ json.RawMessage) error {
	return errors.New("CreateOAuthConfig is not supported in file-based store")
}

// GetApplicationByID implements applicationStoreInterface.
func (f *fileBasedStore) GetApplicationByID(_ context.Context, id string) (*applicationConfigDAO, error) {
	data, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		return nil, model.ApplicationNotFoundError
	}
	app, ok := data.(*applicationConfigDAO)
	if !ok {
		declarativeresource.LogTypeAssertionError("application", id)
		return nil, model.ApplicationDataCorruptedError
	}
	return app, nil
}

// GetOAuthConfigByEntityID implements applicationStoreInterface.
func (f *fileBasedStore) GetOAuthConfigByEntityID(_ context.Context, _ string) (*oauthConfigDAO, error) {
	return nil, model.ApplicationNotFoundError
}

// GetApplicationList implements applicationStoreInterface.
func (f *fileBasedStore) GetApplicationList(_ context.Context) ([]applicationConfigDAO, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	var appList []applicationConfigDAO
	for _, item := range list {
		if app, ok := item.Data.(*applicationConfigDAO); ok {
			app.IsReadOnly = true
			appList = append(appList, *app)
		}
	}
	return appList, nil
}

// GetTotalApplicationCount implements applicationStoreInterface.
func (f *fileBasedStore) GetTotalApplicationCount(_ context.Context) (int, error) {
	return f.GenericFileBasedStore.Count()
}

// UpdateApplication is not supported in file-based store.
func (f *fileBasedStore) UpdateApplication(_ context.Context, _ applicationConfigDAO) error {
	return errors.New("UpdateApplication is not supported in file-based store")
}

// UpdateOAuthConfig is not supported in file-based store.
func (f *fileBasedStore) UpdateOAuthConfig(_ context.Context, _ string, _ json.RawMessage) error {
	return errors.New("UpdateOAuthConfig is not supported in file-based store")
}

// DeleteApplication is not supported in file-based store.
func (f *fileBasedStore) DeleteApplication(_ context.Context, _ string) error {
	return errors.New("DeleteApplication is not supported in file-based store")
}

// DeleteOAuthConfig is not supported in file-based store.
func (f *fileBasedStore) DeleteOAuthConfig(_ context.Context, _ string) error {
	return errors.New("DeleteOAuthConfig is not supported in file-based store")
}

// IsApplicationExists implements applicationStoreInterface.
func (f *fileBasedStore) IsApplicationExists(_ context.Context, id string) (bool, error) {
	_, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// IsApplicationDeclarative returns true for file-based store (all are declarative/immutable).
func (f *fileBasedStore) IsApplicationDeclarative(_ context.Context, id string) bool {
	_, err := f.GenericFileBasedStore.Get(id)
	return err == nil
}

// newFileBasedStore creates a new instance of a file-based store.
func newFileBasedStore() (applicationStoreInterface, transaction.Transactioner) {
	genericStore := declarativeresource.NewGenericFileBasedStore(entity.KeyTypeApplication)
	return &fileBasedStore{
		GenericFileBasedStore: genericStore,
	}, transaction.NewNoOpTransactioner()
}
