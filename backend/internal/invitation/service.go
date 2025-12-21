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
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "InvitationService"

// Service error definitions
var (
	ErrorInvitationNotFound = serviceerror.ServiceError{
		Code:             "INV-40401",
		Error:            "invitation_not_found",
		ErrorDescription: "The requested invitation was not found",
		Type:             serviceerror.ClientErrorType,
	}
	ErrorInvalidToken = serviceerror.ServiceError{
		Code:             "INV-40001",
		Error:            "invalid_token",
		ErrorDescription: "The invitation token is invalid",
		Type:             serviceerror.ClientErrorType,
	}
	ErrorInvitationExpired = serviceerror.ServiceError{
		Code:             "INV-40002",
		Error:            "invitation_expired",
		ErrorDescription: "The invitation has expired",
		Type:             serviceerror.ClientErrorType,
	}
	ErrorInvitationAlreadyRedeemed = serviceerror.ServiceError{
		Code:             "INV-40003",
		Error:            "invitation_already_redeemed",
		ErrorDescription: "The invitation has already been redeemed",
		Type:             serviceerror.ClientErrorType,
	}
	ErrorInvitationRevoked = serviceerror.ServiceError{
		Code:             "INV-40004",
		Error:            "invitation_revoked",
		ErrorDescription: "The invitation has been revoked",
		Type:             serviceerror.ClientErrorType,
	}
	ErrorMissingUserID = serviceerror.ServiceError{
		Code:             "INV-40005",
		Error:            "missing_user_id",
		ErrorDescription: "User ID is required to create an invitation",
		Type:             serviceerror.ClientErrorType,
	}
	ErrorMissingApplicationID = serviceerror.ServiceError{
		Code:             "INV-40006",
		Error:            "missing_application_id",
		ErrorDescription: "Application ID is required to create an invitation",
		Type:             serviceerror.ClientErrorType,
	}
	ErrorInternalServerError = serviceerror.ServiceError{
		Code:             "INV-50001",
		Error:            "internal_server_error",
		ErrorDescription: "An internal server error occurred",
		Type:             serviceerror.ServerErrorType,
	}
)

// InvitationServiceInterface defines the interface for invitation operations.
type InvitationServiceInterface interface {
	CreateInvitation(request CreateInvitationRequest) (*CreateInvitationResponse, *serviceerror.ServiceError)
	ValidateToken(token string) (*ValidateInvitationResponse, *serviceerror.ServiceError)
	RedeemInvitation(token string) *serviceerror.ServiceError
	RevokeInvitation(invitationID string) *serviceerror.ServiceError
	GetInvitation(invitationID string) (*Invitation, *serviceerror.ServiceError)
	GetInvitationByUserID(userID string) (*Invitation, *serviceerror.ServiceError)
	DeleteInvitation(invitationID string) *serviceerror.ServiceError
}

// invitationService is the default implementation of InvitationServiceInterface.
type invitationService struct {
	store      invitationStoreInterface
	serverConf *config.ServerConfig
}

// NewInvitationService creates a new instance of the invitation service.
func NewInvitationService(store invitationStoreInterface, serverConf *config.ServerConfig) InvitationServiceInterface {
	return &invitationService{
		store:      store,
		serverConf: serverConf,
	}
}

// CreateInvitation creates a new invitation for a user.
func (s *invitationService) CreateInvitation(request CreateInvitationRequest) (*CreateInvitationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating invitation", log.String("userID", request.UserID))

	if request.UserID == "" {
		return nil, &ErrorMissingUserID
	}

	if request.ApplicationID == "" {
		return nil, &ErrorMissingApplicationID
	}

	// Generate secure token
	token, err := generateSecureToken(32)
	if err != nil {
		logger.Error("Failed to generate secure token", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	// Calculate expiration time
	expiresIn := request.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = DefaultExpirationSeconds
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	invitation := Invitation{
		ID:            utils.GenerateUUID(),
		UserID:        request.UserID,
		ApplicationID: request.ApplicationID,
		Token:         token,
		Status:        StatusPending,
		ExpiresAt:     expiresAt,
		CreatedAt:     time.Now(),
	}

	if err := s.store.CreateInvitation(invitation); err != nil {
		logger.Error("Failed to create invitation", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	// Generate invite link
	inviteLink := s.generateInviteLink(token)

	logger.Debug("Successfully created invitation",
		log.String("invitationID", invitation.ID),
		log.String("userID", request.UserID))

	return &CreateInvitationResponse{
		InvitationID: invitation.ID,
		UserID:       invitation.UserID,
		Token:        invitation.Token,
		InviteLink:   inviteLink,
		ExpiresAt:    invitation.ExpiresAt,
	}, nil
}

// ValidateToken validates an invitation token and returns the invitation details.
func (s *invitationService) ValidateToken(token string) (*ValidateInvitationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if token == "" {
		return &ValidateInvitationResponse{Valid: false, Message: "Token is required"}, &ErrorInvalidToken
	}

	invitation, err := s.store.GetInvitationByToken(token)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return &ValidateInvitationResponse{Valid: false, Message: "Token not found"}, &ErrorInvalidToken
		}
		logger.Error("Failed to get invitation by token", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	// Check status
	switch invitation.Status {
	case StatusRedeemed:
		return &ValidateInvitationResponse{Valid: false, Message: "Invitation already redeemed"}, &ErrorInvitationAlreadyRedeemed
	case StatusRevoked:
		return &ValidateInvitationResponse{Valid: false, Message: "Invitation has been revoked"}, &ErrorInvitationRevoked
	case StatusExpired:
		return &ValidateInvitationResponse{Valid: false, Message: "Invitation has expired"}, &ErrorInvitationExpired
	}

	// Check expiration
	if time.Now().After(invitation.ExpiresAt) {
		// Update status to expired
		invitation.Status = StatusExpired
		_ = s.store.UpdateInvitation(invitation)
		return &ValidateInvitationResponse{Valid: false, Message: "Invitation has expired"}, &ErrorInvitationExpired
	}

	return &ValidateInvitationResponse{
		Valid:         true,
		UserID:        invitation.UserID,
		ApplicationID: invitation.ApplicationID,
	}, nil
}

// RedeemInvitation marks an invitation as redeemed.
func (s *invitationService) RedeemInvitation(token string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	invitation, err := s.store.GetInvitationByToken(token)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return &ErrorInvalidToken
		}
		logger.Error("Failed to get invitation by token", log.Error(err))
		return &ErrorInternalServerError
	}

	if invitation.Status != StatusPending {
		switch invitation.Status {
		case StatusRedeemed:
			return &ErrorInvitationAlreadyRedeemed
		case StatusRevoked:
			return &ErrorInvitationRevoked
		case StatusExpired:
			return &ErrorInvitationExpired
		}
	}

	if time.Now().After(invitation.ExpiresAt) {
		invitation.Status = StatusExpired
		_ = s.store.UpdateInvitation(invitation)
		return &ErrorInvitationExpired
	}

	now := time.Now()
	invitation.Status = StatusRedeemed
	invitation.RedeemedAt = &now

	if err := s.store.UpdateInvitation(invitation); err != nil {
		logger.Error("Failed to redeem invitation", log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully redeemed invitation", log.String("invitationID", invitation.ID))
	return nil
}

// RevokeInvitation revokes an invitation.
func (s *invitationService) RevokeInvitation(invitationID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	invitation, err := s.store.GetInvitation(invitationID)
	if err != nil {
		if errors.Is(err, ErrInvitationNotFound) {
			return &ErrorInvitationNotFound
		}
		logger.Error("Failed to get invitation", log.Error(err))
		return &ErrorInternalServerError
	}

	if invitation.Status != StatusPending {
		return &ErrorInvitationAlreadyRedeemed
	}

	invitation.Status = StatusRevoked
	if err := s.store.UpdateInvitation(invitation); err != nil {
		logger.Error("Failed to revoke invitation", log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully revoked invitation", log.String("invitationID", invitationID))
	return nil
}

// GetInvitation retrieves an invitation by ID.
func (s *invitationService) GetInvitation(invitationID string) (*Invitation, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	invitation, err := s.store.GetInvitation(invitationID)
	if err != nil {
		if errors.Is(err, ErrInvitationNotFound) {
			return nil, &ErrorInvitationNotFound
		}
		logger.Error("Failed to get invitation", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return invitation, nil
}

// GetInvitationByUserID retrieves the pending invitation for a user.
func (s *invitationService) GetInvitationByUserID(userID string) (*Invitation, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	invitation, err := s.store.GetInvitationByUserID(userID)
	if err != nil {
		if errors.Is(err, ErrInvitationNotFound) {
			return nil, &ErrorInvitationNotFound
		}
		logger.Error("Failed to get invitation by user ID", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return invitation, nil
}

// DeleteInvitation deletes an invitation.
func (s *invitationService) DeleteInvitation(invitationID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := s.store.DeleteInvitation(invitationID); err != nil {
		if errors.Is(err, ErrInvitationNotFound) {
			return &ErrorInvitationNotFound
		}
		logger.Error("Failed to delete invitation", log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully deleted invitation", log.String("invitationID", invitationID))
	return nil
}

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// generateInviteLink generates the invite link URL.
func (s *invitationService) generateInviteLink(token string) string {
	// Use server configuration to build the invite link
	baseURL := "https://localhost:8090" // Default, should come from config
	if s.serverConf != nil {
		baseURL = fmt.Sprintf("https://%s:%d", s.serverConf.Hostname, s.serverConf.Port)
	}
	return fmt.Sprintf("%s/gate/invite?token=%s", baseURL, token)
}
