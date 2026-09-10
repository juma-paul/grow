package executor

import (
	"testing"
	"time"

	"github.com/juma-paul/grow/internal/events"
)

func TestDiffAppendNoResize(t *testing.T) {
	snaps := []Snapshot{
		{Tag: "init", Len: 0, Cap: 4, Addr: 1000},
		{Tag: "append", Len: 1, Cap: 4, Addr: 1000},
	}
	evts := DiffSnapshots(snaps)

	assertEventTypes(t, evts, []string{"append_begin", "append_end"})
}

func TestDiffAppendWithResize(t *testing.T) {
	snaps := []Snapshot{
		{Tag: "before", Len: 4, Cap: 4, Addr: 1000},
		{Tag: "after", Len: 5, Cap: 8, Addr: 2000},
	}
	evts := DiffSnapshots(snaps)

	// append_begin, resize_begin, 4x copy_element, resize_end, append_end
	expected := []string{
		"append_begin", "resize_begin",
		"copy_element", "copy_element", "copy_element", "copy_element",
		"resize_end", "append_end",
	}
	assertEventTypes(t, evts, expected)

	// Check resize fields
	rb := evts[1].(events.ResizeBegin)
	if rb.OldCap != 4 || rb.NewCap != 8 {
		t.Errorf("resize: old=%d new=%d", rb.OldCap, rb.NewCap)
	}
}

func TestDiffAppendInPlaceResize(t *testing.T) {
	snaps := []Snapshot{
		{Tag: "before", Len: 4, Cap: 4, Addr: 1000},
		{Tag: "after", Len: 5, Cap: 8, Addr: 1000}, // same addr = in-place
	}
	evts := DiffSnapshots(snaps)

	// No copy_element events when addr unchanged
	expected := []string{
		"append_begin", "resize_begin", "resize_end", "append_end",
	}
	assertEventTypes(t, evts, expected)
}

func TestDiffPop(t *testing.T) {
	snaps := []Snapshot{
		{Tag: "before", Len: 5, Cap: 8, Addr: 1000},
		{Tag: "after", Len: 4, Cap: 8, Addr: 1000},
	}
	evts := DiffSnapshots(snaps)

	assertEventTypes(t, evts, []string{"pop_begin", "pop_end"})
}

func TestDiffPopWithShrink(t *testing.T) {
	snaps := []Snapshot{
		{Tag: "before", Len: 3, Cap: 8, Addr: 1000},
		{Tag: "after", Len: 2, Cap: 4, Addr: 2000},
	}
	evts := DiffSnapshots(snaps)

	expected := []string{
		"pop_begin", "shrink_begin",
		"copy_element", "copy_element",
		"shrink_end", "pop_end",
	}
	assertEventTypes(t, evts, expected)
}

func TestDiffMultipleAppends(t *testing.T) {
	snaps := []Snapshot{
		{Tag: "init", Len: 0, Cap: 0, Addr: 0},
		{Tag: "append", Len: 1, Cap: 4, Addr: 1000},
		{Tag: "append", Len: 2, Cap: 4, Addr: 1000},
		{Tag: "append", Len: 3, Cap: 4, Addr: 1000},
	}
	evts := DiffSnapshots(snaps)

	// First: append + resize (0→4), rest: append only
	types := eventTypes(evts)
	if types[0] != "append_begin" {
		t.Errorf("first event: %s", types[0])
	}

	resizeCount := 0
	appendCount := 0
	for _, typ := range types {
		if typ == "resize_begin" {
			resizeCount++
		}
		if typ == "append_end" {
			appendCount++
		}
	}
	if resizeCount != 1 {
		t.Errorf("expected 1 resize, got %d", resizeCount)
	}
	if appendCount != 3 {
		t.Errorf("expected 3 appends, got %d", appendCount)
	}
}

func TestDiffEndToEndWithCPython(t *testing.T) {
	src := `lst = []
for i in range(10):
    lst.append(i)
`
	script := InstrumentForObserve(src)
	snaps, err := RunCPython(script, 5*time.Second)
	if err != nil {
		t.Fatalf("RunCPython: %v", err)
	}

	evts := DiffSnapshots(snaps)

	// Should have append_begin/end for each of the 10 appends
	appendEnds := 0
	resizeBegins := 0
	for _, e := range evts {
		switch e.Type() {
		case "append_end":
			appendEnds++
		case "resize_begin":
			resizeBegins++
		}
	}

	if appendEnds != 10 {
		t.Errorf("expected 10 append_end events, got %d", appendEnds)
	}

	// CPython resizes: 0→4, 4→8, 8→16 = 3 resizes
	if resizeBegins != 3 {
		t.Errorf("expected 3 resizes, got %d", resizeBegins)
	}
}

func assertEventTypes(t *testing.T, evts []events.Event, expected []string) {
	t.Helper()
	got := eventTypes(evts)
	if len(got) != len(expected) {
		t.Fatalf("event count: got %d, want %d\ngot:  %v\nwant: %v", len(got), len(expected), got, expected)
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Errorf("event[%d]: got %q, want %q", i, got[i], expected[i])
		}
	}
}

func eventTypes(evts []events.Event) []string {
	out := make([]string, len(evts))
	for i, e := range evts {
		out[i] = e.Type()
	}
	return out
}
