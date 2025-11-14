package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golf-league-manager/server/internal/logger"
)

func TestRequestID(t *testing.T) {
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID was added to context
		requestID := r.Context().Value(logger.RequestIDKey)
		if requestID == nil {
			t.Error("Request ID not found in context")
		}

		// Check if request ID is a string
		if _, ok := requestID.(string); !ok {
			t.Error("Request ID is not a string")
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Check if X-Request-ID header was set in response
	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header not set in response")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestRequestIDPreserved(t *testing.T) {
	existingRequestID := "test-request-id-123"
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(logger.RequestIDKey).(string)
		if requestID != existingRequestID {
			t.Errorf("Expected request ID %s, got %s", existingRequestID, requestID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingRequestID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Check if X-Request-ID header matches the one we sent
	if rr.Header().Get("X-Request-ID") != existingRequestID {
		t.Errorf("Expected X-Request-ID %s, got %s", existingRequestID, rr.Header().Get("X-Request-ID"))
	}
}
