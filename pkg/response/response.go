package response

import (
	"time"
)

// Response represents a standardized API response.
type Response struct {
	Data      interface{} `json:"data,omitempty"`
	Error     *string     `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
	Success   bool        `json:"success"`
}

// Success returns a successful response structure.
func Success(data interface{}) Response {
	return Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// Error returns an error response structure.
func Error(message string) Response {
	return Response{
		Success:   false,
		Error:     &message,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
