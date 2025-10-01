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

package google

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// customServiceError creates a new service error based on an existing error with custom description.
func customServiceError(svcError serviceerror.ServiceError, errorDesc string) *serviceerror.ServiceError {
	err := &serviceerror.ServiceError{
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
