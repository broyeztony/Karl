# Karl Process API Examples

These examples show how `cmd`/`proc`/`run` + channels turn Karl into a practical systems language.

## Files

- `kubernetes_health_report.k`: real `kubectl` wrapper with concurrent namespace checks.
- `text_analytics_pipeline.k`: Unix pipeline composed in Karl with `cmd(...) | cmd(...)`.
- `stream_and_abort.k`: long-running process control (`pid`, `running`, `abort`, `wait`).
- `streaming_pipeline_channels.k`: live pipeline streaming with channel-driven event fan-in.

## Run

```bash
karl run examples/features/processes/text_analytics_pipeline.k
karl run examples/features/processes/stream_and_abort.k
karl run examples/features/processes/streaming_pipeline_channels.k

# requires kubectl + access to a cluster
karl run examples/features/processes/kubernetes_health_report.k
# optional rollout check: <namespace> <deployment>
karl run examples/features/processes/kubernetes_health_report.k -- default api
```

## Why this matters

With this API, Karl can now build:
- operator-style CLIs that orchestrate system tools,
- streaming ETL and log pipelines,
- process supervisors with graceful shutdown and timeouts,
- concurrent command fan-out (for clusters, fleets, CI workers).
