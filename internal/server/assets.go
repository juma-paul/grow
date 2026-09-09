package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed all:dist
var frontendFiles embed.FS

// FrontendHandler returns an http.Handler that serves the embedded
// frontend files, falling back to index.html for client-side routing.
func FrontendHandler() http.Handler {
	dist, err := fs.Sub(frontendFiles, "dist")
	if err != nil {
		log.Fatalf("embedded frontend missing: %v", err)
	}
	return http.FileServer(http.FS(dist))
}
