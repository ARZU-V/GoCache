// File: internal/key/key.go
package key

import (
	"fmt"
	"net/http"
)

// Generate creates a unique cache key for an HTTP request.
// A good key is essential for preventing cache collisions. We include the method,
// host, and the full URL (path + query) to ensure uniqueness.
func Generate(r *http.Request) string {
	return fmt.Sprintf("%s|%s|%s", r.Method, r.Host, r.URL.String())
}