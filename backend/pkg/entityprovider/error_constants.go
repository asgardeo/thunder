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

package entityprovider

// ErrorCode represents an entity provider error code.
type ErrorCode string

// Error codes for entity provider operations.
const (
	ErrorCodeSystemError            ErrorCode = "EP-0001"
	ErrorCodeEntityNotFound         ErrorCode = "EP-0002"
	ErrorCodeInvalidRequestFormat   ErrorCode = "EP-0003"
	ErrorCodeAttributeConflict      ErrorCode = "EP-0004"
	ErrorCodeMissingRequiredFields  ErrorCode = "EP-0005"
	ErrorCodeMissingCredentials     ErrorCode = "EP-0006"
	ErrorCodeNotImplemented         ErrorCode = "EP-0007"
	ErrorCodeAmbiguousEntity        ErrorCode = "EP-0008"
	ErrorCodeSchemaValidationFailed ErrorCode = "EP-0009"
)

// EntityProviderError represents an error returned by the entity provider.
type EntityProviderError struct {
	Code        ErrorCode `json:"code"`
	Message     string    `json:"message"`
	Description string    `json:"description"`
}

func (e *EntityProviderError) Error() string {
	return e.Message + ": " + e.Description
}

// NewEntityProviderError creates a new entity provider error.
func NewEntityProviderError(code ErrorCode, message string, description string) *EntityProviderError {
	return &EntityProviderError{
		Code:        code,
		Message:     message,
		Description: description,
	}
}
