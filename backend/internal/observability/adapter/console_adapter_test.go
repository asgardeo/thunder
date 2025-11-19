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
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
)

func TestNewConsoleAdapter(t *testing.T) {
	adp := InitializeConsoleAdapter()
	if adp == nil {
		t.Fatal("InitializeConsoleAdapter() returned nil")
	}
}

func TestConsoleAdapter_GetName(t *testing.T) {
	adapter := InitializeConsoleAdapter()
	name := adapter.GetName()

	if name == "" {
		t.Error("GetName() returned empty string")
	}

	if name != "ConsoleAdapter" {
		t.Errorf("GetName() = %s, want ConsoleAdapter", name)
	}
}

func TestConsoleAdapter_Write(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	adapter := InitializeConsoleAdapter()

	testData := []byte("test event data")
	err := adapter.Write(testData)

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains the test data
	if !bytes.Contains([]byte(output), testData) {
		t.Errorf("Write() output does not contain expected data. Got: %s", output)
	}

	// Verify newline was added
	if output[len(output)-1] != '\n' {
		t.Error("Write() should append newline to output")
	}
}

func TestConsoleAdapter_WriteMultiple(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	adapter := InitializeConsoleAdapter()

	testData := [][]byte{
		[]byte("event 1"),
		[]byte("event 2"),
		[]byte("event 3"),
	}

	for _, data := range testData {
		err := adapter.Write(data)
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
	}

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify all data was written
	for _, data := range testData {
		if !bytes.Contains([]byte(output), data) {
			t.Errorf("Write() output missing data: %s", data)
		}
	}
}

func TestConsoleAdapter_WriteAfterClose(t *testing.T) {
	adapter := InitializeConsoleAdapter()

	// Close the adapter
	err := adapter.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Try to write after close
	testData := []byte("test data")
	err = adapter.Write(testData)

	if err == nil {
		t.Error("Write() should return error after Close()")
	}

	if err.Error() != "console adapter is closed" {
		t.Errorf("Write() error message = %v, want 'console adapter is closed'", err)
	}
}

func TestConsoleAdapter_Flush(t *testing.T) {
	adapter := InitializeConsoleAdapter()

	// Flush should be a no-op and not return error
	err := adapter.Flush()
	if err != nil {
		t.Errorf("Flush() error = %v, want nil", err)
	}
}

func TestConsoleAdapter_Close(t *testing.T) {
	adapter := InitializeConsoleAdapter()

	err := adapter.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Subsequent close should not error
	err = adapter.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestConsoleAdapter_WriteEmpty(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	adapter := InitializeConsoleAdapter()

	// Write empty data
	err := adapter.Write([]byte{})

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Write() with empty data error = %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should still have newline
	if output != "\n" {
		t.Errorf("Write() with empty data should output newline, got: %q", output)
	}
}

func TestConsoleAdapter_WriteLargeData(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	adapter := InitializeConsoleAdapter()

	// Create large data
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte('A' + (i % 26))
	}

	err := adapter.Write(largeData)

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Write() with large data error = %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.Bytes()

	// Verify size (should be largeData + newline)
	if len(output) != len(largeData)+1 {
		t.Errorf("Write() output size = %d, want %d", len(output), len(largeData)+1)
	}
}

func TestConsoleAdapter_ConcurrentWrites(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	adapter := InitializeConsoleAdapter()

	const numGoroutines = 10
	const writesPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				data := []byte(fmt.Sprintf("goroutine-%d-write-%d", id, j))
				err := adapter.Write(data)
				if err != nil {
					t.Errorf("Concurrent Write() error = %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should not be empty
	if len(output) == 0 {
		t.Error("Concurrent writes produced no output")
	}
}

func TestConsoleAdapter_ThreadSafety(t *testing.T) {
	adapter := InitializeConsoleAdapter()

	var wg sync.WaitGroup
	wg.Add(3)

	// Concurrent writes
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = adapter.Write([]byte("write"))
		}
	}()

	// Concurrent flushes
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = adapter.Flush()
		}
	}()

	// Concurrent close (only once at the end)
	go func() {
		defer wg.Done()
		for i := 0; i < 99; i++ {
			// Just do some work
		}
		_ = adapter.Close()
	}()

	wg.Wait()
	// If we reach here without deadlock or panic, thread safety is good
}

func BenchmarkConsoleAdapter_Write(b *testing.B) {
	// Redirect stdout to /dev/null for benchmarking
	oldStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull
	defer func() {
		os.Stdout = oldStdout
		_ = devNull.Close()
	}()

	adapter := InitializeConsoleAdapter()
	testData := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.Write(testData)
	}
}

func BenchmarkConsoleAdapter_WriteLarge(b *testing.B) {
	// Redirect stdout to /dev/null for benchmarking
	oldStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull
	defer func() {
		os.Stdout = oldStdout
		_ = devNull.Close()
	}()

	adapter := InitializeConsoleAdapter()
	largeData := make([]byte, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.Write(largeData)
	}
}
