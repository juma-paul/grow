package server

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/juma-paul/grow/internal/events"
	"github.com/juma-paul/grow/internal/executor"
	"github.com/juma-paul/grow/internal/simulator"
)

const autoTimeout = 5 * time.Second

// HandleAutoExecute upgrades to WebSocket, reads user Python source,
// rewrites it so list ops go through VisualList, executes in a sandbox,
// coalesces if needed, and streams the events back via EventStream.
func HandleAutoExecute(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "handler", "auto", "error", err)
		return
	}
	defer conn.Close()
	ActiveConnections.Inc()
	defer ActiveConnections.Dec()

	rewrite := r.URL.Query().Get("rewrite") != "false"
	slog.Info("connection opened", "handler", "auto", "rewrite", rewrite)

	_, msg, err := conn.ReadMessage()
	if err != nil {
		slog.Error("read source failed", "handler", "auto", "error", err)
		return
	}

	ExecutionsInFlight.Inc()
	start := time.Now()
	var evts []events.Event
	var runErr error
	if !rewrite {
		evts, runErr = executor.RunWrapper(string(msg), autoTimeout)
	} else {
		evts, runErr = executor.RunAuto(string(msg), autoTimeout)
	}
	duration := time.Since(start)
	ExecutionsInFlight.Dec()
	ExecutionDuration.WithLabelValues("auto").Observe(duration.Seconds())

	evts, coalesced := CoalesceEvents(evts)
	EventsEmitted.WithLabelValues("auto").Add(float64(len(evts)))
	if coalesced > 0 {
		EventsCoalesced.Add(float64(coalesced))
	}

	slog.Info("execution complete",
		"handler", "auto",
		"events", len(evts),
		"coalesced", coalesced,
		"duration", duration,
		"error", runErr,
	)

	stream := NewEventStream()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		stream.WriteTo(conn)
	}()

	for _, e := range evts {
		stream.Send(e)
	}
	stream.Close()
	wg.Wait()

	closeMsg := "done"
	if runErr != nil {
		closeMsg = runErr.Error()
	}
	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, closeMsg),
	)

	conn.SetReadDeadline(time.Now().Add(time.Second))
	conn.ReadMessage()
}

const observeTimeout = 10 * time.Second

// HandleObserve upgrades to WebSocket, reads user Python source,
// instruments it with _snap() calls, runs it under real CPython,
// diffs the snapshots into events, coalesces if needed, and streams
// them back via EventStream.
func HandleObserve(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "handler", "observe", "error", err)
		return
	}
	defer conn.Close()
	ActiveConnections.Inc()
	defer ActiveConnections.Dec()

	slog.Info("connection opened", "handler", "observe")

	_, msg, err := conn.ReadMessage()
	if err != nil {
		slog.Error("read source failed", "handler", "observe", "error", err)
		return
	}

	ExecutionsInFlight.Inc()
	start := time.Now()
	script := executor.InstrumentForObserve(string(msg))
	snaps, runErr := executor.RunCPython(script, observeTimeout)
	duration := time.Since(start)
	ExecutionsInFlight.Dec()
	ExecutionDuration.WithLabelValues("observe").Observe(duration.Seconds())

	evts := executor.DiffSnapshots(snaps)
	evts, coalesced := CoalesceEvents(evts)
	EventsEmitted.WithLabelValues("observe").Add(float64(len(evts)))
	if coalesced > 0 {
		EventsCoalesced.Add(float64(coalesced))
	}

	slog.Info("execution complete",
		"handler", "observe",
		"snapshots", len(snaps),
		"events", len(evts),
		"coalesced", coalesced,
		"duration", duration,
		"error", runErr,
	)

	stream := NewEventStream()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		stream.WriteTo(conn)
	}()

	for _, e := range evts {
		stream.Send(e)
	}
	stream.Close()
	wg.Wait()

	closeMsg := "done"
	if runErr != nil {
		closeMsg = runErr.Error()
	}
	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, closeMsg),
	)

	conn.SetReadDeadline(time.Now().Add(time.Second))
	conn.ReadMessage()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleExecute upgrades to WebSocket and streams events for an
// append scenario via EventStream. The simulator runs in a goroutine
// and sends events to a bounded channel; a writer goroutine drains
// the channel to the WebSocket.
func HandleExecute(w http.ResponseWriter, r *http.Request) {
	count := 20
	if q := r.URL.Query().Get("count"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 1000 {
			count = n
		}
	}

	strategyName := r.URL.Query().Get("strategy")
	if strategyName == "" {
		strategyName = "cpython"
	}
	strategy, err := simulator.StrategyByName(strategyName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, upgradeErr := upgrader.Upgrade(w, r, nil)
	if upgradeErr != nil {
		slog.Error("websocket upgrade failed", "handler", "execute", "error", upgradeErr)
		return
	}
	defer conn.Close()
	ActiveConnections.Inc()
	defer ActiveConnections.Dec()

	slog.Info("connection opened",
		"handler", "execute",
		"count", count,
		"strategy", strategyName,
	)

	ExecutionsInFlight.Inc()
	start := time.Now()
	stream := NewEventStream()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		stream.WriteTo(conn)
	}()

	lst := simulator.NewVisualList(strategy, func(e events.Event) {
		if stream.Err() != nil {
			return
		}
		stream.Send(e)
	})

	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(simulator.OverflowExceeded); !ok {
					panic(r)
				}
			}
		}()
		for i := 0; i < count; i++ {
			lst.Append(i)
			if stream.Err() != nil {
				break
			}
		}
	}()

	stream.Close()
	wg.Wait()

	duration := time.Since(start)
	ExecutionsInFlight.Dec()
	ExecutionDuration.WithLabelValues("execute").Observe(duration.Seconds())

	slog.Info("execution complete",
		"handler", "execute",
		"count", count,
		"strategy", strategyName,
		"duration", duration,
	)

	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "done"),
	)
	conn.SetReadDeadline(time.Now().Add(time.Second))
	conn.ReadMessage()
}
