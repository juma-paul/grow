package simulator

import "github.com/juma-paul/grow/internal/events"

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
