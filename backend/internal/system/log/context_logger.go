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

package log

import (
	"context"
)

// GetLoggerWithContext returns a logger with the trace ID (correlation ID) from the given context.
// This is the RECOMMENDED way to get a logger in HTTP handlers and other request-scoped code
// where a context is available.
//
// The trace ID enables correlation of all logs within a single request flow, making it easier
// to trace and debug issues across service boundaries.
//
// Usage in HTTP handlers:
//
//	func (h *handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
//	    logger := log.GetLoggerWithContext(r.Context()).With(log.String(log.LoggerKeyComponentName, "Handler"))
//	    logger.Info("Processing request")
//	}
//
// Usage in services with context parameter:
//
//	func (s *service) ProcessData(ctx context.Context, data *Data) error {
//	    logger := log.GetLoggerWithContext(ctx).With(log.String(log.LoggerKeyComponentName, "Service"))
//	    logger.Debug("Processing data", log.String("dataId", data.ID))
//	    return nil
//	}
//
// If the context does not contain a trace ID, a new one will be generated.
func GetLoggerWithContext(ctx context.Context) *Logger {
	return GetLogger().WithContext(ctx)
}
