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

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFromDirectory(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// content for file 1 (root)
	file1Content := `
package main

import "github.com/wso2/thunder/backend/internal/system/i18n/core"

var msg1 = core.I18nMessage{
    Key: "key.one",
    DefaultValue: "Value One",
}
`
	// content for file 2 (nested)
	file2Content := `
package sub

import "github.com/wso2/thunder/backend/internal/system/i18n/core"

var msg2 = core.I18nMessage{
    Key: "key.two",
    DefaultValue: "Value Two",
}
`
	// content for file 3 (nested deeper)
	file3Content := `
package deep

import "github.com/wso2/thunder/backend/internal/system/i18n/core"

var msg3 = core.I18nMessage{
    Key: "key.three",
    DefaultValue: "Value Three",
}
`

	// Create files
	if err := os.WriteFile(filepath.Join(tempDir, "file1.go"), []byte(file1Content), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	subDir := filepath.Join(tempDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file2.go"), []byte(file2Content), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	deepDir := filepath.Join(subDir, "deep")
	if err := os.Mkdir(deepDir, 0755); err != nil {
		t.Fatalf("Failed to create deepdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deepDir, "file3.go"), []byte(file3Content), 0644); err != nil {
		t.Fatalf("Failed to create file3: %v", err)
	}

	// Create a test file (should be ignored)
	testFileContent := `
package main
import "github.com/wso2/thunder/backend/internal/system/i18n/core"
var testMsg = core.I18nMessage{Key: "ignored", DefaultValue: "Ignored"}
`
	if err := os.WriteFile(filepath.Join(tempDir, "file_test.go"), []byte(testFileContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run extractor
	extractor := NewExtractor(false)
	messages, err := extractor.ExtractFromDirectory(tempDir)
	if err != nil {
		t.Fatalf("ExtractFromDirectory failed: %v", err)
	}

	// Verify results
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}

	expectedKeys := map[string]string{
		"key.one":   "Value One",
		"key.two":   "Value Two",
		"key.three": "Value Three",
	}

	foundKeys := make(map[string]bool)
	for _, msg := range messages {
		if expectedVal, ok := expectedKeys[msg.Key]; ok {
			if msg.DefaultValue != expectedVal {
				t.Errorf("Mismatch value for key %s: expected %s, got %s", msg.Key, expectedVal, msg.DefaultValue)
			}
			foundKeys[msg.Key] = true
		} else {
			t.Errorf("Unexpected key found: %s", msg.Key)
		}
	}

	if len(foundKeys) != 3 {
		t.Errorf("Not all keys found. Found: %v", foundKeys)
	}
}
