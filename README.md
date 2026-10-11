# grow

A browser-based visualizer for Python's list amortization strategy. Watch how CPython's backing array resizes, copies elements, and maintains O(1) amortized cost — step by step.

## What it shows

- **Array cells** — filled slots (green), empty capacity (grey), mid-resize copies (amber)
- **Resize formula** — CPython's `(n + (n >> 3) + 6) & ~3`
- **Cost graph** — per-operation cost with amortized average line
- **Strategy comparison** — CPython vs. Doubling vs. 1.5× vs. Fixed (no growth)

Pick a preset operation — **append** N items, **insert** at start/middle/end, or **extend** in bulk — and watch the backing array react.

## Quick start

```bash
go run .
# open http://localhost:8080
```

Requires Go 1.26+ and Node 22+ (for building the frontend).

### Build from source

```bash
cd frontend && pnpm install && pnpm build && cd ..
go build -o grow .
./grow
```

## Project structure

```
grow/
  main.go                     # entry point — serves /execute + embedded frontend
  internal/
    simulator/                # growth strategies + VisualList engine
    events/                   # event types with JSON marshalling
    server/                   # HTTP + WebSocket handler, bounded event stream
  frontend/                   # React + TypeScript + Vite + Tailwind
    src/
      store.ts                # Zustand state — events, playback, costs
      components/             # Rail, CanvasArray, CostGraph, PresetsPanel
      hooks/                  # useEventStream, usePlayback
  testdata/                   # CPython 3.14.2 reference fixture
  scripts/                    # verify_cpython.py (reference generator)
```

## How it works

1. The frontend opens a WebSocket to `/execute` with a scenario (`op`, `count`, `pos`, `seed`, `strategy`)
2. The Go backend runs that scenario through `simulator.VisualList`, emitting granular events
3. Events (`append_begin`, `resize_begin`, `copy_element`, `resize_end`, `append_end`, `shift_right`, ...) stream back over a bounded channel
4. The frontend replays events step-by-step with animations, building the cost graph as it goes

The CPython growth formula: `new_cap = (needed + (needed >> 3) + 6) & ~3` — grows by ~12.5% plus padding, aligned to a multiple of 4. The engine models it exactly and is validated against a 1000-entry fixture captured from real CPython 3.14.2.

## Features

- Append / insert / extend scenarios across four growth strategies
- Play / pause / step / scrub through events
- Adjustable playback speed (0.5×–5×)
- Per-operation cost graph with amortized average

## Tech stack

- **Backend:** Go, [gorilla/websocket](https://github.com/gorilla/websocket) for streaming; property tests with [rapid](https://pgregory.net/rapid)
- **Frontend:** React 19, TypeScript, Vite, Tailwind v4, Zustand, Motion, Recharts
