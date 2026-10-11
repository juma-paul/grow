package server

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/juma-paul/grow/internal/events"
	"github.com/juma-paul/grow/internal/simulator"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleExecute upgrades to a WebSocket and streams the events produced by
// running a scenario against the simulator. The scenario is selected by
// query params:
//
//	op=append|insert|extend  operation to run (default append)
//	count=N                  number of operations / items
//	pos=start|middle|end     insert position (insert only)
//	seed=N                   pre-fill the list with N elements first
//	strategy=cpython|doubling|1.5x|nogrowth
//
// The simulator runs on the request goroutine and emits events into a
// bounded channel; a writer goroutine drains the channel to the socket.
func HandleExecute(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	strategyName := q.Get("strategy")
	if strategyName == "" {
		strategyName = "cpython"
	}
	strategy, err := simulator.StrategyByName(strategyName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	op := q.Get("op")
	if op == "" {
		op = "append"
	}
	count := intParam(q.Get("count"), 20, 1, 1000)
	seed := intParam(q.Get("seed"), 0, 0, 1000)
	pos := q.Get("pos")

	conn, upgradeErr := upgrader.Upgrade(w, r, nil)
	if upgradeErr != nil {
		slog.Error("websocket upgrade failed", "error", upgradeErr)
		return
	}
	defer conn.Close()

	slog.Info("connection opened",
		"op", op, "count", count, "seed", seed, "strategy", strategyName)

	stream := NewEventStream()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		stream.WriteTo(conn)
	}()

	lst := simulator.NewVisualListWithLimits(strategy, func(e events.Event) {
		if stream.Err() != nil {
			return
		}
		stream.Send(e)
	}, simulator.Limits{MaxOps: 100_000, MaxAlloc: 10_000_000})

	runScenario(lst, op, count, seed, pos, stream)

	stream.Close()
	wg.Wait()

	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "done"),
	)
	conn.SetReadDeadline(time.Now().Add(time.Second))
	conn.ReadMessage()
}

// runScenario drives the simulator for the chosen operation. A NoGrowth
// overflow or a sandbox limit aborts the run via panic; both are expected
// terminal conditions, so they are recovered here (the triggering event is
// already emitted before the panic).
func runScenario(lst *simulator.VisualList, op string, count, seed int, pos string, stream *EventStream) {
	defer func() {
		if rec := recover(); rec != nil {
			if _, ok := rec.(simulator.OverflowExceeded); ok {
				return
			}
			if _, ok := rec.(simulator.LimitExceeded); ok {
				return
			}
			panic(rec)
		}
	}()

	if seed > 0 {
		lst.Seed(seed)
	}

	switch op {
	case "insert":
		for i := 0; i < count; i++ {
			var idx int
			switch pos {
			case "start":
				idx = 0
			case "middle":
				idx = lst.Len() / 2
			default:
				idx = lst.Len()
			}
			lst.Insert(idx, 90+i)
			if stream.Err() != nil {
				return
			}
		}
	case "extend":
		items := make([]any, count)
		for i := range items {
			items[i] = seed + i
		}
		lst.Extend(items)
	default: // append
		for i := 0; i < count; i++ {
			lst.Append(i)
			if stream.Err() != nil {
				return
			}
		}
	}
}

// intParam parses a query value, returning def when empty, malformed, or
// outside [min, max].
func intParam(s string, def, min, max int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < min || n > max {
		return def
	}
	return n
}
