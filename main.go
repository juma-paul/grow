package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/juma-paul/grow/internal/server"
	"github.com/juma-paul/grow/internal/stats"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	redisAddr := os.Getenv("REDIS_ADDR")
	server.Stats = stats.NewRecorder(redisAddr)
	defer server.Stats.Close()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/execute", server.HandleExecute)
	http.HandleFunc("/auto", server.HandleAutoExecute)
	http.HandleFunc("/observe", server.HandleObserve)

	http.Handle("/", server.FrontendHandler())

	slog.Info("server starting", "addr", "http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
