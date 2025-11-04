// File: internal/admin/admin.go
package admin

import (
	"encoding/json"
	"net/http"
)

// HealthzHandler returns a simple JSON response indicating the service is healthy.
// In a real application, this could be expanded to check database connections or other dependencies.
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
