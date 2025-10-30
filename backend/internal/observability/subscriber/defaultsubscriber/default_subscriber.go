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

// Package defaultsubscriber provides a default implementation of the subscriber interface.
package defaultsubscriber

import (
	"fmt"
	"sync"

	"github.com/asgardeo/thunder/internal/observability/adapter"
	"github.com/asgardeo/thunder/internal/observability/event"
	"github.com/asgardeo/thunder/internal/observability/formatter"
	"github.com/asgardeo/thunder/internal/observability/subscriber"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "DefaultSubscriber"

// DefaultSubscriber is the default implementation of the subscriber interface.
// It supports category-based and event-type filtering.
type DefaultSubscriber struct {
	id         string
	categories []event.EventCategory
	eventTypes map[event.EventType]bool
	formatter  formatter.Formatter
	adapter    adapter.OutputAdapter
	logger     *log.Logger
	mu         sync.RWMutex
}

var _ subscriber.Subscriber = (*DefaultSubscriber)(nil)

// NewDefaultSubscriber creates a new default subscriber with the given formatter and adapter.
// By default, it subscribes to all categories.
func NewDefaultSubscriber(fmt formatter.Formatter, adp adapter.OutputAdapter) *DefaultSubscriber {
	return &DefaultSubscriber{
		id:         utils.GenerateUUID(),
		categories: []event.EventCategory{event.CategoryAll},
		eventTypes: make(map[event.EventType]bool),
		formatter:  fmt,
		adapter:    adp,
		logger:     log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
}

// NewDefaultSubscriberWithCategories creates a subscriber that filters by categories.
func NewDefaultSubscriberWithCategories(
	fmt formatter.Formatter,
	adp adapter.OutputAdapter,
	categories ...event.EventCategory,
) *DefaultSubscriber {
	if len(categories) == 0 {
		categories = []event.EventCategory{event.CategoryAll}
	}
	return &DefaultSubscriber{
		id:         utils.GenerateUUID(),
		categories: categories,
		eventTypes: make(map[event.EventType]bool),
		formatter:  fmt,
		adapter:    adp,
		logger:     log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
}

// NewDefaultSubscriberWithEventTypes creates a subscriber that filters by specific event types.
func NewDefaultSubscriberWithEventTypes(
	fmt formatter.Formatter,
	adp adapter.OutputAdapter,
	eventTypes ...event.EventType,
) *DefaultSubscriber {
	eventTypeMap := make(map[event.EventType]bool)
	for _, et := range eventTypes {
		eventTypeMap[et] = true
	}
	return &DefaultSubscriber{
		id:         utils.GenerateUUID(),
		categories: nil, // Will be determined from event types
		eventTypes: eventTypeMap,
		formatter:  fmt,
		adapter:    adp,
		logger:     log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
}

// GetID returns the unique identifier for this subscriber.
func (ds *DefaultSubscriber) GetID() string {
	return ds.id
}

// GetCategories returns the categories this subscriber is interested in.
func (ds *DefaultSubscriber) GetCategories() []event.EventCategory {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if len(ds.categories) > 0 {
		return ds.categories
	}

	// If no categories but has event types, derive categories from event types
	if len(ds.eventTypes) > 0 {
		categoryMap := make(map[event.EventCategory]bool)
		for eventType := range ds.eventTypes {
			category := event.GetCategory(eventType)
			categoryMap[category] = true
		}

		categories := make([]event.EventCategory, 0, len(categoryMap))
		for category := range categoryMap {
			categories = append(categories, category)
		}
		return categories
	}

	// Default: all categories
	return []event.EventCategory{event.CategoryAll}
}

// SetCategories updates the categories this subscriber is interested in.
func (ds *DefaultSubscriber) SetCategories(categories ...event.EventCategory) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.categories = categories
}

// AddEventTypeFilter adds an event type to the filter list.
func (ds *DefaultSubscriber) AddEventTypeFilter(eventType event.EventType) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if ds.eventTypes == nil {
		ds.eventTypes = make(map[event.EventType]bool)
	}
	ds.eventTypes[eventType] = true
}

// RemoveEventTypeFilter removes an event type from the filter list.
func (ds *DefaultSubscriber) RemoveEventTypeFilter(eventType event.EventType) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.eventTypes, eventType)
}

// shouldProcess determines if the subscriber should process this event.
// This implements subscriber-side filtering.
func (ds *DefaultSubscriber) shouldProcess(evt *event.Event) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	// If specific event types are configured, check them first
	if len(ds.eventTypes) > 0 {
		return ds.eventTypes[event.EventType(evt.Type)]
	}

	// If subscribed to all categories, process everything
	for _, category := range ds.categories {
		if category == event.CategoryAll {
			return true
		}
	}

	// Check if event's category matches any subscribed categories
	eventCategory := evt.GetCategory()
	for _, category := range ds.categories {
		if category == eventCategory {
			return true
		}
	}

	return false
}

// OnEvent is called when a new event is published.
// This implements subscriber-side filtering - events not matching the filter are silently ignored.
func (ds *DefaultSubscriber) OnEvent(evt *event.Event) error {
	if evt == nil {
		return fmt.Errorf("event is nil")
	}

	// Subscriber-side filtering: Check if we should process this event
	if !ds.shouldProcess(evt) {
		// Silently ignore events we're not interested in
		return nil
	}

	// Format the event
	formattedData, err := ds.formatter.Format(evt)
	if err != nil {
		ds.logger.Error("Failed to format event",
			log.String("eventType", evt.Type),
			log.String("eventID", evt.EventID),
			log.Error(err))
		return fmt.Errorf("failed to format event: %w", err)
	}

	// Write to output adapter
	if err := ds.adapter.Write(formattedData); err != nil {
		ds.logger.Error("Failed to write event to adapter",
			log.String("eventType", evt.Type),
			log.String("eventID", evt.EventID),
			log.String("adapterName", ds.adapter.GetName()),
			log.Error(err))
		return fmt.Errorf("failed to write to adapter: %w", err)
	}

	if ds.logger.IsDebugEnabled() {
		ds.logger.Debug("Event processed successfully",
			log.String("eventType", evt.Type),
			log.String("eventID", evt.EventID),
			log.String("traceID", evt.TraceID))
	}

	return nil
}

// Close closes the subscriber and releases resources.
func (ds *DefaultSubscriber) Close() error {
	ds.logger.Info("Closing subscriber", log.String("subscriberID", ds.id))

	// Flush and close adapter
	if err := ds.adapter.Flush(); err != nil {
		ds.logger.Error("Failed to flush adapter", log.Error(err))
	}

	if err := ds.adapter.Close(); err != nil {
		ds.logger.Error("Failed to close adapter", log.Error(err))
		return err
	}

	ds.logger.Info("Subscriber closed", log.String("subscriberID", ds.id))
	return nil
}
