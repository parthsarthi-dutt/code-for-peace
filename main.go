package main

import (
	"fmt"

	"github.com/parthsarthi-dutt/online-judge/api"
	"github.com/parthsarthi-dutt/online-judge/server/database"
	"github.com/parthsarthi-dutt/online-judge/server/queue/worker"
)

func main(){
	database.ConnectRedis()
	fmt.Println("Connected to Database")
	api.ProblemEndpoint()
	api.SubmissionEndpoint()
	worker.StartWorker(database.RDB)
	// worker.StartWorker(database.RDB)

}