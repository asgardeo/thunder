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
	"testing"

	sysContext "github.com/asgardeo/thunder/internal/system/context"
	"github.com/stretchr/testify/suite"
)

type ContextLoggerTestSuite struct {
	suite.Suite
}

func TestContextLoggerSuite(t *testing.T) {
	suite.Run(t, new(ContextLoggerTestSuite))
}

func (suite *ContextLoggerTestSuite) TestGetLoggerWithContext() {
	suite.Run("WithTraceIDInContext", func() {
		// Create context with trace ID
		traceID := "test-trace-id-12345"
		ctx := sysContext.WithTraceID(context.Background(), traceID)

		// Get logger with context
		logger := GetLoggerWithContext(ctx)

		// Verify logger is not nil
		suite.NotNil(logger)
	})

	suite.Run("WithEmptyContext", func() {
		// Create empty context
		ctx := context.Background()

		// Get logger with context - should still work (generates new trace ID)
		logger := GetLoggerWithContext(ctx)

		// Verify logger is not nil
		suite.NotNil(logger)
	})

	suite.Run("WithNilContext", func() {
		// Get logger with nil context - should handle gracefully
		logger := GetLoggerWithContext(nil)

		// Verify logger is not nil
		suite.NotNil(logger)
	})
}
