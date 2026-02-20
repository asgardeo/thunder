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

// Package credentials implements an authentication service for credentials-based authentication.
package credentials

import (
	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	loggerComponentName = "CredentialsAuthnService"
)

// CredentialsAuthnServiceInterface defines the contract for credentials-based authenticator services.
type CredentialsAuthnServiceInterface interface {
	Authenticate(identifiers, credentials map[string]interface{}, metadata *authnprovider.AuthnMetadata) (
		*authnprovider.AuthnResult, *serviceerror.ServiceError)
	GetAttributes(token string, requestedAttributes []string, metadata *authnprovider.GetAttributesMetadata) (
		*authnprovider.GetAttributesResult, *serviceerror.ServiceError)
}

// credentialsAuthnService is the default implementation of CredentialsAuthnServiceInterface.
type credentialsAuthnService struct {
	authnProvider authnprovider.AuthnProviderInterface
	logger        *log.Logger
}

// newCredentialsAuthnService creates a new instance of credentials authenticator service.
func newCredentialsAuthnService(authnProvider authnprovider.AuthnProviderInterface) CredentialsAuthnServiceInterface {
	service := &credentialsAuthnService{
		authnProvider: authnProvider,
	}
	common.RegisterAuthenticator(service.getMetadata())
	service.logger = log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	return service
}

func (c *credentialsAuthnService) Authenticate(identifiers, credentials map[string]interface{},
	metadata *authnprovider.AuthnMetadata) (*authnprovider.AuthnResult, *serviceerror.ServiceError) {
	if len(identifiers) == 0 || len(credentials) == 0 {
		return nil, &ErrorEmptyAttributesOrCredentials
	}

	authnResult, err := c.authnProvider.Authenticate(identifiers, credentials, metadata)
	if err != nil {
		switch err.Code {
		case authnprovider.ErrorCodeAuthenticationFailed:
			return nil, &ErrorInvalidCredentials
		case authnprovider.ErrorCodeUserNotFound:
			return nil, &common.ErrorUserNotFound
		default:
			c.logger.Error("Error occurred while authenticating the user", log.String("errorCode", string(err.Code)),
				log.String("errorDescription", err.Description))
			return nil, &serviceerror.InternalServerError
		}
	}
	return authnResult, nil
}

func (c *credentialsAuthnService) GetAttributes(token string, requestedAttributes []string,
	metadata *authnprovider.GetAttributesMetadata) (*authnprovider.GetAttributesResult, *serviceerror.ServiceError) {
	result, err := c.authnProvider.GetAttributes(token, requestedAttributes, metadata)
	if err != nil {
		switch err.Code {
		case authnprovider.ErrorCodeInvalidToken:
			return nil, &ErrorInvalidToken
		default:
			c.logger.Error("Error occurred while getting attributes", log.String("errorCode", string(err.Code)),
				log.String("errorDescription", err.Description))
			return nil, &serviceerror.InternalServerError
		}
	}
	return result, nil
}

// getMetadata returns the authenticator metadata for credentials authenticator.
func (c *credentialsAuthnService) getMetadata() common.AuthenticatorMeta {
	return common.AuthenticatorMeta{
		Name:    common.AuthenticatorCredentials,
		Factors: []common.AuthenticationFactor{common.FactorKnowledge},
	}
}
