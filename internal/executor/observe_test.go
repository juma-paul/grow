package executor

import (
	"strings"
	"testing"
	"time"
)

func TestInstrumentBasicAppend(t *testing.T) {
	src := `lst = []
for i in range(5):
    lst.append(i)
`
	result := InstrumentForObserve(src)

	if !strings.Contains(result, "_PyListStruct") {
		t.Error("missing preamble")
	}
	if !strings.Contains(result, `_snap(lst, "init_1")`) {
		t.Error("missing init snap")
	}
	if !strings.Contains(result, `_snap(lst, "append_2")`) {
		t.Error("missing append snap")
	}
}

func TestInstrumentListCall(t *testing.T) {
	src := `x = list()
x.append(1)
x.pop()
`
	result := InstrumentForObserve(src)

	if !strings.Contains(result, `_snap(x, "init_1")`) {
		t.Error("missing init snap for list()")
	}
	if !strings.Contains(result, `_snap(x, "append_2")`) {
		t.Error("missing append snap")
	}
	if !strings.Contains(result, `_snap(x, "pop_3")`) {
		t.Error("missing pop snap")
	}
}

func TestInstrumentAugmentedAssign(t *testing.T) {
	src := `a = []
a += [1, 2, 3]
`
	result := InstrumentForObserve(src)

	if !strings.Contains(result, `_snap(a, "extend_2")`) {
		t.Error("missing extend snap for +=")
	}
}

func TestInstrumentPreservesNonListCode(t *testing.T) {
	src := `x = 42
print(x)
`
	result := InstrumentForObserve(src)

	// The preamble defines _snap, so count calls to _snap(x, or _snap(42,
	// There should be zero snap calls for non-list variables
	if strings.Contains(result, `_snap(x,`) {
		t.Error("should not inject snap for non-list code")
	}
	if !strings.Contains(result, "x = 42") {
		t.Error("should preserve original code")
	}
}

func TestInstrumentDoesNotTriggerOnComparison(t *testing.T) {
	src := `x = []
if x == []:
    pass
`
	result := InstrumentForObserve(src)

	count := strings.Count(result, `_snap(x,`)
	if count != 1 {
		t.Errorf("expected 1 snap call (init only), got %d", count)
	}
}

func TestInstrumentEndToEnd(t *testing.T) {
	src := `lst = []
for i in range(10):
    lst.append(i)
`
	script := InstrumentForObserve(src)
	snaps, err := RunCPython(script, 5*time.Second)
	if err != nil {
		t.Fatalf("RunCPython: %v", err)
	}

	// 1 init + 10 appends = 11 snapshots
	if len(snaps) != 11 {
		t.Fatalf("expected 11 snapshots, got %d", len(snaps))
	}

	if snaps[0].Tag != "init_1" {
		t.Errorf("first snap tag: %q", snaps[0].Tag)
	}

	if snaps[0].Len != 0 || snaps[0].Cap != 0 {
		t.Errorf("init snap: len=%d cap=%d", snaps[0].Len, snaps[0].Cap)
	}

	if snaps[10].Len != 10 {
		t.Errorf("final snap len=%d, want 10", snaps[10].Len)
	}

	// Verify growth pattern: 0 → 4 → 8 → 16
	caps := make(map[int]bool)
	for _, s := range snaps {
		caps[s.Cap] = true
	}
	for _, expected := range []int{0, 4, 8, 16} {
		if !caps[expected] {
			t.Errorf("expected to see cap=%d in growth sequence", expected)
		}
	}
}

func TestInstrumentMultipleLists(t *testing.T) {
	src := `a = []
b = []
a.append(1)
b.append(2)
`
	result := InstrumentForObserve(src)

	if !strings.Contains(result, `_snap(a, "init_1")`) {
		t.Error("missing init snap for a")
	}
	if !strings.Contains(result, `_snap(b, "init_2")`) {
		t.Error("missing init snap for b")
	}
	if !strings.Contains(result, `_snap(a, "append_3")`) {
		t.Error("missing append snap for a")
	}
	if !strings.Contains(result, `_snap(b, "append_4")`) {
		t.Error("missing append snap for b")
	}
}
