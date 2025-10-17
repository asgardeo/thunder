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
	UpdateUser(userID string, user *User) (*User, *serviceerror.ServiceError)
	DeleteUser(userID string) *serviceerror.ServiceError
	IdentifyUser(filters map[string]interface{}, userType string) (*string, *serviceerror.ServiceError)
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
	ouService oupkg.OrganizationUnitServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
) UserServiceInterface {
	return &userService{
		userStore:         newUserStore(),
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
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &ErrorOrganizationUnitNotFound
		}
		return nil, logErrorAndReturnServerError(logger,
			"Failed to get organization unit using the handle path from organization service", nil)
	}
	organizationUnitID := ou.ID

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	ouResponse, svcErr := us.ouService.GetOrganizationUnitUsers(organizationUnitID, limit, offset)
	if svcErr != nil {
		return nil, svcErr
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

	if svcErr := us.validateUserAndUniqueness(user.Type, user.Attributes, logger); svcErr != nil {
		return nil, svcErr
	}

	user.ID = utils.GenerateUUID()

	credentials, err := extractCredentials(user)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to create user DTO", err)
	}

	indexedPropertyToColumnNumberMap, svcErr := us.userSchemaService.GetIndexedPropertyToColumnNumberMap(user.Type)
	if svcErr != nil {
		return nil, svcErr
	}

	var parsedAttributes map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &parsedAttributes); err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to parse user attributes", err)
	}

	indexedColumnNameToValueMap := make(map[string]string)

	for attributeName, columnNumber := range indexedPropertyToColumnNumberMap {
		columnName := fmt.Sprintf("indexed_prop_%d_value", columnNumber)
		if value, exists := parsedAttributes[attributeName]; exists {
			if strValue, ok := value.(string); ok {
				indexedColumnNameToValueMap[columnName] = strValue
			} else {
				return nil, logErrorAndReturnServerError(logger,
					fmt.Sprintf("Indexed property %s is not a string", attributeName), err)
			}
		} else {
			// If the indexed property is not present, set it to an empty string
			indexedColumnNameToValueMap[columnName] = ""
		}
	}

	for i := 1; i <= 5; i++ {
		columnName := fmt.Sprintf("indexed_prop_%d_value", i)
		if _, exists := indexedColumnNameToValueMap[columnName]; !exists {
			indexedColumnNameToValueMap[columnName] = ""
		}
	}

	err = us.userStore.CreateUser(*user, credentials, indexedColumnNameToValueMap)
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
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &ErrorOrganizationUnitNotFound
		}
		return nil, logErrorAndReturnServerError(logger,
			"Failed to get organization unit using the handle path from organization service", nil)
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
			credHash := hash.NewCredential([]byte(credValue))

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
func (us *userService) IdentifyUser(filters map[string]interface{}, userType string) (*string,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(filters) == 0 {
		return nil, &ErrorInvalidRequestFormat
	}

	indexedPropToColumnNumberMap, schemaSvcError := us.userSchemaService.GetIndexedPropertyToColumnNumberMap(userType)
	if schemaSvcError != nil {
		return nil, schemaSvcError
	}

	unindexedFilters := make(map[string]interface{})
	indexedFilters := make(map[string]interface{})

	for attributeName, attributeValue := range filters {
		if columnNumber, isIndexed := indexedPropToColumnNumberMap[attributeName]; isIndexed {
			columnName := fmt.Sprintf("indexed_prop_%d_value", columnNumber)
			indexedFilters[columnName] = attributeValue
		} else {
			unindexedFilters[attributeName] = attributeValue
		}
	}

	userID, err := us.userStore.IdentifyUser(unindexedFilters, indexedFilters)
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

	user, storedCredentials, err := us.userStore.VerifyUser(userID)
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
			logger.Debug("No stored credential found for type", log.String("userID", userID), log.String("credType", credType))
			return nil, &ErrorAuthenticationFailed
		}

		verifyingCredential := hash.Credential{
			Algorithm: matchingCredential.StorageAlgo,
			Hash:      matchingCredential.Value,
			Salt:      matchingCredential.Salt,
		}
		hashVerified := hash.Verify([]byte(credValue), verifyingCredential)

		if hashVerified {
			logger.Debug("Credential verified successfully", log.String("userID", userID), log.String("credType", credType))
		} else {
			logger.Debug("Credential verification failed", log.String("userID", userID), log.String("credType", credType))
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

	if len(request.Attributes) == 0 {
		return nil, &ErrorInvalidRequestFormat
	}

	identifyFilters := make(map[string]interface{})
	credentials := make(map[string]interface{})

	for key, value := range request.Attributes {
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

	userID, svcErr := us.IdentifyUser(identifyFilters, request.UserType)
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

// validateUserAndUniqueness validates the user schema and checks for uniqueness.
func (us *userService) validateUserAndUniqueness(
	userType string, attributes []byte, logger *log.Logger,
) *serviceerror.ServiceError {
	isValid, svcErr := us.userSchemaService.ValidateUser(userType, attributes)
	if svcErr != nil {
		return logErrorAndReturnServerError(logger, "Failed to validate user schema", nil)
	}
	if !isValid {
		return &ErrorSchemaValidationFailed
	}

	isValid, svcErr = us.userSchemaService.ValidateUserUniqueness(userType, attributes,
		func(filters map[string]interface{}) (*string, error) {
			userID, svcErr := us.IdentifyUser(filters, userType)
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
