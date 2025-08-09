package types

import (
	"time"

	"github.com/shopspring/decimal"
)

// TransactionCategory defines the type of transaction being performed
type TransactionCategory string

const (
	CategoryDeposit    TransactionCategory = "DEPOSIT"    // Funds being added to wallet
	CategoryTransfer   TransactionCategory = "TRANSFER"   // Funds moving between wallets
	CategoryRefund     TransactionCategory = "REFUND"     // Funds being returned
	CategoryAdjustment TransactionCategory = "ADJUSTMENT" // Manual balance adjustment
	CategoryFee        TransactionCategory = "FEE"        // Transaction fee deduction
)

// TransactionType indicates the direction of funds movement
type TransactionType string

const (
	TypeCredit TransactionType = "CREDIT" // Funds being added
	TypeDebit  TransactionType = "DEBIT"  // Funds being deducted
)

// TransactionStatus represents the current state of a transaction
type TransactionStatus string

const (
	StatusCompleted TransactionStatus = "COMPLETED" // Successfully processed
	StatusFailed    TransactionStatus = "FAILED"    // Processing failed
	StatusPending   TransactionStatus = "PENDING"   // Awaiting processing
)

// TransactionHistory contains the complete record of a wallet transaction
type TransactionHistory struct {
	// ID is the unique identifier for this transaction record
	ID string `json:"id"`

	// WalletID identifies the wallet involved in the transaction
	WalletID string `json:"walletId"`

	// CurrencyCode specifies the wallet's currency at time of transaction
	CurrencyCode string `json:"currencyCode"`

	// InitiatorID identifies who initiated the transaction
	InitiatorID string `json:"initiatorId"`

	// ExternalReference is a reference ID from an external system
	ExternalReference string `json:"externalReference"`

	// Category classifies the type of transaction
	Category TransactionCategory `json:"category"`

	// Description provides human-readable context for the transaction
	Description string `json:"description"`

	// Amount is the principal value moved in this transaction
	Amount decimal.Decimal `json:"amount"`

	// Fee is the processing fee deducted (if any)
	Fee decimal.Decimal `json:"fee"`

	// Type indicates the direction of funds (credit/debit)
	Type TransactionType `json:"type"`

	// BalanceBefore is the wallet's available balance before the transaction
	BalanceBefore decimal.Decimal `json:"balanceBefore"`

	// BalanceAfter is the wallet's available balance after the transaction
	BalanceAfter decimal.Decimal `json:"balanceAfter"`

	// CreatedAt is when the transaction was first requested
	CreatedAt time.Time `json:"initiatedAt"`

	// CompletedAt is when the transaction was finalized
	UpdatedAt time.Time `json:"completedAt"`

	// Status indicates the final disposition of the transaction
	Status TransactionStatus `json:"status"`
}
