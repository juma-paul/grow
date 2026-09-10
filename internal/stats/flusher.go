package stats

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

// Flusher periodically reads counters from Redis and writes durable
// aggregates to Postgres. The flush is idempotent: it uses GETSET to
// atomically read-and-reset Redis counters, so re-running a flush
// that was interrupted picks up only the delta since the last
// successful flush.
type Flusher struct {
	rdb      *redis.Client
	pgConn   string
	interval time.Duration
	stop     chan struct{}
	done     chan struct{}
}

// NewFlusher creates a Flusher. If either addr is empty, returns nil.
func NewFlusher(redisAddr, pgConnStr string, interval time.Duration) *Flusher {
	if redisAddr == "" || pgConnStr == "" {
		return nil
	}
	return &Flusher{
		rdb:      redis.NewClient(&redis.Options{Addr: redisAddr}),
		pgConn:   pgConnStr,
		interval: interval,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Start runs the flush loop in a goroutine.
func (f *Flusher) Start() {
	if f == nil {
		return
	}
	go f.loop()
}

// Stop signals the flush loop to exit and waits for it to finish.
func (f *Flusher) Stop() {
	if f == nil {
		return
	}
	close(f.stop)
	<-f.done
	f.rdb.Close()
}

func (f *Flusher) loop() {
	defer close(f.done)
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := f.flush(); err != nil {
				slog.Error("stats flush failed", "error", err)
			}
		case <-f.stop:
			f.flush()
			return
		}
	}
}

// flush reads counters from Redis (atomically resetting them) and
// upserts the values into Postgres.
func (f *Flusher) flush() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	totalRuns, err := f.getAndReset(ctx, "grow:runs:total")
	if err != nil {
		return fmt.Errorf("read total runs: %w", err)
	}
	totalElements, err := f.getAndReset(ctx, "grow:elements:total")
	if err != nil {
		return fmt.Errorf("read total elements: %w", err)
	}

	if totalRuns == 0 && totalElements == 0 {
		return nil
	}

	conn, err := pgx.Connect(ctx, f.pgConn)
	if err != nil {
		return fmt.Errorf("pg connect: %w", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `
		INSERT INTO usage_stats (id, total_runs, total_elements)
		VALUES (1, $1, $2)
		ON CONFLICT (id) DO UPDATE SET
			total_runs = usage_stats.total_runs + EXCLUDED.total_runs,
			total_elements = usage_stats.total_elements + EXCLUDED.total_elements,
			updated_at = NOW()
	`, totalRuns, totalElements)
	if err != nil {
		return fmt.Errorf("pg upsert: %w", err)
	}

	slog.Info("stats flushed",
		"runs", totalRuns,
		"elements", totalElements,
	)
	return nil
}

func (f *Flusher) getAndReset(ctx context.Context, key string) (int64, error) {
	val, err := f.rdb.GetDel(ctx, key).Int64()
	if err != nil {
		if strings.Contains(err.Error(), "nil") || err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return val, nil
}
