// File: internal/middleware/metrics.go
package middleware

import (
	"go-caching-proxy/internal/metrics"
	"net/http"
	"time"
)

// Metrics records Prometheus metrics for each request.
func Metrics(m *metrics.Metrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start).Seconds()
		m.Latency.Observe(duration)
	})
}
