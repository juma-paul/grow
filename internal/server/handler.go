package server

import (
	"log"
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
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Printf("read source failed: %v", err)
		return
	}

	var evts []events.Event
	var runErr error
	if r.URL.Query().Get("rewrite") == "false" {
		evts, runErr = executor.RunWrapper(string(msg), autoTimeout)
	} else {
		evts, runErr = executor.RunAuto(string(msg), autoTimeout)
	}

	evts, coalesced := CoalesceEvents(evts)
	if coalesced > 0 {
		log.Printf("coalesced %d events", coalesced)
	}

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
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Printf("read source failed: %v", err)
		return
	}

	script := executor.InstrumentForObserve(string(msg))
	snaps, runErr := executor.RunCPython(script, observeTimeout)

	evts := executor.DiffSnapshots(snaps)
	evts, coalesced := CoalesceEvents(evts)
	if coalesced > 0 {
		log.Printf("coalesced %d events", coalesced)
	}

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
		log.Printf("websocket upgrade failed: %v", upgradeErr)
		return
	}
	defer conn.Close()

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

	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "done"),
	)
	conn.SetReadDeadline(time.Now().Add(time.Second))
	conn.ReadMessage()
}
