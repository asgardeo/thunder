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

// Package usertype handles the user type management operations.
package usertype

import (
	"encoding/json"
	"errors"
	"fmt"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/usertype/model"
)

const userTypeLoggerComponentName = "UserTypeService"

// UserTypeServiceInterface defines the interface for the user type service.
type UserTypeServiceInterface interface {
	GetUserTypeList(limit, offset int) (*UserTypeListResponse, *serviceerror.ServiceError)
	CreateUserType(request CreateUserTypeRequest) (*UserType, *serviceerror.ServiceError)
	GetUserType(schemaID string) (*UserType, *serviceerror.ServiceError)
	GetUserTypeByName(schemaName string) (*UserType, *serviceerror.ServiceError)
	UpdateUserType(schemaID string, request UpdateUserTypeRequest) (
		*UserType, *serviceerror.ServiceError)
	DeleteUserType(schemaID string) *serviceerror.ServiceError
	ValidateUser(userType string, userAttributes json.RawMessage) (bool, *serviceerror.ServiceError)
	ValidateUserUniqueness(userType string, userAttributes json.RawMessage,
		identifyUser func(map[string]interface{}) (*string, error)) (bool, *serviceerror.ServiceError)
}

// userTypeService is the default implementation of the UserTypeServiceInterface.
type userTypeService struct {
	userTypeStore userTypeStoreInterface
	ouService     oupkg.OrganizationUnitServiceInterface
}

// newUserTypeService creates a new instance of userTypeService.
func newUserTypeService(ouService oupkg.OrganizationUnitServiceInterface,
	store userTypeStoreInterface) UserTypeServiceInterface {
	return &userTypeService{
		userTypeStore: store,
		ouService:     ouService,
	}
}

// GetUserTypeList lists the user types with pagination.
func (us *userTypeService) GetUserTypeList(limit, offset int) (
	*UserTypeListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeLoggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := us.userTypeStore.GetUserTypeListCount()
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get user type list count", err)
	}

	userTypes, err := us.userTypeStore.GetUserTypeList(limit, offset)
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get user type list", err)
	}

	response := &UserTypeListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(userTypes),
		Types:        userTypes,
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}

	return response, nil
}

// CreateUserType creates a new user type.
func (us *userTypeService) CreateUserType(request CreateUserTypeRequest) (
	*UserType, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeLoggerComponentName))

	if err := declarativeresource.CheckDeclarativeCreate(); err != nil {
		return nil, err
	}

	// Validate the schema definition
	schemaToValidate := UserType{
		Name:               request.Name,
		OrganizationUnitID: request.OrganizationUnitID,
		Schema:             request.Schema,
	}
	if validationErr := validateUserTypeDefinition(schemaToValidate); validationErr != nil {
		logger.Debug("User type validation failed", log.String("name", request.Name))
		return nil, validationErr
	}

	// Ensure organization unit exists
	if svcErr := us.ensureOrganizationUnitExists(request.OrganizationUnitID, logger); svcErr != nil {
		return nil, svcErr
	}

	// Check for name conflicts
	_, err := us.userTypeStore.GetUserTypeByName(request.Name)
	if err == nil {
		return nil, &ErrorUserTypeNameConflict
	} else if !errors.Is(err, ErrUserTypeNotFound) {
		return nil, logAndReturnServerError(logger, "Failed to check existing user type", err)
	}

	id, err := utils.GenerateUUIDv7()
	if err != nil {
		logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	userType := UserType{
		ID:                    id,
		Name:                  request.Name,
		OrganizationUnitID:    request.OrganizationUnitID,
		AllowSelfRegistration: request.AllowSelfRegistration,
		Schema:                request.Schema,
	}

	if err := us.userTypeStore.CreateUserType(userType); err != nil {
		return nil, logAndReturnServerError(logger, "Failed to create user type", err)
	}

	return &userType, nil
}

// GetUserType retrieves a user type by its ID.
func (us *userTypeService) GetUserType(schemaID string) (*UserType, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeLoggerComponentName))

	if schemaID == "" {
		return nil, invalidSchemaRequestError("schema id must not be empty")
	}

	userType, err := us.userTypeStore.GetUserTypeByID(schemaID)
	if err != nil {
		if errors.Is(err, ErrUserTypeNotFound) {
			return nil, &ErrorUserTypeNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to get user type", err)
	}

	return &userType, nil
}

// GetUserTypeByName retrieves a user type by its name.
func (us *userTypeService) GetUserTypeByName(schemaName string) (*UserType, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeLoggerComponentName))

	if schemaName == "" {
		return nil, invalidSchemaRequestError("schema name must not be empty")
	}

	userType, err := us.userTypeStore.GetUserTypeByName(schemaName)
	if err != nil {
		if errors.Is(err, ErrUserTypeNotFound) {
			return nil, &ErrorUserTypeNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to get user type by name", err)
	}

	return &userType, nil
}

// UpdateUserType updates a user type by its ID.
func (us *userTypeService) UpdateUserType(schemaID string, request UpdateUserTypeRequest) (
	*UserType, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeLoggerComponentName))

	if err := declarativeresource.CheckDeclarativeUpdate(); err != nil {
		return nil, err
	}

	if schemaID == "" {
		return nil, invalidSchemaRequestError("schema id must not be empty")
	}

	// Validate the schema definition
	schemaToValidate := UserType{
		Name:               request.Name,
		OrganizationUnitID: request.OrganizationUnitID,
		Schema:             request.Schema,
	}
	if validationErr := validateUserTypeDefinition(schemaToValidate); validationErr != nil {
		logger.Debug("User type validation failed", log.String("id", schemaID))
		return nil, validationErr
	}

	// Ensure organization unit exists
	if svcErr := us.ensureOrganizationUnitExists(request.OrganizationUnitID, logger); svcErr != nil {
		return nil, svcErr
	}

	existingUserType, err := us.userTypeStore.GetUserTypeByID(schemaID)
	if err != nil {
		if errors.Is(err, ErrUserTypeNotFound) {
			return nil, &ErrorUserTypeNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to get existing user type", err)
	}

	if request.Name != existingUserType.Name {
		_, err := us.userTypeStore.GetUserTypeByName(request.Name)
		if err == nil {
			return nil, &ErrorUserTypeNameConflict
		} else if !errors.Is(err, ErrUserTypeNotFound) {
			return nil, logAndReturnServerError(logger, "Failed to check existing user type", err)
		}
	}

	userType := UserType{
		ID:                    schemaID,
		Name:                  request.Name,
		OrganizationUnitID:    request.OrganizationUnitID,
		AllowSelfRegistration: request.AllowSelfRegistration,
		Schema:                request.Schema,
	}

	if err := us.userTypeStore.UpdateUserTypeByID(schemaID, userType); err != nil {
		return nil, logAndReturnServerError(logger, "Failed to update user type", err)
	}

	return &userType, nil
}

// DeleteUserType deletes a user type by its ID.
func (us *userTypeService) DeleteUserType(schemaID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeLoggerComponentName))

	if err := declarativeresource.CheckDeclarativeDelete(); err != nil {
		return err
	}

	if schemaID == "" {
		return invalidSchemaRequestError("schema id must not be empty")
	}

	if err := us.userTypeStore.DeleteUserTypeByID(schemaID); err != nil {
		return logAndReturnServerError(logger, "Failed to delete user type", err)
	}

	return nil
}

// ValidateUser validates user attributes against the user type for the given user type.
func (us *userTypeService) ValidateUser(
	userType string, userAttributes json.RawMessage,
) (bool, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeLoggerComponentName))

	compiledSchema, err := us.getCompiledSchemaForUserType(userType, logger)
	if err != nil {
		if errors.Is(err, ErrUserTypeNotFound) {
			return false, &ErrorUserTypeNotFound
		}
		return false, logAndReturnServerError(logger, "Failed to load user type", err)
	}

	isValid, err := compiledSchema.Validate(userAttributes, logger)
	if err != nil {
		return false, logAndReturnServerError(logger, "Failed to validate user attributes against schema", err)
	}
	if !isValid {
		logger.Debug("Schema validation failed", log.String("userType", userType))
		return false, nil
	}

	logger.Debug("Schema validation successful", log.String("userType", userType))
	return true, nil
}

// ValidateUserUniqueness validates the uniqueness constraints of user attributes.
func (us *userTypeService) ValidateUserUniqueness(
	userType string,
	userAttributes json.RawMessage,
	identifyUser func(map[string]interface{}) (*string, error),
) (bool, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeLoggerComponentName))

	compiledSchema, err := us.getCompiledSchemaForUserType(userType, logger)
	if err != nil {
		if errors.Is(err, ErrUserTypeNotFound) {
			return false, &ErrorUserTypeNotFound
		}
		return false, logAndReturnServerError(logger, "Failed to load user type", err)
	}

	if len(userAttributes) == 0 {
		return true, nil
	}

	var userAttrs map[string]interface{}
	if err := json.Unmarshal(userAttributes, &userAttrs); err != nil {
		return false, logAndReturnServerError(logger, "Failed to unmarshal user attributes", err)
	}

	isValid, err := compiledSchema.ValidateUniqueness(userAttrs, identifyUser, logger)
	if err != nil {
		return false, logAndReturnServerError(logger, "Failed during uniqueness validation", err)
	}
	if !isValid {
		logger.Debug("User attribute failed uniqueness validation", log.String("userType", userType))
		return false, nil
	}

	return true, nil
}

func (us *userTypeService) getCompiledSchemaForUserType(
	userType string,
	logger *log.Logger,
) (*model.Schema, error) {
	if userType == "" {
		return nil, ErrUserTypeNotFound
	}

	userTypeData, err := us.userTypeStore.GetUserTypeByName(userType)
	if err != nil {
		return nil, err
	}

	compiled, err := model.CompileUserType(userTypeData.Schema)
	if err != nil {
		logger.Error("Failed to compile stored user type", log.String("userType", userType), log.Error(err))
		return nil, fmt.Errorf("failed to compile stored user type: %w", err)
	}

	return compiled, nil
}

// ensureOrganizationUnitExists validates that the provided organization unit exists using the OU service.
func (us *userTypeService) ensureOrganizationUnitExists(
	organizationUnitID string,
	logger *log.Logger,
) *serviceerror.ServiceError {
	if us.ouService == nil {
		logger.Error("Organization unit service is not configured for user type operations")
		return &ErrorInternalServerError
	}

	exists, svcErr := us.ouService.IsOrganizationUnitExists(organizationUnitID)
	if svcErr != nil {
		logger.Error("Failed to verify organization unit existence",
			log.String("organizationUnitID", organizationUnitID), log.Any("error", svcErr))
		return &ErrorInternalServerError
	}

	if !exists {
		logger.Debug("Organization unit does not exist",
			log.String("organizationUnitID", organizationUnitID))
		return invalidSchemaRequestError("organization unit id does not exist")
	}

	return nil
}

// validatePaginationParams validates the limit and offset parameters.
func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 1 || limit > serverconst.MaxPageSize {
		return &ErrorInvalidLimit
	}
	if offset < 0 {
		return &ErrorInvalidOffset
	}
	return nil
}

// buildPaginationLinks builds pagination links for the response.
func buildPaginationLinks(limit, offset, totalCount int) []Link {
	links := make([]Link, 0)

	if offset > 0 {
		links = append(links, Link{
			Href: fmt.Sprintf("/user-types?offset=0&limit=%d", limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, Link{
			Href: fmt.Sprintf("/user-types?offset=%d&limit=%d", prevOffset, limit),
			Rel:  "prev",
		})
	}

	if offset+limit < totalCount {
		nextOffset := offset + limit
		links = append(links, Link{
			Href: fmt.Sprintf("/user-types?offset=%d&limit=%d", nextOffset, limit),
			Rel:  "next",
		})
	}

	lastPageOffset := ((totalCount - 1) / limit) * limit
	if offset < lastPageOffset {
		links = append(links, Link{
			Href: fmt.Sprintf("/user-types?offset=%d&limit=%d", lastPageOffset, limit),
			Rel:  "last",
		})
	}

	return links
}

// logAndReturnServerError logs the error and returns a server error.
func logAndReturnServerError(
	logger *log.Logger,
	message string,
	err error,
) *serviceerror.ServiceError {
	logger.Error(message, log.Error(err))
	return &ErrorInternalServerError
}

// validateUserTypeDefinition validates the user type definition without checking OU existence.
// This is used during initialization to validate file-based configurations.
func validateUserTypeDefinition(schema UserType) *serviceerror.ServiceError {
	logger := log.GetLogger()

	if schema.Name == "" {
		logger.Debug("User type validation failed: name is empty")
		return invalidSchemaRequestError("user type name must not be empty")
	}

	if schema.OrganizationUnitID == "" {
		logger.Debug("User type validation failed: organization unit ID is empty")
		return invalidSchemaRequestError("organization unit id must not be empty")
	}

	if !utils.IsValidUUID(schema.OrganizationUnitID) {
		logger.Debug("User type validation failed: invalid organization unit ID format",
			log.String("ouId", schema.OrganizationUnitID))
		return invalidSchemaRequestError("organization unit id is not a valid UUID")
	}

	if len(schema.Schema) == 0 {
		logger.Debug("User type validation failed: schema definition is empty")
		return invalidSchemaRequestError("schema definition must not be empty")
	}

	_, err := model.CompileUserType(schema.Schema)
	if err != nil {
		logger.Debug("User type validation failed: schema compilation error",
			log.Error(err))
		return invalidSchemaRequestError(err.Error())
	}

	return nil
}

func invalidSchemaRequestError(detail string) *serviceerror.ServiceError {
	err := ErrorInvalidUserTypeRequest
	errorDescription := err.ErrorDescription
	if detail != "" {
		errorDescription = fmt.Sprintf("%s: %s", err.ErrorDescription, detail)
	}
	return &serviceerror.ServiceError{
		Code:             err.Code,
		Type:             err.Type,
		Error:            err.Error,
		ErrorDescription: errorDescription,
	}
}
