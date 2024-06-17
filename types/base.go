package types

import (
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
)

// WalletError provides a common base for wallet-related errors.
type WaasError struct{ msg string }

func (e *WaasError) Error() string { return e.msg }

// NewWaasError creates a new WalletError.
func NewWaasError(msg string) *WaasError {
	return &WaasError{msg: msg}
}

// IsWaasError checks if an error is of type WaasError.
func IsWaasError(err error) bool {
	_, ok := err.(*WaasError)
	return ok
}

func NewTransactionID() string {
	return GenerateID("txn_"+time.Now().UTC().Format("20060102")+"_", 8)
}

func GenerateID(prefix string, size int) string {
	return prefix + gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", size)
}

// makes a slice of strings insensitive
func ToLowercaseSlice(strs []string) []string {
	lowercaseStrs := make([]string, len(strs))
	for i, str := range strs {
		lowercaseStrs[i] = strings.TrimSpace(strings.ToLower(str))
	}
	return lowercaseStrs
}
