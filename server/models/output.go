package models

type Output struct {
	Verdict string `json:"verdict"`
	ExecutionTime int64 `json:"execution_time"`
	MemoryUsed int64 `json:"memory_used"`
	Message string `json:"message"`
}