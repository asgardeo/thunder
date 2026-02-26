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

// Package opentelemetry provides OpenTelemetry initialization and configuration.
package opentelemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Config holds OpenTelemetry configuration.
type Config struct {
	Enabled        bool    `json:"enabled"`
	ExporterType   string  `json:"exporter_type"`   // "otlp", "stdout"
	OTLPEndpoint   string  `json:"otlp_endpoint"`   // e.g., "localhost:4317"
	ServiceName    string  `json:"service_name"`    // e.g., "thunder-iam"
	ServiceVersion string  `json:"service_version"` // e.g., "1.0.0"
	Environment    string  `json:"environment"`     // e.g., "production", "development"
	SampleRate     float64 `json:"sample_rate"`     // 0.0 to 1.0 (1.0 = sample all traces)
	Insecure       bool    `json:"insecure"`        // Set to true to disable TLS (not recommended for production)
}

// newTracerProvider creates and configures an OpenTelemetry TracerProvider.
// This is a package-private constructor. Use Initialize() instead.
// This is based on your working sample code pattern.
func newTracerProvider(ctx context.Context, cfg Config) (*sdktrace.TracerProvider, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("OpenTelemetry is disabled")
	}

	// Set defaults
	if cfg.ServiceName == "" {
		cfg.ServiceName = "thunder-iam"
	}
	if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "1.0.0"
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.SampleRate == 0 {
		cfg.SampleRate = 1.0 // Sample all traces by default
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter based on configuration
	var exporter sdktrace.SpanExporter
	switch cfg.ExporterType {
	case "otlp":
		exporter, err = createOTLPExporter(ctx, cfg)
	case "stdout":
		exporter, err = createStdoutExporter()
	default:
		return nil, fmt.Errorf("unsupported exporter type: %s (supported: otlp, stdout)", cfg.ExporterType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create sampler based on sample rate
	var sampler sdktrace.Sampler
	if cfg.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SampleRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	// Create tracer provider with batch span processor (like your sample)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(1*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithSampler(sampler),
	)

	// Set as global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Set up trace context propagation (like your sample)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider, nil
}

// createOTLPExporter creates an OTLP gRPC exporter.
func createOTLPExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	if cfg.OTLPEndpoint == "" {
		return nil, fmt.Errorf("OTLP endpoint is required when using otlp exporter")
	}

	// Configure TLS based on the Insecure setting
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
	}

	if cfg.Insecure {
		// Disable TLS for development/testing
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	// If Insecure is false, the default behavior is to use TLS with system certificates

	return otlptracegrpc.New(ctx, opts...)
}

// createStdoutExporter creates a stdout exporter for testing.
func createStdoutExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
}
