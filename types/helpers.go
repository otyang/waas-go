package types

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// ==== Helpers

type CreditOrDebitWalletOption struct {
	Amount                 decimal.Decimal
	Fee                    decimal.Decimal
	PendTransaction        bool
	TxnCategory            TransactionCategory
	Status                 TransactionStatus
	Narration              *string `json:"narration"`
	UseThisAsTransactionID string
}

func (x *CreditOrDebitWalletOption) Validate() error {
	return nil
}

func newTransactionSummary(wallet *Wallet, opts CreditOrDebitWalletOption, isDebit bool) Transaction {
	if strings.TrimSpace(opts.UseThisAsTransactionID) != "" {
		opts.UseThisAsTransactionID = NewTransactionID()
	}

	var totalAmount decimal.Decimal
	if isDebit {
		totalAmount = opts.Amount.Add(opts.Fee)
	} else {
		totalAmount = opts.Amount.Sub(opts.Fee)
	}

	return Transaction{
		ID:           opts.UseThisAsTransactionID,
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

func TransferWithTxn(fromWallet *Wallet, toWallet *Wallet, opts CreditOrDebitWalletOption) (*Transaction, *Transaction, error) {
	opts.TxnCategory = "transfer"

	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	err := fromWallet.TransferTo(toWallet, opts.Amount, opts.Fee)
	if err != nil {
		return nil, nil, err
	}

	txF := newTransactionSummary(fromWallet, opts, true) // FromTransaction
	txT := newTransactionSummary(toWallet, opts, false)  // ToTransaction

	return &txF, &txT, nil
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

	txF := newTransactionSummary(fromWallet, CreditOrDebitWalletOption{""}, true) // FromTransaction
	txT := newTransactionSummary(toWallet, CreditOrDebitWalletOption{}, false)    // ToTransaction

	return &txF, &txT, nil
}

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
