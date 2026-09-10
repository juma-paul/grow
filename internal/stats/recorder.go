package stats

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Recorder writes usage counters to Redis. All methods are
// fire-and-forget: a Redis failure never blocks or fails the
// caller's execution.
type Recorder struct {
	rdb *redis.Client
}

// NewRecorder creates a Recorder connected to the given Redis address.
// If addr is empty, recording is silently disabled (all methods no-op).
func NewRecorder(addr string) *Recorder {
	if addr == "" {
		return &Recorder{}
	}
	return &Recorder{
		rdb: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

// RecordExecution increments per-strategy and total run counters,
// plus the elements-allocated counter. Fire-and-forget.
func (r *Recorder) RecordExecution(strategy string, elements int) {
	if r.rdb == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	pipe := r.rdb.Pipeline()
	pipe.Incr(ctx, "grow:runs:total")
	pipe.Incr(ctx, fmt.Sprintf("grow:runs:strategy:%s", strategy))
	pipe.IncrBy(ctx, "grow:elements:total", int64(elements))

	today := time.Now().UTC().Format("2006-01-02")
	pipe.Incr(ctx, fmt.Sprintf("grow:runs:daily:%s", today))

	if _, err := pipe.Exec(ctx); err != nil {
		slog.Warn("stats record failed", "error", err)
	}
}

// RecordCountry increments the per-country counter. Fire-and-forget.
func (r *Recorder) RecordCountry(country string) {
	if r.rdb == nil || country == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := r.rdb.Incr(ctx, fmt.Sprintf("grow:countries:%s", country)).Err(); err != nil {
		slog.Warn("country record failed", "error", err)
	}
}

// Close shuts down the Redis connection.
func (r *Recorder) Close() error {
	if r.rdb == nil {
		return nil
	}
	return r.rdb.Close()
}
