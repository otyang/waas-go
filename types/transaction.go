package types

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type (
	TransactionType   string
	TransactionStatus string
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
	ID             string            `json:"id" bun:"id,pk"`
	CustomerID     string            `json:"customerId" bun:",notnull"`
	WalletID       string            `json:"walletId" bun:",notnull"`
	IsDebit        bool              `json:"isDebit" bun:",notnull"`
	Currency       string            `json:"currency" bun:",notnull"`
	Amount         decimal.Decimal   `json:"amount" bun:"type:decimal(24,8),notnull"`
	Fee            decimal.Decimal   `json:"fee" bun:"type:decimal(24,8),notnull"`
	Total          decimal.Decimal   `json:"total" bun:"type:decimal(24,8),notnull"`
	BalanceAfter   decimal.Decimal   `json:"balanceAfter" bun:"type:decimal(24,8),notnull"`
	Type           TransactionType   `json:"type" bun:",notnull"`
	Status         TransactionStatus `json:"status" bun:",notnull"`
	Narration      *string           `json:"narration"`
	CounterpartyID *string           `json:"counterpartyId"`
	ReversedAt     *time.Time        `json:"reversedAt"`
	CreatedAt      time.Time         `json:"createdAt" bun:",notnull"`
	UpdatedAt      time.Time         `json:"updatedAt" bun:",notnull"`
}

// SetCounterpartyID sets the counterparty ID of the transaction.
func (t *Transaction) SetCounterpartyID(id string, reversed bool) *Transaction {
	if trimmedID := strings.TrimSpace(id); trimmedID != "" {
		t.CounterpartyID = &trimmedID

		if reversed {
			_time := time.Now()
			t.ReversedAt = &_time
		}
	}
	return t
}

type TxnSummaryParams struct {
	TransactionID     string
	IsDebit           bool
	Wallet            *Wallet
	Amount            decimal.Decimal
	Fee               decimal.Decimal
	TotalAmount       decimal.Decimal
	TransactionType   TransactionType
	TransactionStatus TransactionStatus
	Narration         string
}

// New transaction
func NewTransactionSummary(params TxnSummaryParams) *Transaction {
	tx := &Transaction{
		ID:             params.TransactionID,
		CustomerID:     params.Wallet.CustomerID,
		WalletID:       params.Wallet.ID,
		IsDebit:        params.IsDebit,
		Currency:       params.Wallet.CurrencyCode,
		Amount:         params.Amount,
		Fee:            params.Fee,
		Total:          params.TotalAmount,
		BalanceAfter:   params.Wallet.TotalBalance(),
		Type:           params.TransactionType,
		Status:         params.TransactionStatus,
		Narration:      nil,
		CounterpartyID: nil,
		ReversedAt:     nil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if strings.TrimSpace(tx.ID) == "" {
		tx.ID = NewTransactionID()
	}

	if strings.TrimSpace(params.Narration) != "" {
		tx.Narration = &params.Narration
	}

	return tx
}
