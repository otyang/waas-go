package response

import (
	"fmt"
	"net/http"
)

// APIError standardizes error responses for APIs
type APIError struct {
	Success        bool           `json:"success"`             // Always false for error responses
	HTTPStatusCode int            `json:"-"`                   // HTTP Status Excluded from JSON
	Message        string         `json:"message"`             // User-facing /Human readable message
	ErrorCode      string         `json:"errorCode,omitempty"` // Error classification
	ErrorDetails   map[string]any `json:"errors,omitempty"`    // Additional error context/details
	InternalError  error          `json:"-"`                   // Original error (for internal logging)
}

// NewAPIError creates a base error with status code and message
func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{
		HTTPStatusCode: statusCode,
		Success:        false,
		Message:        message,
		ErrorDetails:   make(map[string]any),
	}
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.InternalError != nil {
		return fmt.Sprintf("%s (internal: %v)", e.Message, e.InternalError)
	}
	return e.Message
}

// Builder pattern methods
func (e *APIError) Msg(msg string) *APIError { e.Message = msg; return e }
func (e *APIError) Code(t string) *APIError  { e.ErrorCode = t; return e }
func (e *APIError) Err(err error) *APIError  { e.InternalError = err; return e }

func (e *APIError) Detail(key string, value any) *APIError {
	if e.ErrorDetails == nil {
		e.ErrorDetails = make(map[string]any)
	}
	e.ErrorDetails[key] = value
	return e
}

func (e *APIError) Details(details map[string]any) *APIError {
	if e.ErrorDetails == nil {
		e.ErrorDetails = make(map[string]any)
	}
	for k, v := range details {
		e.ErrorDetails[k] = v
	}
	return e
}

// Predefined common errors
var (
	ErrBadRequest          = NewAPIError(http.StatusBadRequest, "Bad Request").Code("bad_request")
	ErrUnauthorized        = NewAPIError(http.StatusUnauthorized, "Unauthorized").Code("unauthorized")
	ErrForbidden           = NewAPIError(http.StatusForbidden, "Forbidden").Code("forbidden")
	ErrNotFound            = NewAPIError(http.StatusNotFound, "Not Found").Code("not_found")
	ErrValidation          = NewAPIError(http.StatusUnprocessableEntity, "Validation Error").Code("validation_error")
	ErrInternalServerError = NewAPIError(http.StatusInternalServerError, "Internal Server Error").Code("server_error")
)

// FromError converts standard errors to APIError
func FromError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return ErrInternalServerError.Err(err)
}
