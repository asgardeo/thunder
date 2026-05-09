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

package model

import (
	"errors"
)

// ApplicationNotFoundError is the error returned when an application is not found.
var ApplicationNotFoundError error = errors.New("application not found")

// ApplicationDataCorruptedError is the error returned when application data is corrupted.
var ApplicationDataCorruptedError error = errors.New("application data is corrupted")

// Constants for MCP tool defaults
var (
	// DefaultUserAttributes are the standard user attributes for application templates.
	DefaultUserAttributes = []string{
		"email", "name", "given_name", "family_name",
		"profile", "picture", "phone_number", "address", "created_at",
	}
	// DefaultScopes are the standard OAuth scopes for application templates.
	DefaultScopes = []string{"openid", "profile", "email"}
)
