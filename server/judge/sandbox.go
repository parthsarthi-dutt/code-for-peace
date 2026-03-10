package judge

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"
)

const sandboxImage = "judge-sandbox"

func CompileInSandbox(submissionPath string) (bool, string) {
	absPath, err := filepath.Abs(submissionPath)
	if err != nil {
		return false, "Failed to resolve path: " + err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Compile to a unique temp name, then atomically rename to "main"
	// This avoids "Text file busy" from a lingering previous container
	tmpBinary := fmt.Sprintf("/sandbox/main_%d", time.Now().UnixNano())

	cmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",
		"--network=none",
		"--memory=256m",
		"--cpus=1",
		"--user=1001:1001",
		"--read-only",
		"--tmpfs=/tmp:size=64m",
		"-v", absPath+":/sandbox:rw",
		sandboxImage,
		"sh", "-c",
		fmt.Sprintf("g++ /sandbox/code.cpp -o %s -O2 -std=c++17 -Wall && mv %s /sandbox/main", tmpBinary, tmpBinary),
	)

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return false, "Compilation Timed Out"
	}
	if err != nil {
		log.Println("Compilation Error:", string(out))
		return false, string(out)
	}

	log.Println("Compilation successful")
	return true, ""
}

func ExecuteInSandbox(submissionPath string, input string, timeLimitSec int64) (string, string, bool) {
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
		sandboxImage,
		"/sandbox/main",
	)

	cmd.Stdin = bytes.NewBufferString(input)
	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		log.Println("TLE: killing leftover container if any")
		return "", "Time Limit Exceeded", true
	}
	if err != nil {
		log.Println("Runtime Error:", string(out))
		return "", "Runtime Error: " + string(out), false
	}

	return string(out), "", false
}