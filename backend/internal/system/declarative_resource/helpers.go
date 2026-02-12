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

package declarativeresource

import (
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// IsDeclarativeModeEnabled checks if declarative resources are enabled in the configuration.
func IsDeclarativeModeEnabled() bool {
	return config.GetThunderRuntime().Config.DeclarativeResources.Enabled
}

// CheckDeclarativeCreate returns an error if declarative mode is enabled and create operation is attempted.
func CheckDeclarativeCreate() *serviceerror.ServiceError {
	if IsDeclarativeModeEnabled() {
		return &ErrorDeclarativeResourceCreateOperation
	}
	return nil
}

// CheckDeclarativeUpdate returns an error if declarative mode is enabled and update operation is attempted.
func CheckDeclarativeUpdate() *serviceerror.ServiceError {
	if IsDeclarativeModeEnabled() {
		return &ErrorDeclarativeResourceUpdateOperation
	}
	return nil
}

// CheckDeclarativeDelete returns an error if declarative mode is enabled and delete operation is attempted.
func CheckDeclarativeDelete() *serviceerror.ServiceError {
	if IsDeclarativeModeEnabled() {
		return &ErrorDeclarativeResourceDeleteOperation
	}
	return nil
}

// CheckDeclarativeCreateI18n returns an i18n error if declarative mode is enabled and create operation is attempted.
func CheckDeclarativeCreateI18n() *serviceerror.I18nServiceError {
	if IsDeclarativeModeEnabled() {
		return &I18nErrorDeclarativeResourceCreateOperation
	}
	return nil
}

// CheckDeclarativeUpdateI18n returns an i18n error if declarative mode is enabled and update operation is attempted.
func CheckDeclarativeUpdateI18n() *serviceerror.I18nServiceError {
	if IsDeclarativeModeEnabled() {
		return &I18nErrorDeclarativeResourceUpdateOperation
	}
	return nil
}

// CheckDeclarativeDeleteI18n returns an i18n error if declarative mode is enabled and delete operation is attempted.
func CheckDeclarativeDeleteI18n() *serviceerror.I18nServiceError {
	if IsDeclarativeModeEnabled() {
		return &I18nErrorDeclarativeResourceDeleteOperation
	}
	return nil
}
