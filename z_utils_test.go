package waas

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTransactionID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{
			name: "Should generate a valid transaction ID",
			want: `^\d{8}_\d{6}_\d{6}_\w{5}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTransactionID()
			matched, err := regexp.MatchString(tt.want, got)
			assert.NoError(t, err)
			assert.True(t, matched)
		})
	}
}

func TestNewOperationError(t *testing.T) {
	t.Parallel()

	err := NewOperationError("failed to transfer funds")

	// Check if the returned error is of type OperationError
	if !IsOperationError(err) {
		t.Errorf("NewOperationError should return an OperationError")
	}

	// Check the underlying error message
	if err.Error() != "failed to transfer funds" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestIsOperationError(t *testing.T) {
	t.Parallel()

	opErr := NewOperationError("operation failed")
	otherErr := errors.New("some other error")

	// Check if IsOperationError correctly identifies OperationError
	if !IsOperationError(opErr) {
		t.Errorf("IsOperationError should return true for OperationError")
	}

	// Check if IsOperationError correctly handles other error types
	if IsOperationError(otherErr) {
		t.Errorf("IsOperationError should return false for non-OperationError errors")
	}
}
