package response

import (
	"encoding/json"
	"net/http"

	"golf-league-manager/server/internal/logger"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains error details for API responses
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", "error", err)
	}
}

// WriteSuccess writes a successful JSON response with 200 OK status
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// WriteCreated writes a successful creation response with 201 Created status
func WriteCreated(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// WriteError writes an error response with the appropriate status code
func WriteError(w http.ResponseWriter, statusCode int, code, message string) {
	WriteJSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// WriteErrorWithDetails writes an error response with additional details
func WriteErrorWithDetails(w http.ResponseWriter, statusCode int, code, message, details string) {
	WriteJSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// WriteBadRequest writes a 400 Bad Request error response
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// WriteUnauthorized writes a 401 Unauthorized error response
func WriteUnauthorized(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// WriteForbidden writes a 403 Forbidden error response
func WriteForbidden(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusForbidden, "FORBIDDEN", message)
}

// WriteNotFound writes a 404 Not Found error response
func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, "NOT_FOUND", message)
}

// WriteConflict writes a 409 Conflict error response
func WriteConflict(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, "CONFLICT", message)
}

// WriteInternalError writes a 500 Internal Server Error response
func WriteInternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
