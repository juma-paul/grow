package simulator

import "github.com/juma-paul/grow/internal/events"

// LimitExceeded is a panic sentinel used to abort gpython execution
// when a sandbox limit is hit.
type LimitExceeded struct {
	Reason string
}

func (e LimitExceeded) Error() string { return e.Reason }

// Limits configures sandbox caps for a VisualList.
type Limits struct {
	MaxOps   int // max operations (0 = unlimited)
	MaxAlloc int // max cumulative allocated capacity (0 = unlimited)
}

func (l *VisualList) checkOpLimit() {
	if l.limits.MaxOps <= 0 {
		return
	}
	l.opCount++
	if l.opCount > l.limits.MaxOps {
		l.emit(events.LimitExceeded{
			Reason: "operation limit exceeded",
			Count:  l.opCount,
		})
		panic(LimitExceeded{Reason: "operation limit exceeded"})
	}
}

func (l *VisualList) checkAllocLimit(newCap int) {
	if l.limits.MaxAlloc <= 0 {
		return
	}
	if l.allocTotal+newCap > l.limits.MaxAlloc {
		l.emit(events.LimitExceeded{
			Reason: "allocation limit exceeded",
			Count:  l.allocTotal + newCap,
		})
		panic(LimitExceeded{Reason: "allocation limit exceeded"})
	}
}
