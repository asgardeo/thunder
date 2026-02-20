/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

// Package userprovider provides user management functionality.
package userprovider

import (
	"context"
	"encoding/json"

	"github.com/asgardeo/thunder/internal/user"
)

type defaultUserProvider struct {
	userSvc user.UserServiceInterface
}

// newDefaultUserProvider creates a new default user provider.
func newDefaultUserProvider(userSvc user.UserServiceInterface) UserProviderInterface {
	return &defaultUserProvider{
		userSvc: userSvc,
	}
}

// IdentifyUser identifies a user based on the given filters.
func (p *defaultUserProvider) IdentifyUser(filters map[string]interface{}) (*string, *UserProviderError) {
	userID, err := p.userSvc.IdentifyUser(context.Background(), filters)
	if err != nil {
		if err.Code == user.ErrorUserNotFound.Code {
			return nil, NewUserProviderError(ErrorCodeUserNotFound, err.Error, err.ErrorDescription)
		}
		return nil, NewUserProviderError(ErrorCodeSystemError, err.Error, err.ErrorDescription)
	}

	return userID, nil
}

// GetUser retrieves a user based on the given user ID.
func (p *defaultUserProvider) GetUser(userID string) (*User, *UserProviderError) {
	userResult, err := p.userSvc.GetUser(context.Background(), userID)
	if err != nil {
		if err.Code == user.ErrorUserNotFound.Code {
			return nil, NewUserProviderError(ErrorCodeUserNotFound, err.Error, err.ErrorDescription)
		}
		return nil, NewUserProviderError(ErrorCodeSystemError, err.Error, err.ErrorDescription)
	}

	return &User{
		UserID:             userResult.ID,
		UserType:           userResult.Type,
		OrganizationUnitID: userResult.OrganizationUnit,
		Attributes:         userResult.Attributes,
	}, nil
}

// GetUserGroups retrieves the groups associated with a user based on the given user ID.
func (p *defaultUserProvider) GetUserGroups(userID string, limit, offset int) (*UserGroupListResponse,
	*UserProviderError) {
	userGroupListResponse, err := p.userSvc.GetUserGroups(context.Background(), userID, limit, offset)
	if err != nil {
		if err.Code == user.ErrorUserNotFound.Code || err.Code == user.ErrorMissingUserID.Code {
			return nil, NewUserProviderError(ErrorCodeUserNotFound, err.Error, err.ErrorDescription)
		}
		return nil, NewUserProviderError(ErrorCodeSystemError, err.Error, err.ErrorDescription)
	}

	groups := make([]UserGroup, len(userGroupListResponse.Groups))
	for i, g := range userGroupListResponse.Groups {
		groups[i] = UserGroup{
			ID:                 g.ID,
			Name:               g.Name,
			OrganizationUnitID: g.OrganizationUnitID,
		}
	}

	links := make([]Link, len(userGroupListResponse.Links))
	for i, l := range userGroupListResponse.Links {
		links[i] = Link{
			Href: l.Href,
			Rel:  l.Rel,
		}
	}

	return &UserGroupListResponse{
		TotalResults: userGroupListResponse.TotalResults,
		StartIndex:   userGroupListResponse.StartIndex,
		Count:        userGroupListResponse.Count,
		Groups:       groups,
		Links:        links,
	}, nil
}

// UpdateUser updates a user based on the given user ID and user update configuration.
func (p *defaultUserProvider) UpdateUser(userID string, userUpdateConfig *User) (*User, *UserProviderError) {
	if userUpdateConfig == nil {
		return nil, NewUserProviderError(ErrorCodeInvalidRequestFormat, "Invalid request",
			"User update configuration cannot be nil")
	}
	updatedUser := &user.User{
		ID:               userID,
		OrganizationUnit: userUpdateConfig.OrganizationUnitID,
		Type:             userUpdateConfig.UserType,
		Attributes:       userUpdateConfig.Attributes,
	}

	userResult, err := p.userSvc.UpdateUser(context.Background(), userID, updatedUser)
	if err != nil {
		switch err.Code {
		case user.ErrorUserNotFound.Code, user.ErrorMissingUserID.Code:
			return nil, NewUserProviderError(ErrorCodeUserNotFound, err.Error, err.ErrorDescription)
		case user.ErrorInvalidRequestFormat.Code:
			return nil, NewUserProviderError(ErrorCodeInvalidRequestFormat, err.Error, err.ErrorDescription)
		case user.ErrorOrganizationUnitMismatch.Code:
			return nil, NewUserProviderError(ErrorCodeOrganizationUnitMismatch, err.Error, err.ErrorDescription)
		case user.ErrorAttributeConflict.Code, user.ErrorEmailConflict.Code:
			return nil, NewUserProviderError(ErrorCodeAttributeConflict, err.Error, err.ErrorDescription)
		default:
			return nil, NewUserProviderError(ErrorCodeSystemError, err.Error, err.ErrorDescription)
		}
	}
	return &User{
		UserID:             userResult.ID,
		UserType:           userResult.Type,
		OrganizationUnitID: userResult.OrganizationUnit,
		Attributes:         userResult.Attributes,
	}, nil
}

// CreateUser creates a new user based on the given user create configuration.
func (p *defaultUserProvider) CreateUser(userCreateConfig *User) (*User, *UserProviderError) {
	if userCreateConfig == nil {
		return nil, NewUserProviderError(ErrorCodeInvalidRequestFormat, "Invalid request",
			"User create configuration cannot be nil")
	}
	newUser := &user.User{
		OrganizationUnit: userCreateConfig.OrganizationUnitID,
		Type:             userCreateConfig.UserType,
		Attributes:       userCreateConfig.Attributes,
	}

	userResult, err := p.userSvc.CreateUser(context.Background(), newUser)
	if err != nil {
		switch err.Code {
		case user.ErrorInvalidRequestFormat.Code:
			return nil, NewUserProviderError(ErrorCodeInvalidRequestFormat, err.Error, err.ErrorDescription)
		case user.ErrorOrganizationUnitMismatch.Code:
			return nil, NewUserProviderError(ErrorCodeOrganizationUnitMismatch, err.Error, err.ErrorDescription)
		case user.ErrorAttributeConflict.Code, user.ErrorEmailConflict.Code:
			return nil, NewUserProviderError(ErrorCodeAttributeConflict, err.Error, err.ErrorDescription)
		case user.ErrorMissingRequiredFields.Code:
			return nil, NewUserProviderError(ErrorCodeMissingRequiredFields, err.Error, err.ErrorDescription)
		case user.ErrorOrganizationUnitNotFound.Code:
			return nil, NewUserProviderError(ErrorCodeOrganizationUnitMismatch, err.Error, err.ErrorDescription)
		default:
			return nil, NewUserProviderError(ErrorCodeSystemError, err.Error, err.ErrorDescription)
		}
	}

	return &User{
		UserID:             userResult.ID,
		UserType:           userResult.Type,
		OrganizationUnitID: userResult.OrganizationUnit,
		Attributes:         userResult.Attributes,
	}, nil
}

// UpdateUserCredentials updates the credentials of a user based on the given user ID and credentials.
func (p *defaultUserProvider) UpdateUserCredentials(userID string, credentials json.RawMessage) *UserProviderError {
	err := p.userSvc.UpdateUserCredentials(context.Background(), userID, credentials)
	if err != nil {
		switch err.Code {
		case user.ErrorInvalidRequestFormat.Code:
			return NewUserProviderError(ErrorCodeInvalidRequestFormat, err.Error, err.ErrorDescription)
		case user.ErrorMissingCredentials.Code:
			return NewUserProviderError(ErrorCodeMissingCredentials, err.Error, err.ErrorDescription)
		case user.ErrorUserNotFound.Code, user.ErrorMissingUserID.Code:
			return NewUserProviderError(ErrorCodeUserNotFound, err.Error, err.ErrorDescription)
		case user.ErrorAuthenticationFailed.Code:
			// Map auth failed (e.g. empty user ID) to invalid request or not found depending on semantics
			return NewUserProviderError(ErrorCodeInvalidRequestFormat, err.Error, err.ErrorDescription)
		default:
			return NewUserProviderError(ErrorCodeSystemError, err.Error, err.ErrorDescription)
		}
	}
	return nil
}

// DeleteUser deletes a user based on the given user ID.
func (p *defaultUserProvider) DeleteUser(userID string) *UserProviderError {
	err := p.userSvc.DeleteUser(context.Background(), userID)
	if err != nil {
		switch err.Code {
		case user.ErrorUserNotFound.Code, user.ErrorMissingUserID.Code:
			return nil
		default:
			return NewUserProviderError(ErrorCodeSystemError, err.Error, err.ErrorDescription)
		}
	}
	return nil
}
