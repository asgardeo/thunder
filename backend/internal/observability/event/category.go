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

// Package event provides event models and types for the observability system.
package event

// EventCategory represents a category for grouping related events.
// Used by the event bus for efficient routing - subscribers declare which
// categories they're interested in, and events are only delivered to
// subscribers that match the event's category.
//
// Benefits:
// - Performance: Skip events with no interested subscribers
// - Flexibility: Subscribers can filter by category, event type, or both
// - Organization: Logical grouping of related events
//
// Example usage:
//
//	subscriber.GetCategories() returns [CategoryAuthentication, CategoryTokens]
//	Event with category "authentication" → only delivered to auth subscribers
type EventCategory string

const (
	// CategoryAuthentication groups all authentication-related events.
	CategoryAuthentication EventCategory = "observability.authentication"

	// CategoryAuthorization groups all authorization-related events.
	CategoryAuthorization EventCategory = "observability.authorization"

	// CategoryFlows groups all flow orchestration events for tracing end-to-end flows.
	CategoryFlows EventCategory = "observability.flows"

	// CategoryAll is a special category that matches all events.
	// Subscribers to this category receive all events regardless of type.
	CategoryAll EventCategory = "observability.all"
)

// eventTypeToCategory maps each event type to its category.
// This enables automatic routing of events to appropriate categories.
var eventTypeToCategory = map[EventType]EventCategory{
	// Authentication events
	EventTypeAuthenticationStarted:        CategoryAuthentication,
	EventTypeAuthenticationMethodSelected: CategoryAuthentication,
	EventTypeCredentialsAuthStarted:       CategoryAuthentication,
	EventTypeCredentialsAuthCompleted:     CategoryAuthentication,
	EventTypeCredentialsAuthFailed:        CategoryAuthentication,
	EventTypeOTPSent:                      CategoryAuthentication,
	EventTypeOTPVerificationStarted:       CategoryAuthentication,
	EventTypeOTPVerified:                  CategoryAuthentication,
	EventTypeOTPVerificationFailed:        CategoryAuthentication,
	EventTypeSocialAuthStarted:            CategoryAuthentication,
	EventTypeSocialAuthCallbackReceived:   CategoryAuthentication,
	EventTypeSocialAuthCompleted:          CategoryAuthentication,
	EventTypeSocialAuthFailed:             CategoryAuthentication,
	EventTypeAuthenticationCompleted:      CategoryAuthentication,
	EventTypeAuthenticationFailed:         CategoryAuthentication,

	// Flow events
	EventTypeFlowStarted:                CategoryFlows,
	EventTypeFlowNodeExecutionStarted:   CategoryFlows,
	EventTypeFlowNodeExecutionCompleted: CategoryFlows,
	EventTypeFlowNodeExecutionFailed:    CategoryFlows,
	EventTypeFlowUserInputRequired:      CategoryFlows,
	EventTypeFlowCompleted:              CategoryFlows,
	EventTypeFlowFailed:                 CategoryFlows,
}

// GetCategory returns the category for a given event type.
// If the event type is not mapped, it panics to prevent unintentional use
// of unmapped event types. All event types must be explicitly mapped to categories.
func GetCategory(eventType EventType) EventCategory {
	if category, exists := eventTypeToCategory[eventType]; exists {
		return category
	}
	// Panic to catch unmapped event types during development/testing
	panic("event type not mapped to category: " + string(eventType) + " - all event types must be explicitly mapped")
}

// GetCategory returns the category for the given event.
// If the event type is not mapped to a category, it panics.
// This ensures all event types are explicitly mapped to prevent unintentional wildcard usage.
func (e *Event) GetCategory() EventCategory {
	return GetCategory(EventType(e.Type))
}

// GetAllCategories returns all defined event categories (excluding CategoryAll).
func GetAllCategories() []EventCategory {
	return []EventCategory{
		CategoryAuthentication,
		CategoryAuthorization,
		CategoryFlows,
	}
}

// IsValidCategory checks if a category is valid.
func IsValidCategory(category EventCategory) bool {
	if category == CategoryAll {
		return true
	}

	for _, validCategory := range GetAllCategories() {
		if category == validCategory {
			return true
		}
	}

	return false
}
