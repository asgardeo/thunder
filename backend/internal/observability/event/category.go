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

// Package event provides event models and types for the analytics system.
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
	CategoryAuthentication EventCategory = "analytics.authentication"

	// CategoryAuthorization groups all authorization-related events.
	CategoryAuthorization EventCategory = "analytics.authorization"

	// CategoryTokens groups all token-related events.
	CategoryTokens EventCategory = "analytics.tokens" // #nosec G101 -- Not a credential, analytics category name

	// CategoryFlows groups all flow execution events.
	CategoryFlows EventCategory = "analytics.flows"

	// CategorySessions groups all session-related events.
	CategorySessions EventCategory = "analytics.sessions"

	// CategoryRegistration groups all user registration events.
	CategoryRegistration EventCategory = "analytics.registration"

	// CategoryAll is a special category that matches all events.
	// Subscribers to this category receive all events regardless of type.
	CategoryAll EventCategory = "analytics.all"
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

	// Authorization events
	EventTypeAuthorizationStarted:       CategoryAuthorization,
	EventTypeAuthorizationValidated:     CategoryAuthorization,
	EventTypeAuthorizationRedirect:      CategoryAuthorization,
	EventTypeAuthorizationCodeGenerated: CategoryAuthorization,
	EventTypeAuthorizationCompleted:     CategoryAuthorization,
	EventTypeAuthorizationFailed:        CategoryAuthorization,

	// Token events
	EventTypeTokenRequestReceived:       CategoryTokens,
	EventTypeTokenRequestValidated:      CategoryTokens,
	EventTypeAuthorizationCodeValidated: CategoryTokens,
	EventTypePKCEValidated:              CategoryTokens,
	EventTypePKCEFailed:                 CategoryTokens,
	EventTypeAccessTokenGenerated:       CategoryTokens,
	EventTypeIDTokenGenerated:           CategoryTokens,
	EventTypeRefreshTokenGenerated:      CategoryTokens,
	EventTypeTokenIssued:                CategoryTokens,
	EventTypeTokenRequestFailed:         CategoryTokens,
	EventTypeRefreshTokenUsed:           CategoryTokens,

	// Flow events
	EventTypeFlowStarted:                CategoryFlows,
	EventTypeFlowNodeExecutionStarted:   CategoryFlows,
	EventTypeFlowNodeExecutionCompleted: CategoryFlows,
	EventTypeFlowNodeExecutionFailed:    CategoryFlows,
	EventTypeFlowUserInputRequired:      CategoryFlows,
	EventTypeFlowCompleted:              CategoryFlows,
	EventTypeFlowFailed:                 CategoryFlows,

	// Registration events
	EventTypeRegistrationStarted:   CategoryRegistration,
	EventTypeUserProvisioned:       CategoryRegistration,
	EventTypeRegistrationCompleted: CategoryRegistration,
	EventTypeRegistrationFailed:    CategoryRegistration,

	// Session events
	EventTypeSessionCreated:   CategorySessions,
	EventTypeSessionUpdated:   CategorySessions,
	EventTypeSessionExpired:   CategorySessions,
	EventTypeSessionDestroyed: CategorySessions,
}

// GetCategory returns the category for a given event type.
// If the event type is not mapped, it returns CategoryAll.
func GetCategory(eventType EventType) EventCategory {
	if category, exists := eventTypeToCategory[eventType]; exists {
		return category
	}
	return CategoryAll
}

// GetCategory returns the category for the given event.
func (e *Event) GetCategory() EventCategory {
	return GetCategory(EventType(e.Type))
}

// GetAllCategories returns all defined event categories (excluding CategoryAll).
func GetAllCategories() []EventCategory {
	return []EventCategory{
		CategoryAuthentication,
		CategoryAuthorization,
		CategoryTokens,
		CategoryFlows,
		CategorySessions,
		CategoryRegistration,
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
