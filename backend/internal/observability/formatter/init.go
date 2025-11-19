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

package formatter

// Initialize creates and returns a formatter instance based on the specified format type.
//
// Supported format types:
//   - "json": JSON formatter (default)
//   - Future: "csv", "text", etc.
//
// Parameters:
//   - formatType: The type of formatter to create ("json", etc.)
//
// Returns:
//   - FormatterInterface: The initialized formatter instance
//
// Example:
//
//	formatter := formatter.Initialize("json")
//	data, err := formatter.Format(event)
func Initialize(formatType string) FormatterInterface {
	switch formatType {
	case "json":
		return newJSONFormatter()
	default:
		// Default to JSON formatter
		return newJSONFormatter()
	}
}
