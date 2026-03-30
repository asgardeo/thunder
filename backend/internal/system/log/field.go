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

package log

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value interface{}
}

// String creates a Field with a string value.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates a Field with an integer value.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a Field with a boolean value.
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Any creates a Field with any value.
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Error creates a Field with an error value.
func Error(value error) Field {
	return Field{Key: "error", Value: value}
}

// MaskedString creates a Field with a masked string value.
func MaskedString(key, value string) Field {
	return Field{Key: key, Value: MaskString(value)}
}
