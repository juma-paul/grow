package executor

import (
	"testing"
	"time"
)

func TestRunCPythonBasicAppends(t *testing.T) {
	script := `
import ctypes, json, sys

class _PyListStruct(ctypes.Structure):
    _fields_ = [
        ("ob_refcnt",  ctypes.c_ssize_t),
        ("ob_type",    ctypes.c_void_p),
        ("ob_size",    ctypes.c_ssize_t),
        ("ob_item",    ctypes.c_void_p),
        ("allocated",  ctypes.c_ssize_t),
    ]

def _snap(lst, tag):
    st = _PyListStruct.from_address(id(lst))
    addr = st.ob_item if st.ob_item is not None else 0
    sys.stdout.write(json.dumps({
        "tag":  tag,
        "len":  st.ob_size,
        "cap":  st.allocated,
        "addr": addr,
    }) + "\n")
    sys.stdout.flush()

lst = []
_snap(lst, "init")
for i in range(5):
    lst.append(i)
    _snap(lst, "append")
`
	snaps, err := RunCPython(script, 5*time.Second)
	if err != nil {
		t.Fatalf("RunCPython: %v", err)
	}

	if len(snaps) != 6 {
		t.Fatalf("expected 6 snapshots, got %d", len(snaps))
	}

	if snaps[0].Tag != "init" || snaps[0].Len != 0 || snaps[0].Cap != 0 {
		t.Errorf("init snapshot: %+v", snaps[0])
	}

	if snaps[5].Len != 5 || snaps[5].Cap < 5 {
		t.Errorf("final snapshot: len=%d cap=%d", snaps[5].Len, snaps[5].Cap)
	}

	// Capacity should grow: 0 → 4 → 8
	if snaps[1].Cap != 4 {
		t.Errorf("after first append: cap=%d, want 4", snaps[1].Cap)
	}
	if snaps[5].Cap != 8 {
		t.Errorf("after 5th append: cap=%d, want 8", snaps[5].Cap)
	}
}

func TestRunCPythonTimeout(t *testing.T) {
	script := `
import time
time.sleep(10)
`
	_, err := RunCPython(script, 500*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !contains(err.Error(), "timed out") {
		t.Errorf("expected timeout message, got: %v", err)
	}
}

func TestRunCPythonSyntaxError(t *testing.T) {
	_, err := RunCPython("def broken(", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for invalid syntax")
	}
}

func TestRunCPythonAddressChangesOnResize(t *testing.T) {
	script := `
import ctypes, json, sys

class _PyListStruct(ctypes.Structure):
    _fields_ = [
        ("ob_refcnt",  ctypes.c_ssize_t),
        ("ob_type",    ctypes.c_void_p),
        ("ob_size",    ctypes.c_ssize_t),
        ("ob_item",    ctypes.c_void_p),
        ("allocated",  ctypes.c_ssize_t),
    ]

def _snap(lst, tag):
    st = _PyListStruct.from_address(id(lst))
    addr = st.ob_item if st.ob_item is not None else 0
    sys.stdout.write(json.dumps({
        "tag":  tag,
        "len":  st.ob_size,
        "cap":  st.allocated,
        "addr": addr,
    }) + "\n")
    sys.stdout.flush()

lst = []
for i in range(20):
    lst.append(i)
    _snap(lst, "append")
`
	snaps, err := RunCPython(script, 5*time.Second)
	if err != nil {
		t.Fatalf("RunCPython: %v", err)
	}

	if len(snaps) != 20 {
		t.Fatalf("expected 20 snapshots, got %d", len(snaps))
	}

	// At least one address change should occur (resize moves the buffer)
	addrChanges := 0
	for i := 1; i < len(snaps); i++ {
		if snaps[i].Addr != snaps[i-1].Addr {
			addrChanges++
		}
	}
	if addrChanges == 0 {
		t.Error("expected at least one address change across 20 appends")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
