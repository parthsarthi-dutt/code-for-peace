package api

import (
	"encoding/json"
	"net/http"

	"github.com/parthsarthi-dutt/online-judge/server/database"
)
func GetAllSubmissionsEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	result, err := database.GetAllSubmissions(id)
	if err != nil {
		http.Error(w, "Failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}