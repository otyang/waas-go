package types

import "errors"

// Wallet state errors
var (
	ErrWalletClosed        = errors.New("wallet is closed and cannot perform transactions")
	ErrWalletAlreadyClosed = errors.New("wallet is already closed")
	ErrWalletNotClosed     = errors.New("wallet is not closed")
	ErrWalletFrozen        = errors.New("wallet is frozen and cannot perform debit transactions")
	ErrWalletAlreadyFrozen = errors.New("wallet is already frozen")
	ErrWalletNotFrozen     = errors.New("wallet is not frozen")
)

// Transaction amount errors
var (
	ErrInvalidAmount        = errors.New("transaction amount must be positive")
	ErrInvalidFee           = errors.New("transaction fee cannot be negative")
	ErrInsufficientFunds    = errors.New("insufficient available balance for transaction")
	ErrInsufficientLien     = errors.New("insufficient lien balance for operation")
	ErrExchangeRateMismatch = errors.New("exchange rate does not match amount conversion")
	ErrInvalidExchangeRate  = errors.New("exchange rate must be positive")
)

// Transaction processing errors
var (
	ErrCurrencyMismatch     = errors.New("currency codes do not match for operation")
	ErrTransactionFailed    = errors.New("transaction processing failed")
	ErrDuplicateTransaction = errors.New("transaction with same ID already exists")
)

// Validation errors
var (
	ErrInvalidWalletID    = errors.New("invalid wallet identifier")
	ErrInvalidCustomerID  = errors.New("invalid customer identifier")
	ErrInvalidCurrency    = errors.New("invalid currency code")
	ErrInvalidDescription = errors.New("transaction description is required")
)

// Lien operation errors
var (
	ErrLienAlreadyExists = errors.New("lien with same reference already exists")
	ErrLienNotFound      = errors.New("no matching lien found for release")
)

// Security errors
var (
	ErrUnauthorizedAccess = errors.New("unauthorized wallet access")
	ErrInvalidSignature   = errors.New("invalid transaction signature")
)

// ErrWalletNotEmpty indicates an operation was attempted that requires
// a zero balance, but the wallet still contains funds
var ErrWalletNotEmpty = errors.New("wallet must have zero balance to perform this operation")
