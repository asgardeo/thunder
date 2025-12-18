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

package event

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name         string
		eventType    EventType
		wantCategory EventCategory
	}{
		// Authentication events
		{
			name:         "token issuance started",
			eventType:    EventTypeTokenIssuanceStarted,
			wantCategory: CategoryAuthentication,
		},
		{
			name:         "token issued",
			eventType:    EventTypeTokenIssued,
			wantCategory: CategoryAuthentication,
		},
		{
			name:         "token issuance failed",
			eventType:    EventTypeTokenIssuanceFailed,
			wantCategory: CategoryAuthentication,
		},

		// Flow events
		{
			name:         "flow started",
			eventType:    EventTypeFlowStarted,
			wantCategory: CategoryFlows,
		},
		{
			name:         "flow completed",
			eventType:    EventTypeFlowCompleted,
			wantCategory: CategoryFlows,
		},
		{
			name:         "flow node execution started",
			eventType:    EventTypeFlowNodeExecutionStarted,
			wantCategory: CategoryFlows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCategory(tt.eventType)
			if err != nil {
				t.Errorf("GetCategory(%s) unexpected error: %v", tt.eventType, err)
			}
			if got != tt.wantCategory {
				t.Errorf("GetCategory(%s) = %s, want %s", tt.eventType, got, tt.wantCategory)
			}
		})
	}
}

func TestEvent_GetCategory(t *testing.T) {
	tests := []struct {
		name         string
		event        *Event
		wantCategory EventCategory
	}{
		{
			name: "authentication event",
			event: &Event{
				TraceID:   "trace-1",
				EventID:   "event-1",
				Type:      string(EventTypeTokenIssuanceStarted),
				Component: "test",
				Timestamp: time.Now(),
			},
			wantCategory: CategoryAuthentication,
		},
		{
			name: "flow event",
			event: &Event{
				TraceID:   "trace-2",
				EventID:   "event-2",
				Type:      string(EventTypeFlowStarted),
				Component: "test",
				Timestamp: time.Now(),
			},
			wantCategory: CategoryFlows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.event.GetCategory()
			if err != nil {
				t.Errorf("Event.GetCategory() unexpected error: %v", err)
			}
			if got != tt.wantCategory {
				t.Errorf("Event.GetCategory() = %s, want %s", got, tt.wantCategory)
			}
		})
	}
}

func TestGetCategory_ReturnsErrorOnUnmappedEventType(t *testing.T) {
	_, err := GetCategory(EventType("unknown.unmapped.event"))
	if err == nil {
		t.Error("GetCategory() should return error for unmapped event type")
	}

	// Verify error message contains expected text
	expectedMsg := "event type not mapped to category"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err.Error())
	}

	// Verify it's the correct error type
	var unmappedErr *UnmappedEventTypeError
	if !errors.As(err, &unmappedErr) {
		t.Errorf("Expected error type *UnmappedEventTypeError, got: %T", err)
	}
}

func TestEvent_GetCategory_ReturnsErrorOnUnmappedEventType(t *testing.T) {
	evt := &Event{
		TraceID:   "trace-1",
		EventID:   "event-1",
		Type:      "unknown.unmapped.event",
		Component: "test",
		Timestamp: time.Now(),
	}

	_, err := evt.GetCategory()
	if err == nil {
		t.Error("Event.GetCategory() should return error for unmapped event type")
	}

	// Verify it's the correct error type
	var unmappedErr *UnmappedEventTypeError
	if !errors.As(err, &unmappedErr) {
		t.Errorf("Expected error type *UnmappedEventTypeError, got: %T", err)
	}
}

func TestGetAllCategories(t *testing.T) {
	categories := GetAllCategories()

	if len(categories) == 0 {
		t.Error("GetAllCategories() returned empty slice")
	}

	// Expected categories (excluding CategoryAll)
	expectedCategories := map[EventCategory]bool{
		CategoryAuthentication: false,
		CategoryAuthorization:  false,
		CategoryFlows:          false,
	}

	for _, cat := range categories {
		if _, exists := expectedCategories[cat]; exists {
			expectedCategories[cat] = true
		}
	}

	// Verify all expected categories were found
	for cat, found := range expectedCategories {
		if !found {
			t.Errorf("Expected category %s not found in GetAllCategories()", cat)
		}
	}

	// Verify CategoryAll is NOT in the list
	for _, cat := range categories {
		if cat == CategoryAll {
			t.Error("CategoryAll should not be in GetAllCategories()")
		}
	}
}

func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		name     string
		category EventCategory
		want     bool
	}{
		{
			name:     "valid authentication category",
			category: CategoryAuthentication,
			want:     true,
		},
		{
			name:     "valid authorization category",
			category: CategoryAuthorization,
			want:     true,
		},
		{
			name:     "valid flows category",
			category: CategoryFlows,
			want:     true,
		},
		{
			name:     "valid CategoryAll",
			category: CategoryAll,
			want:     true,
		},
		{
			name:     "invalid custom category",
			category: EventCategory("observability.custom"),
			want:     false,
		},
		{
			name:     "invalid random string",
			category: EventCategory("not-a-category"),
			want:     false,
		},
		{
			name:     "empty category",
			category: EventCategory(""),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidCategory(tt.category)
			if got != tt.want {
				t.Errorf("IsValidCategory(%s) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

func TestCategoryConstants(t *testing.T) {
	// Verify category constants are not empty
	if CategoryAuthentication == "" {
		t.Error("CategoryAuthentication should not be empty")
	}
	if CategoryAuthorization == "" {
		t.Error("CategoryAuthorization should not be empty")
	}
	if CategoryFlows == "" {
		t.Error("CategoryFlows should not be empty")
	}
	if CategoryAll == "" {
		t.Error("CategoryAll should not be empty")
	}

	// Verify categories have expected prefix
	expectedPrefix := "observability."
	categories := append(GetAllCategories(), CategoryAll)

	for _, cat := range categories {
		catStr := string(cat)
		if !strings.HasPrefix(catStr, expectedPrefix) {
			t.Errorf("Category %s does not start with expected prefix %s", cat, expectedPrefix)
		}
	}
}

func TestEventTypeToCategoryMapping_Comprehensive(t *testing.T) {
	// Verify all defined event types have a category mapping
	allEventTypes := []EventType{
		// Authentication
		EventTypeTokenIssuanceStarted,
		EventTypeTokenIssued,
		EventTypeTokenIssuanceFailed,

		// Flows
		EventTypeFlowStarted,
		EventTypeFlowNodeExecutionStarted,
		EventTypeFlowNodeExecutionCompleted,
		EventTypeFlowNodeExecutionFailed,
		EventTypeFlowUserInputRequired,
		EventTypeFlowCompleted,
		EventTypeFlowFailed,
	}

	for _, eventType := range allEventTypes {
		t.Run(string(eventType), func(t *testing.T) {
			category, err := GetCategory(eventType)
			if err != nil {
				t.Errorf("GetCategory(%s) unexpected error: %v", eventType, err)
			}
			if !IsValidCategory(category) {
				t.Errorf("Event type %s maps to invalid category %s", eventType, category)
			}
		})
	}
}

func TestCategoryToEventTypeConsistency(t *testing.T) {
	// Verify that each category has at least one event type mapped to it
	categoryEventCount := make(map[EventCategory]int)

	allEventTypes := []EventType{
		EventTypeTokenIssuanceStarted, EventTypeTokenIssued, EventTypeTokenIssuanceFailed,
		EventTypeFlowStarted, EventTypeFlowCompleted, EventTypeFlowFailed,
	}

	for _, eventType := range allEventTypes {
		category, err := GetCategory(eventType)
		if err != nil {
			t.Fatalf("GetCategory(%s) unexpected error: %v", eventType, err)
		}
		categoryEventCount[category]++
	}

	// Verify each main category has events
	mainCategories := []EventCategory{
		CategoryAuthentication,
		CategoryFlows,
	}

	for _, cat := range mainCategories {
		count := categoryEventCount[cat]
		if count == 0 {
			t.Errorf("Category %s has no event types mapped to it", cat)
		}
	}
}
