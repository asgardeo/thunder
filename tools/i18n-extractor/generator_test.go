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
	"strings"
	"testing"
)

func TestGenerator_Generate_Success(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "defaults.go")

	messages := []ExtractedMessage{
		{Key: "key.a", DefaultValue: "Value A", SourceFile: "a.go", Line: 10},
		{Key: "key.b", DefaultValue: "Value B", SourceFile: "b.go", Line: 20},
	}

	generator := NewGenerator()
	err := generator.Generate(messages, outputFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, `"key.a": "Value A"`) {
		t.Error("Generated file missing key.a")
	}
	if !strings.Contains(strContent, `"key.b": "Value B"`) {
		t.Error("Generated file missing key.b")
	}
}

func TestGenerator_Generate_DuplicateKeys_SameFile(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "defaults.go")

	messages := []ExtractedMessage{
		{Key: "dup.key", DefaultValue: "Val 1", SourceFile: "file.go", Line: 10},
		{Key: "dup.key", DefaultValue: "Val 2", SourceFile: "file.go", Line: 20},
	}

	generator := NewGenerator()
	err := generator.Generate(messages, outputFile)
	if err == nil {
		t.Fatal("Expected error for duplicate keys, got nil")
	}

	expectedErr := `duplicate keys found:
Key "dup.key" is defined 2 times:
	- file.go:10
	- file.go:20`

	if err.Error() != expectedErr {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expectedErr, err.Error())
	}
}

func TestGenerator_Generate_DuplicateKeys_DifferentFiles(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "defaults.go")

	messages := []ExtractedMessage{
		{Key: "shared.key", DefaultValue: "Val 1", SourceFile: "file1.go", Line: 10},
		{Key: "shared.key", DefaultValue: "Val 2", SourceFile: "file2.go", Line: 15},
	}

	generator := NewGenerator()
	err := generator.Generate(messages, outputFile)
	if err == nil {
		t.Fatal("Expected error for duplicate keys, got nil")
	}

	expectedErrPart := `Key "shared.key" is defined 2 times:`
	if !strings.Contains(err.Error(), expectedErrPart) {
		t.Errorf("Error message does not contain expected part: %s. Got: %s", expectedErrPart, err.Error())
	}
	if !strings.Contains(err.Error(), "file1.go:10") {
		t.Errorf("Error missing file1 location")
	}
	if !strings.Contains(err.Error(), "file2.go:15") {
		t.Errorf("Error missing file2 location")
	}
}

func TestGenerator_Generate_MultipleDuplicates(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "defaults.go")

	messages := []ExtractedMessage{
		{Key: "dup1", DefaultValue: "V1", SourceFile: "f1.go", Line: 1},
		{Key: "dup1", DefaultValue: "V1", SourceFile: "f1.go", Line: 2},
		{Key: "dup2", DefaultValue: "V2", SourceFile: "f2.go", Line: 1},
		{Key: "dup2", DefaultValue: "V2", SourceFile: "f2.go", Line: 2},
	}

	generator := NewGenerator()
	err := generator.Generate(messages, outputFile)
	if err == nil {
		t.Fatal("Expected error for duplicate keys, got nil")
	}

	if !strings.Contains(err.Error(), `Key "dup1" is defined 2 times`) {
		t.Errorf("Error missing dup1 info")
	}
	if !strings.Contains(err.Error(), `Key "dup2" is defined 2 times`) {
		t.Errorf("Error missing dup2 info")
	}
}
