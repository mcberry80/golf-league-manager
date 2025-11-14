package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"

	"golf-league-manager/server/internal/logger"
)

// RateLimiter is a middleware that implements rate limiting per IP address
// using a token bucket algorithm.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter middleware.
// rate is the number of requests per second allowed per IP.
// burst is the maximum burst size.
func NewRateLimiter(r float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(r),
		burst:    burst,
	}
}

// getLimiter returns the rate limiter for a given IP address
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		limiter, exists = rl.limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(rl.rate, rl.burst)
			rl.limiters[ip] = limiter
		}
		rl.mu.Unlock()
	}

	return limiter
}

// Handler returns an HTTP middleware that enforces rate limiting
func (rl *RateLimiter) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract IP address (handle X-Forwarded-For for proxies)
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.Header.Get("X-Real-IP")
			}
			if ip == "" {
				ip = r.RemoteAddr
			}

			limiter := rl.getLimiter(ip)

			if !limiter.Allow() {
				logger.WarnContext(r.Context(), "Rate limit exceeded",
					"ip", ip,
					"path", r.URL.Path,
				)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit returns a rate limiting middleware with sensible defaults.
// Default: 10 requests per second with burst of 20.
func RateLimit() func(http.Handler) http.Handler {
	limiter := NewRateLimiter(10.0, 20)
	return limiter.Handler()
}
