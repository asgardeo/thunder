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

package credentials

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for credentials authentication.
var (
	// ErrorEmptyAttributesOrCredentials is the error when the provided user attributes or credentials are empty.
	ErrorEmptyAttributesOrCredentials = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTH-CRED-1001",
		Error:            "Empty attributes or credentials",
		ErrorDescription: "The user attributes or credentials cannot be empty",
	}
	// ErrorInvalidCredentials is the error when the provided credentials are invalid.
	ErrorInvalidCredentials = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTH-CRED-1002",
		Error:            "Invalid credentials",
		ErrorDescription: "The provided credentials are invalid",
	}
	// ErrorClientErrorFromUserSvcAuthentication is the error when there is a client error from
	// the user service during authentication.
	ErrorClientErrorFromUserSvcAuthentication = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTH-CRED-1003",
		Error:            "authentication failed",
		ErrorDescription: "An error occurred while authenticating the user",
	}
	// ErrorInvalidToken is the error when the provided token is invalid.
	ErrorInvalidToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTH-CRED-1004",
		Error:            "Invalid token",
		ErrorDescription: "The provided token is invalid",
	}
)
