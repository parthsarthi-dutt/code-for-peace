package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/parthsarthi-dutt/online-judge/api"
	apiextensions "github.com/parthsarthi-dutt/online-judge/api-extensions"
	"github.com/parthsarthi-dutt/online-judge/server/auth"
	"github.com/parthsarthi-dutt/online-judge/server/database"
	"github.com/parthsarthi-dutt/online-judge/server/database/postgres"
	"github.com/parthsarthi-dutt/online-judge/server/logging"
	"github.com/parthsarthi-dutt/online-judge/server/ai"
	"github.com/parthsarthi-dutt/online-judge/server/queue/worker"
)

func main(){
	logging.InitLogger()

	err := godotenv.Load()
	if err != nil {
		slog.Warn(".env file not found, using system env vars")
	}
	postgres.ConnectDB()
	slog.Info("Connected to Postgres Database")
	database.ConnectRedis()
	slog.Info("Connected to Redis Database")
	auth.InitGoogleOAuth()
	api.ProblemEndpoint()
	go worker.StartWorker(database.RDB)
	
	http.HandleFunc("/auth/google/login", auth.GoogleLogin)
	http.HandleFunc("/auth/google/callback", auth.GoogleCallback)
	http.Handle("/practice/submit",auth.AuthMiddleware(http.HandlerFunc(api.SubmissionEndpointPractice))) 
	http.Handle("/contest/submit",auth.AuthMiddleware(http.HandlerFunc( api.SubmissionEndpointContest))) 
	http.Handle("/submission", auth.AuthMiddleware(http.HandlerFunc(api.GetSubmissionsEndpoint)))
	http.Handle("/user/submissions", auth.AuthMiddleware(http.HandlerFunc(api.GetAllSubmissionsEndpoint)))
	http.Handle("/api/ai/hint", auth.AuthMiddleware(http.HandlerFunc(ai.GetHint)))
	http.Handle("/api/ai/feedback", auth.AuthMiddleware(http.HandlerFunc(ai.GetFeedback)))

	apiextensions.RegisterExtensions()

	slog.Info("Server running on :8080")

	err = http.ListenAndServe(":8080", apiextensions.IPRateLimitMiddleware(apiextensions.CORSMiddleware(http.DefaultServeMux)))
	if err != nil {
		slog.Error("Server failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// worker.StartWorker(database.RDB)

}