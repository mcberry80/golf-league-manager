package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	allowedOrigins := []string{"http://localhost:3000", "https://example.com"}
	handler := CORS(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		origin         string
		method         string
		expectStatus   int
		expectOrigin   string
	}{
		{
			name:         "Allowed origin",
			origin:       "http://localhost:3000",
			method:       "GET",
			expectStatus: http.StatusOK,
			expectOrigin: "http://localhost:3000",
		},
		{
			name:         "Another allowed origin",
			origin:       "https://example.com",
			method:       "GET",
			expectStatus: http.StatusOK,
			expectOrigin: "https://example.com",
		},
		{
			name:         "Disallowed origin",
			origin:       "https://malicious.com",
			method:       "GET",
			expectStatus: http.StatusOK,
			expectOrigin: "",
		},
		{
			name:         "Preflight request",
			origin:       "http://localhost:3000",
			method:       "OPTIONS",
			expectStatus: http.StatusNoContent,
			expectOrigin: "http://localhost:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, rr.Code)
			}

			if tt.expectOrigin != "" {
				actualOrigin := rr.Header().Get("Access-Control-Allow-Origin")
				if actualOrigin != tt.expectOrigin {
					t.Errorf("Expected CORS origin %s, got %s", tt.expectOrigin, actualOrigin)
				}

				// Check other CORS headers are set
				if rr.Header().Get("Access-Control-Allow-Methods") == "" {
					t.Error("Access-Control-Allow-Methods header not set")
				}
				if rr.Header().Get("Access-Control-Allow-Headers") == "" {
					t.Error("Access-Control-Allow-Headers header not set")
				}
			}
		})
	}
}

func TestCORSWildcard(t *testing.T) {
	handler := CORS([]string{"*"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	actualOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if actualOrigin != "https://any-origin.com" {
		t.Errorf("Expected CORS origin to match request origin, got %s", actualOrigin)
	}
}
