package main

import (
	"log"
	"net/http"

	"github.com/parthsarthi-dutt/online-judge/api"
	"github.com/parthsarthi-dutt/online-judge/server/auth"
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
	auth.InitGoogleOAuth()
	go worker.StartWorker(database.RDB)
	http.HandleFunc("/auth/google/login", auth.GoogleLogin)
	http.HandleFunc("/auth/google/callback", auth.GoogleCallback)
	http.Handle("/practice/submit",auth.AuthMiddleware(http.HandlerFunc(api.SubmissionEndpointPractice))) 
	http.Handle("/contest/submit",auth.AuthMiddleware(http.HandlerFunc( api.SubmissionEndpointContest))) 
	http.Handle("/submission", auth.AuthMiddleware(http.HandlerFunc(api.GetSubmissionsEndpoint)))
	http.Handle("/user/submissions", auth.AuthMiddleware(http.HandlerFunc(api.GetAllSubmissionsEndpoint)))



	log.Println("Server running on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	// worker.StartWorker(database.RDB)

}