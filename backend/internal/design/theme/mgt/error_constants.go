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

package thememgt

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

var (
	// ErrorInvalidThemeData is returned when invalid theme data is provided.
	ErrorInvalidThemeData = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1001",
		Error:            "Invalid theme data",
		ErrorDescription: "The provided theme data is invalid",
	}

	// ErrorInvalidThemeID is returned when an invalid theme ID is provided.
	ErrorInvalidThemeID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1002",
		Error:            "Invalid theme ID",
		ErrorDescription: "The provided theme ID is invalid",
	}

	// ErrorThemeNotFound is returned when a theme is not found.
	ErrorThemeNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1003",
		Error:            "Theme not found",
		ErrorDescription: "The requested theme configuration was not found",
	}

	// ErrorThemeInUse is returned when trying to delete a theme that is being used by applications.
	ErrorThemeInUse = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1004",
		Error:            "Theme in use",
		ErrorDescription: "Cannot delete theme that is currently associated with one or more applications",
	}

	// ErrorMissingDisplayName is returned when display name is not provided.
	ErrorMissingDisplayName = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1005",
		Error:            "Missing display name",
		ErrorDescription: "Display name is required",
	}

	// ErrorMissingTheme is returned when theme field is not provided.
	ErrorMissingTheme = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1006",
		Error:            "Missing theme",
		ErrorDescription: "Theme field is required",
	}

	// ErrorInvalidThemeFormat is returned when theme JSON is invalid.
	ErrorInvalidThemeFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1007",
		Error:            "Invalid theme format",
		ErrorDescription: "Theme must be a valid JSON object",
	}

	// ErrorInvalidLimitValue is returned when limit validation fails in service layer.
	ErrorInvalidLimitValue = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1008",
		Error:            "Invalid limit",
		ErrorDescription: "Limit value is out of valid range",
	}

	// ErrorInvalidOffsetValue is returned when offset validation fails in service layer.
	ErrorInvalidOffsetValue = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1009",
		Error:            "Invalid offset",
		ErrorDescription: "Offset must be non-negative",
	}

	// ErrorInvalidLimitParam is returned when limit parameter cannot be parsed.
	ErrorInvalidLimitParam = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1010",
		Error:            "Invalid limit",
		ErrorDescription: "Limit must be a valid integer",
	}

	// ErrorInvalidOffsetParam is returned when offset parameter cannot be parsed.
	ErrorInvalidOffsetParam = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1011",
		Error:            "Invalid offset",
		ErrorDescription: "Offset must be a valid integer",
	}

	// ErrorCannotUpdateDeclarativeTheme is returned when attempting to update a declarative theme.
	ErrorCannotUpdateDeclarativeTheme = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1012",
		Error:            "Cannot update declarative theme",
		ErrorDescription: "Theme is defined in declarative resources and cannot be modified",
	}

	// ErrorCannotDeleteDeclarativeTheme is returned when attempting to delete a declarative theme.
	ErrorCannotDeleteDeclarativeTheme = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1013",
		Error:            "Cannot delete declarative theme",
		ErrorDescription: "Theme is defined in declarative resources and cannot be deleted",
	}

	// ErrorResultLimitExceededInCompositeMode is returned when composite store result count exceeds max limit.
	ErrorResultLimitExceededInCompositeMode = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "THM-5001",
		Error:            "Result limit exceeded",
		ErrorDescription: "Total count of themes exceeds maximum allowed limit in composite mode",
	}

	// ErrorCannotModifyDeclarativeResource is returned when attempting to modify a declarative theme.
	ErrorCannotModifyDeclarativeResource = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "THM-1014",
		Error:            "Cannot modify declarative resource",
		ErrorDescription: "The theme is declarative and cannot be modified or deleted",
	}
)

// errCannotUpdateDeclarativeTheme is an internal error for composite store operations.
var errCannotUpdateDeclarativeTheme = errors.New("cannot update declarative theme")

// errCannotDeleteDeclarativeTheme is an internal error for composite store operations.
var errCannotDeleteDeclarativeTheme = errors.New("cannot delete declarative theme")

// errResultLimitExceededInCompositeMode is returned when composite store result count exceeds max limit.
var errResultLimitExceededInCompositeMode = errors.New("result limit exceeded in composite mode")
