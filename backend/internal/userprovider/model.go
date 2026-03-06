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

// User represents a user.
type User struct {
	UserID             string          `json:"userId"`
	UserType           string          `json:"userType"`
	OrganizationUnitID string          `json:"ouId"`
	Attributes         json.RawMessage `json:"attributes,omitempty"`
}

// Link represents a link.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// UserGroup represents a user group.
type UserGroup struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	OrganizationUnitID string `json:"ouId"`
}

// UserGroupListResponse represents a response containing a list of user groups.
type UserGroupListResponse struct {
	TotalResults int         `json:"totalResults"`
	StartIndex   int         `json:"startIndex"`
	Count        int         `json:"count"`
	Groups       []UserGroup `json:"groups"`
	Links        []Link      `json:"links"`
}
