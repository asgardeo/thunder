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
			name:         "authentication started",
			eventType:    EventTypeAuthenticationStarted,
			wantCategory: CategoryAuthentication,
		},
		{
			name:         "authentication completed",
			eventType:    EventTypeAuthenticationCompleted,
			wantCategory: CategoryAuthentication,
		},
		{
			name:         "authentication failed",
			eventType:    EventTypeAuthenticationFailed,
			wantCategory: CategoryAuthentication,
		},
		{
			name:         "credentials auth started",
			eventType:    EventTypeCredentialsAuthStarted,
			wantCategory: CategoryAuthentication,
		},
		{
			name:         "OTP sent",
			eventType:    EventTypeOTPSent,
			wantCategory: CategoryAuthentication,
		},
		{
			name:         "social auth started",
			eventType:    EventTypeSocialAuthStarted,
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
			got := GetCategory(tt.eventType)
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
				Type:      string(EventTypeAuthenticationStarted),
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
			got := tt.event.GetCategory()
			if got != tt.wantCategory {
				t.Errorf("Event.GetCategory() = %s, want %s", got, tt.wantCategory)
			}
		})
	}
}

func TestGetCategory_PanicsOnUnmappedEventType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("GetCategory() should panic for unmapped event type")
		} else {
			expectedMsg := "event type not mapped to category"
			if msg, ok := r.(string); !ok || !strings.Contains(msg, expectedMsg) {
				t.Errorf("Expected panic message to contain '%s', got: %v", expectedMsg, r)
			}
		}
	}()

	// This should panic
	GetCategory(EventType("unknown.unmapped.event"))
}

func TestEvent_GetCategory_PanicsOnUnmappedEventType(t *testing.T) {
	evt := &Event{
		TraceID:   "trace-1",
		EventID:   "event-1",
		Type:      "unknown.unmapped.event",
		Component: "test",
		Timestamp: time.Now(),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Event.GetCategory() should panic for unmapped event type")
		}
	}()

	// This should panic
	evt.GetCategory()
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
		EventTypeAuthenticationStarted,
		EventTypeAuthenticationMethodSelected,
		EventTypeCredentialsAuthStarted,
		EventTypeCredentialsAuthCompleted,
		EventTypeCredentialsAuthFailed,
		EventTypeOTPSent,
		EventTypeOTPVerificationStarted,
		EventTypeOTPVerified,
		EventTypeOTPVerificationFailed,
		EventTypeSocialAuthStarted,
		EventTypeSocialAuthCallbackReceived,
		EventTypeSocialAuthCompleted,
		EventTypeSocialAuthFailed,
		EventTypeAuthenticationCompleted,
		EventTypeAuthenticationFailed,

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
			category := GetCategory(eventType)
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
		EventTypeAuthenticationStarted, EventTypeAuthenticationCompleted, EventTypeAuthenticationFailed,
		EventTypeFlowStarted, EventTypeFlowCompleted, EventTypeFlowFailed,
	}

	for _, eventType := range allEventTypes {
		category := GetCategory(eventType)
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
