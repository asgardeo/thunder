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
	"fmt"

	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// notificationStoreInterface defines the interface for notification sender storage operations.
type notificationStoreInterface interface {
	createSender(sender common.NotificationSenderDTO) error
	listSenders() ([]common.NotificationSenderDTO, error)
	getSenderByID(id string) (*common.NotificationSenderDTO, error)
	getSenderByName(name string) (*common.NotificationSenderDTO, error)
	updateSender(id string, sender common.NotificationSenderDTO) error
	deleteSender(id string) error
}

// notificationStore is the implementation of notificationStoreInterface.
type notificationStore struct{}

// getNotificationStore returns a new instance of notificationStoreInterface.
func getNotificationStore() notificationStoreInterface {
	return &notificationStore{}
}

// createSender creates a new notification sender.
func (s *notificationStore) createSender(sender common.NotificationSenderDTO) error {
	for i := range sender.Properties {
		if err := sender.Properties[i].Encrypt(); err != nil {
			return fmt.Errorf("failed to encrypt secret properties: %w", err)
		}
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(queryCreateNotificationSender.Query, sender.Name, sender.ID,
				sender.Description, string(sender.Type), string(sender.Provider))
			return err
		},
	}
	for _, prop := range sender.Properties {
		queries = append(queries, func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(queryCreateNotificationSenderProperty.Query, sender.ID,
				prop.Name, prop.Value, sysutils.BoolToNumString(prop.IsSecret), sysutils.BoolToNumString(prop.IsEncrypted()))
			return err
		})
	}

	if err := executeTransaction(queries); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

// listSenders retrieves all notification senders
func (s *notificationStore) listSenders() ([]common.NotificationSenderDTO, error) {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetAllNotificationSenders)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	senders := make([]common.NotificationSenderDTO, 0, len(results))
	for _, row := range results {
		sender, err := s.buildSenderFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build sender from result row: %w", err)
		}

		// Get properties for the sender
		properties, err := s.GetSenderProperties(sender.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get properties for sender %s: %w", sender.ID, err)
		}

		sender.Properties = properties
		senders = append(senders, *sender)
	}

	return senders, nil
}

// GetSenderProperties retrieves all properties of a notification sender by ID.
func (s *notificationStore) GetSenderProperties(id string) ([]cmodels.Property, error) {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetNotificationSenderProperties, id)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	props, err := buildSenderPropertiesFromResultSet(results, id)
	if err != nil {
		return nil, fmt.Errorf("failed to build sender properties from result set: %w", err)
	}
	return props, nil
}

// getSenderByID retrieves a notification sender by ID.
func (s *notificationStore) getSenderByID(id string) (*common.NotificationSenderDTO, error) {
	return s.getSender(queryGetNotificationSenderByID, id)
}

// getSenderByName retrieves a notification sender by name
func (s *notificationStore) getSenderByName(name string) (*common.NotificationSenderDTO, error) {
	return s.getSender(queryGetNotificationSenderByName, name)
}

// getSender retrieves a notification sender by a specific identifier (ID or name).
func (s *notificationStore) getSender(query dbmodel.DBQuery,
	identifier string) (*common.NotificationSenderDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationStore"))

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(query, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if len(results) == 0 {
		logger.Debug("Notification sender not found", log.String("identifier", identifier))
		return nil, nil
	}
	if len(results) > 1 {
		return nil, fmt.Errorf("multiple senders found for identifier: %s", identifier)
	}

	sender, err := s.buildSenderFromResultRow(results[0])
	if err != nil {
		return nil, fmt.Errorf("failed to build sender from result row: %w", err)
	}

	// Get properties for the sender
	properties, err := s.GetSenderProperties(sender.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties for sender %s: %w", sender.ID, err)
	}
	sender.Properties = properties

	return sender, nil
}

// updateSender updates an existing notification sender.
func (s *notificationStore) updateSender(id string, sender common.NotificationSenderDTO) error {
	for i := range sender.Properties {
		if err := sender.Properties[i].Encrypt(); err != nil {
			return fmt.Errorf("failed to encrypt secret properties: %w", err)
		}
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(queryUpdateNotificationSender.Query, sender.Name, sender.Description,
				string(sender.Provider), id, string(sender.Type))
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(queryDeleteNotificationSenderProperties.Query, id)
			return err
		},
	}
	for _, prop := range sender.Properties {
		queries = append(queries, func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(queryCreateNotificationSenderProperty.Query, id, prop.Name,
				prop.Value, sysutils.BoolToNumString(prop.IsSecret), sysutils.BoolToNumString(prop.IsEncrypted()))
			return err
		})
	}

	if err := executeTransaction(queries); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

// deleteSender deletes a notification sender.
func (s *notificationStore) deleteSender(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationStore"))

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.Execute(queryDeleteNotificationSender, id)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}
	if rowsAffected == 0 {
		logger.Debug("No sender found to delete", log.String("id", id))
	}

	return nil
}

// executeTransaction is a helper function to handle database transactions.
func executeTransaction(queries []func(tx dbmodel.TxInterface) error) error {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, query := range queries {
		if err := query(tx); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %s", rollbackErr.Error()))
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// buildSenderFromResultRow constructs a NotificationSenderDTO from a database result row.
func (s *notificationStore) buildSenderFromResultRow(
	row map[string]interface{}) (*common.NotificationSenderDTO, error) {
	senderID := sysutils.ConvertInterfaceValueToString(row["sender_id"])
	if senderID == "" {
		return nil, fmt.Errorf("failed to parse sender_id as string")
	}

	name := sysutils.ConvertInterfaceValueToString(row["name"])
	if name == "" {
		return nil, fmt.Errorf("failed to parse name as string")
	}

	description := sysutils.ConvertInterfaceValueToString(row["description"])
	if description == "" {
		return nil, fmt.Errorf("failed to parse description as string")
	}

	_type := sysutils.ConvertInterfaceValueToString(row["type"])
	if _type == "" {
		return nil, fmt.Errorf("failed to parse type as string")
	}

	provider := sysutils.ConvertInterfaceValueToString(row["provider"])
	if provider == "" {
		return nil, fmt.Errorf("failed to parse provider as string")
	}

	sender := &common.NotificationSenderDTO{
		ID:          senderID,
		Name:        name,
		Description: description,
		Type:        common.NotificationSenderType(_type),
		Provider:    common.MessageProviderType(provider),
		Properties:  []cmodels.Property{},
	}

	return sender, nil
}

// buildSenderPropertiesFromResultSet builds a list of SenderProperty from the result set.
func buildSenderPropertiesFromResultSet(results []map[string]interface{}, id string) ([]cmodels.Property, error) {
	properties := make([]cmodels.Property, 0, len(results))

	for _, row := range results {
		propName := sysutils.ConvertInterfaceValueToString(row["property_name"])
		if propName == "" {
			return nil, fmt.Errorf("failed to parse property_name as string for sender ID: %s", id)
		}

		propValue := sysutils.ConvertInterfaceValueToString(row["property_value"])
		if propValue == "" {
			return nil, fmt.Errorf("failed to parse property_value as string for sender ID: %s", id)
		}

		isSecretStr := sysutils.ConvertInterfaceValueToString(row["is_secret"])
		if isSecretStr == "" {
			return nil, fmt.Errorf("failed to parse is_secret as string for sender ID: %s", id)
		}
		isSecret := sysutils.NumStringToBool(isSecretStr)

		isEncryptedStr := sysutils.ConvertInterfaceValueToString(row["is_encrypted"])
		if isEncryptedStr == "" {
			return nil, fmt.Errorf("failed to parse is_encrypted as string for sender ID: %s", id)
		}
		isEncrypted := sysutils.NumStringToBool(isEncryptedStr)

		property := cmodels.Property{
			Name:     propName,
			Value:    propValue,
			IsSecret: isSecret,
		}
		property.SetEncrypted(isEncrypted)
		properties = append(properties, property)
	}

	return properties, nil
}
