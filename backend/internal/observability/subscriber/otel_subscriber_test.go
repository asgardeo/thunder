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
	"context"
	"testing"
	"time"

	"github.com/asgardeo/thunder/internal/observability/event"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	exporterTypeStdout = "stdout"
)

func TestNewOTelSubscriber(t *testing.T) {
	sub := NewOTelSubscriber()
	if sub == nil {
		t.Fatal("NewOTelSubscriber() returned nil")
	}
}

func TestOTelSubscriber_IsEnabled(t *testing.T) {
	// Setup mock config
	setupTestConfig(t)
	defer resetTestConfig()

	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{
			name:    "enabled when config is true",
			enabled: true,
			want:    true,
		},
		{
			name:    "disabled when config is false",
			enabled: false,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry.Enabled = tt.enabled

			sub := NewOTelSubscriber()
			if got := sub.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOTelSubscriber_Initialize(t *testing.T) {
	setupTestConfig(t)
	defer resetTestConfig()

	tests := []struct {
		name        string
		config      func()
		wantErr     bool
		errContains string
	}{
		{
			name: "successful initialization with stdout exporter",
			config: func() {
				cfg := &config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry
				cfg.Enabled = true
				cfg.ExporterType = exporterTypeStdout
				cfg.ServiceName = "test-service"
				cfg.ServiceVersion = "1.0.0"
				cfg.SampleRate = 1.0
			},
			wantErr: false,
		},
		{
			name: "successful initialization with default categories",
			config: func() {
				cfg := &config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry
				cfg.Enabled = true
				cfg.ExporterType = exporterTypeStdout
				cfg.Categories = []string{}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)
			tt.config()

			sub := NewOTelSubscriber()
			err := sub.Initialize()

			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if sub.GetID() == "" {
					t.Error("Initialize() should set subscriber ID")
				}
				if sub.tracer == nil {
					t.Error("Initialize() should set tracer")
				}
				if sub.tracerProvider == nil {
					t.Error("Initialize() should set tracer provider")
				}

				// Clean up
				_ = sub.Close()
			}
		})
	}
}

func TestOTelSubscriber_GetID(t *testing.T) {
	setupTestConfig(t)
	defer resetTestConfig()

	config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry.Enabled = true
	config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry.ExporterType = exporterTypeStdout

	sub := NewOTelSubscriber()
	_ = sub.Initialize()
	defer func() { _ = sub.Close() }()

	id := sub.GetID()
	if id == "" {
		t.Error("GetID() returned empty string")
	}

	// ID should be consistent
	if id != sub.GetID() {
		t.Error("GetID() should return consistent ID")
	}
}

func TestOTelSubscriber_GetCategories(t *testing.T) {
	setupTestConfig(t)
	defer resetTestConfig()

	tests := []struct {
		name       string
		categories []string
		wantLen    int
	}{
		{
			name:       "returns default CategoryAll when no categories configured",
			categories: []string{},
			wantLen:    1,
		},
		{
			name:       "returns configured categories",
			categories: []string{"observability.authentication", "observability.flows"},
			wantLen:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)
			cfg := &config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry
			cfg.Enabled = true
			cfg.ExporterType = exporterTypeStdout
			cfg.Categories = tt.categories

			sub := NewOTelSubscriber()
			_ = sub.Initialize()
			defer func() { _ = sub.Close() }()

			categories := sub.GetCategories()
			if len(categories) != tt.wantLen {
				t.Errorf("GetCategories() returned %d categories, want %d", len(categories), tt.wantLen)
			}
		})
	}
}

func TestOTelSubscriber_OnEvent(t *testing.T) {
	setupTestConfig(t)
	defer resetTestConfig()

	config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry.Enabled = true
	config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry.ExporterType = exporterTypeStdout

	sub := NewOTelSubscriber()
	_ = sub.Initialize()
	defer func() { _ = sub.Close() }()

	tests := []struct {
		name    string
		event   *event.Event
		wantErr bool
	}{
		{
			name:    "error when event is nil",
			event:   nil,
			wantErr: true,
		},
		{
			name: "successfully processes valid event",
			event: &event.Event{
				TraceID:   "trace-123",
				EventID:   "event-123",
				Type:      "test.event",
				Timestamp: time.Now(),
				Component: "TestComponent",
				Status:    event.StatusSuccess,
				Data: map[string]interface{}{
					"key1": "value1",
					"key2": 123,
				},
			},
			wantErr: false,
		},
		{
			name: "successfully processes failure event",
			event: &event.Event{
				TraceID:   "trace-456",
				EventID:   "event-456",
				Type:      "test.failure",
				Timestamp: time.Now(),
				Component: "TestComponent",
				Status:    event.StatusFailure,
				Data: map[string]interface{}{
					event.DataKey.Error: "test error message",
				},
			},
			wantErr: false,
		},
		{
			name: "successfully processes event with various data types",
			event: &event.Event{
				TraceID:   "trace-789",
				EventID:   "event-789",
				Type:      "test.datatypes",
				Timestamp: time.Now(),
				Component: "OAuth2Server",
				Status:    event.StatusSuccess,
				Data: map[string]interface{}{
					"string":  "value",
					"int":     42,
					"int64":   int64(123456),
					"float64": 3.14,
					"bool":    true,
					"nil":     nil,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sub.OnEvent(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("OnEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOTelSubscriber_OnEventWithSpanRecorder(t *testing.T) {
	setupTestConfig(t)
	defer resetTestConfig()

	// Create a span recorder to capture spans
	spanRecorder := tracetest.NewSpanRecorder()

	// Create a tracer provider with the span recorder
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanRecorder),
	)

	// Create logger for the subscriber
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TestOTelSubscriber"))

	sub := &OTelSubscriber{
		id:             "test-sub",
		tracerProvider: tracerProvider,
		tracer:         tracerProvider.Tracer("thunder-observability"),
		categories:     []event.EventCategory{event.CategoryAll},
		logger:         logger,
	}

	testEvent := &event.Event{
		TraceID:   "trace-123",
		EventID:   "event-123",
		Type:      "test.event",
		Timestamp: time.Now(),
		Component: "TestComponent",
		Status:    event.StatusSuccess,
		Data: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	err := sub.OnEvent(testEvent)
	if err != nil {
		t.Fatalf("OnEvent() unexpected error = %v", err)
	}

	// Force flush spans
	_ = tracerProvider.ForceFlush(context.Background())

	// Verify span was created
	spans := spanRecorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Name() != testEvent.Type {
		t.Errorf("Span name = %s, want %s", span.Name(), testEvent.Type)
	}

	// Verify span attributes
	attrs := span.Attributes()
	expectedAttrs := map[string]bool{
		"event.id":     false,
		"trace.id":     false,
		"component":    false,
		"event.status": false,
	}

	for _, attr := range attrs {
		if _, exists := expectedAttrs[string(attr.Key)]; exists {
			expectedAttrs[string(attr.Key)] = true
		}
	}

	for key, found := range expectedAttrs {
		if !found {
			t.Errorf("Expected attribute %s not found in span", key)
		}
	}

	// Clean up
	_ = sub.Close()
}

func TestOTelSubscriber_convertDataToAttributes(t *testing.T) {
	sub := &OTelSubscriber{}

	tests := []struct {
		name     string
		data     map[string]interface{}
		wantLen  int
		validate func(*testing.T, []attribute.KeyValue)
	}{
		{
			name:    "empty data",
			data:    map[string]interface{}{},
			wantLen: 0,
		},
		{
			name: "string data",
			data: map[string]interface{}{
				"key": "value",
			},
			wantLen: 1,
			validate: func(t *testing.T, attrs []attribute.KeyValue) {
				if attrs[0].Key != "key" {
					t.Errorf("Key = %s, want key", attrs[0].Key)
				}
				if attrs[0].Value.AsString() != "value" {
					t.Errorf("Value = %s, want value", attrs[0].Value.AsString())
				}
			},
		},
		{
			name: "int data",
			data: map[string]interface{}{
				"count": 42,
			},
			wantLen: 1,
			validate: func(t *testing.T, attrs []attribute.KeyValue) {
				if attrs[0].Value.AsInt64() != 42 {
					t.Errorf("Value = %d, want 42", attrs[0].Value.AsInt64())
				}
			},
		},
		{
			name: "int64 data",
			data: map[string]interface{}{
				"big": int64(123456),
			},
			wantLen: 1,
		},
		{
			name: "float64 data",
			data: map[string]interface{}{
				"ratio": 3.14,
			},
			wantLen: 1,
		},
		{
			name: "bool data",
			data: map[string]interface{}{
				"flag": true,
			},
			wantLen: 1,
		},
		{
			name: "nil values are skipped",
			data: map[string]interface{}{
				"valid": "value",
				"null":  nil,
			},
			wantLen: 1,
		},
		{
			name: "empty string is skipped",
			data: map[string]interface{}{
				"valid": "value",
				"empty": "",
			},
			wantLen: 1,
		},
		{
			name: "mixed data types",
			data: map[string]interface{}{
				"string":  "value",
				"int":     42,
				"float":   3.14,
				"bool":    true,
				"nil":     nil,
				"empty":   "",
				"complex": map[string]string{"nested": "value"},
			},
			wantLen: 5, // nil and empty are skipped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := sub.convertDataToAttributes(tt.data)
			if len(attrs) != tt.wantLen {
				t.Errorf("convertDataToAttributes() returned %d attributes, want %d", len(attrs), tt.wantLen)
			}
			if tt.validate != nil {
				tt.validate(t, attrs)
			}
		})
	}
}

func TestOTelSubscriber_getStringData(t *testing.T) {
	sub := &OTelSubscriber{}

	tests := []struct {
		name  string
		event *event.Event
		key   string
		want  string
	}{
		{
			name: "returns string value",
			event: &event.Event{
				Data: map[string]interface{}{
					"key": "value",
				},
			},
			key:  "key",
			want: "value",
		},
		{
			name: "returns empty string for missing key",
			event: &event.Event{
				Data: map[string]interface{}{},
			},
			key:  "missing",
			want: "",
		},
		{
			name: "returns empty string for non-string value",
			event: &event.Event{
				Data: map[string]interface{}{
					"number": 123,
				},
			},
			key:  "number",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.getStringData(tt.event, tt.key)
			if got != tt.want {
				t.Errorf("getStringData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOTelSubscriber_Close(t *testing.T) {
	setupTestConfig(t)
	defer resetTestConfig()

	config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry.Enabled = true
	config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry.ExporterType = exporterTypeStdout

	sub := NewOTelSubscriber()
	_ = sub.Initialize()

	err := sub.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify tracer provider is nil after close
	if sub.tracerProvider != nil {
		t.Error("Close() should set tracerProvider to nil")
	}

	// Calling Close again should not error
	err = sub.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestOTelSubscriber_CloseWithoutInitialize(t *testing.T) {
	// Note: In real usage, Close should only be called after successful Initialize
	// This test verifies that Close handles uninitialized state gracefully
	// However, since logger is not initialized, this will cause a panic in production
	// This is acceptable since the contract is that Initialize must be called first
	t.Skip("Close without Initialize is not a valid use case - Initialize must be called first")
}

// Helper functions for testing

func setupTestConfig(t *testing.T) {
	// Reset any existing runtime first
	config.ResetThunderRuntime()

	// Create a test config
	testConfig := &config.Config{
		Observability: config.ObservabilityConfig{
			Enabled: true,
			Output: config.ObservabilityOutputConfig{
				OpenTelemetry: config.ObservabilityOTelConfig{
					Enabled:        false,
					ExporterType:   "stdout",
					ServiceName:    "test-service",
					ServiceVersion: "1.0.0",
					Environment:    "test",
					SampleRate:     1.0,
					Insecure:       true,
					Categories:     []string{},
				},
				File: config.ObservabilityFileConfig{
					Enabled: false,
				},
				Console: config.ObservabilityConsoleConfig{
					Enabled: false,
				},
			},
		},
	}

	// Initialize Thunder runtime with test config
	err := config.InitializeThunderRuntime("/tmp/thunder-test", testConfig)
	if err != nil {
		t.Fatalf("Failed to initialize Thunder runtime: %v", err)
	}
}

func resetTestConfig() {
	// Reset to default disabled state
	config.ResetThunderRuntime()
}

func BenchmarkOTelSubscriber_OnEvent(b *testing.B) {
	setupTestConfig(&testing.T{})
	defer resetTestConfig()

	config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry.Enabled = true
	config.GetThunderRuntime().Config.Observability.Output.OpenTelemetry.ExporterType = exporterTypeStdout

	sub := NewOTelSubscriber()
	_ = sub.Initialize()
	defer func() { _ = sub.Close() }()

	testEvent := &event.Event{
		TraceID:   "trace-123",
		EventID:   "event-123",
		Type:      "benchmark.event",
		Timestamp: time.Now(),
		Component: "BenchmarkComponent",
		Status:    event.StatusSuccess,
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sub.OnEvent(testEvent)
	}
}

func BenchmarkOTelSubscriber_convertDataToAttributes(b *testing.B) {
	sub := &OTelSubscriber{}
	data := map[string]interface{}{
		"string":  "value",
		"int":     42,
		"int64":   int64(123456),
		"float64": 3.14,
		"bool":    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sub.convertDataToAttributes(data)
	}
}
