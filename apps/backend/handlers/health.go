package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HandleHealth returns the health status of the service
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "podland-backend",
		"version":   "0.1.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
