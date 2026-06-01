package apiextensions

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/parthsarthi-dutt/online-judge/server/auth"
	"github.com/parthsarthi-dutt/online-judge/server/database/postgres"
	"github.com/parthsarthi-dutt/online-judge/server/proto/evaluation"
	"github.com/parthsarthi-dutt/online-judge/server/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ─── Request / Response types ────────────────────────────────────────

type StartInterviewPayload struct {
	Level    string `json:"level"`
	Duration int    `json:"duration"`
}

type InterviewResponsePayload struct {
	InterviewID  int    `json:"interview_id"`
	AudioBase64  string `json:"audio_base64"`
	TimeUp       bool   `json:"time_up"`
	SystemAction string `json:"system_action"`
	Code         string `json:"code"`
}

type InterviewHistoryEntry struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type InterviewListItem struct {
	ID             int    `json:"id"`
	Level          string `json:"level"`
	Duration       int    `json:"duration"`
	TokensDeducted int    `json:"tokens_deducted"`
	StartedAt      string `json:"started_at"`
	Status         string `json:"status"`
}

// ─── Helpers ─────────────────────────────────────────────────────────

func getInterviewAIClient() (evaluation.EvaluationServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(50*1024*1024),
			grpc.MaxCallSendMsgSize(50*1024*1024),
		),
	)
	if err != nil {
		return nil, nil, err
	}
	return evaluation.NewEvaluationServiceClient(conn), conn, nil
}

func getTokenCost(duration int) int {
	switch duration {
	case 5:
		return 25
	case 10:
		return 40
	case 15:
		return 50
	case 30:
		return 70
	default:
		return -1
	}
}

// ─── Handlers ────────────────────────────────────────────────────────

// StartInterviewHandler starts a new AI interview session.
func StartInterviewHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	userIDStr := strconv.Itoa(userID)

	var payload StartInterviewPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if payload.Level != "easy" && payload.Level != "medium" && payload.Level != "hard" {
		http.Error(w, "Invalid level. Must be easy, medium, or hard", http.StatusBadRequest)
		return
	}
	if payload.Duration != 5 && payload.Duration != 10 && payload.Duration != 15 && payload.Duration != 30 {
		http.Error(w, "Invalid duration. Must be 5, 10, 15, or 30", http.StatusBadRequest)
		return
	}

	// Check tokens
	user, err := repository.GetUserByID(userIDStr)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	cost := getTokenCost(payload.Duration)
	if cost == -1 {
		http.Error(w, "Invalid duration cost", http.StatusBadRequest)
		return
	}
	if user.Tokens < cost {
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":    "Insufficient tokens",
			"required": cost,
			"current":  user.Tokens,
		})
		return
	}

	// Deduct tokens
	if err := repository.UpdateTokens(userIDStr, -cost); err != nil {
		http.Error(w, "Failed to deduct tokens", http.StatusInternalServerError)
		return
	}

	// Call AI service
	client, conn, err := getInterviewAIClient()
	if err != nil {
		// Refund tokens on failure
		repository.UpdateTokens(userIDStr, cost)
		slog.Error("Failed to connect to AI service", slog.String("error", err.Error()))
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	resp, err := client.StartInterviewSession(context.Background(), &evaluation.StartInterviewRequest{
		Level:    payload.Level,
		Duration: int32(payload.Duration),
	})
	if err != nil {
		repository.UpdateTokens(userIDStr, cost)
		slog.Error("AI interview start failed", slog.String("error", err.Error()))
		http.Error(w, "Failed to start interview", http.StatusInternalServerError)
		return
	}

	// Save to DB
	history := []InterviewHistoryEntry{
		{Role: "interviewer", Text: resp.QuestionText},
	}
	historyJSON, _ := json.Marshal(history)

	var interviewID int
	err = postgres.DB.QueryRow(
		context.Background(),
		`INSERT INTO ai_interviews (user_id, level, duration, tokens_deducted, status, history)
		 VALUES ($1, $2, $3, $4, 'active', $5::jsonb)
		 RETURNING id`,
		userID, payload.Level, payload.Duration, cost, string(historyJSON),
	).Scan(&interviewID)

	if err != nil {
		slog.Error("Failed to save interview", slog.String("error", err.Error()))
		http.Error(w, "Failed to save interview session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"interview_id":  interviewID,
		"question_text": resp.QuestionText,
		"audio_base64":  base64.StdEncoding.EncodeToString(resp.AudioBytes),
		"tokens_spent":  cost,
	})
}

// ProcessInterviewResponseHandler processes user audio and returns next question.
func ProcessInterviewResponseHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)

	var payload InterviewResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get interview from DB
	var level string
	var duration int
	var historyJSON string
	var status string
	var dbUserID int

	err := postgres.DB.QueryRow(
		context.Background(),
		`SELECT user_id, level, duration, history::text, status FROM ai_interviews WHERE id = $1`,
		payload.InterviewID,
	).Scan(&dbUserID, &level, &duration, &historyJSON, &status)

	if err != nil {
		http.Error(w, "Interview not found", http.StatusNotFound)
		return
	}
	if dbUserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}
	if status != "active" {
		http.Error(w, "Interview already completed", http.StatusBadRequest)
		return
	}

	// Decode audio (unless it's just a system action)
	var audioBytes []byte
	var errDecode error
	if payload.AudioBase64 != "" {
		audioBytes, errDecode = base64.StdEncoding.DecodeString(payload.AudioBase64)
		if errDecode != nil {
			http.Error(w, "Invalid audio data", http.StatusBadRequest)
			return
		}
	}

	// Call AI service
	client, conn, err := getInterviewAIClient()
	if err != nil {
		slog.Error("Failed to connect to AI service", slog.String("error", err.Error()))
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	resp, err := client.ProcessInterviewResponse(context.Background(), &evaluation.InterviewResponseRequest{
		Level:           level,
		Duration:        int32(duration),
		ChatHistoryJson: historyJSON,
		AudioBytes:      audioBytes,
		TimeUp:          payload.TimeUp,
		SystemAction:    payload.SystemAction,
		Code:            payload.Code,
	})
	if err != nil {
		slog.Error("AI interview processing failed", slog.String("error", err.Error()))
		http.Error(w, "Failed to process response", http.StatusInternalServerError)
		return
	}

	// Update history in DB
	var history []InterviewHistoryEntry
	json.Unmarshal([]byte(historyJSON), &history)

	if resp.UserTranscript != "" {
		history = append(history, InterviewHistoryEntry{Role: "candidate", Text: resp.UserTranscript})
	}
	history = append(history, InterviewHistoryEntry{Role: "interviewer", Text: resp.QuestionText})

	updatedJSON, _ := json.Marshal(history)
	postgres.DB.Exec(
		context.Background(),
		`UPDATE ai_interviews SET history = $1::jsonb WHERE id = $2`,
		string(updatedJSON), payload.InterviewID,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"question_text":   resp.QuestionText,
		"audio_base64":    base64.StdEncoding.EncodeToString(resp.AudioBytes),
		"user_transcript": resp.UserTranscript,
	})
}

// EndInterviewHandler marks an interview as completed.
func EndInterviewHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)

	var payload struct {
		InterviewID int `json:"interview_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var dbUserID int
	var historyJSON string
	var level string

	err := postgres.DB.QueryRow(
		context.Background(),
		`SELECT user_id, history::text, level FROM ai_interviews WHERE id = $1`,
		payload.InterviewID,
	).Scan(&dbUserID, &historyJSON, &level)

	if err != nil {
		http.Error(w, "Interview not found", http.StatusNotFound)
		return
	}
	if dbUserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Generate summary feedback via LLM through gRPC
	// For simplicity, we generate it inline here
	var history []InterviewHistoryEntry
	json.Unmarshal([]byte(historyJSON), &history)

	feedback := generateInterviewFeedback(level, history)

	postgres.DB.Exec(
		context.Background(),
		`UPDATE ai_interviews SET status = 'completed', feedback = $1 WHERE id = $2`,
		feedback, payload.InterviewID,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "completed",
		"feedback": feedback,
	})
}

// GetInterviewsHandler returns all interviews for the authenticated user.
func GetInterviewsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)

	rows, err := postgres.DB.Query(
		context.Background(),
		`SELECT id, level, duration, tokens_deducted, started_at, status
		 FROM ai_interviews WHERE user_id = $1 ORDER BY started_at DESC`,
		userID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch interviews", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var interviews []InterviewListItem
	for rows.Next() {
		var item InterviewListItem
		var startedAt interface{}
		err := rows.Scan(&item.ID, &item.Level, &item.Duration, &item.TokensDeducted, &startedAt, &item.Status)
		if err != nil {
			continue
		}
		if t, ok := startedAt.(interface{ String() string }); ok {
			item.StartedAt = t.String()
		}
		interviews = append(interviews, item)
	}

	if interviews == nil {
		interviews = []InterviewListItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(interviews)
}

// GetInterviewDetailHandler returns full interview details including history.
func GetInterviewDetailHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	interviewIDStr := r.URL.Query().Get("id")
	if interviewIDStr == "" {
		http.Error(w, "Missing interview id", http.StatusBadRequest)
		return
	}

	interviewID, err := strconv.Atoi(interviewIDStr)
	if err != nil {
		http.Error(w, "Invalid interview id", http.StatusBadRequest)
		return
	}

	var dbUserID int
	var level string
	var duration int
	var tokensDeducted int
	var status string
	var historyJSON string
	var feedback string

	err = postgres.DB.QueryRow(
		context.Background(),
		`SELECT user_id, level, duration, tokens_deducted, status, history::text, feedback
		 FROM ai_interviews WHERE id = $1`,
		interviewID,
	).Scan(&dbUserID, &level, &duration, &tokensDeducted, &status, &historyJSON, &feedback)

	if err != nil {
		http.Error(w, "Interview not found", http.StatusNotFound)
		return
	}
	if dbUserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	var history []InterviewHistoryEntry
	json.Unmarshal([]byte(historyJSON), &history)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":              interviewID,
		"level":           level,
		"duration":        duration,
		"tokens_deducted": tokensDeducted,
		"status":          status,
		"history":         history,
		"feedback":        feedback,
	})
}

// ─── Internal: Generate feedback summary via LLM ──────────────────────

func generateInterviewFeedback(level string, history []InterviewHistoryEntry) string {
	conversation := ""
	for _, entry := range history {
		if entry.Role == "interviewer" {
			conversation += "Interviewer: " + entry.Text + "\n"
		} else {
			conversation += "Candidate: " + entry.Text + "\n"
		}
	}

	prompt := `You are evaluating a coding interview transcript.
The interview was at the "` + level + `" difficulty level.

Transcript:
` + conversation + `

Provide a concise performance summary:
1. Overall rating out of 10
2. Strengths (2-3 bullet points)
3. Areas for improvement (2-3 bullet points)
4. Specific topics the candidate should study

**CRITICAL SCORING RULE:**
If the transcript shows that the candidate exited early, did not provide any substantive technical answers, or merely completed the introduction without solving any problems, you MUST give an overall rating of 0/10. Do not invent strengths if none were demonstrated. Be extremely strict about giving a 0 if the interview was essentially empty.

Format your response using Markdown. Use appropriate headings (e.g. ## Performance Summary, ### Strengths), bold text for emphasis, and bullet points. Keep it professional and constructive.`

	// Call the LLM directly via HTTP to Groq API
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "Unable to generate feedback at this time (Missing configuration)."
	}

	payload := map[string]interface{}{
		"model": "llama-3.1-8b-instant",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
	}
	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "Unable to generate feedback at this time."
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "Unable to generate feedback at this time."
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	if choices, ok := res["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content
				}
			}
		}
	}

	return "Unable to parse feedback response."
}
