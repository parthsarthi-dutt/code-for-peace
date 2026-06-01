package judge

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	
)

// CreateAndWriteFile writes the user's source code to disk with the correct
// filename for the language (e.g., code.cpp, code.py, Main.java).
//
// Java requires the filename to match the class name, so we use "Main.java".
func CreateAndWriteFile(code string, basePath string, submissionID string, fileExtension string) {
    // Java requires the public class name to match the file name
    fileName := "code" + fileExtension
    if fileExtension == ".java" {
        fileName = "Main.java"
    }

    dirPath := basePath + submissionID
    filePath := dirPath + "/" + fileName
    slog.Debug("File created", slog.String("path", filePath))
    err := os.MkdirAll(dirPath, 0777)
    if err != nil {
        panic(err)
    }
    // Explicitly chmod to bypass umask so Docker user 1001 can write compiled binaries
    os.Chmod(dirPath, 0777)
    file, err := os.Create(filePath)
    if err != nil {
        panic(err)
    }
    defer file.Close()
    length, err := file.WriteString(code)
    if err != nil {
        panic(err)
    }
    slog.Debug("File written", slog.String("file", file.Name()), slog.Int("length", length))
}

func DeleteFile(basePath string, submissionID string) {
	absPath, _ := filepath.Abs(basePath + submissionID)

	// Delete from inside container to handle Linux-owned files on Windows/Linux
	// rm -rf /sandbox/* deletes any code.cpp, Main.java, code.py, main, Main.class, etc.
	exec.Command("docker", "run",
		"--rm",
		"--user=1001:1001",
		"-v", absPath+":/sandbox:rw",
		"judge-sandbox",
		"sh", "-c", "rm -rf /sandbox/*",
	).Run()

	// Now completely remove the directory from the host
	dirPath := basePath + submissionID
	if err := os.RemoveAll(dirPath); err != nil {
		slog.Error("Error deleting directory", slog.String("directory", dirPath), slog.String("error", err.Error()))
	} else {
		slog.Info("Deleted directory", slog.String("directory", dirPath))
	}
}

func ReadTestFile(path string, test string) string {
    file, err := os.ReadFile(path + test + ".txt")
    if err != nil {
        panic(err)
    }
    return string(file)
}

func ReadExpecFile(path string, test string) string {
    file, err := os.ReadFile(path + test + ".txt")
    if err != nil {
        panic(err)
    }
    return string(file)
}

