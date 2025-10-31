# Observability System

A comprehensive observability platform for Thunder providing event logging, metrics collection, distributed tracing, and health monitoring for authentication and authorization flows.

## Architecture

The observability system uses an **Event Bus** pattern to capture and process events throughout the auth/authz flows with minimal performance impact.

**Key Features:**
- ✅ **Async & Non-blocking** - Events published in goroutines, main thread never blocks
- ✅ **No Queue** - Direct event delivery to subscribers
- ✅ **Category-based routing** - Subscribers filter by event categories
- ✅ **Isolated failures** - Subscriber panics don't affect others

```
Auth/AuthZ Components → Event Bus (Publisher) → Subscribers (async) → Formatters → Output Adapters
                                    ↓
                            Each subscriber in own goroutine
```

## Quick Start

### 1. Publishing Events

```go
import (
    "github.com/asgardeo/thunder/internal/observability"
    "github.com/asgardeo/thunder/internal/observability/event"
)

func handleRequest(ctx context.Context) {
    // Ensure trace ID exists
    ctx = observability.EnsureTraceID(ctx)
    traceID := observability.GetTraceID(ctx)

    // Create and publish event
    evt := event.NewEvent(traceID, string(event.EventTypeAuthenticationStarted), "MyComponent")
    evt.WithData("user_id", "user123")
       .WithStatus(event.StatusInProgress)
       .WithData("message", "Authentication started")

    observability.GetService().PublishEvent(evt)
}
```

### 2. Custom Subscribers

```go
import (
    "github.com/asgardeo/thunder/internal/observability/publisher"
    "github.com/asgardeo/thunder/internal/observability/subscriber/defaultsubscriber"
    "github.com/asgardeo/thunder/internal/observability/formatter/json"
    "github.com/asgardeo/thunder/internal/observability/adapter/file"
    "github.com/asgardeo/thunder/internal/observability/event"
)

// Subscribe to specific event types only
jsonFormatter := jsonformatter.NewJSONFormatter()
fileAdapter, _ := file.NewFileAdapter("/path/to/custom-events.log")
subscriber := defaultsubscriber.NewDefaultSubscriber(jsonFormatter, fileAdapter)

// Option 1: Subscribe to specific event types
publisher.GetPublisher().Subscribe(subscriber,
    event.EventTypeAuthenticationStarted,
    event.EventTypeAuthenticationCompleted,
    event.EventTypeAuthenticationFailed,
)

// Option 2: Subscribe to all events
publisher.GetPublisher().SubscribeAll(subscriber)
```

## Package Structure

```
observability/
├── event/                      # Event model and types
│   ├── event.go               # Core event structure with fluent API
│   └── constants.go           # Event types and component names
├── publisher/                  # Event publisher
│   └── publisher.go           # Singleton publisher with in-memory queue
├── subscriber/                 # Subscriber interfaces and implementations
│   ├── subscriber.go          # Subscriber interface
│   └── defaultsubscriber/     # Default subscriber implementation
│       └── default_subscriber.go
├── formatter/                  # Event formatters
│   ├── formatter.go           # Formatter interface
│   ├── json/                  # JSON formatter
│   │   └── json_formatter.go
│   └── csv/                   # CSV formatter
│       └── csv_formatter.go
├── adapter/                    # Output adapters
│   ├── adapter.go             # OutputAdapter interface
│   ├── file/                  # File output
│   │   └── file_adapter.go
│   └── console/               # Console output
│       └── console_adapter.go
├── examples/                   # Integration examples
│   └── integration_examples.go
├── context.go                 # Context utilities for trace ID
├── service.go                 # Main analytics service
└── README.md                  # This file
```

## Event Types

### Authorization Events
- `AUTHORIZATION_STARTED` - Authorization request received
- `AUTHORIZATION_VALIDATED` - Request validated
- `AUTHORIZATION_CODE_GENERATED` - Auth code created
- `AUTHORIZATION_COMPLETED` - Flow complete
- `AUTHORIZATION_FAILED` - Flow failed

### Authentication Events
- `AUTHENTICATION_STARTED` - Auth flow begins
- `CREDENTIALS_AUTH_STARTED` - Username/password auth
- `CREDENTIALS_AUTH_COMPLETED` - Credentials verified
- `CREDENTIALS_AUTH_FAILED` - Credentials invalid
- `OTP_SENT` - OTP sent to user
- `OTP_VERIFIED` - OTP validated
- `SOCIAL_AUTH_STARTED` - Social login begins
- `SOCIAL_AUTH_COMPLETED` - Social login succeeds
- `AUTHENTICATION_COMPLETED` - Auth complete
- `AUTHENTICATION_FAILED` - Auth failed

### Token Events
- `TOKEN_REQUEST_RECEIVED` - Token request received
- `TOKEN_REQUEST_VALIDATED` - Request validated
- `PKCE_VALIDATED` - PKCE validation successful
- `ACCESS_TOKEN_GENERATED` - Access token created
- `ID_TOKEN_GENERATED` - ID token created
- `REFRESH_TOKEN_GENERATED` - Refresh token created
- `TOKEN_ISSUED` - Tokens issued
- `TOKEN_REQUEST_FAILED` - Request failed

### Flow Events
- `FLOW_STARTED` - Flow execution begins
- `FLOW_NODE_EXECUTION_STARTED` - Node execution starts
- `FLOW_NODE_EXECUTION_COMPLETED` - Node completes
- `FLOW_USER_INPUT_REQUIRED` - User input needed
- `FLOW_COMPLETED` - Flow succeeds
- `FLOW_FAILED` - Flow fails

## Integration Examples

See `examples/integration_examples.go` for complete examples of:
- Authorization handler integration
- Token handler integration
- Flow execution integration
- Authentication service integration

## Output Format

### JSON (Default)
```json
{
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_id": "123e4567-e89b-12d3-a456-426614174000",
  "event_type": "AUTHENTICATION_COMPLETED",
  "timestamp": "2025-10-21T10:15:30Z",
  "component": "AuthenticationService",
  "user_id": "user_123",
  "client_id": "client_456",
  "status": "SUCCESS",
  "duration_ms": 245
}
```

### CSV
```csv
TraceID,EventID,EventType,Timestamp,Component,UserID,ClientID,Status,DurationMS
550e8400...,123e4567...,AUTHENTICATION_COMPLETED,2025-10-21T10:15:30Z,AuthenticationService,user_123,client_456,SUCCESS,245
```

## Querying Events

### Filter by Trace ID
```bash
# Using jq
cat analytics.log | jq 'select(.trace_id == "550e8400-e29b-41d4-a716-446655440000")'

# Get event timeline
cat analytics.log | jq 'select(.trace_id == "550e8400...") | {timestamp, event_type, status}'
```

### Get Failed Authentications
```bash
cat analytics.log | jq 'select(.event_type | contains("FAILED"))'
```

### Calculate Average Duration
```bash
cat analytics.log | jq -s '[.[] | select(.event_type == "TOKEN_ISSUED") | .duration_ms] | add / length'
```

## Performance

- **Non-blocking**: Event publishing returns immediately, never blocks the main thread
- **Async processing**: Each subscriber runs in its own goroutine
- **No queue overhead**: Direct event delivery without buffering
- **Graceful degradation**: Events skipped if no subscribers are interested
- **Parallel processing**: All subscribers process events simultaneously
- **Isolated failures**: One subscriber's failure doesn't affect others

## Extending

### Custom Formatter

```go
type MyFormatter struct{}

func (mf *MyFormatter) Format(evt *event.Event) ([]byte, error) {
    // Custom formatting logic
    return []byte(fmt.Sprintf("%s: %s", evt.EventType, evt.Message)), nil
}

func (mf *MyFormatter) GetName() string {
    return "MyFormatter"
}
```

### Custom Output Adapter

```go
type MyAdapter struct{}

func (ma *MyAdapter) Write(data []byte) error {
    // Custom output logic (send to external service, database, etc.)
    return nil
}

func (ma *MyAdapter) Flush() error {
    return nil
}

func (ma *MyAdapter) Close() error {
    return nil
}

func (ma *MyAdapter) GetName() string {
    return "MyAdapter"
}
```

## Shutdown

Always shut down the analytics service gracefully:

```go
// In main.go or shutdown handler
defer observability.GetService().Shutdown()
```

This ensures:
- Pending events are processed
- Buffers are flushed
- Files are closed properly
- Resources are released

## Future Enhancements

- [ ] Event filtering by type/component/status
- [ ] Event sampling (percentage/rate-based)
- [ ] Batch processing for efficiency
- [ ] Output compression
- [ ] Automatic log rotation
- [ ] Remote adapters (Azure Event Hub, OpenSearch, Kafka)
- [ ] Real-time metrics aggregation
- [ ] Alerting on specific event patterns
- [ ] Distributed tracing with spans
- [ ] OpenTelemetry integration

## Documentation

See `/docs/content/observability-system.md` for comprehensive documentation.
