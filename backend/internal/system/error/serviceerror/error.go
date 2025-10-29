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

// Package serviceerror defines the error structures for the service layer.
package serviceerror

// ServiceErrorType defines the type of service error.
type ServiceErrorType string

const (
	// ClientErrorType denotes the client error type.
	ClientErrorType ServiceErrorType = "client_error"
	// ServerErrorType denotes the server error type.
	ServerErrorType ServiceErrorType = "server_error"
)

// ServiceError defines a generic error structure that can be used across the service layer.
type ServiceError struct {
	Code             string           `json:"code"`
	Type             ServiceErrorType `json:"type"`
	Error            string           `json:"error"`
	ErrorDescription string           `json:"error_description,omitempty"`
}

// CustomServiceError creates a new service error based on an existing error with custom description.
func CustomServiceError(svcError ServiceError, errorDesc string) *ServiceError {
	err := &ServiceError{
		Type:             svcError.Type,
		Code:             svcError.Code,
		Error:            svcError.Error,
		ErrorDescription: svcError.ErrorDescription,
	}
	if errorDesc != "" {
		err.ErrorDescription = errorDesc
	}
	return err
}

// Server errors
var (
	// EncodingError is the error returned when encoding the response.
	ErrorEncodingError = "{Code: \"ENC-5000\",Error: \"Encoding error\"," +
		"ErrorDescription: \"An error occurred while encoding the response\"}"
)
