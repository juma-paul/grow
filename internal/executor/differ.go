package executor

import "github.com/juma-paul/grow/internal/events"

// DiffSnapshots converts a sequence of CPython snapshots into the
// standard event stream used by the frontend animator. Each consecutive
// pair of snapshots produces the events that describe the transition.
func DiffSnapshots(snaps []Snapshot) []events.Event {
	if len(snaps) < 2 {
		return nil
	}

	var out []events.Event

	for i := 1; i < len(snaps); i++ {
		prev := snaps[i-1]
		curr := snaps[i]
		out = append(out, diffPair(prev, curr)...)
	}

	return out
}

func diffPair(prev, curr Snapshot) []events.Event {
	var out []events.Event

	capGrew := curr.Cap > prev.Cap
	capShrank := curr.Cap < prev.Cap
	lenGrew := curr.Len > prev.Len
	lenShrank := curr.Len < prev.Len
	addrChanged := curr.Addr != prev.Addr

	if lenGrew {
		out = append(out, events.AppendBegin{
			Value:    curr.Len - 1,
			Length:   prev.Len,
			Capacity: prev.Cap,
		})

		if capGrew {
			out = append(out, events.ResizeBegin{
				OldCap: prev.Cap,
				NewCap: curr.Cap,
			})

			if addrChanged {
				for j := 0; j < prev.Len; j++ {
					out = append(out, events.CopyElement{
						From:  j,
						To:    j,
						Value: j,
					})
				}
			}

			out = append(out, events.ResizeEnd{
				Cost: prev.Len,
			})
		}

		out = append(out, events.AppendEnd{
			Cost: 1,
		})
	}

	if lenShrank {
		out = append(out, events.PopBegin{
			Length:   prev.Len,
			Capacity: prev.Cap,
		})

		if capShrank {
			out = append(out, events.ShrinkBegin{
				OldCap: prev.Cap,
				NewCap: curr.Cap,
			})

			if addrChanged {
				for j := 0; j < curr.Len; j++ {
					out = append(out, events.CopyElement{
						From:  j,
						To:    j,
						Value: j,
					})
				}
			}

			out = append(out, events.ShrinkEnd{
				Cost: curr.Len,
			})
		}

		out = append(out, events.PopEnd{
			Cost: 1,
		})
	}

	return out
}
