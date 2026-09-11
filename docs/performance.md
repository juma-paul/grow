# Performance characteristics

## Architecture

grow is a single-process Go server. Each WebSocket connection runs its own
goroutine pair: one for simulation/execution, one for writing events to the
client. The simulator itself is CPU-bound (pure Go), while the observe mode
spawns a real `python3` subprocess.

## Where the time goes

| Handler   | Bottleneck           | Typical latency (20 elements) |
|-----------|----------------------|-------------------------------|
| /execute  | Go simulator (CPU)   | < 1 ms                        |
| /auto     | gpython interpreter  | 5–50 ms                       |
| /observe  | python3 subprocess   | 100–500 ms                    |

The `/execute` handler is goroutine-only — it scales linearly with cores until
the goroutine scheduler saturates. The `/auto` handler holds the GIL-equivalent
inside gpython for the duration of execution. The `/observe` handler forks a
real CPython process per request, making it the most expensive path.

## Backpressure

Each connection gets a bounded event channel (capacity 8,192). If the channel
fills, `Send()` drops the event. For high-volume scenarios (> 10,000 events),
`CoalesceEvents` collapses runs of non-resize appends into `AppendBatch`
events, preserving all resize boundaries.

## Why not scale out?

grow is a teaching tool, not a production service. The reasons to stay
single-process:

1. **No shared state in the hot path.** Each simulation is independent — there
   is nothing to coordinate between instances. Adding a load balancer and
   multiple replicas would add complexity with no correctness benefit.

2. **WebSocket affinity is free.** A single process means no sticky-session
   routing. A load-balanced setup would need session affinity or a shared
   event bus, both adding latency.

3. **The real bottleneck is the client.** At high element counts, the browser's
   rendering (DOM updates, canvas redraws, chart re-renders) becomes the limit
   long before the server. Scaling the server doesn't help.

4. **Observe mode is inherently per-machine.** It requires a local CPython
   installation with ctypes access. Distributing this across containers would
   require each container to have a matching Python runtime.

## Monitoring

The server exposes Prometheus metrics at `/metrics`:

- `grow_active_connections` — gauge of open WebSocket connections
- `grow_executions_in_flight` — gauge of running simulations
- `grow_execution_duration_seconds` — histogram by handler
- `grow_events_emitted_total` — counter by handler
- `grow_events_coalesced_total` — counter of coalesced events

A Grafana dashboard is provided at `monitoring/grafana-dashboard.json`.

## Running benchmarks

```bash
# Start the server
go run .

# Quick load test (10 concurrent, 100 total)
go run scripts/loadtest.go -conc 10 -n 100

# Concurrency sweep with CPU tracking
go run scripts/benchmark.go -addr localhost:8080 | tee results.csv

# Generate latency chart
python3 scripts/chart_latency.py results.csv -o docs/latency.svg
```
