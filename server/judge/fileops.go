package judge

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	
)

func CreateAndWriteFile(code string, basePath string, submissionID string) {
    filePath := basePath + submissionID + "/code.cpp"
    log.Println(filePath)
    err := os.MkdirAll(basePath+submissionID, 0755)
    if err != nil {
        panic(err)
    }
    file, err := os.Create(filePath)
    if err != nil {
        panic(err)
    }
    defer file.Close()
    length, err := file.WriteString(code)
    if err != nil {
        panic(err)
    }
    log.Printf("File: %s, Length: %d\n", file.Name(), length)
}

func DeleteFile(basePath string, submissionID string) {
	absPath, _ := filepath.Abs(basePath + submissionID)

	// Delete from inside container to handle Linux-owned files on Windows
	exec.Command("docker", "run",
		"--rm",
		"--user=1001:1001",
		"-v", absPath+":/sandbox:rw",
		"judge-sandbox",
		"sh", "-c", "rm -f /sandbox/code.cpp /sandbox/main",
	).Run()

	// Now remove the directory from host (should be empty or have host-owned files only)
	paths := []string{
		basePath + submissionID + "/code.cpp",
		basePath + submissionID + "/main",
		basePath + submissionID,
	}
	for _, p := range paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			log.Println("Error deleting:", p, err)
		} else {
			log.Println("Deleted:", p)
		}
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

