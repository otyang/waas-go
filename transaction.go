package waas

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Transaction errors
var (
	ErrInvalidTransactionType   = NewWaasError("invalid transaction type")
	ErrInvalidTransactionStatus = NewWaasError("invalid transaction status")
	ErrInvalidTransactionObject = NewWaasError("invalid transaction: nil transaction object")
	ErrUnsupportedReversalType  = NewWaasError("unsupported transaction type for reversal")
	ErrAlreadyReversedTx        = NewWaasError("cannot reverse an already reversed/settled transaction")
)

type Transaction struct {
	ID             string            `json:"id" bun:"id,pk"`
	CustomerID     string            `json:"customerId" bun:",notnull"`
	WalletID       string            `json:"walletId" bun:",notnull"`
	IsDebit        bool              `json:"isDebit" bun:",notnull"`
	Currency       string            `json:"currency" bun:",notnull"`
	Amount         decimal.Decimal   `json:"amount" bun:"type:decimal(24,8),notnull"`
	Fee            decimal.Decimal   `json:"fee" bun:"type:decimal(24,8),notnull"`
	TotalAmount    decimal.Decimal   `json:"totalAmount" bun:"type:decimal(24,8),notnull"`
	BalanceAfter   decimal.Decimal   `json:"balanceAfter" bun:"type:decimal(24,8),notnull"`
	Type           TransactionType   `json:"type" bun:",notnull"`
	Status         TransactionStatus `json:"status" bun:",notnull"`
	Narration      *string           `json:"narration" bun:",notnull"`
	Reversed       bool              `json:"reversed"`
	CounterpartyID *string           `json:"counterpartyId"`
	CreatedAt      time.Time         `json:"createdAt" bun:",notnull"`
	UpdatedAt      time.Time         `json:"updatedAt" bun:",notnull"`
}

// SetNarration sets the narration of the transaction.
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

func (t *Transaction) canBeReversed() error {
	if t == nil {
		return ErrInvalidTransactionObject
	}

	if t.Reversed {
		return ErrAlreadyReversedTx
	}

	// only withdraw or debit transaction can be reversed
	if t.Type != TransactionTypeWithdrawal {
		return ErrUnsupportedReversalType
	}

	return nil
}

func (t *Transaction) Reverse(wallet *Wallet) (*ReverseResponse, error) {
	if err := t.canBeReversed(); err != nil {
		return nil, err
	}

	var newEntry *Transaction

	if t.IsDebit {
		if err := wallet.CreditBalance(t.Amount, t.Fee); err != nil {
			return nil, err
		}

		newEntry = NewTransactionForCreditEntry(wallet, t.Amount, t.Fee, t.Type)
	} else {
		if err := wallet.DebitBalance(t.Amount, t.Fee); err != nil {
			return nil, err
		}

		newEntry = NewTransactionForDebitEntry(wallet, t.Amount, t.Fee, t.Type, TransactionStatusSuccess)
	}

	newEntry.SetCounterpartyID(t.ID)
	newEntry.Reversed = true
	t.Reversed = true
	t.Status = TransactionStatusFailed
	t.SetCounterpartyID(newEntry.ID)

	return &ReverseResponse{
		OldTx:  t,
		NewTx:  newEntry,
		Wallet: wallet,
	}, nil
}

// Creates a new credit transaction entry.
func NewTransactionForCreditEntry(wallet *Wallet, amount, fee decimal.Decimal, txnType TransactionType) *Transaction {
	return &Transaction{
		ID:             NewTransactionID(),
		CustomerID:     wallet.CustomerID,
		WalletID:       wallet.ID,
		Currency:       wallet.CurrencyCode,
		IsDebit:        false,
		Amount:         amount,
		Fee:            fee,
		TotalAmount:    amount,
		BalanceAfter:   wallet.AvailableBalance,
		Type:           txnType,
		Status:         TransactionStatusSuccess,
		Narration:      nil,
		CounterpartyID: nil,
		Reversed:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Creates a new debit transaction entry.
func NewTransactionForDebitEntry(wallet *Wallet, amount, fee decimal.Decimal, txnType TransactionType, txnStatus TransactionStatus,
) *Transaction {
	return &Transaction{
		ID:             NewTransactionID(),
		CustomerID:     wallet.CustomerID,
		WalletID:       wallet.ID,
		Currency:       wallet.CurrencyCode,
		IsDebit:        true,
		Amount:         amount,
		Fee:            fee,
		TotalAmount:    amount.Add(fee),
		BalanceAfter:   wallet.AvailableBalance,
		Type:           txnType,
		Status:         txnStatus,
		Narration:      nil,
		CounterpartyID: nil,
		Reversed:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Creates a new transaction for a transfer.
func NewTransactionForTransfer(fromWallet, toWallet *Wallet, amount, fee decimal.Decimal) (fromTxEntry, toTxEntry *Transaction) {
	// Since transfers are internal and always successful, set the status to success.
	de := NewTransactionForDebitEntry(fromWallet, amount, fee, TransactionTypeTransfer, TransactionStatusSuccess)
	ce := NewTransactionForCreditEntry(toWallet, amount, fee, TransactionTypeTransfer)

	de.SetCounterpartyID(ce.ID)
	ce.SetCounterpartyID(de.ID)

	return de, ce
}

func NewTransactionForSwap(fromWallet, toWallet *Wallet, debitAmount, creditAmount, fee decimal.Decimal) (fromTx, toTx *Transaction) {
	de := NewTransactionForDebitEntry(fromWallet, debitAmount, fee, TransactionTypeSwap, TransactionStatusSuccess)
	ce := NewTransactionForCreditEntry(toWallet, creditAmount, decimal.Zero, TransactionTypeSwap)

	de.SetCounterpartyID(ce.ID)
	ce.SetCounterpartyID(de.ID)

	return de, ce
}
