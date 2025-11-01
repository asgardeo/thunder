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

func TestNewProperty(t *testing.T) {
	prop, err := NewProperty("test", "value", false)
	if err != nil {
		t.Fatalf("NewProperty failed: %v", err)
	}

	if prop.GetName() != "test" {
		t.Errorf("Expected name 'test', got '%s'", prop.GetName())
	}

	if prop.IsMultiValued() {
		t.Error("Expected IsMultiValued() to be false")
	}

	value, err := prop.GetValue()
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}

	if value != "value" {
		t.Errorf("Expected value 'value', got '%s'", value)
	}
}

func TestNewMultiValuedProperty(t *testing.T) {
	values := []string{"openid", "profile", "email"}
	prop, err := NewMultiValuedProperty("scopes", values, false)
	if err != nil {
		t.Fatalf("NewMultiValuedProperty failed: %v", err)
	}

	if prop.GetName() != "scopes" {
		t.Errorf("Expected name 'scopes', got '%s'", prop.GetName())
	}

	if !prop.IsMultiValued() {
		t.Error("Expected IsMultiValued() to be true")
	}

	result, err := prop.GetValues()
	if err != nil {
		t.Fatalf("GetValues failed: %v", err)
	}

	if !reflect.DeepEqual(result, values) {
		t.Errorf("Expected %v, got %v", values, result)
	}
}

func TestProperty_GetValue_OnMultiValued_ShouldError(t *testing.T) {
	prop, _ := NewMultiValuedProperty("scopes", []string{"openid", "profile"}, false)

	_, err := prop.GetValue()
	if err == nil {
		t.Error("Expected error when calling GetValue() on multi-valued property")
	}
}

func TestProperty_GetValues_OnSingleValued_ShouldError(t *testing.T) {
	prop, _ := NewProperty("test", "value", false)

	_, err := prop.GetValues()
	if err == nil {
		t.Error("Expected error when calling GetValues() on single-valued property")
	}
}

func TestSerializeProperties_SingleAndMultiValued(t *testing.T) {
	singleProp, _ := NewProperty("client_id", "abc123", false)
	multiProp, _ := NewMultiValuedProperty("scopes", []string{"openid", "profile"}, false)

	properties := []Property{*singleProp, *multiProp}

	jsonStr, err := SerializePropertiesToJSONArray(properties)
	if err != nil {
		t.Fatalf("SerializePropertiesToJSONArray failed: %v", err)
	}

	// Verify it's valid JSON
	var result []PropertyDTO
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(result))
	}

	// Check single-valued property
	if result[0].Name != "client_id" {
		t.Errorf("Expected name 'client_id', got '%s'", result[0].Name)
	}
	if result[0].IsMultiValued {
		t.Error("Expected IsMultiValued to be false for client_id")
	}
	if result[0].Value != "abc123" {
		t.Errorf("Expected value 'abc123', got '%s'", result[0].Value)
	}

	// Check multi-valued property
	if result[1].Name != "scopes" {
		t.Errorf("Expected name 'scopes', got '%s'", result[1].Name)
	}
	if !result[1].IsMultiValued {
		t.Error("Expected IsMultiValued to be true for scopes")
	}
	if !reflect.DeepEqual(result[1].Values, []string{"openid", "profile"}) {
		t.Errorf("Expected ['openid', 'profile'], got %v", result[1].Values)
	}
}

func TestDeserializeProperties_SingleAndMultiValued(t *testing.T) {
	jsonStr := `[
		{"name":"client_id","value":"abc123","is_secret":false,"is_multi_valued":false},
		{"name":"scopes","values":["openid","profile","email"],"is_secret":false,"is_multi_valued":true}
	]`

	properties, err := DeserializePropertiesFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("DeserializePropertiesFromJSON failed: %v", err)
	}

	if len(properties) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(properties))
	}

	// Check single-valued property
	if properties[0].GetName() != "client_id" {
		t.Errorf("Expected name 'client_id', got '%s'", properties[0].GetName())
	}
	if properties[0].IsMultiValued() {
		t.Error("Expected IsMultiValued() to be false")
	}
	value, _ := properties[0].GetValue()
	if value != "abc123" {
		t.Errorf("Expected value 'abc123', got '%s'", value)
	}

	// Check multi-valued property
	if properties[1].GetName() != "scopes" {
		t.Errorf("Expected name 'scopes', got '%s'", properties[1].GetName())
	}
	if !properties[1].IsMultiValued() {
		t.Error("Expected IsMultiValued() to be true")
	}
	values, _ := properties[1].GetValues()
	expected := []string{"openid", "profile", "email"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Expected %v, got %v", expected, values)
	}
}

func TestPropertyDTO_ToProperty_SingleValued(t *testing.T) {
	dto := &PropertyDTO{
		Name:          "test",
		Value:         "value",
		IsSecret:      false,
		IsMultiValued: false,
	}

	prop, err := dto.ToProperty()
	if err != nil {
		t.Fatalf("ToProperty failed: %v", err)
	}

	if prop.IsMultiValued() {
		t.Error("Expected IsMultiValued() to be false")
	}

	value, _ := prop.GetValue()
	if value != "value" {
		t.Errorf("Expected 'value', got '%s'", value)
	}
}

func TestPropertyDTO_ToProperty_MultiValued(t *testing.T) {
	dto := &PropertyDTO{
		Name:          "scopes",
		Values:        []string{"openid", "profile"},
		IsSecret:      false,
		IsMultiValued: true,
	}

	prop, err := dto.ToProperty()
	if err != nil {
		t.Fatalf("ToProperty failed: %v", err)
	}

	if !prop.IsMultiValued() {
		t.Error("Expected IsMultiValued() to be true")
	}

	values, _ := prop.GetValues()
	if !reflect.DeepEqual(values, []string{"openid", "profile"}) {
		t.Errorf("Expected ['openid', 'profile'], got %v", values)
	}
}

func TestProperty_ToPropertyDTO_SingleValued(t *testing.T) {
	prop, _ := NewProperty("test", "value", false)

	dto, err := prop.ToPropertyDTO()
	if err != nil {
		t.Fatalf("ToPropertyDTO failed: %v", err)
	}

	if dto.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", dto.Name)
	}

	if dto.IsMultiValued {
		t.Error("Expected IsMultiValued to be false")
	}

	if dto.Value != "value" {
		t.Errorf("Expected value 'value', got '%s'", dto.Value)
	}
}

func TestProperty_ToPropertyDTO_MultiValued(t *testing.T) {
	prop, _ := NewMultiValuedProperty("scopes", []string{"openid", "profile"}, false)

	dto, err := prop.ToPropertyDTO()
	if err != nil {
		t.Fatalf("ToPropertyDTO failed: %v", err)
	}

	if dto.Name != "scopes" {
		t.Errorf("Expected name 'scopes', got '%s'", dto.Name)
	}

	if !dto.IsMultiValued {
		t.Error("Expected IsMultiValued to be true")
	}

	if !reflect.DeepEqual(dto.Values, []string{"openid", "profile"}) {
		t.Errorf("Expected ['openid', 'profile'], got %v", dto.Values)
	}
}