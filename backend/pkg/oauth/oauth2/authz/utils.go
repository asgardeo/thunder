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

package authz

import (
	"fmt"
	"strings"
	"time"
)

// parseTimeField parses a time field from the database result.
func parseTimeField(field interface{}, fieldName string) (time.Time, error) {
	const customTimeFormat = "2006-01-02 15:04:05.999999999"

	switch v := field.(type) {
	case string:
		// Handle SQLite datetime strings
		trimmedTime := trimTimeString(v)
		parsedTime, err := time.Parse(customTimeFormat, trimmedTime)
		if err != nil {
			// Try alternative ISO 8601 format as fallback
			parsedTime, err = time.Parse("2006-01-02T15:04:05Z07:00", v)
			if err != nil {
				return time.Time{}, fmt.Errorf("error parsing %s: %w", fieldName, err)
			}
		}
		return parsedTime, nil
	case time.Time:
		return v, nil
	default:
		return time.Time{}, fmt.Errorf("unexpected type for %s", fieldName)
	}
}

// trimTimeString trims extra information from a time string to match the expected format.
func trimTimeString(timeStr string) string {
	parts := strings.SplitN(timeStr, " ", 3)
	if len(parts) >= 2 {
		return parts[0] + " " + parts[1]
	}
	return timeStr
}
