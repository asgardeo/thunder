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

package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ContextTestSuite struct {
	suite.Suite
}

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}

func (s *ContextTestSuite) TestGetTraceID_WithNilContext() {
	traceID := GetTraceID(nil) //nolint:staticcheck // Testing nil context handling
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID, got empty string")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithEmptyContext() {
	ctx := context.Background()
	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID, got empty string")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithExistingTraceID() {
	expectedID := "test-trace-id-123"
	ctx := WithTraceID(context.Background(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestWithTraceID() {
	expectedID := "custom-trace-id"
	ctx := WithTraceID(context.Background(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestWithTraceID_NilContext() {
	expectedID := "custom-trace-id"
	ctx := WithTraceID(nil, expectedID) //nolint:staticcheck // Testing nil context handling

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_CreatesNew() {
	ctx := context.Background()
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID after EnsureTraceID")
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_PreservesExisting() {
	expectedID := "existing-trace-id"
	ctx := WithTraceID(context.Background(), expectedID)
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s to be preserved, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_NilContext() {
	ctx := EnsureTraceID(nil) //nolint:staticcheck // Testing nil context handling

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID after EnsureTraceID with nil context")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithNilContextExplicit() {
	// Test with explicitly nil context
	traceID := GetTraceID(nil) //nolint:staticcheck // Testing nil context handling
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID with nil context, got empty string")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithEmptyTraceIDValue() {
	// Test with empty trace ID value in context
	ctx := context.WithValue(context.Background(), TraceIDKey, "")
	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID when context has empty trace ID, got empty string")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithWrongType() {
	// Test with wrong type in context (not a string)
	ctx := context.WithValue(context.Background(), TraceIDKey, 12345)
	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID when context has wrong type, got empty string")
	}
}

func (s *ContextTestSuite) TestWithTraceID_WithNilContextExplicit() {
	// Test with explicitly nil context
	expectedID := "test-id-nil-ctx"
	ctx := WithTraceID(nil, expectedID) //nolint:staticcheck // Testing nil context handling

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s with nil context, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestWithTraceID_EmptyString() {
	// Test setting empty trace ID
	ctx := WithTraceID(context.Background(), "")

	// Should still set it in context even if empty
	if ctx == nil {
		s.T().Error("Expected non-nil context")
	}

	// GetTraceID should generate a new one since it's empty
	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected GetTraceID to generate new ID when stored value is empty")
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_WithNilContextExplicit() {
	// Test with explicitly nil context
	ctx := EnsureTraceID(nil) //nolint:staticcheck // Testing nil context handling

	if ctx == nil {
		s.T().Error("Expected non-nil context")
	}

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID after EnsureTraceID with nil context")
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_WithEmptyTraceID() {
	// Test with context containing empty trace ID
	ctx := context.WithValue(context.Background(), TraceIDKey, "")
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected EnsureTraceID to generate new ID when empty")
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_WithWrongType() {
	// Test with context containing wrong type for trace ID
	ctx := context.WithValue(context.Background(), TraceIDKey, 999)
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected EnsureTraceID to generate new ID when type is wrong")
	}

	// Verify it's now a string in the context
	if _, ok := ctx.Value(TraceIDKey).(string); !ok {
		s.T().Error("Expected trace ID to be a string after EnsureTraceID")
	}
}

func (s *ContextTestSuite) TestGenerateUUID_Format() {
	uuid := generateUUID()

	// UUID should be in format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(uuid) != 36 {
		s.T().Errorf("Expected UUID length of 36, got %d", len(uuid))
	}

	// Check for dashes at correct positions
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		s.T().Errorf("UUID format incorrect: %s", uuid)
	}
}

func (s *ContextTestSuite) TestGenerateUUID_Uniqueness() {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	if uuid1 == uuid2 {
		s.T().Error("Expected different UUIDs, got same value")
	}
}

func (s *ContextTestSuite) TestGenerateUUID_Version4() {
	uuid := generateUUID()

	// Check version bits (should be 0100 for version 4)
	// The version is in the 7th byte (after 3rd dash at position 14-15)
	// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx where 4 is version
	if uuid[14] != '4' {
		s.T().Errorf("Expected UUID version 4, got version %c", uuid[14])
	}

	// Check variant bits (should be 10xx for RFC 4122)
	// The variant is in the 9th byte (after 4th dash at position 19)
	// Valid values are 8, 9, a, b (binary 10xx)
	variantChar := uuid[19]
	validVariants := map[byte]bool{'8': true, '9': true, 'a': true, 'b': true}
	if !validVariants[variantChar] {
		s.T().Errorf("Expected UUID variant to be 8, 9, a, or b, got %c", variantChar)
	}
}

func (s *ContextTestSuite) TestGenerateUUID_MultipleGenerations() {
	// Generate many UUIDs to ensure no collisions
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		uuid := generateUUID()
		if seen[uuid] {
			s.T().Errorf("Duplicate UUID generated: %s", uuid)
		}
		seen[uuid] = true
	}
}
