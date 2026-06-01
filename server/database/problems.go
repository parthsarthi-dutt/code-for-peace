package database

import (
	"encoding/json"
	"log/slog"

	"github.com/parthsarthi-dutt/online-judge/server/models"
)

var ProblemDB = make(map[string][]byte)

func CreateProblem(problemID string,timeLimit int64,memoryLimit int64){
	slog.Info("Problem Recorded", slog.String("problem_id", problemID))

	problem:=models.Problem{
		
		TimeLimit:timeLimit,
		MemoryLimit:memoryLimit,
		
	}

	data,err:=json.Marshal(problem)
	if(err!=nil){
		panic(err)
	}
	ProblemDB[problemID]=data
	slog.Info("Problem Created Successfully", slog.String("problem_id", problemID))

}

func GetTimeAndMemory(problemID string)[]int64 {
	data, exists := ProblemDB[problemID]
	if !exists {
		panic("problem not found")
	}

	var problem models.Problem
	ans := make([]int64, 2)
	err := json.Unmarshal(data, &problem)
	if err != nil {
		panic(err)
	}
	ans[0] = problem.TimeLimit
	ans[1] = problem.MemoryLimit
	slog.Debug("Problem Info Returned", slog.String("problem_id", problemID))
	return ans
}