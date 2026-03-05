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

	"errors"

	"github.com/asgardeo/thunder/internal/application/model"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

type fileBasedStore struct {
	*declarativeresource.GenericFileBasedStore
}

// Create implements declarativeresource.Storer interface for resource loader
func (f *fileBasedStore) Create(id string, data interface{}) error {
	app := data.(*model.ApplicationProcessedDTO)
	return f.CreateApplication(context.TODO(), *app)
}

// CreateApplication implements applicationStoreInterface.
func (f *fileBasedStore) CreateApplication(ctx context.Context, app model.ApplicationProcessedDTO) error {
	return f.GenericFileBasedStore.Create(app.ID, &app)
}

// DeleteApplication implements applicationStoreInterface.
func (f *fileBasedStore) DeleteApplication(ctx context.Context, id string) error {
	return errors.New("DeleteApplication is not supported in file-based store")
}

// GetApplicationByID implements applicationStoreInterface.
func (f *fileBasedStore) GetApplicationByID(ctx context.Context, id string) (*model.ApplicationProcessedDTO, error) {
	data, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		return nil, model.ApplicationNotFoundError
	}
	app, ok := data.(*model.ApplicationProcessedDTO)
	if !ok {
		declarativeresource.LogTypeAssertionError("application", id)
		return nil, model.ApplicationDataCorruptedError
	}
	return app, nil
}

// GetApplicationByName implements applicationStoreInterface.
func (f *fileBasedStore) GetApplicationByName(
	ctx context.Context, name string) (*model.ApplicationProcessedDTO, error) {
	data, err := f.GenericFileBasedStore.GetByField(name, func(d interface{}) string {
		return d.(*model.ApplicationProcessedDTO).Name
	})
	if err != nil {
		return nil, model.ApplicationNotFoundError
	}
	return data.(*model.ApplicationProcessedDTO), nil
}

// GetApplicationList implements applicationStoreInterface.
func (f *fileBasedStore) GetApplicationList(ctx context.Context) ([]model.BasicApplicationDTO, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	var appList []model.BasicApplicationDTO
	for _, item := range list {
		if app, ok := item.Data.(*model.ApplicationProcessedDTO); ok {
			basicApp := model.BasicApplicationDTO{
				ID:                        app.ID,
				Name:                      app.Name,
				Description:               app.Description,
				AuthFlowID:                app.AuthFlowID,
				RegistrationFlowID:        app.RegistrationFlowID,
				IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
			}
			appList = append(appList, basicApp)
		}
	}
	return appList, nil
}

// GetOAuthApplication implements applicationStoreInterface.
func (f *fileBasedStore) GetOAuthApplication(
	ctx context.Context, clientID string) (*model.OAuthAppConfigProcessedDTO, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	for _, item := range list {
		if app, ok := item.Data.(*model.ApplicationProcessedDTO); ok {
			for _, inbound := range app.InboundAuthConfig {
				if inbound.Type == model.OAuthInboundAuthType {
					if inbound.OAuthAppConfig != nil {
						if inbound.OAuthAppConfig.ClientID == clientID {
							return inbound.OAuthAppConfig, nil
						}
					}
				}
			}
		}
	}

	return nil, model.ApplicationNotFoundError
}

// GetTotalApplicationCount implements applicationStoreInterface.
func (f *fileBasedStore) GetTotalApplicationCount(ctx context.Context) (int, error) {
	return f.GenericFileBasedStore.Count()
}

// UpdateApplication implements applicationStoreInterface.
func (f *fileBasedStore) UpdateApplication(ctx context.Context, existingApp *model.ApplicationProcessedDTO,
	updatedApp *model.ApplicationProcessedDTO) error {
	return errors.New("UpdateApplication is not supported in file-based store")
}

// IsApplicationExists implements applicationStoreInterface.
func (f *fileBasedStore) IsApplicationExists(ctx context.Context, id string) (bool, error) {
	_, err := f.GetApplicationByID(ctx, id)
	if err != nil {
		if errors.Is(err, model.ApplicationNotFoundError) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsApplicationExistsByName implements applicationStoreInterface.
func (f *fileBasedStore) IsApplicationExistsByName(ctx context.Context, name string) (bool, error) {
	_, err := f.GetApplicationByName(ctx, name)
	if err != nil {
		if errors.Is(err, model.ApplicationNotFoundError) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsApplicationDeclarative checks if an application is immutable.
// For file-based store, all applications are declarative/immutable.
func (f *fileBasedStore) IsApplicationDeclarative(ctx context.Context, id string) bool {
	exists, err := f.IsApplicationExists(ctx, id)
	return err == nil && exists
}

// newFileBasedStore creates a new instance of a file-based store.
func newFileBasedStore() applicationStoreInterface {
	genericStore := declarativeresource.NewGenericFileBasedStore(entity.KeyTypeApplication)
	return &fileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}
