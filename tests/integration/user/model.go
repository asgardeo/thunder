/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package user

import (
	"encoding/json"
	"sort"
)

type User struct {
	Id         string          `json:"id"`
	OrgId      string          `json:"org_id"`
	Type       string          `json:"type"`
	Attributes json.RawMessage `json:"attributes"`
}

func compareStringSlices(a, b []string) bool {

	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// compare and validate whether two users have equal content
func (user *User) equals(expectedUser User) bool {
	if user.Id != expectedUser.Id || user.OrgId != expectedUser.OrgId || user.Type != expectedUser.Type {
		return false
	}

	// Compare the Attributes JSON
	var attr1, attr2 map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attr1); err != nil {
		return false
	}
	if err := json.Unmarshal(expectedUser.Attributes, &attr2); err != nil {
		return false
	}

	return compareStringSlices(user.getSortedKeys(attr1), user.getSortedKeys(attr2))
}

// getSortedKeys returns the sorted keys of a map for consistent comparison
func (user *User) getSortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
