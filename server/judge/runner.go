package judge

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/parthsarthi-dutt/online-judge/server/models"
)

// RunFile compiles the submission. Returns false + compile error message on failure.
func RunFile(submissionPath string) bool {
    ok, errMsg := CompileInSandbox(submissionPath)
    if !ok {
        log.Println("Compile error:", errMsg)
    }
    return ok
}

// ExecuteBinary runs all test cases and returns a verdict.
// submissionPath: absolute host path to the submission dir (has code.cpp + main)
// problemPath:    absolute host path to the problem dir (has input/ and output/)
func ExecuteBinary(ProblemPath string, BinaryPath string, timeLimit int64, MemoryLimit int64) (models.Output, error) {
    
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

		output, err, tle := ExecuteInSandbox(BinaryPath, input, timeLimit)
       

       

        // t2 := time.Since(start).Milliseconds()

        if tle {
				log.Println("Time Limit Exceeded")
				log.Printf("Time Limit Exceded on test: %s\n",strings.TrimSuffix(name, ".txt"))
				totalTime:=time.Since(t)
				log.Printf("Verdict:Time Limit Exceded (%d ms)\n\n",totalTime.Milliseconds())
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
					log.Println("Runtime Error")
					log.Println(string(output))
					totalTime:=time.Since(t)
					log.Printf("Verdict: Runtime Error (%d ms)\n\n",totalTime.Milliseconds())
					result:=models.Output{
					Verdict: "Runtime Error",
					ExecutionTime: totalTime.Milliseconds(),
					MemoryUsed: totalTime.Microseconds(),
					Message: "Runtime Error on test "+strings.TrimSuffix(name, ".txt"),
				}
					return result,nil
				}
			log.Printf("Output: %s",string(output))
			if(userOutput==expectedOutput){
				log.Printf("Accepted\n\n")
				Correct++;
			}else{
				log.Printf("Wrong Answer\n\n")
				log.Printf("Wrong Answer on test: %s\n\n",strings.TrimSuffix(name, ".txt"))
				log.Printf("Expected\n%s\n\n",expectedOutput)
				log.Printf("Found\n%s\n\n",userOutput)
				totalTime:=time.Since(t)
				log.Printf("Verdict:Wrong Answer (%d ms)\n\n",totalTime.Milliseconds())
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
	log.Printf("Total TestCases: %d\n",Correct+Incorrect)
	log.Printf("Accepted TestCases: %d\n",Correct)
	log.Printf("Failed TestCases: %d\n\n",Incorrect)
	totalTime:=time.Since(t)
	log.Printf("Verdict: ")
	
	log.Printf("Accepted (%d ms)\n\n",totalTime.Milliseconds())
		result:=models.Output{
		Verdict: "Accepted",
		ExecutionTime: totalTime.Milliseconds(),
		MemoryUsed: totalTime.Microseconds(),
		Message: "All testcases Passed",
		}
	return result,nil
}