# Observability Configuration Guide

This guide explains how to configure Thunder's observability system to enable distributed tracing, logging, and analytics.

## Overview

Thunder's observability system provides:
- **Distributed Tracing** via OpenTelemetry (Jaeger, Tempo, etc.)
- **Console Logging** for development and debugging
- **File-based Analytics** for audit trails and data analysis

## Configuration Location

The observability system is configured in `deployment.yaml` under the `observability` section.

## Configuration Structure

```yaml
observability:
  enabled: true                    # Master switch for observability
  
  output:
    console:
      enabled: true                 # Enable console logging
      format: json                  # "json" or "text"
      categories: []                # Event categories to log
    
    opentelemetry:
      enabled: true                 # Enable OpenTelemetry tracing
      exporter_type: "otlp"         # "otlp" or "stdout"
      otlp_endpoint: "localhost:4317"
      service_name: "thunder-iam"
      service_version: "1.0.0"
      environment: "development"
      sample_rate: 1.0
      insecure: true
      categories: []                # Event categories to trace
    
    file:
      enabled: false                # Enable file-based logging
      file_path: "repository/logs/observability.log"
      format: json
      categories: []
  
  failure_mode: "graceful"          # "graceful" or "strict"
```

## Configuration Options

### Master Switch

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `true` | Global enable/disable for all observability features |

### Console Output

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `console.enabled` | boolean | `true` | Enable console logging |
| `console.format` | string | `"json"` | Output format: `"json"` or `"text"` |
| `console.categories` | array | `[]` | Event categories to log (empty = all) |

**Example:**
```yaml
# File: repository/conf/deployment.yaml
observability:
  enabled: true
  output:
    console:
      enabled: true
      format: json
      categories: ["observability.authentication", "observability.flows"]
```

### OpenTelemetry (Distributed Tracing)

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `opentelemetry.enabled` | boolean | `false` | Enable OpenTelemetry tracing |
| `opentelemetry.exporter_type` | string | `"otlp"` | Exporter type: `"otlp"` or `"stdout"` |
| `opentelemetry.otlp_endpoint` | string | `"localhost:4317"` | OTLP collector endpoint (gRPC) |
| `opentelemetry.service_name` | string | `"thunder-iam"` | Service name in traces |
| `opentelemetry.service_version` | string | `"1.0.0"` | Service version tag |
| `opentelemetry.environment` | string | `"development"` | Environment: `"production"`, `"staging"`, `"development"` |
| `opentelemetry.sample_rate` | float | `1.0` | Sampling rate: `0.0` (none) to `1.0` (100%) |
| `opentelemetry.insecure` | boolean | `false` | Allow non-TLS connections |
| `opentelemetry.categories` | array | `[]` | Event categories to trace (empty = all) |

**Example for Jaeger:**
```yaml
# File: repository/conf/deployment.yaml
observability:
  enabled: true
  output:
    opentelemetry:
      enabled: true
      exporter_type: "otlp"
      otlp_endpoint: "localhost:4317"      # Jaeger's OTLP port
      service_name: "thunder-iam"
      service_version: "1.0.0"
      environment: "production"
      sample_rate: 0.1                     # Sample 10% of requests
      insecure: false                      # Use TLS
      categories: ["observability.all"]
```

**Example for stdout (development):**
```yaml
# File: repository/conf/deployment.yaml
observability:
  enabled: true
  output:
    opentelemetry:
      enabled: true
      exporter_type: "stdout"              # Print traces to console
      service_name: "thunder-iam-dev"
      environment: "development"
      sample_rate: 1.0                     # Trace everything
```

### File-based Analytics

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `file.enabled` | boolean | `false` | Enable file-based logging |
| `file.file_path` | string | `"logs/observability/observability.log"` | Path to the log file |
| `file.format` | string | `"json"` | Output format: `"json"` |
| `file.categories` | array | `[]` | Event categories to log (empty = all) |

**Example:**
```yaml
# File: repository/conf/deployment.yaml
observability:
  enabled: true
  output:
    file:
      enabled: true
      file_path: "repository/logs/observability.log"
      format: json
      categories: ["observability.authorization", "observability.authentication"]
```

### Failure Mode

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `failure_mode` | string | `"graceful"` | How to handle observability failures |

**Values:**
- `"graceful"`: Log errors and continue normal operation
- `"strict"`: Fail the operation if observability fails (not recommended for production)

## Event Categories

Thunder events are grouped into categories. You can filter which categories are sent to each output.

| Category | Description | Example Events |
|----------|-------------|----------------|
| `observability.all` | All events (default) | All events |
| `observability.authentication` | Authentication and authorization | Login, token validation |
| `observability.flows` | Flow execution | Flow started, node execution |
| `observability.authorization` | Authorization decisions | Policy evaluation |

**Category Filtering Example:**
```yaml
# File: repository/conf/deployment.yaml
observability:
  enabled: true
  output:
    console:
      enabled: true
      categories: ["observability.authentication"]  # Only auth events
    
    opentelemetry:
      enabled: true
      categories: ["observability.flows"]          # Only flow events
    
    file:
      enabled: true
      categories: []                               # All events
```

## Common Configurations

### Development Environment
```yaml
# File: repository/conf/deployment.yaml
observability:
  enabled: true
  output:
    console:
      enabled: true
      format: json
      categories: ["observability.all"]
    
    opentelemetry:
      enabled: true
      exporter_type: "stdout"
      service_name: "thunder-dev"
      environment: "development"
      sample_rate: 1.0
  
  failure_mode: "graceful"
```

### Production with Jaeger
```yaml
# File: repository/conf/deployment.yaml
observability:
  enabled: true
  output:
    console:
      enabled: false
    
    opentelemetry:
      enabled: true
      exporter_type: "otlp"
      otlp_endpoint: "jaeger-collector.monitoring.svc:4317"
      service_name: "thunder-iam"
      service_version: "1.2.0"
      environment: "production"
      sample_rate: 0.05                    # 5% sampling for production load
      insecure: false
      categories: ["observability.all"]
    
    file:
      enabled: true
      file_path: "/var/log/thunder/analytics/observability.log"
      format: json
      categories: ["observability.authentication", "observability.authorization"]
  
  failure_mode: "graceful"
```

### Staging with Tempo
```yaml
# File: repository/conf/deployment.yaml
observability:
  enabled: true
  output:
    console:
      enabled: true
      format: text
      categories: ["observability.all"]
    
    opentelemetry:
      enabled: true
      exporter_type: "otlp"
      otlp_endpoint: "tempo.staging.internal:4317"
      service_name: "thunder-iam-staging"
      service_version: "1.2.0-rc1"
      environment: "staging"
      sample_rate: 0.5                     # 50% sampling
      insecure: true
      categories: ["observability.all"]
  
  failure_mode: "graceful"
```

### Disabled (Minimal Overhead)
```yaml
# File: repository/conf/deployment.yaml
observability:
  enabled: false
```

## Setting Up OpenTelemetry Backends

### Jaeger

#### Option 1: Docker Compose (Recommended)
```yaml
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # Jaeger UI
      - "4317:4317"    # OTLP gRPC
    environment:
      - COLLECTOR_OTLP_ENABLED=true
```

#### Option 2: Docker CLI
```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

Then configure Thunder:
```yaml
# File: repository/conf/deployment.yaml
observability:
  output:
    opentelemetry:
      enabled: true
      exporter_type: "otlp"
      otlp_endpoint: "localhost:4317"
```

Access Jaeger UI at: http://localhost:16686

### Grafana Tempo

#### Option 1: Docker Compose (Recommended)
```yaml
services:
  tempo:
    image: grafana/tempo:latest
    command: [ "-config.file=/etc/tempo.yaml" ]
    ports:
      - "4317:4317"    # OTLP gRPC
      - "3200:3200"    # Tempo HTTP
    volumes:
      - ./tempo.yaml:/etc/tempo.yaml
```

#### Option 2: Docker CLI
```bash
# Configuration file is required
docker run -d --name tempo \
  -v $(pwd)/tempo.yaml:/etc/tempo.yaml \
  -p 4317:4317 \
  -p 3200:3200 \
  grafana/tempo:latest \
  -config.file=/etc/tempo.yaml
```

## Troubleshooting

### No Traces Appearing
1. Check `observability.enabled` is `true`
2. Check `opentelemetry.enabled` is `true`
3. Verify the OTLP endpoint is reachable: `telnet localhost 4317`
4. Check Thunder logs for connection errors
5. Verify `sample_rate` is not `0.0`

### High Memory Usage
- Reduce `sample_rate` to a lower value (e.g., `0.1` for 10%)
- Reduce the number of categories being traced

### Events Not Linked in Trace
- Ensure all related events share the same TraceID
- Verify TraceID format is valid (UUID or 32-char hex)

## Performance Considerations

### Sampling
- **Development**: Use `sample_rate: 1.0` (100%)
- **Production**: Use `0.01` to `0.1` (1-10%) depending on load
- High sampling rates increase memory and network usage

### Categories
- Filter to only necessary categories in production
- Use `observability.all` in development only

### Failure Mode
- Always use `"graceful"` in production
- `"strict"` mode should only be used for testing

## Further Reading
- [Adding Observability Events - Developer Guide](../../community/contributing/development/observability.md)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Jaeger Getting Started](https://www.jaegertracing.io/docs/latest/getting-started/)
