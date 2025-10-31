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

// Package json provides JSON formatting for events.
package json

import (
	"encoding/json"

	"github.com/asgardeo/thunder/internal/observability/event"
	"github.com/asgardeo/thunder/internal/observability/formatter"
)

// JSONFormatter formats events as JSON.
type JSONFormatter struct{}

var _ formatter.Formatter = (*JSONFormatter)(nil)

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format formats an event as JSON.
func (jf *JSONFormatter) Format(evt *event.Event) ([]byte, error) {
	return json.Marshal(evt)
}

// GetName returns the name of this formatter.
func (jf *JSONFormatter) GetName() string {
	return "JSONFormatter"
}
