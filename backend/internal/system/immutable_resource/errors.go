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

package immutableresource

import (
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/i18n/core"
)

var (
	// ErrorImmutableResourceCreateOperation is the error returned when
	// an immutable resource create operation is attempted.
	ErrorImmutableResourceCreateOperation = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FBR-1001",
		Error:            "Immutable resource create operation is not allowed",
		ErrorDescription: "Creating immutable resources is not permitted",
	}

	// I18nErrorImmutableResourceCreateOperation is the error returned when
	// an immutable resource create operation is attempted.
	I18nErrorImmutableResourceCreateOperation = serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "FBR-1001",
		Error: core.I18nMessage{
			Key:          "error.immutable_resource.create_operation_not_allowed",
			DefaultValue: "Immutable resource create operation is not allowed",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "error.immutable_resource.create_operation_not_allowed_description",
			DefaultValue: "Creating immutable resources is not permitted",
		},
	}

	// ErrorImmutableResourceUpdateOperation is the error returned when
	// an immutable resource update operation is attempted.
	ErrorImmutableResourceUpdateOperation = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FBR-1002",
		Error:            "Immutable resource update operation is not allowed",
		ErrorDescription: "Updating immutable resources is not permitted",
	}

	// I18nErrorImmutableResourceUpdateOperation is the error returned when
	// an immutable resource update operation is attempted.
	I18nErrorImmutableResourceUpdateOperation = serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "FBR-1002",
		Error: core.I18nMessage{
			Key:          "error.immutable_resource.update_operation_not_allowed",
			DefaultValue: "Immutable resource update operation is not allowed",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "error.immutable_resource.update_operation_not_allowed_description",
			DefaultValue: "Updating immutable resources is not permitted",
		},
	}

	// ErrorImmutableResourceDeleteOperation is the error returned when
	// an immutable resource delete operation is attempted.
	ErrorImmutableResourceDeleteOperation = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FBR-1003",
		Error:            "Immutable resource delete operation is not allowed",
		ErrorDescription: "Deleting immutable resources is not permitted",
	}

	// I18nErrorImmutableResourceDeleteOperation is the error returned when
	// an immutable resource delete operation is attempted.
	I18nErrorImmutableResourceDeleteOperation = serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "FBR-1003",
		Error: core.I18nMessage{
			Key:          "error.immutable_resource.delete_operation_not_allowed",
			DefaultValue: "Immutable resource delete operation is not allowed",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "error.immutable_resource.delete_operation_not_allowed_description",
			DefaultValue: "Deleting immutable resources is not permitted",
		},
	}
)
