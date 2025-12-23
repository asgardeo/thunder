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
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// IsImmutableModeEnabled checks if immutable resources are enabled in the configuration.
func IsImmutableModeEnabled() bool {
	return config.GetThunderRuntime().Config.ImmutableResources.Enabled
}

// CheckImmutableCreate returns an error if immutable mode is enabled and create operation is attempted.
func CheckImmutableCreate() *serviceerror.ServiceError {
	if IsImmutableModeEnabled() {
		return &ErrorImmutableResourceCreateOperation
	}
	return nil
}

// CheckImmutableUpdate returns an error if immutable mode is enabled and update operation is attempted.
func CheckImmutableUpdate() *serviceerror.ServiceError {
	if IsImmutableModeEnabled() {
		return &ErrorImmutableResourceUpdateOperation
	}
	return nil
}

// CheckImmutableDelete returns an error if immutable mode is enabled and delete operation is attempted.
func CheckImmutableDelete() *serviceerror.ServiceError {
	if IsImmutableModeEnabled() {
		return &ErrorImmutableResourceDeleteOperation
	}
	return nil
}

// CheckImmutableCreateI18n returns an i18n error if immutable mode is enabled and create operation is attempted.
func CheckImmutableCreateI18n() *serviceerror.I18nServiceError {
	if IsImmutableModeEnabled() {
		return &I18nErrorImmutableResourceCreateOperation
	}
	return nil
}

// CheckImmutableUpdateI18n returns an i18n error if immutable mode is enabled and update operation is attempted.
func CheckImmutableUpdateI18n() *serviceerror.I18nServiceError {
	if IsImmutableModeEnabled() {
		return &I18nErrorImmutableResourceUpdateOperation
	}
	return nil
}

// CheckImmutableDeleteI18n returns an i18n error if immutable mode is enabled and delete operation is attempted.
func CheckImmutableDeleteI18n() *serviceerror.I18nServiceError {
	if IsImmutableModeEnabled() {
		return &I18nErrorImmutableResourceDeleteOperation
	}
	return nil
}
