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

// Package notification contains the implementation of notification and otp sender services.
package notification

import (
	"github.com/asgardeo/thunder/internal/notification/common"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// NotificationSenderMgtSvcInterface defines the interface for managing notification senders.
type NotificationSenderMgtSvcInterface interface {
	CreateSender(sender common.NotificationSenderDTO) (*common.NotificationSenderDTO,
		*serviceerror.ServiceError)
	ListSenders() ([]common.NotificationSenderDTO, *serviceerror.ServiceError)
	GetSender(id string) (*common.NotificationSenderDTO, *serviceerror.ServiceError)
	GetSenderByName(name string) (*common.NotificationSenderDTO, *serviceerror.ServiceError)
	UpdateSender(id string, sender common.NotificationSenderDTO) (*common.NotificationSenderDTO,
		*serviceerror.ServiceError)
	DeleteSender(id string) *serviceerror.ServiceError
}

// notificationSenderMgtService implements the NotificationSenderMgtSvcInterface.
type notificationSenderMgtService struct {
	notificationStore notificationStoreInterface
}

// newNotificationSenderMgtService returns a new instance of NotificationSenderMgtSvcInterface.
func newNotificationSenderMgtService(store notificationStoreInterface) NotificationSenderMgtSvcInterface {
	return &notificationSenderMgtService{
		notificationStore: store,
	}
}

// NewNotificationSenderMgtService creates a new instance of NotificationSenderMgtSvcInterface.
// [Deprecated: use dependency injection to get the instance instead].
// TODO: Remove this when the flow executors are migrated to the di pattern.
func NewNotificationSenderMgtService() NotificationSenderMgtSvcInterface {
	return newNotificationSenderMgtService(newNotificationStore())
}

// CreateSender creates a new notification sender.
func (s *notificationSenderMgtService) CreateSender(
	sender common.NotificationSenderDTO) (*common.NotificationSenderDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationSenderMgtService"))
	logger.Debug("Creating notification sender", log.String("name", sender.Name))

	if err := declarativeresource.CheckDeclarativeCreate(); err != nil {
		return nil, err
	}

	if err := validateNotificationSender(sender); err != nil {
		return nil, err
	}

	// Check if sender with same name already exists
	senderRetv, err := s.notificationStore.getSenderByName(sender.Name)
	if err != nil {
		logger.Error("Failed to retrieve notification sender", log.String("name", sender.Name),
			log.Error(err))
		return nil, &ErrorInternalServerError
	}
	if senderRetv != nil {
		logger.Debug("Notification sender already exists", log.String("name", sender.Name),
			log.String("id", senderRetv.ID))
		return nil, &ErrorDuplicateSenderName
	}

	id := sysutils.GenerateUUID()
	sender.ID = id

	// Create the sender
	err = s.notificationStore.createSender(sender)
	if err != nil {
		logger.Error("Failed to create notification sender", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return &common.NotificationSenderDTO{
		ID:          sender.ID,
		Name:        sender.Name,
		Description: sender.Description,
		Type:        sender.Type,
		Provider:    sender.Provider,
		Properties:  sender.Properties,
	}, nil
}

// ListSenders retrieves all notification senders.
func (s *notificationSenderMgtService) ListSenders() ([]common.NotificationSenderDTO,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationSenderMgtService"))
	logger.Debug("Listing all notification senders")

	senders, err := s.notificationStore.listSenders()
	if err != nil {
		logger.Error("Failed to list notification senders", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return senders, nil
}

// GetSender retrieves a notification sender by ID.
func (s *notificationSenderMgtService) GetSender(id string) (*common.NotificationSenderDTO,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationSenderMgtService"))
	logger.Debug("Retrieving notification sender", log.String("id", id))

	if id == "" {
		return nil, &ErrorInvalidSenderID
	}

	sender, err := s.notificationStore.getSenderByID(id)
	if err != nil {
		logger.Error("Failed to retrieve notification sender", log.String("id", id), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return sender, nil
}

// GetSenderByName retrieves a notification sender by name.
func (s *notificationSenderMgtService) GetSenderByName(name string) (*common.NotificationSenderDTO,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationSenderMgtService"))
	logger.Debug("Retrieving notification sender by name", log.String("name", name))

	if name == "" {
		return nil, &ErrorInvalidSenderName
	}

	sender, err := s.notificationStore.getSenderByName(name)
	if err != nil {
		logger.Error("Failed to retrieve notification sender", log.String("name", name), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return sender, nil
}

// UpdateSender updates an existing notification sender
func (s *notificationSenderMgtService) UpdateSender(id string,
	sender common.NotificationSenderDTO) (*common.NotificationSenderDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationSenderMgtService"))
	logger.Debug("Updating notification sender", log.String("id", id), log.String("name", sender.Name))

	if err := declarativeresource.CheckDeclarativeUpdate(); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, &ErrorInvalidSenderID
	}
	if err := validateNotificationSender(sender); err != nil {
		return nil, err
	}

	// Check if sender exists
	senderRetv, err := s.notificationStore.getSenderByID(id)
	if err != nil {
		logger.Error("Failed to retrieve notification sender", log.String("id", id), log.Error(err))
		return nil, &ErrorInternalServerError
	}
	if senderRetv == nil {
		logger.Debug("Notification sender not found", log.String("id", id))
		return nil, &ErrorSenderNotFound
	}

	// If the name is being updated, check for duplicates
	if sender.Name != senderRetv.Name {
		senderWithUpdatedName, err := s.notificationStore.getSenderByName(sender.Name)
		if err != nil {
			logger.Error("Failed to retrieve notification sender", log.String("name", sender.Name),
				log.Error(err))
			return nil, &ErrorInternalServerError
		}
		if senderWithUpdatedName != nil && senderWithUpdatedName.ID != id {
			logger.Debug("Another sender with the same name already exists",
				log.String("name", sender.Name), log.String("existingID", senderWithUpdatedName.ID))
			return nil, &ErrorDuplicateSenderName
		}
	}

	// Ensure the type is not changed
	if sender.Type != senderRetv.Type {
		logger.Debug("Attempting to change sender type", log.String("id", id),
			log.String("originalType", string(senderRetv.Type)), log.String("newType", string(sender.Type)))
		return nil, &ErrorSenderTypeUpdateNotAllowed
	}

	// Update the sender
	if err := s.notificationStore.updateSender(id, sender); err != nil {
		logger.Error("Failed to update notification sender", log.String("id", id), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return &common.NotificationSenderDTO{
		ID:          id,
		Name:        sender.Name,
		Description: sender.Description,
		Type:        sender.Type,
		Provider:    sender.Provider,
		Properties:  sender.Properties,
	}, nil
}

// DeleteSender deletes a notification sender
func (s *notificationSenderMgtService) DeleteSender(id string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationSenderMgtService"))
	logger.Debug("Deleting notification sender", log.String("id", id))

	if err := declarativeresource.CheckDeclarativeDelete(); err != nil {
		return err
	}

	if id == "" {
		return &ErrorInvalidSenderID
	}

	if err := s.notificationStore.deleteSender(id); err != nil {
		logger.Error("Failed to delete notification sender", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}

	return nil
}
