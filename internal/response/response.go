package response

import (
	"encoding/json"
	"net/http"
)

// Response is the standard API response structure
type Response struct {
	Status  string      `json:"status"`  // "success" or "error"
	Code    int         `json:"code"`    // HTTP status code
	Message string      `json:"message"` // Human-readable message
	Data    interface{} `json:"data"`    // Response data (can be null)
}

// SuccessResponse returns a success response
func SuccessResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	response := Response{
		Status:  "success",
		Code:    code,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// ErrorResponse returns an error response
func ErrorResponse(w http.ResponseWriter, code int, message string) {
	response := Response{
		Status:  "error",
		Code:    code,
		Message: message,
		Data:    nil,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// ValidationErrorResponse returns a validation error response with details
func ValidationErrorResponse(w http.ResponseWriter, code int, message string, errors map[string]string) {
	response := Response{
		Status:  "error",
		Code:    code,
		Message: message,
		Data:    errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}
