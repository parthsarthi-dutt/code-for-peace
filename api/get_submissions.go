package api

import (
	"encoding/json"
	"net/http"

	"github.com/parthsarthi-dutt/online-judge/server/database"
)

func GetSubmissionsEndpoint(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")

	result,err := database.GetSubmission(id)
	if(err!=nil){
		http.Error(w, "Submission does not exist", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}