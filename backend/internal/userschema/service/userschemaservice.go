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

// Package service provides the implementation for user schema management operations.
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema/constants"
	"github.com/asgardeo/thunder/internal/userschema/model"
	"github.com/asgardeo/thunder/internal/userschema/store"
)

const userSchemaLoggerComponentName = "UserSchemaService"

// UserSchemaServiceInterface defines the interface for the user schema service.
type UserSchemaServiceInterface interface {
	GetUserSchemaList(limit, offset int) (*model.UserSchemaListResponse, *serviceerror.ServiceError)
	CreateUserSchema(request model.CreateUserSchemaRequest) (*model.UserSchema, *serviceerror.ServiceError)
	GetUserSchema(schemaID string) (*model.UserSchema, *serviceerror.ServiceError)
	UpdateUserSchema(schemaID string, request model.UpdateUserSchemaRequest) (
		*model.UserSchema, *serviceerror.ServiceError)
	DeleteUserSchema(schemaID string) *serviceerror.ServiceError
	ValidateUser(userType string, userAttributes json.RawMessage) (bool, *serviceerror.ServiceError)
	ValidateUserUniqueness(userType string, userAttributes json.RawMessage,
		identifyUser func(map[string]interface{}) (*string, error)) (bool, *serviceerror.ServiceError)
}

// UserSchemaService is the default implementation of the UserSchemaServiceInterface.
type UserSchemaService struct{}

// GetUserSchemaService creates a new instance of UserSchemaService.
func GetUserSchemaService() UserSchemaServiceInterface {
	return &UserSchemaService{}
}

// GetUserSchemaList lists the user schemas with pagination.
func (us *UserSchemaService) GetUserSchemaList(limit, offset int) (
	*model.UserSchemaListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := store.GetUserSchemaListCount()
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get user schema list count", err)
	}

	userSchemas, err := store.GetUserSchemaList(limit, offset)
	if err != nil {
		return nil, logAndReturnServerError(logger, "Failed to get user schema list", err)
	}

	response := &model.UserSchemaListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(userSchemas),
		Schemas:      userSchemas,
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}

	return response, nil
}

// CreateUserSchema creates a new user schema.
func (us *UserSchemaService) CreateUserSchema(request model.CreateUserSchemaRequest) (
	*model.UserSchema, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if request.Name == "" {
		return nil, &constants.ErrorInvalidUserSchemaRequest
	}

	if len(request.Schema) == 0 {
		return nil, &constants.ErrorInvalidUserSchemaRequest
	}

	if err := validateJSONSchema(request.Schema); err != nil {
		return nil, &constants.ErrorInvalidUserSchemaRequest
	}

	_, err := store.GetUserSchemaByName(request.Name)
	if err == nil {
		return nil, &constants.ErrorUserSchemaNameConflict
	} else if !errors.Is(err, constants.ErrUserSchemaNotFound) {
		return nil, logAndReturnServerError(logger, "Failed to check existing user schema", err)
	}

	schemaID := utils.GenerateUUID()

	userSchema := model.UserSchema{
		ID:     schemaID,
		Name:   request.Name,
		Schema: request.Schema,
	}

	if err := store.CreateUserSchema(userSchema); err != nil {
		return nil, logAndReturnServerError(logger, "Failed to create user schema", err)
	}

	return &userSchema, nil
}

// GetUserSchema retrieves a user schema by its ID.
func (us *UserSchemaService) GetUserSchema(schemaID string) (*model.UserSchema, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if schemaID == "" {
		return nil, &constants.ErrorInvalidUserSchemaRequest
	}

	userSchema, err := store.GetUserSchemaByID(schemaID)
	if err != nil {
		if errors.Is(err, constants.ErrUserSchemaNotFound) {
			return nil, &constants.ErrorUserSchemaNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to get user schema", err)
	}

	return &userSchema, nil
}

// UpdateUserSchema updates a user schema by its ID.
func (us *UserSchemaService) UpdateUserSchema(schemaID string, request model.UpdateUserSchemaRequest) (
	*model.UserSchema, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if schemaID == "" {
		return nil, &constants.ErrorInvalidUserSchemaRequest
	}

	if request.Name == "" {
		return nil, &constants.ErrorInvalidUserSchemaRequest
	}

	if len(request.Schema) == 0 {
		return nil, &constants.ErrorInvalidUserSchemaRequest
	}

	if err := validateJSONSchema(request.Schema); err != nil {
		return nil, &constants.ErrorInvalidUserSchemaRequest
	}

	existingSchema, err := store.GetUserSchemaByID(schemaID)
	if err != nil {
		if errors.Is(err, constants.ErrUserSchemaNotFound) {
			return nil, &constants.ErrorUserSchemaNotFound
		}
		return nil, logAndReturnServerError(logger, "Failed to get existing user schema", err)
	}

	if request.Name != existingSchema.Name {
		_, err := store.GetUserSchemaByName(request.Name)
		if err == nil {
			return nil, &constants.ErrorUserSchemaNameConflict
		} else if !errors.Is(err, constants.ErrUserSchemaNotFound) {
			return nil, logAndReturnServerError(logger, "Failed to check existing user schema", err)
		}
	}

	userSchema := model.UserSchema{
		ID:     schemaID,
		Name:   request.Name,
		Schema: request.Schema,
	}

	if err := store.UpdateUserSchemaByID(schemaID, userSchema); err != nil {
		return nil, logAndReturnServerError(logger, "Failed to update user schema", err)
	}

	return &userSchema, nil
}

// DeleteUserSchema deletes a user schema by its ID.
func (us *UserSchemaService) DeleteUserSchema(schemaID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if schemaID == "" {
		return &constants.ErrorInvalidUserSchemaRequest
	}

	if err := store.DeleteUserSchemaByID(schemaID); err != nil {
		return logAndReturnServerError(logger, "Failed to delete user schema", err)
	}

	return nil
}

// validatePaginationParams validates the limit and offset parameters.
func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 0 {
		return &constants.ErrorInvalidLimit
	}
	if offset < 0 {
		return &constants.ErrorInvalidOffset
	}
	return nil
}

// buildPaginationLinks builds pagination links for the response.
func buildPaginationLinks(limit, offset, totalCount int) []model.Link {
	links := make([]model.Link, 0)

	if offset > 0 {
		links = append(links, model.Link{
			Href: fmt.Sprintf("/user-schemas?offset=0&limit=%d", limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, model.Link{
			Href: fmt.Sprintf("/user-schemas?offset=%d&limit=%d", prevOffset, limit),
			Rel:  "prev",
		})
	}

	if offset+limit < totalCount {
		nextOffset := offset + limit
		links = append(links, model.Link{
			Href: fmt.Sprintf("/user-schemas?offset=%d&limit=%d", nextOffset, limit),
			Rel:  "next",
		})
	}

	lastPageOffset := ((totalCount - 1) / limit) * limit
	if offset < lastPageOffset {
		links = append(links, model.Link{
			Href: fmt.Sprintf("/user-schemas?offset=%d&limit=%d", lastPageOffset, limit),
			Rel:  "last",
		})
	}

	return links
}

// validateJSONSchema validates that the schema is a valid JSON Schema according to the API specification.
// The schema must be a properties-only object following the UserSchema specification in user.yaml.
func validateJSONSchema(schema json.RawMessage) error {
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if len(schemaMap) == 0 {
		return fmt.Errorf("schema cannot be empty - must contain at least one property definition")
	}

	// The schema is expected to be a properties-only object (additionalProperties map)
	// Each property should follow one of three patterns: leaf, object, or array
	for propName, propDef := range schemaMap {
		if err := validateUserSchemaProperty(propDef); err != nil {
			return fmt.Errorf("invalid property '%s': %w", propName, err)
		}
	}

	return nil
}

// validateUserSchemaProperty validates a property definition according to the UserSchema specification.
// Properties must follow one of three patterns: leaf (string/number/boolean), object, or array.
func validateUserSchemaProperty(propDef interface{}) error {
	propMap, ok := propDef.(map[string]interface{})
	if !ok {
		return fmt.Errorf("property definition must be an object")
	}

	propType, exists := propMap["type"]
	if !exists {
		return fmt.Errorf("missing required 'type' field")
	}

	typeStr, ok := propType.(string)
	if !ok {
		return fmt.Errorf("'type' field must be a string")
	}

	switch typeStr {
	case model.TypeString, model.TypeNumber, model.TypeBoolean:
		return validateLeafProperty(propMap)
	case model.TypeObject:
		return validateObjectProperty(propMap)
	case model.TypeArray:
		return validateArrayProperty(propMap)
	default:
		return fmt.Errorf("invalid type '%s', must be one of: string, number, boolean, object, array", typeStr)
	}
}

// validateLeafProperty validates leaf property definitions (string, number, boolean).
func validateLeafProperty(propMap map[string]interface{}) error {
	// Check for invalid properties
	allowedFields := map[string]bool{
		"type":     true,
		"unique":   true,
		"enum":     true,
		"regex":    true,
		"required": true,
	}
	for field := range propMap {
		if !allowedFields[field] {
			return fmt.Errorf("invalid field '%s' for leaf property", field)
		}
	}

	if unique, exists := propMap["unique"]; exists {
		if _, ok := unique.(bool); !ok {
			return fmt.Errorf("'unique' field must be a boolean")
		}
	}

	if enum, exists := propMap["enum"]; exists {
		enumArray, ok := enum.([]interface{})
		if !ok {
			return fmt.Errorf("'enum' field must be an array")
		}
		if len(enumArray) == 0 {
			return fmt.Errorf("'enum' array cannot be empty")
		}

		// Get the expected type from the property's type field
		propertyType, exists := propMap["type"]
		if !exists {
			return fmt.Errorf("missing required 'type' field")
		}

		expectedType, ok := propertyType.(string)
		if !ok {
			return fmt.Errorf("'type' field must be a string")
		}

		if err := validateEnumItemsType(expectedType, enumArray); err != nil {
			return err
		}
	}

	if regex, exists := propMap["regex"]; exists {
		if _, ok := regex.(string); !ok {
			return fmt.Errorf("'regex' field must be a string")
		}
	}

	if req, exists := propMap["required"]; exists {
		if _, ok := req.(bool); !ok {
			return fmt.Errorf("'required' field must be a boolean")
		}
	}

	return nil
}

// validateObjectProperty validates object property definitions.
func validateObjectProperty(propMap map[string]interface{}) error {
	allowedFields := map[string]bool{
		"type":       true,
		"properties": true,
		"required":   true,
	}
	for field := range propMap {
		if !allowedFields[field] {
			return fmt.Errorf("invalid field '%s' for object property", field)
		}
	}

	properties, exists := propMap["properties"]
	if !exists {
		return fmt.Errorf("missing required 'properties' field for object type")
	}

	propertiesMap, ok := properties.(map[string]interface{})
	if !ok {
		return fmt.Errorf("'properties' field must be an object")
	}

	if req, exists := propMap["required"]; exists {
		if _, ok := req.(bool); !ok {
			return fmt.Errorf("'required' field must be a boolean")
		}
	}

	for nestedPropName, nestedPropDef := range propertiesMap {
		if err := validateUserSchemaProperty(nestedPropDef); err != nil {
			return fmt.Errorf("invalid nested property '%s': %w", nestedPropName, err)
		}
	}

	return nil
}

// validateArrayProperty validates array property definitions.
func validateArrayProperty(propMap map[string]interface{}) error {
	items, exists := propMap["items"]
	if !exists {
		return fmt.Errorf("missing required 'items' field for array type")
	}

	allowedFields := map[string]bool{
		"type":     true,
		"items":    true,
		"required": true,
	}
	for field := range propMap {
		if !allowedFields[field] {
			return fmt.Errorf("invalid field '%s' for array property", field)
		}
	}

	itemsMap, ok := items.(map[string]interface{})
	if !ok {
		return fmt.Errorf("'items' field must be an object")
	}

	if req, exists := propMap["required"]; exists {
		if _, ok := req.(bool); !ok {
			return fmt.Errorf("'required' field must be a boolean")
		}
	}

	itemType, exists := itemsMap["type"]
	if !exists {
		return fmt.Errorf("missing required 'type' field in items definition")
	}

	itemTypeStr, ok := itemType.(string)
	if !ok {
		return fmt.Errorf("'type' field in items must be a string")
	}

	switch itemTypeStr {
	case model.TypeString, model.TypeNumber, model.TypeBoolean:
		return validateArrayLeafItems(itemTypeStr, itemsMap)
	case model.TypeObject:
		return validateObjectProperty(itemsMap)
	default:
		return fmt.Errorf("invalid items type '%s', must be one of: string, number, boolean, object", itemTypeStr)
	}
}

// validateArrayLeafItems validates array items for leaf types (string, number, boolean).
func validateArrayLeafItems(itemType string, itemsMap map[string]interface{}) error {
	itemEnum, exists := itemsMap["enum"]
	if exists {
		itemEnumArray, ok := itemEnum.([]interface{})
		if !ok {
			return fmt.Errorf("'enum' field in items must be an array")
		}
		if len(itemEnumArray) == 0 {
			return fmt.Errorf("'enum' array in items cannot be empty")
		}
		return validateEnumItemsType(itemType, itemEnumArray)
	}
	return nil
}

// validateEnumItemsType validates array enum items for leaf types (string, number, boolean).
func validateEnumItemsType(expectedType string, enumArray []interface{}) error {
	for i, item := range enumArray {
		switch expectedType {
		case model.TypeString:
			if _, ok := item.(string); !ok {
				return fmt.Errorf("'enum' array item at index %d must be a string to match property type", i)
			}
		case model.TypeNumber:
			switch item.(type) {
			case float64, int:
				// Valid number type
			default:
				return fmt.Errorf("'enum' array item at index %d must be a number to match property type", i)
			}
		case model.TypeBoolean:
			if _, ok := item.(bool); !ok {
				return fmt.Errorf("'enum' array item at index %d must be a boolean to match property type", i)
			}
		default:
			return fmt.Errorf("invalid property type '%s' for enum validation", expectedType)
		}
	}
	return nil
}

// logAndReturnServerError logs the error and returns a server error.
func logAndReturnServerError(
	logger *log.Logger,
	message string,
	err error,
) *serviceerror.ServiceError {
	logger.Error(message, log.Error(err))
	return &constants.ErrorInternalServerError
}

// ValidateUser validates user attributes against the user schema for the given user type.
func (us *UserSchemaService) ValidateUser(
	userType string, userAttributes json.RawMessage,
) (bool, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if userType == "" {
		logger.Debug("User type is empty, skipping schema validation")
		return true, nil
	}

	userSchema, err := store.GetUserSchemaByName(userType)
	if err != nil {
		if errors.Is(err, constants.ErrUserSchemaNotFound) {
			logger.Debug("No schema found for user type, skipping validation", log.String("userType", userType))
			return true, nil // Allow users without schema
		}
		return false, logAndReturnServerError(logger, "Failed to get user schema", err)
	}

	isValid, err := validateUserAttributesAgainstSchema(userAttributes, userSchema.Schema, logger)
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
func (u *UserSchemaService) ValidateUserUniqueness(
	userType string,
	userAttributes json.RawMessage,
	identifyUser func(map[string]interface{}) (*string, error),
) (bool, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userSchemaLoggerComponentName))

	if userType == "" {
		return true, nil
	}

	userSchema, err := store.GetUserSchemaByName(userType)
	if err != nil {
		if errors.Is(err, constants.ErrUserSchemaNotFound) {
			return true, nil // No schema, so skip uniqueness validation
		}
		return false, logAndReturnServerError(logger, "Failed to get user schema", err)
	}

	if len(userAttributes) == 0 {
		return true, nil
	}

	var userAttrs map[string]interface{}
	if err := json.Unmarshal(userAttributes, &userAttrs); err != nil {
		return false, logAndReturnServerError(logger, "Failed to unmarshal user attributes", err)
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(userSchema.Schema, &schemaMap); err != nil {
		return false, logAndReturnServerError(logger, "Failed to unmarshal user schema", err)
	}

	if len(schemaMap) == 0 {
		return true, nil
	}

	for propName, propSchema := range schemaMap {
		propSchemaMap, ok := propSchema.(map[string]interface{})
		if !ok {
			continue
		}

		userValue, exists := userAttrs[propName]
		if !exists {
			continue
		}

		isValid, err := validateUserSchemaPropertyForUniqueness(propName, propSchemaMap, userValue, identifyUser, logger)
		if err != nil {
			errMsg := fmt.Sprintf("server error during uniqueness check for '%s'", propName)
			return false, logAndReturnServerError(logger, errMsg, err)
		}

		if !isValid {
			logger.Debug("User attribute failed uniqueness validation",
				log.String("userType", userType), log.String("property", propName))
			return false, nil
		}
	}
	return true, nil
}

// validateUserSchemaPropertyForUniqueness validates a unique property according to the UserSchema specification.
// Properties must follow one of three patterns: leaf (string/number/boolean), object, or array.
func validateUserSchemaPropertyForUniqueness(
	propName string,
	propSchemaMap map[string]interface{},
	value interface{},
	identifyUser func(map[string]interface{}) (*string, error),
	logger *log.Logger,
) (bool, error) {
	propType, exists := propSchemaMap["type"]
	if !exists {
		return false, fmt.Errorf("missing required 'type' field")
	}

	typeStr, ok := propType.(string)
	if !ok {
		return false, fmt.Errorf("invalid 'type' field")
	}

	switch typeStr {
	case "string", "number", "boolean":
		if unique, exists := propSchemaMap["unique"]; exists {
			if uniqueBool, ok := unique.(bool); ok && uniqueBool {
				filter := map[string]interface{}{propName: value}

				existingUserID, err := identifyUser(filter)

				if err != nil {
					return false, err
				}

				if existingUserID != nil {
					return false, nil
				}
			}
		}
		return true, nil
	case "object":
		value, ok := value.(map[string]interface{})
		if !ok {
			logger.Debug("Expected object but got different type",
				log.String("property", propName),
				log.String("value", fmt.Sprintf("%v", value)))
			return false, nil
		}
		properties, exists := propSchemaMap["properties"]
		if !exists {
			return false, fmt.Errorf("missing required 'properties' field")
		}

		propertiesMap, ok := properties.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("invalid 'properties' field")
		}

		// Validate nested properties
		for nestedPropName, nestedPropDef := range propertiesMap {
			nestedPropDefMap, ok := nestedPropDef.(map[string]interface{})
			if !ok {
				continue
			}

			nestedValue, exists := value[nestedPropName]
			if !exists {
				continue
			}

			fullNestedPropName := propName + "." + nestedPropName
			isValid, err := validateUserSchemaPropertyForUniqueness(fullNestedPropName, nestedPropDefMap,
				nestedValue, identifyUser, logger)
			if err != nil {
				return false, err
			}

			if !isValid {
				logger.Debug("User attribute failed uniqueness validation", log.String("property", fullNestedPropName))
				return false, nil
			}
		}

		return true, nil
	case "array":
		// Arrays are not supported for uniqueness validation
		return true, nil
	}

	return true, nil
}

// validateUserAttributesAgainstSchema validates the user attributes against the JSON schema.
func validateUserAttributesAgainstSchema(
	attributes json.RawMessage,
	schema json.RawMessage,
	logger *log.Logger,
) (bool, error) {
	if len(attributes) == 0 {
		logger.Debug("User has no attributes to validate")
		return true, nil
	}

	var userAttrs map[string]interface{}
	if err := json.Unmarshal(attributes, &userAttrs); err != nil {
		return false, fmt.Errorf("failed to unmarshal user attributes: %w", err)
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return false, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	for propName, propDef := range schemaMap {
		propDefMap, ok := propDef.(map[string]interface{})
		if !ok {
			continue
		}

		userValue, exists := userAttrs[propName]
		if !exists {
			isRequired, err := isRequiredField(propDefMap, propName)
			if err != nil {
				return false, err
			}

			if isRequired {
				return false, nil
			}

			continue
		}

		isValid, err := validateUserProperty(propName, userValue, propDefMap, logger)
		if err != nil {
			return false, err
		}

		if !isValid {
			return false, nil
		}
	}

	return true, nil
}

// validateUserProperty validates a single property against its schema definition.
func validateUserProperty(
	propName string,
	value interface{},
	propDef map[string]interface{},
	logger *log.Logger,
) (bool, error) {
	propType, exists := propDef["type"]
	if !exists {
		return false, fmt.Errorf("missing required 'type' field")
	}

	typeStr, ok := propType.(string)
	if !ok {
		return false, fmt.Errorf("invalid 'type' field")
	}

	switch typeStr {
	case "string":
		strValue, ok := value.(string)
		if !ok {
			logger.Debug("Expected string but got different type",
				log.String("property", propName),
				log.String("value", fmt.Sprintf("%v", value)))
			return false, nil
		}
		return validateUserStringProperty(propName, strValue, propDef, logger)
	case "number":
		switch value.(type) {
		case float64, int, int64:
			return validateUserNumberProperty(propName, value, propDef, logger)
		default:
			logger.Debug("Expected number but got different type",
				log.String("property", propName),
				log.String("value", fmt.Sprintf("%v", value)))
			return false, nil
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			logger.Debug("Expected boolean but got different type",
				log.String("property", propName),
				log.String("value", fmt.Sprintf("%v", value)))
			return false, nil
		}
		return true, nil
	case "object":
		valueMap, ok := value.(map[string]interface{})
		if !ok {
			logger.Debug("Expected object but got different type",
				log.String("property", propName),
				log.String("value", fmt.Sprintf("%v", value)))
			return false, nil
		}
		return validateUserObjectProperty(valueMap, propDef, logger)
	case "array":
		valueArray, ok := value.([]interface{})
		if !ok {
			logger.Debug("Expected array but got different type",
				log.String("property", propName),
				log.String("value", fmt.Sprintf("%v", value)))
			return false, nil
		}
		return validateUserArrayProperty(propName, valueArray, propDef, logger)
	default:
		logger.Debug("Unsupported property type", log.String("property", propName), log.String("type", typeStr))
		return false, nil
	}
}

// validateUserStringProperty validates string-specific constraints.
func validateUserStringProperty(
	propName, value string,
	propDef map[string]interface{},
	logger *log.Logger,
) (bool, error) {
	isValid, err := validateEnumConstraint(propName, value, propDef, logger)
	if err != nil {
		return false, err
	}
	if !isValid {
		return false, nil
	}

	isValid, err = validateRegexConstraint(propName, value, propDef, logger)
	if err != nil {
		return false, err
	}
	if !isValid {
		return false, nil
	}

	isValid, err = validateUniquenessConstraint(propName, propDef)
	if err != nil {
		return false, err
	}
	if !isValid {
		return false, nil
	}

	return true, nil
}

// validateUserNumberProperty validates number-specific constraints.
func validateUserNumberProperty(
	propName string,
	value interface{},
	propDef map[string]interface{},
	logger *log.Logger,
) (bool, error) {
	isValid, err := validateEnumConstraint(propName, value, propDef, logger)
	if err != nil {
		return false, err
	}
	if !isValid {
		return false, nil
	}

	isValid, err = validateUniquenessConstraint(propName, propDef)
	if err != nil {
		return false, err
	}
	if !isValid {
		return false, nil
	}

	return true, nil
}

// validateUserObjectProperty validates object-specific constraints.
func validateUserObjectProperty(
	value map[string]interface{},
	propDef map[string]interface{},
	logger *log.Logger,
) (bool, error) {
	properties, exists := propDef["properties"]
	if !exists {
		return false, fmt.Errorf("missing 'properties' field")
	}

	propertiesMap, ok := properties.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid 'properties' field")
	}

	for nestedPropName, nestedPropDef := range propertiesMap {
		nestedPropDefMap, ok := nestedPropDef.(map[string]interface{})
		if !ok {
			continue
		}

		nestedValue, exists := value[nestedPropName]
		if !exists {
			isRequired, err := isRequiredField(nestedPropDefMap, nestedPropName)
			if err != nil {
				return false, err
			}

			if isRequired {
				return false, nil
			}

			continue
		}

		if nestedValue != nil {
			isValid, err := validateUserProperty(nestedPropName, nestedValue, nestedPropDefMap, logger)
			if err != nil {
				return false, err
			}
			if !isValid {
				return false, nil
			}
		}
	}

	return true, nil
}

// validateUserArrayProperty validates array-specific constraints.
func validateUserArrayProperty(
	propName string,
	value []interface{},
	propDef map[string]interface{},
	logger *log.Logger,
) (bool, error) {
	isRequired, err := isRequiredField(propDef, propName)
	if err != nil {
		return false, err
	}
	if isRequired && len(value) == 0 {
		logger.Debug("Array property is required but empty", log.String("property", propName))
		return false, nil
	}

	items, exists := propDef["items"]
	if !exists {
		return false, fmt.Errorf("missing 'items' field")
	}

	itemsMap, ok := items.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid 'items' field")
	}

	for _, item := range value {
		isValid, err := validateUserProperty(propName, item, itemsMap, logger)
		if err != nil {
			return false, err
		}
		if !isValid {
			return false, nil
		}
	}

	return true, nil
}

// validateEnumConstraint validates the enum constraint for a property.
func validateEnumConstraint(
	propName string,
	value interface{},
	propDef map[string]interface{},
	logger *log.Logger,
) (bool, error) {
	if enumValue, exists := propDef["enum"]; exists {
		enumArray, ok := enumValue.([]interface{})
		if !ok {
			return false, fmt.Errorf("invalid 'enum' field")
		}

		for _, enumItem := range enumArray {
			if enumItem == value {
				return true, nil
			}
		}
		logger.Debug("Value not in enum", log.String("property", propName), log.String("value", fmt.Sprintf("%v", value)))
		return false, nil // Value not in enum
	}
	return true, nil
}

// validateRegexConstraint validates the regex pattern constraint for a string property.
func validateRegexConstraint(propName, value string, propDef map[string]interface{}, logger *log.Logger) (bool, error) {
	if regexPattern, exists := propDef["pattern"]; exists {
		pattern, ok := regexPattern.(string)
		if !ok {
			return false, fmt.Errorf("invalid 'pattern' field")
		}

		matched, err := regexp.MatchString(pattern, value)
		if err != nil {
			return false, fmt.Errorf("failed to validate regex pattern: %w", err)
		}

		if !matched {
			logger.Debug("Regex pattern mismatch", log.String("property", propName), log.String("value", value))
			return false, nil // Pattern mismatch
		}
	}
	return true, nil
}

// validateUniquenessConstraint validates the uniqueness constraint for a property.
func validateUniquenessConstraint(propName string, propDef map[string]interface{}) (bool, error) {
	if unique, exists := propDef["unique"]; exists {
		if _, ok := unique.(bool); !ok {
			return false, fmt.Errorf("invalid uniqueness constraint for property '%s'", propName)
		}
	}
	return true, nil
}

// isRequiredField checks if the "required" field is present and valid, and returns whether the property is required.
func isRequiredField(propDefMap map[string]interface{}, propName string) (bool, error) {
	if req, exists := propDefMap["required"]; exists {
		if reqBool, ok := req.(bool); ok {
			return reqBool, nil
		}
		return false, fmt.Errorf("'required' field for property '%s' must be a boolean", propName)
	}
	return false, nil
}
