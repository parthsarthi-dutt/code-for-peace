package models


type Submission struct {
	SubmissionID    string    `json:"submission_id"`
	ProblemID    string    `json:"problem_id"`
	UserID       string    `json:"user_id"`
	Language     string    `json:"language"`
	Code         string    `json:"code"`
	TimeLimit    int64       `json:"time_limit"`
	MemoryLimit  int64       `json:"memory_limit"`
	CreatedAt    string `json:"created_at"`
	Verdict string `json:"verdict"`
	ExecutionTime int64 `json:"execution_time"`
	MemoryUsed int64 `json:"memory_used"`
	Message string `json:"message"`
	Priority string `json:"priority"`
}

