package types

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type (
	TransactionCategory string
	TransactionStatus   string
)

const (
	TransactionStatusNew     TransactionStatus = "NEW"
	TransactionStatusFailed  TransactionStatus = "FAILED"
	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
)

// Transaction errors
var (
	ErrInvalidTransactionObject           = NewWaasError("invalid transaction: nil transaction object")
	ErrTransactionUnsupportedReversalType = NewWaasError("unsupported transaction type for reversal")
	ErrTransactionAlreadyReversed         = NewWaasError("cannot reverse an already reversed transaction")
)

type Transaction struct {
	ID           string              `json:"id" bun:"id,pk"`
	CustomerID   string              `json:"customerId" bun:",notnull"`
	WalletID     string              `json:"walletId" bun:",notnull"`
	IsDebit      bool                `json:"isDebit" bun:",notnull"`
	Currency     string              `json:"currency" bun:",notnull"`
	Amount       decimal.Decimal     `json:"amount" bun:"type:decimal(24,8),notnull"`
	Fee          decimal.Decimal     `json:"fee" bun:"type:decimal(24,8),notnull"`
	Total        decimal.Decimal     `json:"total" bun:"type:decimal(24,8),notnull"`
	BalanceAfter decimal.Decimal     `json:"balanceAfter" bun:"type:decimal(24,8),notnull"`
	Category     TransactionCategory `json:"category" bun:",notnull"`
	Status       TransactionStatus   `json:"status" bun:",notnull"`
	Narration    *string             `json:"narration"`
	ServiceTxnID *string             `json:"serviceTxnId"`
	LinkedTxnID  *string             `json:"linkedTxnId"`
	ReversedAt   *time.Time          `json:"reversedAt"`
	CreatedAt    time.Time           `json:"createdAt" bun:",notnull"`
	UpdatedAt    time.Time           `json:"updatedAt" bun:",notnull"`
}

// SetServiceTxnID sets the counterparty ID of the transaction.
func (t *Transaction) SetServiceTxnID(id string, reversed bool) *Transaction {
	if trimmedID := strings.TrimSpace(id); trimmedID != "" {
		t.ServiceTxnID = &trimmedID

		if reversed {
			_time := time.Now()
			t.ReversedAt = &_time
		}
	}
	return t
}

// SetServiceTxnID sets the counterparty ID of the transaction.
func (t *Transaction) SetLinkedTxnID(id string, reversed bool) *Transaction {
	if trimmedID := strings.TrimSpace(id); trimmedID != "" {
		t.LinkedTxnID = &trimmedID

		if reversed {
			_time := time.Now()
			t.ReversedAt = &_time
		}
	}
	return t
}
