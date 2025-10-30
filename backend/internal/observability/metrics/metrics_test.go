package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestGetMetrics(t *testing.T) {
	m1 := GetMetrics()
	m2 := GetMetrics()

	if m1 != m2 {
		t.Error("GetMetrics should return the same instance")
	}
}

func TestMetricsEnableDisable(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	if !m.IsEnabled() {
		t.Error("Metrics should be enabled by default")
	}

	m.Disable()

	if m.IsEnabled() {
		t.Error("Metrics should be disabled")
	}

	m.Enable()

	if !m.IsEnabled() {
		t.Error("Metrics should be enabled")
	}
}

func TestMetricsEventsPublished(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	if m.GetEventsPublished() != 0 {
		t.Errorf("Expected 0 events published, got %d", m.GetEventsPublished())
	}

	m.IncrementEventsPublished()
	m.IncrementEventsPublished()
	m.IncrementEventsPublished()

	if m.GetEventsPublished() != 3 {
		t.Errorf("Expected 3 events published, got %d", m.GetEventsPublished())
	}
}

func TestMetricsEventsDropped(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	if m.GetEventsDropped() != 0 {
		t.Errorf("Expected 0 events dropped, got %d", m.GetEventsDropped())
	}

	m.IncrementEventsDropped()
	m.IncrementEventsDropped()

	if m.GetEventsDropped() != 2 {
		t.Errorf("Expected 2 events dropped, got %d", m.GetEventsDropped())
	}
}

func TestMetricsEventsSkipped(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	if m.GetEventsSkipped() != 0 {
		t.Errorf("Expected 0 events skipped, got %d", m.GetEventsSkipped())
	}

	m.IncrementEventsSkipped()
	m.IncrementEventsSkipped()
	m.IncrementEventsSkipped()

	if m.GetEventsSkipped() != 3 {
		t.Errorf("Expected 3 events skipped, got %d", m.GetEventsSkipped())
	}
}

func TestMetricsEventsProcessed(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	if m.GetEventsProcessed() != 0 {
		t.Errorf("Expected 0 events processed, got %d", m.GetEventsProcessed())
	}

	m.IncrementEventsProcessed()
	m.IncrementEventsProcessed()
	m.IncrementEventsProcessed()
	m.IncrementEventsProcessed()

	if m.GetEventsProcessed() != 4 {
		t.Errorf("Expected 4 events processed, got %d", m.GetEventsProcessed())
	}
}

func TestMetricsSubscriberErrors(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	if m.GetSubscriberErrors() != 0 {
		t.Errorf("Expected 0 subscriber errors, got %d", m.GetSubscriberErrors())
	}

	m.IncrementSubscriberErrors()

	if m.GetSubscriberErrors() != 1 {
		t.Errorf("Expected 1 subscriber error, got %d", m.GetSubscriberErrors())
	}
}

func TestMetricsQueueDepth(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	if m.GetQueueDepth() != 0 {
		t.Errorf("Expected queue depth 0, got %d", m.GetQueueDepth())
	}

	m.UpdateQueueDepth(10)

	if m.GetQueueDepth() != 10 {
		t.Errorf("Expected queue depth 10, got %d", m.GetQueueDepth())
	}

	m.UpdateQueueDepth(5)

	if m.GetQueueDepth() != 5 {
		t.Errorf("Expected queue depth 5, got %d", m.GetQueueDepth())
	}
}

func TestMetricsQueueCapacity(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	m.SetQueueCapacity(1000)

	if m.GetQueueCapacity() != 1000 {
		t.Errorf("Expected queue capacity 1000, got %d", m.GetQueueCapacity())
	}

	m.SetQueueCapacity(2000)

	if m.GetQueueCapacity() != 2000 {
		t.Errorf("Expected queue capacity 2000, got %d", m.GetQueueCapacity())
	}
}

func TestMetricsQueueUtilization(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	m.SetQueueCapacity(100)
	m.UpdateQueueDepth(0)

	utilization := m.GetQueueUtilization()
	if utilization != 0.0 {
		t.Errorf("Expected 0%% utilization, got %.2f%%", utilization)
	}

	m.UpdateQueueDepth(50)
	utilization = m.GetQueueUtilization()
	if utilization != 50.0 {
		t.Errorf("Expected 50%% utilization, got %.2f%%", utilization)
	}

	m.UpdateQueueDepth(100)
	utilization = m.GetQueueUtilization()
	if utilization != 100.0 {
		t.Errorf("Expected 100%% utilization, got %.2f%%", utilization)
	}
}

func TestMetricsProcessingLatency(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	if m.GetAverageProcessingLatency() != 0 {
		t.Errorf("Expected average latency 0, got %d", m.GetAverageProcessingLatency())
	}

	m.RecordProcessingLatency(100)
	m.RecordProcessingLatency(200)
	m.RecordProcessingLatency(300)

	avgLatency := m.GetAverageProcessingLatency()
	if avgLatency != 200 {
		t.Errorf("Expected average latency 200, got %d", avgLatency)
	}
}

func TestMetricsDropRate(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	// No events published yet
	dropRate := m.GetDropRate()
	if dropRate != 0.0 {
		t.Errorf("Expected drop rate 0%%, got %.2f%%", dropRate)
	}

	// Publish 100 events, drop 10
	for i := 0; i < 100; i++ {
		m.IncrementEventsPublished()
	}
	for i := 0; i < 10; i++ {
		m.IncrementEventsDropped()
	}

	dropRate = m.GetDropRate()
	expected := 10.0
	if dropRate != expected {
		t.Errorf("Expected drop rate %.2f%%, got %.2f%%", expected, dropRate)
	}
}

func TestMetricsSkipRate(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	// No events published yet
	skipRate := m.GetSkipRate()
	if skipRate != 0.0 {
		t.Errorf("Expected skip rate 0%%, got %.2f%%", skipRate)
	}

	// Publish 100 events, skip 20
	for i := 0; i < 100; i++ {
		m.IncrementEventsPublished()
	}
	for i := 0; i < 20; i++ {
		m.IncrementEventsSkipped()
	}

	skipRate = m.GetSkipRate()
	expected := (20.0 / 120.0) * 100.0 // 20 out of (100+20) = 16.67%
	if skipRate < expected-0.1 || skipRate > expected+0.1 {
		t.Errorf("Expected skip rate %.2f%%, got %.2f%%", expected, skipRate)
	}
}

func TestMetricsSuccessRate(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	// No events processed yet
	successRate := m.GetSuccessRate()
	if successRate != 100.0 {
		t.Errorf("Expected success rate 100%%, got %.2f%%", successRate)
	}

	// Process 100 events, 5 errors
	for i := 0; i < 100; i++ {
		m.IncrementEventsProcessed()
	}
	for i := 0; i < 5; i++ {
		m.IncrementSubscriberErrors()
	}

	successRate = m.GetSuccessRate()
	expected := 95.0
	if successRate != expected {
		t.Errorf("Expected success rate %.2f%%, got %.2f%%", expected, successRate)
	}
}

func TestMetricsUptime(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	// Sleep for a bit
	time.Sleep(100 * time.Millisecond)

	uptime := m.GetUptime()
	if uptime < 100*time.Millisecond {
		t.Errorf("Expected uptime >= 100ms, got %v", uptime)
	}

	if uptime > 1*time.Second {
		t.Errorf("Expected uptime < 1s, got %v", uptime)
	}
}

func TestMetricsSnapshot(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	m.SetQueueCapacity(1000)
	m.IncrementEventsPublished()
	m.IncrementEventsPublished()
	m.IncrementEventsDropped()
	m.IncrementEventsSkipped()
	m.IncrementEventsProcessed()
	m.IncrementSubscriberErrors()
	m.UpdateQueueDepth(50)
	m.RecordProcessingLatency(100)

	snapshot := m.GetSnapshot()

	if snapshot.EventsPublished != 2 {
		t.Errorf("Expected 2 events published, got %d", snapshot.EventsPublished)
	}

	if snapshot.EventsDropped != 1 {
		t.Errorf("Expected 1 event dropped, got %d", snapshot.EventsDropped)
	}

	if snapshot.EventsSkipped != 1 {
		t.Errorf("Expected 1 event skipped, got %d", snapshot.EventsSkipped)
	}

	if snapshot.EventsProcessed != 1 {
		t.Errorf("Expected 1 event processed, got %d", snapshot.EventsProcessed)
	}

	if snapshot.SubscriberErrors != 1 {
		t.Errorf("Expected 1 subscriber error, got %d", snapshot.SubscriberErrors)
	}

	if snapshot.QueueDepth != 50 {
		t.Errorf("Expected queue depth 50, got %d", snapshot.QueueDepth)
	}

	if snapshot.QueueCapacity != 1000 {
		t.Errorf("Expected queue capacity 1000, got %d", snapshot.QueueCapacity)
	}

	if snapshot.AverageLatencyMicros != 100 {
		t.Errorf("Expected average latency 100, got %d", snapshot.AverageLatencyMicros)
	}

	if snapshot.Timestamp.IsZero() {
		t.Error("Snapshot timestamp should not be zero")
	}
}

func TestMetricsDisabled(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Disable()

	// These should not increment when disabled
	m.IncrementEventsPublished()
	m.IncrementEventsDropped()
	m.IncrementEventsSkipped()
	m.IncrementEventsProcessed()
	m.IncrementSubscriberErrors()
	m.UpdateQueueDepth(10)
	m.RecordProcessingLatency(100)

	// All should still be zero
	if m.GetEventsPublished() != 0 {
		t.Errorf("Expected 0 events published when disabled, got %d", m.GetEventsPublished())
	}

	if m.GetEventsDropped() != 0 {
		t.Errorf("Expected 0 events dropped when disabled, got %d", m.GetEventsDropped())
	}

	if m.GetEventsSkipped() != 0 {
		t.Errorf("Expected 0 events skipped when disabled, got %d", m.GetEventsSkipped())
	}

	if m.GetEventsProcessed() != 0 {
		t.Errorf("Expected 0 events processed when disabled, got %d", m.GetEventsProcessed())
	}

	if m.GetSubscriberErrors() != 0 {
		t.Errorf("Expected 0 subscriber errors when disabled, got %d", m.GetSubscriberErrors())
	}

	if m.GetQueueDepth() != 0 {
		t.Errorf("Expected queue depth 0 when disabled, got %d", m.GetQueueDepth())
	}

	if m.GetAverageProcessingLatency() != 0 {
		t.Errorf("Expected average latency 0 when disabled, got %d", m.GetAverageProcessingLatency())
	}
}

func TestMetricsConcurrency(t *testing.T) {
	m := GetMetrics()
	m.Reset()
	m.Enable()

	numGoroutines := 100
	operationsPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 4) // 4 different operations

	// Concurrent IncrementEventsPublished
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				m.IncrementEventsPublished()
			}
		}()
	}

	// Concurrent IncrementEventsDropped
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				m.IncrementEventsDropped()
			}
		}()
	}

	// Concurrent IncrementEventsProcessed
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				m.IncrementEventsProcessed()
			}
		}()
	}

	// Concurrent RecordProcessingLatency
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				m.RecordProcessingLatency(100)
			}
		}()
	}

	wg.Wait()

	// #nosec G115 -- Test values are small constants, no overflow risk
	expectedPublished := uint64(int64(numGoroutines) * int64(operationsPerGoroutine))

	if m.GetEventsPublished() != expectedPublished {
		t.Errorf("Expected %d events published, got %d", expectedPublished, m.GetEventsPublished())
	}

	if m.GetEventsDropped() != expectedPublished {
		t.Errorf("Expected %d events dropped, got %d", expectedPublished, m.GetEventsDropped())
	}

	if m.GetEventsProcessed() != expectedPublished {
		t.Errorf("Expected %d events processed, got %d", expectedPublished, m.GetEventsProcessed())
	}
}

func TestMetricsReset(t *testing.T) {
	m := GetMetrics()
	m.Enable()

	// Add some metrics
	m.IncrementEventsPublished()
	m.IncrementEventsDropped()
	m.IncrementEventsSkipped()
	m.IncrementEventsProcessed()
	m.IncrementSubscriberErrors()
	m.UpdateQueueDepth(10)
	m.RecordProcessingLatency(100)

	// Reset
	m.Reset()

	// All should be zero
	if m.GetEventsPublished() != 0 {
		t.Errorf("Expected 0 events published after reset, got %d", m.GetEventsPublished())
	}

	if m.GetEventsDropped() != 0 {
		t.Errorf("Expected 0 events dropped after reset, got %d", m.GetEventsDropped())
	}

	if m.GetEventsSkipped() != 0 {
		t.Errorf("Expected 0 events skipped after reset, got %d", m.GetEventsSkipped())
	}

	if m.GetEventsProcessed() != 0 {
		t.Errorf("Expected 0 events processed after reset, got %d", m.GetEventsProcessed())
	}

	if m.GetSubscriberErrors() != 0 {
		t.Errorf("Expected 0 subscriber errors after reset, got %d", m.GetSubscriberErrors())
	}

	if m.GetQueueDepth() != 0 {
		t.Errorf("Expected queue depth 0 after reset, got %d", m.GetQueueDepth())
	}

	if m.GetAverageProcessingLatency() != 0 {
		t.Errorf("Expected average latency 0 after reset, got %d", m.GetAverageProcessingLatency())
	}
}
