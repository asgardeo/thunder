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

package userprovider

import "encoding/json"

// errNotImplemented is the error returned when a method is not implemented.
var errNotImplemented = NewUserProviderError(
	ErrorCodeNotImplemented,
	"Method Not Implemented",
	"The method is not implemented by the user provider.",
)

// disabledUserProvider is a user provider that returns an error for all methods.
type disabledUserProvider struct{}

// newDisabledUserProvider creates a new DisabledUserProvider.
func newDisabledUserProvider() UserProviderInterface {
	return &disabledUserProvider{}
}

// IdentifyUser returns a not implemented error.
func (p *disabledUserProvider) IdentifyUser(filters map[string]interface{}) (*string, *UserProviderError) {
	return nil, errNotImplemented
}

// GetUser returns a not implemented error.
func (p *disabledUserProvider) GetUser(userID string) (*User, *UserProviderError) {
	return nil, errNotImplemented
}

// GetUserGroups returns a not implemented error.
func (p *disabledUserProvider) GetUserGroups(userID string, limit, offset int) (*UserGroupListResponse,
	*UserProviderError) {
	return nil, errNotImplemented
}

// UpdateUser returns a not implemented error.
func (p *disabledUserProvider) UpdateUser(userID string, user *User) (*User, *UserProviderError) {
	return nil, errNotImplemented
}

// CreateUser returns a not implemented error.
func (p *disabledUserProvider) CreateUser(user *User) (*User, *UserProviderError) {
	return nil, errNotImplemented
}

// UpdateUserCredentials returns a not implemented error.
func (p *disabledUserProvider) UpdateUserCredentials(userID string, credentials json.RawMessage) *UserProviderError {
	return errNotImplemented
}

// DeleteUser returns a not implemented error.
func (p *disabledUserProvider) DeleteUser(userID string) *UserProviderError {
	return errNotImplemented
}
