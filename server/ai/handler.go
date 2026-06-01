package ai

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/parthsarthi-dutt/online-judge/server/auth"
	"github.com/parthsarthi-dutt/online-judge/server/database"
	"github.com/parthsarthi-dutt/online-judge/server/proto/evaluation"
	"github.com/parthsarthi-dutt/online-judge/server/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HintRequestPayload struct {
	ProblemID string `json:"problem_id"`
	UserCode  string `json:"user_code"`
}

func getAIClient() (evaluation.EvaluationServiceClient, *grpc.ClientConn, error) {
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		aiServiceURL = "localhost:50051"
	}
	conn, err := grpc.Dial(aiServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return evaluation.NewEvaluationServiceClient(conn), conn, nil
}

func GetHint(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	userIDStr := strconv.Itoa(userID)

	var payload HintRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := repository.GetUserByID(userIDStr)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	allowed, err := database.CheckRateLimit("ratelimit:ai_hint:"+userIDStr, 5, 86400)
	if err != nil {
		http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Daily limit reached for AI Hints (5/day)", http.StatusTooManyRequests)
		return
	}

	if user.Tokens < 5 {
		http.Error(w, "Insufficient tokens (5 required)", http.StatusForbidden)
		return
	}

	// Read problem statement and editorial code/explanation
	statementPath := "problems/" + payload.ProblemID + "/README.md"
	statement, err := os.ReadFile(statementPath)
	if err != nil {
		statementPath = "problems/" + payload.ProblemID + "/statement.txt"
		statement, err = os.ReadFile(statementPath)
		if err != nil {
			http.Error(w, "Problem statement not found", http.StatusInternalServerError)
			return
		}
	}

	editorialPath := "problems/" + payload.ProblemID + "/editorial.md"

	editorial, err := os.ReadFile(editorialPath)
	if err != nil {
		// Fallback to editorial.cpp if md is not present
		editorialPath = "problems/" + payload.ProblemID + "/editorial.cpp"
		editorial, err = os.ReadFile(editorialPath)
		if err != nil {
			editorial = []byte("Optimal code/explanation not provided.")
		}
	}

	client, conn, err := getAIClient()
	if err != nil {
		slog.Error("Failed to connect to AI service", slog.String("error", err.Error()))
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	req := &evaluation.HintRequest{
		ProblemStatement: string(statement),
		UserCode:         payload.UserCode,
		EditorialCode:    string(editorial),
	}

	resp, err := client.GenerateHint(context.Background(), req)
	if err != nil {
		slog.Error("Failed to generate hint", slog.String("error", err.Error()))
		http.Error(w, "Failed to generate hint", http.StatusInternalServerError)
		return
	}

	// Deduct 5 tokens
	err = repository.UpdateTokens(userIDStr, -5)
	if err != nil {
		slog.Error("Failed to deduct tokens", slog.String("error", err.Error()))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"hint": resp.Hint,
	})
}

type FeedbackRequestPayload struct {
	ProblemID string `json:"problem_id"`
	UserCode  string `json:"user_code"`
}

func GetFeedback(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	userIDStr := strconv.Itoa(userID)

	var payload FeedbackRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := repository.GetUserByID(userIDStr)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	allowed, err := database.CheckRateLimit("ratelimit:ai_feedback:"+userIDStr, 5, 86400)
	if err != nil {
		http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Daily limit reached for AI Feedback (5/day)", http.StatusTooManyRequests)
		return
	}

	if user.Tokens < 3 {
		http.Error(w, "Insufficient tokens (3 required)", http.StatusForbidden)
		return
	}

	// Read problem statement and editorial code/explanation
	statementPath := "problems/" + payload.ProblemID + "/README.md"
	statement, err := os.ReadFile(statementPath)
	if err != nil {
		statementPath = "problems/" + payload.ProblemID + "/statement.txt"
		statement, err = os.ReadFile(statementPath)
		if err != nil {
			http.Error(w, "Problem statement not found", http.StatusInternalServerError)
			return
		}
	}

	editorialPath := "problems/" + payload.ProblemID + "/editorial.md"

	editorial, err := os.ReadFile(editorialPath)
	if err != nil {
		// Fallback to editorial.cpp if md is not present
		editorialPath = "problems/" + payload.ProblemID + "/editorial.cpp"
		editorial, err = os.ReadFile(editorialPath)
		if err != nil {
			editorial = []byte("Optimal code/explanation not provided.")
		}
	}

	client, conn, err := getAIClient()
	if err != nil {
		slog.Error("Failed to connect to AI service", slog.String("error", err.Error()))
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	req := &evaluation.FeedbackRequest{
		ProblemStatement: string(statement),
		UserCode:         payload.UserCode,
		EditorialCode:    string(editorial),
	}

	resp, err := client.GenerateFeedback(context.Background(), req)
	if err != nil {
		slog.Error("Failed to generate feedback", slog.String("error", err.Error()))
		http.Error(w, "Failed to generate feedback", http.StatusInternalServerError)
		return
	}

	// Deduct 3 tokens
	err = repository.UpdateTokens(userIDStr, -3)
	if err != nil {
		slog.Error("Failed to deduct tokens", slog.String("error", err.Error()))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"feedback": resp.Feedback,
	})
}
