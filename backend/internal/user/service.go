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

	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
	"github.com/asgardeo/thunder/internal/system/transaction"
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
	UpdateUser(ctx context.Context, userID string, user *User) (*User, *serviceerror.ServiceError)
	UpdateUserAttributes(ctx context.Context, userID string,
		attributes json.RawMessage) (*User, *serviceerror.ServiceError)
	UpdateUserCredentials(ctx context.Context, userID string,
		credentials json.RawMessage) *serviceerror.ServiceError
	DeleteUser(ctx context.Context, userID string) *serviceerror.ServiceError
	IdentifyUser(ctx context.Context, filters map[string]interface{}) (*string, *serviceerror.ServiceError)
	VerifyUser(ctx context.Context, userID string,
		credentials map[string]interface{}) (*User, *serviceerror.ServiceError)
	AuthenticateUser(ctx context.Context,
		identifiers map[string]interface{},
		credentials map[string]interface{}) (*AuthenticateUserResponse, *serviceerror.ServiceError)
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
	userStore         userStoreInterface
	ouService         oupkg.OrganizationUnitServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
	hashService       hash.HashServiceInterface
	transactioner     transaction.Transactioner
}

// newUserService creates a new instance of userService with injected dependencies.
func newUserService(
	authzService sysauthz.SystemAuthorizationServiceInterface,
	userStore userStoreInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	hashService hash.HashServiceInterface,
	transactioner transaction.Transactioner,
) UserServiceInterface {
	return &userService{
		authzService:      authzService,
		userStore:         userStore,
		ouService:         ouService,
		userSchemaService: userSchemaService,
		hashService:       hashService,
		transactioner:     transactioner,
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
	totalCount, err := us.userStore.GetUserListCount(ctx, filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list count", err)
	}

	users, err := us.userStore.GetUserList(ctx, limit, offset, filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list", err)
	}

	if includeDisplay {
		us.populateUserDisplayNames(ctx, users, logger)
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

	totalCount, err := us.userStore.GetUserListCountByOUIDs(ctx, ouIDs, filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list count", err)
	}

	users, err := us.userStore.GetUserListByOUIDs(ctx, ouIDs, limit, offset, filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list", err)
	}

	if includeDisplay {
		us.populateUserDisplayNames(ctx, users, logger)
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
		fetchedUsers, err := us.userStore.GetUsersByIDs(ctx, userIDs)
		if err != nil {
			logger.Warn("Failed to batch fetch users for display names, skipping display resolution", log.Error(err))
			// Fall back to bare IDs without display — partial display is worse than none.
			users = make([]User, len(ouResponse.Users))
			for i, ouUser := range ouResponse.Users {
				users[i] = User{ID: ouUser.ID}
			}
		} else {
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
						ID:      u.ID,
						Display: utils.ResolveDisplay(u.ID, u.Type, u.Attributes, displayAttrPaths),
					}
				} else {
					users[i] = User{ID: ouUser.ID}
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

	if svcErr := us.validateUserAndUniqueness(ctx, user.Type, user.Attributes, logger, "", false); svcErr != nil {
		return nil, svcErr
	}

	var err error
	user.ID, err = utils.GenerateUUIDv7()
	if err != nil {
		logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	schemaCredentialAttributes, svcErr := us.userSchemaService.GetCredentialAttributes(ctx, user.Type)
	if svcErr != nil {
		if svcErr.Code == userschema.ErrorUserSchemaNotFound.Code {
			return nil, &ErrorUserSchemaNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to get credential attributes from schema",
			fmt.Errorf("schema service error: %s", svcErr.ErrorDescription))
	}

	credentials, err := us.extractCredentials(user, schemaCredentialAttributes)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to create user DTO", err)
	}

	// Use transaction to ensure atomic user creation with indexed attributes
	err = us.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return us.userStore.CreateUser(txCtx, *user, credentials)
	})
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to create user", err)
	}

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

// extractCredentials extracts credentials from user attributes based on schema-defined credential attributes.
// Schema-defined credentials are always hashed. System-managed credentials are also extracted defensively.
func (us *userService) extractCredentials(user *User, schemaCredentialAttributes []string) (Credentials, error) {
	if user.Attributes == nil {
		return Credentials{}, nil
	}

	var attrsMap map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrsMap); err != nil {
		return nil, err
	}

	credentials := make(Credentials)

	// Extract schema-defined credential attributes (always hashed).
	for _, credField := range schemaCredentialAttributes {
		if credValue, ok := attrsMap[credField].(string); ok {
			delete(attrsMap, credField)

			// Skip empty credential values.
			if credValue == "" {
				continue
			}

			credHash, err := us.hashService.Generate([]byte(credValue))
			if err != nil {
				return nil, err
			}

			credential := Credential{
				StorageType: "hash",
				StorageAlgo: credHash.Algorithm,
				StorageAlgoParams: hash.CredParameters{
					Iterations: credHash.Parameters.Iterations,
					KeySize:    credHash.Parameters.KeySize,
					Salt:       credHash.Parameters.Salt,
				},
				Value: credHash.Hash,
			}

			credType := CredentialType(credField)
			if credentials[credType] == nil {
				credentials[credType] = []Credential{}
			}
			credentials[credType] = append(credentials[credType], credential)
		}
	}

	// Extract system-managed credential types defensively.
	for _, credType := range systemManagedCredentialTypes {
		credField := string(credType)
		if credValue, ok := attrsMap[credField].(string); ok {
			delete(attrsMap, credField)

			// Skip empty credential values.
			if credValue == "" {
				continue
			}

			credential := Credential{
				Value: credValue,
			}

			if credentials[credType] == nil {
				credentials[credType] = []Credential{}
			}
			credentials[credType] = append(credentials[credType], credential)
		}
	}

	if len(credentials) > 0 {
		updatedAttrs, err := json.Marshal(attrsMap)
		if err != nil {
			return nil, err
		}
		user.Attributes = updatedAttrs
	}

	return credentials, nil
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

	user, err := us.userStore.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}

	// Check authz using the user's OU ID (fetched from store).
	if svcErr := us.checkUserAccess(ctx, security.ActionReadUser, user.OUID, userID); svcErr != nil {
		return nil, svcErr
	}

	if includeDisplay {
		displayAttrPaths := ResolveDisplayAttributePaths(
			ctx, []string{user.Type}, us.userSchemaService, logger)
		user.Display = utils.ResolveDisplay(
			user.ID, user.Type, user.Attributes, displayAttrPaths)
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
	user, err := as.userStore.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}

	// Check authz using the user's OU ID.
	if svcErr := as.checkUserAccess(ctx, security.ActionReadUser, user.OUID, userID); svcErr != nil {
		return nil, svcErr
	}

	totalCount, err := as.userStore.GetGroupCountForUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to get group count for user", log.MaskedString(log.LoggerKeyUserID, userID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groups, err := as.userStore.GetUserGroups(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user groups", log.String("id", userID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	path := fmt.Sprintf("/users/%s/groups", userID)
	links := utils.BuildPaginationLinks(path, limit, offset, totalCount, "")

	response := &UserGroupListResponse{
		TotalResults: totalCount,
		Groups:       groups,
		StartIndex:   offset + 1,
		Count:        len(groups),
		Links:        links,
	}

	return response, nil
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
	existingUser, err := us.userStore.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}

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

	if us.userSchemaService == nil {
		logger.Error("User schema service is not configured for user operations")
		return nil, &ErrorInternalServerError
	}

	schemaCredentialAttributes, svcErr := us.userSchemaService.GetCredentialAttributes(ctx, user.Type)
	if svcErr != nil {
		if svcErr.Code == userschema.ErrorUserSchemaNotFound.Code {
			return nil, &ErrorUserSchemaNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to get credential attributes from schema",
			fmt.Errorf("schema service error: %s", svcErr.ErrorDescription), log.String("id", userID))
	}

	var capturedSvcErr *serviceerror.ServiceError

	err = us.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if svcErr := us.validateOrganizationUnitForUserType(
			txCtx, user.Type, user.OUID, logger,
		); svcErr != nil {
			capturedSvcErr = svcErr
			return errors.New("rollback for validation error")
		}

		// Validate before extracting credentials so that credential values (e.g. regex)
		// are still checked when present. skipCredentialRequired is true because
		// credentials are optional during updates.
		if svcErr := us.validateUserAndUniqueness(
			txCtx, user.Type, user.Attributes, logger, user.ID, true,
		); svcErr != nil {
			capturedSvcErr = svcErr
			return errors.New("rollback for validation error")
		}

		credentials, err := us.extractCredentials(user, schemaCredentialAttributes)
		if err != nil {
			return fmt.Errorf("failed to extract credentials: %w", err)
		}

		err = us.userStore.UpdateUser(txCtx, user)
		if err != nil {
			return err
		}

		if len(credentials) > 0 {
			_, existingCredentials, err := us.userStore.GetCredentials(txCtx, userID)
			if err != nil {
				return err
			}
			mergedCredentials := us.mergeCredentials(existingCredentials, credentials)
			err = us.userStore.UpdateUserCredentials(txCtx, userID, mergedCredentials)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if capturedSvcErr != nil {
		return nil, capturedSvcErr
	}

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
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
	existingUser, getErr := us.userStore.GetUser(ctx, userID)
	if getErr != nil {
		if errors.Is(getErr, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to get user", getErr, log.String("id", userID))
	}

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

	var updatedUser User
	var capturedSvcErr *serviceerror.ServiceError

	err := us.transactioner.Transact(ctx, func(txCtx context.Context) error {
		existingUser.Attributes = attributes

		if svcErr := us.validateUserAndUniqueness(txCtx, existingUser.Type,
			existingUser.Attributes, logger, userID, true); svcErr != nil {
			capturedSvcErr = svcErr
			return errors.New("rollback for validation error")
		}

		err := us.userStore.UpdateUser(txCtx, &existingUser)
		if err != nil {
			return err
		}
		updatedUser = existingUser
		return nil
	})

	if capturedSvcErr != nil {
		return nil, capturedSvcErr
	}

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to update user attributes", err,
			log.String("id", userID))
	}

	logger.Debug("Successfully updated user attributes", log.String("id", userID))
	return &updatedUser, nil
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
	logger.Debug("Updating user credentials", log.MaskedString(log.LoggerKeyUserID, userID))

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
		log.MaskedString(log.LoggerKeyUserID, userID),
		log.Int("credentialTypesCount", len(credentialsMap)))

	// Fetch user outside the transaction to resolve the OU ID for the authorization check.
	existingUser, err := us.userStore.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.MaskedString(log.LoggerKeyUserID, userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.MaskedString(log.LoggerKeyUserID, userID))
	}

	// Check authz outside the transaction so a denial is returned directly without a rollback.
	if svcErr := us.checkUserAccess(
		ctx, security.ActionUpdateUser, existingUser.OUID, userID); svcErr != nil {
		return svcErr
	}

	// Check if user is declarative (immutable)
	if svcErr := us.checkUserDeclarative(ctx, userID, logger); svcErr != nil {
		return svcErr
	}

	var capturedSvcErr *serviceerror.ServiceError

	err = us.transactioner.Transact(ctx, func(txCtx context.Context) error {
		// Get existing credentials and user info
		existingUser, existingCredentials, err := us.userStore.GetCredentials(txCtx, userID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				logger.Debug("User not found", log.MaskedString(log.LoggerKeyUserID, userID))
				capturedSvcErr = &ErrorUserNotFound
				return errors.New("rollback for user not found")
			}

			return err
		}

		// Get schema credential attributes for the user's type
		if us.userSchemaService == nil {
			return errors.New("user schema service not configured")
		}

		schemaCredentialAttributes, svcErr := us.userSchemaService.GetCredentialAttributes(txCtx, existingUser.Type)
		if svcErr != nil {
			if svcErr.Code == userschema.ErrorUserSchemaNotFound.Code {
				capturedSvcErr = &ErrorUserSchemaNotFound
				return errors.New("rollback for schema not found")
			}
			return fmt.Errorf("schema service error: %s", svcErr.ErrorDescription)
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

		// Process all credential types first (validation and hashing)
		processedCredentials := make(Credentials)
		for credTypeStr, credValue := range credentialsMap {
			credType := CredentialType(credTypeStr)

			// Validate credential type against schema + system-managed types
			if _, valid := validCredentialAttributes[credTypeStr]; !valid {
				logger.Debug("Invalid credential type", log.String("credentialType", credTypeStr))
				errorDesc := fmt.Sprintf("Invalid credential type: %s", credType)
				capturedSvcErr = serviceerror.CustomServiceError(ErrorInvalidCredential, errorDesc)
				return errors.New("rollback for validation error")
			}

			if len(credValue) == 0 {
				capturedSvcErr = &ErrorMissingCredentials
				return errors.New("rollback for validation error")
			}

			// Process and validate credentials for this type
			processed, svcErr := us.processCredentialType(credType, credValue, logger)
			if svcErr != nil {
				capturedSvcErr = svcErr
				return errors.New("rollback for validation error")
			}

			processedCredentials[credType] = processed
		}

		// Merge all processed credentials with existing ones
		updatedCredentials := us.mergeCredentials(existingCredentials, processedCredentials)

		// Update credentials in database
		err = us.userStore.UpdateUserCredentials(txCtx, userID, updatedCredentials)
		if err != nil {
			return err
		}
		return nil
	})

	if capturedSvcErr != nil {
		return capturedSvcErr
	}

	if err != nil {
		return logErrorAndReturnServerError(
			logger,
			"Failed to update user credentials",
			err,
			log.MaskedString(log.LoggerKeyUserID, userID),
		)
	}

	logger.Debug("Successfully batch updated user credentials",
		log.MaskedString(log.LoggerKeyUserID, userID),
		log.Int("credentialTypesCount", len(credentialsMap)))
	return nil
}

// processCredentialType processes and validates credentials for a single credential type.
// It handles parsing, validation, and hashing for credential types that require it.
func (us *userService) processCredentialType(
	credentialType CredentialType,
	credentialValue json.RawMessage,
	logger *log.Logger,
) ([]Credential, *serviceerror.ServiceError) {
	var credentials []Credential

	// Try to parse as array of Credential first
	if err := json.Unmarshal(credentialValue, &credentials); err != nil {
		// If not an array, try parsing as a plain string value
		var stringValue string
		if err := json.Unmarshal(credentialValue, &stringValue); err != nil {
			logger.Debug("Failed to parse credential value",
				log.String("credentialType", string(credentialType)),
				log.Error(err))
			return nil, &ErrorInvalidRequestFormat
		}
		// Convert string value to Credential array
		credentials = []Credential{{Value: stringValue}}
	}

	// System-managed credentials (e.g., passkey) support multiple values.
	// Schema-defined credentials only support a single value.
	if !credentialType.IsSystemManaged() && len(credentials) > 1 {
		logger.Debug("Multiple credentials not supported for this credential type",
			log.String("credentialType", string(credentialType)),
			log.Int("count", len(credentials)))
		errorDesc := fmt.Sprintf("Credential type '%s' does not support multiple credentials. "+
			"Only one credential is allowed.", credentialType)
		return nil, serviceerror.CustomServiceError(ErrorInvalidCredential, errorDesc)
	}

	// Validate credentials
	for i := range credentials {
		if err := us.validateCredential(&credentials[i]); err != nil {
			logger.Debug("Credential validation failed",
				log.String("credentialType", string(credentialType)),
				log.Int("index", i),
				log.Error(err))
			return nil, &ErrorInvalidCredential
		}
	}

	// Schema-defined credentials are always hashed. System-managed credentials are stored as-is.
	if !credentialType.IsSystemManaged() {
		hashedCredentials, svcErr := us.hashCredentials(credentials, credentialType, logger)
		if svcErr != nil {
			return nil, svcErr
		}
		return hashedCredentials, nil
	}

	return credentials, nil
}

// hashCredentials hashes all credentials in the provided list.
func (us *userService) hashCredentials(
	credentials []Credential,
	credType CredentialType,
	logger *log.Logger,
) ([]Credential, *serviceerror.ServiceError) {
	hashedCredentials := make([]Credential, 0, len(credentials))
	for _, cred := range credentials {
		credHash, err := us.hashService.Generate([]byte(cred.Value))
		if err != nil {
			logger.Error("Failed to hash credential",
				log.String("credentialType", string(credType)),
				log.Error(err))
			return nil, &ErrorInternalServerError
		}

		hashedCred := Credential{
			StorageType: "hash",
			StorageAlgo: credHash.Algorithm,
			StorageAlgoParams: hash.CredParameters{
				Iterations: credHash.Parameters.Iterations,
				KeySize:    credHash.Parameters.KeySize,
				Salt:       credHash.Parameters.Salt,
			},
			Value: credHash.Hash,
		}
		hashedCredentials = append(hashedCredentials, hashedCred)
	}

	return hashedCredentials, nil
}

// mergeCredentials merges processed credentials with existing credentials.
// Processed credentials replace existing ones for their types, while other types are preserved.
func (us *userService) mergeCredentials(existing Credentials, processed Credentials) Credentials {
	merged := make(Credentials)

	// Copy existing credentials
	for credType, credList := range existing {
		merged[credType] = append([]Credential{}, credList...)
	}

	// Replace with processed credentials
	for credType, credList := range processed {
		merged[credType] = credList
	}

	return merged
}

// validateCredential validates a single credential.
func (us *userService) validateCredential(credential *Credential) error {
	if credential == nil {
		return errors.New("credential is nil")
	}
	if strings.TrimSpace(credential.Value) == "" {
		return errors.New("credential value is empty")
	}
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
	existingUser, err := us.userStore.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}

	// Check authz using the user's OU ID.
	if svcErr := us.checkUserAccess(
		ctx, security.ActionDeleteUser, existingUser.OUID, userID); svcErr != nil {
		return svcErr
	}

	// Check if user is declarative (immutable)
	if svcErr := us.checkUserDeclarative(ctx, userID, logger); svcErr != nil {
		return svcErr
	}

	err = us.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return us.userStore.DeleteUser(txCtx, userID)
	})

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
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

	userID, err := us.userStore.IdentifyUser(ctx, filters)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found with provided filters")
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to identify user", err)
	}

	return userID, nil
}

// VerifyUser validate the specified user with the given credentials.
func (us *userService) VerifyUser(
	ctx context.Context, userID string, credentials map[string]interface{},
) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if userID == "" {
		return nil, &ErrorMissingUserID
	}

	if len(credentials) == 0 {
		return nil, &ErrorInvalidRequestFormat
	}

	user, storedCredentials, err := us.userStore.GetCredentials(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to verify user", err, log.String("id", userID))
	}

	if len(storedCredentials) == 0 {
		logger.Debug("No credentials found for user", log.MaskedString(log.LoggerKeyUserID, userID))
		return nil, &ErrorAuthenticationFailed
	}

	// Filter credentials to verify: only include those that have stored credential keys.
	credentialsToVerify := make(map[string]string)
	for credType, credValueInterface := range credentials {
		if _, exists := storedCredentials[CredentialType(credType)]; !exists {
			continue
		}

		credValue, ok := credValueInterface.(string)
		if !ok || credValue == "" {
			continue
		}

		credentialsToVerify[credType] = credValue
	}

	if len(credentialsToVerify) == 0 {
		logger.Debug("No valid credentials provided for verification", log.MaskedString(log.LoggerKeyUserID, userID))
		return nil, &ErrorAuthenticationFailed
	}

	for credType, credValue := range credentialsToVerify {
		credList := storedCredentials[CredentialType(credType)]

		// Try to verify against any credential of this type (typically first one)
		verified := false
		for _, storedCred := range credList {
			verifyingCredential := hash.Credential{
				Algorithm: storedCred.StorageAlgo,
				Hash:      storedCred.Value,
				Parameters: hash.CredParameters{
					Salt:       storedCred.StorageAlgoParams.Salt,
					Iterations: storedCred.StorageAlgoParams.Iterations,
					KeySize:    storedCred.StorageAlgoParams.KeySize,
				},
			}
			hashVerified, err := us.hashService.Verify([]byte(credValue), verifyingCredential)

			if err == nil && hashVerified {
				logger.Debug("Credential verified successfully",
					log.MaskedString(log.LoggerKeyUserID, userID), log.String("credType", credType))
				verified = true
				break
			}
		}

		if !verified {
			logger.Debug("Credential verification failed",
				log.MaskedString(log.LoggerKeyUserID, userID), log.String("credType", credType))
			return nil, &ErrorAuthenticationFailed
		}
	}

	logger.Debug("Successfully verified all user credentials", log.String("id", userID))
	return &user, nil
}

// AuthenticateUser authenticates a user by combining identify and verify operations.
// Identifiers are used to find the user, and credentials are verified against stored values.
func (us *userService) AuthenticateUser(
	ctx context.Context,
	identifiers map[string]interface{},
	credentials map[string]interface{},
) (*AuthenticateUserResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Authenticating user")

	if len(identifiers) == 0 {
		return nil, &ErrorMissingRequiredFields
	}

	if len(credentials) == 0 {
		return nil, &ErrorMissingCredentials
	}

	userID, svcErr := us.IdentifyUser(ctx, identifiers)
	if svcErr != nil {
		if svcErr.Code == ErrorUserNotFound.Code {
			return nil, &ErrorUserNotFound
		}
		return nil, svcErr
	}

	user, svcErr := us.VerifyUser(ctx, *userID, credentials)
	if svcErr != nil {
		return nil, svcErr
	}

	logger.Debug("User authenticated successfully", log.MaskedString(log.LoggerKeyUserID, *userID))
	return &AuthenticateUserResponse{
		ID:   user.ID,
		Type: user.Type,
		OUID: user.OUID,
	}, nil
}

// ValidateUserIDs validates that all provided user IDs exist.
func (us *userService) ValidateUserIDs(ctx context.Context, userIDs []string) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(userIDs) == 0 {
		return []string{}, nil
	}

	invalidUserIDs, err := us.userStore.ValidateUserIDs(ctx, userIDs)
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

	users, err := us.userStore.GetUsersByIDs(ctx, uniqueIDs)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get users by IDs", err)
	}

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

	outOfScopeIDs, err := us.userStore.ValidateUserIDsInOUs(ctx, userIDs, ouIDs)
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
		log.MaskedString(log.LoggerKeyUserID, userID),
		log.String("credentialType", credentialType))

	if strings.TrimSpace(userID) == "" {
		return nil, &ErrorMissingUserID
	}

	if strings.TrimSpace(credentialType) == "" {
		logger.Debug("Credential type is empty")
		return nil, &ErrorInvalidRequestFormat
	}

	// Get all credentials for the user
	_, allCredentials, err := us.userStore.GetCredentials(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.MaskedString(log.LoggerKeyUserID, userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(
			logger,
			"Failed to retrieve user credentials",
			err,
			log.MaskedString(log.LoggerKeyUserID, userID),
		)
	}

	// Get credentials of the specified type
	credentials, exists := allCredentials[CredentialType(credentialType)]
	if !exists || len(credentials) == 0 {
		logger.Debug("No credentials found for type",
			log.MaskedString(log.LoggerKeyUserID, userID),
			log.String("credentialType", credentialType))
		// Return empty array
		return []Credential{}, nil
	}

	logger.Debug("Retrieved credentials for type",
		log.MaskedString(log.LoggerKeyUserID, userID),
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

	isDeclarative, err := us.userStore.IsUserDeclarative(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.MaskedString(log.LoggerKeyUserID, userID))
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
	isDeclarative, err := us.userStore.IsUserDeclarative(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return &ErrorUserNotFound
		}
		logger.Error("Failed to check if user is declarative",
			log.MaskedString(log.LoggerKeyUserID, userID), log.Error(err))
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
