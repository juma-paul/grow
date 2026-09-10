package stats

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Snapshot is the cached aggregate served by GET /stats.
type Snapshot struct {
	TotalRuns         int64  `json:"total_runs"`
	RunsToday         int64  `json:"runs_today"`
	ElementsAllocated int64  `json:"elements_allocated"`
	CachedAt          string `json:"cached_at"`
}

// SnapshotCache holds a periodically refreshed stats snapshot
// so the /stats endpoint never queries the database on the hot path.
type SnapshotCache struct {
	rdb      *redis.Client
	mu       sync.RWMutex
	snapshot Snapshot
	stop     chan struct{}
	done     chan struct{}
}

// NewSnapshotCache creates a cache that refreshes from Redis every interval.
// If redisAddr is empty, returns a cache with zero values.
func NewSnapshotCache(redisAddr string, interval time.Duration) *SnapshotCache {
	sc := &SnapshotCache{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	if redisAddr != "" {
		sc.rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
	}
	return sc
}

// Start begins the background refresh loop.
func (sc *SnapshotCache) Start() {
	go sc.loop()
}

// Stop halts the refresh loop.
func (sc *SnapshotCache) Stop() {
	close(sc.stop)
	<-sc.done
	if sc.rdb != nil {
		sc.rdb.Close()
	}
}

// Get returns the current cached snapshot.
func (sc *SnapshotCache) Get() Snapshot {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.snapshot
}

// JSON returns the snapshot as JSON bytes.
func (sc *SnapshotCache) JSON() ([]byte, error) {
	return json.Marshal(sc.Get())
}

func (sc *SnapshotCache) loop() {
	defer close(sc.done)

	sc.refresh()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sc.refresh()
		case <-sc.stop:
			return
		}
	}
}

func (sc *SnapshotCache) refresh() {
	if sc.rdb == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	totalRuns, _ := sc.rdb.Get(ctx, "grow:runs:total").Int64()
	totalElements, _ := sc.rdb.Get(ctx, "grow:elements:total").Int64()

	today := time.Now().UTC().Format("2006-01-02")
	runsToday, _ := sc.rdb.Get(ctx, "grow:runs:daily:"+today).Int64()

	sc.mu.Lock()
	sc.snapshot = Snapshot{
		TotalRuns:         totalRuns,
		RunsToday:         runsToday,
		ElementsAllocated: totalElements,
		CachedAt:          time.Now().UTC().Format(time.RFC3339),
	}
	sc.mu.Unlock()
}
