# grow

A browser-based visualizer for Python's list amortization strategy. Watch how CPython's backing array resizes, copies elements, and maintains O(1) amortized cost — step by step.

<!-- TODO: Replace with an animated GIF of the app in action -->
<!-- ![grow demo](docs/demo.gif) -->

## What it shows

- **Array cells** — filled slots (green), empty capacity (grey), mid-resize copies (amber)
- **Resize formula** — CPython's `(n + (n >> 3) + 6) & ~3` broken down with tooltips
- **Cost graph** — per-append cost with amortized average line
- **Strategy comparison** — CPython vs. Doubling vs. Java 1.5x vs. Fixed (no growth)

Three modes: **Presets** (pick an append count), **Wrapper** (write Python using `VisualList` directly), and **Auto** (write normal Python — list ops are rewritten automatically).

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
  main.go                     # entry point — starts HTTP server
  internal/
    simulator/                # growth strategies + VisualList engine
    events/                   # 14 event types with JSON marshalling
    executor/                 # gpython sandbox, AST rewriter, builtins
    server/                   # HTTP + WebSocket handler, embedded frontend
  frontend/                   # React + TypeScript + Vite + Tailwind
    src/
      store.ts                # Zustand state — events, playback, costs
      components/             # Rail, CanvasArray, CostGraph, panels
      hooks/                  # useEventStream, usePlayback, useTheme, ...
  testdata/                   # CPython 3.14.2 reference fixture
  scripts/                    # verify_cpython.py, evals
```

## How it works

1. The frontend sends an append count (or Python source) over WebSocket
2. The Go backend runs the operations through `simulator.VisualList` (or gpython for user code)
3. Events (`append_begin`, `resize_begin`, `copy_element`, `resize_end`, `append_end`, ...) stream back
4. The frontend replays events step-by-step with animations, building the cost graph as it goes

The CPython growth formula: `new_cap = (needed + (needed >> 3) + 6) & ~3` — grows by ~12.5% plus padding, aligned to a multiple of 4.

## Features

- Play/pause/step/scrub through events
- Adjustable playback speed (0.5x–5x)
- Dark/light theme toggle
- Keyboard shortcuts (Space, arrows, `?` for help)
- Formula tooltips explaining each bitwise operation
- Monaco code editor for wrapper and auto modes
- Resizable bottom dock with code/graph tabs

## Tech stack

- **Backend:** Go, [gpython](https://github.com/nicholasgasior/gpython) for Python execution
- **Frontend:** React 19, TypeScript, Vite, Tailwind v4, Zustand, Motion, Recharts, Monaco Editor
- **CI:** GitHub Actions (Go vet + race tests, frontend typecheck + build)
