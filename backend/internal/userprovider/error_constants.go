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

package userprovider

// ErrorCode represents an error code.
type ErrorCode string

// UserProviderError represents an error returned by the user provider.
type UserProviderError struct {
	Code        ErrorCode `json:"code"`
	Message     string    `json:"message"`
	Description string    `json:"description"`
}

func (e *UserProviderError) Error() string {
	return e.Message + ": " + e.Description
}

// Error codes.
const (
	ErrorCodeSystemError              ErrorCode = "UP-0001"
	ErrorCodeUserNotFound             ErrorCode = "UP-0002"
	ErrorCodeInvalidRequestFormat     ErrorCode = "UP-0003"
	ErrorCodeOrganizationUnitMismatch ErrorCode = "UP-0004"
	ErrorCodeAttributeConflict        ErrorCode = "UP-0005"
	ErrorCodeMissingRequiredFields    ErrorCode = "UP-0006"
	ErrorCodeMissingCredentials       ErrorCode = "UP-0007"
	ErrorCodeNotImplemented           ErrorCode = "UP-0008"
)

// NewUserProviderError creates a new user provider error.
func NewUserProviderError(code ErrorCode, message string, description string) *UserProviderError {
	return &UserProviderError{
		Code:        code,
		Message:     message,
		Description: description,
	}
}
