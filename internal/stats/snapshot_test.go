package stats

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSnapshotCacheNoRedis(t *testing.T) {
	sc := NewSnapshotCache("", "", 10*time.Second)
	sc.Start()
	defer sc.Stop()

	snap := sc.Get()
	if snap.TotalRuns != 0 || snap.RunsToday != 0 || snap.ElementsAllocated != 0 {
		t.Fatalf("expected zero snapshot, got %+v", snap)
	}
}

func TestSnapshotJSON(t *testing.T) {
	sc := NewSnapshotCache("", "", 10*time.Second)
	sc.Start()
	defer sc.Stop()

	data, err := sc.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if snap.TotalRuns != 0 {
		t.Fatalf("expected zero total_runs, got %d", snap.TotalRuns)
	}
}
