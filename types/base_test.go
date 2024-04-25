package types

import (
	"errors"
	"testing"
)

func TestNewWaasError(t *testing.T) {
	t.Parallel()

	err := NewWaasError("failed to transfer funds")

	// Check if the returned error is of type WaasError
	if !IsWaasError(err) {
		t.Errorf("NewWaasError should return a WaasError")
	}

	// Check the underlying error message
	if err.Error() != "failed to transfer funds" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestIsWaasError(t *testing.T) {
	t.Parallel()

	opErr := NewWaasError("operation failed")
	otherErr := errors.New("some other error")

	// Check if IsWaasError correctly identifies WaasError
	if !IsWaasError(opErr) {
		t.Errorf("IsWaasError should return true for WaasError")
	}

	// Check if IsWaasError correctly handles other error types
	if IsWaasError(otherErr) {
		t.Errorf("IsWaasError should return false for non-WaasError errors")
	}
}
