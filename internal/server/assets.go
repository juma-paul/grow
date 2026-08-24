package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var frontendFiles embed.FS

// FrontendHandler returns an http.Handler that serves the embedded
// frontend files, falling back to index.html for client-side routing.
func FrontendHandler() http.Handler {
	dist, _ := fs.Sub(frontendFiles, "dist")
	return http.FileServer(http.FS(dist))
}
