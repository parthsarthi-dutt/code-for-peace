package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func CreateAndWriteFile(s1 string,path string){
		file, err := os.Create(path+"/code.cpp")

	if err != nil {
		panic(err)
	}
	length,err:=file.WriteString(s1)
	if err!=nil{
		panic(err);
	}
	defer file.Close()
	fmt.Printf("File name: %s",file.Name())
	fmt.Printf("\nfile length: %d\n",length)
}

func ReadCodeFile(path string) string{
	file,err:=os.ReadFile(path+"/code.cpp")
	if err!=nil{
		panic(err)
	}
	
	fmt.Printf("content: %s\n",file)
	return string(file)
}
func ReadTestFile(path string,test string) string{
	file,err:=os.ReadFile(path+test+".txt")
	if err!=nil{
		panic(err)
	}
	
	fmt.Printf("input: %s\n",file)
	return string(file)
}
func ReadExpecFile(path string,test string)string{
	file,err:=os.ReadFile(path+test+".txt")
	if err!=nil{
		panic(err)
	}
	
	fmt.Printf("expected: %s\n",file)
	return string(file)
}
func RunFile(path string)bool{
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd:=exec.CommandContext(ctx,"g++",path+"/code.cpp","-o",path+"/main")

	out,err:=cmd.CombinedOutput()
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Println("Time Limit Exceeded")
				
				fmt.Printf("Verdict:Time Limit Exceded\n")
				return false
			}
	if err!=nil{
		fmt.Println("Compilation Error")
		fmt.Println(string(out))
		fmt.Println()
		return false
		// panic(err)
	}
	fmt.Printf("File Compiled")
	return true
}
func ExecuteBinary(path string){
	
	entries, err := os.ReadDir(path+"/input")
		if err != nil {
		panic(err)
	}

	t:=time.Now()
	Correct,Incorrect :=0,0

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			input:=ReadTestFile(path+"/input/",strings.TrimSuffix(name, ".txt"))
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, path+"/main")
			cmd.Stdin = bytes.NewBufferString(input)
			output,err:=cmd.CombinedOutput()

			

			if ctx.Err() == context.DeadlineExceeded {
				fmt.Println("Time Limit Exceeded")
				fmt.Printf("Time Limit Exceded on test: %s\n",strings.TrimSuffix(name, ".txt"))
				totalTime:=time.Since(t)
				fmt.Printf("Verdict:Time Limit Exceded (%d ms)\n\n",totalTime.Milliseconds())
				return
			}
			expected:=ReadExpecFile(path+"/output/",strings.TrimSuffix(name, ".txt"))
			expectedOutput:=strings.TrimSpace(expected)
			userOutput:=strings.TrimSpace(string(output))

				if err != nil {
					fmt.Println("Runtime Error")
					fmt.Println(string(output))
					totalTime:=time.Since(t)
					fmt.Printf("Verdict: Runtime Error (%d ms)\n\n",totalTime.Milliseconds())
					return
				}
			fmt.Printf("Output: %s",string(output))
			if(userOutput==expectedOutput){
				fmt.Printf("Accepted\n\n")
				Correct++;
			}else{
				fmt.Printf("Wrong Answer\n\n")
				fmt.Printf("Wrong Answer on test: %s\n\n",strings.TrimSuffix(name, ".txt"))
				fmt.Printf("Expected\n%s\n\n",expectedOutput)
				fmt.Printf("Found\n%s\n\n",userOutput)
				totalTime:=time.Since(t)
				fmt.Printf("Verdict:Wrong Answer (%d ms)\n\n",totalTime.Milliseconds())
				return
				// Incorrect++;
			}
			
		}
	}
	fmt.Printf("Total TestCases: %d\n",Correct+Incorrect)
	fmt.Printf("Accepted TestCases: %d\n",Correct)
	fmt.Printf("Failed TestCases: %d\n\n",Incorrect)
	totalTime:=time.Since(t)
	fmt.Printf("Verdict: ")
	
	fmt.Printf("Accepted (%d ms)\n\n",totalTime.Milliseconds())
	
	
}
func main(){

	// path:="problems/124-A"
	// ReadCodeFile(path)
	// flag:=RunFile(path)
	// if(flag){ExecuteBinary(path)}
}