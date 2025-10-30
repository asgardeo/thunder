package defaultsubscriber

import (
	"testing"
	"time"

	"github.com/asgardeo/thunder/internal/observability/adapter"
	"github.com/asgardeo/thunder/internal/observability/event"
	"github.com/asgardeo/thunder/internal/observability/formatter"
)

// mockFormatter for testing
type mockFormatter struct {
	formatCalled int
	shouldError  bool
}

func (m *mockFormatter) Format(evt *event.Event) ([]byte, error) {
	m.formatCalled++
	if m.shouldError {
		return nil, &testError{msg: "format error"}
	}
	return []byte(`{"test":"data"}`), nil
}

func (m *mockFormatter) GetName() string {
	return "mock-formatter"
}

// mockAdapter for testing
type mockAdapter struct {
	written     [][]byte
	shouldError bool
}

func (m *mockAdapter) Write(data []byte) error {
	m.written = append(m.written, data)
	if m.shouldError {
		return &testError{msg: "write error"}
	}
	return nil
}

func (m *mockAdapter) Flush() error {
	return nil
}

func (m *mockAdapter) Close() error {
	return nil
}

func (m *mockAdapter) GetName() string {
	return "mock-adapter"
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestNewDefaultSubscriber(t *testing.T) {
	fmt := &mockFormatter{}
	adp := &mockAdapter{}

	sub := NewDefaultSubscriber(fmt, adp)

	if sub.GetID() == "" {
		t.Error("Expected subscriber ID to be generated")
	}

	categories := sub.GetCategories()
	if len(categories) != 1 || categories[0] != event.CategoryAll {
		t.Errorf("Expected subscriber to subscribe to CategoryAll, got %v", categories)
	}
}

func TestNewDefaultSubscriberWithCategories(t *testing.T) {
	fmt := &mockFormatter{}
	adp := &mockAdapter{}

	sub := NewDefaultSubscriberWithCategories(fmt, adp, event.CategoryAuthentication, event.CategoryTokens)

	categories := sub.GetCategories()
	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}
}

func TestNewDefaultSubscriberWithEventTypes(t *testing.T) {
	fmt := &mockFormatter{}
	adp := &mockAdapter{}

	sub := NewDefaultSubscriberWithEventTypes(fmt, adp, event.EventTypeAuthenticationStarted, event.EventTypeTokenIssued)

	// Should derive categories from event types
	categories := sub.GetCategories()
	if len(categories) < 1 {
		t.Error("Expected categories to be derived from event types")
	}
}

func TestSubscriberSideFiltering_CategoryBased(t *testing.T) {
	fmt := &mockFormatter{}
	adp := &mockAdapter{}

	// Create subscriber that only cares about authentication events
	sub := NewDefaultSubscriberWithCategories(fmt, adp, event.CategoryAuthentication)

	// Send authentication event - should be processed
	authEvent := &event.Event{
		EventID:   "auth-1",
		Type:      string(event.EventTypeAuthenticationStarted),
		TraceID:   "trace-1",
		Component: "test",
		Timestamp: time.Now(),
	}

	err := sub.OnEvent(authEvent)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if fmt.formatCalled != 1 {
		t.Errorf("Expected format to be called once, got %d", fmt.formatCalled)
	}

	if len(adp.written) != 1 {
		t.Errorf("Expected 1 write, got %d", len(adp.written))
	}

	// Send token event - should be filtered out
	tokenEvent := &event.Event{
		EventID:   "token-1",
		Type:      string(event.EventTypeTokenIssued),
		TraceID:   "trace-2",
		Component: "test",
		Timestamp: time.Now(),
	}

	err = sub.OnEvent(tokenEvent)
	if err != nil {
		t.Errorf("Expected no error for filtered event, got %v", err)
	}

	// Format should still be called only once (not for filtered event)
	if fmt.formatCalled != 1 {
		t.Errorf("Expected format to still be called once, got %d", fmt.formatCalled)
	}

	if len(adp.written) != 1 {
		t.Errorf("Expected still 1 write, got %d", len(adp.written))
	}
}

func TestSubscriberSideFiltering_EventTypeBased(t *testing.T) {
	fmt := &mockFormatter{}
	adp := &mockAdapter{}

	// Create subscriber that only cares about specific event types
	sub := NewDefaultSubscriberWithEventTypes(
		fmt,
		adp,
		event.EventTypeAuthenticationStarted,
		event.EventTypeAuthenticationCompleted,
	)

	// Send matching event - should be processed
	authStartedEvent := &event.Event{
		EventID:   "auth-1",
		Type:      string(event.EventTypeAuthenticationStarted),
		TraceID:   "trace-1",
		Component: "test",
		Timestamp: time.Now(),
	}

	err := sub.OnEvent(authStartedEvent)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if fmt.formatCalled != 1 {
		t.Errorf("Expected format to be called once, got %d", fmt.formatCalled)
	}

	// Send non-matching event (same category, different type) - should be filtered
	authFailedEvent := &event.Event{
		EventID:   "auth-2",
		Type:      string(event.EventTypeAuthenticationFailed),
		TraceID:   "trace-2",
		Component: "test",
		Timestamp: time.Now(),
	}

	err = sub.OnEvent(authFailedEvent)
	if err != nil {
		t.Errorf("Expected no error for filtered event, got %v", err)
	}

	// Format should still be called only once
	if fmt.formatCalled != 1 {
		t.Errorf("Expected format to still be called once, got %d", fmt.formatCalled)
	}
}

func TestSubscriberSideFiltering_CategoryAll(t *testing.T) {
	fmt := &mockFormatter{}
	adp := &mockAdapter{}

	// Create subscriber that subscribes to all categories
	sub := NewDefaultSubscriber(fmt, adp)

	// Send various events - all should be processed
	events := []*event.Event{
		{
			EventID:   "auth-1",
			Type:      string(event.EventTypeAuthenticationStarted),
			TraceID:   "trace-1",
			Component: "test",
			Timestamp: time.Now(),
		},
		{
			EventID:   "token-1",
			Type:      string(event.EventTypeTokenIssued),
			TraceID:   "trace-2",
			Component: "test",
			Timestamp: time.Now(),
		},
		{
			EventID:   "session-1",
			Type:      string(event.EventTypeSessionCreated),
			TraceID:   "trace-3",
			Component: "test",
			Timestamp: time.Now(),
		},
	}

	for _, evt := range events {
		err := sub.OnEvent(evt)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	if fmt.formatCalled != 3 {
		t.Errorf("Expected format to be called 3 times, got %d", fmt.formatCalled)
	}

	if len(adp.written) != 3 {
		t.Errorf("Expected 3 writes, got %d", len(adp.written))
	}
}

func TestSubscriberDynamicFiltering_AddRemoveEventTypes(t *testing.T) {
	fmt := &mockFormatter{}
	adp := &mockAdapter{}

	// Create subscriber with no event type filters (subscribes to all in a category)
	sub := NewDefaultSubscriberWithCategories(fmt, adp, event.CategoryAuthentication)

	// Add specific event type filter
	sub.AddEventTypeFilter(event.EventTypeAuthenticationStarted)

	// Now it should only process authentication started events
	authStartedEvent := &event.Event{
		EventID:   "auth-1",
		Type:      string(event.EventTypeAuthenticationStarted),
		TraceID:   "trace-1",
		Component: "test",
		Timestamp: time.Now(),
	}

	err := sub.OnEvent(authStartedEvent)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if fmt.formatCalled != 1 {
		t.Errorf("Expected format to be called once, got %d", fmt.formatCalled)
	}

	// Other auth events should be filtered
	authFailedEvent := &event.Event{
		EventID:   "auth-2",
		Type:      string(event.EventTypeAuthenticationFailed),
		TraceID:   "trace-2",
		Component: "test",
		Timestamp: time.Now(),
	}

	err = sub.OnEvent(authFailedEvent)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should still be 1
	if fmt.formatCalled != 1 {
		t.Errorf("Expected format to still be called once, got %d", fmt.formatCalled)
	}

	// Remove the filter
	sub.RemoveEventTypeFilter(event.EventTypeAuthenticationStarted)

	// Now should process all auth events again (no event type filters)
	err = sub.OnEvent(authFailedEvent)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if fmt.formatCalled != 2 {
		t.Errorf("Expected format to be called twice after removing filter, got %d", fmt.formatCalled)
	}
}

func TestSubscriberSetCategories(t *testing.T) {
	fmt := &mockFormatter{}
	adp := &mockAdapter{}

	sub := NewDefaultSubscriber(fmt, adp)

	// Initially subscribes to all
	if len(sub.GetCategories()) != 1 || sub.GetCategories()[0] != event.CategoryAll {
		t.Error("Expected initial subscription to CategoryAll")
	}

	// Change to specific categories
	sub.SetCategories(event.CategoryAuthentication, event.CategoryTokens)

	categories := sub.GetCategories()
	if len(categories) != 2 {
		t.Errorf("Expected 2 categories after SetCategories, got %d", len(categories))
	}
}

// Verify subscriber interface is fully implemented
var _ formatter.Formatter = (*mockFormatter)(nil)
var _ adapter.OutputAdapter = (*mockAdapter)(nil)
