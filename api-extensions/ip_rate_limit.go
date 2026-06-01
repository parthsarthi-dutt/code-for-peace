package apiextensions

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// client represents a single IP address rate limiter
type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*client)
)

// cleanupVisitors removes stale IPs from the map every 5 minutes
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 10*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := clients[ip]
	if !exists {
		// Limit to 5 requests per second, with a burst of 10 requests.
		// This protects against aggressive DDOS bots while allowing normal user interaction.
		limiter := rate.NewLimiter(rate.Limit(5), 10)
		clients[ip] = &client{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// IPRateLimitMiddleware globally limits traffic by IP to protect cloud infrastructure
func IPRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract IP from RemoteAddr or X-Forwarded-For if behind a proxy
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		
		// If behind a load balancer (like AWS ELB), use X-Forwarded-For
		forwarded := r.Header.Get("X-Forwarded-For")
		if forwarded != "" {
			ip = forwarded
		}

		// Exempt localhost from rate limiting for local development
		if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
			next.ServeHTTP(w, r)
			return
		}

		limiter := getLimiter(ip)
		if !limiter.Allow() {
			slog.Warn("Global IP Rate Limit Exceeded", slog.String("ip", ip))
			http.Error(w, `{"error": "Too many requests. Please wait and try again."}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
