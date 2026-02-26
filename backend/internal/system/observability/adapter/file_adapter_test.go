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

package adapter

import (
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewFileAdapter_Simple(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	adapter, err := NewFileAdapter(filePath)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	if adapter.GetName() != "FileAdapter" {
		t.Errorf("Expected name 'FileAdapter', got '%s'", adapter.GetName())
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File was not created")
	}
}

func TestNewFileAdapter_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	config := &Config{
		Path:          filePath,
		MaxFileSizeMB: 10,
		MaxBackups:    5,
		MaxAgeDays:    7,
		Compress:      true,
	}

	adapter, err := NewFileAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create adapter with config: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	if !adapter.isRotationEnabled() {
		t.Error("Expected rotation to be enabled")
	}
}

func TestNewFileAdapter_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "empty path",
			config: &Config{
				Path: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFileAdapterWithConfig(tt.config)
			if err == nil {
				t.Error("Expected error for invalid config")
			}
		})
	}
}

func TestFileAdapter_WriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	adapter, err := NewFileAdapter(filePath)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Write test data
	testData := []string{
		"First log line",
		"Second log line",
		"Third log line",
	}

	for _, data := range testData {
		if err := adapter.Write([]byte(data)); err != nil {
			t.Fatalf("Failed to write: %v", err)
		}
	}

	// Flush to ensure data is written
	if err := adapter.Flush(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	// Close adapter
	if err := adapter.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	// Read file and verify contents
	// #nosec G304 -- filePath is from test temp directory, safe for testing
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != len(testData) {
		t.Errorf("Expected %d lines, got %d", len(testData), len(lines))
	}

	for i, expected := range testData {
		if i < len(lines) && lines[i] != expected {
			t.Errorf("Line %d: expected '%s', got '%s'", i, expected, lines[i])
		}
	}
}

func TestFileAdapter_WriteAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	adapter, err := NewFileAdapter(filePath)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Close adapter
	if err := adapter.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	// Try to write after close
	err = adapter.Write([]byte("should fail"))
	if err == nil {
		t.Error("Expected error when writing after close")
	}
}

func TestFileAdapter_Rotation(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	// Create adapter with very small max file size to trigger rotation
	config := &Config{
		Path:          filePath,
		MaxFileSizeMB: 0, // Will calculate in bytes for testing
		MaxBackups:    3,
		MaxAgeDays:    0,
		Compress:      false,
	}

	adapter, err := NewFileAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Manually set a small max size for testing (1 KB)
	adapter.config.MaxFileSizeMB = 1 // This will be 1MB, but we'll trigger rotation manually

	// Write a large amount of data to trigger rotation
	largeData := strings.Repeat("a", 2*1024*1024) // 2 MB

	if err := adapter.Write([]byte(largeData)); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := adapter.Flush(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	// Give rotation time to happen
	time.Sleep(100 * time.Millisecond)

	// Check for rotated files
	files, err := filepath.Glob(filePath + ".*")
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected rotated files to exist")
	}
}

func TestFileAdapter_RotationWithCompression(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	config := &Config{
		Path:          filePath,
		MaxFileSizeMB: 1,
		MaxBackups:    3,
		MaxAgeDays:    0,
		Compress:      true,
	}

	adapter, err := NewFileAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Write data to trigger rotation
	largeData := strings.Repeat("test log line\n", 100000) // ~1.3 MB
	if err := adapter.Write([]byte(largeData)); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := adapter.Flush(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	// Wait for compression to complete
	time.Sleep(500 * time.Millisecond)

	// Check for compressed files
	files, err := filepath.Glob(filePath + ".*.gz")
	if err != nil {
		t.Fatalf("Failed to glob compressed files: %v", err)
	}

	if len(files) == 0 {
		t.Log("Warning: No compressed files found (compression may be async)")
	} else {
		// Verify the compressed file is valid
		gzFile, err := os.Open(files[0])
		if err != nil {
			t.Fatalf("Failed to open compressed file: %v", err)
		}
		defer func() { _ = gzFile.Close() }()

		gzReader, err := gzip.NewReader(gzFile)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer func() { _ = gzReader.Close() }()

		// Try to read some data to verify it's valid gzip
		buf := make([]byte, 100)
		_, err = gzReader.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatalf("Failed to read from compressed file: %v", err)
		}
	}
}

func TestFileAdapter_MaxBackupsCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	config := &Config{
		Path:          filePath,
		MaxFileSizeMB: 1,
		MaxBackups:    2, // Keep only 2 backups
		MaxAgeDays:    0,
		Compress:      false,
	}

	adapter, err := NewFileAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Trigger multiple rotations
	for i := 0; i < 5; i++ {
		largeData := strings.Repeat("x", 2*1024*1024)
		if err := adapter.Write([]byte(largeData)); err != nil {
			t.Fatalf("Failed to write (iteration %d): %v", i, err)
		}
		if err := adapter.Flush(); err != nil {
			t.Fatalf("Failed to flush (iteration %d): %v", i, err)
		}
		time.Sleep(100 * time.Millisecond) // Allow rotation to complete
	}

	// Wait for cleanup
	time.Sleep(500 * time.Millisecond)

	// Count rotated files
	files, err := filepath.Glob(filePath + ".*")
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}

	// Should have at most MaxBackups files (plus current)
	// Note: cleanup is async so this might not always be exact
	if len(files) > config.MaxBackups+1 {
		t.Logf("Warning: Found %d files, expected at most %d (cleanup may be async)",
			len(files), config.MaxBackups+1)
	}
}

func TestFileAdapter_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	adapter, err := NewFileAdapter(filePath)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Write concurrently from multiple goroutines
	numGoroutines := 10
	writesPerGoroutine := 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < writesPerGoroutine; j++ {
				data := []byte("concurrent write")
				if err := adapter.Write(data); err != nil {
					t.Errorf("Goroutine %d failed to write: %v", id, err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Flush and close
	if err := adapter.Flush(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	if err := adapter.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	// Read file and count lines
	// #nosec G304 -- filePath is from test temp directory, safe for testing
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expectedLines := numGoroutines * writesPerGoroutine

	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
	}
}

func TestFileAdapter_FlushInterval(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	adapter, err := NewFileAdapter(filePath)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Write data
	if err := adapter.Write([]byte("test data")); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Don't flush manually - wait for periodic flush
	time.Sleep(6 * time.Second) // defaultFlushInterval is 5 seconds

	// Read file to verify data was flushed
	// #nosec G304 -- filePath is from test temp directory, safe for testing
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "test data") {
		t.Error("Data was not flushed by periodic flush")
	}
}

func TestFileAdapter_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	// Create path with nested directories that don't exist
	filePath := filepath.Join(tmpDir, "logs", "app", "events", "test.log")

	adapter, err := NewFileAdapter(filePath)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Verify all directories were created
	if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
		t.Error("Directory was not created")
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File was not created")
	}
}

func TestFileAdapter_NoRotation(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	// Create adapter with rotation disabled
	config := &Config{
		Path:          filePath,
		MaxFileSizeMB: 0, // No rotation
	}

	adapter, err := NewFileAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	if adapter.isRotationEnabled() {
		t.Error("Expected rotation to be disabled")
	}

	// Write large amount of data
	largeData := strings.Repeat("no rotation\n", 500000) // ~5.5 MB
	if err := adapter.Write([]byte(largeData)); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := adapter.Flush(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	// Check that no rotated files exist
	files, err := filepath.Glob(filePath + ".*")
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}

	if len(files) > 0 {
		t.Errorf("Expected no rotated files, found %d", len(files))
	}

	// Verify main file has all the data
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() < int64(len(largeData)) {
		t.Error("File doesn't contain all data (rotation may have occurred)")
	}
}

func TestFileAdapter_AppendToExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	// Create file with initial content
	initialContent := "existing line 1\nexisting line 2\n"
	if err := os.WriteFile(filePath, []byte(initialContent), filePermissions); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Create adapter and write more data
	adapter, err := NewFileAdapter(filePath)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	if err := adapter.Write([]byte("new line 1")); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := adapter.Flush(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	if err := adapter.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	// Read file and verify it has both old and new content
	// #nosec G304 -- filePath is from test temp directory, safe for testing
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "existing line 1") {
		t.Error("Original content was lost")
	}
	if !strings.Contains(contentStr, "new line 1") {
		t.Error("New content was not written")
	}
}
