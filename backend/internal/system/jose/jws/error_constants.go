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

package jws

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for JWS operations
var (
	ErrorUnsupportedAlgorithm = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWS-1001",
		Error:            "Unsupported JWS algorithm",
		ErrorDescription: "The specified JWS algorithm is not supported",
	}

	ErrorInvalidSignature = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWS-1002",
		Error:            "Invalid signature",
		ErrorDescription: "The signature is invalid",
	}

	ErrorInvalidFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWS-1003",
		Error:            "Invalid JWS format",
		ErrorDescription: "The JWS token format is invalid",
	}

	ErrorDecodingHeader = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWS-1004",
		Error:            "JWS decode error",
		ErrorDescription: "Error occurred while decoding JWS header",
	}
)
