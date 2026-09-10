package stats

import (
	"testing"
)

func TestRecorderNilRedisNoOps(t *testing.T) {
	r := NewRecorder("")

	// None of these should panic with nil rdb
	r.RecordExecution("cpython", 100)
	r.RecordCountry("US")

	if err := r.Close(); err != nil {
		t.Errorf("Close on nil recorder: %v", err)
	}
}

func TestRecorderFireAndForget(t *testing.T) {
	// Connect to a non-existent Redis — should not block or panic
	r := NewRecorder("localhost:59999")
	defer r.Close()

	r.RecordExecution("doubling", 50)
	r.RecordCountry("KE")
}
