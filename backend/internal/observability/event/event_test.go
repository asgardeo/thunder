package event

import (
	"testing"
	"time"
)

func TestNewEvent(t *testing.T) {
	traceID := "trace-123"
	eventType := string(EventTypeAuthenticationStarted)
	component := "TestComponent"

	evt := NewEvent(traceID, eventType, component)

	if evt == nil {
		t.Fatal("NewEvent returned nil")
	}

	if evt.TraceID != traceID {
		t.Errorf("Expected TraceID %s, got %s", traceID, evt.TraceID)
	}

	if evt.Type != eventType {
		t.Errorf("Expected Type %s, got %s", eventType, evt.Type)
	}

	if evt.Component != component {
		t.Errorf("Expected Component %s, got %s", component, evt.Component)
	}

	if evt.EventID == "" {
		t.Error("EventID should not be empty")
	}

	if evt.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	if evt.Status != StatusInProgress {
		t.Errorf("Expected Status %s, got %s", StatusInProgress, evt.Status)
	}

	if evt.Data == nil {
		t.Error("Data map should be initialized")
	}
}

func TestEventBuilderPattern(t *testing.T) {
	evt := NewEvent("trace-123", string(EventTypeTokenIssued), "TokenHandler")

	result := evt.
		WithStatus(StatusSuccess).
		WithData(DataKey.UserID, "user-456").
		WithData(DataKey.ClientID, "client-789").
		WithData(DataKey.SessionID, "session-abc").
		WithData(DataKey.Message, "Token issued successfully").
		WithData(DataKey.DurationMs, 500).
		WithData(DataKey.IPAddress, "192.168.1.1").
		WithData(DataKey.UserAgent, "Mozilla/5.0")

	if result != evt {
		t.Error("Builder methods should return the same event instance")
	}

	if evt.Status != StatusSuccess {
		t.Errorf("Expected Status %s, got %s", StatusSuccess, evt.Status)
	}

	if evt.Data["user_id"] != "user-456" {
		t.Errorf("Expected Data[user_id] %s, got %v", "user-456", evt.Data["user_id"])
	}

	if evt.Data["client_id"] != "client-789" {
		t.Errorf("Expected Data[client_id] %s, got %v", "client-789", evt.Data["client_id"])
	}

	if evt.Data["session_id"] != "session-abc" {
		t.Errorf("Expected Data[session_id] %s, got %v", "session-abc", evt.Data["session_id"])
	}

	if evt.Data["message"] != "Token issued successfully" {
		t.Errorf("Expected Data[message] %s, got %v", "Token issued successfully", evt.Data["message"])
	}

	if evt.Data["duration_ms"] != 500 {
		t.Errorf("Expected Data[duration_ms] %d, got %v", 500, evt.Data["duration_ms"])
	}

	if evt.Data["ip_address"] != "192.168.1.1" {
		t.Errorf("Expected Data[ip_address] %s, got %v", "192.168.1.1", evt.Data["ip_address"])
	}

	if evt.Data["user_agent"] != "Mozilla/5.0" {
		t.Errorf("Expected Data[user_agent] %s, got %v", "Mozilla/5.0", evt.Data["user_agent"])
	}
}

func TestEventWithDataMap(t *testing.T) {
	evt := NewEvent("trace-123", "user.created", "UserService")

	data := map[string]interface{}{
		"user_id":    "user-123",
		"email":      "user@example.com",
		"created_at": "2025-10-23T10:00:00Z",
		"roles":      []string{"admin", "user"},
	}

	evt.WithDataMap(data)

	if evt.Data["user_id"] != "user-123" {
		t.Error("WithDataMap should set user_id")
	}

	if evt.Data["email"] != "user@example.com" {
		t.Error("WithDataMap should set email")
	}
}

func TestEventValidate(t *testing.T) {
	tests := []struct {
		name    string
		event   *Event
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			event: &Event{
				TraceID:   "trace-123",
				EventID:   "event-456",
				Type:      string(EventTypeAuthenticationStarted),
				Component: "TestComponent",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "nil event",
			event:   nil,
			wantErr: true,
			errMsg:  "event is nil",
		},
		{
			name: "missing trace ID",
			event: &Event{
				EventID:   "event-456",
				Type:      string(EventTypeAuthenticationStarted),
				Component: "TestComponent",
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "trace_id is required",
		},
		{
			name: "missing event ID",
			event: &Event{
				TraceID:   "trace-123",
				Type:      string(EventTypeAuthenticationStarted),
				Component: "TestComponent",
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "event_id is required",
		},
		{
			name: "missing event type",
			event: &Event{
				TraceID:   "trace-123",
				EventID:   "event-456",
				Component: "TestComponent",
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "missing component",
			event: &Event{
				TraceID:   "trace-123",
				EventID:   "event-456",
				Type:      string(EventTypeAuthenticationStarted),
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "component is required",
		},
		{
			name: "missing timestamp",
			event: &Event{
				TraceID:   "trace-123",
				EventID:   "event-456",
				Type:      string(EventTypeAuthenticationStarted),
				Component: "TestComponent",
			},
			wantErr: true,
			errMsg:  "timestamp is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestEventDataNilSafety(t *testing.T) {
	evt := &Event{
		TraceID:   "trace-123",
		EventID:   "event-456",
		Type:      string(EventTypeAuthenticationStarted),
		Component: "TestComponent",
		Timestamp: time.Now(),
		Data:      nil, // Explicitly nil
	}

	// Should initialize map if nil
	evt.WithData(DataKey.Key, "value")

	if evt.Data == nil {
		t.Error("WithData should initialize nil Data map")
	}

	if evt.Data["key"] != "value" {
		t.Errorf("Expected Data[key] %s, got %v", "value", evt.Data["key"])
	}
}

func TestEventTypeConstants(t *testing.T) {
	// Just verify some key constants exist
	if EventTypeAuthenticationStarted == "" {
		t.Error("EventTypeAuthenticationStarted should not be empty")
	}

	if EventTypeTokenIssued == "" {
		t.Error("EventTypeTokenIssued should not be empty")
	}

	if EventTypeAuthorizationStarted == "" {
		t.Error("EventTypeAuthorizationStarted should not be empty")
	}
}

func TestStatusConstants(t *testing.T) {
	// Verify status constants exist
	if StatusSuccess == "" {
		t.Error("StatusSuccess should not be empty")
	}

	if StatusFailure == "" {
		t.Error("StatusFailure should not be empty")
	}

	if StatusInProgress == "" {
		t.Error("StatusInProgress should not be empty")
	}

	if StatusPending == "" {
		t.Error("StatusPending should not be empty")
	}
}
