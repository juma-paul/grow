package stats

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
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
	pgConn   string
	mu       sync.RWMutex
	snapshot Snapshot
	stop     chan struct{}
	done     chan struct{}
}

// NewSnapshotCache creates a cache that refreshes from Redis + Postgres
// every interval. If redisAddr is empty, returns a cache with zero values.
func NewSnapshotCache(redisAddr, pgConnStr string, interval time.Duration) *SnapshotCache {
	sc := &SnapshotCache{
		pgConn: pgConnStr,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Durable totals from Postgres (flusher moves Redis→PG)
	var pgRuns, pgElements int64
	if sc.pgConn != "" {
		if conn, err := pgx.Connect(ctx, sc.pgConn); err == nil {
			conn.QueryRow(ctx,
				"SELECT total_runs, total_elements FROM usage_stats WHERE id = 1",
			).Scan(&pgRuns, &pgElements)
			conn.Close(ctx)
		}
	}

	// Not-yet-flushed delta still in Redis
	var redisRuns, redisElements, runsToday int64
	if sc.rdb != nil {
		redisRuns, _ = sc.rdb.Get(ctx, "grow:runs:total").Int64()
		redisElements, _ = sc.rdb.Get(ctx, "grow:elements:total").Int64()
		today := time.Now().UTC().Format("2006-01-02")
		runsToday, _ = sc.rdb.Get(ctx, "grow:runs:daily:"+today).Int64()
	}

	sc.mu.Lock()
	sc.snapshot = Snapshot{
		TotalRuns:         pgRuns + redisRuns,
		RunsToday:         runsToday,
		ElementsAllocated: pgElements + redisElements,
		CachedAt:          time.Now().UTC().Format(time.RFC3339),
	}
	sc.mu.Unlock()
}
