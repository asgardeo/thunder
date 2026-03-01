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

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

func TestErrorCode_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected bool
	}{
		{
			name:     "Valid code - CodeSystemError",
			code:     CodeSystemError,
			expected: true,
		},
		{
			name:     "Valid code - CodeUserNotFound",
			code:     CodeUserNotFound,
			expected: true,
		},
		{
			name:     "Valid code - CodeInvalidRequestFormat",
			code:     CodeInvalidRequestFormat,
			expected: true,
		},
		{
			name:     "Valid code - CodeOrganizationUnitMismatch",
			code:     CodeOrganizationUnitMismatch,
			expected: true,
		},
		{
			name:     "Valid code - CodeAttributeConflict",
			code:     CodeAttributeConflict,
			expected: true,
		},
		{
			name:     "Valid code - CodeMissingRequiredFields",
			code:     CodeMissingRequiredFields,
			expected: true,
		},
		{
			name:     "Valid code - CodeMissingCredentials",
			code:     CodeMissingCredentials,
			expected: true,
		},
		{
			name:     "Valid code - CodeNotImplemented",
			code:     CodeNotImplemented,
			expected: true,
		},
		{
			name:     "Invalid code - random string",
			code:     ErrorCode("INVALID-001"),
			expected: false,
		},
		{
			name:     "Invalid code - empty string",
			code:     ErrorCode(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.code.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewError_ValidCodes(t *testing.T) {
	tests := []struct {
		name                string
		code                ErrorCode
		description         string
		expectedCode        string
		expectedType        serviceerror.ServiceErrorType
		expectedError       string
		expectedDescription string
	}{
		{
			name:                "CodeSystemError",
			code:                CodeSystemError,
			description:         "Custom system error description",
			expectedCode:        "UP-0001",
			expectedType:        serviceerror.ServerErrorType,
			expectedError:       "System error",
			expectedDescription: "Custom system error description",
		},
		{
			name:                "CodeUserNotFound",
			code:                CodeUserNotFound,
			description:         "User does not exist",
			expectedCode:        "UP-0002",
			expectedType:        serviceerror.ClientErrorType,
			expectedError:       "User not found",
			expectedDescription: "User does not exist",
		},
		{
			name:                "CodeInvalidRequestFormat",
			code:                CodeInvalidRequestFormat,
			description:         "Request body is malformed",
			expectedCode:        "UP-0003",
			expectedType:        serviceerror.ClientErrorType,
			expectedError:       "Invalid request format",
			expectedDescription: "Request body is malformed",
		},
		{
			name:                "CodeOrganizationUnitMismatch",
			code:                CodeOrganizationUnitMismatch,
			description:         "Organization unit does not match user",
			expectedCode:        "UP-0004",
			expectedType:        serviceerror.ClientErrorType,
			expectedError:       "Organization unit mismatch",
			expectedDescription: "Organization unit does not match user",
		},
		{
			name:                "CodeAttributeConflict",
			code:                CodeAttributeConflict,
			description:         "Email already exists",
			expectedCode:        "UP-0005",
			expectedType:        serviceerror.ClientErrorType,
			expectedError:       "Attribute conflict",
			expectedDescription: "Email already exists",
		},
		{
			name:                "CodeMissingRequiredFields",
			code:                CodeMissingRequiredFields,
			description:         "Username is required",
			expectedCode:        "UP-0006",
			expectedType:        serviceerror.ClientErrorType,
			expectedError:       "Missing required fields",
			expectedDescription: "Username is required",
		},
		{
			name:                "CodeMissingCredentials",
			code:                CodeMissingCredentials,
			description:         "Password is required",
			expectedCode:        "UP-0007",
			expectedType:        serviceerror.ClientErrorType,
			expectedError:       "Missing credentials",
			expectedDescription: "Password is required",
		},
		{
			name:                "CodeNotImplemented",
			code:                CodeNotImplemented,
			description:         "Feature is coming soon",
			expectedCode:        "UP-0008",
			expectedType:        serviceerror.ServerErrorType,
			expectedError:       "Not implemented",
			expectedDescription: "Feature is coming soon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code, tt.description)

			assert.NotNil(t, err)
			assert.Equal(t, tt.expectedCode, err.Code)
			assert.Equal(t, tt.expectedType, err.Type)
			assert.Equal(t, tt.expectedError, err.Error)
			assert.Equal(t, tt.expectedDescription, err.ErrorDescription)
		})
	}
}

func TestNewError_InvalidCode_DefaultsToSystemError(t *testing.T) {
	tests := []struct {
		name        string
		code        ErrorCode
		description string
	}{
		{
			name:        "Invalid code - random string",
			code:        ErrorCode("INVALID-001"),
			description: "This should default to system error",
		},
		{
			name:        "Invalid code - empty string",
			code:        ErrorCode(""),
			description: "Empty code defaults to system error",
		},
		{
			name:        "Invalid code - wrong prefix",
			code:        ErrorCode("AUP-0001"),
			description: "Wrong prefix defaults to system error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code, tt.description)

			assert.NotNil(t, err)
			// Should default to CodeSystemError
			assert.Equal(t, "UP-0001", err.Code)
			assert.Equal(t, serviceerror.ServerErrorType, err.Type)
			assert.Equal(t, "System error", err.Error)
			assert.Equal(t, tt.description, err.ErrorDescription)
		})
	}
}

func TestNewError_PreservesDescription(t *testing.T) {
	descriptions := []string{
		"Simple description",
		"",
		"Description with special characters: !@#$%^&*()",
		"Multi-line\ndescription\nwith\nnewlines",
	}

	for _, desc := range descriptions {
		t.Run("Description: "+desc, func(t *testing.T) {
			err := NewError(CodeSystemError, desc)
			assert.Equal(t, desc, err.ErrorDescription)
		})
	}
}

func TestNewError_TypeAndErrorFieldsAreImmutable(t *testing.T) {
	// Test that for each code, the Type and Error fields are always the same
	// regardless of the description provided
	err1 := NewError(CodeUserNotFound, "Description 1")
	err2 := NewError(CodeUserNotFound, "Description 2")

	assert.Equal(t, err1.Code, err2.Code)
	assert.Equal(t, err1.Type, err2.Type)
	assert.Equal(t, err1.Error, err2.Error)
	assert.NotEqual(t, err1.ErrorDescription, err2.ErrorDescription)
}

func TestNewError_AllCodesHaveCorrectTypeMapping(t *testing.T) {
	// Verify that client error codes map to ClientErrorType
	clientErrorCodes := []ErrorCode{
		CodeUserNotFound,
		CodeInvalidRequestFormat,
		CodeOrganizationUnitMismatch,
		CodeAttributeConflict,
		CodeMissingRequiredFields,
		CodeMissingCredentials,
	}

	for _, code := range clientErrorCodes {
		t.Run(string(code)+" is ClientErrorType", func(t *testing.T) {
			err := NewError(code, "test")
			assert.Equal(t, serviceerror.ClientErrorType, err.Type)
		})
	}

	// Verify that server error codes map to ServerErrorType
	serverErrorCodes := []ErrorCode{
		CodeSystemError,
		CodeNotImplemented,
	}

	for _, code := range serverErrorCodes {
		t.Run(string(code)+" is ServerErrorType", func(t *testing.T) {
			err := NewError(code, "test")
			assert.Equal(t, serviceerror.ServerErrorType, err.Type)
		})
	}
}
