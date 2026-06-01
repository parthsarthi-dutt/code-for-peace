package judge

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func CompileInSandbox(submissionPath string, lang LangConfig) (bool, string) {
	absPath, err := filepath.Abs(submissionPath)
	if err != nil {
		return false, "Failed to resolve path: " + err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	compileCmd := lang.CompileCmd
	if lang.FileExtension == ".cpp" {
		tmpBinary := fmt.Sprintf("/sandbox/main_%d", time.Now().UnixNano())
		compileCmd = fmt.Sprintf("g++ /sandbox/code%s -o %s -O2 -std=c++17 -Wall && mv %s /sandbox/main",
			lang.FileExtension, tmpBinary, tmpBinary)
	}

	cmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",
		"--network=none",
		"--memory=256m",
		"--cpus=1",
		"--user=1001:1001",
		"--read-only",
		"--tmpfs=/tmp:size=64m",
		"-v", absPath+":/sandbox:rw",
		lang.Image,
		"sh", "-c", compileCmd,
	)

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return false, "Compilation Timed Out"
	}
	if err != nil {
		slog.Error("Compilation Error", slog.String("output", string(out)))
		cleanOut := strings.ReplaceAll(string(out), "/sandbox/", "")
		return false, cleanOut
	}

	slog.Info("Compilation successful")
	return true, ""
}

// ExecuteInSandbox runs the compiled binary (or interpreted script) inside
// a Docker container with the given input piped to stdin.
//
// Returns (stdout, errorMessage, isTimeLimitExceeded).
//
// Key security: the volume is mounted as READ-ONLY (:ro) during execution
// so the running program cannot modify its own source or binary.
func ExecuteInSandbox(submissionPath string, input string, timeLimitSec int64, lang LangConfig) (string, string, bool) {
	absPath, err := filepath.Abs(submissionPath)
	if err != nil {
		return "", "Failed to resolve path: " + err.Error(), false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimitSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",
		"--network=none",
		"--memory=256m",
		"--cpus=1",
		"--user=1001:1001",
		"--read-only",
		"--tmpfs=/tmp:size=64m",
		"-i",
		"-v", absPath+":/sandbox:ro",
		lang.Image,
		"sh", "-c", lang.RunCmd,
	)

	cmd.Stdin = bytes.NewBufferString(input)
	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		slog.Warn("TLE: killing leftover container if any")
		return "", "Time Limit Exceeded", true
	}
	if err != nil {
		slog.Error("Runtime Error", slog.String("output", string(out)))
		cleanOut := strings.ReplaceAll(string(out), "/sandbox/", "")
		return "", "Runtime Error: " + cleanOut, false
	}

	return string(out), "", false
}