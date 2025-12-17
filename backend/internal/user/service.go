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

// Credential type constants
const (
	CredentialTypePassword = "password"
	CredentialTypePin      = "pin"
	CredentialTypeSecret   = "secret"
	CredentialTypePasskey  = "passkey"
)

// SupportedCredentialFields defines the set of credential field names that are supported.
var supportedCredentialFields = map[string]struct{}{
	CredentialTypePassword: {},
	CredentialTypePin:      {},
	CredentialTypeSecret:   {},
	CredentialTypePasskey:  {},
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
	AddPasskeyCredential(userID string, credential Credential) *serviceerror.ServiceError
	GetUserCredentials(userID string, credentialType string) ([]Credential, *serviceerror.ServiceError)
	GetUserWithCredentials(userID string) (*User, CredentialMap, *serviceerror.ServiceError)
	UpdateCredentialMap(userID string, credentialMap CredentialMap) *serviceerror.ServiceError
	UpdateCredentialInMap(userID string, credentialType string, updatedCredentials []Credential) *serviceerror.ServiceError
}

// userService is the default implementation of the UserServiceInterface.
type userService struct {
	userStore         userStoreInterface
	ouService         oupkg.OrganizationUnitServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
	hashService       hash.HashServiceInterface
}

// newUserService creates a new instance of userService with injected dependencies.
func newUserService(
	userStore userStoreInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	hashService hash.HashServiceInterface,
) UserServiceInterface {
	return &userService{
		userStore:         userStore,
		ouService:         ouService,
		userSchemaService: userSchemaService,
		hashService:       hashService,
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

	credentials, err := us.extractCredentials(user)
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

// extractCredentials extracts the credentials from the user attributes and returns a CredentialMap.
func (us *userService) extractCredentials(user *User) (CredentialMap, error) {
	if user.Attributes == nil {
		return CredentialMap{}, nil
	}

	var attrsMap map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrsMap); err != nil {
		return nil, err
	}

	credentials := make(CredentialMap)

	for credField := range supportedCredentialFields {
		if credValue, ok := attrsMap[credField].(string); ok {
			credHash, err := us.hashService.Generate([]byte(credValue))
			if err != nil {
				return nil, err
			}

			delete(attrsMap, credField)

			credential := Credential{
				CredentialType: credField,
				StorageType:    "hash",
				StorageAlgo:    credHash.Algorithm,
				StorageAlgoParams: hash.CredParameters{
					Iterations: credHash.Parameters.Iterations,
					KeySize:    credHash.Parameters.KeySize,
					Salt:       credHash.Parameters.Salt,
				},
				Value: credHash.Hash,
			}

			// Initialize the credential type array if it doesn't exist
			if credentials[credField] == nil {
				credentials[credField] = []Credential{}
			}
			credentials[credField] = append(credentials[credField], credential)
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

	credentials, err := us.extractCredentials(user)
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
func (us *userService) mergeCredentials(existing CredentialMap, provided CredentialMap) CredentialMap {
	merged := make(CredentialMap)

	// Copy all existing credentials
	for credType, credList := range existing {
		merged[credType] = append([]Credential{}, credList...)
	}

	// Override or add provided credentials
	for credType, credList := range provided {
		merged[credType] = credList
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
		// Get credentials of this type from the map
		credList, exists := storedCredentials[credType]
		if !exists || len(credList) == 0 {
			logger.Debug("No stored credential found for type",
				log.String("userID", userID), log.String("credType", credType))
			return nil, &ErrorAuthenticationFailed
		}

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
			hashVerified := us.hashService.Verify([]byte(credValue), verifyingCredential)

			if hashVerified {
				logger.Debug("Credential verified successfully",
					log.String("userID", userID), log.String("credType", credType))
				verified = true
				break
			}
		}

		if !verified {
		  hashVerified, err := us.hashService.Verify([]byte(credValue), verifyingCredential)

      if err == nil && hashVerified {
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

// AddPasskeyCredential adds a passkey credential to a user's credential map.
func (us *userService) AddPasskeyCredential(userID string, credential Credential) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Adding passkey credential", log.String("userID", log.MaskString(userID)))

	if strings.TrimSpace(userID) == "" {
		return &ErrorMissingUserID
	}

	if credential.CredentialType != CredentialTypePasskey {
		logger.Debug("Invalid credential type for passkey",
			log.String("providedType", credential.CredentialType))
		return &ErrorInvalidCredential
	}

	if strings.TrimSpace(credential.Value) == "" {
		logger.Debug("Empty credential value")
		return &ErrorInvalidCredential
	}

	// Get existing credentials
	_, existingCredentials, err := us.userStore.GetCredentials(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("userID", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(
			logger,
			"Failed to retrieve existing user credentials",
			err,
			log.String("userID", userID),
		)
	}

	// Add the passkey credential to the map
	if existingCredentials == nil {
		existingCredentials = make(CredentialMap)
	}

	// Initialize passkey credentials array if it doesn't exist
	if existingCredentials[CredentialTypePasskey] == nil {
		existingCredentials[CredentialTypePasskey] = []Credential{}
	}

	// Append the new passkey credential
	existingCredentials[CredentialTypePasskey] = append(
		existingCredentials[CredentialTypePasskey],
		credential,
	)

	// Update credentials in the database
	err = us.userStore.UpdateUserCredentials(userID, existingCredentials)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("userID", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(
			logger,
			"Failed to update user credentials",
			err,
			log.String("userID", userID),
		)
	}

	logger.Debug("Successfully added passkey credential", log.String("userID", log.MaskString(userID)))
	return nil
}

// GetUserCredentials retrieves credentials of a specific type for a user.
// If credentialType is empty, returns all credentials.
func (us *userService) GetUserCredentials(userID string, credentialType string) ([]Credential, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving user credentials",
		log.String("userID", log.MaskString(userID)),
		log.String("credentialType", credentialType))

	if strings.TrimSpace(userID) == "" {
		return nil, &ErrorMissingUserID
	}

	// Get all credentials for the user
	_, credentialMap, err := us.userStore.GetCredentials(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
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

	// If no credential type specified, return all credentials
	if strings.TrimSpace(credentialType) == "" {
		allCredentials := []Credential{}
		for _, credList := range credentialMap {
			allCredentials = append(allCredentials, credList...)
		}
		logger.Debug("Retrieved all credentials", log.String("userID", log.MaskString(userID)))
		return allCredentials, nil
	}

	// Return credentials of the specified type
	credentials, exists := credentialMap[credentialType]
	if !exists || len(credentials) == 0 {
		logger.Debug("No credentials found for type",
			log.String("userID", log.MaskString(userID)),
			log.String("credentialType", credentialType))
		return []Credential{}, nil
	}

	logger.Debug("Retrieved credentials for type",
		log.String("userID", log.MaskString(userID)),
		log.String("credentialType", credentialType),
		log.Int("count", len(credentials)))
	return credentials, nil
}

// GetUserWithCredentials retrieves a user along with their complete credential map.
// This method is useful when you need both user information and credentials in a single call.
func (us *userService) GetUserWithCredentials(userID string) (*User, CredentialMap, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving user with credentials", log.String("userID", log.MaskString(userID)))

	if strings.TrimSpace(userID) == "" {
		return nil, nil, &ErrorMissingUserID
	}

	// Get user and credentials from store
	user, credentialMap, err := us.userStore.GetCredentials(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("userID", userID))
			return nil, nil, &ErrorUserNotFound
		}
		return nil, nil, logErrorAndReturnServerError(
			logger,
			"Failed to retrieve user with credentials",
			err,
			log.String("userID", userID),
		)
	}

	logger.Debug("Retrieved user with credentials",
		log.String("userID", log.MaskString(userID)),
		log.Int("credentialTypeCount", len(credentialMap)))
	return &user, credentialMap, nil
}

// UpdateCredentialMap updates the entire credential map for a user.
// This method replaces all existing credentials with the provided credential map.
// Use with caution as it will overwrite all existing credentials.
func (us *userService) UpdateCredentialMap(userID string, credentialMap CredentialMap) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating credential map",
		log.String("userID", log.MaskString(userID)),
		log.Int("credentialTypeCount", len(credentialMap)))

	if strings.TrimSpace(userID) == "" {
		return &ErrorMissingUserID
	}

	if credentialMap == nil {
		logger.Debug("Credential map is nil")
		return &ErrorInvalidRequestFormat
	}

	// Validate credential map structure
	for credType, credentials := range credentialMap {
		if credType == "" {
			logger.Debug("Empty credential type in map")
			return &ErrorInvalidRequestFormat
		}
		for _, cred := range credentials {
			if cred.CredentialType != credType {
				logger.Debug("Credential type mismatch in map",
					log.String("mapKey", credType),
					log.String("credentialType", cred.CredentialType))
				return &ErrorInvalidRequestFormat
			}
		}
	}

	// Update credentials in store
	err := us.userStore.UpdateUserCredentials(userID, credentialMap)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("userID", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(
			logger,
			"Failed to update credential map",
			err,
			log.String("userID", userID),
		)
	}

	logger.Debug("Successfully updated credential map", log.String("userID", log.MaskString(userID)))
	return nil
}

// UpdateCredentialInMap updates credentials of a specific type for a user.
// This method retrieves the existing credential map, updates the credentials for the specified type,
// and saves the updated map back to the store.
func (us *userService) UpdateCredentialInMap(userID string, credentialType string, updatedCredentials []Credential) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating credentials in map",
		log.String("userID", log.MaskString(userID)),
		log.String("credentialType", credentialType),
		log.Int("credentialCount", len(updatedCredentials)))

	if strings.TrimSpace(userID) == "" {
		return &ErrorMissingUserID
	}

	if strings.TrimSpace(credentialType) == "" {
		logger.Debug("Credential type is empty")
		return &ErrorInvalidRequestFormat
	}

	// Validate that all credentials have the correct type
	for _, cred := range updatedCredentials {
		if cred.CredentialType != credentialType {
			logger.Debug("Credential type mismatch",
				log.String("expected", credentialType),
				log.String("actual", cred.CredentialType))
			return &ErrorInvalidRequestFormat
		}
	}

	// Get existing credential map
	_, existingCredMap, err := us.userStore.GetCredentials(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("userID", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(
			logger,
			"Failed to retrieve existing credentials",
			err,
			log.String("userID", userID),
		)
	}

	// Update the credentials for the specified type
	existingCredMap[credentialType] = updatedCredentials

	// Save the updated credential map
	err = us.userStore.UpdateUserCredentials(userID, existingCredMap)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			logger.Debug("User not found", log.String("userID", userID))
			return &ErrorUserNotFound
		}
		return logErrorAndReturnServerError(
			logger,
			"Failed to update credentials in map",
			err,
			log.String("userID", userID),
		)
	}

	logger.Debug("Successfully updated credentials in map",
		log.String("userID", log.MaskString(userID)),
		log.String("credentialType", credentialType))
	return nil
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
