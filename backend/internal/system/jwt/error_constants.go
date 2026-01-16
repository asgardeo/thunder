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

package jwt

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for JWT service
var (
	ErrorDecodingJWTHeader = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT-1001",
		Error:            "JWT decode error",
		ErrorDescription: "Error occurred while decoding JWT header",
	}

	ErrorDecodingJWTPayload = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT-1002",
		Error:            "JWT decode error",
		ErrorDescription: "Error occurred while decoding JWT payload",
	}

	ErrorUnsupportedJWSAlgorithm = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT-1003",
		Error:            "Unsupported JWS algorithm",
		ErrorDescription: "The specified JWS algorithm is not supported",
	}

	ErrorInvalidTokenSignature = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT-1004",
		Error:            "Invalid token signature",
		ErrorDescription: "The JWT token signature is invalid",
	}

	ErrorInvalidJWTFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT-1005",
		Error:            "Invalid JWT format",
		ErrorDescription: "The JWT token format is invalid",
	}

	ErrorNoMatchingJWKFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT-1006",
		Error:            "No matching JWK found",
		ErrorDescription: "No matching JWK found for the given Key ID",
	}

	ErrorTokenExpired = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT-1007",
		Error:            "Token expired",
		ErrorDescription: "The JWT token has expired",
	}

	ErrorFailedToGetJWKS = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT-1008",
		Error:            "Failed to retrieve JWKS",
		ErrorDescription: "Failed to retrieve JWKS from the specified URL",
	}

	ErrorFailedToParseJWKS = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWT-1009",
		Error:            "Failed to parse JWKS",
		ErrorDescription: "Failed to parse JWKS",
	}
)
