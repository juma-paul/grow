package server

import (
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/juma-paul/grow/internal/events"
)

const (
	channelCap    = 8192
	coalesceAfter = 10000
)

// EventStream is a bounded channel between an event producer (simulator,
// executor, or observe pipeline) and the WebSocket writer. When the total
// event count exceeds coalesceAfter, consecutive non-resize appends are
// collapsed into AppendBatch events to keep the stream bounded.
type EventStream struct {
	ch        chan events.Event
	writeErr  error
	coalesced atomic.Int64
}

// NewEventStream returns a stream with a fixed-capacity channel.
func NewEventStream() *EventStream {
	return &EventStream{
		ch: make(chan events.Event, channelCap),
	}
}

// Send enqueues an event. It blocks if the channel is full (the writer
// goroutine is the drain). Call Close when the producer is done.
func (s *EventStream) Send(e events.Event) {
	s.ch <- e
}

// Close signals that no more events will be sent.
func (s *EventStream) Close() {
	close(s.ch)
}

// Err returns the first WebSocket write error, if any.
func (s *EventStream) Err() error {
	return s.writeErr
}

// Coalesced returns the number of events that were collapsed into batches.
func (s *EventStream) Coalesced() int64 {
	return s.coalesced.Load()
}

// WriteTo drains the channel to the WebSocket connection.
// Run this in a goroutine; it returns when the channel is closed.
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

// CoalesceEvents reduces a large event slice by collapsing runs of
// non-resize append pairs into AppendBatch summaries. Resize, shrink,
// copy, and other structural events are always preserved.
//
// If len(evts) <= coalesceAfter, returns evts unchanged.
func CoalesceEvents(evts []events.Event) ([]events.Event, int64) {
	if len(evts) <= coalesceAfter {
		return evts, 0
	}

	var out []events.Event
	var coalesced int64
	i := 0
	for i < len(evts) {
		e := evts[i]
		if ab, ok := e.(events.AppendBegin); ok && !nextIsResize(evts, i) {
			runStart := ab.Length
			count := 0
			for i < len(evts) {
				if _, ok := evts[i].(events.AppendBegin); ok && !nextIsResize(evts, i) {
					count++
					i++ // skip append_begin
					if i < len(evts) {
						i++ // skip append_end
					}
				} else {
					break
				}
			}
			if count == 1 {
				out = append(out,
					events.AppendBegin{Value: nil, Length: runStart, Capacity: 0},
					events.AppendEnd{Cost: 1},
				)
			} else {
				coalesced += int64(count*2 - 1)
				out = append(out, events.AppendBatch{
					Count:   count,
					FromLen: runStart,
					ToLen:   runStart + count,
				})
			}
			continue
		}
		out = append(out, e)
		i++
	}
	return out, coalesced
}

func nextIsResize(evts []events.Event, i int) bool {
	for j := i + 1; j < len(evts) && j <= i+2; j++ {
		switch evts[j].(type) {
		case events.ResizeBegin, events.ShrinkBegin:
			return true
		}
	}
	return false
}
