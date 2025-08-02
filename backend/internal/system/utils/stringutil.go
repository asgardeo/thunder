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

package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ParseStringArray parses a string into a slice of strings using the specified separator.
func ParseStringArray(value string, separator string) []string {
	return ParseTypedStringArray[string](value, separator)
}

// ParseTypedStringArray parses a string into a slice of strings of type T using the specified separator.
func ParseTypedStringArray[T ~string](value string, separator string) []T {
	if value == "" {
		return []T{}
	}
	if separator == "" {
		separator = ","
	}
	parts := strings.Split(value, separator)
	result := make([]T, len(parts))
	for i, p := range parts {
		result[i] = T(strings.TrimSpace(p))
	}
	return result
}

// StringifyStringArray converts a slice of strings into a single string,
// joining the elements with the specified separator. If the slice is empty,
// it returns an empty string. If the separator is empty, it defaults to a comma.
func StringifyStringArray(values []string, separator string) string {
	if len(values) == 0 {
		return ""
	}
	if separator == "" {
		separator = ","
	}
	return strings.Join(values, separator)
}

// ConvertInterfaceMapToStringMap converts a map with string keys and interface{} values
// to a map with string keys and string values. If the value is a slice, it converts
// the slice elements to a comma-separated string.
// If the value is a string, it adds it directly to the output map.
func ConvertInterfaceMapToStringMap(input map[string]interface{}) map[string]string {
	if input == nil {
		return nil
	}

	output := make(map[string]string)
	for key, value := range input {
		output[key] = ConvertInterfaceValueToString(value)
	}
	return output
}

// ConvertInterfaceValueToString converts any interface{} to a string.
// It handles common types like string, bool, int, float, and slices.
// For slices, it concatenates the elements into a comma-separated string.
func ConvertInterfaceValueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	// Directly handle common types.
	switch v := value.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	}

	// Generic slice/array handling (works for []T of any type).
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		var parts []string
		for i := 0; i < val.Len(); i++ {
			parts = append(parts, ConvertInterfaceValueToString(val.Index(i).Interface()))
		}
		return strings.Join(parts, ",")
	}

	// Fallback: default formatting.
	return fmt.Sprintf("%v", value)
}

// MergeStringMaps merges two maps of strings and returns the result.
func MergeStringMaps(dst, src map[string]string) map[string]string {
	if dst == nil {
		dst = make(map[string]string)
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// BoolToNumString converts a boolean value to a numeric string representation.
func BoolToNumString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// NumStringToBool converts a numeric string representation to a boolean value.
func NumStringToBool(s string) bool {
	return s == "1"
}
