package types

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeSwap       TransactionType = "SWAP"
	TransactionTypeTransfer   TransactionType = "TRANSFER"
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
)

type TransactionStatus string

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

func (t *Transaction) SetNarration(narration string) *Transaction {
	if trimmedNarration := strings.TrimSpace(narration); trimmedNarration != "" {
		t.Narration = &trimmedNarration
	}
	return t
}

// SetCounterpartyID sets the counterparty ID of the transaction.
func (t *Transaction) SetCounterpartyID(id string) *Transaction {
	if trimmedID := strings.TrimSpace(id); trimmedID != "" {
		t.CounterpartyID = &trimmedID
	}
	return t
}

func (t *Transaction) SetReversedAt(time time.Time) *Transaction {
	t.ReversedAt = &time
	return t
}

func (t *Transaction) CanBeReversed() error {
	if t == nil {
		return ErrInvalidTransactionObject
	}

	if t.ReversedAt != nil {
		return ErrTransactionAlreadyReversed
	}

	// only withdraw or debit transaction can be reversed
	if t.Type != TransactionTypeWithdrawal {
		return ErrTransactionUnsupportedReversalType
	}

	return nil
}

// New transaction
func NewTransaction(
	isCredit bool,
	wallet *Wallet,
	amount decimal.Decimal,
	fee decimal.Decimal,
	totalAmount decimal.Decimal,
	txnType TransactionType,
	txnStatus TransactionStatus,
) *Transaction {
	return &Transaction{
		ID:             NewTransactionID(),
		CustomerID:     wallet.CustomerID,
		WalletID:       wallet.ID,
		IsDebit:        !isCredit,
		Currency:       wallet.CurrencyCode,
		Amount:         amount,
		Fee:            fee,
		Total:          totalAmount,
		BalanceAfter:   wallet.AvailableBalance,
		Type:           txnType,
		Status:         txnStatus,
		Narration:      nil,
		CounterpartyID: nil,
		ReversedAt:     nil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
