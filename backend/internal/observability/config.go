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

package observability

import (
	"time"

	"github.com/asgardeo/thunder/internal/observability/event"
)

// Config holds the configuration for the analytics system.
type Config struct {
	// Enabled determines whether analytics is enabled.
	// When false, PublishEvent becomes a no-op.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Output configuration
	Output OutputConfig `yaml:"output" json:"output"`

	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics" json:"metrics"`

	// FailureMode determines how to handle initialization failures.
	// "graceful" - log error and disable analytics
	// "strict" - fail fast and return error
	FailureMode string `yaml:"failure_mode" json:"failure_mode"`
}

// OutputConfig holds output-related configuration.
type OutputConfig struct {
	// Type is the output type: file, console, azure, opensearch
	Type string `yaml:"type" json:"type"`

	// File configuration
	File FileOutputConfig `yaml:"file" json:"file"`

	// OpenSearch configuration
	OpenSearch OpenSearchOutputConfig `yaml:"opensearch" json:"opensearch"`

	// Format is the output format: json, csv, xml
	Format string `yaml:"format" json:"format"`
}

// OpenSearchOutputConfig holds OpenSearch output configuration.
type OpenSearchOutputConfig struct {
	// Endpoint is the OpenSearch REST endpoint (e.g. https://localhost:9200).
	Endpoint string `yaml:"endpoint" json:"endpoint"`

	// Index is the index where events will be stored.
	Index string `yaml:"index" json:"index"`

	// Categories limits which categories are forwarded to OpenSearch. Empty means all categories.
	Categories []event.EventCategory `yaml:"categories" json:"categories"`

	// EventTypes limits which event types are forwarded. Empty means all event types.
	EventTypes []event.EventType `yaml:"event_types" json:"event_types"`

	// Username for HTTP basic authentication.
	Username string `yaml:"username" json:"username"`

	// Password for HTTP basic authentication.
	Password string `yaml:"password" json:"password"`

	// APIKey enables API key authentication when provided.
	APIKey string `yaml:"api_key" json:"api_key"`

	// Headers are additional HTTP headers added to each request.
	Headers map[string]string `yaml:"headers" json:"headers"`

	// UseEventID controls whether the event_id should be used as the document ID.
	UseEventID bool `yaml:"use_event_id" json:"use_event_id"`

	// Timeout specifies the HTTP client timeout.
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

// FileOutputConfig holds file output configuration.
type FileOutputConfig struct {
	// Path is the file path for analytics logs.
	Path string `yaml:"path" json:"path"`

	// BufferSize is the buffer size in bytes for file writes.
	BufferSize int `yaml:"buffer_size" json:"buffer_size"`

	// FlushInterval is how often to flush the buffer.
	FlushInterval time.Duration `yaml:"flush_interval" json:"flush_interval"`

	// MaxFileSize is the maximum file size before rotation (in MB).
	MaxFileSize int `yaml:"max_file_size" json:"max_file_size"`

	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `yaml:"max_backups" json:"max_backups"`

	// MaxAge is the maximum number of days to retain old log files.
	MaxAge int `yaml:"max_age" json:"max_age"`

	// Compress determines whether rotated files should be compressed.
	Compress bool `yaml:"compress" json:"compress"`
}

// MetricsConfig holds metrics-related configuration.
type MetricsConfig struct {
	// Enabled determines whether metrics collection is enabled.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// ExportInterval is how often to export metrics.
	ExportInterval time.Duration `yaml:"export_interval" json:"export_interval"`
}

// DefaultConfig returns the default observability configuration.
// Defaults to console output with JSON format for easier development and debugging.
// For production, configure file output with rotation or external adapters (OpenSearch).
func DefaultConfig() *Config {
	return &Config{
		Enabled: true,
		Output: OutputConfig{
			Type:   "console",
			Format: "json",
		},
		Metrics: MetricsConfig{
			Enabled:        true,
			ExportInterval: 60 * time.Second,
		},
		FailureMode: "graceful",
	}
}
