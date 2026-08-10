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
