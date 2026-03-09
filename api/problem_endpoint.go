package api

import "github.com/parthsarthi-dutt/online-judge/server/database"

func ProblemEndpoint() {
	database.CreateProblem("123-A",2,2)
	database.CreateProblem("124-A",2,2)
}