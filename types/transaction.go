package types

import (
	"errors"
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

// TransactionHistory contains a wallet transaction record
type TransactionHistory struct {
	ID                string              `json:"id" bun:",pk"`                                    // Unique transaction ID
	WalletID          string              `json:"walletId" bun:",notnull"`                         // Associated wallet ID
	CurrencyCode      string              `json:"currencyCode" bun:",notnull"`                     // Transaction currency
	InitiatorID       string              `json:"initiatorId" bun:",notnull"`                      // Who initiated the transaction
	ExternalReference string              `json:"externalReference" bun:",notnull"`                // External system reference
	Category          TransactionCategory `json:"category" bun:",notnull"`                         // Transaction type/category
	Description       string              `json:"description" bun:",notnull"`                      // Transaction description
	Amount            decimal.Decimal     `json:"amount" bun:",type:decimal(24,8),notnull"`        // Transaction amount
	Fee               decimal.Decimal     `json:"fee" bun:",type:decimal(24,8),notnull"`           // Processing fee
	Type              TransactionType     `json:"type" bun:",notnull"`                             // Credit/Debit
	BalanceBefore     decimal.Decimal     `json:"balanceBefore" bun:",type:decimal(24,8),notnull"` // Pre-transaction balance
	BalanceAfter      decimal.Decimal     `json:"balanceAfter" bun:",type:decimal(24,8),notnull"`  // Post-transaction balance
	CreatedAt         time.Time           `json:"initiatedAt" bun:",notnull"`                      // Creation timestamp
	UpdatedAt         time.Time           `json:"completedAt" bun:",notnull"`                      // Completion timestamp
	Status            TransactionStatus   `json:"status" bun:",notnull"`                           // Transaction status
}

// Transaction status transition errors
var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrTransactionCompleted    = errors.New("transaction already completed")
)

// MarkAsPending marks the transaction as pending if it's in a valid state
func (t *TransactionHistory) MarkAsPending() error {
	if t.Status == StatusCompleted || t.Status == StatusFailed {
		return ErrTransactionCompleted
	}

	t.Status = StatusPending
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkAsCompleted marks the transaction as successfully completed
func (t *TransactionHistory) MarkAsCompleted() error {
	// Only allow completion from pending state
	if t.Status != StatusPending {
		return ErrInvalidStatusTransition
	}

	t.Status = StatusCompleted
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkAsFailed marks the transaction as failed without modifying description
func (t *TransactionHistory) MarkAsFailed() error {
	// Allow failure from pending state or new transactions
	if t.Status != "" && t.Status != StatusPending {
		return ErrInvalidStatusTransition
	}

	t.Status = StatusFailed
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Revert marks a completed transaction as failed (for refunds/reversals)
func (t *TransactionHistory) Revert() error {
	if t.Status != StatusCompleted {
		return ErrInvalidStatusTransition
	}

	t.Status = StatusFailed
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// CanTransitionTo checks if a status transition is valid
func (t *TransactionHistory) CanTransitionTo(newStatus TransactionStatus) bool {
	switch t.Status {
	case StatusPending:
		return newStatus == StatusCompleted || newStatus == StatusFailed
	case "":
		return newStatus == StatusPending || newStatus == StatusFailed
	default:
		return false
	}
}
