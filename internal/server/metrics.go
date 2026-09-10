package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "grow_active_connections",
		Help: "Number of active WebSocket connections.",
	})

	ExecutionsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "grow_executions_in_flight",
		Help: "Number of executions currently running.",
	})

	ExecutionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grow_execution_duration_seconds",
		Help:    "Execution duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"handler"})

	EventsEmitted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "grow_events_emitted_total",
		Help: "Total events emitted to clients.",
	}, []string{"handler"})

	EventsCoalesced = promauto.NewCounter(prometheus.CounterOpts{
		Name: "grow_events_coalesced_total",
		Help: "Total events collapsed by the coalescing policy.",
	})
)
