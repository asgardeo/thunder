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

package dcr

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// DCR standard service error constants
var (
	// ErrorInvalidRequestFormat is used for nil request validation
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "invalid_client_metadata",
		Error:            "Invalid request format",
		ErrorDescription: "The request format is invalid",
	}

	// ErrorInvalidRedirectURI is the standard error for redirect URI issues
	ErrorInvalidRedirectURI = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "invalid_redirect_uri",
		Error:            "Invalid redirect URI",
		ErrorDescription: "One or more redirect URIs are invalid",
	}

	// ErrorInvalidClientMetadata is the standard error for client metadata issues
	ErrorInvalidClientMetadata = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "invalid_client_metadata",
		Error:            "Invalid client metadata",
		ErrorDescription: "One or more client metadata values are invalid",
	}

	// ErrorJWKSConfigurationConflict is the error returned when both jwks and jwks_uri are provided
	ErrorJWKSConfigurationConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "invalid_client_metadata",
		Error:            "JWKS configuration conflict",
		ErrorDescription: "Cannot specify both 'jwks' and 'jwks_uri' parameters",
	}

	// ErrorServerError is the standard error for server issues
	ErrorServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "server_error",
		Error:            "Server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)
