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

// Package authnprovider provides authentication provider implementations.
package authnprovider

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/asgardeo/thunder/internal/entity"
)

type defaultAuthnProvider struct {
	entitySvc entity.EntityServiceInterface
}

// newDefaultAuthnProvider creates a new default authn provider.
func newDefaultAuthnProvider(entitySvc entity.EntityServiceInterface) AuthnProviderInterface {
	return &defaultAuthnProvider{
		entitySvc: entitySvc,
	}
}

// Authenticate authenticates any entity type (user, app, agent) using the entity service.
func (p *defaultAuthnProvider) Authenticate(
	ctx context.Context,
	identifiers, credentials map[string]interface{},
	metadata *AuthnMetadata,
) (*AuthnResult, *AuthnProviderError) {
	authResult, err := p.entitySvc.AuthenticateEntity(ctx, identifiers, credentials)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, NewError(ErrorCodeUserNotFound, "Entity not found", err.Error())
		}
		if errors.Is(err, entity.ErrAuthenticationFailed) {
			return nil, NewError(ErrorCodeAuthenticationFailed, "Authentication failed", err.Error())
		}
		return nil, NewError(ErrorCodeSystemError, "System error", err.Error())
	}

	// Fetch entity to build available attributes for the result.
	entityResult, getErr := p.entitySvc.GetEntity(ctx, authResult.EntityID)
	if getErr != nil {
		if errors.Is(getErr, entity.ErrEntityNotFound) {
			return nil, NewError(ErrorCodeUserNotFound, "Entity not found", getErr.Error())
		}
		return nil, NewError(ErrorCodeSystemError, "System error", getErr.Error())
	}

	var attributes map[string]interface{}
	if len(entityResult.Attributes) > 0 {
		if err := json.Unmarshal(entityResult.Attributes, &attributes); err != nil {
			return nil, NewError(ErrorCodeSystemError, "Failed to get allowed attributes", err.Error())
		}
	}

	availableAttributes := &AvailableAttributes{
		Attributes:    make(map[string]*AttributeMetadataResponse),
		Verifications: make(map[string]*VerificationResponse),
	}
	for k := range attributes {
		availableAttributes.Attributes[k] = &AttributeMetadataResponse{
			AssuranceMetadataResponse: &AssuranceMetadataResponse{
				IsVerified:     false,
				VerificationID: "",
			},
		}
	}

	return &AuthnResult{
		EntityID:       authResult.EntityID,
		EntityCategory: authResult.EntityCategory.String(),
		EntityType:     authResult.EntityType,
		// Backward-compatible aliases.
		UserID:              authResult.EntityID,
		Token:               authResult.EntityID,
		UserType:            authResult.EntityType,
		OUID:                authResult.OrganizationUnitID,
		AvailableAttributes: availableAttributes,
	}, nil
}

// GetAttributes retrieves the entity attributes using the entity service.
func (p *defaultAuthnProvider) GetAttributes(
	ctx context.Context,
	token string,
	requestedAttributes *RequestedAttributes,
	metadata *GetAttributesMetadata,
) (*GetAttributesResult, *AuthnProviderError) {
	entityID := token

	entityResult, err := p.entitySvc.GetEntity(ctx, entityID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, NewError(ErrorCodeInvalidToken, "Entity not found", err.Error())
		}
		return nil, NewError(ErrorCodeSystemError, "System error", err.Error())
	}

	var allAttributes map[string]interface{}
	if len(entityResult.Attributes) > 0 {
		if err := json.Unmarshal(entityResult.Attributes, &allAttributes); err != nil {
			return nil, NewError(ErrorCodeSystemError, "System Error", "Failed to unmarshal entity attributes")
		}
	}

	attributesResponse := &AttributesResponse{
		Attributes:    make(map[string]*AttributeResponse),
		Verifications: make(map[string]*VerificationResponse),
	}

	if requestedAttributes != nil && len(requestedAttributes.Attributes) > 0 {
		for attrName := range requestedAttributes.Attributes {
			if val, ok := allAttributes[attrName]; ok {
				attributesResponse.Attributes[attrName] = &AttributeResponse{
					Value: val,
					AssuranceMetadataResponse: &AssuranceMetadataResponse{
						IsVerified:     false,
						VerificationID: "",
					},
				}
			}
		}
	} else {
		for attrName, val := range allAttributes {
			attributesResponse.Attributes[attrName] = &AttributeResponse{
				Value: val,
				AssuranceMetadataResponse: &AssuranceMetadataResponse{
					IsVerified:     false,
					VerificationID: "",
				},
			}
		}
	}

	return &GetAttributesResult{
		EntityID:       entityResult.ID,
		EntityCategory: entityResult.Category.String(),
		EntityType:     entityResult.Type,
		// TODO: Remove after refacoring usages. Kept for backward compatibility — aliases for entity fields.
		UserID:             entityResult.ID,
		UserType:           entityResult.Type,
		OUID:               entityResult.OrganizationUnitID,
		AttributesResponse: attributesResponse,
	}, nil
}
