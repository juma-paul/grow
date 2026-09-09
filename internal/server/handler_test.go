package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func readAllEvents(t *testing.T, ws *websocket.Conn) []map[string]any {
	t.Helper()
	var events []map[string]any
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		var evt map[string]any
		if err := json.Unmarshal(msg, &evt); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		events = append(events, evt)
	}
	return events
}

func dialExecute(t *testing.T, server *httptest.Server, query string) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/execute?" + query
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return ws
}

func TestExecuteDefaultStrategy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandleExecute))
	defer srv.Close()

	ws := dialExecute(t, srv, "count=5")
	defer ws.Close()

	events := readAllEvents(t, ws)
	if len(events) == 0 {
		t.Fatal("no events")
	}

	resizes := 0
	for _, e := range events {
		if e["type"] == "resize_begin" {
			resizes++
		}
	}
	if resizes != 2 {
		t.Errorf("cpython default: resize count = %d, want 2", resizes)
	}
}

func TestExecuteDoublingStrategy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandleExecute))
	defer srv.Close()

	ws := dialExecute(t, srv, "count=5&strategy=doubling")
	defer ws.Close()

	events := readAllEvents(t, ws)
	resizes := 0
	for _, e := range events {
		if e["type"] == "resize_begin" {
			resizes++
			if e["new_cap"] == float64(4) || e["new_cap"] == float64(8) {
				continue
			}
			t.Errorf("unexpected doubling resize to %v", e["new_cap"])
		}
	}
	if resizes != 2 {
		t.Errorf("doubling: resize count = %d, want 2", resizes)
	}
}

func TestExecuteNoGrowthOverflow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandleExecute))
	defer srv.Close()

	ws := dialExecute(t, srv, "count=10&strategy=nogrowth")
	defer ws.Close()

	events := readAllEvents(t, ws)
	overflows := 0
	for _, e := range events {
		if e["type"] == "overflow" {
			overflows++
			if e["capacity"] != float64(4) {
				t.Errorf("overflow capacity = %v, want 4", e["capacity"])
			}
		}
	}
	if overflows != 1 {
		t.Errorf("nogrowth: overflow count = %d, want 1", overflows)
	}
}

func TestExecuteUnknownStrategy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandleExecute))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/execute?strategy=bogus"
	_, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err == nil {
		t.Fatal("expected connection to fail for unknown strategy")
	}
	if resp != nil && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
