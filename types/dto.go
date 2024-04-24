package types

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	// CreditWalletOpts defines parameters for crediting a wallet.
	CreditWalletOpts struct {
		WalletID  string
		Amount    decimal.Decimal
		Fee       decimal.Decimal
		Type      TransactionType
		Narration string
	}

	CreditWalletResponse struct {
		Wallet      *Wallet
		Transaction *Transaction
	}

	// DebitWalletOpts defines parameters for debiting a wallet.
	DebitWalletOpts struct {
		WalletID  string
		Amount    decimal.Decimal
		Fee       decimal.Decimal
		Type      TransactionType
		Status    TransactionStatus
		Narration string
	}

	DebitWalletResponse struct {
		Wallet      *Wallet
		Transaction *Transaction
	}

	// TransferOpts defines parameters for transferring funds between wallets.
	TransferRequestOpts struct {
		FromWalletID string          `json:"fromWid"`
		ToWalletID   string          `json:"toWid"`
		Amount       decimal.Decimal `json:"amount"`
		Fee          decimal.Decimal `json:"fee"`
		Narration    string          `json:"narration"`
	}

	TransferResponse struct {
		FromWallet      *Wallet
		ToWallet        *Wallet
		FromTransaction *Transaction
		ToTransaction   *Transaction
	}

	// ReverseOpts defines parameters for reversing a transaction.
	ReverseRequestOpts struct {
		TransactionID string `json:"transactionId"`
	}

	ReverseResponse struct {
		OldTx  *Transaction
		NewTx  *Transaction
		Wallet *Wallet
	}

	// SwapOpts defines parameters for swapping currencies between wallets.
	SwapRequestOpts struct {
		CustomerID       string
		FromCurrencyCode string
		ToCurrencyCode   string
		FromAmount       decimal.Decimal
		FromFee          decimal.Decimal
		ToAmount         decimal.Decimal
	}

	SwapWalletResponse struct {
		FromWallet      *Wallet
		ToWallet        *Wallet
		FromTransaction *Transaction
		ToTransaction   *Transaction
	}

	ListWalletsFilterOpts struct {
		CustomerID    string
		CurrencyCodes []string
		Status        WalletStatus
	}

	ListTransactionsFilterOpts struct {
		Limit      int
		StartDate  time.Time
		EndDate    time.Time
		Direction  string
		CustomerID string
		WalletID   string
		Currency   []string
		IsDebit    *bool
		Type       *TransactionType
		Status     *TransactionStatus
		Reversed   *bool
	}
)

const (
	DirectionAsc  = "asc"
	DirectionDesc = "Desc"
)
