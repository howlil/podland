package handler

import (
	"net/http"
	"time"

	"github.com/podland/backend/pkg/response"
)

// HandleHealth returns the health status of the service
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	pkgresponse.Success(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "podland-backend",
		"version":   "0.1.0",
	})
}
