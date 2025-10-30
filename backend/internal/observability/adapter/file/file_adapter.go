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

// Package file provides a file-based output adapter for analytics events.
package file

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/observability/adapter"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	defaultBufferSize    = 4096
	defaultFlushInterval = 5 * time.Second
	loggerComponentName  = "FileOutputAdapter"
)

// FileAdapter writes events to a file.
type FileAdapter struct {
	filePath    string
	file        *os.File
	writer      *bufio.Writer
	mu          sync.Mutex
	flushTicker *time.Ticker
	stopFlush   chan struct{}
	wg          sync.WaitGroup
	closed      bool
}

var _ adapter.OutputAdapter = (*FileAdapter)(nil)

// NewFileAdapter creates a new file-based output adapter.
func NewFileAdapter(filePath string) (*FileAdapter, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, dirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open file in append mode
	// #nosec G304 -- File path is provided by configuration
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermissions)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	fa := &FileAdapter{
		filePath:    filePath,
		file:        file,
		writer:      bufio.NewWriterSize(file, defaultBufferSize),
		flushTicker: time.NewTicker(defaultFlushInterval),
		stopFlush:   make(chan struct{}),
		closed:      false,
	}

	// Start periodic flushing
	fa.wg.Add(1)
	go fa.periodicFlush()

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Info("File adapter initialized", log.String("filePath", filePath))

	return fa, nil
}

// Write writes data to the file.
func (fa *FileAdapter) Write(data []byte) error {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	if fa.closed {
		return fmt.Errorf("file adapter is closed")
	}

	// Write data
	if _, err := fa.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	// Write newline
	if _, err := fa.writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// Flush flushes buffered data to the file.
func (fa *FileAdapter) Flush() error {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	if fa.closed {
		return nil
	}

	if err := fa.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	if err := fa.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// periodicFlush periodically flushes the buffer to ensure data is persisted.
func (fa *FileAdapter) periodicFlush() {
	defer fa.wg.Done()

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	for {
		select {
		case <-fa.flushTicker.C:
			if err := fa.Flush(); err != nil {
				logger.Error("Failed to flush file", log.Error(err))
			}
		case <-fa.stopFlush:
			return
		}
	}
}

// Close closes the file adapter and releases resources.
func (fa *FileAdapter) Close() error {
	fa.mu.Lock()
	if fa.closed {
		fa.mu.Unlock()
		return nil
	}
	fa.closed = true
	fa.mu.Unlock()

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Info("Closing file adapter", log.String("filePath", fa.filePath))

	// Stop periodic flushing
	fa.flushTicker.Stop()
	close(fa.stopFlush)
	fa.wg.Wait()

	// Final flush
	if err := fa.Flush(); err != nil {
		logger.Error("Failed to perform final flush", log.Error(err))
	}

	// Close file
	if err := fa.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	logger.Info("File adapter closed")
	return nil
}

// GetName returns the name of this adapter.
func (fa *FileAdapter) GetName() string {
	return "FileAdapter"
}
