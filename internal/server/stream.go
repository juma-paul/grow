package server

import (
	"github.com/gorilla/websocket"
	"github.com/juma-paul/grow/internal/events"
)

const channelCap = 8192

// EventStream is a bounded channel between the simulator (producer) and the
// WebSocket writer (consumer). The fixed-capacity channel provides natural
// backpressure: if the socket is slow to drain, the producer blocks on Send
// instead of buffering without limit.
type EventStream struct {
	ch       chan events.Event
	writeErr error
}

// NewEventStream returns a stream with a fixed-capacity channel.
func NewEventStream() *EventStream {
	return &EventStream{ch: make(chan events.Event, channelCap)}
}

// Send enqueues an event, blocking when the channel is full. Call Close when
// the producer is done.
func (s *EventStream) Send(e events.Event) { s.ch <- e }

// Close signals that no more events will be sent.
func (s *EventStream) Close() { close(s.ch) }

// Err returns the first WebSocket write error, if any.
func (s *EventStream) Err() error { return s.writeErr }

// WriteTo drains the channel to the WebSocket connection. Run it in a
// goroutine; it returns when the channel is closed.
func (s *EventStream) WriteTo(conn *websocket.Conn) {
	for e := range s.ch {
		if s.writeErr != nil {
			continue
		}
		data, err := events.Marshal(e)
		if err != nil {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			s.writeErr = err
		}
	}
}
