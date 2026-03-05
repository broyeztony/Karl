# Karl Process API Examples

These examples show how `cmd`/`proc`/`run` + streams make Karl a practical systems language.

## Files

- `kubernetes_health_report.k`: real `kubectl` wrapper with concurrent namespace checks.
- `text_analytics_pipeline.k`: Unix pipeline composed in Karl with `cmd(...) | cmd(...)`.
- `stream_and_abort.k`: long-running process control (`pid`, `running`, `abort`, `wait`).
- `streaming_pipeline_channels.k`: live pipeline streaming with stream reads + channel event fan-in.
- `binary_stream_copy.k`: binary file streaming with `reader`/`writer`/`copy` in `BYTES` mode.
- `binary_process_passthrough.k`: binary bytes passthrough through `proc` using `p.stdin`/`p.stdout`.
- `binary_chunk_loop_reconstruct.k`: manual `read()` chunk loop in `BYTES` mode + `bytesJoin(...)` reconstruction (no `copy`).

## Run

```bash
karl run examples/features/processes/text_analytics_pipeline.k
karl run examples/features/processes/stream_and_abort.k
karl run examples/features/processes/streaming_pipeline_channels.k
karl run examples/features/processes/binary_stream_copy.k
karl run examples/features/processes/binary_process_passthrough.k
karl run examples/features/processes/binary_chunk_loop_reconstruct.k

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
