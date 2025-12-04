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

package notification

import (
	"errors"

	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/file_based_runtime/entity"
	"github.com/asgardeo/thunder/internal/system/log"
)

type notificationFileBasedStore struct {
	storage entity.StoreInterface
}

// createSender implements notificationStoreInterface.
func (f *notificationFileBasedStore) createSender(sender common.NotificationSenderDTO) error {
	senderKey := entity.NewCompositeKey(sender.ID, entity.KeyTypeNotificationSender)
	return f.storage.Set(senderKey, &sender)
}

// deleteSender implements notificationStoreInterface.
func (f *notificationFileBasedStore) deleteSender(id string) error {
	return errors.New("deleteSender is not supported in file-based store")
}

// getSenderByID implements notificationStoreInterface.
func (f *notificationFileBasedStore) getSenderByID(id string) (*common.NotificationSenderDTO, error) {
	entity, err := f.storage.Get(entity.NewCompositeKey(id, entity.KeyTypeNotificationSender))
	if err != nil {
		return nil, err
	}
	sender, ok := entity.Data.(*common.NotificationSenderDTO)
	if !ok {
		log.GetLogger().Error("Type assertion failed while retrieving notification sender by ID",
			log.String("senderID", id))
		return nil, errors.New("notification sender data corrupted")
	}
	return sender, nil
}

// getSenderByName implements notificationStoreInterface.
func (f *notificationFileBasedStore) getSenderByName(name string) (*common.NotificationSenderDTO, error) {
	list, err := f.storage.ListByType(entity.KeyTypeNotificationSender)
	if err != nil {
		return nil, err
	}

	for _, item := range list {
		if sender, ok := item.Data.(*common.NotificationSenderDTO); ok && sender.Name == name {
			return sender, nil
		}
	}

	return nil, nil
}

// listSenders implements notificationStoreInterface.
func (f *notificationFileBasedStore) listSenders() ([]common.NotificationSenderDTO, error) {
	list, err := f.storage.ListByType(entity.KeyTypeNotificationSender)
	if err != nil {
		return nil, err
	}

	var senderList []common.NotificationSenderDTO
	for _, item := range list {
		if sender, ok := item.Data.(*common.NotificationSenderDTO); ok {
			senderList = append(senderList, *sender)
		}
	}
	return senderList, nil
}

// updateSender implements notificationStoreInterface.
func (f *notificationFileBasedStore) updateSender(id string, sender common.NotificationSenderDTO) error {
	return errors.New("updateSender is not supported in file-based store")
}

// newNotificationFileBasedStore creates a new instance of a file-based store.
func newNotificationFileBasedStore() notificationStoreInterface {
	storage := entity.NewStore()
	return &notificationFileBasedStore{
		storage: storage,
	}
}
