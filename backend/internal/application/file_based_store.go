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
	"github.com/asgardeo/thunder/internal/system/file_based_runtime/entity"
	"github.com/asgardeo/thunder/internal/system/log"
)

type fileBasedStore struct {
	storage entity.StoreInterface
}

// CreateApplication implements applicationStoreInterface.
func (f *fileBasedStore) CreateApplication(app model.ApplicationProcessedDTO) error {
	appKey := entity.NewCompositeKey(app.ID, entity.KeyTypeApplication)
	return f.storage.Set(appKey, &app)
}

// DeleteApplication implements applicationStoreInterface.
func (f *fileBasedStore) DeleteApplication(id string) error {
	return errors.New("DeleteApplication is not supported in file-based store")
}

// GetApplicationByID implements applicationStoreInterface.
func (f *fileBasedStore) GetApplicationByID(id string) (*model.ApplicationProcessedDTO, error) {
	entity, err := f.storage.Get(entity.NewCompositeKey(id, entity.KeyTypeApplication))
	if err != nil {
		return nil, err
	}
	app, ok := entity.Data.(*model.ApplicationProcessedDTO)
	if !ok {
		log.GetLogger().Error("Type assertion failed while retrieving application by ID",
			log.String("appID", id))
		return nil, model.ApplicationDataCorruptedError
	}
	return app, nil
}

// GetApplicationByName implements applicationStoreInterface.
func (f *fileBasedStore) GetApplicationByName(name string) (*model.ApplicationProcessedDTO, error) {
	list, err := f.storage.ListByType(entity.KeyTypeApplication)
	if err != nil {
		return nil, err
	}

	for _, item := range list {
		if app, ok := item.Data.(*model.ApplicationProcessedDTO); ok && app.Name == name {
			return app, nil
		}
	}

	return nil, model.ApplicationNotFoundError
}

// GetApplicationList implements applicationStoreInterface.
func (f *fileBasedStore) GetApplicationList() ([]model.BasicApplicationDTO, error) {
	list, err := f.storage.ListByType(entity.KeyTypeApplication)
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
				AuthFlowGraphID:           app.AuthFlowGraphID,
				RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
				IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
			}
			appList = append(appList, basicApp)
		}
	}
	return appList, nil
}

// GetOAuthApplication implements applicationStoreInterface.
func (f *fileBasedStore) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO, error) {
	list, err := f.storage.ListByType(entity.KeyTypeApplication)
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
func (f *fileBasedStore) GetTotalApplicationCount() (int, error) {
	list, err := f.storage.ListByType(entity.KeyTypeApplication)
	if err != nil {
		return 0, err
	}
	return len(list), nil
}

// UpdateApplication implements applicationStoreInterface.
func (f *fileBasedStore) UpdateApplication(existingApp *model.ApplicationProcessedDTO,
	updatedApp *model.ApplicationProcessedDTO) error {
	return errors.New("UpdateApplication is not supported in file-based store")
}

// newFileBasedStore creates a new instance of a file-based store.
func newFileBasedStore() applicationStoreInterface {
	store := entity.GetInstance()
	return &fileBasedStore{
		storage: store,
	}
}
