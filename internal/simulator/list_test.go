package simulator

import (
	"testing"

	"github.com/juma-paul/grow/internal/events"
)

func TestAppend100(t *testing.T) {
	var log []events.Event
	lst := NewVisualList(CPythonGrowth{}, func(e events.Event) {
		log = append(log, e)
	})

	for i := 0; i < 100; i++ {
		lst.Append(i)
	}

	// Check final state (100 items + CPython's over-allocation)
	if lst.length != 100 {
		t.Errorf("length = %d, want 100", lst.length)
	}
	if lst.allocated != 108 {
		t.Errorf("allocated = %d, want 108", lst.allocated)
	}

	// Collect which append numbers triggered a resize
	wantResizeAt := []int{1, 5, 9, 17, 25, 33, 41, 53, 65, 77, 93}
	var gotResizedAt []int

	appendNum := 0
	for _, e := range log {
		if _, ok := e.(events.AppendBegin); ok {
			appendNum++
		}
		if _, ok := e.(events.ResizeBegin); ok {
			gotResizedAt = append(gotResizedAt, appendNum)
		}
	}

	if len(gotResizedAt) != len(wantResizeAt) {
		t.Fatalf("resize count = %d, want = %d", len(gotResizedAt), len(wantResizeAt))
	}
	for i := range wantResizeAt {
		if gotResizedAt[i] != wantResizeAt[i] {
			t.Errorf("resize #%d at append %d, want %d", i+1, gotResizedAt[i], wantResizeAt[i])
		}
	}
}

func TestPopShrink(t *testing.T) {
	var log []events.Event
	lst := NewVisualList(CPythonGrowth{}, func(e events.Event) {
		log = append(log, e)
	})

	// Build up to allocated=76, len=65
	for i := 0; i < 65; i++ {
		lst.Append(i)
	}

	// Verify starting state
	if lst.allocated != 76 {
		t.Fatalf("setup: allocated = %d, want 76", lst.allocated)
	}

	// Clear the log — we only care about pop events
	log = nil

	// Pop 28 times (len goes from 65 to 37)
	for i := 0; i < 28; i++ {
		lst.Pop()
	}

	// Should be exactly one shrink
	var shrinkEvents []events.ShrinkBegin
	for _, e := range log {
		if s, ok := e.(events.ShrinkBegin); ok {
			shrinkEvents = append(shrinkEvents, s)
		}
	}

	if len(shrinkEvents) != 1 {
		t.Fatalf("shrink count = %d, want 1", len(shrinkEvents))
	}

	// Shrink should be from 76 to 44
	if shrinkEvents[0].OldCap != 76 {
		t.Errorf("shrink old_cap = %d, want 76", shrinkEvents[0].OldCap)
	}
	if shrinkEvents[0].NewCap != 44 {
		t.Errorf("shrink new_cap = %d, want 44", shrinkEvents[0].NewCap)
	}

	// Final state
	if lst.length != 37 {
		t.Errorf("length = %d, want 37", lst.length)
	}
	if lst.allocated != 44 {
		t.Errorf("allocated = %d, want 44", lst.allocated)
	}
}
