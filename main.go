package main

import (
	"log"
	"net/http"

	"github.com/parthsarthi-dutt/online-judge/api"
	"github.com/parthsarthi-dutt/online-judge/server/database"
	"github.com/parthsarthi-dutt/online-judge/server/database/postgres"
	"github.com/parthsarthi-dutt/online-judge/server/queue/worker"
)

func main(){
	postgres.ConnectDB()
	log.Println("Connected to Postgres Database")
	database.ConnectRedis()
	log.Println("Connected to Redis Database")
	api.ProblemEndpoint()

	go worker.StartWorker(database.RDB)
	http.HandleFunc("/practice/submit", api.SubmissionEndpointPractice)
	http.HandleFunc("/contest/submit", api.SubmissionEndpointContest)
	http.HandleFunc("/submission", api.GetSubmissionsEndpoint)
	http.HandleFunc("/user/submissions", api.GetAllSubmissionsEndpoint)

	log.Println("Server running on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	// worker.StartWorker(database.RDB)

}