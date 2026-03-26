package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/podland/backend/internal/database"
)

// HandleGetActivity returns the user's activity log
func HandleGetActivity(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	dbWrapper := database.NewDB(db)
	activities, err := dbWrapper.GetUserActivity(userID, 50)
	if err != nil {
		http.Error(w, "Failed to fetch activity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}
