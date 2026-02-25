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

// Package userschema handles the user schema management operations.
package userschema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema/model"
)

const userSchemaLoggerComponentName = "UserSchemaService"

// UserSchemaServiceInterface defines the interface for the user schema service.
type UserSchemaServiceInterface interface {
	GetUserSchemaList(ctx context.Context, limit, offset int) (*UserSchemaListResponse, *serviceerror.ServiceError)
	CreateUserSchema(
		ctx context.Context, request CreateUserSchemaRequest,
	) (*UserSchema, *serviceerror.ServiceError)
	GetUserSchema(ctx context.Context, schemaID string) (*UserSchema, *serviceerror.ServiceError)
	GetUserSchemaByName(
		ctx context.Context, schemaName string,
	) (*UserSchema, *serviceerror.ServiceError)
	UpdateUserSchema(ctx context.Context, schemaID string, request UpdateUserSchemaRequest) (
		*UserSchema, *serviceerror.ServiceError)
	DeleteUserSchema(ctx context.Context, schemaID string) *serviceerror.ServiceError
	ValidateUser(
		ctx context.Context, userType string, userAttributes json.RawMessage,
	) (bool, *serviceerror.ServiceError)
	ValidateUserUniqueness(
		ctx context.Context,
		userType string,
		userAttributes json.RawMessage,
		identifyUser func(map[string]interface{}) (*string, error),
	) (bool, *serviceerror.ServiceError)
	GetCredentialAttributes(
		ctx context.Context, userType string,
	) ([]string, *serviceerror.ServiceError)
}

// userSchemaService is the default implementation of the UserSchemaServiceInterface.
type userSchemaService struct {
	userSchemaStore userSchemaStoreInterface
	ouService       oupkg.OrganizationUnitServiceInterface
	transactioner   transaction.Transactioner
}

// newUserSchemaService creates a new instance of userSchemaService.
func newUserSchemaService(
	ouService oupkg.OrganizationUnitServiceInterface,
	store userSchemaStoreInterface,
	transactioner transaction.Transactioner,
) UserSchemaServiceInterface {
	return &userSchemaService{
		userSchemaStore: store,
		ouService:       ouService,
		transactioner:   transactioner,
	}
}

// GetUserSchemaList lists the user schemas with pagination.
func (us *userSchemaService) GetUserSchemaList(ctx context.Context, limit, offset int) (
	*UserSchemaListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := us.userSchemaStore.GetUserSchemaListCount(ctx)
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get user schema list count", err)
	}

	userSchemas, err := us.userSchemaStore.GetUserSchemaList(ctx, limit, offset)
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get user schema list", err)
	}

	response := &UserSchemaListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(userSchemas),
		Schemas:      userSchemas,
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}

	return response, nil
}

// CreateUserSchema creates a new user schema.
func (us *userSchemaService) CreateUserSchema(
	ctx context.Context, request CreateUserSchemaRequest,
) (*UserSchema, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if err := declarativeresource.CheckDeclarativeCreate(); err != nil {
		return nil, err
	}

	// Validate the schema definition
	schemaToValidate := UserSchema{
		Name:               request.Name,
		OrganizationUnitID: request.OrganizationUnitID,
		Schema:             request.Schema,
	}
	if validationErr := validateUserSchemaDefinition(schemaToValidate); validationErr != nil {
		logger.Debug("User schema validation failed", log.String("name", request.Name))
		return nil, validationErr
	}

	// Ensure organization unit exists
	if svcErr := us.ensureOrganizationUnitExists(request.OrganizationUnitID, logger); svcErr != nil {
		return nil, svcErr
	}

	// Check for name conflicts
	_, err := us.userSchemaStore.GetUserSchemaByName(ctx, request.Name)
	if err == nil {
		return nil, &ErrorUserSchemaNameConflict
	} else if !errors.Is(err, ErrUserSchemaNotFound) {
		return nil, logAndReturnServerError(logger, "Failed to check existing user schema", err)
	}

	id, err := utils.GenerateUUIDv7()
	if err != nil {
		logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	userSchema := UserSchema{
		ID:                    id,
		Name:                  request.Name,
		OrganizationUnitID:    request.OrganizationUnitID,
		AllowSelfRegistration: request.AllowSelfRegistration,
		Schema:                request.Schema,
	}

	if err := us.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := us.userSchemaStore.CreateUserSchema(txCtx, userSchema); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, logAndReturnServerError(logger, "Failed to create user schema", err)
	}

	return &userSchema, nil
}

// GetUserSchema retrieves a user schema by its ID.
func (us *userSchemaService) GetUserSchema(
	ctx context.Context, schemaID string,
) (*UserSchema, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if schemaID == "" {
		return nil, invalidSchemaRequestError("schema id must not be empty")
	}

	userSchema, err := us.userSchemaStore.GetUserSchemaByID(ctx, schemaID)
	if err != nil {
		if errors.Is(err, ErrUserSchemaNotFound) {
			return nil, &ErrorUserSchemaNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to get user schema", err)
	}

	return &userSchema, nil
}

// GetUserSchemaByName retrieves a user schema by its name.
func (us *userSchemaService) GetUserSchemaByName(
	ctx context.Context, schemaName string,
) (*UserSchema, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if schemaName == "" {
		return nil, invalidSchemaRequestError("schema name must not be empty")
	}

	userSchema, err := us.userSchemaStore.GetUserSchemaByName(ctx, schemaName)
	if err != nil {
		if errors.Is(err, ErrUserSchemaNotFound) {
			return nil, &ErrorUserSchemaNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to get user schema by name", err)
	}

	return &userSchema, nil
}

// UpdateUserSchema updates a user schema by its ID.
func (us *userSchemaService) UpdateUserSchema(ctx context.Context, schemaID string, request UpdateUserSchemaRequest) (
	*UserSchema, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if err := declarativeresource.CheckDeclarativeUpdate(); err != nil {
		return nil, err
	}

	if schemaID == "" {
		return nil, invalidSchemaRequestError("schema id must not be empty")
	}

	// Validate the schema definition
	schemaToValidate := UserSchema{
		Name:               request.Name,
		OrganizationUnitID: request.OrganizationUnitID,
		Schema:             request.Schema,
	}
	if validationErr := validateUserSchemaDefinition(schemaToValidate); validationErr != nil {
		logger.Debug("User schema validation failed", log.String("id", schemaID))
		return nil, validationErr
	}

	// Ensure organization unit exists
	if svcErr := us.ensureOrganizationUnitExists(request.OrganizationUnitID, logger); svcErr != nil {
		return nil, svcErr
	}

	existingSchema, err := us.userSchemaStore.GetUserSchemaByID(ctx, schemaID)
	if err != nil {
		if errors.Is(err, ErrUserSchemaNotFound) {
			return nil, &ErrorUserSchemaNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to get existing user schema", err)
	}

	if request.Name != existingSchema.Name {
		_, err := us.userSchemaStore.GetUserSchemaByName(ctx, request.Name)
		if err == nil {
			return nil, &ErrorUserSchemaNameConflict
		} else if !errors.Is(err, ErrUserSchemaNotFound) {
			return nil, logAndReturnServerError(logger, "Failed to check existing user schema", err)
		}
	}

	userSchema := UserSchema{
		ID:                    schemaID,
		Name:                  request.Name,
		OrganizationUnitID:    request.OrganizationUnitID,
		AllowSelfRegistration: request.AllowSelfRegistration,
		Schema:                request.Schema,
	}

	if err := us.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := us.userSchemaStore.UpdateUserSchemaByID(txCtx, schemaID, userSchema); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, logAndReturnServerError(logger, "Failed to update user schema", err)
	}

	return &userSchema, nil
}

// DeleteUserSchema deletes a user schema by its ID.
func (us *userSchemaService) DeleteUserSchema(ctx context.Context, schemaID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if err := declarativeresource.CheckDeclarativeDelete(); err != nil {
		return err
	}

	if schemaID == "" {
		return invalidSchemaRequestError("schema id must not be empty")
	}

	if err := us.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := us.userSchemaStore.DeleteUserSchemaByID(txCtx, schemaID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return logAndReturnServerError(logger, "Failed to delete user schema", err)
	}

	return nil
}

// ValidateUser validates user attributes against the user schema for the given user type.
func (us *userSchemaService) ValidateUser(
	ctx context.Context,
	userType string, userAttributes json.RawMessage,
) (bool, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	compiledSchema, err := us.getCompiledSchemaForUserType(ctx, userType, logger)
	if err != nil {
		if errors.Is(err, ErrUserSchemaNotFound) {
			return false, &ErrorUserSchemaNotFound
		}
		return false, logAndReturnServerError(logger, "Failed to load user schema", err)
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
func (us *userSchemaService) ValidateUserUniqueness(
	ctx context.Context,
	userType string,
	userAttributes json.RawMessage,
	identifyUser func(map[string]interface{}) (*string, error),
) (bool, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	compiledSchema, err := us.getCompiledSchemaForUserType(ctx, userType, logger)
	if err != nil {
		if errors.Is(err, ErrUserSchemaNotFound) {
			return false, &ErrorUserSchemaNotFound
		}
		return false, logAndReturnServerError(logger, "Failed to load user schema", err)
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

// GetCredentialAttributes returns the names of schema properties marked as credentials for a given user type.
func (us *userSchemaService) GetCredentialAttributes(
	ctx context.Context, userType string,
) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	compiledSchema, err := us.getCompiledSchemaForUserType(ctx, userType, logger)
	if err != nil {
		if errors.Is(err, ErrUserSchemaNotFound) {
			return nil, &ErrorUserSchemaNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to load user schema for credential attributes", err)
	}

	return compiledSchema.GetCredentialAttributes(), nil
}

func (us *userSchemaService) getCompiledSchemaForUserType(
	ctx context.Context,
	userType string,
	logger *log.Logger,
) (*model.Schema, error) {
	if userType == "" {
		return nil, ErrUserSchemaNotFound
	}

	userSchema, err := us.userSchemaStore.GetUserSchemaByName(ctx, userType)
	if err != nil {
		return nil, err
	}

	compiled, err := model.CompileUserSchema(userSchema.Schema)
	if err != nil {
		logger.Error("Failed to compile stored user schema", log.String("userType", userType), log.Error(err))
		return nil, fmt.Errorf("failed to compile stored user schema: %w", err)
	}

	return compiled, nil
}

// ensureOrganizationUnitExists validates that the provided organization unit exists using the OU service.
func (us *userSchemaService) ensureOrganizationUnitExists(
	organizationUnitID string,
	logger *log.Logger,
) *serviceerror.ServiceError {
	if us.ouService == nil {
		logger.Error("Organization unit service is not configured for user schema operations")
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
			Href: fmt.Sprintf("/user-schemas?offset=0&limit=%d", limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, Link{
			Href: fmt.Sprintf("/user-schemas?offset=%d&limit=%d", prevOffset, limit),
			Rel:  "prev",
		})
	}

	if offset+limit < totalCount {
		nextOffset := offset + limit
		links = append(links, Link{
			Href: fmt.Sprintf("/user-schemas?offset=%d&limit=%d", nextOffset, limit),
			Rel:  "next",
		})
	}

	lastPageOffset := ((totalCount - 1) / limit) * limit
	if offset < lastPageOffset {
		links = append(links, Link{
			Href: fmt.Sprintf("/user-schemas?offset=%d&limit=%d", lastPageOffset, limit),
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

// validateUserSchemaDefinition validates the user schema definition without checking OU existence.
// This is used during initialization to validate file-based configurations.
func validateUserSchemaDefinition(schema UserSchema) *serviceerror.ServiceError {
	logger := log.GetLogger()

	if schema.Name == "" {
		logger.Debug("User schema validation failed: name is empty")
		return invalidSchemaRequestError("user schema name must not be empty")
	}

	if schema.OrganizationUnitID == "" {
		logger.Debug("User schema validation failed: organization unit ID is empty")
		return invalidSchemaRequestError("organization unit id must not be empty")
	}

	if !utils.IsValidUUID(schema.OrganizationUnitID) {
		logger.Debug("User schema validation failed: invalid organization unit ID format",
			log.String("ouId", schema.OrganizationUnitID))
		return invalidSchemaRequestError("organization unit id is not a valid UUID")
	}

	if len(schema.Schema) == 0 {
		logger.Debug("User schema validation failed: schema definition is empty")
		return invalidSchemaRequestError("schema definition must not be empty")
	}

	_, err := model.CompileUserSchema(schema.Schema)
	if err != nil {
		logger.Debug("User schema validation failed: schema compilation error",
			log.Error(err))
		return invalidSchemaRequestError(err.Error())
	}

	return nil
}

func invalidSchemaRequestError(detail string) *serviceerror.ServiceError {
	err := ErrorInvalidUserSchemaRequest
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
