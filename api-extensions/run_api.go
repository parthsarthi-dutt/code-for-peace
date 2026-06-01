package apiextensions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strconv"

	"github.com/parthsarthi-dutt/online-judge/server/auth"
	"github.com/parthsarthi-dutt/online-judge/server/database"
	"github.com/parthsarthi-dutt/online-judge/server/judge"
)

type RunRequest struct {
	Language  string `json:"language"`
	Code      string `json:"code"`
	InputData string `json:"input_data"`
	TimeLimit float64  `json:"time_limit"`
}

type RunResponse struct {
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	IsTLE   bool   `json:"is_tle"`
}

func RunCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(auth.UserIDKey).(int)
	userIDStr := strconv.Itoa(userID)

	allowed, err := database.CheckRateLimit("ratelimit:run_code:"+userIDStr, 1, 10)
	if err != nil {
		http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Please wait 10 seconds between runs", http.StatusTooManyRequests)
		return
	}

	langConfig, exists := judge.SupportedLanguages[req.Language]
	if !exists {
		http.Error(w, "Unsupported language", http.StatusBadRequest)
		return
	}

	// Create a unique temporary submission folder
	runID := fmt.Sprintf("run_%d", time.Now().UnixNano())
	basePath := "temp_runs/"

	// Write code to disk
	judge.CreateAndWriteFile(req.Code, basePath, runID, langConfig.FileExtension)
	defer judge.DeleteFile(basePath, runID)

	submissionPath := basePath + runID

	// Compile if needed
	if langConfig.NeedsCompilation {
		success, compilerErr := judge.CompileInSandbox(submissionPath, langConfig)
		if !success {
			json.NewEncoder(w).Encode(RunResponse{
				Stdout: "",
				Stderr: "Compilation Error:\n" + compilerErr,
				IsTLE:  false,
			})
			return
		}
	}

	// Default time limit to 2 seconds if not provided
	tl := req.TimeLimit
	if tl <= 0 {
		tl = 2
	}

	// Execute
	stdout, runtimeErr, isTLE := judge.ExecuteInSandbox(submissionPath, req.InputData, int64(tl), langConfig)

	resp := RunResponse{
		Stdout:  stdout,
		Stderr:  runtimeErr,
		IsTLE:   isTLE,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
