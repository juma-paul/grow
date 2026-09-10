package events

import (
	"encoding/json"
	"fmt"
)

// Event is the common interface for all simulator events.
type Event interface {
	Type() string
}

// AppendBegin fires when an append operation starts.
type AppendBegin struct {
	Value     any    `json:"value"`
	Length    int    `json:"length"`
	Capacity  int    `json:"capacity"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (AppendBegin) Type() string { return "append_begin" }

// AppendEnd fires when an append operation completes.
type AppendEnd struct {
	Cost      int    `json:"cost"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (AppendEnd) Type() string { return "append_end" }

// ResizeBegin fires when the array grows to a larger capacity.
type ResizeBegin struct {
	OldCap    int    `json:"old_cap"`
	NewCap    int    `json:"new_cap"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (ResizeBegin) Type() string { return "resize_begin" }

// ResizeEnd fires when a resize completes.
type ResizeEnd struct {
	Cost      int    `json:"cost"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (ResizeEnd) Type() string { return "resize_end" }

// CopyElement fires for each element copied during a resize.
type CopyElement struct {
	From      int    `json:"from"`
	To        int    `json:"to"`
	Value     any    `json:"value"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (CopyElement) Type() string { return "copy_element" }

// ShrinkBegin fires when the array shrinks to a smaller capacity.
type ShrinkBegin struct {
	OldCap    int    `json:"old_cap"`
	NewCap    int    `json:"new_cap"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (ShrinkBegin) Type() string { return "shrink_begin" }

// ShrinkEnd fires when a shrink completes.
type ShrinkEnd struct {
	Cost      int    `json:"cost"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (ShrinkEnd) Type() string { return "shrink_end" }

// PopBegin fires when a pop operation starts.
type PopBegin struct {
	Length    int    `json:"length"`
	Capacity  int    `json:"capacity"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (PopBegin) Type() string { return "pop_begin" }

// PopEnd fires when a pop operation completes.
type PopEnd struct {
	Cost      int    `json:"cost"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (PopEnd) Type() string { return "pop_end" }

// InsertBegin fires when an insert operation starts.
type InsertBegin struct {
	Index     int    `json:"index"`
	Value     any    `json:"value"`
	Length    int    `json:"length"`
	Capacity  int    `json:"capacity"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (InsertBegin) Type() string { return "insert_begin" }

// InsertEnd fires when an insert operation completes.
type InsertEnd struct {
	Cost      int    `json:"cost"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (InsertEnd) Type() string { return "insert_end" }

// ShiftRight fires for each element shifted during an insert.
type ShiftRight struct {
	Index     int    `json:"index"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (ShiftRight) Type() string { return "shift_right" }

// ExtendBegin fires when an extend operation starts.
type ExtendBegin struct {
	Items     int    `json:"items"`
	Length    int    `json:"length"`
	Capacity  int    `json:"capacity"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (ExtendBegin) Type() string { return "extend_begin" }

// ExtendEnd fires when an extend operation completes.
type ExtendEnd struct {
	Cost      int    `json:"cost"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (ExtendEnd) Type() string { return "extend_end" }

// Overflow fires when a NoGrowth strategy cannot accommodate a resize.
type Overflow struct {
	Needed    int    `json:"needed"`
	Capacity  int    `json:"capacity"`
	SourceRef string `json:"source_ref,omitempty"`
}

func (Overflow) Type() string { return "overflow" }

// AppendBatch summarises a run of consecutive non-resize appends
// that were coalesced to reduce event volume.
type AppendBatch struct {
	Count   int `json:"count"`
	FromLen int `json:"from_len"`
	ToLen   int `json:"to_len"`
}

func (AppendBatch) Type() string { return "append_batch" }

// LimitExceeded fires when an execution hits a sandbox cap.
type LimitExceeded struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

func (LimitExceeded) Type() string { return "limit_exceeded" }

// Marshal converts an Event to JSON with a "type" field injected.
// The type value comes from e.Type() and is prepended to the struct's
// serialized fields, producing output like:
//
//	{"type":"resize_begin","old_cap":4,"new_cap":8}
func Marshal(e Event) ([]byte, error) {
	raw, err := json.Marshal(e)

	if err != nil {
		return nil, err
	}

	typeField := fmt.Sprintf(`"type":"%s"`, e.Type())

	inner := string(raw[1 : len(raw)-1]) // strip { and }
	if len(inner) > 0 {
		return []byte("{" + typeField + "," + inner + "}"), nil
	}
	return []byte("{" + typeField + "}"), nil
}
