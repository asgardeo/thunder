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

package subscriber

import (
	"testing"
)

// TestInitialize_RegistryHasSubscribers tests that the registry contains expected subscribers
func TestInitialize_RegistryHasSubscribers(t *testing.T) {
	// Get registered names
	names := GetRegisteredNames()

	// Should have at least the three built-in subscribers
	if len(names) < 3 {
		t.Errorf("Expected at least 3 registered subscribers, got %d", len(names))
	}

	// Check for expected subscriber types
	expectedTypes := map[string]bool{
		"console": false,
		"file":    false,
		"otel":    false,
	}

	for _, name := range names {
		if _, exists := expectedTypes[name]; exists {
			expectedTypes[name] = true
		}
	}

	// Verify all expected types were found
	for typeName, found := range expectedTypes {
		if !found {
			t.Errorf("Expected to find '%s' subscriber in registry", typeName)
		}
	}
}

// TestInitialize_FactoriesCreateInstances tests that all registered factories can create instances
func TestInitialize_FactoriesCreateInstances(t *testing.T) {
	factories := getAllFactories()

	if len(factories) < 3 {
		t.Errorf("Expected at least 3 factories, got %d", len(factories))
	}

	for name, factory := range factories {
		instance := factory()
		if instance == nil {
			t.Errorf("Factory '%s' returned nil instance", name)
			continue
		}

		// Verify instance has required interface methods (they may return zero values until Initialize is called)
		_ = instance.GetID()         // Should not panic
		_ = instance.GetCategories() // Should not panic
		// Note: Can't call IsEnabled() without ThunderRuntime being initialized

		// Verify the instance has the SubscriberInterface
		var _ SubscriberInterface = instance
	}
}
