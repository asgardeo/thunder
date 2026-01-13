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

// Package apierror defines the error structures for the API.
package apierror

import (
	"github.com/asgardeo/thunder/internal/system/i18n/core"
)

// ErrorResponse defines a generic API error response.
type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

// I18nErrorResponse defines an API error response with i18n support.
// This is the new error response type that should be used for services being migrated to i18n.
// Translatable fields use core.Message instead of plain strings.
type I18nErrorResponse struct {
	Code        string           `json:"code"`
	Message     core.I18nMessage `json:"message"`
	Description core.I18nMessage `json:"description"`
}
