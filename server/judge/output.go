package judge

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/parthsarthi-dutt/online-judge/server/models"
)

func CreateAndWriteFile(s1 string,path string,submissionID string){
	filePath:=path+submissionID+"/code.cpp"
	log.Println(filePath)
		err := os.MkdirAll(
			path+submissionID, 0755)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(filePath)

	if err != nil {
		panic(err)
	}
	length,err:=file.WriteString(s1)
	if err!=nil{
		panic(err);
	}
	defer file.Close()
	log.Printf("File name: %s",file.Name())
	log.Printf("\nfile length: %d\n",length)
}
func DeleteFile(path string,submissionID string){
	err := os.Remove(path+submissionID+"/code.cpp")
	if err != nil {
		log.Println("Error deleting directory:", err)
	}
	err = os.Remove(path+submissionID+"/main.exe")
	if err != nil {
		log.Println("Error deleting directory:", err)
	}
	err = os.Remove(path+submissionID)
	if err != nil {
		log.Println("Error deleting directory:", err)
	}else {
		log.Println("Directory deleted successfully")
	}
}
func ReadCodeFile(path string) string{
	file,err:=os.ReadFile(path+"/code.cpp")
	if err!=nil{
		panic(err)
	}
	
	log.Printf("content: %s\n",file)
	return string(file)
}
func ReadTestFile(path string,test string) string{
	file,err:=os.ReadFile(path+test+".txt")
	if err!=nil{
		panic(err)
	}
	
	log.Printf("input: %s\n",file)
	return string(file)
}
func ReadExpecFile(path string,test string)string{
	file,err:=os.ReadFile(path+test+".txt")
	if err!=nil{
		panic(err)
	}
	
	log.Printf("expected: %s\n",file)
	return string(file)
}
func RunFile(path string)bool{
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd:=exec.CommandContext(ctx,"g++",path+"/code.cpp","-o",path+"/main")

	out,err:=cmd.CombinedOutput()
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Time Limit Exceeded")
				
				log.Printf("Verdict:Time Limit Exceded\n")
				return false
			}
	if err!=nil{
		log.Println("Compilation Error")
		log.Println(string(out))
		log.Println()
		return false
		// panic(err)
	}
	log.Printf("File Compiled")
	return true
}

func ExecuteBinary(ProblemPath string,BinaryPath string,timeLimit int64,memoryLimit int64)(models.Output, error){
	
	entries, err := os.ReadDir(ProblemPath+"/input")
		if err != nil {
		panic(err)
	}
	
	t:=time.Now()
	Correct,Incorrect :=0,0

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			input:=ReadTestFile(ProblemPath+"/input/",strings.TrimSuffix(name, ".txt"))
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit)*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, BinaryPath+"/main")
			cmd.Stdin = bytes.NewBufferString(input)
			output,err:=cmd.CombinedOutput()

			

			if ctx.Err() == context.DeadlineExceeded {
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
			expectedOutput:=strings.TrimSpace(expected)
			userOutput:=strings.TrimSpace(string(output))

				if err != nil {
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
func main(){

	// path:="problems/124-A"
	// ReadCodeFile(path)
	// flag:=RunFile(path)
	// if(flag){ExecuteBinary(path)}
}