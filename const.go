package waas

import (
	"errors"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
)

// Defines a new error type for common errors encountered in the wallet module.
type WaasError struct{ error }

func NewWaasError(msg string) error { return &WaasError{errors.New(msg)} }

// IsWaasError checks if an error is of type WaasError.
func IsWaasError(err error) bool {
	_, ok := err.(*WaasError)
	return ok
}

func NewTransactionID() string {
	return time.Now().UTC().Format("20060102") + "_" + GenerateID(6)
}

func GenerateID(size int) string {
	// time.Now().UTC().Format("2006-01-02 15:04:05.999999")
	return gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", size)
}

type (
	TransactionType   string
	TransactionStatus string
)

const (
	TransactionTypeSwap       TransactionType = "SWAP"
	TransactionTypeTransfer   TransactionType = "TRANSFER"
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
)

const (
	TransactionStatusNew     TransactionStatus = "NEW"
	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusFailed  TransactionStatus = "FAILED"
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
)
