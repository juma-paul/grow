package simulator

import (
	"testing"

	"github.com/juma-paul/grow/internal/events"
)

func TestOpLimitAborts(t *testing.T) {
	var log []events.Event
	lst := NewVisualListWithLimits(CPythonGrowth{}, func(e events.Event) {
		log = append(log, e)
	}, Limits{MaxOps: 10})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		le, ok := r.(LimitExceeded)
		if !ok {
			t.Fatalf("expected LimitExceeded, got %T: %v", r, r)
		}
		if le.Reason != "operation limit exceeded" {
			t.Errorf("reason = %q", le.Reason)
		}
		// Check that a LimitExceeded event was emitted
		found := false
		for _, e := range log {
			if e.Type() == "limit_exceeded" {
				found = true
			}
		}
		if !found {
			t.Error("no limit_exceeded event emitted")
		}
	}()

	for i := 0; i < 100; i++ {
		lst.Append(i)
	}
	t.Fatal("loop should not complete")
}

func TestAllocLimitAborts(t *testing.T) {
	var log []events.Event
	lst := NewVisualListWithLimits(CPythonGrowth{}, func(e events.Event) {
		log = append(log, e)
	}, Limits{MaxAlloc: 50})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		le, ok := r.(LimitExceeded)
		if !ok {
			t.Fatalf("expected LimitExceeded, got %T: %v", r, r)
		}
		if le.Reason != "allocation limit exceeded" {
			t.Errorf("reason = %q", le.Reason)
		}
		found := false
		for _, e := range log {
			if e.Type() == "limit_exceeded" {
				found = true
			}
		}
		if !found {
			t.Error("no limit_exceeded event emitted")
		}
	}()

	for i := 0; i < 100; i++ {
		lst.Append(i)
	}
	t.Fatal("loop should not complete")
}

func TestOpLimitUnlimited(t *testing.T) {
	lst := NewVisualList(CPythonGrowth{}, func(e events.Event) {})
	for i := 0; i < 1000; i++ {
		lst.Append(i)
	}
	if lst.Len() != 1000 {
		t.Errorf("Len() = %d, want 1000", lst.Len())
	}
}

func TestBothLimitsOpFiringFirst(t *testing.T) {
	lst := NewVisualListWithLimits(CPythonGrowth{}, func(e events.Event) {
	}, Limits{MaxOps: 5, MaxAlloc: 10_000_000})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		le, ok := r.(LimitExceeded)
		if !ok {
			t.Fatalf("expected LimitExceeded, got %T", r)
		}
		if le.Reason != "operation limit exceeded" {
			t.Errorf("expected op limit, got %q", le.Reason)
		}
	}()

	for i := 0; i < 100; i++ {
		lst.Append(i)
	}
}
