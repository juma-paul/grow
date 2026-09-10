package stats

import (
	"testing"
	"time"
)

func TestRecorderNilRedisNoOps(t *testing.T) {
	r := NewRecorder("")

	r.RecordExecution("cpython", 100)
	r.RecordCountry("US")

	if err := r.Close(); err != nil {
		t.Errorf("Close on nil recorder: %v", err)
	}
}

func TestRecorderFireAndForget(t *testing.T) {
	r := NewRecorder("localhost:59999")
	defer r.Close()

	r.RecordExecution("doubling", 50)
	r.RecordCountry("KE")
}

func TestRecorderNeverBlocks(t *testing.T) {
	r := NewRecorder("localhost:59999")
	defer r.Close()

	done := make(chan struct{})
	go func() {
		r.RecordExecution("cpython", 1000)
		r.RecordCountry("US")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("single RecordExecution blocked >3s — not fire-and-forget")
	}
}

func TestFlusherNilForEmptyAddrs(t *testing.T) {
	f := NewFlusher("", "", 30*time.Second)
	if f != nil {
		t.Fatal("expected nil flusher for empty addrs")
	}
}

func TestFlusherNilSafe(t *testing.T) {
	var f *Flusher
	f.Start()
	f.Stop()
}
