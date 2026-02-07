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

package thememgt

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

var (
	// ErrorThemeNotFound is returned when a theme is not found.
	ErrorThemeNotFound = serviceerror.ServiceError{
		Code:             "THM-1001",
		Error:            "Theme not found",
		ErrorDescription: "The requested theme configuration was not found",
	}

	// ErrorInvalidThemeID is returned when an invalid theme ID is provided.
	ErrorInvalidThemeID = serviceerror.ServiceError{
		Code:             "THM-1002",
		Error:            "Invalid theme ID",
		ErrorDescription: "The provided theme ID is invalid",
	}

	// ErrorInvalidThemeData is returned when invalid theme data is provided.
	ErrorInvalidThemeData = serviceerror.ServiceError{
		Code:             "THM-1003",
		Error:            "Invalid theme data",
		ErrorDescription: "The provided theme data is invalid",
	}

	// ErrorThemeAlreadyExists is returned when trying to create a theme that already exists.
	ErrorThemeAlreadyExists = serviceerror.ServiceError{
		Code:             "THM-1004",
		Error:            "Theme already exists",
		ErrorDescription: "A theme with the same ID already exists",
	}
)
