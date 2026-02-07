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

package layoutmgt

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

var (
	// ErrorLayoutNotFound is returned when a layout is not found.
	ErrorLayoutNotFound = serviceerror.ServiceError{
		Code:             "LAY-1001",
		Error:            "Layout not found",
		ErrorDescription: "The requested layout configuration was not found",
	}

	// ErrorInvalidLayoutID is returned when an invalid layout ID is provided.
	ErrorInvalidLayoutID = serviceerror.ServiceError{
		Code:             "LAY-1002",
		Error:            "Invalid layout ID",
		ErrorDescription: "The provided layout ID is invalid",
	}

	// ErrorInvalidLayoutData is returned when invalid layout data is provided.
	ErrorInvalidLayoutData = serviceerror.ServiceError{
		Code:             "LAY-1003",
		Error:            "Invalid layout data",
		ErrorDescription: "The provided layout data is invalid",
	}

	// ErrorLayoutAlreadyExists is returned when trying to create a layout that already exists.
	ErrorLayoutAlreadyExists = serviceerror.ServiceError{
		Code:             "LAY-1004",
		Error:            "Layout already exists",
		ErrorDescription: "A layout with the same ID already exists",
	}
)
