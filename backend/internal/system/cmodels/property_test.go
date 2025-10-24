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

package cmodels

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPropertyValueDTO_UnmarshalArray(t *testing.T) {
	jsonData := `["openid", "profile", "email"]`

	var pv PropertyValueDTO
	err := json.Unmarshal([]byte(jsonData), &pv)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !pv.IsArray() {
		t.Error("Expected IsArray() to be true")
	}

	expected := []string{"openid", "profile", "email"}
	if !reflect.DeepEqual(pv.multiple, expected) {
		t.Errorf("Expected %v, got %v", expected, pv.multiple)
	}
}

func TestPropertyValueDTO_UnmarshalString(t *testing.T) {
	jsonData := `"client_abc123"`

	var pv PropertyValueDTO
	err := json.Unmarshal([]byte(jsonData), &pv)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if pv.IsArray() {
		t.Error("Expected IsArray() to be false")
	}

	if pv.single != "client_abc123" {
		t.Errorf("Expected 'client_abc123', got '%s'", pv.single)
	}
}

func TestPropertyValueDTO_AsArray_BackwardCompatibility(t *testing.T) {
	// Test old comma-separated format
	jsonData := `"openid,profile,email"`

	var pv PropertyValueDTO
	err := json.Unmarshal([]byte(jsonData), &pv)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	result := pv.AsArray()
	expected := []string{"openid", "profile", "email"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPropertyValueDTO_MarshalArray(t *testing.T) {
	pv := NewArrayPropertyValue([]string{"openid", "profile"})

	jsonData, err := json.Marshal(pv)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `["openid","profile"]`
	if string(jsonData) != expected {
		t.Errorf("Expected %s, got %s", expected, string(jsonData))
	}
}

func TestPropertyValueDTO_MarshalString(t *testing.T) {
	pv := NewStringPropertyValue("test_value")

	jsonData, err := json.Marshal(pv)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `"test_value"`
	if string(jsonData) != expected {
		t.Errorf("Expected %s, got %s", expected, string(jsonData))
	}
}

func TestPropertiesMap_GetSetArray(t *testing.T) {
	props := make(PropertiesMap)

	scopes := []string{"openid", "profile"}
	props.SetArray("scope", scopes)

	result := props.GetArray("scope")
	if !reflect.DeepEqual(result, scopes) {
		t.Errorf("Expected %v, got %v", scopes, result)
	}
}

func TestPropertiesMap_GetSetString(t *testing.T) {
	props := make(PropertiesMap)

	props.SetString("client_id", "abc123")

	result := props.GetString("client_id")
	if result != "abc123" {
		t.Errorf("Expected 'abc123', got '%s'", result)
	}
}

func TestPropertyValueDTO_AsString_Array(t *testing.T) {
	pv := NewArrayPropertyValue([]string{"openid", "profile", "email"})

	result := pv.AsString()
	expected := "openid,profile,email"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestPropertyValueDTO_AsString_String(t *testing.T) {
	pv := NewStringPropertyValue("test_value")

	result := pv.AsString()
	expected := "test_value"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestPropertiesMap_MissingKey(t *testing.T) {
	props := make(PropertiesMap)

	// Test getting non-existent string
	result := props.GetString("nonexistent")
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}

	// Test getting non-existent array
	arrResult := props.GetArray("nonexistent")
	if len(arrResult) != 0 {
		t.Errorf("Expected empty array, got %v", arrResult)
	}
}