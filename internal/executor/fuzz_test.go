package executor

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/juma-paul/grow/internal/events"
	"github.com/juma-paul/grow/internal/simulator"
	"pgregory.net/rapid"

	_ "github.com/go-python/gpython/stdlib"
)

type fuzzOp struct {
	kind  fuzzOpKind
	index int
	count int
}

type fuzzOpKind int

const (
	fuzzAppend fuzzOpKind = iota
	fuzzPop
	fuzzInsert
	fuzzExtend
)

func normalizeOps(ops []fuzzOp) []fuzzOp {
	var result []fuzzOp
	length := 0
	for _, o := range ops {
		switch o.kind {
		case fuzzAppend:
			result = append(result, o)
			length++
		case fuzzPop:
			if length == 0 {
				continue
			}
			result = append(result, o)
			length--
		case fuzzInsert:
			idx := o.index
			if length > 0 {
				idx = idx % (length + 1)
			} else {
				idx = 0
			}
			result = append(result, fuzzOp{kind: fuzzInsert, index: idx})
			length++
		case fuzzExtend:
			c := o.count
			if c <= 0 {
				c = 1
			}
			result = append(result, fuzzOp{kind: fuzzExtend, count: c})
			length += c
		}
	}
	return result
}

func buildWrapperSource(ops []fuzzOp) string {
	var b strings.Builder
	b.WriteString("lst = VisualList()\n")
	for _, o := range ops {
		switch o.kind {
		case fuzzAppend:
			b.WriteString("lst.append(0)\n")
		case fuzzPop:
			b.WriteString("lst.pop()\n")
		case fuzzInsert:
			fmt.Fprintf(&b, "lst.insert(%d, 0)\n", o.index)
		case fuzzExtend:
			items := strings.Repeat("0,", o.count)
			items = items[:len(items)-1]
			fmt.Fprintf(&b, "lst.extend([%s])\n", items)
		}
	}
	return b.String()
}

func runDirect(ops []fuzzOp) []events.Event {
	var collected []events.Event
	vl := simulator.NewVisualList(simulator.CPythonGrowth{}, func(e events.Event) {
		collected = append(collected, e)
	})
	for _, o := range ops {
		switch o.kind {
		case fuzzAppend:
			vl.Append(0)
		case fuzzPop:
			vl.Pop()
		case fuzzInsert:
			vl.Insert(o.index, 0)
		case fuzzExtend:
			items := make([]any, o.count)
			vl.Extend(items)
		}
	}
	return collected
}

func eventsStructurallyEqual(a, b []events.Event) error {
	if len(a) != len(b) {
		return fmt.Errorf("event count: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Type() != b[i].Type() {
			return fmt.Errorf("event %d type: %q vs %q", i, a[i].Type(), b[i].Type())
		}
		switch ea := a[i].(type) {
		case events.AppendBegin:
			eb := b[i].(events.AppendBegin)
			if ea.Length != eb.Length || ea.Capacity != eb.Capacity {
				return fmt.Errorf("event %d AppendBegin: L/C %d/%d vs %d/%d",
					i, ea.Length, ea.Capacity, eb.Length, eb.Capacity)
			}
		case events.AppendEnd:
			eb := b[i].(events.AppendEnd)
			if ea.Cost != eb.Cost {
				return fmt.Errorf("event %d AppendEnd cost: %d vs %d", i, ea.Cost, eb.Cost)
			}
		case events.ResizeBegin:
			eb := b[i].(events.ResizeBegin)
			if ea.OldCap != eb.OldCap || ea.NewCap != eb.NewCap {
				return fmt.Errorf("event %d ResizeBegin: %d->%d vs %d->%d",
					i, ea.OldCap, ea.NewCap, eb.OldCap, eb.NewCap)
			}
		case events.ResizeEnd:
			eb := b[i].(events.ResizeEnd)
			if ea.Cost != eb.Cost {
				return fmt.Errorf("event %d ResizeEnd cost: %d vs %d", i, ea.Cost, eb.Cost)
			}
		case events.ShrinkBegin:
			eb := b[i].(events.ShrinkBegin)
			if ea.OldCap != eb.OldCap || ea.NewCap != eb.NewCap {
				return fmt.Errorf("event %d ShrinkBegin: %d->%d vs %d->%d",
					i, ea.OldCap, ea.NewCap, eb.OldCap, eb.NewCap)
			}
		case events.ShrinkEnd:
			eb := b[i].(events.ShrinkEnd)
			if ea.Cost != eb.Cost {
				return fmt.Errorf("event %d ShrinkEnd cost: %d vs %d", i, ea.Cost, eb.Cost)
			}
		case events.CopyElement:
			eb := b[i].(events.CopyElement)
			if ea.From != eb.From || ea.To != eb.To {
				return fmt.Errorf("event %d CopyElement: %d->%d vs %d->%d",
					i, ea.From, ea.To, eb.From, eb.To)
			}
		case events.PopBegin:
			eb := b[i].(events.PopBegin)
			if ea.Length != eb.Length || ea.Capacity != eb.Capacity {
				return fmt.Errorf("event %d PopBegin: L/C %d/%d vs %d/%d",
					i, ea.Length, ea.Capacity, eb.Length, eb.Capacity)
			}
		case events.PopEnd:
			eb := b[i].(events.PopEnd)
			if ea.Cost != eb.Cost {
				return fmt.Errorf("event %d PopEnd cost: %d vs %d", i, ea.Cost, eb.Cost)
			}
		case events.InsertBegin:
			eb := b[i].(events.InsertBegin)
			if ea.Index != eb.Index || ea.Length != eb.Length || ea.Capacity != eb.Capacity {
				return fmt.Errorf("event %d InsertBegin: I/L/C %d/%d/%d vs %d/%d/%d",
					i, ea.Index, ea.Length, ea.Capacity, eb.Index, eb.Length, eb.Capacity)
			}
		case events.InsertEnd:
			eb := b[i].(events.InsertEnd)
			if ea.Cost != eb.Cost {
				return fmt.Errorf("event %d InsertEnd cost: %d vs %d", i, ea.Cost, eb.Cost)
			}
		case events.ShiftRight:
			eb := b[i].(events.ShiftRight)
			if ea.Index != eb.Index {
				return fmt.Errorf("event %d ShiftRight: %d vs %d", i, ea.Index, eb.Index)
			}
		case events.ExtendBegin:
			eb := b[i].(events.ExtendBegin)
			if ea.Items != eb.Items || ea.Length != eb.Length || ea.Capacity != eb.Capacity {
				return fmt.Errorf("event %d ExtendBegin: N/L/C %d/%d/%d vs %d/%d/%d",
					i, ea.Items, ea.Length, ea.Capacity, eb.Items, eb.Length, eb.Capacity)
			}
		case events.ExtendEnd:
			eb := b[i].(events.ExtendEnd)
			if ea.Cost != eb.Cost {
				return fmt.Errorf("event %d ExtendEnd cost: %d vs %d", i, ea.Cost, eb.Cost)
			}
		}
	}
	return nil
}

func generateCorpusOps(seed int) []fuzzOp {
	rng := rand.New(rand.NewSource(int64(seed)))
	n := 5 + rng.Intn(96)
	ops := make([]fuzzOp, n)
	for i := range ops {
		kind := rng.Intn(4)
		ops[i] = fuzzOp{
			kind:  fuzzOpKind(kind),
			index: rng.Intn(201),
			count: 1 + rng.Intn(10),
		}
	}
	return ops
}

func runParity(t *testing.T, raw []fuzzOp) {
	t.Helper()
	ops := normalizeOps(raw)
	if len(ops) == 0 {
		return
	}

	direct := runDirect(ops)

	src := buildWrapperSource(ops)
	gpythonEvts, err := RunWrapper(src, 10*time.Second)
	if err != nil {
		t.Fatalf("RunWrapper: %v\nsource:\n%s", err, src)
	}

	if err := eventsStructurallyEqual(direct, gpythonEvts); err != nil {
		t.Fatalf("parity: %v\nops=%d direct=%d gpython=%d\nsource:\n%s",
			err, len(ops), len(direct), len(gpythonEvts), src)
	}
}

func TestFuzzCorpus(t *testing.T) {
	for i := range 100 {
		t.Run(fmt.Sprintf("case_%03d", i), func(t *testing.T) {
			raw := generateCorpusOps(i)
			runParity(t, raw)
		})
	}
}

func TestFuzzPropertyWrapperParity(t *testing.T) {
	for round := range 5 {
		t.Run(fmt.Sprintf("round_%d", round), func(t *testing.T) {
			rapid.Check(t, func(t *rapid.T) {
				n := rapid.IntRange(1, 100).Draw(t, "numOps")
				ops := make([]fuzzOp, n)
				for i := range ops {
					kind := rapid.IntRange(0, 3).Draw(t, "opKind")
					ops[i] = fuzzOp{
						kind:  fuzzOpKind(kind),
						index: rapid.IntRange(0, 200).Draw(t, "index"),
						count: rapid.IntRange(1, 10).Draw(t, "count"),
					}
				}

				normalized := normalizeOps(ops)
				if len(normalized) == 0 {
					return
				}

				direct := runDirect(normalized)

				src := buildWrapperSource(normalized)
				gpythonEvts, err := RunWrapper(src, 10*time.Second)
				if err != nil {
					t.Fatalf("RunWrapper: %v", err)
				}

				if err := eventsStructurallyEqual(direct, gpythonEvts); err != nil {
					t.Fatalf("parity: %v\nops=%d direct=%d gpython=%d",
						err, len(normalized), len(direct), len(gpythonEvts))
				}
			})
		})
	}
}
