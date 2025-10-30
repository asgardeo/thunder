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
)

func TestGetTraceID_WithNilContext(t *testing.T) {
	traceID := GetTraceID(context.TODO())
	if traceID == "" {
		t.Error("Expected non-empty trace ID, got empty string")
	}
}

func TestGetTraceID_WithEmptyContext(t *testing.T) {
	ctx := context.Background()
	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected non-empty trace ID, got empty string")
	}
}

func TestGetTraceID_WithExistingTraceID(t *testing.T) {
	expectedID := "test-trace-id-123"
	ctx := WithTraceID(context.Background(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		t.Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func TestWithTraceID(t *testing.T) {
	expectedID := "custom-trace-id"
	ctx := WithTraceID(context.Background(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		t.Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func TestWithTraceID_NilContext(t *testing.T) {
	expectedID := "custom-trace-id"
	ctx := WithTraceID(context.TODO(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		t.Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func TestEnsureTraceID_CreatesNew(t *testing.T) {
	ctx := context.Background()
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected non-empty trace ID after EnsureTraceID")
	}
}

func TestEnsureTraceID_PreservesExisting(t *testing.T) {
	expectedID := "existing-trace-id"
	ctx := WithTraceID(context.Background(), expectedID)
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		t.Errorf("Expected trace ID %s to be preserved, got %s", expectedID, traceID)
	}
}

func TestEnsureTraceID_NilContext(t *testing.T) {
	ctx := EnsureTraceID(context.TODO())

	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected non-empty trace ID after EnsureTraceID with nil context")
	}
}

func TestGenerateUUID_Format(t *testing.T) {
	uuid := generateUUID()

	// UUID should be in format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(uuid) != 36 {
		t.Errorf("Expected UUID length of 36, got %d", len(uuid))
	}

	// Check for dashes at correct positions
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		t.Errorf("UUID format incorrect: %s", uuid)
	}
}

func TestGenerateUUID_Uniqueness(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	if uuid1 == uuid2 {
		t.Error("Expected different UUIDs, got same value")
	}
}
