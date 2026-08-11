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

func TestInsertNoResize(t *testing.T) {
	var log []events.Event
	lst := NewVisualList(CPythonGrowth{}, func(e events.Event) {
		log = append(log, e)
	})

	// Build [1,2,3,4,5] — capacity will be 8, plenty of room
	for i := 1; i <= 5; i++ {
		lst.Append(i)
	}
	log = nil

	// Insert 99 at index 0
	lst.Insert(0, 99)

	// Should be zero ResizeBegin events
	for _, e := range log {
		if _, ok := e.(events.ResizeBegin); ok {
			t.Error("unexpected resize during insert")
		}
	}

	// Should be exactly 5 ShiftRight events
	var shifts []events.ShiftRight
	for _, e := range log {
		if s, ok := e.(events.ShiftRight); ok {
			shifts = append(shifts, s)
		}
	}

	if len(shifts) != 5 {
		t.Fatalf("shift count = %d, want 5", len(shifts))
	}

	// Shifts should be at indices 5, 4, 3, 2, 1 (right to left)
	wantIndices := []int{5, 4, 3, 2, 1}
	for i, s := range shifts {
		if s.Index != wantIndices[i] {
			t.Errorf("shift #%d at index %d, want %d", i+1, s.Index, wantIndices[i])
		}
	}

	// Verify final list content
	if lst.items[0] != 99 {
		t.Errorf("items[0] = %v, want 99", lst.items[0])
	}
	if lst.length != 6 {
		t.Errorf("length = %d, want 6", lst.length)
	}
}
