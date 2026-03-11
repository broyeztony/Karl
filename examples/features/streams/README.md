# Karl Stream Examples

These examples focus on stream pipelines (`|`), with Kubernetes-heavy scenarios and a few local stream demos.

## Requirements

- `kubectl` installed
- current context points to a reachable cluster (minikube, kind, remote, ...)

## Files

- `kubernetes_pod_inventory_streams.k`
  - Build a pod inventory with `readable line -> object` mapping.
  - Uses `lines()`, `filter(...)`, `map(...)`, `collect()`.

- `kubernetes_logs_error_channel.k`
  - Follow deployment logs, keep only error lines, bridge stream output to a channel.
  - Uses `take(...)`, `send(ch)`, and a concurrent consumer task.

- `kubernetes_merge_cluster_feed.k`
  - Merge pod and node feeds into one stream for unified inspection.
  - Uses `merge(...)`, `map(...)`, `take(...)`, `chunk(...)`.

- `kubernetes_phase_counts.k`
  - Aggregate pod phases directly by key.
  - Uses `group_count()`.

- `kubernetes_pod_node_join.k`
  - Join pod and node streams by node name with keyed fan-in.
  - Uses `join(...)` + `map(...)` + `take(...)`.

- `kubernetes_namespaces_debounced_channel.k`
  - Debounce bursty namespace stream values and forward through a channel sink.
  - Uses `debounce(ms)` + `toChannel(ch)` + channel recv loop.

- `kubernetes_pod_phase_partition.k`
  - Split pod stream into running vs non-running groups.
  - Uses `split(pred)` sink on mapped pod objects.

- `kubernetes_pods_tee_spill.k`
  - Duplicate a pod stream to console and disk while preserving downstream flow.
  - Uses `tee(stdout())` + `spill(path)` + implicit lambda shorthand.

- `new_stream_builtins.k`
  - Demonstrates the newly added stream builtins on local NDJSON data.
  - Uses `json()`, `distinct()`, `sort()`, `group_count()`, `reduce_by_key()`, `top()`, `split()`, `exec(...)`.

- `stdin_numbers_top.k`
  - Stream integers from stdin and keep top values.
  - Uses `stdin()`, `lines()`, `map(parseInt(_))`, `top(n)`.

- `http_stream_json_lines.k`
  - Serve NDJSON locally, stream it with `http(url)`, decode, aggregate.
  - Uses `httpServe(...)`, `http(url)`, `json()`, `group_count(...)`.

- `exec_sink_wc.k`
  - Pipe stream content into a subprocess and inspect run status.
  - Uses `exec(...)` sink.

## Run

```bash
karl run examples/features/streams/kubernetes_pod_inventory_streams.k
karl run examples/features/streams/kubernetes_logs_error_channel.k
karl run examples/features/streams/kubernetes_merge_cluster_feed.k
karl run examples/features/streams/kubernetes_phase_counts.k
karl run examples/features/streams/kubernetes_pod_node_join.k
karl run examples/features/streams/kubernetes_namespaces_debounced_channel.k
karl run examples/features/streams/kubernetes_pod_phase_partition.k
karl run examples/features/streams/kubernetes_pods_tee_spill.k
karl run examples/features/streams/new_stream_builtins.k
karl run examples/features/streams/stdin_numbers_top.k < numbers.txt
karl run examples/features/streams/http_stream_json_lines.k
karl run examples/features/streams/exec_sink_wc.k
```
