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

// Package common defines the common models and functions for authentication handling.
package common

import (
	"context"
	"time"

	"github.com/asgardeo/thunder/internal/idp"
)

// AuthenticatedUser represents the user information of an authenticated user.
type AuthenticatedUser struct {
	IsAuthenticated    bool
	UserID             string
	OrganizationUnitID string
	UserType           string
	Attributes         map[string]interface{}
}

// AuthenticationContext represents the context of an authentication session.
type AuthenticationContext struct {
	context.Context
	SessionDataKey     string
	RequestQueryParams map[string]string
	AuthenticatedUser  AuthenticatedUser
	AuthTime           time.Time
}

// AuthenticationResponse represents the response after successful authentication.
type AuthenticationResponse struct {
	ID               string
	Type             string
	OrganizationUnit string
	Assertion        string
}

// AuthenticatorMeta represents an authenticator's metadata including authentication factors.
type AuthenticatorMeta struct {
	// Name is the unique identifier for the authenticator (used in individual authentication APIs)
	Name string
	// Factors represents the authentication factors this authenticator validates
	Factors []AuthenticationFactor
	// AssociatedIDP is the optional identity provider type this authenticator is associated with.
	AssociatedIDP idp.IDPType
}

// AuthenticatorReference represents an engaged authenticator in the authentication flow.
type AuthenticatorReference struct {
	// Authenticator is the name of the authenticator
	Authenticator string `json:"authenticator"`
	// Step is the step number in the flow where this authenticator was engaged
	Step int `json:"step"`
	// Timestamp is the authenticator engaged time (Unix epoch time in seconds)
	Timestamp int64 `json:"timestamp"`
}
