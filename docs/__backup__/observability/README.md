# Observability Guide

Welcome to the Thunder Observability Guide. This section allows you to understand, configure, and monitor the Thunder Identity Server.

## Documentation

*   **[Architecture](architecture.md)**
    *   Learn how the observability system works under the hood.
    *   Understand the high-level data flow and component interaction.
    *   See the architectural diagram.

*   **[Configuration](configuration.md)**
    *   Learn how to enable and configure observability features.
    *   Set up OpenTelemetry exporters (Jaeger, Tempo).
    *   Configure logging output (Console, File).

*   **[Event Reference](events.md)**
    *   View the complete list of events published by the system.
    *   Understand the event schema and data contract.
    *   See what data is available for each event type.

*   **[Analytics Dashboard](analytics.md)**
    *   Set up OpenSearch and Data Prepper for visual analytics.
    *   Deploy the full observability stack using Docker Compose.
    *   Access the pre-defined Trace Analytics dashboards.

## Quick Start

To quickly enable observability in your local development environment, add the following to your `deployment.yaml`:

```yaml
observability:
  enabled: true
  output:
    console:
      enabled: true
      format: json
      categories: ["observability.all"]
```

For more advanced setups, please refer to the [Configuration Guide](configuration.md).
