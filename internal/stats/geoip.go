package stats

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// GeoResolver looks up the country for a request's IP address.
// If the database file is missing, all lookups return "".
type GeoResolver struct {
	db *geoip2.Reader
}

// NewGeoResolver opens a MaxMind GeoLite2-Country database file.
// If path is empty or the file cannot be opened, returns a resolver
// that always returns "" (no country).
func NewGeoResolver(path string) *GeoResolver {
	if path == "" {
		return &GeoResolver{}
	}
	db, err := geoip2.Open(path)
	if err != nil {
		slog.Warn("geoip database not loaded", "path", path, "error", err)
		return &GeoResolver{}
	}
	slog.Info("geoip database loaded", "path", path)
	return &GeoResolver{db: db}
}

// Country returns the ISO country code for the request's IP, or "".
func (g *GeoResolver) Country(r *http.Request) string {
	if g.db == nil {
		return ""
	}
	ipStr := extractIP(r)
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}
	record, err := g.db.Country(ip)
	if err != nil {
		return ""
	}
	return record.Country.IsoCode
}

// Close releases the database resources.
func (g *GeoResolver) Close() error {
	if g.db == nil {
		return nil
	}
	return g.db.Close()
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}
