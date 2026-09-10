package server

import (
	"testing"

	"github.com/juma-paul/grow/internal/events"
)

func TestCoalesceEventsBelowThreshold(t *testing.T) {
	evts := make([]events.Event, 100)
	for i := 0; i < 50; i++ {
		evts[i*2] = events.AppendBegin{Length: i}
		evts[i*2+1] = events.AppendEnd{Cost: 1}
	}

	result, coalesced := CoalesceEvents(evts)
	if coalesced != 0 {
		t.Errorf("expected 0 coalesced, got %d", coalesced)
	}
	if len(result) != 100 {
		t.Errorf("expected 100 events unchanged, got %d", len(result))
	}
}

func TestCoalesceEventsAboveThreshold(t *testing.T) {
	n := coalesceAfter + 1000
	var evts []events.Event
	appendCount := 0
	for i := 0; i < n/2; i++ {
		evts = append(evts, events.AppendBegin{Length: i})
		evts = append(evts, events.AppendEnd{Cost: 1})
		appendCount++
	}

	result, coalesced := CoalesceEvents(evts)
	if coalesced == 0 {
		t.Error("expected coalescing to occur")
	}
	if len(result) >= len(evts) {
		t.Errorf("expected fewer events after coalescing: got %d, original %d", len(result), len(evts))
	}

	batchCount := 0
	for _, e := range result {
		if _, ok := e.(events.AppendBatch); ok {
			batchCount++
		}
	}
	if batchCount == 0 {
		t.Error("expected at least one AppendBatch event")
	}
}

func TestCoalescePreservesResizeEvents(t *testing.T) {
	var evts []events.Event
	resizeIdx := 500

	for i := 0; i < coalesceAfter+2000; i++ {
		if i == resizeIdx {
			evts = append(evts,
				events.AppendBegin{Length: i},
				events.ResizeBegin{OldCap: 4, NewCap: 8},
				events.CopyElement{From: 0, To: 0},
				events.CopyElement{From: 1, To: 1},
				events.CopyElement{From: 2, To: 2},
				events.CopyElement{From: 3, To: 3},
				events.ResizeEnd{Cost: 4},
				events.AppendEnd{Cost: 1},
			)
		} else {
			evts = append(evts, events.AppendBegin{Length: i})
			evts = append(evts, events.AppendEnd{Cost: 1})
		}
	}

	result, _ := CoalesceEvents(evts)

	resizeBegins := 0
	copyElements := 0
	resizeEnds := 0
	for _, e := range result {
		switch e.(type) {
		case events.ResizeBegin:
			resizeBegins++
		case events.CopyElement:
			copyElements++
		case events.ResizeEnd:
			resizeEnds++
		}
	}

	if resizeBegins != 1 {
		t.Errorf("expected 1 resize_begin, got %d", resizeBegins)
	}
	if copyElements != 4 {
		t.Errorf("expected 4 copy_element, got %d", copyElements)
	}
	if resizeEnds != 1 {
		t.Errorf("expected 1 resize_end, got %d", resizeEnds)
	}
}

func TestHighVolumeLoopStaysBounded(t *testing.T) {
	n := 20000
	var evts []events.Event
	cap := 0
	for i := 0; i < n; i++ {
		newCap := nextCPythonCap(i + 1)
		if newCap > cap {
			evts = append(evts, events.AppendBegin{Length: i, Capacity: cap})
			evts = append(evts, events.ResizeBegin{OldCap: cap, NewCap: newCap})
			evts = append(evts, events.CopyElement{From: 0, To: 0})
			evts = append(evts, events.ResizeEnd{Cost: i})
			evts = append(evts, events.AppendEnd{Cost: 1})
			cap = newCap
		} else {
			evts = append(evts, events.AppendBegin{Length: i})
			evts = append(evts, events.AppendEnd{Cost: 1})
		}
	}

	if len(evts) < coalesceAfter {
		t.Fatalf("test setup: expected > %d events, got %d", coalesceAfter, len(evts))
	}

	result, coalesced := CoalesceEvents(evts)

	if len(result) >= len(evts) {
		t.Errorf("coalescing should reduce count: %d -> %d", len(evts), len(result))
	}
	if coalesced == 0 {
		t.Error("expected some events to be coalesced")
	}

	resizeCount := 0
	for _, e := range result {
		if _, ok := e.(events.ResizeBegin); ok {
			resizeCount++
		}
	}
	if resizeCount == 0 {
		t.Error("resize pattern should be preserved")
	}

	origResizes := 0
	for _, e := range evts {
		if _, ok := e.(events.ResizeBegin); ok {
			origResizes++
		}
	}
	if resizeCount != origResizes {
		t.Errorf("resize count changed: original %d, after coalescing %d", origResizes, resizeCount)
	}
}

func nextCPythonCap(needed int) int {
	if needed <= 0 {
		return 0
	}
	newCap := needed + (needed >> 3)
	if needed < 9 {
		newCap += 3
	} else {
		newCap += 6
	}
	newCap = newCap &^ 3
	if newCap < 4 {
		newCap = 4
	}
	return newCap
}

func TestCoalesceBatchFields(t *testing.T) {
	var evts []events.Event
	for i := 0; i < coalesceAfter+100; i++ {
		evts = append(evts, events.AppendBegin{Length: i})
		evts = append(evts, events.AppendEnd{Cost: 1})
	}

	result, _ := CoalesceEvents(evts)

	var batch events.AppendBatch
	found := false
	for _, e := range result {
		if b, ok := e.(events.AppendBatch); ok {
			batch = b
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no AppendBatch found")
	}

	if batch.Count <= 1 {
		t.Errorf("batch count should be > 1, got %d", batch.Count)
	}
	if batch.ToLen != batch.FromLen+batch.Count {
		t.Errorf("to_len (%d) should equal from_len (%d) + count (%d)",
			batch.ToLen, batch.FromLen, batch.Count)
	}
}
