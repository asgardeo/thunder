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

	"github.com/asgardeo/thunder/internal/consent"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
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
	GetUniqueAttributes(
		ctx context.Context, userType string,
	) ([]string, *serviceerror.ServiceError)
	GetDisplayAttributesByNames(
		ctx context.Context, names []string,
	) (map[string]string, *serviceerror.ServiceError)
}

// userSchemaService is the default implementation of the UserSchemaServiceInterface.
type userSchemaService struct {
	userSchemaStore userSchemaStoreInterface
	ouService       oupkg.OrganizationUnitServiceInterface
	transactioner   transaction.Transactioner
	authzService    sysauthz.SystemAuthorizationServiceInterface
	consentService  consent.ConsentServiceInterface
}

// newUserSchemaService creates a new instance of userSchemaService.
func newUserSchemaService(
	ouService oupkg.OrganizationUnitServiceInterface,
	store userSchemaStoreInterface,
	transactioner transaction.Transactioner,
	authzService sysauthz.SystemAuthorizationServiceInterface,
	consentService consent.ConsentServiceInterface,
) UserSchemaServiceInterface {
	return &userSchemaService{
		userSchemaStore: store,
		ouService:       ouService,
		transactioner:   transactioner,
		authzService:    authzService,
		consentService:  consentService,
	}
}

// GetUserSchemaList lists the user schemas with pagination.
func (us *userSchemaService) GetUserSchemaList(ctx context.Context, limit, offset int) (
	*UserSchemaListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	// Resolve the set of user schemas the caller is authorized to see.
	accessible, svcErr := us.getAccessibleResources(ctx, security.ActionListUserSchemas)
	if svcErr != nil {
		return nil, svcErr
	}

	// Unfiltered path: the caller can see all user schemas.
	if accessible.AllAllowed {
		logger.Info("Caller has access to all user schemas, retrieving without OU filtering")
		return us.listAllUserSchemas(ctx, limit, offset, logger)
	}

	// Filtered path: the caller has a restricted set of accessible OUs.
	return us.listAccessibleUserSchemas(ctx, accessible.IDs, limit, offset, logger)
}

// listAllUserSchemas retrieves user schemas without authorization filtering.
func (us *userSchemaService) listAllUserSchemas(
	ctx context.Context, limit, offset int, logger *log.Logger,
) (*UserSchemaListResponse, *serviceerror.ServiceError) {
	totalCount, err := us.userSchemaStore.GetUserSchemaListCount(ctx)
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get user schema list count", err)
	}

	userSchemas, err := us.userSchemaStore.GetUserSchemaList(ctx, limit, offset)
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get user schema list", err)
	}

	return &UserSchemaListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(userSchemas),
		Schemas:      userSchemas,
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}, nil
}

// listAccessibleUserSchemas retrieves only the user schemas belonging to the caller's accessible OUs.
func (us *userSchemaService) listAccessibleUserSchemas(
	ctx context.Context, ouIDs []string, limit, offset int, logger *log.Logger,
) (*UserSchemaListResponse, *serviceerror.ServiceError) {
	if len(ouIDs) == 0 {
		return &UserSchemaListResponse{
			TotalResults: 0,
			StartIndex:   offset + 1,
			Count:        0,
			Schemas:      []UserSchemaListItem{},
			Links:        buildPaginationLinks(limit, offset, 0),
		}, nil
	}

	totalCount, err := us.userSchemaStore.GetUserSchemaListCountByOUIDs(ctx, ouIDs)
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get accessible user schema count", err)
	}

	userSchemas, err := us.userSchemaStore.GetUserSchemaListByOUIDs(ctx, ouIDs, limit, offset)
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get accessible user schema list", err)
	}

	return &UserSchemaListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(userSchemas),
		Schemas:      userSchemas,
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}, nil
}

// CreateUserSchema creates a new user schema.
func (us *userSchemaService) CreateUserSchema(
	ctx context.Context, request CreateUserSchemaRequest,
) (*UserSchema, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if isDeclarativeModeEnabled() {
		return nil, &ErrorCannotModifyDeclarativeResource
	}

	// Validate the schema definition
	schemaToValidate := UserSchema{
		Name:             request.Name,
		OUID:             request.OUID,
		SystemAttributes: request.SystemAttributes,
		Schema:           request.Schema,
	}
	if validationErr := validateUserSchemaDefinition(schemaToValidate); validationErr != nil {
		logger.Debug("User schema validation failed", log.String("name", request.Name))
		return nil, validationErr
	}

	// Ensure organization unit exists
	if svcErr := us.ensureOrganizationUnitExists(
		ctx, request.OUID, logger); svcErr != nil {
		return nil, svcErr
	}

	// Check authorization
	if svcErr := us.checkUserSchemaAccess(
		ctx, security.ActionCreateUserSchema, request.OUID); svcErr != nil {
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
		OUID:                  request.OUID,
		AllowSelfRegistration: request.AllowSelfRegistration,
		SystemAttributes:      request.SystemAttributes,
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

	// Sync consent elements for the created schema
	if us.consentService.IsEnabled() {
		if svcErr := us.syncConsentElementsOnCreate(ctx, userSchema.Schema, logger); svcErr != nil {
			if delErr := us.userSchemaStore.DeleteUserSchemaByID(ctx, userSchema.ID); delErr != nil {
				logger.Error("Failed to compensate schema creation after consent sync failure",
					log.String("schemaID", userSchema.ID), log.Error(delErr))
			}

			return nil, svcErr
		}
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

	// Check authorization
	if svcErr := us.checkUserSchemaAccess(
		ctx, security.ActionReadUserSchema, userSchema.OUID); svcErr != nil {
		return nil, svcErr
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

	// Check authorization
	if svcErr := us.checkUserSchemaAccess(
		ctx, security.ActionReadUserSchema, userSchema.OUID); svcErr != nil {
		return nil, svcErr
	}

	return &userSchema, nil
}

// UpdateUserSchema updates a user schema by its ID.
func (us *userSchemaService) UpdateUserSchema(ctx context.Context, schemaID string, request UpdateUserSchemaRequest) (
	*UserSchema, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if schemaID == "" {
		return nil, invalidSchemaRequestError("schema id must not be empty")
	}

	// Check if schema is declarative (immutable) in composite mode
	if us.userSchemaStore.IsUserSchemaDeclarative(schemaID) {
		return nil, &ErrorCannotModifyDeclarativeResource
	}

	// Validate the schema definition
	schemaToValidate := UserSchema{
		Name:             request.Name,
		OUID:             request.OUID,
		SystemAttributes: request.SystemAttributes,
		Schema:           request.Schema,
	}
	if validationErr := validateUserSchemaDefinition(schemaToValidate); validationErr != nil {
		logger.Debug("User schema validation failed", log.String("id", schemaID))
		return nil, validationErr
	}

	// Ensure organization unit exists
	if svcErr := us.ensureOrganizationUnitExists(
		ctx, request.OUID, logger); svcErr != nil {
		return nil, svcErr
	}

	existingSchema, err := us.userSchemaStore.GetUserSchemaByID(ctx, schemaID)
	if err != nil {
		if errors.Is(err, ErrUserSchemaNotFound) {
			return nil, &ErrorUserSchemaNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to get existing user schema", err)
	}

	// Check authorization
	if svcErr := us.checkUserSchemaAccess(
		ctx, security.ActionUpdateUserSchema, existingSchema.OUID); svcErr != nil {
		return nil, svcErr
	}

	// If OU is being changed, validate access to the target OU as well.
	if request.OUID != existingSchema.OUID {
		if svcErr := us.checkUserSchemaAccess(
			ctx, security.ActionUpdateUserSchema, request.OUID); svcErr != nil {
			return nil, svcErr
		}
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
		OUID:                  request.OUID,
		AllowSelfRegistration: request.AllowSelfRegistration,
		SystemAttributes:      request.SystemAttributes,
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

	// Sync consent elements for the updated schema
	if us.consentService.IsEnabled() {
		if svcErr := us.syncConsentElementsOnUpdate(ctx, existingSchema.Schema,
			userSchema.Schema, logger); svcErr != nil {
			if revertErr := us.userSchemaStore.UpdateUserSchemaByID(ctx, schemaID,
				existingSchema); revertErr != nil {
				logger.Error("Failed to compensate schema update after consent sync failure",
					log.String("schemaID", schemaID), log.Error(revertErr))
			}

			return nil, svcErr
		}
	}

	return &userSchema, nil
}

// DeleteUserSchema deletes a user schema by its ID.
func (us *userSchemaService) DeleteUserSchema(ctx context.Context, schemaID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if schemaID == "" {
		return invalidSchemaRequestError("schema id must not be empty")
	}

	// Fetch the schema to get its OU ID for authorization check.
	existingSchema, err := us.userSchemaStore.GetUserSchemaByID(ctx, schemaID)
	if err != nil {
		if errors.Is(err, ErrUserSchemaNotFound) {
			// Check authorization before revealing whether the schema exists.
			if svcErr := us.checkUserSchemaAccess(
				ctx, security.ActionDeleteUserSchema, ""); svcErr != nil {
				return svcErr
			}
			// Authorized caller — schema doesn't exist, return nil for idempotent delete.
			return nil
		}
		return logAndReturnServerError(logger, "Failed to get user schema for delete", err)
	}

	// Check authorization against the schema's OU.
	if svcErr := us.checkUserSchemaAccess(
		ctx, security.ActionDeleteUserSchema, existingSchema.OUID); svcErr != nil {
		return svcErr
	}

	// Check if schema is declarative (immutable) in composite mode
	if us.userSchemaStore.IsUserSchemaDeclarative(schemaID) {
		return &ErrorCannotModifyDeclarativeResource
	}

	// Extract attribute names from the existing schema before deletion to identify associated
	// consent elements for cleanup
	var attributeNames []string
	if us.consentService.IsEnabled() {
		attrNames, err := extractAttributeNames(existingSchema.Schema)
		if err != nil {
			logger.Error("Failed to extract attribute names for consent cleanup; proceeding with schema deletion",
				log.String("schemaID", schemaID), log.Any("error", err))
			attributeNames = []string{}
		} else {
			attributeNames = attrNames
		}
	}

	if err := us.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return us.userSchemaStore.DeleteUserSchemaByID(txCtx, schemaID)
	}); err != nil {
		return logAndReturnServerError(logger, "Failed to delete user schema", err)
	}

	// Sync consent elements for the deleted schema by deleting the associated consent elements
	// If consent deletion fails, we log the error but do NOT re-create the schema
	// since orphaned consent elements are safe and won't cause active harm.
	if us.consentService.IsEnabled() && len(attributeNames) > 0 {
		if svcErr := us.deleteConsentElements(ctx, attributeNames, logger); svcErr != nil {
			logger.Error("Failed to delete consent elements for removed schema attributes; "+
				"orphaned consent elements may remain but schema deletion succeeded",
				log.Any("attributeNames", attributeNames), log.Any("error", svcErr))
		}
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

// GetUniqueAttributes returns the names of schema properties marked as unique for a given user type.
func (us *userSchemaService) GetUniqueAttributes(
	ctx context.Context, userType string,
) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	compiledSchema, err := us.getCompiledSchemaForUserType(ctx, userType, logger)
	if err != nil {
		if errors.Is(err, ErrUserSchemaNotFound) {
			return nil, &ErrorUserSchemaNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to load user schema for unique attributes", err)
	}

	return compiledSchema.GetUniqueAttributes(), nil
}

// GetDisplayAttributesByNames returns display attributes for multiple user schemas by name.
func (us *userSchemaService) GetDisplayAttributesByNames(
	ctx context.Context, names []string,
) (map[string]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if len(names) == 0 {
		return map[string]string{}, nil
	}

	result, err := us.userSchemaStore.GetDisplayAttributesByNames(ctx, names)
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get display attributes by names", err)
	}

	return result, nil
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

// checkUserSchemaAccess validates that the caller is authorized to perform the given action on a user schema.
// Pass the user schema's OU ID to scope the authorization check to the caller's organization unit membership.
func (us *userSchemaService) checkUserSchemaAccess(
	ctx context.Context, action security.Action, ouID string,
) *serviceerror.ServiceError {
	if us.authzService == nil {
		return nil
	}
	allowed, svcErr := us.authzService.IsActionAllowed(ctx, action,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeUserSchema, OUID: ouID})
	if svcErr != nil {
		return &ErrorInternalServerError
	}
	if !allowed {
		return &serviceerror.ErrorUnauthorized
	}
	return nil
}

// getAccessibleResources returns the set of OU IDs the caller is permitted to access for user schemas.
func (us *userSchemaService) getAccessibleResources(
	ctx context.Context, action security.Action,
) (*sysauthz.AccessibleResources, *serviceerror.ServiceError) {
	if us.authzService == nil {
		return &sysauthz.AccessibleResources{AllAllowed: true}, nil
	}
	accessible, svcErr := us.authzService.GetAccessibleResources(
		ctx, action, security.ResourceTypeUserSchema)
	if svcErr != nil {
		return nil, &ErrorInternalServerError
	}
	return accessible, nil
}

// ensureOrganizationUnitExists validates that the provided organization unit exists using the OU service.
func (us *userSchemaService) ensureOrganizationUnitExists(
	ctx context.Context,
	oUID string,
	logger *log.Logger,
) *serviceerror.ServiceError {
	if us.ouService == nil {
		logger.Error("Organization unit service is not configured for user schema operations")
		return &ErrorInternalServerError
	}

	exists, svcErr := us.ouService.IsOrganizationUnitExists(ctx, oUID)
	if svcErr != nil {
		logger.Error("Failed to verify organization unit existence",
			log.String("oUID", oUID), log.Any("error", svcErr))
		return &ErrorInternalServerError
	}

	if !exists {
		logger.Debug("Organization unit does not exist",
			log.String("oUID", oUID))
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

	if schema.OUID == "" {
		logger.Debug("User schema validation failed: organization unit ID is empty")
		return invalidSchemaRequestError("organization unit id must not be empty")
	}

	if len(schema.Schema) == 0 {
		logger.Debug("User schema validation failed: schema definition is empty")
		return invalidSchemaRequestError("schema definition must not be empty")
	}

	compiledSchema, err := model.CompileUserSchema(schema.Schema)
	if err != nil {
		logger.Debug("User schema validation failed: schema compilation error",
			log.Error(err))
		return invalidSchemaRequestError(err.Error())
	}

	return validateSystemAttributes(compiledSchema, schema.SystemAttributes)
}

// validateSystemAttributes validates the system attributes against the compiled schema.
func validateSystemAttributes(
	compiledSchema *model.Schema, systemAttrs *SystemAttributes,
) *serviceerror.ServiceError {
	if systemAttrs == nil {
		return nil
	}

	return validateDisplayAttribute(compiledSchema, systemAttrs.Display)
}

// validateDisplayAttribute validates that the display attribute, if provided,
// references an existing, displayable, non-credential attribute in the compiled schema.
// Only string and number types are considered displayable.
func validateDisplayAttribute(
	compiledSchema *model.Schema, display string,
) *serviceerror.ServiceError {
	if display == "" {
		return nil
	}

	switch compiledSchema.ValidateAsDisplayAttribute(display) {
	case model.DisplayAttributeNotFound:
		return &ErrorInvalidDisplayAttribute
	case model.DisplayAttributeNotDisplayable:
		return &ErrorNonDisplayableAttribute
	case model.DisplayAttributeIsCredential:
		return &ErrorCredentialDisplayAttribute
	default:
		return nil
	}
}

// syncConsentElementsOnCreate creates missing consent elements for a new schema creation.
func (us *userSchemaService) syncConsentElementsOnCreate(ctx context.Context,
	schema json.RawMessage, logger *log.Logger) *serviceerror.ServiceError {
	// TODO: Replace "default" with the schema's actual OU when applications are associated with OUs.
	const ouID = "default"

	logger.Debug("Synchronizing consent elements for the new schema", log.String("ouID", ouID))

	names, err := extractAttributeNames(schema)
	if err != nil {
		return err
	}

	if len(names) > 0 {
		logger.Debug("Creating missing consent elements for the new schema",
			log.String("ouID", ouID), log.Int("elementCount", len(names)))
		if svcErr := us.createMissingConsentElements(ctx, ouID, names, logger); svcErr != nil {
			return svcErr
		}
	}

	return nil
}

// syncConsentElementsOnUpdate reconciles consent elements when a schema is updated.
// It creates elements that were added and deletes elements that were removed.
func (us *userSchemaService) syncConsentElementsOnUpdate(ctx context.Context,
	oldSchema, newSchema json.RawMessage, logger *log.Logger) *serviceerror.ServiceError {
	// TODO: Replace "default" with the schema's actual OU when applications are associated with OUs.
	const ouID = "default"

	logger.Debug("Synchronizing consent elements for the updated schema", log.String("ouID", ouID))

	oldAttrs, err := extractAttributeNamesAsMap(oldSchema)
	if err != nil {
		return err
	}

	newAttrs, err := extractAttributeNamesAsMap(newSchema)
	if err != nil {
		return err
	}

	// Create consent elements for new attributes that were added in the updated schema.
	// createMissingConsentElements method will handle filtering out existing elements, so we can pass all
	// new attribute names here. This ensures that even consent service was disabled when creating the schema,
	// the necessary consent elements are created when updating the schema with consent service enabled.
	requiredNames := make([]string, 0, len(newAttrs))
	for name := range newAttrs {
		requiredNames = append(requiredNames, name)
	}

	if len(requiredNames) > 0 {
		logger.Debug("Ensuring consent elements exist for all requested attributes",
			log.String("ouID", ouID), log.Int("requiredAttributesCount", len(requiredNames)))
		if err := us.createMissingConsentElements(ctx, ouID, requiredNames, logger); err != nil {
			return err
		}
	}

	// Delete variables that are no longer part of the current payload
	var removedNames []string
	for name := range oldAttrs {
		if _, exists := newAttrs[name]; !exists {
			removedNames = append(removedNames, name)
		}
	}

	return us.deleteConsentElements(ctx, removedNames, logger)
}

// createMissingConsentElements validates a list of consent element names and creates only
// the missing ones.
// nolint:unparam // ouID is always "default" in current usage but kept for future flexibility
func (us *userSchemaService) createMissingConsentElements(ctx context.Context,
	ouID string, names []string, logger *log.Logger) *serviceerror.ServiceError {
	if len(names) == 0 {
		logger.Debug("No consent elements to create for the schema", log.String("ouID", ouID))
		return nil
	}

	logger.Debug("Validating consent elements for the schema attributes",
		log.String("ouID", ouID), log.Int("elementCount", len(names)))

	validNames, err := us.consentService.ValidateConsentElements(ctx, ouID, names)
	if err != nil {
		return wrapConsentServiceError(err, logger)
	}

	// Create a map of existing elements for fast lookup
	existingMap := make(map[string]bool, len(validNames))
	for _, name := range validNames {
		existingMap[name] = true
	}

	// Filter out the existing elements
	var elementsToCreate []consent.ConsentElementInput
	for _, name := range names {
		if !existingMap[name] {
			elementsToCreate = append(elementsToCreate, consent.ConsentElementInput{
				Name:      name,
				Namespace: consent.NamespaceAttribute,
			})
		}
	}

	if len(elementsToCreate) > 0 {
		logger.Debug("Creating new consent elements for the schema attributes",
			log.String("ouID", ouID), log.Int("elementCount", len(elementsToCreate)))
		if _, err := us.consentService.CreateConsentElements(ctx, ouID, elementsToCreate); err != nil {
			return wrapConsentServiceError(err, logger)
		}
	}

	return nil
}

// deleteConsentElements removes a list of consent elements associated with the given attribute names.
func (us *userSchemaService) deleteConsentElements(ctx context.Context,
	attributeNames []string, logger *log.Logger) *serviceerror.ServiceError {
	// TODO: Replace "default" with the schema's actual OU when applications are associated with OUs.
	const ouID = "default"

	logger.Debug("Deleting consent elements for the removed schema attributes",
		log.String("ouID", ouID), log.Int("elementCount", len(attributeNames)))

	if len(attributeNames) == 0 {
		logger.Debug("No consent elements to delete for the schema", log.String("ouID", ouID))
		return nil
	}

	for _, attrName := range attributeNames {
		// List existing consent elements for the removed attribute to find their IDs for deletion
		existing, err := us.consentService.ListConsentElements(ctx, ouID, consent.NamespaceAttribute, attrName)
		if err != nil {
			return wrapConsentServiceError(err, logger)
		}

		// Delete the first element if the list is not empty.
		// We assume there is only one consent element per attribute name.
		// TODO: This should be revisited when user type separation is onboarded to consent elements.
		if len(existing) > 0 {
			logger.Debug("Deleting consent element for the removed schema attribute",
				log.String("ouID", ouID), log.String("attribute", attrName), log.String("elementID", existing[0].ID))
			if err := us.consentService.DeleteConsentElement(ctx, ouID, existing[0].ID); err != nil {
				// Silently ignore the error if it's due to associated purposes, but log a warning.
				// The same attribute can exist in a different schema and purpose can be associated with that,
				// so we should not block the schema update in that case.
				// If it's not associated with a purpose, but exists in a different schema, we still delete it,
				// as the consent element can be created again when configuring attribute for a application.
				if err.Code == consent.ErrorDeletingConsentElementWithAssociatedPurpose.Code {
					logger.Warn("Cannot delete consent element for removed attribute due to associated purposes",
						log.String("attribute", attrName), log.String("elementID", existing[0].ID),
						log.String("error", err.ErrorDescription.DefaultValue))
					continue
				}

				return wrapConsentServiceError(err, logger)
			}
		}
	}

	return nil
}

// invalidSchemaRequestError creates a service error for invalid user schema requests
// with an optional detail message.
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

// extractAttributeNames returns the set of attribute names from a schema JSON as a string slice.
func extractAttributeNames(schema json.RawMessage) ([]string, *serviceerror.ServiceError) {
	if len(schema) == 0 {
		return nil, nil
	}

	var schemaMap map[string]json.RawMessage
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return nil, invalidSchemaRequestError("invalid schema json: " + err.Error())
	}

	names := make([]string, 0, len(schemaMap))
	for name := range schemaMap {
		names = append(names, name)
	}

	return names, nil
}

// extractAttributeNamesAsMap returns the set of attribute names from a schema JSON as a map
// for last lookups.
func extractAttributeNamesAsMap(schema json.RawMessage) (map[string]bool, *serviceerror.ServiceError) {
	result := make(map[string]bool)
	if len(schema) == 0 {
		return result, nil
	}

	var schemaMap map[string]json.RawMessage
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return nil, invalidSchemaRequestError("invalid schema json: " + err.Error())
	}

	for name := range schemaMap {
		result[name] = true
	}

	return result, nil
}

// wrapConsentServiceError converts an I18nServiceError from the consent service into a ServiceError
// for the user schema service.
func wrapConsentServiceError(err *serviceerror.I18nServiceError, logger *log.Logger) *serviceerror.ServiceError {
	if err == nil {
		return nil
	}

	if err.Type == serviceerror.ClientErrorType {
		logger.Debug("Failed to sync consent elements for the schema changes", log.Any("error", err))
		return serviceerror.CustomServiceError(ErrorConsentSyncFailed,
			fmt.Sprintf(ErrorConsentSyncFailed.ErrorDescription+" : code - %s", err.Code))
	}

	logger.Error("Failed to sync consent elements for the schema changes", log.Any("error", err))
	return &ErrorInternalServerError
}
