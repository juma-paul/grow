package server

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/juma-paul/grow/internal/events"
	"github.com/juma-paul/grow/internal/executor"
	"github.com/juma-paul/grow/internal/simulator"
)

const autoTimeout = 5 * time.Second

// HandleAutoExecute upgrades to WebSocket, reads user Python source,
// rewrites it so list ops go through VisualList, executes in a sandbox,
// and streams the collected events back.
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

	evts, runErr := executor.RunAuto(string(msg), autoTimeout)

	for _, e := range evts {
		data, err := events.Marshal(e)
		if err != nil {
			log.Printf("marshal error: %v", err)
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			break
		}
	}

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
// append scenario. Accepts ?count=N&strategy=NAME query params.
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

	var writeErr error
	lst := simulator.NewVisualList(strategy, func(e events.Event) {
		if writeErr != nil {
			return
		}
		data, err := events.Marshal(e)
		if err != nil {
			log.Printf("marshal error: %v", err)
			return
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			writeErr = err
		}
	})

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(simulator.OverflowExceeded); !ok {
				panic(r)
			}
		}
		conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "done"),
		)
		conn.SetReadDeadline(time.Now().Add(time.Second))
		conn.ReadMessage()
	}()

	for i := 0; i < count; i++ {
		lst.Append(i)
		if writeErr != nil {
			break
		}
	}
}
