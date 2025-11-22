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

// Package idp provides the implementation for identity provider management operations.
package idp

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	filebasedruntime "github.com/asgardeo/thunder/internal/system/file_based_runtime"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// IDPServiceInterface defines the interface for the IdP service.
type IDPServiceInterface interface {
	CreateIdentityProvider(idp *IDPDTO) (*IDPDTO, *serviceerror.ServiceError)
	GetIdentityProviderList() ([]BasicIDPDTO, *serviceerror.ServiceError)
	GetIdentityProvider(idpID string) (*IDPDTO, *serviceerror.ServiceError)
	GetIdentityProviderByName(idpName string) (*IDPDTO, *serviceerror.ServiceError)
	UpdateIdentityProvider(idpID string, idp *IDPDTO) (*IDPDTO, *serviceerror.ServiceError)
	DeleteIdentityProvider(idpID string) *serviceerror.ServiceError
}

// idpService is the default implementation of the IdPServiceInterface.
type idpService struct {
	idpStore idpStoreInterface
}

// newIDPService creates a new instance of IdPService.
func newIDPService(idpStore idpStoreInterface) IDPServiceInterface {
	return &idpService{
		idpStore: idpStore,
	}
}

// CreateIdentityProvider creates a new Identity Provider.
func (is *idpService) CreateIdentityProvider(idp *IDPDTO) (*IDPDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))
	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		return nil, &filebasedruntime.ErrorImmutableResourceCreateOperation
	}

	if svcErr := is.validateIDP(idp); svcErr != nil {
		return nil, svcErr
	}

	// Check if an identity provider with the same name already exists
	existingIDP, err := is.idpStore.GetIdentityProviderByName(idp.Name)
	if err != nil && !errors.Is(err, ErrIDPNotFound) {
		logger.Error("Failed to check existing identity provider by name", log.Error(err),
			log.String("idpName", idp.Name))
		return nil, &serviceerror.InternalServerError
	}
	if existingIDP != nil {
		return nil, &ErrorIDPAlreadyExists
	}

	// Create the IdP in the database.
	idp.ID = utils.GenerateUUID()
	err = is.idpStore.CreateIdentityProvider(*idp)
	if err != nil {
		logger.Error("Failed to create IdP", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return idp, nil
}

// GetIdentityProviderList retrieves the list of all Identity Providers.
func (is *idpService) GetIdentityProviderList() ([]BasicIDPDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))

	idps, err := is.idpStore.GetIdentityProviderList()
	if err != nil {
		logger.Error("Failed to get identity provider list", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return idps, nil
}

// GetIdentityProvider retrieves an identity provider by its ID.
func (is *idpService) GetIdentityProvider(idpID string) (*IDPDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))

	if strings.TrimSpace(idpID) == "" {
		return nil, &ErrorInvalidIDPID
	}

	idp, err := is.idpStore.GetIdentityProvider(idpID)
	if err != nil {
		if errors.Is(err, ErrIDPNotFound) {
			return nil, &ErrorIDPNotFound
		}
		logger.Error("Failed to get identity provider", log.String("idpID", idpID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return idp, nil
}

// GetIdentityProviderByName retrieves an identity provider by its name.
func (is *idpService) GetIdentityProviderByName(idpName string) (*IDPDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))

	if strings.TrimSpace(idpName) == "" {
		return nil, &ErrorInvalidIDPName
	}

	idp, err := is.idpStore.GetIdentityProviderByName(idpName)
	if err != nil {
		if errors.Is(err, ErrIDPNotFound) {
			return nil, &ErrorIDPNotFound
		}
		logger.Error("Failed to get identity provider by name", log.String("idpName", idpName), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return idp, nil
}

// UpdateIdentityProvider updates an existing Identity Provider.
func (is *idpService) UpdateIdentityProvider(idpID string, idp *IDPDTO) (*IDPDTO,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))
	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		return nil, &filebasedruntime.ErrorImmutableResourceUpdateOperation
	}

	if strings.TrimSpace(idpID) == "" {
		return nil, &ErrorInvalidIDPID
	}
	if svcErr := is.validateIDP(idp); svcErr != nil {
		return nil, svcErr
	}

	// Check if the identity provider exists
	existingIDP, err := is.idpStore.GetIdentityProvider(idpID)
	if err != nil {
		if errors.Is(err, ErrIDPNotFound) {
			return nil, &ErrorIDPNotFound
		}
		logger.Error("Failed to get identity provider for update", log.String("idpID", idpID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if existingIDP == nil {
		return nil, &ErrorIDPNotFound
	}

	// If the name is being updated, check whether another IdP with the same name exists
	if existingIDP.Name != idp.Name {
		existingIDPByName, err := is.idpStore.GetIdentityProviderByName(idp.Name)
		if err != nil && !errors.Is(err, ErrIDPNotFound) {
			logger.Error("Failed to check existing identity provider by name", log.Error(err),
				log.String("idpName", idp.Name))
			return nil, &serviceerror.InternalServerError
		}
		if existingIDPByName != nil {
			return nil, &ErrorIDPAlreadyExists
		}
	}

	idp.ID = idpID
	err = is.idpStore.UpdateIdentityProvider(idp)
	if err != nil {
		logger.Error("Failed to update identity provider", log.Error(err), log.String("idpID", idpID))
		return nil, &serviceerror.InternalServerError
	}

	return idp, nil
}

// DeleteIdentityProvider deletes an Identity Provider by its ID.
func (is *idpService) DeleteIdentityProvider(idpID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))
	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		return &filebasedruntime.ErrorImmutableResourceDeleteOperation
	}

	if strings.TrimSpace(idpID) == "" {
		return &ErrorInvalidIDPID
	}

	// Check if the identity provider exists
	_, err := is.idpStore.GetIdentityProvider(idpID)
	if err != nil {
		if errors.Is(err, ErrIDPNotFound) {
			return nil
		}
		logger.Error("Failed to get identity provider for deletion", log.Error(err), log.String("idpID", idpID))
		return &serviceerror.InternalServerError
	}

	err = is.idpStore.DeleteIdentityProvider(idpID)
	if err != nil {
		logger.Error("Failed to delete identity provider", log.Error(err), log.String("idpID", idpID))
		return &serviceerror.InternalServerError
	}

	return nil
}

// validateIDP validates the identity provider details.
func (is *idpService) validateIDP(idp *IDPDTO) *serviceerror.ServiceError {
	if idp == nil {
		return &ErrorIDPNil
	}
	if strings.TrimSpace(idp.Name) == "" {
		return &ErrorInvalidIDPName
	}

	// Validate identity provider type
	if strings.TrimSpace(string(idp.Type)) == "" {
		return &ErrorInvalidIDPType
	}
	isValidType := slices.Contains(supportedIDPTypes, idp.Type)
	if !isValidType {
		return &ErrorInvalidIDPType
	}

	return validateIDPProperties(idp.Properties)
}

// validateIDPProperties validates the identity provider properties.
func validateIDPProperties(properties []cmodels.Property) *serviceerror.ServiceError {
	if len(properties) == 0 {
		return nil
	}
	for _, property := range properties {
		if strings.TrimSpace(property.GetName()) == "" {
			return serviceerror.CustomServiceError(ErrorInvalidIDPProperty,
				"property names cannot be empty")
		}
		propertyValue, err := property.GetValue()
		if err != nil {
			return serviceerror.CustomServiceError(ErrorInvalidIDPProperty,
				fmt.Sprintf("failed to get value for property '%s': %v", property.GetName(), err))
		}
		if strings.TrimSpace(propertyValue) == "" {
			return serviceerror.CustomServiceError(ErrorInvalidIDPProperty,
				fmt.Sprintf("property value cannot be empty for property '%s'", property.GetName()))
		}
		if !slices.Contains(supportedIDPProperties, property.GetName()) {
			return serviceerror.CustomServiceError(ErrorUnsupportedIDPProperty,
				fmt.Sprintf("property '%s' is not supported", property.GetName()))
		}
	}
	return nil
}
