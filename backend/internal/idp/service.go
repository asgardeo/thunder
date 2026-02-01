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
	"strings"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
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
	logger   *log.Logger
}

// newIDPService creates a new instance of IdPService.
func newIDPService(idpStore idpStoreInterface) IDPServiceInterface {
	return &idpService{
		idpStore: idpStore,
		logger:   log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPService")),
	}
}

// CreateIdentityProvider creates a new Identity Provider.
func (is *idpService) CreateIdentityProvider(idp *IDPDTO) (*IDPDTO, *serviceerror.ServiceError) {
	logger := is.logger
	if err := declarativeresource.CheckDeclarativeCreate(); err != nil {
		return nil, err
	}

	if svcErr := validateIDP(idp, logger); svcErr != nil {
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
	logger := is.logger
	idps, err := is.idpStore.GetIdentityProviderList()
	if err != nil {
		logger.Error("Failed to get identity provider list", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return idps, nil
}

// GetIdentityProvider retrieves an identity provider by its ID.
func (is *idpService) GetIdentityProvider(idpID string) (*IDPDTO, *serviceerror.ServiceError) {
	logger := is.logger
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
	logger := is.logger
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
	logger := is.logger
	if err := declarativeresource.CheckDeclarativeUpdate(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(idpID) == "" {
		return nil, &ErrorInvalidIDPID
	}
	if svcErr := validateIDP(idp, logger); svcErr != nil {
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

// DeleteIdentityProvider deletes an identity provider.
func (is *idpService) DeleteIdentityProvider(idpID string) *serviceerror.ServiceError {
	logger := is.logger
	if err := declarativeresource.CheckDeclarativeDelete(); err != nil {
		return err
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
