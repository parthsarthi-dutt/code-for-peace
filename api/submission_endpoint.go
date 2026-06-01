package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/parthsarthi-dutt/online-judge/server/auth"
	"github.com/parthsarthi-dutt/online-judge/server/database"
	"github.com/parthsarthi-dutt/online-judge/server/models"
)

func SubmissionEndpointContest(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// extract user from JWT context
	userID := r.Context().Value(auth.UsernameKey).(string)

	allowed, errRate := database.CheckRateLimit("ratelimit:submit_code:"+userID, 1, 15)
	if errRate != nil {
		http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Please wait 15 seconds between submissions", http.StatusTooManyRequests)
		return
	}

	submissionID := r.FormValue("submission_id")
	problemID := r.FormValue("problem_id")
	language := r.FormValue("language")
	code := r.FormValue("code")

	fmt.Println("submissionID:", submissionID)
	fmt.Println("problemID:", problemID)
	fmt.Println("userID:", userID)
	fmt.Println("language:", language)
	fmt.Println("code length:", len(code))

	if submissionID == "" || problemID == "" || language == "" || code == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	database.UserSubmission(
		submissionID,
		problemID,
		code,
		userID,
		language,
		"contest",
	)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Submission received"))
}
func SubmissionEndpointPractice(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var submission models.Submission

	err := json.NewDecoder(r.Body).Decode(&submission)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	allowed, errRate := database.CheckRateLimit("ratelimit:submit_code:"+submission.UserID, 1, 15)
	if errRate != nil {
		http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Please wait 15 seconds between submissions", http.StatusTooManyRequests)
		return
	}
	
	database.UserSubmission(
		submission.SubmissionID,
		submission.ProblemID,
		submission.Code,
		submission.UserID,
		submission.Language,
		"practice",
	)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Submission received"))
}