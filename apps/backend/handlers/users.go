package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/podland/backend/internal/database"
)

// HandleGetMe returns the current authenticated user
func HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value("user_id")
	if userIDRaw == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbWrapper := database.NewDB(db)
	user, err := dbWrapper.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleGetUser returns a user by ID
func HandleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	dbWrapper := database.NewDB(db)
	user, err := dbWrapper.GetUserByID(id.String())
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleConfirmNIM confirms or updates the user's NIM
func HandleConfirmNIM(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value("user_id")
	if userIDRaw == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		NIM string `json:"nim"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NIM == "" {
		http.Error(w, "NIM is required", http.StatusBadRequest)
		return
	}

	dbWrapper := database.NewDB(db)

	// Update NIM (this also recalculates role based on NIM)
	if err := dbWrapper.UpdateUserNIM(userID, req.NIM); err != nil {
		http.Error(w, "Failed to update NIM", http.StatusInternalServerError)
		return
	}

	// Log activity
	_ = dbWrapper.CreateActivityLog(userID, "nim_confirmed", map[string]interface{}{
		"nim": req.NIM,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
