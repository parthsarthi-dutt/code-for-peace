package apiextensions

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ProblemMeta struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Difficulty  string   `json:"difficulty"`
	Tags        []string `json:"tags"`
	TimeLimit   float64  `json:"time_limit"`
	MemoryLimit float64  `json:"memory_limit"`
}

type ProblemDetail struct {
	ProblemMeta
	Statement    string       `json:"statement"`
	SampleCases  []SampleCase `json:"sample_cases"`
}

type SampleCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

const problemsDir = "problems"

func ListProblemsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(problemsDir)
	if err != nil {
		slog.Error("Failed to read problems directory", slog.String("error", err.Error()))
		http.Error(w, "Failed to read problems", http.StatusInternalServerError)
		return
	}

	var problems []ProblemMeta

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metaPath := filepath.Join(problemsDir, entry.Name(), "metadata.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			slog.Warn("Skipping problem (no metadata)", slog.String("directory", entry.Name()))
			continue
		}

		var meta ProblemMeta
		err = json.Unmarshal(data, &meta)
		if err != nil {
			slog.Warn("Invalid metadata", slog.String("directory", entry.Name()), slog.String("error", err.Error()))
			continue
		}

		meta.ID = entry.Name() // Enforce unique ID from directory name
		problems = append(problems, meta)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problems)
}

func GetProblemHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing problem id", http.StatusBadRequest)
		return
	}

	problemDir := filepath.Join(problemsDir, id)

	// read metadata
	metaPath := filepath.Join(problemDir, "metadata.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		http.Error(w, "Problem not found", http.StatusNotFound)
		return
	}

	var meta ProblemMeta
	json.Unmarshal(metaData, &meta)

	// read statement (prefer README.md, fallback to statement.txt)
	stmtPath := filepath.Join(problemDir, "README.md")
	stmtData, err := os.ReadFile(stmtPath)
	if err != nil {
		stmtPath = filepath.Join(problemDir, "statement.txt")
		stmtData, err = os.ReadFile(stmtPath)
	}
	
	statement := ""
	if err == nil {
		statement = string(stmtData)
	}

	// read sample cases
	var samples []SampleCase
	inputDir := filepath.Join(problemDir, "input")
	outputDir := filepath.Join(problemDir, "output")

	inputEntries, err := os.ReadDir(inputDir)
	if err == nil {
		for _, ie := range inputEntries {
			if ie.IsDir() {
				continue
			}

			inputData, err := os.ReadFile(filepath.Join(inputDir, ie.Name()))
			if err != nil {
				continue
			}

			// match output file with same name
			outName := ie.Name()
			ext := filepath.Ext(outName)
			baseName := strings.TrimSuffix(outName, ext)
			_ = baseName

			outputData, err := os.ReadFile(filepath.Join(outputDir, ie.Name()))
			if err != nil {
				continue
			}

			samples = append(samples, SampleCase{
				Input:  string(inputData),
				Output: string(outputData),
			})
		}
	}

	detail := ProblemDetail{
		ProblemMeta: meta,
		Statement:   statement,
		SampleCases: samples,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}
