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

// Package file provides file-based output adapters for analytics events.
package file

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/observability/adapter"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	defaultMaxFileSize          = 100 * 1024 * 1024 // 100 MB
	defaultMaxBackups           = 10
	defaultMaxAge               = 30 // days
	megabyte                    = 1024 * 1024
	rotatingLoggerComponentName = "RotatingFileAdapter"
	dirPermissions              = 0750
	filePermissions             = 0600
)

// RotatingFileAdapter writes events to a file with rotation support.
type RotatingFileAdapter struct {
	filePath    string
	maxFileSize int64 // in bytes
	maxBackups  int
	maxAge      int // in days
	compress    bool
	file        *os.File
	currentSize int64
	mu          sync.Mutex
	flushTicker *time.Ticker
	stopFlush   chan struct{}
	wg          sync.WaitGroup
	closed      bool
}

var _ adapter.OutputAdapter = (*RotatingFileAdapter)(nil)

// RotatingFileConfig holds configuration for rotating file adapter.
type RotatingFileConfig struct {
	FilePath      string
	MaxFileSizeMB int // in megabytes
	MaxBackups    int
	MaxAgeDays    int
	Compress      bool
}

// NewRotatingFileAdapter creates a new rotating file adapter.
func NewRotatingFileAdapter(config *RotatingFileConfig) (*RotatingFileAdapter, error) {
	if config == nil {
		config = &RotatingFileConfig{
			MaxFileSizeMB: 100,
			MaxBackups:    defaultMaxBackups,
			MaxAgeDays:    defaultMaxAge,
			Compress:      true,
		}
	}

	// Set defaults
	if config.MaxFileSizeMB <= 0 {
		config.MaxFileSizeMB = 100
	}
	if config.MaxBackups < 0 {
		config.MaxBackups = defaultMaxBackups
	}
	if config.MaxAgeDays < 0 {
		config.MaxAgeDays = defaultMaxAge
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, dirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open file in append mode
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermissions)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", config.FilePath, err)
	}

	// Get current file size
	fileInfo, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	rfa := &RotatingFileAdapter{
		filePath:    config.FilePath,
		maxFileSize: int64(config.MaxFileSizeMB) * megabyte,
		maxBackups:  config.MaxBackups,
		maxAge:      config.MaxAgeDays,
		compress:    config.Compress,
		file:        file,
		currentSize: fileInfo.Size(),
		flushTicker: time.NewTicker(defaultFlushInterval),
		stopFlush:   make(chan struct{}),
		closed:      false,
	}

	// Start periodic flushing
	rfa.wg.Add(1)
	go rfa.periodicFlush()

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, rotatingLoggerComponentName))
	logger.Info("Rotating file adapter initialized",
		log.String("filePath", config.FilePath),
		log.Int("maxFileSizeMB", config.MaxFileSizeMB),
		log.Int("maxBackups", config.MaxBackups),
		log.Int("maxAgeDays", config.MaxAgeDays),
		log.Bool("compress", config.Compress))

	return rfa, nil
}

// Write writes data to the file, rotating if necessary.
func (rfa *RotatingFileAdapter) Write(data []byte) error {
	rfa.mu.Lock()
	defer rfa.mu.Unlock()

	if rfa.closed {
		return fmt.Errorf("rotating file adapter is closed")
	}

	// Check if rotation is needed
	dataSize := int64(len(data)) + 1 // +1 for newline
	if rfa.currentSize+dataSize > rfa.maxFileSize {
		if err := rfa.rotate(); err != nil {
			logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, rotatingLoggerComponentName))
			logger.Error("Failed to rotate file", log.Error(err))
			// Continue writing to current file even if rotation fails
		}
	}

	// Write data
	n, err := rfa.file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	rfa.currentSize += int64(n)

	// Write newline
	n, err = rfa.file.WriteString("\n")
	if err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}
	rfa.currentSize += int64(n)

	return nil
}

// rotate rotates the log file.
func (rfa *RotatingFileAdapter) rotate() error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, rotatingLoggerComponentName))

	// Close current file
	if err := rfa.file.Close(); err != nil {
		logger.Error("Failed to close file during rotation", log.Error(err))
	}

	// Generate rotated filename with timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	rotatedPath := rfa.filePath + "." + timestamp

	// Rename current file
	if err := os.Rename(rfa.filePath, rotatedPath); err != nil {
		logger.Error("Failed to rename file during rotation", log.Error(err))
		// Try to reopen original file
		file, err := os.OpenFile(rfa.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermissions)
		if err != nil {
			return fmt.Errorf("failed to reopen file after rotation error: %w", err)
		}
		rfa.file = file
		rfa.currentSize = 0
		return fmt.Errorf("failed to rename file: %w", err)
	}

	// Compress rotated file if enabled
	if rfa.compress {
		go rfa.compressFile(rotatedPath)
	}

	// Create new file
	file, err := os.OpenFile(rfa.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermissions)
	if err != nil {
		return fmt.Errorf("failed to create new file after rotation: %w", err)
	}

	rfa.file = file
	rfa.currentSize = 0

	logger.Info("File rotated successfully", log.String("rotatedFile", rotatedPath))

	// Clean up old files
	go rfa.cleanup()

	return nil
}

// compressFile compresses a log file using gzip.
func (rfa *RotatingFileAdapter) compressFile(filePath string) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, rotatingLoggerComponentName))

	// Open source file
	src, err := os.Open(filePath) // #nosec G304 -- File path is controlled internally by rotation logic
	if err != nil {
		logger.Error("Failed to open file for compression", log.String("filePath", filePath), log.Error(err))
		return
	}
	defer func() {
		if closeErr := src.Close(); closeErr != nil {
			logger.Error("Failed to close source file", log.Error(closeErr))
		}
	}()

	// Create compressed file
	compressedPath := filePath + ".gz"
	dst, err := os.Create(compressedPath) // #nosec G304 -- File path is derived from controlled internal path
	if err != nil {
		logger.Error("Failed to create compressed file", log.String("compressedPath", compressedPath), log.Error(err))
		return
	}
	defer func() {
		if closeErr := dst.Close(); closeErr != nil {
			logger.Error("Failed to close destination file", log.Error(closeErr))
		}
	}()

	// Create gzip writer
	gzWriter := gzip.NewWriter(dst)
	defer func() {
		if closeErr := gzWriter.Close(); closeErr != nil {
			logger.Error("Failed to close gzip writer", log.Error(closeErr))
		}
	}()

	// Copy and compress
	if _, err := io.Copy(gzWriter, src); err != nil {
		logger.Error("Failed to compress file", log.String("filePath", filePath), log.Error(err))
		return
	}

	// Close gzip writer to flush
	if err := gzWriter.Close(); err != nil {
		logger.Error("Failed to close gzip writer", log.Error(err))
		return
	}

	// Remove original file
	if err := os.Remove(filePath); err != nil {
		logger.Error("Failed to remove original file after compression", log.String("filePath", filePath), log.Error(err))
		return
	}

	logger.Info("File compressed successfully", log.String("compressedPath", compressedPath))
}

// cleanup removes old log files based on maxBackups and maxAge.
func (rfa *RotatingFileAdapter) cleanup() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, rotatingLoggerComponentName))

	dir := filepath.Dir(rfa.filePath)
	baseName := filepath.Base(rfa.filePath)

	// Find all rotated log files
	files, err := filepath.Glob(filepath.Join(dir, baseName+".*"))
	if err != nil {
		logger.Error("Failed to list rotated files", log.Error(err))
		return
	}

	// Sort files by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		fi, err1 := os.Stat(files[i])
		fj, err2 := os.Stat(files[j])
		if err1 != nil || err2 != nil {
			return false
		}
		return fi.ModTime().After(fj.ModTime())
	})

	now := time.Now()
	removed := 0

	for i, file := range files {
		// Skip if it's the current file
		if file == rfa.filePath {
			continue
		}

		shouldRemove := false

		// Check maxBackups
		if rfa.maxBackups > 0 && i >= rfa.maxBackups {
			shouldRemove = true
		}

		// Check maxAge
		if rfa.maxAge > 0 {
			fileInfo, err := os.Stat(file)
			if err == nil {
				age := now.Sub(fileInfo.ModTime())
				if age > time.Duration(rfa.maxAge)*24*time.Hour {
					shouldRemove = true
				}
			}
		}

		if shouldRemove {
			if err := os.Remove(file); err != nil {
				logger.Error("Failed to remove old log file", log.String("filePath", file), log.Error(err))
			} else {
				removed++
			}
		}
	}

	if removed > 0 {
		logger.Info("Cleaned up old log files", log.Int("removedCount", removed))
	}
}

// Flush flushes buffered data to the file.
func (rfa *RotatingFileAdapter) Flush() error {
	rfa.mu.Lock()
	defer rfa.mu.Unlock()

	if rfa.closed {
		return nil
	}

	if err := rfa.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// periodicFlush periodically flushes the buffer.
func (rfa *RotatingFileAdapter) periodicFlush() {
	defer rfa.wg.Done()

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, rotatingLoggerComponentName))

	for {
		select {
		case <-rfa.flushTicker.C:
			if err := rfa.Flush(); err != nil {
				logger.Error("Failed to flush file", log.Error(err))
			}
		case <-rfa.stopFlush:
			return
		}
	}
}

// Close closes the rotating file adapter and releases resources.
func (rfa *RotatingFileAdapter) Close() error {
	rfa.mu.Lock()
	if rfa.closed {
		rfa.mu.Unlock()
		return nil
	}
	rfa.closed = true
	rfa.mu.Unlock()

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, rotatingLoggerComponentName))
	logger.Info("Closing rotating file adapter", log.String("filePath", rfa.filePath))

	// Stop periodic flushing
	rfa.flushTicker.Stop()
	close(rfa.stopFlush)
	rfa.wg.Wait()

	// Final flush
	if err := rfa.Flush(); err != nil {
		logger.Error("Failed to perform final flush", log.Error(err))
	}

	// Close file
	if err := rfa.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	logger.Info("Rotating file adapter closed")
	return nil
}

// GetName returns the name of this adapter.
func (rfa *RotatingFileAdapter) GetName() string {
	return "RotatingFileAdapter"
}

// GetCurrentFileSize returns the current file size in bytes.
func (rfa *RotatingFileAdapter) GetCurrentFileSize() int64 {
	rfa.mu.Lock()
	defer rfa.mu.Unlock()
	return rfa.currentSize
}

// GetRotatedFiles returns a list of rotated log files.
func (rfa *RotatingFileAdapter) GetRotatedFiles() ([]string, error) {
	dir := filepath.Dir(rfa.filePath)
	baseName := filepath.Base(rfa.filePath)

	files, err := filepath.Glob(filepath.Join(dir, baseName+".*"))
	if err != nil {
		return nil, err
	}

	// Filter out non-rotated files
	var rotatedFiles []string
	for _, file := range files {
		if file != rfa.filePath && !strings.HasSuffix(file, ".tmp") {
			rotatedFiles = append(rotatedFiles, file)
		}
	}

	return rotatedFiles, nil
}
