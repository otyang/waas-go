package waas

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// Transaction errors
var (
	ErrInvalidTransactionType   = NewWaasError("invalid transaction type")
	ErrInvalidTransactionStatus = NewWaasError("invalid transaction status")
	ErrInvalidTransactionObject = NewWaasError("invalid transaction: nil transaction object")
	ErrUnsupportedReversalType  = NewWaasError("unsupported transaction type for reversal")
	ErrReverseSettledTx         = NewWaasError("cannot reverse an already reversed/settled transaction")
)

type Transaction struct {
	mutex          sync.Mutex
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
	IdempotencyId  string            `json:"idempotencyId" bun:",notnull"` //(used during reversal)
	CreatedAt      time.Time         `json:"createdAt" bun:",notnull"`
	UpdatedAt      time.Time         `json:"updatedAt" bun:",notnull"`
}

// SetNarration sets the narration of the transaction.
func (t *Transaction) SetNarration(narration string) *Transaction {
	t.Narration = &narration
	return t
}

// SetCounterpartyID sets the counterparty ID of the transaction.
func (t *Transaction) SetCounterpartyID(id string) *Transaction {
	t.CounterpartyID = &id
	return t
}

func (t *Transaction) CanBeReversed() error {
	if t == nil {
		return ErrInvalidTransactionObject
	}

	if t.Reversed {
		return ErrReverseSettledTx
	}

	// only withdraw or debit transaction can be reversed
	if t.Type != TransactionTypeWithdrawal {
		return ErrUnsupportedReversalType
	}

	return nil
}

func (t *Transaction) Reverse(wallet *Wallet) (*ReverseResponse, error) {
	// t.mutex.Lock()
	// defer t.mutex.Unlock()

	if err := t.CanBeReversed(); err != nil {
		return nil, err
	}

	if t.IsDebit {
		if err := wallet.CreditBalance(t.Amount, t.Fee); err != nil {
			return nil, err
		}

		t.Reversed = true
		t.Status = TransactionStatusFailed
	} else {
		if err := wallet.DebitBalance(t.Amount, t.Fee); err != nil {
			return nil, err
		}

		t.Reversed = true
		t.Status = TransactionStatusFailed
	}

	newEntry, err := NewTransactionEntry(wallet, !t.IsDebit, t.Amount, t.Fee, t.Type, TransactionStatusSuccess)
	if err != nil {
		return nil, err
	}

	newEntry.Reversed = true
	newEntry.SetCounterpartyID(t.ID)
	t.SetCounterpartyID(newEntry.ID)

	return &ReverseResponse{
		OldUpdatedTx:  t,
		NewTx:         newEntry,
		UpdatedWallet: wallet,
	}, nil
}

// =====

// ==

// newTransactionEntry creates a new transaction entry.
func NewTransactionEntry(
	wallet *Wallet, forCredit bool, amount decimal.Decimal, fee decimal.Decimal, txnType TransactionType, txnStatus TransactionStatus,
) (*Transaction, error) {
	if wallet == nil {
		return nil, ErrWalletInvalid
	}

	if err := txnStatus.IsValid(); err != nil {
		return nil, err
	}

	if err := txnType.IsValid(); err != nil {
		return nil, err
	}

	return &Transaction{
		mutex:          sync.Mutex{},
		ID:             NewTransactionID(),
		CustomerID:     wallet.CustomerID,
		WalletID:       wallet.ID,
		Currency:       wallet.CurrencyCode,
		IsDebit:        !forCredit,
		Amount:         amount,
		Fee:            fee,
		TotalAmount:    amount.Add(fee), // ==='
		BalanceAfter:   wallet.AvailableBalance,
		Type:           txnType,
		Status:         txnStatus,
		Narration:      nil,
		CounterpartyID: nil,
		Reversed:       false,
		IdempotencyId:  GenerateID(7),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

// Creates a new credit transaction entry.
func NewTransactionForCreditEntry(wallet *Wallet, amount, fee decimal.Decimal, txnType TransactionType) (*Transaction, error) {
	return NewTransactionEntry(wallet, true, amount, fee, txnType, TransactionStatusSuccess)
}

// Creates a new debit transaction entry.
func NewTransactionForDebitEntry(
	wallet *Wallet, amount, fee decimal.Decimal, txnType TransactionType, txnStatus TransactionStatus,
) (*Transaction, error) {
	return NewTransactionEntry(wallet, false, amount, fee, txnType, txnStatus)
}

// Creates a new transaction for a transfer.
func NewTransactionForTransfer(
	fromWallet, toWallet *Wallet, amount, fee decimal.Decimal,
) (fromTxEntry, toTxEntry *Transaction, err error) {
	// Since transfers are internal and always successful, set the status to success.
	de, err := NewTransactionEntry(fromWallet, false, amount, fee, TransactionTypeTransfer, TransactionStatusSuccess)
	if err != nil {
		return nil, nil, err
	}

	ce, err := NewTransactionEntry(toWallet, true, amount, fee, TransactionTypeTransfer, TransactionStatusSuccess)
	if err != nil {
		return nil, nil, err
	}

	de.SetCounterpartyID(ce.ID)
	ce.SetCounterpartyID(de.ID)

	return de, ce, nil
}

func NewTransactionForSwap(
	fromWallet, toWallet *Wallet, debitAmount, creditAmount, fee decimal.Decimal,
) (fromTxEntry, toTxEntry *Transaction, err error) {
	de, err := NewTransactionEntry(fromWallet, false, debitAmount, fee, TransactionTypeSwap, TransactionStatusSuccess)
	if err != nil {
		return nil, nil, err
	}

	ce, err := NewTransactionEntry(toWallet, true, creditAmount, decimal.Zero, TransactionTypeSwap, TransactionStatusSuccess)
	if err != nil {
		return nil, nil, err
	}

	de.SetCounterpartyID(ce.ID)
	ce.SetCounterpartyID(de.ID)

	return de, ce, nil
}
