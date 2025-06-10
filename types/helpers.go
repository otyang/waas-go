package types

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// ========================================= Helpers
type CreditOrDebitWalletOption struct {
	Amount                         decimal.Decimal
	Fee                            decimal.Decimal
	PendTransaction                bool
	TxnCategory                    TransactionCategory
	Status                         TransactionStatus
	Narration                      *string `json:"narration"`
	OptionalUseThisAsTransactionID string  // if empty it autogenerates new id
	OptionalLinkedTxnID            *string
}

func (x *CreditOrDebitWalletOption) Validate() error {
	if x.TxnCategory == "" {
		return errors.New("transaction category parameter shouldn't be empty")
	}

	if x.Narration == nil {
		return errors.New("transaction narration parameter shouldn't be empty")
	}

	if x.PendTransaction {
		x.Status = TransactionStatusPending
	}

	return nil
}

func newTransactionSummary(wallet *Wallet, opts CreditOrDebitWalletOption, isDebit bool) Transaction {
	if strings.TrimSpace(opts.OptionalUseThisAsTransactionID) != "" {
		opts.OptionalUseThisAsTransactionID = NewTransactionID()
	}

	var totalAmount decimal.Decimal
	if isDebit {
		totalAmount = opts.Amount.Add(opts.Fee)
	} else {
		totalAmount = opts.Amount.Sub(opts.Fee)
	}

	return Transaction{
		ID:           opts.OptionalUseThisAsTransactionID,
		CustomerID:   wallet.CustomerID,
		WalletID:     wallet.ID,
		IsDebit:      isDebit,
		Currency:     wallet.CurrencyCode,
		Amount:       opts.Amount,
		Fee:          opts.Fee,
		Total:        totalAmount,
		BalanceAfter: wallet.AvailableBalance,
		Category:     opts.TxnCategory,
		Status:       opts.Status,
		Narration:    opts.Narration,
		ServiceTxnID: nil,
		LinkedTxnID:  opts.OptionalLinkedTxnID,
		ReversedAt:   nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Debit Balance from a wallet and generates Transaction.
func DebitBalanceWithTxn(wlt *Wallet, opts CreditOrDebitWalletOption) (*Transaction, *Wallet, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	err := wlt.DebitBalance(opts.Amount, opts.Fee)
	if err != nil {
		return nil, nil, err
	}

	tx := newTransactionSummary(wlt, opts, true)

	return &tx, wlt, nil
}

// Credit Balance to a wallet and generates Transaction.
func CreditBalanceWithTxn(wlt *Wallet, opts CreditOrDebitWalletOption) (*Transaction, *Wallet, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	err := wlt.CreditBalance(opts.Amount, opts.Fee)
	if err != nil {
		return nil, nil, err
	}

	tx := newTransactionSummary(wlt, opts, false)
	return &tx, wlt, nil
}

// Credit Balance to a wallet and generates Transaction.
func ReverseTxWithTxn(txn *Transaction, wlt *Wallet) (*Transaction, *Wallet, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	if !txn.ReversedAt.IsZero() {
		//
	}

	reversedAt := time.Now()
	txn.Status = TransactionStatusFailed
	txn.ReversedAt = &reversedAt

	if txn.IsDebit {
		rT, rW, err := CreditBalanceWithTxn(wlt, CreditOrDebitWalletOption{
			Amount:          txn.Amount,
			Fee:             txn.Fee,
			PendTransaction: false,
			TxnCategory:     txn.Category,
			Status:          TransactionStatusSuccess, //Reversal are alwayssucceful
			Narration:       txn.Narration,
		})
		if err != nil {
			return nil, nil, err
		}

		return txn, rT, rW, err
	}

	if !txn.IsDebit {
		rT, rW, err := DebitBalanceWithTxn(wlt, CreditOrDebitWalletOption{
			Amount:          txn.Amount,
			Fee:             txn.Fee,
			PendTransaction: false,
			TxnCategory:     txn.Category,
			Status:          TransactionStatusSuccess, //Reversal are alwayssucceful
			Narration:       txn.Narration,
		})
		if err != nil {
			return nil, nil, err
		}

		return txn, rT, rW, err
	}

	tx := newTransactionSummary(wlt, opts, false)
	return &tx, wlt, nil
}

// =============================== transfer helpers
type TransferWalletOption struct {
	Amount          decimal.Decimal
	Fee             decimal.Decimal
	PendTransaction bool
	TxnCategory     TransactionCategory
	Status          TransactionStatus
	Narration       *string `json:"narration"`
}

func (x *TransferWalletOption) Validate() error {
	x.TxnCategory = "transfer"

	if x.Narration == nil {
		return errors.New("transaction narration shouldn't be empty")
	}

	if x.PendTransaction {
		x.Status = TransactionStatusPending
	}

	return nil
}

func (x *TransferWalletOption) ToTxnSummary() CreditOrDebitWalletOption {
	return CreditOrDebitWalletOption{
		Amount:          x.Amount,
		Fee:             x.Fee,
		PendTransaction: x.PendTransaction,
		TxnCategory:     x.TxnCategory,
		Status:          x.Status,
		Narration:       x.Narration,
	}
}

func TransferWithTxn(fromWallet *Wallet, toWallet *Wallet, opts TransferWalletOption) (*Transaction, *Transaction, error) {
	opts.TxnCategory = "transfer"

	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	err := fromWallet.TransferTo(toWallet, opts.Amount, opts.Fee)
	if err != nil {
		return nil, nil, err
	}

	txF := newTransactionSummary(fromWallet, opts.ToTxnSummary(), true) // FromTransaction
	txT := newTransactionSummary(toWallet, opts.ToTxnSummary(), false)  // ToTransaction

	txF.SetLinkedTxnID(txT.ID, false) // set linked ID
	txT.SetLinkedTxnID(txF.ID, false) // set linked ID

	return &txF, &txT, nil
}

// =============================== swap helpers
type SwapWalletOption struct {
	FromAmount      decimal.Decimal
	ToAmount        decimal.Decimal
	Fee             decimal.Decimal
	PendTransaction bool
	TxnCategory     TransactionCategory
	Status          TransactionStatus
	Narration       *string `json:"narration"`
}

func (x *SwapWalletOption) Validate() error {
	return nil
}

func (x *SwapWalletOption) ToDebitWalletParams() CreditOrDebitWalletOption {
	return CreditOrDebitWalletOption{
		Amount:                         x.FromAmount,
		Fee:                            x.Fee,
		PendTransaction:                x.PendTransaction,
		TxnCategory:                    x.TxnCategory,
		Status:                         x.Status,
		Narration:                      x.Narration,
		OptionalUseThisAsTransactionID: NewTransactionID(),
	}
}

func (x *SwapWalletOption) ToCreditWalletParams() CreditOrDebitWalletOption {
	return CreditOrDebitWalletOption{
		Amount:                         x.ToAmount,
		Fee:                            decimal.Zero,
		PendTransaction:                x.PendTransaction,
		TxnCategory:                    x.TxnCategory,
		Status:                         x.Status,
		Narration:                      x.Narration,
		OptionalUseThisAsTransactionID: NewTransactionID(),
	}
}

func SwapWithTxn(fromWallet *Wallet, toWallet *Wallet, opts SwapWalletOption) (*Transaction, *Transaction, error) {
	opts.TxnCategory = "swap"

	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	err := fromWallet.Swap(toWallet, opts.FromAmount, opts.ToAmount, opts.Fee)
	if err != nil {
		return nil, nil, err
	}

	txF := newTransactionSummary(fromWallet, opts.ToDebitWalletParams(), true) // FromTransaction
	txT := newTransactionSummary(toWallet, opts.ToCreditWalletParams(), false) // ToTransaction

	txF.SetLinkedTxnID(txT.ID, false) // set linked ID
	txT.SetLinkedTxnID(txF.ID, false) // set linked ID

	return &txF, &txT, nil
}
