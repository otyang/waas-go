package response

// render

import (
	"encoding/json"
	"net/http"
)

// APISuccess standardizes success responses for HTTP APIs
type APISuccess struct {
	Success    bool           `json:"success"`        // Always true for successful responses
	StatusCode int            `json:"-"`              // Excluded from JSON
	Message    string         `json:"message"`        // User-facing message
	Data       any            `json:"data"`           // Main response payload
	Meta       map[string]any `json:"meta,omitempty"` // Additional metadata
}

// NewAPISuccess creates a base success response with status code and message
func NewAPISuccess(statusCode int, message string) *APISuccess {
	return &APISuccess{
		Success:    true,
		StatusCode: statusCode,
		Message:    message,
		Meta:       make(map[string]any),
	}
}

// Msg sets the success message
func (s *APISuccess) Msg(msg string) *APISuccess {
	s.Message = msg
	return s
}

// WithData sets the main response data
func (s *APISuccess) WithData(data any) *APISuccess {
	s.Data = data
	return s
}

// WithMeta adds a single metadata key-value pair
func (s *APISuccess) WithMeta(key string, value any) *APISuccess {
	if s.Meta == nil {
		s.Meta = make(map[string]any)
	}
	s.Meta[key] = value
	return s
}

// WithMetas adds multiple metadata key-value pairs
func (s *APISuccess) WithMetas(meta map[string]any) *APISuccess {
	if s.Meta == nil {
		s.Meta = make(map[string]any)
	}
	for k, v := range meta {
		s.Meta[k] = v
	}
	return s
}

// Write sends the success response as JSON
func (s *APISuccess) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s.StatusCode)
	return json.NewEncoder(w).Encode(s)
}

// Predefined common success responses
var (
	SuccessOK        = NewAPISuccess(http.StatusOK, "Success")
	SuccessCreated   = NewAPISuccess(http.StatusCreated, "Resource created")
	SuccessAccepted  = NewAPISuccess(http.StatusAccepted, "Request accepted")
	SuccessNoContent = NewAPISuccess(http.StatusNoContent, "No content") // Note: Will typically exclude body
)

// ==================

func ExampleUsage(w http.ResponseWriter, r *http.Request) {
	results := []string{"result1", "result2", "result3"}

	err := NewAPISuccess(http.StatusOK, "Search completed").
		WithData(results).
		WithMeta("query", r.URL.Query().Get("q")).
		WithMeta("resultsCount", len(results)).
		Write(w)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}

	// Use predefined success response
	_ = SuccessCreated.
		WithData(map[string]string{"id": "789"}).
		WithMeta("location", "/users/789")
}
