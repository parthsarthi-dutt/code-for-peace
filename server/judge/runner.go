package judge

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/parthsarthi-dutt/online-judge/server/models"
)

// RunFile compiles the submission if the language requires it.
// For interpreted languages (Python), this is a no-op that returns true.
//
// Interview answer: "Not all languages need compilation. Python is interpreted,
// so RunFile checks NeedsCompilation and skips the Docker compile step entirely."
func RunFile(submissionPath string, lang LangConfig) (bool, string) {
    if !lang.NeedsCompilation {
        slog.Info("Language does not need compilation, skipping")
        return true, ""
    }
    ok, errMsg := CompileInSandbox(submissionPath, lang)
    if !ok {
        slog.Error("Compile error", slog.String("error", errMsg))
    }
    return ok, errMsg
}

// ExecuteBinary runs all test cases and returns a verdict.
// submissionPath: absolute host path to the submission dir (has code.cpp + main)
// problemPath:    absolute host path to the problem dir (has input/ and output/)
func ExecuteBinary(ProblemPath string, BinaryPath string, timeLimit int64, MemoryLimit int64, lang LangConfig) (models.Output, error) {
    
	entries, err := os.ReadDir(ProblemPath+"/input")
    if err != nil {
        panic(err)
    }

    t := time.Now()
	Correct,Incorrect:=0,0
	for _, entry := range entries {
		if !entry.IsDir() {
		name := entry.Name()
		input:=ReadTestFile(ProblemPath+"/input/",strings.TrimSuffix(name, ".txt"))

		output, err, tle := ExecuteInSandbox(BinaryPath, input, timeLimit, lang)
       

       

        // t2 := time.Since(start).Milliseconds()

        if tle {
				slog.Warn("Time Limit Exceeded on test", slog.String("test", strings.TrimSuffix(name, ".txt")))
				totalTime:=time.Since(t)
				slog.Info("Verdict: Time Limit Exceeded", slog.Int64("execution_ms", totalTime.Milliseconds()))
				result:=models.Output{
					Verdict: "Time Limit Exceeded",
					ExecutionTime: totalTime.Milliseconds(),
					MemoryUsed: totalTime.Microseconds(),
					Message: "Time Limit Exceeded on test "+strings.TrimSuffix(name, ".txt"),
				}
				return result,nil
			}
			expected:=ReadExpecFile(ProblemPath+"/output/",strings.TrimSuffix(name, ".txt"))
			expectedOutput := strings.ReplaceAll(strings.TrimSpace(expected), "\r\n", "\n")
			userOutput := strings.ReplaceAll(strings.TrimSpace(string(output)), "\r\n", "\n")

				if err!="" {
					slog.Warn("Runtime Error", slog.String("test", strings.TrimSuffix(name, ".txt")), slog.String("output", string(output)))
					totalTime:=time.Since(t)
					slog.Info("Verdict: Runtime Error", slog.Int64("execution_ms", totalTime.Milliseconds()))
					result:=models.Output{
					Verdict: "Runtime Error",
					ExecutionTime: totalTime.Milliseconds(),
					MemoryUsed: totalTime.Microseconds(),
					Message: "Runtime Error on test "+strings.TrimSuffix(name, ".txt"),
				}
					return result,nil
				}
			slog.Debug("Execution output", slog.String("output", string(output)))
			if(userOutput==expectedOutput){
				slog.Info("Testcase Accepted", slog.String("test", strings.TrimSuffix(name, ".txt")))
				Correct++;
			}else{
				slog.Warn("Wrong Answer on test", slog.String("test", strings.TrimSuffix(name, ".txt")))
				slog.Debug("Wrong Answer Details", slog.String("expected", expectedOutput), slog.String("found", userOutput))
				totalTime:=time.Since(t)
				slog.Info("Verdict: Wrong Answer", slog.Int64("execution_ms", totalTime.Milliseconds()))
				result:=models.Output{
					Verdict: "Wrong Answer",
					ExecutionTime: totalTime.Milliseconds(),
					MemoryUsed: totalTime.Microseconds(),
					Message: "Wrong Answer on test "+strings.TrimSuffix(name, ".txt")+" , Expected Output: "+expectedOutput+" , Found Output: "+userOutput,
				}
				return result,nil
				// Incorrect++;
			}
		}
		
	}
	slog.Info("Execution finished", slog.Int("total_testcases", Correct+Incorrect), slog.Int("accepted", Correct), slog.Int("failed", Incorrect))
	totalTime:=time.Since(t)
	slog.Info("Verdict: Accepted", slog.Int64("execution_ms", totalTime.Milliseconds()))
		result:=models.Output{
		Verdict: "Accepted",
		ExecutionTime: totalTime.Milliseconds(),
		MemoryUsed: totalTime.Microseconds(),
		Message: "All testcases Passed",
		}
	return result,nil
}