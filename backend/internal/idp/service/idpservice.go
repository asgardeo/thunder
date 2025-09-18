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

// Package service provides the implementation for IdP management operations.
package service

import (
	"errors"
	"strings"

	"github.com/asgardeo/thunder/internal/idp/constants"
	"github.com/asgardeo/thunder/internal/idp/model"
	"github.com/asgardeo/thunder/internal/idp/store"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// IDPServiceInterface defines the interface for the IdP service.
type IDPServiceInterface interface {
	CreateIdentityProvider(idp *model.IdpDTO) (*model.IdpDTO, *serviceerror.ServiceError)
	GetIdentityProviderList() ([]model.BasicIdpDTO, *serviceerror.ServiceError)
	GetIdentityProvider(idpID string) (*model.IdpDTO, *serviceerror.ServiceError)
	GetIdentityProviderByName(idpName string) (*model.IdpDTO, *serviceerror.ServiceError)
	UpdateIdentityProvider(idpID string, idp *model.IdpDTO) (*model.IdpDTO, *serviceerror.ServiceError)
	DeleteIdentityProvider(idpID string) *serviceerror.ServiceError
}

// IDPService is the default implementation of the IdPServiceInterface.
type IDPService struct {
	IDPStore store.IDPStoreInterface
}

// NewIDPService creates a new instance of IdPService.
func NewIDPService() IDPServiceInterface {
	return &IDPService{
		IDPStore: store.NewIDPStore(),
	}
}

// CreateIdentityProvider creates a new Identity Provider.
func (is *IDPService) CreateIdentityProvider(idp *model.IdpDTO) (*model.IdpDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))

	if idp == nil {
		return nil, &constants.ErrorIDPNil
	}

	if strings.TrimSpace(idp.Name) == "" {
		return nil, &constants.ErrorInvalidIDPName
	}

	// Check if an identity provider with the same name already exists
	existingIDP, err := is.IDPStore.GetIdentityProviderByName(idp.Name)
	if err != nil && !errors.Is(err, constants.ErrIDPNotFound) {
		logger.Error("Failed to check existing identity provider by name", log.Error(err),
			log.String("idpName", idp.Name))
		return nil, &constants.ErrorInternalServerError
	}
	if existingIDP != nil {
		return nil, &constants.ErrorIDPAlreadyExists
	}

	// Validate properties
	if svcErr := validateIDPProperties(idp.Properties); svcErr != nil {
		return nil, svcErr
	}

	idp.ID = utils.GenerateUUID()

	// Create the IdP in the database.
	err = is.IDPStore.CreateIdentityProvider(*idp)
	if err != nil {
		logger.Error("Failed to create IdP", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return idp, nil
}

// GetIdentityProviderList retrieves the list of all Identity Providers.
func (is *IDPService) GetIdentityProviderList() ([]model.BasicIdpDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))

	idps, err := is.IDPStore.GetIdentityProviderList()
	if err != nil {
		logger.Error("Failed to get identity provider list", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return idps, nil
}

// GetIdentityProvider retrieves an identity provider by its ID.
func (is *IDPService) GetIdentityProvider(idpID string) (*model.IdpDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))

	if strings.TrimSpace(idpID) == "" {
		return nil, &constants.ErrorInvalidIDPID
	}

	idp, err := is.IDPStore.GetIdentityProvider(idpID)
	if err != nil {
		if errors.Is(err, constants.ErrIDPNotFound) {
			return nil, &constants.ErrorIDPNotFound
		}
		logger.Error("Failed to get identity provider", log.String("idpID", idpID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return idp, nil
}

// GetIdentityProviderByName retrieves an identity provider by its name.
func (is *IDPService) GetIdentityProviderByName(idpName string) (*model.IdpDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))

	if strings.TrimSpace(idpName) == "" {
		return nil, &constants.ErrorInvalidIDPName
	}

	idp, err := is.IDPStore.GetIdentityProviderByName(idpName)
	if err != nil {
		if errors.Is(err, constants.ErrIDPNotFound) {
			return nil, &constants.ErrorIDPNotFound
		}
		logger.Error("Failed to get identity provider by name", log.String("idpName", idpName), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return idp, nil
}

// UpdateIdentityProvider updates an existing Identity Provider.
func (is *IDPService) UpdateIdentityProvider(idpID string, idp *model.IdpDTO) (*model.IdpDTO,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))

	if strings.TrimSpace(idpID) == "" {
		return nil, &constants.ErrorInvalidIDPID
	}

	if idp == nil {
		return nil, &constants.ErrorIDPNil
	}

	if strings.TrimSpace(idp.Name) == "" {
		return nil, &constants.ErrorInvalidIDPName
	}

	// Check if the identity provider exists
	existingIDP, err := is.IDPStore.GetIdentityProvider(idpID)
	if err != nil {
		if errors.Is(err, constants.ErrIDPNotFound) {
			return nil, &constants.ErrorIDPNotFound
		}
		logger.Error("Failed to get identity provider for update", log.String("idpID", idpID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}
	if existingIDP == nil {
		return nil, &constants.ErrorIDPNotFound
	}

	// If the name is being updated, check whether another IdP with the same name exists
	if existingIDP.Name != idp.Name {
		existingIDPByName, err := is.IDPStore.GetIdentityProviderByName(idp.Name)
		if err != nil && !errors.Is(err, constants.ErrIDPNotFound) {
			logger.Error("Failed to check existing identity provider by name", log.Error(err),
				log.String("idpName", idp.Name))
			return nil, &constants.ErrorInternalServerError
		}
		if existingIDPByName != nil {
			return nil, &constants.ErrorIDPAlreadyExists
		}
	}

	// Validate properties
	if svcErr := validateIDPProperties(idp.Properties); svcErr != nil {
		return nil, svcErr
	}

	idp.ID = idpID

	err = is.IDPStore.UpdateIdentityProvider(idp)
	if err != nil {
		logger.Error("Failed to update identity provider", log.Error(err), log.String("idpID", idpID))
		return nil, &constants.ErrorInternalServerError
	}

	return idp, nil
}

// DeleteIdentityProvider deletes an Identity Provider by its ID.
func (is *IDPService) DeleteIdentityProvider(idpID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService"))

	if strings.TrimSpace(idpID) == "" {
		return &constants.ErrorInvalidIDPID
	}

	// Check if the identity provider exists
	_, err := is.IDPStore.GetIdentityProvider(idpID)
	if err != nil {
		if errors.Is(err, constants.ErrIDPNotFound) {
			return nil
		}
		logger.Error("Failed to get identity provider for deletion", log.Error(err), log.String("idpID", idpID))
		return &constants.ErrorInternalServerError
	}

	err = is.IDPStore.DeleteIdentityProvider(idpID)
	if err != nil {
		logger.Error("Failed to delete identity provider", log.Error(err), log.String("idpID", idpID))
		return &constants.ErrorInternalServerError
	}

	return nil
}

// validateIDPProperties validates the identity provider properties.
func validateIDPProperties(properties []model.IdpProperty) *serviceerror.ServiceError {
	for _, property := range properties {
		if strings.TrimSpace(property.Name) == "" {
			return &constants.ErrorInvalidIDPProperties
		}
	}
	return nil
}
