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

// Package metrics provides metrics collection for the analytics system.
package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "AnalyticsMetrics"

// Metrics holds metrics for the analytics system.
type Metrics struct {
	// Events published
	eventsPublished atomic.Uint64

	// Events dropped (queue full)
	eventsDropped atomic.Uint64

	// Events skipped (no subscribers)
	eventsSkipped atomic.Uint64

	// Events processed by subscribers
	eventsProcessed atomic.Uint64

	// Subscriber errors
	subscriberErrors atomic.Uint64

	// Queue depth (current size)
	queueDepth atomic.Int64

	// Queue capacity
	queueCapacity int64

	// Subscriber processing latency (in microseconds)
	totalProcessingLatency atomic.Uint64
	processingCount        atomic.Uint64

	// Start time
	startTime time.Time

	// Enabled flag
	enabled bool

	mu sync.RWMutex
}

var (
	metricsInstance *Metrics
	metricsOnce     sync.Once
)

// GetMetrics returns the singleton metrics instance.
func GetMetrics() *Metrics {
	metricsOnce.Do(func() {
		metricsInstance = &Metrics{
			startTime:     time.Now(),
			enabled:       true,
			queueCapacity: 1000, // default
		}
	})
	return metricsInstance
}

// SetQueueCapacity sets the queue capacity for metrics calculations.
func (m *Metrics) SetQueueCapacity(capacity int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queueCapacity = capacity
}

// Enable enables metrics collection.
func (m *Metrics) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables metrics collection.
func (m *Metrics) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// IsEnabled returns whether metrics collection is enabled.
func (m *Metrics) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// IncrementEventsPublished increments the events published counter.
func (m *Metrics) IncrementEventsPublished() {
	if !m.IsEnabled() {
		return
	}
	m.eventsPublished.Add(1)
}

// IncrementEventsDropped increments the events dropped counter.
func (m *Metrics) IncrementEventsDropped() {
	if !m.IsEnabled() {
		return
	}
	m.eventsDropped.Add(1)
}

// IncrementEventsSkipped increments the events skipped counter (no subscribers).
func (m *Metrics) IncrementEventsSkipped() {
	if !m.IsEnabled() {
		return
	}
	m.eventsSkipped.Add(1)
}

// IncrementEventsProcessed increments the events processed counter.
func (m *Metrics) IncrementEventsProcessed() {
	if !m.IsEnabled() {
		return
	}
	m.eventsProcessed.Add(1)
}

// IncrementSubscriberErrors increments the subscriber errors counter.
func (m *Metrics) IncrementSubscriberErrors() {
	if !m.IsEnabled() {
		return
	}
	m.subscriberErrors.Add(1)
}

// UpdateQueueDepth updates the current queue depth.
func (m *Metrics) UpdateQueueDepth(depth int64) {
	if !m.IsEnabled() {
		return
	}
	m.queueDepth.Store(depth)
}

// RecordProcessingLatency records subscriber processing latency.
func (m *Metrics) RecordProcessingLatency(latencyMicros uint64) {
	if !m.IsEnabled() {
		return
	}
	m.totalProcessingLatency.Add(latencyMicros)
	m.processingCount.Add(1)
}

// GetEventsPublished returns the number of events published.
func (m *Metrics) GetEventsPublished() uint64 {
	return m.eventsPublished.Load()
}

// GetEventsDropped returns the number of events dropped.
func (m *Metrics) GetEventsDropped() uint64 {
	return m.eventsDropped.Load()
}

// GetEventsSkipped returns the number of events skipped (no subscribers).
func (m *Metrics) GetEventsSkipped() uint64 {
	return m.eventsSkipped.Load()
}

// GetEventsProcessed returns the number of events processed.
func (m *Metrics) GetEventsProcessed() uint64 {
	return m.eventsProcessed.Load()
}

// GetSubscriberErrors returns the number of subscriber errors.
func (m *Metrics) GetSubscriberErrors() uint64 {
	return m.subscriberErrors.Load()
}

// GetQueueDepth returns the current queue depth.
func (m *Metrics) GetQueueDepth() int64 {
	return m.queueDepth.Load()
}

// GetQueueCapacity returns the queue capacity.
func (m *Metrics) GetQueueCapacity() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.queueCapacity
}

// GetQueueUtilization returns the queue utilization percentage (0-100).
func (m *Metrics) GetQueueUtilization() float64 {
	capacity := m.GetQueueCapacity()
	if capacity == 0 {
		return 0
	}
	depth := float64(m.GetQueueDepth())
	return (depth / float64(capacity)) * 100.0
}

// GetAverageProcessingLatency returns the average processing latency in microseconds.
func (m *Metrics) GetAverageProcessingLatency() uint64 {
	count := m.processingCount.Load()
	if count == 0 {
		return 0
	}
	total := m.totalProcessingLatency.Load()
	return total / count
}

// GetUptime returns how long the metrics have been running.
func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// GetDropRate returns the event drop rate percentage.
func (m *Metrics) GetDropRate() float64 {
	published := float64(m.GetEventsPublished())
	if published == 0 {
		return 0
	}
	dropped := float64(m.GetEventsDropped())
	return (dropped / published) * 100.0
}

// GetSkipRate returns the event skip rate percentage (events with no subscribers).
func (m *Metrics) GetSkipRate() float64 {
	total := float64(m.GetEventsPublished() + m.GetEventsSkipped())
	if total == 0 {
		return 0
	}
	skipped := float64(m.GetEventsSkipped())
	return (skipped / total) * 100.0
}

// GetSuccessRate returns the event success rate percentage.
func (m *Metrics) GetSuccessRate() float64 {
	processed := float64(m.GetEventsProcessed())
	if processed == 0 {
		return 100.0
	}
	errors := float64(m.GetSubscriberErrors())
	return ((processed - errors) / processed) * 100.0
}

// Snapshot returns a snapshot of current metrics.
type Snapshot struct {
	EventsPublished      uint64        `json:"events_published"`
	EventsDropped        uint64        `json:"events_dropped"`
	EventsSkipped        uint64        `json:"events_skipped"`
	EventsProcessed      uint64        `json:"events_processed"`
	SubscriberErrors     uint64        `json:"subscriber_errors"`
	QueueDepth           int64         `json:"queue_depth"`
	QueueCapacity        int64         `json:"queue_capacity"`
	QueueUtilization     float64       `json:"queue_utilization_percent"`
	AverageLatencyMicros uint64        `json:"average_latency_micros"`
	DropRate             float64       `json:"drop_rate_percent"`
	SkipRate             float64       `json:"skip_rate_percent"`
	SuccessRate          float64       `json:"success_rate_percent"`
	Uptime               time.Duration `json:"uptime"`
	Timestamp            time.Time     `json:"timestamp"`
}

// GetSnapshot returns a snapshot of current metrics.
func (m *Metrics) GetSnapshot() *Snapshot {
	return &Snapshot{
		EventsPublished:      m.GetEventsPublished(),
		EventsDropped:        m.GetEventsDropped(),
		EventsSkipped:        m.GetEventsSkipped(),
		EventsProcessed:      m.GetEventsProcessed(),
		SubscriberErrors:     m.GetSubscriberErrors(),
		QueueDepth:           m.GetQueueDepth(),
		QueueCapacity:        m.GetQueueCapacity(),
		QueueUtilization:     m.GetQueueUtilization(),
		AverageLatencyMicros: m.GetAverageProcessingLatency(),
		DropRate:             m.GetDropRate(),
		SkipRate:             m.GetSkipRate(),
		SuccessRate:          m.GetSuccessRate(),
		Uptime:               m.GetUptime(),
		Timestamp:            time.Now(),
	}
}

// LogSnapshot logs the current metrics snapshot.
func (m *Metrics) LogSnapshot() {
	if !m.IsEnabled() {
		return
	}

	snapshot := m.GetSnapshot()
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	logger.Info("Analytics metrics snapshot",
		log.Any("events_published", snapshot.EventsPublished),
		log.Any("events_dropped", snapshot.EventsDropped),
		log.Any("events_skipped", snapshot.EventsSkipped),
		log.Any("events_processed", snapshot.EventsProcessed),
		log.Any("subscriber_errors", snapshot.SubscriberErrors),
		log.Any("queue_depth", snapshot.QueueDepth),
		log.Any("queue_capacity", snapshot.QueueCapacity),
		log.Any("queue_utilization_percent", snapshot.QueueUtilization),
		log.Any("avg_latency_micros", snapshot.AverageLatencyMicros),
		log.Any("drop_rate_percent", snapshot.DropRate),
		log.Any("skip_rate_percent", snapshot.SkipRate),
		log.Any("success_rate_percent", snapshot.SuccessRate),
		log.Any("uptime", snapshot.Uptime))
}

// Reset resets all metrics counters (primarily for testing).
func (m *Metrics) Reset() {
	m.eventsPublished.Store(0)
	m.eventsDropped.Store(0)
	m.eventsSkipped.Store(0)
	m.eventsProcessed.Store(0)
	m.subscriberErrors.Store(0)
	m.queueDepth.Store(0)
	m.totalProcessingLatency.Store(0)
	m.processingCount.Store(0)
	m.startTime = time.Now()
}
