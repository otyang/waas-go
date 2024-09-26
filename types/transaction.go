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
	ID           string            `json:"id" bun:"id,pk"`
	CustomerID   string            `json:"customerId" bun:",notnull"`
	WalletID     string            `json:"walletId" bun:",notnull"`
	IsDebit      bool              `json:"isDebit" bun:",notnull"`
	Currency     string            `json:"currency" bun:",notnull"`
	Amount       decimal.Decimal   `json:"amount" bun:"type:decimal(24,8),notnull"`
	Fee          decimal.Decimal   `json:"fee" bun:"type:decimal(24,8),notnull"`
	Total        decimal.Decimal   `json:"total" bun:"type:decimal(24,8),notnull"`
	BalanceAfter decimal.Decimal   `json:"balanceAfter" bun:"type:decimal(24,8),notnull"`
	Type         TransactionType   `json:"type" bun:",notnull"`
	Status       TransactionStatus `json:"status" bun:",notnull"`
	Narration    *string           `json:"narration"`
	ServiceTxnID *string           `json:"serviceTxnId" bun:",notnull"`
	ReversedAt   *time.Time        `json:"reversedAt"`
	CreatedAt    time.Time         `json:"createdAt" bun:",notnull"`
	UpdatedAt    time.Time         `json:"updatedAt" bun:",notnull"`
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
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// New transaction
func NewTransactionSummary(params TxnSummaryParams) *Transaction {
	tx := &Transaction{
		ID:           params.TransactionID,
		CustomerID:   params.Wallet.CustomerID,
		WalletID:     params.Wallet.ID,
		IsDebit:      params.IsDebit,
		Currency:     params.Wallet.CurrencyCode,
		Amount:       params.Amount,
		Fee:          params.Fee,
		Total:        params.TotalAmount,
		BalanceAfter: params.Wallet.TotalBalance(),
		Type:         params.TransactionType,
		Status:       params.TransactionStatus,
		Narration:    nil,
		ServiceTxnID: nil,
		ReversedAt:   nil,
		CreatedAt:    params.CreatedAt,
		UpdatedAt:    params.UpdatedAt,
	}

	if strings.TrimSpace(tx.ID) == "" {
		tx.ID = NewTransactionID()
	}

	if strings.TrimSpace(params.Narration) != "" {
		tx.Narration = &params.Narration
	}

	if params.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}

	if params.UpdatedAt.IsZero() {
		tx.UpdatedAt = time.Now()
	}

	return tx
}

// Creates a new credit transaction entry.
func NewTransactionForCreditEntry(wallet *Wallet, amount, fee decimal.Decimal, txnType TransactionType) *Transaction {
	return &Transaction{
		ID:           NewTransactionID(),
		CustomerID:   wallet.CustomerID,
		WalletID:     wallet.ID,
		IsDebit:      false,
		Currency:     wallet.CurrencyCode,
		Amount:       amount,
		Fee:          fee,
		Total:        amount.Add(fee),
		BalanceAfter: wallet.AvailableBalance,
		Type:         txnType,
		Status:       TransactionStatusSuccess,
		Narration:    nil,
		ServiceTxnID: nil,
		ReversedAt:   nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
