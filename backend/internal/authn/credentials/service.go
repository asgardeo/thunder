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
	"context"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	loggerComponentName = "CredentialsAuthnService"
)

// CredentialsAuthnServiceInterface defines the contract for credentials-based authenticator services.
type CredentialsAuthnServiceInterface interface {
	Authenticate(attributes map[string]interface{}) (*user.User, *serviceerror.ServiceError)
}

// credentialsAuthnService is the default implementation of CredentialsAuthnServiceInterface.
type credentialsAuthnService struct {
	userService user.UserServiceInterface
}

// newCredentialsAuthnService creates a new instance of credentials authenticator service.
func newCredentialsAuthnService(userSvc user.UserServiceInterface) CredentialsAuthnServiceInterface {
	service := &credentialsAuthnService{
		userService: userSvc,
	}
	common.RegisterAuthenticator(service.getMetadata())

	return service
}

// Authenticate authenticates a user using credentials.
func (c *credentialsAuthnService) Authenticate(attributes map[string]interface{}) (
	*user.User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Authenticating user with credentials")

	if len(attributes) == 0 {
		return nil, &ErrorEmptyAttributesOrCredentials
	}

	authRequest := user.AuthenticateUserRequest(attributes)
	authResponse, svcErr := c.userService.AuthenticateUser(context.TODO(), authRequest)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			switch svcErr.Code {
			case user.ErrorUserNotFound.Code:
				return nil, &common.ErrorUserNotFound
			case user.ErrorAuthenticationFailed.Code:
				return nil, &ErrorInvalidCredentials
			default:
				return nil, serviceerror.CustomServiceError(
					ErrorClientErrorFromUserSvcAuthentication, svcErr.ErrorDescription)
			}
		}

		logger.Error("Error occurred while authenticating the user", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, &serviceerror.InternalServerError
	}

	// Fetch the user details
	user, svcErr := c.userService.GetUser(context.TODO(), authResponse.ID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			return nil, serviceerror.CustomServiceError(
				ErrorClientErrorFromUserSvcAuthentication, svcErr.ErrorDescription)
		}

		logger.Error("Error occurred while retrieving the user", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, &serviceerror.InternalServerError
	}

	return user, nil
}

// getMetadata returns the authenticator metadata for credentials authenticator.
func (c *credentialsAuthnService) getMetadata() common.AuthenticatorMeta {
	return common.AuthenticatorMeta{
		Name:    common.AuthenticatorCredentials,
		Factors: []common.AuthenticationFactor{common.FactorKnowledge},
	}
}
