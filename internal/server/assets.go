package server

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
)

//go:embed all:dist
var frontendFiles embed.FS

// FrontendHandler returns an http.Handler that serves the embedded
// frontend files, falling back to index.html for client-side routing.
func FrontendHandler() http.Handler {
	dist, err := fs.Sub(frontendFiles, "dist")
	if err != nil {
		slog.Error("embedded frontend missing", "error", err)
		os.Exit(1)
	}
	return http.FileServer(http.FS(dist))
}
