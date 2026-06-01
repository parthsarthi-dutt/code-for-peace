package apiextensions

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/parthsarthi-dutt/online-judge/server/auth"
	"github.com/parthsarthi-dutt/online-judge/server/database/postgres"
	"github.com/parthsarthi-dutt/online-judge/server/repository"
)

type UnlockRequest struct {
	ProblemID string `json:"problem_id"`
}

func UnlockEditorialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDInt := r.Context().Value(auth.UserIDKey).(int)
	userIDStr := strconv.Itoa(userIDInt)

	var req UnlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check if already unlocked
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM unlocked_editorials WHERE user_id=$1 AND problem_id=$2)`
	postgres.DB.QueryRow(context.Background(), checkQuery, userIDInt, req.ProblemID).Scan(&exists)
	if exists {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Already unlocked"})
		return
	}

	// Check tokens
	user, err := repository.GetUserByID(userIDStr)
	if err != nil || user.Tokens < 10 {
		http.Error(w, "Insufficient tokens", http.StatusPaymentRequired)
		return
	}

	// Deduct tokens and unlock
	err = repository.UpdateTokens(userIDStr, -10)
	if err != nil {
		http.Error(w, "Failed to deduct tokens", http.StatusInternalServerError)
		return
	}

	insertQuery := `INSERT INTO unlocked_editorials (user_id, problem_id) VALUES ($1, $2)`
	postgres.DB.Exec(context.Background(), insertQuery, userIDInt, req.ProblemID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Unlocked successfully"})
}

func GetEditorialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	problemID := r.URL.Query().Get("id")
	if problemID == "" {
		http.Error(w, "Missing problem ID", http.StatusBadRequest)
		return
	}

	userIDInt := r.Context().Value(auth.UserIDKey).(int)
	userIDStr := strconv.Itoa(userIDInt)

	// Check if user has solved it
	isSolved, _ := repository.HasUserSolvedProblem(userIDStr, problemID, "")
	
	// Check if user has unlocked it
	var isUnlocked bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM unlocked_editorials WHERE user_id=$1 AND problem_id=$2)`
	postgres.DB.QueryRow(context.Background(), checkQuery, userIDInt, problemID).Scan(&isUnlocked)

	if !isSolved && !isUnlocked {
		http.Error(w, "Editorial locked", http.StatusForbidden)
		return
	}

	// Read editorial.md
	content, err := os.ReadFile("problems/" + problemID + "/editorial.md")
	if err != nil {
		// Fallback to README.md or editorial.cpp? Wait, user asked for editorial.md
		http.Error(w, "Editorial not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": string(content)})
}
