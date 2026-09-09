package simulator

import (
	"fmt"

	"github.com/juma-paul/grow/internal/events"
)

// VisualList wraps a dynamic array with event emission for visualization.
type VisualList struct {
	items      []any
	length     int
	allocated  int
	strategy   GrowthStrategy
	emit       func(events.Event)
	limits     Limits
	opCount    int
	allocTotal int
}

func (l *VisualList) resize(newLen int) {
	if l.allocated >= newLen && newLen >= l.allocated>>1 {
		l.length = newLen
		return
	}

	newCap := l.strategy.NextCapacity(newLen)
	if newCap < 0 {
		l.emit(events.Overflow{Needed: newLen, Capacity: l.allocated, SourceRef: "list_resize:overflow"})
		panic(OverflowExceeded{})
	}
	oldCap := l.allocated
	if newCap == oldCap {
		l.length = newLen
		return
	}
	growing := newCap > oldCap

	if growing {
		l.emit(events.ResizeBegin{OldCap: oldCap, NewCap: newCap, SourceRef: "list_resize:growth-formula"})
	} else {
		l.emit(events.ShrinkBegin{OldCap: oldCap, NewCap: newCap, SourceRef: "list_resize:shrink-check"})
	}

	l.checkAllocLimit(newCap)
	l.allocTotal += newCap

	newItems := make([]any, newCap)
	toCopy := min(l.length, newLen)

	for i := 0; i < toCopy; i++ {
		newItems[i] = l.items[i]
		l.emit(events.CopyElement{From: i, To: i, Value: l.items[i], SourceRef: "list_resize:memcpy"})
	}

	l.items = newItems
	l.allocated = newCap
	l.length = newLen

	if growing {
		l.emit(events.ResizeEnd{Cost: toCopy, SourceRef: "list_resize:growth-formula"})
	} else {
		l.emit(events.ShrinkEnd{Cost: toCopy, SourceRef: "list_resize:shrink-check"})
	}
}

// Len returns the number of elements in the list.
func (l *VisualList) Len() int { return l.length }

func NewVisualList(strategy GrowthStrategy, emit func(events.Event)) *VisualList {
	return &VisualList{
		strategy: strategy,
		emit:     emit,
	}
}

// NewVisualListWithLimits creates a VisualList with sandbox caps.
func NewVisualListWithLimits(strategy GrowthStrategy, emit func(events.Event), limits Limits) *VisualList {
	return &VisualList{
		strategy: strategy,
		emit:     emit,
		limits:   limits,
	}
}

func (l *VisualList) Append(value any) {
	l.checkOpLimit()
	l.emit(events.AppendBegin{
		Value:     value,
		Length:    l.length,
		Capacity:  l.allocated,
		SourceRef: "list_append:PyList_Append",
	})

	oldLen := l.length
	l.resize(l.length + 1)
	l.items[oldLen] = value

	l.emit(events.AppendEnd{Cost: 1, SourceRef: "list_append:PyList_Append"})
}

func (l *VisualList) Pop() any {
	l.checkOpLimit()
	if l.length == 0 {
		panic("pop from empty list")
	}

	l.emit(events.PopBegin{
		Length:    l.length,
		Capacity:  l.allocated,
		SourceRef: "list_pop:PyList_Pop",
	})

	l.length--

	value := l.items[l.length]
	l.items[l.length] = nil
	l.resize(l.length)

	l.emit(events.PopEnd{Cost: 1, SourceRef: "list_pop:PyList_Pop"})

	return value
}

func (l *VisualList) Insert(index int, value any) {
	l.checkOpLimit()
	if index < 0 || index > l.length {
		panic(fmt.Sprintf("insert index %d out of range for list of length %d", index, l.length))
	}

	l.emit(events.InsertBegin{
		Index:     index,
		Value:     value,
		Length:    l.length,
		Capacity:  l.allocated,
		SourceRef: "list_insert:ins1",
	})

	l.resize(l.length + 1)

	for i := l.length - 1; i > index; i-- {
		l.items[i] = l.items[i-1]
		l.emit(events.ShiftRight{Index: i, SourceRef: "list_insert:ins1"})
	}

	l.items[index] = value

	l.emit(events.InsertEnd{Cost: l.length - index, SourceRef: "list_insert:ins1"})
}

func (l *VisualList) Extend(items []any) {
	l.checkOpLimit()
	hint := len(items)

	l.emit(events.ExtendBegin{
		Items:     hint,
		Length:    l.length,
		Capacity:  l.allocated,
		SourceRef: "list_extend:PyList_Extend",
	})

	l.resize(l.length + hint)

	for i, v := range items {
		l.items[l.length-hint+i] = v
	}

	l.emit(events.ExtendEnd{Cost: hint, SourceRef: "list_extend:PyList_Extend"})
}
