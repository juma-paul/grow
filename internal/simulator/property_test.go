package simulator

import (
	"testing"

	"github.com/juma-paul/grow/internal/events"
	"pgregory.net/rapid"
)

// refAllocated is a standalone reference impl of CPython's list allocation.
// It tracks what allocated should be after a sequence of operations.
func refAllocated(ops []op) int {
	length := 0
	allocated := 0
	strat := CPythonGrowth{}

	for _, o := range ops {
		switch o.kind {
		case opAppend:
			newLen := length + 1
			if allocated >= newLen && newLen >= allocated>>1 {
				length = newLen
			} else {
				allocated = strat.NextCapacity(newLen)
				length = newLen
			}
		case opPop:
			if length == 0 {
				continue
			}
			newLen := length - 1
			length = newLen
			if allocated >= newLen && newLen >= allocated>>1 {
				// no resize
			} else {
				allocated = strat.NextCapacity(newLen)
			}
		case opInsert:
			idx := o.index
			if idx < 0 || idx > length {
				continue
			}
			newLen := length + 1
			if allocated >= newLen && newLen >= allocated>>1 {
				length = newLen
			} else {
				allocated = strat.NextCapacity(newLen)
				length = newLen
			}
		case opExtend:
			count := o.count
			if count <= 0 {
				continue
			}
			newLen := length + count
			if allocated >= newLen && newLen >= allocated>>1 {
				length = newLen
			} else {
				allocated = strat.NextCapacity(newLen)
				length = newLen
			}
		}
	}

	return allocated
}

type opKind int

const (
	opAppend opKind = iota
	opPop
	opInsert
	opExtend
)

type op struct {
	kind  opKind
	index int
	count int
}

func genOps(t *rapid.T) []op {
	n := rapid.IntRange(1, 500).Draw(t, "numOps")
	ops := make([]op, n)
	for i := range ops {
		kind := rapid.IntRange(0, 3).Draw(t, "opKind")
		ops[i] = op{
			kind:  opKind(kind),
			index: rapid.IntRange(0, 500).Draw(t, "index"),
			count: rapid.IntRange(1, 20).Draw(t, "count"),
		}
	}
	return ops
}

func applyOps(vl *VisualList, ops []op) {
	for _, o := range ops {
		switch o.kind {
		case opAppend:
			vl.Append(0)
		case opPop:
			if vl.Len() == 0 {
				continue
			}
			vl.Pop()
		case opInsert:
			idx := o.index
			if idx < 0 || idx > vl.Len() {
				continue
			}
			vl.Insert(idx, 0)
		case opExtend:
			items := make([]any, o.count)
			vl.Extend(items)
		}
	}
}

func TestPropertyResizeCapMatchesFormula(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ops := genOps(t)

		strat := CPythonGrowth{}
		var pendingNewLen int
		resizeIdx := 0

		collect := func(e events.Event) {
			switch ev := e.(type) {
			case events.AppendBegin:
				pendingNewLen = ev.Length + 1
			case events.InsertBegin:
				pendingNewLen = ev.Length + 1
			case events.ExtendBegin:
				pendingNewLen = ev.Length + ev.Items
			case events.ResizeBegin:
				expected := strat.NextCapacity(pendingNewLen)
				if ev.NewCap != expected {
					t.Fatalf("resize #%d: new_cap=%d but NextCapacity(%d)=%d",
						resizeIdx, ev.NewCap, pendingNewLen, expected)
				}
				resizeIdx++
			}
		}

		vl := NewVisualList(strat, collect)
		applyOps(vl, ops)
	})
}

func TestPropertyShrinkFiresIffConditionMet(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ops := genOps(t)

		noop := func(events.Event) {}
		vl := NewVisualList(CPythonGrowth{}, noop)

		type popRecord struct {
			newLen   int
			oldAlloc int
			shrunk   bool
		}
		var popRecords []popRecord

		for _, o := range ops {
			switch o.kind {
			case opAppend:
				vl.Append(0)
			case opPop:
				if vl.Len() == 0 {
					continue
				}
				beforeLen := vl.length
				beforeAlloc := vl.allocated
				vl.Pop()
				popRecords = append(popRecords, popRecord{
					newLen:   beforeLen - 1,
					oldAlloc: beforeAlloc,
					shrunk:   vl.allocated != beforeAlloc,
				})
			case opInsert:
				if o.index < 0 || o.index > vl.Len() {
					continue
				}
				vl.Insert(o.index, 0)
			case opExtend:
				vl.Extend(make([]any, o.count))
			}
		}

		strat := CPythonGrowth{}
		for i, pr := range popRecords {
			belowHalf := pr.newLen < pr.oldAlloc>>1
			newCap := strat.NextCapacity(pr.newLen)
			wouldChange := newCap != pr.oldAlloc
			shouldShrink := belowHalf && wouldChange
			if pr.shrunk && !shouldShrink {
				t.Fatalf("pop #%d: shrunk but condition not met (newLen=%d, oldAlloc=%d, newCap=%d)",
					i, pr.newLen, pr.oldAlloc, newCap)
			}
			if !pr.shrunk && shouldShrink {
				t.Fatalf("pop #%d: did NOT shrink but condition met (newLen=%d < oldAlloc/2=%d, newCap=%d != oldAlloc=%d)",
					i, pr.newLen, pr.oldAlloc>>1, newCap, pr.oldAlloc)
			}
		}
	})
}

func TestPropertyAllocatedMatchesReference(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ops := genOps(t)

		noop := func(events.Event) {}
		vl := NewVisualList(CPythonGrowth{}, noop)
		applyOps(vl, ops)

		expected := refAllocated(ops)
		if vl.allocated != expected {
			t.Fatalf("allocated mismatch: simulator=%d reference=%d", vl.allocated, expected)
		}
	})
}
