package waas

import (
	"errors"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
)

// Defines a new error type for common errors encountered in the wallet module.
type OperationError struct{ error }

func NewOperationError(msg string) error { return &OperationError{errors.New(msg)} }

// IsOperationError checks if an error is of type OperationError.
func IsOperationError(err error) bool {
	_, ok := err.(*OperationError)
	return ok
}

func NewTransactionID() string {
	return time.Now().UTC().Format("20060102") + GenerateID(6)
}

func GenerateID(size int) string {
	return gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", size)
}

type (
	TransactionType   string
	TransactionStatus string
)

const (
	TransactionTypeSwap       TransactionType = "swap"
	TransactionTypeTransfer   TransactionType = "transfer"
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
)

const (
	TransactionStatusNew     TransactionStatus = "new"
	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusFailed  TransactionStatus = "failed"
	TransactionStatusSuccess TransactionStatus = "success"
)

func (t TransactionType) IsValid() error {
	switch t {
	case
		TransactionTypeSwap, TransactionTypeTransfer, TransactionTypeDeposit, TransactionTypeWithdrawal:
		return nil
	}
	return ErrInvalidTransactionType
}

func (t TransactionStatus) IsValid() error {
	switch t {
	case TransactionStatusNew, TransactionStatusPending, TransactionStatusFailed, TransactionStatusSuccess:
		return nil
	}
	return ErrInvalidTransactionStatus
}
