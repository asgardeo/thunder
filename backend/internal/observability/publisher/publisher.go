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

// Package publisher provides the event bus mechanism for the observability system.
package publisher

import (
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/observability/event"
	"github.com/asgardeo/thunder/internal/observability/metrics"
	"github.com/asgardeo/thunder/internal/observability/subscriber"
	"github.com/asgardeo/thunder/internal/system/log"
)

// CategoryPublisher is an event bus interface with category-based routing.
// Events are published synchronously to all interested subscribers.
type CategoryPublisher interface {
	// Publish publishes an event immediately to all interested subscribers.
	Publish(event *event.Event)

	// Subscribe adds a subscriber (subscriber decides what categories it wants).
	Subscribe(sub subscriber.Subscriber)

	// Unsubscribe removes a subscriber.
	Unsubscribe(sub subscriber.Subscriber)

	// GetActiveCategories returns categories that have at least one subscriber.
	GetActiveCategories() []event.EventCategory

	// Shutdown gracefully shuts down the publisher.
	Shutdown()
}

// categoryEventPublisher implements category-based event bus with smart routing.
// This is a simple synchronous event bus - no queuing, no async processing.
// Events are published directly to subscribers who can decide whether to process them.
type categoryEventPublisher struct {
	subscribers           map[string]subscriber.Subscriber // subscriberID -> subscriber
	subscribersByCategory map[event.EventCategory][]string // category -> []subscriberIDs
	mu                    sync.RWMutex
	isShutdown            bool
}

const publisherLogComponent = "CategoryEventBus"

// NewCategoryPublisher creates a new category-based event bus.
func NewCategoryPublisher() CategoryPublisher {
	return &categoryEventPublisher{
		subscribers:           make(map[string]subscriber.Subscriber),
		subscribersByCategory: make(map[event.EventCategory][]string),
		isShutdown:            false,
	}
}

// Publish publishes an event to all interested subscribers.
// This method returns immediately without blocking the caller.
// Each subscriber receives the event in its own goroutine, ensuring:
// - No blocking of the main thread
// - Parallel processing by all subscribers
// - Isolated failures (one subscriber's panic doesn't affect others)
func (p *categoryEventPublisher) Publish(evt *event.Event) {
	if evt == nil {
		return
	}

	// Validate event
	if err := evt.Validate(); err != nil {
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, publisherLogComponent))
		logger.Warn("Invalid event, skipping publish", log.Error(err))
		return
	}

	// Get event category
	category := evt.GetCategory()

	p.mu.RLock()
	if p.isShutdown {
		p.mu.RUnlock()
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, publisherLogComponent))
		logger.Warn("Attempted to publish event after shutdown",
			log.String("eventType", evt.Type),
			log.String("category", string(category)))
		return
	}

	// Smart Publishing: Check if anyone cares about this category
	hasSubscribers := false
	if subscribers, exists := p.subscribersByCategory[category]; exists && len(subscribers) > 0 {
		hasSubscribers = true
	}
	// Also check for CategoryAll subscribers
	if allSubscribers, exists := p.subscribersByCategory[event.CategoryAll]; exists && len(allSubscribers) > 0 {
		hasSubscribers = true
	}

	if !hasSubscribers {
		// No subscribers for this category - skip it!
		p.mu.RUnlock()
		metrics.GetMetrics().IncrementEventsSkipped()
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, publisherLogComponent))
		if logger.IsDebugEnabled() {
			logger.Debug("No subscribers for event category, skipping",
				log.String("eventType", evt.Type),
				log.String("category", string(category)))
		}
		return
	}

	// Get ALL subscribers (they will filter themselves)
	allSubscribers := make([]subscriber.Subscriber, 0, len(p.subscribers))
	for _, sub := range p.subscribers {
		allSubscribers = append(allSubscribers, sub)
	}
	p.mu.RUnlock()

	// Event has interested subscribers
	metrics.GetMetrics().IncrementEventsPublished()

	// Broadcast to all subscribers asynchronously
	// Each subscriber runs in its own goroutine to avoid blocking
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, publisherLogComponent))
	for _, sub := range allSubscribers {
		go func(s subscriber.Subscriber) {
			startTime := time.Now()
			defer func() {
				if r := recover(); r != nil {
					metrics.GetMetrics().IncrementSubscriberErrors()
					logger.Error("Subscriber panicked while handling event",
						log.String("subscriberID", s.GetID()),
						log.Any("panic", r))
				}
			}()

			// Subscriber will filter the event itself based on its interests
			if err := s.OnEvent(evt); err != nil {
				metrics.GetMetrics().IncrementSubscriberErrors()
				logger.Error("Subscriber failed to handle event",
					log.String("subscriberID", s.GetID()),
					log.String("eventType", evt.Type),
					log.Error(err))
			} else {
				// Record successful processing
				metrics.GetMetrics().IncrementEventsProcessed()
				latency := time.Since(startTime).Microseconds()
				if latency >= 0 {
					metrics.GetMetrics().RecordProcessingLatency(uint64(latency))
				}
			}
		}(sub)
	}
}

// Subscribe adds a subscriber.
// The subscriber's GetCategories() method determines what categories it receives.
func (p *categoryEventPublisher) Subscribe(sub subscriber.Subscriber) {
	if sub == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	subscriberID := sub.GetID()
	p.subscribers[subscriberID] = sub

	// Register subscriber for its interested categories
	categories := sub.GetCategories()
	for _, category := range categories {
		if p.subscribersByCategory[category] == nil {
			p.subscribersByCategory[category] = make([]string, 0)
		}
		p.subscribersByCategory[category] = append(p.subscribersByCategory[category], subscriberID)
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, publisherLogComponent))
	logger.Info("Subscriber registered",
		log.String("subscriberID", subscriberID),
		log.Int("categoryCount", len(categories)),
		log.Any("categories", categories))
}

// Unsubscribe removes a subscriber.
func (p *categoryEventPublisher) Unsubscribe(sub subscriber.Subscriber) {
	if sub == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	subscriberID := sub.GetID()

	// Remove from subscribers map
	delete(p.subscribers, subscriberID)

	// Remove from all category lists
	for category, subscriberList := range p.subscribersByCategory {
		newList := make([]string, 0, len(subscriberList))
		for _, id := range subscriberList {
			if id != subscriberID {
				newList = append(newList, id)
			}
		}
		p.subscribersByCategory[category] = newList

		// Clean up empty category lists
		if len(newList) == 0 {
			delete(p.subscribersByCategory, category)
		}
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, publisherLogComponent))
	logger.Info("Subscriber unregistered", log.String("subscriberID", subscriberID))
}

// GetActiveCategories returns categories that have at least one subscriber.
func (p *categoryEventPublisher) GetActiveCategories() []event.EventCategory {
	p.mu.RLock()
	defer p.mu.RUnlock()

	categories := make([]event.EventCategory, 0, len(p.subscribersByCategory))
	for category, subscribers := range p.subscribersByCategory {
		if len(subscribers) > 0 {
			categories = append(categories, category)
		}
	}

	return categories
}

// Shutdown gracefully shuts down the publisher and all subscribers.
func (p *categoryEventPublisher) Shutdown() {
	p.mu.Lock()
	if p.isShutdown {
		p.mu.Unlock()
		return
	}
	p.isShutdown = true

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, publisherLogComponent))
	logger.Info("Shutting down event bus")

	// Get all subscribers
	subscribers := make([]subscriber.Subscriber, 0, len(p.subscribers))
	for _, sub := range p.subscribers {
		subscribers = append(subscribers, sub)
	}
	p.mu.Unlock()

	// Close all subscribers
	for _, sub := range subscribers {
		if err := sub.Close(); err != nil {
			logger.Error("Error closing subscriber",
				log.String("subscriberID", sub.GetID()),
				log.Error(err))
		}
	}

	logger.Info("Event bus shutdown complete")
}
