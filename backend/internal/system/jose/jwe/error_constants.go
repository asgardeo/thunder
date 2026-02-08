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

package jwe

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for JWE service
var (
	// ErrorDecodingJWE is the error returned when decoding the JWE token fails.
	ErrorDecodingJWE = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWE-1001",
		Error:            "JWE decode error",
		ErrorDescription: "Error occurred while decoding JWE token",
	}

	// ErrorJWEDecryptionFailed is the error returned when the JWE token decryption fails.
	ErrorJWEDecryptionFailed = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWE-1002",
		Error:            "JWE decryption failed",
		ErrorDescription: "Failed to decrypt the JWE token",
	}

	// ErrorUnsupportedJWEAlgorithm is the error returned when the JWE algorithm is unsupported.
	ErrorUnsupportedJWEAlgorithm = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWE-1003",
		Error:            "Unsupported JWE algorithm",
		ErrorDescription: "The specified JWE algorithm is not supported",
	}

	// ErrorUnsupportedEncryptionAlgorithm is the error returned when the encryption algorithm is unsupported.
	ErrorUnsupportedEncryptionAlgorithm = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "JWE-1004",
		Error:            "Unsupported encryption algorithm",
		ErrorDescription: "The specified encryption algorithm is not supported",
	}
)
