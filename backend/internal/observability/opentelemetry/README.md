# OpenTelemetry Integration

This package provides OpenTelemetry (OTel) integration for the Thunder observability system.

## Architecture

The OTel integration follows a **Subscriber Pattern** instead of using formatters and adapters:

```
Event → Publisher → OTel Subscriber → OTel SDK → Exporter → Backend
                          ↓
                    (No Formatter!)
                    (No Adapter!)
                          ↓
                   Span Management
```

### Why No Formatter/Adapter?

- **Formatters** produce bytes (JSON, CSV)
- **OTel** produces spans (in-memory objects)
- **OTel SDK** handles serialization internally
- **Subscriber** maintains stateful span hierarchies across multiple events

## How It Works

### 1. Event Publishing (Your Code)
```go
// Publish authentication started event
evt := event.NewEvent(flowID, string(event.EventTypeFlowStarted), "FlowService")
evt.WithData(event.DataKey.FlowID, flowID)
observability.GetService().PublishEvent(evt)
```

### 2. OTel Subscriber Receives Event
```go
// OTel subscriber converts event to span
func (o *OTelSubscriber) handleFlowStarted(evt *event.Event) {
    ctx, span := o.tracer.Start(ctx, "authentication.flow")
    span.SetAttributes(...)
    o.storeSpan(evt.TraceID, span, ctx) // Keep span open!
}
```

### 3. Later Events Complete the Span
```go
// When flow completes
func (o *OTelSubscriber) handleFlowCompleted(evt *event.Event) {
    span := o.getAndRemoveSpan(evt.TraceID)
    span.SetAttributes(...)
    span.End() // Close the span
}
```

### 4. OTel SDK Sends to Backend
The OTel SDK automatically batches and exports spans to Jaeger/Tempo/etc.

## Configuration

Add to your `observability.yaml` or runtime config:

```yaml
observability:
  enabled: true

  # Traditional output (file/console) - still works!
  output:
    type: "file"
    format: "json"
    file:
      path: "/var/log/thunder/analytics.log"

  # OpenTelemetry configuration (NEW!)
  opentelemetry:
    enabled: true
    exporter_type: "otlp"           # "otlp" or "stdout"
    otlp_endpoint: "localhost:4317" # Jaeger/Tempo collector
    service_name: "thunder-iam"
    service_version: "1.0.0"
    environment: "production"
    sample_rate: 1.0                # 1.0 = sample all traces
```

## Supported Backends

### 1. Jaeger (Recommended for Development)
```bash
# Run Jaeger all-in-one
docker run -d --name jaeger \
  -p 4317:4317 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest

# View traces at http://localhost:16686
```

### 2. Grafana Tempo (Recommended for Production)
```bash
# Run Tempo with docker-compose
# See: https://grafana.com/docs/tempo/latest/setup/docker-compose/
```

### 3. Stdout (For Testing)
```yaml
opentelemetry:
  enabled: true
  exporter_type: "stdout"  # Prints spans to console
```

## Trace Hierarchy

The OTel subscriber creates the following span hierarchy:

```
Root Span: authentication.flow
├── Span: oauth.authorization
│   └── Event: authorization_code_generated
├── Span: flow.node.task_execution
│   ├── Span: auth.credentials
│   │   ├── Attribute: username
│   │   ├── Attribute: user_id
│   │   └── Duration: 100ms
│   └── Attribute: node_type
└── Span: oauth.token
    ├── Event: authorization_code_validated
    ├── Event: pkce_validated
    ├── Event: access_token_generated
    ├── Event: id_token_generated
    └── Duration: 50ms
```

## Event Type Mapping

| Event Type | OTel Action | Span Name |
|-----------|-------------|-----------|
| `FLOW_STARTED` | Create root span | `authentication.flow` |
| `FLOW_COMPLETED` | End root span | - |
| `FLOW_NODE_EXECUTION_STARTED` | Create child span | `flow.node.{type}` |
| `FLOW_NODE_EXECUTION_COMPLETED` | End child span | - |
| `CREDENTIALS_AUTH_STARTED` | Create auth span | `auth.credentials` |
| `CREDENTIALS_AUTH_COMPLETED` | End auth span | - |
| `AUTHORIZATION_STARTED` | Create authz span | `oauth.authorization` |
| `AUTHORIZATION_COMPLETED` | End authz span | - |
| `TOKEN_REQUEST_RECEIVED` | Create token span | `oauth.token` |
| `TOKEN_ISSUED` | End token span | - |

## Benefits

### Distributed Tracing
- Track authentication across multiple services
- See the complete user journey
- Identify performance bottlenecks

### Visual Analysis
- View traces in Jaeger/Grafana UI
- See span timings and relationships
- Filter by service, operation, tags

### No Code Changes
- Just enable in configuration
- Existing event publishing works as-is
- Runs alongside file/console output

## Example: Complete Authentication Trace

```
Trace ID: abc123-456-def
Service: thunder-iam
Duration: 250ms

├─ authentication.flow (250ms)
│  ├─ oauth.authorization (50ms)
│  │  └─ Events: code_generated
│  ├─ flow.node.task_execution (150ms)
│  │  └─ auth.credentials (140ms)
│  │     ├─ username: john@example.com
│  │     ├─ user_id: user-789
│  │     └─ status: success
│  └─ oauth.token (45ms)
│     ├─ grant_type: authorization_code
│     ├─ Events: code_validated, pkce_validated
│     ├─ Events: access_token_generated, id_token_generated
│     └─ status: success
```

## Troubleshooting

### Spans Not Appearing in Backend

1. **Check configuration**:
   ```bash
   # Verify endpoint is accessible
   nc -zv localhost 4317
   ```

2. **Enable stdout exporter** to see spans locally:
   ```yaml
   exporter_type: "stdout"
   ```

3. **Check logs**:
   ```bash
   grep "OpenTelemetry" /var/log/thunder/server.log
   ```

### Orphaned Spans

If you see warnings about "orphaned spans" during shutdown:
- Some events started but never completed
- Check for missing COMPLETED/FAILED events
- Review error handling in your code

### High Memory Usage

If many spans are kept in memory:
- Check `sample_rate` (reduce from 1.0 to 0.1 for 10% sampling)
- Ensure flows complete (check for stuck flows)
- Review span retention in your backend

## Development

### Running Tests
```bash
cd backend
go test ./internal/observability/subscriber/otelsubscriber/...
```

### Adding New Event Types

1. Add event constant in `event/constants.go`
2. Add handler in `otelsubscriber/otel_subscriber.go`:
   ```go
   case event.EventTypeMyNewEvent:
       return o.handleMyNewEvent(evt)
   ```
3. Implement handler following existing patterns

## References

- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Grafana Tempo](https://grafana.com/docs/tempo/latest/)
- [Thunder Observability Architecture](../README.md)
