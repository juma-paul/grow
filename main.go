package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/juma-paul/grow/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})

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
