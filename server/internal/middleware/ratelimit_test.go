package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	// Create a rate limiter with very low limits for testing
	limiter := NewRateLimiter(1.0, 2) // 1 req/s, burst of 2
	handler := limiter.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First two requests should succeed (burst)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, rr.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr.Code)
	}
}

func TestRateLimitRecovery(t *testing.T) {
	// Create a rate limiter with low limits
	limiter := NewRateLimiter(2.0, 1) // 2 req/s, burst of 1
	handler := limiter.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First request: Expected status 200, got %d", rr1.Code)
	}

	// Second immediate request should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: Expected status 429, got %d", rr2.Code)
	}

	// Wait for rate limit to recover (500ms for 2 req/s)
	time.Sleep(600 * time.Millisecond)

	// Third request should succeed after recovery
	req3 := httptest.NewRequest("GET", "/test", nil)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("Third request after recovery: Expected status 200, got %d", rr3.Code)
	}
}

func TestRateLimitDifferentIPs(t *testing.T) {
	// Create a rate limiter
	limiter := NewRateLimiter(1.0, 1) // 1 req/s, burst of 1
	handler := limiter.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request from first IP
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First IP: Expected status 200, got %d", rr1.Code)
	}

	// Request from second IP should also succeed (different limiter)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Second IP: Expected status 200, got %d", rr2.Code)
	}
}
