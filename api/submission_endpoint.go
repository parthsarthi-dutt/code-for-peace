package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/parthsarthi-dutt/online-judge/server/database"
	"github.com/parthsarthi-dutt/online-judge/server/models"
)

func SubmissionEndpointContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	submissionID := r.FormValue("submission_id")
	problemID := r.FormValue("problem_id")
	userID := r.FormValue("user_id")
	language := r.FormValue("language")
	code := r.FormValue("code")

	// Debug print (very useful right now)
	fmt.Println("submissionID:", submissionID)
	fmt.Println("problemID:", problemID)
	fmt.Println("userID:", userID)
	fmt.Println("language:", language)
	fmt.Println("code length:", len(code))

	if submissionID == "" || problemID == "" || userID == "" || language == "" || code == "" {
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