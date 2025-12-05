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
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema"
)

const loggerComponentName = "UserService"

// SupportedCredentialFields defines the set of credential field names that are supported.
var supportedCredentialFields = map[string]struct{}{
	"password": {},
	"pin":      {},
	"secret":   {},
}

// UserServiceInterface defines the interface for the user service.
type UserServiceInterface interface {
	GetUserList(limit, offset int, filters map[string]interface{}) (*UserListResponse, *serviceerror.ServiceError)
	GetUsersByPath(handlePath string, limit, offset int,
		filters map[string]interface{}) (*UserListResponse, *serviceerror.ServiceError)
	CreateUser(user *User) (*User, *serviceerror.ServiceError)
	CreateUserByPath(handlePath string, request CreateUserByPathRequest) (*User, *serviceerror.ServiceError)
	GetUser(userID string) (*User, *serviceerror.ServiceError)
	GetUserGroups(userID string, limit, offset int) (*UserGroupListResponse, *serviceerror.ServiceError)
	UpdateUser(userID string, user *User) (*User, *serviceerror.ServiceError)
	UpdateUserAttributes(userID string, attributes json.RawMessage) (*User, *serviceerror.ServiceError)
	UpdateUserCredentials(userID string, attributes json.RawMessage) *serviceerror.ServiceError
	DeleteUser(userID string) *serviceerror.ServiceError
	IdentifyUser(filters map[string]interface{}) (*string, *serviceerror.ServiceError)
	VerifyUser(userID string, credentials map[string]interface{}) (*User, *serviceerror.ServiceError)
	AuthenticateUser(request AuthenticateUserRequest) (*AuthenticateUserResponse, *serviceerror.ServiceError)
	ValidateUserIDs(userIDs []string) ([]string, *serviceerror.ServiceError)
}

// userService is the default implementation of the UserServiceInterface.
type userService struct {
	userStore         userStoreInterface
	ouService         oupkg.OrganizationUnitServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
}

// newUserService creates a new instance of userService with injected dependencies.
func newUserService(
	userStore userStoreInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
) UserServiceInterface {
	return &userService{
		userStore:         userStore,
		ouService:         ouService,
		userSchemaService: userSchemaService,
	}
}

// GetUserList lists the users.
func (us *userService) GetUserList(limit, offset int,
	filters map[string]interface{}) (*UserListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := us.userStore.GetUserListCount(filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list count", err)
	}

	users, err := us.userStore.GetUserList(limit, offset, filters)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list", err)
	}

	response := &UserListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(users),
		Users:        users,
		Links:        buildPaginationLinks("/users", limit, offset, totalCount),
	}

	return response, nil
}

// GetUsersByPath retrieves a list of users by hierarchical handle path.
func (us *userService) GetUsersByPath(
	handlePath string, limit, offset int, filters map[string]interface{},
) (*UserListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Getting users by path", log.String("path", handlePath))

	serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, svcErr := us.ouService.GetOrganizationUnitByPath(handlePath)
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
	organizationUnitID := ou.ID

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	ouResponse, svcErr := us.ouService.GetOrganizationUnitUsers(organizationUnitID, limit, offset)
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
			log.String("organizationUnitID", organizationUnitID),
			log.Int("limit", limit),
			log.Int("offset", offset),
		)
	}

	users := make([]User, len(ouResponse.Users))
	for i, ouUser := range ouResponse.Users {
		users[i] = User{
			ID: ouUser.ID,
		}
	}

	response := &UserListResponse{
		TotalResults: ouResponse.TotalResults,
		StartIndex:   ouResponse.StartIndex,
		Count:        ouResponse.Count,
		Users:        users,
		Links:        buildTreePaginationLinks(handlePath, limit, offset, ouResponse.TotalResults),
	}

	return response, nil
}

// CreateUser creates the user.
func (us *userService) CreateUser(user *User) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if user == nil {
		return nil, &ErrorInvalidRequestFormat
	}

	if svcErr := us.validateOrganizationUnitForUserType(user.Type, user.OrganizationUnit, logger); svcErr != nil {
		return nil, svcErr
	}

	if svcErr := us.validateUserAndUniqueness(user.Type, user.Attributes, logger); svcErr != nil {
		return nil, svcErr
	}

	user.ID = utils.GenerateUUID()

	credentials, err := extractCredentials(user)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to create user DTO", err)
	}

	err = us.userStore.CreateUser(*user, credentials)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to create user", err)
	}

	logger.Debug("Successfully created user", log.String("id", user.ID))
	return user, nil
}

// CreateUserByPath creates a new user under the organization unit specified by the handle path.
func (us *userService) CreateUserByPath(
	handlePath string, request CreateUserByPathRequest,
) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating user by path", log.String("path", handlePath), log.String("type", request.Type))

	serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, svcErr := us.ouService.GetOrganizationUnitByPath(handlePath)
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
		OrganizationUnit: ou.ID,
		Type:             request.Type,
		Attributes:       request.Attributes,
	}

	return us.CreateUser(user)
}

// extractCredentials extracts the credentials from the user attributes and returns a Credentials array.
func extractCredentials(user *User) ([]Credential, error) {
	if user.Attributes == nil {
		return []Credential{}, nil
	}

	var attrsMap map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrsMap); err != nil {
		return nil, err
	}

	var credentials []Credential

	for credField := range supportedCredentialFields {
		if credValue, ok := attrsMap[credField].(string); ok {
			credHash := hash.Generate([]byte(credValue))

			delete(attrsMap, credField)

			credential := Credential{
				CredentialType: credField,
				StorageType:    "hash",
				StorageAlgo:    credHash.Algorithm,
				Value:          credHash.Hash,
				Salt:           credHash.Salt,
			}

			credentials = append(credentials, credential)
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

// GetUser get the user for given user id.
func (us *userService) GetUser(userID string) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving user", log.String("id", userID))

	if userID == "" {
		return nil, &ErrorMissingUserID
	}

	user, err := us.userStore.GetUser(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}

	logger.Debug("Successfully retrieved user", log.String("id", userID))
	return &user, nil
}

// GetUserGroups retrieves groups of a user with pagination.
func (as *userService) GetUserGroups(userID string, limit, offset int) (
	*UserGroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if userID == "" {
		return nil, &ErrorMissingUserID
	}

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	invalidUserIDs, err := as.userStore.ValidateUserIDs([]string{userID})
	if err != nil {
		logger.Error("Failed to validate user IDs", log.String("error", err.Error()))
		return nil, &ErrorInternalServerError
	}

	if len(invalidUserIDs) > 0 {
		logger.Debug("User not found", log.String("id", userID))
		return nil, &ErrorUserNotFound
	}

	totalCount, err := as.userStore.GetGroupCountForUser(userID)
	if err != nil {
		logger.Error("Failed to get group count for user", log.String("userID", userID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groups, err := as.userStore.GetUserGroups(userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user groups", log.String("id", userID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	path := fmt.Sprintf("/users/%s/groups", userID)
	links := buildPaginationLinks(path, limit, offset, totalCount)

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
func (us *userService) UpdateUser(userID string, user *User) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating user", log.String("id", userID))

	if userID == "" {
		return nil, &ErrorMissingUserID
	}

	if user == nil {
		return nil, &ErrorInvalidRequestFormat
	}

	if svcErr := us.validateOrganizationUnitForUserType(user.Type, user.OrganizationUnit, logger); svcErr != nil {
		return nil, svcErr
	}

	if svcErr := us.validateUserAndUniqueness(user.Type, user.Attributes, logger); svcErr != nil {
		return nil, svcErr
	}

	err := us.userStore.UpdateUser(user)
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
	userID string, attributes json.RawMessage,
) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating user attributes", log.String("id", userID))

	if strings.TrimSpace(userID) == "" {
		return nil, &ErrorMissingUserID
	}

	if len(attributes) == 0 {
		return nil, &ErrorInvalidRequestFormat
	}

	hasCredentials, svcErr := us.containsCredentialFields(attributes)
	if svcErr != nil {
		return nil, svcErr
	}
	if hasCredentials {
		return nil, &ErrorInvalidRequestFormat
	}

	user, err := us.userStore.GetUser(userID)
	if errors.Is(err, ErrUserNotFound) {
		logger.Debug("User not found", log.String("id", userID))
		return nil, &ErrorUserNotFound
	}

	user.Attributes = attributes

	if svcErr := us.validateUserAndUniqueness(user.Type, user.Attributes, logger); svcErr != nil {
		return nil, svcErr
	}

	err = us.userStore.UpdateUser(&user)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to update user attributes", err,
			log.String("id", userID))
	}

	logger.Debug("Successfully updated user attributes", log.String("id", userID))
	return &user, nil
}

// containsCredentialFields checks whether the attributes include credential fields.
func (us *userService) containsCredentialFields(attributes json.RawMessage) (bool, *serviceerror.ServiceError) {
	if len(attributes) == 0 {
		return false, nil
	}

	var attrs map[string]any
	if err := json.Unmarshal(attributes, &attrs); err != nil {
		return false, &ErrorInvalidRequestFormat
	}

	for credField := range supportedCredentialFields {
		if _, ok := attrs[credField]; ok {
			return true, nil
		}
	}

	return false, nil
}

// UpdateUserCredentials updates the credentials of a user.
func (us *userService) UpdateUserCredentials(userID string, attributes json.RawMessage) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating user credentials", log.String("id", userID))

	if strings.TrimSpace(userID) == "" {
		return &ErrorAuthenticationFailed
	}

	if len(attributes) == 0 {
		return &ErrorMissingCredentials
	}

	var attrs map[string]interface{}
	if err := json.Unmarshal(attributes, &attrs); err != nil {
		return &ErrorInvalidRequestFormat
	}
	if len(attrs) == 0 {
		return &ErrorMissingCredentials
	}

	var invalidCredentialFields []string
	for key := range attrs {
		if _, found := supportedCredentialFields[key]; !found {
			invalidCredentialFields = append(invalidCredentialFields, key)
		}
	}
	if len(invalidCredentialFields) > 0 {
		errorDesc := fmt.Sprintf(
			"Invalid credential fields in request: %s",
			strings.Join(invalidCredentialFields, ", "),
		)
		return serviceerror.CustomServiceError(ErrorInvalidCredential, errorDesc)
	}

	user := &User{
		ID:         userID,
		Attributes: attributes,
	}

	credentials, err := extractCredentials(user)
	if err != nil {
		return logErrorAndReturnServerError(
			logger,
			"Failed to extract credentials from request",
			err,
			log.String("userID", userID),
		)
	}

	if len(credentials) == 0 {
		return &ErrorMissingCredentials
	}

	_, existingCredentials, err := us.userStore.GetCredentials(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(
			logger,
			"Failed to retrieve existing user credentials",
			err,
			log.String("id", userID),
		)
	}

	mergedCredentials := us.mergeCredentials(existingCredentials, credentials)

	err = us.userStore.UpdateUserCredentials(userID, mergedCredentials)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to update user credentials", err, log.String("id", userID))
	}

	logger.Debug("Successfully updated user credentials", log.String("id", userID))
	return nil
}

// mergeCredentials combines existing credentials with provided credentials, preferring provided values.
func (us *userService) mergeCredentials(existing []Credential, provided []Credential) []Credential {
	providedMap := make(map[string]Credential, len(provided))
	providedOrder := make([]string, 0, len(provided))
	for _, cred := range provided {
		if _, exists := providedMap[cred.CredentialType]; !exists {
			providedOrder = append(providedOrder, cred.CredentialType)
		}
		providedMap[cred.CredentialType] = cred
	}

	merged := make([]Credential, 0, len(existing)+len(providedMap))

	for _, storedCred := range existing {
		if updatedCred, found := providedMap[storedCred.CredentialType]; found {
			merged = append(merged, updatedCred)
			delete(providedMap, storedCred.CredentialType)
		} else {
			merged = append(merged, storedCred)
		}
	}

	for _, credType := range providedOrder {
		if cred, found := providedMap[credType]; found {
			merged = append(merged, cred)
		}
	}

	return merged
}

// DeleteUser delete the user for given user id.
func (us *userService) DeleteUser(userID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Deleting user", log.String("id", userID))

	if userID == "" {
		return &ErrorMissingUserID
	}

	err := us.userStore.DeleteUser(userID)
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
func (us *userService) IdentifyUser(filters map[string]interface{}) (*string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(filters) == 0 {
		return nil, &ErrorInvalidRequestFormat
	}

	userID, err := us.userStore.IdentifyUser(filters)
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
	userID string, credentials map[string]interface{},
) (*User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if userID == "" {
		return nil, &ErrorMissingUserID
	}

	if len(credentials) == 0 {
		return nil, &ErrorInvalidRequestFormat
	}

	credentialsToVerify := make(map[string]string)

	for credType, credValueInterface := range credentials {
		if _, isSupported := supportedCredentialFields[credType]; !isSupported {
			continue
		}

		credValue, ok := credValueInterface.(string)
		if !ok || credValue == "" {
			continue
		}

		credentialsToVerify[credType] = credValue
	}

	if len(credentialsToVerify) == 0 {
		logger.Debug("No valid credentials provided for verification", log.String("userID", userID))
		return nil, &ErrorAuthenticationFailed
	}

	user, storedCredentials, err := us.userStore.GetCredentials(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to verify user", err, log.String("id", userID))
	}

	if len(storedCredentials) == 0 {
		logger.Debug("No credentials found for user", log.String("userID", userID))
		return nil, &ErrorAuthenticationFailed
	}

	for credType, credValue := range credentialsToVerify {
		var matchingCredential *Credential
		for _, storedCred := range storedCredentials {
			if storedCred.CredentialType == credType {
				matchingCredential = &storedCred
				break
			}
		}

		if matchingCredential == nil {
			logger.Debug("No stored credential found for type",
				log.String("userID", userID), log.String("credType", credType))
			return nil, &ErrorAuthenticationFailed
		}

		verifyingCredential := hash.Credential{
			Algorithm: matchingCredential.StorageAlgo,
			Hash:      matchingCredential.Value,
			Salt:      matchingCredential.Salt,
		}
		hashVerified := hash.Verify([]byte(credValue), verifyingCredential)

		if hashVerified {
			logger.Debug("Credential verified successfully",
				log.String("userID", userID), log.String("credType", credType))
		} else {
			logger.Debug("Credential verification failed",
				log.String("userID", userID), log.String("credType", credType))
			return nil, &ErrorAuthenticationFailed
		}
	}

	logger.Debug("Successfully verified all user credentials", log.String("id", userID))
	return &user, nil
}

// AuthenticateUser authenticates a user by combining identify and verify operations.
func (us *userService) AuthenticateUser(
	request AuthenticateUserRequest,
) (*AuthenticateUserResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(request) == 0 {
		return nil, &ErrorInvalidRequestFormat
	}

	identifyFilters := make(map[string]interface{})
	credentials := make(map[string]interface{})

	for key, value := range request {
		if _, isCredential := supportedCredentialFields[key]; isCredential {
			credentials[key] = value
		} else {
			identifyFilters[key] = value
		}
	}

	if len(identifyFilters) == 0 {
		return nil, &ErrorMissingRequiredFields
	}
	if len(credentials) == 0 {
		return nil, &ErrorMissingCredentials
	}

	userID, svcErr := us.IdentifyUser(identifyFilters)
	if svcErr != nil {
		if svcErr.Code == ErrorUserNotFound.Code {
			return nil, &ErrorUserNotFound
		}
		return nil, svcErr
	}

	user, svcErr := us.VerifyUser(*userID, credentials)
	if svcErr != nil {
		return nil, svcErr
	}

	logger.Debug("User authenticated successfully", log.String("userID", *userID))
	return &AuthenticateUserResponse{
		ID:               user.ID,
		Type:             user.Type,
		OrganizationUnit: user.OrganizationUnit,
	}, nil
}

// ValidateUserIDs validates that all provided user IDs exist.
func (us *userService) ValidateUserIDs(userIDs []string) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(userIDs) == 0 {
		return []string{}, nil
	}

	invalidUserIDs, err := us.userStore.ValidateUserIDs(userIDs)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to validate user IDs", err)
	}

	return invalidUserIDs, nil
}

// validateOrganizationUnitForUserType ensures that the organization unit ID is valid and belongs to the user type.
func (us *userService) validateOrganizationUnitForUserType(
	userType, organizationUnitID string, logger *log.Logger,
) *serviceerror.ServiceError {
	if strings.TrimSpace(userType) == "" {
		return &ErrorUserSchemaNotFound
	}

	if strings.TrimSpace(organizationUnitID) == "" || !utils.IsValidUUID(organizationUnitID) {
		return &ErrorInvalidOrganizationUnitID
	}

	if us.ouService == nil {
		logger.Error("Organization unit service is not configured for user operations")
		return &ErrorInternalServerError
	}

	exists, svcErr := us.ouService.IsOrganizationUnitExists(organizationUnitID)
	if svcErr != nil {
		return mapOUServiceError(
			svcErr,
			logger,
			"verifying organization unit existence",
			map[string]*serviceerror.ServiceError{
				oupkg.ErrorOrganizationUnitNotFound.Code:  &ErrorOrganizationUnitNotFound,
				oupkg.ErrorInvalidRequestFormat.Code:      &ErrorInvalidOrganizationUnitID,
				oupkg.ErrorMissingOrganizationUnitID.Code: &ErrorInvalidOrganizationUnitID,
			},
			log.String("organizationUnitID", organizationUnitID),
		)
	}
	if !exists {
		return &ErrorOrganizationUnitNotFound
	}

	if us.userSchemaService == nil {
		logger.Error("User schema service is not configured for user operations")
		return &ErrorInternalServerError
	}

	userSchema, svcErr := us.userSchemaService.GetUserSchemaByName(userType)
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

	if userSchema.OrganizationUnitID == organizationUnitID {
		return nil
	}

	isParent, svcErr := us.ouService.IsParent(userSchema.OrganizationUnitID, organizationUnitID)
	if svcErr != nil {
		return mapOUServiceError(
			svcErr,
			logger,
			"validating organization unit hierarchy",
			map[string]*serviceerror.ServiceError{
				oupkg.ErrorOrganizationUnitNotFound.Code: &ErrorOrganizationUnitNotFound,
			},
			log.String("userType", userType),
			log.String("organizationUnitID", organizationUnitID),
			log.String("schemaOrganizationUnitID", userSchema.OrganizationUnitID),
		)
	}

	if !isParent {
		logger.Debug("Organization unit mismatch for user type",
			log.String("userType", userType),
			log.String("organizationUnitID", organizationUnitID),
			log.String("schemaOrganizationUnitID", userSchema.OrganizationUnitID))
		return &ErrorOrganizationUnitMismatch
	}

	return nil
}

// validateUserAndUniqueness validates the user schema and checks for uniqueness.
func (us *userService) validateUserAndUniqueness(
	userType string, attributes []byte, logger *log.Logger,
) *serviceerror.ServiceError {
	isValid, svcErr := us.userSchemaService.ValidateUser(userType, attributes)
	if svcErr != nil {
		if svcErr.Code == userschema.ErrorUserSchemaNotFound.Code {
			return &ErrorUserSchemaNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to validate user schema", nil)
	}
	if !isValid {
		return &ErrorSchemaValidationFailed
	}

	isValid, svcErr = us.userSchemaService.ValidateUserUniqueness(userType, attributes,
		func(filters map[string]interface{}) (*string, error) {
			userID, svcErr := us.IdentifyUser(filters)
			if svcErr != nil {
				if svcErr.Code == ErrorUserNotFound.Code {
					return nil, nil
				} else {
					return nil, errors.New(svcErr.Error)
				}
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

// buildPaginationLinks builds pagination links for the response.
func buildPaginationLinks(path string, limit, offset, totalResults int) []Link {
	links := make([]Link, 0)

	if offset > 0 {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=0&limit=%d", path, limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", path, prevOffset, limit),
			Rel:  "prev",
		})
	}

	if offset+limit < totalResults {
		nextOffset := offset + limit
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", path, nextOffset, limit),
			Rel:  "next",
		})
	}

	lastPageOffset := ((totalResults - 1) / limit) * limit
	if offset < lastPageOffset {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", path, lastPageOffset, limit),
			Rel:  "last",
		})
	}

	return links
}

// buildTreePaginationLinks builds pagination links for user responses.
func buildTreePaginationLinks(handlePath string, limit, offset, totalResults int) []Link {
	path := fmt.Sprintf("/users/tree/%s", path.Clean(handlePath))
	return buildPaginationLinks(path, limit, offset, totalResults)
}
