package apiextensions

import (
	"net/http"

	"github.com/parthsarthi-dutt/online-judge/server/auth"
)

func RegisterExtensions() {

	// public endpoints
	http.HandleFunc("/api/problems", ListProblemsHandler)
	http.HandleFunc("/api/problem", GetProblemHandler)

	// authenticated endpoints
	http.Handle("/api/user/profile", auth.AuthMiddleware(http.HandlerFunc(GetUserProfileHandler)))
	http.Handle("/api/user/profile/update", auth.AuthMiddleware(http.HandlerFunc(UpdateUserProfileHandler)))
	http.Handle("/api/judge/run", auth.AuthMiddleware(http.HandlerFunc(RunCodeHandler)))
	http.Handle("/api/problem/editorial", auth.AuthMiddleware(http.HandlerFunc(GetEditorialHandler)))
	http.Handle("/api/problem/editorial/unlock", auth.AuthMiddleware(http.HandlerFunc(UnlockEditorialHandler)))

	// AI Interview endpoints
	http.Handle("/api/interview/start", auth.AuthMiddleware(http.HandlerFunc(StartInterviewHandler)))
	http.Handle("/api/interview/respond", auth.AuthMiddleware(http.HandlerFunc(ProcessInterviewResponseHandler)))
	http.Handle("/api/interview/end", auth.AuthMiddleware(http.HandlerFunc(EndInterviewHandler)))
	http.Handle("/api/interview/list", auth.AuthMiddleware(http.HandlerFunc(GetInterviewsHandler)))
	http.Handle("/api/interview/detail", auth.AuthMiddleware(http.HandlerFunc(GetInterviewDetailHandler)))
}

