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

// Package invitation provides functionality for managing user invitations.
package invitation

import "time"

// Invitation status constants
const (
	// StatusPending indicates the invitation is waiting to be redeemed
	StatusPending = "pending"
	// StatusRedeemed indicates the invitation has been successfully redeemed
	StatusRedeemed = "redeemed"
	// StatusExpired indicates the invitation has expired
	StatusExpired = "expired"
	// StatusRevoked indicates the invitation was revoked by admin
	StatusRevoked = "revoked"
)

// Default expiration time for invitations (7 days in seconds)
const DefaultExpirationSeconds = 7 * 24 * 60 * 60

// Invitation represents an invitation for a user to complete their registration.
type Invitation struct {
	ID            string     `json:"id"`
	UserID        string     `json:"userId"`
	ApplicationID string     `json:"applicationId"`
	Token         string     `json:"token"`
	Status        string     `json:"status"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	RedeemedAt    *time.Time `json:"redeemedAt,omitempty"`
}

// CreateInvitationRequest represents the request to create an invitation.
type CreateInvitationRequest struct {
	UserID        string `json:"userId"`
	ApplicationID string `json:"applicationId"`
	ExpiresIn     int    `json:"expiresIn,omitempty"` // Seconds until expiration, defaults to 7 days
}

// CreateInvitationResponse represents the response after creating an invitation.
type CreateInvitationResponse struct {
	InvitationID string    `json:"invitationId"`
	UserID       string    `json:"userId"`
	Token        string    `json:"token"`
	InviteLink   string    `json:"inviteLink"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// ValidateInvitationResponse represents the response from validating an invitation token.
type ValidateInvitationResponse struct {
	Valid         bool   `json:"valid"`
	UserID        string `json:"userId,omitempty"`
	ApplicationID string `json:"applicationId,omitempty"`
	FlowID        string `json:"flowId,omitempty"`
	Message       string `json:"message,omitempty"`
}

// InvitationListResponse represents a paginated list of invitations.
type InvitationListResponse struct {
	TotalResults int          `json:"totalResults"`
	StartIndex   int          `json:"startIndex"`
	Count        int          `json:"count"`
	Invitations  []Invitation `json:"invitations"`
}
