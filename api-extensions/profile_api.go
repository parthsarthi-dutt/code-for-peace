package apiextensions

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/parthsarthi-dutt/online-judge/server/auth"
	"github.com/parthsarthi-dutt/online-judge/server/repository"
)

type ProfileResponse struct {
	UserID        int            `json:"user_id"`
	Username      string         `json:"username"`
	Tokens        int            `json:"tokens"`
	TotalSolved   int            `json:"total_solved"`
	EasySolved    int            `json:"easy_solved"`
	MediumSolved  int            `json:"medium_solved"`
	HardSolved    int            `json:"hard_solved"`
	CurrentStreak int            `json:"current_streak"`
	HighestStreak int            `json:"highest_streak"`
	HeatmapData   map[string]int `json:"heatmap_data"`
	RecentSolves  []RecentSolve  `json:"recent_solves"`
}

type RecentSolve struct {
	ProblemID     string `json:"problem_id"`
	SubmissionID  string `json:"submission_id"`
	SolvedAt      string `json:"solved_at"`
	Verdict       string `json:"verdict"`
	Code          string `json:"code"`
	ExecutionTime int64  `json:"execution_time"`
	MemoryUsed    int64  `json:"memory_used"`
}

func GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(auth.UserIDKey).(int)
	username := r.Context().Value(auth.UsernameKey).(string)

	submissions, err := repository.GetAllSubmissions(strconv.Itoa(userID))
	if err != nil {
		http.Error(w, "Failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	difficultyMap := loadDifficultyMap()

	solvedProblems := make(map[string]bool)
	heatmapData := make(map[string]int)
	var recentSolves []RecentSolve

	easySolved := 0
	mediumSolved := 0
	hardSolved := 0

	for _, s := range submissions {
		day := s.CreatedAt.Format("2006-01-02")
		heatmapData[day]++

		if s.Verdict == "Accepted" && !solvedProblems[s.ProblemID] {
			solvedProblems[s.ProblemID] = true

			diff := difficultyMap[s.ProblemID]
			switch diff {
			case "Easy":
				easySolved++
			case "Medium":
				mediumSolved++
			case "Hard":
				hardSolved++
			}

			recentSolves = append(recentSolves, RecentSolve{
				ProblemID:     s.ProblemID,
				SubmissionID:  s.SubmissionID,
				SolvedAt:      s.CreatedAt.Format("2006-01-02 15:04:05"),
				Verdict:       s.Verdict,
				Code:          s.Code,
				ExecutionTime: s.ExecutionTime,
				MemoryUsed:    s.MemoryUsed,
			})
		}
	}

	if len(recentSolves) > 20 {
		recentSolves = recentSolves[len(recentSolves)-20:]
	}

	userModel, err := repository.GetUserByID(strconv.Itoa(userID))
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	profile := ProfileResponse{
		UserID:        userID,
		Username:      username,
		Tokens:        userModel.Tokens,
		TotalSolved:   len(solvedProblems),
		EasySolved:    easySolved,
		MediumSolved:  mediumSolved,
		HardSolved:    hardSolved,
		CurrentStreak: userModel.CurrentStreak,
		HighestStreak: userModel.HighestStreak,
		HeatmapData:   heatmapData,
		RecentSolves:  recentSolves,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func loadDifficultyMap() map[string]string {
	dm := make(map[string]string)

	entries, err := os.ReadDir(problemsDir)
	if err != nil {
		return dm
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metaPath := filepath.Join(problemsDir, entry.Name(), "metadata.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		var meta ProblemMeta
		json.Unmarshal(data, &meta)
		dm[meta.ID] = meta.Difficulty
	}

	return dm
}

type UpdateProfileRequest struct {
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

func UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(auth.UserIDKey).(int)

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username cannot be empty", http.StatusBadRequest)
		return
	}

	err := repository.UpdateUserProfile(strconv.Itoa(userID), req.Username, req.AvatarURL)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}
