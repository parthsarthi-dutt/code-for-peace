package api

import "github.com/parthsarthi-dutt/online-judge/server/database"

func ProblemEndpoint() {
	
	database.CreateProblem("123-A",2,200)
	database.CreateProblem("124-A",2,200)
	database.CreateProblem("125-A",2,200)
	database.CreateProblem("126-A",2,200)
	database.CreateProblem("127-A",2,200)
	database.CreateProblem("128-A",2,200)
	database.CreateProblem("129-A",2,200)
}