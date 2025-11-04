// File: internal/middleware/logging.go
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logging is a middleware that logs the start and end of each request,
// along with its duration and other useful information.
func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the incoming request
		logger.Info("request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		// Call the next handler in the chain
		next.ServeHTTP(w, r)

		// Log the completed request
		logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}