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
	"fmt"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// Common errors for invitation store operations
var (
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrTokenNotFound      = errors.New("invitation token not found")
	ErrInvitationExpired  = errors.New("invitation has expired")
)

// Database query constants
var (
	QueryCreateInvitation = dbmodel.DBQuery{
		ID:    "INV-001",
		Query: `INSERT INTO INVITATION (INVITATION_ID, USER_ID, APPLICATION_ID, TOKEN, STATUS, EXPIRES_AT, CREATED_AT, DEPLOYMENT_ID) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
	}
	QueryGetInvitationByID = dbmodel.DBQuery{
		ID:    "INV-002",
		Query: `SELECT INVITATION_ID, USER_ID, APPLICATION_ID, TOKEN, STATUS, EXPIRES_AT, CREATED_AT, REDEEMED_AT FROM INVITATION WHERE INVITATION_ID = $1 AND DEPLOYMENT_ID = $2`,
	}
	QueryGetInvitationByToken = dbmodel.DBQuery{
		ID:    "INV-003",
		Query: `SELECT INVITATION_ID, USER_ID, APPLICATION_ID, TOKEN, STATUS, EXPIRES_AT, CREATED_AT, REDEEMED_AT FROM INVITATION WHERE TOKEN = $1 AND DEPLOYMENT_ID = $2`,
	}
	QueryGetInvitationByUserID = dbmodel.DBQuery{
		ID:    "INV-004",
		Query: `SELECT INVITATION_ID, USER_ID, APPLICATION_ID, TOKEN, STATUS, EXPIRES_AT, CREATED_AT, REDEEMED_AT FROM INVITATION WHERE USER_ID = $1 AND STATUS = $2 AND DEPLOYMENT_ID = $3 ORDER BY CREATED_AT DESC LIMIT 1`,
	}
	QueryUpdateInvitation = dbmodel.DBQuery{
		ID:    "INV-005",
		Query: `UPDATE INVITATION SET STATUS = $1, REDEEMED_AT = $2 WHERE INVITATION_ID = $3 AND DEPLOYMENT_ID = $4`,
	}
	QueryDeleteInvitation = dbmodel.DBQuery{
		ID:    "INV-006",
		Query: `DELETE FROM INVITATION WHERE INVITATION_ID = $1 AND DEPLOYMENT_ID = $2`,
	}
	QueryGetInvitationCount = dbmodel.DBQuery{
		ID:    "INV-007",
		Query: `SELECT COUNT(*) as total FROM INVITATION WHERE DEPLOYMENT_ID = $1`,
	}
	QueryGetInvitationList = dbmodel.DBQuery{
		ID:    "INV-008",
		Query: `SELECT INVITATION_ID, USER_ID, APPLICATION_ID, TOKEN, STATUS, EXPIRES_AT, CREATED_AT, REDEEMED_AT FROM INVITATION WHERE DEPLOYMENT_ID = $1 ORDER BY CREATED_AT DESC LIMIT $2 OFFSET $3`,
	}
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
	deploymentID string
}

// newInvitationStore creates a new instance of the invitation store.
func newInvitationStore() invitationStoreInterface {
	runtime := config.GetThunderRuntime()
	return &invitationStore{
		deploymentID: runtime.Config.Server.Identifier,
	}
}

// CreateInvitation creates a new invitation in the database.
func (s *invitationStore) CreateInvitation(invitation Invitation) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	dbClient, err := provider.GetDBProvider().GetIdentityDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Execute(
		QueryCreateInvitation,
		invitation.ID,
		invitation.UserID,
		invitation.ApplicationID,
		invitation.Token,
		invitation.Status,
		invitation.ExpiresAt,
		invitation.CreatedAt,
		s.deploymentID,
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

	dbClient, err := provider.GetDBProvider().GetIdentityDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryGetInvitationByID, invitationID, s.deploymentID)
	if err != nil {
		logger.Error("Failed to query invitation", log.Error(err))
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrInvitationNotFound
	}

	invitation, err := buildInvitationFromRow(results[0])
	if err != nil {
		logger.Error("Failed to build invitation from row", log.Error(err))
		return nil, err
	}

	return invitation, nil
}

// GetInvitationByToken retrieves an invitation by its token.
func (s *invitationStore) GetInvitationByToken(token string) (*Invitation, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	dbClient, err := provider.GetDBProvider().GetIdentityDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryGetInvitationByToken, token, s.deploymentID)
	if err != nil {
		logger.Error("Failed to query invitation by token", log.Error(err))
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrTokenNotFound
	}

	invitation, err := buildInvitationFromRow(results[0])
	if err != nil {
		logger.Error("Failed to build invitation from row", log.Error(err))
		return nil, err
	}

	return invitation, nil
}

// GetInvitationByUserID retrieves the latest pending invitation for a user.
func (s *invitationStore) GetInvitationByUserID(userID string) (*Invitation, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	dbClient, err := provider.GetDBProvider().GetIdentityDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryGetInvitationByUserID, userID, StatusPending, s.deploymentID)
	if err != nil {
		logger.Error("Failed to query invitation by user ID", log.Error(err))
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrInvitationNotFound
	}

	invitation, err := buildInvitationFromRow(results[0])
	if err != nil {
		logger.Error("Failed to build invitation from row", log.Error(err))
		return nil, err
	}

	return invitation, nil
}

// UpdateInvitation updates an existing invitation.
func (s *invitationStore) UpdateInvitation(invitation *Invitation) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InvitationStore"))

	dbClient, err := provider.GetDBProvider().GetIdentityDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.Execute(
		QueryUpdateInvitation,
		invitation.Status,
		invitation.RedeemedAt,
		invitation.ID,
		s.deploymentID,
	)
	if err != nil {
		logger.Error("Failed to update invitation", log.Error(err))
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

	dbClient, err := provider.GetDBProvider().GetIdentityDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.Execute(QueryDeleteInvitation, invitationID, s.deploymentID)
	if err != nil {
		logger.Error("Failed to delete invitation", log.Error(err))
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

	dbClient, err := provider.GetDBProvider().GetIdentityDBClient()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get database client: %w", err)
	}

	// Get total count
	countResults, err := dbClient.Query(QueryGetInvitationCount, s.deploymentID)
	if err != nil {
		logger.Error("Failed to count invitations", log.Error(err))
		return nil, 0, err
	}

	var totalCount int
	if len(countResults) > 0 {
		if count, ok := countResults[0]["total"].(int64); ok {
			totalCount = int(count)
		}
	}

	// Get invitations
	results, err := dbClient.Query(QueryGetInvitationList, s.deploymentID, limit, offset)
	if err != nil {
		logger.Error("Failed to query invitations", log.Error(err))
		return nil, 0, err
	}

	invitations := make([]Invitation, 0, len(results))
	for _, row := range results {
		invitation, err := buildInvitationFromRow(row)
		if err != nil {
			logger.Error("Failed to build invitation from row", log.Error(err))
			return nil, 0, err
		}
		invitations = append(invitations, *invitation)
	}

	return invitations, totalCount, nil
}

// buildInvitationFromRow constructs an Invitation from a database result row.
func buildInvitationFromRow(row map[string]interface{}) (*Invitation, error) {
	invitation := &Invitation{}

	// Parse invitation ID
	if id, ok := row["invitation_id"].(string); ok {
		invitation.ID = id
	} else {
		return nil, fmt.Errorf("failed to parse invitation_id as string")
	}

	// Parse user ID
	if userID, ok := row["user_id"].(string); ok {
		invitation.UserID = userID
	} else {
		return nil, fmt.Errorf("failed to parse user_id as string")
	}

	// Parse application ID
	if appID, ok := row["application_id"].(string); ok {
		invitation.ApplicationID = appID
	} else {
		return nil, fmt.Errorf("failed to parse application_id as string")
	}

	// Parse token
	if token, ok := row["token"].(string); ok {
		invitation.Token = token
	} else {
		return nil, fmt.Errorf("failed to parse token as string")
	}

	// Parse status
	if status, ok := row["status"].(string); ok {
		invitation.Status = status
	} else {
		return nil, fmt.Errorf("failed to parse status as string")
	}

	// Parse expires_at
	if expiresAt, ok := row["expires_at"].(time.Time); ok {
		invitation.ExpiresAt = expiresAt
	} else if expiresAtStr, ok := row["expires_at"].(string); ok {
		parsed, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expires_at: %w", err)
		}
		invitation.ExpiresAt = parsed
	}

	// Parse created_at
	if createdAt, ok := row["created_at"].(time.Time); ok {
		invitation.CreatedAt = createdAt
	} else if createdAtStr, ok := row["created_at"].(string); ok {
		parsed, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		invitation.CreatedAt = parsed
	}

	// Parse redeemed_at (optional)
	if redeemedAt, ok := row["redeemed_at"].(time.Time); ok {
		invitation.RedeemedAt = &redeemedAt
	} else if redeemedAtStr, ok := row["redeemed_at"].(string); ok && redeemedAtStr != "" {
		parsed, err := time.Parse(time.RFC3339, redeemedAtStr)
		if err == nil {
			invitation.RedeemedAt = &parsed
		}
	}

	return invitation, nil
}
