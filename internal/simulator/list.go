package simulator

import (
	"fmt"

	"github.com/juma-paul/grow/internal/events"
)

type VisualList struct {
	items     []any
	length    int
	allocated int
	strategy  GrowthStrategy
	emit      func(events.Event)
}

func (l *VisualList) resize(newLen int) {
	if l.allocated >= newLen && newLen >= l.allocated>>1 {
		l.length = newLen
		return
	}

	newCap := l.strategy.NextCapacity(newLen)
	oldCap := l.allocated
	growing := newCap > oldCap

	if growing {
		l.emit(events.ResizeBegin{OldCap: oldCap, NewCap: newCap})
	} else {
		l.emit(events.ShrinkBegin{OldCap: oldCap, NewCap: newCap})
	}

	newItems := make([]any, newCap)
	toCopy := l.length

	if newLen < toCopy {
		toCopy = newLen
	}

	for i := 0; i < toCopy; i++ {
		newItems[i] = l.items[i]
		l.emit(events.CopyElement{From: i, To: i, Value: l.items[i]})
	}

	l.items = newItems
	l.allocated = newCap
	l.length = newLen

	if growing {
		l.emit(events.ResizeEnd{Cost: toCopy})
	} else {
		l.emit(events.ShrinkEnd{Cost: toCopy})
	}
}

func NewVisualList(strategy GrowthStrategy, emit func(events.Event)) *VisualList {
	return &VisualList{
		items:     nil,
		length:    0,
		allocated: 0,
		strategy:  strategy,
		emit:      emit,
	}
}

func (l *VisualList) Append(value any) {
	l.emit(events.AppendBegin{
		Value:    value,
		Length:   l.length,
		Capacity: l.allocated,
	})

	oldLen := l.length
	l.resize(l.length + 1)
	l.items[oldLen] = value

	l.emit(events.AppendEnd{Cost: 1})
}

func (l *VisualList) Pop() any {
	if l.length == 0 {
		panic("pop from empty list")
	}

	l.emit(events.PopBegin{
		Length:   l.length,
		Capacity: l.allocated,
	})

	l.length--

	value := l.items[l.length]
	// clear the reference
	l.items[l.length] = nil
	l.resize(l.length)

	l.emit(events.PopEnd{Cost: 1})

	return value
}

func (l *VisualList) Insert(index int, value any) {
	if index < 0 || index > l.length {
		panic(fmt.Sprintf("insert index %d out of range for list of length %d", index, l.length))
	}

	l.emit(events.InsertBegin{
		Index:    index,
		Value:    value,
		Length:   l.length,
		Capacity: l.allocated,
	})

	l.resize(l.length + 1)

	// Shift elements right, from end toward insertion point
	for i := l.length - 1; i > index; i-- {
		l.items[i] = l.items[i-1]
		l.emit(events.ShiftRight{Index: i})
	}

	l.items[index] = value

	l.emit(events.InsertEnd{Cost: l.length - index})
}

func (l *VisualList) Extend(items []any) {
	hint := len(items)

	l.emit(events.ExtendBegin{
		Items:    hint,
		Length:   l.length,
		Capacity: l.allocated,
	})

	// Pre-size: one resize for all items (the length-hint optimization)
	l.resize(l.length + hint)

	// Place each element — no further resizes needed
	for i, v := range items {
		l.items[l.length-hint+i] = v
	}

	l.emit(events.ExtendEnd{Cost: hint})
}
