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

package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	resourceTypeUser = "user"
	paramTypeUser    = "User"
)

// userExporter implements declarativeresource.ResourceExporter for users.
type userExporter struct {
	service UserServiceInterface
}

// newUserExporter creates a new user exporter.
func newUserExporter(service UserServiceInterface) *userExporter {
	return &userExporter{service: service}
}

// GetResourceType returns the resource type for users.
func (e *userExporter) GetResourceType() string {
	return resourceTypeUser
}

// GetParameterizerType returns the parameterizer type for users.
func (e *userExporter) GetParameterizerType() string {
	return paramTypeUser
}

// GetAllResourceIDs retrieves all user IDs from the database store.
// In composite mode, this excludes declarative (YAML-based) users.
func (e *userExporter) GetAllResourceIDs(ctx context.Context) ([]string, *serviceerror.ServiceError) {
	offset := 0
	limit := serverconst.MaxPageSize
	ids := []string{}

	for {
		users, err := e.service.GetUserList(ctx, limit, offset, nil)
		if err != nil {
			return nil, err
		}

		for _, user := range users.Users {
			isDeclarative, svcErr := e.service.IsUserDeclarative(ctx, user.ID)
			if svcErr != nil {
				return nil, svcErr
			}
			if !isDeclarative {
				ids = append(ids, user.ID)
			}
		}

		offset += len(users.Users)

		// Continue fetching while we get results; stop only on empty page
		if len(users.Users) == 0 {
			break
		}
	}

	return ids, nil
}

// GetResourceByID retrieves a user by its ID.
func (e *userExporter) GetResourceByID(
	ctx context.Context, id string) (interface{}, string, *serviceerror.ServiceError) {
	user, err := e.service.GetUser(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// Extract username from attributes for identification
	var username string
	var attrs map[string]interface{}
	if len(user.Attributes) > 0 {
		if jsonErr := json.Unmarshal(user.Attributes, &attrs); jsonErr == nil {
			if un, ok := attrs["username"].(string); ok {
				username = un
			}
		}
	}

	// Convert User.Attributes (json.RawMessage) to map for export
	var attributesMap map[string]interface{}
	if len(user.Attributes) > 0 {
		if jsonErr := json.Unmarshal(user.Attributes, &attributesMap); jsonErr != nil {
			attributesMap = make(map[string]interface{})
		}
	} else {
		attributesMap = make(map[string]interface{})
	}

	// Create export structure with credentials as placeholders
	// The parameterizer will replace actual credential values with template variables
	exportUser := &userDeclarativeResource{
		ID:               user.ID,
		Type:             user.Type,
		OrganizationUnit: user.OrganizationUnit,
		Attributes:       attributesMap,
		Credentials:      make(map[string]interface{}), // Empty credentials - will be filled with placeholders
	}

	return exportUser, username, nil
}

// ValidateResource validates a user resource.
func (e *userExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *declarativeresource.ExportError) {
	user, ok := resource.(*userDeclarativeResource)
	if !ok {
		return "", declarativeresource.CreateTypeError(resourceTypeUser, id)
	}

	// Extract username for validation
	var username string
	if un, ok := user.Attributes["username"].(string); ok {
		username = un
	}

	if username == "" {
		logger.Warn("USER_VALIDATION_ERROR: Missing username",
			log.String("userID", id))
		return "", &declarativeresource.ExportError{
			ResourceType: resourceTypeUser,
			ResourceID:   id,
			Code:         "USER_VALIDATION_ERROR",
			Error:        fmt.Sprintf("User '%s' validation failed: missing username", id),
		}
	}

	return username, nil
}

// GetResourceRules returns the parameterization rules for users.
func (e *userExporter) GetResourceRules() *declarativeresource.ResourceRules {
	return &declarativeresource.ResourceRules{
		Variables:             []string{},
		ArrayVariables:        []string{},
		DynamicPropertyFields: []string{"Credentials"},
	}
}

// loadDeclarativeResources loads immutable user resources from files.
// The dbStore parameter is optional (can be nil) and is used for duplicate checking in composite mode.
func loadDeclarativeResources(fileStore *userFileBasedStore, dbStore userStoreInterface) error {
	resourceConfig := declarativeresource.ResourceConfig{
		ResourceType:  "User",
		DirectoryName: "users",
		Parser:        parseToUserWrapper,
		Validator: func(data interface{}) error {
			return validateUserWrapper(data, fileStore, dbStore)
		},
		IDExtractor: func(data interface{}) string {
			// Use safe type assertion to prevent panic
			if v, ok := data.(*userResource); ok {
				return v.User.ID
			}
			// Log error and return empty string if type assertion fails
			log.GetLogger().Error("IDExtractor: type assertion failed for userResource")
			return ""
		},
	}

	loader := declarativeresource.NewResourceLoader(resourceConfig, fileStore)
	if err := loader.LoadResources(); err != nil {
		return fmt.Errorf("failed to load user resources: %w", err)
	}

	return nil
}

// parseToUserWrapper wraps parseToUser to match the expected signature.
func parseToUserWrapper(data []byte) (interface{}, error) {
	return parseToUser(data)
}

type userDeclarativeResource struct {
	ID               string                 `yaml:"id"`
	Type             string                 `yaml:"type"`
	OrganizationUnit string                 `yaml:"ou_id"`
	Attributes       map[string]interface{} `yaml:"attributes"`
	Credentials      map[string]interface{} `yaml:"credentials,omitempty"` // Flexible format for YAML
}

// parseToUser parses YAML data to userResource.
func parseToUser(data []byte) (*userResource, error) {
	var userRes userDeclarativeResource
	if err := yaml.Unmarshal(data, &userRes); err != nil {
		return nil, err
	}

	// Convert attributes map to JSON
	attributesJSON, err := json.Marshal(userRes.Attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attributes: %w", err)
	}

	user := User{
		ID:               userRes.ID,
		Type:             userRes.Type,
		OrganizationUnit: userRes.OrganizationUnit,
		Attributes:       json.RawMessage(attributesJSON),
	}

	// Parse and hash credentials
	credentials, err := parseCredentials(userRes.Credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	resource := &userResource{
		User:        user,
		Credentials: credentials,
	}

	return resource, nil
}

// parseCredentials parses credentials from YAML and hashes plain text values.
// Supports two formats:
// 1. Simple format: credentials: { password: "plaintext" }
// 2. Full format: credentials: { password: [{ storageType: "hash", value: "hashed", ... }] }
func parseCredentials(credentialsMap map[string]interface{}) (Credentials, error) {
	if len(credentialsMap) == 0 {
		return make(Credentials), nil
	}

	credentials := make(Credentials)
	hashService := hash.Initialize()

	for credType, credValue := range credentialsMap {
		credentialType := CredentialType(credType)

		switch v := credValue.(type) {
		case string:
			// Simple format: plain text password that needs hashing
			if v == "" {
				continue
			}

			if credentialType.IsSystemManaged() {
				credentials[credentialType] = []Credential{{Value: v}}
				continue
			}

			hashedCred, err := hashService.Generate([]byte(v))
			if err != nil {
				return nil, fmt.Errorf("failed to hash credential %s: %w", credType, err)
			}

			credential := Credential{
				StorageType: "hash",
				StorageAlgo: hashedCred.Algorithm,
				StorageAlgoParams: hash.CredParameters{
					Iterations: hashedCred.Parameters.Iterations,
					KeySize:    hashedCred.Parameters.KeySize,
					Salt:       hashedCred.Parameters.Salt,
				},
				Value: hashedCred.Hash,
			}

			credentials[credentialType] = []Credential{credential}

		case []interface{}:
			// Full format: array of credential objects
			var credList []Credential
			for _, item := range v {
				credMap, ok := item.(map[string]interface{})
				if !ok {
					// Try map[interface{}]interface{} (YAML unmarshaling)
					if credMapAny, ok := item.(map[interface{}]interface{}); ok {
						credMap = make(map[string]interface{})
						for k, val := range credMapAny {
							if keyStr, ok := k.(string); ok {
								credMap[keyStr] = val
							}
						}
					} else {
						return nil, fmt.Errorf("invalid credential format for %s", credType)
					}
				}

				cred, err := parseCredentialObject(credMap, hashService, credentialType)
				if err != nil {
					return nil, fmt.Errorf("failed to parse credential %s: %w", credType, err)
				}
				credList = append(credList, cred)
			}
			credentials[credentialType] = credList

		default:
			return nil, fmt.Errorf("unsupported credential format for %s", credType)
		}
	}

	return credentials, nil
}

// parseCredentialObject parses a single credential object.
// If the value is plain text and no hash info is provided, it will hash it.
func parseCredentialObject(
	credMap map[string]interface{},
	hashService hash.HashServiceInterface,
	credentialType CredentialType,
) (Credential, error) {
	value, hasValue := credMap["value"].(string)
	if !hasValue || value == "" {
		return Credential{}, fmt.Errorf("credential value is required")
	}

	storageType, _ := credMap["storageType"].(string)
	storageAlgo, _ := credMap["storageAlgo"].(string)
	systemManaged, _ := credMap["systemManaged"].(bool)

	if credentialType.IsSystemManaged() || systemManaged || storageType == "system" {
		if storageType == "" {
			storageType = "system"
		}
		return Credential{
			StorageType: storageType,
			StorageAlgo: hash.CredAlgorithm(storageAlgo),
			Value:       value,
		}, nil
	}

	// If storage type is not specified or is not "hash", treat as plain text and hash it
	if storageType == "" || storageType != "hash" {
		hashedCred, err := hashService.Generate([]byte(value))
		if err != nil {
			return Credential{}, fmt.Errorf("failed to hash credential: %w", err)
		}

		return Credential{
			StorageType: "hash",
			StorageAlgo: hashedCred.Algorithm,
			StorageAlgoParams: hash.CredParameters{
				Iterations: hashedCred.Parameters.Iterations,
				KeySize:    hashedCred.Parameters.KeySize,
				Salt:       hashedCred.Parameters.Salt,
			},
			Value: hashedCred.Hash,
		}, nil
	}

	// Parse pre-hashed credential
	paramsMap, _ := credMap["storageAlgoParams"].(map[string]interface{})
	if paramsMap == nil {
		// Try map[interface{}]interface{} format
		if paramsMapAny, ok := credMap["storageAlgoParams"].(map[interface{}]interface{}); ok {
			paramsMap = make(map[string]interface{})
			for k, v := range paramsMapAny {
				if keyStr, ok := k.(string); ok {
					paramsMap[keyStr] = v
				}
			}
		}
	}

	iterations, _ := paramsMap["iterations"].(int)
	keySize, _ := paramsMap["keySize"].(int)
	salt, _ := paramsMap["salt"].(string)

	return Credential{
		StorageType: storageType,
		StorageAlgo: hash.CredAlgorithm(storageAlgo),
		StorageAlgoParams: hash.CredParameters{
			Iterations: iterations,
			KeySize:    keySize,
			Salt:       salt,
		},
		Value: value,
	}, nil
}

// validateUserWrapper validates user declarative resources and checks for duplicates.
func validateUserWrapper(data interface{}, fileStore *userFileBasedStore, dbStore userStoreInterface) error {
	resource, ok := data.(*userResource)
	if !ok {
		return fmt.Errorf("invalid type: expected *userResource")
	}

	user := resource.User

	if user.ID == "" {
		return fmt.Errorf("user ID is required")
	}
	if user.Type == "" {
		return fmt.Errorf("user type is required")
	}
	if user.OrganizationUnit == "" {
		return fmt.Errorf("organization unit ID is required")
	}

	// Validate attributes exist
	if len(user.Attributes) == 0 {
		return fmt.Errorf("user attributes are required")
	}

	// Extract and validate username
	var attrs map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
		return fmt.Errorf("failed to parse user attributes: %w", err)
	}

	username, hasUsername := attrs["username"]
	if !hasUsername || username == "" {
		return fmt.Errorf("username is required in user attributes")
	}

	// Check for duplicates in file store
	if fileStore != nil {
		if existingData, err := fileStore.GenericFileBasedStore.Get(user.ID); err == nil && existingData != nil {
			return fmt.Errorf("duplicate user ID '%s': user already exists in declarative resources", user.ID)
		}
	}

	// Check for duplicates in database store
	if dbStore != nil {
		_, err := dbStore.GetUser(context.Background(), user.ID)
		if err == nil {
			return fmt.Errorf("duplicate user ID '%s': user already exists in the database store", user.ID)
		}
		if !errors.Is(err, ErrUserNotFound) {
			// Fail loudly on DB errors during duplicate check
			return fmt.Errorf("checking user existence for '%s': %w", user.ID, err)
		}
	}

	return nil
}
