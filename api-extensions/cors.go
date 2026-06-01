package apiextensions

import (
	"net/http"
	"strings"
)

// allowedOrigins lists all frontend origins that are permitted.
var allowedOrigins = map[string]bool{
	"http://localhost:5173": true,
	"http://localhost:5174": true,
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] || strings.HasPrefix(origin, "http://localhost:") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
