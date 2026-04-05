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

package entity

import "errors"

// Error variables for entity operations.
var (
	// ErrEntityNotFound is returned when the entity is not found in the system.
	ErrEntityNotFound = errors.New("entity not found")

	// ErrAuthenticationFailed is returned when entity credential verification fails.
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrSchemaValidationFailed is returned when entity attributes fail schema validation.
	ErrSchemaValidationFailed = errors.New("schema validation failed")

	// ErrAttributeConflict is returned when entity attributes conflict with an existing entity.
	ErrAttributeConflict = errors.New("attribute conflict")

	// ErrInvalidCredential is returned when a credential value is invalid.
	ErrInvalidCredential = errors.New("invalid credential")

	// ErrBadAttributesInRequest is returned when the attributes in the request are invalid.
	ErrBadAttributesInRequest = errors.New("failed to marshal attributes")

	// errResultLimitExceededInCompositeMode is returned when the result limit is exceeded in composite mode.
	errResultLimitExceededInCompositeMode = errors.New("result limit exceeded in composite mode")
)
