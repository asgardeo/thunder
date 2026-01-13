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

// Package core provides internationalization support.
package core

// I18nMessage represents a translatable message with a key and default value.
// Developers can use this struct inline when defining error messages or other
// translatable content. The extraction script will find all I18nMessage literals
// and generate the defaults.go file.
type I18nMessage struct {
	Key          string `json:"key"`
	DefaultValue string `json:"defaultValue"`
}

// String returns the default value of the message.
// This is useful for logging or when translation is not available.
func (m I18nMessage) String() string {
	return m.DefaultValue
}

// IsEmpty returns true if the message has no key set.
func (m I18nMessage) IsEmpty() bool {
	return m.Key == ""
}
