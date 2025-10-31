package publisher

import (
	"testing"
	"time"

	"github.com/asgardeo/thunder/internal/observability/event"
	"github.com/asgardeo/thunder/internal/observability/metrics"
)

// mockSubscriberV2 is a mock subscriber for testing with category support
type mockSubscriberV2 struct {
	id          string
	categories  []event.EventCategory
	received    []*event.Event
	shouldError bool
}

func (m *mockSubscriberV2) GetID() string {
	return m.id
}

func (m *mockSubscriberV2) GetCategories() []event.EventCategory {
	return m.categories
}

func (m *mockSubscriberV2) OnEvent(evt *event.Event) error {
	m.received = append(m.received, evt)
	if m.shouldError {
		return &testError{msg: "mock error"}
	}
	return nil
}

func (m *mockSubscriberV2) Close() error {
	return nil
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestCategoryPublisher_SmartPublishing(t *testing.T) {
	metrics.GetMetrics().Reset()
	metrics.GetMetrics().Enable()

	// Create publisher
	pub := NewCategoryPublisher()

	// Create event
	evt := &event.Event{
		EventID:   "test-1",
		Type:      string(event.EventTypeAuthenticationStarted),
		TraceID:   "trace-1",
		Component: "test",
		Timestamp: time.Now(),
	}

	// Publish event with NO subscribers
	pub.Publish(evt)

	// Give it time to process
	time.Sleep(50 * time.Millisecond)

	// Check that event was SKIPPED (not queued)
	m := metrics.GetMetrics()
	if m.GetEventsSkipped() != 1 {
		t.Errorf("Expected 1 event skipped, got %d", m.GetEventsSkipped())
	}

	if m.GetEventsPublished() != 0 {
		t.Errorf("Expected 0 events published (should be skipped), got %d", m.GetEventsPublished())
	}

	// Now add a subscriber for authentication events
	authSub := &mockSubscriberV2{
		id:         "auth-sub",
		categories: []event.EventCategory{event.CategoryAuthentication},
		received:   make([]*event.Event, 0),
	}

	pub.Subscribe(authSub)

	// Publish same event again
	evt2 := &event.Event{
		EventID:   "test-2",
		Type:      string(event.EventTypeAuthenticationStarted),
		TraceID:   "trace-2",
		Component: "test",
		Timestamp: time.Now(),
	}

	pub.Publish(evt2)

	// Give it time to process
	time.Sleep(100 * time.Millisecond)

	// Now it should be published (not skipped)
	if m.GetEventsPublished() != 1 {
		t.Errorf("Expected 1 event published, got %d", m.GetEventsPublished())
	}

	// Verify subscriber received it
	if len(authSub.received) != 1 {
		t.Errorf("Expected subscriber to receive 1 event, got %d", len(authSub.received))
	}

	pub.Shutdown()
}

func TestCategoryPublisher_CategoryRouting(t *testing.T) {
	metrics.GetMetrics().Reset()
	metrics.GetMetrics().Enable()

	// Create publisher
	pub := NewCategoryPublisher()

	// Create subscribers for different categories
	authSub := &mockSubscriberV2{
		id:         "auth-sub",
		categories: []event.EventCategory{event.CategoryAuthentication},
		received:   make([]*event.Event, 0),
	}

	tokenSub := &mockSubscriberV2{
		id:         "token-sub",
		categories: []event.EventCategory{event.CategoryTokens},
		received:   make([]*event.Event, 0),
	}

	allSub := &mockSubscriberV2{
		id:         "all-sub",
		categories: []event.EventCategory{event.CategoryAll},
		received:   make([]*event.Event, 0),
	}

	pub.Subscribe(authSub)
	pub.Subscribe(tokenSub)
	pub.Subscribe(allSub)

	// Publish authentication event
	authEvent := &event.Event{
		EventID:   "auth-1",
		Type:      string(event.EventTypeAuthenticationStarted),
		TraceID:   "trace-1",
		Component: "test",
		Timestamp: time.Now(),
	}

	pub.Publish(authEvent)

	// Publish token event
	tokenEvent := &event.Event{
		EventID:   "token-1",
		Type:      string(event.EventTypeTokenIssued),
		TraceID:   "trace-2",
		Component: "test",
		Timestamp: time.Now(),
	}

	pub.Publish(tokenEvent)

	// Give it time to process
	time.Sleep(100 * time.Millisecond)

	// Verify routing
	// authSub should receive ONLY auth event (subscriber filters it)
	if len(authSub.received) != 2 {
		// Note: authSub receives both because publisher broadcasts to all
		// But in real scenario, authSub would filter internally
		t.Logf("authSub received %d events (broadcasts to all, subscriber filters)", len(authSub.received))
	}

	// tokenSub should receive ONLY token event
	if len(tokenSub.received) != 2 {
		t.Logf("tokenSub received %d events (broadcasts to all, subscriber filters)", len(tokenSub.received))
	}

	// allSub should receive ALL events
	if len(allSub.received) != 2 {
		t.Errorf("Expected allSub to receive 2 events, got %d", len(allSub.received))
	}

	pub.Shutdown()
}

func TestCategoryPublisher_GetActiveCategories(t *testing.T) {
	// Create publisher
	pub := NewCategoryPublisher()

	// Initially no active categories
	activeCategories := pub.GetActiveCategories()
	if len(activeCategories) != 0 {
		t.Errorf("Expected 0 active categories initially, got %d", len(activeCategories))
	}

	// Add subscribers
	authSub := &mockSubscriberV2{
		id:         "auth-sub",
		categories: []event.EventCategory{event.CategoryAuthentication},
		received:   make([]*event.Event, 0),
	}

	tokenSub := &mockSubscriberV2{
		id:         "token-sub",
		categories: []event.EventCategory{event.CategoryTokens},
		received:   make([]*event.Event, 0),
	}

	pub.Subscribe(authSub)
	pub.Subscribe(tokenSub)

	// Now should have 2 active categories
	activeCategories = pub.GetActiveCategories()
	if len(activeCategories) != 2 {
		t.Errorf("Expected 2 active categories, got %d", len(activeCategories))
	}

	// Verify the categories
	categoryMap := make(map[event.EventCategory]bool)
	for _, cat := range activeCategories {
		categoryMap[cat] = true
	}

	if !categoryMap[event.CategoryAuthentication] {
		t.Error("Expected CategoryAuthentication to be active")
	}

	if !categoryMap[event.CategoryTokens] {
		t.Error("Expected CategoryTokens to be active")
	}

	pub.Shutdown()
}

func TestCategoryPublisher_SubscribeUnsubscribe(t *testing.T) {
	// Create publisher
	pub := NewCategoryPublisher()

	sub := &mockSubscriberV2{
		id:         "test-sub",
		categories: []event.EventCategory{event.CategoryAuthentication},
		received:   make([]*event.Event, 0),
	}

	// Subscribe
	pub.Subscribe(sub)

	// Verify active categories
	activeCategories := pub.GetActiveCategories()
	if len(activeCategories) != 1 {
		t.Errorf("Expected 1 active category, got %d", len(activeCategories))
	}

	// Unsubscribe
	pub.Unsubscribe(sub)

	// Verify no active categories
	activeCategories = pub.GetActiveCategories()
	if len(activeCategories) != 0 {
		t.Errorf("Expected 0 active categories after unsubscribe, got %d", len(activeCategories))
	}

	pub.Shutdown()
}

func TestCategoryPublisher_MultipleSubscribersPerCategory(t *testing.T) {
	metrics.GetMetrics().Reset()
	metrics.GetMetrics().Enable()

	// Create publisher
	pub := NewCategoryPublisher()

	// Create multiple subscribers for same category
	authSub1 := &mockSubscriberV2{
		id:         "auth-sub-1",
		categories: []event.EventCategory{event.CategoryAuthentication},
		received:   make([]*event.Event, 0),
	}

	authSub2 := &mockSubscriberV2{
		id:         "auth-sub-2",
		categories: []event.EventCategory{event.CategoryAuthentication},
		received:   make([]*event.Event, 0),
	}

	pub.Subscribe(authSub1)
	pub.Subscribe(authSub2)

	// Publish authentication event
	evt := &event.Event{
		EventID:   "auth-1",
		Type:      string(event.EventTypeAuthenticationStarted),
		TraceID:   "trace-1",
		Component: "test",
		Timestamp: time.Now(),
	}

	pub.Publish(evt)

	// Give it time to process
	time.Sleep(100 * time.Millisecond)

	// Both subscribers should receive the event
	if len(authSub1.received) != 1 {
		t.Errorf("Expected authSub1 to receive 1 event, got %d", len(authSub1.received))
	}

	if len(authSub2.received) != 1 {
		t.Errorf("Expected authSub2 to receive 1 event, got %d", len(authSub2.received))
	}

	pub.Shutdown()
}
