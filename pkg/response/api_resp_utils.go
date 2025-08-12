package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// IsJsonError identifies and categorizes JSON parsing errors with detailed messages.
// Returns:
//   - bool: true if the error is JSON-related
//   - error: enriched error message (or original error if not JSON-related)
//
// Handles all standard JSON error types with precise error messages including:
//   - Syntax errors with character position
//   - Type mismatches with field name and position
//   - EOF cases (empty or truncated data)
//   - Size limit errors (when using limited readers)
func IsJsonError(err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var sizeErr *http.MaxBytesError

	switch {
	case errors.As(err, &syntaxErr):
		return true, fmt.Errorf("malformed JSON at position %d: %v", syntaxErr.Offset, syntaxErr)

	case errors.As(err, &typeErr):
		msg := fmt.Sprintf("invalid JSON type for field %q at position %d", typeErr.Field, typeErr.Offset)
		if typeErr.Value != "" {
			msg += fmt.Sprintf(" (got %q)", typeErr.Value)
		}
		return true, fmt.Errorf("%s: expected %s", msg, typeErr.Type)

	case errors.Is(err, io.ErrUnexpectedEOF):
		return true, errors.New("truncated JSON data")

	case errors.Is(err, io.EOF):
		return true, errors.New("empty JSON body")

	case errors.As(err, &sizeErr):
		return true, fmt.Errorf("JSON payload exceeds size limit (%d bytes)", sizeErr.Limit)

	default:
		return false, err
	}
}

// func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
//     if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
//         if isJson, jsonErr := IsJsonError(err); isJson {
//             // Return 400 with specific JSON error
//             return NewAPIError(http.StatusBadRequest, jsonErr.Error())
//         }
//         // Return 500 for other errors
//         return ErrInternalServerError.Err(err)
//     }
//     return nil
// }

// Example error outputs:
// - "malformed JSON at position 42: invalid character '}'"
// - "invalid JSON type for field 'age' at position 102 (got 'string'): expected number"
// - "JSON payload exceeds size limit (1048576 bytes)"
