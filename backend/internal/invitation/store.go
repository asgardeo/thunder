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

package invitation

import (
	"errors"
	"time"

	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// Common errors for invitation store operations
var (
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrTokenNotFound      = errors.New("invitation token not found")
	ErrInvitationExpired  = errors.New("invitation has expired")
)

// invitationStoreInterface defines the interface for invitation data access operations.
type invitationStoreInterface interface {
	CreateInvitation(invitation Invitation) error
	GetInvitation(invitationID string) (*Invitation, error)
	GetInvitationByToken(token string) (*Invitation, error)
	GetInvitationByUserID(userID string) (*Invitation, error)
	UpdateInvitation(invitation *Invitation) error
	DeleteInvitation(invitationID string) error
	GetInvitationList(limit, offset int) ([]Invitation, int, error)
}

// invitationStore is the default implementation of invitationStoreInterface.
type invitationStore struct {
	dbClient provider.DBClientInterface
}

// newInvitationStore creates a new instance of the invitation store.
func newInvitationStore(dbClient provider.DBClientInterface) invitationStoreInterface {
	return &invitationStore{
		dbClient: dbClient,
	}
}

// CreateInvitation creates a new invitation in the database.
func (s *invitationStore) CreateInvitation(invitation Invitation) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	query := `
		INSERT INTO invitations (id, user_id, application_id, token, status, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.dbClient.Execute(
		query,
		invitation.ID,
		invitation.UserID,
		invitation.ApplicationID,
		invitation.Token,
		invitation.Status,
		invitation.ExpiresAt,
		invitation.CreatedAt,
	)

	if err != nil {
		logger.Error("Failed to create invitation", log.Error(err))
		return err
	}

	logger.Debug("Successfully created invitation", log.String("id", invitation.ID))
	return nil
}

// GetInvitation retrieves an invitation by its ID.
func (s *invitationStore) GetInvitation(invitationID string) (*Invitation, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	query := `
		SELECT id, user_id, application_id, token, status, expires_at, created_at, redeemed_at
		FROM invitations
		WHERE id = ?
	`

	rows, err := s.dbClient.Query(query, invitationID)
	if err != nil {
		logger.Error("Failed to query invitation", log.Error(err))
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrInvitationNotFound
	}

	invitation, err := scanInvitation(rows)
	if err != nil {
		logger.Error("Failed to scan invitation", log.Error(err))
		return nil, err
	}

	return invitation, nil
}

// GetInvitationByToken retrieves an invitation by its token.
func (s *invitationStore) GetInvitationByToken(token string) (*Invitation, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	query := `
		SELECT id, user_id, application_id, token, status, expires_at, created_at, redeemed_at
		FROM invitations
		WHERE token = ?
	`

	rows, err := s.dbClient.Query(query, token)
	if err != nil {
		logger.Error("Failed to query invitation by token", log.Error(err))
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrTokenNotFound
	}

	invitation, err := scanInvitation(rows)
	if err != nil {
		logger.Error("Failed to scan invitation", log.Error(err))
		return nil, err
	}

	return invitation, nil
}

// GetInvitationByUserID retrieves the latest pending invitation for a user.
func (s *invitationStore) GetInvitationByUserID(userID string) (*Invitation, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	query := `
		SELECT id, user_id, application_id, token, status, expires_at, created_at, redeemed_at
		FROM invitations
		WHERE user_id = ? AND status = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	rows, err := s.dbClient.Query(query, userID, StatusPending)
	if err != nil {
		logger.Error("Failed to query invitation by user ID", log.Error(err))
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrInvitationNotFound
	}

	invitation, err := scanInvitation(rows)
	if err != nil {
		logger.Error("Failed to scan invitation", log.Error(err))
		return nil, err
	}

	return invitation, nil
}

// UpdateInvitation updates an existing invitation.
func (s *invitationStore) UpdateInvitation(invitation *Invitation) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	query := `
		UPDATE invitations
		SET status = ?, redeemed_at = ?
		WHERE id = ?
	`

	result, err := s.dbClient.Execute(query, invitation.Status, invitation.RedeemedAt, invitation.ID)
	if err != nil {
		logger.Error("Failed to update invitation", log.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", log.Error(err))
		return err
	}

	if rowsAffected == 0 {
		return ErrInvitationNotFound
	}

	logger.Debug("Successfully updated invitation", log.String("id", invitation.ID))
	return nil
}

// DeleteInvitation deletes an invitation by its ID.
func (s *invitationStore) DeleteInvitation(invitationID string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	query := `DELETE FROM invitations WHERE id = ?`

	result, err := s.dbClient.Execute(query, invitationID)
	if err != nil {
		logger.Error("Failed to delete invitation", log.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", log.Error(err))
		return err
	}

	if rowsAffected == 0 {
		return ErrInvitationNotFound
	}

	logger.Debug("Successfully deleted invitation", log.String("id", invitationID))
	return nil
}

// GetInvitationList retrieves a paginated list of invitations.
func (s *invitationStore) GetInvitationList(limit, offset int) ([]Invitation, int, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	// Get total count
	countQuery := `SELECT COUNT(*) FROM invitations`
	countRows, err := s.dbClient.Query(countQuery)
	if err != nil {
		logger.Error("Failed to count invitations", log.Error(err))
		return nil, 0, err
	}
	defer countRows.Close()

	var totalCount int
	if countRows.Next() {
		if err := countRows.Scan(&totalCount); err != nil {
			logger.Error("Failed to scan count", log.Error(err))
			return nil, 0, err
		}
	}

	// Get invitations
	query := `
		SELECT id, user_id, application_id, token, status, expires_at, created_at, redeemed_at
		FROM invitations
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.dbClient.Query(query, limit, offset)
	if err != nil {
		logger.Error("Failed to query invitations", log.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var invitations []Invitation
	for rows.Next() {
		invitation, err := scanInvitation(rows)
		if err != nil {
			logger.Error("Failed to scan invitation", log.Error(err))
			return nil, 0, err
		}
		invitations = append(invitations, *invitation)
	}

	return invitations, totalCount, nil
}

// scanInvitation scans a database row into an Invitation struct.
func scanInvitation(rows provider.RowsInterface) (*Invitation, error) {
	var invitation Invitation
	var redeemedAt *time.Time

	err := rows.Scan(
		&invitation.ID,
		&invitation.UserID,
		&invitation.ApplicationID,
		&invitation.Token,
		&invitation.Status,
		&invitation.ExpiresAt,
		&invitation.CreatedAt,
		&redeemedAt,
	)
	if err != nil {
		return nil, err
	}

	invitation.RedeemedAt = redeemedAt
	return &invitation, nil
}
