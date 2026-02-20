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

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
)

type defaultAuthnProvider struct {
	userSvc user.UserServiceInterface
}

// newDefaultAuthnProvider creates a new internal user authn provider.
func newDefaultAuthnProvider(userSvc user.UserServiceInterface) AuthnProviderInterface {
	return &defaultAuthnProvider{
		userSvc: userSvc,
	}
}

// Authenticate authenticates the user using the internal user service.
func (p *defaultAuthnProvider) Authenticate(
	identifiers, credentials map[string]interface{},
	metadata *AuthnMetadata,
) (*AuthnResult, *AuthnProviderError) {
	request := make(user.AuthenticateUserRequest)
	for k, v := range identifiers {
		request[k] = v
	}
	for k, v := range credentials {
		request[k] = v
	}

	authResponse, authErr := p.userSvc.AuthenticateUser(context.Background(), request)
	if authErr != nil {
		if authErr.Type == serviceerror.ClientErrorType {
			if authErr.Code == user.ErrorUserNotFound.Code {
				return nil, NewError(ErrorCodeUserNotFound, authErr.Error, authErr.ErrorDescription)
			}
			return nil, NewError(ErrorCodeAuthenticationFailed, authErr.Error, authErr.ErrorDescription)
		}
		return nil, NewError(ErrorCodeSystemError, authErr.Error, authErr.ErrorDescription)
	}

	userResult, getUserErr := p.userSvc.GetUser(context.Background(), authResponse.ID)
	if getUserErr != nil {
		if getUserErr.Code == user.ErrorUserNotFound.Code {
			return nil, NewError(ErrorCodeUserNotFound, getUserErr.Error, getUserErr.ErrorDescription)
		}
		return nil, NewError(ErrorCodeSystemError, getUserErr.Error, getUserErr.ErrorDescription)
	}

	var attributes map[string]interface{}
	if len(userResult.Attributes) > 0 {
		if err := json.Unmarshal(userResult.Attributes, &attributes); err != nil {
			return nil, NewError(ErrorCodeSystemError, "Failed to get allowed attributes", err.Error())
		}
	}

	availableAttributes := make([]AvailableAttribute, 0)
	for k := range attributes {
		availableAttributes = append(availableAttributes, AvailableAttribute{
			Name:        k,
			DisplayName: k,
			Verified:    false,
		})
	}

	return &AuthnResult{
		UserID:              authResponse.ID,
		Token:               authResponse.ID,
		UserType:            userResult.Type,
		OrganizationUnitID:  userResult.OrganizationUnit,
		AvailableAttributes: availableAttributes,
	}, nil
}

// GetAttributes retrieves the user attributes using the internal user service.
func (p *defaultAuthnProvider) GetAttributes(
	token string,
	requestedAttributes []string,
	metadata *GetAttributesMetadata,
) (*GetAttributesResult, *AuthnProviderError) {
	userID := token

	userResult, authErr := p.userSvc.GetUser(context.Background(), userID)
	if authErr != nil {
		if authErr.Type == serviceerror.ClientErrorType {
			return nil, NewError(ErrorCodeInvalidToken, authErr.Error, authErr.ErrorDescription)
		}
		return nil, NewError(ErrorCodeSystemError, authErr.Error, authErr.ErrorDescription)
	}

	var attributes json.RawMessage
	if len(requestedAttributes) > 0 {
		var allAttributes map[string]interface{}
		if len(userResult.Attributes) > 0 {
			if err := json.Unmarshal(userResult.Attributes, &allAttributes); err != nil {
				return nil, NewError(ErrorCodeSystemError, "System Error", "Failed to unmarshal user attributes")
			}
		}

		filteredAttributes := make(map[string]interface{})
		for _, attr := range requestedAttributes {
			if val, ok := allAttributes[attr]; ok {
				filteredAttributes[attr] = val
			}
		}

		var err error
		attributes, err = json.Marshal(filteredAttributes)
		if err != nil {
			return nil, NewError(ErrorCodeSystemError, "System Error", "Failed to marshal filtered user attributes")
		}
	} else {
		attributes = userResult.Attributes
	}

	return &GetAttributesResult{
		UserID:             userResult.ID,
		UserType:           userResult.Type,
		OrganizationUnitID: userResult.OrganizationUnit,
		Attributes:         attributes,
	}, nil
}
