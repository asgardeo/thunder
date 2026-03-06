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

package opentelemetry

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Initialize creates and configures a TracerProvider based on the provided configuration,
// sets it as the global tracer provider, and configures trace context propagation.
//
// Parameters:
//   - ctx: Context for initialization (used for exporter setup)
//   - cfg: OpenTelemetry configuration including exporter type, endpoint, service details, etc.
//
// Returns:
//   - *sdktrace.TracerProvider: The initialized tracer provider
//   - error: Error if initialization fails (e.g., invalid config, exporter creation failure)
//
// Example:
//
//	cfg := opentelemetry.Config{
//	    Enabled:      true,
//	    ExporterType: "otlp",
//	    OTLPEndpoint: "localhost:4317",
//	    ServiceName:  "thunder-iam",
//	}
//	provider, err := opentelemetry.Initialize(ctx, cfg)
//	if err != nil {
//	    return err
//	}
//	defer provider.Shutdown(ctx)
func Initialize(ctx context.Context, cfg Config) (*sdktrace.TracerProvider, error) {
	return newTracerProvider(ctx, cfg)
}
