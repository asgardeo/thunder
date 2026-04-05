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

// Package user provides user management functionality.
package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/asgardeo/thunder/internal/entity"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema"
)

const loggerComponentName = "UserService"

// UserServiceInterface defines the interface for the user service.
type UserServiceInterface interface {
	GetUserList(ctx context.Context, limit, offset int,
		filters map[string]interface{}, includeDisplay bool) (*UserListResponse, *serviceerror.ServiceError)
	GetUsersByPath(ctx context.Context, handlePath string, limit, offset int,
		filters map[string]interface{}, includeDisplay bool) (*UserListResponse, *serviceerror.ServiceError)
	CreateUser(ctx context.Context, user *User) (*User, *serviceerror.ServiceError)
	CreateUserByPath(ctx context.Context, handlePath string,
		request CreateUserByPathRequest) (*User, *serviceerror.ServiceError)
	GetUser(ctx context.Context, userID string, includeDisplay bool) (*User, *serviceerror.ServiceError)
	GetUserGroups(ctx context.Context, userID string,
		limit, offset int) (*UserGroupListResponse, *serviceerror.ServiceError)
	GetTransitiveUserGroups(ctx context.Context, userID string) ([]UserGroup, *serviceerror.ServiceError)
	UpdateUser(ctx context.Context, userID string, user *User) (*User, *serviceerror.ServiceError)
	UpdateUserAttributes(ctx context.Context, userID string,
		attributes json.RawMessage) (*User, *serviceerror.ServiceError)
	UpdateUserCredentials(ctx context.Context, userID string,
		credentials json.RawMessage) *serviceerror.ServiceError
	DeleteUser(ctx context.Context, userID string) *serviceerror.ServiceError
	IdentifyUser(ctx context.Context, filters map[string]interface{}) (*string, *serviceerror.ServiceError)
	ValidateUserIDs(ctx context.Context, userIDs []string) ([]string, *serviceerror.ServiceError)
	GetUsersByIDs(ctx context.Context, userIDs []string) (map[string]*User, *serviceerror.ServiceError)
	ValidateUserIDsInOUs(ctx context.Context, userIDs []string,
		ouIDs []string) ([]string, *serviceerror.ServiceError)
	GetUserCredentialsByType(ctx context.Context, userID string,
		credentialType string) ([]Credential, *serviceerror.ServiceError)
	IsUserDeclarative(ctx context.Context, userID string) (bool, *serviceerror.ServiceError)
}

// userService is the default implementation of the UserServiceInterface.
type userService struct {
	authzService      sysauthz.SystemAuthorizationServiceInterface
	entityService     entity.EntityServiceInterface
	ouService         oupkg.OrganizationUnitServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
}

// newUserService creates a new instance of userService with injected dependencies.
func newUserService(
	authzService sysauthz.SystemAuthorizationServiceInterface,
	entityService entity.EntityServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
) UserServiceInterface {
	return &userService{
		authzService:      authzService,
		entityService:     entityService,
		ouService:         ouService,
		userSchemaService: userSchemaService,
	}
}

// GetUserList retrieves a list of users with pagination and filtering.
func (us *userService) GetUserList(ctx context.Context, limit, offset int,
	filters map[string]interface{}, includeDisplay bool) (*UserListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	// Resolve the set of organization units the caller is authorized to list users from.
	accessible, svcErr := us.authzService.GetAccessibleResources(
		ctx, security.ActionListUsers, security.ResourceTypeOU)
	if svcErr != nil {
		logger.Error("Failed to resolve accessible resources for listing users", log.Any("error", svcErr))
		return nil, &ErrorInternalServerError
	}

	// Unfiltered path: system-level caller — return all users.
	if accessible.AllAllowed {
		return us.listAllUsers(ctx, limit, offset, filters, includeDisplay, logger)
	}

	// Filtered path: return users belonging to the accessible OUs.
	return us.listUsersByOUIDs(ctx, accessible.IDs, limit, offset, filters, includeDisplay, logger)
}

// listAllUsers retrieves users without OU filtering.
func (us *userService) listAllUsers(
	ctx context.Context, limit, offset int, filters map[string]interface{},
	includeDisplay bool, logger *log.Logger,
) (*UserListResponse, *serviceerror.ServiceError) {
	totalCount, err := us.entityService.GetEntityListCount(ctx, entity.EntityCategoryUser, filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list count", err)
	}

	entities, err := us.entityService.GetEntityList(ctx, entity.EntityCategoryUser, limit, offset, filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list", err)
	}

	users := entitiesToUsers(entities)
	if includeDisplay {
		us.populateUserDisplayNames(ctx, users, logger)
		us.populateOUHandles(ctx, users, logger)
	}

	return buildUserListResponse(users, totalCount, limit, offset, utils.DisplayQueryParam(includeDisplay)), nil
}

// listUsersByOUIDs retrieves users scoped to the given organization unit IDs.
func (us *userService) listUsersByOUIDs(
	ctx context.Context, ouIDs []string, limit, offset int, filters map[string]interface{},
	includeDisplay bool, logger *log.Logger,
) (*UserListResponse, *serviceerror.ServiceError) {
	displayQuery := utils.DisplayQueryParam(includeDisplay)

	if len(ouIDs) == 0 {
		return buildUserListResponse([]User{}, 0, limit, offset, displayQuery), nil
	}

	totalCount, err := us.entityService.GetEntityListCountByOUIDs(ctx, entity.EntityCategoryUser, ouIDs, filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list count", err)
	}

	entities, err := us.entityService.GetEntityListByOUIDs(
		ctx, entity.EntityCategoryUser, ouIDs, limit, offset, filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list", err)
	}

	users := entitiesToUsers(entities)
	if includeDisplay {
		us.populateUserDisplayNames(ctx, users, logger)
		us.populateOUHandles(ctx, users, logger)
	}

	return buildUserListResponse(users, totalCount, limit, offset, displayQuery), nil
}

// buildUserListResponse constructs a paginated UserListResponse.
func buildUserListResponse(users []User, totalCount, limit, offset int, displayQuery string) *UserListResponse {
	return &UserListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(users),
		Users:        users,
		Links:        utils.BuildPaginationLinks("/users", limit, offset, totalCount, displayQuery),
	}
}

// GetUsersByPath retrieves a list of users by hierarchical handle path.
func (us *userService) GetUsersByPath(
	ctx context.Context, handlePath string, limit, offset int, filters map[string]interface{},
	includeDisplay bool,
) (*UserListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Getting users by path", log.String("path", handlePath))

	serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, svcErr := us.ouService.GetOrganizationUnitByPath(ctx, handlePath)
	if svcErr != nil {
		return nil, mapOUServiceError(
			svcErr,
			logger,
			"resolving organization unit by path",
			map[string]*serviceerror.ServiceError{
				oupkg.ErrorOrganizationUnitNotFound.Code: &ErrorOrganizationUnitNotFound,
				oupkg.ErrorInvalidHandlePath.Code:        &ErrorInvalidHandlePath,
			},
			log.String("path", handlePath),
		)
	}
	oUID := ou.ID

	// Check if caller is authorized to list users in the resolved OU.
	if svcErr := us.checkUserAccess(ctx, security.ActionListUsers, oUID, ""); svcErr != nil {
		return nil, svcErr
	}

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	ouResponse, svcErr := us.ouService.GetOrganizationUnitUsers(ctx, oUID, limit, offset, false)
	if svcErr != nil {
		return nil, mapOUServiceError(
			svcErr,
			logger,
			"listing organization unit users",
			map[string]*serviceerror.ServiceError{
				oupkg.ErrorOrganizationUnitNotFound.Code: &ErrorOrganizationUnitNotFound,
				oupkg.ErrorInvalidLimit.Code:             &ErrorInvalidLimit,
				oupkg.ErrorInvalidOffset.Code:            &ErrorInvalidOffset,
			},
			log.String("oUID", oUID),
			log.Int("limit", limit),
			log.Int("offset", offset),
		)
	}
	if ouResponse == nil {
		return &UserListResponse{}, nil
	}

	var users []User
	if includeDisplay && len(ouResponse.Users) > 0 {
		// Batch-fetch full user data to resolve display names.
		userIDs := make([]string, len(ouResponse.Users))
		for i, ouUser := range ouResponse.Users {
			userIDs[i] = ouUser.ID
		}
		fetchedEntities, err := us.entityService.GetEntitiesByIDs(ctx, userIDs)
		if err != nil {
			logger.Warn("Failed to batch fetch users for display names, skipping display resolution", log.Error(err))
			// Fall back to bare IDs without display — partial display is worse than none.
			users = make([]User, len(ouResponse.Users))
			for i, ouUser := range ouResponse.Users {
				users[i] = User{ID: ouUser.ID, OUHandle: ou.Handle}
			}
		} else {
			fetchedUsers := entitiesToUsers(fetchedEntities)
			// Build an ID-keyed map for display resolution, but only expose ID + Display.
			userMap := make(map[string]User, len(fetchedUsers))
			for _, u := range fetchedUsers {
				userMap[u.ID] = u
			}

			// Resolve display attribute paths for the fetched user types.
			userTypes := make([]string, 0, len(fetchedUsers))
			for _, u := range fetchedUsers {
				userTypes = append(userTypes, u.Type)
			}
			displayAttrPaths := ResolveDisplayAttributePaths(ctx, userTypes, us.userSchemaService, logger)

			users = make([]User, len(ouResponse.Users))
			for i, ouUser := range ouResponse.Users {
				if u, ok := userMap[ouUser.ID]; ok {
					users[i] = User{
						ID:       u.ID,
						OUHandle: ou.Handle,
						Display:  utils.ResolveDisplay(u.ID, u.Type, u.Attributes, displayAttrPaths),
					}
				} else {
					users[i] = User{ID: ouUser.ID, OUHandle: ou.Handle}
				}
			}
		}
	} else {
		users = make([]User, len(ouResponse.Users))
		for i, ouUser := range ouResponse.Users {
			users[i] = User{ID: ouUser.ID}
		}
	}

	response := &UserListResponse{
		TotalResults: ouResponse.TotalResults,
		StartIndex:   ouResponse.StartIndex,
		Count:        ouResponse.Count,
		Users:        users,
		Links: buildTreePaginationLinks(
			handlePath, limit, offset, ouResponse.TotalResults, utils.DisplayQueryParam(includeDisplay)),
	}

	return response, nil
}

// CreateUser creates the user.
func (us *userService) CreateUser(ctx context.Context, user *User) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if user == nil {
		return nil, &ErrorInvalidRequestFormat
	}

	// Check if caller is authorized to create users in the target OU.
	if svcErr := us.checkUserAccess(ctx, security.ActionCreateUser, user.OUID, ""); svcErr != nil {
		return nil, svcErr
	}

	if svcErr := us.validateOrganizationUnitForUserType(ctx, user.Type, user.OUID, logger); svcErr != nil {
		return nil, svcErr
	}

	// Schema validation and uniqueness checks are handled by entity service in CreateEntity.

	var err error
	user.ID, err = utils.GenerateUUIDv7()
	if err != nil {
		logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	e := userToEntity(user)
	created, err := us.entityService.CreateEntity(ctx, e, nil)
	if err != nil {
		if svcErr := mapEntityError(err); svcErr != nil {
			return nil, svcErr
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to create user", err)
	}

	// Sync cleaned attributes back — entity service removed credential fields from Attributes.
	user.Attributes = created.Attributes

	logger.Debug("Successfully created user", log.String("id", user.ID))
	return user, nil
}

// CreateUserByPath creates a new user under the organization unit specified by the handle path.
func (us *userService) CreateUserByPath(
	ctx context.Context, handlePath string, request CreateUserByPathRequest,
) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating user by path", log.String("path", handlePath), log.String("type", request.Type))

	serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, svcErr := us.ouService.GetOrganizationUnitByPath(ctx, handlePath)
	if svcErr != nil {
		return nil, mapOUServiceError(
			svcErr,
			logger,
			"resolving organization unit by path",
			map[string]*serviceerror.ServiceError{
				oupkg.ErrorOrganizationUnitNotFound.Code: &ErrorOrganizationUnitNotFound,
				oupkg.ErrorInvalidHandlePath.Code:        &ErrorInvalidHandlePath,
			},
			log.String("path", handlePath),
		)
	}

	user := &User{
		OUID:       ou.ID,
		Type:       request.Type,
		Attributes: request.Attributes,
	}

	return us.CreateUser(ctx, user)
}

// GetUser retrieves a user by ID.
func (us *userService) GetUser(
	ctx context.Context, userID string, includeDisplay bool,
) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving user", log.String("id", userID))

	if userID == "" {
		return nil, &ErrorMissingUserID
	}

	e, err := us.entityService.GetEntity(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}
	if e.Category != entity.EntityCategoryUser {
		return nil, &ErrorUserNotFound
	}
	user := entityToUser(e)

	// Check authz using the user's OU ID (fetched from store).
	if svcErr := us.checkUserAccess(ctx, security.ActionReadUser, user.OUID, userID); svcErr != nil {
		return nil, svcErr
	}

	if includeDisplay {
		displayAttrPaths := ResolveDisplayAttributePaths(
			ctx, []string{user.Type}, us.userSchemaService, logger)
		user.Display = utils.ResolveDisplay(
			user.ID, user.Type, user.Attributes, displayAttrPaths)

		handleMap, svcErr := us.ouService.GetOrganizationUnitHandlesByIDs(ctx, []string{user.OUID})
		if svcErr != nil {
			logger.Warn("Failed to resolve OU handle for user, skipping",
				log.Any("error", svcErr))
		} else if handle, ok := handleMap[user.OUID]; ok {
			user.OUHandle = handle
		}
	}

	logger.Debug("Successfully retrieved user", log.String("id", userID))
	return &user, nil
}

// GetUserGroups retrieves groups of a user with pagination.
func (as *userService) GetUserGroups(ctx context.Context, userID string, limit, offset int) (
	*UserGroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if userID == "" {
		return nil, &ErrorMissingUserID
	}

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	// Fetch user to resolve the OU ID for the authorization check.
	userEntity, err := as.entityService.GetEntity(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}
	if userEntity.Category != entity.EntityCategoryUser {
		return nil, &ErrorUserNotFound
	}

	// Check authz using the user's OU ID.
	if svcErr := as.checkUserAccess(
		ctx, security.ActionReadUser, userEntity.OrganizationUnitID, userID); svcErr != nil {
		return nil, svcErr
	}

	totalCount, err := as.entityService.GetGroupCountForEntity(ctx, userID)
	if err != nil {
		logger.Error("Failed to get group count for user", log.String("userID", userID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	entityGroups, err := as.entityService.GetEntityGroups(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user groups", log.String("id", userID), log.Error(err))
		return nil, &ErrorInternalServerError
	}
	path := fmt.Sprintf("/users/%s/groups", userID)
	links := utils.BuildPaginationLinks(path, limit, offset, totalCount, "")

	response := &UserGroupListResponse{
		TotalResults: totalCount,
		Groups:       entityGroups,
		StartIndex:   offset + 1,
		Count:        len(entityGroups),
		Links:        links,
	}

	return response, nil
}

// GetTransitiveUserGroups retrieves all groups a user belongs to, including groups inherited
// through nested group membership. This is used internally for auth flows (token claims,
// authorization checks) and does not support pagination.
func (as *userService) GetTransitiveUserGroups(ctx context.Context, userID string) (
	[]UserGroup, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if userID == "" {
		return nil, &ErrorMissingUserID
	}

	entityGroups, err := as.entityService.GetTransitiveEntityGroups(ctx, userID)
	if err != nil {
		logger.Error("Failed to get transitive user groups", log.String("userID", userID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groups := make([]UserGroup, 0, len(entityGroups))
	for _, g := range entityGroups {
		groups = append(groups, UserGroup{ID: g.ID, Name: g.Name, OUID: g.OUID})
	}
	return groups, nil
}

// UpdateUser update the user for given user id.
func (us *userService) UpdateUser(ctx context.Context, userID string, user *User) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating user", log.String("id", userID))

	if userID == "" {
		return nil, &ErrorMissingUserID
	}

	if user == nil {
		return nil, &ErrorInvalidRequestFormat
	}

	// Fetch the existing user to obtain its OU ID for the authorization check.
	existingEntity, err := us.entityService.GetEntity(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}
	if existingEntity.Category != entity.EntityCategoryUser {
		return nil, &ErrorUserNotFound
	}
	existingUser := entityToUser(existingEntity)

	// Check authz using the existing user's OU ID.
	if svcErr := us.checkUserAccess(
		ctx, security.ActionUpdateUser, existingUser.OUID, userID); svcErr != nil {
		return nil, svcErr
	}

	// If the user is moving to a different OU, require authorization for the destination OU as well.
	if user.OUID != existingUser.OUID {
		if svcErr := us.checkUserAccess(
			ctx, security.ActionUpdateUser, user.OUID, userID); svcErr != nil {
			return nil, svcErr
		}
	}

	// Check if user is declarative (immutable)
	if svcErr := us.checkUserDeclarative(ctx, userID, logger); svcErr != nil {
		return nil, svcErr
	}

	// Ensure the user object has the correct ID
	user.ID = userID

	if svcErr := us.validateOrganizationUnitForUserType(
		ctx, user.Type, user.OUID, logger,
	); svcErr != nil {
		return nil, svcErr
	}

	// Entity service handles schema validation, credential extraction from attributes,
	// hashing, merging with existing credentials, and entity update.
	e := userToEntity(user)
	e.SystemAttributes = existingEntity.SystemAttributes
	_, err = us.entityService.UpdateEntity(ctx, userID, e)
	if err != nil {
		if svcErr := mapEntityError(err); svcErr != nil {
			return nil, svcErr
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to update user", err, log.String("id", userID))
	}

	logger.Debug("Successfully updated user", log.String("id", userID))
	return user, nil
}

// UpdateUserAttributes updates only the attributes of a user while preserving immutable fields.
func (us *userService) UpdateUserAttributes(
	ctx context.Context, userID string, attributes json.RawMessage,
) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating user attributes", log.String("id", userID))

	if strings.TrimSpace(userID) == "" {
		return nil, &ErrorMissingUserID
	}

	if len(attributes) == 0 {
		return nil, &ErrorInvalidRequestFormat
	}

	// Pre-fetch user to get the type for credential field lookup (outside transaction).
	existingEntity, getErr := us.entityService.GetEntity(ctx, userID)
	if getErr != nil {
		if errors.Is(getErr, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to get user", getErr, log.String("id", userID))
	}
	if existingEntity.Category != entity.EntityCategoryUser {
		return nil, &ErrorUserNotFound
	}
	existingUser := entityToUser(existingEntity)

	if us.userSchemaService == nil {
		logger.Error("User schema service is not configured for user operations")
		return nil, &ErrorInternalServerError
	}

	schemaCredentialAttributes, svcErr := us.userSchemaService.GetCredentialAttributes(ctx, existingUser.Type)
	if svcErr != nil {
		if svcErr.Code == userschema.ErrorUserSchemaNotFound.Code {
			return nil, &ErrorUserSchemaNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to get credential attributes from schema",
			fmt.Errorf("schema service error: %s", svcErr.ErrorDescription), log.String("id", userID))
	}

	hasCredentials, svcErr := us.containsCredentialAttributes(attributes, schemaCredentialAttributes)
	if svcErr != nil {
		return nil, svcErr
	}
	if hasCredentials {
		return nil, &ErrorInvalidRequestFormat
	}

	// Check authz outside the transaction so a denial is returned directly without a rollback.
	if svcErr := us.checkUserAccess(
		ctx, security.ActionUpdateUser, existingUser.OUID, userID); svcErr != nil {
		return nil, svcErr
	}

	// Check if user is declarative (immutable)
	if svcErr := us.checkUserDeclarative(ctx, userID, logger); svcErr != nil {
		return nil, svcErr
	}

	existingUser.Attributes = attributes

	if svcErr := us.validateUserAndUniqueness(ctx, existingUser.Type,
		existingUser.Attributes, logger, userID, true); svcErr != nil {
		return nil, svcErr
	}

	if err := us.entityService.UpdateAttributes(ctx, userID, attributes); err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to update user attributes", err,
			log.String("id", userID))
	}

	logger.Debug("Successfully updated user attributes", log.String("id", userID))
	return &existingUser, nil
}

// containsCredentialAttributes checks whether the attributes include credential attributes
// (either schema-defined or system-managed).
func (us *userService) containsCredentialAttributes(
	attributes json.RawMessage, schemaCredentialAttributes []string,
) (bool, *serviceerror.ServiceError) {
	if len(attributes) == 0 {
		return false, nil
	}

	var attrs map[string]any
	if err := json.Unmarshal(attributes, &attrs); err != nil {
		return false, &ErrorInvalidRequestFormat
	}

	for _, credField := range schemaCredentialAttributes {
		if _, ok := attrs[credField]; ok {
			return true, nil
		}
	}

	for _, credType := range systemManagedCredentialTypes {
		if _, ok := attrs[string(credType)]; ok {
			return true, nil
		}
	}

	return false, nil
}

// UpdateUserCredentials updates the credentials of a user.
func (us *userService) UpdateUserCredentials(
	ctx context.Context,
	userID string,
	credentials json.RawMessage,
) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating user credentials", log.String("userID", userID))

	if strings.TrimSpace(userID) == "" {
		return &ErrorAuthenticationFailed
	}

	if len(credentials) == 0 {
		return &ErrorMissingCredentials
	}

	// Parse credentials to extract credential types
	var credentialsMap map[string]json.RawMessage
	if err := json.Unmarshal(credentials, &credentialsMap); err != nil {
		logger.Debug("Failed to parse credentials", log.Error(err))
		return &ErrorInvalidRequestFormat
	}

	if len(credentialsMap) == 0 {
		return &ErrorMissingCredentials
	}

	// Delegate to batch update method
	return us.batchUpdateUserCredentials(ctx, userID, credentialsMap)
}

// batchUpdateUserCredentials updates multiple user credentials within a single transaction.
func (us *userService) batchUpdateUserCredentials(
	ctx context.Context,
	userID string,
	credentialsMap map[string]json.RawMessage,
) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Batch updating user credentials",
		log.String("userID", userID),
		log.Int("credentialTypesCount", len(credentialsMap)))

	// Fetch user outside the transaction to resolve the OU ID for the authorization check.
	existingEntity, err := us.entityService.GetEntity(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("userID", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("userID", userID))
	}
	if existingEntity.Category != entity.EntityCategoryUser {
		return &ErrorUserNotFound
	}
	existingUser := entityToUser(existingEntity)

	// Check authz outside the transaction so a denial is returned directly without a rollback.
	if svcErr := us.checkUserAccess(
		ctx, security.ActionUpdateUser, existingUser.OUID, userID); svcErr != nil {
		return svcErr
	}

	// Check if user is declarative (immutable)
	if svcErr := us.checkUserDeclarative(ctx, userID, logger); svcErr != nil {
		return svcErr
	}

	// Get schema credential attributes for the user's type
	if us.userSchemaService == nil {
		return logErrorAndReturnServerError(logger, "User schema service not configured",
			errors.New("user schema service not configured"), log.String("userID", userID))
	}

	schemaCredentialAttributes, svcErr := us.userSchemaService.GetCredentialAttributes(ctx, existingUser.Type)
	if svcErr != nil {
		if svcErr.Code == userschema.ErrorUserSchemaNotFound.Code {
			return &ErrorUserSchemaNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to get credential attributes from schema",
			fmt.Errorf("schema service error: %s", svcErr.ErrorDescription), log.String("userID", userID))
	}

	// Build set of valid credential field names
	validCredentialAttributes := make(
		map[string]struct{}, len(schemaCredentialAttributes)+len(systemManagedCredentialTypes))
	for _, field := range schemaCredentialAttributes {
		validCredentialAttributes[field] = struct{}{}
	}
	for _, credType := range systemManagedCredentialTypes {
		validCredentialAttributes[string(credType)] = struct{}{}
	}

	// Validate credential types and build plaintext map for entity service.
	plaintextCreds := make(map[string]interface{})
	for credTypeStr, credValue := range credentialsMap {
		// Validate credential type against schema + system-managed types.
		if _, valid := validCredentialAttributes[credTypeStr]; !valid {
			logger.Debug("Invalid credential type", log.String("credentialType", credTypeStr))
			errorDesc := fmt.Sprintf("Invalid credential type: %s", credTypeStr)
			return serviceerror.CustomServiceError(ErrorInvalidCredential, errorDesc)
		}

		if len(credValue) == 0 {
			return &ErrorMissingCredentials
		}

		// Try to parse as a plain string value.
		var stringValue string
		if err := json.Unmarshal(credValue, &stringValue); err == nil {
			plaintextCreds[credTypeStr] = stringValue
		} else {
			// Pass structured values through as-is (e.g., passkey objects).
			var structuredValue interface{}
			if err := json.Unmarshal(credValue, &structuredValue); err != nil {
				return &ErrorInvalidRequestFormat
			}
			plaintextCreds[credTypeStr] = structuredValue
		}
	}

	// Delegate hashing and merging to entity service.
	plaintextJSON, err := json.Marshal(plaintextCreds)
	if err != nil {
		return logErrorAndReturnServerError(logger, "Failed to marshal credentials", err,
			log.String("userID", userID))
	}
	if err = us.entityService.UpdateSystemCredentials(ctx, userID, plaintextJSON); err != nil {
		if svcErr := mapEntityError(err); svcErr != nil {
			return svcErr
		}
		return logErrorAndReturnServerError(logger, "Failed to update user credentials", err,
			log.String("userID", userID))
	}

	logger.Debug("Successfully updated user credentials",
		log.String("userID", userID),
		log.Int("credentialTypesCount", len(credentialsMap)))
	return nil
}

// DeleteUser delete the user for given user id.
func (us *userService) DeleteUser(ctx context.Context, userID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Deleting user", log.String("id", userID))

	if userID == "" {
		return &ErrorMissingUserID
	}

	// Fetch the user to resolve the OU ID for the authorization check.
	existingEntity, err := us.entityService.GetEntity(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}
	if existingEntity.Category != entity.EntityCategoryUser {
		return &ErrorUserNotFound
	}
	existingUser := entityToUser(existingEntity)

	// Check authz using the user's OU ID.
	if svcErr := us.checkUserAccess(
		ctx, security.ActionDeleteUser, existingUser.OUID, userID); svcErr != nil {
		return svcErr
	}

	// Check if user is declarative (immutable)
	if svcErr := us.checkUserDeclarative(ctx, userID, logger); svcErr != nil {
		return svcErr
	}

	err = us.entityService.DeleteEntity(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to delete user", err, log.String("id", userID))
	}

	logger.Debug("Successfully deleted user", log.String("id", userID))
	return nil
}

// IdentifyUser identifies a user with the given filters.
func (us *userService) IdentifyUser(ctx context.Context,
	filters map[string]interface{}) (*string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(filters) == 0 {
		return nil, &ErrorInvalidRequestFormat
	}

	userID, err := us.entityService.IdentifyEntity(ctx, filters)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found with provided filters")
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to identify user", err)
	}

	return userID, nil
}

// ValidateUserIDs validates that all provided user IDs exist.
func (us *userService) ValidateUserIDs(ctx context.Context, userIDs []string) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(userIDs) == 0 {
		return []string{}, nil
	}

	invalidUserIDs, err := us.entityService.ValidateEntityIDs(ctx, userIDs)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to validate user IDs", err)
	}

	return invalidUserIDs, nil
}

// GetUsersByIDs retrieves users by a list of IDs.
func (us *userService) GetUsersByIDs(
	ctx context.Context, userIDs []string,
) (map[string]*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(userIDs) == 0 {
		return map[string]*User{}, nil
	}

	// Deduplicate IDs before passing to store.
	seen := make(map[string]struct{}, len(userIDs))
	uniqueIDs := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	entities, err := us.entityService.GetEntitiesByIDs(ctx, uniqueIDs)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get users by IDs", err)
	}

	users := entitiesToUsers(entities)
	result := make(map[string]*User, len(users))
	for i := range users {
		result[users[i].ID] = &users[i]
	}

	return result, nil
}

// populateUserDisplayNames resolves display names for a slice of users in-place.
// It batch-fetches display attribute paths from the user schema service and extracts the
// display value from each user's attributes. Falls back to user ID if extraction fails.
func (us *userService) populateUserDisplayNames(ctx context.Context, users []User, logger *log.Logger) {
	// Collect user types for display attribute resolution.
	userTypes := make([]string, 0, len(users))
	for _, u := range users {
		userTypes = append(userTypes, u.Type)
	}

	displayAttrPaths := ResolveDisplayAttributePaths(
		ctx, userTypes, us.userSchemaService, logger)

	// Resolve display for each user.
	for i := range users {
		users[i].Display = utils.ResolveDisplay(
			users[i].ID, users[i].Type, users[i].Attributes, displayAttrPaths)
	}
}

// populateOUHandles resolves OU handles for a slice of users in-place.
func (us *userService) populateOUHandles(ctx context.Context, users []User, logger *log.Logger) {
	ouIDs := make([]string, 0, len(users))
	seen := make(map[string]bool, len(users))
	for _, u := range users {
		if u.OUID != "" && !seen[u.OUID] {
			ouIDs = append(ouIDs, u.OUID)
			seen[u.OUID] = true
		}
	}

	handleMap, svcErr := us.ouService.GetOrganizationUnitHandlesByIDs(ctx, ouIDs)
	if svcErr != nil {
		logger.Warn("Failed to resolve OU handles, skipping", log.Any("error", svcErr))
		return
	}

	for i := range users {
		if handle, ok := handleMap[users[i].OUID]; ok {
			users[i].OUHandle = handle
		}
	}
}

// ValidateUserIDsInOUs validates that all provided user IDs belong to one of the given OUs.
// Returns IDs that are outside the allowed OU scope.
func (us *userService) ValidateUserIDsInOUs(
	ctx context.Context, userIDs []string, ouIDs []string,
) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(userIDs) == 0 {
		return []string{}, nil
	}
	if len(ouIDs) == 0 {
		// No accessible OUs — all IDs are out of scope.
		return append([]string{}, userIDs...), nil
	}

	outOfScopeIDs, err := us.entityService.ValidateEntityIDsInOUs(ctx, userIDs, ouIDs)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to validate user IDs in OUs", err)
	}
	return outOfScopeIDs, nil
}

// GetUserCredentialsByType retrieves credentials of a specific type for a user.
// Returns an empty array if no credentials of the specified type exist.
func (us *userService) GetUserCredentialsByType(
	ctx context.Context,
	userID string,
	credentialType string,
) ([]Credential, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving user credentials by type",
		log.String("userID", log.MaskString(userID)),
		log.String("credentialType", credentialType))

	if strings.TrimSpace(userID) == "" {
		return nil, &ErrorMissingUserID
	}

	if strings.TrimSpace(credentialType) == "" {
		logger.Debug("Credential type is empty")
		return nil, &ErrorInvalidRequestFormat
	}

	// Get all credentials for the user (from both CREDENTIALS and SYSTEM_CREDENTIALS columns).
	credEntity, schemaCredsJSON, systemCredsJSON, err := us.entityService.GetEntityWithCredentials(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("userID", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(
			logger,
			"Failed to retrieve user credentials",
			err,
			log.String("userID", userID),
		)
	}
	if credEntity.Category != entity.EntityCategoryUser {
		return nil, &ErrorUserNotFound
	}
	allCredentials := make(Credentials)
	if len(schemaCredsJSON) > 0 {
		if schemaCreds, unmarshalErr := jsonToCredentials(schemaCredsJSON); unmarshalErr == nil {
			for k, v := range schemaCreds {
				allCredentials[k] = v
			}
		}
	}
	if len(systemCredsJSON) > 0 {
		if sysCreds, unmarshalErr := jsonToCredentials(systemCredsJSON); unmarshalErr == nil {
			for k, v := range sysCreds {
				allCredentials[k] = v
			}
		}
	}

	// Get credentials of the specified type
	credentials, exists := allCredentials[CredentialType(credentialType)]
	if !exists || len(credentials) == 0 {
		logger.Debug("No credentials found for type",
			log.String("userID", log.MaskString(userID)),
			log.String("credentialType", credentialType))
		// Return empty array
		return []Credential{}, nil
	}

	logger.Debug("Retrieved credentials for type",
		log.String("userID", log.MaskString(userID)),
		log.String("credentialType", credentialType),
		log.Int("count", len(credentials)))

	return credentials, nil
}

// IsUserDeclarative checks if a user is immutable (declarative) or mutable.
func (us *userService) IsUserDeclarative(ctx context.Context, userID string) (bool, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if strings.TrimSpace(userID) == "" {
		return false, &ErrorMissingUserID
	}

	isDeclarative, err := us.entityService.IsEntityDeclarative(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			logger.Debug("User not found", log.String("userID", userID))
			return false, &ErrorUserNotFound
		}
		return false, logErrorAndReturnServerError(logger, "Failed to check if user is declarative", err)
	}

	return isDeclarative, nil
}

// validateOrganizationUnitForUserType ensures that the organization unit ID is valid and belongs to the user type.
func (us *userService) validateOrganizationUnitForUserType(
	ctx context.Context, userType, oUID string, logger *log.Logger,
) *serviceerror.ServiceError {
	if strings.TrimSpace(userType) == "" {
		return &ErrorUserSchemaNotFound
	}

	if strings.TrimSpace(oUID) == "" {
		return &ErrorInvalidOUID
	}

	if us.ouService == nil {
		logger.Error("Organization unit service is not configured for user operations")
		return &ErrorInternalServerError
	}

	exists, svcErr := us.ouService.IsOrganizationUnitExists(ctx, oUID)
	if svcErr != nil {
		return mapOUServiceError(
			svcErr,
			logger,
			"verifying organization unit existence",
			map[string]*serviceerror.ServiceError{
				oupkg.ErrorOrganizationUnitNotFound.Code: &ErrorOrganizationUnitNotFound,
				oupkg.ErrorInvalidRequestFormat.Code:     &ErrorInvalidOUID,
				oupkg.ErrorMissingOUID.Code:              &ErrorInvalidOUID,
			},
			log.String("oUID", oUID),
		)
	}
	if !exists {
		return &ErrorOrganizationUnitNotFound
	}

	if us.userSchemaService == nil {
		logger.Error("User schema service is not configured for user operations")
		return &ErrorInternalServerError
	}

	userSchema, svcErr := us.userSchemaService.GetUserSchemaByName(ctx, userType)
	if svcErr != nil {
		if svcErr.Code == userschema.ErrorUserSchemaNotFound.Code {
			return &ErrorUserSchemaNotFound
		}
		logger.Error("Failed to retrieve user schema",
			log.String("userType", userType), log.Any("error", svcErr))
		return &ErrorInternalServerError
	}

	if userSchema == nil {
		logger.Error("User schema service returned nil response", log.String("userType", userType))
		return &ErrorInternalServerError
	}

	if userSchema.OUID == oUID {
		return nil
	}

	isParent, svcErr := us.ouService.IsParent(ctx, userSchema.OUID, oUID)
	if svcErr != nil {
		return mapOUServiceError(
			svcErr,
			logger,
			"validating organization unit hierarchy",
			map[string]*serviceerror.ServiceError{
				oupkg.ErrorOrganizationUnitNotFound.Code: &ErrorOrganizationUnitNotFound,
			},
			log.String("userType", userType),
			log.String("oUID", oUID),
			log.String("schemaOUID", userSchema.OUID),
		)
	}

	if !isParent {
		logger.Debug("Organization unit mismatch for user type",
			log.String("userType", userType),
			log.String("oUID", oUID),
			log.String("schemaOUID", userSchema.OUID))
		return &ErrorOrganizationUnitMismatch
	}

	return nil
}

// validateUserAndUniqueness validates the user schema and checks for uniqueness.
func (us *userService) validateUserAndUniqueness(
	ctx context.Context, userType string, attributes []byte, logger *log.Logger, excludeUserID string,
	skipCredentialRequired bool,
) *serviceerror.ServiceError {
	isValid, svcErr := us.userSchemaService.ValidateUser(ctx, userType, attributes, skipCredentialRequired)
	if svcErr != nil {
		if svcErr.Code == userschema.ErrorUserSchemaNotFound.Code {
			return &ErrorUserSchemaNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to validate user schema", nil)
	}
	if !isValid {
		return &ErrorSchemaValidationFailed
	}

	isValid, svcErr = us.userSchemaService.ValidateUserUniqueness(ctx, userType, attributes,
		func(filters map[string]interface{}) (*string, error) {
			userID, svcErr := us.IdentifyUser(ctx, filters)
			if svcErr != nil {
				if svcErr.Code == ErrorUserNotFound.Code {
					return nil, nil
				} else {
					return nil, errors.New(svcErr.Error)
				}
			}
			if excludeUserID != "" && userID != nil && *userID == excludeUserID {
				return nil, nil
			}
			return userID, nil
		})
	if svcErr != nil {
		if svcErr.Code == userschema.ErrorUserSchemaNotFound.Code {
			return &ErrorUserSchemaNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to validate user schema", nil)
	}

	if !isValid {
		return &ErrorAttributeConflict
	}

	return nil
}

// validateAndProcessHandlePath validates and processes the handle path.
func validateAndProcessHandlePath(handlePath string) *serviceerror.ServiceError {
	if strings.TrimSpace(handlePath) == "" {
		return &ErrorInvalidHandlePath
	}

	handles := strings.Split(strings.Trim(handlePath, "/"), "/")
	if len(handles) == 0 {
		return &ErrorInvalidHandlePath
	}

	for _, handle := range handles {
		if strings.TrimSpace(handle) == "" {
			return &ErrorInvalidHandlePath
		}
	}
	return nil
}

// validatePaginationParams validates pagination parameters.
func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 1 || limit > serverconst.MaxPageSize {
		return &ErrorInvalidLimit
	}
	if offset < 0 {
		return &ErrorInvalidOffset
	}
	return nil
}

// logErrorAndReturnServerError logs the error and returns a server error.
func logErrorAndReturnServerError(
	logger *log.Logger,
	message string,
	err error,
	additionalFields ...log.Field,
) *serviceerror.ServiceError {
	fields := additionalFields
	if err != nil {
		fields = append(fields, log.Error(err))
	}
	logger.Error(message, fields...)
	return &ErrorInternalServerError
}

// mapEntityError maps entity service errors to user service errors.
// Returns nil if the error is not a recognized entity error.
func mapEntityError(err error) *serviceerror.ServiceError {
	switch {
	case errors.Is(err, entity.ErrEntityNotFound):
		return &ErrorUserNotFound
	case errors.Is(err, entity.ErrAuthenticationFailed):
		return &ErrorAuthenticationFailed
	case errors.Is(err, entity.ErrSchemaValidationFailed):
		return &ErrorSchemaValidationFailed
	case errors.Is(err, entity.ErrAttributeConflict):
		return &ErrorAttributeConflict
	case errors.Is(err, entity.ErrInvalidCredential):
		return &ErrorInvalidCredential
	default:
		return nil
	}
}

// mapOUServiceError converts organization unit service errors to user service errors.
func mapOUServiceError(
	svcErr *serviceerror.ServiceError,
	logger *log.Logger,
	context string,
	mappings map[string]*serviceerror.ServiceError,
	fields ...log.Field,
) *serviceerror.ServiceError {
	if svcErr == nil {
		return nil
	}

	if mappedErr, ok := mappings[svcErr.Code]; ok {
		return mappedErr
	}

	if svcErr.Type == serviceerror.ClientErrorType {
		logFields := append([]log.Field{}, fields...)
		logFields = append(logFields, log.Any("error", svcErr))
		logger.Error(fmt.Sprintf("Unexpected organization unit client error while %s", context), logFields...)
		return &ErrorInternalServerError
	}

	logFields := append([]log.Field{}, fields...)
	logFields = append(logFields, log.Any("error", svcErr))
	logger.Error(fmt.Sprintf("Organization unit service error while %s", context), logFields...)
	return &ErrorInternalServerError
}

// checkUserDeclarative checks if a user is declarative and returns an error if it is.
func (us *userService) checkUserDeclarative(
	ctx context.Context, userID string, logger *log.Logger,
) *serviceerror.ServiceError {
	isDeclarative, err := us.entityService.IsEntityDeclarative(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return &ErrorUserNotFound
		}
		logger.Error("Failed to check if user is declarative",
			log.String("userID", userID), log.Error(err))
		return &ErrorInternalServerError
	}
	if isDeclarative {
		return &ErrorCannotModifyDeclarativeResource
	}
	return nil
}

// checkUserAccess validates that the caller is authorized to perform the given action on a user.
func (us *userService) checkUserAccess(
	ctx context.Context, action security.Action, ouID string, resourceID string,
) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	allowed, svcErr := us.authzService.IsActionAllowed(ctx, action,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeUser, OUID: ouID, ResourceID: resourceID})
	if svcErr != nil {
		logger.Error("Failed to check authorization for action",
			log.String("action", string(action)), log.Any("error", svcErr))
		return &ErrorInternalServerError
	}
	if !allowed {
		return &serviceerror.ErrorUnauthorized
	}
	return nil
}

// buildTreePaginationLinks builds pagination links for user responses.
func buildTreePaginationLinks(handlePath string, limit, offset, totalResults int, displayQuery string) []utils.Link {
	treePath := fmt.Sprintf("/users/tree/%s", path.Clean(handlePath))
	return utils.BuildPaginationLinks(treePath, limit, offset, totalResults, displayQuery)
}
