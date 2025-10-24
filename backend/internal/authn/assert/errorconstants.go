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

package assert

import (
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client errors for authentication assertion operations.
var (
	// ErrorNoAuthenticators is the error returned when no authenticators are provided.
	ErrorNoAuthenticators = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AST-1001",
		Error:            "No authenticators",
		ErrorDescription: "Cannot generate assertion without authenticators",
	}
	// ErrorInvalidAuthenticator is the error returned when authenticator name is invalid.
	ErrorInvalidAuthenticator = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AST-1002",
		Error:            "Invalid authenticator",
		ErrorDescription: "Authenticator name cannot be empty",
	}
)

// Server errors for authentication assertion operations.
var (
	// ErrorInternalServerError is the error returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "AST-5000",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)
