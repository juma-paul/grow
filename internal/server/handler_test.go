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

func countType(events []map[string]any, typ string) int {
	n := 0
	for _, e := range events {
		if e["type"] == typ {
			n++
		}
	}
	return n
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
	if resizes := countType(events, "resize_begin"); resizes != 2 {
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

func TestExecuteInsertSeedsThenShifts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandleExecute))
	defer srv.Close()

	// Seed 8 elements, then insert 3 times at the start. Each insert into a
	// non-empty list shifts the existing elements right.
	ws := dialExecute(t, srv, "op=insert&seed=8&count=3&pos=start")
	defer ws.Close()

	events := readAllEvents(t, ws)
	if inits := countType(events, "init"); inits != 1 {
		t.Errorf("init count = %d, want 1", inits)
	}
	if ends := countType(events, "insert_end"); ends != 3 {
		t.Errorf("insert_end count = %d, want 3", ends)
	}
	if shifts := countType(events, "shift_right"); shifts == 0 {
		t.Error("expected shift_right events from inserting at the start")
	}
}

func TestExecuteExtendBulkResize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandleExecute))
	defer srv.Close()

	ws := dialExecute(t, srv, "op=extend&seed=5&count=10")
	defer ws.Close()

	events := readAllEvents(t, ws)
	if ends := countType(events, "extend_end"); ends != 1 {
		t.Errorf("extend_end count = %d, want 1", ends)
	}
	if begins := countType(events, "extend_begin"); begins != 1 {
		t.Errorf("extend_begin count = %d, want 1", begins)
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
