package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/juma-paul/grow/internal/events"
	"github.com/juma-paul/grow/internal/simulator"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleExecute upgrades to WebSocket and streams events for a
// hardcoded "append 100" scenario.
func HandleExecute(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	lst := simulator.NewVisualList(simulator.CPythonGrowth{}, func(e events.Event) {
		data, err := events.Marshal(e)
		if err != nil {
			log.Printf("marshal error: %v", err)
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("write error: %v", err)
		}
	})

	for i := 0; i < 100; i++ {
		lst.Append(i)
	}

	// Signal clean shutdown to the client
	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "done"),
	)

	// Complete the close handshake by reading the client's close frame.
	conn.SetReadDeadline(time.Now().Add(time.Second))
	conn.ReadMessage()
}
